"""Next.js context provider — detects Next.js projects and enriches symbols with framework metadata.

When a Next.js project is detected (via next.config.js/ts/mjs), this provider:
1. Parses app/ for App Router file-based routing → enriches components with route metadata
2. Parses app/api/ for API route handlers → enriches with endpoint metadata
3. Detects middleware.ts at project root
4. Exposes route metadata via get_metadata()
"""

import logging
import re
from pathlib import Path
from typing import Optional

from .base import ContextProvider, FileContext, register_provider

logger = logging.getLogger(__name__)

# Next.js special files (App Router)
_NEXT_SPECIAL_FILES = frozenset({
    "page", "layout", "loading", "error", "not-found",
    "template", "default", "route", "middleware",
})


def _next_route_from_path(file_path: str, app_dir: str = "app") -> str:
    """Convert a Next.js app/ file path to a route string.

    Examples:
        'app/page.tsx'              → '/'
        'app/users/page.tsx'        → '/users'
        'app/users/[id]/page.tsx'   → '/users/:id'
        'app/posts/[...slug]/page.tsx' → '/posts/*'
    """
    rel = file_path
    if rel.startswith(app_dir + "/"):
        rel = rel[len(app_dir) + 1:]
    # Remove filename (page.tsx, layout.tsx, etc.)
    parts = rel.split("/")
    parts = parts[:-1]  # Remove the file itself

    segments = []
    for part in parts:
        if part.startswith("[..."):
            segments.append("*")
        elif part.startswith("[") and part.endswith("]"):
            param = part[1:-1]
            segments.append(f":{param}")
        elif part.startswith("(") and part.endswith(")"):
            continue  # Route group — does not affect URL
        else:
            segments.append(part)

    return "/" + "/".join(segments) if segments else "/"


def _next_api_methods_from_content(content: str) -> list[str]:
    """Extract exported HTTP method names from a Next.js route handler.

    Example:
        export async function GET(request: Request) { ... }
        export async function POST(request: Request) { ... }
        → ['GET', 'POST']
    """
    methods = []
    for m in re.finditer(
        r"export\s+(?:async\s+)?function\s+(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\b",
        content,
    ):
        methods.append(m.group(1))
    return methods


@register_provider
class NextjsContextProvider(ContextProvider):
    """Context provider for Next.js projects (App Router).

    Detects next.config.js/ts/mjs, then parses:
    - app/**/page.tsx    → page routes with route metadata
    - app/**/layout.tsx  → layout nesting
    - app/api/**/route.ts → API handlers with HTTP methods
    """

    def __init__(self) -> None:
        self._folder: Optional[Path] = None
        self._file_contexts: dict[str, FileContext] = {}
        self._route_metadata: dict[str, dict] = {}
        self._api_metadata: dict[str, dict] = {}

    @property
    def name(self) -> str:
        return "nextjs"

    def detect(self, folder_path: Path) -> bool:
        for config_name in ("next.config.js", "next.config.ts", "next.config.mjs"):
            if (folder_path / config_name).exists():
                self._folder = folder_path
                return True
        return False

    def load(self, folder_path: Path) -> None:
        if self._folder is None:
            self._folder = folder_path

        self._parse_app_router(folder_path)
        self._parse_middleware(folder_path)

    def _parse_app_router(self, folder_path: Path) -> None:
        """Parse app/ directory for pages, layouts, and API routes."""
        app_dir = folder_path / "app"
        if not app_dir.is_dir():
            return

        for src_file in sorted(app_dir.rglob("*")):
            if src_file.suffix not in (".tsx", ".ts", ".jsx", ".js"):
                continue
            stem = src_file.stem
            if stem not in _NEXT_SPECIAL_FILES:
                continue

            rel_path = str(src_file.relative_to(folder_path)).replace("\\", "/")
            route = _next_route_from_path(rel_path)

            if stem == "route":
                # API route handler
                try:
                    content = src_file.read_text("utf-8", errors="replace")
                except Exception:
                    content = ""
                methods = _next_api_methods_from_content(content) or ["ALL"]
                endpoint = "/api" + route if not route.startswith("/api") else route
                # Check if this is actually under app/api/
                if "api/" in rel_path:
                    key = f"{', '.join(methods)} {endpoint}"
                    self._file_contexts[rel_path] = FileContext(
                        description=f"Next.js API route handler. Endpoint: {endpoint}. Methods: {', '.join(methods)}.",
                        tags=["nextjs-api", "endpoint", "app-router"],
                        properties={"methods": ", ".join(methods), "endpoint": endpoint},
                    )
                    self._api_metadata[key] = {"handler": rel_path, "methods": methods}
            elif stem == "page":
                # Page component
                params = re.findall(r":(\w+)", route)
                desc = f"Next.js page. Route: {route}"
                if params:
                    desc += f" (params: {', '.join(params)})"
                self._file_contexts[rel_path] = FileContext(
                    description=desc,
                    tags=["nextjs-page", "route", "app-router"],
                    properties={"route": route},
                )
                self._route_metadata[route] = {"page": rel_path, "method": "GET"}
            elif stem == "layout":
                self._file_contexts[rel_path] = FileContext(
                    description=f"Next.js layout for {route or '/'}",
                    tags=["nextjs-layout", "app-router"],
                    properties={"scope": route or "/"},
                )
            elif stem in ("loading", "error", "not-found"):
                self._file_contexts[rel_path] = FileContext(
                    description=f"Next.js {stem} boundary for {route or '/'}",
                    tags=[f"nextjs-{stem}", "app-router"],
                    properties={"scope": route or "/"},
                )

        logger.info(
            "Next.js: parsed %d page routes, %d API routes",
            len(self._route_metadata), len(self._api_metadata),
        )

    def _parse_middleware(self, folder_path: Path) -> None:
        """Detect middleware.ts at project root."""
        for name in ("middleware.ts", "middleware.js"):
            mw_path = folder_path / name
            if mw_path.exists():
                try:
                    content = mw_path.read_text("utf-8", errors="replace")
                except Exception:
                    content = ""
                # Extract matcher config
                matcher_match = re.search(
                    r"matcher\s*:\s*\[([^\]]+)\]", content
                )
                matchers = []
                if matcher_match:
                    matchers = re.findall(r"""['"]([^'"]+)['"]""", matcher_match.group(1))

                rel_path = str(mw_path.relative_to(folder_path)).replace("\\", "/")
                props: dict[str, str] = {}
                if matchers:
                    props["matcher"] = ", ".join(matchers)
                self._file_contexts[rel_path] = FileContext(
                    description="Next.js middleware",
                    tags=["nextjs-middleware"],
                    properties=props,
                )
                break

    def get_file_context(self, file_path: str) -> Optional[FileContext]:
        return self._file_contexts.get(file_path)

    def get_extra_imports(self) -> dict[str, list[dict]]:
        return {}  # Next.js doesn't have auto-imports like Nuxt

    def get_metadata(self) -> dict:
        meta: dict = {}
        if self._route_metadata:
            meta["nextjs_routes"] = self._route_metadata
        if self._api_metadata:
            meta["nextjs_api"] = self._api_metadata
        return meta

    def stats(self) -> dict:
        return {
            "page_routes": len(self._route_metadata),
            "api_routes": len(self._api_metadata),
        }
