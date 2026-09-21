' open-lms.vbs - detect the real private IPv4 (exclude virtual NICs) and open the LMS page.
Option Explicit
Dim sh, ip
Set sh = CreateObject("WScript.Shell")
ip = PrimaryIP()
If ip = "" Then WScript.Quit 1
sh.Run "rundll32 url.dll,FileProtocolHandler http://" & ip & ":8765/"

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
