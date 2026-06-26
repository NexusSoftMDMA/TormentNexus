---
id: t-b19e3d
title: Scarf spawns a new Hermes session on every chat, losing all previous context
status: todo
added: 2026-06-13
source: gh#99
---

## Description

> Imported from gh#99 — https://github.com/awizemann/scarf/issues/99

Environment
    - Scarf version: latest (as of 2026-05-22)
    - Hermes Agent: 2.9.1 (37)
    - OS: macOS 15.7.7

    Current behavior
    Every time I open or return to a chat in Scarf, it spawns a brand-new Hermes session instead of continuing the previous one. This causes complete loss of:
    - Conversation history
    - Active task state (TODO lists, work-in-progress)
    - Loaded skills and their configurations
    - Any accumulated context (file paths, decisions, etc.)

    Evidence from state.db
    Two sessions created within ~1 hour for the same chat thread:

    | Session ID             | Title                | Parent       | Messages | Created |
    |------------------------|----------------------|--------------|----------|---------|
    | 96d5d88f-...           | "Родитеские чаты"    | —            | 34       | 12:50   |
    | 20260522_134609_26bd26 | "Родитеские чаты #2" | 96d5d88f-... | 10       | 13:46   |

    The second session has parent_session_id pointing to the first, yet Scarf did not pass --continue or --resume, so Hermes started from scratch.

    Additional impact: context compaction inside sessions
    Even when staying inside one Scarf session, once it grows past ~30–40 messages, Hermes triggers [CONTEXT COMPACTION], collapsing early turns into a summary. After compaction, the model starts repeating old answers (e.g., responding about PDF support again even though that topic was already resolved 30 turns ago).

    Expected behavior
    Scarf should:
    1. Persist the Hermes session_id in its own chat state.
    2. On every subsequent interaction with the same chat thread, pass --resume <session_id> to the Hermes CLI / ACP call.
    3. Optionally expose a manual /new or /reset command if the user explicitly wants a fresh session.

    Why this matters
    Hermes is designed as a stateful agent with persistent memory, skills, and sessions. Scarf currently treats it as stateless, which breaks the core value proposition for any non-trivial task.

    Workaround (for users until fixed)
    Use Telegram (or CLI) for long-running tasks — those channels properly keep sessions alive. Use Scarf only for quick, self-contained queries.

## Plan



## Artifacts



