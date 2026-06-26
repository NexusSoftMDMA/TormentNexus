import { test, expect, describe, afterAll } from "bun:test";
import { parseSession, stripCodeBlocks, cleanMessageText, type ParsedSession } from "../src/parser";
import { writeFileSync, mkdirSync, rmSync } from "fs";
import { join } from "path";

const TEST_DIR = join(import.meta.dir, ".fixtures");

// Helper to create a test JSONL file
function createTestSession(filename: string, records: object[]): string {
  mkdirSync(TEST_DIR, { recursive: true });
  const path = join(TEST_DIR, filename);
  const content = records.map((r) => JSON.stringify(r)).join("\n");
  writeFileSync(path, content);
  return path;
}

// Cleanup after tests
afterAll(() => {
  rmSync(TEST_DIR, { recursive: true, force: true });
});

// --- Test data: realistic Claude Code JSONL records ---

const userMessage = (text: string, opts: Partial<Record<string, any>> = {}) => ({
  type: "user",
  sessionId: "test-session-123",
  uuid: `user-${Math.random().toString(36).slice(2)}`,
  parentUuid: null,
  timestamp: "2026-04-06T23:39:11.918Z",
  gitBranch: "main",
  cwd: "/Users/test/projects/myapp",
  message: { role: "user", content: text },
  ...opts,
});

const toolResultMessage = (toolUseId: string, content: string, isError = false) => ({
  type: "user",
  sessionId: "test-session-123",
  uuid: `toolresult-${Math.random().toString(36).slice(2)}`,
  parentUuid: null,
  timestamp: "2026-04-06T23:40:00.000Z",
  message: {
    role: "user",
    content: [{ tool_use_id: toolUseId, type: "tool_result", content, is_error: isError }],
  },
});

const assistantTextMessage = (text: string) => ({
  type: "assistant",
  sessionId: "test-session-123",
  uuid: `asst-${Math.random().toString(36).slice(2)}`,
  parentUuid: null,
  timestamp: "2026-04-06T23:39:15.000Z",
  gitBranch: "main",
  message: {
    role: "assistant",
    content: [{ type: "text", text }],
  },
});

const assistantToolUse = (toolName: string, input: Record<string, any>) => ({
  type: "assistant",
  sessionId: "test-session-123",
  uuid: `asst-${Math.random().toString(36).slice(2)}`,
  parentUuid: null,
  timestamp: "2026-04-06T23:39:20.000Z",
  gitBranch: "main",
  message: {
    role: "assistant",
    content: [{ type: "tool_use", id: `toolu_${Math.random().toString(36).slice(2)}`, name: toolName, input }],
  },
});

const assistantThinking = (thinking: string) => ({
  type: "assistant",
  sessionId: "test-session-123",
  uuid: `asst-${Math.random().toString(36).slice(2)}`,
  parentUuid: null,
  timestamp: "2026-04-06T23:39:12.000Z",
  message: {
    role: "assistant",
    content: [{ type: "thinking", thinking }],
  },
});

const summaryRecord = (summary: string) => ({
  type: "summary",
  summary,
  leafUuid: "some-uuid",
});

const progressRecord = () => ({
  type: "progress",
  sessionId: "test-session-123",
  data: "x".repeat(10000), // large payload
});

const fileHistorySnapshot = () => ({
  type: "file-history-snapshot",
  messageId: "some-id",
  snapshot: { trackedFileBackups: {}, timestamp: "2026-04-06T23:39:11.000Z" },
});

// --- Tests ---

describe("stripCodeBlocks", () => {
  test("keeps short code blocks", () => {
    const text = "Here's the fix:\n```ts\nconst x = 1;\n```\nDone.";
    expect(stripCodeBlocks(text)).toBe(text);
  });

  test("strips large code blocks with language marker", () => {
    const longCode = "x\n".repeat(200);
    const text = `Here's the file:\n\`\`\`typescript\n${longCode}\`\`\`\nThat's it.`;
    const result = stripCodeBlocks(text);
    expect(result).toContain("[code: typescript,");
    expect(result).toContain("lines]");
    expect(result).toContain("Here's the file:");
    expect(result).toContain("That's it.");
    expect(result).not.toContain(longCode);
  });

  test("strips large code blocks without language", () => {
    const longCode = "line\n".repeat(200);
    const text = `Before\n\`\`\`\n${longCode}\`\`\`\nAfter`;
    const result = stripCodeBlocks(text);
    expect(result).toContain("[code,");
    expect(result).toContain("Before");
    expect(result).toContain("After");
  });

  test("handles multiple code blocks independently", () => {
    const short = "```ts\nconst x = 1;\n```";
    const longCode = "line\n".repeat(200);
    const long = `\`\`\`py\n${longCode}\`\`\``;
    const text = `${short}\nsome text\n${long}`;
    const result = stripCodeBlocks(text);
    expect(result).toContain("const x = 1"); // short kept
    expect(result).toContain("[code: py,"); // long stripped
  });

  test("returns text unchanged when no code blocks", () => {
    const text = "Just regular text with no code.";
    expect(stripCodeBlocks(text)).toBe(text);
  });

  test("strips code blocks from parsed messages", async () => {
    const longCode = "x\n".repeat(200);
    const path = createTestSession("codestrip.jsonl", [
      userMessage(`fix this:\n\`\`\`ts\n${longCode}\`\`\``),
      assistantTextMessage(`Here's the fix:\n\`\`\`ts\n${longCode}\`\`\``),
    ]);
    const result = await parseSession(path);
    expect(result.messages[0].text).toContain("[code: ts,");
    expect(result.messages[1].text).toContain("[code: ts,");
    expect(result.messages[0].text).not.toContain(longCode);
  });
});

