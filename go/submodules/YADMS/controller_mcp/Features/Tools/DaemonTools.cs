using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Threading.Tasks;
using System.Windows.Forms;
using Microsoft.Win32;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class DaemonTools
    {
        [McpServerTool, Description("Installs the current MCP Server application as a background Daemon that runs automatically when the user logs in, hiding completely in the system tray.")]
        public static CallToolResult InstallAsService()
        {
            try
            {
                using (RegistryKey rk = Registry.CurrentUser.OpenSubKey("SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run", true))
                {
                    rk.SetValue("ControllerMcpDaemon", "\"" + Application.ExecutablePath + "\" --daemon");
                }
                
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = "Successfully registered Controller MCP to start as a background daemon on logon." } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to install daemon: {ex.Message}" } } };
            }
        }
        [McpServerTool, Description("Uninstalls the background Daemon so it no longer runs automatically on logon.")]
        public static CallToolResult RemoveAsService()
        {
            try
            {
                using (RegistryKey rk = Registry.CurrentUser.OpenSubKey("SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run", true))
                {
                    if (rk.GetValue("ControllerMcpDaemon") != null)
                        rk.DeleteValue("ControllerMcpDaemon", false);
                }
                
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = "Successfully removed Controller MCP from background startup." } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to remove daemon: {ex.Message}" } } };
            }
        }

        public static bool IsServiceInstalled()
        {
            try
            {
                using (RegistryKey rk = Registry.CurrentUser.OpenSubKey("SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run", false))
                {
                    return rk != null && rk.GetValue("ControllerMcpDaemon") != null;
                }
            }
            catch
            {
                return false;
            }
        }
    }
}
