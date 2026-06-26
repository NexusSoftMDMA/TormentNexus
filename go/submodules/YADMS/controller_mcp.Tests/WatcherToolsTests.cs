using System;
using System.IO;
using System.Text.Json;
using System.Threading;
using System.Linq;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;

namespace controller_mcp.Tests
{
    [Collection("StateBackup")] // Uses static StateBackupManager internally
    public class WatcherToolsTests : IDisposable
    {
        private string _testDir;

        public WatcherToolsTests()
        {
            _testDir = Path.Combine(Path.GetTempPath(), "WatcherToolsTests_" + Guid.NewGuid());
            Directory.CreateDirectory(_testDir);
        }

        public void Dispose()
        {
            WatcherTools.StopAll();
            try
            {
                if (Directory.Exists(_testDir)) Directory.Delete(_testDir, true);
            }
            catch { }
        }

        [Fact]
        public void WatcherTools_EndToEnd_DetectsFileCreation()
        {
            // Start watcher
            var startRes = WatcherTools.StartDirectoryWatcher(_testDir);
            Assert.True(startRes.IsError != true);

            var content = startRes.Content.First() as TextContentBlock;
            var json = JsonDocument.Parse(content.Text);
            string watcherId = json.RootElement.GetProperty("watcher_id").GetString();
            Assert.NotNull(watcherId);

            // Create file
            string testFile = Path.Combine(_testDir, "test.txt");
            File.WriteAllText(testFile, "Hello World");

            // Give FileSystemWatcher time to trigger and buffer
            Thread.Sleep(500);

            // Poll watcher
            var pollRes = WatcherTools.PollWatcherEvents(watcherId);
            Assert.True(pollRes.IsError != true);
            var pollContent = pollRes.Content.First() as TextContentBlock;
            
            Assert.Contains("[Created]", pollContent.Text);
            Assert.Contains("test.txt", pollContent.Text);

            // Stop watcher
            var stopRes = WatcherTools.StopDirectoryWatcher(watcherId);
            Assert.True(stopRes.IsError != true);
        }
    
        [Fact] public void WatcherTools_StartWatching_FailsGracefullyOnInvalidPath() { var result = WatcherTools.StartDirectoryWatcher("Z:\\invalid\\path"); Assert.True(result.IsError == true); }
    }
}
