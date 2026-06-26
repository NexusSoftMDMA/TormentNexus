using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class AnalyticsManagerTests : IDisposable
    {
        public AnalyticsManagerTests()
        {
            AnalyticsManager.Restore(new AnalyticsData());
        }

        public void Dispose()
        {
            AnalyticsManager.Restore(new AnalyticsData());
        }

        [Fact]
        public void Analytics_TracksConcurrentRequests_WithoutDataLoss()
        {
            int iterations = 10000;
            
            Parallel.For(0, iterations, i =>
            {
                AnalyticsManager.TrackRequest();
            });

            Assert.Equal(iterations, AnalyticsManager.Current.TotalRequestsProcessed);
        }

        [Fact]
        public void Analytics_TracksConcurrentBytes_WithoutDataLoss()
        {
            int iterations = 5000;
            long bytesPerRequest = 1024;

            Parallel.For(0, iterations, i =>
            {
                AnalyticsManager.TrackBytesSent(bytesPerRequest);
            });

            Assert.Equal(iterations * bytesPerRequest, AnalyticsManager.Current.TotalBytesSent);
        }
    
        [Fact] public void AnalyticsManager_LogToolInvocation_FailsGracefullyOnNull() { AnalyticsManager.TrackToolInvocation(null); Assert.True(AnalyticsManager.Current.TotalRequestsProcessed >= 0); }
    }
}
