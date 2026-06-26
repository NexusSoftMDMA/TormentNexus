"""Identity encoder — never transforms, always emits JSON."""

from __future__ import annotations

ENCODING_ID = "json"


def encode(response: dict) -> tuple[None, str]:
    """Passthrough: return sentinel indicating caller should emit the raw dict."""
    return None, ENCODING_ID


def decode(payload: str) -> dict:
    import json as _json
    return _json.loads(payload)
