@echo off
call "C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Auxiliary\Build\vcvars64.bat" >nul
cd /d "C:\Users\Lenovo\Desktop\PJ\lms-modern\tools\dllexp"
copy /y cargo_copy.exe cargo_fixed.exe >nul
mt -nologo -manifest fixed.manifest -outputresource:cargo_fixed.exe;#1
echo FIX_DONE
