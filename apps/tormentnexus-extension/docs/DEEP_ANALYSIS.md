# Deep Technical Analysis & Architecture Overview

**Version**: 0.7.1
**Date**: 2026-02-12

This document provides a comprehensive technical analysis of the `tormentnexus-extension` codebase, intended for advanced AI models and senior developers. It covers architectural patterns, state management, data flow, and specific feature implementations.
This document provides a comprehensive technical analysis of the `tormentnexus-extension` codebase, intended for advanced AI models and senior developers. It covers architectural patterns, state management, data flow, and specific feature implementations.

## 1. Architecture Overview

The project is a **Chrome/Edge/Firefox Extension** built with a **React** frontend (injected via Content Script) and a background service worker. It follows a "Thick Client" architecture where most logic resides in the React app (Sidebar), communicating with a local MCP Proxy via the background script.

### Key Components

*   **`chrome-extension/`**: The extension shell.
    *   **Background Script (`src/background/`)**: Acts as the bridge between the Content Script and the MCP Proxy (SSE/WebSocket/Streamable HTTP). It handles connection management, heartbeats, analytics, and Context Menu events.
    *   **MCP Client (`src/mcpclient/`)**: Protocol implementation with plugin-based transport (SSE, WebSocket, Streamable HTTP).
    *   **Manifest (`manifest.ts`)**: Defines permissions (`storage`, `clipboardWrite`, `contextMenus`) and host matches for 13 platforms.
*   **`pages/content/`**: The core application.
    *   **`Sidebar.tsx`**: The main React entry point. Manages visibility, theming, routing (14 tabs), push mode, and resize.
    *   **`SidebarManager.tsx`**: Shadow DOM management, injection lifecycle.
    *   **Plugin Adapters (`src/plugins/adapters/`)**: 16 per-platform adapters (ChatGPT, Gemini, Grok, DeepSeek, OpenRouter, Kimi, etc.).
    *   **Render Prescript (`src/render_prescript/`)**: DOM rendering, function call detection, and streaming tool call parsing.
    *   **Services (`src/services/`)**: Business logic (AutomationService for auto-execute).

### Transport Layer

The extension supports three transport protocols to communicate with MCP servers:
1. **SSE** (Server-Sent Events) — default, `http://localhost:3006/sse`
2. **WebSocket** — `ws://localhost:3006/message`
3. **Streamable HTTP** — `http://localhost:3006/mcp`

## 2. State Management (Zustand)

We use **Zustand** with `persist` middleware for state management. This ensures state survives page reloads.

| Store | File | Purpose |
|-------|------|---------|
| `ui.store.ts` | 15KB | Sidebar visibility, theme, position, size, push mode, notifications |
| `connection.store.ts` | 6KB | Connection status, URL, retry logic, transport type |
| `tool.store.ts` | 9KB | Available tools, execution history, favorites, sorting |
| `app.store.ts` | 5KB | Global app state, initialization flags |
| `config.store.ts` | 11KB | User preferences, feature flags, remote config |
| `activity.store.ts` | 1KB | Activity log entries |
| `adapter.store.ts` | 11KB | Platform adapter registration, active adapter |
| `profile.store.ts` | 2KB | Multi-profile server configurations |
| `toast.store.ts` | 1KB | Toast notification queue |

**Additional stores in `lib/`**: `macro.store.ts` (user macros), `context.store.ts` (text snippets).

**Hook Pattern**: We use `useShallow` in `hooks/useStores.ts` to prevent unnecessary re-renders. Always prefer using these composed hooks in components.

**Critical Pattern**: Use `useUIStore.getState()` for non-React access (e.g., in Services).

## 3. Feature Deep Dives

### A. Agentic Mode (Macros)
*   **Location**: `lib/macro.runner.ts`, `components/sidebar/Macros/`
*   **Components**: `MacroBuilder.tsx` (create/edit), `MacroList.tsx` (list/manage)
*   **Logic**: The `MacroRunner` executes a list of steps:
    *   **`tool`**: Calls `mcpClient.callTool`.
    *   **`condition`**: Evaluates simple JS expressions (safe parsing, no `eval`). Supports branching (`goto`, `stop`).
    *   **`set_variable`**: Stores data in a local `env` object for use in subsequent steps (`{{env.varName}}`).
