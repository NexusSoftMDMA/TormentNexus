"""Phase 6 — HTTP live-ingest endpoint route handlers.

Mounts three POST routes on the existing Starlette HTTP transport so
production systems can ship runtime signals to a running jcm instance
in real time instead of via nightly file imports:

* ``POST /runtime/otel``    — OTLP/JSON spans (Phase 1 wire format)
* ``POST /runtime/sql``     — pg_stat_statements CSV / JSON-Lines (Phase 4)
* ``POST /runtime/stack``   — Python / JVM / Node stacks (Phase 5)

The handlers are deliberately thin: parse → hand to the corresponding
``ingest_*_stream`` orchestrator. **Same** redaction chokepoint, **same**
upserts, **same** FIFO eviction, **same** response envelope as the
file-based ``import-trace`` CLI. The HTTP and file paths are
interchangeable.

Security model
--------------

The endpoint is **off by default**. Two keys must turn before traffic
flows:

1. ``JCODEMUNCH_HTTP_TOKEN`` — bearer auth (already required by the
   existing HTTP MCP transport when bound to a non-loopback host). The
   route handlers themselves don't enforce auth — that's the job of the
   existing ``_make_auth_middleware()`` Starlette middleware which already
   sits in front of every Starlette route in the transport.
2. ``JCODEMUNCH_RUNTIME_INGEST_ENABLED=1`` — explicit opt-in to the write
   side. Read-only MCP tools don't enable the ingest endpoints — operators
   have to flip the second flag.

Per-request body size cap (default 5 MB; configurable via
``JCODEMUNCH_RUNTIME_INGEST_MAX_BODY_BYTES``) prevents DoS via giant
payloads. ``Content-Encoding: gzip`` is honoured so collectors that
compress on the wire don't have to decompress before forwarding.

Concurrency
-----------

A per-repo asyncio.Lock serialises writes against the same SQLite
database file. SQLite's WAL mode would handle multiple writers via
BEGIN IMMEDIATE retry, but explicit serialisation is cheaper than
retry-storms under load and keeps the upsert order deterministic. One
lock per ``(owner, name)`` pair; idle locks are garbage-collected when
no longer referenced.
"""

from __future__ import annotations

import asyncio
import gzip
import logging
import os
from collections import OrderedDict
from typing import TYPE_CHECKING, Any, Optional

from .. import config as _config_mod
from ..storage import IndexStore
from .ingest import ingest_otel_stream
from .sql_ingest import ingest_sql_log_stream
from .stack_ingest import ingest_stack_log_stream

if TYPE_CHECKING:
    from starlette.requests import Request
    from starlette.responses import JSONResponse

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Per-repo write mutex
# ---------------------------------------------------------------------------


class _RepoLockRegistry:
    """LRU-bounded registry of asyncio.Lock objects keyed by owner/name.

    A small bound (256 repos) is plenty in practice — the registry only
    grows for repos with active inflight ingest requests. The LRU is a
    safety valve against a memory leak if some upstream caller fires a
    request per random slug.
    """

    __slots__ = ("_locks", "_max_size", "_outer_lock")

    def __init__(self, max_size: int = 256) -> None:
        self._locks: OrderedDict[str, asyncio.Lock] = OrderedDict()
        self._max_size = max_size
        self._outer_lock = asyncio.Lock()

    async def get(self, owner: str, name: str) -> asyncio.Lock:
        key = f"{owner}/{name}"
        async with self._outer_lock:
            lock = self._locks.get(key)
            if lock is None:
                lock = asyncio.Lock()
                self._locks[key] = lock
            else:
                self._locks.move_to_end(key)
            while len(self._locks) > self._max_size:
                # Evict the LRU repo lock. Safe even if another coroutine
                # is awaiting it — they hold their own reference.
                self._locks.popitem(last=False)
            return lock


_REGISTRY = _RepoLockRegistry()


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _runtime_ingest_enabled() -> bool:
    """Whether the HTTP ingest routes should accept traffic."""
    val = _config_mod.get("runtime_ingest_enabled", False)
    if isinstance(val, str):
        return val.strip().lower() in ("1", "true", "yes", "on")
    return bool(val)


def _max_body_bytes() -> int:
    """Per-request body cap (post-decompression)."""
    val = _config_mod.get("runtime_ingest_max_body_bytes", 5_242_880)
    try:
        return max(1024, int(val))
    except (TypeError, ValueError):
        return 5_242_880


def _http_auth_token_present() -> bool:
    """Whether bearer auth is actually enforceable.

    ``BearerAuthMiddleware`` (server.py) only checks the token *if*
    ``JCODEMUNCH_HTTP_TOKEN`` is set; when it's unset the middleware is a no-op
    and an enabled write endpoint would accept unauthenticated writes. We read
    the same source the middleware reads so the fail-closed decision here agrees
    with what the middleware would (or wouldn't) enforce.
    """
    return bool(os.environ.get("JCODEMUNCH_HTTP_TOKEN"))


