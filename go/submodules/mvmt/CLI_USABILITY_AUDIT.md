# CLI Usability Audit

## Main User Journeys

### First-Time Run

Expected path:
- `mvmt serve`
- missing config triggers guided setup
- user adds one or more mounts
- server starts locally, optionally with `-i`

Current friction:
- Top-level help explains commands but gives no examples.
- The description says "personal local data" instead of the current README language around selected local folders and mounts.
- `config setup` is discoverable, but examples do not show the faster `serve --path <dir>` path.

### Common Successful Workflow

Expected path:
- `mvmt mounts add notes ~/notes --mount-path /notes --read-only`
- `mvmt reindex`
- `mvmt serve -i`
- `mvmt token`

Current friction:
- Command help lists flags but not common command shapes.
- `mounts` output is human-readable only, which makes scripts parse terminal formatting.
- `reindex` failure for empty mounts does not tell the user what to do next.

### Invalid Input / Typo

Expected path:
- mistyped command or invalid flags should show the nearest valid command and a next step.

Current friction:
- Unknown commands show help, but no suggestion.
- Errors rely on the user reading the whole help block.

### Missing Config / Missing Auth / Missing File

Expected path:
- missing config points to `mvmt config setup` or a direct one-off command.
- missing token points to `mvmt serve` or `mvmt token rotate`.
- missing mount roots are reported by `doctor`.

Current friction:
- Most messages have next steps.
- `reindex` has weaker next-step guidance than other commands.

### Cancellation / Ctrl-C

Expected path:
- `serve -i` should treat first Ctrl-C as cancel/escape and second Ctrl-C as quit.
- Prompt cancellation should not print a stack trace.

Current friction:
- Interactive mode already has double Ctrl-C handling and cancel handling.
- Non-interactive destructive commands do not have an automation-friendly confirmation bypass.

### Non-Interactive CI / Script Usage

Expected path:
- read-only commands should support machine output.
- destructive commands should support explicit confirmation flags.

Current friction:
- `doctor --json` exists.
- `mounts list` lacks `--json`.
- `mounts remove` always prompts, so scripts cannot remove a mount safely without an interactive prompt.
- Nested commands that accept `--config` can be confusing when the option is parsed on the parent command instead of the child command.

### Power-User Repeated Usage

Expected path:
- short workflows should be easy to discover from help.
- repeated `mounts add/edit/remove`, `reindex`, and `serve -i` should have predictable output.

Current friction:
- Help text lacks examples.
- Command descriptions are sometimes accurate but not concrete.

## Proposed Improvements

1. High impact / low risk: Add examples to top-level and high-traffic command help.
2. High impact / low risk: Add `--json` to `mvmt mounts` and `mvmt mounts list`.
3. High impact / low risk: Add `--yes` to `mvmt mounts remove` to make scripted removal explicit and non-interactive.
4. Medium impact / low risk: Add typo suggestions for unknown commands.
5. Medium impact / low risk: Improve `mvmt reindex` empty-config guidance.
6. Medium impact / low risk: Make nested commands inherit `--config` consistently.

## Done When

- [x] Top-level help uses README-consistent wording and includes examples.
- [x] `serve`, `mounts add`, and `mounts remove` help include examples.
- [x] Unknown command output suggests close matches when available.
- [x] `mvmt mounts --json` and `mvmt mounts list --json` produce stable JSON.
- [x] `mvmt mounts remove <name> --yes` removes without prompting.
- [x] Nested mount commands honor `--config` whether it is parsed on the parent or child command.
- [x] `mvmt reindex` explains how to add a mount when none exist.
- [x] README command table reflects new additive flags.
- [x] Tests cover JSON mount output, `--yes` removal, help examples, and reindex guidance.
- [x] `npm run verify` passes.
