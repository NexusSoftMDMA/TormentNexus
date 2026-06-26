# CLI Task Surface — `chorus run` and `chorus review`

## Problem

Today Chorus has no CLI subcommand for starting a chat. Users can only drive the daemon
via the cockpit (web UI on `:5050`) or via MCP from another tool. This blocks three
real workflows:

1. **Headless / automation contexts** — cron jobs, CI pipelines, shell scripts that
   want to run a Chorus review and gate on the verdict
2. **Pipe-friendly composition with `/work` and other harnesses** — `echo "$DIFF" |
   chorus review --json` is the unlock; today these harnesses have no clean way to
   delegate review to Chorus
3. **Terminal-native users** — people who live in shells and don't want to leave for a
   browser to start a chat

## Approach

Two subcommands that mirror the doer-vs-no-doer template split (see
`review-only-mode.md`):

```
chorus run "fix the bug in foo.ts"        # full-pipeline templates (doer required)
chorus review --file diff.patch           # review-only templates (artifact, no doer)
chorus review < diff.patch                # same, stdin-friendly for pipes
```

Same daemon, same templates table, same chat lifecycle as cockpit/MCP entry points.
Only the entry-point shape differs.

## Subcommand specs

### `chorus run`

```
chorus run <task...>                              # task as positional args
chorus run --template <id> <task...>              # explicit template pick
chorus run --repo <path> <task...>                # enables ship phase
chorus run --no-ship <task...>                    # disable ship even if template enables it
chorus run --attach <chat-id>                     # re-attach to a chat already running
chorus run <task...> --quiet                      # only print final verdict
chorus run <task...> --json                       # JSON Lines events to stdout
```

**Template resolution.** `--template` explicit; otherwise pick the user's configured
default (settings key `cli.run.defaultTemplate`); otherwise fall back to a designated
built-in (`code-review`).

**Reject.** If resolved template's first phase is `review_only`, fail with:
> Template 'review-only' is review-only — use `chorus review` instead.

**Repo / ship.** `--repo` enables ship phase if the template has it; `--no-ship`
forces skip. No default repo (user opts in).

### `chorus review`

```
chorus review --file <path>                       # artifact from file
chorus review                                     # artifact from stdin
chorus review --template <id> --file <path>       # explicit template pick
chorus review --file <path> --findings-only       # strip agent output, just findings block
chorus review --file <path> --out <dir>           # per-reviewer findings written to dir
chorus review --file <path> --json                # JSON Lines events to stdout
chorus review --file <path> --quiet               # only print final verdict
```

**Template resolution.** `--template` explicit; otherwise built-in `review-only`.

**Reject.** If resolved template's first phase is NOT `review_only`, fail with:
> Template 'code-review' requires a doer — use `chorus run` instead.

**Artifact source.** `--file` OR stdin. Reject if both provided. Reject if both empty.
Reject if size exceeds template's `artifact.maxBytes` with a clear error showing the
limit.

## Output modes

### Default (interactive, TTY detected)

Stream prefixed lines per participant:

```
[doer/claude] Reading the file...
[doer/claude] The bug is at line 42, replacing X with Y...
[doer/claude] DONE
[reviewer-1/codex] Looking at the diff...
[reviewer-2/gemini] The fix looks correct, but...
[reviewer-1/codex] DONE
[reviewer-2/gemini] DONE

✓ Agreed (codex + gemini)
```

ANSI color when TTY, plain text when piped (auto-detect).

### `--quiet`

Suppress streaming. Print only the final verdict block:

```
✓ Agreed (codex + gemini)

Findings:
- ...
```

### `--json`

JSON Lines to stdout, one event per line. Stable schema for harness consumers:

```json
{"event":"chat_created","chatId":"abc123","template":"review-only"}
{"event":"phase_started","phase":"review","participants":["codex","gemini"]}
{"event":"phase_event","participant":"codex","kind":"output","text":"..."}
{"event":"phase_event","participant":"codex","kind":"done"}
{"event":"phase_event","participant":"gemini","kind":"done"}
{"event":"converged","verdict":"agree","reviewers":["codex","gemini"]}
{"event":"chat_done","chatId":"abc123","verdict":"agree"}
```

`--quiet` and `--json` are mutually exclusive (json overrides).

## Exit codes

```
0 = converged with `agree` verdict
1 = converged with `disagree` verdict
2 = chat errored (precheck fail, CLI crashed, timeout, daemon unreachable)
3 = bad invocation (wrong subcommand for template kind, missing artifact, etc.)
```

This makes shell composition clean:

```bash
chorus review --file diff.patch && git push origin feat/foo
echo "$DIFF" | chorus review --quiet || { echo "review failed"; exit 1; }
```

## Daemon dependency

