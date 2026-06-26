"""Import all data from ../bobbybookmarks into tormentnexus.db"""

import sqlite3
from pathlib import Path

SRC = Path("../bobbybookmarks")
DST = Path("tormentnexus.db")


def log(msg):
    print(f"  {msg}")


def safe_import(
    src_db,
    table,
    dst_table=None,
    columns=None,
    where=None,
    transform=None,
    conflict="IGNORE",
):
    """Copy data from one DB to another."""
    src = sqlite3.connect(str(SRC / src_db))
    dst = sqlite3.connect(str(DST))

    if dst_table is None:
        dst_table = table

    try:
        # Check source
        src_rows = src.execute(f'SELECT COUNT(*) FROM "{table}"').fetchone()[0]
        if src_rows == 0:
            src.close()
            dst.close()
            return 0

        # Check destination - create table if needed
        if columns:
            col_defs = ", ".join(f'"{c}" TEXT' for c in columns)
            dst.execute(f'CREATE TABLE IF NOT EXISTS "{dst_table}" ({col_defs})')

        # Copy data
        if columns:
            sel = ", ".join(f'"{c}"' for c in columns)
            placeholders = ", ".join("?" for _ in columns)
        else:
            sel = "*"
            placeholders = None

        query = f'SELECT {sel} FROM "{table}"'
        if where:
            query += f" WHERE {where}"

        rows = src.execute(query).fetchall()

        if placeholders:
            dst.executemany(
                f'INSERT OR {conflict} INTO "{dst_table}" VALUES ({placeholders})',
                [transform(r) if transform else r for r in rows],
            )
        else:
            # Dynamic columns
            for r in rows:
                try:
                    dst.execute(
                        f'INSERT OR {conflict} INTO "{dst_table}" VALUES ({",".join("?" for _ in r)})',
                        r,
                    )
                except:
                    pass

        dst.commit()
        imported = len(rows)
        log(f"{src_db}.{table} -> {dst_table}: {imported} rows")
        src.close()
        dst.close()
        return imported
    except Exception as e:
        log(f"{src_db}.{table} -> {dst_table}: ERROR {e}")
        src.close()
        dst.close()
        return 0


def copy_bookmarks():
    """Copy bookmarks with url/url dedup."""
    src = sqlite3.connect(str(SRC / "bookmarks.db"))
    dst = sqlite3.connect(str(DST))

    dst.execute("""
        CREATE TABLE IF NOT EXISTS bobby_bookmarks (
            id INTEGER PRIMARY KEY,
            url TEXT,
            url2 TEXT,
            title TEXT,
            description TEXT,
            tags TEXT,
            source TEXT,
            is_duplicate INTEGER DEFAULT 0,
            imported_at TEXT DEFAULT CURRENT_TIMESTAMP
        )
    """)

    rows = src.execute("SELECT * FROM bookmarks").fetchall()
    imported = 0
    for r in rows:
        try:
            dst.execute(
                "INSERT OR IGNORE INTO bobby_bookmarks (id, url, url2, title, description, tags) VALUES (?, ?, ?, ?, ?, ?)",
                (
                    r[0],
                    r[1] if len(r) > 1 else "",
                    r[2] if len(r) > 2 else "",
                    r[3] if len(r) > 3 else "",
                    r[4] if len(r) > 4 else "",
                    str(r[5]) if len(r) > 5 else "",
                ),
            )
            imported += 1
        except:
            pass
    dst.commit()
    log(f"bookmarks.db.bookmarks -> bobby_bookmarks: {imported} rows")
    src.close()
    dst.close()
    return imported


def copy_atlas():
    src = sqlite3.connect(str(SRC / "atlas.db"))
    dst = sqlite3.connect(str(DST))
    dst.execute(
        "CREATE TABLE IF NOT EXISTS bobby_atlas_entries (id INTEGER, key TEXT, value TEXT, layer TEXT)"
    )
    rows = src.execute("SELECT * FROM entries").fetchall()
    imported = 0
    for r in rows:
        try:
            dst.execute(
                "INSERT OR IGNORE INTO bobby_atlas_entries VALUES (?, ?, ?, ?)",
                (
                    r[0],
                    str(r[1])[:500],
                    str(r[2])[:5000],
                    str(r[3]) if len(r) > 3 else "",
                ),
            )
            imported += 1
        except:
            pass
    dst.commit()
    log(f"atlas.db.entries -> bobby_atlas_entries: {imported} rows")
    src.close()
    dst.close()


def copy_all():
    print("=== Importing from ../bobbybookmarks ===")

    copy_bookmarks()
    copy_atlas()

    safe_import("bookmarks.db", "embeddings", "bobby_embeddings")
    safe_import("bookmarks.db", "clusters", "bobby_clusters")
    safe_import("bookmarks.db", "debates", "bobby_debates")
    safe_import("bookmarks.db", "nebula_map", "bobby_nebula_map")
    safe_import("bookmarks.db", "agent_heartbeats", "bobby_agent_heartbeats")

    safe_import("catalog.db", "catalog_entries", "bobby_catalog_entries")

    print()
    dst = sqlite3.connect(str(DST))
    for table in dst.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'bobby_%'"
    ).fetchall():
        count = dst.execute(f'SELECT COUNT(*) FROM "{table[0]}"').fetchone()[0]
        print(f"  {table[0]}: {count} rows")
    dst.close()


if __name__ == "__main__":
    copy_all()
