---
id: t-b8a6c3
title: Performance and reliability issues on large state.db (lag, crashes, missing sessions)
status: todo
added: 2026-06-13
source: gh#61
---

## Description

> Imported from gh#61 — https://github.com/awizemann/scarf/issues/61

## Environment

- Scarf v2.5.2 (also reproducible on v2.5.1)
- Hermes v0.11.0 (2026.4.23)
- macOS 15 (M4 MacBook Air, 24GB RAM)
- state.db: 69MB, 63 sessions, 5,000+ messages

## Problems

### 1. Typing lag and unresponsive buttons in chat

The chat input bar has severe input lag — keystrokes take 300-500ms to appear, and sidebar/settings buttons frequently require multiple clicks. This gets worse as the session message count grows.

**Likely cause:** `ChatViewModel.startACPSession()` calls `richChatViewModel.loadSessionHistory()` synchronously on `@MainActor` (line 478 of `ChatViewModel.swift`). For sessions with hundreds of messages and tool calls, this blocks the main thread for seconds. During ACP prompts, the event loop (`startACPEventLoop`) and health monitor also run on `@MainActor`, creating contention with user input processing.

### 2. ACP approval callback crashes with Hermes v0.11

Every prompt that triggers a dangerous command approval fails with:

```
TypeError: make_approval_callback.<locals>._callback() got an unexpected keyword argument 'allow_permanent'
```

Hermes v0.11 added an `allow_permanent` parameter to the approval callback interface. `acp_adapter/permissions.py` line 43 has a fixed signature of `(command: str, description: str)` that rejects extra kwargs. The prompt silently hangs — no error surfaced to the UI.

**Fix:** Change the callback signature to accept `**kwargs`.

### 3. Session loading failures ("no session ID" / "ACP request initialize timed out")

When the gateway is also running (even the correct profile-aware instance), ACP session initialization frequently times out. Scarf shows "Starting..." indefinitely or falls back to the error state. This is intermittent but frequent — roughly 1 in 3 session starts fail.

**Likely cause:** The gateway and ACP subprocess both access `state.db` in WAL mode. Under concurrent load from the gateway's Discord sync + skill registration + cron scheduler, ACP's `session/new` or `session/load` calls stall on database locks. Scarf's 30-second timeout on `initialize` fires before the lock releases.

### 4. Sessions created in WebUI/gateway don't appear in Scarf

Of 63 total sessions (40 webui, 7 acp, 8 cli, 5 discord, 3 api_server), Scarf only shows the 7 ACP ones plus a handful of recent CLI sessions. This is confusing — users expect their sessions to be available regardless of which client created them.

**Likely cause:** ACP's `session/load` uses a different session ID format than gateway-created sessions. Gateway/webui sessions have IDs that ACP doesn't recognize, so `loadSession()` returns nil and Scarf falls back to creating a new session. The old conversation history never loads.

### 5. state.db WAL journal grows unboundedly

After 48 hours of mixed Scarf + WebUI usage, `state.db-wal` grew to 65MB — matching the database itself at 69MB. Reads pass through the WAL, so a bloated journal effectively doubles the I/O cost of every query. `PRAGMA wal_checkpoint(TRUNCATE)` fixes it temporarily, but it regrows within a day of normal use.

**Suggestions:** Either run periodic checkpoints automatically, enable `PRAGMA auto_vacuum`, or add a maintenance task that triggers after gateway restarts or session pruning.

### 6. No system notification when reply completes

The "Bell on Complete" setting plays a sound, but there's no banner/notification center alert. Users who tab away from Scarf while waiting for long agent replies have no way to know when the response is ready without manually checking. The `NotificationRouter` skeleton exists but is hard-gated behind `apnsEnabled = false`.

## Reproduction

1. Have a state.db with 40+ sessions and 3,000+ messages (typical after 1-2 weeks of use)
2. Open Scarf, navigate to Chat
3. Try typing in the input bar — note the lag
4. Click sidebar items (Settings, Sessions, Skills) — note unresponsiveness
5. Resume a session with 50+ messages — observe multi-second freeze before the transcript appears
6. Start a session from WebUI, then try to open it in Scarf — it won't appear

## Related

- #60 — Add toggle to hide inspector pane (fixed 3-pane layout wastes space on smaller windows)

## Plan



## Artifacts



