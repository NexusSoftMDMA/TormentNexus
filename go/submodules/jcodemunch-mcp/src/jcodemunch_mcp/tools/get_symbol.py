"""Get symbol source code."""

import hashlib
import os
import subprocess
import time
from pathlib import Path
from typing import Optional

from ..storage import IndexStore, record_savings, estimate_savings, cost_avoided as _cost_avoided
from ._utils import index_status_to_tool_error, resolve_repo, resolve_fqn


def _make_meta(timing_ms: float, **kwargs) -> dict:
    """Build a _meta envelope dict."""
    meta = {"timing_ms": round(timing_ms, 1)}
    meta.update(kwargs)
    return meta


def _utf8_safe_truncate(text: str, max_bytes: int) -> str:
    """Truncate ``text`` to at most ``max_bytes`` UTF-8 bytes without splitting a
    multibyte character (a trailing partial sequence is dropped)."""
    if max_bytes <= 0:
        return ""
    encoded = text.encode("utf-8")
    if len(encoded) <= max_bytes:
        return text
    return encoded[:max_bytes].decode("utf-8", errors="ignore")


def _bound_source(
    source: str,
    symbol_line: int,
    symbol_end_line: int,
    source_start_line: Optional[int],
    source_end_line: Optional[int],
    max_source_lines: Optional[int],
    max_source_bytes: Optional[int],
    remaining_total_bytes: Optional[int],
) -> dict:
    """Shape one symbol's full body into a bounded slice + truncation metadata.

    Identity and verification still refer to the full indexed body; this only
    shapes the returned ``source`` and authors the server-side truncation
    contract. All line numbers are absolute file lines (matching ``line`` /
    ``end_line`` and ``get_file_content(start_line=, end_line=)``).

    Bounds apply in order — explicit line range (clamped to the symbol body) →
    ``max_source_lines`` → ``max_source_bytes`` → batch ``remaining_total_bytes``
    — so a later, tighter bound supersedes the reason of an earlier one.

    Returns ``{text, truncated, reason, range, total_range, total_lines,
    total_bytes}``.
    """
    full_lines = source.split("\n")
    total_lines = len(full_lines)
    total_bytes = len(source.encode("utf-8"))
    total_range = {"start_line": symbol_line, "end_line": symbol_end_line}

    # 1) Explicit absolute line range, clamped to the symbol body.
    rel_start = 0
    rel_end = total_lines  # exclusive
    reason = None
    if source_start_line is not None:
        rel_start = max(0, min(source_start_line - symbol_line, total_lines))
        if rel_start > 0:
            reason = "source_range"
    if source_end_line is not None:
        rel_end = max(rel_start, min(source_end_line - symbol_line + 1, total_lines))
        if rel_end < total_lines:
            reason = "source_range"
    sliced_lines = full_lines[rel_start:rel_end]
    start_abs = symbol_line + rel_start

    # 2) Max line cap on the (possibly range-limited) slice.
    if max_source_lines is not None and len(sliced_lines) > max_source_lines:
        sliced_lines = sliced_lines[:max_source_lines]
        reason = "max_source_lines"

    text = "\n".join(sliced_lines)

    # 3) Per-symbol byte cap (UTF-8 safe).
    if max_source_bytes is not None and len(text.encode("utf-8")) > max_source_bytes:
        text = _utf8_safe_truncate(text, max_source_bytes)
        reason = "max_source_bytes"

    # 4) Batch total-byte cap (caller-supplied running budget; overrides).
    if remaining_total_bytes is not None and len(text.encode("utf-8")) > remaining_total_bytes:
        text = _utf8_safe_truncate(text, remaining_total_bytes)
        reason = "max_total_source_bytes"

    truncated = text != source
    # Returned absolute range: the last line may be byte-truncated but is still
    # partially present, so it counts. An empty slice returns an empty range.
    end_abs = (start_abs + text.count("\n")) if text else (start_abs - 1)
    return {
        "text": text,
        "truncated": truncated,
        "reason": reason if truncated else None,
        "range": {"start_line": start_abs, "end_line": end_abs},
        "total_range": total_range,
        "total_lines": total_lines,
        "total_bytes": total_bytes,
    }


