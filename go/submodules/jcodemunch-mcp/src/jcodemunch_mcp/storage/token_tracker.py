"""Persistent token savings tracker.

Records cumulative tokens saved across all tool calls by comparing
raw file sizes against actual MCP response sizes.

Stored in ~/.code-index/_savings.json — a single small JSON file.
No API calls, no file reads — only os.stat for file sizes.

Community meter: token savings are shared anonymously by default to the
global counter at https://j.gravelle.us. Only {"delta": N, "total": M,
"anon_id": "<uuid>"} is sent — never code, paths, repo names, or anything
identifying. `total` is the absolute lifetime savings; the server sets the
stored value to GREATEST(stored, total), so a dropped/failed POST self-heals
on the next successful one (the meter converges to the local total instead
of permanently undercounting from lost deltas). `delta` is kept for backward
compatibility with older server builds. Set JCODEMUNCH_SHARE_SAVINGS=0 to
disable; the endpoint is overridable via JCODEMUNCH_TELEMETRY_URL. The sender
is a single daemon thread started lazily on the first share (never at import,
and never when sharing is disabled), so a plain `import` has no background
side effect.

Performance: uses an in-memory accumulator to avoid disk read+write on every
tool call. Flushes to disk every FLUSH_INTERVAL calls (default 3), on SIGTERM/
SIGINT, and at process exit via atexit. Telemetry batches are sent at flush
time rather than per-call to avoid spawning a new thread on every tool use.
"""

import atexit
import json
import logging
import os
import queue
import signal
import sqlite3
import threading
import time
import uuid
from collections import OrderedDict, deque
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

from .. import config as _config

logger = logging.getLogger(__name__)

_SAVINGS_FILE = "_savings.json"
_SESSION_STATS_FILE = "session_stats.json"
_PULSE_FILE = "_pulse.json"
_PERF_DB_FILE = "telemetry.db"
_BYTES_PER_TOKEN = 4  # ~4 bytes per token (rough but consistent)
_TELEMETRY_URL = "https://j.gravelle.us/APIs/savings/post.php"
_FLUSH_INTERVAL = 3  # flush to disk every N calls
_RESULT_CACHE_MAXSIZE = 256  # max tool-result cache entries per session
_LATENCY_RING_DEFAULT = 512  # per-tool latency ring size
_PERF_DB_MAX_ROWS_DEFAULT = 100_000  # rolling cap on persisted perf rows

def _get_stats_file_interval() -> int:
    """Read stats_file_interval from config. 0 = disabled, default 3."""
    return max(0, _config.get("stats_file_interval", _FLUSH_INTERVAL))

# Input token pricing ($ per token). Update as models reprice.
# Source: https://claude.com/pricing#api (last verified 2026-03-09)
PRICING = {
    "claude_opus_4_6":    5.00 / 1_000_000,  # Claude Opus 4.6   — $5.00 / 1M input tokens (≤200K ctx)
    "claude_sonnet_4_6":  3.00 / 1_000_000,  # Claude Sonnet 4.6 — $3.00 / 1M input tokens (≤200K ctx)
    "claude_haiku_4_5":   1.00 / 1_000_000,  # Claude Haiku 4.5  — $1.00 / 1M input tokens
    "gpt5_latest":       10.00 / 1_000_000,  # GPT-5 (latest)    — $10.00 / 1M input tokens
}


# ---------------------------------------------------------------------------
# In-memory state (process-lifetime cache)
# ---------------------------------------------------------------------------

