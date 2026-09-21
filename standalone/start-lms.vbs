' start-lms.vbs - double-click launcher: start the server hidden (no console window),
' then open the default browser at the current private IP. Safe to run repeatedly.
Option Explicit
Dim fso, dir, f, exe, sh, env, ip
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
' Hidden window, do not wait. If the server is already running this extra copy exits at once.
sh.Run """" & exe & """", 0, False
WScript.Sleep 1800
ip = PrimaryIP()
If ip <> "" Then sh.Run "rundll32 url.dll,FileProtocolHandler http://" & ip & ":8765/"

Function PrimaryIP()
  Dim wmi, nics, nic, cfgs, c, a
  Set wmi = GetObject("winmgmts:\\.\root\cimv2")
  Set nics = wmi.ExecQuery("Select Index, Name from Win32_NetworkAdapter where IPEnabled=True")
  PrimaryIP = ""
  For Each nic In nics
    If IsRealNic(nic.Name) Then
      Set cfgs = wmi.ExecQuery("Select IPAddress from Win32_NetworkAdapterConfiguration where Index=" & nic.Index)
      For Each c In cfgs
        If Not IsNull(c.IPAddress) Then
          For Each a In c.IPAddress
            If InStr(a, ":") = 0 Then
              If Left(a, 7) <> "169.254" And Left(a, 4) <> "127." Then
                If PrimaryIP = "" Then PrimaryIP = a
              End If
            End If
          Next
        End If
      Next
    End If
  Next
End Function

Function IsRealNic(nm)
  Dim n
  n = LCase(nm)
  IsRealNic = True
  If InStr(n, "virtualbox") > 0 Or InStr(n, "vmware") > 0 Or InStr(n, "hyper-v") > 0 Then IsRealNic = False
  If InStr(n, "loopback") > 0 Or InStr(n, "pseudo") > 0 Or InStr(n, "docker") > 0 Then IsRealNic = False
End Function
