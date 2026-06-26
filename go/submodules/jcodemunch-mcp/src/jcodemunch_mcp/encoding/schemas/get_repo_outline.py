"""Compact encoder for get_repo_outline — biggest savings target (repo topology)."""

from .. import schema_driven as sd

TOOLS = ("get_repo_outline",)
ENCODING_ID = "ro1"

_TABLES = [
    sd.TableSpec(
        key="files",
        tag="f",
        cols=["file", "language", "symbol_count", "line_count", "summary"],
        intern=["file"],
        types={"symbol_count": "int", "line_count": "int"},
    ),
    sd.TableSpec(
        key="directories",
        tag="d",
        cols=["path", "file_count", "summary"],
        intern=["path"],
        types={"file_count": "int"},
    ),
]
_SCALARS = (
    "repo", "source_root", "language", "file_count", "symbol_count",
    "indexed_at", "git_head", "display_name", "staleness_warning",
)
_META = (
    "timing_ms", "tokens_saved", "total_tokens_saved", "is_stale",
)
_JSON = ("languages", "tree", "stats")


def encode(tool: str, response: dict) -> tuple[str, str]:
    return sd.encode(
        tool, response, ENCODING_ID, _TABLES, _SCALARS,
        meta_keys=_META, json_blobs=_JSON,
    )


def decode(payload: str) -> dict:
    return sd.decode(
        payload, _TABLES, _SCALARS, meta_keys=_META, json_blobs=_JSON,
    )