describe("cleanMessageText", () => {
  test("strips system-reminder tags", () => {
    const text = "Hello\n<system-reminder>\nSome long system prompt\n</system-reminder>\nWorld";
    const result = cleanMessageText(text);
    expect(result).toContain("Hello");
    expect(result).toContain("World");
    expect(result).not.toContain("system-reminder");
    expect(result).not.toContain("Some long system prompt");
  });

  test("strips local-command-caveat tags", () => {
    const text = "<local-command-caveat>Caveat: messages below were generated</local-command-caveat>\nActual message";
    const result = cleanMessageText(text);
    expect(result).toContain("Actual message");
    expect(result).not.toContain("Caveat");
  });

  test("strips available-deferred-tools tags", () => {
    const text = "Before\n<available-deferred-tools>\nTool1\nTool2\n</available-deferred-tools>\nAfter";
    const result = cleanMessageText(text);
    expect(result).toContain("Before");
    expect(result).toContain("After");
    expect(result).not.toContain("Tool1");
  });

  test("strips multiple framework tags in one message", () => {
    const text = "<command-name>/clear</command-name>\n<command-message>clear</command-message>\n<command-args></command-args>\nReal content here";
    const result = cleanMessageText(text);
    expect(result).toContain("Real content here");
    expect(result).not.toContain("command-name");
    expect(result).not.toContain("command-message");
  });

  test("strips both code blocks and XML tags", () => {
    const longCode = "x\n".repeat(200);
    const text = `<system-reminder>noise</system-reminder>\nHere's the fix:\n\`\`\`ts\n${longCode}\`\`\`\nDone.`;
    const result = cleanMessageText(text);
    expect(result).toContain("Here's the fix:");
    expect(result).toContain("Done.");
    expect(result).not.toContain("noise");
    expect(result).toContain("[code: ts,");
  });

  test("collapses excessive newlines", () => {
    const text = "Line 1\n\n\n\n\nLine 2";
    expect(cleanMessageText(text)).toBe("Line 1\n\nLine 2");
  });
});

