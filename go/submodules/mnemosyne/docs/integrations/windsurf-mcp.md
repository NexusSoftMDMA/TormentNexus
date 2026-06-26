# Mnemosyne + Windsurf

Add persistent memory to Windsurf via MCP.

## Setup

1. Install Mnemosyne with MCP support:

```bash
pip install "mnemosyne-memory[mcp]"
```

2. Add to `.windsurf/mcp_config.json` in your project root:

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

3. Restart Windsurf. Tools register automatically.

## Usage

Ask Windsurf's agent:
- "Remember my commit message style"
- "What did we learn about the API rate limits?"
- "Recall any stored context about this codebase"
