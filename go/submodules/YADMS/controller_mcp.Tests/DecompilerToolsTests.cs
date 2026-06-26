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
    public class DecompilerToolsTests
    {
        [Fact]
        public async Task DecompilerTools_AnalyzeDotNetAssembly_SucceedsOnSelf()
        {
            string selfPath = typeof(DecompilerTools).Assembly.Location;
            
            var result = await DecompilerTools.AnalyzeDotNetAssembly(selfPath);
            Assert.True(result.IsError != true);
            Assert.NotNull(result.Content);

            string json = ((TextContentBlock)result.Content[0]).Text;
            var doc = JsonDocument.Parse(json);
            
            bool foundDecompilerClass = false;
            foreach (var type in doc.RootElement.EnumerateArray())
            {
                if (type.GetProperty("ClassName").GetString() == "DecompilerTools")
                {
                    foundDecompilerClass = true;
                    break;
                }
            }

            Assert.True(foundDecompilerClass, "Failed to find DecompilerTools class within its own assembly.");
        }
    
        [Fact] public async Task DecompilerTools_AnalyzeDotNetAssembly_FailsGracefullyOnInvalidPath() { var result = await DecompilerTools.AnalyzeDotNetAssembly("Z:\\invalid\\path.dll"); Assert.True(result.IsError == true); }
    }
}
#endif
