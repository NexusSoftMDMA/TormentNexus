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
using Mono.Cecil;
using Mono.Cecil.Cil;

namespace controller_mcp.Features.Tools
{
    public static class DecompilerTools
    {
        [McpServerTool, Description("Decompiles a compiled .NET .dll or .exe file and lists all Classes and Methods contained within it.")]
        public static async Task<CallToolResult> AnalyzeDotNetAssembly(string file_path)
        {
            return await Task.Run(() =>
            {
                try
                {
                    string safePath = InputValidator.ValidateFilePath(file_path, nameof(file_path));

                    if (!File.Exists(safePath))
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File not found: {safePath}" } } };

                    var resultList = new List<object>();

                    using (var module = ModuleDefinition.ReadModule(safePath))
                    {
                        foreach (var type in module.Types)
                        {
                            if (type.Name == "<Module>") continue;

                            var methods = new List<string>();
                            foreach (var method in type.Methods)
                            {
                                methods.Add($"{method.Name}({string.Join(", ", method.Parameters.Select(p => p.ParameterType.Name))})");
                            }

                            resultList.Add(new
                            {
                                Namespace = type.Namespace,
                                ClassName = type.Name,
                                Methods = methods
                            });
                        }
                    }

                    string json = System.Text.Json.JsonSerializer.Serialize(resultList, new System.Text.Json.JsonSerializerOptions { WriteIndented = true });
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = json } } };
                }
                catch (BadImageFormatException bex)
                {
                    AuditLogger.Log(LogLevel.WARN, "DecompilerTools", bex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File is not a valid .NET Assembly or is corrupted. Details: {bex.Message}" } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "DecompilerTools", ex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Decompilation failed: {ex.Message}" } } };
                }
            });
        }

        [McpServerTool, Description("Injects raw Intermediate Language (IL) logic into every method of a compiled .NET .dll, then recompiles it. Specifically, it injects a Console.WriteLine('TELEMETRY: [MethodName]') at the very start of every function, allowing you to instantly profile exactly what a foreign .dll is doing.")]
        public static async Task<CallToolResult> InjectTelemetryPayload(string file_path, string output_path)
        {
            return await Task.Run(() =>
            {
                try
                {
                    string safeInputPath = InputValidator.ValidateFilePath(file_path, nameof(file_path));
                    string safeOutputPath = InputValidator.ValidateFilePath(output_path, nameof(output_path));

                    if (!File.Exists(safeInputPath))
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File not found: {safeInputPath}" } } };

                    using (var module = ModuleDefinition.ReadModule(safeInputPath))
                    {
                        // Get a reference to System.Console.WriteLine(string)
                        var consoleType = module.ImportReference(typeof(System.Console));
                        var writeLineMethod = module.ImportReference(
                            typeof(System.Console).GetMethod("WriteLine", new[] { typeof(string) })
                        );

                        int modifiedMethodsCount = 0;

                        foreach (var type in module.Types)
                        {
                            if (type.Name == "<Module>") continue;

                            foreach (var method in type.Methods)
                            {
                                if (!method.HasBody) continue;

                                var ilProcessor = method.Body.GetILProcessor();
                                var firstInstruction = method.Body.Instructions.FirstOrDefault();

                                if (firstInstruction != null)
                                {
                                    // Inject: Console.WriteLine("TELEMETRY: Namespace.Class.Method")
                                    string payloadStr = $"TELEMETRY: {type.FullName}.{method.Name}";
                                    var strInstruction = ilProcessor.Create(OpCodes.Ldstr, payloadStr);
                                    var callInstruction = ilProcessor.Create(OpCodes.Call, writeLineMethod);

                                    ilProcessor.InsertBefore(firstInstruction, strInstruction);
                                    ilProcessor.InsertAfter(strInstruction, callInstruction);

                                    modifiedMethodsCount++;
                                }
                            }
                        }

                        module.Write(output_path);

                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully injected IL telemetry into {modifiedMethodsCount} methods. Modified assembly saved to: {output_path}" } } };
                    }
                }
                catch (BadImageFormatException bex)
                {
                    AuditLogger.Log(LogLevel.WARN, "DecompilerTools", bex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File is not a valid .NET Assembly. {bex.Message}" } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "DecompilerTools", ex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"IL Injection failed: {ex.Message}" } } };
                }
            });
        }
    }
}
#endif
