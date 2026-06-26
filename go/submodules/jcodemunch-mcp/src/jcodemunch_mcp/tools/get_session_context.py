"""Get session context tool — thin wrapper around SessionJournal."""

from typing import Optional


def get_session_context(
    max_files: int = 50,
    max_queries: int = 20,
    storage_path: Optional[str] = None,  # noqa: ARG001 - for API consistency
) -> dict:
    """Get the current session context.

    Returns information about files accessed, searches performed, and edits
    registered during the current MCP session.

    Args:
        max_files: Maximum number of files to return in files_accessed.
        max_queries: Maximum number of queries to return in recent_searches.
        storage_path: Ignored (for API consistency with other tools).

    Returns:
        Dict with:
            - files_accessed: List of {file, reads, last_tool}
            - recent_searches: List of {query, count, result_count}
            - files_edited: List of {file, edits}
            - tool_calls: Dict of {tool_name: count}
            - session_duration_s: Seconds since session start
            - total_unique_files: Total unique files accessed
            - total_unique_queries: Total unique queries performed
    """
    from .session_journal import get_journal
    import time


    start = time.perf_counter()
    journal = get_journal()
    result = journal.get_context(max_files=max_files, max_queries=max_queries, max_edits=20)  # Use default for backward compatibility
    result["_meta"] = {
        "timing_ms": round((time.perf_counter() - start) * 1000, 1),
    }
    return result