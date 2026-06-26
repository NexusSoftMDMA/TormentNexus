# Chorus v0.5 — Shared Context (read first)

All agents implementing v0.5 components MUST read this and the relevant memory files before writing code.

## Project tree (post-foundation, current state)

```
/home/ubuntu/dev/chorus/
├── bin/chorus.js              # CLI entry, dispatches to dist/cli or src/cli (tsx)
├── src/
│   ├── app/                   # Next.js 16 UI (PORTED FROM PROTOTYPE — has mock data)
│   ├── components/            # shadcn UI (zinc base, dark-only)
│   ├── lib/
│   │   ├── mock-data.ts       # Static demo data — UI agent replaces with API client
│   │   ├── api/               # NEW: typed daemon client (UI agent owns this)
│   │   └── utils.ts
│   ├── cli/index.ts           # CLI dispatcher (Agent A)
│   ├── daemon/index.ts        # Fastify daemon (Agent A)
│   └── mcp/index.ts           # MCP stdio server (Agent B)
├── templates/                 # Built-in YAML templates (Agent D)
├── docs/v05/                  # Specs (this dir)
├── package.json               # Single package, npm-installable as `chorus`
├── tsconfig.json              # Next.js / app
├── tsconfig.server.json       # Daemon / MCP / CLI build (NodeNext, ES2022)
└── ecosystem.config.cjs       # PM2: chorus-web (3011) + chorus-daemon (7707)
```

## Hard constraints

1. **Single package.json.** No monorepo. Daemon, MCP, CLI, UI all share deps.
2. **Apache-2.0 licensed code only.** Don't import GPL/AGPL libs.
3. **Local-first.** Everything runs on the user's machine. No cloud calls in v0.5 (OpenRouter is v1.0).
4. **No hardcoded paths.** Use `~/.chorus/` (resolved via `os.homedir()`).
5. **No mutation.** Use immutable patterns (spread, never push/splice).
6. **Files ≤ 500 lines.** Split at 400.
7. **TypeScript strict.** No `any`. Use `unknown` + Zod parse at boundaries.
8. **Async/await.** No raw promise chains.
9. **Use the existing shadcn components** in `src/components/ui/` — don't add new UI libs.
10. **Don't touch other agents' subtrees.** Stay in your own directory.

## Memory files to read for context

- `/home/ubuntu/.claude/projects/-home-ubuntu/memory/chorus_v05_scope.md` — locked feature list
- `/home/ubuntu/.claude/projects/-home-ubuntu/memory/chorus_strategy.md` — overall product
- `/home/ubuntu/.claude/projects/-home-ubuntu/memory/chorus_phase_schema.md` — phase primitive (doer/reviewer/inputs/iterate)
- `/home/ubuntu/.claude/projects/-home-ubuntu/memory/chorus_terminology.md` — vocabulary (Balanced column)
- `/home/ubuntu/.claude/projects/-home-ubuntu/memory/chorus_tmux_session_lifecycle.md` — fresh-per-task tmux
- `/home/ubuntu/.claude/projects/-home-ubuntu/memory/chorus_credentials_model.md` — credentials vault
- `/home/ubuntu/.claude/projects/-home-ubuntu/memory/chorus_run_page_architecture.md` — UI run page state machine

## Done = green typecheck

```
cd /home/ubuntu/dev/chorus
pnpm typecheck   # MUST pass
pnpm lint        # MUST pass
```

If your component has runtime entrypoints, smoke-test them by running them once:
- Daemon: `pnpm dev:daemon` then `curl http://127.0.0.1:7707/health`
- MCP:    `pnpm dev:mcp <<< '{"jsonrpc":"2.0","method":"initialize",...}'`
- UI:     `pnpm dev` then `curl http://127.0.0.1:3011/`
- Templates: `node -e 'console.log(JSON.parse(require("yaml").parse(require("fs").readFileSync("templates/code-review.yaml", "utf8"))))'`
