# Universal Agent Installer Prompt

`jcm install <client>` covers the agent/IDE environments we ship first-class
support for: Claude Code, Claude Desktop, Cursor, Windsurf, Continue. Other
clients — Codex CLI, Cline, Roo Code, JetBrains AI, Aider, Zed, Goose, custom
in-house agents — have their own instruction mechanisms (project rules,
custom instructions, system prompts, plugin manifests, agent.md files, etc.)
that we don't (yet) have a CLI installer for.

Rather than guess which clients matter, we ship the prompt below for the
gap. Paste it into your agent and let *it* figure out where to install
jCodemunch (and jDocMunch) guidance for its own environment. The prompt is
environment-agnostic by design.

When the prompt finishes, it emits a **compatibility report**. If your
environment isn't covered by a first-class installer yet, paste that report
into a new GitHub issue (https://github.com/jgravelle/jcodemunch-mcp/issues) —
that's how we decide which client gets first-class support next.

---

## Known first-class targets — use these instead

Before running the universal prompt, check whether your environment is
already covered by a first-class CLI installer. If it is, the CLI is faster,
better tested, and the right answer:

| Environment | Use this instead |
|---|---|
| Claude Code | `jcm install claude-code --skills` (writes the Claude Agent Skill bundle) |
| Claude Desktop | `jcm install claude-desktop` |
| Cursor | `jcm install cursor` (writes `.cursor/rules/jcodemunch.mdc`) |
| Windsurf | `jcm install windsurf` |
| Continue | `jcm install continue` |
| All of the above at once | `jcm install all` |

Run `jcm install --list` to enumerate, and `jcm install-status --json` to
check what's already wired.

**The universal prompt below is for everything else** — Codex CLI, Cline,
Roo Code, JetBrains AI, Aider, Zed, Goose, custom in-house agents, and any
agent/IDE not in the table above. If you run the universal prompt against
one of the known targets, the prompt's Step 1 environment detection will
correctly tell you to use the CLI installer instead.

The one-click install badges at the top of the README handle a *different*
layer — they wire the MCP server itself into a client's config (so the
tools become callable). The universal prompt handles the guidance layer
(teaching the agent *when* to call those tools). You may need both:
badge/CLI for the wiring, then the universal prompt for the guidance — if
your environment isn't in the first-class table.

---

## How to use it

