---
name: scarf-cron
description: Schedule a recurring Hermes cron job for the active project
argumentHint: <what the job should do, e.g. "fetch latest hn comments daily">
version: 1.0.0
---

The user wants to register a scheduled (cron) job for the active Scarf project. The active project's path is in the `<!-- scarf-project -->` AGENTS.md block; read it first. If no project is active, ask which project the job belongs to (or whether they want a global job).

User's request: {{argument | default: "(no specifics yet — ask the user what the job should do and how often)"}}

What you need from the user:

1. **What the job does** — the prompt the agent will receive each tick.
2. **Schedule** — natural-language ("every weekday at 9am") or cron expression ("0 9 * * 1-5").
3. **Delivery** — where results land. Options: `print` (just log), or a messaging platform like `telegram`, `discord`, `slack`, `signal`, etc. Use `all` to fan out to every connected platform (Hermes v0.14+).
4. **(Optional) Model** — the global default is used otherwise.
5. **(Optional) Context-from** — chain on another job's output (read its name from `hermes cron list --json`).

Then run:

```bash
hermes cron create \
  --name "<short descriptive name>" \
  --schedule "<schedule>" \
  --prompt "<the prompt>" \
  --workdir "<project.path>" \
  --deliver "<delivery>"
```

Always pass `--workdir <project.path>` for a project-scoped job — it makes the spawned agent inherit AGENTS.md, the dashboard, and resolve relative paths against the project's files. Without it the job runs against `$HOME` and the project context is lost.

Confirm the job landed by running `hermes cron list` (or `--json` for parseable output) and reporting the new entry to the user. Mention they can see + manage it in Scarf's Cron sidebar.
