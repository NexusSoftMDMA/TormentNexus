"""``digest`` — agent stand-up briefing.

Composes a tight (~200 token) markdown briefing from existing tools so an
agent walking into a session knows the load-bearing changes since its
last visit + the current high-risk surface area, without having to
discover any of it via cold exploration.

Design properties:
  - **Agent-facing, not human-facing.** Every line either references a
    symbol_id the agent can immediately query, names a tool to call, or
    flags a regression worth investigating.
  - **Composes existing tools, doesn't recompute.** Reuses
    ``get_changed_symbols``, ``get_hotspots``, ``find_dead_code`` so the
    briefing tracks tool improvements automatically.
  - **State-aware.** Tracks the SHA the agent last saw per repo at
    ``~/.code-index/digest_state/<repo>.json``; on next call surfaces
    the *delta*, not a fresh snapshot.

Surfaces this module powers:
  1. ``mcp__jcodemunch__digest`` — agent calls during a session
  2. ``jcodemunch-mcp digest`` CLI — developer at standup
  3. (v2) Hook injection at SubagentStart — deferred
"""

from __future__ import annotations

import datetime as _dt
import json
import logging
import os
import time
from pathlib import Path
from typing import Optional

from ..storage import IndexStore
from ._utils import index_status_to_tool_error, resolve_repo

logger = logging.getLogger(__name__)

_DIGEST_STATE_SUBDIR = "digest_state"


def _state_path(owner: str, name: str, base_path: Optional[str] = None) -> Path:
    """Resolve the per-repo state file location."""
    if base_path:
        root = Path(base_path)
    else:
        root = Path(os.environ.get("CODE_INDEX_PATH") or Path.home() / ".code-index")
    return root / _DIGEST_STATE_SUBDIR / f"{owner}--{name}.json"


def _read_state(path: Path) -> dict:
    if not path.exists():
        return {}
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        logger.debug("could not read digest state at %s", path, exc_info=True)
        return {}


def _write_state(path: Path, payload: dict) -> None:
    try:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(payload, indent=2), encoding="utf-8")
    except OSError:
        logger.debug("could not write digest state at %s", path, exc_info=True)


def _git_head(source_root: str) -> Optional[str]:
    """Return current HEAD SHA, or None if not a git repo."""
    import subprocess
    try:
        r = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            cwd=source_root,
            capture_output=True,
            text=True,
            timeout=5,
            stdin=subprocess.DEVNULL,
        )
        if r.returncode == 0:
            return r.stdout.strip() or None
    except (subprocess.TimeoutExpired, OSError, FileNotFoundError):
        return None
    return None


def _short_sha(sha: str) -> str:
    return sha[:7] if sha and len(sha) >= 7 else sha


def _truncate_symbol_id(sym_id: str, max_len: int = 60) -> str:
    if len(sym_id) <= max_len:
        return sym_id
    return "..." + sym_id[-(max_len - 3):]


