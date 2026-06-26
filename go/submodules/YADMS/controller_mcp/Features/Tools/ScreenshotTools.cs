using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;
using System.Windows.Forms;

namespace controller_mcp.Features.Tools
{
    [McpServerToolType]
    public static class ScreenshotTools
    {
        [McpServerTool, Description("Captures all connected screens stitched together into a single image.")]
        public static CallToolResult CaptureAllScreens(string save_directory = null, string filename = null)
        {
            if (!string.IsNullOrWhiteSpace(filename) && string.IsNullOrWhiteSpace(save_directory))
                throw new ArgumentException("save_directory must be provided if filename is set.");
            Rectangle totalSize = Rectangle.Empty;
            foreach (Screen screen in Screen.AllScreens)
                totalSize = Rectangle.Union(totalSize, screen.Bounds);

            using (var bmp = new Bitmap(totalSize.Width, totalSize.Height, PixelFormat.Format32bppArgb))
            {
                using (var g = Graphics.FromImage(bmp))
                {
                    g.CopyFromScreen(SystemInformation.VirtualScreen.Left, SystemInformation.VirtualScreen.Top, 0, 0, SystemInformation.VirtualScreen.Size);
                }
                return CreateImageResult(bmp, save_directory, filename);
            }
        }

        [McpServerTool, Description("Captures a single full screen monitor. Provide an index (default 0 for primary).")]
        public static CallToolResult CaptureFullScreen(int screenIndex = 0, string save_directory = null, string filename = null)
        {
            if (!string.IsNullOrWhiteSpace(filename) && string.IsNullOrWhiteSpace(save_directory))
                throw new ArgumentException("save_directory must be provided if filename is set.");
            var screens = Screen.AllScreens;
            if (screenIndex < 0 || screenIndex >= screens.Length)
                screenIndex = 0;

            var bounds = screens[screenIndex].Bounds;
            using (var bmp = new Bitmap(bounds.Width, bounds.Height, PixelFormat.Format32bppArgb))
            {
                using (var g = Graphics.FromImage(bmp))
                {
                    g.CopyFromScreen(bounds.Location, Point.Empty, bounds.Size, CopyPixelOperation.SourceCopy);
                }
                return CreateImageResult(bmp, save_directory, filename);
            }
        }