def _ingest_auth_error() -> Optional["JSONResponse"]:
    """Fail closed when an ingest endpoint is enabled but no token is set.

    Enabling ingest is the first key of the documented two-key turn; the bearer
    token is the second. Without it the write endpoint would accept
    unauthenticated writes (only a startup warning today), so we refuse (503)
    rather than warn. Returns ``None`` when a token is set (the middleware then
    enforces it), else the 503 response to short-circuit the handler.
    """
    if _http_auth_token_present():
        return None
    return _json(
        {
            "error": (
                "ingest endpoint enabled but JCODEMUNCH_HTTP_TOKEN is not set; "
                "refusing unauthenticated writes. Set JCODEMUNCH_HTTP_TOKEN on the "
                "transport (the second key of the two-key turn) to enable writes."
            )
        },
        status=503,
    )


def _resolve_repo_param(request: "Request") -> Optional[tuple[str, str]]:
    """Extract owner/name from ``X-JCM-Repo`` header or ``?repo=`` query.

    Returns ``None`` when neither is set; the caller emits a 400.
    """
    candidate = (
        request.headers.get("x-jcm-repo")
        or request.query_params.get("repo")
        or ""
    ).strip()
    if not candidate:
        return None
    if "/" in candidate:
        owner, _, name = candidate.partition("/")
        owner = owner.strip()
        name = name.strip()
        if owner and name:
            return owner, name
    # Bare names default to the "local" owner (matches the index_folder convention)
    if candidate:
        return "local", candidate
    return None


async def _read_body(request: "Request") -> tuple[Optional[bytes], Optional[str]]:
    """Read the request body, honouring Content-Encoding: gzip + size cap.

    Returns ``(body, error)``. On size-cap violation returns
    ``(None, "...")`` so the caller emits a 413.
    """
    cap = _max_body_bytes()
    encoding = (request.headers.get("content-encoding") or "").strip().lower()
    raw = await request.body()
    if len(raw) > cap and encoding != "gzip":
        return None, (
            f"request body too large: {len(raw)} bytes > {cap} cap. "
            f"Increase JCODEMUNCH_RUNTIME_INGEST_MAX_BODY_BYTES or split the payload."
        )
    if encoding == "gzip":
        try:
            decoded = gzip.decompress(raw)
        except OSError as exc:
            return None, f"gzip decode failed: {exc}"
        if len(decoded) > cap:
            return None, (
                f"decompressed body too large: {len(decoded)} bytes > {cap} cap "
                f"(on-wire was {len(raw)} bytes — declared Content-Encoding: gzip). "
                f"This is the gzip-bomb guard; tune JCODEMUNCH_RUNTIME_INGEST_MAX_BODY_BYTES "
                f"if your real workload genuinely needs more."
            )
        return decoded, None
    return raw, None


def _json(payload: dict[str, Any], status: int = 200) -> "JSONResponse":
    from starlette.responses import JSONResponse
    return JSONResponse(payload, status_code=status)


def _shared_handler_setup(
    request: "Request",
) -> tuple[Optional[tuple[str, str, str]], Optional["JSONResponse"]]:
    """Run the gate / repo-resolve / body-read sequence shared by all 3 routes.

    Returns ``((owner, name, body), None)`` on success or
    ``(None, JSONResponse)`` with the appropriate error response.
    """
    if not _runtime_ingest_enabled():
        return None, _json(
            {
                "error": (
                    "runtime ingest endpoint is disabled. Set "
                    "JCODEMUNCH_RUNTIME_INGEST_ENABLED=1 (or runtime_ingest_enabled=true "
                    "in config.jsonc) AND ensure JCODEMUNCH_HTTP_TOKEN is set on the "
                    "transport."
                )
            },
            status=503,
        )
    auth_err = _ingest_auth_error()
    if auth_err is not None:
        return None, auth_err
    repo = _resolve_repo_param(request)
    if repo is None:
        return None, _json(
            {
                "error": (
                    "missing repo identifier. Pass X-JCM-Repo: owner/name header "
                    "or ?repo=owner/name query string."
                )
            },
            status=400,
        )
    owner, name = repo
    return (owner, name, ""), None


async def _resolve_db_path(owner: str, name: str) -> Optional[str]:
    """Locate the per-repo SQLite DB. Returns None when the repo isn't indexed."""
    storage_path = os.environ.get("CODE_INDEX_PATH")
    store = IndexStore(base_path=storage_path)
    db_path = store._sqlite._db_path(owner, name)  # type: ignore[attr-defined]
    if not db_path.exists():
        return None
    return str(db_path)


# ---------------------------------------------------------------------------
# Route handlers
# ---------------------------------------------------------------------------


