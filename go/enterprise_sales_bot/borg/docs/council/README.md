# cli-orchestrator Council

A multi-model AI council for autonomous development guidance. Multiple AI supervisors debate and vote on code changes through a democratic process.

This document describes the legacy standalone council architecture that has now been assimilated into the TormentNexus monorepo. The canonical product name is **cli-orchestrator**. Legacy `AUTOPILOT_*` environment variables remain supported for compatibility, but new deployments should prefer `TORMENTNEXUS_ORCHESTRATOR_*`. This is no longer documented as a separate `tormentnexus-orchestrator` repository.
This document describes the legacy standalone council architecture that has now been assimilated into the tormentnexus monorepo. The canonical product name is **cli-orchestrator**. Legacy `AUTOPILOT_*` environment variables remain supported for compatibility, but new deployments should prefer `TORMENTNEXUS_ORCHESTRATOR_*`. This is no longer documented as a separate `tormentnexus-orchestrator` repository.

## Architecture

```
packages/
├── server/     # Hono + Bun backend API
├── cli/        # Ink (React) terminal UI  
└── shared/     # TypeScript types
```

## Features

- **Multi-Model Council**: GPT-4, Claude, Gemini, DeepSeek, Grok, Qwen, Kimi
- **Weighted Voting**: Supervisor weights + confidence scores
- **8 Consensus Modes**: simple-majority, supermajority, unanimous, weighted, ceo-override, ceo-veto, hybrid-ceo-majority, ranked-choice
- **Smart Pilot**: Auto-continue when council approves
- **Hook System**: Intercept debate/guidance flow via webhooks
- **Multi-CLI Support**: Auto-detect opencode, claude, aider, cursor, continue, cody, copilot
- **Session Health Monitoring**: Automatic health checks with auto-recovery
- **Session Templates**: Pre-configured templates (default, review, debug, feature)
- **Tag Management**: Organize sessions with tags
- **Environment Variables**: Per-session and global environment variable management
- **Log Rotation**: Automatic log pruning by count and age
- **TUI Dashboard**: Real-time supervisor status and debate history
- **Web Dashboard**: Modern dark UI with real-time WebSocket updates
- **Session Persistence**: Auto-save and restore sessions on restart

## Quick Start

```bash
# Install dependencies
bun install

# Set at least one API key
export OPENAI_API_KEY="sk-..."

# Start server
cd packages/server && bun run dev

# In another terminal, start CLI
cd packages/cli && bun run dev
```

## Environment Variables

### Supervisor API Keys (auto-registers supervisors)
```bash
OPENAI_API_KEY          # GPT-4o supervisor
ANTHROPIC_API_KEY       # Claude supervisor  
DEEPSEEK_API_KEY        # DeepSeek supervisor
GEMINI_API_KEY          # Gemini supervisor
GROK_API_KEY            # Grok supervisor (or XAI_API_KEY)
QWEN_API_KEY            # Qwen supervisor
KIMI_API_KEY            # Kimi supervisor (or MOONSHOT_API_KEY)
```

### Server Configuration
```bash
TORMENTNEXUS_ORCHESTRATOR_PORT=3847           # API server port
TORMENTNEXUS_ORCHESTRATOR_HOST=0.0.0.0        # Bind address
TORMENTNEXUS_ORCHESTRATOR_BASE_PORT=4096      # Session base port
TORMENTNEXUS_ORCHESTRATOR_DEBATE_ROUNDS=2     # Number of debate rounds
TORMENTNEXUS_ORCHESTRATOR_CONSENSUS=0.7       # Consensus threshold (70%)
TORMENTNEXUS_ORCHESTRATOR_SMART_PILOT=false   # Enable auto-continue
TORMENTNEXUS_ORCHESTRATOR_PORT=3847           # API server port
TORMENTNEXUS_ORCHESTRATOR_HOST=0.0.0.0        # Bind address
TORMENTNEXUS_ORCHESTRATOR_BASE_PORT=4096      # Session base port
TORMENTNEXUS_ORCHESTRATOR_DEBATE_ROUNDS=2     # Number of debate rounds
TORMENTNEXUS_ORCHESTRATOR_CONSENSUS=0.7       # Consensus threshold (70%)
TORMENTNEXUS_ORCHESTRATOR_SMART_PILOT=false   # Enable auto-continue

# Legacy compatibility aliases still accepted:
AUTOPILOT_PORT=3847
AUTOPILOT_HOST=0.0.0.0
AUTOPILOT_BASE_PORT=4096
AUTOPILOT_DEBATE_ROUNDS=2
AUTOPILOT_CONSENSUS=0.7
AUTOPILOT_SMART_PILOT=false
```

