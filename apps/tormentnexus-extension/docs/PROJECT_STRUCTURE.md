# Project Structure & Dependencies

This document provides a detailed overview of the `tormentnexus-extension` monorepo structure, its submodules (packages), and their relationships.
This document provides a detailed overview of the `tormentnexus-extension` monorepo structure, its submodules (packages), and their relationships.

## Directory Layout

The project is structured as a monorepo managed by `turbo` and `pnpm`.

```
tormentnexus-extension/
tormentnexus-extension/
├── chrome-extension/       # The core Chrome Extension logic (manifest, background, build config)
│   ├── src/
│   │   ├── background/     # Service worker logic (connection management, context menus, analytics)
│   │   └── mcpclient/      # MCP protocol implementation (SSE, WebSocket, Streamable HTTP plugins)
│   ├── utils/              # Build plugins (manifest, assets)
│   ├── public/             # Static assets (icons)
│   └── manifest.ts         # Extension manifest source
│
├── pages/
│   └── content/            # The main UI injected into web pages (Sidebar, React App)
│       └── src/
│           ├── components/
│           │   ├── sidebar/ # 14 tab components (Dashboard, Tools, Macros, Context, Settings...)
│           │   └── ui/      # Shared shadcn components (Button, Card, Dialog)
│           ├── hooks/       # Custom React hooks (useMcpCommunication, useKeyboardShortcuts)
│           ├── lib/         # Core logic (MacroRunner, context/macro stores)
│           ├── plugins/
│           │   └── adapters/ # 16 per-platform adapters (ChatGPT, Gemini, Grok, etc.)
│           ├── services/    # Business logic (AutomationService)
│           ├── stores/      # 10 Zustand state stores
│           ├── types/       # TypeScript interfaces
│           └── render_prescript/ # DOM rendering, function call detection
│
├── packages/               # Shared internal packages (12 workspaces)
│   ├── dev-utils/          # Manifest parser, build helpers
│   ├── env/                # Environment configuration (IS_FIREFOX, IS_DEV)
│   ├── hmr/                # Hot Module Replacement logic
│   ├── i18n/               # Internationalization infrastructure
│   ├── module-manager/     # Module lifecycle management
│   ├── shared/             # Logger, utilities
│   ├── storage/            # Chrome Storage wrappers
│   ├── tailwind-config/    # Shared Tailwind configuration
│   ├── tsconfig/           # Shared TypeScript configuration
│   ├── ui/                 # Shared UI components (shadcn/ui)
│   ├── vite-config/        # Shared Vite configuration
│   └── zipper/             # ZIP packaging for distribution
│
├── docs/                   # Project documentation (11 files)
├── scripts/                # Build utilities
└── bash-scripts/           # Shell maintenance scripts
```

## Submodules / Workspaces

The project uses pnpm workspaces. These are not git submodules but internal packages.

| Package | Path | Version | Description |
| :--- | :--- | :--- | :--- |
| `chrome-extension` | `chrome-extension/` | 0.7.1 | The build entry point for the extension. |
| `content` | `pages/content/` | 0.7.1 | The frontend UI (Sidebar) injected into pages. |
| `@extension/shared` | `packages/shared/` | workspace:* | Shared utilities and logger. |
| `@extension/storage` | `packages/storage/` | workspace:* | Type-safe wrappers for `chrome.storage`. |
| `@extension/env` | `packages/env/` | workspace:* | Environment variable handling (IS_FIREFOX, IS_DEV). |
| `@extension/ui` | `packages/ui/` | workspace:* | Shared UI components (shadcn/ui). |
| `@extension/hmr` | `packages/hmr/` | workspace:* | Hot Module Replacement. |
| `@extension/i18n` | `packages/i18n/` | workspace:* | Internationalization. |
| `@extension/dev-utils` | `packages/dev-utils/` | workspace:* | Manifest parser, build plugins. |
| `@extension/vite-config` | `packages/vite-config/` | workspace:* | Shared Vite configuration. |
| `@extension/tsconfig` | `packages/tsconfig/` | workspace:* | Shared TypeScript configuration. |

## Build System

-   **Build Tool**: Vite (v6.1.0)
-   **Monorepo Manager**: Turbo (v2.4.2)
-   **Package Manager**: pnpm (v9.15.1)
-   **Node.js**: ≥22.12.0

## Versioning

The single source of truth for the project version is the `VERSION` file in the root directory.
Currently: `0.7.1`

When updating the version:
1.  Update `VERSION`.
2.  Run `pnpm update-version` (or manually update `package.json` and `chrome-extension/package.json`).
3.  Update `CHANGELOG.md`.
