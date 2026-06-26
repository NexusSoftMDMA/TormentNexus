"""Standalone Prefab dashboard served via FastAPI.

Start with: ogham dashboard --port 3113
"""

from __future__ import annotations

import json
from typing import Any

from fastapi import FastAPI, Query
from fastapi.responses import HTMLResponse, JSONResponse

from ogham.database import (
    get_all_memories_full,
    get_backend,
    list_recent_memories,
    query_audit_log,
)


def create_app(profile: str = "default") -> FastAPI:
    app = FastAPI(title="Ogham Dashboard", docs_url=None, redoc_url=None)

    @app.get("/api/stats")
    def api_stats() -> Any:
        return get_backend().get_memory_stats(profile)

    @app.get("/api/memories")
    def api_memories(limit: int = Query(default=50, le=200)) -> Any:
        return JSONResponse(
            json.loads(json.dumps(list_recent_memories(profile=profile, limit=limit), default=str))
        )

    @app.get("/api/audit")
    def api_audit(
        limit: int = Query(default=50, le=200),
        operation: str | None = Query(default=None),
    ) -> Any:
        return JSONResponse(
            json.loads(
                json.dumps(
                    query_audit_log(profile=profile, limit=limit, operation=operation), default=str
                )
            )
        )

    @app.get("/api/decay")
    def api_decay() -> Any:
        return JSONResponse(
            json.loads(json.dumps(get_all_memories_full(profile=profile), default=str))
        )

    @app.get("/", response_class=HTMLResponse)
    def index() -> HTMLResponse:
        from prefab_ui.app import PrefabApp
        from prefab_ui.components import (
            Card,
            CardContent,
            CardHeader,
            CardTitle,
            Column,
            Dashboard,
            DashboardItem,
            DataTable,
            Muted,
            Row,
            Text,
        )
        from prefab_ui.components.charts import PieChart
        from prefab_ui.components.data_table import DataTableColumn
        from prefab_ui.components.metric import Metric
        from prefab_ui.themes import Theme

        stats = get_backend().get_memory_stats(profile)
        total = stats.get("total", 0)
        orphans = stats.get("relationships", {}).get("orphan_count", 0)
        sources = stats.get("sources", {})
        tags_count = stats.get("tagging", {}).get("distinct_tag_count", 0)
        decay_eligible = stats.get("decay", {}).get("eligible_count", 0)
        decay_floor = stats.get("decay", {}).get("floor_count", 0)
        connected_pct = int((total - orphans) / total * 100) if total > 0 else 0

        source_data = [
            {"source": k, "count": v} for k, v in sorted(sources.items(), key=lambda x: -x[1])[:8]
        ]

        from prefab_ui.components import Badge, Code, Separator
        from prefab_ui.components import Column as Col
        from prefab_ui.components.data_table import ExpandableRow

        memories = list_recent_memories(profile=profile, limit=100)
        mem_rows = []
        for m in memories:
            all_tags = m.get("tags", [])
            full_content = m.get("content", "")
            with Col(gap=3) as detail:
                Code(full_content, language="text")
                if all_tags:
                    Separator()
                    with Row(gap=1, css_class="flex-wrap"):
                        for t in all_tags[:10]:
                            Badge(t, variant="secondary")

            mem_rows.append(
                ExpandableRow(
                    {
                        "created_at": str(m.get("created_at", ""))[:19],
                        "content": full_content[:120],
                        "tags": ", ".join(all_tags[:3]),
                        "source": m.get("source", ""),
                    },
                    detail=detail,
                )
            )

        ogham_theme = Theme(
            mode="dark",
            accent="#c8a84e",
            css="body { background: #0d1117; color: #c9d1d9; }",
        )

        with PrefabApp(
            title=f"Ogham -- {profile}",
            state={"profile": profile},
            css_class="p-6 max-w-6xl mx-auto dark",
            theme=ogham_theme,
        ) as prefab_app:
            with Column(gap=6):
                with Row(align="center", css_class="justify-between"):
                    Text(
                        f"Ogham -- {profile}",
                        css_class="text-2xl font-bold",
                    )
                    Muted(f"{total:,} memories")

                with Dashboard(columns=12, row_height="auto", gap=4):
                    with DashboardItem(col=1, row=1, col_span=3):
                        with Card():
                            with CardContent():
                                Metric(
                                    label="Memories",
                                    value=f"{total:,}",
                                )

                    with DashboardItem(col=4, row=1, col_span=3):
                        with Card():
                            with CardContent():
                                Metric(
                                    label="Connected",
                                    value=f"{connected_pct}%",
                                    delta=f"{orphans} orphans",
                                    trend="up" if connected_pct > 50 else "down",
                                    trend_sentiment="positive"
                                    if connected_pct > 50
                                    else "negative",
                                )

                    with DashboardItem(col=7, row=1, col_span=3):
                        with Card():
                            with CardContent():
                                Metric(
                                    label="Tags",
                                    value=f"{tags_count:,}",
                                )

                    with DashboardItem(col=10, row=1, col_span=3):
                        with Card():
                            with CardContent():
                                Metric(
                                    label="Decay",
                                    value=str(decay_eligible),
                                    delta=f"{decay_floor} at floor",
                                    trend="down" if decay_eligible > total // 2 else "up",
                                    trend_sentiment="negative"
                                    if decay_eligible > total // 2
                                    else "positive",
                                )

                    if source_data:
                        with DashboardItem(col=1, row=2, col_span=5):
                            with Card(css_class="h-full"):
                                with CardHeader():
                                    CardTitle("Sources")
                                with CardContent():
                                    PieChart(
                                        data=source_data,
                                        data_key="count",
                                        name_key="source",
                                        inner_radius=50,
                                        show_legend=True,
                                        height=250,
                                    )

                    with DashboardItem(col=1, row=3, col_span=12):
                        with Card():
                            with CardHeader():
                                CardTitle("Recent Memories")
                            with CardContent():
                                DataTable(
                                    rows=mem_rows,
                                    columns=[
                                        DataTableColumn(
                                            key="created_at",
                                            header="Time",
                                            sortable=True,
                                        ),
                                        DataTableColumn(
                                            key="content",
                                            header="Content",
                                        ),
                                        DataTableColumn(
                                            key="tags",
                                            header="Tags",
                                        ),
                                        DataTableColumn(
                                            key="source",
                                            header="Source",
                                            sortable=True,
                                        ),
                                    ],
                                    search=True,
                                    paginated=True,
                                    page_size=10,
                                )

        return HTMLResponse(prefab_app.html())

    return app
