using System;
using System.Threading;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;

namespace controller_mcp.Tests
{
    public class CronToolsTests : IDisposable
    {
        public CronToolsTests()
        {
            CronTools.StopAll();
        }

        public void Dispose()
        {
            CronTools.StopAll();
        }

        [Fact]
        public void ScheduleTask_InvalidInterval_ReturnsError()
        {
            var res = CronTools.ScheduleRecurringTask("command", "echo test", 0);
            
            Assert.True(res.IsError == true);
            Assert.Contains("Interval must be at least 1 second", ((TextContentBlock)res.Content[0]).Text);
        }

        [Fact]
        public void ScheduleTask_AddsToInternalList()
        {
            var res = CronTools.ScheduleRecurringTask("command", "echo test", 5);
            Assert.True(res.IsError != true);

            var listRes = CronTools.ListScheduledTasks();
            string json = ((TextContentBlock)listRes.Content[0]).Text;

            Assert.Contains("echo test", json);
            Assert.Contains("command", json);
        }

        [Fact]
        public void CancelTask_RemovesFromInternalList()
        {
            var res = CronTools.ScheduleRecurringTask("command", "echo test", 5);
            string json = ((TextContentBlock)res.Content[0]).Text;
            
            // Extract task_id from json: {"status":"scheduled", "task_id":"...", "action":"command", "interval":5}
            var root = System.Text.Json.JsonDocument.Parse(json).RootElement;
            string taskId = root.GetProperty("task_id").GetString();

            var cancelRes = CronTools.CancelScheduledTask(taskId);
            Assert.True(cancelRes.IsError != true);

            var listRes = CronTools.ListScheduledTasks();
            string listJson = ((TextContentBlock)listRes.Content[0]).Text;
            Assert.DoesNotContain(taskId, listJson);
        }
        [Fact]
        public void ScheduleRecurringTask_HttpActionWithInvalidUrl_ReturnsError()
        {
            var result = CronTools.ScheduleRecurringTask("http", "ftp://bad-url", 5, "");
            Assert.True(result.IsError);
            Assert.Contains("invalid scheme", ((TextContentBlock)result.Content[0]).Text);
        }
    }
}
