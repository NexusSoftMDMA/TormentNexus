"""Compact encoder for get_call_hierarchy."""

from .. import schema_driven as sd

TOOLS = ("get_call_hierarchy",)
ENCODING_ID = "ch1"

_TABLES = [
    sd.TableSpec(
        key="callers",
        tag="c",
        cols=["id", "name", "kind", "file", "line", "depth", "resolution"],
        intern=["file", "id"],
        types={"line": "int", "depth": "int"},
    ),
    sd.TableSpec(
        key="callees",
        tag="e",
        cols=["id", "name", "kind", "file", "line", "depth", "resolution"],
        intern=["file", "id"],
        types={"line": "int", "depth": "int"},
    ),
    sd.TableSpec(
        key="dispatches",
        tag="d",
        cols=["id", "name", "kind", "file", "line", "depth", "resolution"],
        intern=["file", "id"],
        types={"line": "int", "depth": "int"},
    ),
]
_SCALARS = ("repo", "direction", "depth", "depth_reached", "caller_count", "callee_count")
_NESTED = {"symbol": ["id", "name", "kind", "file", "line"]}
_META = ("timing_ms", "methodology", "confidence_level", "source", "tip")
_JSON = ("resolution_tiers",)


def encode(tool: str, response: dict) -> tuple[str, str]:
    return sd.encode(
        tool, response, ENCODING_ID, _TABLES, _SCALARS,
        nested_dicts=_NESTED, meta_keys=_META, json_blobs=_JSON,
    )


def decode(payload: str) -> dict:
    return sd.decode(
        payload, _TABLES, _SCALARS,
        nested_dicts=_NESTED, meta_keys=_META, json_blobs=_JSON,
    )
