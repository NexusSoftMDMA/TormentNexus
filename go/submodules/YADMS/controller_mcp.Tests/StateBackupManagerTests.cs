using System;
using System.IO;
using Xunit;
using controller_mcp.Features.Tools;
using System.Threading;

namespace controller_mcp.Tests
{
    [Collection("StateBackup")]
    public class StateBackupManagerTests : IDisposable
    {
        public StateBackupManagerTests()
        {
            // Clear singleton state before each test
            StateBackupManager.CurrentState = new DaemonStateBackup();
        }

        public void Dispose()
        {
            StateBackupManager.CurrentState = new DaemonStateBackup();
        }

        [Fact]
        public void SaveState_Debouncer_ExecutesAfterDelay()
        {
            StateBackupManager.AddTerminal("echo test", "C:\\", 12345);
            
            // Allow Debouncer to flush
            Thread.Sleep(1500);

            string path = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "state_backups.dat");
            Assert.True(File.Exists(path));
        }

        [Fact]
        public void RestoreState_WithNullJsonFields_DoesNotCrash()
        {
            // 1. Construct a bad JSON payload with explicitly null fields
            string badJson = "{\"Terminals\": null, \"Watchers\": null, \"CronTasks\": null, \"WebSockets\": null, \"SshSessions\": null, \"Analytics\": {\"ToolInvocations\": null}}";

            // 2. Use Reflection to access private AES encryption in StateBackupManager
            var type = typeof(StateBackupManager);
            var getKeyMethod = type.GetMethod("GetOrCreateMasterKey", System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Static);
            var encryptMethod = type.GetMethod("AesEncrypt", System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Static);

            byte[] key = (byte[])getKeyMethod.Invoke(null, null);
            byte[] encryptedData = (byte[])encryptMethod.Invoke(null, new object[] { badJson, key });

            // 3. Write corrupted state to disk
            string path = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "state_backups.dat");
            File.WriteAllBytes(path, encryptedData);

            // 4. Trigger RestoreState, which should deserialize the bad JSON and attempt to loop.
            // If the null-coalescing boundaries are missing, this will throw a NullReferenceException.
            var ex = Record.Exception(() => StateBackupManager.RestoreState());

            // 5. Allow background Task.Run in RestoreState to execute
            Thread.Sleep(1000);

            // 6. Assert no crash
            Assert.Null(ex);
        }

        [Fact]
        public void RestoreState_WithCorruptJson_DoesNotCrash()
        {
            // 1. Construct a completely mangled JSON payload
            string badJson = "{ \"Terminals\": [ { \"Command\": \"echo bad\", \"Wait, this isn't json! ] }";

            var type = typeof(StateBackupManager);
            var getKeyMethod = type.GetMethod("GetOrCreateMasterKey", System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Static);
            var encryptMethod = type.GetMethod("AesEncrypt", System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Static);

            byte[] key = (byte[])getKeyMethod.Invoke(null, null);
            byte[] encryptedData = (byte[])encryptMethod.Invoke(null, new object[] { badJson, key });

            string path = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "state_backups.dat");
            File.WriteAllBytes(path, encryptedData);

            // This should intercept JsonException and not crash
            var ex = Record.Exception(() => StateBackupManager.RestoreState());

            Thread.Sleep(500);

            Assert.Null(ex);
        }

        [Fact]
        public void ImportMasterKey_InvalidBase64_ThrowsFormatException()
        {
            Assert.Throws<FormatException>(() => StateBackupManager.ImportMasterKey("This is not valid base64!!!@@@"));
        }

        [Fact]
        public void ImportMasterKey_InvalidByteLength_ThrowsArgumentException()
        {
            // Valid base64, but only encodes 4 bytes ("test"), not 32 bytes
            string shortKey = Convert.ToBase64String(System.Text.Encoding.UTF8.GetBytes("test"));
            var ex = Assert.Throws<ArgumentException>(() => StateBackupManager.ImportMasterKey(shortKey));
            Assert.Contains("must be exactly 256 bits", ex.Message);
        }

        [Fact]
        public void AddAndRemovePcap_UpdatesStateCorrectly()
        {
            string id = Guid.NewGuid().ToString();
            StateBackupManager.AddPcap(id, 0, "tcp", 1234);
            
            Assert.Contains(StateBackupManager.CurrentState.Pcaps, p => p.Id == id);
            
            StateBackupManager.RemovePcap(id);
            Assert.DoesNotContain(StateBackupManager.CurrentState.Pcaps, p => p.Id == id);
        }

        [Fact]
        public void AddAndRemoveRecording_UpdatesStateCorrectly()
        {
            string id = Guid.NewGuid().ToString();
            StateBackupManager.AddRecording(id, 9999, DateTime.Now, "out.mp4", "Target", -1);
            
            Assert.Contains(StateBackupManager.CurrentState.Recordings, r => r.Id == id);
            
            StateBackupManager.RemoveRecording(id);
            Assert.DoesNotContain(StateBackupManager.CurrentState.Recordings, r => r.Id == id);
        }

        [Fact]
        public void ToolDisconnects_EradicateZombies_ByRemovingFromStateBackup()
        {
            // Terminal
            StateBackupManager.AddTerminal("zombie_cmd", "C:\\", 1111);
            StateBackupManager.RemoveTerminal("zombie_cmd");
            Assert.DoesNotContain(StateBackupManager.CurrentState.Terminals, t => t.Command == "zombie_cmd");

            // Watcher
            StateBackupManager.AddWatcher("C:\\zombie_dir", "*.*", true);
            StateBackupManager.RemoveWatcher("C:\\zombie_dir");
            Assert.DoesNotContain(StateBackupManager.CurrentState.Watchers, w => w.DirectoryPath == "C:\\zombie_dir");

            // WebSocket
            StateBackupManager.AddWebSocket("ws://zombie_socket");
            StateBackupManager.RemoveWebSocket("ws://zombie_socket");
            Assert.DoesNotContain(StateBackupManager.CurrentState.WebSockets, ws => ws.Url == "ws://zombie_socket");

            // SSH
            StateBackupManager.AddSsh("zombie_host", 22, "user", "pass", "");
            StateBackupManager.RemoveSsh("zombie_host");
            Assert.DoesNotContain(StateBackupManager.CurrentState.SshSessions, ssh => ssh.Host == "zombie_host");
        }
    }
}
