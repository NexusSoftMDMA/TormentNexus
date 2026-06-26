#!/usr/bin/env python3
import sqlite3
import re
import os
import hashlib
from pathlib import Path

# Paths
WORKSPACE = Path(r"c:\Users\hyper\workspace\tormentnexus")
HOME = Path(os.path.expanduser("~"))

# Source folders to scan for SKILL.md/PROMPT.md files (strictly defined to be fast)
SCAN_DIRS = [
    WORKSPACE / "packages" / "core" / "src" / "skills",
    WORKSPACE / ".tormentnexus" / "skills",
    HOME / ".tormentnexus" / "skills",
    WORKSPACE / ".agent" / "skills",
    WORKSPACE / ".borg" / "skills",
    WORKSPACE / "skill-registry",
    WORKSPACE / "taskplane-tasks",
    WORKSPACE / "hermes-addons",
    WORKSPACE / "mcp-assimilation",
    WORKSPACE / "third_party" / "rtk" / ".claude" / "skills",
]

# Database paths
SKILLS_DB_PATH = WORKSPACE / ".tormentnexus" / "skills.db"
CATALOG_DB_PATH = WORKSPACE / "catalog.db"

# Ignored directories during walk
IGNORED_DIRS = {
    "node_modules", ".venv", ".git", ".ruff_cache", ".pytest_cache",
    "site-packages", "__pycache__", ".next", ".turbo", "dist", "bin"
}

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

def clean_content(text):
    """Normalize whitespace and lowercase to compare content core logic."""
    # Remove markdown titles and frontmatter
    text = re.sub(r"^---[\s\S]*?---", "", text)
    text = re.sub(r"^#\s+.*", "", text, flags=re.MULTILINE)
    # Remove spacing and normalize
    return "".join(text.split()).lower()

def extract_skill_info(file_path: Path) -> dict:
    try:
        content = file_path.read_text(encoding="utf-8", errors="replace")
    except Exception as e:
        print(f"Error reading {file_path}: {e}")
        return None

    # Determine if it's a prompt
    is_prompt = "prompt" in file_path.name.lower() or "prompt" in file_path.parent.name.lower()
    category = "prompt" if is_prompt else "skill"
    description = ""
    name = file_path.parent.name

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

    clean_id = re.sub(r'[^a-zA-Z0-9_-]', '', name.lower().replace(" ", "_").replace("-", "_")).strip()
    if not clean_id:
        clean_id = "skill_" + hashlib.md5(name.encode()).hexdigest()[:6]

    cleaned_body = clean_content(content)
    content_hash = hashlib.sha256(cleaned_body.encode('utf-8')).hexdigest()

    return {
        "id": clean_id,
        "name": name,
        "description": description or f"Structured runbook for {name}",
        "category": category,
        "content": content,
        "tokens": tokenize(content),
        "content_hash": content_hash,
        "source_path": file_path
    }

