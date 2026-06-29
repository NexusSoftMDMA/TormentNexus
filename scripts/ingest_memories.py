#!/usr/bin/env python3
"""
Ingest memories from external DBs into the TormentNexus L2/L3 vaults.

Sources:
  - tormentnexus.db       → imported_session_memories (57K+ records)
  - atlas.db              → atlas entries (17K+ records)
  - bobbybookmarks/atlas.db → same atlas, submodule copy

Usage:
  python scripts/ingest_memories.py [--dry-run] [--limit N]
"""

import argparse
import json
import os
import sqlite3
import uuid
from datetime import datetime, timezone
from pathlib import Path

WORKSPACE = Path(__file__).resolve().parent.parent
GO_SIDECAR = os.environ.get("TORMENTNEXUS_SIDECAR", "http://127.0.0.1:7778")

SOURCE_DBS = [
    (
        "TormentNexus sessions",
        WORKSPACE / "tormentnexus.db",
        "SELECT uuid, imported_session_uuid, content, created_at FROM imported_session_memories ORDER BY created_at DESC",
    ),
    (
        "BobbyBookmarks atlas",
        WORKSPACE / "bobbybookmarks" / "atlas.db",
        "SELECT rowid, page_title, short_description, created_at FROM entries ORDER BY created_at DESC",
    ),
    (
        "Local atlas",
        WORKSPACE / "data" / "atlas.db",
        "SELECT rowid, page_title, short_description, created_at FROM entries ORDER BY created_at DESC",
    ),
]


def fmt_time(ts):
    """Convert a DB timestamp to RFC3339."""
    if not ts:
        return datetime.now(timezone.utc).isoformat()
    try:
        if isinstance(ts, (int, float)):
            return datetime.fromtimestamp(ts, tz=timezone.utc).isoformat()
        return str(ts)
    except Exception:
        return datetime.now(timezone.utc).isoformat()


def make_memory(content, source_id, source_table, created_at):
    """Build a memory record dict matching L2VaultRecord."""
    return {
        "id": str(uuid.uuid5(uuid.NAMESPACE_DNS, f"{source_table}:{source_id}")),
        "session_id": source_table,
        "memory_type": "long_term",
        "memory_kind": "fact",
        "category": "imported",
        "tags": source_table,
        "source_url": "",
        "content": str(content)[:10000] if content else "",
        "importance": 0.5,
        "heat_score": 50.0,
        "last_accessed_at": datetime.now(timezone.utc).isoformat(),
        "created_at": fmt_time(created_at),
    }


def push_memory(record, dry_run):
    """POST a memory record to the Go sidecar."""
    if dry_run:
        return True
    import urllib.request

    data = json.dumps(record).encode("utf-8")
    req = urllib.request.Request(
        f"{GO_SIDECAR}/api/memory/add",
        data=data,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            return resp.status in (200, 201)
    except Exception:
        return False


def main():
    parser = argparse.ArgumentParser(
        description="Ingest memories into TormentNexus vault"
    )
    parser.add_argument("--dry-run", action="store_true", help="Scan only, don't push")
    parser.add_argument("--limit", type=int, default=0, help="Max memories per source")
    args = parser.parse_args()

    total_pushed = 0
    total_skipped = 0

    for label, db_path, query in SOURCE_DBS:
        if not db_path.exists():
            print(f"  ..  {label}: not found ({db_path})")
            continue

        try:
            conn = sqlite3.connect(str(db_path))
            conn.row_factory = sqlite3.Row
            cursor = conn.execute(query)
            rows = cursor.fetchall()
            count = len(rows)
            print(f"\n{'=' * 60}")
            print(f"[STORE] {label}: {count} records")
            print(f"{'=' * 60}")

            source_table = label.lower().replace(" ", "_")

            for i, row in enumerate(rows):
                if args.limit and i >= args.limit:
                    print(f"  (stopped at limit {args.limit})")
                    break

                content = row[2] if len(row) > 2 else str(dict(row))
                source_id = str(row[0])
                created_at = row[3] if len(row) > 3 else None

                record = make_memory(content, source_id, source_table, created_at)

                if args.dry_run:
                    total_skipped += 1
                    if i < 3:
                        print(f"  [PUSH] Would push: {record['content'][:60]}...")
                    continue

                ok = push_memory(record, dry_run=False)
                if ok:
                    total_pushed += 1
                else:
                    total_skipped += 1

                if (i + 1) % 100 == 0:
                    print(f"  Progress: {i + 1}/{count}")

            conn.close()

        except Exception as e:
            print(f"  !! Error reading {label}: {e}")

    print(f"\n{'=' * 60}")
    if args.dry_run:
        print(f"[DONE] DRY RUN: {total_skipped} memories scanned (not pushed)")
    else:
        print(f"[DONE] Done: {total_pushed} pushed, {total_skipped} skipped/errors")
    print(f"{'=' * 60}")


if __name__ == "__main__":
    main()
