# Provider Configuration

mnem ships with a bundled ONNX MiniLM-L6-v2 embedder that runs in-process. No Ollama, no API keys, no configuration needed to get started.

When you're ready to switch - to a larger model, a hosted API, or a different architecture - change one section in `.mnem/config.toml`. No rebuild, no fork.

## Supported providers

| Provider | Type | Notes |
|---|---|---|
| ONNX (bundled) | Embedding | Default; runs in-process, no network |
| Ollama | Embedding | Local; requires Ollama running |
| OpenAI | Embedding | Hosted; requires `OPENAI_API_KEY` |
| Cohere | Embedding + Rerank | Hosted; requires `COHERE_API_KEY` |
| Voyage | Rerank | Hosted; requires `VOYAGE_API_KEY` |
| SPLADE-ONNX | Sparse | Local; feature-gated |

## Example config

```toml
[embed]
provider = "ollama"
model    = "nomic-embed-text"
base_url = "http://localhost:11434"
```

API keys are read from environment variables - never written to disk:

```bash
export OPENAI_API_KEY=sk-...
mnem config set embed.provider openai
mnem config set embed.model text-embedding-3-small
```

## Switching providers

After changing the embedder, existing nodes need their embeddings recomputed:

```bash
mnem reindex          # recompute embeddings for all nodes with the new provider
```

## See also

- [Configuration reference](../src/configuration.md)
- [Embedding providers guide](../src/guides/embed-providers.md)
