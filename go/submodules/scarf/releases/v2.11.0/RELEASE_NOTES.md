# Scarf v2.11.0

A coordinated catch-up to **Hermes v0.16.0 (2026.6.5)**. The headline is a correctness fix for the first `state.db` schema change since v0.11 — Hermes can now **soft-delete messages** when you rewind a conversation, and Scarf needs to hide them to show you what the agent actually sees. Alongside it: live session titles, a "rewound" indicator, the Spotify-skill-became-a-plugin move, two real gateway-config bugs fixed, and a handful of v0.16 quick wins. Every new surface is capability-gated or schema-detected, so pre-v0.16 hosts render byte-identical to v2.10.3. All flag, config, and wire shapes were verified against the live Hermes v0.16 source before implementation.

## Rewound messages are hidden, matching what the agent sees

v0.16 is the first `state.db` schema change since v0.11: the `messages` table gained an `active` column. When you run `/undo` in Hermes, the rewound messages are flipped to inactive (a soft delete) rather than being removed — so the agent no longer sees them, but they're still on disk. Scarf was reading every row regardless, which meant a rewound chat showed you messages the agent had already discarded.

Scarf now filters to active messages on **every** message-read path — chat transcript, resume, skeleton loaders, the lot. The filter is schema-gated: Scarf detects the column first (a `PRAGMA` check locally, a preflight on remote hosts), so connecting to a pre-v0.16 Hermes that doesn't have the column still works exactly as before instead of erroring with "no such column." The net effect: what you see in a rewound conversation now matches what the agent is actually working from.

## Session titles update live

When Hermes generates or regenerates a session's title, v0.16 now pushes an ACP notification (`session_info_update`). Scarf listens for it and updates the title in the Mac chat sidebar in place — no need to navigate away and back to see a freshly-titled session get its name. (Title-setting itself remains a Hermes/gateway concern; there's no ACP path to set a title from Scarf, by design.)

## "Rewound ×N" badge on sessions

v0.16 tracks how many times a session has been rewound in a new `sessions.rewind_count` column. Scarf schema-detects it and surfaces a **"rewound ×N"** badge in the chat sidebar, the Sessions table, and the dashboard rows, so a heavily-edited session is visible at a glance. Like the message filter, this is schema-gated and read by column name, so older hosts simply don't show the badge.

## Spotify sign-in moved to Plugins

In v0.16, Spotify is no longer a skill — it became a built-in tool with a native plugin. Scarf follows suit: on a v0.16 host, the Spotify sign-in affordance now lives in the **Plugins** view instead of Skills (the `hermes auth spotify` PKCE flow is unchanged). The old Skills-view path is kept for pre-v0.16 hosts, so both render correctly depending on what you're connected to.

## Two gateway-config bugs fixed

Both were silent failures found while verifying Scarf's writes against the live v0.16 source:

- **Platform allowlists now actually apply.** Scarf was reading and writing channel/chat/room allowlists under `gateway.platforms.<platform>.*`, but Hermes enforces them from the **top-level** platform section (`slack.allowed_channels`, `telegram.allowed_chats`, `matrix.allowed_rooms`, `dingtalk.allowed_chats`). The mismatch meant an allowlist you set in Scarf silently never took effect. Scarf now writes to the correct top-level location, merging into the existing platform block so your other settings are preserved. (This also fixes DingTalk, which uses `allowed_chats`, not `allowed_rooms`.)
- **The gateway digest works again.** `hermes gateway list --json` doesn't exist in v0.16 — the command prints a text table and errors on the flag — so Scarf's cross-profile gateway digest was silently hiding. Scarf now runs the plain command and parses the text output (status, profile, PID), so the digest shows up.

## v0.16 quick wins

- **Optimize sessions database** — a new Health-view action wired to `hermes sessions optimize` (FTS merge + VACUUM), gated on the host advertising it.
- **Kanban goal pill** — cards now decode `goal_mode` / `goal_max_turns` and render a **"Goal · N"** pill when a task is running in goal mode.
- **Session rename** is gated on the v0.16 `hermes sessions rename` capability (the rename path Scarf already drove).

## Under the hood

- **Config-shape corrections, verified against live v0.16.** OpenRouter response-caching is now written as a scalar bool — so toggling it *off* actually takes effect (the old nested shape wrote a value Hermes always read as "on"). The MCP SSE server-add path was corrected to match the real `hermes mcp add` flags (URL on add, then `transport: sse` stamped into the YAML). A dead Kanban `verify` verb was removed (Reject now routes through the real `comment` + `archive` verbs). The `curator list-archived` text output is parsed directly instead of retrying a non-existent `--json` flag. ACP `/goal` and `/subgoal` were de-advertised from the slash-command surface, since they're gateway-only in v0.16 and no-op over ACP.
- **New providers and models flow automatically.** AWS Bedrock is registered as an overlay-only provider; the other v0.16 providers self-surface from the model catalog with no Scarf change, and new models (deepseek-v4-flash, MiniMax-M3, qwen3.7-plus, gemini-3.5-flash) appear automatically. The upstream catalog now refreshes hourly.
- **New capability flags** — `hasSessionsRename`, `hasSessionsOptimize`, `hasKanbanGoalMode`, `hasInsightsCommand`, `hasDashboardCommand`, plus the `isV016OrLater` predicate — gate every surface above.
- **628 ScarfCore tests** pass, including new soft-delete SQL-shape, allowlist sibling-preservation, config round-trip, and gateway text-parser tests.

## Upgrade notes

- Sparkle will offer this update automatically on next launch (or **Scarf → Check for Updates**).
- macOS 14.6+ (Sonoma) deployment target unchanged.
- **Fully backward compatible.** Every v0.16 surface is capability-gated or schema-detected; if you're still on Hermes v0.15.x or earlier, Scarf renders exactly as v2.10.3 did. There are no data migrations.
- **iOS testers:** the shared-core changes (the `messages.active` filter, rewind count, config-shape fixes) ride in ScarfCore; a ScarfGo TestFlight build carrying them is queued separately on the iOS track.
