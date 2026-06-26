using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Management;
using System.Text.Json;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class ResourceTools
    {
        [McpServerTool, Description("Pulls live CPU load, available RAM, and GPU details using WMI and Performance Counters.")]
        public static async Task<CallToolResult> GetHardwareMetrics()
        {
            return await Task.Run(() =>
            {
                var metrics = new Dictionary<string, object>();

                try
                {
                    // CPU Usage
                    using (PerformanceCounter cpuCounter = new PerformanceCounter("Processor", "% Processor Time", "_Total"))
                    {
                        cpuCounter.NextValue();
                        System.Threading.Thread.Sleep(500); // 500ms delay to get a true reading
                        metrics["CPU_Load_Percent"] = Math.Round(cpuCounter.NextValue(), 2);
                    }

                    // RAM
                    using (ManagementObjectSearcher searcher = new ManagementObjectSearcher("SELECT FreePhysicalMemory, TotalVisibleMemorySize FROM Win32_OperatingSystem"))
                    {
                        foreach (ManagementObject mo in searcher.Get())
                        {
                            long freeMemKb = Convert.ToInt64(mo["FreePhysicalMemory"]);
                            long totalMemKb = Convert.ToInt64(mo["TotalVisibleMemorySize"]);
                            metrics["RAM_Total_MB"] = totalMemKb / 1024;
                            metrics["RAM_Free_MB"] = freeMemKb / 1024;
                            metrics["RAM_Usage_Percent"] = Math.Round((1.0 - ((double)freeMemKb / totalMemKb)) * 100.0, 2);
                        }
                    }

                    // GPU
                    var gpus = new List<object>();
                    using (ManagementObjectSearcher searcher = new ManagementObjectSearcher("SELECT Name, AdapterRAM, DriverVersion FROM Win32_VideoController"))
                    {
                        foreach (ManagementObject mo in searcher.Get())
                        {
                            string name = mo["Name"]?.ToString();
                            long vramBytes = 0;
                            if (mo["AdapterRAM"] != null)
                            {
                                long.TryParse(mo["AdapterRAM"].ToString(), out vramBytes);
                            }
                            
                            gpus.Add(new
                            {
                                Name = name,
                                VRAM_MB = vramBytes / 1024 / 1024,
                                Driver = mo["DriverVersion"]?.ToString()
                            });
                        }
                    }
                    metrics["GPUs"] = gpus;
                }
                catch (Exception ex)
                {
                    metrics["Error"] = ex.Message;
                }

                string json = JsonSerializer.Serialize(metrics, new JsonSerializerOptions { WriteIndented = true });
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
                };
            });
        }
    }
}
