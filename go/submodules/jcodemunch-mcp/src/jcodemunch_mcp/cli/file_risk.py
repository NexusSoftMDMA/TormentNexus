"""``jcodemunch-mcp file-risk`` — print per-symbol risk JSON for a file.

Thin CLI surface used by the v1.89.0 VS Code risk-density gutter
(``vscode-extension/src/riskGutter.ts``). Output is the full
``get_file_risk`` JSON to stdout; errors go to stderr with non-zero exit.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Optional


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Print per-symbol risk for a file as JSON.",
    )
    parser.add_argument(
        "file",
        help="Path to the file within an indexed repo (absolute or repo-relative).",
    )
    parser.add_argument(
        "--repo", default=None,
        help="Repo identifier. Defaults to autodetecting from the file's path.",
    )
    parser.add_argument(
        "--storage-path", default=None,
        help="Override index storage location.",
    )
    args = parser.parse_args(argv)

    file_arg = args.file
    repo_arg = args.repo

    # Auto-resolve the repo from the file's parent if not given.
    if repo_arg is None:
        from ..tools.resolve_repo import resolve_repo as _resolve
        path = Path(file_arg).resolve()
        if not path.exists():
            print(f"error: file not found: {file_arg}", file=sys.stderr)
            return 2
        # Walk up to find an indexed root. Cheapest path: ask resolve_repo
        # for the file's directory and let it climb to a known repo root.
        try:
            resolved = _resolve(str(path.parent), storage_path=args.storage_path)
        except Exception as e:
            print(f"error resolving repo for {file_arg}: {e}", file=sys.stderr)
            return 2
        if not resolved.get("indexed") or not resolved.get("repo"):
            print(
                f"error: file '{file_arg}' is not in an indexed repo. "
                f"Run `jcodemunch-mcp index <repo-root>` first.",
                file=sys.stderr,
            )
            return 2
        repo_arg = resolved["repo"]
        # Convert absolute file path to repo-relative for the lookup.
        source_root = resolved.get("source_root")
        if source_root:
            try:
                file_arg = str(path.relative_to(Path(source_root))).replace("\\", "/")
            except ValueError:
                # File isn't under the source_root the index recorded —
                # pass through as-is and let get_file_risk error.
                pass

    from ..tools.get_file_risk import get_file_risk
    result = get_file_risk(
        repo=repo_arg,
        file_path=file_arg,
        storage_path=args.storage_path,
    )

    if "error" in result:
        print(f"error: {result['error']}", file=sys.stderr)
        return 1

    sys.stdout.write(json.dumps(result, indent=2))
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