async def handle_otel(request: "Request") -> "JSONResponse":
    """``POST /runtime/otel`` — accept OTLP/JSON spans and ingest them."""
    setup, err_resp = _shared_handler_setup(request)
    if err_resp is not None:
        return err_resp
    owner, name, _ = setup  # type: ignore[misc]
    body, body_err = await _read_body(request)
    if body_err is not None:
        return _json({"error": body_err}, status=413)
    db_path = await _resolve_db_path(owner, name)
    if db_path is None:
        return _json(
            {"error": f"repository {owner}/{name} is not indexed; run `jcodemunch-mcp index` first."},
            status=404,
        )
    text = (body or b"").decode("utf-8", errors="replace")
    return await _run_with_lock(
        owner,
        name,
        lambda: ingest_otel_stream(
            db_path=db_path,
            text=text,
            redact_enabled=bool(_config_mod.get("runtime_redact_enabled", True)),
            max_rows=int(_config_mod.get("runtime_max_rows", 100_000)),
        ),
        source="otel",
    )


async def handle_sql(request: "Request") -> "JSONResponse":
    """``POST /runtime/sql`` — accept pg_stat_statements / SQL JSON-Lines."""
    setup, err_resp = _shared_handler_setup(request)
    if err_resp is not None:
        return err_resp
    owner, name, _ = setup  # type: ignore[misc]
    body, body_err = await _read_body(request)
    if body_err is not None:
        return _json({"error": body_err}, status=413)
    db_path = await _resolve_db_path(owner, name)
    if db_path is None:
        return _json(
            {"error": f"repository {owner}/{name} is not indexed; run `jcodemunch-mcp index` first."},
            status=404,
        )
    text = (body or b"").decode("utf-8", errors="replace")
    fmt = (request.query_params.get("fmt") or "auto").strip().lower()
    if fmt not in ("auto", "csv", "jsonl"):
        return _json({"error": f"unknown ?fmt={fmt!r}; valid: auto / csv / jsonl"}, status=400)
    return await _run_with_lock(
        owner,
        name,
        lambda: ingest_sql_log_stream(
            db_path=db_path,
            text=text,
            fmt=fmt,
            redact_enabled=bool(_config_mod.get("runtime_redact_enabled", True)),
            max_rows=int(_config_mod.get("runtime_max_rows", 100_000)),
        ),
        source="sql_log",
    )


async def handle_stack(request: "Request") -> "JSONResponse":
    """``POST /runtime/stack`` — accept Python / JVM / Node.js stack logs."""
    setup, err_resp = _shared_handler_setup(request)
    if err_resp is not None:
        return err_resp
    owner, name, _ = setup  # type: ignore[misc]
    body, body_err = await _read_body(request)
    if body_err is not None:
        return _json({"error": body_err}, status=413)
    db_path = await _resolve_db_path(owner, name)
    if db_path is None:
        return _json(
            {"error": f"repository {owner}/{name} is not indexed; run `jcodemunch-mcp index` first."},
            status=404,
        )
    text = (body or b"").decode("utf-8", errors="replace")
    fmt = (request.query_params.get("fmt") or "auto").strip().lower()
    if fmt not in ("auto", "plain", "jsonl"):
        return _json({"error": f"unknown ?fmt={fmt!r}; valid: auto / plain / jsonl"}, status=400)
    return await _run_with_lock(
        owner,
        name,
        lambda: ingest_stack_log_stream(
            db_path=db_path,
            text=text,
            fmt=fmt,
            redact_enabled=bool(_config_mod.get("runtime_redact_enabled", True)),
            max_rows=int(_config_mod.get("runtime_max_rows", 100_000)),
        ),
        source="stack_log",
    )


async def _run_with_lock(
    owner: str,
    name: str,
    work,
    *,
    source: str,
) -> "JSONResponse":
    """Take the per-repo write lock; offload the synchronous ingest to a thread."""
    lock = await _REGISTRY.get(owner, name)
    async with lock:
        try:
            result = await asyncio.to_thread(work)
        except FileNotFoundError as exc:
            return _json({"error": str(exc)}, status=404)
        except Exception as exc:  # pragma: no cover — surfaced as 500
            logger.warning("ingest %s failed for %s/%s: %s", source, owner, name, exc, exc_info=True)
            return _json({"error": f"ingest failed: {exc}"}, status=500)
    return _json({"success": True, "repo": f"{owner}/{name}", "source": source, **result})


def make_runtime_routes() -> list:
    """Build the Starlette Route objects for the three runtime POST endpoints.

    Imported lazily so users without the ``[http]`` extra don't pay the
    import cost. Returns ``[]`` when starlette isn't installed.
    """
    try:
        from starlette.routing import Route
    except ImportError:
        return []
    return [
        Route("/runtime/otel", endpoint=handle_otel, methods=["POST"]),
        Route("/runtime/sql", endpoint=handle_sql, methods=["POST"]),
        Route("/runtime/stack", endpoint=handle_stack, methods=["POST"]),
    ]
