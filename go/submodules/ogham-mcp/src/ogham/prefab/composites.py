"""Ogham Prefab composite wrappers.

Each function builds a Prefab component tree from Ogham data.
Composites enforce the Ogham theme (navy + gold) so all dashboards
look consistent. Requires prefab-ui >= 0.1.0.

These are designed to be called from MCP tool handlers that return
Prefab apps. They will not work until FastMCP >= 3.2.5 ships with
the task_redis_prefix fix.

Phase 1 composites (use native Prefab components):
  - profile_health_dashboard
  - audit_viewer
  - decay_chart
  - memory_search_table

Phase 2 composites (use Embed for custom JS):
  - timeline_viewer
  - entity_graph
  - calendar_view
"""

from __future__ import annotations

from typing import Any


def profile_health_dashboard(stats: dict[str, Any]) -> Any:
    """Cards + badges + progress bars showing profile health counters.

    Expects output from get_memory_stats() with health counters
    (relationships.orphan_count, tagging.*, decay.*).
    """
    from prefab_ui.components import (
        Badge,
        Card,
        CardContent,
        CardHeader,
        CardTitle,
        Column,
        Grid,
        Metric,
        Progress,
        Row,
        Text,
    )

    total = stats.get("total", 0)
    orphans = stats.get("relationships", {}).get("orphan_count", 0)
    untagged = stats.get("tagging", {}).get("untagged_count", 0)
    distinct_tags = stats.get("tagging", {}).get("distinct_tag_count", 0)
    decay_eligible = stats.get("decay", {}).get("eligible_count", 0)
    decay_floor = stats.get("decay", {}).get("floor_count", 0)

    connected_pct = ((total - orphans) / total * 100) if total > 0 else 0
    tagged_pct = ((total - untagged) / total * 100) if total > 0 else 0

    with Grid(columns=3, gap=4) as grid:
        with Card():
            with CardHeader():
                CardTitle("Profile Overview")
            with CardContent():
                Metric(value=str(total), label="Total memories")
                with Row(gap=2):
                    Badge(f"{len(stats.get('sources', {}))} sources", variant="default")
                    Badge(f"{distinct_tags} tags", variant="secondary")

        with Card():
            with CardHeader():
                CardTitle("Graph Health")
            with CardContent():
                with Column(gap=2):
                    Text(f"{total - orphans} connected, {orphans} orphans")
                    Progress(value=int(connected_pct))
                    Text(
                        f"{connected_pct:.0f}% connected", css_class="text-xs text-muted-foreground"
                    )

        with Card():
            with CardHeader():
                CardTitle("Decay Status")
            with CardContent():
                with Column(gap=2):
                    Metric(value=str(decay_eligible), label="Eligible for decay")
                    Metric(value=str(decay_floor), label="At floor (0.05)")
                    Text(f"{tagged_pct:.0f}% tagged", css_class="text-xs text-muted-foreground")
                    Progress(value=int(tagged_pct))

    return grid


def audit_viewer(events: list[dict[str, Any]], limit: int = 50) -> Any:
    """DataTable showing recent audit log events."""
    from prefab_ui.components import DataTable
    from prefab_ui.components.data_table import DataTableColumn

    columns = [
        DataTableColumn(key="created_at", header="Time", sortable=True),
        DataTableColumn(key="operation", header="Operation", sortable=True),
        DataTableColumn(key="profile", header="Profile"),
        DataTableColumn(key="resource_id", header="Resource ID"),
        DataTableColumn(key="outcome", header="Outcome"),
    ]

    rows = []
    for e in events[:limit]:
        rows.append(
            {
                "created_at": str(e.get("created_at", ""))[:19],
                "operation": e.get("operation", ""),
                "profile": e.get("profile", ""),
                "resource_id": str(e.get("resource_id", ""))[:8],
                "outcome": e.get("outcome", ""),
            }
        )

    return DataTable(rows=rows, columns=columns, search=True)


def decay_chart(memories: list[dict[str, Any]]) -> Any:
    """LineChart showing importance distribution over time.

    Expects a list of memories with 'importance', 'created_at',
    and optionally 'access_count' fields.
    """
    from prefab_ui.components import Column, Text
    from prefab_ui.components.charts import ChartSeries, LineChart

    buckets: dict[str, list[float]] = {}
    for m in memories:
        date = str(m.get("created_at", ""))[:10]
        if date:
            buckets.setdefault(date, []).append(m.get("importance", 0.5))

    data = []
    for date in sorted(buckets.keys())[-30:]:
        vals = buckets[date]
        data.append(
            {
                "date": date,
                "avg_importance": round(sum(vals) / len(vals), 3),
                "count": len(vals),
            }
        )

    with Column(gap=2) as col:
        Text("Importance over time (last 30 days)", css_class="text-sm font-medium")
        LineChart(
            data=data,
            series=[
                ChartSeries(data_key="avg_importance", label="Avg Importance", color="#c8a84e"),
                ChartSeries(data_key="count", label="Count", color="#6b8e6b"),
            ],
            x_axis="date",
            height=300,
            show_legend=True,
        )

    return col


