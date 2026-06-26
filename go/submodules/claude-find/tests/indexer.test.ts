import { test, expect, describe, afterAll, beforeEach } from "bun:test";
import { indexSessions, indexSession, type IndexProgress } from "../src/indexer";
import { createDatabase, type ClaudeFindDB } from "../src/db";
import { rmSync, mkdirSync, writeFileSync } from "fs";
import { join } from "path";

const TEST_DIR = join(import.meta.dir, ".fixtures-indexer");
const TEST_DB_PATH = join(import.meta.dir, ".test-indexer.db");

let db: ClaudeFindDB;

// Helper to create a fake session JSONL file
function createFakeSession(
  dir: string,
  sessionId: string,
  records: object[]
): string {
  mkdirSync(dir, { recursive: true });
  const path = join(dir, `${sessionId}.jsonl`);
  writeFileSync(path, records.map((r) => JSON.stringify(r)).join("\n"));
  return path;
}

function userMsg(text: string, opts: Record<string, any> = {}) {
  return {
    type: "user",
    sessionId: opts.sessionId || "test-session",
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
    sessionId: "test-session",
    uuid: `a-${Math.random().toString(36).slice(2)}`,
    parentUuid: null,
    timestamp: "2026-04-06T10:05:00.000Z",
    gitBranch: "main",
    message: {
      role: "assistant",
      content: [{ type: "text", text }],
    },
  };
}

function toolUseMsg(toolName: string, filePath: string) {
  return {
    type: "assistant",
    sessionId: "test-session",
    uuid: `a-${Math.random().toString(36).slice(2)}`,
    parentUuid: null,
    timestamp: "2026-04-06T10:05:00.000Z",
    gitBranch: "main",
    message: {
      role: "assistant",
      content: [
        { type: "tool_use", id: "toolu_123", name: toolName, input: { file_path: filePath } },
      ],
    },
  };
}

function summaryRecord(title: string) {
  return { type: "summary", summary: title, leafUuid: "leaf-123" };
}

beforeEach(() => {
  if (db) db.close();
  rmSync(TEST_DB_PATH, { force: true });
  rmSync(TEST_DIR, { recursive: true, force: true });
  db = createDatabase(TEST_DB_PATH);
});

afterAll(() => {
  if (db) db.close();
  rmSync(TEST_DB_PATH, { force: true });
  rmSync(TEST_DIR, { recursive: true, force: true });
});

