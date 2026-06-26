---
name: arc-sync
description: >
  Manage MCP servers via Arc Relay. Use this skill for ANY MCP server operation:
  adding, removing, listing, syncing, configuring, or troubleshooting MCP servers.
  Triggers on: MCP, .mcp.json, .codex/config.toml, server configuration, missing tools, relay.
  Do NOT manually edit .mcp.json or .codex/config.toml - always use arc-sync commands instead.
user-invocable: true
disable-model-invocation: false
allowed-tools: Bash(arc-sync *)
argument-hint: [list|add <server>|remove <server>|reset|status|setup-claude|setup-codex|setup-project|server add|server remove|server start|server stop]
---

# MCP Server Management via arc-sync

MCP servers in this environment are managed by Arc Relay. **Never edit `.mcp.json` or `.codex/config.toml` directly** - use `arc-sync` commands.

## First-run check

Before running any arc-sync command, check if arc-sync is configured:

1. Run `arc-sync status --json 2>/dev/null`. If it fails or returns `{"error": ...}`:
   - Check if `arc-sync` binary exists: `which arc-sync`
   - If not installed: tell the user to ask their admin for an install command (invite token), or run `arc-sync init` if they have credentials
   - If installed but not configured: run `arc-sync init` to set up

2. If status succeeds but shows missing Claude integration, suggest:
   - `arc-sync setup-claude` for personal skill installation
   - `arc-sync setup-codex` for personal Codex CLI instructions
   - `arc-sync setup-project` for team-shared project instructions

## Current project status
!`arc-sync status --json 2>/dev/null || echo '{"error": "arc-sync not installed or not configured - run: arc-sync init"}'`

## Commands

| Command | Description |
|---------|-------------|
| `arc-sync` | Interactive sync - add new relay servers to project |
| `arc-sync list` | Show all servers (use `--json` for machine-readable) |
| `arc-sync add <name>` | Add a specific server to this project |
| `arc-sync remove <name>` | Remove a server from this project |
| `arc-sync reset` | Clear the skip list for this project |
| `arc-sync status` | Show config and project details |
| `arc-sync setup-claude` | Install Claude Code skill and CLAUDE.md instructions |
| `arc-sync setup-codex` | Install Codex CLI AGENTS.md instructions |
| `arc-sync setup-project` | Add MCP instructions to project .claude/CLAUDE.md and AGENTS.md |
| `arc-sync server add <name> --type remote <url>` | Register a remote MCP server on the relay |
| `arc-sync server add <name> --type stdio --build python --package <pkg>` | Register an auto-build server |
| `arc-sync server add <name> --type stdio --image <img>` | Register a Docker stdio server |
| `arc-sync server add <name> --type http --image <img> --port <p>` | Register a Docker HTTP server |
| `arc-sync server remove <name>` | Delete a server from the relay |
| `arc-sync server start <name>` | Start a relay server |
| `arc-sync server stop <name>` | Stop a relay server |

Use `--non-interactive` or `-y` for automation. Use `--dry-run` for preview.

## Session sync behavior

`arc-sync` manages `.mcp.json` and `.codex/config.toml`, which are project MCP configs for Claude Code and Codex CLI. Agent sessions load MCP connections at session start, so changes do not take effect in the current conversation.

**After `arc-sync remove <name>`:**
- Stop using that server's tools for the rest of this conversation. Treat it as disconnected.
- Tell the user to start a new conversation if they need to confirm it's fully gone.

**After `arc-sync add <name>` or interactive sync that adds servers:**
- The new server's tools will NOT be available in this conversation.
- Tell the user to start a new conversation to pick up the new tools.

**After `arc-sync server stop <name>`:**
- Stop using that server's tools. The proxy will reject requests to a stopped server.

## When to use this skill

Use this skill whenever the conversation involves:
- **MCP** servers, tools, configuration, or `.mcp.json`
- **Codex** MCP configuration or `.codex/config.toml`
- Adding, removing, or configuring tool servers
- Missing tools or "tool not found" errors
- Server health, status, or connectivity issues
- Any mention of relay or server management

**Always prefer `arc-sync` over manually editing `.mcp.json` or `.codex/config.toml`.**

## Usage from arguments

If arguments are provided, run: `arc-sync $ARGUMENTS`
Otherwise, run `arc-sync list` first to show status, then ask the user what they'd like to do.
