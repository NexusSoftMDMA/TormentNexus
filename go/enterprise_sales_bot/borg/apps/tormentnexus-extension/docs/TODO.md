# TODO: Short-Term Tasks & Bug Fixes

> This file tracks fine-grained, actionable items. For long-term structural plans, see [ROADMAP.md](./ROADMAP.md).

## Critical / High Priority

- [x] **Plugin Registry is a placeholder**: `app.store.ts:10` — `initializePluginRegistry` is a no-op stub. Either implement the plugin system or remove the placeholder to avoid confusion.
- [x] **Feature Flags UI**: `config.store.ts` has a full `FeatureFlag` system with rollout percentages, targeting (versions, regions, user segments), and remote notifications — but **none of it is exposed in the Settings UI**. Wire it up or document it as internal-only.
- [x] **Remote Notifications not surfaced**: `config.store.ts` has `RemoteNotification` with actions, targeting, and campaign support. The store logic exists but there is no UI component to display or manage these notifications.
- [x] ~~**Dead reconnection code in mcpHandler.ts**~~: Replaced 1096 lines of dead code with 17-line deprecation notice in v0.7.2.
- [x] **Macro store not in stores/ directory**: The macro store lives in `lib/macro.store.ts` instead of `stores/macro.store.ts`. Consider relocating for consistency with the 10 other stores.
- [x] **Context store not in stores/ directory**: Same issue — `lib/context.store.ts` should be moved to `stores/` for discoverability.

## Medium Priority

- [x] ~~**Prompt Templates**~~: Implemented in v0.7.2 (prompt.store.ts + PromptTemplates.tsx + Sidebar integration).
- [x] **Resource Browser**: Implement a tab to browse MCP server resources (Phase 4 roadmap item).
- [x] **MANUAL.md refresh**: The user manual (`docs/MANUAL.md`) likely needs updating to cover Macros, Context Manager, Command Palette, and new sidebar features added since v0.6.0.
- [x] **Accessibility audit**: Run `axe-core` against the sidebar Shadow DOM and fix violations to achieve WCAG 2.1 AA compliance.
- [x] **Virtual scrolling for Activity Log**: `ActivityLog` renders all items. For lists >1000 entries, implement virtual scrolling (e.g., `react-virtuoso`).
- [x] **Test suite setup**: Add Vitest unit tests for stores and services; Playwright for e2e sidebar interactions.
## Low Priority / Nice to Have

- [ ] **i18n activation**: The `packages/i18n` infrastructure exists but has no translations loaded. Generate English strings file and wire it up.
- [ ] **Cloud Sync**: Sync macros and context across devices via Chrome Sync API or external backend.
- [ ] **Multi-Proxy**: Connect to multiple MCP servers simultaneously.
- [x] **Tool Chaining Visual Builder**: Replace the linear Macro steps with a node-based graph editor.
- [x] **Clean up UNIVERSAL_LLM_INSTRUCTIONS.md**: This file exists at the root of `docs/` and may duplicate or conflict with `AGENTS.md`. Consolidate or remove.

## Code Hygiene

- [x] ~~**Remove 500+ commented lines from mcpHandler.ts**~~: Removed 1096-line `mcpHandler.ts` and 933-line `backgroundCommunication.ts`. Both replaced with deprecation notices. Original code in git history.
- [x] **Consolidate duplicate theme sync logic**: `ui.store.ts:340-360` subscribes to `app.store` for theme sync. The comment says "Ensure this logic is robust or handled by a single source of truth for theme." — resolve this ambiguity.
- [x] **`helpers.ts:91` temporary visual indicator**: A "temporary visual indicator to help identify Shadow DOM boundaries" is still in production code. Remove or guard behind a dev flag.
