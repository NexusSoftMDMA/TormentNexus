"""
Parse and assimilate MCP servers from bobbybookmarks files into the state database.
"""

import re
import json
import sqlite3
from pathlib import Path

WORKSPACE = Path(r"C:\Users\hyper\workspace")
BOBBY = WORKSPACE / "bobbybookmarks"
TORMENT = WORKSPACE / "tormentnexus"
DB_PATH = TORMENT / "data" / "assimilation_state.db"
TOOLS_DIR = TORMENT / "go" / "internal" / "tools"

RANKED_FILE = BOBBY / "mcp_servers_ranked.md"
ORGANIZED_FILE = BOBBY / "mcp_servers_organized.md"


def parse_ranked(path):
    """Parse value-ranked MCP servers file."""
    text = path.read_text("utf-8", errors="replace")
    entries = {}

    pat = re.compile(
        r"^\s*\d+\.\s+\*?\*?\[(\d+)\]\*?\*?\s+`sig:(\d+)`\s+(⭐\s+)?"
        r"\[([^\]]+)\]\(([^)]+)\)",
        re.MULTILINE,
    )

    for m in pat.finditer(text):
        score = int(m.group(1))
        signal = int(m.group(2))
        title = m.group(4)
        url = m.group(5).rstrip("/")

        name = (
            title.split(":")[0].strip()
            if ":" in title
            else title.split("/")[-1]
            if "/" in title
            else title
        )
        if len(name) > 100:
            name = name[:100]

        if name in entries:
            continue

        entries[name] = dict(
            name=name,
            url=url,
            score=score,
            signal=signal,
            description=title,
            category="",
            subcategory="",
        )

    print(f"  ranked: {len(entries)} entries")
    return entries


def parse_organized(path):
    """Parse category-organized MCP servers file."""
    text = path.read_text("utf-8", errors="replace")
    entries = {}

    cat_pat = re.compile(r"^##\s+(.+)$", re.MULTILINE)
    entry_pat = re.compile(
        r"^\s*\d+\.\s+`sig:(\d+)`\s+(⭐\s+)?\[([^\]]+)\]\(([^)]+)\)", re.MULTILINE
    )

    # Split text by category headers
    sections = cat_pat.split(text)
    # sections[0] is before first header, then alternating [title, body, title, body, ...]
    current_cat = ""
    for i, section in enumerate(sections):
        section = section.strip()
        if not section:
            continue
        if i % 2 == 1:
            # This is a category title (odd index in split result)
            current_cat = section
        else:
            # This is body text (even index, after a title)
            for m in entry_pat.finditer(section):
                signal = int(m.group(1))
                title = m.group(3)
                url = m.group(4).rstrip("/")
                name = (
                    title.split(":")[0].strip()
                    if ":" in title
                    else title.split("/")[-1]
                    if "/" in title
                    else title
                )
                if len(name) > 100:
                    name = name[:100]

                if name in entries:
                    if current_cat and not entries[name]["category"]:
                        entries[name]["category"] = current_cat
                else:
                    entries[name] = dict(
                        name=name,
                        url=url,
                        score=0,
                        signal=signal,
                        description=title,
                        category=current_cat,
                        subcategory="",
                    )

    print(f"  organized: {len(entries)} entries")
    return entries


def merge(ranked, organized):
    merged = {}
    all_names = set(ranked.keys()) | set(organized.keys())
    for name in all_names:
        r = ranked.get(name, {})
        o = organized.get(name, {})
        merged[name] = dict(
            name=name,
            url=r.get("url", o.get("url", "")),
            score=r.get("score", 0),
            signal=max(r.get("signal", 0), o.get("signal", 0)),
            description=r.get("description", o.get("description", "")),
            category=o.get("category", r.get("category", "general")),
            subcategory="",
        )
    print(f"  merged: {len(merged)} unique")
    return list(merged.values())


def check_go():
    existing = set()
    if TOOLS_DIR.exists():
        for f in TOOLS_DIR.iterdir():
            if f.suffix == ".go" and not f.name.endswith("_test.go"):
                stem = f.stem.lower().replace("_", "-").replace(".", "-")
                existing.add(stem)
    print(f"  go tools: {len(existing)} files")
    return existing


