# Agent A — Daemon + CLI

You own: `src/daemon/`, `src/cli/`, `src/lib/db/` (new), `bin/chorus.js`, `templates/seed.ts` (new).

## Build

### 1. CLI (`src/cli/index.ts`, called by `bin/chorus.js`)

```
chorus init                # one-time: create ~/.chorus/, seed DB, copy built-in templates
chorus start [--ui]        # spawn daemon (PM2-style fork), optionally open browser
chorus ui                  # open http://127.0.0.1:3011 in default browser
chorus mcp                 # exec the MCP server on stdio (called by orchestrators)
chorus stop                # SIGTERM the daemon
chorus status              # daemon health check (HTTP /health on :7707)
```

Use `commander`. Use `open` (npm) for `chorus ui`. Spawn daemon with `child_process.spawn(detached:true, stdio:'ignore')` and write PID to `~/.chorus/daemon.pid`.

### 2. Daemon (`src/daemon/index.ts`)

Fastify server on `127.0.0.1:7707` (override via `CHORUS_DAEMON_PORT`).

Routes:

| Method | Path | Purpose |
|---|---|---|
| GET    | `/health` | `{ ok: true, version, uptime }` |
| GET    | `/chats` | list chats (status filter, limit, offset) |
| GET    | `/chats/:id` | one chat with phase state |
| POST   | `/chats` | create chat `{ work, templateId, files? }` |
| POST   | `/chats/:id/cancel` | mark cancelled, kill tmux |
| POST   | `/chats/:id/resume` | answer a blocking question `{ answer }` |
| GET    | `/chats/:id/stream` | SSE stream of phase events |
| GET    | `/templates` | list templates (built-in + user) |
| GET    | `/templates/:id` | one template (YAML→JSON) |
| POST   | `/templates` | save user template `{ id, yaml }` |
| GET    | `/blocked` | chats waiting on user (`status='blocked'`) |
| GET    | `/settings` | get settings JSON |
| PUT    | `/settings` | upsert settings JSON |
| GET    | `/secrets` | list credentials (no values, just names + provider) |
| PUT    | `/secrets/:provider` | upsert credential value (encrypted at rest later; v0.5: plain in DB ok) |

CORS: allow `http://127.0.0.1:3011` only (the local UI).

### 3. SQLite schema (`src/lib/db/schema.sql` + `src/lib/db/index.ts`)

Use `better-sqlite3`. DB file: `~/.chorus/chorus.db`.

```sql
CREATE TABLE chats (
  id TEXT PRIMARY KEY,                    -- ULID
  work TEXT NOT NULL,                     -- the prompt
  template_id TEXT NOT NULL,
  status TEXT NOT NULL,                   -- drafting|reviewing|approved|merged|blocked|cancelled|failed
  current_phase_idx INTEGER DEFAULT 0,
  yolo BOOLEAN DEFAULT 0,
  attached_files TEXT,                    -- JSON array
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL,
  finished_at INTEGER
);

CREATE TABLE phase_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  chat_id TEXT NOT NULL REFERENCES chats(id),
  phase_idx INTEGER NOT NULL,
  phase_kind TEXT NOT NULL,               -- plan|spec|tests|implement|review|verify|divergence
  role TEXT NOT NULL,                     -- doer|reviewer
  agent_id TEXT,                          -- e.g. claude-opus-4-7, codex-gpt-5
  state TEXT NOT NULL,                    -- drafting|submitted|reviewing|approved|revising|blocked
  output TEXT,                            -- JSON or text
  cost_usd REAL DEFAULT 0,
  tokens_in INTEGER DEFAULT 0,
  tokens_out INTEGER DEFAULT 0,
  started_at INTEGER NOT NULL,
  finished_at INTEGER
);

CREATE TABLE templates (
  id TEXT PRIMARY KEY,                    -- e.g. code-review, t-red-green
  source TEXT NOT NULL,                   -- builtin|user
  yaml TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE TABLE settings (
  key TEXT PRIMARY KEY,                   -- 'permissions', 'privacy', 'webhooks', etc
  value TEXT NOT NULL                     -- JSON
);

CREATE TABLE secrets (
  provider TEXT PRIMARY KEY,              -- 'anthropic' | 'openai' | 'openrouter' | 'xai' | 'google'
  kind TEXT NOT NULL,                     -- 'api_key' | 'cli_subscription'
  value TEXT NOT NULL,                    -- API key or path to CLI binary
  meta TEXT,                              -- JSON: { models?: string[] }
  updated_at INTEGER NOT NULL
);

CREATE INDEX idx_chats_status ON chats(status);
CREATE INDEX idx_phase_events_chat ON phase_events(chat_id, phase_idx);
```

DB module exports a typed wrapper: `getDb()`, `chats.create()`, `chats.list()`, `chats.cancel()`, etc. Use Zod schemas for validation at the API boundary.

### 4. Seed (`src/lib/db/seed.ts`)

Reads YAML files from `templates/*.yaml` (Agent D writes these), inserts into `templates` table with `source='builtin'`.

### 5. tmux session manager (`src/daemon/tmux.ts`)

Use `child_process` to call `tmux` shell commands. Sessions named `chorus-<chatId>`. Reaper runs every 5 min via `setInterval` — kill sessions with no output for >10 min OR no associated active chat.

For v0.5, the actual LLM calls are STUBBED — fake a 2-3s delay then emit fake output. Real CLI integration is v0.6. Just establish the tmux lifecycle pattern.

### 6. Error handling

- Every route returns `{ ok: true, data }` or `{ ok: false, error: { code, message } }`.
- Codes: `not_found`, `validation`, `db_error`, `tmux_error`, `internal`.

## Don't touch

- `src/app/`, `src/components/` — Agent C
- `src/mcp/` — Agent B
- `templates/*.yaml` — Agent D

## Acceptance

```bash
cd /home/ubuntu/dev/chorus
pnpm typecheck                                  # passes
pnpm dev:daemon &                               # starts
sleep 2 && curl -s http://127.0.0.1:7707/health # returns {ok:true, ...}
curl -s http://127.0.0.1:7707/templates         # returns the 4 builtins
```
