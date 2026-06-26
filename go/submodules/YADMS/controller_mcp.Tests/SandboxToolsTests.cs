using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;

namespace controller_mcp.Tests
{
    public class SandboxToolsTests
    {
        [Fact]
        public async Task LaunchSandboxedProcess_DirectoryTraversal_ReturnsError()
        {
            // Act
            var result = await SandboxTools.LaunchSandboxedProcess("../../Windows/System32/cmd.exe");

            // Assert
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((TextContentBlock)result.Content[0]).Text);
        }
    
        [Fact] public async Task SandboxTools_Execute_FailsGracefullyOnInvalidPath() { var result = await SandboxTools.LaunchSandboxedProcess("Z:\\invalid\\path.exe"); Assert.True(result.IsError == true); }
    }
}
