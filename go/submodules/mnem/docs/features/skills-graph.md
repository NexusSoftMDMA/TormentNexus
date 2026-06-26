# Skills Graph

Agent conventions today live in flat files - `.cursorrules`, `AGENTS.md`, `CLAUDE.md`. They're useful, but they have no structure: you can't query them, diff them against a teammate's, branch them for an experiment, or merge two versions together.

mnem replaces them with a versioned, branchable, mergeable knowledge graph.

## What you can do

```bash
# Commit a convention as a typed, queryable node
mnem add node --label Convention -s "All REST endpoints are versioned under /v1/"

# Later: retrieve it in any session, on any machine
mnem retrieve "API versioning convention"

# Export your graph, share it, import a teammate's
mnem export my-conventions.car
mnem import teammate-conventions.car

# Diff the two before merging
mnem diff HEAD <their-commit-cid>

# Merge selectively
mnem merge their-branch --strategy=theirs
```

## Why this matters

A flat file is either entirely present or entirely absent in a session. A graph node is retrievable by meaning - the agent finds the convention when it's relevant, not because you remembered to include the file.

Conventions also evolve. With flat files, you lose the history of why a rule exists. With mnem commits, every change has a timestamp, an author, and an optional message.

## See also

- [mnem add](../src/cli.md)
- [mnem retrieve](../src/cli.md)
- [mnem export / import](../src/cli.md)
