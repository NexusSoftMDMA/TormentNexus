using System;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using System.Text.Json;

namespace controller_mcp.Tests
{
    [Collection("TestTarget")]
    public class DumpToolsTests : IDisposable
    {
        private readonly TestTargetHelper _target;

        public DumpToolsTests()
        {
            _target = new TestTargetHelper(hidden: true);
        }

        public void Dispose()
        {
            _target.Dispose();
        }

        [Fact]
        public async Task DumpTools_CreateAndAnalyzeDump_Succeeds()
        {
            int pid = _target.TargetProcess.Id;

            // 1. Create Dump
            var createResult = await DumpTools.CreateCrashDump(pid);
            Assert.True(createResult.IsError != true, $"Dump creation failed: {((TextContentBlock)createResult.Content?[0])?.Text}");
            
            // Extract dump path from text
            string content = ((TextContentBlock)createResult.Content[0]).Text;
            Assert.Contains("Successfully generated dump file:", content);
            string dumpPath = content.Replace("Successfully generated dump file: ", "").Trim();

            Assert.True(File.Exists(dumpPath), "Dump file was not physically written to disk.");

            // 2. Analyze Dump
            var analyzeResult = await DumpTools.AnalyzeDotNetDump(dumpPath);
            Assert.True(analyzeResult.IsError != true, $"Dump analysis failed: {((TextContentBlock)analyzeResult.Content?[0])?.Text}");
            
            string report = ((TextContentBlock)analyzeResult.Content[0]).Text;
            Assert.Contains("=== .NET Crash Dump Analysis ===", report);

            // Cleanup
            if (File.Exists(dumpPath))
            {
                File.Delete(dumpPath);
            }
        }
    
        [Fact] public async Task DumpTools_CreateCrashDump_FailsGracefullyOnInvalidPid() { var result = await DumpTools.CreateCrashDump(-9999); Assert.True(result.IsError == true); }
    }
}
