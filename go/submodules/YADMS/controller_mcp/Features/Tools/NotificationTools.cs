using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Threading.Tasks;
using Microsoft.Toolkit.Uwp.Notifications;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class NotificationTools
    {
        [McpServerTool, Description("Shows a native OS notification. Uses Toast Notifications on Windows 10/11, and gracefully falls back to System Tray Balloon Tips on Windows 7/8.")]
        public static async Task<CallToolResult> ShowNotification(string title, string message)
        {
            return await Task.Run(() =>
            {
                try
                {
                    // Attempt Windows 10/11 Native Toast
                    new ToastContentBuilder()
                        .AddText(title)
                        .AddText(message)
                        .Show();
                    
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = "Toast notification sent." } } };
                }
                catch
                {
                    try
                    {
                        // Fallback to Windows 7 System Tray Balloon
                        Form1.Instance?.ShowBalloonTip(title, message);
                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = "Balloon tip notification sent (Win 7 fallback)." } } };
                    }
                    catch (Exception ex)
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to send notification: {ex.Message}" } } };
                    }
                }
            });
        }
    }
}
