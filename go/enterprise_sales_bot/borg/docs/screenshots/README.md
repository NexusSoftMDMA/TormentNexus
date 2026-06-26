# Screenshot Capture Guide

Use this folder for repository-facing screenshots referenced from the root `README.md`.

## Canonical filenames

- `dashboard-home.png`
- `mcp-registry.png`
- `mcp-search.png`
- `mcp-inspector.png`
- `billing.png`
- `github-actions.png`

## Capture standards

- Resolution: 1920×1080 (or 1440p with similar aspect ratio)
- Theme: dark mode where possible
- Browser zoom: 100%
- Hide personal data, secrets, and API keys
- Prefer stable states (no transient toasts/loaders unless intentionally highlighted)

## Workflow

1. Capture screenshots using consistent viewport settings.
2. Name files exactly as listed above.
3. Replace existing files in this directory.
4. Update the status checkboxes in the root `README.md` screenshot table.

## Pre-commit sanity check

- Ensure each screenshot opens correctly.
- Ensure file names match the README table paths exactly.
- Keep file sizes reasonable for GitHub rendering (generally under 1.5 MB per image).

## Validation command

Run the repository screenshot validator before committing:

`pnpm run check:screenshots`

Use strict mode to fail when any required screenshot is missing:

`pnpm run check:screenshots:strict`

## Status sync command

Automatically update the screenshot status column in root `README.md`:

`pnpm run sync:screenshot-status`

Check whether status is already synced (no file writes):

`pnpm run check:screenshot-status-sync`

## One-command refresh

Run sync + validation together:

`pnpm run visuals:refresh`

For release-level enforcement (fails if any required screenshot is missing):

`pnpm run visuals:refresh:strict`

Verify-only mode (no writes):

`pnpm run visuals:verify`

Verify-only strict mode (no writes, fails if screenshots are missing):

`pnpm run visuals:verify:strict`

## Command matrix (all the stuff)

| Goal | Command | Writes files | Fails on missing screenshots |
|---|---|---:|---:|
| Sync status only | `pnpm run sync:screenshot-status` | ✅ | ❌ |
| Check status sync only | `pnpm run check:screenshot-status-sync` | ❌ | ❌ |
| Validate screenshots (warn mode) | `pnpm run check:screenshots` | ❌ | ❌ |
| Validate screenshots (strict) | `pnpm run check:screenshots:strict` | ❌ | ✅ |
| Refresh visuals (daily) | `pnpm run visuals:refresh` | ✅ | ❌ |
| Refresh visuals (release) | `pnpm run visuals:refresh:strict` | ✅ | ✅ |
| Verify visuals (CI no-write) | `pnpm run visuals:verify` | ❌ | ❌ |
| Verify visuals (CI no-write strict) | `pnpm run visuals:verify:strict` | ❌ | ✅ |
| Do all visuals checks (daily) | `pnpm run visuals:all` | ✅ | ❌ |
| Do all visuals checks (release) | `pnpm run visuals:all:strict` | ✅ | ✅ |

## Release gate integration

- Default gate (warn-level visuals): `pnpm run check:release-gate:ci`
- Strict visuals gate: `pnpm run check:release-gate:ci:strict-visuals`
