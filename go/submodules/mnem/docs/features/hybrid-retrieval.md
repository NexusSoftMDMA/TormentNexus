# Hybrid Retrieval

mnem retrieves across three lanes simultaneously and fuses the results:

1. **Vector** - HNSW nearest-neighbour search over dense embeddings. Finds semantically similar content even when wording differs.
2. **Sparse** - BM25 keyword matching. Exact term overlap; strong for proper nouns, IDs, and precise queries.
3. **Graph** - multi-hop traversal over authored edges. Follows relationships between nodes to find connected context.

Results from all active lanes are fused via Reciprocal Rank Fusion (RRF) into a single ranked list.

## Graph traversal is optional

Vector search handles most queries well. Turn graph traversal on when:

- Queries span multiple documents or entities
- You need multi-hop reasoning ("what does Alice's team own?")
- The answer requires following a chain of relationships

```bash
# Dense baseline (fast, works well for most queries)
mnem retrieve "what did we decide about rate limiting"

# Add multi-hop graph traversal
mnem retrieve "..." --graph-expand 20

# Full stack: graph + community detection + Personalised PageRank + cross-encoder rerank
mnem retrieve "..." --graph-expand 20 --community-filter --graph-mode ppr --rerank cohere:rerank-english-v3.0
```

## Flags reference

| Flag | What it does |
|---|---|
| `--graph-expand <N>` | Add N graph neighbours of top-K seeds |
| `--graph-mode ppr` | Personalised PageRank scoring (vs default decay) |
| `--community-filter` | Run Leiden community detection; drop low-coverage communities |
| `--vector-cap <N>` | Widen the dense candidate pool (default 256) |
| `--rerank <provider:model>` | Post-fusion cross-encoder rerank |

## See also

- [CLI reference - retrieve](../src/cli.md)
- [GraphRAG](../../README.md#graphrag)
- [Embedding providers](../src/guides/embed-providers.md)