## API Endpoints

### Council
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/council/status` | Get council status and config |
| POST | `/api/council/config` | Update full council configuration |
| POST | `/api/council/debate` | Trigger debate on a task |
| POST | `/api/council/toggle` | Enable/disable council |
| POST | `/api/council/add-mock` | Add mock supervisor (testing) |
| POST | `/api/council/supervisors` | Add real supervisors |
| DELETE | `/api/council/supervisors` | Remove all supervisors |
| GET | `/api/council/consensus-modes` | List available consensus modes |
| POST | `/api/council/consensus-mode` | Set consensus mode |
| POST | `/api/council/lead-supervisor` | Set lead supervisor (CEO) |
| POST | `/api/council/fallback-chain` | Set fallback supervisor chain |
| POST | `/api/council/supervisor-weight` | Set individual supervisor weight |

### Sessions
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/sessions` | List all sessions |
| GET | `/api/sessions/active` | List active sessions |
| GET | `/api/sessions/stats` | Get session statistics |
| GET | `/api/sessions/templates` | List session templates |
| GET | `/api/sessions/persisted` | Get persisted sessions list |
| GET | `/api/sessions/by-cli/:cliType` | List sessions by CLI type |
| GET | `/api/sessions/by-tag/:tag` | Filter sessions by tag |
| GET | `/api/sessions/by-template/:template` | Filter sessions by template |
| GET | `/api/sessions/:id` | Get specific session |
| POST | `/api/sessions/start` | Start new session |
| POST | `/api/sessions/from-template/:name` | Start session from template |
| POST | `/api/sessions/:id/stop` | Stop session |
| POST | `/api/sessions/:id/resume` | Resume stopped session |
| DELETE | `/api/sessions/:id` | Delete session |
| POST | `/api/sessions/:id/guidance` | Send guidance to session |
| GET | `/api/sessions/:id/logs` | Get session logs |
| GET | `/api/sessions/:id/logs/export` | Export session logs (json/csv/text) |
| GET | `/api/sessions/logs/export-all` | Export all logs |
| PUT | `/api/sessions/:id/tags` | Set session tags (replace all) |
| POST | `/api/sessions/:id/tags` | Add tag to session |
| DELETE | `/api/sessions/:id/tags/:tag` | Remove tag from session |
| POST | `/api/sessions/bulk/start` | Bulk start sessions |
| POST | `/api/sessions/bulk/stop` | Bulk stop all sessions |
| POST | `/api/sessions/bulk/resume` | Bulk resume sessions |

### CLI Tools
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/cli/tools` | List all detected CLI tools |
| GET | `/api/cli/tools/available` | List available (installed) CLI tools |
| GET | `/api/cli/tools/:type` | Get specific CLI tool info |
| POST | `/api/cli/tools/refresh` | Re-scan for CLI tools |
| POST | `/api/cli/tools/custom` | Register custom CLI tool |

### Health Monitoring
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health/sessions` | Get health status of all sessions |
| GET | `/api/health/sessions/:id` | Get health status of specific session |
| GET | `/api/health/stats` | Get session, log, and health statistics |
| GET | `/api/health/server` | Get server health (uptime, memory) |
| POST | `/api/health/sessions/:id/check` | Force health check on session |

