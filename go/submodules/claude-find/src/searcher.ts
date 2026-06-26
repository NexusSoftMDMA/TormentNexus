import { getEmbedding } from "./embeddings";
import type { ClaudeFindDB } from "./db";
import { basename } from "path";

export interface SearchOptions {
  query: string;
  maxSessions?: number;
  maxChunks?: number;
  currentProject?: string;
  currentBranch?: string;
}

export interface SearchChunk {
  text: string;
  msgStart: number;
  msgEnd: number;
  isCompactSummary: boolean;
  score: number;
}

export interface SearchResult {
  sessionId: string;
  title: string | null;
  branch: string | null;
  projectPath: string;
  createdAt: string | null;
  updatedAt: string | null;
  messageCount: number;
  isArchived: boolean;
  filesTouched: string[];
  chunks: SearchChunk[];
  score: number;
}

const RRF_K = 60;
const COMPACT_SUMMARY_BOOST = 1.5;
const CURRENT_PROJECT_BOOST = 1.2;
const CURRENT_BRANCH_BOOST = 1.1;

/**
 * Hybrid search: semantic + FTS5 keyword, merged via Reciprocal Rank Fusion.
 * Returns results grouped by session, ranked by best chunk score.
 */
export async function search(
  db: ClaudeFindDB,
  options: SearchOptions
): Promise<SearchResult[]> {
  const {
    query,
    maxSessions = 3,
    maxChunks = 2,
    currentProject,
    currentBranch,
  } = options;

  // Wider search window — we'll narrow after grouping
  const searchLimit = maxSessions * maxChunks * 5;

  // 1. Semantic search — enrich query with project context to match document-side enrichment
  let enrichedQuery = query;
  if (currentProject) {
    enrichedQuery = `${basename(currentProject)}: ${query}`;
  }
  const queryEmbedding = await getEmbedding(enrichedQuery, "query");
  const vectorResults = db.searchVectors(queryEmbedding, searchLimit);

  // 2. Keyword search — sanitize query for FTS5 syntax, degrade gracefully on error
  let ftsResults: any[] = [];
  try {
    const ftsQuery = query
      .split(/\s+/)
      .filter(Boolean)
      .map((term) => term.replace(/[^\w]/g, ""))
      .filter((term) => term && !/^(AND|OR|NOT|NEAR)$/i.test(term))
      .map((term) => `"${term}"`)
      .join(" ");
    if (ftsQuery) {
      ftsResults = db.searchFTS(ftsQuery, searchLimit);
    }
  } catch (err) {
    console.warn(`[claude-find] FTS5 fallback for query "${query}":`, err instanceof Error ? err.message : err);
  }

  // 3. Reciprocal Rank Fusion
  const chunkScores = new Map<number, number>();

  for (let i = 0; i < vectorResults.length; i++) {
    const chunkId = vectorResults[i].chunk_id;
    const rrfScore = 1 / (RRF_K + i);
    chunkScores.set(chunkId, (chunkScores.get(chunkId) || 0) + rrfScore);
  }

  for (let i = 0; i < ftsResults.length; i++) {
    const chunkId = ftsResults[i].chunk_id;
    const rrfScore = 1 / (RRF_K + i);
    chunkScores.set(chunkId, (chunkScores.get(chunkId) || 0) + rrfScore);
  }

  if (chunkScores.size === 0) return [];

  // 4. Fetch chunk details and apply boosts
  const sessionCache = new Map<string, any>();

  const getSessionCached = (sessionId: string) => {
    if (!sessionCache.has(sessionId)) {
      sessionCache.set(sessionId, db.getSession(sessionId));
    }
    return sessionCache.get(sessionId);
  };

  const scoredChunks: Array<{
    chunkId: number;
    sessionId: string;
    text: string;
    msgStart: number;
    msgEnd: number;
    isCompactSummary: boolean;
    score: number;
  }> = [];

  for (const [chunkId, baseScore] of chunkScores) {
    const chunk = db.getChunkById(chunkId) as any;

    if (!chunk) continue;

    let score = baseScore;

    // Boost compact summary chunks
    if (chunk.is_compact_summary) {
      score *= COMPACT_SUMMARY_BOOST;
    }

    // Context-aware filtering and boosts
    const session = getSessionCached(chunk.session_id);
    if (session && currentProject) {
      const projectPath = session.project_path || "";
      const isMatch = projectPath === currentProject || projectPath.includes(currentProject);
      if (!isMatch) {
        // Filter out non-matching projects (not just boost)
        continue;
      }
      score *= CURRENT_PROJECT_BOOST;
    }
    if (session && currentBranch && session.branch === currentBranch) {
      score *= CURRENT_BRANCH_BOOST;
    }

    scoredChunks.push({
      chunkId,
      sessionId: chunk.session_id,
      text: chunk.text,
      msgStart: chunk.msg_start,
      msgEnd: chunk.msg_end,
      isCompactSummary: !!chunk.is_compact_summary,
      score,
    });
  }

  // 5. Group by session, take top chunks per session
  const sessionMap = new Map<string, typeof scoredChunks>();

  for (const chunk of scoredChunks) {
    const existing = sessionMap.get(chunk.sessionId) || [];
    existing.push(chunk);
    sessionMap.set(chunk.sessionId, existing);
  }

  // 6. Build results: sort chunks within each session, take top N
  const results: SearchResult[] = [];

  for (const [sessionId, chunks] of sessionMap) {
    const session = getSessionCached(sessionId);
    if (!session) continue;

    // Sort chunks by score descending, take top maxChunks
    const topChunks = chunks
      .sort((a, b) => b.score - a.score)
      .slice(0, maxChunks);

    // Session score = best chunk score
    const sessionScore = topChunks[0].score;

    // Get files touched
    const files = db.getSessionFiles(sessionId);
    const uniqueFiles = [...new Set(files.map((f: any) => f.file_path))];

    results.push({
      sessionId,
      title: session.title,
      branch: session.branch,
      projectPath: session.project_path,
      createdAt: session.created_at,
      updatedAt: session.updated_at,
      messageCount: session.message_count,
      isArchived: !!session.is_archived,
      filesTouched: uniqueFiles,
      chunks: topChunks.map((c) => ({
        text: c.text,
        msgStart: c.msgStart,
        msgEnd: c.msgEnd,
        isCompactSummary: c.isCompactSummary,
        score: c.score,
      })),
      score: sessionScore,
    });
  }

  // 7. Sort sessions by score descending, return top N
  return results
    .sort((a, b) => b.score - a.score)
    .slice(0, maxSessions);
}
