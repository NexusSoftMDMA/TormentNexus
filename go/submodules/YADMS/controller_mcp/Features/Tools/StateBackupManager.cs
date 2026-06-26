using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.IO;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace controller_mcp.Features.Tools
{
    public class DaemonStateBackup
    {
        public List<CronTaskBackup> CronTasks { get; set; } = new List<CronTaskBackup>();
        public List<WatcherBackup> Watchers { get; set; } = new List<WatcherBackup>();
        public List<TerminalBackup> Terminals { get; set; } = new List<TerminalBackup>();
        public List<WebSocketBackup> WebSockets { get; set; } = new List<WebSocketBackup>();
        public List<SshBackup> SshSessions { get; set; } = new List<SshBackup>();
        public List<RecordingBackup> Recordings { get; set; } = new List<RecordingBackup>();
        public List<PcapBackup> Pcaps { get; set; } = new List<PcapBackup>();
        public AnalyticsData Analytics { get; set; } = new AnalyticsData();
    }

    public class CronTaskBackup
    {
        public string ActionType { get; set; }
        public string Target { get; set; }
        public string Payload { get; set; }
        public int IntervalSeconds { get; set; }
    }

    public class WatcherBackup
    {
        public string DirectoryPath { get; set; }
        public string Filter { get; set; }
        public bool IncludeSubdirectories { get; set; }
    }

    public class TerminalBackup
    {
        public string Command { get; set; }
        public string WorkingDirectory { get; set; }
        public int ProcessId { get; set; }
    }

    public class WebSocketBackup
    {
        public string Url { get; set; }
    }

    public class SshBackup
    {
        public string Host { get; set; }
        public int Port { get; set; }
        public string Username { get; set; }
        public string Password { get; set; }
        public string PrivateKeyPath { get; set; }
    }

    public class RecordingBackup
    {
        public string Id { get; set; }
        public int ProcessId { get; set; }
        public DateTime StartTime { get; set; }
        public string OutputPath { get; set; }
        public string TargetName { get; set; }
        public int IsolateAudioPid { get; set; }
    }

    public class PcapBackup
    {
        public string Id { get; set; }
        public int DeviceIndex { get; set; }
        public string Filter { get; set; }
        public int TargetPid { get; set; }
    }

    public static class StateBackupManager
    {
        private static readonly string BackupFilePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "state_backups.dat");
        private static readonly string MasterKeyFilePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "master_key.dat");
        private static readonly object _fileLock = new object();
        private static readonly System.Threading.Timer _saveDebouncer = new System.Threading.Timer(SaveStateSynchronous, null, System.Threading.Timeout.Infinite, System.Threading.Timeout.Infinite);
        public static DaemonStateBackup CurrentState { get; set; } = new DaemonStateBackup();

        private static byte[] GetOrCreateMasterKey()
        {
            if (File.Exists(MasterKeyFilePath))
            {
                try
                {
                    byte[] encryptedKey = File.ReadAllBytes(MasterKeyFilePath);
                    return ProtectedData.Unprotect(encryptedKey, null, DataProtectionScope.LocalMachine);
                }
                catch (Exception ex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "StateBackup", $"Failed to unseal Master Key via DPAPI. If this daemon was moved from another PC, you must import the exported key. Error: {ex.Message}");
                    throw;
                }
            }

            // Generate new 32-byte (256-bit) AES key
            byte[] newKey = new byte[32];
            using (var rng = new RNGCryptoServiceProvider())
            {
                rng.GetBytes(newKey);
            }

            byte[] protectedKey = ProtectedData.Protect(newKey, null, DataProtectionScope.LocalMachine);
            File.WriteAllBytes(MasterKeyFilePath, protectedKey);
            AuditLogger.Log(LogLevel.INFO, "StateBackup", "Generated new DPAPI-sealed Master Key for State Backup.");
            return newKey;
        }

        public static string ExportMasterKey()
        {
            byte[] key = GetOrCreateMasterKey();
            return Convert.ToBase64String(key);
        }

        public static void ImportMasterKey(string base64Key)
        {
            byte[] newKey = Convert.FromBase64String(base64Key);
            if (newKey.Length != 32) throw new ArgumentException("Master Key must be exactly 256 bits (32 bytes).");

            byte[] protectedKey = ProtectedData.Protect(newKey, null, DataProtectionScope.LocalMachine);
            File.WriteAllBytes(MasterKeyFilePath, protectedKey);
            AuditLogger.Log(LogLevel.INFO, "StateBackup", "Successfully imported and sealed Master Key via DPAPI.");
        }

        private static byte[] AesEncrypt(string plaintext, byte[] key)
        {
            using (var aes = new AesCryptoServiceProvider { Key = key, Mode = CipherMode.CBC, Padding = PaddingMode.PKCS7 })
            {
                aes.GenerateIV();
                var encryptor = aes.CreateEncryptor();
                byte[] plainBytes = Encoding.UTF8.GetBytes(plaintext);
                byte[] cipherBytes = encryptor.TransformFinalBlock(plainBytes, 0, plainBytes.Length);
                
                byte[] resultBytes = new byte[aes.IV.Length + cipherBytes.Length];
                Buffer.BlockCopy(aes.IV, 0, resultBytes, 0, aes.IV.Length);
                Buffer.BlockCopy(cipherBytes, 0, resultBytes, aes.IV.Length, cipherBytes.Length);
                return resultBytes;
            }
        }

        private static string AesDecrypt(byte[] payload, byte[] key)
        {
            using (var aes = new AesCryptoServiceProvider { Key = key, Mode = CipherMode.CBC, Padding = PaddingMode.PKCS7 })
            {
                byte[] iv = new byte[16];
                byte[] cipher = new byte[payload.Length - 16];
                Buffer.BlockCopy(payload, 0, iv, 0, 16);
                Buffer.BlockCopy(payload, 16, cipher, 0, cipher.Length);

                aes.IV = iv;
                var decryptor = aes.CreateDecryptor();
                byte[] plainBytes = decryptor.TransformFinalBlock(cipher, 0, cipher.Length);
                return Encoding.UTF8.GetString(plainBytes);
            }
        }

        public static void SaveState()
        {
            _saveDebouncer.Change(1000, System.Threading.Timeout.Infinite);
        }

        private static void SaveStateSynchronous(object state)
        {
            try
            {
                lock (_fileLock)
                {
                    CurrentState.Analytics = AnalyticsManager.Current;
                    string json = JsonSerializer.Serialize(CurrentState, new JsonSerializerOptions { WriteIndented = false });
                    byte[] key = GetOrCreateMasterKey();
                    byte[] encryptedData = AesEncrypt(json, key);
                    File.WriteAllBytes(BackupFilePath, encryptedData);
                }
            }
            catch (Exception ex)
            {
                AuditLogger.Log(LogLevel.ERROR, "StateBackup", $"Failed to securely save state: {ex.Message}");
            }
        }

        public static void AddCronTask(string actionType, string target, string payload, int intervalSeconds)
        {
            lock (_fileLock)
            {
                CurrentState.CronTasks.Add(new CronTaskBackup { ActionType = actionType, Target = target, Payload = payload, IntervalSeconds = intervalSeconds });
            }
            SaveState();
        }

        public static void RemoveCronTask(string target)
        {
            lock (_fileLock)
            {
                CurrentState.CronTasks.RemoveAll(x => x.Target == target);
            }
            SaveState();
        }

        public static void AddWatcher(string path, string filter, bool subdirs)
        {
            lock (_fileLock)
            {
                CurrentState.Watchers.Add(new WatcherBackup { DirectoryPath = path, Filter = filter, IncludeSubdirectories = subdirs });
            }
            SaveState();
        }

        public static void RemoveWatcher(string path)
        {
            lock (_fileLock)
            {
                CurrentState.Watchers.RemoveAll(x => x.DirectoryPath == path);
            }
            SaveState();
        }

        public static void AddTerminal(string cmd, string cwd, int processId)
        {
            lock (_fileLock)
            {
                CurrentState.Terminals.Add(new TerminalBackup { Command = cmd, WorkingDirectory = cwd, ProcessId = processId });
            }
            SaveState();
        }

        public static void RemoveTerminal(string cmd)
        {
            lock (_fileLock)
            {
                CurrentState.Terminals.RemoveAll(x => x.Command == cmd);
            }
            SaveState();
        }

        public static void AddWebSocket(string url)
        {
            lock (_fileLock)
            {
                CurrentState.WebSockets.Add(new WebSocketBackup { Url = url });
            }
            SaveState();
        }

        public static void RemoveWebSocket(string url)
        {
            lock (_fileLock)
            {
                CurrentState.WebSockets.RemoveAll(x => x.Url == url);
            }
            SaveState();
        }

        public static void AddSsh(string host, int port, string user, string pass, string keyPath)
        {
            lock (_fileLock)
            {
                CurrentState.SshSessions.Add(new SshBackup { Host = host, Port = port, Username = user, Password = pass, PrivateKeyPath = keyPath });
            }
            SaveState();
        }

        public static void RemoveSsh(string host)
        {
            lock (_fileLock)
            {
                CurrentState.SshSessions.RemoveAll(x => x.Host == host);
            }
            SaveState();
        }

        public static void AddRecording(string id, int processId, DateTime startTime, string outputPath, string targetName, int isolateAudioPid)
        {
            lock (_fileLock)
            {
                CurrentState.Recordings.Add(new RecordingBackup { Id = id, ProcessId = processId, StartTime = startTime, OutputPath = outputPath, TargetName = targetName, IsolateAudioPid = isolateAudioPid });
            }
            SaveState();
        }

        public static void RemoveRecording(string id)
        {
            lock (_fileLock)
            {
                CurrentState.Recordings.RemoveAll(x => x.Id == id);
            }
            SaveState();
        }

        public static void AddPcap(string id, int deviceIndex, string filter, int targetPid)
        {
            lock (_fileLock)
            {
                CurrentState.Pcaps.Add(new PcapBackup { Id = id, DeviceIndex = deviceIndex, Filter = filter, TargetPid = targetPid });
            }
            SaveState();
        }

        public static void RemovePcap(string id)
        {
            lock (_fileLock)
            {
                CurrentState.Pcaps.RemoveAll(x => x.Id == id);
            }
            SaveState();
        }

        public static void RestoreState()
        {
            if (!File.Exists(BackupFilePath)) return;

            try
            {
                string json;
                lock (_fileLock)
                {
                    byte[] encryptedData = File.ReadAllBytes(BackupFilePath);
                    byte[] key = GetOrCreateMasterKey();
                    json = AesDecrypt(encryptedData, key);
                }

                var backup = JsonSerializer.Deserialize<DaemonStateBackup>(json);
                if (backup == null) return;

                AnalyticsManager.Restore(backup.Analytics);

                // Clear current lists because the tools will re-register themselves when they are recreated.
                CurrentState = new DaemonStateBackup();

                Task.Run(async () =>
                {
                    foreach (var t in backup.Terminals ?? new List<TerminalBackup>()) 
                    { 
                        try 
                        { 
                            if (t.ProcessId > 0)
                            {
                                var oldProcess = System.Diagnostics.Process.GetProcessById(t.ProcessId);
                                if (oldProcess.ProcessName.Contains("cmd") || oldProcess.ProcessName.Contains("powershell"))
                                {
                                    var psi = new System.Diagnostics.ProcessStartInfo("taskkill", $"/F /T /PID {t.ProcessId}")
                                    {
                                        CreateNoWindow = true,
                                        UseShellExecute = false
                                    };
                                    System.Diagnostics.Process.Start(psi)?.WaitForExit();
                                }
                            }
                        } 
                        catch { } 
                        try { TerminalTools.TerminalStart(t.Command, t.WorkingDirectory); } catch { } 
                    }
                    foreach (var w in backup.Watchers ?? new List<WatcherBackup>()) { try { WatcherTools.StartDirectoryWatcher(w.DirectoryPath, w.Filter, w.IncludeSubdirectories); } catch { } }
                    foreach (var c in backup.CronTasks ?? new List<CronTaskBackup>()) { try { CronTools.ScheduleRecurringTask(c.ActionType, c.Target, c.IntervalSeconds, c.Payload); } catch { } }
                    foreach (var ws in backup.WebSockets ?? new List<WebSocketBackup>()) { try { await WebSocketTools.WebsocketConnect(ws.Url); } catch { } }
                    foreach (var ssh in backup.SshSessions ?? new List<SshBackup>()) { try { SshTools.SshConnect(ssh.Host, ssh.Port, ssh.Username, ssh.Password, ssh.PrivateKeyPath); } catch { } }
                    foreach (var r in backup.Recordings ?? new List<RecordingBackup>()) { try { StatefulRecordingTools.ResumeRecording(r); } catch { } }
                    foreach (var p in backup.Pcaps ?? new List<PcapBackup>()) { try { PcapTools.RestartPacketCapture(p); } catch { } }

                    AuditLogger.Log(LogLevel.INFO, "StateBackup", "Successfully decrypted and restored state from disk.");
                });
            }
            catch (Exception ex)
            {
                AuditLogger.Log(LogLevel.ERROR, "StateBackup", $"Failed to decrypt and restore state: {ex.Message}");
            }
        }
    }
}
