@echo off
rem 无窗口、以 SYSTEM 方式常驻图书管理系统；不自动弹浏览器（浏览器由用户自行打开）。
set LMS_OPEN_BROWSER=0
cd /d "%~dp0"
for %%E in ("%~dp0"*.exe) do "%%~fE"
