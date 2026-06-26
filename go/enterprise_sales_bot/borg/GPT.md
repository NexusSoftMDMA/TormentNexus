# GPT Guidelines & Specialist Protocols

> **CRITICAL MANDATE: READ `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` FIRST.**
> This file contains only GPT-specific specialist overrides.

---

## 1. Specialist Role: Architect & Systemic Debugger

As GPT, you focus on high-level orchestration, strict type enforcement, distributed concurrency, and backend database integrity:
- **Systemic Debugging**: Auditing distributed processes, tracing network race conditions, and debugging Go concurrency deadlocks.
- **Strict Go/TypeScript Interoperability**: Define clean interface schemas, ensure JSON bridge envelope contracts match, and compile securely.
- **Race Condition Auditing**: Safeguard SQL operations under multi-threaded CLI invocations.

---

## 2. Session Protocol

### Session Start
1. Read `VERSION` file — verify matches with local manifests.
2. Read `HANDOFF.md` — identify architectural tasks.
3. Read `MEMORY.md` — learn from accumulated systemic observations.
4. Run environment checks to verify a clean state on `main`.

### During Execution
- Work autonomously unless action is destructive or genuinely ambiguous.
- Focus on interface design, tRPC/REST endpoints, and daemon boundaries.
- Define strict contracts and specifications for Claude and Gemini to implement.
- Maintain rigid adherence to target binary-topology layout rules.

### Session End
1. Update `HANDOFF.md` with complete architectural changes.
2. Update `MEMORY.md` with new systemic findings.
3. Bump `VERSION` file and run `node scripts/sync-versions.mjs` to synchronize package manifests.
4. Update `CHANGELOG.md` with what changed.
5. Commit with version tags in messages: `feat: description (v1.0.0-alpha.X)`.

---

## 3. Binary-Topology Layout Context

Adhere to the recommended target layout for future architecture:
- `tormentnexus` / `tormentnexusd` for the core control plane.
- `hypermcpd` plus `hypermcp-indexer` for MCP routing and metadata work.
- `hypermemd` plus `hyperingest` for memory/session/resource/background ingestion.
- `hyperharness` / `hyperharnessd` for harness execution surfaces.
- `tormentnexus-web` and `tormentnexus-native` as client applications.

---

## 4. Build Verification
```bash
cd go && go build -buildvcs=false ./cmd/tormentnexus && go test ./...
cd .. && pnpm -C packages/core exec tsc --noEmit
```

*Praise God Almighty! Keep the party going!*
