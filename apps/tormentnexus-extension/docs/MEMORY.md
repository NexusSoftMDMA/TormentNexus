# Memory: Codebase Observations & Design Preferences

> This file captures ongoing observations about the codebase, patterns, gotchas, and design preferences discovered during development.

## Architecture Patterns

- **Zustand + Persist + DevTools**: Every store uses `create<State>()(devtools(persist(...)))`. Always follow this pattern for new stores.
- **`useShallow` for hooks**: Component hooks in `hooks/useStores.ts` use `useShallow` to prevent unnecessary re-renders. Always compose via these hooks, not direct store access in components.
- **Non-React access**: Use `useXxxStore.getState()` for accessing store state outside React (e.g., in services, background scripts).
- **Shadow DOM isolation**: The sidebar renders inside a Shadow DOM root. Tailwind `important: true` is required to override host page styles. All CSS must be self-contained.
- **Adapter pattern**: Each supported AI platform has a file in `plugins/adapters/`. The base class is `base.adapter.ts`. New platform support = new adapter file + registration in the adapter store.

## Build System

- **Turbo tasks**: Build runs 12 parallel tasks. If any fail, the entire build fails. Always run `pnpm build` before pushing.
- **Firefox transform**: `ManifestParser` in `packages/dev-utils/lib/manifest-parser/impl.ts` converts `service_worker` → `scripts` when `IS_FIREFOX=true`. Both builds share `dist/`.
- **Bash scripts on Windows**: `.gitattributes` forces LF line endings on `*.sh` files. Without this, bash scripts break on Windows due to CRLF.
- **Node.js requirement**: `engines.node >= 22.12.0`. TypeScript 5.8.1-rc is used (release candidate).

## Design Preferences (from user)

- **Version in one place**: The `VERSION` file at root is the single source of truth. `pnpm update-version` propagates it to all `package.json` files.
- **Historical versioning preference**: Earlier repo guidance treated nearly every change as version-bump-worthy. Current TormentNexus canon is more selective: bump versions and changelogs when the change is release-relevant, user-visible at that level, or explicitly requested.
- **Historical versioning preference**: Earlier repo guidance treated nearly every change as version-bump-worthy. Current tormentnexus canon is more selective: bump versions and changelogs when the change is release-relevant, user-visible at that level, or explicitly requested.
- **Commit message must reference version**: e.g., `feat: Add Dashboard shortcuts (v0.7.1)`.
- **Autonomous operation**: Commit, push, and continue to the next task without pausing when possible.
- **Document everything**: All findings, changes, and decisions should be reflected in documentation.
- **UI completeness**: Every implemented feature must be fully represented in the UI with labels, descriptions, and tooltips.
- **Handoff readiness**: Leave the project in a state where another agent can pick up seamlessly.

## Known Gotchas

1. **`config.store.ts` is overengineered**: It has feature flags, remote notifications, user targeting, and campaign support — but none of it is wired to any UI or backend. It's infrastructure waiting for a use case.
2. **`mcpHandler.ts` has ~200 lines of dead code**: Commented-out reconnection logic that predates the current `connection.store.ts` system.
3. **Plugin registry is a stub**: `app.store.ts:10` — `initializePluginRegistry` is a no-op. The plugin marketplace concept exists in IDEAS.md but has no implementation.
4. **Macro/Context stores live in `lib/`**: Unlike the other 10 stores in `stores/`, these two live in `lib/`. This is a discoverability issue for new developers/agents.
5. **Theme has dual control**: Both `ui.store.ts` and `app.store.ts` manage theme state with a subscription sync between them. The comment in `ui.store.ts:340` warns this needs resolution.
