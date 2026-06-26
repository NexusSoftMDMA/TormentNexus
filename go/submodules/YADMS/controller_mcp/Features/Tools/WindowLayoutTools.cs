using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class WindowLayoutTools
    {
        [DllImport("user32.dll", SetLastError = true)]
        static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int X, int Y, int cx, int cy, uint uFlags);

        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool IsWindowVisible(IntPtr hWnd);

        private static IntPtr FindWindowByName(string targetName)
        {
            IntPtr hwnd = IntPtr.Zero;
            ScreenshotTools.EnumDelegate filter = delegate (IntPtr hWnd, int lParam)
            {
                if (!IsWindowVisible(hWnd)) return true;
                StringBuilder sb = new StringBuilder(255);
                ScreenshotTools.GetWindowText(hWnd, sb, sb.Capacity + 1);
                string title = sb.ToString();

                ScreenshotTools.GetWindowThreadProcessId(hWnd, out uint pid);
                string processName = "";
                try
                {
                    using (var proc = Process.GetProcessById((int)pid))
                    {
                        processName = proc.ProcessName;
                    }
                }
                catch { }

                if (title.IndexOf(targetName, StringComparison.OrdinalIgnoreCase) >= 0 ||
                    processName.IndexOf(targetName, StringComparison.OrdinalIgnoreCase) >= 0)
                {
                    hwnd = hWnd;
                    return false;
                }
                return true;
            };

            ScreenshotTools.EnumWindows(filter, 0);
            return hwnd;
        }

        [McpServerTool, Description("Programmatically moves and resizes a specific window on the physical desktop. Provide window name/process, X and Y coordinates, and width/height.")]
        public static async Task<CallToolResult> SetWindowPosition(string target_name, int x, int y, int width, int height)
        {
            return await Task.Run(() =>
            {
                try
                {
                    IntPtr hwnd = FindWindowByName(target_name);
                    if (hwnd == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find window matching '{target_name}'." } } };

                    // SWP_NOZORDER = 0x0004
                    bool success = SetWindowPos(hwnd, IntPtr.Zero, x, y, width, height, 0x0004);

                    if (success)
                    {
                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully moved '{target_name}' to {x},{y} ({width}x{height})." } } };
                    }
                    else
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to move window. Error code: {Marshal.GetLastWin32Error()}" } } };
                    }
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"SetWindowPosition Failed: {ex.Message}" } } };
                }
            });
        }
    }
}
