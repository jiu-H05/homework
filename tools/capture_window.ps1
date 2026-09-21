# capture_window.ps1 -Title <窗口标题> -Out <PNG路径>
param(
  [Parameter(Mandatory=$true)][string]$Title,
  [Parameter(Mandatory=$true)][string]$Out
)
Add-Type -AssemblyName System.Drawing
$src = @"
using System;
using System.Runtime.InteropServices;
public class WinCap {
  [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowsProc cb, IntPtr lParam);
  public delegate bool EnumWindowsProc(IntPtr hWnd, IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hWnd);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetWindowText(IntPtr hWnd, System.Text.StringBuilder s, int n);
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT r);
  [DllImport("user32.dll")] public static extern bool PrintWindow(IntPtr hWnd, IntPtr hdcBlt, uint flags);
  [StructLayout(LayoutKind.Sequential)] public struct RECT { public int Left, Top, Right, Bottom; }
}
"@
Add-Type -TypeDefinition $src -ReferencedAssemblies System.Drawing

$found = [IntPtr]::Zero
$cb = [WinCap+EnumWindowsProc]{
  param($hWnd,$lp)
  if([WinCap]::IsWindowVisible($hWnd)){
    $sb = New-Object System.Text.StringBuilder 512
    [void][WinCap]::GetWindowText($hWnd,$sb,$sb.Capacity)
    if($sb.ToString() -eq $Title){ $script:found = $hWnd }
  }
  return $true
}
[void][WinCap]::EnumWindows($cb,[IntPtr]::Zero)

if($found -eq [IntPtr]::Zero){ Write-Output "WINDOW_NOT_FOUND"; exit 1 }
$r = New-Object WinCap+RECT
[void][WinCap]::GetWindowRect($found,[ref]$r)
$w = $r.Right - $r.Left; $h = $r.Bottom - $r.Top
$bmp = New-Object System.Drawing.Bitmap $w,$h
$g = [System.Drawing.Graphics]::FromImage($bmp)
$hdc = $g.GetHdc()
$ok = [WinCap]::PrintWindow($found,$hdc,3)
$g.ReleaseHdc($hdc)
$bmp.Save($Out,[System.Drawing.Imaging.ImageFormat]::Png)
$g.Dispose(); $bmp.Dispose()
Write-Output ("CAPTURED ok=" + $ok + " size=" + $w + "x" + $h + " -> " + $Out)
