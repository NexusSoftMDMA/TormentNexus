"""Seat-side org reporting: compute this seat's savings and record it.

Reuses the receipt machinery (transcript-derived savings) for the payload.
Records to the local org store today (this machine acting as the org host); the
remote HTTP transport (POST to JCODEMUNCH_ORG_ENDPOINT) is the next increment
and will reuse the same record_seat_report sink on the host.
"""

from __future__ import annotations

import os
import socket
from typing import Optional

from .store import record_seat_report


def _seat_id() -> str:
    return os.environ.get("JCODEMUNCH_CLIENT_ID") or socket.gethostname() or "unknown-seat"


def _post_to_endpoint(endpoint: str, payload: dict) -> dict:
    """Ship a seat report to a remote org host's POST /org/report."""
    url = endpoint.rstrip("/") + "/org/report"
    headers = {"Content-Type": "application/json"}
    token = os.environ.get("JCODEMUNCH_HTTP_TOKEN")
    if token:
        headers["Authorization"] = f"Bearer {token}"
    try:
        import httpx
        resp = httpx.post(url, json=payload, headers=headers, timeout=15)
    except Exception as exc:
        return {"reported": False, "transport": "http", "endpoint": url, "error": str(exc), **payload}
    ok = resp.status_code == 200
    out = {"reported": ok, "transport": "http", "endpoint": url, "status_code": resp.status_code, **payload}
    if not ok:
        out["error"] = resp.text[:300]
    return out


def run_org_report(
    *,
    model: str = "opus",
    org_id: Optional[str] = None,
    seat_id: Optional[str] = None,
    endpoint: Optional[str] = None,
    storage_path: Optional[str] = None,
) -> dict:
    """Compute all-time savings for this seat and report them under its org.

    If an endpoint is given (``--endpoint`` or ``JCODEMUNCH_ORG_ENDPOINT``), POST
    to the remote org host; otherwise record into the local store (this machine
    acting as the org host).
    """
    org_id = org_id or os.environ.get("JCODEMUNCH_ORG_ID", "")
    if not org_id:
        return {"error": "JCODEMUNCH_ORG_ID is not set", "hint": "set JCODEMUNCH_ORG_ID to your org identifier"}
    seat_id = seat_id or _seat_id()

    from ..cli.receipt import iter_calls, aggregate, dollar_savings, _projects_root

    agg = aggregate(iter_calls(_projects_root()))
    tokens = int(agg["totals"]["savings_tokens"])
    calls = int(agg["totals"]["calls"])
    usd = dollar_savings(tokens, model)
    payload = {"org_id": org_id, "seat_id": seat_id, "tokens_saved": tokens, "usd": usd, "calls": calls}

    endpoint = endpoint or os.environ.get("JCODEMUNCH_ORG_ENDPOINT", "")
    if endpoint:
        return _post_to_endpoint(endpoint, payload)

    record_seat_report(org_id, seat_id, tokens, usd, calls, storage_path=storage_path)
    return {"recorded": True, "transport": "local", **payload}
