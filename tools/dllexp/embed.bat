@echo off
call "C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Auxiliary\Build\vcvars64.bat" >nul
cd /d "C:\Users\Lenovo\Desktop\PJ\lms-modern\tools\dllexp"
copy /y myexe.exe myexe_utf8.exe >nul
mt -nologo -manifest utf8.manifest -outputresource:myexe_utf8.exe;#1
echo EMBED_DONE
