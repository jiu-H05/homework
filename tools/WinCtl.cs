using System;
using System.Collections.Generic;
using System.Text;
using System.Runtime.InteropServices;

public static class WinCtl
{
    private delegate bool EnumProc(IntPtr h, IntPtr l);

    [DllImport("user32.dll")] private static extern bool EnumWindows(EnumProc cb, IntPtr l);
    [DllImport("user32.dll")] private static extern int GetWindowText(IntPtr h, StringBuilder s, int n);
    [DllImport("user32.dll")] private static extern bool IsWindowVisible(IntPtr h);
    [DllImport("user32.dll")] private static extern bool ShowWindow(IntPtr h, int n);
    [DllImport("user32.dll")] private static extern bool MoveWindow(IntPtr h, int x, int y, int w, int ht, bool r);
    [DllImport("user32.dll")] private static extern bool SetForegroundWindow(IntPtr h);
    [DllImport("user32.dll")] private static extern bool SetCursorPos(int x, int y);
    [DllImport("user32.dll")] private static extern void mouse_event(uint flags, uint dx, uint dy, uint data, UIntPtr extra);
    private const uint MOVE = 0x0001, LEFTDOWN = 0x0002, LEFTUP = 0x0004;

    public static void Click(int x, int y)
    {
        SetCursorPos(x, y);
        System.Threading.Thread.Sleep(120);
        mouse_event(LEFTDOWN, 0, 0, 0, UIntPtr.Zero);
        System.Threading.Thread.Sleep(60);
        mouse_event(LEFTUP, 0, 0, 0, UIntPtr.Zero);
        System.Threading.Thread.Sleep(250);
    }

    public static void Send(string keys)
    {
        System.Windows.Forms.SendKeys.SendWait(keys);
        System.Threading.Thread.Sleep(150);
    }

    // 找到标题完全等于 title 的窗口，最小化其他有标题窗口，恢复并前置该窗口。
    public static string FocusOnly(string title, int x, int y, int w, int ht)
    {
        IntPtr app = IntPtr.Zero;
        var others = new List<IntPtr>();
        EnumWindows((h, l) =>
        {
            if (IsWindowVisible(h))
            {
                var sb = new StringBuilder(256);
                GetWindowText(h, sb, 256);
                string t = sb.ToString();
                if (t.Length > 0)
                {
                    if (t == title) app = h;
                    else others.Add(h);
                }
            }
            return true;
        }, IntPtr.Zero);

        foreach (var o in others) ShowWindow(o, 6);
        if (app == IntPtr.Zero) return "NOT FOUND";
        ShowWindow(app, 9);
        MoveWindow(app, x, y, w, ht, true);
        SetForegroundWindow(app);
        return "OK " + app;
    }

    public static void Shot(string outPath)
    {
        var b = System.Windows.Forms.Screen.PrimaryScreen.Bounds;
        using (var bmp = new System.Drawing.Bitmap(b.Width, b.Height))
        using (var g = System.Drawing.Graphics.FromImage(bmp))
        {
            g.CopyFromScreen(b.Location, System.Drawing.Point.Empty, b.Size);
            bmp.Save(outPath);
        }
    }
}
