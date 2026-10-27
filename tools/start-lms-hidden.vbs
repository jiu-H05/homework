Dim fso, sh, base, cmdPath
Set fso = CreateObject("Scripting.FileSystemObject")
Set sh = CreateObject("WScript.Shell")
base = fso.GetParentFolderName(WScript.ScriptFullName)
cmdPath = base & "\start-lms-named.cmd"
If Not fso.FileExists(cmdPath) Then WScript.Quit 1
sh.Run "cmd /c """ & cmdPath & """", 0, False