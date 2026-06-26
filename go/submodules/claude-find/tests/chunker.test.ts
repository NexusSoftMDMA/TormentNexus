import { test, expect, describe } from "bun:test";
import { chunkSession, type Chunk } from "../src/chunker";
import type { ParsedMessage } from "../src/parser";

// Helper to create messages
function msg(role: "user" | "assistant", text: string, index: number): ParsedMessage {
  return { role, text, timestamp: "2026-04-06T10:00:00.000Z", index };
}

// Helper to create messages with specific char count (4 chars ≈ 1 token)
function msgOfSize(role: "user" | "assistant", tokenCount: number, index: number): ParsedMessage {
  return { role, text: "x".repeat(tokenCount * 4), timestamp: "2026-04-06T10:00:00.000Z", index };
}

describe("chunkSession", () => {
  test("returns a single chunk for a small session", () => {
    const messages: ParsedMessage[] = [
      msg("user", "fix the auth bug", 0),
      msg("assistant", "I found the issue in session.ts", 1),
      msg("user", "great, ship it", 2),
    ];

    const chunks = chunkSession(messages, []);

    expect(chunks.length).toBe(1);
    expect(chunks[0].msgStart).toBe(0);
    expect(chunks[0].msgEnd).toBe(2);
    expect(chunks[0].text).toContain("fix the auth bug");
    expect(chunks[0].text).toContain("I found the issue");
    expect(chunks[0].isCompactSummary).toBe(false);
  });

  test("splits into multiple chunks when exceeding token limit", () => {
    // Create messages that together exceed ~1500 tokens
    const messages: ParsedMessage[] = [
      msgOfSize("user", 400, 0),
      msgOfSize("assistant", 400, 1),
      msgOfSize("user", 400, 2),
      msgOfSize("assistant", 400, 3),  // ~1600 tokens total at this point
      msgOfSize("user", 400, 4),
      msgOfSize("assistant", 400, 5),
    ];

    const chunks = chunkSession(messages, []);

    expect(chunks.length).toBeGreaterThan(1);
    // Every message should be in exactly one chunk
    const allIndices = chunks.flatMap((c) =>
      Array.from({ length: c.msgEnd - c.msgStart + 1 }, (_, i) => c.msgStart + i)
    );
    expect(allIndices).toContain(0);
    expect(allIndices).toContain(5);
  });

  test("never splits a user message from its immediate assistant response", () => {
    // Create a scenario where the split point falls between user and assistant
    const messages: ParsedMessage[] = [
      msgOfSize("user", 600, 0),
      msgOfSize("assistant", 600, 1),
      // ~1200 tokens — close to limit
      msgOfSize("user", 200, 2),  // this pushes to ~1400
      msgOfSize("assistant", 300, 3), // this pushes past 1500, but should stay with user msg 2
      msgOfSize("user", 400, 4),
      msgOfSize("assistant", 400, 5),
    ];

    const chunks = chunkSession(messages, []);

    // For each chunk, check that no user message at the end is missing its response
    for (const chunk of chunks) {
      // Get messages in this chunk
      const chunkMsgs = messages.filter((m) => m.index >= chunk.msgStart && m.index <= chunk.msgEnd);
      const lastMsg = chunkMsgs[chunkMsgs.length - 1];

      // If chunk ends with a user message, the next message should NOT be an assistant
      // (i.e., we didn't split a pair) — unless it's the very last message in the session
      if (lastMsg.role === "user" && lastMsg.index < messages.length - 1) {
        const nextMsg = messages[lastMsg.index + 1];
        // The next message shouldn't be an assistant response to this user message
        // Actually, if chunk ends on a user msg, that's OK only if there's no assistant after it
        // This is hard to test precisely, so we test the inverse:
        // assistant messages should never be the FIRST message of a chunk (unless it's the first chunk)
        if (chunk !== chunks[0]) {
          const firstMsgInChunk = messages.find((m) => m.index === chunk.msgStart);
          // It's acceptable for an assistant to start a chunk only if it's the overall first message
          // or if the previous message was also an assistant (consecutive assistant messages)
          if (firstMsgInChunk && firstMsgInChunk.role === "assistant" && chunk.msgStart > 0) {
            const prevMsg = messages[chunk.msgStart - 1];
            // Previous message should be an assistant too (not a user),
            // because we should keep user+assistant pairs together
            expect(prevMsg.role).not.toBe("user");
          }
        }
      }
    }
  });

  test("compact summaries get their own dedicated chunks", () => {
    const messages: ParsedMessage[] = [
      msg("user", "fix the bug", 0),
      msg("assistant", "working on it", 1),
    ];
    const compactSummaries = [
      "Task overview: Fixed auth flash bug. Key decision: Added loading state during async token validation.",
    ];

    const chunks = chunkSession(messages, compactSummaries);

    const summaryChunks = chunks.filter((c) => c.isCompactSummary);
    const regularChunks = chunks.filter((c) => !c.isCompactSummary);

    expect(summaryChunks.length).toBe(1);
    expect(summaryChunks[0].text).toContain("Fixed auth flash bug");
    expect(summaryChunks[0].isCompactSummary).toBe(true);
    expect(regularChunks.length).toBe(1);
  });

  test("multiple compact summaries each get their own chunk", () => {
    const messages: ParsedMessage[] = [
      msg("user", "msg", 0),
    ];
    const compactSummaries = [
      "First summary about auth work",
      "Second summary about payment refactor",
    ];

    const chunks = chunkSession(messages, compactSummaries);

    const summaryChunks = chunks.filter((c) => c.isCompactSummary);
    expect(summaryChunks.length).toBe(2);
    expect(summaryChunks[0].text).toContain("auth work");
    expect(summaryChunks[1].text).toContain("payment refactor");
  });

  test("handles empty message list", () => {
    const chunks = chunkSession([], []);
    expect(chunks.length).toBe(0);
  });

  test("handles empty messages with compact summaries", () => {
    const chunks = chunkSession([], ["Summary of the session"]);
    expect(chunks.length).toBe(1);
    expect(chunks[0].isCompactSummary).toBe(true);
  });

  test("chunk text includes role labels", () => {
    const messages: ParsedMessage[] = [
      msg("user", "what is this bug", 0),
      msg("assistant", "the issue is in auth.ts", 1),
    ];

    const chunks = chunkSession(messages, []);

    expect(chunks[0].text).toContain("USER:");
    expect(chunks[0].text).toContain("CLAUDE:");
  });

  test("chunks cover all messages from first to last", () => {
    const messages: ParsedMessage[] = Array.from({ length: 30 }, (_, i) =>
      msgOfSize(i % 2 === 0 ? "user" : "assistant", 200, i)
    );

    const chunks = chunkSession(messages, []);
    const regularChunks = chunks.filter((c) => !c.isCompactSummary);

    // Verify first and last message indices are covered
    expect(regularChunks[0].msgStart).toBe(0);
    expect(regularChunks[regularChunks.length - 1].msgEnd).toBe(29);

    // Verify multiple chunks were created (30 messages at 200 tokens each won't fit in one)
    expect(regularChunks.length).toBeGreaterThan(1);

    // Verify no chunk exceeds the target size by much
    for (const chunk of regularChunks) {
      expect(chunk.text.length).toBeLessThan(15000);
    }
  });

  test("splits oversized single message into multiple chunks", () => {
    // One message that's ~3x the target size
    const messages: ParsedMessage[] = [
      msg("user", "line\n".repeat(5000), 0),
      msg("assistant", "ok", 1),
    ];

    const chunks = chunkSession(messages, []);
    const regularChunks = chunks.filter((c) => !c.isCompactSummary);

    // Should be split into multiple chunks, not one giant one
    expect(regularChunks.length).toBeGreaterThan(1);
    // No chunk should exceed the target size by much
    for (const chunk of regularChunks) {
      expect(chunk.text.length).toBeLessThan(15000);
    }
  });

  test("splits oversized compact summary", () => {
    const hugeSummary = "summary line\n".repeat(2000);
    const chunks = chunkSession([], [hugeSummary]);

    expect(chunks.length).toBeGreaterThan(1);
    for (const chunk of chunks) {
      expect(chunk.text.length).toBeLessThan(15000);
      expect(chunk.isCompactSummary).toBe(true);
    }
  });
});
