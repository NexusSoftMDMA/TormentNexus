# Improvement Plan

This pass prioritizes readability, modularity, DRY, KISS, and small verified changes.

## Selected Improvements

### 1. Centralize mount path policy helpers

Status: complete

Why:
- `LocalFolderStorageProvider` currently owns storage behavior and low-level path-pattern policy helpers.
- Pulling pattern matching and global secret-path decisions into a focused module gives storage one fewer reason to change.

Done when:
- [x] Path-pattern matching and global secret-path checks live in a dedicated module.
- [x] `LocalFolderStorageProvider` delegates to that module.
- [x] Existing storage behavior remains unchanged.
- [x] Focused tests cover global sensitive-path rejection and pattern matching.

### 2. Remove duplicate config writing in tunnel CLI

Status: complete

Why:
- `src/cli/tunnel.ts` has a private `saveConfig` implementation even though `src/config/loader.ts` exports the canonical writer.
- Keeping one config writer avoids permission-mode drift and reduces copy-paste maintenance.

Done when:
- [x] `src/cli/tunnel.ts` imports `saveConfig` from `src/config/loader.ts`.
- [x] The private duplicate writer and now-unused imports are removed.
- [x] Existing tunnel/config tests pass.

### 3. Isolate OAuth client registration validation

Status: complete

Why:
- `OAuthStore.registerClient` currently normalizes, enforces registry limits, checks duplicates, mutates state, and handles persistence rollback in one method.
- Extracting normalization and limit checks improves readability without changing OAuth behavior.

Done when:
- [x] Client registration normalization is handled by a focused helper.
- [x] Registry limit checks are handled by a focused helper.
- [x] Existing OAuth/server registration tests pass.

## Verification Checklist

- [x] Focused tests pass after each completed item.
- [x] `npm run verify` passes after all selected improvements.
