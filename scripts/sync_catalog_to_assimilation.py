#!/usr/bin/env python3
"""
Sync missing catalog.published_mcp_servers into assimilation_state.db:mcp_servers.
New entries get status='pending' so the swarm can pick them up.
Incremental: skips names that already exist.
"""

import re
import sqlite3
import sys

CATALOG_DB = "catalog.db"
ASSIM_DB = "data/assimilation_state.db"


def slugify(name):
    """Convert display_name to our standard snake_case slug."""
    s = name.lower().strip()
    s = re.sub(r"[^a-z0-9\s-]", "", s)
    s = re.sub(r"[\s-]+", "_", s)
    s = s.strip("_")
    if not s:
        return None
    # Avoid too-long names
    if len(s) > 150:
        s = s[:150]
    return s


def main():
    c_cat = sqlite3.connect(CATALOG_DB).cursor()
    c_assim = sqlite3.connect(ASSIM_DB).cursor()

    # Fetch existing names (lowered) for fast lookup
    c_assim.execute("SELECT name, github_url FROM mcp_servers")
    existing = set()
    existing_urls = set()
    for row in c_assim.fetchall():
        existing.add(row[0].lower().strip())
        if row[1]:
            existing_urls.add(row[1].rstrip("/").lower())

    # Fetch all catalog servers
    c_cat.execute("""
        SELECT canonical_id, display_name, repository_url, stars, description
        FROM published_mcp_servers
        ORDER BY stars DESC
    """)
    rows = c_cat.fetchall()

    inserted = 0
    skipped_existing_name = 0
    skipped_no_name = 0
    skipped_url_exists = 0

    for row in rows:
        canonical_id, display_name, repo_url, stars, description = row
        if not display_name or not display_name.strip():
            skipped_no_name += 1
            continue

        name = slugify(display_name)
        if not name:
            skipped_no_name += 1
            continue

        # Skip if name already exists in assimilation
        if name in existing:
            skipped_existing_name += 1
            continue

        # Skip if same repo URL already exists
        repo_normalized = repo_url.rstrip("/").lower() if repo_url else ""
        if repo_normalized and repo_normalized in existing_urls:
            skipped_url_exists += 1
            continue

        # Insert new pending entry
        try:
            c_assim.execute(
                """
                INSERT INTO mcp_servers (name, github_url, score, status, notes, added_at)
                VALUES (?, ?, ?, 'pending', ?, datetime('now'))
            """,
                (
                    name,
                    repo_url or "",
                    stars or 0,
                    (description or "")[:500] if description else "",
                ),
            )
            existing.add(name)
            if repo_normalized:
                existing_urls.add(repo_normalized)
            inserted += 1
        except sqlite3.IntegrityError as e:
            # Race: name already inserted by another process
            print(f"  INTEGRITY ERROR on {name}: {e}", file=sys.stderr)
            continue

    c_assim.connection.commit()

    print(f"Candidates processed: {len(rows)}")
    print(f"Inserted (new pending): {inserted}")
    print(f"Skipped (name exists): {skipped_existing_name}")
    print(f"Skipped (no name): {skipped_no_name}")
    print(f"Skipped (URL exists): {skipped_url_exists}")

    # Verify new total
    c_assim.execute("SELECT COUNT(*) FROM mcp_servers")
    new_total = c_assim.fetchone()[0]
    print(f"New total assimilation rows: {new_total}")

    c_cat.connection.close()
    c_assim.connection.close()


if __name__ == "__main__":
    main()
