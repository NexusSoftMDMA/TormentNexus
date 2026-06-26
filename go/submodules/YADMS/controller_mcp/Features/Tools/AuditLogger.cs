using System;
using System.IO;
using System.IO.Compression;
using System.Text.Json;
using System.Threading.Tasks;
using System.Collections.Concurrent;
using System.Threading;

namespace controller_mcp.Features.Tools
{
    public enum LogLevel { DEBUG, INFO, WARN, ERROR, FATAL }

    public static class AuditLogger
    {
        private static string _logDirectory;
        private static string _logFilePath;
        private static readonly int MaxFileSizeBytes = 10 * 1024 * 1024; // 10MB
        
        private static readonly BlockingCollection<string> _logQueue = new BlockingCollection<string>();
        private static bool _isProcessing = false;

        // Event for the UI to subscribe to
        public static event Action<string> OnLogWritten;

        public static void Initialize(string logDir = null)
        {
            Reconfigure(logDir);
            
            if (!_isProcessing)
            {
                _isProcessing = true;
                Task.Factory.StartNew(ProcessLogQueue, TaskCreationOptions.LongRunning);
            }
        }

        public static void Reconfigure(string logDir)
        {
            if (!string.IsNullOrEmpty(logDir))
            {
                try { _logDirectory = Path.GetFullPath(logDir); } catch { _logDirectory = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "logs"); }
            }
            else
            {
                _logDirectory = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "logs");
            }

            try
            {
                if (!Directory.Exists(_logDirectory))
                {
                    Directory.CreateDirectory(_logDirectory);
                }
            }
            catch
            {
                // Fallback to Temp directory if BaseDirectory is restricted
                try
                {
                    _logDirectory = Path.Combine(Path.GetTempPath(), "ControllerMCP_Logs");
                    if (!Directory.Exists(_logDirectory))
                        Directory.CreateDirectory(_logDirectory);
                }
                catch
                {
                    // If everything fails, disable file logging to prevent crashes
                    _logDirectory = null;
                }
            }

            if (_logDirectory != null)
            {
                _logFilePath = Path.Combine(_logDirectory, $"mcp_audit_{DateTime.Now:yyyyMMdd}.log");
            }
            else
            {
                _logFilePath = null;
            }
        }

        public static void Log(LogLevel level, string category, string message)
        {
            var settings = AppSettings.Load(suppressLogging: true);
            if (level == LogLevel.DEBUG && !settings.EnableDebugLogging)
                return; // Skip debug logs if disabled

            string timestamp = DateTime.UtcNow.ToString("O");
            string formattedLog = $"[{timestamp}] [{level}] [{category}] {message}";

            // Fire event for the UI Console
            OnLogWritten?.Invoke(formattedLog);

            var logObj = new
            {
                Timestamp = timestamp,
                Level = level.ToString(),
                Category = category,
                Message = message
            };

            string json = JsonSerializer.Serialize(logObj);
            _logQueue.Add(json);
        }

        private static void ProcessLogQueue()
        {
            foreach (var json in _logQueue.GetConsumingEnumerable())
            {
                try
                {
                    if (string.IsNullOrEmpty(_logFilePath)) continue;

                    if (File.Exists(_logFilePath))
                    {
                        var info = new FileInfo(_logFilePath);
                        if (info.Length > MaxFileSizeBytes)
                        {
                            string backupFile = Path.Combine(_logDirectory, $"mcp_audit_{DateTime.Now:yyyyMMdd_HHmmss}.zip");
                            using (var archive = ZipFile.Open(backupFile, ZipArchiveMode.Create))
                            {
                                archive.CreateEntryFromFile(_logFilePath, Path.GetFileName(_logFilePath));
                            }
                            File.Delete(_logFilePath);
                        }
                    }

                    File.AppendAllText(_logFilePath, json + Environment.NewLine);
                }
                catch (Exception ex) 
                { 
                    try 
                    { 
                        File.AppendAllText(Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "audit_error.txt"), ex.ToString() + Environment.NewLine); 
                    } 
                    catch { } // Swallow entirely to prevent Task loop termination
                }
            }
        }

        // Backwards compatibility for existing tools
        public static void LogSystemEvent(string category, string message)
        {
            Log(LogLevel.INFO, category, message);
        }

        public static void LogJsonRpc(string direction, string rawJson)
        {
            var settings = AppSettings.Load(suppressLogging: true);
            if (!settings.EnableDebugLogging) return; // RPC is very noisy, treat as debug

            string payloadStr = rawJson;
            if (!string.IsNullOrEmpty(payloadStr) && payloadStr.Length > 5000)
            {
                payloadStr = payloadStr.Substring(0, 5000) + "... [TRUNCATED_FOR_LOGS: Payload exceeded 5000 chars]";
            }

            Log(LogLevel.DEBUG, "JSON_RPC", $"[{direction}] {payloadStr}");
        }
    }
}
