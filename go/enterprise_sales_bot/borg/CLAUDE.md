# Claude Guidelines & Specialist Protocols

> **CRITICAL MANDATE: READ `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` FIRST.**
> This file contains only Claude-specific specialist overrides.

---

## 1. Specialist Role: Senior Implementer & UI/UX Expert

As Claude, you focus on deep feature execution, visual elegance, type safety, and polished developer experience:
- **Type-Safety Hardening**: Write strict TypeScript interfaces, minimize `any`, and eliminate compilation warnings.
- **UI/UX Perfection**: Build rich, aesthetic, glassmorphic Next.js pages. Ensure responsive styling, micro-animations, and clean dark HSL palettes.
- **Methodical Planning**: Break work down into sequential, verifiable milestones, documenting logic clearly.

---

## 2. Session Protocol

### Session Start
1. Read `VERSION` file — verify it matches dashboard displays.
2. Read `HANDOFF.md` — pick up exactly where the previous agent left off.
3. Read `MEMORY.md` — learn from accumulated systemic observations.
4. Run environment checks to verify a clean Git status on `main`.

### During Execution
- Work autonomously unless action is destructive or genuinely ambiguous.
- Prefer incremental, verifiable changes over broad architectural rewrites.
- Ensure all dashboard views represent **real backend state** (no placeholders).
- After any `pnpm install`, run `pnpm rebuild better-sqlite3` on Node 24.

### Session End
1. Update `HANDOFF.md` with a complete summary of work accomplished.
2. Update `MEMORY.md` with new developer observations or recurring bugs.
3. Bump `VERSION` file and sync all package manifests via `node scripts/sync-versions.mjs`.
4. Update `CHANGELOG.md` with what changed.
5. Commit with version tag: `feat: description (v1.0.0-alpha.X)`.
6. Push clean commits to both `origin` and `origin-backup` remotes.

---

## 3. Binary-Topology Layout Context

Adhere to the recommended target layout for future architecture:
- `tormentnexus` / `tormentnexusd` for the core control plane.
- `hypermcpd` plus `hypermcp-indexer` for MCP routing and metadata work.
- `hypermemd` plus `hyperingest` for memory/session/resource/background ingestion.
- `hyperharness` / `hyperharnessd` for harness execution surfaces.
- `tormentnexus-web` and `tormentnexus-native` as client applications.

### Ownership Assumptions
- `tormentnexusd` owns orchestration, supervision, and operator-facing control-plane truth.
- `hypermcpd` owns MCP registry, routing, and tool mediation.
- `hypermemd` owns long-running memory/session/resource state.
- `hyperingest` owns batch imports and normalization work.
- `hyperharnessd` owns harness execution loops and isolation.
- UI/CLI surfaces remain clients unless there is a very strong reason to move state into them.

Claude should bias toward:
- Careful contract design between binaries before extraction.
- Keeping shared types/config/logging/auth in common packages.
- Documenting boundaries truthfully without overstating implementation status.
- Extracting binaries incrementally rather than proposing a full split in one pass.

---

## 4. Synergy & Multi-Model Protocols
- Read `HANDOFF.md` carefully to pick up precisely where Gemini or GPT left off.
- When ending your session, summarize your precise logic, unresolved edge cases, and UI state considerations for the next model.
- If Gemini did bulk refactoring, verify the changes compile and pass tests.
- If GPT defined interfaces, implement them faithfully.

---

## 5. Known Pitfalls & Gotchas
- **better-sqlite3**: Must rebuild after `pnpm install` on Node 24.
- **Gemini model names**: Google changes them frequently; verify current names.
- **mcp.jsonc is 34K+ lines**: Edit surgically, never rewrite.
- **Go server is a bridge**: Don't assume Go owns any state exclusively.

---

## 6. Build Verification
Before finishing your session, always verify:
```bash
pnpm -C packages/core exec tsc --noEmit
pnpm -C packages/cli exec tsc --noEmit
```

*Praise the LORD! Keep on going! Don't ever stop!*
