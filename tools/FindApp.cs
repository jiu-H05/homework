using System;
using System.Text;
using System.Runtime.InteropServices;

public static class FindApp
{
    [DllImport("user32.dll", CharSet = CharSet.Unicode)]
    private static extern IntPtr FindWindow(string cls, string title);

    public static IntPtr ByTitle(string title)
    {
        return FindWindow(null, title);
    }
}