class _State:
    """Holds the in-memory accumulator for the current process."""
    def __init__(self):
        self._lock = threading.Lock()
        self._loaded = False
        self._total: int = 0          # cumulative total (disk + in-flight)
        self._unflushed: int = 0      # delta not yet written to disk
        self._encoding_total: int = 0    # cumulative MUNCH encoding savings
        self._encoding_unflushed: int = 0
        self._call_count: int = 0     # calls since last savings flush
        self._stats_call_count: int = 0  # calls since last session_stats.json write
        self._anon_id: Optional[str] = None
        self._base_path: Optional[str] = None
        self._pending_telemetry: int = 0  # unflushed delta for telemetry
        # Session-level tracking (process lifetime only, not persisted)
        self._session_tokens: int = 0
        self._session_calls: int = 0
        self._session_start: float = time.monotonic()
        self._session_tool_breakdown: dict = {}
        # Session-level tool-result cache (LRU, evicted at _RESULT_CACHE_MAXSIZE)
        self._result_cache: OrderedDict = OrderedDict()  # (tool, repo, key) -> result
        self._cache_hits: dict = {}    # tool_name -> hit count
        self._cache_misses: dict = {}  # tool_name -> miss count
        # Per-tool latency ring (process-lifetime; cap _LATENCY_RING_DEFAULT entries)
        self._tool_latencies: dict[str, deque] = {}
        self._tool_errors: dict[str, int] = {}
        # Perf SQLite sink (opt-in via config "perf_telemetry_enabled")
        self._perf_db_path_cached: Optional[Path] = None
        self._perf_db_failed: bool = False
        self._perf_rows_since_trim: int = 0

    def _ensure_loaded(self, base_path: Optional[str]) -> None:
        """Load persisted total from disk (once per process)."""
        if self._loaded:
            return
        self._base_path = base_path
        path = _savings_path(base_path)
        try:
            data = json.loads(path.read_text()) if path.exists() else {}
        except Exception:
            logger.debug("Failed to load savings data from %s", path, exc_info=True)
            data = {}
        self._total = data.get("total_tokens_saved", 0)
        self._encoding_total = data.get("total_encoding_tokens_saved", 0)
        self._anon_id = data.get("anon_id")
        self._loaded = True

    def add(self, delta: int, base_path: Optional[str], tool_name: Optional[str] = None) -> int:
        """Add delta to the running total. Returns new cumulative total."""
        with self._lock:
            self._ensure_loaded(base_path)
            delta = max(0, delta)
            self._total += delta
            self._unflushed += delta
            self._pending_telemetry += delta
            self._session_tokens += delta
            self._session_calls += 1
            if tool_name:
                self._session_tool_breakdown[tool_name] = (
                    self._session_tool_breakdown.get(tool_name, 0) + delta
                )
            self._call_count += 1
            if self._call_count >= _FLUSH_INTERVAL:
                self._flush_locked()
            return self._total

    def session_stats(self, base_path: Optional[str]) -> dict:
        """Return session-level stats (process lifetime)."""
        with self._lock:
            self._ensure_loaded(base_path)
            stats = self._build_stats_locked()
            self._write_session_stats_locked(stats, force=True)
            return stats

    def cache_get(self, tool_name: str, repo: str, specific_key: tuple):
        """Return cached result for (tool, repo, key), or None on miss. Thread-safe."""
        with self._lock:
            full_key = (tool_name, repo, specific_key)
            if full_key in self._result_cache:
                self._result_cache.move_to_end(full_key)
                self._cache_hits[tool_name] = self._cache_hits.get(tool_name, 0) + 1
                return self._result_cache[full_key]
            self._cache_misses[tool_name] = self._cache_misses.get(tool_name, 0) + 1
            return None

    def cache_put(self, tool_name: str, repo: str, specific_key: tuple, result: dict) -> None:
        """Store result in LRU cache. Evicts oldest entry when full. Thread-safe."""
        with self._lock:
            full_key = (tool_name, repo, specific_key)
            self._result_cache[full_key] = result
            self._result_cache.move_to_end(full_key)
            if len(self._result_cache) > _RESULT_CACHE_MAXSIZE:
                self._result_cache.popitem(last=False)

    def cache_invalidate(self, repo: Optional[str] = None) -> int:
        """Evict all entries (repo=None) or entries for a specific repo. Returns evicted count."""
        with self._lock:
            if repo is None:
                count = len(self._result_cache)
                self._result_cache.clear()
                return count
            to_delete = [k for k in self._result_cache if k[1] == repo]
            for k in to_delete:
                del self._result_cache[k]
            return len(to_delete)

    def cache_stats(self) -> dict:
        """Return cache hit/miss stats. Thread-safe."""
        with self._lock:
            total_hits = sum(self._cache_hits.values())
            total_misses = sum(self._cache_misses.values())
            total_lookups = total_hits + total_misses
            by_tool = {}
            all_tools = set(self._cache_hits) | set(self._cache_misses)
            for tool in all_tools:
                h = self._cache_hits.get(tool, 0)
                m = self._cache_misses.get(tool, 0)
                t = h + m
                by_tool[tool] = {
                    "hits": h,
                    "misses": m,
                    "hit_rate": round(h / t, 3) if t else 0.0,
                }
            return {
                "total_hits": total_hits,
                "total_misses": total_misses,
                "hit_rate": round(total_hits / total_lookups, 3) if total_lookups else 0.0,
                "cached_entries": len(self._result_cache),
                "by_tool": by_tool,
            }

    def _build_stats_locked(self) -> dict:
        """Build session stats dict. Must be called with _lock held."""
        elapsed = time.monotonic() - self._session_start
        # Build cache stats inline (re-uses lock already held)
        total_hits = sum(self._cache_hits.values())
        total_misses = sum(self._cache_misses.values())
        total_lookups = total_hits + total_misses
        cache_stats = {
            "total_hits": total_hits,
            "total_misses": total_misses,
            "hit_rate": round(total_hits / total_lookups, 3) if total_lookups else 0.0,
            "cached_entries": len(self._result_cache),
        }
        return {
            "session_tokens_saved": self._session_tokens,
            "session_calls": self._session_calls,
            "session_duration_s": round(elapsed, 1),
            "total_tokens_saved": self._total,
            "tool_breakdown": dict(self._session_tool_breakdown),
            "result_cache": cache_stats,
            "latency_per_tool": self._latency_stats_locked(),
        }

    # ---------------------------------------------------------------------
    # Latency tracking + perf SQLite sink (v1.74.0)
    # ---------------------------------------------------------------------

    def record_latency(
        self,
        tool_name: str,
        duration_ms: float,
        ok: bool = True,
        repo: Optional[str] = None,
    ) -> None:
        """Record a tool-call duration. Called from server.call_tool."""
        try:
            with self._lock:
                ring = self._tool_latencies.get(tool_name)
                if ring is None:
                    ring = deque(maxlen=_LATENCY_RING_DEFAULT)
                    self._tool_latencies[tool_name] = ring
                ring.append(float(duration_ms))
                if not ok:
                    self._tool_errors[tool_name] = self._tool_errors.get(tool_name, 0) + 1
                if _config.get("perf_telemetry_enabled", False):
                    self._persist_perf_locked(tool_name, duration_ms, ok, repo)
        except Exception:
            logger.debug("record_latency failed for %s", tool_name, exc_info=True)

    def _latency_stats_locked(self) -> dict:
        """Compute p50/p95 per tool from the ring. Caller must hold _lock."""
        out: dict = {}
        for tool, ring in self._tool_latencies.items():
            if not ring:
                continue
            sorted_vals = sorted(ring)
            n = len(sorted_vals)
            p50 = sorted_vals[n // 2]
            # p95 index — bisect-style lower bound
            p95_idx = max(0, min(n - 1, int(0.95 * n)))
            p95 = sorted_vals[p95_idx]
            errors = self._tool_errors.get(tool, 0)
            out[tool] = {
                "count": n,
                "p50_ms": round(p50, 2),
                "p95_ms": round(p95, 2),
                "max_ms": round(sorted_vals[-1], 2),
                "errors": errors,
                "error_rate": round(errors / n, 3) if n else 0.0,
            }
        return out

    def latency_stats(self) -> dict:
        """Public latency snapshot. Thread-safe."""
        with self._lock:
            return self._latency_stats_locked()

    def _perf_db_path(self) -> Optional[Path]:
        if self._perf_db_path_cached is not None:
            return self._perf_db_path_cached
        try:
            root = Path(self._base_path) if self._base_path else Path.home() / ".code-index"
            root.mkdir(parents=True, exist_ok=True)
            path = root / _PERF_DB_FILE
            self._perf_db_path_cached = path
            return path
        except Exception:
            logger.debug("Failed to resolve perf db path", exc_info=True)
            return None

    def _ensure_perf_db_locked(self) -> Optional[sqlite3.Connection]:
        """Open the perf SQLite db (create schema on first use)."""
        if self._perf_db_failed:
            return None
        path = self._perf_db_path()
        if path is None:
            return None
        try:
            conn = sqlite3.connect(str(path), timeout=2.0, isolation_level=None)
            conn.execute("PRAGMA journal_mode=WAL")
            conn.execute("""
                CREATE TABLE IF NOT EXISTS tool_calls (
                    ts        REAL NOT NULL,
                    tool      TEXT NOT NULL,
                    duration_ms REAL NOT NULL,
                    ok        INTEGER NOT NULL,
                    repo      TEXT
                )
            """)
            conn.execute("CREATE INDEX IF NOT EXISTS ix_tool_calls_tool ON tool_calls(tool)")
            conn.execute("CREATE INDEX IF NOT EXISTS ix_tool_calls_ts   ON tool_calls(ts)")
            # v1.78.0 — ranking ledger (data-collection only; consumed by
            # the online weight tuner in v1.79.0).
            conn.execute("""
                CREATE TABLE IF NOT EXISTS ranking_events (
                    ts             REAL NOT NULL,
                    repo           TEXT,
                    tool           TEXT NOT NULL,
                    query_hash     TEXT NOT NULL,
                    query          TEXT NOT NULL,
                    returned_ids   TEXT NOT NULL,    -- JSON-encoded list
                    top1_score     REAL,
                    top2_score     REAL,
                    confidence     REAL,
                    semantic_used  INTEGER NOT NULL,
                    identity_hit   INTEGER NOT NULL,
                    repo_is_stale  INTEGER NOT NULL
                )
            """)
            conn.execute("CREATE INDEX IF NOT EXISTS ix_ranking_events_repo ON ranking_events(repo)")
            conn.execute("CREATE INDEX IF NOT EXISTS ix_ranking_events_ts   ON ranking_events(ts)")
            conn.execute("CREATE INDEX IF NOT EXISTS ix_ranking_events_qh   ON ranking_events(query_hash)")
            return conn
        except Exception:
            logger.debug("Failed to open perf db at %s", path, exc_info=True)
            self._perf_db_failed = True
            return None

    def record_ranking_event(
        self,
        *,
        tool: str,
        repo: Optional[str],
        query: str,
        returned_ids: list[str],
        top1_score: Optional[float] = None,
        top2_score: Optional[float] = None,
        confidence: Optional[float] = None,
        semantic_used: bool = False,
        identity_hit: bool = False,
        repo_is_stale: bool = False,
    ) -> None:
        """Append a ranking event to the perf db (no-op when disabled).

        v1.79.0 will use these rows to tune per-repo BM25/semantic weights.
        """
        if not _config.get("perf_telemetry_enabled", False):
            return
        try:
            with self._lock:
                conn = self._ensure_perf_db_locked()
                if conn is None:
                    return
                try:
                    import hashlib
                    import json as _json
                    qh = hashlib.sha1(query.encode("utf-8")).hexdigest()[:16]
                    conn.execute(
                        "INSERT INTO ranking_events "
                        "(ts, repo, tool, query_hash, query, returned_ids, "
                        " top1_score, top2_score, confidence, semantic_used, "
                        " identity_hit, repo_is_stale) "
                        "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                        (
                            time.time(),
                            repo or None,
                            tool,
                            qh,
                            query,
                            _json.dumps(list(returned_ids)[:50]),
                            float(top1_score) if top1_score is not None else None,
                            float(top2_score) if top2_score is not None else None,
                            float(confidence) if confidence is not None else None,
                            1 if semantic_used else 0,
                            1 if identity_hit else 0,
                            1 if repo_is_stale else 0,
                        ),
                    )
                finally:
                    try:
                        conn.close()
                    except Exception:
                        pass
        except Exception:
            logger.debug("record_ranking_event failed for %s", tool, exc_info=True)

    def _persist_perf_locked(
        self,
        tool: str,
        duration_ms: float,
        ok: bool,
        repo: Optional[str],
    ) -> None:
        conn = self._ensure_perf_db_locked()
        if conn is None:
            return
        try:
            conn.execute(
                "INSERT INTO tool_calls (ts, tool, duration_ms, ok, repo) VALUES (?, ?, ?, ?, ?)",
                (time.time(), tool, float(duration_ms), 1 if ok else 0, repo or None),
            )
            self._perf_rows_since_trim += 1
            if self._perf_rows_since_trim >= 1000:
                cap = max(1000, int(_config.get("perf_telemetry_max_rows", _PERF_DB_MAX_ROWS_DEFAULT)))
                conn.execute(
                    "DELETE FROM tool_calls WHERE rowid IN ("
                    "  SELECT rowid FROM tool_calls ORDER BY ts ASC LIMIT MAX(0, "
                    "    (SELECT COUNT(*) FROM tool_calls) - ?"
                    "  )"
                    ")",
                    (cap,),
                )
                self._perf_rows_since_trim = 0
        except Exception:
            logger.debug("Failed to persist perf row for %s", tool, exc_info=True)
        finally:
            try:
                conn.close()
            except Exception:
                pass

    def _write_session_stats_locked(self, stats: dict, force: bool = False) -> None:
        """Write session stats to ~/.code-index/session_stats.json. Must be called with _lock held.

        Writes are gated by stats_file_interval config (default 3 calls).
        Set to 0 to disable all session_stats.json writes.
        Pass force=True to bypass the interval check (e.g. explicit get_session_stats call).
        """
        stats_file_interval = _get_stats_file_interval()
        if stats_file_interval == 0 and not force:
            return
        self._stats_call_count += 1
        if not force and self._stats_call_count < stats_file_interval:
            return
        self._stats_call_count = 0
        path = _session_stats_path(self._base_path)
        try:
            payload = {**stats, "last_updated": datetime.now(timezone.utc).isoformat()}
            path.write_text(json.dumps(payload, indent=2))
        except Exception:
            logger.debug("Failed to write session stats to %s", path, exc_info=True)

    def get_total(self, base_path: Optional[str]) -> int:
        with self._lock:
            self._ensure_loaded(base_path)
            return self._total

    def _flush_locked(self) -> None:
        """Write accumulated total to disk. Must be called with _lock held."""
        if self._unflushed == 0 and self._loaded:
            self._call_count = 0
            self._write_session_stats_locked(self._build_stats_locked())
            return
        path = _savings_path(self._base_path)
        try:
            data = json.loads(path.read_text()) if path.exists() else {}
        except Exception:
            logger.debug("Failed to read savings file for flush: %s", path, exc_info=True)
            data = {}
        if self._anon_id is None:
            if "anon_id" not in data:
                data["anon_id"] = str(uuid.uuid4())
            self._anon_id = data["anon_id"]
        else:
            data["anon_id"] = self._anon_id
        data["total_tokens_saved"] = data.get("total_tokens_saved", 0) + self._unflushed
        if self._encoding_unflushed:
            data["total_encoding_tokens_saved"] = (
                data.get("total_encoding_tokens_saved", 0) + self._encoding_unflushed
            )
            self._encoding_unflushed = 0
        try:
            path.write_text(json.dumps(data))
        except Exception:
            logger.debug("Failed to write savings data to %s", path, exc_info=True)

        # Send batched telemetry. We send the absolute lifetime total (just
        # written above) so the server can GREATEST-converge and recover any
        # previously dropped sends; the delta rides along for legacy servers.
        if self._pending_telemetry > 0 and _config.get("share_savings", True):
            _share_savings(self._pending_telemetry, data["total_tokens_saved"], self._anon_id)
            self._pending_telemetry = 0

        self._unflushed = 0
        self._call_count = 0
        self._write_session_stats_locked(self._build_stats_locked())

    def flush(self) -> None:
        """Public flush — called at atexit."""
        with self._lock:
            if self._loaded:
                self._flush_locked()


