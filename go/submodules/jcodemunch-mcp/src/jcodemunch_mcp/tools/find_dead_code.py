"""Dead code detection — find files and symbols unreachable from any entry point."""

from __future__ import annotations

import fnmatch
import json
import re
import time
from typing import Optional

from ..storage import IndexStore
from ..parser.imports import resolve_specifier
from ._utils import index_status_to_tool_error, resolve_repo
from ..parser.context._route_utils import ENTRY_POINT_DECORATOR_RE


# ---------------------------------------------------------------------------
# Entry-point heuristics
# ---------------------------------------------------------------------------

_ENTRY_POINT_FILENAMES = frozenset({
    "__main__.py",
    "conftest.py",
    "manage.py",
    "wsgi.py",
    "asgi.py",
    "setup.py",
    "app.py",
    "main.py",
    "run.py",
    "cli.py",
    "celery.py",
    "Makefile",
})

_MAIN_GUARD_RE = re.compile(r'if\s+__name__\s*==\s*["\']__main__["\']')


def _is_entry_point_filename(file_path: str) -> bool:
    filename = file_path.replace("\\", "/").rsplit("/", 1)[-1]
    return filename in _ENTRY_POINT_FILENAMES


def _is_init_file(file_path: str) -> bool:
    filename = file_path.replace("\\", "/").rsplit("/", 1)[-1]
    return filename == "__init__.py"


def _is_test_file(file_path: str) -> bool:
    fp = file_path.replace("\\", "/")
    fn = fp.rsplit("/", 1)[-1]
    base = fn.rsplit(".", 1)[0] if "." in fn else fn
    return (
        "/tests/" in fp
        or "/test/" in fp
        or "/__tests__/" in fp
        or fn.startswith("test_")
        or fn.endswith("_test.py")
        or fn == "conftest.py"
        or base.endswith(".spec")    # foo.spec.ts, foo.spec.js
        or base.endswith(".test")    # foo.test.ts, foo.test.js
    )


def _matches_any_pattern(file_path: str, patterns: list[str]) -> bool:
    fp_fwd = file_path.replace("\\", "/")
    for pat in patterns:
        if fnmatch.fnmatch(fp_fwd, pat) or fnmatch.fnmatch(fp_fwd.rsplit("/", 1)[-1], pat):
            return True
    return False


def _package_json_entries(index, store, owner: str, repo_name: str) -> set[str]:
    """Return source files referenced by any ``package.json``'s ``main`` /
    ``module`` / ``exports`` / ``bin`` field. JS-library equivalent of the
    Python ``app.py``/``main.py`` filename heuristic. (Backported from
    get_dead_code_v2 in v1.80.8 — sverklo bench parity.)
    """
    entries: set[str] = set()
    source_files = frozenset(index.source_files)
    for f in index.source_files:
        fn = f.replace("\\", "/").rsplit("/", 1)[-1]
        if fn != "package.json":
            continue
        content = store.get_file_content(owner, repo_name, f)
        if not content:
            continue
        try:
            pkg = json.loads(content)
        except (ValueError, TypeError):
            continue
        if not isinstance(pkg, dict):
            continue
        candidates: list[str] = []
        for key in ("main", "module", "browser"):
            v = pkg.get(key)
            if isinstance(v, str):
                candidates.append(v)
        exports = pkg.get("exports")
        if isinstance(exports, str):
            candidates.append(exports)
        elif isinstance(exports, dict):
            def _walk_exports(node):
                if isinstance(node, str):
                    candidates.append(node)
                elif isinstance(node, dict):
                    for v in node.values():
                        _walk_exports(v)
            _walk_exports(exports)
        bins = pkg.get("bin")
        if isinstance(bins, str):
            candidates.append(bins)
        elif isinstance(bins, dict):
            candidates.extend(v for v in bins.values() if isinstance(v, str))

        pkg_dir = f.replace("\\", "/").rsplit("/", 1)[0] if "/" in f else ""
        for cand in candidates:
            cand = cand.lstrip("./").replace("\\", "/")
            joined = f"{pkg_dir}/{cand}" if pkg_dir else cand
            joined = joined.lstrip("/")
            if joined in source_files:
                entries.add(joined)
                continue
            for ext in ("", ".js", ".ts", ".mjs", ".cjs", ".jsx", ".tsx",
                        "/index.js", "/index.ts", "/index.mjs",
                        "/index.cjs"):
                trial = joined + ext
                if trial in source_files:
                    entries.add(trial)
                    break
    return entries


def _has_entry_point_decorator(sym: dict) -> bool:
    for dec in sym.get("decorators") or []:
        if ENTRY_POINT_DECORATOR_RE.search(str(dec)):
            return True
    return False


def _build_reverse_adjacency(
    imports: dict, source_files: frozenset, alias_map: Optional[dict] = None,
    psr4_map: Optional[dict] = None,
) -> dict[str, list[str]]:
    """Return {file: [files_that_import_it]} from raw import data."""
    rev: dict[str, list[str]] = {}
    for src_file, file_imports in imports.items():
        for imp in file_imports:
            target = resolve_specifier(imp["specifier"], src_file, source_files, alias_map, psr4_map)
            if target and target != src_file:
                rev.setdefault(target, []).append(src_file)
    return {k: list(dict.fromkeys(v)) for k, v in rev.items()}


# ---------------------------------------------------------------------------
# Public tool function
# ---------------------------------------------------------------------------