def compose_digest(
    repo: str,
    *,
    since_sha: Optional[str] = None,
    max_changed_files: int = 5,
    max_hotspots: int = 3,
    max_dead_code: int = 3,
    storage_path: Optional[str] = None,
) -> dict:
    """Compose a since-last-session briefing for a repo.

    Args:
        repo: Repo identifier (owner/name, full id, or bare display name).
        since_sha: Override the last-seen SHA (default: state file or
            falls back to "no prior session known").
        max_changed_files: Cap on changed-files list.
        max_hotspots: Cap on hotspot list.
        max_dead_code: Cap on dead-code candidates.
        storage_path: Index storage override.

    Returns:
        Dict with:
            briefing: Markdown string for agent context injection
            structured: Per-section data for programmatic access
            _meta: Timing info
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

    state_path = _state_path(owner, name, storage_path)
    state = _read_state(state_path)
    prior_sha = since_sha or state.get("git_head")
    prior_session_at = state.get("session_at")

    current_head: Optional[str] = None
    if index.source_root:
        current_head = _git_head(index.source_root)

    structured: dict = {
        "repo": f"{owner}/{name}",
        "current_head": current_head,
        "prior_head": prior_sha,
        "prior_session_at": prior_session_at,
        "languages": _languages_summary(index),
        "n_files": len(index.source_files),
        "n_symbols": len(index.symbols),
    }

    # Section 1: since-last-session delta (only when we have prior SHA + git)
    delta = {}
    if prior_sha and current_head and prior_sha != current_head and index.source_root:
        delta = _compose_delta(
            owner, name, prior_sha, current_head, max_changed_files, storage_path,
        )
        structured["delta"] = delta

    # Section 2: hotspots (always)
    hotspots = _compose_hotspots(owner, name, max_hotspots, storage_path)
    structured["hotspots"] = hotspots

    # Section 3: dead-code candidates (always)
    dead = _compose_dead_code(owner, name, max_dead_code, storage_path)
    structured["dead_code"] = dead

    # Section 4: retrieval-regret summary (only when the ledger has clusters)
    regret = _compose_regret(f"{owner}/{name}", storage_path)
    if regret:
        structured["regret"] = regret

    briefing = _render_markdown(structured)

    # Persist current state so the next call computes a fresh delta.
    _write_state(state_path, {
        "git_head": current_head,
        "session_at": _dt.datetime.now(_dt.timezone.utc).isoformat(),
    })

    elapsed_ms = (time.perf_counter() - start) * 1000.0
    return {
        "briefing": briefing,
        "structured": structured,
        "_meta": {"timing_ms": round(elapsed_ms, 1)},
    }


def _languages_summary(index, max_langs: int = 4) -> str:
    """Top-N language list by symbol count."""
    counts: dict[str, int] = {}
    for sym in index.symbols:
        lang = sym.get("language")
        if lang:
            counts[lang] = counts.get(lang, 0) + 1
    if not counts:
        return ""
    ranked = sorted(counts.items(), key=lambda kv: kv[1], reverse=True)
    return ", ".join(name for name, _ in ranked[:max_langs])


def _compose_delta(
    owner: str,
    name: str,
    since_sha: str,
    until_sha: str,
    max_files: int,
    storage_path: Optional[str],
) -> dict:
    """Compose changed-symbols delta. Failures degrade silently."""
    try:
        from .get_changed_symbols import get_changed_symbols
        result = get_changed_symbols(
            f"{owner}/{name}",
            since_sha=since_sha,
            until_sha=until_sha,
            include_blast_radius=False,
            suppress_meta=True,
            storage_path=storage_path,
        )
        if "error" in result:
            return {"error": result["error"]}
        files = (result.get("changed_files") or [])[:max_files]
        added = (result.get("added_symbols") or [])[:max_files]
        modified = (result.get("changed_symbols") or [])[:max_files]
        removed = (result.get("removed_symbols") or [])[:max_files]
        return {
            "files": files,
            "added": added,
            "modified": modified,
            "removed": removed,
            "from_sha": result.get("from_sha"),
            "to_sha": result.get("to_sha"),
        }
    except Exception:
        logger.debug("compose_delta failed", exc_info=True)
        return {}


def _compose_hotspots(
    owner: str, name: str, max_n: int, storage_path: Optional[str],
) -> list[dict]:
    """Top-N hotspots by complexity × churn. Failures degrade silently."""
    try:
        from .get_hotspots import get_hotspots
        result = get_hotspots(
            f"{owner}/{name}",
            top_n=max_n,
            storage_path=storage_path,
        )
        return (result.get("hotspots") or [])[:max_n]
    except Exception:
        logger.debug("compose_hotspots failed", exc_info=True)
        return []


def _compose_dead_code(
    owner: str, name: str, max_n: int, storage_path: Optional[str],
) -> list[dict]:
    """Top-N dead-code candidates. Failures degrade silently."""
    try:
        from .find_dead_code import find_dead_code
        result = find_dead_code(
            f"{owner}/{name}",
            storage_path=storage_path,
        )
        candidates = result.get("dead_symbols") or result.get("candidates") or []
        return candidates[:max_n]
    except Exception:
        logger.debug("compose_dead_code failed", exc_info=True)
        return []


def _compose_regret(repo: str, storage_path: Optional[str]) -> Optional[dict]:
    """One-line retrieval-regret summary from the ranking ledger. Returns None
    when telemetry is off or no regret clusters cross threshold (so the digest
    line only appears when there's something to act on). Failures degrade
    silently."""
    try:
        from ..retrieval.regret import analyze_regret
        out = analyze_regret(repo, storage_path=storage_path)
        clusters = out.get("clusters") or []
        if not clusters:
            return None
        top = clusters[0]
        return {
            "count": len(clusters),
            "events": out.get("events_analyzed", 0),
            "top_signal": top.get("signal"),
            "top_severity": top.get("severity"),
        }
    except Exception:
        logger.debug("compose_regret failed", exc_info=True)
        return None


def _render_markdown(s: dict) -> str:
    """Render the briefing as a tight markdown digest."""
    lines: list[str] = [f"## jCodemunch digest — {s['repo']}"]

    head_short = _short_sha(s.get("current_head") or "")
    head_label = f"@ {head_short}" if head_short else ""
    lines.append(
        f"**{s['n_symbols']:,} symbols across {s['n_files']:,} files** "
        f"({s.get('languages') or 'unknown'}) {head_label}".rstrip()
    )

    delta = s.get("delta") or {}
    if delta and not delta.get("error"):
        from_sha = _short_sha(delta.get("from_sha") or "")
        to_sha = _short_sha(delta.get("to_sha") or "")
        files = delta.get("files") or []
        added = delta.get("added") or []
        modified = delta.get("modified") or []
        removed = delta.get("removed") or []
        if files or added or modified or removed:
            lines.append(f"\n### Since {from_sha}…{to_sha}")
            if files:
                lines.append(f"**Files changed:** {len(files)}")
                for f in files[:5]:
                    lines.append(f"- `{f}`")
            if added:
                lines.append(f"**Added symbols:** {len(added)}")
                for sym in added[:3]:
                    sid = sym.get("symbol_id") or sym.get("name", "?")
                    lines.append(f"- `{_truncate_symbol_id(sid)}`")
            if modified:
                lines.append(f"**Modified symbols:** {len(modified)}")
                for sym in modified[:3]:
                    sid = sym.get("symbol_id") or sym.get("name", "?")
                    lines.append(f"- `{_truncate_symbol_id(sid)}`")
            if removed:
                lines.append(f"**Removed symbols:** {len(removed)}")
                for sym in removed[:3]:
                    sid = sym.get("symbol_id") or sym.get("name", "?")
                    lines.append(f"- `{_truncate_symbol_id(sid)}`")
    elif s.get("prior_head") is None:
        lines.append("\n_(first session — no prior digest state; future calls will surface deltas)_")

    hotspots = s.get("hotspots") or []
    if hotspots:
        lines.append("\n### Risk surface (hotspots)")
        for h in hotspots:
            sid = h.get("symbol_id") or h.get("name", "?")
            score = h.get("hotspot_score", 0)
            lines.append(f"- `{_truncate_symbol_id(sid)}` — score {score:.1f}")
        lines.append("Drill in: `get_symbol_complexity` / `get_call_hierarchy`.")

    dead = s.get("dead_code") or []
    if dead:
        lines.append("\n### Dead-code candidates")
        for d in dead[:3]:
            sid = d.get("symbol_id") or d.get("name", "?")
            lines.append(f"- `{_truncate_symbol_id(sid)}`")
        lines.append("Verify with `check_references` before removal.")

    regret = s.get("regret")
    if regret:
        lines.append(
            f"\n### Retrieval regret\n{regret['count']} regret cluster(s) this window; "
            f"top: {regret['top_signal']} ({regret['top_severity']}). "
            f"Run `reflect` for suggested fixes."
        )

    lines.append(
        "\n_Composed from get_changed_symbols + get_hotspots + find_dead_code. "
        "Call those tools directly for full data._"
    )
    return "\n".join(lines)
