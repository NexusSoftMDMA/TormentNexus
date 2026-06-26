"""Preflight check: is it safe to delete this symbol?

Combines find_importers (with cross-repo), check_references (text + import search),
find_dead_code confidence, runtime evidence (when Phase 7 traces exist), and
entry-point heuristics into a single verdict + actionable recommendation.

Verdict tiers (most-permissive first):
  - safe_to_delete         — no importers, no refs, dead-code confidence ≥0.9, no runtime hits
  - test_coverage_only     — only test files reference it (orphan; consider removing tests too)
  - internal_only          — refs exist only within the symbol's own file
  - internal_uses_blocking — referenced by other symbols in this repo (refactor first)
  - external_uses_blocking — imported by other files in this repo
  - cross_repo_blocking    — used by other indexed repos (highest static severity)
  - runtime_observed       — Phase 7 traces show this code runs (red flag regardless of static refs)
  - entry_point            — decorator/main pattern suggests external invocation
"""

from __future__ import annotations

import logging
import re
import time
from typing import Optional

from ..storage import IndexStore, record_savings, estimate_savings, cost_avoided
from ._utils import index_status_to_tool_error, resolve_repo

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Severity scoring for individual blockers (1-5, higher = more dangerous)
# ---------------------------------------------------------------------------
_SEVERITY_CROSS_REPO = 5
_SEVERITY_RUNTIME = 5
_SEVERITY_ENTRY_POINT = 5
_SEVERITY_EXTERNAL_IMPORT = 4
_SEVERITY_INTERNAL_REF = 3
_SEVERITY_TEST_ONLY = 2

# Decorator patterns suggesting external invocation
_ENTRY_DECORATOR_RE = re.compile(
    r"\b(route|get|post|put|patch|delete|command|task|signal|"
    r"event|listener|handler|subscribe|on|receiver|websocket|"
    r"endpoint|api|view|mount|app|cli|main|fixture)\b",
    re.IGNORECASE,
)

_TEST_FILE_RE = re.compile(r"(^|[/\\])(test_|tests?[/\\]|_test\.|conftest\.py)", re.IGNORECASE)


def _is_test_file(file_path: str) -> bool:
    return bool(_TEST_FILE_RE.search(file_path or ""))


def _resolve_target(index, symbol: str) -> Optional[dict]:
    """Resolve a symbol id or name to one symbol dict."""
    for sym in index.symbols:
        if sym.get("id") == symbol:
            return sym
    candidates = [s for s in index.symbols if s.get("name") == symbol]
    if not candidates:
        return None
    # Prefer non-import kinds with the largest body
    candidates.sort(key=lambda s: (
        s.get("kind") == "import",
        -int(s.get("byte_length", 0) or 0),
    ))
    return candidates[0]


def _detect_entry_point(target: dict) -> Optional[str]:
    """Return the matched entry-point indicator if target looks like one."""
    decorators = target.get("decorators") or []
    if isinstance(decorators, str):
        decorators = [decorators]
    for dec in decorators:
        dec_str = str(dec) if not isinstance(dec, dict) else (dec.get("name") or "")
        if dec_str and _ENTRY_DECORATOR_RE.search(dec_str):
            return f"decorator:{dec_str}"
    # __main__ / main heuristics
    name = (target.get("name") or "").lower()
    if name in {"main", "__main__", "run", "serve", "cli", "app"}:
        return f"name:{name}"
    return None


def _runtime_hits(store: IndexStore, owner: str, name: str, symbol_id: str) -> Optional[int]:
    """Best-effort runtime hit count over the indexed trace window."""
    try:
        import sqlite3  # noqa: PLC0415
        db_path = store._sqlite._db_path(owner, name)
        if not db_path.exists():
            return None
        conn = sqlite3.connect(f"file:{db_path}?mode=ro&immutable=1", uri=True)
        try:
            cur = conn.execute(
                "SELECT COALESCE(SUM(hit_count), 0) FROM runtime_calls WHERE symbol_id = ?",
                (symbol_id,),
            )
            row = cur.fetchone()
            return int(row[0]) if row and row[0] else None
        finally:
            conn.close()
    except Exception as exc:  # noqa: BLE001
        logger.debug("check_delete_safe: runtime hits skipped: %s", exc, exc_info=True)
        return None


