# Agent B — MCP Server

You own: `src/mcp/`.

## Build

stdio MCP server using `@modelcontextprotocol/sdk`. Exposes 7 tools to orchestrators (Claude Code / Codex / Cursor). All tools call the local daemon REST on `http://127.0.0.1:7707`.

### The 7 tools

| Tool | Input | Output |
|---|---|---|
| `create_chat` | `{ work: string, template?: string, files?: string[] }` (template defaults to `code-review`) | `{ chatId, status, url }` |
| `wait_for_chat` | `{ chatId: string, timeoutSec?: number }` (default 600) | `{ status, verdict?, summary?, blocked? }` — blocks until terminal state |
| `get_chat_status` | `{ chatId: string }` | `{ status, phase, progress, blocked? }` — non-blocking |
| `list_blocked` | `{}` | `{ chats: [{ chatId, work, blockedReason, since }] }` |
| `resume_chat` | `{ chatId: string, answer: string }` | `{ ok: true, status }` |
| `cancel_chat` | `{ chatId: string }` | `{ ok: true }` |
| `list_templates` | `{}` | `{ templates: [{ id, name, description, lineages: [...] }] }` |

`wait_for_chat` polls `/chats/:id/stream` (SSE) — emit progress as MCP `notifications/message` events at each phase transition, and resolve when status flips to terminal (`approved` / `merged` / `blocked` / `cancelled` / `failed`).

### Files

- `src/mcp/index.ts` — main entry, sets up server, registers tools, runs stdio transport
- `src/mcp/tools.ts` — one function per tool, each takes a typed input via Zod schema
- `src/mcp/client.ts` — fetch wrapper around the daemon API (with retries on connection refused: tells user to run `chorus start` first)

### Constraints

- All input validated via Zod before the daemon call.
- All errors translated to `McpError` with stable codes.
- If daemon is down, return a clear error: `"Chorus daemon not running. Run 'chorus start' first."`.
- Keep the file count tight: 3 files, ~600 lines total target.

## Don't touch

- `src/daemon/` — Agent A
- `src/app/`, `src/components/` — Agent C
- `templates/*.yaml` — Agent D

## Acceptance

```bash
cd /home/ubuntu/dev/chorus
pnpm typecheck

# With daemon running on :7707, this MCP request should list 4 templates:
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_templates","arguments":{}}}' | pnpm dev:mcp
```

## Reference

Read the existing prototype's MCP_TOOLS in `src/lib/mock-data.ts` for tool descriptions/copy. The actual implementations are NEW — those mocks are just for the UI Connect page.
