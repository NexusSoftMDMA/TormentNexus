# Scarf v2.9.1

Patch release covering [#97](https://github.com/awizemann/scarf/issues/97) — the v2.9.0 mid-chat model switcher and per-project model preset binding both fired `session/set_model` with the bare model ID, never the colon-prefixed `<provider>:<model>` wire format Hermes's ACP adapter expects. For model IDs Hermes can auto-detect from the name alone (e.g. `claude-opus-4-7`), this worked. For less-obvious IDs (e.g. OpenRouter's `<author>/<model>` slugs like `inclusionai/ring-2.6-1t`), Hermes's `_resolve_model_selection` would fall through to `detect_provider_for_model` and infer the wrong provider — the session ended up routing through whatever provider was previously active, and the user got an opaque rejection from the wrong upstream.

## Fix

- **`ACPClient.setSessionModel`** gains an optional `providerID:` parameter. When supplied, the wire `modelId` is encoded as `"<provider>:<model>"` (matching `acp_adapter/server.py`'s `_encode_model_choice`). When `nil` or empty, falls back to the bare-model wire shape so existing call sites that genuinely don't know the provider keep working.
- **`ChatViewModel.switchModelPreset`** and **`ChatViewModel.applyProjectModelPreset`** now pass `preset.providerID` through. The "Use global default" mid-chat path additionally reads `config.provider` from `config.yaml` so reverting to the global default lands on the same provider the CLI would.
- New public `ACPClient.encodeModelChoice(modelID:providerID:)` static helper exposes the encoding logic for tests + any future call sites that need the wire form without firing the RPC.
- 6 new tests in `M1ACPTests` cover the encoding rules: bare model when provider absent / empty / whitespace, colon-prefix when both present, provider lower-casing (matches Hermes), whitespace trimming, empty-model no-op semantics.

## Not affected

- The `/model <X>` slash command typed in Scarf chat — Scarf passes the literal text through; the parsing bug lives entirely in Hermes's ACP `_cmd_model` handler (per the RCA in [#97](https://github.com/awizemann/scarf/issues/97)). Workaround: use the new chat-header model badge in v2.9 instead, which now resolves correctly via this fix.
- Pre-v0.13 hosts — no `session/set_model` RPC; preset apply silently no-ops as before.
- Model presets saved before v2.9.0 with empty `providerID` — fall back to the bare-model wire shape, identical to v2.9.0 behavior. Re-save the preset to opt into colon-encoding.

## Hermes compatibility

Targets Hermes v0.14.0 (v2026.5.16). The fix is forward-compatible — Hermes's `_resolve_model_selection` accepts both the colon-prefixed and bare-model wire shapes on every release of v0.13+, so v2.9.1 works against v0.13 hosts too.
