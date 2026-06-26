# Mnemosyne + OpenAI Codex CLI

Connect Codex CLI to Mnemosyne for long-term memory across coding sessions.

## Setup

1. Install Mnemosyne with MCP support:

```bash
pip install "mnemosyne-memory[mcp]"
```

2. Add to `.codex/mcp.json` in your project root:

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

3. Restart Codex CLI. Tools appear automatically.

## Usage

Ask Codex:
- "Remember my preferred test framework is pytest"
- "Recall our discussion about migration strategy"
- "What preferences do you have stored for me?"

## Memory Banks

Codex projects can use separate memory banks to keep context isolated per project:

```json
{
  "mcpServers": {
    "mnemosyne": {
      "command": "mnemosyne",
      "args": ["mcp", "--bank", "codex-{{project-name}}"],
      "env": {}
    }
  }
}
```
