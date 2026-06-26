import type { Chunk } from "./chunker";
import type { SessionMetadata, FileTouch } from "./parser";
import { basename } from "path";

/**
 * Format an ISO timestamp as a coarse date like "Apr 2026".
 */
function formatCoarseDate(isoString: string | null): string | null {
  if (!isoString) return null;
  try {
    const d = new Date(isoString);
    return d.toLocaleDateString("en-US", { month: "short", year: "numeric" });
  } catch {
    return null;
  }
}

/**
 * Extract unique file basenames, prioritizing edits/writes over reads.
 */
function topFileBasenames(files: FileTouch[], limit: number): string[] {
  const seen = new Set<string>();
  const result: string[] = [];

  // Edits and writes first (higher signal)
  for (const f of files) {
    if (f.operation !== "read") {
      const name = basename(f.path);
      if (!seen.has(name)) {
        seen.add(name);
        result.push(name);
      }
    }
  }
  // Then reads
  for (const f of files) {
    if (f.operation === "read") {
      const name = basename(f.path);
      if (!seen.has(name)) {
        seen.add(name);
        result.push(name);
      }
    }
  }

  return result.slice(0, limit);
}

/**
 * Prepend a structured metadata prefix to a chunk's text for better
 * retrieval quality. The prefix gives the embedding model and FTS5
 * context about which project, branch, files, and time period the
 * conversation relates to.
 *
 * Returns the original text unchanged if no metadata is available.
 */
export function enrichChunkText(
  chunk: Chunk,
  metadata: SessionMetadata,
  filesTouched: FileTouch[]
): string {
  const parts: string[] = [];

  if (metadata.projectPath) {
    parts.push(`Project: ${basename(metadata.projectPath)}`);
  }

  if (metadata.branch && metadata.branch !== "main" && metadata.branch !== "master") {
    parts.push(`Branch: ${metadata.branch}`);
  }

  if (chunk.isCompactSummary && metadata.title) {
    parts.push(`Topic: ${metadata.title}`);
  }

  const date = formatCoarseDate(metadata.createdAt);
  if (date) {
    parts.push(date);
  }

  const fileNames = topFileBasenames(filesTouched, 5);
  if (fileNames.length > 0) {
    parts.push(`Files: ${fileNames.join(", ")}`);
  }

  if (parts.length === 0) return chunk.text;

  return `[${parts.join(" | ")}]\n${chunk.text}`;
}
