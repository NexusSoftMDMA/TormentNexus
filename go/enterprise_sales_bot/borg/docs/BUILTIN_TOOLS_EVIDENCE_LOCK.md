# Built-In Tools Evidence Lock

> Date: 2026-03-19  
> Purpose: source-pin every competitor tool contract before declaring parity complete.

## How to Use

For each platform, fill all required fields:

- **Primary source URL** (official docs/repo)
- **Version/commit/date pin**
- **Exact tool names**
- **Parameter schemas**
- **Return payload shape**
- **Approval/permission model**
- **Known caveats**

A platform is only marked **Locked** when all fields are complete and reviewed.

---

## Related Resources (Phase 1: Foundation)

- **[TORMENTNEXUS_MASTER_INDEX.jsonc](../TORMENTNEXUS_MASTER_INDEX.jsonc)** — Master project index with full evidence lock state tracking, phase status, and artifact inventory.
- **[TORMENTNEXUS_MASTER_INDEX.jsonc](../TORMENTNEXUS_MASTER_INDEX.jsonc)** — Master project index with full evidence lock state tracking, phase status, and artifact inventory.
- **[EVIDENCE_LOCK_STATUS_2026_03_19.md](./EVIDENCE_LOCK_STATUS_2026_03_19.md)** — Comprehensive status snapshot with metrics and next steps.
- **[VERSION_PINS.jsonc](./fixtures/VERSION_PINS.jsonc)** — Central registry for version pins across all L2 platforms. Used in CI validation.
- **[TOOL_CONTRACTS.md](./fixtures/TOOL_CONTRACTS.md)** — Golden fixture template capturing tool signatures, permission patterns, and hook semantics.
- **[L3_PROMOTION_STRATEGY.md](./fixtures/L3_PROMOTION_STRATEGY.md)** — Systematic path from L2 → L3 for each platform (version capture + response fixtures + CI validation).
- **[CI Workflow: validate-evidence-lock.yml](../.github/workflows/validate-evidence-lock.yml)** — Automated validation gate for lock integrity and version drift.

**Phase 1 Status (2026-03-19):** Infrastructure created for version pinning and fixture automation. All L2 platforms ready for version capture and L3 promotion. Next: Execute Phase 2 (golden fixture population) on first platform release.

---

## Lock Rubric (L0–L3)

- **L0 (Unlocked)**: no trustworthy first-party evidence.
- **L1 (Partial, integration-level)**: archive/internal evidence only.
- **L2 (Partial, first-party)**: official docs captured for commands/tools/permissions, but missing strict version/contract pin and/or response payload contract.
- **L3 (Locked)**: first-party docs + version/commit pin + exact command/tool/permission schema + payload contract + reviewer signoff.

---

## First-Party Verification Queue (Current)

| Platform | Level | First-party sources captured | Blocking gap to L3 |
| --- | --- | --- | --- |
| OpenCode | L3 | In-repo canonical evidence set | Revalidate against latest upstream release |
| GitHub Copilot CLI | L2 | GitHub Docs command/programmatic/ACP references | Runtime release pin + normalized response payload contract fixture |
| Gemini CLI | L2 | Google Gemini CLI command + tools references (captured in current pass) | Version pin + return payload fixture coverage |
| Codex CLI | L2 | OpenAI CLI reference + slash commands + approvals/security + changelog | Explicit built-in tool I/O fixture set and stable release pin strategy |
| Claude Code | L2 | Anthropic hooks + permissions references | Version pin and serialized response fixture matrix |
| Cursor | L2 | Cursor hooks reference (events, matchers, permission decisions, tool categories) | Version pin + full built-in tool argument/return contract set |
| VS Code + Copilot IDE Agent | L0 | None yet in this lock file | Official API/tool contract mapping + approvals model |
| Windsurf | L1 | Onboarding/install docs only | First-party built-in tool manifest + permissions + schemas |
| Kiro | L2 | Kiro hooks + MCP docs | Version pin + response payload contracts |
| Antigravity | L1 | Archive-only integration evidence | First-party tool manifest, permissions, and schema evidence |

