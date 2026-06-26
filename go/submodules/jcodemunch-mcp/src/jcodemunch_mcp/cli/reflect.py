"""`jcodemunch-mcp reflect` — surface retrieval regret as suggested corrections.

Thin CLI over ``tools.suggest_corrections``: mines the ranking ledger for the
given repo and prints a prioritized, explainable set of SUGGESTED config
corrections (routing/vocabulary/freshness/stale-config) plus a dry-run
ranking-weight proposal. Read-only by charter — never writes a user file unless
you pass ``--apply-weights`` (which touches only the tuning.jsonc sidecar).
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Optional


def _resolve_repo_arg(repo_arg: str, storage_path: Optional[str]) -> str:
    """Path-like argument -> indexed repo id (digest/receipt ergonomics)."""
    if repo_arg in (".", "..") or "/" in repo_arg or "\\" in repo_arg:
        path = Path(repo_arg).resolve()
        if path.exists():
            try:
                from ..tools.resolve_repo import resolve_repo as _resolve
                resolved = _resolve(str(path), storage_path)
                if resolved.get("indexed") and resolved.get("repo"):
                    return resolved["repo"]
            except Exception as e:  # noqa: BLE001
                print(f"error resolving '{repo_arg}': {e}", file=sys.stderr)
    return repo_arg


_SEV_MARK = {"high": "!!", "error": "!!", "medium": "!", "warning": "!", "low": "-", "info": "-"}


def _render_human(out: dict) -> str:
    lines: list[str] = []
    repo = out.get("repo")
    lines.append(f"# Retrieval reflection — {repo}")
    lines.append("")
    if not out.get("telemetry_present"):
        lines.append(out.get("hint") or "Telemetry is off; nothing to reflect on.")
        return "\n".join(lines)

    lines.append(
        f"Analyzed {out.get('events_analyzed', 0)} ranking events"
        + (f" over the last {out['window_days']} days." if out.get("window_days") else " (all time).")
    )
    lines.append("")

    corrections = out.get("corrections") or []
    if not corrections:
        lines.append("No retrieval regret above threshold. Nothing to suggest.")
    else:
        lines.append(f"## {len(corrections)} suggested correction(s)")
        lines.append("")
        for i, c in enumerate(corrections, 1):
            mark = _SEV_MARK.get(c.get("severity"), "-")
            lines.append(f"### {i}. [{mark} {c.get('severity')}] {c.get('kind')}")
            lines.append(f"  cause: {c.get('cause')}")
            lines.append(f"  fix:   {c.get('recommended_action')}")
            patch = c.get("suggested_patch")
            if patch:
                lines.append("  suggested patch (apply yourself):")
                lines.extend("    " + ln for ln in patch.splitlines())
            lines.append("")

    wp = out.get("weight_proposal")
    if isinstance(wp, dict) and wp.get("before") is not None:
        applied = wp.get("applied")
        verb = "applied" if applied else "proposed (dry-run)"
        lines.append(f"## Ranking-weight {verb}")
        lines.append(f"  before: {wp.get('before')}")
        lines.append(f"  after:  {wp.get('after')}")
        if wp.get("reason"):
            lines.append(f"  note:   {wp.get('reason')}")
        lines.append("")

    lines.append("(Charter: suggestions only — no file was written.)")
    return "\n".join(lines)


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Surface retrieval regret as suggested config corrections.",
    )
    parser.add_argument("repo", nargs="?", default=".",
        help="Repo identifier (path, owner/name, or bare display name). Defaults to '.' (cwd).")
    parser.add_argument("--project-path", default=None,
        help="Directory holding the config files to target. Defaults to cwd.")
    parser.add_argument("--window-days", type=int, default=30,
        help="Rolling ledger window to mine (default 30).")
    parser.add_argument("--all", dest="all_time", action="store_true",
        help="Analyze the full ledger, ignoring the window.")
    parser.add_argument("--apply-weights", action="store_true",
        help="Persist the ranking-weight proposal to tuning.jsonc (sidecar, not user source).")
    parser.add_argument("--json", action="store_true",
        help="Emit the structured payload as JSON instead of the human report.")
    parser.add_argument("--storage-path", default=None,
        help="Override index storage location.")
    args = parser.parse_args(argv)

    repo = _resolve_repo_arg(args.repo, args.storage_path)

    from ..tools.suggest_corrections import suggest_corrections
    out = suggest_corrections(
        repo=repo,
        project_path=args.project_path,
        storage_path=args.storage_path,
        window_days=args.window_days,
        all_time=args.all_time,
        apply_weights=args.apply_weights,
    )

    if args.json:
        print(json.dumps(out, indent=2, default=str))
    else:
        # Some Windows consoles default to cp1252 and can't encode every char
        # (em-dashes etc.); re-encode through stdout's encoding with a
        # backslash fallback so the report prints rather than crashing.
        text = _render_human(out)
        enc = getattr(sys.stdout, "encoding", None) or "utf-8"
        sys.stdout.write(text.encode(enc, errors="backslashreplace").decode(enc))
        sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
