#!/usr/bin/env python3
"""
Phase 3: Parse MCP Comprehensive List → catalog DB
Also extract links_backlog from atlas.db
"""

import sqlite3
import re
import time
from pathlib import Path

PROJECT = Path(r"C:\Users\hyper\workspace\bobbybookmarks")
DB_PATH = PROJECT / "tormentnexus.db"
MCP_LIST = PROJECT / "research/mcp_comprehensive_list.md"
ATLAS_DB = PROJECT / "atlas.db"


def parse_mcp_list():
    """Parse the MCP comprehensive list into structured entries."""
    with open(MCP_LIST, "r", encoding="utf-8", errors="replace") as f:
        content = f.read()

    entries = []
    lines = content.split("\n")

    current_entry = {}
    for line in lines:
        line = line.strip()

        # Skip section headers
        if line.startswith("===") or line.startswith("---"):
            continue

        # Match: sig=   XX | Title
        m = re.match(r"sig=\s*(\d+)\s*\|\s*(.*)", line)
        if m:
            if current_entry.get("url") and current_entry.get("title"):
                entries.append(current_entry)
            current_entry = {
                "relevance_score": int(m.group(1)),
                "title": m.group(2).strip(),
                "description": "",
            }
            continue

        # Match: URL: https://...
        m = re.match(r"URL:\s*(https?://\S+)", line)
        if m and current_entry:
            current_entry["url"] = m.group(1)
            continue

        # Description line (text that's not empty and not a section header)
        if (
            line
            and not line.startswith("sig=")
            and not line.startswith("URL:")
            and not line.startswith("#")
            and not line.startswith("==")
            and current_entry.get("url")
        ):
            if current_entry.get("description"):
                current_entry["description"] += " " + line
            else:
                current_entry["description"] = line

    # Don't forget last entry
    if current_entry.get("url") and current_entry.get("title"):
        entries.append(current_entry)

    return entries


def extract_catalog_from_atlas():
    """Extract URLs from atlas.db for links_backlog recovery."""
    conn = sqlite3.connect(str(ATLAS_DB))
    cur = conn.cursor()
    cur.execute("SELECT COUNT(*) FROM entries")
    total = cur.fetchone()[0]
    print(f"ATLAS entries: {total}")

    # Get sample entries
    cur.execute(
        "SELECT url, page_title, short_description, main_features FROM entries LIMIT 10"
    )
    print("\nSample ATLAS entries:")
    for row in cur.fetchall():
        print(f"  {row[0][:80]} | {row[1][:40] if row[1] else ''}")

    conn.close()
    return total


def create_catalog_db(entries):
    """Create a catalog from parsed MCP entries."""
    conn = sqlite3.connect(str(PROJECT / "catalog.db"))
    cur = conn.cursor()

    cur.execute("""
        CREATE TABLE IF NOT EXISTS catalog_entries (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT,
            url TEXT UNIQUE,
            description TEXT,
            relevance_score INTEGER DEFAULT 0,
            category TEXT DEFAULT 'mcp',
            ingested_at INTEGER
        )
    """)

    now = int(time.time())
    inserted = 0
    for entry in entries:
        try:
            cur.execute(
                """
                INSERT OR IGNORE INTO catalog_entries 
                (title, url, description, relevance_score, category, ingested_at)
                VALUES (?, ?, ?, ?, ?, ?)
            """,
                (
                    entry.get("title", "")[:500],
                    entry.get("url", ""),
                    entry.get("description", "")[:1000],
                    entry.get("relevance_score", 0),
                    "mcp",
                    now,
                ),
            )
            if cur.rowcount > 0:
                inserted += 1
        except Exception as e:
            print(f"  Error inserting {entry.get('url', '')}: {e}")

    conn.commit()
    total = cur.execute("SELECT COUNT(*) FROM catalog_entries").fetchone()[0]
    print(f"\nCatalog: {inserted} new, {total} total entries")
    conn.close()
    return inserted


if __name__ == "__main__":
    print("=" * 60)
    print("Phase 3: Rebuilding MCP Catalog from Comprehensive List")
    print("=" * 60)

    # Parse MCP list
    print("\n1. Parsing MCP comprehensive list...")
    entries = parse_mcp_list()
    print(f"   Parsed {len(entries)} MCP entries")

    # Create catalog
    print("\n2. Creating catalog DB...")
    inserted = create_catalog_db(entries)

    # Check Atlas
    print("\n3. Checking Atlas DB for links...")
    atlas_count = extract_catalog_from_atlas()

    print(
        f"\nPhase 3 complete: {len(entries)} entries parsed, {inserted} inserted into catalog"
    )