---

## OpenCode — ✅ Locked (L3, repo-sourced)

- Primary source (in-repo): `archive/docs/RESEARCH_COMPETITORS.md`
- Version pin: research snapshot in repo (2026-02 era)
- Exact built-ins captured:
  - `ls(path, recursive)`
  - `grep(pattern, path, include, literal_text)`
  - `read(file_path)`
  - `view(file_path, offset, limit)`
  - `edit(file_path, ...)`
  - `patch(file_path, diff)`
  - `diagnostics(file_path)`
  - `bash(command, timeout)`
  - `fetch(url, format)`
  - `agent(prompt)`
- Permission model captured:
  - Allow once / allow session / deny
  - Non-interactive auto-approve mode
- Caveat:
  - Revalidate against latest upstream release cadence.

---

## GitHub Copilot CLI — 🟡 Partially Locked (L2, first-party)

- First-party references:
  - https://docs.github.com/en/copilot/how-tos/use-copilot-agents/use-copilot-cli
  - https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference
  - https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference
  - https://docs.github.com/en/copilot/reference/copilot-cli-reference/acp-server
- Last verified date: 2026-03-19
- Captured command/flag schema evidence:
  - `copilot`, `copilot help`, `copilot login/logout`, `copilot plugin` command family
  - approval/permissions flags: `--allow-all`, `--allow-tool`, `--deny-tool`, `--allow-url`, `--deny-url`
  - execution modes: interactive, `-p/--prompt` programmatic, ACP server via `--acp`
  - slash command inventory includes `/plan`, `/permissions`, `/mcp`, `/agent`, `/review`, `/usage`, etc.
- Captured tool/permission contract evidence:
  - tool permission pattern syntax (`Kind(argument)`), deny precedence over allow
  - session-level approval responses (`y`, `n`, `!`, `#`, `?`)
  - hook contracts (`preToolUse`, `agentStop`, `subagentStop`) with decision fields and tool names (`bash`, `powershell`, `view`, `edit`, `create`, `glob`, `grep`, `web_fetch`, `task`)
- Missing for L3:
  - runtime release pin policy (e.g., `copilot version` capture in evidence fixture)
  - normalized, machine-readable response payload fixture contract across interactive/programmatic/ACP flows

---

## Gemini CLI — 🟡 Partially Locked (L2, first-party)

- First-party references (captured in current pass):
  - Gemini CLI command reference (slash command inventory)
  - Gemini CLI tools reference (built-in tools and safety/confirmation semantics)
- Last verified date: 2026-03-19
- Captured evidence:
  - explicit slash commands including context/tools/MCP/permissions/planning surfaces
  - explicit built-in tool families (shell/file/web/planning/memory/interaction)
  - mutating tool confirmation and security posture documented
- Missing for L3:
  - stable release/version pin and changelog-based drift process
  - formal return payload fixture matrix per built-in tool

---

## Codex CLI — 🟡 Partially Locked (L2, first-party)

- First-party references:
  - https://developers.openai.com/codex/cli/reference
  - https://developers.openai.com/codex/cli/slash-commands
  - https://developers.openai.com/codex/agent-approvals-security
  - https://developers.openai.com/codex/changelog
- Last verified date: 2026-03-19
- Captured command/flag schema evidence:
  - global flags include `--ask-for-approval`, `--sandbox`, `--full-auto`, `--yolo`, `--search`, `--add-dir`
  - command inventory includes `codex exec`, `resume`, `fork`, `mcp`, `mcp-server`, `sandbox`, `cloud`, `features`, `execpolicy`
  - slash command inventory includes `/permissions`, `/plan`, `/mcp`, `/status`, `/diff`, `/review`, `/agent`, `/ps`
- Captured approval/sandbox semantics:
  - explicit sandbox-vs-approval layering
  - predefined combinations (`read-only`, `workspace-write`, `danger-full-access`) and policy behavior
  - protected path semantics and network/web-search risk controls
- Version evidence:
  - changelog includes explicit CLI release entries (example observed: `0.115.0`)
