"""Per-tool compact encoders.

Each module exposes:
    ENCODING_ID: str         — versioned id (e.g. "ch1")
    TOOLS: tuple[str, ...]   — tool names this encoder handles
    encode(tool, response) -> tuple[str, str]  — (payload, encoding_id)
    decode(payload) -> dict

The registry is auto-populated by walking this package at import time.
"""

from __future__ import annotations

from . import registry  # noqa: F401
