"""The Counter: an adaptive tool surface for jcodemunch-mcp.

Problem this solves
-------------------
jcm exposes ~83 MCP tools. The host serializes every resident tool's schema
into the model's context on every turn (a fixed per-turn token tax), and the
model must select one tool out of ~83 (dispatch dilution). Both costs scale
with tool count and work against jcm's own token-efficiency thesis.

The Counter is a small, stable front door that fronts the full catalog without
removing any capability:

  * ``order(action, args)`` -- single dispatch verb. Re-enters the normal
    tool pipeline for the chosen action. Read-only by default at the boundary:
    state-changing actions require an explicit opt-in, and exec/file-write
    verbs are refused unconditionally (a forward-looking charter tripwire --
    jcm ships none today, and the Counter must never become the surface that
    introduces one).
  * ``menu(query, tier)`` -- discovery. Search/browse the action catalog and
    return compact entries, so all ~83 schemas need not stay resident.
  * ``route(task, execute)`` -- intent to action. Map a natural-language task
    to the best catalog action(s); optionally dispatch the top one. Composes
    with ``assemble_task_context`` / ``plan_turn`` (it recommends them for
    context-gathering intents); it does not replace them.

This module is pure logic with no server import (keeps the dependency one-way:
server.py imports counter, never the reverse). server.py owns the Tool
registration, the live catalog, and call_tool re-dispatch; it hands plain data
to the helpers here.
"""

from __future__ import annotations

import re
from typing import Iterable, Optional

# The front-door tool names. These are never themselves dispatchable via
# ``order`` (no front-door recursion).
FRONT_DOOR: frozenset[str] = frozenset({"order", "menu", "route"})

# Actions that change persistent index / embedding / session / config state.
# These are charter-safe (none write the user's source files or execute code),
# but ``order`` requires an explicit ``allow_state_change=true`` before
# dispatching one, so the front door reads as read-only by default.
STATE_CHANGING_ACTIONS: frozenset[str] = frozenset({
    "index_repo", "index_folder", "index_file",
    "invalidate_cache", "register_edit", "tune_weights",
    "set_tool_tier", "announce_model", "embed_repo",
    "import_runtime_signal", "summarize_repo",
})

# Forward-looking tripwire. ``order`` refuses to dispatch any action whose name
# matches one of these verbs, even if such a tool were somehow added to the
# catalog later. jcm is read-only by charter and ships none of these; the gate
# exists so the consolidation layer can never silently become an exec/mutation
# backdoor (the line write-enabled competitors cross). This is the "safety
# surface" property: the dispatcher is a charter checkpoint, not just ergonomics.
_FORBIDDEN_VERB_RE = re.compile(
    r"(^|[._-])(exec|shell|run_command|spawn|eval|"
    r"write_file|edit_file|patch|apply_patch|delete_file|rm|mv|chmod)($|[._-])",
    re.IGNORECASE,
)


def is_state_changing(action: str) -> bool:
    return action in STATE_CHANGING_ACTIONS


def forbidden_reason(action: str) -> Optional[str]:
    """Return a rejection reason if *action* matches the exec/write tripwire."""
    if _FORBIDDEN_VERB_RE.search(action or ""):
        return (
            f"'{action}' names a write/exec verb. The Counter is a read-only "
            f"dispatch surface by charter and refuses to route execution or "
            f"file-mutation actions."
        )
    return None


def order_gate(
    action: str,
    catalog_names: Iterable[str],
    allow_state_change: bool,
) -> Optional[str]:
    """Validate an ``order`` request. Return an error string, or None if OK.

    Order of checks matters: structural (front door / unknown) before charter
    (tripwire) before policy (state-change opt-in), so the message an agent
    sees is the most actionable one.
    """
    if not action or not isinstance(action, str):
        return "order requires a non-empty 'action' name. Call 'menu' to list actions."
    if action in FRONT_DOOR:
        return f"'{action}' is a front-door tool and cannot be dispatched through order."
    names = set(catalog_names)
    if action not in names:
        return (
            f"Unknown action '{action}'. Call 'menu' (optionally with a query) "
            f"to discover valid actions."
        )
    tripwire = forbidden_reason(action)
    if tripwire is not None:
        return tripwire
    if is_state_changing(action) and not allow_state_change:
        return (
            f"'{action}' changes index/session state. Re-issue with "
            f"allow_state_change=true to proceed. (Read-only actions need no opt-in.)"
        )
    return None


