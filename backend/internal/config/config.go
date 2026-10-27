// Package config 负责运行配置：监听地址、数据文件路径、令牌密钥等。
package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
)

// Config 保存后端运行参数。
type Config struct {
	Addr       string // HTTP 监听地址
	DBPath     string // SQLite 数据库路径
	TokenSecret []byte // 登录令牌 HMAC 密钥
}

// appDir 返回可执行文件（或开发时项目）所在目录，保证数据库与程序同目录、便于便携。
func appDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

// Load 从环境变量读取配置，未设置时使用默认值。
func Load() *Config {
	addr := os.Getenv("LMS_ADDR")
	if addr == "" {
		// 默认监听所有网卡，局域网内可用私网 IP 访问（不走 localhost）。
		addr = "0.0.0.0:8765"
	}
	dbPath := os.Getenv("LMS_DB")
	if dbPath == "" {
		// 优先找 exe 上级目录的 standalone/library_data.db（部署结构：tools/ 与 standalone/ 同级）。
		// 找不到才退化到 exe 同目录（便携模式或 go run 开发场景）。
		if exe, err := os.Executable(); err == nil {
			candidate := filepath.Join(filepath.Dir(exe), "..", "standalone", "library_data.db")
			if abs, err := filepath.Abs(candidate); err == nil {
				if _, err := os.Stat(abs); err == nil {
					dbPath = abs
				}
			}
		}
		if dbPath == "" {
			dbPath = filepath.Join(appDir(), "library_data.db")
		}
	}
	secret := os.Getenv("LMS_SECRET")
	if secret == "" {
		// 未配置密钥时随机生成（每次启动不同，重启后已签发令牌失效），
		// 避免使用可预测的固定默认值被用来伪造令牌。
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			log.Fatalf("无法生成随机 JWT 密钥: %v", err)
		}
		secret = hex.EncodeToString(buf)
		log.Println("[警告] 未设置 LMS_SECRET：已生成一次性随机密钥，重启后令牌失效；请通过环境变量或 tools/lms-secret.txt 固定配置")
	}
	return &Config{
		Addr:        addr,
		DBPath:      dbPath,
		TokenSecret: []byte(secret),
	}
}
