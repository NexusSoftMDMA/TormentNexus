import { test, expect, describe, afterAll, beforeEach } from "bun:test";
import { createDatabase, type ClaudeFindDB } from "../src/db";
import { rmSync } from "fs";
import { join } from "path";

const TEST_DB_PATH = join(import.meta.dir, ".test-index.db");

let db: ClaudeFindDB;

beforeEach(() => {
  // Close prior handle before creating a new one
  if (db) db.close();
  rmSync(TEST_DB_PATH, { force: true });
  db = createDatabase(TEST_DB_PATH);
});

afterAll(() => {
  rmSync(TEST_DB_PATH, { force: true });
});

describe("createDatabase", () => {
  test("creates database file and returns handle", () => {
    expect(db).toBeDefined();
    expect(db.db).toBeDefined();
  });

  test("sets WAL journal mode", () => {
    const result = db.db.prepare("PRAGMA journal_mode").get() as any;
    expect(result.journal_mode).toBe("wal");
  });

  test("creates sessions table", () => {
    const tables = db.db
      .prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'")
      .get() as any;
    expect(tables).toBeDefined();
    expect(tables.name).toBe("sessions");
  });

  test("creates session_files table", () => {
    const tables = db.db
      .prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='session_files'")
      .get() as any;
    expect(tables).toBeDefined();
  });

  test("creates chunks table", () => {
    const tables = db.db
      .prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='chunks'")
      .get() as any;
    expect(tables).toBeDefined();
  });

  test("creates FTS5 virtual table", () => {
    const tables = db.db
      .prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='chunks_fts'")
      .get() as any;
    expect(tables).toBeDefined();
  });
});

describe("session operations", () => {
  test("inserts and retrieves a session", () => {
    db.insertSession({
      id: "abc-123",
      projectPath: "/Users/test/myapp",
      branch: "main",
      title: "Fix auth bug",
      messageCount: 42,
      createdAt: "2026-04-06T10:00:00.000Z",
      updatedAt: "2026-04-06T12:00:00.000Z",
      fileSize: 5000,
    });

    const session = db.getSession("abc-123");
    expect(session).toBeDefined();
    expect(session!.id).toBe("abc-123");
    expect(session!.project_path).toBe("/Users/test/myapp");
    expect(session!.branch).toBe("main");
    expect(session!.title).toBe("Fix auth bug");
    expect(session!.message_count).toBe(42);
  });

  test("checks if session exists by id", () => {
    db.insertSession({
      id: "exists-123",
      projectPath: "/test",
      branch: null,
      title: null,
      messageCount: 1,
      createdAt: "2026-04-06T10:00:00.000Z",
      updatedAt: "2026-04-06T10:00:00.000Z",
      fileSize: 100,
    });

    expect(db.sessionExists("exists-123")).toBe(true);
    expect(db.sessionExists("nope-456")).toBe(false);
  });

  test("marks session as archived", () => {
    db.insertSession({
      id: "archive-me",
      projectPath: "/test",
      branch: null,
      title: null,
      messageCount: 1,
      createdAt: "2026-04-06T10:00:00.000Z",
      updatedAt: "2026-04-06T10:00:00.000Z",
      fileSize: 100,
    });

    db.markArchived("archive-me");

    const session = db.getSession("archive-me");
    expect(session!.is_archived).toBe(1);
    expect(session!.archived_at).toBeDefined();
  });
});

describe("session_files operations", () => {
  test("inserts and queries files for a session", () => {
    db.insertSession({
      id: "session-with-files",
      projectPath: "/test",
      branch: "main",
      title: null,
      messageCount: 10,
      createdAt: "2026-04-06T10:00:00.000Z",
      updatedAt: "2026-04-06T10:00:00.000Z",
      fileSize: 1000,
    });

    db.insertSessionFile("session-with-files", "/src/auth.ts", "read");
    db.insertSessionFile("session-with-files", "/src/auth.ts", "edit");
    db.insertSessionFile("session-with-files", "/src/login.tsx", "write");

    const files = db.getSessionFiles("session-with-files");
    expect(files.length).toBe(3);
    expect(files.some((f: any) => f.file_path === "/src/auth.ts" && f.operation === "read")).toBe(true);
    expect(files.some((f: any) => f.file_path === "/src/auth.ts" && f.operation === "edit")).toBe(true);
    expect(files.some((f: any) => f.file_path === "/src/login.tsx" && f.operation === "write")).toBe(true);
  });

  test("finds sessions by file path", () => {
    db.insertSession({ id: "s1", projectPath: "/test", branch: null, title: null, messageCount: 1, createdAt: "2026-04-01T00:00:00Z", updatedAt: "2026-04-01T00:00:00Z", fileSize: 100 });
    db.insertSession({ id: "s2", projectPath: "/test", branch: null, title: null, messageCount: 1, createdAt: "2026-04-02T00:00:00Z", updatedAt: "2026-04-02T00:00:00Z", fileSize: 100 });

    db.insertSessionFile("s1", "/src/auth.ts", "edit");
    db.insertSessionFile("s2", "/src/auth.ts", "read");
    db.insertSessionFile("s2", "/src/other.ts", "edit");

    const sessions = db.findSessionsByFile("/src/auth.ts");
    expect(sessions.length).toBe(2);
  });
});

