from __future__ import annotations

import json
from pathlib import Path

from ogham.okf.viewer import load_asset
from ogham.okf.viewer.concepts import parse_bundle
from ogham.okf.viewer.graph import build_elements

_HTML_TEMPLATE = """\
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>{title} -- OKF viewer</title>
<style>
{css}
</style>
</head>
<body>
<header>
  <h1>{title}</h1>
  <small>OKF v0.1 -- Ogham bundle viewer</small>
</header>
<div id="graph"></div>
<aside id="panel"></aside>
<script>
{cytoscape}
</script>
<script>
window.__OKF_ELEMENTS__ = {elements_json};
</script>
<script>
{viewer_js}
</script>
</body>
</html>
"""


def _escape_html(text: str) -> str:
    return (
        text.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;").replace('"', "&quot;")
    )


def _safe_json(elements: list[dict]) -> str:
    return json.dumps(elements, ensure_ascii=False).replace("</", "<\\/")


def render_html(elements: list[dict], bundle_title: str) -> str:
    return _HTML_TEMPLATE.format(
        title=_escape_html(bundle_title),
        css=load_asset("viewer.css"),
        cytoscape=load_asset("cytoscape.min.js"),
        viewer_js=load_asset("viewer.js"),
        elements_json=_safe_json(elements),
    )


def build_viewer(bundle_dir: Path, out_path: Path | None = None) -> Path:
    bundle_dir = Path(bundle_dir)
    if out_path is None:
        out_path = bundle_dir / "viewer.html"
    else:
        out_path = Path(out_path)
    concepts = parse_bundle(bundle_dir)
    elements = build_elements(concepts)
    html = render_html(elements, bundle_dir.name)
    out_path.write_text(html, encoding="utf-8")
    return out_path
