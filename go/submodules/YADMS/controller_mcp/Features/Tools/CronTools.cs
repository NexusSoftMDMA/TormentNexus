using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Threading;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public class ScheduledTask : IDisposable
    {
        public string Id { get; set; }
        public string ActionType { get; set; }
        public string Target { get; set; }
        public string Payload { get; set; }
        public int IntervalSeconds { get; set; }
        public System.Threading.Timer Timer { get; set; }

        public void Dispose()
        {
            Timer?.Dispose();
        }
    }

    public static class CronTools
    {
        private static readonly ConcurrentDictionary<string, ScheduledTask> _tasks = new ConcurrentDictionary<string, ScheduledTask>();

        public static void StopAll()
        {
            foreach (var kvp in _tasks)
            {
                try { kvp.Value.Dispose(); } catch { }
            }
            _tasks.Clear();
        }

        private static async void TimerCallback(object state)
        {
            if (!(state is ScheduledTask task)) return;

            string resultLog = $"[CronTask {task.Id}] Executed {task.ActionType} -> {task.Target}. ";
            
            try
            {
                if (task.ActionType.Equals("command", StringComparison.OrdinalIgnoreCase))
                {
                    ProcessStartInfo psi = new ProcessStartInfo
                    {
                        FileName = "cmd.exe",
                        Arguments = $"/c {task.Target}",
                        UseShellExecute = false,
                        CreateNoWindow = true,
                        RedirectStandardOutput = true,
                        RedirectStandardError = true
                    };
                    using (Process p = Process.Start(psi))
                    {
                        if (!p.WaitForExit(10000)) // 10s timeout
                        {
                            try { p.Kill(); } catch { }
                            resultLog += "Timed out and killed.";
                        }
                        else
                        {
                            resultLog += "Success.";
                        }
                    }
                }
                else if (task.ActionType.Equals("macro", StringComparison.OrdinalIgnoreCase))
                {
                    var root = System.Text.Json.JsonDocument.Parse(task.Target).RootElement;
                    InputTools.ExecuteMacro(root);
                    resultLog += "Success.";
                }
                else if (task.ActionType.Equals("websocket", StringComparison.OrdinalIgnoreCase))
                {
                    await WebSocketTools.WebsocketSend(task.Target, task.Payload);
                    resultLog += "Success.";
                }
                else if (task.ActionType.Equals("http", StringComparison.OrdinalIgnoreCase))
                {
                    await HttpTools.MakeHttpRequest("GET", task.Target, "{}", task.Payload);
                    resultLog += "Success.";
                }
            }
            catch (Exception ex)
            {
                resultLog += $"Failed: {ex.Message}";
            }

            // Log to debug console
            try
            {
                if (Form1.Instance != null)
                {
                    Form1.Instance.Log(resultLog);
                }
            }
            catch { }
        }

        [McpServerTool, Description("Spins up a background timer to silently execute a specific action on a recurring interval. Action types: 'command', 'macro', 'websocket', 'http'. Target is the command, json path, ws id, or url.")]
        public static CallToolResult ScheduleRecurringTask(string action_type, string target, int interval_seconds, string payload = "")
        {
            try
            {
                if (interval_seconds < 1)
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Interval must be at least 1 second." } } };

                string safeTarget = target;
                if (action_type.Equals("command", StringComparison.OrdinalIgnoreCase))
                {
                    safeTarget = InputValidator.SanitizeCommand(target);
                }
                else if (action_type.Equals("http", StringComparison.OrdinalIgnoreCase) || action_type.Equals("websocket", StringComparison.OrdinalIgnoreCase))
                {
                    safeTarget = InputValidator.ValidateUrl(target, nameof(target));
                }

                string id = Guid.NewGuid().ToString();

                var scheduledTask = new ScheduledTask
                {
                    Id = id,
                    ActionType = action_type,
                    Target = safeTarget,
                    Payload = payload,
                    IntervalSeconds = interval_seconds
                };

                scheduledTask.Timer = new System.Threading.Timer(TimerCallback, scheduledTask, TimeSpan.FromSeconds(interval_seconds), TimeSpan.FromSeconds(interval_seconds));

                _tasks.TryAdd(id, scheduledTask);
                StateBackupManager.AddCronTask(action_type, safeTarget, payload, interval_seconds);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { 
                        new TextContentBlock { Text = $"{{\"status\":\"scheduled\", \"task_id\":\"{id}\", \"action\":\"{action_type}\", \"interval\":{interval_seconds}}}" } 
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }

        [McpServerTool, Description("Lists all currently active scheduled background tasks.")]
        public static CallToolResult ListScheduledTasks()
        {
            try
            {
                List<object> list = new List<object>();
                foreach (var kvp in _tasks)
                {
                    list.Add(new
                    {
                        TaskId = kvp.Key,
                        ActionType = kvp.Value.ActionType,
                        Target = kvp.Value.Target,
                        IntervalSeconds = kvp.Value.IntervalSeconds
                    });
                }

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = System.Text.Json.JsonSerializer.Serialize(list) } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }

        [McpServerTool, Description("Cancels and removes a scheduled background task by ID.")]
        public static CallToolResult CancelScheduledTask(string task_id)
        {
            try
            {
                if (_tasks.TryRemove(task_id, out ScheduledTask task))
                {
                    StateBackupManager.RemoveCronTask(task.Target);
                    task.Dispose();
                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = $"Task {task_id} successfully cancelled." } }
                    };
                }

                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active task found with ID '{task_id}'." } } };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }
    }
}
