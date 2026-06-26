using System;
using System.IO;
using Xunit;
using controller_mcp;

namespace controller_mcp.Tests
{
    public class AppSettingsTests : IDisposable
    {
        private string _testConfigDir;
        
        public AppSettingsTests()
        {
            _testConfigDir = Path.Combine(Path.GetTempPath(), "ControllerMCP_Tests_" + Guid.NewGuid());
            Directory.CreateDirectory(_testConfigDir);
            
            // We cannot easily mock the static ConfigPath if it relies on Environment, 
            // but we can test the serialization structure manually.
        }

        public void Dispose()
        {
            if (Directory.Exists(_testConfigDir))
            {
                Directory.Delete(_testConfigDir, true);
            }
        }

        [Fact]
        public void AppSettings_DefaultValues_AreCorrect()
        {
            var settings = new AppSettings();
            
            Assert.False(settings.EnableDebugLogging);
            Assert.Equal("", settings.FFmpegPath);
            Assert.Contains("logs", settings.LogDirectory);
        }

        [Fact]
        public void AppSettings_Serialization_Works()
        {
            var settings = new AppSettings
            {
                EnableDebugLogging = true,
                FFmpegPath = "C:\\ffmpeg.exe",
                LogDirectory = "C:\\logs"
            };

            string json = System.Text.Json.JsonSerializer.Serialize(settings);
            
            Assert.Contains("true", json);
            Assert.Contains("C:\\\\ffmpeg.exe", json);

            var deserialized = System.Text.Json.JsonSerializer.Deserialize<AppSettings>(json);
            
            Assert.True(deserialized.EnableDebugLogging);
            Assert.Equal("C:\\ffmpeg.exe", deserialized.FFmpegPath);
            Assert.Equal("C:\\logs", deserialized.LogDirectory);
        }
    
        [Fact] public void AppSettings_Load_FailsGracefullyOnCorruptedFile() { File.WriteAllText("config.json", "{invalid_json_!@#"); var res = AppSettings.Load(); Assert.NotNull(res); }
    }
}