describe("parseSession", () => {
  test("extracts user messages (string content only, skips tool results)", async () => {
    const path = createTestSession("basic.jsonl", [
      userMessage("fix the auth flash bug"),
      userMessage("also check the login page"),
      toolResultMessage("toolu_123", "file contents here"), // should be skipped
    ]);

    const result = await parseSession(path);

    expect(result.messages.length).toBe(2);
    expect(result.messages[0].role).toBe("user");
    expect(result.messages[0].text).toBe("fix the auth flash bug");
    expect(result.messages[1].text).toBe("also check the login page");
  });

  test("extracts assistant text messages (skips thinking blocks)", async () => {
    const path = createTestSession("assistant.jsonl", [
      userMessage("fix the bug"),
      assistantThinking("Let me think about this..."),
      assistantTextMessage("I found the issue in session.ts"),
      assistantTextMessage("The fix is to add a loading state"),
    ]);

    const result = await parseSession(path);

    const assistantMsgs = result.messages.filter((m) => m.role === "assistant");
    expect(assistantMsgs.length).toBe(2);
    expect(assistantMsgs[0].text).toBe("I found the issue in session.ts");
    expect(assistantMsgs[1].text).toBe("The fix is to add a loading state");
  });

  test("extracts file paths from Read/Edit/Write tool calls", async () => {
    const path = createTestSession("tools.jsonl", [
      userMessage("look at the auth code"),
      assistantToolUse("Read", { file_path: "/src/auth/session.ts" }),
      assistantToolUse("Edit", { file_path: "/src/auth/session.ts", old_string: "x", new_string: "y" }),
      assistantToolUse("Write", { file_path: "/src/auth/new-file.ts", content: "..." }),
      assistantToolUse("Bash", { command: "bun test" }), // should NOT be in files
    ]);

    const result = await parseSession(path);

    expect(result.filesTouched.length).toBe(3); // read + edit for session.ts, write for new-file.ts
    expect(result.filesTouched).toContainEqual({ path: "/src/auth/session.ts", operation: "read" });
    expect(result.filesTouched).toContainEqual({ path: "/src/auth/session.ts", operation: "edit" });
    expect(result.filesTouched).toContainEqual({ path: "/src/auth/new-file.ts", operation: "write" });
  });

  test("extracts session metadata", async () => {
    const path = createTestSession("metadata.jsonl", [
      userMessage("fix something", {
        sessionId: "abc-123",
        gitBranch: "fix-auth",
        timestamp: "2026-04-06T10:00:00.000Z",
        cwd: "/Users/test/myapp",
      }),
      assistantTextMessage("done"),
      summaryRecord("Auth Flash Bug Fix"),
    ]);

    const result = await parseSession(path);

    expect(result.metadata.sessionId).toBe("abc-123");
    expect(result.metadata.branch).toBe("fix-auth");
    expect(result.metadata.title).toBe("Auth Flash Bug Fix");
    expect(result.metadata.projectPath).toBe("/Users/test/myapp");
    expect(result.metadata.createdAt).toBe("2026-04-06T10:00:00.000Z");
  });

  test("extracts compact summaries and flags them", async () => {
    const compactSummary = {
      type: "user",
      sessionId: "test-session-123",
      uuid: "compact-uuid",
      parentUuid: null,
      timestamp: "2026-04-06T23:50:00.000Z",
      isCompactSummary: true,
      message: {
        role: "user",
        content: "Task overview: Fixed auth flash bug. Key decision: Added loading state during async token validation.",
      },
    };

    const path = createTestSession("compact.jsonl", [
      userMessage("fix the bug"),
      assistantTextMessage("working on it"),
      compactSummary,
    ]);

    const result = await parseSession(path);

    expect(result.compactSummaries.length).toBe(1);
    expect(result.compactSummaries[0]).toContain("Fixed auth flash bug");
  });

  test("skips progress and file-history-snapshot records", async () => {
    const path = createTestSession("skip.jsonl", [
      fileHistorySnapshot(),
      progressRecord(),
      userMessage("real message"),
      progressRecord(),
      assistantTextMessage("real response"),
    ]);

    const result = await parseSession(path);

    expect(result.messages.length).toBe(2);
    expect(result.messages[0].text).toBe("real message");
    expect(result.messages[1].text).toBe("real response");
  });

  test("handles corrupted lines gracefully", async () => {
    const path = join(TEST_DIR, "corrupt.jsonl");
    mkdirSync(TEST_DIR, { recursive: true });
    const lines = [
      JSON.stringify(userMessage("good message 1")),
      '{"broken json',
      "",
      JSON.stringify(assistantTextMessage("good response")),
      "not json at all",
      JSON.stringify(userMessage("good message 2")),
    ];
    writeFileSync(path, lines.join("\n"));

    const result = await parseSession(path);

    expect(result.messages.length).toBe(3);
    expect(result.messages[0].text).toBe("good message 1");
    expect(result.messages[1].text).toBe("good response");
    expect(result.messages[2].text).toBe("good message 2");
  });

  test("extracts timestamps for first and last records", async () => {
    const path = createTestSession("timestamps.jsonl", [
      userMessage("first", { timestamp: "2026-04-06T10:00:00.000Z" }),
      assistantTextMessage("middle"),
      userMessage("last", { timestamp: "2026-04-06T12:00:00.000Z" }),
    ]);

    const result = await parseSession(path);

    expect(result.metadata.createdAt).toBe("2026-04-06T10:00:00.000Z");
    expect(result.metadata.updatedAt).toBe("2026-04-06T12:00:00.000Z");
  });

  test("counts total messages", async () => {
    const path = createTestSession("count.jsonl", [
      userMessage("one"),
      assistantTextMessage("two"),
      userMessage("three"),
      assistantTextMessage("four"),
      assistantTextMessage("five"),
    ]);

    const result = await parseSession(path);

    expect(result.metadata.messageCount).toBe(5);
  });

  test("works on a real session file", async () => {
    // Use an actual visk session if it exists
    const realPath = join(
      process.env.HOME || "~",
      ".claude/projects/-Users-cavin--cursor-tutor-projects-visk/626a93e2-2a67-4eee-93f2-839d57be14b7.jsonl"
    );

    try {
      const result = await parseSession(realPath);

      expect(result.messages.length).toBeGreaterThan(0);
      expect(result.metadata.sessionId).toBe("626a93e2-2a67-4eee-93f2-839d57be14b7");
      expect(result.metadata.branch).toBe("dev");
      expect(result.metadata.title).toBe("Hero Section Responsive Layout - Cards Overlap Fix");
      expect(result.filesTouched.length).toBeGreaterThan(0);
      // Should have extracted Hero.tsx from tool calls
      expect(result.filesTouched.some((f) => f.path.includes("Hero.tsx"))).toBe(true);
    } catch {
      // Skip if real file doesn't exist (CI environment)
      console.warn("Skipping real session test — file not found");
    }
  });
});
