"""get_call_hierarchy: callers and callees for any indexed symbol, N levels deep."""

import time
from typing import Optional

from ..storage import IndexStore
from ._utils import index_status_to_tool_error, resolve_repo
from .get_blast_radius import _build_reverse_adjacency, _find_symbol
from ._call_graph import build_symbols_by_file, bfs_callers, bfs_callees


def get_call_hierarchy(
    repo: str,
    symbol_id: str,
    direction: str = "both",
    depth: int = 3,
    storage_path: Optional[str] = None,
) -> dict:
    """Return incoming callers and outgoing callees for a symbol, N levels deep.

    Uses AST-derived call detection — no LSP required. Callers are found by
    scanning symbols in files that import the target's module; callees are found
    by matching imported-symbol names against the target's source body.

    Args:
        repo: Repository identifier (owner/repo or just repo name).
        symbol_id: Symbol name or full ID to analyse. Use search_symbols to find IDs.
        direction: 'callers' | 'callees' | 'both'. Default 'both'.
        depth: Maximum hops to traverse (1–5). Default 3.
        storage_path: Custom storage path.

    Returns:
        Dict with symbol info, callers list, callees list, depth_reached, and _meta.
        Each caller/callee entry includes {id, name, kind, file, line, depth}.
    """
    depth = max(1, min(depth, 5))
    if direction not in ("callers", "callees", "both"):
        direction = "both"
    start = time.perf_counter()

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
                "to enable call hierarchy analysis."
            )
        }

    matches = _find_symbol(index, symbol_id)
    if not matches:
        return {"error": f"Symbol not found: '{symbol_id}'. Try search_symbols first."}
    if len(matches) > 1:
        ambiguous = [{"name": s["name"], "file": s["file"], "id": s["id"]} for s in matches]
        return {
            "error": (
                f"Ambiguous symbol '{symbol_id}': found {len(matches)} definitions. "
                "Use the symbol 'id' field to disambiguate."
            ),
            "candidates": ambiguous,
        }

    sym = matches[0]
    symbols_by_file = build_symbols_by_file(index)
    reverse_adj = _build_reverse_adjacency(
        index.imports,
        frozenset(index.source_files),
        getattr(index, "alias_map", None),
        getattr(index, "psr4_map", None),
    )

    callers: list[dict] = []
    callees: list[dict] = []
    depth_reached = 0

    if direction in ("callers", "both"):
        callers, dr = bfs_callers(
            index, store, owner, name, sym, reverse_adj, symbols_by_file, depth
        )
        depth_reached = max(depth_reached, dr)

    if direction in ("callees", "both"):
        callees, dr = bfs_callees(
            index, store, owner, name, sym, symbols_by_file, depth
        )
        depth_reached = max(depth_reached, dr)

    elapsed = (time.perf_counter() - start) * 1000

    # Build dispatches section from dispatch edges
    ctx_meta = getattr(index, "context_metadata", None) or {}
    dispatch_edge_data = ctx_meta.get("dispatch_edges", [])
    dispatches: list[dict] = []
    if dispatch_edge_data:
        # Group by (interface_name, method_name)
        grouped: dict[tuple[str, str], list[dict]] = {}
        for de in dispatch_edge_data:
            key = (de.get("interface_name", ""), de.get("method_name", ""))
            grouped.setdefault(key, []).append(de)
        for (iface, method), impls in grouped.items():
            dispatches.append({
                "interface": iface,
                "method": method,
                "implementations": [
                    {
                        "name": imp.get("impl_name", ""),
                        "file": imp.get("impl_file", ""),
                        "line": imp.get("impl_line", 0),
                    }
                    for imp in impls
                ],
            })

    # Determine methodology based on available data
    get_callers = getattr(index, "get_callers_by_name", None)
    callers_by_name = get_callers() if get_callers else None
    has_call_data = bool(callers_by_name)
    has_lsp_data = bool(ctx_meta.get("lsp_edges"))
    has_dispatch_data = bool(dispatch_edge_data)
    if has_dispatch_data:
        methodology = "lsp_dispatch_enriched"
        confidence = "high"
        source = "lsp_bridge + dispatch_resolution + ast_call_references"
        tip = (
            "LSP dispatch-enriched: compiler-grade resolution via language servers with "
            "interface/trait dispatch resolution — concrete implementations of interface "
            "methods are resolved via textDocument/implementation. Each edge has a "
            "'resolution' field: lsp_dispatch (interface dispatch), lsp_resolved "
            "(compiler-grade), ast_resolved (direct AST), ast_inferred (import graph), "
            "or text_matched (heuristic)."
        )
    elif has_lsp_data:
        methodology = "lsp_enriched"
        confidence = "high"
        source = "lsp_bridge + ast_call_references"
        tip = (
            "LSP-enriched: compiler-grade resolution via language servers (pyright, gopls, "
            "typescript-language-server, rust-analyzer) for highest confidence, with AST "
            "call_references and text heuristic as fallback layers. Each edge has a "
            "'resolution' field: lsp_resolved (compiler-grade), ast_resolved (direct AST), "
            "ast_inferred (import graph), or text_matched (heuristic)."
        )
    elif has_call_data:
        methodology = "ast_call_references"
        confidence = "medium"
        source = "ast_call_references"
        tip = (
            "AST-based: call references extracted from tree-sitter AST during indexing. "
            "More precise than text heuristic, but still approximate for dynamic dispatch. "
            "Each edge has a 'resolution' field: ast_resolved (direct AST match), "
            "ast_inferred (resolved via import graph), or text_matched (heuristic). "
            "Enable LSP enrichment for compiler-grade resolution."
        )
    else:
        methodology = "text_heuristic"
        confidence = "low"
        source = "text_heuristic"
        tip = (
            "Text-heuristic: callers = symbols in importing files that mention this "
            "name as a word token; callees = imported symbols mentioned in this "
            "symbol's body. May have false positives for common names or dynamic "
            "dispatch. Use get_impact_preview for a transitive 'what breaks?' view."
        )

    # Summarize resolution tiers across all edges
    resolution_counts: dict[str, int] = {}
    for edge in callers + callees:
        r = edge.get("resolution", "unknown")
        resolution_counts[r] = resolution_counts.get(r, 0) + 1

    response: dict = {
        "repo": f"{owner}/{name}",
        "symbol": {
            "id": sym.get("id", ""),
            "name": sym.get("name", ""),
            "kind": sym.get("kind", ""),
            "file": sym.get("file", ""),
            "line": sym.get("line", 0),
        },
        "direction": direction,
        "depth": depth,
        "depth_reached": depth_reached,
        "caller_count": len(callers),
        "callee_count": len(callees),
        "callers": callers,
        "callees": callees,
        "dispatches": dispatches,
        "_meta": {
            "timing_ms": round(elapsed, 1),
            "methodology": methodology,
            "confidence_level": confidence,
            "source": source,
            "resolution_tiers": resolution_counts,
            "tip": tip,
        },
    }

    # Phase 2: runtime confidence — zero-cost no-op when no traces ingested.
    from ..runtime.confidence import attach_runtime_confidence as _attach_runtime
    _db_path_str = str(store._sqlite._db_path(owner, name))
    _stamped: list[dict] = [response["symbol"], *callers, *callees]
    _summary = _attach_runtime(_stamped, _db_path_str, id_field="id")
    if _summary:
        response["_meta"]["runtime_freshness"] = _summary
    return response
