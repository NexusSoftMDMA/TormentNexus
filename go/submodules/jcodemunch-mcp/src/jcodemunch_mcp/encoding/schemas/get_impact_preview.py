"""Compact encoder for get_impact_preview."""

from .. import schema_driven as sd

TOOLS = ("get_impact_preview",)
ENCODING_ID = "ip1"

_TABLES = [
    sd.TableSpec(
        key="affected_symbols",
        tag="s",
        cols=["id", "name", "kind", "file", "line", "depth"],
        intern=["file", "id"],
        types={"line": "int", "depth": "int"},
    ),
]
_SCALARS = ("repo", "affected_files", "affected_symbol_count")
_NESTED = {"symbol": ["id", "name", "kind", "file", "line"]}
_META = ("timing_ms",)
_JSON = ("affected_by_file", "call_chains")


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
