"""Compact encoder for get_ranked_context."""

from .. import schema_driven as sd

TOOLS = ("get_ranked_context",)
ENCODING_ID = "rc1"

_TABLES = [
    sd.TableSpec(
        key="context_items",
        tag="i",
        cols=["id", "name", "kind", "file", "line", "score", "token_cost", "summary"],
        intern=["file", "id"],
        types={"line": "int", "score": "float", "token_cost": "int"},
    ),
]
_SCALARS = (
    "total_tokens", "budget_tokens", "items_included", "items_considered",
    "query", "repo",
)
_META = (
    "timing_ms", "tokens_saved", "total_tokens_saved", "fusion",
)
_JSON = ("channels",)


def encode(tool: str, response: dict) -> tuple[str, str]:
    return sd.encode(
        tool, response, ENCODING_ID, _TABLES, _SCALARS,
        meta_keys=_META, json_blobs=_JSON,
    )


def decode(payload: str) -> dict:
    return sd.decode(
        payload, _TABLES, _SCALARS, meta_keys=_META, json_blobs=_JSON,
    )
