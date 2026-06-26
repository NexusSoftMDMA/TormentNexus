# mnem Format Specification

Version: 0.1  
Status: canonical - implementations MUST conform.

Two independent implementations that follow this document must produce
byte-identical objects for the same input. That property is the whole point.

---

## §1 Scope

This document covers:

- §2  Content identifiers (CIDs)
- §3  Canonical encoding (DAG-CBOR)
- §4  Object types (Node, Edge, Commit, Operation, View, Tombstone, …)
- §5  Prolly tree chunking algorithm
- §6  Concurrency model (CAS, linearize)
- §7  Repository layout and backend contract
- §8  Retrieval protocol
- §9  Cryptographic signatures and revocation

---

## §2 Content Identifiers

mnem uses **CIDv1** (the [multiformats CID spec](https://github.com/multiformats/cid))
to address every persistent object.

### §2.1 Structure

A CID bundles three self-describing fields:

| Field     | Value for structured objects | Value for raw blobs |
|-----------|------------------------------|---------------------|
| Version   | 1                            | 1                   |
| Codec     | `0x71` (DAG-CBOR)            | `0x55` (Raw)        |
| Multihash | SHA-256 or BLAKE3 of content | same                |

Two CIDs are equal if and only if all three fields match.

### §2.2 Text form

The canonical text form is **base32 lowercase** with a `b` multibase prefix,
e.g. `bafkreicxxxxxxx…`. Implementations MUST use this encoding when
serializing CIDs as strings (JSON values, CLI output, log lines).

### §2.3 Stable identifiers

Named `NodeId`, `EdgeId`, `ChangeId`, `OperationId`. Encoded as **16-byte
byte strings** (major type 2 in CBOR). They survive content mutations -
editing a node's text does not change its `NodeId`. Do not confuse stable
identifiers with content-addressed CIDs.

---

## §3 Canonical Encoding

### §3.1 DAG-CBOR

All structured objects (Nodes, Edges, Trees, Commits, Operations, Views) are
serialized as **DAG-CBOR** for hashing and on-disk storage. The content hash
is computed over the canonical CBOR bytes. Every CID for a structured object
uses codec `0x71`.

Requirements:

- Map keys MUST be sorted (lexicographic, as required by DAG-CBOR).
- Integer encoding MUST use the smallest form.
- Indefinite-length items are forbidden.
- The `dagjson` representation is a debug/inspection format only - it is
  never hashed, never stored as canonical content.

### §3.2 Forward-compatibility extension map

Every object type carries an `extra` field typed as a CBOR map with string
keys. Decoders MUST preserve unknown `extra` keys and MUST round-trip them
unchanged. This allows additive fields without a major version bump.

---

## §4 Object Types

### §4.1 Node

The unit of content. Fields:

| Field       | Type                     | Notes                                       |
|-------------|--------------------------|---------------------------------------------|
| `id`        | NodeId (16 bytes)        | Stable; set once at creation                |
| `text`      | string                   | The semantic content                        |
| `label`     | string                   | Scope namespace; empty string = unlabelled  |
| `metadata`  | map(string → any)        | Caller-defined opaque tags                  |
| `embedding` | absent on the Node object | Lives in the per-commit sidecar bucket      |
| `extra`     | map(string → any)        | Forward-compat; see §3.2                    |

The embedding is stored in a sidecar, not on the node CID, so two nodes with
identical text but different embedding models share one CID.

### §4.2 Edge

A directed labelled edge between two nodes. The hash of the edge does not
depend on node properties - editing a node's text does not invalidate its
outgoing edges.

Fields: `id` (EdgeId), `src` (NodeId), `dst` (NodeId), `label` (string),
`metadata`, `extra`.

### §4.3 Prolly tree chunks

The node and edge stores are Merkle DAGs of `TreeChunk` objects. Two chunk
kinds exist: `"leaf"` and `"internal"`. The `_kind` discriminator MUST be
present. Chunks with `_kind` absent or with mismatched leaf/internal
invariants are rejected.

Chunk CIDs use codec `0x71`. See §5 for the chunking algorithm.

### §4.4 Commit

A commit is an immutable snapshot of the graph. Fields:

| Field       | Type            | Notes                                      |
|-------------|-----------------|--------------------------------------------|
| `id`        | ChangeId        | Stable logical change identity             |
| `parent`    | CID or null     | Parent commit; null for the root           |
| `root`      | CID             | Root chunk of the Prolly tree at this time |
| `timestamp` | uint64 (ms UTC) | Wall time at commit creation               |
| `message`   | string          | Human-readable                             |
| `signature` | bytes or absent | See §9.1                                   |
| `extra`     | map             | §3.2                                       |

### §4.5 Operation

The unit of the op-log. Represents a single logical mutation (ingest, edit,
tombstone). Keys in the operation's `payload` map MUST be strings (CBOR major
type 3).

### §4.6 View

A snapshot of the mutable state of a repository: the set of ref heads and the
revocation list. The `kind` discriminator distinguishes `"normal"` from
`"conflicted"` ref targets:

- `"normal"`: a single head CID.
- `"conflicted"`: two or more concurrent head CIDs (lexicographic tie-break
  as a secondary sort; see conflict detector B4.2).

The root view (empty repository) has empty heads and empty refs.

### §4.8 Secondary index

The top-level ANN / vector index aggregator. Added in mnem/0.2. Absent in
0.1.x repositories; decoders MUST treat absence as an empty index.

### §4.9 Edge storage order

Edges in a leaf chunk are stored sorted by `(label, src_or_dst, edge_cid)`.
Implementations MUST preserve this sort order when writing chunks.

### §4.10 Tombstones

A tombstone marks a node as logically deleted ("forget"). Tombstone-wins
semantics apply: a tombstone always supersedes a live node entry. Tombstones
are durable - they survive merges, compactions, and replica sync.

---

## §5 Prolly Tree Chunking

### §5.1 Boundary probability

mnem uses a **logistic CDF** to decide chunk boundaries. The target mean chunk
size and the CDF parameters are fixed constants; see `mnem-core` source.

### §5.2 Rolling hash window

Window size: **64 bytes = four 16-byte keys**. The boundary hash is computed
over the last 64 bytes of the concatenation of keys in order.

---

## §6 Concurrency Model

### §6.1 Heads

A repository's "current state" is the set of op-heads per `§7.3`'s insertion
ordering.

### §6.4 Compare-and-swap on refs

`update_ref(expected_cid, new_cid)` is the only mutation path for refs.
**No lost update**: if two writers race, exactly one wins and the other
observes an unexpected CID, retries or returns a conflict. Implementations
MUST NOT update a ref without first verifying `expected_cid`.

### §6.5 Linearize mode

When `linearize = true` (opt-in), the implementation re-reads op-heads
immediately before the CAS attempt to reduce the conflict window. Default is
`false`.

---

## §7 Repository Layout and Backend Contract

### §7.3 Write ordering

New entries MUST be inserted before delete/tombstone markers for the same
key in a single batch. This ordering (`new-then-delete`) ensures the
repository remains consistent if the process crashes mid-write.

### §7.5 Fresh repository initialization

A newly initialized repository has:

- One root View with empty heads and empty refs.
- No commits, no nodes, no edges.

The root View CID is the sole ref in the initial state.

---

## §8 Retrieval Protocol

`retrieve` fans out across up to three lanes and fuses ranked results:

1. **Vector** - HNSW over per-commit sidecar embeddings.
2. **Sparse** - BM25 / SPLADE (feature-gated; absent = disabled).
3. **Graph** - n-hop traversal over authored edges, optional PPR scoring.

The `SPEC §retrieve.response-budget` governs the aggregate byte cap on
returned content. Implementations MUST respect this budget.

---

## §9 Cryptographic Signatures and Revocation

### §9.1 Signing

Signatures are **Ed25519** and are computed over the **canonical DAG-CBOR**
encoding of the commit with the `signature` field absent. The signed bytes
are the same bytes that determine the commit CID.

The `signature` field, when present, is a 64-byte byte string.

### §9.2 Revocation

A key entry in the revocation list has a `revoked_at` timestamp.
Commits signed by the key whose `timestamp <= revoked_at` remain **valid**.
Only commits with `timestamp > revoked_at` are rejected.
