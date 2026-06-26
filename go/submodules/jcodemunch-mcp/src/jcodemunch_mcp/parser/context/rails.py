"""Rails context provider for Ruby on Rails projects.

This provider detects Rails projects and parses routes.rb to extract
resource routes and verb routes.
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

# Rails resources: resources :users
_RAILS_RESOURCES = re.compile(
    r"resources?\s+:(?P<name>\w+)",
)

# Rails verb routes: get 'path', to: 'controller#action'
_RAILS_VERB = re.compile(
    r"(?P<verb>get|post|put|patch|delete)\s+[\"'](?P<path>[^\"']+)[\"']"
    r"\s*,\s*to:\s*[\"']?(?P<controller>\w+)#(?P<action>\w+)[\"']?",
    re.IGNORECASE,
)

# Rails namespace: namespace :admin do
_RAILS_NAMESPACE = re.compile(
    r"namespace\s+:(?P<name>\w+)",
)

# Rails scope: scope '/api' do
_RAILS_SCOPE = re.compile(
    r"scope\s+['\"](?P<path>[^\"']+)['\"]",
)


# RESTful route expansion for a resource
_RESTFUL_ROUTES = [
    ("GET", ""),           # index
    ("GET", "/new"),       # new
    ("POST", ""),          # create
    ("GET", "/:id"),       # show
    ("GET", "/:id/edit"),  # edit
    ("PATCH", "/:id"),     # update
    ("DELETE", "/:id"),    # destroy
]


@register_provider
class RailsProvider(ContextProvider):
    """Context provider for Rails projects.

    Detects via Gemfile + rails gem AND config/routes.rb, then parses
    routes.rb to extract resource routes and verb routes.
    """

    def __init__(self) -> None:
        self._folder: Optional[Path] = None
        self._file_contexts: dict[str, FileContext] = {}
        self._extra_imports: dict[str, list[dict]] = {}

    @property
    def name(self) -> str:
        return "rails"

    def detect(self, folder: Path) -> bool:
        """Detect Rails via Gemfile with rails gem AND config/routes.rb."""
        routes_rb = folder / "config" / "routes.rb"
        if not routes_rb.exists():
            return False
        if not has_dependency(folder, "rails", ["Gemfile"]):
            return False
        logger.info("Rails provider: detected Rails project")
        return True

    def load(self, folder: Path) -> None:
        """Parse routes.rb to extract routes."""
        self._folder = folder
        routes_rb = folder / "config" / "routes.rb"
        if not routes_rb.exists():
            return

        try:
            content = routes_rb.read_text("utf-8", errors="replace")
        except Exception:
            return

        rel_path = "config/routes.rb"
        routes = self._parse_routes_rb(content)

        if routes:
            ctx = make_route_file_context("rails", routes, kind="route")
            self._file_contexts[rel_path] = ctx

            # Create import edges from routes.rb to controllers
            self._extract_controller_edges(content, rel_path)

        logger.info("Rails provider: found %d routes", len(routes))

    def _parse_routes_rb(self, content: str) -> list[dict]:
        """Parse routes.rb content to extract routes."""
        routes = []
        namespace_stack = []  # Stack of active namespace prefixes
        scope_stack = []      # Stack of active scope prefixes

        lines = content.splitlines()
        i = 0
        while i < len(lines):
            line = lines[i].strip()

            # Skip empty lines and comments
            if not line or line.startswith("#"):
                i += 1
                continue

            # Handle namespace
            ns_match = re.search(r"namespace\s+:(?P<name>\w+)", line)
            if ns_match:
                namespace_stack.append(ns_match.group("name"))
                i += 1
                continue

            # Handle scope
            sc_match = re.search(r"scope\s+['\"](?P<path>[^\"']+)['\"]", line)
            if sc_match:
                scope_stack.append(sc_match.group("path"))
                i += 1
                continue

            # Handle resources
            res_match = re.search(r"resources?\s+:(?P<name>\w+)", line)
            if res_match:
                resource_name = res_match.group("name")
                prefix = self._build_prefix(namespace_stack, scope_stack)
                expanded = self._expand_resource_routes(resource_name, prefix)
                routes.extend(expanded)
                i += 1
                continue

            # Handle singular resources (resource :user)
            singular_match = re.search(r"resource\s+:(?P<name>\w+)", line)
            if singular_match:
                resource_name = singular_match.group("name")
                prefix = self._build_prefix(namespace_stack, scope_stack)
                expanded = self._expand_singular_resource(resource_name, prefix)
                routes.extend(expanded)
                i += 1
                continue

            # Handle verb routes (get 'path', to: 'controller#action')
            verb_match = re.search(
                r"(?P<verb>get|post|put|patch|delete)\s+['\"](?P<path>[^\"']+)['\"]"
                r"\s*,\s*to:\s*['\"]?(?P<controller>\w+)#(?P<action>\w+)['\"]?",
                line,
                re.IGNORECASE,
            )
            if verb_match:
                prefix = self._build_prefix(namespace_stack, scope_stack)
                path = verb_match.group("path")
                if not path.startswith("/"):
                    path = "/" + path
                full_path = prefix + path
                verb = verb_match.group("verb").upper()
                controller = verb_match.group("controller")
                action = verb_match.group("action")
                routes.append({
                    "verb": verb,
                    "path": full_path,
                    "controller": controller,
                    "action": action,
                })

            # Check for block endings
            if line == "end" and (namespace_stack or scope_stack):
                if namespace_stack:
                    namespace_stack.pop()
                elif scope_stack:
                    scope_stack.pop()

            i += 1

        return routes

    def _build_prefix(self, namespace_stack: list[str], scope_stack: list[str]) -> str:
        """Build URL prefix from namespace and scope stacks."""
        parts = []
        for ns in namespace_stack:
            parts.append(f"/{ns}")
        for sc in scope_stack:
            if not sc.startswith("/"):
                sc = "/" + sc
            parts.append(sc)
        return "".join(parts) if parts else ""

    def _expand_resource_routes(self, resource: str, prefix: str = "") -> list[dict]:
        """Expand a resources declaration into individual RESTful routes."""
        routes = []
        base = f"/{resource}"

        for verb, suffix in _RESTFUL_ROUTES:
            path = prefix + base + suffix
            routes.append({
                "verb": verb,
                "path": path,
                "resource": resource,
            })

        return routes

    def _expand_singular_resource(self, resource: str, prefix: str = "") -> list[dict]:
        """Expand a singular resource declaration."""
        routes = []
        base = f"/{resource}"

        singular_routes = [
            ("GET", ""),        # show
            ("GET", "/new"),    # new
            ("POST", ""),       # create
            ("GET", "/edit"),   # edit
            ("PATCH", ""),      # update
            ("DELETE", ""),     # destroy
        ]

        for verb, suffix in singular_routes:
            path = prefix + base + suffix
            routes.append({
                "verb": verb,
                "path": path,
                "resource": resource,
            })

        return routes

    def _extract_controller_edges(self, content: str, from_file: str) -> None:
        """Extract controller references from routes.rb and create import edges."""
        controller_map: dict[str, str] = {}

        # Find all controller#action references
        for m in re.finditer(r"(?P<controller>\w+)#(?P<action>\w+)", content):
            ctrl = m.group("controller")
            # Map to controller file: UsersController -> app/controllers/users_controller.rb
            ctrl_file = _controller_to_file(ctrl)
            controller_map[ctrl] = ctrl_file

        # Create edges
        for ctrl, ctrl_file in controller_map.items():
            self._extra_imports.setdefault(from_file, []).append({
                "specifier": ctrl_file,
                "names": [ctrl],
            })

    def get_file_context(self, file_path: str) -> Optional[FileContext]:
        """Look up FileContext for a file."""
        if file_path in self._file_contexts:
            return self._file_contexts[file_path]
        return None

    def get_extra_imports(self) -> dict[str, list[dict]]:
        """Return controller import edges."""
        return self._extra_imports

    def get_metadata(self) -> dict:
        """Return metadata."""
        return {"rails_provider": {"route_files": len(self._file_contexts)}}

    def stats(self) -> dict:
        """Return provider stats."""
        return {"route_files": len(self._file_contexts)}


# ---------------------------------------------------------------------------
# Helper functions
# ---------------------------------------------------------------------------

def _to_underscore(name: str) -> str:
    """Convert CamelCase to snake_case."""
    import re
    return re.sub(r"(?<!^)(?=[A-Z])", "_", name).lower()


def _controller_to_file(controller: str) -> str:
    """Convert controller name to file path."""
    return f"app/controllers/{_to_underscore(controller)}_controller.rb"