# --- menu: catalog search -------------------------------------------------- #

_WORD_RE = re.compile(r"[a-z0-9]+")


def _tokens(text: str) -> list[str]:
    return _WORD_RE.findall((text or "").lower())


def _first_sentence(desc: str, limit: int = 160) -> str:
    desc = (desc or "").strip().replace("\n", " ")
    # Cut at the first sentence boundary, else hard-truncate.
    m = re.search(r"(?<=[.!?])\s", desc)
    out = desc[: m.start() + 1] if m else desc
    if len(out) > limit:
        out = out[: limit - 1].rstrip() + "…"
    return out


def _required_args(schema: dict) -> list[str]:
    if not isinstance(schema, dict):
        return []
    req = schema.get("required")
    return list(req) if isinstance(req, list) else []


def catalog_entry(name: str, description: str, schema: dict) -> dict:
    """Compact, dense menu row for one action."""
    return {
        "action": name,
        "summary": _first_sentence(description),
        "required": _required_args(schema),
        "state_changing": is_state_changing(name),
    }


def score_action(
    query_tokens: list[str],
    name: str,
    description: str,
    weights: Optional[dict[str, float]] = None,
) -> float:
    """Heuristic relevance of an action to a query. Higher is better.

    Name hits dominate (an agent usually has a verb in mind); description word
    overlap breaks ties. Each query token is scaled by its idf ``weights`` so a
    rare, discriminating term ("calls") outranks a ubiquitous one ("symbol").
    Deterministic, no embeddings -- in the jMRI idiom.
    """
    if not query_tokens:
        return 0.0
    name_l = name.lower()
    name_toks = set(_tokens(name))
    desc_toks = set(_tokens(description))
    score = 0.0
    for qt in query_tokens:
        w = weights.get(qt, 1.0) if weights else 1.0
        if qt == name_l:
            score += 10.0 * w
        elif qt in name_l:
            score += 4.0 * w
        elif qt in name_toks:
            score += 3.0 * w
        if qt in desc_toks:
            score += 1.0 * w
    return score


def _idf_weights(query_tokens: list[str], rows: list[dict]) -> dict[str, float]:
    """Inverse document frequency of each query token across the catalog
    (name + description). Rare tokens weigh more; tokens in every row weigh ~0.
    """
    import math
    n = max(1, len(rows))
    docs = [set(_tokens(r["action"])) | set(_tokens(r.get("_description", r.get("summary", "")))) for r in rows]
    weights: dict[str, float] = {}
    for qt in set(query_tokens):
        df = sum(1 for d in docs if qt in d)
        # +1 smoothing; floor at a small positive so a common term still counts.
        weights[qt] = max(0.15, math.log((n + 1) / (df + 1)) + 0.3)
    return weights


def search_catalog(
    catalog: list[dict],
    query: Optional[str],
    limit: int,
) -> list[dict]:
    """Rank/filter catalog rows for *query*. ``catalog`` rows are
    ``{"action", "summary", "required", "state_changing", "_description"}``.
    With no query, return the catalog in stable order (capped at limit).
    """
    rows = [r for r in catalog if r["action"] not in FRONT_DOOR]
    if not query:
        return rows[:limit]
    qt = _tokens(query)
    weights = _idf_weights(qt, rows)
    scored = []
    for r in rows:
        s = score_action(qt, r["action"], r.get("_description", r["summary"]), weights)
        if s > 0:
            scored.append((s, r))
    scored.sort(key=lambda x: (-x[0], x[1]["action"]))
    return [r for _, r in scored[:limit]]


