from __future__ import annotations

from ogham.okf.viewer.concepts import ViewerConcept

_TYPE_PALETTE = {
    "Decision": "#D4A843",
    "Memory": "#4ADE80",
    "Pattern": "#60A5FA",
    "Gotcha": "#F87171",
    "Reference": "#A78BFA",
    "Topic Summary": "#FB923C",
}
_DEFAULT_COLOR = "#94a3b8"
_MIN_SIZE = 30
_MAX_SIZE_BONUS = 60
_BODY_DIVISOR = 200


def build_elements(concepts: list[ViewerConcept]) -> list[dict]:
    ids = {c.id for c in concepts}
    nodes: list[dict] = []
    for c in concepts:
        nodes.append(
            {
                "data": {
                    "id": c.id,
                    "label": c.title or c.id,
                    "type": c.type,
                    "tags": list(c.tags),
                    "color": _TYPE_PALETTE.get(c.type, _DEFAULT_COLOR),
                    "size": _MIN_SIZE + min(_MAX_SIZE_BONUS, len(c.body) // _BODY_DIVISOR),
                }
            }
        )
    edges: list[dict] = []
    for c in concepts:
        seen: set[str] = set()
        for target in c.links_to:
            if target == c.id or target in seen or target not in ids:
                continue
            seen.add(target)
            edges.append({"data": {"id": f"{c.id}->{target}", "source": c.id, "target": target}})
    return nodes + edges
