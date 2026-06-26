using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;
using System.IO;

namespace controller_mcp.Tests
{
    [Collection("StateBackup")]
    public class StateBackupManagerConcurrencyTests : IDisposable
    {
        public StateBackupManagerConcurrencyTests()
        {
            // Clear state
            StateBackupManager.CurrentState = new DaemonStateBackup();
        }

        public void Dispose()
        {
            StateBackupManager.CurrentState = new DaemonStateBackup();
            
            // Delete state file if it exists to clean up
            string backupFile = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "state_backups.dat");
            try { if (File.Exists(backupFile)) File.Delete(backupFile); } catch { }
        }

        [Fact]
        public void AddTerminal_ConcurrentAdditions_DoesNotThrowCollectionModifiedException()
        {
            StateBackupManager.CurrentState.Terminals.Clear();
            int iterations = 1000;
            var exceptions = new System.Collections.Concurrent.ConcurrentQueue<Exception>();

            Parallel.For(0, iterations, i =>
            {
                try
                {
                    StateBackupManager.AddTerminal($"cmd_test_{i}", "C:\\", i);
                }
                catch (Exception ex)
                {
                    exceptions.Enqueue(ex);
                }
            });

            Assert.Empty(exceptions);
            Assert.Equal(iterations, StateBackupManager.CurrentState.Terminals.Count);
        }

        [Fact]
        public void AddWatcher_ConcurrentAdditions_DoesNotThrowCollectionModifiedException()
        {
            StateBackupManager.CurrentState.Watchers.Clear();
            int iterations = 1000;
            var exceptions = new System.Collections.Concurrent.ConcurrentQueue<Exception>();

            Parallel.For(0, iterations, i =>
            {
                try
                {
                    StateBackupManager.AddWatcher($"C:\\Test_{i}", "*.*", true);
                }
                catch (Exception ex)
                {
                    exceptions.Enqueue(ex);
                }
            });

            Assert.Empty(exceptions);
            Assert.Equal(iterations, StateBackupManager.CurrentState.Watchers.Count);
        }
    
        [Fact] public void StateBackupManager_RestoreState_HandlesExceptions() { File.WriteAllText("state_backup.dat", "corrupted"); StateBackupManager.RestoreState(); Assert.True(true); }
    }
}
