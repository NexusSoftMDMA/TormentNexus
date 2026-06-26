#if GAME_HACKING
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class EngineTools
    {
        [McpServerTool, Description("Scans the loaded memory modules of a target process to determine if it is running Unity (Mono/IL2CPP) or Unreal Engine. Returns the Hexadecimal Base Address of the engine runtime module, which is required for advanced Cheat Engine pointer scanning or DLL injection.")]
        public static async Task<CallToolResult> ScanGameEngine(int target_pid)
        {
            return await Task.Run(() =>
            {
                try
                {
                    using (Process targetProcess = Process.GetProcessById(target_pid))
                    {
                        var detectedModules = new List<object>();
                        string engineType = "Unknown";

                        foreach (ProcessModule module in targetProcess.Modules)
                    {
                        string modName = module.ModuleName.ToLowerInvariant();
                        
                        // Unity Mono
                        if (modName == "mono.dll" || modName == "mono-2.0-bdwgc.dll")
                        {
                            engineType = "Unity (Mono)";
                            detectedModules.Add(new { Name = module.ModuleName, BaseAddress = $"0x{module.BaseAddress.ToInt64():X}" });
                        }
                        // Unity IL2CPP
                        else if (modName == "gameassembly.dll")
                        {
                            engineType = "Unity (IL2CPP)";
                            detectedModules.Add(new { Name = module.ModuleName, BaseAddress = $"0x{module.BaseAddress.ToInt64():X}" });
                        }
                        // Unreal Engine
                        else if (modName.StartsWith("unreal") || modName.Contains("ue4") || modName.Contains("ue5"))
                        {
                            engineType = "Unreal Engine";
                            detectedModules.Add(new { Name = module.ModuleName, BaseAddress = $"0x{module.BaseAddress.ToInt64():X}" });
                        }
                    }

                    var result = new
                    {
                        pid = target_pid,
                        process_name = targetProcess.ProcessName,
                        engine_detected = engineType,
                        critical_modules = detectedModules
                    };

                    string json = System.Text.Json.JsonSerializer.Serialize(result, new System.Text.Json.JsonSerializerOptions { WriteIndented = true });
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = json } } };
                    }
                }
                catch (UnauthorizedAccessException uex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "EngineTools", uex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Access Denied: You must run the server as Administrator to read foreign process modules. {uex.Message}" } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "EngineTools", ex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Engine scan failed: {ex.Message}" } } };
                }
            });
        }
    }
}
#endif
