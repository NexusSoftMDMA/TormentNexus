"""Public decoder entry point — rehydrate MUNCH payloads to JSON-compatible dicts.

Usage:
    from jcodemunch_mcp.encoding.decoder import decode
    obj = decode(payload_str)

Dispatches by encoding id from the header. Falls back to JSON parsing
when the payload isn't a MUNCH envelope.
"""

from __future__ import annotations

import json

from .format import HEADER_PREFIX, parse_header


def decode(payload: str) -> dict:
    if not payload.startswith(HEADER_PREFIX):
        return json.loads(payload)
    head = payload.splitlines()[0]
    meta = parse_header(head)
    enc = meta.get("enc", "gen1")
    module = _load_schema(enc)
    return module.decode(payload)


def _load_schema(encoding_id: str):
    from . import generic
    if encoding_id == generic.ENCODING_ID:
        return generic
    try:
        from .schemas import registry
        return registry.get(encoding_id)
    except Exception:
        return generic
