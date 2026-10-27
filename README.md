# 图书管理系统（Library Management System, LMS）

一个现代前后端分离的图书管理系统：本机运行 Go 后端 API，Tauri 瘦客户端（Windows）连接，通过 Cloudflare **Named Tunnel** 以固定域名 **https://lms.hyx0211.xyz** 发布到公网，支持多用户同时使用。

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go (net/http) + SQLite (WAL) + JWT |
| 前端 | 原生 HTML5 / ES Modules / CSS，Apple Glass 风格 |
| 客户端 | Tauri 2（Windows NSIS、Android arm64） |
| 公网发布 | Cloudflare **Named Tunnel**（固定域名 https://lms.hyx0211.xyz）；产品发布页可选临时隧道 |
| 数据 | 单文件 SQLite，应用内一键导出/导入 |

## 目录结构

```
lms-modern/
├── backend/             # Go 后端源码（纯 REST API，入口 cmd/server）
├── frontend/src/        # 前端源码（HTML/JS/CSS）
│   ├── js/              # api / net / views / components
│   └── src-tauri/       # Tauri Rust 壳
├── standalone/          # 运行时数据库 library_data.db（已 gitignore）
└── tools/               # 部署工具
    ├── lms-server.exe        # 编译后的后端（gitignore）
    ├── cloudflared.exe       # 隧道工具（gitignore）
    ├── cloudflared-named.yml # 命名隧道配置（固定域名，不含密钥）
    ├── start-lms-named.cmd   # 后端 + 命名隧道 一键启动（推荐）
    ├── start-lms-hidden.vbs  # 隐藏窗口启动器（登录自启调用）
    ├── lms-secret.txt        # JWT 密钥（gitignore，勿提交）
    ├── landing/              # 产品发布页静态站
    └── start-public-server.cmd # 临时隧道版一键启动（备用）
```

## 运行方法

### 1. 启动后端

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
| LMS_ADDR | 0.0.0.0:8765 | 监听地址 |
| LMS_DB | ../standalone/library_data.db | 数据库路径 |
| LMS_SECRET | 见 `tools\lms-secret.txt`（已 gitignore，勿提交） | JWT 密钥，生产必须随机 |

验证：浏览器访问 `http://localhost:8765/api/health`，返回 `{"status":"ok"}` 即正常。

### 2. 启动公网隧道（命名隧道，固定域名）

固定公网地址：**https://lms.hyx0211.xyz**（重启不变）。

首次配置（每个域名只需一次）：

```bash
cd tools
cloudflared.exe tunnel login                      # 浏览器里授权该域名
cloudflared.exe tunnel create lms                 # 创建隧道，记下 tunnel UUID
cloudflared.exe tunnel route dns lms lms.hyx0211.xyz   # 建 DNS CNAME（必须，否则域名不解析）
```

然后把 UUID 与 `credentials-file` 路径填进 `tools/cloudflared-named.yml`。

日常启动：

```bat
tools\start-lms-named.cmd
```

开机/登录自启：Windows 计划任务 `LMS-NamedTunnel` 调 `wscript tools\start-lms-hidden.vbs`（隐藏窗口，无需管理员）。

> 临时隧道（地址每次变化，仅用于临时测试）：
> `cloudflared.exe tunnel --url http://localhost:8765`

产品发布页（可选）单独起一条临时隧道：

```bash
python -m http.server 8888 --directory tools/landing
cloudflared.exe tunnel --url http://localhost:8888
```

### 3. 运行客户端

- **Windows**：安装 `图书管理系统_1.0.0_x64-setup.exe`，首次启动填写固定地址 `https://lms.hyx0211.xyz`。
- **Android**：源码与工程已就绪，当前**未纳入交付**（本机缺 Android SDK/NDK）。

客户端不硬编码服务器地址，首次连接后保存在本地，可在「服务器设置」中修改。

## 默认账号

| 角色 | 账号 | 密码 |
|---|---|---|
| 管理员 | admin | admin123 |
| 读者 | reader | reader123 |

> ⚠️ **部署到公网后请立即修改默认口令**，并将 JWT 密钥写入 `tools/lms-secret.txt`（已 gitignore）。
> 密钥轮换后所有已签发令牌立即失效，需要重新登录。

## 注意

- 通过 Cloudflare 公网访问时，边缘的安全策略可能拦截非浏览器 User-Agent（返回 `1010`）；
  桌面客户端走 WebView2 不受影响，脚本化调用请带上浏览器 UA。
- 后端默认监听 `0.0.0.0:8765`，仅建议在内网或隧道后使用，不要直接暴露 8765 到公网。

## 功能

- 管理员：控制台统计（638 种 / 1923 册）、图书管理（分类/地区/位置/册数）、读者管理、借阅办理、操作日志、数据库备份/恢复。
- 读者：馆藏浏览与筛选、自助借阅、一键续借、我的借阅、个人信息。

## 数据迁移

新电脑安装客户端后，在管理端「备份与恢复」中上传从旧机导出的 `.db` 文件即可，所有图书、读者、借阅记录完全一致。