*   **Safety**: Max step limit (1000) prevents infinite loops.

### B. Context Manager
*   **Location**: `components/sidebar/ContextManager/ContextManager.tsx`, `lib/context.store.ts`
*   **Flow**:
    1.  User selects text → Right Click → "Save to MCP Context".
    2.  Background script catches event → Broadcasts `mcp:save-context`.
    3.  `Sidebar` (via EventBus) listens and notifies user.
    4.  `InputArea` listens and opens `ContextManager` with pre-filled text.
*   **Storage**: `context.store.ts` persists items to `localStorage`.

### C. Dynamic Theming
*   **Location**: `tailwind.config.ts`, `Sidebar.tsx`
*   **Mechanism**: Tailwind maps `primary-*` colors to CSS variables (e.g., `--color-primary-500`). `Sidebar.tsx` injects these variables into the root element's `style` attribute based on the user's selection in `Settings.tsx`.
*   **Palette**: Indigo (default), Blue, Green, Purple, Red, Orange.

### D. Auto-Execute Whitelist
*   **Location**: `services/automation.service.ts`
*   **Logic**:
    *   Checks `autoExecuteDelay` to determine if feature is globally active.
    *   Checks `autoExecuteWhitelist` array.
    *   **Rule**: If whitelist has items, *only* allowed tools run. If whitelist is empty, *no* tools auto-execute (Safe by default).

### E. Command Palette
*   **Location**: `components/sidebar/CommandPalette/CommandPalette.tsx`
*   **Trigger**: `/` key in sidebar
*   **Purpose**: Quick-access fuzzy search for tools and actions.

### F. FunctionBlock Parser
*   **Location**: `src/render_prescript/`
*   **Purpose**: Detects and parses MCP tool calls from AI model output rendered in the DOM.
*   **Supported formats**: JSON-Lines (`function_call_start`/`function_call_end`), XML (`<function_calls>`/`<invoke>`), and raw JSON objects.
*   **Streaming**: Monitors DOM mutations in real-time to detect tool calls as they stream in.

## 4. Communication Bridge

The `ContextBridge` pattern abstracts `chrome.runtime.sendMessage`.

*   **Content Script**: Sends `mcp:command` messages.
*   **Background**: Processes command → calls MCP Proxy → returns response.
*   **Events**: Background broadcasts events (e.g., `connection:status-changed`) which `McpClient` listens for and updates stores.

## 5. Cross-Browser Support

| Browser | Build Command | Manifest Differences |
|---------|---------------|---------------------|
| Chrome / Edge | `pnpm build` | `background.service_worker` |
| Firefox | `pnpm build:firefox` | `background.scripts`, `content_security_policy`, `browser_specific_settings.gecko` |

The `ManifestParser` in `packages/dev-utils/lib/manifest-parser/impl.ts` handles the conversion automatically. Both builds output to `dist/`.

## 6. Known Limitations & Future Work

1.  **Tailwind Shadow DOM**: We use a custom Tailwind config. Ensure `important: true` is maintained to override host page styles.
2.  **CSP Compatibility**: The `MacroRunner` avoids `new Function` where possible, but complex logic might still trigger strict CSPs on some sites.
3.  **Performance**: `ActivityLog` currently renders all items. For lists >1000 items, consider adding virtual scrolling.
4.  **Shared `dist/`**: Chrome and Firefox builds overwrite the same `dist/` directory. Only one can exist at a time.

## 7. Testing Protocol

*   **Unit**: Test Stores and Services (logic). Vitest configured in `chrome-extension/`.
*   **Integration**: Test `McpClient` → Background communication.
*   **E2E**: Playwright scripts (in `e2e/`) should verify the Sidebar interaction flows.
