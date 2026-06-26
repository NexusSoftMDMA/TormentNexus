import { test, expect, describe, afterAll, beforeAll } from "bun:test";
import { formatSearchResults } from "../src/server";
import type { SearchResult } from "../src/searcher";

describe("formatSearchResults", () => {
  test("formats results with session metadata and chunk text", () => {
    const results: SearchResult[] = [
      {
        sessionId: "abc-123",
        title: "Auth Flash Bug Fix",
        branch: "fix-auth",
        projectPath: "/Users/test/myapp",
        createdAt: "2026-04-20T10:00:00.000Z",
        updatedAt: "2026-04-20T12:00:00.000Z",
        messageCount: 47,
        isArchived: false,
        filesTouched: ["/src/auth/session.ts", "/src/components/Login.tsx"],
        chunks: [
          {
            text: "USER: fix the auth flash bug\nCLAUDE: I found the issue in session.ts",
            msgStart: 0,
            msgEnd: 5,
            isCompactSummary: false,
            score: 0.92,
          },
        ],
        score: 0.92,
      },
    ];

    const formatted = formatSearchResults(results);

    expect(formatted).toContain("Auth Flash Bug Fix");
    expect(formatted).toContain("fix-auth");
    expect(formatted).toContain("/Users/test/myapp");
    expect(formatted).toContain("session.ts");
    expect(formatted).toContain("fix the auth flash bug");
    expect(formatted).toContain("47 messages");
  });

  test("shows archived badge for archived sessions", () => {
    const results: SearchResult[] = [
      {
        sessionId: "old-123",
        title: "Old Session",
        branch: "main",
        projectPath: "/test",
        createdAt: "2026-02-01T10:00:00.000Z",
        updatedAt: "2026-02-01T12:00:00.000Z",
        messageCount: 10,
        isArchived: true,
        filesTouched: [],
        chunks: [{ text: "old content", msgStart: 0, msgEnd: 2, isCompactSummary: false, score: 0.5 }],
        score: 0.5,
      },
    ];

    const formatted = formatSearchResults(results);
    expect(formatted).toContain("archived");
  });

  test("labels compact summary chunks", () => {
    const results: SearchResult[] = [
      {
        sessionId: "compact-123",
        title: "Session with Summary",
        branch: "main",
        projectPath: "/test",
        createdAt: "2026-04-20T10:00:00.000Z",
        updatedAt: "2026-04-20T12:00:00.000Z",
        messageCount: 100,
        isArchived: false,
        filesTouched: [],
        chunks: [
          { text: "Task overview: Fixed the auth bug", msgStart: -1, msgEnd: -1, isCompactSummary: true, score: 0.9 },
          { text: "USER: regular conversation", msgStart: 0, msgEnd: 5, isCompactSummary: false, score: 0.7 },
        ],
        score: 0.9,
      },
    ];

    const formatted = formatSearchResults(results);
    expect(formatted).toContain("Compact Summary");
  });

  test("returns empty message for no results", () => {
    const formatted = formatSearchResults([]);
    expect(formatted).toContain("No matching sessions found");
  });

  test("formats multiple sessions", () => {
    const results: SearchResult[] = [
      {
        sessionId: "s1",
        title: "First Session",
        branch: "main",
        projectPath: "/app1",
        createdAt: "2026-04-20T10:00:00.000Z",
        updatedAt: "2026-04-20T10:00:00.000Z",
        messageCount: 10,
        isArchived: false,
        filesTouched: [],
        chunks: [{ text: "first content", msgStart: 0, msgEnd: 2, isCompactSummary: false, score: 0.9 }],
        score: 0.9,
      },
      {
        sessionId: "s2",
        title: "Second Session",
        branch: "dev",
        projectPath: "/app2",
        createdAt: "2026-04-19T10:00:00.000Z",
        updatedAt: "2026-04-19T10:00:00.000Z",
        messageCount: 5,
        isArchived: false,
        filesTouched: [],
        chunks: [{ text: "second content", msgStart: 0, msgEnd: 1, isCompactSummary: false, score: 0.7 }],
        score: 0.7,
      },
    ];

    const formatted = formatSearchResults(results);
    expect(formatted).toContain("First Session");
    expect(formatted).toContain("Second Session");
    expect(formatted).toContain("Found 2 relevant sessions");
  });
});
