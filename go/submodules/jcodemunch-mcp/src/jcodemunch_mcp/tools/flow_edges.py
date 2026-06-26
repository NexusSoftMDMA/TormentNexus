"""flow_edges — language-agnostic framework flow-edge resolver.

The AST call graph is structurally blind to the two most common ways a web
framework wires a request to the code that handles it:

  * **route -> handler** — a route registration binds a URL string to a handler
    that is *referenced*, never called or decorated, so no call edge exists.
    Examples (one shape, many stacks)::

        # Django urls.py
        path("users/", views.list_users)
        re_path(r"^u/(?P<pk>\\d+)$", views.UserView.as_view())
        # Express / Fastify / Koa
        router.get("/users", listUsers)
        # Flask
        app.add_url_rule("/users", view_func=list_users)
        # Rails routes.rb
        get "/users", to: "users#index"

    Decorator-bound handlers (``@app.route`` / ``@GetMapping`` / ``@Get()``)
    already surface as gateways in :func:`get_signal_chains`, so they are
    deliberately *not* re-emitted here — this resolver fills the dispatch gap
    the call graph cannot see.

  * **render -> view** — a handler renders a template named by a string literal
    (``render(request, "page.html")``, ``render_template("x.html")``,
    ``res.render("index")``, ``return view("home")``). The template is a string,
    invisible to the call graph; this edge connects the rendering symbol to the
    template (and, when the template file is itself indexed, to that file).

This is **one shape-keyed resolver, not a plugin per framework**: detection keys
on the structural shape of the registration/render call, and resolution reuses
the index's own symbol table + import edges (the same machinery
``find_direct_callees`` uses). It is a pure read path — no reindex, no persisted
state, nothing written to the index.

The emitted edges are consumed by :func:`get_signal_chains` (string-dispatched
handlers become gateways; rendered templates attach as a ``views`` annotation),
and the resolver is reusable by any caller that wants typed request-flow edges.
"""

from __future__ import annotations

import logging
import re
from typing import Optional

from ..parser.imports import resolve_specifier
from ._call_graph import _ContentCache, _symbol_body, build_symbols_by_file

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Structural shape detectors (route registration → verb, path, handler ref)
# ---------------------------------------------------------------------------

# Express / Fastify / Koa: `router.get("/x", handler)` — handler is a bare
# reference (followed by `,` for more middleware, or `)`). Inline functions are
# excluded by the trailing-char anchor + the keyword skip-set below.
_RE_EXPRESS = re.compile(
    r"\b(?:app|router|api|server|route|routes|r|fastify|koa)\s*\.\s*"
    r"(?P<verb>get|post|put|delete|patch|head|options|all)\s*\(\s*"
    r"(?P<q>['\"`])(?P<path>[^'\"`]+)(?P=q)\s*,\s*"
    r"(?P<handler>[A-Za-z_$][\w.$]*)\s*[),]",
    re.IGNORECASE,
)

# Django urls.py: `path("x/", views.foo)` / `re_path(r"^x$", Foo.as_view())`.
# Case-sensitive so handler/View names are preserved verbatim (the dispatcher
# function names path/re_path/url are lowercase regardless).
_RE_DJANGO = re.compile(
    r"\b(?:re_path|path|url)\s*\(\s*"
    r"r?(?P<q>['\"])(?P<path>[^'\"]*)(?P=q)\s*,\s*"
    r"(?P<handler>[A-Za-z_][\w.]*)",
)

# Flask: `app.add_url_rule("/x", view_func=foo)` (endpoint kwarg may sit between).
_RE_FLASK_ADD = re.compile(
    r"\.add_url_rule\s*\(\s*(?P<q>['\"])(?P<path>[^'\"]*)(?P=q)"
    r"[^)]*?view_func\s*=\s*(?P<handler>[A-Za-z_][\w.]*)",
)

