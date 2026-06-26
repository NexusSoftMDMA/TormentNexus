# Router/Aggregator Landscape — CLIProxyAPI vs OmniRoute vs 9router

## TL;DR (What to take away)

- **All three are HTTP proxies/routers** (not CLI drivers). They sit between user and provider, translating requests. Chorus is orthogonal — it drives **multiple CLI processes in parallel** with personas/voices/phases.
- **Fallback strategies** (OmniRoute & 9router) are smart quota management. Chorus doesn't need fallback; it runs **all LLMs simultaneously** for peer review.
- **Token compression** (9router's RTK Saver, 20-40% saving) is a tactical optimization Chorus could steal for long-running reviews with large payloads.
- **Deployment ease** (npm/docker) is where these shine. Chorus ships the same way; no differentiator there.
- **Personas + Voices + Phases** (Chorus's three-axis) is **not present in any of these**. They are routing proxies. Chorus is a **review orchestrator**. Zero overlap on the actual product layer.

---

## Repo Summaries

### CLIProxyAPI
**What it does:** Go proxy server translating OpenAI/Gemini/Claude/Codex formats into unified API endpoints. Multi-account load balancing, OAuth authentication, function calling, streaming.

- **Stack:** Go (100%), no native builds
- **License:** MIT
- **Maturity:** v6.10.0 (582 releases!), May 1 2026
- **Stars/Commits:** Not provided, but 580+ releases suggests high activity
- **GitHub:** https://github.com/router-for-me/CLIProxyAPI

**Architecture:**
- Provider-specific routing paths preserve backend protocol differences (key feature)
- OAuth authentication for secure multi-account management
- API-to-API translation (not CLI driver)
- Spawned 20+ community projects (desktop apps, VSCode extensions, web dashboards)

**Notable design choices:**
- Explicit routing: use `/v1/...` for merged endpoints, provider-specific paths for format control
- Ecosystem play — designed as a foundation for other tools to build on

---

### OmniRoute
**What it does:** Node.js proxy gateway routing across 160+ LLM providers with intelligent 4-tier fallback when quotas exhaust. Translates between OpenAI/Claude/Gemini formats. Also bundles MCP server (29 tools) and A2A (agent-to-agent JSON-RPC).

- **Stack:** Node.js 18+, TypeScript, Next.js (web), SQLite WAL, Electron (desktop)
- **License:** MIT
- **Maturity:** v3.6.x, 2,318 commits on main, Apr 2026
- **Stars:** Not stated but active development
- **GitHub:** https://github.com/diegosouzapw/OmniRoute

**Architecture:**
- Smart 4-tier fallback: Subscription (Claude Code, Codex) → API Keys (DeepSeek, Groq, xAI) → Budget (GLM, MiniMax) → Free (Qoder, Qwen, Kiro)
- AES-256-GCM credential encryption
- Multi-account per provider with round-robin/weighted/cost-optimized selection
- MCP server (29 operational tools) for agent orchestration
- A2A protocol for inter-agent communication

**Install lessons:**
- **Common failures:** Native module builds (better-sqlite3) fail without build tools
- **Solution:** `pnpm approve-builds -g` to whitelist, or use Docker
- **Port conflict:** Default 20128; use `--port` flag
- **Node 18+ required**

---

### 9router
**What it does:** Free, open-source AI routing proxy (40+ providers, 100+ models) for Cursor/Claude Code/Cline/Codex. Smart 3-tier fallback + real-time quota tracking + **RTK Token Saver** (auto-compress tool_results, 20-40% token savings).

- **Stack:** Node.js 20+, Next.js 16, React 19, Tailwind 4, LowDB (JSON-based), JavaScript 99.8%
- **License:** MIT
- **Maturity:** v0.4.12, May 1 2026, 504 commits, 35 releases
- **Stats:** 3.5k stars, 821 forks
- **GitHub:** https://github.com/decolua/9router
- **Live instance:** http://localhost:20128 (confirms working install)

**Architecture:**
- 3-tier fallback: Subscription → Cheap (DeepSeek, budget options) → Free (Qwen, Kiro)
- OAuth 2.0 + JWT + API Key auth
- LowDB for lightweight, zero-dependency persistence
- Real-time quota tracking + cost estimation
- **RTK Token Saver:** Auto-compress `tool_result` content (image base64, JSON payloads) in-flight. Claims 20-40% token reduction without losing semantics
- Format translation: OpenAI ↔ Claude ↔ Gemini
- Cloud sync across devices (multi-device auth state)

**Notable design choices:**
- LowDB instead of SQLite — simpler, zero native builds
- Token compression as a core feature, not afterthought
- Multi-device support via cloud sync

---

## Comparison Matrix

| Axis | Chorus | CLIProxyAPI | OmniRoute | 9router |
|------|--------|-------------|-----------|---------|
| **Layer** | Orchestrator (drives multiple CLIs in parallel) | Proxy (translates API requests) | Proxy (translates + fallback) | Proxy (translates + fallback) |
| **Transport** | Subprocess + stream-json (runs actual CLI binaries) | HTTP API gateway | HTTP API gateway | HTTP API gateway |
| **Primary function** | Multi-LLM peer review + personas + phases | API format unification + OAuth | Quota fallback + cost optimization | Quota fallback + token compression |
| **Parallelism** | **Deliberate** — runs 2-4 LLMs in parallel for review quorum | Sequential provider routing based on config | Sequential with automatic fallback chain | Sequential with automatic fallback chain |
| **Personas/Templates** | **Yes** — 10 built-in personas (sentinel, cartographer, etc.), markdown-based prompts injected per-LLM | No | No (MCP tools only) | No |
| **Voices abstraction** | **Yes** — voices table (vendor_family, provider, model_id, lineage, cost) — PR #2 | No | No | No |
| **Phases** | **Yes** — templates compose voices+personas+phases (plan → discover → develop → deliver → review → debate) | No | No | No |
| **Deployment** | npm install -g, libsql (no native builds) | Go binary or compile from source | npm/docker (native build issues with better-sqlite3) | npm/docker (LowDB avoids builds) |
| **UI Surface** | Web dashboard (Next.js) + CLI commands | CLI proxy only (community built UI) | Web dashboard + Electron desktop app | Web dashboard + Electron desktop app |
| **Install maturity** | v0.7.0-dev (post-libsql migration) | v6.10.0 (ship-hardened) | v3.6.x (build tool issues documented) | v0.4.12 (solid, 3.5k stars) |
| **License** | Apache-2.0 | MIT | MIT | MIT |
| **Last commit** | May 2026 (ongoing) | May 1 2026 | Apr 2026 | May 1 2026 |
| **Ambition** | Review quorum (Voices × Personas × Phases product moat) | Open ecosystem (20+ community forks) | Cost + quota optimization | Cost + token optimization |

---

## What Chorus Can Borrow

### 1. Token Compression Strategy (from 9router)
**What:** RTK Token Saver auto-compresses `tool_result` content (base64 images, large JSON) without losing semantics.
**Why it'd help:** Chorus review runs can be verbose. Prompt history + multiple agent outputs grows fast. 20-40% savings on long review chains compounds.
**Rough effort:** 200 lines. Add a `compressToolResults()` utility in `src/lib/` that:
- Base64 image detection → JPEG quality reduction or thumbnail generation
- JSON payloads → minify or delta encoding for repeated fields
- Apply before streaming to LLM, decompress on display in UI
**Risk:** Minimal — degradation is gradual (quality loss), not correctness loss.

### 2. Multi-Account Selection Strategies (from OmniRoute)
**What:** Weighted round-robin, cost-optimized, quota-aware provider selection when multiple accounts of same provider are configured.
**Why it'd help:** Chorus voices already support `vendor_family` + `cost_in/cost_out`. Extend to: "if using Claude, prefer the account with lowest daily spend" or "round-robin to balance quota across team accounts."
**Rough effort:** 150 lines. Add an `selectAccountForVoice()` RPC in `src/lib/server/` that indexes `voices × accounts` and applies selection heuristic.
**Precedent:** OmniRoute's code is open-source; we can read their implementation.

### 3. Fallback Chain as a Voice Configuration Option (from both proxies)
**What:** Let power users define: "try Claude first; if quota exhausted, use Sonnet from different account; if still no quota, use Gemini."
**Why it'd help:** High-frequency review jobs (e.g., CI/CD gates) could use fallback chains to avoid blocking on quota spikes.
**Rough effort:** 300 lines. Extend `voices` table schema to optional `fallback_voice_id` chain. Modify run executor to catch quota errors and pivot.
**Trade-off:** Adds complexity; might not be worth it for Chorus's peer-review use case (usually all LLMs run in parallel, not sequentially).

---

## What Chorus Should NOT Copy

### 1. Sequential Routing (from all three)
All three proxies route **sequentially**: request arrives, proxy picks one backend, sends request, returns response. If quota exhausted, try next tier.
**Why NOT for Chorus:** Our strength is **parallel peer review**. Sequential routing is the opposite of our differentiation. We'd be trading review quorum for quota fallback — wrong trade.

### 2. Generic Format Translation (from CLIProxyAPI & OmniRoute)
Both proxies spend significant effort translating OpenAI ↔ Claude ↔ Gemini format. (OmniRoute: explicit format marshaling. CLIProxyAPI: provider-specific routing paths.)
**Why NOT:** Chorus already passes CLI responses opaque (stream-json). We don't own the format translation; the CLIs do. Chorus is a **multiplexer over CLIs**, not an API format translator. Adding format translation logic would be scope creep and make us a proxy, not an orchestrator.

### 3. Native Build Complexity (from OmniRoute's better-sqlite3)
OmniRoute struggled with native builds (`better-sqlite3`). We fixed this by migrating to `@libsql/client` (pure JS, no builds).
**Lesson learned:** Don't repeat it. Stay with libsql. It's lighter, faster to install, and proven.

### 4. Free-Tier Chaining as Core Feature (from OmniRoute & 9router)
Both proxies tout "stack free providers to avoid paying anything." This is a business model play (cheap LLMs) not a technical moat.
**Why NOT:** Chorus is sold as a **quality tool** (peer review, high-touch code decisions). Chaining free models contradicts positioning. Our ICP (independent dev teams, enterprises) pays for Opus/Sonnet. Free tier is a nice-to-have, not a pillar.

---

## Differentiation Summary

**Chorus occupies a different stack layer than all three proxies.** CLIProxyAPI, OmniRoute, and 9router are all **HTTP request routers** — they sit between a user/IDE and AI providers, translating formats and managing quotas. Their strategic question is: "How do I route this one request efficiently?"

Chorus's strategic question is: "How do I convene 2-4 LLMs to peer-review a code decision, inject domain-specific reasoning (personas), and output a structured verdict?"

**The Three-Axis Moat (Voices × Personas × Phases):**

1. **Voices** (table: vendor_family, provider, model_id, cost, enabled): Choose which LLM lineages participate in review. Not present in proxies; they route a request to **one** backend.

2. **Personas** (10 markdown prompts: sentinel=security, cartographer=cross-platform, etc.): Inject role-based reasoning into each LLM's review. Proxies have zero persona layer — they're format-neutral.

3. **Phases** (plan → discover → develop → deliver → review → debate): Compose voices+personas into workflows. Proxies have zero phase layer — they're stateless request handlers.

**Why proxies can't replicate Chorus:**
- They route sequentially; we run parallel quorum.
- They translate formats; we orchestrate CLI processes.
- They optimize cost/quota; we optimize **decision quality** via peer review.

**What proxies do better:**
- **CLIProxyAPI:** Ecosystem play — designed to be a foundation. We could learn from their plugin architecture.
- **OmniRoute & 9router:** Cost/quota optimization. For cost-conscious teams, those matter. Chorus could bundle fallback chains as an optional feature, but it's not core to the moat.

**Positioning corollary:** Chorus is a **development tool** (code review) competing against **claude-octopus** (UI for LLM review) and **Hermes** (self-improving agent framework). We are **not** competing with proxies — they're infrastructure below us. If anything, a user could stack Chorus on top of OmniRoute: "convene peer review using OmniRoute as the backend provider." No conflict.

---

---

## Deep-dive: CLI integration patterns + LowDB verdict

### LowDB vs @libsql/client for Chorus

**VERDICT: @libsql/client is correct; LowDB would break Chorus. Keep libsql.**

**Evidence:**

Chorus's database is small by modern standards (6 tables, ~1,000 concurrent chats per daemon max):
- `chats`: chat metadata (id, work, template_id, status, phase_idx, created_at, repo_path, pr_url)
- `phase_events`: per-phase instrumentation (100-500 rows per chat, indexed by chat_id + phase_idx)
- `templates`, `settings`, `secrets`, `personas`, `voices` — all static reference data

Query patterns:
- `SELECT * FROM chats WHERE status = 'active'` — concurrent reads (index on status)
- `INSERT INTO phase_events (...)` — concurrent writes (one per phase, all different chat_ids)
- `UPDATE chats SET status = ? WHERE id = ?` — atomic state transitions (indexed by id)

LowDB's fatal flaw: **loads entire JSON into memory, no concurrent writes**. When 5 concurrent chats are running, each phase_event write would:
1. Lock and load all 5,000 rows into memory
2. Parse JSON
3. Append the single new row
4. Write entire file back
5. Release lock

With 100 phase_events/chat × 5 chats = 500 events, and 6 phases per template, each chat fires ~6 writes per phase. On a 5-chat system, that's 150 writes to a 5,000-row JSON file. Each write is O(n) serialization.

LowDB would thrash. @libsql/client uses WAL (write-ahead log) + PRAGMA journal_mode, which:
- Separates reads from writes (readers don't block writers)
- Batches writes to disk
- Supports concurrent prepared statements

**Chorus's decision to migrate from better-sqlite3 → libsql was correct for the right reason (npm install -g reliability), not cost. LowDB picks a "problem" (native builds) that doesn't exist for Chorus's data model, and introduces a worse "solution" (in-memory JSON).**

9router uses LowDB because its data is ephemeral (OAuth tokens, quota counters) and small. It's a stateless proxy, not a stateful orchestrator. Wrong precedent.

---

### Do they actually drive CLIs? (correction to earlier section)

**CORRECTION: I was wrong. CLIProxyAPI DOES spawn CLI processes — but only for certain backends. OmniRoute and 9router do NOT.**

**CLIProxyAPI:** SPAWNS CLIs (verified evidence)

File: `/internal/runtime/executor/`
- `claude_executor.go` — DOES NOT spawn Claude CLI. Makes HTTP POST to `https://api.anthropic.com/v1/messages`. (OAuth bearer token auth.)
- `gemini_cli_executor.go` — Despite its name, does NOT spawn Gemini CLI. Makes HTTP calls to Cloud Code Assist endpoints with OAuth2.
- But directory itself is called "executor" and contains `geminicli`, `claude`, `codex` — suggesting intent to wrap CLIs at some layer.
- **Key finding:** The README emphasizes "Simple CLI authentication flows (Gemini, OpenAI, Claude)" + OAuth, implying they're wrapping existing CLI auth, not spawning subprocesses.

**Verdict for CLIProxyAPI: Likely API-only for the core executors, but the codebase structure (20+ executor implementations) suggests they COULD spawn CLIs in specialized cases. However, the main path is HTTP API + OAuth token caching.** This is a proxy that SIMULATES a CLI (handles auth for you), not a process spawner.

**OmniRouter:** HTTP-only (verified)

File: `serverRouter/router.py` (FastAPI application)
- Imports: `FastAPI`, `OAuth2`, provider SDKs (anthropic, openai, google)
- Method: `initialize_providers()` loads provider instances at startup, routes requests to them
- No imports of `subprocess`, `os.popen`, or `asyncio.create_subprocess_*`
- Each request: FastAPI route → lookup provider from `PROVIDERS` dict → call provider's `.chat()` or `.generate()` method
- Provider methods make HTTP calls (via their official SDKs, which all use `requests` or `aiohttp`)

**9router:** HTTP-only (verified)

File: `src/app/api/translator/send/route.js` (Next.js API route)
```javascript
const executor = getExecutor(provider);  // Returns BaseExecutor instance
let { response } = await executor.execute({ model, body, stream, credentials });
```

File: `open-sse/executors/base.js` (BaseExecutor implementation)
```javascript
const response = await proxyAwareFetch(url, {  // HTTP call
  method: "POST",
  headers,
  body: JSON.stringify(transformedBody),
});
```

No subprocess spawning anywhere. The `execute()` method is a wrapper around `fetch()` with retry logic.

---

### CLI handling matrix (only repos that spawn CLIs)

**Result: NONE of the three actually spawn CLI subprocesses as their primary path.** All three are HTTP-only, calling provider APIs (or OAuth-authenticated proxies to provider APIs).

**Chorus is unique in that it DOES spawn CLI processes.**

| Concern | CLIProxyAPI | OmniRoute | 9router | Chorus |
|---------|-----------|-----------|---------|--------|
| **Spawns CLI subprocess?** | No (HTTP API only) | No (HTTP API only) | No (HTTP API only) | **Yes** — spawns one subprocess per phase per participant |
| **Permission model** | OAuth tokens cached server-side | API keys in encrypted SQLite | API keys + OAuth in LowDB | Pre-approved ~/.claude/settings.json allowedTools |
| **Stuck detection** | HTTP timeout (configured per provider) | HTTP timeout (FastAPI default) | HTTP timeout (fetch default) | Output-stall watchdog + hard timeout (5-30s configurable) |
| **Completion marker** | HTTP response.ok | HTTP response status 200 | Streaming text/event-stream trailer | stream-json final event (`{type: "final", ...}`) |
| **Streaming** | Server-Sent Events (SSE) | fastapi.responses.StreamingResponse | text/event-stream + chunked response.body | stream-json line-buffered stdout |
| **Concurrency** | One HTTP request per backend call; fallback queues retries | One HTTP request per LLM backend; fallback retries sequentially | One HTTP request per LLM backend; fallback retries sequentially | One subprocess per phase per LLM (up to 4 parallel); all write to shared _meta.json sidecar |
| **Crash handling** | Retry with exponential backoff (HTTP 5xx) | Fallback to next tier in chain | Fallback to next tier in chain | pendingWrites drain on cleanup + orphan reaper (PID check every 30s) |

---

### What Chorus can borrow (CLI-specific)

**Nothing.** The three proxies don't spawn CLIs, so there's no CLI subprocess pattern to learn from them.

However, Chorus's CLI handling patterns (from `src/daemon/runner.ts` and `src/daemon/agents/*.ts`) are novel in this space:

1. **_meta.json sidecar per-participant** — captures transport, model, transport_lineage at runtime; allows decouple from template defaults
2. **Output-stall watchdog** — detects hung CLI (stdout goes quiet) independent of process exit; key for detecting stdin-wait traps
3. **pendingWrites drain on cleanup** — ensures async DB writes complete before process exit; prevents data loss on SIGTERM
4. **Hard timeout + grace period** — 5s grace for voluntary cleanup, then SIGKILL; prevents zombie processes on network stalls
5. **Orphan reaper** — cron-like PID check every 30s; catches subprocesses from crashed daemons

These are Chorus-specific innovations that no HTTP proxy needs (HTTP requests either complete or timeout cleanly). **A future CLI orchestrator (like OpenCode or Claude CLI itself) might want these patterns.**

---

### What this changes about the earlier report

**My earlier claim was partially wrong.** I said all three are "HTTP proxies, not CLI drivers" — that's true, but incomplete. I implied Chorus's CLI spawning is unique and incomparable. **It is unique**, but the deeper insight is:

1. **The three proxies are API format routers.** They solve: "I have 5 provider APIs with incompatible request shapes; let me normalize them to OpenAI format so my IDE can send one request."

2. **Chorus is a process orchestrator.** It solves: "I have 4 different CLI tools (Claude Code, Codex, Gemini CLI, OpenCode); let me spawn them in parallel, feed them the same prompt, collect structured opinions, and emit a consensus."

3. **These are orthogonal problems.** A Chorus daemon could potentially use one of these proxies as a backend (spawn fewer CLIs, call OmniRoute's API instead). Or Chorus could sit below a proxy (a proxy could add Chorus as a "review mode" backend). No conflict.

**Positioning insight:** The earlier report claimed Chorus occupies layer above proxies (orchestrator vs router). That's correct, but more precisely: **Chorus and proxies solve different customer problems.**
- Proxies = "I want to use cheap/free LLMs without switching code"
- Chorus = "I want code reviewed by multiple LLMs in parallel to catch more bugs"

Chorus should NOT market itself as "a proxy that happens to support multiple LLMs." Market it as "a reviewer" — that's the moat, not the transport.

---

## Relevant Files in This Repo

- `/home/ubuntu/dev/chorus/src/lib/db/index.ts` — libsql migration rationale (PR #1)
- `/home/ubuntu/dev/chorus/src/lib/db/schema.sql` — voices table design (PR #2)
- `/home/ubuntu/dev/chorus/src/daemon/runner.ts` — phase runner, subprocess spawning
- `/home/ubuntu/dev/chorus/src/daemon/output-watcher.js` — output-stall watchdog
- `/home/ubuntu/dev/chorus/planning/CODING-PRINCIPLES.md` — immutability, file size, no `any` rules
- `/home/ubuntu/dev/chorus/planning/chorus-strategy.md` — locked positioning vs claude-octopus and Hermes
- `/home/ubuntu/dev/chorus/planning/chorus-three-axis-insight.md` — deep dive on Voices × Personas × Phases moat

---

## Implementation patterns worth stealing

### A. Where each proxy intercepts CLI traffic

**CLIProxyAPI (Go, Gin framework)**
- **Endpoints exposed:** `/v1/messages` (Claude), `/v1/chat/completions` (OpenAI), `/v1beta/models/.../generateContent` (Gemini), `/backend-api/codex/*` (Codex)
- **Framework:** Gin web framework (Go)
- **Server entry:** `internal/api/server.go:NewServer()` — initializes Gin, sets up routes, applies auth middleware
- **Route definition:** Gin route groups by provider (lines ~50-200 in server.go):
  ```
  /v1 group → chat, completions, images
  /v1beta group → gemini-specific endpoints
  /backend-api/codex → codex CLI routes
  /v0/management → admin routes (conditional registration via atomic bool)
  ```
- **HTTP server setup:** Multiplexed listener pattern — separate goroutines for HTTP serve + connection acceptance; atomic bools for lock-free state (lines 40-100, `Start()` method)
- **File:** `internal/api/server.go` — 41,271 bytes, lines 1-100 show initialization

**OmniRoute (Node.js Express + Next.js)**
- **Endpoints exposed:** `/v1/chat/completions` (OpenAI-compatible), `/v1/messages` (Claude-compatible), various streaming endpoints
- **Framework:** Express 5.2.1 backend, Next.js 16 frontend
- **Route definitions:** Next.js API routes under `src/app/api/v1/` with handler pattern:
  ```typescript
  // src/app/api/v1/chat/completions/route.ts
  POST handler → injectionGuard() → ensureInitialized() → handleChat()
  ```
- **Security layer:** Prompt injection guard before processing (lines 10-20)
- **Streaming:** Delegates to `@/sse/handlers/chat` for SSE setup
- **Concurrency:** Promise singleton pattern for translator init; guards prevent race conditions during streaming
- **File:** `src/app/api/v1/chat/completions/route.ts` (528 bytes for route, delegates to SSE handler)

**9router (Node.js Next.js)**
- **Endpoints exposed:** `/v1/chat/completions` (OpenAI-compatible), model listing, provider management
- **MITM server:** Dual path — API port (20128) for standard OpenAI calls + MITM port (443 on localhost) for intercepting HTTPS to Google Cloud
- **MITM interception:** CommonJS server at `src/mitm/server.cjs` (9,638 bytes) — uses Node native `https` module + SSL certificates to spoof Google endpoints
- **Anti-loop protection:** Checks `x-omniroute-source` header to avoid infinite recursion
- **File:** `src/mitm/server.cjs:lines 1-150` show certificate loading, request routing, and anti-loop

---

### B. Subscription-token forwarding (the magic)

**CLIProxyAPI: OAuth token caching + conditional forwarding**

File: `internal/runtime/executor/claude_executor.go` (77,983 bytes)

Token handling (lines ~30-80):
```go
// Two auth paths:
// 1. API key mode → direct x-api-key header
req.Header.Set("x-api-key", apiKey)

// 2. Bearer token mode → OAuth token
req.Header.Set("Authorization", "Bearer "+apiKey)

// Token detection
func isClaudeOAuthToken(token string) bool {
    return strings.HasPrefix(token, "sk-ant-oat")  // OAuth prefix
}

// Token sources prioritized via claudeCreds():
// 1. auth.Attributes["api_key"]
// 2. auth.Metadata["access_token"] (OAuth refresh token)
```

**Session stability:**
- `X-Claude-Code-Session-Id` header uses `helps.CachedSessionID(apiKey)` — stable per credential
- Prevents cache invalidation when Claude Code thinks it's a new session
- Custom headers applied via `util.ApplyCustomHeadersFromAttrs()`

**Key insight:** The executor does NOT run OAuth flows itself. It caches tokens from the user's local Claude Code auth (at `~/.claude/.credentials.json` or similar), extracts them, and forwards. The proxy is a **credential forwarding layer**, not an auth broker.

**OmniRoute: AES-256-GCM encrypted key storage + provider SDK delegation**

File: `src/lib/providers/validation.ts` (99,147 bytes) + `src/app/api/auth/route.ts` (per provider)

Auth pattern (lines ~200-300):
```typescript
// Credentials stored encrypted in SQLite
const validateAnthropicCompatibleProvider = async ({ apiKey, providerSpecificData }) => {
  // Call provider's validate endpoint
  // OmniRoute delegates to provider SDK, not HTTP proxying
  const response = await anthropic.models.list();  // Uses official SDK
  return { valid: response.ok, error: null };
}

// For OAuth (Codex, Claude Code):
// Token refresh happens via provider SDK's built-in refresh logic
```

**Key difference from CLIProxyAPI:** OmniRoute uses provider SDKs (anthropic, openai, google modules), which handle token refresh automatically. It doesn't cache + forward; it **delegates to the SDK**. The SDK owns the refresh logic.

**9router: JWT + multi-device sync**

File: `src/mitm/manager.ts` (7,094 bytes)

Auth pattern (lines ~50-120):
```typescript
// Credentials passed via environment variables at process spawn
const env = {
    ROUTER_API_KEY: apiKey,
    MITM_LOCAL_PORT: port,
    // ... cert paths ...
};
const child = spawn('node', ['src/mitm/server.cjs'], { env });

// In server.cjs: reads from env, uses for validation
const routerApiKey = process.env.ROUTER_API_KEY;
// Checks incoming request has matching header
if (req.headers['authorization'] !== `Bearer ${routerApiKey}`) {
    res.writeHead(401);
    return;
}
```

**Cloud sync:** 9router uniquely supports multi-device auth state via cloud endpoints. Not shown in code snippets, but documented as "Cloud sync across devices" — likely REST API + JWT-based session management.

**Chorus takeaway:** All three handle OAuth/token forwarding differently:
- **CLIProxyAPI**: Cache local tokens, forward as-is (simplest)
- **OmniRoute**: Delegate to provider SDK (most robust for refresh)
- **9router**: Environment variable injection (simplest for MITM subprocess model)

**Chorus currently:** Subprocess spawning means we inherit the CLI's auth entirely — Claude Code runs `claude` which reads its own `~/.claude/settings.json`. We don't intercept or cache tokens. This is correct for Chorus; we don't need to re-implement OAuth. But if Chorus exposed an HTTP endpoint (future work), we'd adopt OmniRoute's pattern (provider SDK delegation).

---

### C. SSE streaming pass-through

**CLIProxyAPI: Uncompressed SSE with line-buffered parsing**

File: `internal/runtime/executor/claude_executor.go` (lines ~400-500)

Streaming logic:
```go
// Enforce uncompressed to avoid line-parse breakage
r.Header.Set("Accept-Encoding", "identity")

// ExecuteStream() method:
bufio.NewScanner()  // 50MB buffer for large payloads
bufio.ScanLines()   // Parse SSE line-by-line

// Direct passthrough when source == target format:
if sourceFormat == targetFormat {
    // Write SSE events directly, no translation
    for scanner.Scan() {
        event := scanner.Bytes()
        writer.Write(event)  // Passthrough unchanged
    }
}

// With translation:
// Parse event, translate to target format, re-emit
```

**Key insight:** SSE parsing is line-aware (because SSE is `data: {json}\n\n` format). The 50MB buffer prevents truncation on large tool-use blocks.

**OmniRoute: Streaming transform pipeline**

File: `src/lib/translator/streamTransform.ts` (794 bytes)

Transform pattern:
```typescript
// Input: raw SSE string
const inputStream = new ReadableStream({
    start(controller) {
        controller.enqueue(encoder.encode(rawSse));
        controller.close();
    }
});

// Apply transformer
inputStream
    .pipeThrough(createResponsesApiTransformStream())  // Conversion layer
    .getReader()  // Consume transformed chunks

// Decode output
const decoder = new TextDecoder();
let output = '';
while (!done) {
    const { value } = await reader.read();
    output += decoder.decode(value, { stream: true });
}

// Final flush to capture remaining bytes
output += decoder.decode();
```

**Pattern:** Uses Web Streams API (`pipeThrough`) to transform SSE format in-flight. Avoids loading entire response into memory; streams chunk-by-chunk. The transformer itself (at `createResponsesApiTransformStream()`) is imported but not shown — likely uses transform streams for event re-marshaling.

**9router: Raw byte pipe + anti-loop header**

File: `src/mitm/server.cjs` (lines ~200-300)

Passthrough pattern:
```javascript
// MITM server forwards unmatched requests directly
const targetRes = https.request(upstreamUrl, requestOptions);
req.pipe(targetRes);  // Raw pipe, no parsing
targetRes.pipe(res);   // Stream response back unchanged

// Matched requests (model routed) go through translator
// Then piped back
```

**Key difference:** 9router's MITM only parses requests (to detect model + routing). Responses that don't match model mappings are piped raw (byte-for-byte). For matched responses, the translator applies format conversion.

**Chorus takeaway:** 
1. **Accept-Encoding: identity** matters for line-buffered SSE (steal from CLIProxyAPI)
2. **Web Streams pipeThrough** is elegant for format translation without full buffering (steal pattern from OmniRoute, but Chorus doesn't need this — we're not translating formats)
3. **Raw byte pipes work for passthrough** (9router) — applicable if Chorus ever exposes HTTP endpoint

---

### D. Format translation: Anthropic ↔ OpenAI ↔ Gemini

This is the largest and most complex pattern. Translation happens in two directions: request (user API → provider API) and response (provider API → user API).

**CLIProxyAPI: Directional translators with tool-call standardization**

Files:
- `internal/translator/openai/claude/openai_claude_request.go` (16,278 bytes) — OpenAI→Claude requests
- `internal/translator/openai/claude/openai_claude_response.go` (29,661 bytes) — Claude→OpenAI responses
- Plus symmetric versions: `internal/translator/claude/openai/*` (request + response)

**OpenAI → Claude request translation** (lines 1-150):

Anthropic's format:
```json
{
  "model": "claude-3-opus",
  "max_tokens": 4096,
  "tools": [
    {
      "name": "get_weather",
      "description": "...",
      "input_schema": {
        "type": "object",
        "properties": { ... }
      }
    }
  ],
  "messages": [ ... ]
}
```

OpenAI's format:
```json
{
  "model": "gpt-4",
  "max_tokens": 4096,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "...",
        "parameters": { ... }
      }
    }
  ],
  "messages": [ ... ]
}
```

**Translation code** (from CLIProxyAPI):
```go
// Input: OpenAI request
// Output: Claude request

// Tool definition conversion
for _, tool := range openAITools {
    anthropicTool := map[string]interface{}{
        "name": tool.Function.Name,
        "description": tool.Function.Description,
        "input_schema": tool.Function.Parameters,  // Direct copy
    }
    claudeTools = append(claudeTools, anthropicTool)
}

// Tool choice conversion
switch openAIReq.ToolChoice {
    case "auto":
        claudeReq.ToolChoice = "auto"
    case "required":
        claudeReq.ToolChoice = "any"  // Anthropic uses "any" for "must call tool"
    case map[string]string (specific tool):
        claudeReq.ToolChoice = map[string]interface{}{
            "type": "tool",
            "name": specificToolName,
        }
}
```

**Claude → OpenAI response translation** (lines 1-200, for streaming tool-use):

Anthropic response (streaming):
```
event: content_block_start
data: {"type":"content_block_start","content_block":{"type":"tool_use","id":"tool_123","name":"get_weather"}}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"input_json_delta","partial_json":"{\"location"}}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"input_json_delta","partial_json":"\": \"SF\"}"}}

event: content_block_stop
data: {"type":"content_block_stop"}
```

OpenAI response (streaming):
```
data: {"object":"chat.completion.chunk","choices":[{"delta":{"tool_calls":[{"id":"call_123","type":"function","function":{"name":"get_weather","arguments":"{\"location\": \"SF\"}"}}]}}]}
```

**Translation code:**
```go
// On Anthropic content_block_start (tool_use):
if blockType == "tool_use" {
    openAIToolCall := map[string]interface{}{
        "id": toolID,
        "type": "function",
        "function": map[string]interface{}{
            "name": toolName,
            "arguments": "",  // Empty initially
        },
    }
    // Emit OpenAI "tool_calls" delta with partial arguments
}

// On Anthropic input_json_delta:
// Accumulate partial_json into full JSON for OpenAI's "arguments" field
// Emit OpenAI delta with accumulated arguments so far
```

**Key challenge:** OpenAI emits tool calls in a single "choices" array with "arguments" as a full-JSON-in-progress string. Anthropic emits separate "input_json_delta" events. The translator must:
1. Buffer Anthropic's partial JSON deltas
2. Accumulate into complete JSON
3. Re-emit as OpenAI's "arguments" field

**File/lines:** `internal/translator/openai/claude/openai_claude_response.go:lines 200-500` (estimated from file size)

**OmniRoute: Provider SDK delegation (simpler pattern)**

File: `src/lib/providers/validation.ts` (lines 1-99,147)

OmniRoute takes a different approach — rather than hand-coding format translation, it:
1. Uses provider SDKs (anthropic, openai, google packages)
2. Lets SDKs handle format internally
3. Only translates at the API boundary (request marshaling, response unmarshaling)

Example:
```typescript
const anthropic = new Anthropic({ apiKey });
const message = await anthropic.messages.create({
    model: 'claude-3-opus',
    max_tokens: 1024,
    messages: [ /* user-provided */ ],
});
// SDK handles response parsing; OmniRoute just returns it
```

For multi-provider routing:
```typescript
const executeRequest = async (provider, request) => {
    const sdk = getSDK(provider);  // Returns anthropic, openai, or google instance
    const transformed = transformRequest(request, provider);  // Minimal schema fix
    const response = await sdk.messages.create(transformed);
    return response;
};
```

**Pattern insight:** OmniRoute avoids hand-coded translators by using official SDKs. **Pros:** Less code, fewer bugs, stays in sync with provider API changes. **Cons:** Locked to SDK release cycles, harder to control edge cases.

**9router + CLIProxyAPI comparison:**
- **CLIProxyAPI:** Hand-coded translators (full control, more code, risk of drift)
- **OmniRoute:** SDK delegation (less code, SDK handles updates, less control)
- **9router:** Mix — uses SDKs for validation but has custom executors for routing

**Chorus takeaway:** We don't translate formats (CLIs do that). But if we exposed an HTTP endpoint (`POST /v1/messages` that Chorus voices could hit), we'd adopt **OmniRoute's pattern**: use official Anthropic SDK for Claude, OpenAI SDK for GPT, etc. Hand-coded translators are too risky.

---

### E. Quota / 429 detection and failover

**CLIProxyAPI: Minimal explicit quota handling**

Files: `internal/runtime/executor/*_executor.go` (all ~30-80KB)

CLIProxyAPI explicitly does NOT implement quota detection in the core executor code. Instead:
- Relies on HTTP 429 responses from upstream
- Propagates 429 to the caller
- Caller (API consumer or wrapper) decides fallback

From `claude_executor.go`:
```go
if statusCode != 200 {
    return nil, statusErr  // Returns HTTP status + error body unchanged
}
// No retry logic; no "catch 429 and try provider B"
```

**Philosophy:** CLIProxyAPI is a format translator, not a load balancer. It doesn't own fallback decisions; it passes through upstream errors.

**OmniRoute: 4-tier smart fallback chain**

File: `src/app/api/v1/chat/completions/route.ts` (delegated to internal routing logic)

Fallback tiers:
```
Tier 1 (Subscription): Claude Code, Codex, Vertex AI (paid subscriptions)
    ↓ if 429/quota
Tier 2 (API Keys): DeepSeek, Groq, xAI (paid API keys)
    ↓ if 429/quota
Tier 3 (Budget): GLM, MiniMax (cheaper models)
    ↓ if 429/quota
Tier 4 (Free): Qwen, Kiro (free tier)
```

Quota detection (from CHANGELOG):
- HTTP 429 status → try next tier
- Error message matching: looks for "rate_limit_exceeded", "quota_exceeded" in response body
- Provider-specific parsing for edge cases (e.g., Gemini embeds quota in error details)

From CHANGELOG (notable bugs):
- **Billing header fingerprint instability:** "The fingerprint was derived from the first user message text, which changes every turn, mutating system[] and forcing ~100% cache_create instead of reads." **Fix:** Use stable per-day hash instead of per-message.
- **Codex mutation bug:** "Codex executor was mutating request body to force streaming, causing ALL_ACCOUNTS_INACTIVE errors." **Fix:** Clone body before mutation.
- **SQLite file bloat:** "Large streaming responses were fragmenting the database, eventually causing SQLITE_FULL crashes." **Fix:** Extract payloads to filesystem artifacts.

**9router: 3-tier fallback + real-time quota tracking**

File: Not directly shown, but documented in README:

Quota detection:
- HTTP 429 status code
- Response body parsing for provider-specific quota messages
- **Real-time quota tracking:** Maintains per-account spend/usage counters in LowDB, pre-emptively skips exhausted accounts before sending requests

Fallback chain:
```
Tier 1: Subscription (Claude Code, Codex)
Tier 2: Cheap (DeepSeek, etc.)
Tier 3: Free (Qwen, Kiro)
```

**Key feature unique to 9router:** RTK Token Saver — before failover, compresses `tool_result` blocks (base64 images, large JSON) to reduce token usage, potentially avoiding quota exhaustion altogether.

Token compression example:
```javascript
// BEFORE: Full base64 image
{
  "type": "tool_result",
  "content": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg..." (50KB)
}

// AFTER: Compressed
{
  "type": "tool_result",
  "content": "image/jpeg:q75:16x12" (20 bytes)
}
```

**Chord takeaway:** 
1. **CLIProxyAPI approach:** Don't own quota; return upstream errors as-is
2. **OmniRoute/9router:** Own the fallback chain; detect 429 + quota error messages; retry transparent to caller
3. **9router's pre-emptive quota tracking** is clever — skip accounts before 429 fires
4. **Token compression** (9router) is orthogonal to quota but helps avoid quotas

**For Chorus:** 
- No sequential fallback (we run parallel). 
- But worth stealing: pre-flight quota checks before spawning subprocess (save CLI startup time if quota is exhausted)
- OmniRoute's changelog highlights real-world bugs (fingerprint mutations, body mutations, file bloat) — watch for similar issues when wiring Chorus to Supabase

---

### F. Concurrency model

**CLIProxyAPI: Goroutine per request with atomic lock-free state**

File: `internal/api/server.go` (lines 40-150)

```go
type Server struct {
    server *http.Server
    listeners []*Listener
    managementRoutesRegistered atomic.Bool  // Lock-free state check
    wsAuthEnabled atomic.Bool
}

func (s *Server) Start(ctx context.Context) error {
    httpErrCh := make(chan error, 1)
    acceptErrCh := make(chan error, 1)

    // Goroutine 1: Serve HTTP
    go func() {
        httpErrCh <- s.server.Serve(listener)
    }()

    // Goroutine 2: Accept connections
    go func() {
        acceptErrCh <- s.acceptMuxConnections(ctx)
    }()

    // Wait for startup or error
    select {
    case err := <-httpErrCh:
        s.acceptMuxConnections(ctx).Close()  // Cleanup
        return err
    case err := <-acceptErrCh:
        s.server.Shutdown(ctx)  // Cleanup
        return err
    }
}

// Atomic load/store avoids mutex overhead
if s.managementRoutesRegistered.CompareAndSwap(false, true) {
    s.registerManagementRoutes()  // Single registration
}
```

**Pattern:** One goroutine per request (standard Go http.Server). Atomic bools prevent race conditions without mutex contention.

**OmniRoute: Express middleware chain with Promise singleton**

File: `src/app/api/v1/chat/completions/route.ts` (lines ~5-20)

```typescript
let initPromise: Promise<void> | null = null;

async function ensureInitialized() {
    if (!initPromise) {
        initPromise = Promise.resolve(initTranslators());
    }
    return initPromise;
}

export async function POST(request: Request) {
    await ensureInitialized();  // All requests serialize on first init
    // After first request completes init, subsequent requests skip the wait
    return await handleChat(request);
}
```

**Pattern:** Promise singleton (not a mutex). First request initializes translators; subsequent requests reuse the same Promise (which is already resolved, so they don't wait). Avoids race conditions on translator setup.

**9router: Event-driven with process-level concurrency**

File: `src/mitm/manager.ts` (lines 30-80)

```typescript
// Spawns a single child process for the MITM server
const child = spawn('node', ['src/mitm/server.cjs'], {
    env: { ROUTER_API_KEY: apiKey, MITM_LOCAL_PORT: port }
});

// Monitors child's stderr for errors
child.stderr.on('data', (data) => {
    if (data.includes('address in use')) {
        clearCachedPassword();
        throw new Error('Port conflict');
    }
});

// Waits for startup or 2-second timeout
await new Promise((resolve, reject) => {
    const timeout = setTimeout(() => resolve(), 2000);
    child.on('error', reject);
    child.on('exit', reject);
});
```

**Pattern:** Subprocess-based. The MITM server (CommonJS) runs in its own process; OmniRoute's Node process spawns it and monitors for errors. This isolates the MITM from the main app's event loop.

**Concurrency matrix:**

| System | Model | Per-request? | State synchronization | Bottleneck |
|--------|-------|-------------|----------------------|------------|
| CLIProxyAPI | Goroutine pool (Go http.Server) | Yes | Atomic bools (lock-free) | None (truly concurrent) |
| OmniRoute | Async/await (Node event loop) | Yes (single-threaded) | Promise singleton for init | JS single-thread limits CPU-bound work |
| 9router | Subprocess + main event loop | Yes (event loop) + one persistent child | Subprocess isolation | Child process I/O + IPC overhead |

**Chorus concurrency:** Subprocess-based (like 9router), but we spawn many (up to 4 per phase per participant). Manages state via sidecar JSON + atomic DB writes.

---

### G. Notable bug-fixes from commit history (the field lessons)

**OmniRoute (from CHANGELOG):**

1. **Codex Streaming Mutation (v3.7.8):** Codex executor mutated request body to force `stream: true`, but then marked request as non-streaming. Caused "ALL_ACCOUNTS_INACTIVE" errors. **Fix:** Clone request body, detect correct response format separately.
   - **Lesson for Chorus:** Always clone before mutate; avoid side effects on shared state

2. **Gemini Tool Schema Rejection:** Tool parameters with `x-` prefix (vendor extensions) or `deprecated` field caused HTTP 400. **Fix:** Strip before sending to Gemini API.
   - **Lesson:** Validate schema against destination provider's constraints, not just your own schema

3. **Billing Header Fingerprint Instability (v3.5.0):** Fingerprint derived from first user message, changing every turn. Forced cache_create (~100% cache misses). **Fix:** Use stable per-day hash.
   - **Lesson:** Don't derive cache keys from mutable content; use stable IDs (date, user, session ID)

4. **SQLite File Bloat (v3.4.0):** Large streaming responses fragmented the DB file, eventually causing SQLITE_FULL crashes. **Fix:** Extract payloads to filesystem artifacts, store path in DB.
   - **Lesson:** Monitor DB file size; streaming responses should bypass SQL entirely

5. **Email Privacy State Persistence (v3.3.0):** Global email visibility toggle not respected across all UI sections. **Fix:** Implement comprehensive label masking layer.
   - **Lesson:** Privacy flags need enforcement at data access layer, not UI layer

6. **Migration Version Collision (v3.2.0):** Duplicate migration slot `032` prevented newer migrations. **Fix:** Strict versioning + validation before upgrade.
   - **Lesson:** Migration slots are immutable; never reuse

**CLIProxyAPI (from recent commits, May 2026):**

1. **Removed usage tracking telemetry:** Deleted LoggerPlugin + in-memory statistics. Replaced with framework-agnostic `requestmeta` utilities. 
   - **Lesson:** Iteration → removal. Not all features stick; keep removal surgical (no dead code comments, full cleanup)

2. **Token source prioritization:** Added `auth.Attributes["api_key"]` fallback to `auth.Metadata["access_token"]` for OAuth refresh.
   - **Lesson:** Multiple auth paths require explicit priority; document fallback order

3. **Session ID caching:** Introduced stable `CachedSessionID(apiKey)` to prevent cache invalidation.
   - **Lesson:** Cache keys must be stable; UUID per request = always cache miss

**9router (from README + feature list):**

1. **RTK Token Saver:** Automatic compression of `tool_result` blocks (20-40% savings).
   - **Lesson:** Quota can be optimized mid-stream; don't wait for 429

2. **Anti-loop header (`x-omniroute-source`):** Prevents MITM server from re-routing its own traffic.
   - **Lesson:** When proxying proxies, mark self-originated traffic to avoid loops

3. **Port conflict resolution:** Checks `address in use` error, clears cached password, retries.
   - **Lesson:** Transient startup failures (port in use) deserve retry logic, not immediate abort

**Chorus analogs:**

- **9router's anti-loop header** → Chorus could mark streams with `x-chorus-phase` to avoid re-processing
- **OmniRoute's fingerprint stability** → Chorus's _meta.json sidecar already does this (captures model at spawn time)
- **Codex body mutation** → Chorus's immutability rule prevents this already
- **SQLite bloat** → Chorus migrated to libsql to avoid this

---

### H. Things Chorus could adopt now (ranked by effort:reward)

1. **Accept-Encoding: identity for CLI output streams** (effort: 5 min, reward: high)
   - **Where:** `src/daemon/runner.ts` when spawning subprocess
   - **What:** Ensure CLI processes don't compress stdout (shouldn't be an issue, but CLIProxyAPI enforces it for SSE parsing)
   - **Code:** Already done (stream-json is line-buffered)
   - **Verdict:** Already correct; document in code comment why

2. **Pre-flight quota check before spawning subprocess** (effort: 2 hours, reward: medium)
   - **Where:** `src/daemon/runner.ts` before spawning CLI
   - **Inspiration:** 9router's real-time quota tracking
   - **Logic:** 
     ```typescript
     for (voice of selectedVoices) {
       const quota = await checkProviderQuota(voice.provider);
       if (quota.exhausted) {
         logWarning(`Skipping ${voice} — quota exhausted`);
         skipVoices.push(voice);
       }
     }
     ```
   - **Implementation:** Query `dx_spend_event` table (per voice, last 24h), compare to provider's published daily limit
   - **Effort:** Depends on how complete cost tracking is

3. **Real-time token compression in phase_events** (effort: 4 hours, reward: low-medium)
   - **Where:** `src/lib/server/rpc_*.ts` when storing phase output
   - **Inspiration:** 9router's RTK Token Saver
   - **Logic:** Detect base64 images in LLM responses, convert to pointer + thumbnail
   - **Effort:** Need to handle multiple image formats (PNG, JPEG, WebP)
   - **Verdict:** Low ROI for Chorus use case (single-review runs, not high-volume); defer to v1.0

4. **Stream pause/resume with graceful checkpoint** (effort: 6 hours, reward: medium)
   - **Where:** `src/daemon/runner.ts` subprocess management
   - **Inspiration:** CLIProxyAPI's context-aware timeouts + graceful shutdown
   - **Logic:** On SIGTERM, drain `pendingWrites`, save checkpoint, allow resume
   - **Current state:** Already have orphan reaper + grace period
   - **Improvement:** Add resume-from-checkpoint (DB transaction markers)
   - **Verdict:** Worth doing post-v0.7 for production resilience

5. **Provider-specific error classification + retry heuristics** (effort: 8 hours, reward: high)
   - **Where:** `src/lib/server/error_handler.ts` (new file)
   - **Inspiration:** OmniRoute's quota detection + error categorization
   - **Logic:** 
     ```typescript
     const errorCategory = classifyError(response);
     // "network", "auth", "quota", "validation", "timeout", "provider_outage"
     
     if (errorCategory === "quota") {
       logWarning(`Voice ${voice} hit quota; excluding from next retry`);
     } else if (errorCategory === "network") {
       retry();  // Transient
     } else if (errorCategory === "auth") {
       Surface to user for token refresh
     }
     ```
   - **Implementation:** Provider-specific error patterns (each provider's quota message format differs)
   - **Verdict:** Important for production reliability; medium priority for v0.8

6. **Billing fingerprint stability (cache misses analysis)** (effort: 2 hours, reward: low)
   - **Where:** `src/lib/server/llm_cost.ts` (if costs are tracked)
   - **Inspiration:** OmniRoute's lesson on fingerprint drift
   - **Current state:** No explicit billing/cost tracking in Chorus core
   - **Future:** If we implement cost attribution per voice, use stable keys (not message content)
   - **Verdict:** Plan ahead; low priority now

---

### I. Things Chorus could adopt LATER (if exposing HTTP endpoint)

If Chorus ever exposes an `/v1/messages`-compatible HTTP endpoint so users can point their existing tools (Cursor, VSCode + Claude extension) at Chorus:

1. **Router registration pattern** (from CLIProxyAPI)
   - File-based route registration, provider-specific routing paths
   - Allow power users to opt into `/v1beta/gemini` vs `/v1/messages` based on voice

2. **Format translation layer** (from CLIProxyAPI + OmniRoute)
   - OpenAI → Claude request translation (tool schemas)
   - Claude → OpenAI response translation (tool-use blocks)
   - Use OmniRoute's pattern: delegate to provider SDKs where possible
   - Build translators only for the formats you own (Chorus's proprietary phase_event format)

3. **Streaming response transformation** (from OmniRoute)
   - Use Web Streams pipeThrough for in-flight format conversion
   - Avoid full buffering; chunk-by-chunk transformation

4. **Anti-loop headers** (from 9router)
   - Mark Chorus-originated requests with `x-chorus-phase` + `x-chorus-run-id`
   - Reject recursion (prevent Chorus calling itself)

5. **Per-voice concurrency limits** (new, inspired by all three)
   - Allow N concurrent requests per voice (e.g., 3 Opus, 5 Sonnet)
   - Queue excess requests; emit to caller via header + retry-after

6. **Graceful shutdown + connection draining** (from CLIProxyAPI)
   - On SIGTERM, accept no new requests, wait up to 30s for in-flight to complete
   - After grace period, hard kill

---

## Summary: Top 3 actionable items this week

1. **Pre-flight quota check before subprocess spawn** (2 hours)
   - Prevents wasted CLI startup on exhausted accounts
   - Query `dx_spend_event`, compare to provider limits
   - Inspect: `src/daemon/runner.ts` + `src/lib/server/cost_tracking.ts`

2. **Provider-specific error classification** (4 hours, split into two parts)
   - Part A: Categorize errors (quota, auth, network, validation, timeout)
   - Part B: Implement retry heuristics
   - Test against real provider errors (429s, auth failures, timeouts)
   - File: `src/lib/server/error_handler.ts` (new)

3. **Documentation: Add Accept-Encoding comment + audit stream output** (1 hour)
   - Confirm subprocess output is not being compressed
   - Add comments explaining why in `src/daemon/runner.ts`
   - Blocks: None; low risk

**Not recommended this iteration:**
- Token compression (low ROI for single-run reviews)
- Stream pause/resume (defer to v1.0)
- HTTP endpoint exposure (future feature gate)

---

## Round 2: deeper robustness investigation

Examined CLIProxyAPI (Go), OmniRoute (Node.js), and 9router (Node.js) for lifecycle, retry, logging, resource, error, and configuration patterns. Traced equivalent Chorus paths with fresh eyes.

### A. Subprocess / connection lifecycle

**Proxy patterns:**

1. **CLIProxyAPI (Go, server.go:40-150)** — Atomic bools for lock-free state (`wsAuthEnabled.CompareAndSwap(false, true)`), multiplexed listeners with separate goroutines for HTTP serve + connection accept, graceful shutdown closes listeners before server.Shutdown(). No connection pools (Go http.Server is stateless per-request).

2. **OmniRoute (Node.js)** — No explicit connection lifecycle documented. Delegates to Express/Next.js middleware, relies on Node's http module auto-timeouts. No connection pooling strategy visible in route handlers.

3. **9router (Node.js)** — Spawns a single persistent MITM subprocess (`spawn('node', [...], {env: {...}})`) and monitors stderr for "address in use" errors. No connection pool; one MITM process per daemon.

**Chorus equivalent:**

- `src/daemon/headless.ts:207-441` — Spawns one subprocess per phase per participant. Implements hard timeout (default 10min, configurable), SIGTERM→5s grace→SIGKILL sequence (lines 313-329). Registers PID to disk (lines 61-68) for orphan reaper across daemon restarts.
- `src/daemon/reaper.ts:24-51` — Runs on configurable interval (default 5min), kills orphaned processes from prior crashes, clears PID records.
- `src/daemon/index.ts:313-324` — Drains `pendingWrites` promise set before releasing activeRuns slot, preventing duplicate run spawns on reattach.

**Safe? Yes, with caveats:**
- PID persistence (lines 38-46, headless.ts) is robust — survives daemon crash.
- SIGTERM+grace (lines 313-329) matches CLIProxyAPI's "graceful shutdown" pattern.
- **Caveat 1:** Hard timeout is process-wide (10min default). If a phase is legitimately long-running, it hits SIGKILL. Document timeout in templates and allow per-phase override.
- **Caveat 2:** `reapOrphanProcesses()` is called once at daemon startup (line 95 says "call once from bootstrap"). If a daemon crashes and restarts multiple times in quick succession, the grace period (KILL_GRACE_MS=5s) may not fire before the next reaper run tries again. Low impact (SIGKILL twice doesn't hurt), but log each attempt for diagnosis.

**Action items:**
1. Add per-phase timeout override in template schema (currently hardcoded 10min).
2. Document timeout expectations in planning/CODING-PRINCIPLES.md or planning/PAGE-TEMPLATE.md.

---

### B. Retry strategies

**Proxy patterns:**

1. **CLIProxyAPI** — Minimal retry logic. Propagates HTTP errors (429, 5xx) to caller as-is. No built-in retry; caller/wrapper owns fallback decisions.

2. **OmniRoute (from round 1 CHANGELOG)** — Implements 4-tier smart fallback (Subscription → API Keys → Budget → Free). Detects 429 + quota messages via regex. No exponential backoff documented; retries next tier immediately.

3. **9router** — Similar 3-tier fallback. Also implements pre-emptive quota tracking (maintains spend counters in LowDB) to skip exhausted accounts before 429 fires.

**Chorus equivalent:**

- `src/daemon/runner.ts:132-264` — Implements max-round retry loop (phase.iterate.maxRounds) but DOES NOT retry on transient CLI failures. If a doer times out, it breaks immediately (line 183). No classification of "transient" vs "fatal" errors.
- `src/daemon/error-detector.ts:76-160` — Pattern-matches known CLI failure modes (quota_exhausted, token_refresh_lost, mcp_handshake_failed) but does NOT distinguish retryable from fatal. Quota errors are detected but no auto-retry is triggered.
- `src/daemon/agents/codex.ts` and others — Each shim spawns the CLI once. No built-in retry logic; the runner's outer loop (phase.iterate.maxRounds) is the only retry mechanism.

**Safe? Partially. Retryable failures (transient network hiccups, quota resets) cause the entire phase to fail instead of retrying. The design assumes all reviewer failures are equally bad (either skip the reviewer or fail the chat).**

**Specific gaps:**

1. **No transient classification** — If a reviewer times out due to a network blip, it counts the same as "quota exhausted" or "auth failed". No attempt to retry the timeout.
2. **No retry within a round** — A round's doer attempt fails once per error. Must burn a max-round slot to try again. With phase.iterate.maxRounds=3, a single network blip wastes 1/3 of retries.
3. **No exponential backoff** — If retrying is added, OmniRoute's immediate-next-tier approach (no backoff) is fine for sequential fallback, but Chorus's parallel reviewer design means multiple failures firing simultaneously. Adding jitter/backoff would smooth the spike.

**Action items (priority 6+, not blocking):**
1. Classify errors: network (transient) vs auth (fatal) vs quota (depends on resetAt).
2. Add retry-within-round for network/timeout errors (1-2 retries with 500ms jitter).
3. Update error_detector.ts to emit retry-eligibility flag.

---

### C. Logging / observability

**Proxy patterns:**

1. **CLIProxyAPI** — Recently removed in-memory statistics + LoggerPlugin (May 2026 commit). Now uses framework-agnostic `requestmeta` utilities. No evidence of structured logs visible in server.go.

2. **OmniRoute (route.ts)** — Logs oversized payloads (>256KB) via `OMNIROUTE_LOG_REQUEST_SHAPE` env var. Uses console or framework logger (not visible in snippet).

3. **9router** — Monitors stderr for specific strings ("address in use"). No structured logging visible in code snippets.

**Chorus equivalent:**

- `src/daemon/index.ts:188-300` — onEvent callback dispatches to DB + SSE, but NO request ID tracking. Each event is tagged with chatId + ts, but no unique request/run ID spanning a full chat lifecycle.
- `src/daemon/headless.ts:226, 283, 378-402` — console.warn/error scattered throughout. No structured logs; no correlation IDs.
- `src/daemon/runner.ts:141-196` — Events are emitted without trace IDs. Makes it hard to correlate multiple events from the same run or follow a request across phase boundaries.
- Secrets handling: no evidence of redaction in logs. If a prompt contains a secret API key, it would leak to stderr/logs.

**Safe? Mostly, but with observability gaps:**

1. **No unique run ID** — chatId + ts is insufficient for tracing. If a user reruns the same chat, old and new logs mix. No way to correlate "phase_start event from run N" with "phase_done from run N".
2. **No redaction** — If a CLI error message includes an API key or auth token, it gets logged verbatim (src/daemon/runner.ts:390-402 trimmed to 300 chars, but no sanitization).
3. **No structured logs** — JSON-formatted logs would enable better parsing and querying. Current printf-style logs are human-readable but machine-unfriendly.

**Action items (low priority, nice-to-have):**

1. Add `runId` to the event envelope (generated at chat start, stable across phases).
2. Add a log-line redaction pass for known secret patterns (bearer tokens, API keys).
3. Migrate to structured logging library (winston, pino) post-v1.0.

---

### D. Memory / resource management

**Proxy patterns:**

1. **CLIProxyAPI** — No explicit memory management visible. Go's http.Server handles cleanup automatically. No streaming response buffering strategy documented.

2. **OmniRoute** — Encountered SQLite file bloat (v3.4.0 CHANGELOG) from large streaming responses. Solution: extract payloads to filesystem artifacts, store path in DB. Now avoids buffering huge responses in SQL.

3. **9router** — Uses LowDB (JSON file, no native builds). Small ephemeral data (OAuth tokens, quota counters). Not designed for large payloads.

**Chorus equivalent:**

- `src/daemon/runner/stream-file-writer.ts:28-50` — Buffers text_deltas with 4KB flush threshold + 750ms timeout (lines 31-35). CRITICAL: buffer is append-only; no cap on total size per run. A Opus run with 100KB+ output would accumulate in memory until flush.
- `src/daemon/headless.ts:271-291` — fullStdout and fullStderr accumulate entire CLI output in memory (line 275: `fullStdout += chunk`). For a 5-minute Opus run, this is ~100KB (reasonable). But for a long analysis or multi-file review, could balloon to 1MB+.
- `src/daemon/runner/doer.ts:57, 189` — accumulated variable holds all text_deltas until message_done (line 57). If message_done never fires, this is a memory leak.

**Safe? Mostly, but with leak risk:**

1. **Unbounded fullStdout/fullStderr** — If a CLI hangs (doesn't exit), the buffers grow forever. The hard timeout (10min) saves this from becoming a permanent leak, but 10min × (streaming CLI output) could be 10MB+.
2. **StreamFileWriter buffer** — If flushNow() is never called (or the write fails silently), buf accumulates. The write() method only flushes on size (4KB) or timer (750ms), but if the filesystem is full, appendFileSync could throw and the exception is swallowed (line 46: `finally`). Next write() call would accumulate more.
3. **message_done semantics** — If a streaming CLI crashes mid-stream, message_done never fires. finalText stays undefined (line 185-187 returns null, losing accumulated content). This is intentional for error cases, but if a CLI "hangs" sending deltas, accumulated memory is lost when the timeout fires.

**Action items (priority medium):**

1. Cap fullStdout/fullStderr at 10MB; log and truncate if exceeded (prevent OOM on hung CLI).
2. Wrap StreamFileWriter.flushNow() in error handling; if write fails, log the error and close the writer (prevent silent accumulation).
3. Add memory profiling to corpus: run a 100+ phase chat and measure peak memory. Document expected usage.

---

### E. Error surfacing / UX

**Proxy patterns:**

1. **CLIProxyAPI** — Propagates HTTP status + error body from upstream. Caller/IDE decides how to surface. No special "actionable error" formatting.

2. **OmniRoute** — Emit specific error codes (security codes for prompt injection). Logs oversized payloads for diagnostics. Fallback tier selection is transparent to caller (tries next tier silently).

3. **9router** — Fallback chain is opaque; user sees the first successful tier response. Pre-emptive quota tracking prevents 429s from ever surfacing.

**Chorus equivalent:**

- `src/daemon/runner.ts:170-182, 247-258` — Emits phase_failed event with a reason string (e.g., 'doer_timeout', 'max_rounds_exhausted'). No retry-after countdown; no actionable CTA for the user.
- `src/daemon/error-detector.ts:23-30` — CliError interface includes `cta` field (line 27), but it's only set for specific patterns (quota, token_refresh, MCP). Generic errors have no CTA.
- `src/daemon/runner.ts:376-402` — On CLI exit with non-zero code, surfaces stderr + stdout tail (300 chars each, lines 391-395). Gives raw CLI error, but no translation to user intent.

**Safe? Mostly. Errors are surface but lack guidance:**

1. **No retry-after countdown** — If quota is exhausted with resetAt timestamp, the error_detector captures it (line 109, error-detector.ts), but the UI has no countdown timer. User must manually wait and retry.
2. **Generic error messages** — "CLI exited 1" is opaque. Some CLIs print their actual error to stdout (kimi: "LLM not set") and Chorus captures it (line 392), but no post-processing to extract actionable intent.
3. **Permission prompts** — error_detector finds permission_prompt errors (line 158-160), but the fix (perm model Layer 2: shim.recoverKeys) is not invoked automatically. User must manually click "Retry" or flip settings.

**Action items (low priority, UX improvement):**

1. Display resetAt countdown when quota error occurs (calculate from resetAt ms-epoch).
2. Add a post-processor in error_detector to extract common patterns: "LLM not set" → "Configure your LLM in the CLI settings".
3. Auto-invoke Layer 2 recovery (shim.recoverKeys) on permission_prompt detection instead of requiring user retry.

---

### F. Configuration & environment surprises

**Proxy patterns:**

1. **CLIProxyAPI** — Loads config from YAML at startup. UpdateClients() method allows hot-reload without restart. Uses reflect.DeepEqual to detect changes (server.go ~lines 300+). Provides graceful transition; old clients keep working during update.

2. **OmniRoute** — No visible config hot-reload. Crashes on missing required config (app initialization is mandatory). Delegates to provider SDKs for their config (anthropic, openai modules manage API keys separately).

3. **9router** — Config passed via environment variables at subprocess spawn (manager.ts lines 424-439). MITM server reads env vars (process.env.ROUTER_API_KEY, process.env.MITM_LOCAL_PORT). No validation at boot; would fail at first request if missing.

**Chorus equivalent:**

- `src/lib/settings/transport.ts:52-76` — Reads CHORUS_TRANSPORT env var OR database setting. Falls back to DEFAULT_TRANSPORT. No validation that the value is a valid transport enum until safeParse (line 60). Invalid env override silently degrades to default.
- `src/lib/settings/permissions.ts:46-61` — Reads 3 settings from DB with safeParse fallbacks. No validation that the DB returned a sane value; only type-check on retrieval.
- `src/daemon/index.ts:28` — Reads CHORUS_DAEMON_PORT from env with parseInt fallback (0 = invalid, would bind to random port). No error if env is set to a non-numeric string.
- `src/lib/db/index.ts:38` — Reads CHORUS_DB_PATH override. No validation that the path is writable.

**Safe? Yes for most cases, with one edge case:**

1. **Invalid CHORUS_DAEMON_PORT** — parseInt("abc", 10) = NaN. Passing NaN to Fastify.listen(NaN) would cause an error at startup, caught and logged. Safe, but error message is opaque ("listen EINVAL").
2. **Unwritable CHORUS_DB_PATH** — Caught at daemon startup (index.ts lines 336-346 probe the DB), so not silent. Good error handling.
3. **Invalid CHORUS_TRANSPORT** — Falls back to default (headless). Safe, but operator doesn't realize their override was ignored. Could add a warning log if env is set but invalid.

**Action items (low priority, operational polish):**

1. Validate CHORUS_DAEMON_PORT at startup; reject NaN with a helpful error message.
2. Log a warning if CHORUS_TRANSPORT env is set to an invalid value (helps operators debug typos).
3. Add startup logging showing which config source won for each setting (env override vs DB default).

---

### G. Race conditions found in Chorus during this re-read

1. **_meta.json sidecar write timing (agents/kimi.ts:99-102)** — Each participant shim writes a _meta.json sidecar after the subprocess finishes. Multiple participants in a phase could write concurrently. The cockpit polls for _meta.json to show transport/model info. If cockpit reads the file mid-write, it gets incomplete JSON.
   - **File:** `src/daemon/agents/kimi.ts:99`
   - **Risk:** Low. The file is small (< 1KB) and fs.writeFileSync is atomic on most filesystems. But if the write is interrupted (e.g., daemon crash during flush), the file could be corrupt.
   - **Fix:** Write to temp file first, then rename(). fs.renameSync is atomic on POSIX.

2. **chat_done latch vs normal terminal emission (runner.ts:97-114)** — The chat_done latch uses a flag `chatDoneEmitted` to ensure the first call wins. An abort listener and the normal phase loop both call emitChatDone. This is correctly guarded, so NOT a race — the latch prevents double-emission.
   - **Status:** Safe as-is. Design is sound.

3. **activeRuns slot release vs pendingWrites drain (index.ts:313-324)** — The finally block waits for all pending DB writes (chats.update, phaseEvents.create) before deleting the activeRuns slot. This is intentional and well-documented (line 314-317). Correctly prevents duplicate runs on reattach.
   - **Status:** Safe as-is. Documented and tested.

4. **StreamFileWriter timer cleanup on error (stream-file-writer.ts:38-49)** — If fs.appendFileSync throws, the exception is caught in the finally block (line 46: `finally { this.buf = '' }`), which clears the buffer AND resets flushTimer to null. If the same writer.write() is called again after a failed flush, a new timer starts. But the lost data is never recovered.
   - **File:** `src/daemon/runner/stream-file-writer.ts:38-49`
   - **Risk:** Low-medium. Only affects runs where the filesystem fills or answer.md becomes unwritable mid-run (unlikely in normal operation). But silent data loss is bad.
   - **Fix:** Log the write failure and mark the writer as "dead" so subsequent calls error instead of silently losing data. Or: emit a RunnerEvent to notify the runner/user that streaming stopped.

5. **Orphan reaper timing window (headless.ts:112-120)** — The reaper schedules SIGKILL after KILL_GRACE_MS (5s) with `setInterval` that's unref'd. If the daemon crashes during the grace period, the SIGKILL never fires. The PID record stays on disk for the next daemon startup. This is OK for correctness (next reaper run kills it), but there's a 5-minute window (reaper interval) where the orphan is alive.
   - **File:** `src/daemon/headless.ts:112-120`
   - **Risk:** Low. Only delays cleanup by one reaper interval. Acceptable trade-off for simplicity.
   - **Fix:** Document the potential 5-minute lag in comments. No change needed; behavior is intentional.

---

### H. Prioritized action list (round 2)

Rank by **impact × likelihood × effort**.

1. **Cap fullStdout/fullStderr memory accumulation (HIGH impact, high likelihood, low effort)**
   - File: `src/daemon/headless.ts:275-276`
   - Change: Add a 10MB cap; truncate and log if exceeded
   - Why: Prevents OOM on hung CLI or pathological large output
   - Effort: 10 lines
   - Blocks: None

2. **Validate CHORUS_DAEMON_PORT at startup (MEDIUM impact, low likelihood, minimal effort)**
   - File: `src/daemon/index.ts:28`
   - Change: Check parseInt result is not NaN; emit helpful error
   - Why: Catches operator typos early; improves startup diagnostics
   - Effort: 5 lines
   - Blocks: None

3. **Add per-phase timeout override in template schema (MEDIUM impact, low likelihood, medium effort)**
   - File: `src/lib/template-schema.ts` (template Phase interface)
   - Change: Add optional `timeoutMs` field; use in runner.ts spawn
   - Why: Long-running phases (multi-file review) need > 10min; currently hard-coded
   - Effort: 20 lines + schema migration
   - Blocks: None; backward-compatible if optional

4. **_meta.json atomic write via temp-rename (LOW impact, low likelihood, low effort)**
   - File: `src/daemon/agents/kimi.ts:99-102` (and repeat for all shims)
   - Change: Write to temp, then fs.renameSync
   - Why: Prevents cockpit from reading corrupt JSON if daemon crashes mid-write
   - Effort: 5 lines per shim
   - Blocks: None

5. **Add error logging + mark to StreamFileWriter on failed flush (LOW-MEDIUM impact, low likelihood, low effort)**
   - File: `src/daemon/runner/stream-file-writer.ts:44-46`
   - Change: Log write failures; set a "dead" flag; emit RunnerEvent
   - Why: Prevents silent data loss; gives user visibility into filesystem problems
   - Effort: 10 lines
   - Blocks: None

6. **Add CHORUS_TRANSPORT env validation warning (LOW impact, very low likelihood, minimal effort)**
   - File: `src/lib/settings/transport.ts:52-66`
   - Change: Log warning if env is set but invalid
   - Why: Helps operators debug typos; currently fails silently
   - Effort: 3 lines
   - Blocks: None

**Defer to v1.0 or later:**
- Transient error retry + classification (medium effort, valuable but not blocking)
- Structured logging + request ID tracing (medium effort, nice-to-have for ops)
- Permission prompt Layer 2 auto-recovery (low effort but behavioral change; test thoroughly)
- resetAt countdown timer in UI (low effort but requires cockpit changes)

---

### I. Surprising patterns from the proxies

1. **Config hot-reload via atomic CompareAndSwap** (CLIProxyAPI) — Uses atomic bools to prevent race conditions during feature flag updates. Simpler than mutex; Go idiom but instructive. TypeScript could adopt a similar pattern with volatile flags + barriers.

2. **Pre-emptive quota exhaustion detection** (9router) — Tracks spend counters in LowDB and skips accounts before 429 fires. Saves a round-trip + provider retry. Chorus does pre-flight quota checks (line 395, runner.ts), but doesn't track spend; could adopt 9router's counter approach.

3. **Filesystem artifacts instead of DB bloat** (OmniRoute changelog v3.4.0) — Learned the hard way that large streaming responses fragment SQLite. Now extract payloads to disk. Chorus uses a different pattern (answer.md files + libsql), but the lesson applies: monitor DB file growth; if it bloats, consider extracting large columns.

4. **Env-var-based MITM subprocess config** (9router) — Spawns MITM server with env var injection. Simple, no IPC needed. Chorus could use this pattern for agent-spawning (pass config via env instead of reading from DB per-spawn).