def load_db(entries, existing_go):
    DB_PATH.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(str(DB_PATH))
    conn.executescript("""
        CREATE TABLE IF NOT EXISTS mcp_servers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            github_url TEXT DEFAULT '',
            score INTEGER DEFAULT 0,
            signal INTEGER DEFAULT 0,
            category TEXT DEFAULT '',
            subcategory TEXT DEFAULT '',
            description TEXT DEFAULT '',
            tags TEXT DEFAULT '[]',
            status TEXT DEFAULT 'pending',
            go_file TEXT DEFAULT '',
            tools_exposed TEXT DEFAULT '[]',
            notes TEXT DEFAULT '',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS skills (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            description TEXT DEFAULT '',
            category TEXT DEFAULT 'general',
            frontmatter TEXT DEFAULT '',
            content TEXT DEFAULT '',
            version INTEGER DEFAULT 1,
            similarity_score INTEGER DEFAULT 100,
            canonical_id INTEGER REFERENCES skills(id),
            status TEXT DEFAULT 'active',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS hermes_addons (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            category TEXT DEFAULT 'community',
            description TEXT DEFAULT '',
            go_file TEXT DEFAULT '',
            skill_name TEXT DEFAULT '',
            status TEXT DEFAULT 'pending',
            notes TEXT DEFAULT '',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS prompt_library (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            description TEXT DEFAULT '',
            category TEXT DEFAULT 'general',
            content TEXT NOT NULL,
            tags TEXT DEFAULT '[]',
            usage_count INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_mcp_status ON mcp_servers(status);
        CREATE INDEX IF NOT EXISTS idx_mcp_category ON mcp_servers(category);
        CREATE INDEX IF NOT EXISTS idx_mcp_score ON mcp_servers(score);
    """)
    conn.commit()

    inserted = 0
    pending_count = 0
    implemented_count = 0

    for e in entries:
        name_norm = e["name"].lower().replace("_", "-").replace(".", "-")
        url_norm = e["url"].lower().replace("https://github.com/", "").replace("/", "-")
        url_norm = url_norm.replace("_", "-").replace(".", "-")

        is_go = any(
            g in name_norm or name_norm in g or g in url_norm for g in existing_go
        )
        status = "implemented" if is_go else "pending"
        go_file = ""
        if is_go:
            match = next(
                (
                    g
                    for g in existing_go
                    if g in name_norm or name_norm in g or g in url_norm
                ),
                None,
            )
            if match:
                go_file = f"go/internal/tools/{match}.go"

        try:
            conn.execute(
                """
                INSERT OR IGNORE INTO mcp_servers
                    (name, github_url, score, signal, category, description, status, go_file)
                VALUES (?,?,?,?,?,?,?,?)
            """,
                (
                    e["name"][:200],
                    e["url"],
                    e["score"],
                    e["signal"],
                    e["category"][:100],
                    e["description"][:500],
                    status,
                    go_file,
                ),
            )
            inserted += 1
            if status == "pending":
                pending_count += 1
            else:
                implemented_count += 1
        except Exception as ex:
            print(f"  ERROR {e['name']}: {ex}")

    conn.commit()
    total = conn.execute("SELECT COUNT(*) FROM mcp_servers").fetchone()[0]
    conn.close()

    print("\n  DB SUMMARY:")
    print(f"    Total:      {total}")
    print(f"    Inserted:    {inserted}")
    print(f"    Already Go:  {implemented_count}")
    print(f"    Pending:     {pending_count}")

    return dict(
        total=total,
        inserted=inserted,
        go_implemented=implemented_count,
        pending=pending_count,
    )


if __name__ == "__main__":
    print("=== MCP Server Assimilation ===\n")
    print("1. Parsing ranked file...")
    ranked = parse_ranked(RANKED_FILE)

    print("\n2. Parsing organized file...")
    organized = parse_organized(ORGANIZED_FILE)

    print("\n3. Merging...")
    entries = merge(ranked, organized)

    print("\n4. Checking existing Go tools...")
    existing_go = check_go()

    print("\n5. Loading into SQLite DB...")
    stats = load_db(entries, existing_go)

    print("\n=== DONE ===")
    print(f"  DB: {DB_PATH}")
    print(json.dumps(stats, indent=2))