describe("chunk operations", () => {
  test("inserts and retrieves chunks for a session", () => {
    db.insertSession({ id: "chunked", projectPath: "/test", branch: null, title: null, messageCount: 50, createdAt: "2026-04-06T10:00:00Z", updatedAt: "2026-04-06T12:00:00Z", fileSize: 5000 });

    const chunkId1 = db.insertChunk("chunked", 0, 19, "First chunk of conversation text", false);
    const chunkId2 = db.insertChunk("chunked", 20, 39, "Second chunk about auth", false);
    const chunkId3 = db.insertChunk("chunked", -1, -1, "Compact summary: Fixed auth bug", true);

    expect(chunkId1).toBeGreaterThan(0);
    expect(chunkId2).toBeGreaterThan(0);
    expect(chunkId3).toBeGreaterThan(0);

    const chunks = db.getChunksForSession("chunked");
    expect(chunks.length).toBe(3);
    expect(chunks[2].is_compact_summary).toBe(1);
  });
});

describe("FTS5 search", () => {
  test("finds chunks by keyword", () => {
    db.insertSession({ id: "fts-test", projectPath: "/test", branch: null, title: null, messageCount: 10, createdAt: "2026-04-06T10:00:00Z", updatedAt: "2026-04-06T10:00:00Z", fileSize: 1000 });

    const id1 = db.insertChunk("fts-test", 0, 9, "Debugging the authentication flash bug in the login page", false);
    const id2 = db.insertChunk("fts-test", 10, 19, "Refactoring the payment processing middleware", false);

    const results = db.searchFTS("authentication");
    expect(results.length).toBe(1);
    expect(results[0].chunk_id).toBe(id1);
  });

  test("FTS search returns ranked results", () => {
    db.insertSession({ id: "fts-rank", projectPath: "/test", branch: null, title: null, messageCount: 30, createdAt: "2026-04-06T10:00:00Z", updatedAt: "2026-04-06T10:00:00Z", fileSize: 2000 });

    db.insertChunk("fts-rank", 0, 9, "The auth module handles authentication", false);
    db.insertChunk("fts-rank", 10, 19, "Authentication and authorization in the auth middleware with auth tokens", false);
    db.insertChunk("fts-rank", 20, 29, "Payment processing pipeline", false);

    const results = db.searchFTS("auth");
    expect(results.length).toBe(2); // "Payment" chunk shouldn't match
  });
});

describe("vector operations", () => {
  test("inserts and queries vectors", () => {
    db.insertSession({ id: "vec-test", projectPath: "/test", branch: null, title: null, messageCount: 10, createdAt: "2026-04-06T10:00:00Z", updatedAt: "2026-04-06T10:00:00Z", fileSize: 1000 });

    const id1 = db.insertChunk("vec-test", 0, 9, "auth discussion", false);
    const id2 = db.insertChunk("vec-test", 10, 19, "payment processing", false);

    // Insert fake 1024-dim vectors
    const vec1 = new Float32Array(1024).fill(0.1);
    const vec2 = new Float32Array(1024).fill(0.9);

    db.insertVector(id1, vec1);
    db.insertVector(id2, vec2);

    // Query with a vector close to vec1
    const queryVec = new Float32Array(1024).fill(0.15);
    const results = db.searchVectors(queryVec, 2);

    expect(results.length).toBe(2);
    // vec1 (0.1) should be closer to query (0.15) than vec2 (0.9)
    expect(results[0].chunk_id).toBe(id1);
  });
});

describe("listing sessions", () => {
  test("lists all indexed session IDs", () => {
    db.insertSession({ id: "s1", projectPath: "/a", branch: null, title: null, messageCount: 1, createdAt: "2026-04-01T00:00:00Z", updatedAt: "2026-04-01T00:00:00Z", fileSize: 100 });
    db.insertSession({ id: "s2", projectPath: "/b", branch: null, title: null, messageCount: 1, createdAt: "2026-04-02T00:00:00Z", updatedAt: "2026-04-02T00:00:00Z", fileSize: 200 });

    const ids = db.getAllSessionIds();
    expect(ids).toContain("s1");
    expect(ids).toContain("s2");
  });
});