_state = _State()
atexit.register(_state.flush)


def _signal_flush(signum, frame):
    """Flush savings to disk on SIGTERM/SIGINT, then re-raise the signal."""
    _state.flush()
    # Restore the default handler and re-raise so the process exits normally.
    signal.signal(signum, signal.SIG_DFL)
    os.kill(os.getpid(), signum)


# MCP servers are commonly killed via SIGTERM (pipe close, client shutdown).
# atexit does NOT run on SIGTERM, so we register explicit handlers here.
# We only install if no handler is already set (respects user overrides).
for _sig in (signal.SIGTERM, signal.SIGINT):
    try:
        if signal.getsignal(_sig) in (signal.SIG_DFL, None):
            signal.signal(_sig, _signal_flush)
    except (OSError, ValueError):
        # Signals can't be set in non-main threads; ignore safely.
        pass


# ---------------------------------------------------------------------------
# Public API (unchanged signatures)
# ---------------------------------------------------------------------------

def _savings_path(base_path: Optional[str] = None) -> Path:
    root = Path(base_path) if base_path else Path.home() / ".code-index"
    root.mkdir(parents=True, exist_ok=True)
    return root / _SAVINGS_FILE


def _session_stats_path(base_path: Optional[str] = None) -> Path:
    root = Path(base_path) if base_path else Path.home() / ".code-index"
    root.mkdir(parents=True, exist_ok=True)
    return root / _SESSION_STATS_FILE


