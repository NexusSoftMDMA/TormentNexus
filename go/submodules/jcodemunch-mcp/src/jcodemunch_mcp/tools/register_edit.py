"""Register file edits to invalidate caches."""

import time
import logging
from typing import Optional

from ..storage import IndexStore
from ._utils import index_status_to_tool_error, resolve_repo

_logger = logging.getLogger(__name__)


def register_edit(
    repo: str,
    file_paths: list[str],
    reindex: bool = False,
    storage_path: Optional[str] = None,
    _journal=None,  # For testing - inject journal
) -> dict:
    """Register file edits to invalidate caches.

    Call this after editing files to:
    1. Record the edit in the session journal
    2. Clear BM25 cache for the repo
    3. Invalidate search result cache for the repo

    Args:
        repo: Repository identifier.
        file_paths: List of file paths that were edited.
        reindex: If True, also reindex the files (calls index_file for each).
        storage_path: Custom storage path.

    Returns:
        Dict with registered count, invalidated_symbols, bm25_cache_cleared, and _meta.
    """
    start = time.perf_counter()

    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return {"error": str(e)}

    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name)
    if not index:
        return index_status_to_tool_error(store.inspect_index(owner, name))

    # Get journal for recording
    if _journal is None:
        from .session_journal import get_journal
        journal = get_journal()
    else:
        journal = _journal

    # Record edits in journal
    for fp in file_paths:
        journal.record_edit(fp)

    # Invalidate symbol tokens for edited files
    edited_set = set(file_paths)
    invalidated_symbols = 0
    for sym in index.symbols:
        if sym.get("file") in edited_set:
            # Clear cached tokens
            sym.pop("_tokens", None)
            sym.pop("_tf", None)
            sym.pop("_dl", None)
            invalidated_symbols += 1

    # Clear BM25 cache
    index._bm25_cache.clear()
    bm25_cleared = True

    # Clear import name index
    index._import_name_index = None

    # Invalidate search result cache for this repo
    from .search_symbols import result_cache_invalidate_repo
    result_cache_invalidate_repo(f"{owner}/{name}")

    # Optionally reindex
    if reindex:
        from .index_file import index_file
        for fp in file_paths:
            try:
                index_file(path=fp, storage_path=storage_path)
            except Exception as e:
                _logger.debug("Failed to reindex %s: %s", fp, e)

    elapsed = (time.perf_counter() - start) * 1000

    return {
        "registered": len(file_paths),
        "invalidated_symbols": invalidated_symbols,
        "bm25_cache_cleared": bm25_cleared,
        "_meta": {
            "timing_ms": round(elapsed, 1),
        },
    }