### Environment Variables
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/env/sessions/:id` | Get session environment variables |
| POST | `/api/env/sessions/:id` | Set session environment variable |
| DELETE | `/api/env/sessions/:id/:key` | Remove session environment variable |
| GET | `/api/env/global` | Get global environment variables |
| POST | `/api/env/global` | Set global environment variable |
| DELETE | `/api/env/global/:key` | Remove global environment variable |

### Dynamic Supervisor Selection
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/dynamic-selection/status` | Get status and statistics |
| POST | `/api/dynamic-selection/toggle` | Enable/disable dynamic selection |
| GET | `/api/dynamic-selection/profiles` | List supervisor profiles |
| GET | `/api/dynamic-selection/profiles/:name` | Get specific supervisor profile |
| POST | `/api/dynamic-selection/profiles` | Add/update supervisor profile |
| DELETE | `/api/dynamic-selection/profiles/:name` | Remove supervisor profile |
| GET | `/api/dynamic-selection/templates` | List team templates |
| GET | `/api/dynamic-selection/templates/:name` | Get specific team template |
| POST | `/api/dynamic-selection/templates` | Add/update team template |
| DELETE | `/api/dynamic-selection/templates/:name` | Remove team template |
| POST | `/api/dynamic-selection/detect` | Detect task type from description |
| POST | `/api/dynamic-selection/select` | Select optimal team for a task |
| POST | `/api/dynamic-selection/sync-supervisors` | Sync available supervisors from council |

### Human Veto (Council Chair)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/veto/status` | Get veto status and statistics |
| POST | `/api/veto/toggle` | Enable/disable human veto |
| GET | `/api/veto/config` | Get veto configuration |
| POST | `/api/veto/config` | Update veto configuration |
| GET | `/api/veto/pending` | List pending decisions awaiting veto |
| GET | `/api/veto/pending/:id` | Get specific pending decision |
| POST | `/api/veto/pending/:id/approve` | Approve a pending decision |
| POST | `/api/veto/pending/:id/reject` | Reject a pending decision |
| POST | `/api/veto/pending/:id/redebate` | Request re-debate for a decision |
| GET | `/api/veto/history` | Get veto decision history |

### Plugins
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/plugins/status` | Get plugin system status |
| GET | `/api/plugins/config` | Get plugin configuration |
| POST | `/api/plugins/config` | Update plugin configuration |
| GET | `/api/plugins/list` | List all loaded plugins |
| GET | `/api/plugins/:name` | Get specific plugin details |
| POST | `/api/plugins/load` | Load plugin from path or manifest |
| POST | `/api/plugins/load-all` | Load all plugins from directory |
| DELETE | `/api/plugins/:name` | Unload a plugin |
| POST | `/api/plugins/:name/reload` | Reload a plugin (hot-reload) |
| POST | `/api/plugins/:name/enable` | Enable a plugin |
| POST | `/api/plugins/:name/disable` | Disable a plugin |
| GET | `/api/plugins/supervisors/all` | Get all supervisors from all plugins |
| GET | `/api/plugins/:name/supervisors` | Get supervisors from specific plugin |
| POST | `/api/plugins/validate` | Validate a plugin manifest |
| GET | `/api/plugins/sample-manifest` | Get sample plugin manifest template |

### Debate History
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/debate-history/status` | Get status and statistics |
| GET | `/api/debate-history/config` | Get configuration |
| POST | `/api/debate-history/config` | Update configuration |
| POST | `/api/debate-history/toggle` | Enable/disable history persistence |
| GET | `/api/debate-history/stats` | Get detailed debate statistics |
| GET | `/api/debate-history/list` | Query debate records with filters |
| GET | `/api/debate-history/debates/:id` | Get specific debate record |
| DELETE | `/api/debate-history/debates/:id` | Delete a debate record |
| GET | `/api/debate-history/supervisor/:name` | Get supervisor voting history |
| GET | `/api/debate-history/export/json` | Export history as JSON |
| GET | `/api/debate-history/export/csv` | Export history as CSV |
| DELETE | `/api/debate-history/clear` | Clear all debate history |
| POST | `/api/debate-history/initialize` | Initialize/reload from disk |

