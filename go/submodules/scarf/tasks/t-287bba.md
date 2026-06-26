---
id: t-287bba
title: Informational: /model change in Scarf not working properly - Bugs reported to Hermes
status: todo
added: 2026-06-13
source: gh#97
---

## Description

> Imported from gh#97 — https://github.com/awizemann/scarf/issues/97

Hi,

When using Scarf (or any ACP client like Zed) to send `/model inclusionai/ring-2.6-1t --provider openrouter` to Hermes, the `--provider openrouter` flags are **completely ignored**. The entire string `"inclusionai/ring-2.6-1t --provider openrouter"` is treated as a literal model ID and passed to the **current default provider** (OpenAI Codex in my personal case), which rejects it with an error.

The CLI path (`/model` typed in the Hermes terminal) correctly parses `--provider`/`--global` flags and routes the model to the specified provider. The ACP adapter path does not — it sends the raw, unparsed args string straight through to model resolution.


## Root Cause Analysis (Code)

The `_cmd_model` handler at `acp_adapter/server.py:1510`:

```python
def _cmd_model(self, args: str, state: SessionState) -> str:
    # ...
    current_provider = getattr(state.agent, "provider", None) or "openrouter"
    # BUG: passes entire args string including "--provider openrouter" flags
    target_provider, new_model = self._resolve_model_selection(args, current_provider)
    # ...
```

`parse_model_input()` in `models.py:1663` does not understand `--provider` flags. It sees `"inclusionai/ring-2.6-1t --provider openrouter"`, finds no colon-based `provider:model` syntax, and returns `(current_provider, "inclusionai/ring-2.6-1t --provider openrouter")`. The current provider is `openai-codex` (per config), so it tries to send `--provider openrouter` as part of the model name to OpenAI Codex, which fails.

## Proposed Fix

Modified `_cmd_model` in `acp_adapter/server.py` to call `parse_model_flags()` before resolving, matching the CLI behavior:

```python
from hermes_cli.model_switch import parse_model_flags

model_input, explicit_provider, _persist_global = parse_model_flags(args)

current_provider = getattr(state.agent, "provider", None) or "openrouter"
if explicit_provider:
    target_provider = explicit_provider
else:
    target_provider = current_provider

new_model = model_input or state.model or ""

if new_model:
    target_provider_resolved, new_model = self._resolve_model_selection(
        new_model, target_provider
    )
else:
    target_provider_resolved = target_provider
```

## Verification

Tested `parse_model_flags()` with all relevant input patterns:

| Input | model_input | explicit_provider | persist_global |
|-------|-------------|-------------------|---------------|
| `"inclusionai/ring-2.6-1t --provider openrouter"` | `"inclusionai/ring-2.6-1t"` | `"openrouter"` | `False` |
| `"sonnet --provider anthropic --global"` | `"sonnet"` | `"anthropic"` | `True` |
| `"--provider openrouter"` | `""` | `"openrouter"` | `False` |
| `"anthropic:claude-sonnet-4-6"` | `"anthropic:claude-sonnet-4-6"` | `""` | `False` |
| `""` | `""` | `""` | `False` |

Existing `provider:model` colon syntax continues to work correctly (preserves backward compatibility with the existing test at `tests/acp/test_server.py:1404`).

## Plan



## Artifacts



