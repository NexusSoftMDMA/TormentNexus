# Code Health Audit

Date: 2026-04-29
Branch: `quality/code-health-pass`

Scope: local-first mount runtime, CLI command layer, text index, router, tests, and docs visible in this repository.

## Findings

### 1. Router initialization is not idempotent

- Problem: `ToolRouter.initialize()` appends context tool definitions every time it is called.
- Evidence from the code: `src/apps/mcp/router.ts` pushes into `this.contextToolDefinitions` without checking whether initialization already happened.
- User/developer impact: repeated initialization in a long-running process or test helper can expose duplicate tools and make behavior depend on call order.
- Risk level: Medium.
- Proposed fix: make initialization idempotent by tracking whether the router has already initialized.
- Acceptance criteria: calling `initialize()` twice returns the same tool list as calling it once.
- Files likely involved: `src/apps/mcp/router.ts`, `tests/router.test.ts`.
- Status: Completed.

### 2. `expected_hash` writes can create a missing file

- Problem: `TextContextIndex.write(path, content, expectedHash)` allows the write when the target file is missing.
- Evidence from the code: `src/context/text-index.ts` catches the failed read and only rethrows if `provider.exists()` returns true.
- User/developer impact: a client that passes a hash from a previous read can silently recreate a deleted file instead of detecting stale state.
- Risk level: Medium.
- Proposed fix: if `expected_hash` is present, require the current file to exist and match that hash.
- Acceptance criteria: missing target with `expected_hash` fails; matching hash still permits a write; unmatched hash still fails.
- Files likely involved: `src/context/text-index.ts`, `tests/text-index.test.ts`.
- Status: Completed.

### 3. Nested `--config` inheritance only checks one parent command

- Problem: nested commands inherit `--config` from their direct parent only.
- Evidence from the code: `bin/mvmt.ts` uses `command.parent?.opts()` in `withInheritedConfig`.
- User/developer impact: deeper commands such as `mvmt tunnel --config cfg logs stream` can ignore the config path if the option is parsed on a grandparent command.
- Risk level: Low.
- Proposed fix: walk command ancestors until a config value is found.
- Acceptance criteria: nested command helpers inherit config from any ancestor command.
- Files likely involved: `bin/mvmt.ts`, `src/cli/command-options.ts`, `tests/cli-command-options.test.ts`.
- Status: Completed.

### 4. Tool argument validation silently ignores invalid arrays

- Problem: optional string array inputs accept any non-array as undefined and drop non-string array values.
- Evidence from the code: `optionalStringArray()` in the MCP app tool helpers filters invalid entries instead of rejecting them.
- User/developer impact: malformed client calls can behave like broader default calls, making debugging harder.
- Risk level: Medium.
- Proposed fix: return a structured tool error when `mounts` is present but is not an array of non-empty strings.
- Acceptance criteria: `search({ mounts: "notes" })` and mixed arrays return `isError: true`.
- Files likely involved: `src/apps/mcp/tools/helpers.ts`, `tests/router.test.ts`.
- Status: Backlog.

### 5. Index snapshot parsing is unvalidated

- Problem: `readSnapshot()` casts parsed JSON to `TextIndexSnapshot` without checking version or shape.
- Evidence from the code: `JSON.parse(...) as TextIndexSnapshot` in `src/context/text-index.ts`.
- User/developer impact: corrupt or future-version index files can produce confusing runtime failures later in search.
- Risk level: Medium.
- Proposed fix: validate snapshot version and array fields before using it.
- Acceptance criteria: malformed snapshots fail with a clear error; missing snapshots still behave as empty.
- Files likely involved: `src/context/text-index.ts`, `tests/text-index.test.ts`.
- Status: Backlog.

### 6. Search scoring has no deterministic tie-break on chunk position

- Problem: equal scores sort by path only; chunks from the same file keep implementation-dependent ordering.
- Evidence from the code: `search()` sorts by score then path in `src/context/text-index.ts`.
- User/developer impact: repeated searches can return confusing ordering when multiple chunks in one file match equally.
- Risk level: Low.
- Proposed fix: include `chunk_id` or parsed offset in the tie-break.
- Acceptance criteria: equal-score chunks from one file return in file order.
- Files likely involved: `src/context/text-index.ts`, `tests/text-index.test.ts`.
- Status: Backlog.

### 7. Rebuild cost is paid after every write and remove

- Problem: each write/remove performs a full index rebuild.
- Evidence from the code: `TextContextIndex.write()` and `remove()` both call `await this.rebuild()`.
- User/developer impact: large mount sets will make simple writes slow.
- Risk level: Medium.
- Proposed fix: add targeted index updates or queue rebuild work after write/remove.
- Acceptance criteria: writing one file updates index state without walking unrelated mounts.
- Files likely involved: `src/context/text-index.ts`, `tests/text-index.test.ts`.
- Status: Backlog.

