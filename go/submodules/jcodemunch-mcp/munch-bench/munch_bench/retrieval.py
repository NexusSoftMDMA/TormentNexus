"""Retrieval layer — wraps jCodeMunch for benchmark evaluation."""

from __future__ import annotations

import time
from dataclasses import dataclass, field
from typing import Optional

from jcodemunch_mcp.tools.get_ranked_context import get_ranked_context
from jcodemunch_mcp.tools.list_repos import list_repos


@dataclass
class RetrievalResult:
    """Result of a single retrieval call."""

    symbols_returned: list[str]
    context_text: str
    token_count: int
    wall_time_s: float
    raw: dict = field(default_factory=dict)


def find_repo_id(repo: str, storage_path: Optional[str] = None) -> Optional[str]:
    """Find the indexed repo ID matching a repo name or owner/name."""
    result = list_repos(storage_path=storage_path)
    for entry in result.get("repos", []):
        repo_id = entry["repo"]
        if repo_id == repo:
            return repo_id
        if "/" in repo_id and repo_id.split("/", 1)[1] == repo:
            return repo_id
        if entry.get("display_name") == repo:
            return repo_id
    return None


def retrieve(
    repo_id: str,
    query: str,
    token_budget: int = 8000,
    storage_path: Optional[str] = None,
) -> RetrievalResult:
    """Run retrieval against the jCodeMunch index."""
    t0 = time.perf_counter()
    result = get_ranked_context(
        repo=repo_id,
        query=query,
        token_budget=token_budget,
        strategy="combined",
        fusion=True,
        storage_path=storage_path,
    )
    elapsed = time.perf_counter() - t0

    items = result.get("context_items", [])
    symbols = []
    for item in items:
        # symbol_id format: "path::name#kind" — extract the name part
        sid = item.get("symbol_id", "")
        if "::" in sid:
            name_part = sid.split("::", 1)[1]
            # Strip #kind suffix if present
            if "#" in name_part:
                name_part = name_part.rsplit("#", 1)[0]
            symbols.append(name_part)
        else:
            symbols.append(item.get("symbol", sid))

    parts = []
    for item in items:
        sid = item.get("symbol_id", "?")
        source = item.get("source", "")
        parts.append(f"# {sid}\n```\n{source}\n```")
    context_text = "\n\n".join(parts)

    return RetrievalResult(
        symbols_returned=symbols,
        context_text=context_text,
        token_count=result.get("total_tokens", 0),
        wall_time_s=elapsed,
        raw=result,
    )
