import sqlite3
import re
import uuid
import json

DB_PATH = r"c:\Users\hyper\workspace\borg\tormentnexus.db"
MD_PATH = r"c:\Users\hyper\workspace\bobbybookmarks\research\mcp_comprehensive_list.md"

def extract_github_info(url):
    m = re.search(r'github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)', url, re.IGNORECASE)
    if m:
        owner = m.group(1).lower().strip()
        repo = m.group(2).lower().strip()
        # Clean common github noise
        if owner in ["features", "marketplace", "pricing", "collections", "topics", "trending", "notifications", "settings", "orgs"]:
            return None, None
        if repo in ["issues", "pulls", "wiki", "releases", "actions", "projects"]:
            return None, None
        return owner, repo
    return None, None

def main():
    print("=== INGESTING COMPREHENSIVE LIST ===")
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()

    # Load existing to avoid duplicates
    c.execute("SELECT canonical_id FROM published_mcp_servers")
    existing_cids = {row[0] for row in c.fetchall()}
    print(f"Loaded {len(existing_cids)} existing servers from database.")

    with open(MD_PATH, 'r', encoding='utf-8') as f:
        content = f.read()

    # Extract all github URLs
    urls = re.findall(r'https?://github\.com/[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+', content)
    
    discovered_repos = {}
    for url in urls:
        owner, repo = extract_github_info(url)
        if owner and repo:
            cid = f"github/{owner}/{repo}"
            discovered_repos[cid] = {
                "owner": owner,
                "repo": repo,
                "url": f"https://github.com/{owner}/{repo}",
                "display": repo.replace("-", " ").replace("_", " ").title()
            }
            
    print(f"Found {len(discovered_repos)} unique GitHub repos in the markdown.")

    added = 0
    skipped = 0

    for cid, s in discovered_repos.items():
        if cid in existing_cids:
            skipped += 1
            continue

        server_uuid = str(uuid.uuid4())
        tags_json = json.dumps(["github-repo", "comprehensive-list"])
        categories_json = json.dumps(["comprehensive-imported"])

        c.execute("""
            INSERT INTO published_mcp_servers (
                uuid, canonical_id, display_name, description, repository_url, homepage_url,
                transport, install_method, tags, categories, status, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, 'stdio', 'unknown', ?, ?, 'discovered', strftime('%s', 'now'), strftime('%s', 'now'))
        """, (
            server_uuid, cid, s["display"], f"Discovered MCP repository: {s['owner']}/{s['repo']} from comprehensive list",
            s["url"], s["url"], tags_json, categories_json
        ))

        # Provenance
        src_uuid = str(uuid.uuid4())
        c.execute("""
            INSERT INTO published_mcp_server_sources (uuid, server_uuid, source_name, source_url, raw_payload)
            VALUES (?, ?, 'comprehensive-list', ?, ?)
        """, (src_uuid, server_uuid, MD_PATH, json.dumps(s)))

        added += 1
        existing_cids.add(cid)

    conn.commit()
    conn.close()

    print(f"\nIngestion Complete:")
    print(f"  Added:           {added} new MCP servers")
    print(f"  Skipped (Dupes): {skipped}")
    print(f"  Total Registered: {len(existing_cids)}")

if __name__ == "__main__":
    main()
