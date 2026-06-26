# Changelog

All notable changes to this project will be documented in this file.

## [0.7.2] - 2026-02-26

### Added
- **Prompt Templates**: New sidebar tab for saving and reusing common AI prompts with search, copy-to-clipboard, and full CRUD. (`prompt.store.ts`, `PromptTemplates.tsx`)
- **7 new sidebar icons**: `arrow-left`, `download`, `save`, `plus`, `git-branch`, `clock`, `database` added to `Icon.tsx`.
- **Documentation**: Created `TODO.md`, `MEMORY.md`, `DEPLOY.md`. Rewrote `AGENTS.md` with universal LLM instructions. Updated `IDEAS.md`.

### Fixed
- **Hardcoded version in app.store.ts**: Changed `'0.1.0'` event emission to read from `chrome.runtime.getManifest()`.
- **Unsafe `require()` in ui.store.ts**: Replaced dynamic `require('./config.store')` with static import.
- **Type safety**: Eliminated 5× `as any` casts in ui.store.ts notification handling.
- **Missing `UserPreferences` fields**: Added `accentColor` and `autoExecuteWhitelist` to the TypeScript interface.
- **Logger naming**: Changed logger name from `'initializePluginRegistry'` to `'AppStore'`.

## [0.7.1] - 2026-02-12

### Added
- **Input, Textarea, Select UI components**: Created missing sidebar form components (`Input.tsx`, `Textarea.tsx`, `Select.tsx`) used by ContextManager and MacroBuilder.
- **icon-16.png**: Generated 16×16 icon for extension manifest (resized from icon-128).

### Fixed
- **CRLF line endings**: Converted bash scripts (`update_version.sh`, `set_global_env.sh`, `copy_env.sh`) to LF. Added `.gitattributes` to prevent recurrence.
- **Duplicate variable declaration**: Removed unused `root` variable in Sidebar.tsx accent color theming.
- **Invalid JSX escape**: Fixed backslash escape in MacroBuilder.tsx placeholder attribute using template literal syntax.
- **Missing UI barrel exports**: Added `Input`, `Textarea`, `Select` exports to `ui/index.ts` so ContextManager and MacroBuilder imports resolve.

## [0.7.0] - 2026-02-12

### Added
- **Dashboard Version Display**: The Dashboard tab now shows the current version badge prominently.
- **Keyboard Shortcuts Reference**: The Dashboard includes a quick-reference card for all keyboard shortcuts.

### Fixed
- **Trusted Tools Whitelist**: The Safety & Whitelist UI in Settings now persists trusted tools to the store instead of using ephemeral local state. The "Add" button is now fully functional.
- **`trustedTools` Type**: Added `trustedTools?: string[]` to the `UserPreferences` interface, resolving previously silent type errors.
- **Keyboard Shortcuts Hook**: Removed broken `useSidebarState` import and unused `useToastStore` reference from `useKeyboardShortcuts.ts`.
- **Version Sync**: Synchronized all version references (VERSION, package.json, DASHBOARD.md, Settings export) to 0.7.0.
- **Version Script**: `update_version.sh` now writes to the `VERSION` file in addition to updating package.json files.

### Changed
- **Documentation Overhaul**: Rewrote `AGENTS.md` as comprehensive universal LLM instructions. Created `GEMINI.md` and `GPT.md`. Expanded `VISION.md`, `ROADMAP.md`, and `DASHBOARD.md`.

## [0.6.0] - 2024-05-22

### Added
- **Analytics Dashboard**: New sidebar tab showing high-level usage stats (Total Runs, Success Rate, Most Used Tool).
- **Rich Renderer**: Visualizes tool outputs in Activity Logs (JSON tree, Markdown, Images).
- **Keyboard Shortcuts**: Global shortcuts for power users (`Alt+Shift+S` toggle sidebar, `/` search, etc.).
- **Auto-Execute Whitelist**: Safe mode for automation, allowing only trusted tools to run automatically.

## [0.5.9] - 2024-05-22

### Added
- **Activity Log**: Comprehensive logging system with UI.
- **Notifications**: Toast notification system.
- **Settings**: Enhanced UI with sliders and grouping.
- **Documentation**: New `VISION.md`, `ROADMAP.md`, and agent instruction files.

### Changed
- **Server Status**: Added ping latency test.
- **Tools**: Added favorites and sorting.
