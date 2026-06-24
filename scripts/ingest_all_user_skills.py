#!/usr/bin/env python3
import sqlite3
import re
import os
import sys
from pathlib import Path

# Paths
HOME = Path(os.path.expanduser("~"))
USER_SKILLS_DIR = HOME / ".tormentnexus" / "skills"
DB_PATH = Path(".tormentnexus/skills.db")

def tokenize(text):
    """Tokenize text into lowercase words of length >= 2."""
    return set(re.findall(r"\b[a-z0-9]{2,}\b", text.lower()))

def calculate_jaccard(set1, set2):
    """Calculate Jaccard similarity between two token sets."""
    if not set1 or not set2:
        return 0.0
    intersection = len(set1.intersection(set2))
    union = len(set1.union(set2))
    return float(intersection) / union

def extract_skill_info(skill_dir: Path) -> dict:
    md_file = skill_dir / "SKILL.md"
    if not md_file.exists():
        return None
    try:
        content = md_file.read_text(encoding="utf-8", errors="replace")
    except Exception as e:
        print(f"Error reading {md_file}: {e}")
        return None

    category = "general"
    description = ""
    name = skill_dir.name

    # Parse yaml frontmatter
    fm_match = re.match(r"^---\n(.*?)\n---", content, re.DOTALL)
    if fm_match:
        fm_text = fm_match.group(1)
        # Parse name
        m = re.search(r"^name:\s*(.+)", fm_text, re.MULTILINE | re.IGNORECASE)
        if m:
            name = m.group(1).strip().strip("\"'")
        # Parse category
        m = re.search(r"^category:\s*(.+)", fm_text, re.IGNORECASE | re.MULTILINE)
        if m:
            category = m.group(1).strip().strip("\"'")
        # Parse description
        m = re.search(r"^description:\s*(.+)", fm_text, re.IGNORECASE | re.MULTILINE)
        if m:
            description = m.group(1).strip().strip("\"'")[:300]

    if not description:
        # Fallback: find first non-empty line after frontmatter
        body = content[fm_match.end():] if fm_match else content
        for line in body.splitlines():
            line = line.strip()
            if line and not line.startswith("#") and not line.startswith("---"):
                description = line[:300]
                break

    return {
        "id": skill_dir.name,
        "name": name,
        "description": description,
        "category": category,
        "content": content,
        "tokens": tokenize(content)
    }

def main():
    if not USER_SKILLS_DIR.exists():
        print(f"Error: User skills directory not found at {USER_SKILLS_DIR}")
        sys.exit(1)

    print(f"Scanning skills in {USER_SKILLS_DIR}...")
    os.makedirs(DB_PATH.parent, exist_ok=True)
    
    conn = sqlite3.connect(str(DB_PATH))
    # Drop existing table if any to ensure clean schema rebuild
    conn.execute("DROP TABLE IF EXISTS skills")
    # Ensure database schema matches instructions
    conn.executescript("""
        CREATE TABLE IF NOT EXISTS skills (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            skill_id TEXT UNIQUE NOT NULL,
            name TEXT NOT NULL,
            description TEXT DEFAULT '',
            category TEXT DEFAULT 'general',
            content TEXT DEFAULT '',
            version INTEGER DEFAULT 1,
            similarity INTEGER DEFAULT 100,
            canonical_id INTEGER REFERENCES skills(id),
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    """)

    # Load existing skills to build our cache
    existing_skills = []
    cursor = conn.execute("SELECT id, skill_id, content FROM skills WHERE canonical_id IS NULL")
    for row in cursor.fetchall():
        existing_skills.append({
            "db_id": row[0],
            "skill_id": row[1],
            "tokens": tokenize(row[2])
        })
    print(f"Loaded {len(existing_skills)} existing canonical skills from database.")

    skills_to_process = []
    for d in USER_SKILLS_DIR.iterdir():
        if d.is_dir():
            info = extract_skill_info(d)
            if info:
                skills_to_process.append(info)

    print(f"Found {len(skills_to_process)} skills on disk to process.")
    
    new_canonical_count = 0
    duplicate_count = 0
    
    # Process and deduplicate with Jaccard similarity >= 90%
    for idx, skill in enumerate(skills_to_process):
        if idx % 200 == 0 and idx > 0:
            print(f"Processed {idx}/{len(skills_to_process)} skills...")

        # Find if it is similar to any existing canonical skill
        max_sim = 0.0
        best_match_db_id = None
        
        for canonical in existing_skills:
            sim = calculate_jaccard(skill["tokens"], canonical["tokens"])
            if sim > max_sim:
                max_sim = sim
                best_match_db_id = canonical["db_id"]

        if max_sim >= 0.90:
            # Duplicate
            duplicate_count += 1
            similarity_pct = int(max_sim * 100)
            try:
                conn.execute("""
                    INSERT INTO skills (skill_id, name, description, category, content, similarity, canonical_id)
                    VALUES (?, ?, ?, ?, ?, ?, ?)
                """, (skill["id"], skill["name"], skill["description"], skill["category"], skill["content"], similarity_pct, best_match_db_id))
            except sqlite3.IntegrityError:
                # Already exists in DB, skip or update
                pass
        else:
            # New canonical skill
            new_canonical_count += 1
            try:
                cursor = conn.cursor()
                cursor.execute("""
                    INSERT INTO skills (skill_id, name, description, category, content, similarity, canonical_id)
                    VALUES (?, ?, ?, ?, ?, 100, NULL)
                """, (skill["id"], skill["name"], skill["description"], skill["category"], skill["content"]))
                db_id = cursor.lastrowid
                
                # Add to existing_skills cache for future comparisons
                existing_skills.append({
                    "db_id": db_id,
                    "skill_id": skill["id"],
                    "tokens": skill["tokens"]
                })
            except sqlite3.IntegrityError:
                pass

    conn.commit()
    conn.close()
    
    print("\n=== Ingestion & Deduplication Summary ===")
    print(f"Total skills processed: {len(skills_to_process)}")
    print(f"New canonical skills added: {new_canonical_count}")
    print(f"Duplicate skills identified (>=90% similarity): {duplicate_count}")
    print("Done!")

if __name__ == "__main__":
    main()