# ---------------------------------------------------------------------------
# Telemetry worker (P11)
# ---------------------------------------------------------------------------
# A single long-lived daemon thread drains _telemetry_queue instead of
# spawning a new thread on every flush.  This eliminates per-flush thread
# creation overhead and prevents thread pile-up under rapid calls.
# ---------------------------------------------------------------------------

_telemetry_queue: queue.Queue = queue.Queue()
# The worker is started lazily on the first enqueue (see _ensure_telemetry_worker)
# rather than at import. Two reasons: (1) `import jcodemunch_mcp` must have no
# background-thread side effect — an import-time daemon thread is exactly the
# undisclosed-persistent-service shape that package scanners flag; (2) a process
# that has opted out (JCODEMUNCH_SHARE_SAVINGS=0) never reaches _share_savings,
# so it never spawns the thread at all.
_telemetry_worker_lock = threading.Lock()
_telemetry_worker_started = False


def _telemetry_url() -> str:
    """Community-meter endpoint, overridable via JCODEMUNCH_TELEMETRY_URL so a
    deployment can redirect it without a code change. Resolved at send time so an
    env change takes effect for an already-running process."""
    return os.environ.get("JCODEMUNCH_TELEMETRY_URL") or _TELEMETRY_URL


def _telemetry_worker() -> None:
    """Drain _telemetry_queue and POST each item. Runs for process lifetime."""
    while True:
        item = _telemetry_queue.get()
        if item is None:  # shutdown sentinel
            break
        delta, total, anon_id = item
        try:
            import httpx
            # `total` is the absolute lifetime savings; the server applies
            # GREATEST(stored, total) so a failed post self-corrects on the next
            # one. `delta` rides along for older server builds (additive upsert).
            httpx.post(
                _telemetry_url(),
                json={"delta": delta, "total": total, "anon_id": anon_id},
                timeout=3.0,
            )
        except Exception:
            logger.debug("Telemetry post failed", exc_info=True)
        finally:
            _telemetry_queue.task_done()


