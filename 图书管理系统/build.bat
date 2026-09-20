@echo off
chcp 65001 >nul
REM ============================================================
REM 图书管理系统 - Windows 打包脚本（生成单文件 .exe）
REM 需先安装 pyinstaller：python -m pip install pyinstaller
REM 若存在 app.ico 则使用应用图标，否则使用默认图标。
REM ============================================================
echo [1/2] 开始打包图书管理系统 ...
if exist app.ico (
    python -m PyInstaller --noconfirm --clean --onefile --windowed --name LibrarySystem --icon app.ico main.py
) else (
    python -m PyInstaller --noconfirm --clean --onefile --windowed --name LibrarySystem main.py
)
if errorlevel 1 (
    echo 打包失败，请检查错误信息。
    pause
    exit /b 1
)
echo [2/2] 打包完成，可执行文件位于 dist\LibrarySystem.exe
pause
