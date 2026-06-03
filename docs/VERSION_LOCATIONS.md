# TormentNexus Version Number Locations

This document tracks all locations where the TormentNexus version number is hardcoded or referenced. When performing a major version bump (like the `0.90.0` milestone), ensure all these locations are synchronized.
# tormentnexus Version Number Locations

This document tracks all locations where the tormentnexus version number is hardcoded or referenced. When performing a major version bump (like the `0.90.0` milestone), ensure all these locations are synchronized.

## 1. Primary Version Sources
- `VERSION`: (Current: `0.99.3`) - Authoritative plain-text version.
- `VERSION.md`: (Current: `0.99.3`) - Version history and sync checklist.
- `CHANGELOG.md`: (Current: `0.99.3`) - Narrative history of changes.
- `HANDOFF.md`: (Current: `0.99.3`) - Active session context.

## 2. Package Manifests (`package.json`)
All these currently reference `0.99.3`:
- `package.json` (Root)
- `apps/tormentnexus-extension/package.json`
- `apps/tormentnexus-extension/package.json`
- `apps/vscode/package.json`
- `apps/web/package.json`
- `packages/adk/package.json`
- `packages/agents/package.json`
- `packages/ai/package.json`
- `packages/tormentnexus-supervisor/package.json`
- `packages/tormentnexus-supervisor/package.json`
- `packages/browser/package.json`
- `packages/browser-extension/package.json`
- `packages/cli/package.json`
- `packages/core/package.json`
- `packages/mcp-client/package.json`
- `packages/mcp-registry/package.json`
- `packages/mcp-router-cli/package.json`
- `packages/memory/package.json`
- `packages/search/package.json`
- `packages/supervisor-plugin/package.json`
- `packages/tools/package.json`
- `packages/tsconfig/package.json`
- `packages/types/package.json`
- `packages/ui/package.json`
- `packages/vscode/package.json`

## 3. Web UI Fallbacks & Branding
- `apps/web/src/components/Navigation.tsx`: Fallback for `NEXT_PUBLIC_TORMENTNEXUS_VERSION`.
- `apps/web/src/components/mcp/nav-config.ts`: Hardcoded branding string "TormentNexus 0.99.3 Core".
- `apps/web/src/components/Navigation.tsx`: Fallback for `NEXT_PUBLIC_TORMENTNEXUS_VERSION`.
- `apps/web/src/components/mcp/nav-config.ts`: Hardcoded branding string "tormentnexus 0.99.3 Core".

## 4. CLI & Core Runtime Fallbacks
- `packages/cli/src/version.ts`: Returns hardcoded version string.
- `packages/core/src/Router.ts`: Initial status version.
- `packages/core/src/MCPServer.ts`: Server identity metadata.
- `packages/core/src/stdioLoader.ts`: Loader identity metadata.
- `packages/core/src/routers/openWebUIRouter.ts`: Router status metadata.
- `packages/core/src/services/AgentMemoryService.ts`: Service identity metadata.
- `packages/core/src/services/mcp-client.service.ts`: Client identity metadata.

## 5. Other Components (Fixed Versions)
These versions are typically independent of the main TormentNexus version:
These versions are typically independent of the main tormentnexus version:
- `packages/tormentnexus/package.json`: (Currently `10.5.7`)
- `archive/OmniRoute/package.json`: (Currently `2.3.1`)
- Various submodules in `submodules/` or `archive/submodules/`.

## 6. Maintenance Scripts
Scripts like `scripts/generate_dashboard.py` and `scripts/update_submodules_doc.ts` dynamically extract versions from git tags or commit hashes, but may have fallback logic.
