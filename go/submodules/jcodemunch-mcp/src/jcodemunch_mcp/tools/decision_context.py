"""decision_context — surface architecture-decision context from the git record.

When an agent runs impact analysis (``get_blast_radius`` / ``get_impact_preview``)
it learns *what* breaks. This resolver adds *why the code is the way it is*: it
mines the commit history of the symbol-under-analysis and its impacted files for
**decision-bearing commits** — reverts, performance rewrites, refactors, renames,
and root-cause bugfixes — and surfaces a compact, deduped, recency-ranked digest
plus a one-line volatility read ("this surface has 3 reverts + 2 perf rewrites in
180d — review the rationale before changing").

Design — deliberately clean-room:
  * The source is the **durable git commit record**, not ephemeral agent chat
    logs. It is language- and agent-agnostic and already half-parsed: this reuses
    ``get_symbol_provenance``'s commit classifier verbatim.
  * It is **surface-only**: a read of history attached to the response, never a
    persisted decision graph and never a write to the user's tree
    (read-only charter). Theirs writes a decision graph; ours surfaces and forgets.

Pure read path. Bounded cost (capped files × capped commits per file), so it is
opt-in on the impact tools (``include_decisions``).
"""

from __future__ import annotations

import logging
from collections import Counter
from typing import Iterable, Optional

from .get_symbol_provenance import _run_git, _classify_commit, _extract_intent

logger = logging.getLogger(__name__)

# Commit categories (from get_symbol_provenance's classifier) that signal a
# recorded decision / tradeoff / consequence worth surfacing during impact
# analysis. Low-signal categories (docs/test/config/feature/evolution) are
# intentionally excluded — they describe routine change, not a decision.
_DECISION_CATEGORIES = frozenset({"revert", "perf", "refactor", "rename", "bugfix"})

# Surface order: a revert means a prior change was undone (highest signal); a
# refactor/perf is a deliberate restructuring; a rename moved an API; a bugfix
# carries root-cause history.
_CATEGORY_WEIGHT = {"revert": 5, "refactor": 4, "perf": 4, "rename": 3, "bugfix": 2}

_DELIM = "---DECISION_DELIM---"


def _git_available(cwd: str) -> bool:
    rc, _, _ = _run_git(["rev-parse", "--git-dir"], cwd=cwd)
    return rc == 0


def _decisions_for_file(
    cwd: str, rel_file: str, window_days: int, max_per_file: int
) -> list[dict]:
    """Decision-bearing commits touching *rel_file* within the window."""
    log_args = [
        "log",
        "--follow",  # track the file across renames
        "--no-merges",
        f"-n{max_per_file}",
        f"--since={window_days} days ago",
        f"--format=%H|%an|%aI|%s%n%b%n{_DELIM}",
        "--",
        rel_file,
    ]
    rc, out, _ = _run_git(log_args, cwd=cwd, timeout=20)
    if rc != 0 or not out:
        return []

    found: list[dict] = []
    for block in out.split(_DELIM):
        block = block.strip()
        if not block:
            continue
        lines = block.splitlines()
        parts = lines[0].split("|", 3)
        if len(parts) < 4:
            continue
        sha, author, date, subject = parts
        category = _classify_commit(subject)
        if category not in _DECISION_CATEGORIES:
            continue
        full_message = subject + "\n" + "\n".join(lines[1:]) if len(lines) > 1 else subject
        found.append({
            "sha": sha[:12],
            "author": author,
            "date": date[:10],
            "category": category,
            "subject": subject,
            "intent": _extract_intent(full_message),
            "file": rel_file,
        })
    return found


def _summary(
    decisions: list[dict], by_category: dict, volatility: str, window_days: int, n_files: int
) -> str:
    if not decisions:
        return (
            f"No decision-bearing commits (revert/perf/refactor/rename/bugfix) in the "
            f"last {window_days}d across {n_files} examined file(s)."
        )
    breakdown = ", ".join(
        f"{count} {cat}" for cat, count in sorted(by_category.items(), key=lambda kv: -kv[1])
    )
    head = (
        f"{len(decisions)} decision-bearing commit(s) in {window_days}d across "
        f"{n_files} file(s): {breakdown}."
    )
    tail = {
        "high": " High-churn history — review the rationale before changing.",
        "moderate": " Some change history worth reviewing before changing.",
        "low": "",
        "none": "",
    }.get(volatility, "")
    return head + tail


def resolve_decision_context(
    cwd: Optional[str],
    files: Iterable[str],
    *,
    window_days: int = 180,
    max_decisions: int = 12,
    max_files: int = 8,
    max_per_file: int = 20,
) -> dict:
    """Surface decision-bearing commits for *files* from the git record.

    Read-only; nothing is persisted. Bounded to *max_files* files (the focal
    file should be passed first) × *max_per_file* commits each.

    Returns ``{available: True, surfaced_from, window_days, files_examined,
    decision_count, by_category, volatility, summary, decisions, note}`` or an
    ``{available: False, reason, hint}`` honest-hint shape when there is no local
    git working tree (e.g. a GitHub-indexed repo).
    """
    if not cwd:
        return {
            "available": False,
            "reason": "no_local_git",
            "hint": (
                "Decision context needs a locally indexed git repo (index_folder); "
                "GitHub-indexed repos have no local working tree to read history from."
            ),
        }
    if not _git_available(cwd):
        return {
            "available": False,
            "reason": "not_a_git_repo",
            "hint": "The indexed source root is not a git working tree; no commit history to surface.",
        }

    # Preserve order, dedupe, cap. The focal file should be first in *files*.
    seen: set[str] = set()
    examined: list[str] = []
    for f in files:
        if f and f not in seen:
            seen.add(f)
            examined.append(f)
            if len(examined) >= max_files:
                break

    by_sha: dict[str, dict] = {}
    for rel in examined:
        try:
            rows = _decisions_for_file(cwd, rel, window_days, max_per_file)
        except Exception:
            logger.debug("decision scan failed for %s", rel, exc_info=True)
            rows = []
        for d in rows:
            existing = by_sha.get(d["sha"])
            if existing is None:
                d["files"] = [d.pop("file")]
                by_sha[d["sha"]] = d
            else:
                fn = d.get("file")
                if fn and fn not in existing["files"]:
                    existing["files"].append(fn)

    decisions = list(by_sha.values())
    # Rank by category weight (desc), then recency (date desc).
    decisions.sort(
        key=lambda d: (_CATEGORY_WEIGHT.get(d["category"], 0), d["date"]),
        reverse=True,
    )
    decisions = decisions[:max_decisions]

    by_category = dict(Counter(d["category"] for d in decisions))
    reverts = by_category.get("revert", 0)
    total = len(decisions)
    if reverts >= 2 or total >= 8:
        volatility = "high"
    elif total >= 3:
        volatility = "moderate"
    elif total >= 1:
        volatility = "low"
    else:
        volatility = "none"

    return {
        "available": True,
        "surfaced_from": "git_history",
        "window_days": window_days,
        "files_examined": len(examined),
        "decision_count": len(decisions),
        "by_category": by_category,
        "volatility": volatility,
        "summary": _summary(decisions, by_category, volatility, window_days, len(examined)),
        "decisions": decisions,
        "note": "Surfaced read-only from the commit record; nothing is persisted.",
    }
