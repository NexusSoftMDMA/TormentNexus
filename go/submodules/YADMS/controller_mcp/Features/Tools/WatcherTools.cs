using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Text.Json;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public class WatcherSession : IDisposable
    {
        public string Id { get; set; }
        public string Path { get; set; }
        public FileSystemWatcher Watcher { get; set; }
        public ConcurrentQueue<string> Events { get; set; } = new ConcurrentQueue<string>();

        public void Dispose()
        {
            try
            {
                if (Watcher != null)
                {
                    Watcher.EnableRaisingEvents = false;
                    Watcher.Dispose();
                }
            }
            catch { }
        }
    }

    public static class WatcherTools
    {
        private static readonly ConcurrentDictionary<string, WatcherSession> _sessions = new ConcurrentDictionary<string, WatcherSession>();

        public static void StopAll()
        {
            foreach (var kvp in _sessions)
            {
                try { kvp.Value.Dispose(); } catch { }
            }
            _sessions.Clear();
        }

        [McpServerTool, Description("Starts a background FileSystemWatcher on a specific directory. Returns a watcher_id to poll later.")]
        public static CallToolResult StartDirectoryWatcher(string directory_path, string filter = "*.*", bool include_subdirectories = true)
        {
            try
            {
                string safePath = InputValidator.ValidateFilePath(directory_path, nameof(directory_path));

                if (!Directory.Exists(safePath))
                    throw new DirectoryNotFoundException($"Directory not found: {safePath}");

                string id = Guid.NewGuid().ToString();
                
                var watcher = new FileSystemWatcher(safePath, filter)
                {
                    IncludeSubdirectories = include_subdirectories,
                    NotifyFilter = NotifyFilters.FileName | NotifyFilters.DirectoryName | NotifyFilters.LastWrite | NotifyFilters.Size
                };

                var session = new WatcherSession
                {
                    Id = id,
                    Path = safePath,
                    Watcher = watcher
                };

                // Attach event handlers
                watcher.Created += (s, e) => session.Events.Enqueue($"[Created] {e.FullPath.Replace("\\", "\\\\")}");
                watcher.Changed += (s, e) => session.Events.Enqueue($"[Changed] {e.FullPath.Replace("\\", "\\\\")}");
                watcher.Deleted += (s, e) => session.Events.Enqueue($"[Deleted] {e.FullPath.Replace("\\", "\\\\")}");
                watcher.Error += (s, e) => {
                    session.Events.Enqueue($"[Error] FileSystemWatcher Error: {e.GetException()?.Message}");
                    StopDirectoryWatcher(id);
                };

                // Start watching
                watcher.EnableRaisingEvents = true;

                _sessions.TryAdd(id, session);
                StateBackupManager.AddWatcher(safePath, filter, include_subdirectories);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { 
                        new TextContentBlock { Text = $"{{\"status\":\"started\", \"watcher_id\":\"{id}\", \"path\":\"{safePath.Replace("\\", "\\\\")}\"}}" } 
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to start watcher: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Retrieves all file system events captured since the last time this was called. Clears the queue by default.")]
        public static CallToolResult PollWatcherEvents(string watcher_id, bool clear_buffer = true)
        {
            if (_sessions.TryGetValue(watcher_id, out WatcherSession session))
            {
                List<string> retrievedEvents = new List<string>();

                if (clear_buffer)
                {
                    while (session.Events.TryDequeue(out string ev))
                    {
                        retrievedEvents.Add(ev);
                    }
                }
                else
                {
                    retrievedEvents.AddRange(session.Events);
                }

                string json = JsonSerializer.Serialize(new 
                {
                    status = session.Watcher.EnableRaisingEvents ? "watching" : "disabled",
                    event_count = retrievedEvents.Count,
                    events = retrievedEvents
                });

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
                };
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active watcher found with ID '{watcher_id}'." } } };
        }

        [McpServerTool, Description("Gracefully shuts down a FileSystemWatcher and frees resources.")]
        public static CallToolResult StopDirectoryWatcher(string watcher_id)
        {
            if (_sessions.TryRemove(watcher_id, out WatcherSession session))
            {
                StateBackupManager.RemoveWatcher(session.Path);
                session.Dispose();

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = $"Watcher {watcher_id} stopped successfully." } }
                };
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active watcher found with ID '{watcher_id}'." } } };
        }
    }
}