def memory_search_table(results: list[dict[str, Any]]) -> Any:
    """DataTable showing search results with relevance scores."""
    from prefab_ui.components import DataTable
    from prefab_ui.components.data_table import DataTableColumn

    columns = [
        DataTableColumn(key="relevance", header="Score", sortable=True),
        DataTableColumn(key="content", header="Content"),
        DataTableColumn(key="created_at", header="Created", sortable=True),
        DataTableColumn(key="tags", header="Tags"),
        DataTableColumn(key="source", header="Source"),
    ]

    rows = []
    for r in results:
        rows.append(
            {
                "relevance": f"{r.get('relevance', r.get('similarity', 0)):.3f}",
                "content": r.get("content", "")[:120],
                "created_at": str(r.get("created_at", ""))[:10],
                "tags": ", ".join(r.get("tags", [])[:3]),
                "source": r.get("source", ""),
            }
        )

    return DataTable(rows=rows, columns=columns, search=True)


# --- Phase 2: Embed-based composites (custom JS) ---


def timeline_viewer(memories: list[dict[str, Any]]) -> Any:
    """Interactive timeline using vis-timeline via Embed.

    Renders memories on a zoomable timeline grouped by source.
    """
    from prefab_ui.components import Embed

    items_js = []
    for m in memories:
        items_js.append(
            {
                "id": str(m.get("id", ""))[:8],
                "content": m.get("content", "")[:60].replace('"', '\\"'),
                "start": str(m.get("created_at", ""))[:19],
                "group": m.get("source", "unknown"),
            }
        )

    groups = list({m.get("source", "unknown") for m in memories})
    groups_js = [{"id": g, "content": g} for g in sorted(groups)]

    import json

    vis_css = "https://unpkg.com/vis-timeline@7.7.3/styles/vis-timeline-graph2d.min.css"
    vis_js = "https://unpkg.com/vis-timeline@7.7.3/standalone/umd/vis-timeline-graph2d.min.js"

    html = f"""<!DOCTYPE html>
<html><head>
<link href="{vis_css}" rel="stylesheet">
<script src="{vis_js}"></script>
<style>
body{{margin:0;font-family:Inter,sans-serif;background:#0d1117;color:#c9d1d9}}
.vis-timeline{{border-color:#30363d}}
.vis-item{{background:#c8a84e;color:#0d1117;border:none;border-radius:4px}}
.vis-item.vis-selected{{background:#d4af37}}
.vis-labelset .vis-label{{color:#9ca3af}}
</style>
</head><body>
<div id="tl" style="width:100%;height:100%"></div>
<script>
var items=new vis.DataSet({json.dumps(items_js)});
var groups=new vis.DataSet({json.dumps(groups_js)});
new vis.Timeline(document.getElementById('tl'),items,groups,{{
  stack:true,zoomMin:86400000,zoomMax:31536000000
}});
</script></body></html>"""

    return Embed(html=html, height="400px", sandbox="allow-scripts")


def entity_graph(memories: list[dict[str, Any]], relationships: list[dict[str, Any]]) -> Any:
    """Interactive knowledge graph using vis-network via Embed.

    Renders memories as nodes and relationships as edges.
    """
    from prefab_ui.components import Embed

    nodes = []
    seen = set()
    for m in memories:
        mid = str(m.get("id", ""))
        if mid not in seen:
            seen.add(mid)
            nodes.append(
                {
                    "id": mid[:8],
                    "label": m.get("content", "")[:40],
                    "color": "#c8a84e",
                }
            )

    edges = []
    for r in relationships:
        edges.append(
            {
                "from": str(r.get("source_id", ""))[:8],
                "to": str(r.get("target_id", ""))[:8],
                "label": r.get("relationship", ""),
                "color": {
                    "color": "#6b8e6b" if r.get("relationship") != "contradicts" else "#f85149"
                },
            }
        )

    import json

    html = f"""<!DOCTYPE html>
<html><head>
<script src="https://unpkg.com/vis-network@9.1.9/standalone/umd/vis-network.min.js"></script>
<style>body{{margin:0;background:#0d1117}}#graph{{width:100%;height:100%}}</style>
</head><body>
<div id="graph"></div>
<script>
var nodes=new vis.DataSet({json.dumps(nodes)});
var edges=new vis.DataSet({json.dumps(edges)});
new vis.Network(document.getElementById('graph'),{{nodes:nodes,edges:edges}},{{
  nodes:{{shape:'dot',size:16,font:{{color:'#c9d1d9',size:11}}}},
  edges:{{arrows:'to',font:{{color:'#9ca3af',size:9}}}},
  physics:{{solver:'forceAtlas2Based'}},
  layout:{{improvedLayout:true}}
}});
</script></body></html>"""

    return Embed(html=html, height="500px", sandbox="allow-scripts")
