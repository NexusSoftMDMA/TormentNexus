# TormentNexus/TormentNexus Ideas & Future Enhancements

## Phase 2: Autonomy Loop (Self-Healing)
- [ ] **Verify Loop**: After applying a fix via `HealerService`, automatically run relevant tests. If tests fail, feed the new error back into the healer for a second iteration.
- [ ] **Idle Healing**: Background agent that scans for known bugs or TODOs during low CPU usage periods and proposes fixes to the L2 Vault.
- [ ] **Fix Confidence Scoring**: Rank suggested fixes based on historical "success heat" of similar patterns.

## Phase 3: Skill Intelligence
- [ ] **Progressive Skill Disclosure**: Apply the same LRU/Ranking system used for tools to the system's "Skills" (instruction sets).
- [ ] **Skill Evolution**: Automatically refine skill instructions based on model performance metrics.

## Enterprise Features
- [ ] **DLP Layer**: Implement a regex/LLM-based PII scrubber that intercepts prompt streams before they reach public providers.
- [ ] **Audit Ledger**: Store hash-verified logs of every tool call in a dedicated SQLite table for compliance.
- [ ] **Fleet Dashboard**: Visualizer for pi-pods/vLLM cluster status.

## UI/UX
- [ ] **Reasoning HUD**: A "Live Graph" of the model's current thought process, showing tool selection and memory retrieval in real-time.
- [ ] **Heat Map**: Visualize memory "Heat" in the dashboard, showing which areas of the codebase or knowledge base are currently "hot."
