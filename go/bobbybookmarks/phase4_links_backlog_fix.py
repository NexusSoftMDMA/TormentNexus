#!/usr/bin/env python3
"""
Phase 4: Rebuild links_backlog from atlas.db entries (fixed)
"""

import sqlite3
import json
import uuid as uuid_mod
import time
from pathlib import Path

PROJECT = Path(r"C:\Users\hyper\workspace\bobbybookmarks")
DB_PATH = PROJECT / "tormentnexus.db"
ATLAS_DB = PROJECT / "atlas.db"


def extract_links_backlog():
    """Extract all URLs from atlas.db into the links_backlog table."""
    conn = sqlite3.connect(str(DB_PATH))
    cur = conn.cursor()

    cur.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name='links_backlog'"
    )
    if not cur.fetchone():
        print("links_backlog table doesn't exist in tormentnexus.db, skipping")
        conn.close()
        return 0

    cur.execute("PRAGMA table_info(links_backlog)")
    cols = [row[1] for row in cur.fetchall()]
    print(f"links_backlog columns: {len(cols)}")

    cur.execute("SELECT COUNT(*) FROM links_backlog")
    existing = cur.fetchone()[0]
    print(f"Existing links_backlog entries: {existing}")

    atlas = sqlite3.connect(str(ATLAS_DB))
    atlas_cur = atlas.cursor()
    atlas_cur.execute(
        "SELECT id, url, page_title, short_description, main_features FROM entries"
    )

    now_ms = int(time.time() * 1000)
    inserted = 0
    batch = []

    for row in atlas_cur.fetchall():
        entry_id, url, title, short_desc, long_desc = row
        title = title or ""
        description = short_desc or long_desc or ""

        batch.append(
            (
                str(uuid_mod.uuid4()),
                url,
                "",
                title[:500],
                description[:1000],
                "",
                "atlas-rebuild",
                0,
                "",
                "pending",
                0,
                "",
                "",
                now_ms,
                0,
                0,
                "",
                json.dumps({"source": "atlas", "entry_id": entry_id}),
                "",
                0,
                "",
                "",
                "",
                "",
                0,
                now_ms,
                now_ms,
            )
        )

        if len(batch) >= 100:
            try:
                conn.executemany(
                    """
                    INSERT OR IGNORE INTO links_backlog 
                    (uuid, url, normalized_url, title, description, tags, source, is_duplicate, duplicate_of,
                     research_status, http_status, page_title, page_description, discovered_at, 
                     last_attempted_at, completed_at, error_message, metadata, favicon_url, researched_at,
                     cluster_id, bobbybookmarks_bookmark_id, import_session_id, raw_payload, 
                     synced_at, created_at, updated_at)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                    batch,
                )
                conn.commit()
                inserted += len(batch)
                print(f"  Inserted {inserted} links...")
            except Exception as e:
                print(f"  Batch error: {e}")
            batch = []

    if batch:
        try:
            conn.executemany(
                """
                INSERT OR IGNORE INTO links_backlog 
                (uuid, url, normalized_url, title, description, tags, source, is_duplicate, duplicate_of,
                 research_status, http_status, page_title, page_description, discovered_at, 
                 last_attempted_at, completed_at, error_message, metadata, favicon_url, researched_at,
                 cluster_id, bobbybookmarks_bookmark_id, import_session_id, raw_payload, 
                 synced_at, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
                batch,
            )
            conn.commit()
            inserted += len(batch)
        except Exception as e:
            print(f"  Final batch error: {e}")

    cur.execute("SELECT COUNT(*) FROM links_backlog")
    total_after = cur.fetchone()[0]
    print(f"Links backlog: {existing} -> {total_after} (+{total_after - existing})")

    atlas.close()
    conn.close()
    return inserted


if __name__ == "__main__":
    print("=" * 60)
    print("Phase 4: Rebuilding links_backlog from ATLAS (final)")
    print("=" * 60)
    count = extract_links_backlog()
    print(f"Phase 4 complete: {count} links inserted")
