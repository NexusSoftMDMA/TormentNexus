"""Go Router provider for Gin, Chi, Echo, and Fiber.

This provider detects Go frameworks that use call-expression routing
(e.g., r.GET("/path", handler)) and extracts routes and middleware.
"""

from __future__ import annotations

import logging
import re
from pathlib import Path
from typing import Optional

from ._route_utils import make_route_file_context
from .base import ContextProvider, FileContext, register_provider

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Detection
# ---------------------------------------------------------------------------

_GO_FRAMEWORKS = {
    "gin": "github.com/gin-gonic/gin",
    "chi": "github.com/go-chi/chi",
    "echo": "github.com/labstack/echo",
    "fiber": "github.com/gofiber/fiber",
}

# Pattern: r.GET("/path", handler) or r.Post("/path", handler)
_GO_ROUTE = re.compile(
    r'(?:\w+)\s*\.\s*(?P<method>GET|Get|POST|Post|PUT|Put|PATCH|Patch|DELETE|Delete|HEAD|Head|OPTIONS|Options|Any)\s*\(\s*'
    r'"(?P<path>[^"]+)"',
)

# Pattern: r.Use(middleware) or r.Middleware(middleware)
_GO_MIDDLEWARE = re.compile(
    r'(?:\w+)\s*\.\s*(?:Use|Middleware)\s*\(\s*(?P<handler>[^)]+)\)',
)

# Pattern: r.Group("/prefix")
_GO_GROUP = re.compile(
    r'(?:\w+)\s*\.\s*Group\s*\(\s*"(?P<prefix>[^"]+)"',
)


@register_provider
class GoRoutersProvider(ContextProvider):
    """Context provider for Gin, Chi, Echo, and Fiber frameworks.

    Detects via go.mod and extracts routes/middleware from Go source files.
    """

    def __init__(self) -> None:
        self._folder: Optional[Path] = None
        self._framework: Optional[str] = None
        self._file_contexts: dict[str, FileContext] = {}

    @property
    def name(self) -> str:
        return "go-routers"

    def detect(self, folder: Path) -> bool:
        """Detect Go frameworks via go.mod."""
        go_mod = folder / "go.mod"
        if not go_mod.exists():
            return False

        try:
            content = go_mod.read_text("utf-8", errors="replace")
        except Exception:
            return False

        for fw_name, module_path in _GO_FRAMEWORKS.items():
            if module_path in content:
                self._framework = fw_name
                logger.info("Go routers provider: detected %s", fw_name)
                return True

        return False

    def load(self, folder: Path) -> None:
        """Scan Go files for routes and middleware."""
        self._folder = folder
        framework = self._framework or "gin"
        skip_dirs = ("vendor/", "testdata/", ".git/")

        route_count = 0
        file_count = 0

        for file_path in folder.glob("**/*.go"):
            rel_path = str(file_path.relative_to(folder)).replace("\\", "/")
            if any(skip in rel_path for skip in skip_dirs):
                continue

            try:
                content = file_path.read_text("utf-8", errors="replace")
            except Exception:
                continue

            routes = self._extract_routes(content)
            if routes:
                ctx = make_route_file_context(framework, routes, kind="route")
                self._file_contexts[rel_path] = ctx
                route_count += len(routes)
                file_count += 1

        logger.info(
            "Go routers provider: found %d routes in %d files (%s)",
            route_count, file_count, framework,
        )

    def _extract_routes(self, content: str) -> list[dict]:
        """Extract route dicts from content."""
        routes = []
        method_map = {
            "GET": "GET", "Get": "GET",
            "POST": "POST", "Post": "POST",
            "PUT": "PUT", "Put": "PUT",
            "PATCH": "PATCH", "Patch": "PATCH",
            "DELETE": "DELETE", "Delete": "DELETE",
            "HEAD": "HEAD", "Head": "HEAD",
            "OPTIONS": "OPTIONS", "Options": "OPTIONS",
            "ANY": "ANY", "Any": "ANY",
        }

        for m in _GO_ROUTE.finditer(content):
            verb_raw = m.group("method")
            verb = method_map.get(verb_raw, verb_raw.upper())
            path = m.group("path") or "/"
            routes.append({"verb": verb, "path": path})

        return routes

    def get_file_context(self, file_path: str) -> Optional[FileContext]:
        """Look up FileContext for a file."""
        if file_path in self._file_contexts:
            return self._file_contexts[file_path]
        stem = Path(file_path).stem
        for key, ctx in self._file_contexts.items():
            if Path(key).stem == stem:
                return ctx
        return None

    def get_extra_imports(self) -> dict[str, list[dict]]:
        """Return extra imports (currently empty for Go)."""
        return {}

    def get_metadata(self) -> dict:
        """Return route summary metadata."""
        return {"go_framework": self._framework or "gin"}

    def stats(self) -> dict:
        """Return provider stats."""
        return {
            "framework": self._framework or "gin",
            "route_files": len(self._file_contexts),
        }