def _runtime_data_present(store: IndexStore, owner: str, name: str) -> bool:
    """Has *any* runtime trace been ingested for this repo?

    Distinct from :func:`_runtime_hits` — which conflates "no traces at
    all" with "this particular symbol has zero hits in traces that
    exist." This helper lets ``safe_to_delete`` verdicts caveat
    themselves honestly when the runtime channel was simply never
    populated.
    """
    try:
        import sqlite3  # noqa: PLC0415
        db_path = store._sqlite._db_path(owner, name)
        if not db_path.exists():
            return False
        conn = sqlite3.connect(f"file:{db_path}?mode=ro&immutable=1", uri=True)
        try:
            row = conn.execute("SELECT 1 FROM runtime_calls LIMIT 1").fetchone()
            return row is not None
        finally:
            conn.close()
    except Exception as exc:  # noqa: BLE001
        logger.debug("check_delete_safe: runtime probe skipped: %s", exc, exc_info=True)
        return False


def _check_dead_code_conf(repo: str, target_id: str, storage_path: Optional[str]) -> float:
    """Look up find_dead_code's confidence score for this symbol."""
    try:
        from .find_dead_code import find_dead_code  # noqa: PLC0415
        out = find_dead_code(
            repo, granularity="symbol", min_confidence=0.0, include_tests=False,
            storage_path=storage_path,
        )
        entries = out.get("dead_symbols") or out.get("results") or []
        for e in entries:
            if e.get("symbol_id") == target_id:
                return float(e.get("confidence", 0.0))
    except Exception as exc:  # noqa: BLE001
        logger.debug("check_delete_safe: find_dead_code lookup skipped: %s", exc, exc_info=True)
    return 0.0


