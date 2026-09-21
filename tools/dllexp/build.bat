@echo off
call "C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Auxiliary\Build\vcvars64.bat" >nul
cd /d "C:\Users\Lenovo\Desktop\PJ\lms-modern\tools\dllexp"
cl /nologo /LD mydll.c
cl /nologo myexe.c /link mydll.lib
echo BUILD_DONE
