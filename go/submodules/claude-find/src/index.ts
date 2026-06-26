import { startServer } from "./server";
import { createDatabase } from "./db";
import { indexSessions } from "./indexer";
import { join } from "path";
import { existsSync } from "fs";

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

const command = process.argv[2];

async function main() {
  switch (command) {
    case "serve":
      await startServer();
      break;

    case "index": {
      if (existsSync(DB_PATH)) {
        const { unlinkSync } = await import("fs");
        unlinkSync(DB_PATH);
      }
      console.log("Indexing Claude Code sessions...\n");
      const db = createDatabase(DB_PATH);
      let indexed = 0;
      let skipped = 0;
      let errors = 0;
      const startTime = Date.now();
      await indexSessions(db, CLAUDE_PROJECTS_DIR, (p) => {
        if (p.status === "indexing") indexed++;
        else if (p.status === "skipped") skipped++;
        else if (p.status === "error") errors++;

        if (p.total === 0) {
          if (p.status === "done") console.log("\nNo sessions found.");
          return;
        }
        const pct = Math.round((p.current / p.total) * 100);
        const filled = Math.round(pct / 4);
        const bar = "\u2588".repeat(filled) + "\u2591".repeat(25 - filled);
        const elapsed = Math.round((Date.now() - startTime) / 1000);
        process.stdout.write(`\r  [${bar}] ${p.current}/${p.total} (${pct}%) ${elapsed}s     `);

        if (p.status === "done") {
          const totalTime = Math.round((Date.now() - startTime) / 1000);
          console.log(`\n\nDone in ${totalTime}s. ${indexed} indexed, ${skipped} unchanged, ${errors} errors.`);
        }
      });
      db.close();
      break;
    }

    case "status": {
      if (!existsSync(DB_PATH)) {
        console.log("No index found. Run 'claude-find index' or start the MCP server (indexes automatically on startup).");
        break;
      }
      const db = createDatabase(DB_PATH);
      const ids = db.getAllSessionIds();
      const archived = ids.filter((id) => {
        const s = db.getSession(id);
        return s?.is_archived;
      });
      console.log(`Sessions indexed: ${ids.length} (${ids.length - archived.length} live, ${archived.length} archived)`);
      const stat = Bun.file(DB_PATH);
      console.log(`Index size: ${(stat.size / 1024 / 1024).toFixed(1)} MB`);
      console.log(`Index path: ${DB_PATH}`);
      db.close();
      break;
    }

    case "install":
    case "setup": {
      const platform = process.platform;
      const check = (ok: boolean, msg: string) => console.log(ok ? `  ✓ ${msg}` : `  ✗ ${msg}`);
      let failed = false;

      console.log("\nclaude-find setup\n");

      // 1. Check Bun (if we're running, it's there)
      check(true, "Bun installed");

      // 2. Check Ollama installed
      const ollamaInstalled = Bun.which("ollama") !== null;
      if (!ollamaInstalled) {
        check(false, "Ollama not found");
        if (platform === "darwin") {
          console.log("\n    brew install ollama\n");
        } else if (platform === "win32") {
          console.log("\n    Download from https://ollama.com/download/windows\n");
        } else {
          console.log("\n    curl -fsSL https://ollama.com/install.sh | sh\n");
        }
        failed = true;
      } else {
        check(true, "Ollama installed");

        // 3. Check Ollama running
        let ollamaRunning = false;
        try {
          const resp = await fetch("http://localhost:11434/api/tags", { signal: AbortSignal.timeout(2000) });
          ollamaRunning = resp.ok;
        } catch {}

        if (!ollamaRunning) {
          // Try to start it
          console.log("  … Starting Ollama");
          Bun.spawn(["ollama", "serve"], { stdout: "ignore", stderr: "ignore" }).unref();
          // Wait for it to come up
          for (let i = 0; i < 10; i++) {
            await Bun.sleep(1000);
            try {
              const resp = await fetch("http://localhost:11434/api/tags", { signal: AbortSignal.timeout(1000) });
              if (resp.ok) { ollamaRunning = true; break; }
            } catch {}
          }
        }

        if (!ollamaRunning) {
          check(false, "Ollama not running (could not start automatically)");
          if (platform === "darwin") {
            console.log("\n    brew services start ollama\n");
          } else {
            console.log("\n    ollama serve &\n");
          }
          failed = true;
        } else {
          check(true, "Ollama running");

          // 4. Check/pull model
          try {
            const resp = await fetch("http://localhost:11434/api/tags");
            const data = await resp.json() as any;
            const models = data.models || [];
            const hasModel = models.some((m: any) => m.name?.startsWith("qwen3-embedding"));
            if (!hasModel) {
              console.log("  … Pulling qwen3-embedding:0.6b (639 MB)");
              const pullResp = await fetch("http://localhost:11434/api/pull", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ model: "qwen3-embedding:0.6b" }),
              });
              if (!pullResp.ok || !pullResp.body) throw new Error("Pull failed");
              const reader = pullResp.body.getReader();
              while (true) {
                const { done } = await reader.read();
                if (done) break;
              }
              check(true, "Model qwen3-embedding:0.6b ready");
            } else {
              check(true, "Model qwen3-embedding:0.6b ready");
            }
          } catch (err) {
            check(false, "Failed to pull model");
            console.log("\n    ollama pull qwen3-embedding:0.6b\n");
            failed = true;
          }
        }
      }

      if (failed) {
        console.log("Fix the above and re-run: bunx claude-find setup");
        process.exit(1);
      }

      // 5. Set session retention to permanent
      const settingsPath = join(process.env.HOME || "~", ".claude", "settings.json");
      try {
        let settings: any = {};
        if (existsSync(settingsPath)) {
          settings = JSON.parse(await Bun.file(settingsPath).text());
        }
        if (!settings.cleanupPeriodDays || settings.cleanupPeriodDays === 30) {
          settings.cleanupPeriodDays = 99999;
          await Bun.write(settingsPath, JSON.stringify(settings, null, 2) + "\n");
          check(true, "Session retention set to permanent (cleanupPeriodDays: 99999)");
        } else {
          check(true, `Session retention: ${settings.cleanupPeriodDays} days (custom, keeping yours)`);
        }
      } catch {
        check(false, "Could not update session retention");
        console.log(`\n    Add "cleanupPeriodDays": 99999 to ${settingsPath}\n`);
      }

      // 6. Register MCP server
      console.log("  … Registering MCP server");
      const proc = Bun.spawn(
        ["claude", "mcp", "add", "--scope", "user", "claude-find", "--", "bunx", "--bun", "claude-find", "serve"],
        { stdout: "pipe", stderr: "pipe" }
      );
      await proc.exited;
      const stderr = await new Response(proc.stderr).text();
      if (proc.exitCode === 0) {
        check(true, "MCP server registered");
      } else if (stderr.includes("already exists")) {
        check(true, "MCP server registered");
      } else {
        check(false, "Could not register MCP server");
        if (stderr.trim()) console.log(`    ${stderr.trim()}`);
        console.log("\n    claude mcp add claude-find -- bunx --bun claude-find serve\n");
        process.exit(1);
      }

      console.log("\nReady! Use /find in any Claude Code session.\n");
      break;
    }

    default:
      console.log(`claude-find — Pull deep memory from across your Claude Code sessions

Usage:
  claude-find setup     Install everything and register with Claude Code
  claude-find index     Rebuild index from scratch
  claude-find status    Show index statistics
`);
  }
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
