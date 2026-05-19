# AGENTS — Nexus Kernel & HyperCode Contributor Guide

> **CRITICAL: ALL AGENTS MUST READ `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` BEFORE PROCEEDING.**

This file serves as the primary coordination point for multi-agent workflows and human operators.

## 1. Multi-Agent Handoff Protocol
- Agents communicate primarily through `HANDOFF.md`.
- Document exactly what you did, what failed, and what the next agent must do.
- Update `MEMORY.md` with new systemic observations or recurring bugs.
- **Cycle**: Read → Strategize → Execute → Validate → Commit → Handoff.

## 2. Model Specializations
| Model | Strengths |
|---|---|
| **Gemini** | Speed, massive context analysis, recursive maintenance scripts, bulk refactoring. |
| **Claude** | Deep implementation, UI/UX perfection, precise documentation, complex type-safety. |
| **GPT** | Architecture design, systemic debugging, Go porting authority, strict enforcement. |
| **Gemini** | Speed, recursive scripts, massive context processing, repo maintenance, bulk refactoring |
| **Claude** | Deep implementation, UI/UX perfection, documentation, styling, type safety |
| **GPT** | Architecture, systemic debugging, strict type enforcement, Go porting |

### Iteration Cycle
**Read → Strategize → Execute → Validate → Commit → Handoff. Never stop the party.**


Please summarize anything you have learned during this session that was not obvious at the start, or anything else you would like to inform yourself upon newer sessions.

## 2. Session Protocol

### Session Start
1. Read `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`
2. Read `VERSION` file — verify it matches `package.json` and dashboard display
3. Read `HANDOFF.md` — pick up where previous agent left off
4. Read `MEMORY.md` — learn from accumulated observations
5. Run `git fetch --all && git status` — verify clean state on `main`
6. Understand repo structure before making changes

### During Execution
- Work autonomously unless action is destructive or genuinely ambiguous
- Prefer small, verifiable changes over broad rewrites
- Use parallel tool calls when safe
- Keep status labels and documentation honest
- After any `pnpm install`, run `pnpm rebuild better-sqlite3` on Node 24

### Session End
1. Update `HANDOFF.md` with complete session summary
2. Update `MEMORY.md` with new observations
3. Bump `VERSION` file and sync all `package.json` files
4. Update `CHANGELOG.md` with what changed
5. Commit with version number in message: `feat: description (v1.0.0-alpha.X)`
6. Push to both remotes: `origin` and `borg-upstream`
7. Update `TODO.md` and `ROADMAP.md` if priorities changed

Please include https://github.com/robertpelloni/jules-autopilot as a dashboard for integrating google jules cloud dev env, and also https://github.com/robertpelloni/opencode-autopilot, and please also make a dashboard with links to all convenient major AI tools, like a section for the billing for API key for gemini/vertex/aistudio/cloud, azure, openrouter, chatgpt/openai api key, claude api key, and then also links to the pro plan membership subscription management page, and then use the spreadsheet i linked with all the ai tools on another dashboard page to detect all installed tools and link to their homepage, and then detect whether they are connected using account oauth or api key, and show usage for each provider both account and api key and show budgets, copilot plus subscription, and then also make a dashboard page similar to the jules-app page for copilot cloud, claude chrome, claude cloud, openai codex cloud, and other providers like blocks, basically useful dashboard pages that organize everything into one place both locally and remote.

## 3. Version Management
- **Single Source of Truth**: `VERSION.md` contains the master version string.
- All `package.json` files must sync to this value.
- Commit messages must include the version number: `feat: description (v1.0.0-alpha.X)`.

## 4. Required Runtime Ports
| Service | Port | Purpose |
|---|---|---|
| Next.js Dashboard | 3000 | Web observation deck |
| Socket.io | 3001 | Real-time swarm signals |
| tRPC Bridge | 4100 | TypeScript Control Plane API |
| Nexus Go Kernel | 4300 | Authoritative native sidecar |

## 5. Development Rituals
1. **Sync**: `git fetch --all && git status` (Verify clean `main`).
2. **Read**: `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`.
3. **Check**: `VERSION.md` vs. local dashboard state.
4. **Learn**: `HANDOFF.md` and `MEMORY.md` from previous session.

*Praise the LORD!!! Keep on goin'! Don't ever stop! Don't stop the party!!!*
