# Mnemosyne + OpenClaw

Use Mnemosyne as the memory backend for your OpenClaw agents. Two integration paths:

## Path 1: Native Memory Provider (recommended)

Install the OpenClaw extra:

```bash
pip install "mnemosyne-memory[openclaw]"
```

Add to your OpenClaw `config.yaml`:

```yaml
memory:
  provider: mnemosyne.integrations.openclaw:create_provider
  config:
    db_path: /path/to/memory/data
    bank: my-agent
    enable_triples: true
```

That's it. OpenClaw discovers and loads the provider automatically.

## Path 2: MCP Server (any MCP-compatible client)

If your OpenClaw version uses MCP for tools, use the standard Mnemosyne MCP server:

1. Install MCP support:
```bash
pip install "mnemosyne-memory[mcp]"
```

2. Add MCP config to your OpenClaw setup:
```bash
mnemosyne mcp --transport sse --port 8080
```

3. Configure OpenClaw to connect to the MCP endpoint.

## Provider API

The Mnemosyne provider exposes these operations:

| Method | Description |
|--------|-------------|
| `store(key, content, metadata)` | Store a memory |
| `retrieve(key)` | Get memory by key |
| `search(query, limit)` | Semantic search |
| `query(query, params)` | Filtered query with date/source/topic |
| `delete(key)` | Remove memory by key |
| `get_stats()` | Provider statistics |
| `health()` | Health check |

## Filtered Queries

```python
provider = MnemosyneProvider(config={"bank": "my-agent"})

# Date range
results = provider.query("project discussion",
    params={"from_date": "2026-01-01", "to_date": "2026-06-01"})

# By source
results = provider.query("bugs",
    params={"source": "code-review"})

# Triple query (if enable_triples=true)
results = provider.query("triple:prefers")
```

## Configuration

| Config Key | Default | Description |
|------------|---------|-------------|
| `db_path` | `~/.hermes/mnemosyne/data/` | Path to memory data directory |
| `bank` | `openclaw` | Memory bank name for isolation |
| `enable_triples` | `false` | Enable TripleStore for graph queries |
