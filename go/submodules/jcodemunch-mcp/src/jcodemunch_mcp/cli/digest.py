"""``jcodemunch-mcp digest`` — agent stand-up briefing CLI surface.

Calls into :func:`jcodemunch_mcp.tools.digest.compose_digest` and prints
the resulting markdown to stdout. The actual composition logic lives in
``tools/digest.py`` so the MCP tool surface and this CLI share one
implementation.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Optional


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Print a since-last-session briefing for a repo.",
    )
    parser.add_argument(
        "repo",
        nargs="?",
        default=".",
        help="Repo identifier (path, owner/name, or bare display name). "
        "Defaults to '.'  — resolves the current working directory.",
    )
    parser.add_argument(
        "--since-sha",
        default=None,
        help="Override the last-seen SHA (for re-running a delta)",
    )
    parser.add_argument(
        "--max-changed-files", type=int, default=5,
        help="Cap on changed-files list (default 5)",
    )
    parser.add_argument(
        "--max-hotspots", type=int, default=3,
        help="Cap on hotspot list (default 3)",
    )
    parser.add_argument(
        "--max-dead-code", type=int, default=3,
        help="Cap on dead-code candidates (default 3)",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Emit the structured payload as JSON instead of markdown.",
    )
    parser.add_argument(
        "--storage-path",
        default=None,
        help="Override index storage location.",
    )
    args = parser.parse_args(argv)

    # If the repo argument looks like a path, try to resolve it to a repo
    # ID via tools.resolve_repo first; that's the standard CLI ergonomics
    # ("run from inside my project, get the briefing for it").
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

    from ..tools.digest import compose_digest
    result = compose_digest(
        repo_arg,
        since_sha=args.since_sha,
        max_changed_files=args.max_changed_files,
        max_hotspots=args.max_hotspots,
        max_dead_code=args.max_dead_code,
        storage_path=args.storage_path,
    )

    if "error" in result:
        print(f"error: {result['error']}", file=sys.stderr)
        return 1

    payload = (
        json.dumps(result["structured"], indent=2)
        if args.json
        else result["briefing"]
    )
    # Some Windows consoles default to cp1252 which can't encode every
    # character from the markdown. Re-encode through stdout's reported
    # encoding with backslash-escape fallback so the briefing still
    # prints rather than blowing up the whole call.
    enc = getattr(sys.stdout, "encoding", None) or "utf-8"
    sys.stdout.write(payload.encode(enc, errors="backslashreplace").decode(enc))
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
