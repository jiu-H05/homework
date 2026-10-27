@echo off
setlocal
cd /d "%~dp0"

rem LMS backend + Cloudflare named tunnel (fixed domain https://lms.hyx0211.xyz)
set LMS_ADDR=0.0.0.0:8765
set LMS_DB=%~dp0..\standalone\library_data.db
rem JWT secret is read from lms-secret.txt (gitignored, not committed)
if exist "%~dp0lms-secret.txt" ( set /p LMS_SECRET=<"%~dp0lms-secret.txt" )
set LMS_OPEN_BROWSER=0

echo [1/2] starting backend on 0.0.0.0:8765 ...
start "" /b "%~dp0lms-server.exe"
timeout /t 4 /nobreak >nul

echo [2/2] starting named tunnel (lms -> lms.hyx0211.xyz) ...
"%~dp0cloudflared.exe" --config "%~dp0cloudflared-named.yml" tunnel run lms