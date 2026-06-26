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
    public static class WndProcTools
    {
        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool IsWindowVisible(IntPtr hWnd);

        [DllImport("user32.dll", SetLastError = true)]
        private static extern bool PostMessage(IntPtr hWnd, uint Msg, IntPtr wParam, IntPtr lParam);

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

        [McpServerTool, Description("Injects a raw hardware window message (like WM_KEYDOWN) directly into a background application's message loop without bringing it to the foreground.")]
        public static async Task<CallToolResult> SendRawWindowMessage(string target_name, uint msg, int wparam, int lparam)
        {
            return await Task.Run(() =>
            {
                try
                {
                    IntPtr hwnd = FindWindowByName(target_name);
                    if (hwnd == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find window matching '{target_name}'." } } };

                    bool success = PostMessage(hwnd, msg, (IntPtr)wparam, (IntPtr)lparam);

                    if (success)
                    {
                        return new CallToolResult
                        {
                            Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully posted message {msg} (0x{msg:X}) to window {hwnd:X}." } }
                        };
                    }
                    else
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to post message. Error code: {Marshal.GetLastWin32Error()}" } } };
                    }
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"WndProc Injection Failed: {ex.Message}" } } };
                }
            });
        }
    }
}
