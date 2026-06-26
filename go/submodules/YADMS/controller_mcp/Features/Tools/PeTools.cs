#if GAME_HACKING
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;
using PeNet;

namespace controller_mcp.Features.Tools
{
    public static class PeTools
    {
        [McpServerTool, Description("Mathematically dissects a Windows .exe or .dll file without executing it. Returns its Win32 API imports, internal architecture, and compile timestamp. Highly useful for offline malware analysis.")]
        public static async Task<CallToolResult> AnalyzeExecutable(string file_path)
        {
            return await Task.Run(() =>
            {
                try
                {
                    string safePath = InputValidator.ValidateFilePath(file_path, nameof(file_path));

                    if (!File.Exists(safePath))
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File not found: {safePath}" } } };

                    var peFile = new PeFile(safePath);

                    bool is32Bit = peFile.Is32Bit;
                    bool is64Bit = peFile.Is64Bit;
                    bool isDll = peFile.IsDll;
                    bool isExe = peFile.IsExe;

                    // Extract imported DLLs and their functions
                    var imports = new Dictionary<string, List<string>>();
                    if (peFile.ImportedFunctions != null)
                    {
                        foreach (var import in peFile.ImportedFunctions)
                        {
                            if (!string.IsNullOrEmpty(import.DLL))
                            {
                                if (!imports.ContainsKey(import.DLL))
                                    imports[import.DLL] = new List<string>();

                                if (!string.IsNullOrEmpty(import.Name))
                                    imports[import.DLL].Add(import.Name);
                            }
                        }
                    }

                    // Look for suspicious APIs commonly used by malware
                    var suspiciousApis = new List<string> { 
                        "CreateRemoteThread", "VirtualAllocEx", "WriteProcessMemory", 
                        "SetWindowsHookExA", "SetWindowsHookExW", "ReadProcessMemory",
                        "URLDownloadToFileA", "InternetOpenA" 
                    };
                    var foundSuspicious = new List<string>();

                    foreach (var funcs in imports.Values)
                    {
                        foreach (var func in funcs)
                        {
                            if (suspiciousApis.Contains(func))
                            {
                                foundSuspicious.Add(func);
                            }
                        }
                    }

                    var result = new
                    {
                        file = Path.GetFileName(file_path),
                        architecture = is64Bit ? "x64" : (is32Bit ? "x86" : "Unknown"),
                        type = isDll ? "DLL" : (isExe ? "EXE" : "Unknown"),
                        imported_modules_count = imports.Count,
                        suspicious_apis_detected = foundSuspicious.Distinct().ToList(),
                        full_imports = imports
                    };

                    string json = System.Text.Json.JsonSerializer.Serialize(result, new System.Text.Json.JsonSerializerOptions { WriteIndented = true });
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = json } } };
                }
                catch (UnauthorizedAccessException uex)
                {
                    AuditLogger.Log(LogLevel.WARN, "PeTools", uex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Access Denied reading file: {uex.Message}" } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "PeTools", ex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"PE Analysis failed: {ex.Message}" } } };
                }
            });
        }
    }
}
#endif
