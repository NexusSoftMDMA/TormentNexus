"""Django context provider for Django projects.

This provider detects Django projects and extracts URL patterns, DRF decorators,
and middleware information.
"""

from __future__ import annotations

import logging
import re
from pathlib import Path
from typing import Optional

from ._route_utils import has_dependency, make_route_file_context
from .base import ContextProvider, FileContext, register_provider

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Patterns
# ---------------------------------------------------------------------------

# Django URL pattern: path('url', view, name='name')
_DJANGO_PATH = re.compile(
    r"path\s*\(\s*[\"'](?P<url>[^\"']*)[\"']"
    r"\s*,\s*(?P<view>[^,)]+)"
    r"(?:\s*,\s*name\s*=\s*[\"'](?P<name>[^\"']+)[\"'])?",
)

# Django include: include('module.urls')
_DJANGO_INCLUDE = re.compile(
    r"include\s*\(\s*[\"'](?P<module>[^\"']+)[\"']",
)

# DRF @api_view decorator
_DRF_API_VIEW = re.compile(
    r"@api_view\s*\(\s*\[(?P<methods>[^\]]+)\]\s*\)",
)

# DRF @action decorator
_DRF_ACTION = re.compile(
    r"@action\s*\(\s*(?:detail\s*=\s*(?P<detail>True|False)[,\s]*)?\)",
)

# Django middleware in settings
_DJANGO_MIDDLEWARE = re.compile(
    r"['\"](?P<mw>[^'\"]+\.[^'\"]+)['\"](?:\s*,)?",
)