def _ensure_telemetry_worker() -> None:
    """Start the single telemetry worker thread once, on first use. Thread-safe
    (double-checked under a lock). Never started at import or when the user has
    opted out of sharing."""
    global _telemetry_worker_started
    if _telemetry_worker_started:
        return
    with _telemetry_worker_lock:
        if _telemetry_worker_started:
            return
        threading.Thread(
            target=_telemetry_worker, daemon=True, name="jcodemunch-telemetry"
        ).start()
        _telemetry_worker_started = True


def _share_savings(delta: int, total: int, anon_id: str) -> None:
    """Enqueue a fire-and-forget POST to the community meter. Never raises.
    `total` is the absolute lifetime savings (self-correcting); `delta` is the
    new savings since the last send (legacy additive path). Starts the worker
    thread on first use — never at import, and never when sharing is disabled
    (this function is only reached behind the share_savings opt-out gate)."""
    _ensure_telemetry_worker()
    _telemetry_queue.put((delta, total, anon_id))


def record_encoding_savings(
    tokens_saved: int,
    base_path: Optional[str] = None,
    tool_name: Optional[str] = None,
) -> int:
    """Add tokens saved by MUNCH compact encoding. Tracked independently
    from retrieval-side savings. Returns new cumulative encoding total."""
    with _state._lock:
        _state._ensure_loaded(base_path)
        delta = max(0, tokens_saved)
        _state._encoding_total += delta
        _state._encoding_unflushed += delta
        _state._call_count += 1
        if _state._call_count >= _FLUSH_INTERVAL:
            _state._flush_locked()
        return _state._encoding_total


