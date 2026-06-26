# Agent C — UI Port (mock-data → daemon API)

You own: `src/app/`, `src/components/`, `src/lib/mock-data.ts` (deleted), `src/lib/api/` (new), `src/lib/types.ts` (new).

## Build

Replace the prototype's static mock data with a typed API client that calls the local daemon on `http://127.0.0.1:7707`. The UI itself stays largely the same — same routes, same components, same look.

### 1. Type contract (`src/lib/types.ts`)

Extract types from the existing `src/lib/mock-data.ts` (Chat, Template, Phase, etc.) and lock them as the shared contract. Daemon JSON responses must conform.

Key types: `Chat`, `Phase`, `PhaseEvent`, `Template`, `TemplatePhase`, `Settings`, `Secret`. Keep field names `camelCase` on the wire (daemon translates from `snake_case` SQLite columns).

### 2. API client (`src/lib/api/`)

```
src/lib/api/
  index.ts        # re-exports
  client.ts       # fetch wrapper, base URL from env or window.location, error envelope handling
  chats.ts        # createChat(), listChats(), getChat(), cancelChat(), resumeChat(), streamChat()
  templates.ts    # listTemplates(), getTemplate(), saveTemplate()
  settings.ts     # getSettings(), updateSettings()
  secrets.ts      # listSecrets(), upsertSecret()
```

Use the native `fetch` (Next.js handles it). Server components use absolute URL `http://127.0.0.1:7707`; client components use relative `/api/proxy/*` if needed (or just CORS to localhost).

For SSE streams (live run page), use `EventSource` in client components.

### 3. Page-by-page swap

| Page | Mock import (current) | Replace with |
|---|---|---|
| `/` | `MOCK_PROJECTS, MOCK_AGENTS, MOCK_RUNS` | `listChats({ limit: 10, status: 'active' })` |
| `/runs` (NEW) | n/a | `listChats({ limit: 50 })` paginated |
| `/runs/[id]` | `MOCK_TASK_RUN, ROUND_2_REVIEWERS, etc` | `getChat(id) + streamChat(id)` SSE |
| `/new` | `MOCK_TEMPLATES` | `listTemplates()` |
| `/templates` | `MOCK_TEMPLATES` | `listTemplates()`, `saveTemplate()` |
| `/settings` | static | `getSettings(), updateSettings(), listSecrets()` |
| `/connect` | `BLOCKED_CHATS` | `GET /blocked` |
| `/onboarding` | static | calls `/secrets` PUT during the auth step |

Drop `/projects` route entirely — deferred to v0.6.

### 4. Live run page

The `/runs/[id]` page has streaming reviewer cards. Use `EventSource` to consume `/chats/:id/stream` SSE. On each `phase_event` server event, append to the local state. Keep all the existing animations + cost meter + Pause/Cancel UI — just change the data source.

While the daemon is producing fake output (Agent A's stub), this still works end-to-end.

### 5. Error UX

If the daemon is down (`fetch` fails / 502 from nginx), show a top banner: *"Chorus daemon not running. Run `chorus start` in a terminal."* with a "Retry" button. Don't break the whole page.

### 6. Brand cleanup

Find any leftover "Murmur" brand text in the UI (titles, sidebar, settings paths like `~/.murmur/`) and replace with "Chorus" / `~/.chorus/`. Real path references in comments stay as is — this is the brand layer only.

## Don't touch

- `src/daemon/` — Agent A
- `src/mcp/` — Agent B
- `templates/*.yaml` — Agent D
- `bin/`, `ecosystem.config.cjs`, `package.json`

## Acceptance

```bash
cd /home/ubuntu/dev/chorus
pnpm typecheck                            # passes
pnpm lint                                 # passes
# With daemon running:
pnpm dev                                  # next dev on :3011
curl -s http://127.0.0.1:3011/ | grep -i "chorus"   # not "murmur"
```

Visit `http://127.0.0.1:3011/` and click through each page — no broken imports, no `MOCK_*` references remaining in `src/lib/`.

## Reference

The existing prototype is the visual ground truth. Don't redesign — just swap data layer.
