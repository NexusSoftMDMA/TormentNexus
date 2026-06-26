#if GAME_HACKING
using System;
using System.Linq;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using System.Text.Json;

namespace controller_mcp.Tests
{
    [Collection("TestTarget")]
    public class MemoryToolsTests : IDisposable
    {
        private readonly TestTargetHelper _target;

        public MemoryToolsTests()
        {
            _target = new TestTargetHelper(hidden: true);
        }

        public void Dispose()
        {
            _target.Dispose();
        }

        [Fact(Skip = "Managed heap memory scanning via Win32 API is non-deterministic due to .NET GC page protections.")]
        public async Task MemoryTools_ScanAndWriteLiveMemory_SucceedsOnTestTarget()
        {
            int targetPid = _target.TargetProcess.Id;
            int initialValue = 1337;
            int newValue = 9999;

            // 1. Scan for the known 1337 integer
            var scanResult = await MemoryTools.ScanLiveMemoryInt32(targetPid, initialValue);
            
            Assert.True(scanResult.IsError != true);
            Assert.NotNull(scanResult.Content);
            
            string jsonResult = ((TextContentBlock)scanResult.Content[0]).Text;
            var doc = JsonDocument.Parse(jsonResult);
            
            int matches = doc.RootElement.GetProperty("matches_found").GetInt32();
            Assert.True(matches > 0, "Failed to find the 1337 value in the TestTarget memory.");

            // Get the first matched address
            var addresses = doc.RootElement.GetProperty("addresses").EnumerateArray().Select(a => a.GetString()).ToList();
            string firstAddress = addresses[0];

            // 2. Overwrite it with 9999
            var writeResult = await MemoryTools.WriteLiveMemoryInt32(targetPid, firstAddress, newValue);
            Assert.True(writeResult.IsError != true);
            Assert.Contains(newValue.ToString(), ((TextContentBlock)writeResult.Content[0]).Text);

            // 3. Scan for 9999 to confirm it was written
            var verifyResult = await MemoryTools.ScanLiveMemoryInt32(targetPid, newValue);
            Assert.True(verifyResult.IsError != true);
            
            string verifyJson = ((TextContentBlock)verifyResult.Content[0]).Text;
            var verifyDoc = JsonDocument.Parse(verifyJson);
            
            int verifyMatches = verifyDoc.RootElement.GetProperty("matches_found").GetInt32();
            var verifyAddresses = verifyDoc.RootElement.GetProperty("addresses").EnumerateArray().Select(a => a.GetString()).ToList();
            
            Assert.True(verifyMatches > 0, "Failed to find the written 9999 value.");
            Assert.Contains(firstAddress, verifyAddresses);
        }
    
        [Fact] public async Task MemoryTools_ScanLiveMemory_FailsGracefullyOnInvalidPid() { var result = await MemoryTools.ScanLiveMemoryInt32(-9999, 1337); Assert.True(result.IsError == true); }
    }
}
#endif
