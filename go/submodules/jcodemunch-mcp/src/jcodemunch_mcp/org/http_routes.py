"""HTTP ingest route for org-rollup telemetry (team SKU).

``POST /org/report`` — a seat ships its savings to the org host, which records
them via :func:`record_seat_report`. The cross-machine transport over the
transport-agnostic store.

Security model mirrors the runtime ingest routes: **off by default**, two keys
to turn:

1. ``JCODEMUNCH_HTTP_TOKEN`` — bearer auth, enforced by the existing Starlette
   auth middleware in front of every route (not per-handler).
2. ``JCODEMUNCH_ORG_INGEST_ENABLED=1`` — explicit opt-in to the write side.

Body size cap + gzip handling are reused from the runtime ingest helpers.
"""

from __future__ import annotations

import asyncio
import json
import logging
from typing import TYPE_CHECKING

from .. import config as _config_mod
from ..runtime.http_routes import _read_body, _json, _ingest_auth_error
from .store import record_seat_report

if TYPE_CHECKING:
    from starlette.requests import Request
    from starlette.responses import JSONResponse

logger = logging.getLogger(__name__)


def _org_ingest_enabled() -> bool:
    """Whether the org host accepts seat reports over HTTP."""
    val = _config_mod.get("org_ingest_enabled", False)
    if isinstance(val, str):
        return val.strip().lower() in ("1", "true", "yes", "on")
    return bool(val)


async def handle_org_report(request: "Request") -> "JSONResponse":
    """``POST /org/report`` — record a seat's savings under its org."""
    if not _org_ingest_enabled():
        return _json(
            {"error": "org ingest is disabled; set JCODEMUNCH_ORG_INGEST_ENABLED=1 on the org host"},
            status=403,
        )
    auth_err = _ingest_auth_error()
    if auth_err is not None:
        return auth_err
    body, err = await _read_body(request)
    if err:
        return _json({"error": err}, status=413)
    try:
        payload = json.loads(body or b"{}")
    except (ValueError, TypeError):
        return _json({"error": "invalid JSON body"}, status=400)
    if not isinstance(payload, dict):
        return _json({"error": "body must be a JSON object"}, status=400)

    org_id = str(payload.get("org_id", "")).strip()
    seat_id = str(payload.get("seat_id", "")).strip()
    if not org_id or not seat_id:
        return _json({"error": "org_id and seat_id are required"}, status=400)
    try:
        tokens = int(payload.get("tokens_saved", 0))
        usd = float(payload.get("usd", 0.0))
        calls = int(payload.get("calls", 0))
    except (TypeError, ValueError):
        return _json({"error": "tokens_saved/usd/calls must be numeric"}, status=400)
    date = payload.get("date") or None

    try:
        result = await asyncio.to_thread(
            record_seat_report, org_id, seat_id, tokens, usd, calls, date=date,
        )
    except ValueError as exc:
        return _json({"error": str(exc)}, status=400)
    except Exception as exc:  # pragma: no cover — surfaced as 500
        logger.warning("org report ingest failed for %s/%s: %s", org_id, seat_id, exc, exc_info=True)
        return _json({"error": f"record failed: {exc}"}, status=500)
    return _json({"success": True, **result})


def make_org_routes() -> list:
    """Build the Starlette Route for the org ingest endpoint. ``[]`` if starlette
    isn't installed (imported lazily, like the runtime routes)."""
    try:
        from starlette.routing import Route
    except ImportError:
        return []
    return [Route("/org/report", endpoint=handle_org_report, methods=["POST"])]