### Smart Pilot
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/smart-pilot/status` | Get smart pilot status |
| POST | `/api/smart-pilot/toggle` | Toggle auto-continue |
| POST | `/api/smart-pilot/config` | Update settings |

### Hooks
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/hooks` | List registered hooks |
| POST | `/api/hooks/register` | Register webhook |
| DELETE | `/api/hooks/:id` | Unregister hook |

### WebSocket
Connect to `ws://localhost:3847/ws` for real-time updates:
- `session_update` - Session status changes
- `session_health` - Session health status changes
- `council_decision` - Debate results
- `log` - Log entries
- `error` - Error messages

## CLI Usage

```bash
# Start TUI dashboard
cd packages/cli && bun run dev
```

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `1` | Dashboard view |
| `2` | Logs view |
| `3` | Council view |
| `4` | Smart Pilot view |
| `r` | Refresh |
| `t` | Toggle council |
| `p` | Toggle smart pilot |
| `q` | Quit |

## Weighted Voting

Each supervisor has:
- **Weight** (0-2x): Prioritizes certain supervisors
- **Confidence** (0-1): Self-reported certainty in vote

**Weighted Consensus Formula:**
```
consensus = Σ(approved × weight × confidence) / Σ(weight)
```

Strong dissent (rejection with confidence > 0.7) blocks auto-approval.

## Hook System

Register webhooks to intercept the debate flow:

```bash
# Register a pre-debate hook
curl -X POST http://localhost:3847/api/hooks/register \
  -H "Content-Type: application/json" \
  -d '{
    "phase": "pre-debate",
    "webhookUrl": "http://my-service/hook",
    "priority": 10
  }'
```

**Hook Phases:**
- `pre-debate` - Before debate starts (can block)
- `post-debate` - After debate, before guidance (can modify decision)
- `pre-guidance` - Before sending guidance (can modify)
- `post-guidance` - After guidance sent
- `on-error` - On error during flow

**Webhook Response Format:**
```json
{
  "continue": true,
  "modifiedGuidance": { ... },
  "modifiedDecision": { ... },
  "reason": "Optional explanation"
}
```

## Testing

```bash
# Run unit tests (18 council tests)
cd packages/server && bun test src/services/__tests__/

# Run integration tests (13 API tests, requires running server)
cd packages/server && bun test src/__tests__/

# Run E2E test
bun run scripts/test-e2e.ts
```

## Development

```bash
# Typecheck all packages
bun run typecheck

# Build shared types
cd packages/shared && bun run build

# Dev server with hot reload
cd packages/server && bun run dev

# Dev CLI with hot reload
cd packages/cli && bun run dev
```

## Supervisor Configuration

### Adding via API
```bash
curl -X POST http://localhost:3847/api/council/supervisors \
  -H "Content-Type: application/json" \
  -d '{
    "supervisors": [{
      "name": "GPT-4o",
      "provider": "openai",
      "model": "gpt-4o",
      "weight": 1.5
    }]
  }'
```

### Adding via Config File
Create `.autopilot/config.json`:
```json
{
  "council": {
    "supervisors": [
      {
        "name": "GPT-4o",
        "provider": "openai",
        "model": "gpt-4o",
        "weight": 1.5
      },
      {
        "name": "Claude",
        "provider": "anthropic",
        "model": "claude-3-5-sonnet-20241022",
        "weight": 1.0
      }
    ],
    "debateRounds": 2,
    "consensusThreshold": 0.7,
    "weightedVoting": true,
    "smartPilot": false
  }
}
```

## License

MIT
