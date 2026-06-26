using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class SshToolsTests : IDisposable
    {
        public void Dispose()
        {
            SshTools.StopAll();
        }

        [Fact]
        public void SshConnect_InvalidHost_ReturnsError()
        {
            var res = SshTools.SshConnect("invalid.local.domain.test", 22, "user", "pass");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public void SshSend_NonExistentSession_ReturnsError()
        {
            var res = SshTools.SshSend("INVALID_ID_12345", "ls -la");
            Assert.True(res.IsError == true);
        }
    }
}
