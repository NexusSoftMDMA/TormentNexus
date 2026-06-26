using System;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using System.Diagnostics;

namespace controller_mcp.Tests
{
    public class StatefulRecordingToolsTests
    {
        [Fact]
        public async System.Threading.Tasks.Task StartRecording_NonExistentWindow_ReturnsError()
        {
            var res = await StatefulRecordingTools.StartRecording("window", "NON_EXISTENT_WINDOW_12345", "C:\\out", "test.mp4");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public async System.Threading.Tasks.Task StopRecording_NonExistentId_ReturnsError()
        {
            var res = await StatefulRecordingTools.StopRecording("INVALID_ID_12345");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public async System.Threading.Tasks.Task ResumeRecording_WithValidPid_ReattachesSuccessfully()
        {
            // Arrange
            string testId = Guid.NewGuid().ToString();
            
            // We need a long running process that won't exit immediately to simulate ffmpeg
            var psi = new ProcessStartInfo
            {
                FileName = "cmd.exe",
                Arguments = "/c ping 127.0.0.1 -n 10",
                CreateNoWindow = true,
                UseShellExecute = false
            };
            var proc = Process.Start(psi);

            var backup = new RecordingBackup
            {
                Id = testId,
                ProcessId = proc.Id,
                StartTime = DateTime.Now,
                OutputPath = "test.mp4",
                TargetName = "TestTarget",
                IsolateAudioPid = -1
            };

            // Act
            StatefulRecordingTools.ResumeRecording(backup);
            
            // Assert
            var statusResult = StatefulRecordingTools.CheckRecordingStatus(testId);
            if (statusResult.IsError == true) 
            {
                var txt = ((ModelContextProtocol.Protocol.TextContentBlock)statusResult.Content[0]).Text;
                throw new Exception($"Failed to resume. CheckRecordingStatus returned error: {txt}");
            }
            Assert.True(statusResult.IsError != true);
            Assert.Contains("running", ((ModelContextProtocol.Protocol.TextContentBlock)statusResult.Content[0]).Text);

            // Cleanup
            var stopRes = await StatefulRecordingTools.StopRecording(testId);
            if (!proc.HasExited)
            {
                proc.Kill();
            }
        }
    }
}
