using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows.Automation;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class AccessibilityTools
    {
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

        private static object WalkTree(AutomationElement element, int depth = 0)
        {
            if (depth > 5) return null; // Prevent massive hangs on huge UI trees

            var children = new List<object>();
            var condition = TreeWalker.ControlViewWalker;
            var child = condition.GetFirstChild(element);

            while (child != null)
            {
                var childObj = WalkTree(child, depth + 1);
                if (childObj != null) children.Add(childObj);
                child = condition.GetNextSibling(child);
            }

            return new
            {
                AutomationId = element.Current.AutomationId,
                Name = element.Current.Name,
                ControlType = element.Current.ControlType.ProgrammaticName,
                IsEnabled = element.Current.IsEnabled,
                Children = children.Count > 0 ? children : null
            };
        }

        [McpServerTool, Description("Inspects the UI hierarchy of a target application (by window name or process). Returns a JSON tree of AutomationIds and Names.")]
        public static async Task<CallToolResult> InspectUiTree(string target_name)
        {
            return await Task.Run(() =>
            {
                try
                {
                    IntPtr hwnd = FindWindowByName(target_name);
                    if (hwnd == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find window matching '{target_name}'." } } };

                    var rootElement = AutomationElement.FromHandle(hwnd);
                    var tree = WalkTree(rootElement);

                    string json = JsonSerializer.Serialize(tree, new JsonSerializerOptions { WriteIndented = true });
                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
                    };
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"UI Inspection Failed: {ex.Message}" } } };
                }
            });
        }

        private static AutomationElement FindElement(AutomationElement root, string automationIdOrName)
        {
            var condition = new OrCondition(
                new PropertyCondition(AutomationElement.AutomationIdProperty, automationIdOrName),
                new PropertyCondition(AutomationElement.NameProperty, automationIdOrName)
            );
            return root.FindFirst(TreeScope.Descendants, condition);
        }

        [McpServerTool, Description("Finds a UI element by AutomationId or Name and invokes its default action (like a button click) regardless of where it is on screen.")]
        public static async Task<CallToolResult> InvokeUiElement(string target_name, string automation_id_or_name)
        {
            return await Task.Run(() =>
            {
                try
                {
                    IntPtr hwnd = FindWindowByName(target_name);
                    if (hwnd == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find window matching '{target_name}'." } } };

                    var rootElement = AutomationElement.FromHandle(hwnd);
                    var element = FindElement(rootElement, automation_id_or_name);

                    if (element == null)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find element '{automation_id_or_name}' inside the target window." } } };

                    if (element.TryGetCurrentPattern(InvokePattern.Pattern, out object invokePattern))
                    {
                        ((InvokePattern)invokePattern).Invoke();
                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully invoked '{automation_id_or_name}'." } } };
                    }
                    else if (element.TryGetCurrentPattern(TogglePattern.Pattern, out object togglePattern))
                    {
                        ((TogglePattern)togglePattern).Toggle();
                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully toggled '{automation_id_or_name}'." } } };
                    }
                    
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Element '{automation_id_or_name}' does not support Invoke or Toggle patterns." } } };
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"UI Invocation Failed: {ex.Message}" } } };
                }
            });
        }
    }
}