# Rails routes.rb: `get "/x", to: "users#index"`.
_RE_RAILS = re.compile(
    r"\b(?P<verb>get|post|put|patch|delete)\s+"
    r"(?P<q>['\"])(?P<path>[^'\"]+)(?P=q)\s*,\s*to:\s*"
    r"(?P<q2>['\"])(?P<controller>[^'\"#]+)#(?P<action>[A-Za-z_]\w*)(?P=q2)",
)

# (regex, shape-name, implicit-verb) — implicit verb used when the regex has no
# `verb` group (Django path() carries no HTTP method).
_ROUTE_DETECTORS = [
    (_RE_EXPRESS, "express", None),
    (_RE_DJANGO, "django", "PATH"),
    (_RE_FLASK_ADD, "flask_add_url", "ANY"),
    (_RE_RAILS, "rails", None),
]

# Cheap substring pre-gate: skip the regex pass on files that can't contain a
# route registration. Lowercased contains-any.
_ROUTE_GATE = (
    ".get(", ".post(", ".put(", ".delete(", ".patch(", ".all(",
    "path(", "re_path(", "url(", "add_url_rule(", ", to:",
)

# Handler tokens that are language keywords / framework objects, not a symbol.
_HANDLER_SKIP = frozenset({
    "function", "async", "await", "require", "express", "lambda",
    "self", "cls", "this", "new", "render", "redirect", "next",
})


# ---------------------------------------------------------------------------
# render -> view detection
# ---------------------------------------------------------------------------

# The render-family call openings we recognise. We capture the *first* string
# literal in the call args, which is the template name across all of these
# (Django render()'s first literal is the template — `request` is not a literal).
_RE_RENDER_CALL = re.compile(
    r"\b(?:render_template_string|render_template|render_to_response"
    r"|res\s*\.\s*render|render|view)\s*\(",
)
_RE_FIRST_STRING = re.compile(r"['\"]([^'\"]+)['\"]")

# Template-ish extensions used to (a) keep bare `render(` from matching React/DOM
# render calls whose first literal isn't a template, and (b) resolve a template
# string to an indexed file.
_TEMPLATE_EXTS = (
    ".html", ".htm", ".jinja", ".jinja2", ".j2", ".twig", ".ejs", ".erb",
    ".hbs", ".handlebars", ".mustache", ".pug", ".jade", ".njk", ".liquid",
    ".blade.php", ".vue", ".svelte", ".astro", ".tpl", ".tmpl", ".haml", ".slim",
)

_RENDER_GATE = ("render", "view(")


# ---------------------------------------------------------------------------
# Resolution helpers
# ---------------------------------------------------------------------------

def _candidate_files(index, sym_file: str) -> set[str]:
    """Files whose symbols a reference in *sym_file* could resolve to.

    The declaring file's resolved imports plus the file itself — exactly the
    reachable set ``find_direct_callees`` uses for text-matched callees.
    """
    source_files_fs = frozenset(index.source_files)
    alias_map = getattr(index, "alias_map", {}) or {}
    psr4_map = getattr(index, "psr4_map", None)
    files: set[str] = {sym_file}
    for imp in (index.imports or {}).get(sym_file, []):
        try:
            target = resolve_specifier(
                imp.get("specifier", ""), sym_file, source_files_fs, alias_map, psr4_map
            )
        except Exception:  # pragma: no cover - defensive; resolver is best-effort
            logger.debug("resolve_specifier failed for %r", imp, exc_info=True)
            target = None
        if target:
            files.add(target)
    return files


