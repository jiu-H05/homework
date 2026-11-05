# 图书管理系统（Library Management System, LMS）

一个现代前后端分离的图书管理系统：Go 后端纯 REST API + SQLite，Tauri 瘦客户端（Windows）连接，
后端部署在阿里云 ECS 上由 systemd 常驻，通过自有域名 **https://lms.hyx0211.xyz** 对外提供服务
（Cloudflare 代理提供 HTTPS，Origin Rule 回源 8080）。

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go (net/http) + SQLite (WAL) + JWT + FTS5 全文检索 |
| 前端 | 原生 HTML5 / ES Modules / CSS，Apple Glass 风格 |
| 客户端 | Tauri 2（Windows，NSIS 打包） |
| 部署 | 阿里云 ECS（Ubuntu 22.04）+ systemd 常驻；Cloudflare 代理 + Origin Rule（回源 8080） |
| 数据 | 单文件 SQLite，应用内一键导出/导入 |

## 目录结构

```
lms-modern/
├── backend/             # Go 后端源码（纯 REST API，入口 cmd/server）
├── frontend/src/        # 前端源码（HTML/JS/CSS）
│   ├── js/              # api / net / views / components
│   └── src-tauri/       # Tauri Rust 壳
├── deploy/              # 服务器部署（systemd 单元 + deploy.sh + 说明）
├── standalone/          # 运行时数据库 library_data.db（已 gitignore）
└── tools/               # 开发与运维脚本
    ├── landing/              # 产品发布页静态站
    ├── start-lms-named.cmd   # 备用：本机 + Cloudflare 隧道方式启动
    └── start-public-server.cmd # 备用：临时隧道方式启动（仅测试）
```

## 运行方法

### 1. 本地启动后端

```bash
# 方式一：直接运行编译好的二进制
cd tools
lms-server.exe
# 默认监听 0.0.0.0:8765

# 方式二：从源码编译运行（入口在 backend/cmd/server）
cd backend
go run ./cmd/server
```

环境变量：

| 变量 | 默认 | 说明 |
|---|---|---|
| LMS_ADDR | 0.0.0.0:8765 | 监听地址（服务器上部署时用 0.0.0.0:8080） |
| LMS_DB | ../standalone/library_data.db | 数据库路径 |
| LMS_SECRET | 见 `tools/lms-secret.txt`（已 gitignore，勿提交） | JWT 密钥；未设置时启动随机生成，重启后令牌失效 |

验证：`http://localhost:8765/api/health` 返回 `{"status":"ok"}` 即正常。

### 2. 部署到服务器（公网发布，当前生产方式）

架构：

```
客户端 ──https──▶ Cloudflare（边缘证书 + 代理）
                    │  Origin Rule：回源端口 → 8080
                    ▼
              阿里云 ECS（Ubuntu）─ systemd: lms-server ─ :8080  ─ SQLite
```

步骤见 **[deploy/README.md](deploy/README.md)**，要点：

```bash
# 服务器上（Ubuntu 22.04）
scp -r deploy root@<你的服务器IP>:/root/
cd /root/deploy && chmod +x deploy.sh && sudo sh deploy.sh
curl http://127.0.0.1:8080/api/health
```

云侧两项配置：
1. **阿里云安全组**放行 `TCP 8080`；
2. **Cloudflare**：SSL/TLS 模式设为 **Flexible**；
   Rules → Origin Rules → `Hostname equals lms.hyx0211.xyz` → **Destination Port 重写为 8080**。

> **备案说明**：ECS 位于境内地域，**域名未完成 ICP 备案时**，云厂商会拦截以该域名发起的访问（**含 8080 等非标准端口**，返回备案提示页）。备案通过后自动恢复；备案期间可直接用 `http://<服务器IP>:8080` 访问。

### 3. 运行客户端

- **Windows**：安装 `图书管理系统_1.0.0_x64-setup.exe`，首次启动填写 `https://lms.hyx0211.xyz`。
- **Android**：源码与工程已就绪，当前**未纳入交付**（本机缺 Android SDK/NDK）。

客户端不硬编码服务器地址，首次连接后保存在本地，可在「服务器设置」中修改。

### 备用：本机 + 隧道方式

不部署服务器时，也可以在本机起后端并用 Cloudflare 隧道临时暴露：

```bat
tools\start-lms-named.cmd        # 命名隧道（需先配置 tools/cloudflared-named.yml）
tools\start-public-server.cmd    # 临时隧道（地址每次变化，仅测试）
```

## 默认账号

首次初始化数据库时会创建管理员 `admin` 与示例读者 `reader`（初始口令见
`backend/internal/store/seed.go`，**仅在新建空库时生效**）。

> ⚠️ **部署到公网前必须修改初始口令**；JWT 密钥请写入 `tools/lms-secret.txt`（已 gitignore）。
> 未配置 `LMS_SECRET` 时后端会生成一次性随机密钥，重启后已签发令牌失效。

## 运维（服务器上）

```bash
systemctl status lms-server      # 状态（已 enabled，开机自启）
systemctl restart lms-server     # 重启
journalctl -u lms-server -f      # 实时日志
cp /opt/lms/library_data.db /opt/lms/backup-$(date +%F).db   # 备份（建议先 stop）
```

## 功能

- 管理员：控制台统计（638 种 / 1923 册）、图书管理（分类/地区/位置/册数）、读者管理、借阅办理、操作日志、数据库备份/恢复。
- 读者：馆藏浏览与筛选、自助借阅、一键续借、我的借阅、个人信息。

## 数据迁移

新机器部署后，在管理端「备份与恢复」中上传从旧机导出的 `.db` 文件即可，所有图书、读者、借阅记录完全一致。

## 注意

- 通过 Cloudflare 公网访问时，边缘安全策略可能拦截非浏览器 User-Agent（返回 `1010`）；
  桌面客户端走 WebView2 不受影响，脚本化调用请带上浏览器 UA。
- 后端默认监听 `0.0.0.0:8765`（服务器部署时用 8080），仅建议在内网或 Cloudflare 代理之后使用，
  不要直接把数据库端口裸奔到公网。
