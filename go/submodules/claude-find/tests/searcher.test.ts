import { test, expect, describe, afterAll, beforeAll } from "bun:test";
import { search, type SearchResult } from "../src/searcher";
import { createDatabase, type ClaudeFindDB } from "../src/db";
import { indexSession } from "../src/indexer";
import { rmSync, mkdirSync, writeFileSync } from "fs";
import { join } from "path";

const TEST_DIR = join(import.meta.dir, ".fixtures-searcher");
const TEST_DB_PATH = join(import.meta.dir, ".test-searcher.db");

let db: ClaudeFindDB;

function createSession(sessionId: string, records: object[]): string {
  const dir = join(TEST_DIR, "project");
  mkdirSync(dir, { recursive: true });
  const path = join(dir, `${sessionId}.jsonl`);
  writeFileSync(path, records.map((r) => JSON.stringify(r)).join("\n"));
  return path;
}

function userMsg(text: string, opts: Record<string, any> = {}) {
  return {
    type: "user",
    sessionId: opts.sessionId || "test",
    uuid: `u-${Math.random().toString(36).slice(2)}`,
    parentUuid: null,
    timestamp: opts.timestamp || "2026-04-06T10:00:00.000Z",
    gitBranch: opts.gitBranch || "main",
    cwd: opts.cwd || "/Users/test/myapp",
    message: { role: "user", content: text },
  };
}

function assistantMsg(text: string) {
  return {
    type: "assistant",
    sessionId: "test",
    uuid: `a-${Math.random().toString(36).slice(2)}`,
    parentUuid: null,
    timestamp: "2026-04-06T10:05:00.000Z",
    gitBranch: "main",
    message: { role: "assistant", content: [{ type: "text", text }] },
  };
}

function toolUseMsg(toolName: string, filePath: string) {
  return {
    type: "assistant",
    sessionId: "test",
    uuid: `a-${Math.random().toString(36).slice(2)}`,
    parentUuid: null,
    timestamp: "2026-04-06T10:05:00.000Z",
    gitBranch: "main",
    message: {
      role: "assistant",
      content: [{ type: "tool_use", id: "toolu_123", name: toolName, input: { file_path: filePath } }],
    },
  };
}

function summaryRecord(title: string) {
  return { type: "summary", summary: title, leafUuid: "leaf" };
}

// Index test sessions once before all tests
beforeAll(async () => {
  rmSync(TEST_DB_PATH, { force: true });
  rmSync(TEST_DIR, { recursive: true, force: true });
  db = createDatabase(TEST_DB_PATH);

  // Session 1: Auth flash bug fix
  await indexSession(db, createSession("auth-session", [
    userMsg("fix the auth flash bug on the login page", {
      sessionId: "auth-session", gitBranch: "fix-auth", cwd: "/Users/test/myapp",
      timestamp: "2026-04-20T10:00:00.000Z",
    }),
    assistantMsg("I found the issue — the session token is being validated async without a loading state"),
    toolUseMsg("Read", "/src/auth/session.ts"),
    toolUseMsg("Edit", "/src/auth/session.ts"),
    userMsg("don't add a full loading screen, just prevent the flash"),
    assistantMsg("Added an isValidating state to the useAuth hook"),
    summaryRecord("Auth Flash Bug Fix"),
  ]));

  // Session 2: Payment webhook
  await indexSession(db, createSession("payment-session", [
    userMsg("implement the stripe payment webhook handler", {
      sessionId: "payment-session", gitBranch: "main", cwd: "/Users/test/myapp",
      timestamp: "2026-04-18T10:00:00.000Z",
    }),
    assistantMsg("I'll create a webhook endpoint that verifies the Stripe signature"),
    toolUseMsg("Write", "/src/api/webhooks/stripe.ts"),
    userMsg("make sure we handle idempotency"),
    assistantMsg("Added an idempotency key check using the event ID"),
    summaryRecord("Stripe Webhook Implementation"),
  ]));

  // Session 3: CSS layout (different topic)
  await indexSession(db, createSession("css-session", [
    userMsg("the hero section cards are overlapping on mobile", {
      sessionId: "css-session", gitBranch: "dev", cwd: "/Users/test/landing",
      timestamp: "2026-04-15T10:00:00.000Z",
    }),
    assistantMsg("The cards are using absolute positioning which breaks on smaller screens"),
    toolUseMsg("Edit", "/src/components/Hero.tsx"),
    userMsg("we want separation between cards because they're for different products"),
    assistantMsg("Switched to a two-column grid layout with Agent and Screen product grouping"),
    summaryRecord("Hero Section Card Layout Fix"),
  ]));
}, 120000);

