"""Match Go files to MCP servers and rebuild assimilation DB status."""
import os
import re
import sqlite3
from collections import defaultdict

tools_dir = "go/internal/tools"
db = sqlite3.connect("data/assimilation_state.db")

def name_to_filename(name):
    n = name
    for p in ["mcp-server-", "mcp-", "@", "server-", "mcp_"]:
        if n.lower().startswith(p):
            n = n[len(p):]
    n = re.sub(r"[/.\-]+", "_", n)
    n = re.sub(r"[^a-zA-Z0-9_]", "", n).lower()
    return n[:60]

# Build lookup: short_hash -> [server_names]
print("Building lookup index...")
server_map = defaultdict(list)
server_names = [s[0] for s in db.execute("SELECT name FROM mcp_servers").fetchall()]
for sname in server_names:
    fn = name_to_filename(sname)
    server_map[fn[:15]].append(sname)
    server_map[fn[:10]].append(sname)

matched = 0
unmatched = 0
real_count = 0
stub_count = 0

print(f"Scanning {len([f for f in os.listdir(tools_dir) if f.endswith('.go')])} Go files...")

for fn in sorted(os.listdir(tools_dir)):
    if not fn.endswith(".go"):
        continue
    fp = os.path.join(tools_dir, fn)
    try:
        with open(fp, "r", encoding="utf-8", errors="replace") as f:
            content = f.read()
    except:
        continue
    
    if not content.strip():
        continue
    
    has_build_ignore = "//go:build ignore" in content
    has_handlers = "func Handle" in content
    
    if has_handlers and not has_build_ignore:
        kind = "real"
        real_count += 1
    elif has_handlers and has_build_ignore:
        kind = "stub_h"
        stub_count += 1
    elif has_build_ignore:
        kind = "stub"
        stub_count += 1
    else:
        continue
    
    go_stem = fn[:-3]
    
    # Find matching server
    candidates = server_map.get(go_stem[:15], []) or server_map.get(go_stem[:10], [])
    
    # Filter to best match
    best = None
    for c in candidates:
        cf = name_to_filename(c)
        if go_stem == cf:
            best = c
            break
        elif go_stem.startswith(cf) or cf.startswith(go_stem):
            best = c
            break
    
    if best:
        db.execute(
            "UPDATE mcp_servers SET status='implemented', notes='recovered' WHERE name=?",
            (best,)
        )
        matched += 1
    else:
        unmatched += 1
        if kind == "real":
            display = go_stem.replace("_", " ").title()
            db.execute(
                "INSERT OR IGNORE INTO mcp_servers (name, github_url, score, status, notes) VALUES (?, '', 0, 'implemented', 'orphan')",
                (display,)
            )

db.commit()

total_impl = db.execute("SELECT COUNT(*) FROM mcp_servers WHERE status='implemented'").fetchone()[0]
total_pend = db.execute("SELECT COUNT(*) FROM mcp_servers WHERE status='pending'").fetchone()[0]
total_fail = db.execute("SELECT COUNT(*) FROM mcp_servers WHERE status='failed'").fetchone()[0]

print("\n=== Results ===")
print(f"Matched:  {matched}")
print(f"Unmatched: {unmatched}")
print("")
print(f"Go files: {real_count} real + {stub_count} stubs = {real_count + stub_count}")
print("")
print("DB now:")
print(f"  Implemented: {total_impl}")
print(f"  Pending:     {total_pend}")
print(f"  Failed:      {total_fail}")

db.close()
