# Versioned Memory

mnem treats every write as a commit - the same way git treats every save as a snapshot.

Every node and edge you add becomes part of a content-addressed commit in a Merkle DAG. The graph has branches you can switch between, a log you can walk backwards, and a diff you can run between any two points in time.

Two agents working on the same graph offline? When they reconnect, mnem reconciles their histories through a 3-way merge - no fact is silently overwritten, and the full provenance of every node is preserved.

## What this means in practice

```bash
mnem diff HEAD <cid>           # see exactly what an agent added or removed in a session
mnem log                       # walk the op-log backwards, operation by operation
mnem branch create experiment  # experiment without touching the main graph
mnem merge experiment          # fold results back in when you're satisfied
mnem revert <cid>              # undo a bad batch of facts without losing the audit trail
```

## Why it matters

Standard agent-memory systems are write-only append logs or search indexes with no history. You can't see what changed between sessions, you can't undo a bad ingest, and you can't run two experiments in parallel.

mnem's commit model gives you the same safety net your code already has.

## See also

- [CLI reference](../src/cli.md)
- [Core concepts](../src/core-concepts.md)
- [mnem diff / mnem show](../src/cli.md)
