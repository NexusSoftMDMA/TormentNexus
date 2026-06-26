"""Shared helpers for tool modules."""

import logging
import threading
from typing import Optional

from ..storage import IndexStore

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Bare-name resolution cache (P5)
# ---------------------------------------------------------------------------
# Keyed by storage base_path string.
# Value: (dir_mtime: float, mapping: dict[bare_name -> sorted list of owner/name])
# Invalidated whenever the base_path directory mtime changes (repo added/removed).
# ---------------------------------------------------------------------------
_bare_name_cache: dict[str, tuple[float, dict[str, list[str]]]] = {}
_BARE_NAME_LOCK = threading.Lock()


def _get_bare_name_map(store: IndexStore) -> dict[str, list[str]]:
    """Return a cached bare-name → [owner/name] mapping for the store's base_path.

    Rebuilds when the directory mtime changes (repo indexed or cache invalidated).
    Cost when warm: one stat() call instead of N db reads.
    """
    path_str = str(store.base_path)
    try:
        mtime = store.base_path.stat().st_mtime
    except OSError:
        mtime = 0.0

    with _BARE_NAME_LOCK:
        cached = _bare_name_cache.get(path_str)
        if cached and cached[0] == mtime:
            return cached[1]

    # Miss: rebuild without holding the lock (list_repos does I/O)
    mapping: dict[str, list[str]] = {}
    for repo_entry in store.list_repos():
        owner_name = repo_entry["repo"]
        if not owner_name or "/" not in owner_name:
            continue
        _, repo_name = owner_name.split("/", 1)
        for key in (repo_name, repo_entry.get("display_name")):
            if key:
                mapping.setdefault(key, []).append(owner_name)

    # Deduplicate and sort so output is deterministic
    mapping = {k: sorted(set(v)) for k, v in mapping.items()}
    with _BARE_NAME_LOCK:
        _bare_name_cache[path_str] = (mtime, mapping)
    return mapping


def resolve_repo(repo: str, storage_path: Optional[str] = None) -> tuple[str, str]:
    """Resolve an indexed repository id or unique bare display/name.

    Raises ValueError if the repo is not found or the bare name is ambiguous.
    """
    if "/" in repo:
        return repo.split("/", 1)

    store = IndexStore(base_path=storage_path)
    mapping = _get_bare_name_map(store)
    candidates = mapping.get(repo, [])

    if not candidates:
        raise ValueError(f"Repository not found: {repo}")
    if len(candidates) > 1:
        raise ValueError(
            f"Ambiguous repository name: {repo}. Use one of: {', '.join(candidates)}"
        )

    return candidates[0].split("/", 1)


def index_status_to_tool_error(status) -> dict:
    """Convert an index status probe into a consistent tool error."""
    hint = status.hint or "Re-index this repository to rebuild the index."
    return {
        "error": f"Repository index is not loadable: {status.repo}",
        "repo": status.repo,
        "index_present": status.index_present,
        "loadable": status.loadable,
        "status": status.status,
        "load_error": status.load_error or status.status,
        "hint": hint,
    }


def load_repo_index_or_error(
    repo: str,
    storage_path: Optional[str] = None,
    branch: str = "",
) -> tuple[Optional[object], Optional[dict], Optional[object]]:
    """Resolve and load a repo index, returning a structured error on failure."""
    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return None, {"error": str(e)}, None

    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name, branch=branch)
    if index is not None:
        return index, None, None

    status = store.inspect_index(owner, name, branch=branch)
    return None, index_status_to_tool_error(status), status


def resolve_fqn(
    repo: str, fqn: str, storage_path: Optional[str] = None
) -> tuple[Optional[str], Optional[str]]:
    """Resolve a PHP FQN to a jcodemunch symbol_id.

    Returns ``(symbol_id, None)`` on success or ``(None, error_message)`` on failure.
    """
    from ..parser.fqn import fqn_to_symbol
    from ..parser.imports import build_psr4_map

    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return None, f"Repository not found: {e}"
    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name)
    if not index:
        status = store.inspect_index(owner, name)
        err = index_status_to_tool_error(status)
        return None, f"{err['error']} ({err['load_error']}). {err['hint']}"
    if not getattr(index, "source_root", None):
        return None, "Index has no source_root (remote indexes don't support FQN resolution)"
    psr4 = build_psr4_map(index.source_root)
    if not psr4:
        return None, "No PSR-4 autoload config found in composer.json"
    resolved = fqn_to_symbol(fqn, psr4, frozenset(index.source_files))
    if not resolved:
        return None, f"FQN '{fqn}' could not be resolved. File not in index or namespace mismatch."
    return resolved, None
