import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { createDatabase } from "./db";
import { indexSessions } from "./indexer";
import { search, type SearchResult } from "./searcher";
import { join } from "path";
import { existsSync } from "fs";

// Track indexing progress at module scope so global error handlers can update it
let indexProgress = { current: 0, total: 0, done: false };

// Prevent background indexing errors from crashing the MCP server
process.on("uncaughtException", (err) => {
  console.error("[claude-find] Uncaught exception:", err);
  indexProgress.done = true;
});
process.on("unhandledRejection", (err) => {
  console.error("[claude-find] Unhandled rejection:", err);
  indexProgress.done = true;
});

const CLAUDE_PROJECTS_DIR = join(
  process.env.HOME || "~",
  ".claude",
  "projects"
);
const DB_PATH = join(
  process.env.HOME || "~",
  ".claude-find",
  "index.db"
);

/**
 * Format search results as readable text for Claude.
 */
export function formatSearchResults(results: SearchResult[]): string {
  if (results.length === 0) {
    return "No matching sessions found.";
  }

  const lines: string[] = [];
  lines.push(`Found ${results.length} relevant session${results.length > 1 ? "s" : ""}:\n`);

  for (const result of results) {
    lines.push("---");
    lines.push(`**${result.title || "Untitled session"}**${result.isArchived ? " [archived]" : ""}`);
    lines.push(`Project: ${result.projectPath} | Branch: ${result.branch || "unknown"} | ${result.messageCount} messages`);

    if (result.createdAt) {
      const date = new Date(result.createdAt).toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
      });
      lines.push(`Date: ${date}`);
    }

    if (result.filesTouched.length > 0) {
      lines.push(`Files: ${result.filesTouched.join(", ")}`);
    }

    lines.push("");

    for (const chunk of result.chunks) {
      if (chunk.isCompactSummary) {
        lines.push(`[Compact Summary]`);
      }
      lines.push(chunk.text);
      lines.push("");
    }
  }

  return lines.join("\n");
}

/**
 * Start the MCP server.
 */
export async function startServer(): Promise<void> {
  const server = new McpServer({
    name: "claude-find",
    version: "0.1.0",
  });

  let db: ReturnType<typeof createDatabase> | null = null;

  function getDb() {
    if (!db) {
      db = createDatabase(DB_PATH);
    }
    return db;
  }

  // Start indexing in the background immediately — searches can run against partial results
  if (existsSync(CLAUDE_PROJECTS_DIR)) {
    const database = getDb();
    console.error("[claude-find] Starting background indexing...");
    indexSessions(database, CLAUDE_PROJECTS_DIR, (p) => {
      indexProgress.current = p.current;
      indexProgress.total = p.total;
      if (p.status === "indexing") {
        console.error(`[claude-find] Indexing ${p.current}/${p.total}: ${p.sessionId}`);
      }
      if (p.status === "done") {
        indexProgress.done = true;
        console.error("[claude-find] Indexing complete.");
      }
    }).catch((err) => {
      console.error("[claude-find] Indexing error:", err);
      indexProgress.done = true;
    });
  } else {
    indexProgress.done = true;
  }

  server.tool(
    "search_sessions",
    "Search the full conversation history from past Claude Code sessions stored in ~/.claude/projects/. This tool has access to the complete raw transcripts of all previous sessions — including the actual back-and-forth discussion, reasoning, failed approaches, user constraints, and code decisions. Use this tool FIRST whenever the user mentions anything from a past session, asks 'what did we discuss', 'pull in context from', 'remember when we', 'how did we handle', or references any prior work. This tool searches semantically — the user doesn't need to remember exact words. Much more detailed than built-in memory.",
    {
      query: z.string().describe("What to search for — natural language description of the past session or topic"),
      max_sessions: z.number().min(1).max(5).optional().default(3).describe("Max sessions to return (default 3, max 5)"),
      max_chunks: z.number().min(1).max(3).optional().default(3).describe("Max conversation chunks per session (default 3, max 3)"),
      scope: z.enum(["current", "all"]).optional().default("current").describe("'current' searches only the current project (default), 'all' searches across all projects. Use 'all' when user says 'across all projects' or doesn't specify a project."),
      project_filter: z.string().optional().describe("Filter to a specific project by name (e.g. 'visk', 'myapp'). Use when user says 'in visk' or 'in the payments project'."),
    },
    async ({ query, max_sessions, max_chunks, scope, project_filter }) => {
      try {
        const database = getDb();

        // Determine project scope
        let currentProject: string | undefined;
        if (project_filter) {
          currentProject = project_filter;
        } else if (scope === "current") {
          currentProject = process.cwd();
        }

        const results = await search(database, {
          query,
          maxSessions: max_sessions,
          maxChunks: max_chunks,
          currentProject,
        });

        let formatted = formatSearchResults(results);

        if (!indexProgress.done) {
          formatted += `\n\n---\n_Indexing in progress (${indexProgress.current}/${indexProgress.total} sessions) — results may be incomplete._`;
        }

        return {
          content: [{ type: "text" as const, text: formatted }],
        };
      } catch (err) {
        console.error("[claude-find] Search error:", err);
        return {
          content: [
            {
              type: "text" as const,
              text: `Error searching sessions: ${err instanceof Error ? err.message : String(err)}`,
            },
          ],
          isError: true,
        };
      }
    }
  );

  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("[claude-find] MCP server running on stdio");

  // Bun exits when the event loop drains. StdioServerTransport doesn't register
  // as an active handle, so keep the process alive explicitly.
  setInterval(() => {}, 2 ** 31 - 1);
}
