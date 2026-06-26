"""SQLite-backed org-rollup store. Transport-agnostic: callers (a local
`org-report`, or — later — an HTTP ingest route) write seat reports here; the
org host aggregates with ``org_rollup``.

Data is deliberately minimal — per-seat token/dollar/call counts by day. No
code, no content, no file paths leave a seat; only aggregate savings numbers.
"""

from __future__ import annotations

import datetime
import os
import sqlite3
from pathlib import Path
from typing import Optional


def _db_path(storage_path: Optional[str] = None) -> Path:
    base = storage_path or os.environ.get("CODE_INDEX_PATH", str(Path.home() / ".code-index"))
    return Path(base) / "org_savings.db"


def _connect(storage_path: Optional[str] = None) -> sqlite3.Connection:
    p = _db_path(storage_path)
    p.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(str(p))
    conn.execute(
        """CREATE TABLE IF NOT EXISTS org_savings (
            org_id TEXT NOT NULL,
            seat_id TEXT NOT NULL,
            date TEXT NOT NULL,
            tokens_saved INTEGER NOT NULL,
            usd REAL NOT NULL,
            calls INTEGER NOT NULL,
            updated_at TEXT NOT NULL,
            PRIMARY KEY (org_id, seat_id, date)
        )"""
    )
    return conn


def record_seat_report(
    org_id: str,
    seat_id: str,
    tokens_saved: int,
    usd: float,
    calls: int,
    *,
    date: Optional[str] = None,
    storage_path: Optional[str] = None,
) -> dict:
    """Upsert one seat's savings for a day (latest report for that day wins)."""
    if not org_id or not seat_id:
        raise ValueError("org_id and seat_id are required")
    date = date or datetime.date.today().isoformat()
    now = datetime.datetime.now(datetime.timezone.utc).isoformat()
    conn = _connect(storage_path)
    try:
        conn.execute(
            "INSERT INTO org_savings (org_id, seat_id, date, tokens_saved, usd, calls, updated_at) "
            "VALUES (?, ?, ?, ?, ?, ?, ?) "
            "ON CONFLICT(org_id, seat_id, date) DO UPDATE SET "
            "tokens_saved=excluded.tokens_saved, usd=excluded.usd, "
            "calls=excluded.calls, updated_at=excluded.updated_at",
            (org_id, seat_id, date, int(tokens_saved), float(usd), int(calls), now),
        )
        conn.commit()
    finally:
        conn.close()
    return {"org_id": org_id, "seat_id": seat_id, "date": date}


def org_rollup(org_id: str, *, storage_path: Optional[str] = None) -> dict:
    """Aggregate per-seat savings for an org → {org_id, seats[], totals}."""
    conn = _connect(storage_path)
    try:
        rows = conn.execute(
            "SELECT seat_id, SUM(tokens_saved), SUM(usd), SUM(calls), MAX(updated_at) "
            "FROM org_savings WHERE org_id = ? GROUP BY seat_id "
            "ORDER BY SUM(tokens_saved) DESC",
            (org_id,),
        ).fetchall()
    finally:
        conn.close()
    seats = [
        {
            "seat_id": r[0],
            "tokens_saved": int(r[1] or 0),
            "usd": round(r[2] or 0.0, 4),
            "calls": int(r[3] or 0),
            "last_seen": r[4],
        }
        for r in rows
    ]
    totals = {
        "tokens_saved": sum(s["tokens_saved"] for s in seats),
        "usd": round(sum(s["usd"] for s in seats), 4),
        "calls": sum(s["calls"] for s in seats),
        "seat_count": len(seats),
    }
    return {"org_id": org_id, "seats": seats, "totals": totals}