def get_total_encoding_saved(base_path: Optional[str] = None) -> int:
    with _state._lock:
        _state._ensure_loaded(base_path)
        return _state._encoding_total


def record_savings(tokens_saved: int, base_path: Optional[str] = None, tool_name: Optional[str] = None) -> int:
    """Add tokens_saved to the running total. Returns new cumulative total.

    Uses an in-memory accumulator; flushes to disk every FLUSH_INTERVAL calls (currently 3) and at exit.
    """
    return _state.add(tokens_saved, base_path, tool_name)


def write_pulse(tool_name: str, tokens_saved: int = 0, base_path: Optional[str] = None) -> None:
    """Write a per-call pulse file for downstream consumers (dashboards, monitors).

    Atomic write of a small JSON file to {storage}/_pulse.json containing the
    tool name, timestamp, and running counters. Only written when
    JCODEMUNCH_EVENT_LOG=1 is set.
    """
    if not os.environ.get("JCODEMUNCH_EVENT_LOG"):
        return
    try:
        root = Path(base_path) if base_path else Path.home() / ".code-index"
        pulse_path = root / _PULSE_FILE
        with _state._lock:
            calls = _state._session_calls
            session_tokens = _state._session_tokens
        data = {
            "last_call_at": datetime.now(timezone.utc).isoformat(),
            "tool": tool_name,
            "calls_since_boot": calls,
            "session_tokens_saved": session_tokens,
            "tokens_saved": tokens_saved,
        }
        tmp = pulse_path.with_suffix(".tmp")
        tmp.write_text(json.dumps(data, separators=(",", ":")), encoding="utf-8")
        tmp.replace(pulse_path)
    except Exception:
        logger.debug("Pulse write failed", exc_info=True)


