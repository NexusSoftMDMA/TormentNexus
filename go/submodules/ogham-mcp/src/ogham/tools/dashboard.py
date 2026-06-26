"""MCP dashboard tools -- visual composites rendered via Prefab."""

from __future__ import annotations

from typing import Any

from ogham.app import mcp
from ogham.tools.memory import get_active_profile


def _build_profile_health(profile: str) -> Any:
    from ogham.database import get_backend
    from ogham.prefab.composites import profile_health_dashboard

    stats = get_backend().get_memory_stats(profile)
    return profile_health_dashboard(stats)


def _build_audit_log(profile: str, limit: int = 50) -> Any:
    from ogham.database import query_audit_log
    from ogham.prefab.composites import audit_viewer

    events = query_audit_log(profile=profile, limit=limit)
    return audit_viewer(events, limit=limit)


def _build_decay_chart(profile: str) -> Any:
    from ogham.database import get_all_memories_full
    from ogham.prefab.composites import decay_chart

    memories = get_all_memories_full(profile=profile)
    return decay_chart(memories)


@mcp.tool(app=True)
def show_profile_health() -> Any:
    """Visual dashboard showing profile health: memory count, graph
    connectivity, decay status, tagging coverage."""
    return _build_profile_health(get_active_profile())


@mcp.tool(app=True)
def show_audit_log(limit: int = 50) -> Any:
    """Interactive audit log table with sorting and search."""
    return _build_audit_log(get_active_profile(), limit=limit)


@mcp.tool(app=True)
def show_decay_chart() -> Any:
    """Line chart showing memory importance trends over the last 30 days."""
    return _build_decay_chart(get_active_profile())
