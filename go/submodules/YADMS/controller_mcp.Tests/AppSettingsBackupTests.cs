using System;
using System.IO;
using Xunit;
using controller_mcp;

[assembly: CollectionBehavior(DisableTestParallelization = true)]

namespace controller_mcp.Tests
{
    public class AppSettingsBackupTests
    {
        [Fact]
        public void AppSettings_Save_CreatesBackupFile()
        {
            var settings = new AppSettings { EnableDebugLogging = true, FFmpegPath = "TEST_PATH" };
            settings.Save();

            // ConfigPath relies on Environment.GetFolderPath, so we verify against the known locations
            Assert.True(File.Exists(AppSettings.ConfigPath));
            
            // If we save a second time, it should copy to .bak
            settings.EnableDebugLogging = false;
            settings.Save();

            Assert.True(File.Exists(AppSettings.ConfigBackupPath));
        }

        [Fact]
        public void AppSettings_Load_FallsBackToBackupOnCorruption()
        {
            if (File.Exists(AppSettings.ConfigPath)) File.Delete(AppSettings.ConfigPath);
            if (File.Exists(AppSettings.ConfigBackupPath)) File.Delete(AppSettings.ConfigBackupPath);

            var settings = new AppSettings { EnableDebugLogging = true, FFmpegPath = "BACKUP_VALUE" };
            settings.Save(); // Writes config.json
            settings.Save(); // Copies config.json to config.bak

            // Corrupt config.json
            File.WriteAllText(AppSettings.ConfigPath, "{ INVALID JSON DATA }");

            // Clear the static cache so it actually reads from disk
            typeof(AppSettings).GetField("_current", System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Static).SetValue(null, null);

            // Load should intercept corruption and deserialize config.bak
            var loaded = AppSettings.Load();

            Assert.True(loaded.EnableDebugLogging);
            Assert.Equal("BACKUP_VALUE", loaded.FFmpegPath);
        }
    
        [Fact] public void AppSettingsBackup_Load_FailsGracefullyOnCorruptedFile() { File.WriteAllText("config.json", "{invalid_json_!@#"); var res = AppSettings.Load(); Assert.NotNull(res); }
    }
}
