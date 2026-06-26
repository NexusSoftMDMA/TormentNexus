"""Compact encoder for search_ast."""

from .. import schema_driven as sd

TOOLS = ("search_ast",)
ENCODING_ID = "sa1"

_TABLES = [
    sd.TableSpec(
        key="results",
        tag="a",
        cols=["file", "line", "match_type", "snippet", "symbol_id", "symbol_name"],
        intern=["file", "symbol_id"],
        types={"line": "int"},
    ),
]
_SCALARS = ("result_count", "query", "repo", "category", "pattern")
_META = (
    "timing_ms", "files_searched", "truncated",
    "tokens_saved", "total_tokens_saved",
)


def encode(tool: str, response: dict) -> tuple[str, str]:
    return sd.encode(tool, response, ENCODING_ID, _TABLES, _SCALARS, meta_keys=_META)


def decode(payload: str) -> dict:
    return sd.decode(payload, _TABLES, _SCALARS, meta_keys=_META)
