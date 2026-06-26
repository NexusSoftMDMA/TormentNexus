"""Compact encoder for get_dependency_graph — adjacency list as CSV."""

from .. import schema_driven as sd

TOOLS = ("get_dependency_graph",)
ENCODING_ID = "dg1"

_TABLES = [
    sd.TableSpec(
        key="edges",
        tag="e",
        cols=["from", "to", "depth"],
        intern=["from", "to"],
        types={"depth": "int"},
    ),
    sd.TableSpec(
        key="cross_repo_edges",
        tag="x",
        cols=["from_file", "to_repo", "specifier", "depth"],
        intern=["from_file", "specifier"],
        types={"depth": "int"},
    ),
]
_SCALARS = (
    "repo", "file", "direction", "depth", "depth_reached",
    "node_count", "edge_count",
)
_META = ("timing_ms", "truncated", "cross_repo", "tokens_saved", "total_tokens_saved")
_JSON = ("nodes", "neighbors")


def encode(tool: str, response: dict) -> tuple[str, str]:
    return sd.encode(
        tool, response, ENCODING_ID, _TABLES, _SCALARS,
        meta_keys=_META, json_blobs=_JSON,
    )


def decode(payload: str) -> dict:
    return sd.decode(
        payload, _TABLES, _SCALARS, meta_keys=_META, json_blobs=_JSON,
    )
