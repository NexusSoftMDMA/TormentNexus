"""`jcodemunch-mcp delivery` — durable-change delivery over a window.

Thin CLI over ``tools.get_delivery_metrics``: reports how many non-merge commits
in the window landed and stuck (durable) vs were reverted or re-touched
(churn-back), plus a category breakdown. Pass ``--cost`` (AI spend over the same
window) to print the headline cost-per-durable-change — how much got done for how
little, instead of rewarding raw activity. Read-only git archaeology.
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


def _render_human(out: dict, cost: Optional[float]) -> str:
    if out.get("error"):
        return f"error: {out['error']}"

    lines: list[str] = []
    repo = out.get("repo")
    wd = out.get("window_days")
    lines.append(f"# Delivery — {repo} (last {wd}d)")
    lines.append("")
    lines.append(out.get("assessment", ""))
    lines.append("")
    lines.append(f"  durable          {out.get('commits_durable', 0)}")
    lines.append(f"  reworked         {out.get('commits_reworked', 0)}  (churn-back within "
                 f"{out.get('rework_horizon_days')}d)")
    lines.append(f"  reverted         {out.get('commits_reverted', 0)}")
    lines.append(f"  revert-authored  {out.get('commits_revert_authored', 0)}")
    lines.append(f"  total commits    {out.get('commits_total', 0)}")
    lines.append(f"  durable rate     {int(out.get('durable_rate', 0) * 100)}%"
                 f"   rework rate {int(out.get('rework_rate', 0) * 100)}%")
    prov = out.get("commits_provisional", 0)
    if prov:
        lines.append(f"  provisional      {prov} (too recent to be final)")
    bycat = out.get("by_category") or {}
    if bycat:
        lines.append("")
        lines.append("  durable by kind: " + ", ".join(f"{k} {v}" for k, v in bycat.items()))

    if cost is not None:
        durable = out.get("commits_durable", 0)
        lines.append("")
        if durable > 0:
            lines.append(f"  ${cost:.2f} of AI spend / {durable} durable change(s) "
                         f"= ${cost / durable:.2f} per durable change")
        else:
            lines.append(f"  ${cost:.2f} of AI spend / 0 durable changes "
                         f"(nothing settled to attribute spend to yet)")
    return "\n".join(lines)


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Report durable-change delivery (and optional cost-per-outcome) over a window.",
    )
    parser.add_argument("repo", nargs="?", default=".",
        help="Repo identifier (path, owner/name, or bare display name). Defaults to '.' (cwd).")
    parser.add_argument("--window-days", type=int, default=30,
        help="Look-back window in days (default 30).")
    parser.add_argument("--rework-horizon-days", type=int, default=14,
        help="Days within which a re-touch counts as churn-back (default 14).")
    parser.add_argument("--cost", type=float, default=None,
        help="AI spend (dollars) over the same window; prints cost-per-durable-change.")
    parser.add_argument("--json", action="store_true",
        help="Emit the structured payload as JSON instead of the human report.")
    parser.add_argument("--storage-path", default=None,
        help="Override index storage location.")
    args = parser.parse_args(argv)

    repo = _resolve_repo_arg(args.repo, args.storage_path)

    from ..tools.get_delivery_metrics import get_delivery_metrics
    out = get_delivery_metrics(
        repo=repo,
        window_days=args.window_days,
        rework_horizon_days=args.rework_horizon_days,
        storage_path=args.storage_path,
    )

    if args.cost is not None and not out.get("error"):
        durable = out.get("commits_durable", 0)
        out["cost_input"] = args.cost
        out["cost_per_durable_change"] = (
            round(args.cost / durable, 4) if durable > 0 else None
        )

    if args.json:
        print(json.dumps(out, indent=2, default=str))
    else:
        text = _render_human(out, args.cost)
        enc = getattr(sys.stdout, "encoding", None) or "utf-8"
        sys.stdout.write(text.encode(enc, errors="backslashreplace").decode(enc))
        sys.stdout.write("\n")
    return 0 if not out.get("error") else 1


if __name__ == "__main__":
    sys.exit(main())
