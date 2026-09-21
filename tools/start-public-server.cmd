@echo off
chcp 65001 >nul
rem ============================================
rem 图书系统公网服务一键启动
rem 用法：双击本脚本。窗口保持开着，公网才有效。
rem 客户端在"服务器设置"里填窗口里显示的 https://xxxx.trycloudflare.com
rem ============================================
cd /d "%~dp0"

set LMS_ADDR=0.0.0.0:8765
set LMS_DB=%~dp0..\standalone\library_data.db
set LMS_SECRET=lms-local-secret-2026
set LMS_OPEN_BROWSER=0

echo [1/2] 启动本地后端...
start "lms-server" /min "%~dp0lms-server.exe"
timeout /t 5 /nobreak >nul

echo [2/2] 建立 Cloudflare 公网隧道...
echo 下方 https://xxxx.trycloudflare.com 就是公网地址：
echo.
"%~dp0cloudflared.exe" tunnel --url http://localhost:8765
pause
