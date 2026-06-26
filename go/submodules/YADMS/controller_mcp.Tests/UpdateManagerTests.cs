using System;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class UpdateManagerTests
    {
        [Fact]
        public void ParseReleaseResponse_ValidRelease_ReturnsAssetUrl()
        {
            // Arrange
            string json = @"{
                ""tag_name"": ""v1.0.142"",
                ""assets"": [
                    { ""browser_download_url"": ""https://github.com/blazium-games/YADMS/releases/download/v1.0.142/YADMS_Installer_Lite_Standard_v1.0.142.exe"" }
                ]
            }";

            // Act
            var result = UpdateManager.ParseReleaseResponse(json, "1.0.141", "Lite_Standard");

            // Assert
            Assert.True(result.UpdateAvailable);
            Assert.Equal("1.0.142", result.LatestVersion);
            Assert.Equal("https://github.com/blazium-games/YADMS/releases/download/v1.0.142/YADMS_Installer_Lite_Standard_v1.0.142.exe", result.DownloadUrl);
        }

        [Fact]
        public void ParseReleaseResponse_OlderVersion_ReturnsFalse()
        {
            // Arrange
            string json = @"{
                ""tag_name"": ""v1.0.140"",
                ""assets"": [
                    { ""browser_download_url"": ""https://github.com/blazium-games/YADMS/releases/download/v1.0.140/YADMS_Installer_Lite_Standard_v1.0.140.exe"" }
                ]
            }";

            // Act
            var result = UpdateManager.ParseReleaseResponse(json, "1.0.141", "Lite_Standard");

            // Assert
            Assert.False(result.UpdateAvailable);
            Assert.Equal("1.0.140", result.LatestVersion);
            Assert.Null(result.DownloadUrl);
        }

        [Fact]
        public void ParseReleaseResponse_NoAssets_ReturnsGracefully()
        {
            // Arrange
            string json = @"{
                ""tag_name"": ""v1.0.145"",
                ""assets"": []
            }";

            // Act
            var result = UpdateManager.ParseReleaseResponse(json, "1.0.141", "Lite_Standard");

            // Assert
            Assert.True(result.UpdateAvailable); // Version is newer
            Assert.Equal("1.0.145", result.LatestVersion);
            Assert.Null(result.DownloadUrl); // But no asset found
        }

        [Fact]
        public void ParseReleaseResponse_MalformedJson_FailsSafely()
        {
            // Arrange
            string json = "Not JSON at all";

            // Act
            var result = UpdateManager.ParseReleaseResponse(json, "1.0.141", "Lite_Standard");

            // Assert
            Assert.False(result.UpdateAvailable);
            Assert.Null(result.LatestVersion);
            Assert.Null(result.DownloadUrl);
        }
    }
}
