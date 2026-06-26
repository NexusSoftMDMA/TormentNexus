#if GAME_HACKING
using System;
using System.IO;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using System.Text.Json;

namespace controller_mcp.Tests
{
    public class PeToolsTests
    {
        [Fact]
        public async Task PeTools_AnalyzeExecutable_SucceedsOnSelf()
        {
            string selfPath = typeof(PeTools).Assembly.Location;
            
            var result = await PeTools.AnalyzeExecutable(selfPath);
            Assert.True(result.IsError != true);
            Assert.NotNull(result.Content);

            string json = ((TextContentBlock)result.Content[0]).Text;
            var doc = JsonDocument.Parse(json);
            
            string type = doc.RootElement.GetProperty("type").GetString();
            Assert.True(type == "DLL" || type == "EXE");

            int importsCount = doc.RootElement.GetProperty("imported_modules_count").GetInt32();
            Assert.True(importsCount > 0, "No imported modules found, which is mathematically impossible for a managed assembly.");
        }
    
        [Fact] public async Task PeTools_AnalyzeExecutable_FailsGracefullyOnInvalidPath() { var result = await PeTools.AnalyzeExecutable("Z:\\invalid\\path.exe"); Assert.True(result.IsError == true); }
    }
}
#endif
