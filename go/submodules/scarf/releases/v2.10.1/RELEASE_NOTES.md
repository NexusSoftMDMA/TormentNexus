# Scarf v2.10.1

A "projects fundamentals" maintenance release built on user feedback after v2.10.0 shipped. Six interlocking fixes that make project work in Scarf feel deterministic again: the slash menu surfaces what you can actually do, the new-project wizard reliably triggers the right skill, the AGENTS.md block teaches the agent what Scarf actually offers, the Skills sidebar finally shows the bundled `scarf-template-author`, and a new `/scarf-*` family of slash commands is available in every chat. Plus a Health-view capability diagnostic so the next time something looks sparse, you can tell at a glance whether the version gate is alive.

## Global `/scarf-*` slash commands

Six bundled slash commands that drive Scarf-specific project workflows — available in **every chat** (pre-session, global, and project-scoped), not just per-project:

- `/scarf-new <one-liner>` — kicks off the `scarf-template-author` interview to scaffold a new project from scratch.
- `/scarf-help` — concise tour of Scarf's feature surface (dashboard widgets, Kanban, model presets, slash commands, cron, etc.).
- `/scarf-dashboard <change>` — design or edit the active project's `dashboard.json`.
- `/scarf-widget <kind>` — add a single widget to the active dashboard.
- `/scarf-cron <description>` — schedule a recurring `hermes cron` job for the active project.
- `/scarf-export` — prepare + run the `.scarftemplate` export of the active project.

The commands live as `.md` files at `~/.hermes/scarf/slash-commands/`. A new `SlashCommandBootstrapService` copies the bundled set into that directory on app launch, with the same idempotent + version-gated upgrade pattern as `SkillBootstrapService`: a frontmatter `version: x.y.z` field is the source of truth, hand-edited copies are preserved, future shipped updates land automatically when their version is newer. Per-project commands of the same name (at `<project>/.scarf/slash-commands/`) still win — author your own `/scarf-help` and it overrides the bundled one.

## Skills sidebar: `scarf-template-author` now appears

The bundled `scarf-template-author` skill was being installed at `~/.hermes/skills/scarf-template-author/SKILL.md` (flat), but Scarf's `SkillsScanner` expects the `<category>/<skill>/SKILL.md` two-level layout, so the skill never showed in the Skills sidebar — even though Hermes itself loaded it correctly. `SkillBootstrapService` now installs into `~/.hermes/skills/scarf/<skill>/` and auto-migrates the existing flat install on first launch. The skill shows up under a new "scarf" category in the Skills view; no user action needed.

## Pre-session slash menu

Before v2.10.1, opening a fresh app and typing `/` in the chat input collapsed the menu down to just `/new` — every session-required command (`/clear`, `/compact`, `/cost`, `/model`, `/tools`, `/reload-skills`, `/help`, `/exit`, `/yolo`, `/sessions`, `/codex-runtime`, `/steer`, `/goal`, `/queue`, `/subgoal`) was filtered out and indistinguishable from "the menu is broken". Now those commands all stay in the menu, greyed out, with a tooltip that reads **"Available once a chat is open. Press Return on `/new` (or click an existing session) to start one."** The pattern matches the existing `/steer`-on-pre-v0.13-idle grey-out; only the trigger is different.

## New-project wizard hand-off

The "New Project from Scratch" wizard's kickoff prompt was a single polite sentence (*"Use the `scarf-template-author` skill to walk me through configuring it…"*) that agents routinely treated as a suggestion rather than an instruction — colder models would reply conversationally without firing the skill. The new prompt anchors on `SKILL:` and `PROJECT_PATH:` markers that pattern-trained agents recognize as invocation, lists the skill's expected interview stages explicitly, and tells the agent to "Start with question 1." If the user filled in the optional one-liner, it's threaded as the answer to question 1 so the agent jumps straight to question 2.