def check_delete_safe(
    repo: str,
    symbol: str,
    cross_repo: bool = True,
    include_runtime: bool = True,
    storage_path: Optional[str] = None,
) -> dict:
    """Composite preflight: can this symbol be deleted safely?

    Returns one verdict tier, a confidence score, a ranked list of blockers,
    and a one-line recommended action. Reuses find_importers + check_references
    + find_dead_code + runtime evidence; never mutates the codebase.

    Args:
        repo: Repository identifier.
        symbol: Symbol id or name to evaluate.
        cross_repo: Include other indexed repos in the analysis (default True).
        include_runtime: Consult runtime_calls for production evidence (default True).
        storage_path: Custom storage path.

    Returns:
        Dict with ``verdict``, ``confidence``, ``blockers`` list, ``recommended_action``,
        per-signal counts, and ``_meta``.
    """
    start = time.perf_counter()

    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return {"error": str(e)}

    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name)
    if not index:
        return index_status_to_tool_error(store.inspect_index(owner, name))

    target = _resolve_target(index, symbol)
    if target is None:
        return {"error": f"Symbol not found: {symbol}"}

    target_id = target["id"]
    target_name = target.get("name", "")
    target_file = target.get("file", "")

    blockers: list[dict] = []

    # ── Signal 1: entry-point indicator ─────────────────────────────────
    entry_signal = _detect_entry_point(target)
    if entry_signal:
        blockers.append({
            "kind": "entry_point",
            "detail": entry_signal,
            "severity": _SEVERITY_ENTRY_POINT,
        })

    # ── Signal 2: file-level importers (cross_repo when requested) ─────
    # Test-file importers are tracked separately so the verdict can correctly
    # downgrade to test_coverage_only when nothing but tests imports the file.
    external_import_count = 0
    test_import_count = 0
    cross_repo_count = 0
    try:
        from .find_importers import find_importers  # noqa: PLC0415
        importers_out = find_importers(
            repo=f"{owner}/{name}", file_path=target_file,
            cross_repo=cross_repo, storage_path=storage_path,
        )
        for entry in importers_out.get("importers", []) or []:
            if entry.get("cross_repo"):
                cross_repo_count += 1
                blockers.append({
                    "kind": "cross_repo_import",
                    "repo": entry.get("source_repo", ""),
                    "file": entry.get("file", ""),
                    "severity": _SEVERITY_CROSS_REPO,
                })
            else:
                imp_file = entry.get("file", "")
                if imp_file and imp_file != target_file:
                    if _is_test_file(imp_file):
                        test_import_count += 1
                        blockers.append({
                            "kind": "test_import",
                            "file": imp_file,
                            "severity": _SEVERITY_TEST_ONLY,
                        })
                    else:
                        external_import_count += 1
                        blockers.append({
                            "kind": "external_import",
                            "file": imp_file,
                            "severity": _SEVERITY_EXTERNAL_IMPORT,
                        })
    except Exception as exc:  # noqa: BLE001
        logger.debug("check_delete_safe: find_importers skipped: %s", exc, exc_info=True)

    # ── Signal 3: identifier text refs (catches duck-typed callers) ────
    internal_ref_count = 0
    test_ref_count = 0
    try:
        from .check_references import check_references  # noqa: PLC0415
        # Batch form (identifiers=[...]) so check_references returns its grouped
        # `results` shape — singular (identifier=...) returns a flat response with
        # no `results` key, which this loop would silently read as empty (#338).
        ref_out = check_references(
            repo=f"{owner}/{name}", identifiers=[target_name],
            search_content=True, max_content_results=20,
            storage_path=storage_path,
        )
        for entry in ref_out.get("results", []) or []:
            for ref in entry.get("content_references", []) or []:
                ref_file = ref.get("file", "")
                if not ref_file or ref_file == target_file:
                    continue
                if _is_test_file(ref_file):
                    test_ref_count += 1
                    if test_ref_count <= 3:
                        blockers.append({
                            "kind": "test_reference",
                            "file": ref_file,
                            "line": ref.get("line", 0),
                            "severity": _SEVERITY_TEST_ONLY,
                        })
                else:
                    internal_ref_count += 1
                    if internal_ref_count <= 3:
                        blockers.append({
                            "kind": "internal_reference",
                            "file": ref_file,
                            "line": ref.get("line", 0),
                            "severity": _SEVERITY_INTERNAL_REF,
                        })
    except Exception as exc:  # noqa: BLE001
        logger.debug("check_delete_safe: check_references skipped: %s", exc, exc_info=True)

    # ── Signal 4: dead-code confidence ─────────────────────────────────
    dead_code_conf = _check_dead_code_conf(f"{owner}/{name}", target_id, storage_path)

    # ── Signal 5: runtime evidence (Phase 7) ────────────────────────────
    runtime_hits = _runtime_hits(store, owner, name, target_id) if include_runtime else None
    runtime_data_present = _runtime_data_present(store, owner, name) if include_runtime else False
    if runtime_hits and runtime_hits > 0:
        blockers.append({
            "kind": "runtime_observed",
            "hit_count": runtime_hits,
            "severity": _SEVERITY_RUNTIME,
        })

    # ── Verdict selection ──────────────────────────────────────────────
    # Order matters — most restrictive first. Tests are counted separately
    # from external imports so test-only consumption downgrades the verdict.
    total_test_signals = test_ref_count + test_import_count
    total_external_signals = external_import_count
    total_internal_signals = internal_ref_count

    if runtime_hits and runtime_hits > 0:
        verdict = "runtime_observed"
    elif entry_signal:
        verdict = "entry_point"
    elif cross_repo_count > 0:
        verdict = "cross_repo_blocking"
    elif total_external_signals > 0:
        verdict = "external_uses_blocking"
    elif total_internal_signals > 0:
        verdict = "internal_uses_blocking"
    elif total_test_signals > 0:
        verdict = "test_coverage_only"
    elif dead_code_conf >= 0.9:
        verdict = "safe_to_delete"
    elif (total_internal_signals == 0 and total_external_signals == 0
          and total_test_signals == 0):
        # No refs at all, but dead-code analysis didn't reach high confidence.
        # Still surface as safe with a slightly lower confidence score.
        verdict = "safe_to_delete"
    else:
        verdict = "internal_only"

    # ── Confidence ─────────────────────────────────────────────────────
    # Start at dead-code confidence (or 0.5 baseline) and decay per blocker.
    confidence = max(0.5, dead_code_conf)
    if verdict == "safe_to_delete":
        confidence = max(confidence, 0.85 if dead_code_conf < 0.9 else 0.95)
    elif verdict == "runtime_observed":
        confidence = 0.05  # nearly certain unsafe
    elif verdict == "cross_repo_blocking":
        confidence = 0.10
    elif verdict == "entry_point":
        confidence = 0.20
    elif verdict == "external_uses_blocking":
        confidence = 0.25
    elif verdict == "internal_uses_blocking":
        confidence = 0.45
    elif verdict == "test_coverage_only":
        confidence = 0.65
    elif verdict == "internal_only":
        confidence = 0.55

    # ── Recommended action ─────────────────────────────────────────────
    # Honest-hint caveat: when the verdict relies on the *absence* of
    # runtime evidence but no traces have ever been ingested for this
    # repo, the runtime channel can't actually prove safety — only
    # static signals can. Surface that in the recommended_action so
    # operators don't read "safe" as "we checked production traffic."
    safe_action = "No callers, refs, or runtime hits found — deletion appears safe."
    if include_runtime and not runtime_data_present:
        safe_action = (
            "No callers or refs found. Static signals only — no runtime traces "
            "ingested for this repo, so production traffic was not consulted. "
            "Run `import-trace` against representative traffic to strengthen this verdict."
        )

    actions = {
        "safe_to_delete": safe_action,
        "test_coverage_only": "Only tests reference this symbol. Remove the tests alongside it.",
        "internal_only": "Refs exist only in the same file. Safe with local refactor.",
        "internal_uses_blocking": (
            f"{internal_ref_count} internal reference(s) found. Rename/refactor callers first."
        ),
        "external_uses_blocking": (
            f"{external_import_count} other file(s) in this repo import this. Update importers first."
        ),
        "cross_repo_blocking": (
            f"{cross_repo_count} other repo(s) in the suite depend on this. Coordinate a deprecation."
        ),
        "runtime_observed": (
            f"Runtime traces show {runtime_hits} hits — this code runs in production. "
            "Investigate why static analysis missed the callers."
        ),
        "entry_point": (
            f"Entry-point indicator ({entry_signal}) — invoked externally by framework/CLI/protocol. "
            "Never delete blindly; verify routing config."
        ),
    }

    # Rank blockers by severity, truncate to top 5
    blockers.sort(key=lambda b: -b.get("severity", 0))
    blockers_out = blockers[:5]

    # Token-savings ledger (cheap response)
    raw_bytes = int(target.get("byte_length", 0) or 0) + 1000
    response_bytes = 800
    tokens_saved = estimate_savings(raw_bytes, response_bytes)
    total_saved = record_savings(tokens_saved, tool_name="check_delete_safe")

    elapsed = (time.perf_counter() - start) * 1000

    result = {
        "verdict": verdict,
        "confidence": round(confidence, 2),
        "target": {
            "symbol_id": target_id,
            "name": target_name,
            "kind": target.get("kind", ""),
            "file": target_file,
            "line": target.get("line", 0),
        },
        "blockers": blockers_out,
        "recommended_action": actions.get(verdict, "Review blockers before deletion."),
        "signals": {
            "external_import_count": external_import_count,
            "test_import_count": test_import_count,
            "cross_repo_count": cross_repo_count,
            "internal_ref_count": internal_ref_count,
            "test_ref_count": test_ref_count,
            "dead_code_confidence": round(dead_code_conf, 3),
            "entry_point": entry_signal,
        },
        "_meta": {
            "timing_ms": round(elapsed, 1),
            "tokens_saved": tokens_saved,
            "total_tokens_saved": total_saved,
            **cost_avoided(tokens_saved, total_saved),
        },
    }
    if runtime_hits is not None:
        result["signals"]["runtime_hits"] = runtime_hits
    if include_runtime:
        result["signals"]["runtime_data_present"] = runtime_data_present
    return result
