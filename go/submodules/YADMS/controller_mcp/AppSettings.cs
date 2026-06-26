using System;
using System.IO;
using System.Text.Json;

namespace controller_mcp
{
    public class AppSettings
    {
        private static AppSettings _current;
        public string LogDirectory { get; set; } = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "logs");
        public string FFmpegPath { get; set; } = "";
        public bool EnableDebugLogging { get; set; } = false;
        public bool MasterKeyExported { get; set; } = false;
        public bool EnableAutoUpdate { get; set; } = true;
        public DateTime LastUpdateCheck { get; set; } = DateTime.MinValue;

        public static string ConfigDir => Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData), "ControllerMCP");
        public static string ConfigPath => Path.Combine(ConfigDir, "config.json");
        public static string ConfigBackupPath => Path.Combine(ConfigDir, "config.bak");

        public static AppSettings Load(bool suppressLogging = false)
        {
            if (_current != null) return _current;
            if (File.Exists(ConfigPath))
            {
                try
                {
                    string json = File.ReadAllText(ConfigPath);
                    var settings = JsonSerializer.Deserialize<AppSettings>(json);
                    if (settings != null) 
                    {
                        _current = settings;
                        return _current;
                    }
                }
                catch (Exception ex)
                {
                    if (!suppressLogging)
                        controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.ERROR, "System", $"Failed to load primary config.json: {ex.Message}");
                } // Fallback to bak
            }

            if (File.Exists(ConfigBackupPath))
            {
                try
                {
                    string json = File.ReadAllText(ConfigBackupPath);
                    var settings = JsonSerializer.Deserialize<AppSettings>(json);
                    if (settings != null)
                    {
                        _current = settings;
                        return _current;
                    }
                }
                catch (Exception ex)
                {
                    if (!suppressLogging)
                        controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.ERROR, "System", $"Failed to load backup config.bak: {ex.Message}");
                }
            }

            _current = new AppSettings();
            return _current;
        }

        public void Save()
        {
            try
            {
                if (!Directory.Exists(ConfigDir))
                    Directory.CreateDirectory(ConfigDir);

                if (File.Exists(ConfigPath))
                {
                    File.Copy(ConfigPath, ConfigBackupPath, true);
                }

                string json = JsonSerializer.Serialize(this, new JsonSerializerOptions { WriteIndented = true });
                File.WriteAllText(ConfigPath, json);
                _current = this;
            }
            catch (Exception ex)
            {
                controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.ERROR, "System", $"Critical failure saving configuration: {ex.Message}");
            }
        }
    }
}
