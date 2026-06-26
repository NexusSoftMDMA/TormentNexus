# Mnemosyne + Cursor

Connect Mnemosyne to Cursor as an MCP tool for persistent memory across sessions.

## Setup

1. Install Mnemosyne with MCP support:

```bash
pip install "mnemosyne-memory[mcp]"
```

2. Add to Cursor's MCP config at `.cursor/mcp.json` in your project root:

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

3. Restart Cursor. The Mnemosyne tools (remember, recall, forget) appear automatically in your agent's tool palette.

## Usage

Ask Cursor's agent:
- "Remember that I prefer tab-width 4 for Go files"
- "What do you know about my development preferences?"
- "Recall any notes about the deployment pipeline"

## Tips

- Memories persist across projects when using the default bank
- Use different banks per project for isolation:
  ```json
  {
    "mcpServers": {
      "mnemosyne": {
        "command": "mnemosyne",
        "args": ["mcp", "--bank", "project-xyz"],
        "env": {}
      }
    }
  }
  ```
- To check your memory stats: ask "how many memories do I have?"
