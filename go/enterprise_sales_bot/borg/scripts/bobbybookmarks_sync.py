"""
BobbyBookmarks Sync Worker
===========================
Integrates the bobbybookmarks subproject into the TormentNexus ecosystem:
1. Syncs/merges bookmark databases into tormentnexus.db
2. Deduplicates across all data sources
3. Runs as a persistent worker using freellm for any LLM calls
4. Reports sync status periodically
"""

import json
import sqlite3
import time
import urllib.request
import urllib.error
from datetime import datetime
from pathlib import Path

WORKSPACE = Path(__file__).resolve().parent.parent
BOBBY_DIR = WORKSPACE.parent / "bobbybookmarks"  # ../bobbybookmarks
SYNC_INTERVAL = 3600  # 1 hour between sync cycles
LLM_PROXY = "http://localhost:4000/v1/chat/completions"

LOG_PATH = WORKSPACE / "data" / "bobbybookmarks_sync.log"


def log(msg, level="INFO"):
    ts = datetime.now().strftime("%H:%M:%S")
    line = f"[{ts}][{level}] {msg}"
    with open(LOG_PATH, "a", encoding="utf-8") as f:
        f.write(line + "\n")
    print(line)


def llm_call(prompt, max_tokens=200):
    """Call freellm with indefinite timeout."""
    payload = json.dumps(
        {
            "model": "free-llm",
            "messages": [{"role": "user", "content": prompt}],
            "max_tokens": max_tokens,
            "temperature": 0.1,
        }
    ).encode()
    req = urllib.request.Request(
        LLM_PROXY, data=payload, headers={"Content-Type": "application/json"}
    )
    try:
        with urllib.request.urlopen(req, timeout=3600) as r:
            return json.loads(r.read())["choices"][0]["message"]["content"]
    except Exception as e:
        log(f"LLM error: {e}", "WARN")
        return None


def get_table_counts(db_path):
    """Get row counts for all tables in a database."""
    try:
        conn = sqlite3.connect(str(db_path))
        tables = conn.execute(
            "SELECT name FROM sqlite_master WHERE type='table'"
        ).fetchall()
        counts = {}
        for (t,) in tables:
            if t != "sqlite_sequence":
                counts[t] = conn.execute(f'SELECT COUNT(*) FROM "{t}"').fetchone()[0]
        conn.close()
        return counts
    except Exception as e:
        return {"error": str(e)}


def merge_bookmarks_db():
    """Merge bookmarks.db into tormentnexus.db."""
    src = BOBBY_DIR / "bookmarks.db"
    dst = WORKSPACE / "tormentnexus.db"

    if not src.exists():
        log(f"bookmarks.db not found at {src}", "WARN")
        return 0

    log("Merging bookmarks.db -> tormentnexus.db")

    src_counts = get_table_counts(src)
    log(f"Source tables: {json.dumps(src_counts, default=str)[:200]}")

    conn_src = sqlite3.connect(str(src))
    conn_dst = sqlite3.connect(str(dst))

    total_imported = 0

    # Merge bookmarks table
    try:
        bookmarks = conn_src.execute("SELECT * FROM bookmarks").fetchall()
        columns = [
            d[0] for d in conn_src.execute("PRAGMA table_info(bookmarks)").fetchall()
        ]
        log(f"Bookmarks: {len(bookmarks)} rows, columns: {columns}")

        # Create bookmarks table in destination if not exists
        conn_dst.execute("""
            CREATE TABLE IF NOT EXISTS bobby_bookmarks (
                id INTEGER PRIMARY KEY,
                url TEXT UNIQUE,
                title TEXT,
                description TEXT,
                tags TEXT,
                source TEXT,
                created_at TEXT,
                updated_at TEXT,
                normalized_url TEXT,
                is_duplicate INTEGER DEFAULT 0,
                research_status TEXT DEFAULT 'pending',
                imported_at TEXT DEFAULT CURRENT_TIMESTAMP
            )
        """)

        imported = 0
        for row in bookmarks:
            try:
                conn_dst.execute(
                    "INSERT OR IGNORE INTO bobby_bookmarks (id, url, title, description, tags) VALUES (?, ?, ?, ?, ?)",
                    (row[0], row[1], row[2], row[3], row[4]),
                )
                if conn_dst.total_changes > 0:
                    imported += 1
            except Exception:
                pass
        conn_dst.commit()
        log(f"Imported {imported} bookmarks")
        total_imported += imported
    except Exception as e:
        log(f"Bookmark merge error: {e}", "ERR")

    # Merge imported_sessions
    try:
        if "imported_sessions" in src_counts:
            sessions = conn_src.execute("SELECT * FROM imported_sessions").fetchall()
            log(f"Imported sessions: {len(sessions)} rows")

            conn_dst.execute("""
                CREATE TABLE IF NOT EXISTS bobby_imported_sessions (
                    id INTEGER PRIMARY KEY,
                    source_tool TEXT,
                    source_path TEXT UNIQUE,
                    title TEXT,
                    transcript TEXT,
                    imported_at TEXT DEFAULT CURRENT_TIMESTAMP
                )
            """)

            s_imported = 0
            for row in sessions:
                try:
                    conn_dst.execute(
                        "INSERT OR IGNORE INTO bobby_imported_sessions (id, source_tool, source_path, title, transcript) VALUES (?, ?, ?, ?, ?)",
                        (
                            row[0],
                            row[1],
                            row[2],
                            row[3],
                            str(row[4])[:1000] if row[4] else "",
                        ),
                    )
                    if conn_dst.total_changes > 0:
                        s_imported += 1
                except Exception:
                    pass
            conn_dst.commit()
            log(f"Imported {s_imported} sessions")
            total_imported += s_imported
    except Exception as e:
        log(f"Session import error: {e}", "WARN")

    conn_src.close()
    conn_dst.close()
    return total_imported