# --- route: intent to action ----------------------------------------------- #

# Ordered intent rules. First match whose pattern hits the task wins as the
# primary recommendation; remaining matches become alternates. Each rule is
# (compiled_pattern, action, why). Kept deterministic and legible -- this is a
# curated map, not a learned model, consistent with the read-only charter.
_INTENT_RULES: list[tuple[re.Pattern, str, str]] = [
    (re.compile(r"\b(who )?calls?\b|\bcallers?\b|\bcall(ed)? by\b|\bcall (graph|hierarchy)\b", re.I),
     "get_call_hierarchy", "Trace callers/callees of a symbol."),
    (re.compile(r"\bused? (by|where)\b|\breferences?\b|\bwhere is .* used\b", re.I),
     "find_references", "Find where an identifier is referenced."),
    (re.compile(r"\b(blast|impact|break|breaks?|affect|ripple|what changes)\b", re.I),
     "get_blast_radius", "Show what a change to a symbol would affect."),
    (re.compile(r"\bdead code\b|\bunused\b|\bunreachable\b", re.I),
     "find_dead_code", "Find unreachable/unused code."),
    (re.compile(r"\boutline\b|\bstructure of\b|\bwhat'?s in .*\bfile\b|\bsymbols in\b", re.I),
     "get_file_outline", "List the symbols/structure of a file."),
    (re.compile(r"\b(string|text|literal|config value|comment|grep|regex)\b", re.I),
     "search_text", "Full-text search across file contents."),
    (re.compile(r"\bclass (hierarchy|tree)\b|\bsubclass|\bsuperclass|\binherit", re.I),
     "get_class_hierarchy", "Show a class inheritance hierarchy."),
    (re.compile(r"\bdependenc|\bimport graph\b|\bwhat imports\b", re.I),
     "get_dependency_graph", "Map file-level import dependencies."),
    (re.compile(r"\bhealth\b|\bhotspot|\bcomplexit|\bchurn\b|\brisk\b", re.I),
     "get_repo_health", "Repo-level health, hotspots, and risk."),
    (re.compile(r"\bplan\b|\bwhere (do|should) i (start|begin)\b|\bcontext for\b|\bonboard\b|\bunderstand the\b", re.I),
     "assemble_task_context", "Single-call task-scoped context assembly."),
    (re.compile(r"\b(find|locate|where is|look up|search for|definition of)\b", re.I),
     "search_symbols", "Find a symbol by name."),
]

# Repo-scoped actions whose primary query arg is named differently. Used by
# route(execute=true) to shape args from (repo, task).
_QUERY_ARG: dict[str, str] = {
    "search_symbols": "query",
    "search_text": "query",
    "assemble_task_context": "task",
    "plan_turn": "query",
    "get_file_outline": "file_path",
}


def classify_intent(task: str, catalog_names: Iterable[str]) -> list[dict]:
    """Return ranked recommended actions for a task.

    Combines the curated intent rules (high precision) with a catalog-search
    fallback (high recall), de-duplicated, primary first. Each row is
    ``{"action", "why"}``. Only actions present in the live catalog survive.
    """
    names = set(catalog_names)
    out: list[dict] = []
    seen: set[str] = set()
    for pat, action, why in _INTENT_RULES:
        if action in names and action not in seen and pat.search(task or ""):
            out.append({"action": action, "why": why})
            seen.add(action)
    return out


def shape_execute_args(action: str, repo: Optional[str], task: str) -> Optional[dict]:
    """Build a best-effort argument dict to dispatch *action* from (repo, task).

    Returns None when the action's inputs can't be satisfied from route's
    inputs (caller should then recommend rather than execute).
    """
    qarg = _QUERY_ARG.get(action)
    if qarg is None:
        return None
    if action == "get_file_outline":
        # Needs a concrete file path, which a free-form task rarely provides.
        return None
    if not repo:
        return None
    return {"repo": repo, qarg: task}
