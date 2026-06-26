# Handoff: TormentNexus Extension v0.7.2
# Handoff: tormentnexus Extension v0.7.2

> **Read [AGENTS.md](./AGENTS.md) first.** It contains all universal instructions.
> This file documents the latest session state for incoming agents.

## Session Summary (2026-02-26)

### What Was Done
1. **Documentation Overhaul**: Created `TODO.md`, `MEMORY.md`, `DEPLOY.md`. Rewrote `AGENTS.md` with universal LLM instructions. Updated `IDEAS.md`, `DASHBOARD.md`, `ROADMAP.md`, `CHANGELOG.md`.
2. **Code Quality Audit**: Fixed hardcoded version `'0.1.0'` in `app.store.ts`, replaced unsafe `require()` with static import in `ui.store.ts`, eliminated 5× `as any` casts, added missing `UserPreferences` fields (`accentColor`, `autoExecuteWhitelist`), added 7 missing icon types+SVGs to `Icon.tsx`.
3. **Prompt Templates Feature** (Phase 4 roadmap): New sidebar tab for saving/reusing prompts. Created `prompt.store.ts` (Zustand CRUD store), `PromptTemplates.tsx` (list/editor/search/copy-to-clipboard), wired into `Sidebar.tsx` tab system.
4. **Version Bump**: `0.7.1` → `0.7.2`. All package.json files, VERSION file, and CHANGELOG updated.

### Current State
- **Version**: `0.7.2`
- **Build**: Passes 12/12 (Chrome/Edge). Firefox build also works via `pnpm build:firefox`.
- **Git**: All changes committed and pushed to `main` (d8f6d3d).

### Next Priorities (from ROADMAP.md Phase 4)
1. **Notification Center** — Wire up the existing `config.store.ts` remote notification and feature flag infrastructure to a visible UI panel.
2. **Resource Browser** — Browse MCP server resources (schema, templates).
3. **Test Suite** — Set up Vitest for stores/services, Playwright for e2e.
4. **Accessibility Audit** — `axe-core` scan of the sidebar Shadow DOM.

### Known Issues (from TODO.md)
- `mcpHandler.ts` has ~200 lines of dead reconnection code (commented out).
- Dual theme sync between `ui.store.ts` and `app.store.ts` (comment says "Ensure this logic is robust").
- Macro/Context stores live in `lib/` instead of `stores/` (inconsistency).
- Plugin registry is a no-op stub in `app.store.ts`.
- Pre-existing lint errors in `Sidebar.tsx` (lines 347-348: `context:save` event type, line 424: `toggleCommandPalette`, line 1140: `Tool` type mismatch).

### Critical Files
| File | Purpose |
|------|---------|
| `pages/content/src/components/sidebar/Sidebar.tsx` | Main sidebar (1260 lines, tabs, state) |
| `pages/content/src/stores/*.ts` | 11 Zustand stores |
| `pages/content/src/plugins/adapters/*.ts` | 16 platform adapters |
| `chrome-extension/src/background/index.ts` | Background service worker |
| `pages/content/src/render_prescript/` | DOM rendering & tool detection |
