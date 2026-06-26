"""Encoding A/B benchmark — measure MUNCH compact output vs JSON.

Runs a set of representative response fixtures through the encoding dispatcher
with format="compact" and reports JSON bytes, compact bytes, % savings, and an
estimated token savings (bytes/4 heuristic). Used to validate the PRD's
≥30% median / ≥50% graph-tool targets.

Usage:
    python -m munch_bench.encoding_bench
    python -m munch_bench.encoding_bench --json  # emit machine-readable JSON
"""

from __future__ import annotations

import argparse
import json
import statistics
from typing import Any

from jcodemunch_mcp.encoding import encode_response


def _fixture_find_references(n: int = 20) -> dict:
    return {
        "repo": "acme/app",
        "identifier": "get_user",
        "reference_count": n,
        "references": [
            {
                "file": f"src/service/handlers/file_{i:03d}.py",
                "line": i + 1,
                "column": (i % 8) * 4,
                "specifier": "models.user",
                "kind": "import",
            }
            for i in range(n)
        ],
        "_meta": {"timing_ms": 3.1, "truncated": False},
    }


def _fixture_dependency_graph(n: int = 24) -> dict:
    return {
        "repo": "acme/app",
        "file": "src/main.py",
        "direction": "both",
        "depth": 3,
        "depth_reached": 3,
        "node_count": n + 1,
        "edge_count": n,
        "edges": [
            {
                "from": "src/main.py",
                "to": f"src/lib/module_{i:03d}.py",
                "depth": 1 + (i % 3),
            }
            for i in range(n)
        ],
        "cross_repo_edges": [],
        "_meta": {"timing_ms": 2.1, "truncated": False, "cross_repo": False},
    }


def _fixture_call_hierarchy(n: int = 16) -> dict:
    return {
        "repo": "acme/app",
        "symbol": {"id": "sym1", "name": "foo", "kind": "function", "file": "x.py", "line": 1},
        "direction": "both",
        "depth": 3,
        "depth_reached": 3,
        "caller_count": n,
        "callee_count": n,
        "callers": [
            {
                "id": f"c{i}",
                "name": f"caller_{i}",
                "kind": "function",
                "file": f"src/callers/file_{i:03d}.py",
                "line": 10 + i,
                "depth": 1 + (i % 3),
                "resolution": "ast",
            }
            for i in range(n)
        ],
        "callees": [
            {
                "id": f"e{i}",
                "name": f"callee_{i}",
                "kind": "function",
                "file": f"src/callees/file_{i:03d}.py",
                "line": 20 + i,
                "depth": 1 + (i % 3),
                "resolution": "lsp",
            }
            for i in range(n)
        ],
        "dispatches": [],
        "_meta": {"timing_ms": 4.0, "methodology": "ast+lsp"},
    }


def _fixture_search_symbols(n: int = 25) -> dict:
    return {
        "result_count": n,
        "query": "user",
        "results": [
            {
                "id": f"s{i}",
                "name": f"handler_{i}",
                "kind": "function",
                "file": f"src/models/user_{i:03d}.py",
                "line": i + 1,
                "score": round(0.99 - i * 0.01, 3),
                "signature": f"def handler_{i}(request)",
                "summary": "Handles a user request",
            }
            for i in range(n)
        ],
        "_meta": {"timing_ms": 1.3, "total_symbols": 1200, "truncated": False},
    }


def _fixture_get_repo_outline(n: int = 40) -> dict:
    return {
        "repo": "acme/app",
        "source_root": "/tmp/app",
        "file_count": n,
        "symbol_count": n * 5,
        "files": [
            {
                "file": f"src/module/feature_{i:03d}.py",
                "language": "python",
                "symbol_count": 5,
                "line_count": 40 + i,
                "summary": f"feature {i}",
            }
            for i in range(n)
        ],
        "_meta": {"timing_ms": 2.0, "is_stale": False},
    }


def _fixture_generic_hotspots(n: int = 20) -> dict:
    return {
        "repo": "acme/app",
        "window_days": 30,
        "hotspots": [
            {
                "name": f"handler_{i}",
                "file": f"src/service/hot_{i:03d}.py",
                "complexity": 5 + (i % 10),
                "churn": 3 + (i % 7),
                "risk_score": round((5 + i) * 0.17, 3),
            }
            for i in range(n)
        ],
    }


_FIXTURES = [
    ("find_references", _fixture_find_references()),
    ("get_dependency_graph", _fixture_dependency_graph()),
    ("get_call_hierarchy", _fixture_call_hierarchy()),
    ("search_symbols", _fixture_search_symbols()),
    ("get_repo_outline", _fixture_get_repo_outline()),
    ("get_hotspots (generic)", _fixture_generic_hotspots()),
]


def _run() -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for tool, resp in _FIXTURES:
        base_tool = tool.split(" ", 1)[0]
        payload, meta = encode_response(base_tool, resp, "compact")
        json_bytes = meta["json_bytes"]
        compact_bytes = meta["encoded_bytes"]
        savings = 1.0 - (compact_bytes / json_bytes)
        rows.append({
            "tool": tool,
            "encoding": meta["encoding"],
            "json_bytes": json_bytes,
            "compact_bytes": compact_bytes,
            "savings_pct": round(savings * 100, 1),
            "tokens_saved_est": (json_bytes - compact_bytes) // 4,
        })
    return rows


def _print_table(rows: list[dict[str, Any]]) -> None:
    cols = ["tool", "encoding", "json_bytes", "compact_bytes", "savings_pct", "tokens_saved_est"]
    widths = {c: max(len(c), max(len(str(r[c])) for r in rows)) for c in cols}
    header = "  ".join(c.ljust(widths[c]) for c in cols)
    print(header)
    print("  ".join("-" * widths[c] for c in cols))
    for r in rows:
        print("  ".join(str(r[c]).ljust(widths[c]) for c in cols))
    savings = [r["savings_pct"] for r in rows]
    print()
    print(f"median savings: {statistics.median(savings):.1f}%")
    print(f"mean savings:   {statistics.mean(savings):.1f}%")
    print(f"min / max:      {min(savings):.1f}% / {max(savings):.1f}%")


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--json", action="store_true", help="emit JSON instead of a table")
    args = ap.parse_args()
    rows = _run()
    if args.json:
        print(json.dumps(rows, indent=2))
    else:
        _print_table(rows)


if __name__ == "__main__":
    main()
