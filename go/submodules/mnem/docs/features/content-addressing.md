# Content Addressing

Every object in mnem - nodes, edges, commits, tree chunks - has an address derived from its content. The same bytes always produce the same address (CID), on any machine.

This isn't a slogan. It means:

- **Determinism**: two independent systems ingesting the same document produce byte-identical graphs.
- **Deduplication**: identical content automatically collapses to one CID; storage doesn't grow from repeated ingests of the same data.
- **Auditability**: you can verify any object by recomputing its CID from its bytes. Nothing can be silently altered.
- **Replayability**: given the same input, you can reproduce any historical state of the graph from scratch.

## How it works

Objects are encoded as DAG-CBOR (a canonical binary format) and hashed with BLAKE3. The resulting CID uniquely identifies the content - the same content always produces the same CID, and different content never produces the same one.

Stable identifiers (`NodeId`, `EdgeId`) are separate from CIDs. Editing a node's text changes its CID (new content) but not its `NodeId` (stable logical identity). This lets you track a node's history across mutations.

## See also

- [Format Specification](../SPEC.md)
- [Core concepts](../src/core-concepts.md)