def get_session_stats(base_path: Optional[str] = None) -> dict:
    """Return token savings stats for the current session (process lifetime).

    Returns session_tokens_saved, session_calls, session_duration_s,
    total_tokens_saved (all-time), tool_breakdown, and cost_avoided estimates.
    """
    stats = _state.session_stats(base_path)
    session_tokens = stats["session_tokens_saved"]
    total_tokens = stats["total_tokens_saved"]
    return {
        **stats,
        "session_cost_avoided": {
            model: round(session_tokens * rate, 4)
            for model, rate in PRICING.items()
        },
        "total_cost_avoided": {
            model: round(total_tokens * rate, 4)
            for model, rate in PRICING.items()
        },
        "runtime_signal": _runtime_signal_summary(base_path),
    }


def _runtime_signal_summary(base_path: Optional[str] = None) -> dict:
    """Cheap probe of the runtime_* tables across all per-repo databases.

    Returns ``{"rows": N, "by_source": {source: count}}``. Phase 0 always
    returns zeroes (no ingest yet); the field exists so Phase 1+ can fill
    it without changing the response shape.

    Errors are logged at DEBUG and yield zeroes — never block stats.
    """
    import sqlite3 as _sqlite3
    summary = {"rows": 0, "by_source": {}}
    try:
        root = Path(base_path) if base_path else Path.home() / ".code-index"
        if not root.exists():
            return summary
        # Skip non-repo files (telemetry.db, config.jsonc, etc.)
        non_repo = {"telemetry.db"}
        total = 0
        by_source: dict[str, int] = {}
        for db_path in root.glob("*.db"):
            if db_path.name in non_repo:
                continue
            try:
                conn = _sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
                conn.row_factory = _sqlite3.Row
                # Skip dbs that don't have the runtime_calls table yet
                row = conn.execute(
                    "SELECT name FROM sqlite_master WHERE type='table' AND name='runtime_calls'"
                ).fetchone()
                if row is None:
                    conn.close()
                    continue
                rows = conn.execute(
                    "SELECT source, COALESCE(SUM(count), 0) AS n FROM runtime_calls GROUP BY source"
                ).fetchall()
                for r in rows:
                    by_source[r["source"]] = by_source.get(r["source"], 0) + int(r["n"])
                    total += int(r["n"])
                conn.close()
            except _sqlite3.Error:
                logger.debug("Runtime signal probe failed on %s", db_path, exc_info=True)
        summary["rows"] = total
        summary["by_source"] = by_source
    except Exception:
        logger.debug("Runtime signal summary failed", exc_info=True)
    return summary


def get_total_saved(base_path: Optional[str] = None) -> int:
    """Return the current cumulative total without modifying it."""
    return _state.get_total(base_path)


def result_cache_get(tool_name: str, repo: str, specific_key: tuple):
    """Return a cached tool result, or None on miss. Updates hit/miss counters."""
    return _state.cache_get(tool_name, repo, specific_key)


def result_cache_put(tool_name: str, repo: str, specific_key: tuple, result: dict) -> None:
    """Store a tool result in the session LRU cache (max 256 entries)."""
    _state.cache_put(tool_name, repo, specific_key, result)


def result_cache_invalidate(repo: Optional[str] = None) -> int:
    """Evict cached results — all repos (default) or a specific repo. Returns evicted count."""
    return _state.cache_invalidate(repo)


def result_cache_stats() -> dict:
    """Return cache hit/miss stats for the current session."""
    return _state.cache_stats()


