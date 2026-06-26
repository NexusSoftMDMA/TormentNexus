using System;
using System.IO;

namespace controller_mcp.Features.Tools
{
    public static class InputValidator
    {
        public static string ValidateFilePath(string path, string paramName)
        {
            if (string.IsNullOrWhiteSpace(path))
                throw new ArgumentException($"Parameter '{paramName}' cannot be empty.");

            if (path.IndexOfAny(Path.GetInvalidPathChars()) >= 0)
                throw new ArgumentException($"Parameter '{paramName}' contains invalid path characters.");

            if (path.Contains(".."))
                throw new ArgumentException($"Parameter '{paramName}' cannot contain directory traversal sequences.");

            return path;
        }

        public static string EscapeArgument(string arg)
        {
            if (string.IsNullOrEmpty(arg))
                return "\"\"";

            // Prevents Argument Injection by replacing internal double-quotes 
            // and securely wrapping the final argument string.
            return "\"" + arg.Replace("\"", "\\\"") + "\"";
        }

        public static int ValidatePid(int pid, string paramName)
        {
            if (pid < 0)
                throw new ArgumentException($"Parameter '{paramName}' cannot be negative.");
            return pid;
        }

        public static string ValidateUrl(string url, string paramName)
        {
            if (string.IsNullOrWhiteSpace(url))
                throw new ArgumentException($"Parameter '{paramName}' cannot be empty.");

            if (!Uri.TryCreate(url, UriKind.Absolute, out Uri result))
                throw new ArgumentException($"Parameter '{paramName}' must be a valid absolute URL.");

            if (result.Scheme != Uri.UriSchemeHttp && 
                result.Scheme != Uri.UriSchemeHttps && 
                result.Scheme != "ws" && 
                result.Scheme != "wss")
            {
                throw new ArgumentException($"Parameter '{paramName}' has an invalid scheme (must be http, https, ws, or wss).");
            }

            return url;
        }

        public static string SanitizeCommand(string command)
        {
            if (string.IsNullOrWhiteSpace(command))
                throw new ArgumentException("Command cannot be empty.");
            
            return command.Trim();
        }
    }
}