def _resolve_handler(
    index,
    symbols_by_file: dict[str, list[dict]],
    decl_file: str,
    handler_ref: str,
) -> Optional[dict]:
    """Resolve a route's handler reference to a symbol dict, or None.

    ``handler_ref`` may be dotted (``views.foo``, ``views.UserView.as_view``,
    ``UsersController#index`` already split by the caller). Resolution prefers a
    symbol whose file's stem matches the dotted module prefix (``views`` ->
    ``.../views.py``), then any candidate file, then a repo-wide name fallback.
    """
    # `views.UserView.as_view` -> drop trailing `as_view`; `views.foo` -> `foo`.
    parts = [p for p in handler_ref.split(".") if p]
    if not parts:
        return None
    if parts[-1] in ("as_view", "as_view()") and len(parts) >= 2:
        parts = parts[:-1]
    name = parts[-1]
    if not name or name in _HANDLER_SKIP:
        return None
    module_hint = parts[-2] if len(parts) >= 2 else None

    candidates = _candidate_files(index, decl_file)

    def _match_in(files: set[str], require_hint: bool) -> Optional[dict]:
        best: Optional[dict] = None
        for f in files:
            if require_hint and module_hint:
                stem = f.replace("\\", "/").rsplit("/", 1)[-1].rsplit(".", 1)[0]
                if stem != module_hint and module_hint not in f.replace("\\", "/").split("/"):
                    continue
            for s in symbols_by_file.get(f, []):
                if s.get("name") != name:
                    continue
                if s.get("kind") not in ("function", "method", "class"):
                    continue
                # Prefer a class for `Foo.as_view`, else first function/method.
                if best is None or s.get("kind") == "class":
                    best = s
        return best

    # 1) candidate file whose module name matches the dotted hint
    hit = _match_in(candidates, require_hint=True) if module_hint else None
    # 2) any candidate (imports + own file)
    if hit is None:
        hit = _match_in(candidates, require_hint=False)
    # 3) repo-wide unique name fallback (handles re-exports the import graph missed)
    if hit is None:
        repo_hits = [
            s for syms in symbols_by_file.values() for s in syms
            if s.get("name") == name and s.get("kind") in ("function", "method", "class")
        ]
        if len(repo_hits) == 1:
            hit = repo_hits[0]
    return hit


def _resolve_template(index, template: str) -> Optional[str]:
    """Resolve a template string to an indexed source file, or None.

    Templates are frequently not indexed (jcm indexes code), so an unresolved
    return is normal and the edge is still emitted with the raw name.
    """
    norm = template.replace("\\", "/").lstrip("/")
    norm_low = norm.lower()
    best: Optional[str] = None
    for f in index.source_files:
        fl = f.replace("\\", "/").lower()
        if fl == norm_low or fl.endswith("/" + norm_low):
            return f  # exact / suffix path match wins outright
        if best is None and "/" not in norm_low:
            # bare name like `index` or `home.html`: match by basename/stem
            base = fl.rsplit("/", 1)[-1]
            if base == norm_low or base.rsplit(".", 1)[0] == norm_low.rsplit(".", 1)[0]:
                best = f
    return best


def _looks_like_template(value: str) -> bool:
    low = value.lower()
    if any(low.endswith(ext) for ext in _TEMPLATE_EXTS):
        return True
    # bare names (no extension) are accepted only when they look like a view id
    # (a path segment or a plain identifier), never arbitrary prose.
    if "." in low:
        return False
    return bool(re.fullmatch(r"[\w][\w/.\-]*", value))


# ---------------------------------------------------------------------------
# Public resolver
# ---------------------------------------------------------------------------

def resolve_flow_edges(
    index,
    store,
    owner: str,
    repo_name: str,
    *,
    kinds: tuple[str, ...] = ("route", "render"),
    symbols_by_file: Optional[dict[str, list[dict]]] = None,
    content_cache: Optional["_ContentCache"] = None,
) -> list[dict]:
    """Resolve typed framework flow edges over the loaded *index*.

    Args:
        index:           A loaded ``CodeIndex``.
        store:           ``IndexStore`` for file-content reads.
        owner, repo_name: Resolved repo identity (for content reads).
        kinds:           Which edge families to resolve: ``"route"`` and/or
                         ``"render"``.
        symbols_by_file: Optional prebuilt ``{file: [symbols]}`` to reuse.
        content_cache:   Optional ``_ContentCache`` to share file reads.

    Returns:
        List of edge dicts. Each has ``type`` (``"route->handler"`` /
        ``"render->view"``), ``label``, ``framework_shape``, ``confidence``,
        ``resolution`` (``"resolved"`` / ``"unresolved"``), the source site
        (``src_file``/``line``), and target fields (``dst_id``/``dst_name``/
        ``dst_file``, plus ``verb``/``path`` for routes).
    """
    if symbols_by_file is None:
        symbols_by_file = build_symbols_by_file(index)
    if content_cache is None:
        content_cache = _ContentCache(store, owner, repo_name)

    edges: list[dict] = []
    if "route" in kinds:
        edges.extend(_resolve_routes(index, symbols_by_file, content_cache))
    if "render" in kinds:
        edges.extend(_resolve_renders(index, symbols_by_file, content_cache))
    return edges


