from importlib.resources import files

from ogham.okf.viewer.concepts import ViewerConcept, extract_links, parse_bundle
from ogham.okf.viewer.graph import build_elements

_ASSET_NAMES = frozenset({"cytoscape.min.js", "viewer.css", "viewer.js"})


def load_asset(name: str) -> str:
    if name not in _ASSET_NAMES:
        raise FileNotFoundError(f"Unknown viewer asset: {name}")
    return (files("ogham.okf.viewer.assets") / name).read_text(encoding="utf-8")


# Imported after load_asset is defined to avoid a circular import.
from ogham.okf.viewer.render import build_viewer, render_html  # noqa: E402

__all__ = [
    "ViewerConcept",
    "build_elements",
    "build_viewer",
    "extract_links",
    "load_asset",
    "parse_bundle",
    "render_html",
]
