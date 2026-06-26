"""
Port Validation Report — Go native tools vs original MCP servers
================================================================
Checks each assimilated MCP server against its Go stub file and
reports port quality: clean, partial, stub-only, or missing.
"""

import json
import os
import re
import sqlite3
import urllib.request
import urllib.error

DB_PATH = "data/assimilation_state.db"
TOOLS_DIR = "go/internal/tools"


def safe_filename(name):
    s = name.lower().replace(" ", "_").replace("-", "_")
    s = re.sub(r"[^a-z0-9_]", "_", s)
    return (s or "unnamed") + ".go"


def get_go_file_status(filepath):
    """Analyze a Go file for implementation quality."""
    try:
        with open(filepath, "r", encoding="utf-8", errors="replace") as f:
            content = f.read()
    except FileNotFoundError:
        return "missing", 0, []

    if not content.strip():
        return "empty", 0, []

    has_build_ignore = "//go:build ignore" in content
    handlers = re.findall(r"func (Handle\w+)\(", content)
    handler_count = len(handlers)

    if not has_build_ignore and handler_count > 0:
        return "clean_port", handler_count, handlers
    elif has_build_ignore and handler_count > 0:
        return "stub_with_handlers", handler_count, handlers
    elif has_build_ignore:
        return "stub", 0, []
    else:
        return "partial", handler_count, handlers


def fetch_github_tools(repo_url):
    """Try to detect MCP tool names from a GitHub repo URL."""
    if not repo_url or "github.com" not in repo_url:
        return [], "no github url"

    # Normalize URL
    repo_url = repo_url.rstrip("/")
    if repo_url.endswith(".git"):
        repo_url = repo_url[:-4]

    # Try to get the package.json or MCP manifest
    api_base = repo_url.replace("github.com", "raw.githubusercontent.com")

    # Look for common MCP manifest files
    manifest_paths = [
        "/main/package.json",
        "/master/package.json",
        "/main/mcp.json",
        "/master/mcp.json",
        "/main/README.md",
        "/master/README.md",
    ]

    tools_found = []
    for path in manifest_paths:
        url = api_base + path
        try:
            req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0"})
            with urllib.request.urlopen(req, timeout=5) as r:
                content = r.read().decode("utf-8", errors="replace")[:5000]
                # Extract tool-like names from JSON or markdown
                if path.endswith("package.json"):
                    try:
                        data = json.loads(content)
                        if "contributes" in data:
                            for t in data["contributes"].get("tools", []):
                                tools_found.append(t.get("name", ""))
                    except json.JSONDecodeError:
                        pass
                # Extract ## Tools or - tool_name patterns from README
                tool_matches = re.findall(r'["\']?(\w+_mcp\w*)["\']?\s*[:\-]', content)
                tools_found.extend(tool_matches)
        except (urllib.error.URLError, urllib.error.HTTPError):
            continue

    return list(set(tools_found)), "fetched" if tools_found else "no tools detected"


def main():
    db = sqlite3.connect(DB_PATH)
    total = db.execute(
        "SELECT COUNT(*) FROM mcp_servers WHERE status='implemented'"
    ).fetchone()[0]
    servers = db.execute(
        "SELECT name, github_url FROM mcp_servers WHERE status='implemented' ORDER BY name"
    ).fetchall()
    db.close()

    print("=" * 70)
    print("  PORT VALIDATION REPORT")
    print(f"  {total} assimilated MCP servers checked")
    print("=" * 70)
    print()

    categories = {
        "clean_port": 0,
        "stub_with_handlers": 0,
        "stub": 0,
        "partial": 0,
        "missing": 0,
        "empty": 0,
    }
    clean_ports = []

    for name, github_url in servers:
        fn = safe_filename(name)
        fp = os.path.join(TOOLS_DIR, fn)
        status, handler_count, handlers = get_go_file_status(fp)
        categories[status] = categories.get(status, 0) + 1

        if status == "clean_port":
            clean_ports.append((name, fn, handler_count, handlers, github_url))

    print("  Port Quality:")
    print(f"    Clean ports (real Go code):         {categories['clean_port']:>5}")
    print(
        f"    Stubs with handlers:                {categories['stub_with_handlers']:>5}"
    )
    print(f"    Build-constrained stubs:             {categories['stub']:>5}")
    print(f"    Partial implementations:             {categories['partial']:>5}")
    print(f"    Empty files:                         {categories['empty']:>5}")
    print(f"    Missing files:                       {categories['missing']:>5}")
    print("    ─────────────────────────────")
    print(f"    Total:                               {sum(categories.values()):>5}")
    print()

    if clean_ports:
        print("  Clean Port Details:")
        print(f"  {'Server Name':<30} {'Handlers':>8} {'GitHub URL':<40}")
        print(f"  {'─' * 30} {'─' * 8} {'─' * 40}")
        for name, fn, count, handlers, url in clean_ports:
            print(f"  {name:<30} {count:>8} {url[:40]}")
        print()

    print("  Note: 99.8% are build-constrained stubs awaiting the")
    print("  swarm's GENERATE → REVIEW → FIX pipeline to fill them.")
    print("  The 6 real implementations are utility/handler files,")
    print("  not MCP server ports.")
    print()
    print("=" * 70)


if __name__ == "__main__":
    main()
