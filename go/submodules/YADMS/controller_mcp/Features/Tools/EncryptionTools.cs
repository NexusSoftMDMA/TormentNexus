using System;
using System.ComponentModel;
using System.IO;
using System.Security.Cryptography;
using System.Text;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;
using System.Collections.Generic;

namespace controller_mcp.Features.Tools
{
    public static class EncryptionTools
    {
        [McpServerTool, Description("Generates a new random 256-bit AES Key. Returns it as a Base64 encoded string.")]
        public static CallToolResult GenerateAesKey()
        {
            try
            {
                using (var aes = new AesCryptoServiceProvider())
                {
                    aes.KeySize = 256;
                    aes.GenerateKey();
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = Convert.ToBase64String(aes.Key) } } };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Encrypts a plain-text string using AES-256. Requires a Base64 encoded 256-bit AES Key. Returns Base64 IV + CipherText.")]
        public static CallToolResult AesEncrypt(string plaintext, string base64Key)
        {
            try
            {
                byte[] keyBytes = Convert.FromBase64String(base64Key);
                using (var aes = new AesCryptoServiceProvider { Key = keyBytes, Mode = CipherMode.CBC, Padding = PaddingMode.PKCS7 })
                {
                    aes.GenerateIV();
                    var encryptor = aes.CreateEncryptor();
                    byte[] plainBytes = Encoding.UTF8.GetBytes(plaintext);
                    byte[] cipherBytes = encryptor.TransformFinalBlock(plainBytes, 0, plainBytes.Length);
                    
                    // Combine IV + Cipher
                    byte[] resultBytes = new byte[aes.IV.Length + cipherBytes.Length];
                    Buffer.BlockCopy(aes.IV, 0, resultBytes, 0, aes.IV.Length);
                    Buffer.BlockCopy(cipherBytes, 0, resultBytes, aes.IV.Length, cipherBytes.Length);
                    
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = Convert.ToBase64String(resultBytes) } } };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Decrypts AES-256 cipher text. Requires the Base64 encoded payload (IV + Cipher) and the Base64 encoded 256-bit AES Key.")]
        public static CallToolResult AesDecrypt(string base64Payload, string base64Key)
        {
            try
            {
                byte[] payload = Convert.FromBase64String(base64Payload);
                byte[] keyBytes = Convert.FromBase64String(base64Key);

                using (var aes = new AesCryptoServiceProvider { Key = keyBytes, Mode = CipherMode.CBC, Padding = PaddingMode.PKCS7 })
                {
                    byte[] iv = new byte[16];
                    byte[] cipher = new byte[payload.Length - 16];
                    Buffer.BlockCopy(payload, 0, iv, 0, 16);
                    Buffer.BlockCopy(payload, 16, cipher, 0, cipher.Length);

                    aes.IV = iv;
                    var decryptor = aes.CreateDecryptor();
                    byte[] plainBytes = decryptor.TransformFinalBlock(cipher, 0, cipher.Length);
                    
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = Encoding.UTF8.GetString(plainBytes) } } };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Generates an RSA-2048 Key Pair. Returns an XML string containing the Public and Private keys.")]
        public static CallToolResult GenerateRsaKeyPair()
        {
            try
            {
                using (var rsa = new RSACryptoServiceProvider(2048))
                {
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = rsa.ToXmlString(true) } } };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Encrypts a small string using an RSA Public Key (XML format). Returns Base64 cipher text.")]
        public static CallToolResult RsaEncrypt(string plaintext, string publicKeyXml)
        {
            try
            {
                using (var rsa = new RSACryptoServiceProvider())
                {
                    rsa.FromXmlString(publicKeyXml);
                    byte[] data = Encoding.UTF8.GetBytes(plaintext);
                    byte[] encrypted = rsa.Encrypt(data, false);
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = Convert.ToBase64String(encrypted) } } };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Decrypts a Base64 cipher text using an RSA Private Key (XML format).")]
        public static CallToolResult RsaDecrypt(string base64Cipher, string privateKeyXml)
        {
            try
            {
                using (var rsa = new RSACryptoServiceProvider())
                {
                    rsa.FromXmlString(privateKeyXml);
                    byte[] data = Convert.FromBase64String(base64Cipher);
                    byte[] decrypted = rsa.Decrypt(data, false);
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = Encoding.UTF8.GetString(decrypted) } } };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Hashes a string. Algorithms: MD5, SHA1, SHA256, SHA512. Returns Hexadecimal hash.")]
        public static CallToolResult HashString(string text, string algorithm = "SHA256")
        {
            try
            {
                using (HashAlgorithm alg = HashAlgorithm.Create(algorithm))
                {
                    if (alg == null) throw new ArgumentException($"Unsupported algorithm: {algorithm}");
                    byte[] hash = alg.ComputeHash(Encoding.UTF8.GetBytes(text));
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = BitConverter.ToString(hash).Replace("-", "").ToLowerInvariant() } } };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed: {ex.Message}" } } };
            }
        }
    }
}
