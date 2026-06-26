using System;
using System.Threading;
using System.Text.Json;
using System.Linq;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;

namespace controller_mcp.Tests
{
    [Collection("StateBackup")]
    public class TerminalToolsTests : IDisposable
    {
        public void Dispose()
        {
            TerminalTools.StopAll();
        }

        [Fact]
        public void TerminalTools_StartAndExecute_BuffersOutput()
        {
            // Start terminal
            var startRes = TerminalTools.TerminalStart();
            Assert.True(startRes.IsError != true);

            var startContent = startRes.Content.First() as TextContentBlock;
            var startJson = JsonDocument.Parse(startContent.Text);
            string terminalId = startJson.RootElement.GetProperty("terminal_id").GetString();
            Assert.NotNull(terminalId);

            // Run an echo command
            var runRes = TerminalTools.TerminalSend(terminalId, "echo TEST_TERMINAL_OUTPUT");
            Assert.True(runRes.IsError != true);

            // Allow the cmd.exe process to output and enqueue to buffer
            Thread.Sleep(1000);

            // Poll output
            var pollRes = TerminalTools.TerminalReceive(terminalId);
            Assert.True(pollRes.IsError != true);
            var pollContent = pollRes.Content.First() as TextContentBlock;

            Assert.Contains("TEST_TERMINAL_OUTPUT", pollContent.Text);

            // Stop terminal
            var stopRes = TerminalTools.TerminalKill(terminalId);
            Assert.True(stopRes.IsError != true);
        }
        [Fact]
        public void TerminalStart_DirectoryTraversal_ReturnsError()
        {
            var result = TerminalTools.TerminalStart("echo test", "../../Windows/System32");
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((TextContentBlock)result.Content[0]).Text);
        }
    
        [Fact] public async Task TerminalTools_RunCommand_FailsGracefullyOnInvalidCommand() { var result = TerminalTools.TerminalStart("NON_EXISTENT_COMMAND_XYZ", ""); Assert.True(result.IsError == true); }
    }
}