@register_provider
class DjangoProvider(ContextProvider):
    """Context provider for Django projects.

    Detects via manage.py + django dependency, then parses:
    - **/urls.py for URL patterns
    - DRF decorators in views.py
    - settings.py for middleware
    """

    def __init__(self) -> None:
        self._folder: Optional[Path] = None
        self._file_contexts: dict[str, FileContext] = {}
        self._extra_imports: dict[str, list[dict]] = {}

    @property
    def name(self) -> str:
        return "django"

    def detect(self, folder: Path) -> bool:
        """Detect Django via manage.py + django dependency."""
        if not (folder / "manage.py").exists():
            return False
        if not has_dependency(folder, "django", ["requirements.txt", "pyproject.toml"]):
            return False
        logger.info("Django provider: detected Django project")
        return True

    def load(self, folder: Path) -> None:
        """Scan for Django URL patterns and DRF decorators."""
        self._folder = folder
        url_count = 0
        drf_count = 0

        # Scan for urls.py files
        for urls_file in folder.glob("**/urls.py"):
            if self._skip_file(urls_file):
                continue
            try:
                content = urls_file.read_text("utf-8", errors="replace")
            except Exception:
                continue

            rel_path = str(urls_file.relative_to(folder)).replace("\\", "/")
            routes = self._extract_url_patterns(content)
            if routes:
                ctx = make_route_file_context("django", routes, kind="url")
                self._file_contexts[rel_path] = ctx
                url_count += len(routes)

                # Create import edges for include() statements
                self._extract_include_edges(content, rel_path)

        # Scan for DRF decorators in views
        for views_file in folder.glob("**/views.py"):
            if self._skip_file(views_file):
                continue
            try:
                content = views_file.read_text("utf-8", errors="replace")
            except Exception:
                continue

            drf_routes = self._extract_drf_routes(content)
            if drf_routes:
                rel_path = str(views_file.relative_to(folder)).replace("\\", "/")
                if rel_path in self._file_contexts:
                    # Merge with existing URL context
                    existing = self._file_contexts[rel_path]
                    combined_routes = self._parse_routes(existing.properties.get("routes", ""))
                    combined_routes.extend(drf_routes)
                    ctx = make_route_file_context("django", combined_routes, kind="drf")
                else:
                    ctx = make_route_file_context("django", drf_routes, kind="drf")
                self._file_contexts[rel_path] = ctx
                drf_count += len(drf_routes)

        logger.info("Django provider: found %d URL patterns, %d DRF routes", url_count, drf_count)

    def _skip_file(self, file_path: Path) -> bool:
        """Check if a file should be skipped."""
        rel = str(file_path)
        skip = ("venv/", ".venv/", "env/", "node_modules/", "migrations/", "__pycache__/")
        return any(s in rel for s in skip)

    def _extract_url_patterns(self, content: str) -> list[dict]:
        """Extract URL patterns from urls.py content."""
        routes = []

        for m in _DJANGO_PATH.finditer(content):
            url = m.group("url") or "/"
            view = m.group("view") or ""
            name = m.group("name") or ""
            route_str = f"{view}"
            if name:
                route_str += f" ({name})"
            routes.append({"verb": "PATH", "path": url, "view": view, "name": name})

        return routes

    def _extract_include_edges(self, content: str, from_file: str) -> None:
        """Create import edges for include() statements."""
        for m in _DJANGO_INCLUDE.finditer(content):
            module = m.group("module")
            # Try to resolve the module to a file path
            resolved = self._resolve_include_module(module, from_file)
            if resolved:
                self._extra_imports.setdefault(from_file, []).append({
                    "specifier": resolved,
                    "names": [],
                })

    def _resolve_include_module(self, module: str, from_file: str) -> Optional[str]:
        """Resolve an include() module path to a file path."""
        if self._folder is None:
            return None

        # Convert module path to file path
        # e.g., 'api.urls' -> 'api/urls.py'
        parts = module.split(".")
        if not parts:
            return None

        from_dir = (self._folder / from_file).parent

        # Try as relative path first
        for i in range(len(parts), 0, -1):
            subpath = "/".join(parts[:i])
            candidate = from_dir / subpath
            for ext in ("", "/urls.py", "/urls/index.py"):
                if ext:
                    test_path = candidate / ext if not candidate.suffix else candidate
                else:
                    test_path = candidate
                test_path = candidate.parent / (candidate.name + ".py") if candidate.suffix and not candidate.name.endswith(".py") else candidate
                # Simple approach: try the module as a direct .py file
                direct_py = from_dir / f"{subpath}.py"
                if direct_py.exists():
                    return str(direct_py.relative_to(self._folder)).replace("\\", "/")
                urls_py = from_dir / subpath / "urls.py"
                if urls_py.exists():
                    return str(urls_py.relative_to(self._folder)).replace("\\", "/")

        return None

    def _extract_drf_routes(self, content: str) -> list[dict]:
        """Extract DRF @api_view and @action routes."""
        routes = []

        # @api_view(['GET', 'POST'])
        for m in _DRF_API_VIEW.finditer(content):
            methods_str = m.group("methods")
            methods = [m.strip().strip("'\"") for m in methods_str.split(",")]
            # Find the function that follows this decorator
            func_match = re.search(r"def\s+(\w+)\s*\(", content[m.end():m.end() + 200])
            if func_match:
                func_name = func_match.group(1)
                for method in methods:
                    routes.append({
                        "verb": method.upper(),
                        "path": f"@api_view {func_name}",
                        "view": func_name,
                    })

        # @action(detail=True/False)
        for m in _DRF_ACTION.finditer(content):
            detail = m.group("detail")
            func_match = re.search(r"def\s+(\w+)\s*\(", content[m.end():m.end() + 200])
            if func_match:
                func_name = func_match.group(1)
                routes.append({
                    "verb": "ACTION",
                    "path": f"@action {func_name} (detail={detail})",
                    "view": func_name,
                })

        return routes

    def _parse_routes(self, routes_str: str) -> list[dict]:
        """Parse route strings back into route dicts."""
        routes = []
        if not routes_str:
            return routes
        for part in routes_str.split(";"):
            part = part.strip()
            if not part:
                continue
            # Format: "VERB path" or "path (name)"
            parts = part.split()
            if len(parts) >= 2:
                verb = parts[0]
                path = parts[1]
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
        """Return include() edges."""
        return self._extra_imports

    def get_metadata(self) -> dict:
        """Return metadata."""
        return {"django_provider": {"url_files": len(self._file_contexts)}}

    def stats(self) -> dict:
        """Return provider stats."""
        return {"url_files": len(self._file_contexts)}
