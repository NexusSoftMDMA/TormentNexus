# AGENTS — TormentNexus Kernel & TormentNexus Contributor Guide

> **CRITICAL: ALL AGENTS MUST READ `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` BEFORE PROCEEDING.**

This file serves as the primary coordination point for multi-agent workflows and human operators.

---

## 1. Multi-Agent Handoff Protocol
- Agents communicate primarily through `HANDOFF.md`.
- Document exactly what you did, what failed, and what the next agent must do.
- Update `MEMORY.md` with new systemic observations or recurring bugs.
- **Cycle**: Read → Strategize → Execute → Validate → Commit → Handoff.

---

## 2. Model Specializations

| Model | Strengths | Focus Areas |
|---|---|---|
| **Gemini** | Speed, massive context processing, repo maintenance | Bulk refactoring, recursive scripts, context analysis |
| **Claude** | UI/UX perfection, documentation, deep feature execution | Responsive layouts, type safety, precise documentation |
| **GPT** | Systemic architecture, distributed debugging, race conditions | Go/TS bridge contracts, DB migration, concurrency safety |

---

## 3. Session Protocol

### Session Start
1. Read `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` to load canonical rules.
2. Read the `VERSION` file to check dashboard synchronization.
3. Read `HANDOFF.md` to resume exactly where the previous agent left off.
4. Read `MEMORY.md` to review accumulated multi-agent insights.
5. Run git checks to ensure workspace cleanliness.

### During Execution
- Work autonomously unless changes are destructive or highly ambiguous.
- Prefer small, incremental, easily verifiable commits.
- Ensure loading, error, and empty states are represented across all dashboard interfaces.
- After any `pnpm install`, run `pnpm rebuild better-sqlite3` on Node 24.

### Session End
1. Update `HANDOFF.md` with a complete, detailed session summary.
2. Update `MEMORY.md` with new developer observations or gotchas.
3. Bump the `VERSION` file and synchronize workspaces using `node scripts/sync-versions.mjs`.
4. Update `CHANGELOG.md` with recent feature implementations.
5. Commit clean changes with version tag: `feat: description (v1.0.0-alpha.X)`.
6. Push commits to `origin` and `tormentnexus-upstream` remotes.

---

## 4. Required Runtime Ports

| Service | Port | Purpose |
|---|---|---|
| Next.js Dashboard | 3000 | Web observation deck |
| Socket.io | 3001 | Real-time swarm signals |
| tRPC Bridge | 4100 | TypeScript Control Plane API |
| TormentNexus Go Kernel | 4300 | Authoritative native sidecar |

*Praise the LORD! Keep on going! Don't ever stop! Don't stop the party!!!*
