# 部署到云服务器（公网发布）

当前生产架构：**阿里云 ECS（Ubuntu 22.04）+ systemd + Cloudflare 代理（Origin Rule 回源 8080）**。

## 1. 交叉编译 Linux 后端（在开发机上）

```bash
cd backend
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o ../deploy/lms-server ./cmd/server
```

（本仓库的 `deploy/lms-server` 不入库，需自行编译；`tools/lms-server.exe` 是 Windows 版。）

## 2. 上传并安装

```bash
# 开发机
scp -r deploy root@<你的服务器IP>:/root/

# 服务器
cd /root/deploy && chmod +x deploy.sh && sudo sh deploy.sh
curl http://127.0.0.1:8080/api/health      # {"status":"ok"}
```

服务安装为 `lms-server`，用户 `lms`，数据在 `/opt/lms/`，密钥在 `/opt/lms/lms.env`（600）。

## 3. 云侧配置

| 项 | 配置 |
|---|---|
| 安全组 | 入方向放行 **TCP 8080**（源 0.0.0.0/0） |
| DNS | `lms.<你的域名>` A 记录指向服务器公网 IP |
| SSL 模式 | Cloudflare SSL/TLS → **Flexible**（源站是 HTTP:8080） |
| Origin Rule | Rules → Origin Rules → `Hostname equals lms.<你的域名>` → **Destination Port → 8080** |

> **备案说明**：境内地域 ECS 上，域名未完成 ICP 备案时，云厂商会拦截以该域名发起的访问（**含 8080 等非标准端口**）。备案通过后自动恢复；备案期间可用 `http://<服务器IP>:8080` 直连。

## 4. 验证

```bash
curl https://lms.<你的域名>/api/health     # {"status":"ok"}
```

## 5. 运维

```bash
systemctl status lms-server
systemctl restart lms-server
journalctl -u lms-server -f
cp /opt/lms/library_data.db /opt/lms/backup-$(date +%F).db   # 先 systemctl stop 更稳妥
```
