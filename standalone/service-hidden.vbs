' service-hidden.vbs - hidden, windowless, persistent launcher for the LMS server (current user).
Option Explicit
Dim fso, dir, f, exe, sh, env
Set fso = CreateObject("Scripting.FileSystemObject")
dir = fso.GetParentFolderName(WScript.ScriptFullName)
exe = ""
For Each f In fso.GetFolder(dir).Files
  If LCase(fso.GetExtensionName(f.Name)) = "exe" Then exe = f.Path
Next
If exe = "" Then WScript.Quit 1
Set sh = CreateObject("WScript.Shell")
sh.CurrentDirectory = dir
Set env = sh.Environment("PROCESS")
env("LMS_OPEN_BROWSER") = "0"
' style 0 = hidden window; True = wait until the server exits (keeps the task alive).
sh.Run """" & exe & """", 0, True