def _verify_against_git_sha(
    cached_source: str,
    source_root: Optional[str],
    file_path: str,
    line: int,
    end_line: int,
) -> str:
    """Compare cached source against the working-tree git HEAD content (P1.6).

    Returns one of:
    - ``"git_sha_match"``      — the cached source matches the HEAD slice
                                  of the same file (lines line..end_line).
    - ``"git_sha_mismatch"``   — the file exists in HEAD but the slice differs.
    - ``"git_unavailable"``    — source_root unknown, file isn't tracked in
                                  HEAD, or git is unreachable from this env.

    This is an externally-attested verification mode: the comparison target
    comes from git, not from the same cache the symbol's content_hash was
    derived from. The default ``verify_against="cache"`` mode is self-referential
    and only catches incoherent tamper of ``~/.code-index/<repo>/``; this mode
    catches divergence between the cache and the upstream source.
    """
    if not source_root or not file_path:
        return "git_unavailable"
    root = Path(source_root)
    if not (root / ".git").exists() and not (root / ".git").is_file():
        # Not a git working tree (or worktree pointing elsewhere; bail rather
        # than guess).
        return "git_unavailable"
    try:
        result = subprocess.run(
            ["git", "-C", str(root), "show", f"HEAD:{file_path}"],
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
            # Windows stdio-MCP deadlock guard: never inherit the JSON-RPC pipe
            # as the git child's stdin (Git-for-Windows' cmd\git.exe wrapper
            # blocks forever holding the handle, even for commands that don't
            # read stdin). Mirrors the redirect across the other git spawns.
            stdin=subprocess.DEVNULL,
        )
    except (FileNotFoundError, subprocess.TimeoutExpired, OSError):
        return "git_unavailable"
    if result.returncode != 0:
        # File not in HEAD (untracked, new file, deleted from HEAD, etc.)
        return "git_unavailable"
    head_content = result.stdout
    if not head_content:
        return "git_unavailable"
    head_lines = head_content.split("\n")
    if line < 1 or end_line < line or end_line > len(head_lines):
        # Symbol line range no longer falls within the HEAD file shape; treat
        # as divergence rather than match.
        return "git_sha_mismatch"
    head_slice = "\n".join(head_lines[line - 1:end_line])
    cached_slice = cached_source.rstrip("\n")
    head_slice = head_slice.rstrip("\n")
    return "git_sha_match" if head_slice == cached_slice else "git_sha_mismatch"


