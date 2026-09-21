// Package main 是图书管理系统后端服务入口。
package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"lms/backend/internal/config"
	"lms/backend/internal/httpapi"
	"lms/backend/internal/service"
	"lms/backend/internal/store"
)

// isVirtual 判断网卡是否为虚拟 / 仅主机网卡（VirtualBox、VMware、Hyper-V 等），这些地址不能用于局域网访问。
func isVirtual(name string) bool {
	n := strings.ToLower(name)
	for _, kw := range []string{
		"virtualbox", "vbox", "vmware", "vmnet", "hyper-v", "vethernet",
		"loopback", "pseudo", "tap-windows", "wintun", "docker",
	} {
		if strings.Contains(n, kw) {
			return true
		}
	}
	return false
}

// primaryIP 返回通往默认路由的源 IPv4，即真实联网（WLAN/以太网）网卡的私网地址。
// UDP Dial 不会真正发包，仅让操作系统按路由表选出出口 IP。
func primaryIP() string {
	for _, target := range []string{"8.8.8.8:80", "114.114.114.114:80", "223.5.5.5:80"} {
		c, err := net.Dial("udp", target)
		if err != nil {
			continue
		}
		defer c.Close()
		if a, ok := c.LocalAddr().(*net.UDPAddr); ok && a.IP != nil {
			if ip4 := a.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return ""
}

// privateURLs 返回可用于访问的地址：真实联网 IP 排最前，其后是其他物理网卡，已排除虚拟网卡与环回/链路本地地址。
func privateURLs(addr string) []string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	add := func(ip string) {
		if ip == "" || seen[ip] {
			return
		}
		seen[ip] = true
		out = append(out, "http://"+ip+":"+port+"/")
	}

	add(primaryIP()) // 默认路由对应的真实私网 IP，作为首选（自动打开用）

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, ifc := range ifaces {
			if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 || isVirtual(ifc.Name) {
				continue
			}
			addrs, err := ifc.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addrs {
				var ip net.IP
				switch v := a.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil || ip.IsLoopback() {
					continue
				}
				if ip4 := ip.To4(); ip4 != nil && !ip4.IsLinkLocalUnicast() {
					add(ip4.String())
				}
			}
		}
	}
	return out
}

// openBrowser 调用系统默认浏览器打开指定 URL。
func openBrowser(u string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	case "darwin":
		return exec.Command("open", u).Start()
	default:
		return exec.Command("xdg-open", u).Start()
	}
}

func main() {
	cfg := config.Load()

	st, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	srv := httpapi.New(svc, cfg.TokenSecret)

	log.Printf("图书管理系统已启动，监听 %s（数据库：%s）", cfg.Addr, cfg.DBPath)
	urls := privateURLs(cfg.Addr)
	if len(urls) > 0 {
		log.Printf("请在浏览器中打开以下任一地址（页面与接口同端口），首选地址已自动打开：")
		for _, u := range urls {
			log.Printf("  %s", u)
		}
	} else {
		_, port, _ := net.SplitHostPort(cfg.Addr)
		log.Printf("未检测到联网 IPv4，可在本机用 http://127.0.0.1:%s/ 打开", port)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- http.ListenAndServe(cfg.Addr, srv.Handler()) }()
	time.Sleep(400 * time.Millisecond)

	// 自动打开“真实联网私网 IP”；IP 变动后重新双击程序即可。设 LMS_OPEN_BROWSER=0 可关闭。
	if os.Getenv("LMS_OPEN_BROWSER") != "0" && len(urls) > 0 {
		if err := openBrowser(urls[0]); err != nil {
			log.Printf("自动打开浏览器失败，请手动复制上面的首选地址：%v", err)
		}
	}

	if err := <-errCh; err != nil {
		log.Fatalf("服务退出: %v", err)
	}
}
