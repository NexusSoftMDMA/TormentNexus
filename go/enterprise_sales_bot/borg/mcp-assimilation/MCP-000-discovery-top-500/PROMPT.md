# Task: MCP-000 - Discovery: Build Top 500 MCP Server Ranked List

## Goal
Parse `~/.tormentnexus/mcp-cache.json` (the full 26MB catalog with ~7000 entries) and
`~/.tormentnexus/mcp.json` (60 active high-priority servers) to build a ranked list of
the top 500 MCP servers to assimilate as native Go modules.

## Steps

### 1. Parse mcp-cache.json safely (use Python, not PowerShell - file has duplicate keys)

```python
import json, sqlite3, os
from pathlib import Path

# Handle duplicate keys by keeping last value
def parse_with_duplicates(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        text = f.read()
    # Use object_pairs_hook to collect all entries
    entries = []
    def handler(pairs):
        entries.extend(pairs)
        return dict(pairs)
    json.loads(text, object_pairs_hook=handler)
    return entries

cache_path = Path.home() / '.tormentnexus' / 'mcp-cache.json'
active_path = Path.home() / '.tormentnexus' / 'mcp.json'
```

### 2. Score and rank each server
- Active in mcp.json: +50 pts
- Has GitHub URL with real repo: +20 pts
- Name contains high-value keywords (filesystem, github, search, browser, db, memory, code, ai, llm): +15 pts
- Not a duplicate name variant: +15 pts

### 3. Output to mcp-assimilation/TOP_500_SERVERS.json
Format: `[{"rank": 1, "name": "...", "github_url": "...", "score": 95, "priority": "critical|high|medium"}, ...]`

### 4. Insert all 500 into the state tracking DB
DB path: `data/assimilation_state.db`
Table: `mcp_servers` (name, github_url, status='pending')

Also check existing Go tools in `go/internal/tools/` - if a server is already implemented,
set status='implemented' in the DB and skip it from the queue.

### 5. Print final summary
- Total entries in mcp-cache.json
- Top 500 selected
- Already implemented: N
- Pending implementation: N
