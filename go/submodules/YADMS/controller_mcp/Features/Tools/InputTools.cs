using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.IO;
using System.Runtime.InteropServices;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using System.Windows.Forms;
using System.Text;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class InputTools
    {
        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool SetForegroundWindow(IntPtr hWnd);

        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);

        [DllImport("user32.dll")]
        private static extern bool IsIconic(IntPtr hWnd);

        [DllImport("user32.dll", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool SetCursorPos(int X, int Y);

        [DllImport("user32.dll")]
        private static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint cButtons, uint dwExtraInfo);

        private const uint MOUSEEVENTF_LEFTDOWN = 0x0002;
        private const uint MOUSEEVENTF_LEFTUP = 0x0004;
        private const uint MOUSEEVENTF_RIGHTDOWN = 0x0008;
        private const uint MOUSEEVENTF_RIGHTUP = 0x0010;

        private static void PerformClick(string button)
        {
            if (button.Equals("right", StringComparison.OrdinalIgnoreCase))
            {
                mouse_event(MOUSEEVENTF_RIGHTDOWN, 0, 0, 0, 0);
                mouse_event(MOUSEEVENTF_RIGHTUP, 0, 0, 0, 0);
            }
            else
            {
                mouse_event(MOUSEEVENTF_LEFTDOWN, 0, 0, 0, 0);
                mouse_event(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0);
            }
        }

        private static void MoveCursor(int x, int y, int monitor_index)
        {
            var screens = Screen.AllScreens;
            if (monitor_index >= 0 && monitor_index < screens.Length)
            {
                var bounds = screens[monitor_index].Bounds;
                SetCursorPos(bounds.Left + x, bounds.Top + y);
            }
            else
            {
                SetCursorPos(x, y);
            }
        }

        [McpServerTool, Description("Moves the mouse and/or performs clicks at a specific coordinate.")]
        public static CallToolResult MouseAction(
            string action = "left_click", 
            int? x = null, 
            int? y = null, 
            int monitor_index = 0)
        {
            if (x.HasValue && y.HasValue)
            {
                MoveCursor(x.Value, y.Value, monitor_index);
                Thread.Sleep(50); // Small delay to let OS catch up
            }

            if (action.Equals("left_click", StringComparison.OrdinalIgnoreCase))
            {
                PerformClick("left");
            }
            else if (action.Equals("right_click", StringComparison.OrdinalIgnoreCase))
            {
                PerformClick("right");
            }
            else if (action.Equals("double_click", StringComparison.OrdinalIgnoreCase))
            {
                PerformClick("left");
                Thread.Sleep(50);
                PerformClick("left");
            }
            
            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = $"Executed mouse action: {action}" } }
            };
        }

        [McpServerTool, Description("Types a sequence of characters or keys using SendKeys syntax (e.g. '^{c}' for Ctrl+C).")]
        public static CallToolResult KeyboardType(string text, int delay_ms = 50)
        {
            if (string.IsNullOrEmpty(text))
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Text cannot be empty." } } };
            }

            // SendKeys.SendWait handles typing synchronously. 
            // We cannot easily inject a precise delay_ms between characters without parsing SendKeys syntax manually.
            // But we can just use SendWait which is highly robust.
            try
            {
                SendKeys.SendWait(text);
                Thread.Sleep(delay_ms); // Optional post-typing delay
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"SendKeys failed: {ex.Message}" } } };
            }

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully typed: {text}" } }
            };
        }

        [McpServerTool, Description("Executes a JSON-defined macro of sequential mouse/keyboard actions with precise delays.")]
        public static CallToolResult ExecuteMacro(JsonElement actions)
        {
            if (actions.ValueKind != JsonValueKind.Array)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "actions parameter must be a JSON array." } } };
            }

            foreach (var action in actions.EnumerateArray())
            {
                string type = action.GetProperty("type").GetString();

                switch (type.ToLower())
                {
                    case "move":
                        int x = action.GetProperty("x").GetInt32();
                        int y = action.GetProperty("y").GetInt32();
                        int monitor = 0;
                        if (action.TryGetProperty("monitor", out var mProp))
                            monitor = mProp.GetInt32();
                        MoveCursor(x, y, monitor);
                        break;
                    case "click":
                        string button = "left";
                        if (action.TryGetProperty("button", out var btnProp))
                            button = btnProp.GetString();
                        PerformClick(button);
                        break;
                    case "double_click":
                        PerformClick("left");
                        Thread.Sleep(50);
                        PerformClick("left");
                        break;
                    case "delay":
                        int ms = action.GetProperty("ms").GetInt32();
                        Thread.Sleep(ms);
                        break;
                    case "type":
                    case "press":
                        string text = "";
                        if (action.TryGetProperty("text", out var textProp))
                            text = textProp.GetString();
                        else if (action.TryGetProperty("key", out var keyProp))
                            text = keyProp.GetString();
                        
                        if (!string.IsNullOrEmpty(text))
                        {
                            SendKeys.SendWait(text);
                        }
                        
                        if (action.TryGetProperty("delay_ms", out var delayProp))
                            Thread.Sleep(delayProp.GetInt32());
                        break;
                    default:
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Unknown macro action type: {type}" } } };
                }
            }

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = "Macro execution complete." } }
            };
        }

        private static readonly Dictionary<char, string> QwertyMap = new Dictionary<char, string>
        {
            // Numbers & Top symbols
            {'1', "2q"}, {'2', "13qw"}, {'3', "24we"}, {'4', "35er"}, {'5', "46rt"},
            {'6', "57ty"}, {'7', "68yu"}, {'8', "79ui"}, {'9', "80io"}, {'0', "9-op"},
            {'-', "0=p["}, {'=', "-]"},
            
            // Top row
            {'q', "12wa"}, {'w', "23qeas"}, {'e', "34wrsd"}, {'r', "45etdf"}, {'t', "56ryfg"},
            {'y', "67tugh"}, {'u', "78yihj"}, {'i', "89uojk"}, {'o', "90ipkl"}, {'p', "0-o["},
            {'[', "-=p]"}, {']', "=["},

            // Home row
            {'a', "qwsz"}, {'s', "weadzx"}, {'d', "ersfxc"}, {'f', "rtdgcv"}, {'g', "tyfhvb"},
            {'h', "yugjbn"}, {'j', "uihknm"}, {'k', "iojlm,"}, {'l', "opk;.,"}, {';', "p[l'/"},
            {'\'', "];"},

            // Bottom row
            {'z', "asx"}, {'x', "sdzc"}, {'c', "dfxv"}, {'v', "fgcb"}, {'b', "ghvn"},
            {'n', "hjbm"}, {'m', "jkn,"}, {',', "klm."}, {'.', "l;,/"}, {'/', ";.,"}
        };

        private static string EscapeSendKeys(char c)
        {
            if (c == '\n') return "{ENTER}";
            if (c == '\r') return ""; 
            if (c == ' ') return " ";
            if (c == '\t') return "{TAB}";

            string special = "+^%~(){}[]";
            if (special.Contains(c.ToString())) return "{" + c + "}";
            return c.ToString();
        }

        [McpServerTool, Description("Types the contents of an ASCII text file into a target window, simulating physical human typing with realistic randomized delays and adjacent-key typos/backspaces.")]
        public static async Task<CallToolResult> ShadowTypeFile(
            string file_path, 
            string window_search_term = null, 
            int typo_chance_percent = 5, 
            int base_delay_ms = 50,
            IProgress<ProgressNotificationValue> progress = null)
        {
            string safePath = InputValidator.ValidateFilePath(file_path, nameof(file_path));

            if (!System.IO.File.Exists(safePath))
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File not found: {safePath}" } } };
            }

            if (!string.IsNullOrEmpty(window_search_term))
            {
                IntPtr hwnd = IntPtr.Zero;
                ScreenshotTools.EnumDelegate filter = delegate (IntPtr hWnd, int lParam)
                {
                    if (!ScreenshotTools.IsWindowVisible(hWnd)) return true;
                    StringBuilder sb = new StringBuilder(255);
                    ScreenshotTools.GetWindowText(hWnd, sb, sb.Capacity + 1);
                    string title = sb.ToString();

                    ScreenshotTools.GetWindowThreadProcessId(hWnd, out uint pid);
                    string processName = "";
                    try
                    {
                        using (var process = System.Diagnostics.Process.GetProcessById((int)pid))
                        {
                            processName = process.ProcessName;
                        }
                    }
                    catch { }

                    if (title.IndexOf(window_search_term, StringComparison.OrdinalIgnoreCase) >= 0 ||
                        processName.IndexOf(window_search_term, StringComparison.OrdinalIgnoreCase) >= 0)
                    {
                        hwnd = hWnd;
                        return false;
                    }
                    return true;
                };

                ScreenshotTools.EnumWindows(filter, 0);

                if (hwnd == IntPtr.Zero)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find an open window matching '{window_search_term}'." } } };
                }

                if (IsIconic(hwnd))
                {
                    ShowWindow(hwnd, 9); // SW_RESTORE = 9
                    await Task.Delay(300);
                }
                SetForegroundWindow(hwnd);
                await Task.Delay(500); // Wait for window to focus
            }

            string text = System.IO.File.ReadAllText(safePath);
            
            // Run typing on a background thread so we don't stall the MCP server
            await Task.Run(() => 
            {
                Random rng = new Random();
                int totalChars = text.Length;
                
                for (int i = 0; i < totalChars; i++)
                {
                    char c = text[i];
                    if (c == '\r') continue; // Skip raw CR, SendKeys {ENTER} handles newlines.

                    if (progress != null && i % Math.Max(1, totalChars / 20) == 0)
                    {
                        progress.Report(new ProgressNotificationValue { Progress = (float)((double)i / totalChars * 100.0), Message = $"Typing... {i}/{totalChars}" });
                    }

                    char lowerC = char.ToLower(c);
                    bool isUpper = char.IsUpper(c);

                    if (QwertyMap.ContainsKey(lowerC) && rng.Next(0, 100) < typo_chance_percent)
                    {
                        // Simulate a typo
                        string adj = QwertyMap[lowerC];
                        char typo = adj[rng.Next(adj.Length)];
                        if (isUpper) typo = char.ToUpper(typo);

                        SendKeys.SendWait(EscapeSendKeys(typo));
                        Thread.Sleep(base_delay_ms + rng.Next(-10, 30)); // Reaction delay
                        
                        SendKeys.SendWait("{BACKSPACE}");
                        Thread.Sleep(base_delay_ms + rng.Next(-10, 20)); // Correction delay
                    }

                    // Type correct char
                    string escaped = EscapeSendKeys(c);
                    if (!string.IsNullOrEmpty(escaped))
                    {
                        SendKeys.SendWait(escaped);
                    }

                    // Randomized human delay
                    Thread.Sleep(base_delay_ms + rng.Next((int)(-base_delay_ms * 0.3), (int)(base_delay_ms * 0.8)));
                }
            });

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully shadow-typed {file_path}." } }
            };
        }
    }
}
