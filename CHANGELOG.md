# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-alpha.56] - 2026-05-10

### Added
- **Search Alignment**: Implemented `SearchRankingService.ts` in the TypeScript orchestrator to proxy tool searches to the Go sidecar's BM25/Cosine ranking engine, ensuring unified search results across all interfaces.
- **Hardened Execution**: Refactored `ShellService.ts` to use `spawn` with explicit argument arrays for all execution paths, eliminating `child_process.exec` and mitigating shell-injection risks.

### Changed
- **Performance**: Validated and integrated `O(1)` batch hydration for `ToolSetsRepository` and `ToolChainsRepository`, reducing database roundtrips.

## [1.0.0-alpha.55] - 2026-05-10

### Added
- **Stream Stabilization**: Introduced `SignalBuffer` to debounce high-frequency events like `USER_ACTIVITY`.
- **Resilient Subscriptions**: Added exponential backoff reconnection policy to `TRPCProvider`.

## [1.0.0-alpha.53] - 2026-05-08

### Changed
- **Infrastructure**: Standardized Go sidecar port to 4300 and standardized import paths to `github.com/borghq/borg-go`.
- **UX**: Restored Alt+Enter keyboard shortcut for instant plan approval.

