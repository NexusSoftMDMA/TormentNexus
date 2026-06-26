import type { ParsedMessage } from "./parser";
import { cleanMessageText } from "./parser";

export interface Chunk {
  msgStart: number;
  msgEnd: number;
  text: string;
  isCompactSummary: boolean;
}

// Qwen3-Embedding has 32K token context via Ollama. At ~4 chars/token for
// natural language, 8000 chars ≈ 2000 tokens — well within limits and ~2x
// larger than before for more coherent chunks.
const TARGET_CHARS = 8000;

function formatMessage(msg: ParsedMessage): string {
  const label = msg.role === "user" ? "USER" : "CLAUDE";
  return `${label}: ${msg.text}`;
}

/**
 * Split text that exceeds TARGET_CHARS at newline boundaries.
 */
function splitOversized(text: string): string[] {
  if (text.length <= TARGET_CHARS) return [text];

  const parts: string[] = [];
  let remaining = text;
  while (remaining.length > TARGET_CHARS) {
    let splitAt = remaining.lastIndexOf("\n", TARGET_CHARS);
    if (splitAt <= 0) {
      parts.push(remaining.substring(0, TARGET_CHARS));
      remaining = remaining.substring(TARGET_CHARS);
    } else {
      parts.push(remaining.substring(0, splitAt));
      remaining = remaining.substring(splitAt + 1); // skip newline
    }
  }
  if (remaining) parts.push(remaining);
  return parts;
}

/**
 * Split a parsed session into chunks.
 * - Compact summaries get their own dedicated chunks
 * - Message chunks target ~1500 tokens
 * - Never splits a user message from its immediate assistant response
 */
export function chunkSession(
  messages: ParsedMessage[],
  compactSummaries: string[]
): Chunk[] {
  const chunks: Chunk[] = [];

  // Compact summaries each get their own prioritized chunk
  for (const summary of compactSummaries) {
    const stripped = cleanMessageText(summary);
    for (const part of splitOversized(stripped)) {
      chunks.push({
        msgStart: -1,
        msgEnd: -1,
        text: part,
        isCompactSummary: true,
      });
    }
  }

  if (messages.length === 0) return chunks;

  // Helper: flush accumulated text as one or more chunks, splitting if oversized
  function flushChunks(texts: string[], start: number, end: number) {
    const joined = texts.join("\n");
    for (const part of splitOversized(joined)) {
      chunks.push({ msgStart: start, msgEnd: end, text: part, isCompactSummary: false });
    }
  }

  // Split messages into chunks, respecting user+assistant pairs
  let currentTexts: string[] = [];
  let currentChars = 0;
  let chunkStart = messages[0].index;

  for (let i = 0; i < messages.length; i++) {
    const msg = messages[i];
    const formatted = formatMessage(msg);
    const msgChars = formatted.length;

    // Would adding this message exceed the target?
    if (currentChars + msgChars > TARGET_CHARS && currentTexts.length > 0) {
      // Check: is this an assistant message following a user message?
      // If so, keep them together (don't split the pair)
      const prevMsg = i > 0 ? messages[i - 1] : null;
      const isResponseToPrev = msg.role === "assistant" && prevMsg?.role === "user";

      if (!isResponseToPrev) {
        flushChunks(currentTexts, chunkStart, messages[i - 1].index);
        currentTexts = [];
        currentChars = 0;
        chunkStart = msg.index;
      }
    }

    currentTexts.push(formatted);
    currentChars += msgChars;
  }

  // Flush remaining messages
  if (currentTexts.length > 0) {
    flushChunks(currentTexts, chunkStart, messages[messages.length - 1].index);
  }

  return chunks;
}
