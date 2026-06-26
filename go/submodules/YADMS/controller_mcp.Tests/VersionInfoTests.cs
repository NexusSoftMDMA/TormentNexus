using System;
using System.Text.RegularExpressions;
using Xunit;
using controller_mcp;

namespace controller_mcp.Tests
{
    public class VersionInfoTests
    {
        [Fact]
        public void VersionInfo_CurrentVersion_IsCorrectFormat()
        {
            string version = VersionInfo.CurrentVersion;
            Assert.False(string.IsNullOrWhiteSpace(version));
            
            // Should match semantic versioning format (e.g. 1.0.141)
            var regex = new Regex(@"^\d+\.\d+\.\d+$");
            Assert.Matches(regex, version);
        }
    }
}
