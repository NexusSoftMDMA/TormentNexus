// Permissive chatId validator — the runner uses ULIDs (26-char Base32)
// but we keep older fixtures + the MCP create_chat surface in mind, so
// allow any short alphanumeric/dash string. Belt-and-braces against
// unbounded user input becoming a filesystem path or log file name (DoS
// via 100 MB id).
const CHAT_ID_RE = /^[A-Za-z0-9_-]{1,64}$/;

export function isValidChatId(value: unknown): value is string {
  return typeof value === 'string' && CHAT_ID_RE.test(value);
}