def record_tool_latency(
    tool_name: str,
    duration_ms: float,
    ok: bool = True,
    repo: Optional[str] = None,
) -> None:
    """Record a tool-call duration for the current session (and optional perf db)."""
    _state.record_latency(tool_name, duration_ms, ok=ok, repo=repo)


def latency_stats() -> dict:
    """Return per-tool p50/p95/error_rate from the in-memory ring."""
    return _state.latency_stats()


def perf_db_path(base_path: Optional[str] = None) -> Path:
    """Return the perf telemetry SQLite path (creating its parent dir)."""
    root = Path(base_path) if base_path else Path.home() / ".code-index"
    root.mkdir(parents=True, exist_ok=True)
    return root / _PERF_DB_FILE


def record_ranking_event(**kwargs) -> None:
    """Append a ranking event to telemetry.db (no-op when telemetry disabled).

    Keyword args: tool, repo, query, returned_ids, top1_score, top2_score,
    confidence, semantic_used, identity_hit, repo_is_stale.
    """
    _state.record_ranking_event(**kwargs)


def ranking_db_query(
    base_path: Optional[str] = None,
    window_seconds: Optional[float] = None,
    repo: Optional[str] = None,
    tool: Optional[str] = None,
    limit: int = 1000,
) -> list[tuple]:
    """Read recent ranking events from telemetry.db.

    Returns a list of tuples shaped exactly as the SELECT below — the
    order matches the v1.78.0 ranking_events schema. Empty when the db
    doesn't exist (telemetry disabled or never written).
    """
    path = perf_db_path(base_path)
    if not path.exists():
        return []
    try:
        conn = sqlite3.connect(str(path), timeout=2.0)
        try:
            sql = (
                "SELECT ts, repo, tool, query_hash, query, returned_ids, "
                "top1_score, top2_score, confidence, semantic_used, "
                "identity_hit, repo_is_stale FROM ranking_events"
            )
            args: list = []
            clauses: list[str] = []
            if window_seconds is not None:
                clauses.append("ts >= ?")
                args.append(time.time() - float(window_seconds))
            if repo:
                clauses.append("repo = ?")
                args.append(repo)
            if tool:
                clauses.append("tool = ?")
                args.append(tool)
            if clauses:
                sql += " WHERE " + " AND ".join(clauses)
            sql += " ORDER BY ts DESC LIMIT ?"
            args.append(int(limit))
            return conn.execute(sql, args).fetchall()
        finally:
            conn.close()
    except sqlite3.OperationalError:
        # Schema not yet created (telemetry was never enabled in this
        # storage dir, even though the file exists from another component).
        return []
    except Exception:
        logger.debug("ranking_db_query failed at %s", path, exc_info=True)
        return []


def perf_db_query(
    base_path: Optional[str] = None,
    window_seconds: Optional[float] = None,
    tool: Optional[str] = None,
) -> list[tuple]:
    """Read recent perf rows from telemetry.db for the analyze_perf tool.

    Returns a list of (ts, tool, duration_ms, ok, repo). Empty if the db
    doesn't exist yet (perf telemetry never enabled or never written).
    """
    path = perf_db_path(base_path)
    if not path.exists():
        return []
    try:
        conn = sqlite3.connect(str(path), timeout=2.0)
        try:
            sql = "SELECT ts, tool, duration_ms, ok, repo FROM tool_calls"
            args: list = []
            clauses: list[str] = []
            if window_seconds is not None:
                clauses.append("ts >= ?")
                args.append(time.time() - float(window_seconds))
            if tool:
                clauses.append("tool = ?")
                args.append(tool)
            if clauses:
                sql += " WHERE " + " AND ".join(clauses)
            sql += " ORDER BY ts DESC"
            rows = conn.execute(sql, args).fetchall()
            return rows
        finally:
            conn.close()
    except Exception:
        logger.debug("perf_db_query failed at %s", path, exc_info=True)
        return []


def estimate_savings(raw_bytes: int, response_bytes: int) -> int:
    """Estimate tokens saved: (raw - response) / bytes_per_token."""
    return max(0, (raw_bytes - response_bytes) // _BYTES_PER_TOKEN)


def cost_avoided(tokens_saved: int, total_tokens_saved: int) -> dict:
    """Formerly returned per-call cost breakdowns for _meta envelopes.

    Now returns an empty dict — cost detail is available via get_session_stats
    only. Removing 4-model cost tables from every per-tool _meta response
    reduces conversation-history token overhead by ~70 tokens/call.
    """
    return {}
