using System;
using System.Collections.Generic;
using System.ComponentModel;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;
using System.Text.Json;

namespace controller_mcp.Features.Tools
{
    public static class SystemTools
    {
        [McpServerTool, Description("Generates and returns a random RFC 4122 version 4 UUID.")]
        public static CallToolResult GenerateUuid()
        {
            string uuid = Guid.NewGuid().ToString();
            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = uuid } }
            };
        }

        [McpServerTool, Description("Gets system information including Config Location, Log Directory, FFmpeg Directory, Build Type, and SemVer version.")]
        public static CallToolResult GetSystemInfo()
        {
            var settings = AppSettings.Load();
            
#if GAME_HACKING
            bool gameHacking = true;
#else
            bool gameHacking = false;
#endif

#if BUILD_WITH_FFMPEG
            bool buildWithFfmpeg = true;
            string ffmpegPath = "Embedded (Build Flag)";
#else
            bool buildWithFfmpeg = false;
            string ffmpegPath = settings.FFmpegPath;
#endif

            string version = "1.0.0";
            try { version = System.IO.File.ReadAllText("version.txt").Trim(); } catch { }

            var sysInfo = new
            {
                Version = version,
                ConfigLocation = AppSettings.ConfigPath,
                LogDirectory = settings.LogDirectory,
                FFmpegPath = ffmpegPath,
                BuildFlags = new
                {
                    GAME_HACKING = gameHacking,
                    BUILD_WITH_FFMPEG = buildWithFfmpeg
                }
            };

            string json = JsonSerializer.Serialize(sysInfo, new JsonSerializerOptions { WriteIndented = true });
            return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = json } } };
        }

        [McpServerTool, Description("Gets the current system timestamp in multiple formats (ISO 8601, Unix epoch, Human Readable).")]
        public static CallToolResult GetTimestamps()
        {
            DateTime now = DateTime.UtcNow;
            
            var timestamps = new
            {
                iso_8601 = now.ToString("yyyy-MM-ddTHH:mm:ssZ"),
                unix_epoch = new DateTimeOffset(now).ToUnixTimeSeconds(),
                unix_epoch_ms = new DateTimeOffset(now).ToUnixTimeMilliseconds(),
                human_readable = now.ToString("dddd, MMMM dd, yyyy h:mm:ss tt UTC")
            };

            string json = JsonSerializer.Serialize(timestamps, new JsonSerializerOptions { WriteIndented = true });

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
            };
        }
    }
}
