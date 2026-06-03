## 🏗️ Technical Architecture

### Overview
TormentNexus Orchestrator is the council and session-supervision subsystem inside the TormentNexus monorepo. This document describes the original standalone architecture that informed the integrated runtime now hosted in TormentNexus core and the web dashboard.
tormentnexus Orchestrator is the council and session-supervision subsystem inside the tormentnexus monorepo. This document describes the original standalone architecture that informed the integrated runtime now hosted in tormentnexus core and the web dashboard.

### Monorepo Structure

```
packages/
├── server/          # Hono + Bun backend API (port 3847)
│   └── src/
│       ├── index.ts           # Entry point, Hono app setup
│       ├── routes/            # API route handlers
│       ├── services/          # Core business logic
│       └── supervisors/       # LLM provider adapters
├── cli/             # Ink (React) terminal UI dashboard
│   └── src/
│       ├── app.tsx            # Main TUI application
│       ├── components/        # React components for TUI
│       └── hooks/             # Custom React hooks
├── shared/          # TypeScript types shared across packages
│   └── src/
│       └── types.ts           # Shared interfaces and types
public/
└── index.html       # Web dashboard (single-file vanilla JS)
```

### Core Components

#### 1. Backend (`packages/server/`)

**Entry Point (`index.ts`)**
- Hono web framework with Bun runtime
- Serves static frontend from `public/`
- WebSocket support for real-time updates
- Auto-registers supervisors from environment API keys

**Routes (`routes/`)**
| File | Purpose |
|------|---------|
| `council.ts` | Council management, debate triggers, supervisor config |
| `sessions.ts` | Session CRUD, bulk operations, templates, tags |
| `cli.ts` | CLI tool detection and registry |
| `health.ts` | Health monitoring, auto-recovery |
| `env.ts` | Environment variable management |
| `hooks.ts` | Webhook registration and management |
| `smart-pilot.ts` | Auto-continue configuration |

**Services (`services/`)**
| File | Purpose |
|------|---------|
| `session-manager.ts` | Process spawning, monitoring, guidance delivery |
| `council.ts` | Supervisor management, debate orchestration, consensus |
| `cli-registry.ts` | Auto-detect CLI tools (opencode, claude, aider, etc.) |
| `health-monitor.ts` | Session health checks, auto-recovery |
| `log-rotation.ts` | Log pruning by count and age |
| `environment-manager.ts` | Per-session and global env vars |
| `hook-manager.ts` | Webhook lifecycle management |
| `smart-pilot.ts` | Auto-continue logic |
| `session-persistence.ts` | Save/restore sessions on restart |

**Supervisors (`supervisors/`)**
LLM provider adapters implementing the `ISupervisor` interface:
- `openai.ts` - GPT-4o
- `anthropic.ts` - Claude
- `google.ts` - Gemini
- `deepseek.ts` - DeepSeek
- `grok.ts` - Grok (xAI)
- `qwen.ts` - Qwen (Alibaba)
- `kimi.ts` - Kimi (Moonshot)
- `mock.ts` - Mock supervisor for testing

#### 2. CLI Dashboard (`packages/cli/`)

Terminal UI built with Ink (React for CLI):
- Real-time supervisor status display
- Session management controls
- Debate history viewer
- Smart Pilot controls

#### 3. Web Dashboard (`public/index.html`)

Single-file vanilla JS application:
- Dark theme UI (Zinc/Shadcn style)
- WebSocket real-time updates
- Session health monitoring
- Council configuration panel

#### 4. Shared Types (`packages/shared/`)

TypeScript interfaces used across packages:
- `Session`, `SessionStatus`, `SessionHealth`
- `Supervisor`, `SupervisorVote`, `DebateResult`
- `CouncilConfig`, `ConsensusMode`
- `CLITool`, `CLIType`
- `HookPhase`, `HookPayload`

### 🔄 Data Flow

1. **Session Start**: User starts session via API/Dashboard
2. **Process Spawn**: SessionManager spawns CLI process on unique port (4096+)
3. **Monitoring**: Health monitor polls sessions every 10 seconds
4. **AI Turn Detection**: Manager detects when AI finishes a turn
5. **Council Debate**: 
   - Pre-debate hooks fire (can block)
   - Supervisors review context in parallel
   - Debate rounds with round-robin discussion
   - Consensus calculation based on mode (weighted, CEO, etc.)
   - Post-debate hooks fire (can modify decision)
6. **Guidance Delivery**:
   - Pre-guidance hooks fire (can modify)
   - Guidance sent to CLI session as user prompt
   - Post-guidance hooks fire
7. **Smart Pilot**: If enabled and approved, auto-continue cycle repeats

### 🗳️ Consensus Modes

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

### 🛠️ Configuration

**Environment Variables**
```bash
# Supervisor API Keys (auto-registers)
OPENAI_API_KEY, ANTHROPIC_API_KEY, DEEPSEEK_API_KEY,
GEMINI_API_KEY, GROK_API_KEY, QWEN_API_KEY, KIMI_API_KEY

# Server Config
TORMENTNEXUS_ORCHESTRATOR_PORT=3847
TORMENTNEXUS_ORCHESTRATOR_HOST=0.0.0.0
TORMENTNEXUS_ORCHESTRATOR_BASE_PORT=4096
TORMENTNEXUS_ORCHESTRATOR_DEBATE_ROUNDS=2
TORMENTNEXUS_ORCHESTRATOR_CONSENSUS=0.7
TORMENTNEXUS_ORCHESTRATOR_SMART_PILOT=false
TORMENTNEXUS_ORCHESTRATOR_PORT=3847
TORMENTNEXUS_ORCHESTRATOR_HOST=0.0.0.0
TORMENTNEXUS_ORCHESTRATOR_BASE_PORT=4096
TORMENTNEXUS_ORCHESTRATOR_DEBATE_ROUNDS=2
TORMENTNEXUS_ORCHESTRATOR_CONSENSUS=0.7
TORMENTNEXUS_ORCHESTRATOR_SMART_PILOT=false

# Legacy compatibility aliases
AUTOPILOT_PORT=3847
AUTOPILOT_HOST=0.0.0.0
AUTOPILOT_BASE_PORT=4096
AUTOPILOT_DEBATE_ROUNDS=2
AUTOPILOT_CONSENSUS=0.7
AUTOPILOT_SMART_PILOT=false
```

**Config File** (`.tormentnexus/orchestrator/config.json`, with `.autopilot/config.json` still supported as a legacy fallback)
**Config File** (`.tormentnexus/orchestrator/config.json`, with `.autopilot/config.json` still supported as a legacy fallback)
```json
{
  "council": {
    "supervisors": [...],
    "debateRounds": 2,
    "consensusThreshold": 0.7,
    "consensusMode": "weighted",
    "weightedVoting": true,
    "smartPilot": false
  }
}
```

### 🔌 Hook System

Webhooks intercept the debate/guidance flow at 5 phases:
1. `pre-debate` - Before debate (can block)
2. `post-debate` - After debate (can modify decision)
3. `pre-guidance` - Before sending guidance (can modify)
4. `post-guidance` - After guidance sent
5. `on-error` - On error during flow

### 📊 Health Monitoring

- Automatic health checks every 30 seconds
- Session states: `healthy`, `unhealthy`, `degraded`, `unknown`
- Auto-recovery attempts for unhealthy sessions
- Metrics: response time, memory usage, error count
