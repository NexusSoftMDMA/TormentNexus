# TormentNexus: Future Memory Systems Specification

This document details the architectural patterns of advanced memory systems found in the BobbyBookmarks ecosystem that are not currently implemented in TormentNexus, and provides a blueprint for their Go-native integration.

---

## 1. GraphRAG & Relational Memory Links
* **Reference Systems**: *AutoMem, MemVault, core, mcp-memory-service*
* **What it is**: Linking memory entries together with explicit relational edges (e.g. `PREFERS_OVER`, `DERIVED_FROM`, `CAUSED_BY`) instead of treating them as isolated vector points.
* **Why it matters**: Allows the agent to perform multi-hop relational queries (e.g., following a chain of dependencies to discover the context behind a user's coding preference) rather than just flat similarity scans.
* **Go Blueprint**:
  - Create a `l2_relations` table:
    ```sql
    CREATE TABLE IF NOT EXISTS l2_relations (
        source_id TEXT NOT NULL,
        target_id TEXT NOT NULL,
        relation_type TEXT NOT NULL,
        weight REAL DEFAULT 1.0,
        PRIMARY KEY (source_id, target_id, relation_type),
        FOREIGN KEY (source_id) REFERENCES l2_vault(id) ON DELETE CASCADE,
        FOREIGN KEY (target_id) REFERENCES l2_vault(id) ON DELETE CASCADE
    );
    ```

---

## 2. Asynchronous "Sleep Cycle" Consolidation
* **Reference Systems**: *MemVault, AutoMem*
* **What it is**: A periodic, heavy background worker that runs offline to optimize, deduplicate, and summarize the memory graph.
* **Why it matters**: Ingestion happens in real-time under low latency constraint. Over time, memory accumulates duplicate facts and minor contradictions. A sleep cycle aggregates detailed step-by-step logs into generalized semantic facts and prunes stale nodes.
* **Go Blueprint**:
  - Add an asynchronous ticker worker in `go/internal/memorystore/consolidator.go`.
  - Every 6 hours, it queries all memories from the last interval, runs a local LLM summarization prompt, resolves duplicates using semantic Jaccard logic, and writes clean consolidated facts back to L2.

---

## 3. Biomimetic Forgetting Curves (Temporal Decay)
* **Reference Systems**: *CortexGraph (mnemex), hippo-memory*
* **What it is**: Emulating human memory decay curves. Memory records naturally decay their relevance/heat score over time unless they are reinforced by active recall or reinforcement commands.
* **Why it matters**: Prevents the LLM context window from being flooded with ancient, low-value experiences that happen to share vector similarity.
* **Go Blueprint**:
  - Implement a decay function based on the Ebbinghaus forgetting curve formula:
    \[
    R = e^{-\frac{t}{S}}
    \]
    Where \(R\) is retrievability, \(t\) is elapsed time since last access, and \(S\) is memory strength (stability).
  - Run a nightly background update:
    ```sql
    UPDATE l2_vault 
    SET heat_score = heat_score * exp(-julianday('now') + julianday(last_accessed_at)) 
    WHERE memory_type != 'archive';
    ```

---

## 4. Cognitive Layer Partitioning
* **Reference Systems**: *Hindsight*
* **What it is**: Segmenting memory into three distinct biological tiers:
  1. **World**: Static global facts.
  2. **Experiences**: Chronological trace of specific agent actions/events.
  3. **Mental Models**: Generalized understandings distilled through self-reflection (e.g. "User gets annoyed when tabs are converted to spaces").
* **Why it matters**: Allows distinct query rules; agents can reflect on mental models directly to guide behavior instead of scanning raw logs.
* **Go Blueprint**:
  - Map this to our `memory_kind` attribute (`world_fact`, `experience`, `mental_model`) and add a reflection loop command `/reflect` that prompts the LLM to synthesize mental models from recent experiences.
