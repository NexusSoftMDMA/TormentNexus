import { parseSession } from "./parser";
import { chunkSession } from "./chunker";
import { enrichChunkText } from "./enrichment";
import { getEmbeddings } from "./embeddings";
import type { ClaudeFindDB } from "./db";
import { existsSync, statSync, readdirSync } from "fs";
import { join, basename } from "path";

export interface IndexProgress {
  current: number;
  total: number;
  sessionId: string;
  status: "indexing" | "skipped" | "done" | "error";
}

type ProgressCallback = (progress: IndexProgress) => void;

/**
 * Index a single session JSONL file.
 * Returns false if skipped (already indexed with same file size).
 */
export async function indexSession(
  db: ClaudeFindDB,
  filePath: string
): Promise<boolean> {
  // Use filename as the canonical session ID (matches claude --resume)
  const sessionId = basename(filePath, ".jsonl");
  const stat = statSync(filePath);

  // Skip if already indexed with same file size (no changes)
  if (db.sessionExists(sessionId)) {
    const existing = db.getSession(sessionId);
    if (existing.file_size === stat.size) {
      return false;
    }
  }

  // Parse the session
  const parsed = await parseSession(filePath);

  // Chunk the messages — embed everything for full coverage
  const chunks = chunkSession(parsed.messages, parsed.compactSummaries);

  if (chunks.length === 0) return false;

  // Enrich chunks with session context (project, branch, files, date) for better retrieval
  const enrichedTexts = chunks.map((chunk) =>
    enrichChunkText(chunk, parsed.metadata, parsed.filesTouched)
  );

  // Embed in small batches — tensor dispose prevents memory leaks
  const BATCH_SIZE = 4;
  const embeddings: (Float32Array | null)[] = [];
  for (let i = 0; i < chunks.length; i += BATCH_SIZE) {
    const batch = enrichedTexts.slice(i, i + BATCH_SIZE);
    const batchEmbeddings = await getEmbeddings(batch, "document");
    embeddings.push(...batchEmbeddings);
  }

  // Write all data in a single transaction — if anything fails,
  // everything rolls back so the next run retries cleanly
  const writeAll = db.db.transaction(() => {
    db.insertSession({
      id: sessionId,
      projectPath: parsed.metadata.projectPath || "",
      branch: parsed.metadata.branch,
      title: parsed.metadata.title,
      messageCount: parsed.metadata.messageCount,
      createdAt: parsed.metadata.createdAt,
      updatedAt: parsed.metadata.updatedAt,
      fileSize: stat.size,
    });

    for (const file of parsed.filesTouched) {
      db.insertSessionFile(sessionId, file.path, file.operation);
    }

    for (let i = 0; i < chunks.length; i++) {
      const chunk = chunks[i];
      const chunkId = db.insertChunk(
        sessionId,
        chunk.msgStart,
        chunk.msgEnd,
        enrichedTexts[i],
        chunk.isCompactSummary
      );
      if (embeddings[i]) {
        db.insertVector(chunkId, embeddings[i]);
      }
    }
  });

  writeAll();
  return true;
}

/**
 * Scan a Claude Code projects directory and index all sessions.
 * Detects archived sessions (JSONL deleted but still in index).
 */
export async function indexSessions(
  db: ClaudeFindDB,
  projectsDir: string,
  onProgress?: ProgressCallback
): Promise<void> {
  // Find all JSONL files across all project directories
  const sessionFiles: string[] = [];

  if (!existsSync(projectsDir)) return;

  const projectDirs = readdirSync(projectsDir, { withFileTypes: true })
    .filter((d) => d.isDirectory())
    .map((d) => join(projectsDir, d.name));

  for (const projectDir of projectDirs) {
    const files = readdirSync(projectDir)
      .filter((f) => f.endsWith(".jsonl"))
      .map((f) => join(projectDir, f));
    sessionFiles.push(...files);
  }

  const total = sessionFiles.length;

  // Index each session
  for (let i = 0; i < sessionFiles.length; i++) {
    const filePath = sessionFiles[i];
    const sessionId = basename(filePath, ".jsonl");

    try {
      const indexed = await indexSession(db, filePath);
      onProgress?.({
        current: i + 1,
        total,
        sessionId,
        status: indexed ? "indexing" : "skipped",
      });
    } catch (err) {
      console.error(`[claude-find] Error indexing ${sessionId}:`, err);
      onProgress?.({
        current: i + 1,
        total,
        sessionId,
        status: "error",
      });
    }
  }

  // Detect archived sessions: indexed sessions whose JSONL no longer exists
  const indexedIds = db.getAllSessionIds();
  const existingFiles = new Set(
    sessionFiles.map((f) => basename(f, ".jsonl"))
  );

  for (const id of indexedIds) {
    const session = db.getSession(id);
    if (!session.is_archived && !existingFiles.has(id)) {
      db.markArchived(id);
    }
  }

  onProgress?.({
    current: total,
    total,
    sessionId: "",
    status: "done",
  });
}