def get_symbol_source(
    repo: str,
    symbol_id: Optional[str] = None,
    symbol_ids: Optional[list[str]] = None,
    verify: bool = False,
    context_lines: int = 0,
    storage_path: Optional[str] = None,
    fqn: Optional[str] = None,
    verify_against: str = "cache",
    source_start_line: Optional[int] = None,
    source_end_line: Optional[int] = None,
    max_source_lines: Optional[int] = None,
    max_source_bytes: Optional[int] = None,
    max_total_source_bytes: Optional[int] = None,
) -> dict:
    """Get full source of one or more symbols by ID.

    Pass symbol_id (string) for one symbol — returns flat symbol object.
    Pass symbol_ids (array) for batch — returns {symbols, errors}.
    Both modes support verify and context_lines.
    Pass fqn (PHP FQN like 'App\\Models\\User') to resolve via PSR-4.

    Bounded-source mode (all optional, default off — when none are supplied the
    response is byte-for-byte the full-source default). Lets large symbols or
    broad batches return an explicitly-labeled source *slice* so a downstream
    client/context clip can't silently hand the agent a partial body:

    - ``source_start_line`` / ``source_end_line``: absolute file line numbers
      (same frame as ``line`` / ``end_line``), clamped to the symbol body.
    - ``max_source_lines``: keep at most the first N lines of the (ranged) slice.
    - ``max_source_bytes``: UTF-8-safe per-symbol byte cap.
    - ``max_total_source_bytes``: batch cap across all returned symbols, so a
      large batch returns bounded entries instead of an N x per-symbol blowup;
      oversized symbols come back partial, never dropped.

    When a bound shortens the source, the entry carries server-authored metadata:
    ``source_truncated``, ``source_range``, ``source_total_range``,
    ``source_total_lines``, ``source_total_bytes``, ``source_truncated_reason``,
    and ``source_is_bounded_view`` (so a verified entry shows the returned source
    is a slice, not the verified bytes). ``verify`` always hashes the *full*
    indexed body. ``context_lines`` may not be combined with any source bound
    (rejected) so it can never expand the payload past the requested bound.
    """
    # FQN resolution: translate PHP FQN → symbol_id
    if fqn and symbol_id is None and symbol_ids is None:
        resolved, fqn_error = resolve_fqn(repo, fqn, storage_path)
        if resolved is None:
            return {"error": fqn_error or f"Could not resolve FQN '{fqn}'."}
        symbol_id = resolved

    # Normalize: some MCP clients send symbol_ids=[] alongside symbol_id when they mean singular mode
    if symbol_id is not None and symbol_ids is not None and len(symbol_ids) == 0:
        symbol_ids = None
    if symbol_id is None and symbol_ids is None:
        return {"error": "Provide symbol_id (string), symbol_ids (array), or fqn (PHP FQN)."}
    if symbol_id is not None and symbol_ids is not None:
        return {"error": "Provide symbol_id or symbol_ids, not both."}

    batch_mode = symbol_ids is not None
    ids = symbol_ids if batch_mode else [symbol_id]

    start = time.perf_counter()
    context_lines = max(0, min(context_lines, 50))

    # Bounded-source mode: validated up-front so a bad bound rejects fast and the
    # contract is unambiguous (see docstring). Default (no bounds) is untouched.
    bounds_requested = any(
        v is not None for v in (
            source_start_line, source_end_line,
            max_source_lines, max_source_bytes, max_total_source_bytes,
        )
    )
    if bounds_requested:
        if context_lines > 0:
            return {"error": (
                "context_lines cannot be combined with source bounds "
                "(source_start_line / source_end_line / max_source_lines / "
                "max_source_bytes / max_total_source_bytes); it would expand the "
                "payload past the requested bound. Request context in a separate "
                "unbounded call."
            )}
        for label, val in (
            ("source_start_line", source_start_line),
            ("source_end_line", source_end_line),
            ("max_source_lines", max_source_lines),
            ("max_source_bytes", max_source_bytes),
            ("max_total_source_bytes", max_total_source_bytes),
        ):
            if val is not None and val < 1:
                return {"error": f"{label} must be >= 1 when provided."}
        if (source_start_line is not None and source_end_line is not None
                and source_end_line < source_start_line):
            return {"error": "source_end_line must be >= source_start_line."}

    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return {"error": str(e)}

    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name)

    if not index:
        return index_status_to_tool_error(store.inspect_index(owner, name))

    symbols_out = []
    errors_out = []
    seen_files: set = set()
    raw_bytes = 0
    response_bytes = 0
    total_source_used = 0  # running byte total for the batch max_total_source_bytes cap

    for sid in ids:
        symbol = index.get_symbol(sid)

        if not symbol:
            errors_out.append({"id": sid, "error": f"Symbol not found: {sid}"})
            continue

        source = store.get_symbol_content(owner, name, sid, _index=index)
        content_dir = store._content_dir(owner, name)
        file_full_path = content_dir / symbol["file"]

        context_before = ""
        context_after = ""
        if context_lines > 0 and source and file_full_path.exists():
            try:
                all_lines = file_full_path.read_text(encoding="utf-8", errors="replace").split("\n")
                s_line = symbol["line"] - 1  # 0-indexed
                e_line = symbol["end_line"]   # exclusive
                before_start = max(0, s_line - context_lines)
                after_end = min(len(all_lines), e_line + context_lines)
                if before_start < s_line:
                    context_before = "\n".join(all_lines[before_start:s_line])
                if e_line < after_end:
                    context_after = "\n".join(all_lines[e_line:after_end])
            except Exception:
                pass

        # Bounded-source mode shapes the returned `source` into an explicitly
        # labeled slice; `source` (the variable) stays the full body so verify
        # below still hashes the complete indexed bytes.
        display_source = source or ""
        bound_meta = None
        if bounds_requested and source:
            remaining = None
            if max_total_source_bytes is not None:
                remaining = max(0, max_total_source_bytes - total_source_used)
            bound_meta = _bound_source(
                source,
                symbol["line"],
                symbol["end_line"],
                source_start_line,
                source_end_line,
                max_source_lines,
                max_source_bytes,
                remaining,
            )
            display_source = bound_meta["text"]
            total_source_used += len(display_source.encode("utf-8"))

        entry = {
            "id": symbol["id"],
            "kind": symbol["kind"],
            "name": symbol["name"],
            "file": symbol["file"],
            "line": symbol["line"],
            "end_line": symbol["end_line"],
            "signature": symbol["signature"],
            "decorators": symbol.get("decorators", []),
            "docstring": symbol.get("docstring", ""),
            "content_hash": symbol.get("content_hash", ""),
            "source": display_source,
        }
        if bound_meta is not None:
            entry["source_truncated"] = bound_meta["truncated"]
            if bound_meta["truncated"]:
                # Verified entries: flag that `source` is a slice, not the bytes
                # `content_verified` attests to (which is always the full body).
                entry["source_is_bounded_view"] = True
                entry["source_range"] = bound_meta["range"]
                entry["source_total_range"] = bound_meta["total_range"]
                entry["source_total_lines"] = bound_meta["total_lines"]
                entry["source_total_bytes"] = bound_meta["total_bytes"]
                entry["source_truncated_reason"] = bound_meta["reason"]
        # P1.4: distinguish "empty source" from "no body cached because we're
        # in metadata_only mode" so downstream agents don't treat the empty
        # string as the symbol's actual source.
        if not source:
            try:
                from .. import config as _cfg
                if _cfg.get("cache_mode", "full") == "metadata_only":
                    entry["source_status"] = "metadata_only_mode"
            except Exception:
                pass
        if context_before:
            entry["context_before"] = context_before
        if context_after:
            entry["context_after"] = context_after

        if verify and source:
            actual_hash = hashlib.sha256(source.encode("utf-8")).hexdigest()
            stored_hash = symbol.get("content_hash", "")
            entry["content_verified"] = actual_hash == stored_hash if stored_hash else None
            # P1.6: externally-attested mode compares cached source against the
            # working-tree git HEAD slice of the same file. Surfaced alongside
            # the cache-only verification so callers can see both signals.
            if verify_against == "git_sha":
                entry["git_sha_verification"] = _verify_against_git_sha(
                    cached_source=source,
                    source_root=getattr(index, "source_root", None),
                    file_path=symbol["file"],
                    line=symbol["line"],
                    end_line=symbol["end_line"],
                )

        symbols_out.append(entry)

        # Accumulate token savings
        f = symbol["file"]
        if f not in seen_files:
            seen_files.add(f)
            try:
                raw_bytes += os.path.getsize(file_full_path)
            except OSError:
                pass
        response_bytes += symbol.get("byte_length", 0)

    tokens_saved = estimate_savings(raw_bytes, response_bytes)
    total_saved = record_savings(tokens_saved, tool_name="get_symbol_source")
    elapsed = (time.perf_counter() - start) * 1000
    meta = _make_meta(elapsed, tokens_saved=tokens_saved, total_tokens_saved=total_saved,
                      **_cost_avoided(tokens_saved, total_saved))

    from ..retrieval.freshness import FreshnessProbe as _FreshnessProbe
    _probe = _FreshnessProbe(
        source_root=getattr(index, "source_root", "") or None,
        indexed_at=getattr(index, "indexed_at", ""),
        index_sha=getattr(index, "git_head", None),
        file_mtimes=getattr(index, "file_mtimes", None),
    )
    _probe.annotate(symbols_out)

    # Phase 2: runtime confidence — zero-cost no-op when no traces ingested.
    from ..runtime.confidence import attach_runtime_confidence as _attach_runtime
    _runtime_summary = _attach_runtime(
        symbols_out,
        str(store._sqlite._db_path(owner, name)),
        id_field="id",
    )

    if batch_mode:
        meta["symbol_count"] = len(symbols_out)
        meta["freshness"] = _probe.summary(symbols_out)
        if _runtime_summary:
            meta["runtime_freshness"] = _runtime_summary
        return {"symbols": symbols_out, "errors": errors_out, "_meta": meta}

    # Single mode: flat object or error
    if errors_out:
        return {"error": errors_out[0]["error"]}
    result = symbols_out[0]
    meta["hint"] = "Use get_context_bundle(symbol_id) to retrieve source + imports in one call"
    meta["freshness"] = _probe.summary(symbols_out)
    if _runtime_summary:
        meta["runtime_freshness"] = _runtime_summary
    result["_meta"] = meta
    return result