Also: `NewProjectViewModel.commit()` now runs `SkillBootstrapService.ensureBundledSkillsInstalled()` as a synchronous preflight before handing off to chat, so the bundled skill is **guaranteed** on disk before Hermes's `session/new` runs its skill-index scan. The launch-time bootstrap is a detached task that might not have finished yet on cold launch or slow remotes; this makes the wizard self-contained.

## AGENTS.md `scarf-project` block: Scarf platform reference

The Scarf-managed block in `<project>/AGENTS.md` previously surfaced only project bookkeeping (path, dashboard location, template id, config field names, cron jobs, Kanban tenant, uninstall manifest). Agents walked into projects with no idea what Scarf actually offers beyond bare Hermes — the most common failure mode being agents proposing a shell script when a dashboard widget would do the job in one line of JSON.

The block now appends a static "Scarf platform reference" section describing the dashboard widget vocabulary, project slash command authoring (the `<!-- scarf-slash:<name> -->` expansion marker), the Kanban tenant convention, model presets, the typed configuration schema with Keychain-backed secrets, cron `--workdir` scoping, where Hermes loads skills from, and the export-to-template flow. The section is idempotent (byte-identical across refreshes), secret-safe (no values appear, only schema field names), and capped around 30 lines so it doesn't crowd the agent's context.

`ProjectTemplateInstaller` now also refreshes the block on install (previously only chat-start did) so a freshly-installed template project has the block on disk before the user opens its first chat.

## Health view: capabilities diagnostic panel

The Health view gets a new "Hermes Capabilities" panel at the top of its scroll area showing:

- The raw `hermes --version` line as parsed by Scarf
- A `… · N capabilities active` summary with per-release flag rows (v0.12 / v0.13 / v0.14 / v0.15) that show green when the connected Hermes is at or above that release
- A **Re-detect** button that re-runs `hermes --version` on demand
- An explanatory note for when a flag is missing ("UI for that release is hidden because the connected Hermes is older")

The capabilities store also now auto-refreshes when Scarf returns to the foreground (`NSApplication.didBecomeActive`) — so if you run `hermes update` in a Terminal while Scarf is backgrounded, the slash menu, Kanban tab, and other version-gated UI pick up the new version without needing a Scarf relaunch.

## Repo memory: Memophant

Internal note for contributors: as of this release Scarf's repo memory (the developer-facing notes, wiki, design system, and CLAUDE.md guidance for AI coding sessions) is managed via [Memophant](https://github.com/awizemann), a memory manager I built for exactly this. Several recent commits on `main` show as "via Memophant" — those are repo-memory migrations and consolidations, not Scarf app changes. Memophant will be open-sourced shortly; until then the only user-visible artifact is a new managed block at the bottom of `CLAUDE.md` describing the layout, plus a `wiki/` working directory and a `TASKS.md` kanban file. Nothing in Scarf the app depends on Memophant; it's a workflow tool for the repo.

## Hermes compatibility

Unchanged from v2.10.0 — targets Hermes v0.15.0 (v2026.5.28). Pre-v0.15 hosts continue to work; the v2.10.1 changes are all Scarf-side (slash command bundle, skill install layout, AGENTS.md block content, Health panel) and require no Hermes upgrade.

## Upgrade notes

- The Sparkle appcast at `https://awizemann.github.io/scarf/appcast.xml` will offer this update automatically on next launch.
- macOS 14.6+ (Sonoma) deployment target unchanged.
- **One-time migration runs at first launch**: the existing flat `~/.hermes/skills/scarf-template-author/` install is removed and re-installed under `~/.hermes/skills/scarf/scarf-template-author/`. Idempotent and non-fatal — if the migration fails (locked file, permissions), the install proceeds and the old flat copy stays where it was (the agent still loads it, and the Skills view will pick up the categorized copy on the next launch).
- No breaking changes to `~/.hermes/` state. The new `~/.hermes/scarf/slash-commands/` directory is created lazily on first bootstrap.
