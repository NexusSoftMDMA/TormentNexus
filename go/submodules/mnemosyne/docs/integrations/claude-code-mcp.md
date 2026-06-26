# Mnemosyne + Claude Code

Give Claude Code persistent memory across sessions. It remembers your preferences, project context, and decisions.

## Setup

1. Install Mnemosyne with MCP support:

```bash
pip install "mnemosyne-memory[mcp]"
```

2. Add to `claude.json` in your project root:

```json
{
  "mcpServers": {
    "mnemosyne": {
      "command": "mnemosyne",
      "args": ["mcp"],
      "env": {}
    }
  }
}
```

3. Restart Claude Code. The tools register automatically.

## Usage

In Claude Code:

```
/remember User prefers pnpm over npm
/recall project architecture decisions
```

Or ask naturally:
- "Remember that I use 2-space indentation"
- "What do you know about our auth system?"
- "Recall the discussion about the database schema"

## CLI shortcuts

Claude Code slash commands:
- `mnemosyne_remember` — Store a memory
- `mnemosyne_recall` — Search memories
- `mnemosyne_forget` — Delete a memory

## Project isolation

Use per-project memory banks to keep contexts separate:

```json
{
  "mcpServers": {
    "mnemosyne": {
      "command": "mnemosyne",
      "args": ["mcp", "--bank", "my-project"],
      "env": {}
    }
  }
}
```
