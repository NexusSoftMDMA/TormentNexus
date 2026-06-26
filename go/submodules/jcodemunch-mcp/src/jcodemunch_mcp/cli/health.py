"""``jcodemunch-mcp health`` — print get_repo_health JSON to stdout.

Thin CLI surface over the existing ``get_repo_health`` MCP tool, so CI
pipelines (notably the v1.88.0 health-radar GitHub Action) can extract
the six-axis radar without writing a Python wrapper.

Output is the full ``get_repo_health`` response (including the
``radar`` sub-field) as JSON to stdout. Errors go to stderr; non-zero
exit on resolve/index failures.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Optional


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Print get_repo_health JSON for a repo.",
    )
    parser.add_argument(
        "repo",
        nargs="?",
        default=".",
        help="Repo identifier (path, owner/name, or bare display name). "
        "Defaults to '.' (resolves cwd).",
    )
    parser.add_argument(
        "--days", type=int, default=90,
        help="Churn look-back window in days (default 90).",
    )
    parser.add_argument(
        "--radar-only",
        action="store_true",
        help="Emit only the `radar` sub-field instead of the full health response.",
    )
    parser.add_argument(
        "--storage-path", default=None,
        help="Override index storage location.",
    )
    args = parser.parse_args(argv)

    # Resolve path-style repo args via tools.resolve_repo (same UX as `digest`).
    repo_arg = args.repo
    if repo_arg in (".", "..") or "/" in repo_arg or "\\" in repo_arg:
        path = Path(repo_arg).resolve()
        if path.exists():
            try:
                from ..tools.resolve_repo import resolve_repo as _resolve
                resolved = _resolve(str(path), storage_path=args.storage_path)
                if resolved.get("indexed") and resolved.get("repo"):
                    repo_arg = resolved["repo"]
                else:
                    print(
                        f"error: path '{path}' is not indexed. "
                        f"Run `jcodemunch-mcp index '{path}'` first.",
                        file=sys.stderr,
                    )
                    return 2
            except Exception as e:
                print(f"error resolving '{repo_arg}': {e}", file=sys.stderr)
                return 2

    from ..tools.get_repo_health import get_repo_health
    result = get_repo_health(
        repo=repo_arg,
        days=args.days,
        storage_path=args.storage_path,
    )

    if "error" in result:
        print(f"error: {result['error']}", file=sys.stderr)
        return 1

    payload = result["radar"] if args.radar_only else result
    sys.stdout.write(json.dumps(payload, indent=2))
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
