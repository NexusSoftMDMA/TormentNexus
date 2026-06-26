# Dashboard

## Project Structure

tormentnexus Orchestrator is a TypeScript monorepo using Bun as the runtime.

```
packages/
├── server/              # Hono backend API (port 3847)
│   └── src/
│       ├── index.ts              # Entry point
│       ├── routes/               # API route handlers
│       │   ├── council.ts        # Council management
│       │   ├── sessions.ts       # Session CRUD
│       │   ├── cli.ts            # CLI tool detection
│       │   ├── health.ts         # Health monitoring
│       │   ├── env.ts            # Environment variables
│       │   ├── hooks.ts          # Webhooks
│       │   └── smart-pilot.ts    # Auto-continue
│       ├── services/             # Business logic
│       │   ├── session-manager.ts
│       │   ├── council.ts
│       │   ├── cli-registry.ts
│       │   ├── health-monitor.ts
│       │   ├── log-rotation.ts
│       │   ├── environment-manager.ts
│       │   ├── hook-manager.ts
│       │   ├── smart-pilot.ts
│       │   └── session-persistence.ts
│       └── supervisors/          # LLM adapters
│           ├── openai.ts
│           ├── anthropic.ts
│           ├── google.ts
│           ├── deepseek.ts
│           ├── grok.ts
│           ├── qwen.ts
│           ├── kimi.ts
│           └── mock.ts
├── cli/                 # Ink (React) terminal UI
│   └── src/
│       ├── app.tsx
│       ├── components/
│       └── hooks/
└── shared/              # Shared TypeScript types
    └── src/
        └── types.ts

public/
└── index.html           # Web dashboard (vanilla JS)
```

## Supervisor Status

| Provider | Status | Class | API Key Env Var |
|----------|--------|-------|-----------------|
| **OpenAI** | ✅ Active | `OpenAISupervisor` | `OPENAI_API_KEY` |
| **Anthropic** | ✅ Active | `AnthropicSupervisor` | `ANTHROPIC_API_KEY` |
| **Google** | ✅ Active | `GoogleSupervisor` | `GEMINI_API_KEY` |
| **DeepSeek** | ✅ Active | `DeepSeekSupervisor` | `DEEPSEEK_API_KEY` |
| **Grok** | ✅ Active | `GrokSupervisor` | `GROK_API_KEY` / `XAI_API_KEY` |
| **Qwen** | ✅ Active | `QwenSupervisor` | `QWEN_API_KEY` |
| **Kimi** | ✅ Active | `KimiSupervisor` | `KIMI_API_KEY` / `MOONSHOT_API_KEY` |
| **Mock** | ✅ Testing | `MockSupervisor` | (none) |

## Supported CLI Tools

| CLI | Auto-Detect | Command |
|-----|-------------|---------|
| OpenCode | ✅ Yes | `opencode` |
| Claude Code | ✅ Yes | `claude` |
| Aider | ✅ Yes | `aider` |
| Cursor | ✅ Yes | `cursor` |
| Continue | ✅ Yes | `continue` |
| Cody | ✅ Yes | `cody` |
| Copilot | ✅ Yes | `github-copilot` |

## Consensus Modes

| Mode | Description |
|------|-------------|
| `simple-majority` | >50% approval |
| `supermajority` | ≥67% approval |
| `unanimous` | 100% approval |
| `weighted` | Weight × confidence voting |
| `ceo-override` | Lead supervisor can override |
| `ceo-veto` | Lead supervisor can veto |
| `hybrid-ceo-majority` | CEO override + majority fallback |
| `ranked-choice` | Ranked preference voting |

## Quick Commands

```bash
# Install dependencies
bun install

# Build shared types (required first)
cd packages/shared && bun run build

# Start server (development)
cd packages/server && bun run dev

# Start CLI dashboard
cd packages/cli && bun run dev

# Run tests
cd packages/server && bun test
```

## Latest Build

- **Version**: 1.1.0
- **Date**: January 2026
- **Runtime**: Bun
- **Framework**: Hono