`chorus run` and `chorus review` **require the daemon to already be running.** If the
daemon is not running, fail with:

> Daemon not running. Start it with `chorus start`.

Auto-starting the daemon from the task subcommand is rejected — creates lifecycle
confusion (who owns the daemon? does Ctrl-C kill it? does it persist after the task?).
Explicit `chorus start` keeps ownership clear.

## Cancellation / Ctrl-C

When the user hits Ctrl-C in `chorus run` or `chorus review`:

1. Local CLI sends `DELETE /chats/<chatId>` to the daemon (which already exists for
   cockpit cancellation)
2. Daemon kills participant subprocesses and marks the chat cancelled
3. CLI prints "Cancelled." and exits with code 130 (standard SIGINT exit)

The CLI does NOT exit before the daemon confirms cancellation — otherwise zombie
subprocesses leak. Add a 2s timeout on the cancel ack; if exceeded, exit with a
warning ("daemon did not ack cancel; chat may still be running").

## `--attach` re-connection

`chorus run --attach <chat-id>` re-subscribes to a chat already running in the daemon.
Use case: terminal disconnected, user wants to reconnect to a long-running task without
losing the live stream. SSE subscription is idempotent — multiple subscribers per chat
already work. CLI just opens the SSE connection and prints from the current point.

If the chat already finished, print the final verdict and exit with the matching code.

## /work integration (the unlock)

The `/work` orchestrator currently spawns review fleets via tmux + per-agent shell
processes. With `chorus review --json`, /work can delegate the entire fan-out to
Chorus:

```bash
chorus review --file packed-diff.txt --json --template tri-review > findings.jsonl

# Parse last line for verdict
verdict=$(tail -1 findings.jsonl | jq -r '.verdict')
```

This is a *significant* simplification of /work's review path. Worth a follow-up
investigation: does /work still need its own tmux-based fleet, or can it call Chorus
for everything review-shaped? Probably yes for the `plan` and `major-bug` modes.

## Template listing

A small companion subcommand to discover templates:

```
chorus templates ls                     # list all templates with kind + role summary
chorus templates show <id>              # full template YAML
```

Helps users learn which templates are review-only vs full-pipeline without opening the
cockpit.

## Out of scope

- **Auto-start daemon from task subcommand** — explicit `chorus start` stays the model
- **TUI mode with rich panes** — `--json` + external tools (e.g. pipe to a TUI like
  `gum` or `fzf`) is the composable answer
- **Resumable chats with disk-backed transcript replay** — `--attach` covers the
  current-session case; full historical replay belongs to a future `chorus chats`
  subcommand
- **Multi-template fan-out from one invocation** (`chorus run --templates a,b,c`) —
  shell loops cover this until evidence shows it's painful
- **Chat history browsing** (`chorus chats ls`, `chorus chats show <id>`) — separate
  follow-up, not blocking task surface

## Risks

- **Terminal width and wrapping** — multi-participant streaming can produce visual
  spaghetti on narrow terminals. Mitigation: each line carries a `[participant]` prefix
  so users can grep / `awk '/^\[reviewer-1\]/'` to filter.
- **ANSI color in pipes** — auto-detect TTY (`process.stdout.isTTY`); strip color when
  piping. Standard hygiene.
- **JSON schema drift** — `--json` events become a public contract once /work depends
  on them. Version the event shape from day one (`{"event":"...","schema":1,...}`) so
  future changes can be additive without breaking consumers.
- **Long stdin reads on `chorus review`** — if the user pipes a multi-MB diff,
  reading stdin should be streaming, not buffered. Use a streaming read with the
  `maxBytes` cap as the early-exit limit.
- **Daemon version mismatch** — CLI binary and daemon binary can drift if user
  upgrades one without the other. Add a daemon `/version` check on first connect; warn
  if they don't match.

## Order of work

1. **`chorus review`** first — smaller surface (no doer logic, no ship phase, no repo
   handling), unlocks /work integration immediately
2. **`chorus run`** second — adds the doer + ship + repo handling
3. **`chorus templates ls/show`** third — discovery layer
4. **`--json` event schema versioning + docs** before the first PR ships, so consumers
   can rely on the contract from day one

Each is roughly half a day. The full set is ~2 days. Depends on `review-only-mode.md`
shipping first (the substrate that `chorus review` drives).

## Estimated effort

| Subcommand | Effort | Depends on |
|---|---|---|
| `chorus review` | ½ day | review-only-mode.md substrate |
| `chorus run` | ½ day | nothing (existing templates work today) |
| `chorus templates ls/show` | ¼ day | nothing |
| JSON event schema + docs | ¼ day | nothing |
| /work migration to use `chorus review` | ½ day (in /work repo) | `chorus review` shipped |

Total: ~2 days of focused work, splittable into four independent PRs.