def _resolve_routes(index, symbols_by_file, content_cache) -> list[dict]:
    edges: list[dict] = []
    # Symbol ids already reachable as decorator gateways are skipped: those
    # handlers surface on their own; we only want the dispatch the graph misses.
    for decl_file in index.source_files:
        content = content_cache.content(decl_file)
        if not content:
            continue
        low = content.lower()
        if not any(tok in low for tok in _ROUTE_GATE):
            continue
        for regex, shape, implicit_verb in _ROUTE_DETECTORS:
            for m in regex.finditer(content):
                gd = m.groupdict()
                if shape == "rails":
                    handler_ref = f"{gd['controller']}.{gd['action']}"
                else:
                    handler_ref = gd.get("handler", "")
                if not handler_ref:
                    continue
                head = handler_ref.split(".")[0].lower()
                if head in _HANDLER_SKIP:
                    continue
                target = _resolve_handler(index, symbols_by_file, decl_file, handler_ref)
                verb = (gd.get("verb") or implicit_verb or "ANY").upper()
                path = gd.get("path", "")
                line = content.count("\n", 0, m.start()) + 1
                label = f"{verb} {path}".strip()
                if target is None:
                    edges.append({
                        "type": "route->handler",
                        "src_file": decl_file,
                        "line": line,
                        "dst_id": None,
                        "dst_name": handler_ref,
                        "dst_file": None,
                        "verb": verb,
                        "path": path,
                        "framework_shape": shape,
                        "label": f"{label} -> {handler_ref}",
                        "evidence": m.group(0).strip()[:160],
                        "confidence": 0.3,
                        "resolution": "unresolved",
                    })
                    continue
                same_file = target.get("file") == decl_file
                edges.append({
                    "type": "route->handler",
                    "src_file": decl_file,
                    "line": line,
                    "dst_id": target.get("id"),
                    "dst_name": target.get("name"),
                    "dst_file": target.get("file"),
                    "verb": verb,
                    "path": path,
                    "framework_shape": shape,
                    "label": f"{label} -> {target.get('name')}",
                    "evidence": m.group(0).strip()[:160],
                    "confidence": 0.7 if same_file else 0.8,
                    "resolution": "resolved",
                })
    return edges


def _resolve_renders(index, symbols_by_file, content_cache) -> list[dict]:
    edges: list[dict] = []
    for sym in index.symbols:
        if sym.get("kind") not in ("function", "method"):
            continue
        sym_file = sym.get("file", "")
        if not sym_file:
            continue
        body = _symbol_body(content_cache.lines(sym_file), sym)
        if not body or not any(tok in body.lower() for tok in _RENDER_GATE):
            continue
        for cm in _RE_RENDER_CALL.finditer(body):
            window = body[cm.end():cm.end() + 200]
            sm = _RE_FIRST_STRING.search(window)
            if not sm:
                continue
            template = sm.group(1)
            if not _looks_like_template(template):
                continue
            dst_file = _resolve_template(index, template)
            edges.append({
                "type": "render->view",
                "src_id": sym.get("id"),
                "src_name": sym.get("name"),
                "src_file": sym_file,
                "line": sym.get("line", 0),
                "dst_id": None,
                "dst_name": template,
                "dst_file": dst_file,
                "framework_shape": "render",
                "label": f"{sym.get('name')} -> {template}",
                "evidence": cm.group(0).strip() + window[:sm.end()].strip(),
                "confidence": 0.7 if dst_file else 0.4,
                "resolution": "resolved" if dst_file else "unresolved",
            })
    return edges
