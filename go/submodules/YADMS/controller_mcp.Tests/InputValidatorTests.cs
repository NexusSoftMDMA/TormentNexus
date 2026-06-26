using System;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class InputValidatorTests
    {
        [Theory]
        [InlineData("C:\\valid\\path", true)]
        [InlineData("/var/log/syslog", true)]
        [InlineData("C:\\invalid\\..\\path", false)]
        [InlineData("../../etc/passwd", false)]
        public void SanitizePath_PreventsTraversal(string path, bool isValid)
        {
            if (isValid)
            {
                Assert.Equal(path, InputValidator.ValidateFilePath(path, "path"));
            }
            else
            {
                Assert.Throws<ArgumentException>(() => InputValidator.ValidateFilePath(path, "path"));
            }
        }

        [Theory]
        [InlineData("https://github.com", true)]
        [InlineData("ftp://internal.server", false)] // Invalid scheme
        [InlineData("javascript:alert(1)", false)] // XSS
        [InlineData("http://localhost:8080", true)]
        public void SanitizeUrl_PreventsInjections(string url, bool isValid)
        {
            if (isValid)
            {
                Assert.Equal(url, InputValidator.ValidateUrl(url, "url"));
            }
            else
            {
                Assert.Throws<ArgumentException>(() => InputValidator.ValidateUrl(url, "url"));
            }
        }
    }
}
