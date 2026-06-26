import { test, expect, describe } from "bun:test";
import { enrichChunkText } from "../src/enrichment";
import type { Chunk } from "../src/chunker";
import type { SessionMetadata, FileTouch } from "../src/parser";

function makeChunk(text: string, isCompactSummary = false): Chunk {
  return {
    msgStart: isCompactSummary ? -1 : 0,
    msgEnd: isCompactSummary ? -1 : 1,
    text,
    isCompactSummary,
  };
}

const fullMetadata: SessionMetadata = {
  sessionId: "test-session",
  branch: "fix-auth",
  title: "Auth Flash Bug Fix",
  projectPath: "/Users/cavin/code/visk",
  createdAt: "2026-04-15T10:00:00.000Z",
  updatedAt: "2026-04-15T11:00:00.000Z",
  messageCount: 12,
};

const fullFiles: FileTouch[] = [
  { path: "/src/auth/session.ts", operation: "edit" },
  { path: "/src/components/Login.tsx", operation: "write" },
  { path: "/src/utils/helpers.ts", operation: "read" },
];

describe("enrichChunkText", () => {
  test("prepends full metadata prefix to conversation chunk", () => {
    const chunk = makeChunk("USER: fix the auth bug\nCLAUDE: I found the issue");
    const result = enrichChunkText(chunk, fullMetadata, fullFiles);

    expect(result).toStartWith("[");
    expect(result).toContain("Project: visk");
    expect(result).toContain("Branch: fix-auth");
    expect(result).toContain("Apr 2026");
    expect(result).toContain("Files: session.ts, Login.tsx, helpers.ts");
    expect(result).toContain("USER: fix the auth bug");
    // Conversation chunks should NOT include title
    expect(result).not.toContain("Topic:");
  });

  test("includes title only on compact summary chunks", () => {
    const chunk = makeChunk("Task overview: Fixed the auth flash bug", true);
    const result = enrichChunkText(chunk, fullMetadata, fullFiles);

    expect(result).toContain("Topic: Auth Flash Bug Fix");
  });

  test("skips main branch", () => {
    const metadata = { ...fullMetadata, branch: "main" };
    const result = enrichChunkText(makeChunk("hello"), metadata, []);

    expect(result).not.toContain("Branch:");
    expect(result).not.toContain("main");
  });

  test("skips master branch", () => {
    const metadata = { ...fullMetadata, branch: "master" };
    const result = enrichChunkText(makeChunk("hello"), metadata, []);

    expect(result).not.toContain("Branch:");
    expect(result).not.toContain("master");
  });

  test("handles missing fields gracefully", () => {
    const metadata: SessionMetadata = {
      sessionId: "test",
      branch: null,
      title: null,
      projectPath: null,
      createdAt: null,
      updatedAt: null,
      messageCount: 0,
    };
    const chunk = makeChunk("USER: hello");
    const result = enrichChunkText(chunk, metadata, []);

    // No metadata available — returns raw text unchanged
    expect(result).toBe("USER: hello");
  });

  test("prioritizes edits/writes over reads in file list", () => {
    const files: FileTouch[] = [
      { path: "/a/read-first.ts", operation: "read" },
      { path: "/b/edited.ts", operation: "edit" },
      { path: "/c/written.ts", operation: "write" },
      { path: "/d/also-read.ts", operation: "read" },
    ];
    const result = enrichChunkText(makeChunk("hello"), fullMetadata, files);

    // edits/writes should appear before reads
    const editIdx = result.indexOf("edited.ts");
    const writeIdx = result.indexOf("written.ts");
    const readIdx = result.indexOf("read-first.ts");
    expect(editIdx).toBeLessThan(readIdx);
    expect(writeIdx).toBeLessThan(readIdx);
  });

  test("deduplicates file basenames", () => {
    const files: FileTouch[] = [
      { path: "/src/auth/session.ts", operation: "read" },
      { path: "/src/auth/session.ts", operation: "edit" },
      { path: "/src/other/session.ts", operation: "read" },
    ];
    const result = enrichChunkText(makeChunk("hello"), fullMetadata, files);

    // "session.ts" should appear only once
    const matches = result.match(/session\.ts/g);
    expect(matches?.length).toBe(1);
  });

  test("limits to 5 files", () => {
    const files: FileTouch[] = Array.from({ length: 10 }, (_, i) => ({
      path: `/src/file${i}.ts`,
      operation: "edit" as const,
    }));
    const result = enrichChunkText(makeChunk("hello"), fullMetadata, files);

    // Count comma-separated files in the Files: section
    const filesMatch = result.match(/Files: ([^\]]+)/);
    expect(filesMatch).toBeTruthy();
    const fileCount = filesMatch![1].split(", ").length;
    expect(fileCount).toBeLessThanOrEqual(5);
  });

  test("uses basename only, not full path", () => {
    const files: FileTouch[] = [
      { path: "/Users/cavin/code/visk/src/deeply/nested/Component.tsx", operation: "edit" },
    ];
    const result = enrichChunkText(makeChunk("hello"), fullMetadata, files);

    expect(result).toContain("Component.tsx");
    expect(result).not.toContain("/Users/cavin");
    expect(result).not.toContain("deeply/nested");
  });

  test("prefix is reasonably short (under 60 tokens)", () => {
    const chunk = makeChunk("USER: fix the bug");
    const result = enrichChunkText(chunk, fullMetadata, fullFiles);

    // Extract just the prefix (before the original text)
    const prefix = result.slice(0, result.indexOf(chunk.text));
    // Rough token estimate: 4 chars per token
    const estimatedTokens = prefix.length / 4;
    expect(estimatedTokens).toBeLessThan(60);
  });
});
