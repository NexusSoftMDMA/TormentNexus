"""Regenerate assimilation_state.db from catalog.db"""

import sqlite3

cat = sqlite3.connect("catalog.db")
assim = sqlite3.connect("data/assimilation_state.db")

assim.execute("""
    CREATE TABLE IF NOT EXISTS mcp_servers (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        github_url TEXT,
        score REAL DEFAULT 0,
        status TEXT DEFAULT 'pending',
        go_file TEXT,
        tools_exposed TEXT,
        notes TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
""")

servers = cat.execute(
    "SELECT display_name, repository_url, stars FROM published_mcp_servers"
).fetchall()
print(f"Copying {len(servers)} servers from catalog.db")

for name, url, score in servers:
    if url and url.startswith("git+"):
        url = url[4:]
    if url and url.startswith("git@"):
        url = "https://" + url.split("@")[1].replace(":", "/")
    status = "implemented" if score and score > 0 else "pending"
    assim.execute(
        "INSERT INTO mcp_servers (name, github_url, score, status) VALUES (?, ?, ?, ?)",
        (name, url, score or 0, status),
    )

assim.commit()
total = assim.execute("SELECT COUNT(*) FROM mcp_servers").fetchone()[0]
done = assim.execute(
    "SELECT COUNT(*) FROM mcp_servers WHERE status='implemented'"
).fetchone()[0]
pend = assim.execute(
    "SELECT COUNT(*) FROM mcp_servers WHERE status='pending'"
).fetchone()[0]
print(f"Total: {total} | Implemented: {done} | Pending: {pend}")

assim.close()
cat.close()