def find_dead_code(
    repo: str,
    granularity: str = "symbol",
    min_confidence: float = 0.8,
    include_tests: bool = False,
    entry_point_patterns: Optional[list[str]] = None,
    storage_path: Optional[str] = None,
) -> dict:
    """Find dead code — files and symbols with no importers and no entry-point role.

    Args:
        repo: Repository identifier (owner/repo or bare name).
        granularity: "symbol" returns dead symbols (default); "file" returns dead files only.
        min_confidence: Minimum confidence threshold (0.0–1.0). Default 0.8.
        include_tests: Treat test files as live roots (default false).
        entry_point_patterns: Additional glob patterns to treat as live roots.
        storage_path: Custom storage path.
    """
    start = time.perf_counter()
    entry_point_patterns = entry_point_patterns or []

    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return {"error": str(e)}

    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name)
    if not index:
        return index_status_to_tool_error(store.inspect_index(owner, name))

    if index.imports is None:
        return {
            "error": (
                "No import data available. Re-index with jcodemunch-mcp >= 1.3.0 "
                "to enable dead code analysis."
            )
        }

    source_files = frozenset(index.source_files)
    rev = _build_reverse_adjacency(index.imports, source_files, index.alias_map, getattr(index, "psr4_map", None))

    # -----------------------------------------------------------------------
    # Phase 1: identify live roots by filename pattern + package.json (no I/O
    # for the filename pass; package.json parsing is bounded by # of manifests).
    # -----------------------------------------------------------------------
    live_roots: set[str] = set()
    for f in index.source_files:
        if _is_entry_point_filename(f):
            live_roots.add(f)
        elif _is_init_file(f):
            live_roots.add(f)
        elif include_tests and _is_test_file(f):
            live_roots.add(f)
        elif entry_point_patterns and _matches_any_pattern(f, entry_point_patterns):
            live_roots.add(f)
    # JS-library entry points: whatever package.json declares as main/module/
    # exports/bin. Without this, library files like Express's lib/express.js
    # had only `index.js` as importer, that index had no further importers,
    # so the file was misclassified as `all_importers_dead` at confidence 0.7.
    pkg_entries = _package_json_entries(index, store, owner, name)
    live_roots.update(pkg_entries)

    # -----------------------------------------------------------------------
    # Phase 2: content check for `if __name__ == "__main__"` (Python only,
    # only for files not yet classified as live and with zero importers)
    # -----------------------------------------------------------------------
    for f in index.source_files:
        if f in live_roots or rev.get(f):
            continue
        if not (f.endswith(".py") or f.endswith(".pyw")):
            continue
        content = store.get_file_content(owner, name, f)
        if content and _MAIN_GUARD_RE.search(content):
            live_roots.add(f)

    # -----------------------------------------------------------------------
    # Phase 3: classify files
    # -----------------------------------------------------------------------
    # Pre-compute which files have only dead importers (for cascading 0.7 case)
    # A file's importers are "all dead" when each importer has zero importers
    # of its own and is not a live root — simple one-hop check, avoids deep BFS.

    dead_files: list[dict] = []

    for f in sorted(index.source_files):
        if f in live_roots:
            continue
        if not include_tests and _is_test_file(f):
            continue

        importers = rev.get(f, [])

        if not importers:
            confidence = 0.9 if _is_test_file(f) else 1.0
            reason = "zero_importers"
        else:
            # Check for cascading dead code: all importers themselves have
            # zero importers and are not live roots
            all_dead = all(
                not rev.get(imp) and imp not in live_roots
                for imp in importers
            )
            if all_dead:
                confidence = 0.7
                reason = "all_importers_dead"
            else:
                continue  # file is reachable, skip

        if confidence < min_confidence:
            continue

        dead_files.append({
            "file": f,
            "confidence": confidence,
            "reason": reason,
            "importer_count": len(importers),
        })

    # -----------------------------------------------------------------------
    # Phase 4: symbol-level results
    # -----------------------------------------------------------------------
    dead_symbols: list[dict] = []

    if granularity == "symbol":
        dead_file_map = {d["file"]: d for d in dead_files}

        for sym in index.symbols:
            sym_file = sym.get("file", "")
            if sym_file not in dead_file_map:
                continue

            file_entry = dead_file_map[sym_file]
            confidence = file_entry["confidence"]
            reason = file_entry["reason"]

            # Framework decorators lower confidence
            if _has_entry_point_decorator(sym):
                confidence = min(confidence, 0.5)
                reason = "framework_decorator"

            if confidence < min_confidence:
                continue

            dead_symbols.append({
                "symbol_id": sym.get("id", ""),
                "file": sym_file,
                "kind": sym.get("kind", ""),
                "confidence": confidence,
                "reason": reason,
            })

    # -----------------------------------------------------------------------
    # Response
    # -----------------------------------------------------------------------
    elapsed = (time.perf_counter() - start) * 1000

    sample_roots = sorted(live_roots)[:5]
    analysis_notes = [
        f"Entry points detected: {len(live_roots)}",
        f"Total files analyzed: {len(index.source_files)}",
    ]
    if sample_roots:
        analysis_notes.append(f"Sample entry points: {', '.join(sample_roots)}")

    result: dict = {
        "repo": f"{owner}/{name}",
        "granularity": granularity,
        "min_confidence": min_confidence,
        "dead_symbols": dead_symbols,
        "dead_files": dead_files,
        "dead_file_count": len(dead_files),
        "dead_symbol_count": len(dead_symbols),
        "live_root_count": len(live_roots),
        "analysis_notes": analysis_notes,
        "_meta": {"timing_ms": round(elapsed, 1)},
    }
    return result
