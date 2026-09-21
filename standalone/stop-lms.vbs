' stop-lms.vbs - stop the hidden LMS server by matching its install path (no console window to close).
Option Explicit
Dim wmi, procs, p
Set wmi = GetObject("winmgmts:\\.\root\cimv2")
Set procs = wmi.ExecQuery("Select ExecutablePath from Win32_Process")
For Each p In procs
  If Not IsNull(p.ExecutablePath) Then
    If InStr(LCase(p.ExecutablePath), "\lms-modern\standalone\") > 0 Then p.Terminate
  End If
Next