- Missing for L3:
  - pinned fixture set for built-in tool call inputs/outputs across modes
  - finalized release pinning rule adopted in this repo (exact CLI version + verification script)

## Claude Code — 🟡 Partially Locked (L2, first-party)

- First-party references:
  - https://code.claude.com/docs/en/hooks
  - https://code.claude.com/docs/en/permissions
- Last verified date: 2026-03-19
- Captured tool/permission schema evidence:
  - permission modes: `default`, `acceptEdits`, `plan`, `dontAsk`, `bypassPermissions`
  - rule syntax for `Bash`, `Read/Edit`, `WebFetch`, `MCP`, `Agent(...)`
  - precedence model (deny/ask/allow and managed settings behavior)
  - hook lifecycle events and decision outputs (`allow/deny/ask`, `block/allow`, `continue` semantics)
  - pre-tool built-ins and argument fields documented (`Bash`, `Write`, `Edit`, `Read`, `Glob`, `Grep`, `WebFetch`, `WebSearch`, `Agent`)
- Missing for L3:
  - reproducible release/version pin captured in lock fixture
  - canonical return payload fixture set for each built-in tool/event

## Cursor — 🟡 Partially Locked (L2, first-party)

- First-party references:
  - Cursor hooks reference (captured in current pass)
- Last verified date: 2026-03-19
- Captured evidence:
  - hook events (`preToolUse`, `postToolUse`, `beforeMCPExecution`, etc.)
  - matcher behavior and tool category matching (Shell/Read/Write/Grep/Delete/Task/MCP patterns)
  - permission decisions and schema fields (`allow`/`deny`/`ask` patterns where applicable)
  - hook config contract (`version`, `timeout`, `loop_limit`, `failClosed`, handlers)
- Missing for L3:
  - explicit release/version pin
  - full built-in tool argument/return contracts outside hook-surface documentation

## VS Code + Copilot IDE Agent — ❌ Unlocked (L0)

- Required evidence:
  - official docs + version pin
  - exact built-ins/signatures
  - approval/permission model semantics

## Windsurf — 🟡 Partially Locked (L1)

- First-party references:
  - Windsurf getting-started/onboarding docs (captured)
- Last verified date: 2026-03-19
- Captured evidence:
  - install/update and onboarding behavior
- Missing for L2/L3:
  - official built-in tools manifest
  - exact tool signatures and return shapes
  - permissions/approvals model details

## Kiro — 🟡 Partially Locked (L2, first-party)

- First-party references:
  - Kiro hooks docs
  - Kiro MCP docs
- Last verified date: 2026-03-19
- Captured evidence:
  - hook trigger taxonomy (pre/post tool use, file/task/manual triggers)
  - built-in category filters and MCP interaction model
- Missing for L3:
  - release/version pin
  - explicit return payload contracts for core built-ins

## Antigravity — 🟡 Partially Locked (L1)

- Primary source (archive/integration-level):
  - `archive/OmniRoute/docs/FEATURES.md`
  - `archive/OmniRoute/docs/CODEBASE_DOCUMENTATION.md`
  - `archive/docs/submodules/config_repos-Setup_Ultimate_OpenCode.md`
- Last verified date: 2026-03-19
- Captured evidence:
  - provider/integration existence and executor behavior notes
- Missing for L2/L3:
  - first-party built-in tools manifest + permissions and schema contracts
  - version/release pin from official source

---

## Evidence Quality Note

This file now contains a mixed evidence set:

- **L2 sections** are grounded in first-party docs fetched in this session.
- **L1 sections** remain archive/integration-level and are not yet authoritative contracts.

Do not claim parity complete until L1/L0 entries are promoted.

---

## TormentNexus Readiness Gate
## tormentnexus Readiness Gate

Do not claim “first-class parity complete” until:

- [ ] All target platforms are **L3 / Locked**
- [ ] Golden fixtures exist for tool call/response compatibility
- [ ] Alias profiles pass CI
- [ ] Permission model equivalence tests pass
- [ ] Changelog includes parity delta per release