using System;
using System.Collections.Concurrent;

namespace controller_mcp.Features.Tools
{
    public class AnalyticsData
    {
        public DateTime StartTime { get; set; } = DateTime.UtcNow;
        public long TotalRequestsProcessed;
        public long TotalErrors;
        public ConcurrentDictionary<string, long> ToolInvocations { get; set; } = new ConcurrentDictionary<string, long>();
        public long TotalBytesSent;
        public long TotalBytesReceived;
    }

    public static class AnalyticsManager
    {
        // This holds the runtime state. It gets injected from StateBackupManager on boot.
        public static AnalyticsData Current { get; private set; } = new AnalyticsData();

        public static void Restore(AnalyticsData data)
        {
            if (data != null)
            {
                Current = data;
                Current.ToolInvocations = Current.ToolInvocations ?? new ConcurrentDictionary<string, long>();
                
                // Preserve start time from original boot if it was restored
                if (Current.StartTime == default)
                {
                    Current.StartTime = DateTime.UtcNow;
                }
            }
        }

        public static void TrackRequest()
        {
            System.Threading.Interlocked.Increment(ref Current.TotalRequestsProcessed);
            StateBackupManager.SaveState();
        }

        public static void TrackError()
        {
            System.Threading.Interlocked.Increment(ref Current.TotalErrors);
            StateBackupManager.SaveState();
        }

        public static void TrackToolInvocation(string toolName)
        {
            if (string.IsNullOrWhiteSpace(toolName)) return;
            Current.ToolInvocations.AddOrUpdate(toolName, 1, (key, count) => count + 1);
            StateBackupManager.SaveState();
        }

        public static void TrackBytesSent(long bytes)
        {
            System.Threading.Interlocked.Add(ref Current.TotalBytesSent, bytes);
            StateBackupManager.SaveState(); // Note: Debouncer prevents thrashing
        }

        public static void TrackBytesReceived(long bytes)
        {
            System.Threading.Interlocked.Add(ref Current.TotalBytesReceived, bytes);
            StateBackupManager.SaveState();
        }
    }
}
