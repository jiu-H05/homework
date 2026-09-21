@echo off
call "C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Auxiliary\Build\vcvars64.bat" >nul
set BIN=C:\Users\Lenovo\.rustup\toolchains\1.88.0-x86_64-pc-windows-msvc\bin
set FIX=C:\Users\Lenovo\Desktop\PJ\lms-modern\tools\dllexp\fixed.manifest
mt -nologo -manifest %FIX% -outputresource:%BIN%\rustc.exe;#1
mt -nologo -manifest %FIX% -outputresource:%BIN%\cargo.exe;#1
echo TOOLCHAIN_FIXED