def merge_atlas_db():
    """Merge atlas.db entries into tormentnexus.db."""
    src = BOBBY_DIR / "atlas.db"
    dst = WORKSPACE / "tormentnexus.db"

    if not src.exists():
        return 0

    log("Merging atlas.db")
    conn_src = sqlite3.connect(str(src))
    conn_dst = sqlite3.connect(str(dst))

    total = 0
    try:
        entries = conn_src.execute("SELECT * FROM entries").fetchall()
        columns = [
            d[0] for d in conn_src.execute("PRAGMA table_info(entries)").fetchall()
        ]
        log(f"Atlas entries: {len(entries)} rows")

        conn_dst.execute("""
            CREATE TABLE IF NOT EXISTS bobby_atlas_entries (
                id INTEGER PRIMARY KEY,
                key TEXT UNIQUE,
                value TEXT,
                layer TEXT,
                created_at TEXT DEFAULT CURRENT_TIMESTAMP
            )
        """)

        for row in entries:
            try:
                conn_dst.execute(
                    "INSERT OR IGNORE INTO bobby_atlas_entries (id, key, value) VALUES (?, ?, ?)",
                    (
                        row[0],
                        str(row[1])[:500],
                        str(row[2])[:5000] if len(row) > 2 else "",
                    ),
                )
                if conn_dst.total_changes > 0:
                    total += 1
            except Exception:
                pass
        conn_dst.commit()
        log(f"Imported {total} atlas entries")
    except Exception as e:
        log(f"Atlas merge error: {e}", "WARN")

    conn_src.close()
    conn_dst.close()
    return total


def deduplicate():
    """Deduplicate bookmarks based on URL normalization."""
    dst = WORKSPACE / "tormentnexus.db"
    conn = sqlite3.connect(str(dst))

    try:
        # Mark duplicates in bobby_bookmarks
        conn.execute("""
            UPDATE bobby_bookmarks SET is_duplicate = 1
            WHERE url IN (
                SELECT url FROM bobby_bookmarks
                GROUP BY url HAVING COUNT(*) > 1
            ) AND id NOT IN (
                SELECT MIN(id) FROM bobby_bookmarks GROUP BY url
            )
        """)
        dup_count = conn.total_changes
        conn.commit()
        if dup_count:
            log(f"Marked {dup_count} duplicate bookmarks")

        # Count remaining
        total = conn.execute(
            "SELECT COUNT(*) FROM bobby_bookmarks WHERE is_duplicate = 0"
        ).fetchone()[0]
        log(f"Unique bookmarks: {total}")
    except Exception as e:
        log(f"Dedup error: {e}", "WARN")

    conn.close()
    return total


def generate_report():
    """Use freellm to generate a summary of the merged data."""
    dst = WORKSPACE / "tormentnexus.db"
    conn = sqlite3.connect(str(dst))

    try:
        bookmark_count = conn.execute(
            "SELECT COUNT(*) FROM bobby_bookmarks"
        ).fetchone()[0]
        unique_count = conn.execute(
            "SELECT COUNT(*) FROM bobby_bookmarks WHERE is_duplicate = 0"
        ).fetchone()[0]
        dup_count = bookmark_count - unique_count
        atlas_count = conn.execute(
            "SELECT COUNT(*) FROM bobby_atlas_entries"
        ).fetchone()[0]
        session_count = conn.execute(
            "SELECT COUNT(*) FROM bobby_imported_sessions"
        ).fetchone()[0]
    except Exception:
        bookmark_count = unique_count = dup_count = atlas_count = session_count = 0

    conn.close()

    prompt = f"""Summarize this data sync between BobbyBookmarks and TormentNexus:
- Bookmarks: {bookmark_count} total ({unique_count} unique, {dup_count} duplicates)
- Atlas entries: {atlas_count}
- Imported sessions: {session_count}
- Sync time: {datetime.now().isoformat()}

Write a brief 2-3 sentence summary of the sync results."""

    result = llm_call(prompt, max_tokens=150)
    if result:
        log(f"LLM Summary: {result[:200]}")


def main():
    log("=" * 60)
    log("BOBBYBOOKMARKS SYNC WORKER STARTED")
    log(f"BobbyBookmarks dir: {BOBBY_DIR}")
    log(f"Sync interval: {SYNC_INTERVAL}s")
    log("=" * 60)

    while True:
        cycle_start = time.time()
        log(f"--- Sync cycle {datetime.now().isoformat()} ---")

        # Step 1: Merge bookmarks.db
        bm_count = merge_bookmarks_db()

        # Step 2: Merge atlas.db
        at_count = merge_atlas_db()

        # Step 3: Deduplicate
        uniq = deduplicate()

        # Step 4: Generate AI summary via freellm
        generate_report()

        elapsed = time.time() - cycle_start
        log(f"Cycle complete: {bm_count + at_count} items merged, {elapsed:.1f}s")
        log(f"Next sync in {SYNC_INTERVAL}s...")
        log("")

        time.sleep(SYNC_INTERVAL)


if __name__ == "__main__":
    main()
