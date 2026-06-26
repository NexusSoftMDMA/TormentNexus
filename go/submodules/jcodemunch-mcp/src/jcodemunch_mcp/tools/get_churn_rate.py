"""get_churn_rate — count git commits touching a file or symbol over a time window.

Uses ``git log`` against the repo's local working tree (requires the repo to have been
indexed with ``index_folder``; GitHub-indexed repos have no local source root).

When a symbol_id is provided the file path is resolved from the index; churn is measured
at file level (git does not track line ranges across renames, so file-level churn is the
most reliable approximation).
"""

from __future__ import annotations

import logging
import subprocess
import time
from typing import Optional

from ..storage import IndexStore
from ._utils import resolve_repo

logger = logging.getLogger(__name__)


def _run_git(args: list[str], cwd: str, timeout: int = 15) -> tuple[int, str, str]:
    try:
        r = subprocess.run(
            ["git"] + args,
            cwd=cwd, capture_output=True, text=True,
            timeout=timeout, stdin=subprocess.DEVNULL,
        )
        return r.returncode, r.stdout.strip(), r.stderr.strip()
    except FileNotFoundError:
        return -1, "", "git not found on PATH"
    except subprocess.TimeoutExpired:
        return -2, "", "git command timed out"
    except Exception as exc:
        logger.debug("git subprocess error: %s", exc, exc_info=True)
        return -3, "", str(exc)


def get_churn_rate(
    repo: str,
    target: str,
    days: int = 90,
    storage_path: Optional[str] = None,
) -> dict:
    """Return git churn metrics for a file or symbol.

    *target* is either a relative file path within the repo (e.g. ``src/utils.py``) or
    a full symbol ID as returned by ``search_symbols``.  When a symbol ID is given the
    file path is resolved from the index and churn is measured at file level.

    Args:
        repo:         Repository identifier (owner/repo or bare name).
        target:       Relative file path **or** symbol ID.
        days:         Look-back window in days (default 90).
        storage_path: Optional index storage path override.

    Returns:
        ``{repo, target, target_type, file, commits, authors, first_seen,
           last_modified, churn_per_week, assessment, _meta}``

        ``assessment`` is ``"stable"`` (≤1 commit/week), ``"active"`` (≤3/week),
        or ``"volatile"`` (>3/week).
    """
    t0 = time.perf_counter()
    days = max(1, days)

    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return {"error": str(e)}

    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name)
    if index is None:
        return {"error": f"No index found for {repo!r}. Run index_folder first."}

    if not index.source_root:
        return {
            "error": (
                "get_churn_rate requires a locally indexed repo (index_folder). "
                "GitHub-indexed repos (index_repo) do not have a local git working tree."
            )
        }
    cwd = index.source_root

    # Determine file path from target
    target_type = "file"
    file_path = target
    sym_name = None

    # If target looks like a symbol ID (contains "::" or is found in index)
    sym = next((s for s in index.symbols if s.get("id") == target), None)
    if sym is not None:
        file_path = sym.get("file", "")
        sym_name = sym.get("name", "")
        target_type = "symbol"
        if not file_path:
            return {"error": f"Symbol {target!r} has no file in index."}

    # Verify git availability
    rc, _, err = _run_git(["rev-parse", "--git-dir"], cwd=cwd)
    if rc != 0:
        if rc == -1:
            return {"error": "git not found on PATH."}
        return {"error": f"Not a git repository: {err}"}

    # Fetch commits in the window: format hash|author_email|ISO date
    rc2, log_out, log_err = _run_git(
        [
            "log",
            "--follow",
            f"--since={days} days ago",
            "--format=%H|%ae|%aI",
            "--",
            file_path,
        ],
        cwd=cwd,
        timeout=30,
    )
    if rc2 not in (0, 128):  # 128 = no commits but still OK
        return {"error": f"git log failed: {log_err}"}

    commits_raw = [line for line in log_out.splitlines() if line.strip()] if log_out else []
    commit_count = len(commits_raw)

    authors: list[str] = sorted(
        {parts[1] for line in commits_raw if len((parts := line.split("|"))) >= 2}
    )
    dates = [
        parts[2]
        for line in commits_raw
        if len((parts := line.split("|"))) >= 3 and parts[2]
    ]
    last_modified = dates[0] if dates else None   # git log is newest-first

    # First-ever commit (beyond the window)
    rc3, first_out, _ = _run_git(
        ["log", "--follow", "--diff-filter=A", "--format=%aI", "--", file_path],
        cwd=cwd,
        timeout=30,
    )
    first_seen: Optional[str] = None
    if rc3 == 0 and first_out:
        first_seen = first_out.splitlines()[-1].strip() or None  # oldest last

    churn_per_week = round(commit_count / (days / 7), 2) if days > 0 else 0.0
    if churn_per_week <= 1.0:
        assessment = "stable"
    elif churn_per_week <= 3.0:
        assessment = "active"
    else:
        assessment = "volatile"

    result: dict = {
        "repo": f"{owner}/{name}",
        "target": target,
        "target_type": target_type,
        "file": file_path,
        "commits": commit_count,
        "authors": authors,
        "first_seen": first_seen,
        "last_modified": last_modified,
        "days": days,
        "churn_per_week": churn_per_week,
        "assessment": assessment,
        "_meta": {
            "timing_ms": round((time.perf_counter() - t0) * 1000, 1),
            "methodology": "git_log",
            "confidence_level": "high",
        },
    }
    if sym_name:
        result["symbol_name"] = sym_name
    return result
