@echo off
set LMS_ADDR=0.0.0.0:8765
set LMS_DB=C:\Users\Lenovo\AppData\Roaming\com.lms.desktop\library_data.db
rem JWT secret is read from lms-secret.txt (gitignored, not committed)
if exist "%~dp0lms-secret.txt" ( set /p LMS_SECRET=<"%~dp0lms-secret.txt" )
"C:\Users\Lenovo\Desktop\PJ\lms-modern\frontend\src-tauri\binaries\lms-server-x86_64-pc-windows-msvc.exe"
