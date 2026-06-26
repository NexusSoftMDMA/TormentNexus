# v2.8.0 Coordinator Review — Hermes v0.13.0 catch-up

**Status:** all 8 work-stream plans drafted; WS-1 (capability flags) committed on branch `ws-1-capabilities-v0.13` (PR #80). This document is the coordinator's cross-stream review compiled from each per-stream plan's _Open Questions_ section, file inventory, and confidence rating. It exists so the user can review the v2.8.0 implementation surface in one read instead of eight.

## Plan inventory

| Stream | File | Lines | Confidence | Open Q's | Files touched | Branch |
| --- | --- | --: | --- | --: | --: | --- |
| WS-2 | [WS-2-goals-and-queue-plan.md](WS-2-goals-and-queue-plan.md) | ~600 | medium-high | 7 | ~6 | `ws-2-goals-and-queue` |
| WS-3 | [WS-3-kanban-v0.13-plan.md](WS-3-kanban-v0.13-plan.md) | 947 | medium-high | 7 | 12 (1 new) | `ws-3-kanban-v0.13` |
| WS-4 | [WS-4-curator-archive-plan.md](WS-4-curator-archive-plan.md) | 561 | medium-high | 6 | ~10 | `ws-4-curator-archive` |
| WS-5 | [WS-5-gateway-v0.13-plan.md](WS-5-gateway-v0.13-plan.md) | 520 | medium-high | 8 | ~17 (5 new) | `ws-5-gateway-v0.13` |
| WS-6 | [WS-6-providers-v0.13-plan.md](WS-6-providers-v0.13-plan.md) | 625 | high (arch) / medium (key) | 5 | 8 | `ws-6-providers-v0.13` |
| WS-7 | [WS-7-settings-v0.13-plan.md](WS-7-settings-v0.13-plan.md) | 628 | medium-high | 8 | 17 | `ws-7-settings-v0.13` |
| WS-8 | [WS-8-ux-v0.13-plan.md](WS-8-ux-v0.13-plan.md) | 580 | high (5 of 6) / medium (1) | 5 | 12 | `ws-8-ux-v0.13` |
| WS-9 | [WS-9-ios-v0.13-plan.md](WS-9-ios-v0.13-plan.md) | 926 | medium-high | 8 | 7 | `ws-9-ios-v0.13` |

**Total v2.8.0 surface:** ~89 files touched (with overlap; net unique ~75), ~5400 lines of plan, 54 distinct open questions across 8 streams.

## Cross-stream collisions (coordinator-tracked)

These files appear in more than one work-stream and need explicit sequencing:

| File | Streams | Resolution |
| --- | --- | --- |
| `RichChatViewModel.swift` | WS-2 (`/goal`/`/queue`), WS-8 (`/new <name>` help text) | WS-8 lands AFTER WS-2; the `/new <name>` change is one-line and rebases trivially. |
| `SessionInfoBar` (chat status bar) | WS-2 (queue chip), WS-8 (compression count) | Both add SwiftUI children to the same HStack — order-independent. WS-8 lands after WS-2 to avoid file-level conflicts. |
| `HermesCapabilities.swift` | WS-1 (all flags), WS-8 + WS-9 (request `isV013OrLater` helper) | Decided: add `isV013OrLater` helper to WS-1 PR (one-line, lands cleanly). See _Decision A_ below. |
| `HermesConfig` model | WS-5 (gateway allowlists), WS-6 (`image_gen.model`, `openrouter.response_cache`), WS-7 (mcp/cron/web-tools/profiles) | Each work-stream extends a different namespace. Touch the same file; merge resolution mechanical. |
| iOS surfaces | WS-9 consumes WS-2/WS-3/WS-4/WS-5 model fields | WS-9 lands LAST in the v2.8.0 cycle. Hard sequencing constraint. |

## Open-questions matrix (cluster-organized)

Of 54 questions across the 8 plans, **45 are wire-shape unknowns** that can only be resolved by inspecting a real Hermes v0.13.0 install (i.e. they need a v0.13 host to dogfood against, since the release notes don't pin every CLI flag, JSON field, or YAML key). The remaining 9 are Scarf-side architectural choices that the agents already recommended; they need user adjudication.

### Cluster A — wire-shape unknowns (resolve at integration time, not before implementation starts)

These are the questions where each plan agent gave a best-inference default, marked the spot with a `// TODO` comment, and recommended verification when a v0.13 host is reachable. The implementation can proceed safely with these defaults; if any are wrong, the fix is a one-line edit + a new test fixture.

- **WS-2:** goal-state read-back channel (Q1), `/queue --clear` syntax (Q2), `/queue` argument shape (Q5), `/goal` non-interruptive on the wire (Q7)
- **WS-3:** hallucination verb name (Q1), diagnostics location (task vs run, Q2), `set_max_retries` post-create (Q3), failure-counter unification field (Q4), darwin-zombie kind (Q5), default `max_retries` value (Q6), `kanban diagnose <id>` verb (Q7)
- **WS-4:** `prune --dry-run` flag (Q1), `--json` on read verbs (Q2), single-skill prune (Q3), sync-run timeout (Q4)
- **WS-5:** Google Chat platform identifier (Q1), allowlist YAML key path (Q2), `gateway list --json` shape (Q3), `[[as_document]]` discoverability (Q6)
- **WS-6:** `openrouter.response_cache.enabled` exact key (Q1), default value (Q2), grok rename old-slot redirect (Q4), `models_dev_cache.json` refresh on clean install (Q5)
- **WS-7:** MCP transport names (Q1), `sse_read_timeout` default (Q2), `--transport sse` flag spelling (Q3), `--no-agent` toggle-off shape (Q4), argparse + `--no-agent` (Q5), web-tools backend lists (Q6), `web_tools.backend` legacy fallback (Q7), `--no-skills` × `--clone-all` interaction (Q8)
- **WS-8:** compression-count wire field name (Q1), xAI TTS config keys (Q2), `display.language` empty-string vs `"en"` default (Q3)

**Recommended resolution:** proceed with implementation against the agents' inferred defaults. Each implementation agent should be briefed to mark its TODO callsites. A coordinator pass before merging WS-2…WS-9 (after the user has dogfooded a v0.13 host) confirms or fixes each in <30 minutes total.

### Cluster B — Scarf-side architectural choices (need user adjudication)

These are the 9 questions where the user's input directly shapes the implementation:

| ID | Question | Agent's recommendation |
| --- | --- | --- |
| **A** | Add `isV013OrLater` helper to WS-1? | **Yes** — both WS-8 and WS-9 want it. One-line addition. Land in the existing WS-1 PR before merging. |
| **B** | "Auto-resumed from checkpoint" indicator | **Defer to v2.8.1** (WS-2 Q3). Hermes v0.13's auto-resume signal isn't documented; surfacing it requires a wire-format we don't have yet. |
| **C** | `/queue --clear` button when syntax unconfirmed | **Remove the "Clear all" button from the queue popover until syntax is confirmed.** Local-only clear that lies about server state is worse than no button. |
| **D** | Curator prune confirm UX | **Custom sheet matching template-uninstall** (WS-4 Q5). Enumerated list + asymmetric keyboard shortcut, no typed-name confirmation. |
| **E** | Filter Yuanbao + Teams platforms on pre-v0.12? | **Keep current behavior** (WS-5 Q4). Don't change v0.12 host UX in a v0.13 work-stream. Document the asymmetry. |
| **F** | Capability flag for slash-command notice TTL | **Proxy through `hasGatewayBusyAckToggle ‖ hasGatewayRestartNotification`** (WS-5 Q5). A dedicated flag is YAGNI. |
| **G** | Rename `MessagingGatewayViewModel`? | **Apply rename if <5 callsites change.** Otherwise keep the type name and rely on user-facing label. |
| **H** | Profile `--no-skills` + `--clone-all` interaction | **Conservative: disable `--no-skills` toggle when `--clone-all` is on.** Argparse may reject anyway. |
| **I** | Implementation parallelism — 8 PRs in parallel worktrees, or sequential review? | Recommend **parallel worktree implementation** with **sequential coordinator review** (one PR at a time merging into main). Parallel impl = ~3-4 days of agent-time; sequential review = the natural throttle for production safety. |

### Cluster C — out-of-scope deferrals (no decision needed)

These were identified during planning but the agents already deferred them with sound rationale:

- WS-2: optimistic-vs-authoritative goal reconciliation
- WS-3: failure-counter unification field rendering
- WS-6: Arcee Trinity Large Thinking temperature/compression overrides surface
- WS-7: `web_tools.backend` legacy migration prompt
- WS-9: deep-links from v0.13-features sheet, hallucination-badge tap-target alert
- All streams: iOS write surfaces (always deferred)

## Recommended next steps (post-review)

Once the user resolves Cluster B questions A–I:

1. **Patch WS-1 PR #80** with the `isV013OrLater` helper (Decision A). One commit, one push.
2. **Spawn 8 implementation agents in parallel** (Decision I), each in an isolated worktree:
   - Each agent gets its plan file + the answers to relevant Cluster B questions + the WS-1 commit ref.
   - Each agent produces a single PR from its branch.
   - Branch names match the plan inventory table.
3. **Coordinator-review each PR sequentially** in dependency order:
   - Wave 1 (WS-2, WS-3, WS-4, WS-5) — review one at a time, merge in any order
   - Wave 2 (WS-6, WS-7, WS-8) — same
   - Wave 3 (WS-9) — last; consumes Wave 1+2 model fields
4. **WS-10 release** after WS-9 merges:
   - Update CLAUDE.md (already partially done in WS-1)
   - Update wiki via `scripts/wiki.sh`
   - Write `releases/v2.8.0/RELEASE_NOTES.md`
   - Run `scripts/release.sh v2.8.0 --draft` to validate
   - Run `scripts/release.sh v2.8.0` for the full promotion

## Risk register

- **Production app, thousands of users.** Each PR must build clean, all tests green, manual smoke against a v0.13 host before merge.
- **Cluster A wire-shape risk.** Mitigated by tolerant decoders + capability gates; if any guess is wrong, pre-v0.13 hosts still work and v0.13 hosts surface a benign decode-failure (UI hides instead of crashes).
- **Sparkle update path.** v2.8.0 is delivered via the existing Sparkle appcast; there's no migration path for users on pre-v0.12 Hermes hosts (their v0.13-only surfaces stay hidden).
- **No data migrations.** Per CLAUDE.md, schema is unchanged from v0.11/v0.12 across this release. Per-project `manifest.json` and Scarf-owned sidecars at `~/.hermes/scarf/` are untouched.

## Estimate

- WS-1: shipped (PR #80 awaiting merge after Decision A)
- Wave 1 implementation: ~3 days agent-time × 4 streams in parallel = ~3 calendar days
- Wave 2 implementation: ~2 days agent-time × 3 streams in parallel = ~2 calendar days
- WS-9 implementation: ~2 days agent-time
- WS-10 release coordination: ~½ day

**Calendar-time estimate: ~8 days** with parallel implementation + sequential review. The bottleneck is coordinator review at PR-merge boundaries, not agent throughput.
