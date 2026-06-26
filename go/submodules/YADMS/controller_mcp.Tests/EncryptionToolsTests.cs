using System;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using System.Text.Json;

namespace controller_mcp.Tests
{
    public class EncryptionToolsTests
    {
        [Fact]
        public void AesEncryptDecrypt_RoundTrip_Success()
        {
            // Generate Key
            var keyResult = EncryptionTools.GenerateAesKey();
            Assert.True(keyResult.IsError != true);
            string base64Key = ((ModelContextProtocol.Protocol.TextContentBlock)keyResult.Content[0]).Text;

            // Encrypt
            string originalText = "Top Secret 123!";
            var encResult = EncryptionTools.AesEncrypt(originalText, base64Key);
            Assert.True(encResult.IsError != true);
            string cipherText = ((ModelContextProtocol.Protocol.TextContentBlock)encResult.Content[0]).Text;

            // Decrypt
            var decResult = EncryptionTools.AesDecrypt(cipherText, base64Key);
            Assert.True(decResult.IsError != true);
            string plainText = ((ModelContextProtocol.Protocol.TextContentBlock)decResult.Content[0]).Text;

            Assert.Equal(originalText, plainText);
        }

        [Fact]
        public void HashString_ReturnsCorrectFormat()
        {
            var res = EncryptionTools.HashString("password", "SHA256");
            Assert.True(res.IsError != true);
            string hash = ((ModelContextProtocol.Protocol.TextContentBlock)res.Content[0]).Text;

            Assert.Equal(64, hash.Length); // SHA256 hex is 64 chars
        }
    
        [Fact] public void EncryptionTools_Decrypt_FailsGracefullyOnInvalidBase64() { var result = EncryptionTools.AesDecrypt("invalid", "invalid"); Assert.True(result.IsError == true); }
    }
}
