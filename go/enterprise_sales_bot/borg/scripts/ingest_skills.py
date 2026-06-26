import sqlite3
import re
import os
from pathlib import Path

SKILLS_DIR = Path(".agent/skills")
DB_PATH = Path(".tormentnexus/skills.db")

def extract_skill_info(skill_dir: Path) -> dict:
    md_file = skill_dir / "SKILL.md"
    if not md_file.exists(): return None
    content = md_file.read_text(encoding="utf-8", errors="replace")
    category, description = "general", ""
    fm_match = re.match(r"^---\n(.*?)\n---", content, re.DOTALL)
    if fm_match:
        fm_text = fm_match.group(1)
        m = re.search(r"category:\s*(.+)", fm_text, re.IGNORECASE)
        if m: category = m.group(1).strip().strip("\"'")
        m = re.search(r"description:\s*(.+)", fm_text, re.IGNORECASE)
        if m: description = m.group(1).strip().strip("\"'")[:200]
    if not description:
        for line in content.splitlines():
            line = line.strip()
            if line and not line.startswith("#") and not line.startswith("---"):
                description = line[:200]
                break
    return dict(name=skill_dir.name, description=description, category=category, frontmatter=content[:800], content=content)

def main():
    if not SKILLS_DIR.exists(): return
    os.makedirs(DB_PATH.parent, exist_ok=True)
    conn = sqlite3.connect(str(DB_PATH))
    conn.executescript("""
        CREATE TABLE IF NOT EXISTS skills (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            description TEXT DEFAULT '',
            category TEXT DEFAULT 'general',
            frontmatter TEXT DEFAULT '',
            content TEXT DEFAULT '',
            version INTEGER DEFAULT 1,
            similarity INTEGER DEFAULT 100,
            canonical_id INTEGER REFERENCES skills(id),
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    """)
    for d in SKILLS_DIR.iterdir():
        if d.is_dir():
            info = extract_skill_info(d)
            if info:
                conn.execute("INSERT OR IGNORE INTO skills (name, description, category, frontmatter, content) VALUES (?,?,?,?,?)",
                             (info["name"], info["description"], info["category"], info["frontmatter"], info["content"]))
    conn.commit()
    conn.close()

if __name__ == "__main__": main()