### 8. Router mixes permission checks, dispatch, and auditing

- Problem: `ToolRouter` has multiple reasons to change.
- Evidence from the code: `src/apps/mcp/router.ts` parses arguments, checks permissions, dispatches storage calls, and records audit entries.
- User/developer impact: small changes to one concern are harder to review and test in isolation.
- Risk level: Medium.
- Proposed fix: split context tool definitions and handlers into per-tool modules with a small registry.
- Acceptance criteria: router tests remain behavior-focused; adding a new tool requires adding a module and registering it in `src/apps/mcp/tools/index.ts`.
- Files likely involved: `src/apps/mcp/router.ts`, `src/apps/mcp/tools/*`, `tests/router.test.ts`.
- Status: Completed for context tool definitions and handlers. Permission helpers and audit remain in `ToolRouter`.

### 9. CLI command registration is concentrated in one large file

- Problem: `bin/mvmt.ts` owns command definitions, examples, compatibility aliases, and shared option behavior.
- Evidence from the code: the executable file contains all top-level and nested command registration.
- User/developer impact: adding commands increases merge conflicts and makes command-specific tests less local.
- Risk level: Low.
- Proposed fix: extract command registration helpers by command group once more CLI flags are added.
- Acceptance criteria: each command group has a small registration function and existing CLI subprocess tests still pass.
- Files likely involved: `bin/mvmt.ts`, `src/cli/*`, `tests/cli-usability.test.ts`.
- Status: Backlog.

### 10. Interactive prompt flows have limited subprocess coverage

- Problem: interactive control paths are mostly unit-tested rather than exercised through a spawned CLI.
- Evidence from the code: `tests/interactive.test.ts` checks helpers and command handlers, but not a full prompt session.
- User/developer impact: regressions in prompt wiring, cancellation, and output can slip through.
- Risk level: Medium.
- Proposed fix: add one pty-style smoke test if the test environment can run it reliably.
- Acceptance criteria: test covers startup, `help`, and clean exit without depending on color output.
- Files likely involved: `tests/interactive.test.ts`, `bin/mvmt.ts`.
- Status: Backlog.

### 11. Some tests assert broad object shapes instead of exact behavior

- Problem: several tests use `expect.objectContaining` where exact output contracts matter.
- Evidence from the code: router and server tests parse JSON responses but often assert only a subset.
- User/developer impact: accidental extra fields or missing important fields may not be caught.
- Risk level: Low.
- Proposed fix: tighten assertions on stable public tool outputs.
- Acceptance criteria: key tool responses have exact or near-exact shape tests.
- Files likely involved: `tests/router.test.ts`, `tests/server.test.ts`.
- Status: Backlog.

### 12. Doctor and config summary output are human-only in some paths

- Problem: `doctor --json` exists, but other config diagnostics are formatted only for terminals.
- Evidence from the code: `src/cli/config.ts` prints summaries directly.
- User/developer impact: scripts must parse display text for config checks.
- Risk level: Low.
- Proposed fix: add JSON output only where there is a real scripting use case.
- Acceptance criteria: no behavior change for default output; JSON mode is covered by subprocess tests.
- Files likely involved: `src/cli/config.ts`, `bin/mvmt.ts`, `tests/config-cli.test.ts`.
- Status: Backlog.

### 13. Accessibility is mostly terminal readability, not browser UI

- Problem: the project has no browser UI, so accessibility concerns are terminal-focused.
- Evidence from the code: CLI/TUI output uses chalk colors but generally includes text labels.
- User/developer impact: color-only status would be hard to read; current output mostly avoids this.
- Risk level: Low.
- Proposed fix: keep labels such as `enabled`, `disabled`, `read-only`, and avoid future color-only states.
- Acceptance criteria: new CLI output remains understandable with color disabled.
- Files likely involved: `src/cli/*`, `tests/*cli*.test.ts`.
- Status: Backlog.

## Verification Log

- 2026-04-29: `npm test -- --run tests/router.test.ts tests/text-index.test.ts tests/cli-command-options.test.ts tests/cli-usability.test.ts` passed.
- 2026-04-29: `npm run verify` passed.
- 2026-04-29: `npm run verify` passed after final audit update.
- 2026-04-29: removed leaked non-project audit finding.
- 2026-04-29: `npm test -- --run tests/router.test.ts` passed after context tool registry extraction.
- 2026-04-29: `npm test -- --run tests/context-tools.test.ts tests/router.test.ts tests/server.test.ts` passed after per-tool module extraction.
- 2026-04-29: `npm run verify` passed after per-tool module extraction.
- New findings discovered during implementation: none.
