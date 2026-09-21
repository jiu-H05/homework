using System;
using System.Text;
using System.Runtime.InteropServices;

public static class EnumChild
{
    private delegate bool EnumProc(IntPtr h, IntPtr l);

    [DllImport("user32.dll")] private static extern bool EnumChildWindows(IntPtr parent, EnumProc cb, IntPtr l);
    [DllImport("user32.dll")] private static extern int GetClassName(IntPtr h, StringBuilder s, int n);
    [DllImport("user32.dll")] private static extern bool GetWindowRect(IntPtr h, out RECT r);
    [StructLayout(LayoutKind.Sequential)] private struct RECT { public int Left, Top, Right, Bottom; }

    public static string Dump(IntPtr parent)
    {
        var sb = new StringBuilder();
        EnumChildWindows(parent, (h, l) =>
        {
            var cn = new StringBuilder(256);
            GetClassName(h, cn, 256);
            RECT r; GetWindowRect(h, out r);
            sb.Append(h).Append(" class='").Append(cn).Append("' rect=(")
              .Append(r.Left).Append(",").Append(r.Top).Append(",").Append(r.Right).Append(",").Append(r.Bottom).Append(")\n");
            return true;
        }, IntPtr.Zero);
        return sb.ToString();
    }
}
