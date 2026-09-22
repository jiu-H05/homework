# 图书管理系统（Library Management System, LMS）

一个现代前后端分离的图书管理系统：本机运行 Go 后端 API，Tauri 瘦客户端（Windows / Android）连接，通过 Cloudflare Tunnel 发布到公网，支持多用户同时使用。

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go (net/http) + SQLite (WAL) + JWT |
| 前端 | 原生 HTML5 / ES Modules / CSS，Apple Glass 风格 |
| 客户端 | Tauri 2（Windows NSIS、Android arm64） |
| 公网发布 | Cloudflare Tunnel（双隧道：纯后端 API + 产品发布页） |
| 数据 | 单文件 SQLite，应用内一键导出/导入 |

## 目录结构

```
lms-modern/
├── server/              # Go 后端源码（纯 REST API）
├── frontend/src/        # 前端源码（HTML/JS/CSS）
│   ├── js/              # api / net / views / components
│   └── src-tauri/       # Tauri Rust 壳
├── standalone/          # 运行时数据库 library_data.db（已 gitignore）
└── tools/               # 部署工具
    ├── lms-server.exe   # 编译后的后端（gitignore）
    ├── cloudflared.exe  # 隧道工具（gitignore）
    ├── landing/         # 产品发布页静态站
    └── start-public-server.cmd  # 一键公网启动脚本
```

## 运行方法

### 1. 启动后端

```bash
# 方式一：直接运行编译好的二进制
cd tools
lms-server.exe
# 默认监听 0.0.0.0:8765

# 方式二：从源码编译运行
cd server
go run .
```

环境变量：

| 变量 | 默认 | 说明 |
|---|---|---|
| LMS_ADDR | 0.0.0.0:8765 | 监听地址 |
| LMS_DB | ../standalone/library_data.db | 数据库路径 |
| LMS_SECRET | lms-local-secret-2026 | JWT 密钥 |

验证：浏览器访问 `http://localhost:8765/api/health`，返回 `{"status":"ok"}` 即正常。

### 2. 启动公网隧道（可选，用于跨网/多端访问）

```bash
cd tools
cloudflared.exe tunnel --url http://localhost:8765
# 终端会打印一个 https://xxxx.trycloudflare.com 地址
```

产品发布页单独起一条隧道：

```bash
python -m http.server 8888 --directory tools/landing
cloudflared.exe tunnel --url http://localhost:8888
```

### 3. 运行客户端

- **Windows**：安装 `图书管理系统_1.0.0_x64-setup.exe`，首次启动填写后端地址（如上面的 trycloudflare 地址）。
- **Android**：安装 arm64 APK，同样填写后端地址。

客户端不硬编码服务器地址，首次连接后保存在本地，可在「服务器设置」中修改。

## 默认账号

| 角色 | 账号 | 密码 |
|---|---|---|
| 管理员 | admin | admin123 |
| 读者 | reader | reader123 |

## 功能

- 管理员：控制台统计（638 种 / 1923 册）、图书管理（分类/地区/位置/册数）、读者管理、借阅办理、操作日志、数据库备份/恢复。
- 读者：馆藏浏览与筛选、自助借阅、一键续借、我的借阅、个人信息。

## 数据迁移

新电脑安装客户端后，在管理端「备份与恢复」中上传从旧机导出的 `.db` 文件即可，所有图书、读者、借阅记录完全一致。