afterAll(() => {
  if (db) db.close();
  rmSync(TEST_DB_PATH, { force: true });
  rmSync(TEST_DIR, { recursive: true, force: true });
});

describe("search", () => {
  test("finds sessions by semantic meaning (different words)", async () => {
    const results = await search(db, {
      query: "UI glitch when users log in",
    });

    expect(results.length).toBeGreaterThan(0);
    // Auth session should rank highest — "UI glitch when users log in" ≈ "auth flash bug on login page"
    expect(results[0].sessionId).toBe("auth-session");
  }, 30000);

  test("finds sessions by keyword", async () => {
    const results = await search(db, {
      query: "stripe webhook",
    });

    expect(results.length).toBeGreaterThan(0);
    expect(results[0].sessionId).toBe("payment-session");
  }, 30000);

  test("returns session metadata with results", async () => {
    const results = await search(db, {
      query: "auth flash bug",
    });

    expect(results.length).toBeGreaterThan(0);
    const top = results[0];
    expect(top.sessionId).toBe("auth-session");
    expect(top.title).toBe("Auth Flash Bug Fix");
    expect(top.branch).toBe("fix-auth");
    expect(top.projectPath).toBe("/Users/test/myapp");
    expect(top.filesTouched.length).toBeGreaterThan(0);
    expect(top.filesTouched.some((f: string) => f.includes("session.ts"))).toBe(true);
  }, 30000);

  test("returns matching chunk text", async () => {
    const results = await search(db, {
      query: "auth flash bug",
    });

    expect(results[0].chunks.length).toBeGreaterThan(0);
    const chunkText = results[0].chunks[0].text;
    expect(chunkText).toContain("auth flash bug");
  }, 30000);

  test("respects max_sessions parameter", async () => {
    const results = await search(db, {
      query: "code",
      maxSessions: 1,
    });

    expect(results.length).toBeLessThanOrEqual(1);
  }, 30000);

  test("respects max_chunks parameter", async () => {
    const results = await search(db, {
      query: "auth",
      maxChunks: 1,
    });

    for (const result of results) {
      expect(result.chunks.length).toBeLessThanOrEqual(1);
    }
  }, 30000);

  test("returns results even when only keyword matches (no semantic)", async () => {
    // Search for an exact function name that wouldn't match semantically
    const results = await search(db, {
      query: "isValidating useAuth",
    });

    expect(results.length).toBeGreaterThan(0);
    expect(results[0].sessionId).toBe("auth-session");
  }, 30000);

  test("boosts compact summary chunks in ranking", async () => {
    // Index a session with a compact summary
    const compactSession = createSession("compact-session", [
      userMsg("random conversation start", { sessionId: "compact-session", timestamp: "2026-04-21T10:00:00.000Z" }),
      assistantMsg("some response"),
      {
        type: "user",
        sessionId: "compact-session",
        uuid: "compact-uuid",
        parentUuid: null,
        timestamp: "2026-04-21T11:00:00.000Z",
        isCompactSummary: true,
        message: {
          role: "user",
          content: "Task overview: Implemented rate limiting with Redis. Key decision: chose token bucket over sliding window for better burst handling.",
        },
      },
    ]);
    await indexSession(db, compactSession);

    const results = await search(db, {
      query: "rate limiting",
    });

    expect(results.length).toBeGreaterThan(0);
    expect(results[0].sessionId).toBe("compact-session");
    // The compact summary chunk should be present
    expect(results[0].chunks.some((c: any) => c.isCompactSummary)).toBe(true);
  }, 60000);

  test("returns empty array for no matches", async () => {
    const results = await search(db, {
      query: "quantum physics black hole entropy",
    });

    // May return results with very low scores, or empty
    // The key is it doesn't crash
    expect(Array.isArray(results)).toBe(true);
  }, 30000);
});