def main():
    print("=== STARTING COMPREHENSIVE RESOURCE ASSIMILATION ===")

    # Initialize / Connect to DB
    SKILLS_DB_PATH.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(str(SKILLS_DB_PATH))
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

    # Load existing skill IDs to perform incremental ingestion
    existing_ids = {r[0] for r in conn.execute("SELECT skill_id FROM skills").fetchall()}
    print(f"Loaded {len(existing_ids)} existing resource IDs from database.")

    # 1. Scan filesystem for all SKILL.md/PROMPT.md resources
    all_resources = []
    seen_paths = set()

    for scan_dir in SCAN_DIRS:
        if not scan_dir.exists():
            continue
        print(f"Scanning directory: {scan_dir}")
        for root, dirs, files in os.walk(scan_dir):
            # Limit search depth to 3 levels to avoid traversing massive subrepos
            depth = len(Path(root).relative_to(scan_dir).parts)
            if depth > 3:
                dirs.clear()
                continue
                
            dirs[:] = [d for d in dirs if d not in IGNORED_DIRS]
            for file in files:
                if file.lower() in ("skill.md", "skill.markdown", "prompt.md", "prompt.markdown"):
                    fpath = Path(root) / file
                    
                    # Deduplicate by path
                    abs_path = fpath.resolve()
                    if abs_path in seen_paths:
                        continue
                    seen_paths.add(abs_path)
                    
                    # Fast check: skip if folder name is already in existing_ids
                    clean_id = re.sub(r'[^a-zA-Z0-9_-]', '', fpath.parent.name.lower().replace(" ", "_").replace("-", "_")).strip()
                    if clean_id in existing_ids:
                        continue
                        
                    info = extract_skill_info(fpath)
                    if info:
                        if info["id"] not in existing_ids:
                            all_resources.append(info)

    print(f"Found {len(all_resources)} new file-based skill/prompt resources to process.")

    # 2. Extract MCP servers from databases and convert to markdown resources
    mcp_servers_count = 0
    if CATALOG_DB_PATH.exists():
        try:
            print(f"Querying MCP servers from {CATALOG_DB_PATH}...")
            conn_catalog = sqlite3.connect(str(CATALOG_DB_PATH))
            cursor = conn_catalog.cursor()
            cursor.execute("""
                SELECT s.display_name, s.description, s.repository_url, s.homepage_url, r.template, r.required_secrets
                FROM published_mcp_servers s
                LEFT JOIN published_mcp_config_recipes r ON s.uuid = r.server_uuid
                WHERE s.display_name IS NOT NULL
            """)
            for name, desc, repo, home, recipe, secrets in cursor.fetchall():
                clean_id = f"mcp_server_{name.lower().replace(' ', '_').replace('-', '_')}"
                clean_id = re.sub(r'[^a-zA-Z0-9_-]', '', clean_id).strip()
                
                # Incrementally skip already stored MCP servers
                if clean_id in existing_ids:
                    continue
                    
                desc = desc or "No description provided."
                repo = repo or home or "N/A"
                recipe = recipe or "N/A"
                secrets = secrets or "[]"
                
                content = f"""---
name: mcp-server-{name.lower().replace(' ', '-')}
description: MCP Server wrapper for {name}
category: mcp_server
---

# MCP Server: {name}

{desc}

- **Repository**: {repo}
- **Required Secrets**: {secrets}

## Installation Recipe
```json
{recipe}
```
"""
                cleaned_body = clean_content(content)
                content_hash = hashlib.sha256(cleaned_body.encode('utf-8')).hexdigest()

                all_resources.append({
                    "id": clean_id,
                    "name": f"MCP: {name}",
                    "description": desc[:300],
                    "category": "mcp_server",
                    "content": content,
                    "tokens": tokenize(content),
                    "content_hash": content_hash,
                    "source_path": None
                })
                mcp_servers_count += 1
            conn_catalog.close()
        except Exception as e:
            print(f"Error querying catalog.db: {e}")
            
    print(f"Ingested {mcp_servers_count} new MCP server resources from catalog.db.")

    if len(all_resources) == 0:
        print("No new resources to ingest. Deduplication complete.")
        conn.close()
        return

    # Load existing canonical skills for Jaccard comparisons and fast exact matching
    existing_canonical = []
    seen_hashes = {} # maps content_hash -> db_id
    
    for r in conn.execute("SELECT id, skill_id, content, category FROM skills").fetchall():
        db_id, skill_id, content, category = r
        cleaned_body = clean_content(content)
        h = hashlib.sha256(cleaned_body.encode('utf-8')).hexdigest()
        seen_hashes[h] = db_id
        if category != 'mcp_server':
            existing_canonical.append({
                "db_id": db_id,
                "id": skill_id,
                "tokens": tokenize(content)
            })

    print(f"Loaded {len(existing_canonical)} canonical versions for deduplication.")

    seen_mcp_ids = {} # maps clean_id to DB id
    for r in conn.execute("SELECT id, skill_id FROM skills WHERE category='mcp_server'").fetchall():
        seen_mcp_ids[r[1]] = r[0]

    new_canonical_count = 0
    duplicate_count = 0

    print("Deduplicating and inserting into database...")
    for idx, resource in enumerate(all_resources):
        if idx % 1000 == 0 and idx > 0:
            print(f"  Processed {idx}/{len(all_resources)} resources...")

        # Exact Content Hash check first (extremely fast O(1))
        chash = resource["content_hash"]
        if chash in seen_hashes:
            duplicate_count += 1
            try:
                conn.execute("""
                    INSERT OR IGNORE INTO skills (skill_id, name, description, category, content, similarity, canonical_id)
                    VALUES (?, ?, ?, ?, ?, 100, ?)
                """, (resource["id"] + f"_dup_{duplicate_count}", resource["name"], resource["description"], resource["category"], resource["content"], seen_hashes[chash]))
            except sqlite3.IntegrityError:
                pass
            continue

        if resource["category"] == "mcp_server":
            # Fast deduplicate for MCP servers by clean_id
            mcp_id = resource["id"]
            if mcp_id in seen_mcp_ids:
                duplicate_count += 1
                try:
                    conn.execute("""
                        INSERT OR IGNORE INTO skills (skill_id, name, description, category, content, similarity, canonical_id)
                        VALUES (?, ?, ?, ?, ?, 100, ?)
                    """, (resource["id"] + f"_dup_{duplicate_count}", resource["name"], resource["description"], resource["category"], resource["content"], seen_mcp_ids[mcp_id]))
                except sqlite3.IntegrityError:
                    pass
            else:
                new_canonical_count += 1
                try:
                    cursor = conn.cursor()
                    cursor.execute("""
                        INSERT OR IGNORE INTO skills (skill_id, name, description, category, content, similarity, canonical_id)
                        VALUES (?, ?, ?, ?, ?, 100, NULL)
                    """, (resource["id"], resource["name"], resource["description"], resource["category"], resource["content"]))
                    db_id = cursor.lastrowid
                    if db_id:
                        seen_mcp_ids[mcp_id] = db_id
                        seen_hashes[chash] = db_id
                except sqlite3.IntegrityError:
                    pass
        else:
            # Jaccard deduplicate for skills/prompts (fast, since total is small)
            max_sim = 0.0
            best_match_db_id = None
            
            len_a = len(resource["tokens"])
            for canonical in existing_canonical:
                len_b = len(canonical["tokens"])
                
                # Mathematical pruning: Jaccard similarity <= min(len_a, len_b) / max(len_a, len_b)
                if len_a == 0 or len_b == 0:
                    continue
                if min(len_a, len_b) / max(len_a, len_b) < 0.90:
                    continue
                    
                sim = calculate_jaccard(resource["tokens"], canonical["tokens"])
                if sim > max_sim:
                    max_sim = sim
                    best_match_db_id = canonical["db_id"]

            if max_sim >= 0.90:
                # Duplicate
                duplicate_count += 1
                similarity_pct = int(max_sim * 100)
                try:
                    conn.execute("""
                        INSERT OR IGNORE INTO skills (skill_id, name, description, category, content, similarity, canonical_id)
                        VALUES (?, ?, ?, ?, ?, ?, ?)
                    """, (resource["id"], resource["name"], resource["description"], resource["category"], resource["content"], similarity_pct, best_match_db_id))
                except sqlite3.IntegrityError:
                    pass
            else:
                # New canonical resource
                new_canonical_count += 1
                try:
                    cursor = conn.cursor()
                    cursor.execute("""
                        INSERT OR IGNORE INTO skills (skill_id, name, description, category, content, similarity, canonical_id)
                        VALUES (?, ?, ?, ?, ?, 100, NULL)
                    """, (resource["id"], resource["name"], resource["description"], resource["category"], resource["content"]))
                    db_id = cursor.lastrowid
                    
                    if db_id:
                        existing_canonical.append({
                            "db_id": db_id,
                            "id": resource["id"],
                            "tokens": resource["tokens"]
                        })
                        seen_hashes[chash] = db_id
                        
                        # Write to filesystem only if file does not exist to optimize disk I/O speed
                        for dest_folder in [WORKSPACE / ".tormentnexus" / "skills", HOME / ".tormentnexus" / "skills"]:
                            skill_folder = dest_folder / resource["id"]
                            target_file = skill_folder / "SKILL.md"
                            if not target_file.exists():
                                skill_folder.mkdir(parents=True, exist_ok=True)
                                target_file.write_text(resource["content"], encoding="utf-8")
                except sqlite3.IntegrityError:
                    pass

    conn.commit()
    conn.close()

    print("\n=== Ingestion & Deduplication Summary ===")
    print(f"Total Resources Processed: {len(all_resources)}")
    print(f"New Canonical Resources Indexed: {new_canonical_count}")
    print(f"Duplicate Resources Identified (>=90% similarity): {duplicate_count}")
    print("Done!")

if __name__ == "__main__":
    main()
