using System;
using System.Text;
using System.Management;
using System.ComponentModel;
using ModelContextProtocol.Server;
using System.Reflection;

namespace controller_mcp.Features.Prompts
{
    [McpServerPromptType]
    public static class SystemInfoPrompt
    {
        [McpServerPrompt, Description("Generates a comprehensive summary of the system hardware, OS, and MCP Controller details.")]
        public static string GetSystemInfo()
        {
            var sb = new StringBuilder();
            sb.AppendLine("## System Information");
            sb.AppendLine($"**OS Version:** {Environment.OSVersion}");
            sb.AppendLine($"**Machine Name:** {Environment.MachineName}");
            sb.AppendLine($"**.NET Runtime:** {Environment.Version}");
            
            sb.AppendLine("\n## Hardware Data");
            try
            {
                using (var searcher = new ManagementObjectSearcher("SELECT Name, NumberOfCores FROM Win32_Processor"))
                {
                    foreach (var item in searcher.Get())
                    {
                        sb.AppendLine($"**CPU:** {item["Name"]} ({item["NumberOfCores"]} Cores)");
                    }
                }
            }
            catch (Exception ex) { sb.AppendLine($"**CPU:** [Error retrieving CPU: {ex.Message}]"); }

            try
            {
                using (var searcher = new ManagementObjectSearcher("SELECT TotalPhysicalMemory FROM Win32_ComputerSystem"))
                {
                    foreach (var item in searcher.Get())
                    {
                        var bytes = Convert.ToDouble(item["TotalPhysicalMemory"]);
                        var gb = Math.Round(bytes / 1073741824, 2);
                        sb.AppendLine($"**RAM:** {gb} GB");
                    }
                }
            }
            catch (Exception ex) { sb.AppendLine($"**RAM:** [Error retrieving RAM: {ex.Message}]"); }

            try
            {
                using (var searcher = new ManagementObjectSearcher("SELECT Manufacturer, Product FROM Win32_BaseBoard"))
                {
                    foreach (var item in searcher.Get())
                    {
                        sb.AppendLine($"**Motherboard:** {item["Manufacturer"]} - {item["Product"]}");
                    }
                }
            }
            catch (Exception ex) { sb.AppendLine($"**Motherboard:** [Error retrieving Motherboard: {ex.Message}]"); }

            sb.AppendLine("\n## MCP Controller Server Information");
            sb.AppendLine($"**Controller Location:** {AppDomain.CurrentDomain.BaseDirectory}");
            sb.AppendLine($"**Executable Name:** {AppDomain.CurrentDomain.FriendlyName}");
            sb.AppendLine($"**Protocol:** Model Context Protocol (HTTP SSE)");
            sb.AppendLine($"**SDK Version:** v1.2.0 (Official ModelContextProtocol C# SDK)");
            sb.AppendLine("This server provides deep system integration, including full screenshotting capabilities (multi-monitor, region, and application-specific captures).");

            return sb.ToString();
        }
    }
}