describe("indexSession", () => {
  test("indexes a single session file into the database", async () => {
    const sessionDir = join(TEST_DIR, "project-a");
    const filePath = createFakeSession(sessionDir, "session-abc", [
      userMsg("fix the auth flash bug", { sessionId: "session-abc" }),
      assistantMsg("I found the issue in session.ts"),
      toolUseMsg("Read", "/src/auth/session.ts"),
      toolUseMsg("Edit", "/src/auth/session.ts"),
      userMsg("looks good, ship it"),
      assistantMsg("Done, committed the fix"),
      summaryRecord("Auth Flash Bug Fix"),
    ]);

    await indexSession(db, filePath);

    // Session should be in DB
    expect(db.sessionExists("session-abc")).toBe(true);

    const session = db.getSession("session-abc");
    expect(session.title).toBe("Auth Flash Bug Fix");
    expect(session.branch).toBe("main");
    expect(session.message_count).toBe(4); // 2 user text + 2 assistant text

    // Files should be tracked
    const files = db.getSessionFiles("session-abc");
    expect(files.some((f: any) => f.file_path === "/src/auth/session.ts")).toBe(true);

    // Chunks should exist
    const chunks = db.getChunksForSession("session-abc");
    expect(chunks.length).toBeGreaterThan(0);

    // Chunk text should contain the conversation
    const allText = chunks.map((c: any) => c.text).join(" ");
    expect(allText).toContain("auth flash bug");

    // Chunks should be enriched with metadata prefix
    expect(allText).toContain("[Project: myapp");
  }, 60000); // allow time for embedding model load

  test("stores embeddings for each chunk", async () => {
    const sessionDir = join(TEST_DIR, "project-b");
    createFakeSession(sessionDir, "session-emb", [
      userMsg("implement the payment webhook handler", { sessionId: "session-emb" }),
      assistantMsg("I'll create a webhook endpoint for Stripe"),
    ]);

    await indexSession(db, join(sessionDir, "session-emb.jsonl"));

    // Should be able to search by vector
    const chunks = db.getChunksForSession("session-emb");
    expect(chunks.length).toBeGreaterThan(0);

    // Vector search should return results (embedding was stored)
    // We need to import getEmbedding to create a query vector
    const { getEmbedding } = await import("../src/embeddings");
    const queryVec = await getEmbedding("stripe payment webhook", "query");
    const results = db.searchVectors(queryVec, 5);
    expect(results.length).toBeGreaterThan(0);
  }, 60000);

  test("populates FTS5 index for keyword search", async () => {
    const sessionDir = join(TEST_DIR, "project-c");
    createFakeSession(sessionDir, "session-fts", [
      userMsg("debug the authentication middleware", { sessionId: "session-fts" }),
      assistantMsg("The middleware is rejecting valid JWT tokens"),
    ]);

    await indexSession(db, join(sessionDir, "session-fts.jsonl"));

    const results = db.searchFTS("authentication");
    expect(results.length).toBeGreaterThan(0);
  }, 60000);

  test("skips already-indexed sessions (same file size)", async () => {
    const sessionDir = join(TEST_DIR, "project-d");
    const filePath = createFakeSession(sessionDir, "session-skip", [
      userMsg("hello", { sessionId: "session-skip" }),
      assistantMsg("hi there"),
    ]);

    // Index once
    await indexSession(db, filePath);
    const firstChunks = db.getChunksForSession("session-skip");

    // Index again — should skip
    const skipped = await indexSession(db, filePath);
    expect(skipped).toBe(false); // returns false when skipped

    // Chunks should be the same
    const secondChunks = db.getChunksForSession("session-skip");
    expect(secondChunks.length).toBe(firstChunks.length);
  }, 60000);
});

describe("indexSessions", () => {
  test("indexes all sessions in a projects directory", async () => {
    const projectsDir = join(TEST_DIR, "projects");
    const projectA = join(projectsDir, "-Users-test-myapp");
    const projectB = join(projectsDir, "-Users-test-other");

    createFakeSession(projectA, "s1", [
      userMsg("fix auth", { sessionId: "s1", cwd: "/Users/test/myapp" }),
      assistantMsg("fixing auth"),
    ]);
    createFakeSession(projectA, "s2", [
      userMsg("add tests", { sessionId: "s2", cwd: "/Users/test/myapp" }),
      assistantMsg("adding tests"),
    ]);
    createFakeSession(projectB, "s3", [
      userMsg("deploy", { sessionId: "s3", cwd: "/Users/test/other" }),
      assistantMsg("deploying"),
    ]);

    const progress: IndexProgress[] = [];
    await indexSessions(db, projectsDir, (p) => progress.push(p));

    expect(db.sessionExists("s1")).toBe(true);
    expect(db.sessionExists("s2")).toBe(true);
    expect(db.sessionExists("s3")).toBe(true);

    // Progress callback should have been called
    expect(progress.length).toBeGreaterThan(0);
    expect(progress[progress.length - 1].total).toBe(3);
  }, 120000);

  test("detects archived sessions when JSONL is deleted", async () => {
    const projectsDir = join(TEST_DIR, "projects-archive");
    const projectDir = join(projectsDir, "-Users-test-app");

    const filePath = createFakeSession(projectDir, "will-delete", [
      userMsg("temp session", { sessionId: "will-delete" }),
      assistantMsg("response"),
    ]);

    // Index it
    await indexSessions(db, projectsDir);
    expect(db.sessionExists("will-delete")).toBe(true);
    expect(db.getSession("will-delete").is_archived).toBe(0);

    // Delete the JSONL file
    rmSync(filePath);

    // Re-index — should detect deletion and mark archived
    await indexSessions(db, projectsDir);
    const session = db.getSession("will-delete");
    expect(session.is_archived).toBe(1);
    expect(session.archived_at).toBeDefined();

    // Chunks should still exist (preserved)
    const chunks = db.getChunksForSession("will-delete");
    expect(chunks.length).toBeGreaterThan(0);
  }, 120000);
});
