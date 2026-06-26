"""Retriever — bridges gcm CLI to jCodeMunch index + ranked context."""

import asyncio
import os
import sys
from typing import Optional

from ..tools.list_repos import list_repos
from ..tools.get_ranked_context import get_ranked_context
from ..tools.index_folder import index_folder


def _is_github_repo(repo: str) -> bool:
    """True if repo looks like owner/name (not a local path).

    GitHub repos have exactly one slash, no backslashes, no dots or colons
    at the start (ruling out relative/absolute paths), and don't exist on disk.
    """
    if repo.count("/") != 1:
        return False
    owner, name = repo.split("/", 1)
    if not owner or not name:
        return False
    # Rule out obvious local paths
    if repo.startswith((".", "/", "\\")) or "\\" in repo or ":" in repo:
        return False
    return not os.path.exists(repo)


def _find_indexed_repo(repo: str, storage_path: Optional[str] = None) -> Optional[str]:
    """Return the full owner/name if repo is already indexed, else None."""
    result = list_repos(storage_path=storage_path)
    for entry in result.get("repos", []):
        repo_id = entry["repo"]
        # Exact match
        if repo_id == repo:
            return repo_id
        # Bare name match (e.g. "flask" matches "pallets/flask")
        if "/" in repo_id and repo_id.split("/", 1)[1] == repo:
            return repo_id
        # Display name match
        if entry.get("display_name") == repo:
            return repo_id
    return None


def ensure_indexed(
    repo: str,
    storage_path: Optional[str] = None,
    github_token: Optional[str] = None,
    verbose: bool = False,
) -> tuple[Optional[str], Optional[str]]:
    """Ensure repo is indexed. Returns (repo_id, None) on success or (None, error).

    Handles three cases:
    1. Already indexed — returns immediately
    2. GitHub repo (owner/name) — indexes via index_repo
    3. Local path — indexes via index_folder
    """
    # Check if already indexed
    existing = _find_indexed_repo(repo, storage_path)
    if existing:
        return existing, None

    # GitHub repo
    if _is_github_repo(repo):
        if verbose:
            print(f"Indexing {repo} from GitHub...", file=sys.stderr)
        from ..tools.index_repo import index_repo
        result = asyncio.run(index_repo(
            url=repo,
            use_ai_summaries=False,
            github_token=github_token,
            storage_path=storage_path,
        ))
        if not result.get("success", False):
            return None, result.get("error", "Indexing failed")
        # Re-check
        existing = _find_indexed_repo(repo, storage_path)
        if existing:
            return existing, None
        return None, "Indexing succeeded but repo not found in store"

    # Local path
    local_path = os.path.abspath(repo) if not os.path.isabs(repo) else repo
    if os.path.isdir(local_path):
        if verbose:
            print(f"Indexing {local_path}...", file=sys.stderr)
        result = index_folder(
            path=local_path,
            use_ai_summaries=False,
            storage_path=storage_path,
        )
        if not result.get("success", False):
            return None, result.get("error", "Indexing failed")
        existing = _find_indexed_repo(local_path, storage_path)
        if existing is None:
            # index_folder uses folder name as display_name; try repo field
            for entry in list_repos(storage_path=storage_path).get("repos", []):
                source = entry.get("source_root", "")
                if source and os.path.normcase(os.path.normpath(source)) == os.path.normcase(os.path.normpath(local_path)):
                    return entry["repo"], None
        if existing:
            return existing, None
        return None, "Indexing succeeded but repo not found in store"

    return None, f"Not a GitHub repo or local directory: {repo}"


def retrieve_context(
    repo_id: str,
    query: str,
    token_budget: int = 8000,
    storage_path: Optional[str] = None,
) -> tuple[str, dict]:
    """Retrieve ranked context for a query. Returns (formatted_context, raw_result)."""
    result = get_ranked_context(
        repo=repo_id,
        query=query,
        token_budget=token_budget,
        strategy="combined",
        fusion=True,
        storage_path=storage_path,
    )

    if "error" in result:
        return "", result

    items = result.get("context_items", [])
    if not items:
        return "", result

    # Format context for the LLM
    parts = []
    for item in items:
        header = f"# {item.get('file', '?')} :: {item.get('symbol', '?')} ({item.get('kind', '?')})"
        source = item.get("source", "")
        parts.append(f"{header}\n```\n{source}\n```")

    formatted = "\n\n".join(parts)
    return formatted, result