1. Confirm the jCodemunch MCP server is already configured in your agent
   (run `jcm install <client>` first if it isn't, or wire the MCP manually).
2. Open a fresh session with your agent.
3. Paste the prompt verbatim under the divider below.
4. Answer any clarifying questions the agent asks (most should be answerable
   from inspecting your filesystem and config).
5. Copy the final compatibility report and either keep it for your own records
   or attach it to a feature request.

The prompt does **not** assume you're running Claude. It does **not** assume
file writes are allowed. It does **not** assume the agent has a "skills"
concept. It discovers all of that before installing anything.

---

## The prompt

```
You are installing jCodemunch (and, if present, jDocMunch) usage guidance into
your own environment. Do not assume you are Claude, Codex, Cursor, or any
specific client. Discover before you act. If a step can't be verified, say so
explicitly rather than claiming success.

# Step 1 — Identify the environment

Determine, from observable signals only:

- Which agent/IDE runtime is hosting this conversation (Claude Code, Codex,
  Cursor, Windsurf, Continue, Cline, JetBrains AI, a custom in-house agent,
  something else).
- Which version, if exposed.
- Whether tool/function calling is available and which tools you can see right
  now. List the jCodemunch tool names you have access to (they will be
  prefixed with mcp__jcodemunch__ or similar).
- Whether jDocMunch tools are also available (mcp__jdocmunch__*).

If you cannot identify the environment with reasonable confidence, ask the
user one targeted question. Do not guess.

**Short-circuit for known first-class targets.** If the detected environment
is Claude Code, Claude Desktop, Cursor, Windsurf, or Continue, STOP. A CLI
installer already exists. Tell the user to run the appropriate command and
do not proceed with the rest of this prompt:

- Claude Code: `jcm install claude-code --skills`
- Claude Desktop: `jcm install claude-desktop`
- Cursor: `jcm install cursor`
- Windsurf: `jcm install windsurf`
- Continue: `jcm install continue`

Only continue past Step 1 if the environment is NOT in that list.

# Step 2 — Discover the instruction mechanism

Every agent environment has at least one supported way to install persistent
guidance. Examples (non-exhaustive):

- Claude Code: ~/.claude/skills/<name>/SKILL.md (global) or
  ./.claude/skills/<name>/SKILL.md (project); CLAUDE.md at repo root.
- Codex: project rules, agents.md, or per-repo system prompts depending on
  the host.
- Cursor: .cursorrules at repo root, or ~/.cursor/rules/.
- Windsurf: .windsurfrules at repo root.
- Continue: ~/.continue/config.json system prompt or .continue/ project rules.
- Cline / Roo: .clinerules, custom instructions.
- JetBrains AI: project rules / custom instructions panel.
- Generic: a top-level AGENTS.md / .agent/INSTRUCTIONS.md convention.

For the environment you identified in Step 1, determine:

a. Which of those mechanisms it actually honours. Check the documentation if
   you have web access, or inspect existing files in the user's home directory
   and current repo for telltales (e.g. .cursorrules existing implies Cursor).
b. Whether the environment supports BOTH user-level (global, across all
   projects) AND project-level (current repo only) scopes. If yes, plan to
   install at both unless the user opts out.
c. Whether the environment allows you to write files. If file writes are
   blocked, you must produce manual-install instructions (exact paths + exact
   file contents) instead of writing.

# Step 3 — Respect the responsibility split

The MCP server does the work. The installed instructions only teach the agent
*when* to call which MCP tool. Do not duplicate MCP logic into the installed
file. Specifically:

- jCodemunch MCP provides: indexing, symbol/text search, file outlines, symbol
  source retrieval, import/reference graphs, blast-radius analysis, refactor
  planning, edit registration, dead-code detection, task-context assembly,
  health/risk scoring, and a live-policy tool jcodemunch_guide.
- jDocMunch MCP provides: documentation indexing, section search, TOC
  browsing, document outlines, section retrieval.
- Your installed instructions should be a thin onboarding layer that tells the
  agent: "for code navigation, prefer jCodemunch tools over Read/Grep/Glob;
  for indexed documentation navigation, prefer jDocMunch tools."

Use the live-policy tool when available:

- If jcodemunch_guide is exposed, instruct the agent to call it for the
  authoritative tool-selection decision tree rather than hard-coding a stale
  copy in the installed file.

# Step 4 — Preserve native-tool exceptions

The installed guidance MUST allow the agent to use its native shell / file
tools for the following cases, even when jCodemunch is available:

- Exact, known file paths (no search needed).
- Test files when running tests or reading test output.
- Command output (running commands, reading logs).
- Files not in any indexed repo (generated files, local-only files).
- Files outside the index.
- Pre-edit line-number verification before calling an Edit tool.
- Reading complete process-control files (README, CONTRIBUTING, AGENTS.md,
  CLAUDE.md, etc.) when the repo or workflow requires them to be read in full.
  Indexed section retrieval IS NOT a substitute for a complete read of those
  files.

If the installed file omits these exceptions, the agent will over-call MCP
tools for trivial known-path reads, which wastes tokens and frustrates users.

# Step 5 — Draft the installed file contents

Compose a single file (or the smallest set of files the environment requires)
that includes, at minimum:

1. A short identity line (what jCodemunch / jDocMunch are).
2. The trigger rule ("for code exploration, prefer these tools over native
   Read/Grep/Glob").
3. A compact decision tree for the most common operations (symbol search,
   file outline, source retrieval, import graph, reference search, blast
   radius, dead code, edit registration).
4. The native-tool exceptions from Step 4.
5. A pointer to jcodemunch_guide / jdocmunch_guide for live policy.

Format the file in whatever syntax the target environment expects (YAML
frontmatter + markdown for Claude skills; plain markdown for .cursorrules /
.windsurfrules; JSON for Continue config; etc.). Do not invent syntax — match
what the environment documents.

# Step 6 — Preserve existing user content

Before writing:

- If a file at the target path already exists, READ it completely.
- If it contains user-authored content unrelated to jCodemunch, do NOT
  overwrite. Either merge (preserve their content, append yours in a clearly
  marked section) or refuse and report.
- If it contains a prior jCodemunch section, replace only that section.
- Always make a .bak copy of any file you modify, unless the user opts out.

# Step 7 — Verify or refuse

Do not claim installation succeeded unless you actually verified the file
exists at the target path with the expected contents. If writes were blocked,
say "manual install required" and produce the exact paths and file contents
the user must create themselves.

# Step 8 — Emit a compatibility report

End your response with a fenced markdown report in this exact shape:

```
## jCodemunch agent-environment compatibility report

**Detected environment:** <name + version>
**Instruction mechanism used:** <skills | project rules | custom instructions | plugin manifest | other>
**Scopes installed:** <global | project | both | none — manual install>
**Target paths:**
- <path 1>
- <path 2>
**jCodemunch MCP tools detected:** <count, or "not configured">
**jDocMunch MCP tools detected:** <count, or "not configured">
**Files created or modified:**
- <path — created | merged | replaced jcm section | manual instructions only>
**Verification performed:** <yes — file exists with expected content | no — writes blocked>
**Limitations / caveats:** <anything the user should know>

### Upstream issue summary

If first-class jCodemunch support for this environment would be useful, the
maintainer would need:

- <one-line description of the install mechanism>
- <one-line description of the file format>
- <one-line description of the install paths>
- <anything environment-specific that would surprise an installer>
```

Paste this report into a new GitHub issue at
https://github.com/jgravelle/jcodemunch-mcp/issues if you want first-class
installer support for this environment.
```

---

## Notes for maintainers

This doc is intentionally environment-agnostic. The prompt does the
environment-specific work at runtime, using whatever the agent itself can
observe. That's the whole point — we don't have to know every client up front,
and the compatibility reports come back as demand signal that drives which
client gets first-class support next.

If you're considering a first-class installer for a specific client, the
compatibility reports filed in the issue tracker are the input. Five reports
for the same environment is a much stronger signal than a feature request
asking for it abstractly.
