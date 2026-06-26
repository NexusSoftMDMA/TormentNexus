using System;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class WindowAndProcessToolsTests
    {
        [Fact]
        public void KillProcess_NonExistent_ReturnsError()
        {
            var res = WindowAndProcessTools.KillProcess(-9999);
            Assert.True(res.IsError == true);
        }

        [Fact]
        public void CloseWindow_NonExistent_ReturnsError()
        {
            var res = WindowAndProcessTools.CloseWindow(-9999);
            Assert.True(res.IsError == true);
        }
        [Fact]
        public void StartDetachedProcess_DirectoryTraversal_ReturnsError()
        {
            var result = WindowAndProcessTools.StartDetachedProcess("../../Windows/System32/cmd.exe", "");
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((ModelContextProtocol.Protocol.TextContentBlock)result.Content[0]).Text);
        }
    }
}
