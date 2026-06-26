# Agent Rules

This file is the entry point for any AI coding agent working on mvmt.
If you are Claude Code, Codex, Cursor, Copilot, or any other coding agent,
read and follow these rules before making any changes.

## Project context

mvmt is a local-first MCP data hub. Read `README.md` for what it does,
`docs/architecture.md` for how it works, and `CONTRIBUTING.md` for
development setup and the test matrix.

## Rules

### 1. Keep diffs small and reviewable

- One concern per commit. Do not bundle unrelated changes.
- Prefer editing existing files over creating new ones.
- Do not refactor, rename, reformat, or add comments to code you are not
  changing for the task at hand.
- Do not add speculative features, abstractions, or configurability beyond
  what was requested.
- If a change touches more than ~5 files or ~150 lines, stop and break it
  into smaller steps.

### 2. Run verification before declaring done

Run:

```bash
npm run verify
```

This is the PR gate. It runs the TypeScript build, full test suite,
whitespace checks, package dry-run, and a runtime smoke test.

Coverage requirements:

- **100% feature coverage**: every behavior your change introduces or
  modifies must have a test that exercises it, including error/rejection
  paths.
- **80% spatial coverage**: at minimum 80% of the lines you added or
  changed must be hit by at least one test. Run `npm run test:coverage`
  to check.

Do not skip verification. Do not report a task as complete without
showing that verification passed.

### 3. Security-sensitive changes need extra tests

If your change touches any of these areas, add focused tests for both
the success and failure/rejection paths:

- Token generation or validation
- OAuth/PKCE flow
- Origin checks
- Write gates
- Filesystem path validation (traversal, symlinks)
- Environment scrubbing
- Audit logging
- Connector process spawning or shutdown

Default to less access. Require explicit user consent for writes.

### 4. PR description format

Every pull request must include:

```
## Why
<What problem does this solve or what goal does it advance?>

## What changed
<Bullet list of behavior changes, not file-level diffs>

## How
<Brief explanation of the approach and any non-obvious decisions>

## Changed files
<List every file touched, one per line, with a short note>

## Verification
<Paste the output of `npm run verify` showing it passed>
<If coverage-relevant, paste the output of `npm run test:coverage`>
```

### 5. Commit messages

Use short, direct commit messages. Prepend with agent name:

```
[claude] fix daily note symlink escape
[codex] add shutdown timeout guard
```

### 6. What not to do

- Do not push to main without PR review.
- Do not modify `.env`, credentials, or token files.
- Do not add dependencies without discussing it first.
- Do not change the security model (auth, write gates, origin checks)
  without explicit approval.
- Do not generate or guess URLs, API keys, or secrets.
- Do not amend existing commits; create new ones.
