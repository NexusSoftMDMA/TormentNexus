"""List indexed repositories."""

import time
from typing import Optional

from ..storage import IndexStore


def list_repos(storage_path: Optional[str] = None) -> dict:
    """List all indexed repositories.

    Returns:
        Dict with count, list of repos, and _meta envelope.
    """
    start = time.perf_counter()
    store = IndexStore(base_path=storage_path)
    repos = store.list_repos()
    elapsed = (time.perf_counter() - start) * 1000

    return {
        "count": len(repos),
        "repos": repos,
        "_meta": {
            "timing_ms": round(elapsed, 1),
        },
    }


def repos_report(storage_path: Optional[str] = None) -> list[dict]:
    """Cockpit view of indexed repos: per-repo counts + freshness + watcher state.

    Joins `list_repos` metadata (counts, languages, indexed_at) with
    `get_watch_status` (staleness + watcher lock holder), keyed by source_root.
    Structured for the jMunch Console index/watcher cockpit, but general-purpose.
    Watch status only covers discovered repos, so a repo it doesn't cover
    defaults to fresh/idle (no staleness signal available).
    """
    store = IndexStore(base_path=storage_path)
    repos = store.list_repos()
    try:
        from .get_watch_status import get_watch_status
        ws = get_watch_status(storage_path)
        ws_by_root = {r.get("source_root"): r for r in ws.get("repos", [])}
    except Exception:
        ws_by_root = {}

    report: list[dict] = []
    for r in repos:
        w = ws_by_root.get(r.get("source_root", ""), {})
        if w.get("reindex_in_progress"):
            watcher_state = "reindexing"
        elif w.get("watcher_holder"):
            watcher_state = "watching"
        else:
            watcher_state = "idle"
        holder = w.get("watcher_holder") or {}
        report.append({
            "repo_id": r.get("repo", ""),
            "display_name": r.get("display_name") or r.get("repo", ""),
            "source_root": r.get("source_root", ""),
            "file_count": r.get("file_count", 0),
            "symbol_count": r.get("symbol_count", 0),
            "languages": r.get("languages", {}) or {},
            "indexed_at": r.get("indexed_at", ""),
            "freshness": "stale_index" if w.get("index_stale") else "fresh",
            "watcher_state": watcher_state,
            "lock_holder": holder.get("client_id"),
        })
    return report
