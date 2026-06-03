# Version
1.0.0-alpha.62

## Version Sync Checklist

When bumping tormentnexus, update every active product-facing version reference below in the same change set. This list was refreshed during the `1.0.0-alpha.1` normalization pass to match the files that were actually updated.

### 1. Canonical version files
- `VERSION`
- `VERSION.md`

### 2. Top-level branding and package identity
- `README.md`
- `package.json`

### 3. App package manifests
- `apps/tormentnexus-extension/package.json`
- `apps/vscode/package.json`
- `apps/web/package.json`

### 4. Package manifests
- `packages/adk/package.json`
- `packages/agents/package.json`
- `packages/ai/package.json`
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

### 5. Web UI fallback and branding strings
- `apps/web/src/components/Navigation.tsx`
- `apps/web/src/components/mcp/nav-config.ts`

### 6. CLI and runtime fallback strings
- `packages/cli/src/version.ts`
- `packages/core/src/Router.ts`
- `packages/core/src/MCPServer.ts`
- `packages/core/src/stdioLoader.ts`
- `packages/core/src/routers/openWebUIRouter.ts`
- `packages/core/src/services/AgentMemoryService.ts`
- `packages/core/src/services/mcp-client.service.ts`

### 7. Tests that assert version metadata
- `packages/core/src/bridge/bridge-manifest.test.ts`

## Verification

After updating the files above:

1. Search active files for the previous version string to confirm none remain.
2. Run `pnpm run typecheck`.
3. If the version bump also changed release notes, verify `CHANGELOG.md` and `VERSION` agree with the new version.