        [McpServerTool, Description("Captures a specific rectangular region of the screen.")]
        public static CallToolResult CaptureScreenRegion(int x, int y, int width, int height, string save_directory = null, string filename = null)
        {
            if (string.IsNullOrEmpty(save_directory) && !string.IsNullOrEmpty(filename))
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "save_directory must be provided if filename is set." } } };
            if (width <= 0 || height <= 0)
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Width and height must be positive." } } };

            using (var bmp = new Bitmap(width, height, PixelFormat.Format32bppArgb))
            {
                using (var g = Graphics.FromImage(bmp))
                {
                    g.CopyFromScreen(x, y, 0, 0, new Size(width, height), CopyPixelOperation.SourceCopy);
                }
                return CreateImageResult(bmp, save_directory, filename);
            }
        }

        [McpServerTool, Description("Captures a specific application window by matching its title or executable name (fuzzy match).")]
        public static CallToolResult CaptureWindow(string search_term, string save_directory = null, string filename = null)
        {
            if (!string.IsNullOrWhiteSpace(filename) && string.IsNullOrWhiteSpace(save_directory))
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "save_directory must be provided if filename is set." } } };
            IntPtr hwnd = IntPtr.Zero;

            // Search windows
            EnumDelegate filter = delegate (IntPtr hWnd, int lParam)
            {
                if (!IsWindowVisible(hWnd)) return true;

                StringBuilder sb = new StringBuilder(255);
                GetWindowText(hWnd, sb, sb.Capacity + 1);
                string title = sb.ToString();

                GetWindowThreadProcessId(hWnd, out uint pid);
                string processName = "";
                try
                {
                    using (var process = Process.GetProcessById((int)pid))
                    {
                        processName = process.ProcessName;
                    }
                }
                catch { }

                if (title.IndexOf(search_term, StringComparison.OrdinalIgnoreCase) >= 0 ||
                    processName.IndexOf(search_term, StringComparison.OrdinalIgnoreCase) >= 0)
                {
                    hwnd = hWnd;
                    return false; // stop searching
                }
                return true;
            };

            EnumWindows(filter, 0);

            if (hwnd == IntPtr.Zero)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find an open window matching '{search_term}'." } } };
            }
            bool wasMinimized = false;
            if (IsIconic(hwnd))
            {
                wasMinimized = true;
                ShowWindow(hwnd, SW_RESTORE);
                System.Threading.Thread.Sleep(300); // Wait for animation
            }

            // Put window in foreground
            SetForegroundWindow(hwnd);
            System.Threading.Thread.Sleep(200); // Give it a moment to render

            if (GetWindowRect(hwnd, out RECT rect))
            {
                int width = rect.Right - rect.Left;
                int height = rect.Bottom - rect.Top;
                
                if (width <= 0 || height <= 0)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Window found but it has invalid dimensions." } } };
                }
                using (var bmp = new Bitmap(width, height, PixelFormat.Format32bppArgb))
                {
                    using (var gfx = Graphics.FromImage(bmp))
                    {
                        gfx.CopyFromScreen(rect.Left, rect.Top, 0, 0, new Size(width, height), CopyPixelOperation.SourceCopy);
                    }
                    
                    if (wasMinimized)
                    {
                        ShowWindow(hwnd, SW_MINIMIZE);
                    }
                    
                    return CreateImageResult(bmp, save_directory, filename);
                }
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Failed to get window boundaries." } } };
        }

        private static CallToolResult CreateImageResult(Bitmap bmp, string save_directory = null, string filename = null)
        {
            if (!string.IsNullOrWhiteSpace(save_directory))
            {
                string safeDir = null;
                string safeName = null;
                try
                {
                    safeDir = InputValidator.ValidateFilePath(save_directory, nameof(save_directory));
                    
                    string finalName = string.IsNullOrWhiteSpace(filename) 
                        ? $"screenshot_{DateTime.Now:yyyyMMdd_HHmmss}.png" 
                        : filename;
                        
                    if (!finalName.EndsWith(".png", StringComparison.OrdinalIgnoreCase))
                        finalName += ".png";

                    safeName = InputValidator.ValidateFilePath(finalName, nameof(filename));
                }
                catch (ArgumentException ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
                }

                try
                {
                    if (!Directory.Exists(safeDir)) Directory.CreateDirectory(safeDir);
                    string path = Path.Combine(safeDir, safeName);
                    bmp.Save(path, ImageFormat.Png);
                }
                catch { } // Ignore save errors to avoid crashing the tool
            }

            using (var ms = new MemoryStream())
            {
                bmp.Save(ms, ImageFormat.Png);
                // The C# SDK's ContentBlock JSON converter treats the Data byte array as a UTF-8 string containing the Base64 data!
                // We must encode our Base64 string as UTF-8 bytes to ensure it serializes properly.
                string base64 = Convert.ToBase64String(ms.ToArray());
                byte[] dataBytes = System.Text.Encoding.UTF8.GetBytes(base64);

                return new CallToolResult
                {
                    Content = new System.Collections.Generic.List<ContentBlock>
                    {
                        new ImageContentBlock
                        {
                            Data = dataBytes,
                            MimeType = "image/png"
                        }
                    }
                };
            }
        }

        #region Win32 P/Invoke
        public delegate bool EnumDelegate(IntPtr hWnd, int lParam);

        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        public static extern bool IsWindowVisible(IntPtr hWnd);

        [DllImport("user32.dll", EntryPoint = "EnumWindows", ExactSpelling = false, CharSet = CharSet.Auto, SetLastError = true)]
        public static extern bool EnumWindows(EnumDelegate lpEnumFunc, int lParam);

        [DllImport("user32.dll", EntryPoint = "GetWindowText", ExactSpelling = false, CharSet = CharSet.Auto, SetLastError = true)]
        public static extern int GetWindowText(IntPtr hWnd, StringBuilder lpWindowText, int nMaxCount);

        [DllImport("user32.dll", SetLastError = true)]
        public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);

        [StructLayout(LayoutKind.Sequential)]
        public struct RECT
        {
            public int Left;
            public int Top;
            public int Right;
            public int Bottom;
        }

        [DllImport("user32.dll", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);

        [DllImport("user32.dll")]
        public static extern bool SetForegroundWindow(IntPtr hWnd);

        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        public static extern bool IsIconic(IntPtr hWnd);

        [DllImport("user32.dll")]
        public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
        
        public const int SW_RESTORE = 9;
        public const int SW_MINIMIZE = 6;
        #endregion
    }
}
