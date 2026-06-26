"""
MCP Port Validator — Verifies Go implementations match original MCP servers.
Uses the same LLM proxy as the swarm for review/validation.

For each assimilated server:
  1. Reads the Go stub file
  2. Fetches the original MCP server's manifest from GitHub
  3. Validates the port is correct
  4. Reports clean / partial / mismatch
"""

import json
import os
import re
import sqlite3
import threading
import urllib.request
import urllib.error
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
from pathlib import Path

DB_PATH = "data/assimilation_state.db"
TOOLS_DIR = Path("go/internal/tools")
PROXY_URL = "http://localhost:4000/v1/chat/completions"
MAX_WORKERS = 3
VALIDATION_LOG = Path("data/port_validation.log")

log_lock = threading.Lock()


def log(msg, level="INFO"):
    ts = datetime.now().strftime("%H:%M:%S")
    with log_lock:
        with open(VALIDATION_LOG, "a", encoding="utf-8") as f:
            f.write(f"[{ts}][{level}] {msg}\n")
        print(f"[{ts}][{level}] {msg}")


def query_llm(prompt, model="free-llm", max_tokens=500):
    """Query the LLM proxy for validation."""
    payload = json.dumps(
        {
            "model": model,
            "messages": [{"role": "user", "content": prompt}],
            "max_tokens": max_tokens,
            "temperature": 0.1,
        }
    ).encode()

    req = urllib.request.Request(
        PROXY_URL,
        data=payload,
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=120) as r:
            data = json.loads(r.read())
            return data["choices"][0]["message"]["content"]
    except Exception as e:
        return f"LLM_ERROR: {e}"


def find_go_file(server_name):
    """Find the Go file for a given server name."""
    safe = server_name.lower().replace(" ", "_").replace("-", "_")
    safe = re.sub(r"[^a-z0-9_]", "_", safe)

    for fn in os.listdir(TOOLS_DIR):
        if not fn.endswith(".go"):
            continue
        if safe[:15] in fn.lower() or safe.replace("_", "") in fn.lower().replace(
            "_", ""
        ):
            return fn
    return None


def get_go_content(filename):
    """Read Go file content."""
    fp = TOOLS_DIR / filename
    try:
        with open(fp, "r", encoding="utf-8", errors="replace") as f:
            return f.read()
    except FileNotFoundError:
        return ""


def fetch_github_readme(github_url):
    """Fetch the README from a GitHub repo for MCP tool info."""
    if not github_url or "github.com" not in github_url:
        return ""

    url = github_url.rstrip("/").replace("github.com", "raw.githubusercontent.com")

    for branch in ["main", "master"]:
        for path in ["README.md", "mcp.json", "package.json"]:
            try:
                req = urllib.request.Request(
                    f"{url}/{branch}/{path}",
                    headers={"User-Agent": "Mozilla/5.0"},
                )
                with urllib.request.urlopen(req, timeout=10) as r:
                    content = r.read().decode("utf-8", errors="replace")[:8000]
                    if content.strip():
                        return f"[{path} from {branch}]\n" + content
            except Exception:
                continue
    return ""


def validate_server(name, github_url):
    """Validate a single server's Go port."""
    log(f"Validating: {name}")

    go_fn = find_go_file(name)
    if not go_fn:
        return name, "MISSING", "no Go file found"

    go_code = get_go_content(go_fn)
    if not go_code.strip():
        return name, "EMPTY_STUB", "Go file is empty"

    if "//go:build ignore" in go_code:
        # Build-constrained stub — check if it at least has the right package
        if "package tools" in go_code:
            return name, "STUB", "build-constrained stub"
        return name, "INVALID_STUB", "missing package declaration"

    # Real implementation — validate against original MCP source
    original_info = fetch_github_readme(github_url) if github_url else ""

    if not original_info:
        return name, "NO_ORIGINAL", "cannot fetch original MCP source"

    # Have LLM validate the port
    prompt = f"""You are validating a Go port of an MCP server.

Original MCP server ({name}):
Source: {github_url}
Description/Manifest:
{original_info[:3000]}

Go implementation:
```go
{go_code[:4000]}
```

Evaluate if the Go code is a clean port of the original MCP server.
Respond with exactly ONE line:
- CLEAN_PORT if the Go code correctly implements the original MCP tools
- PARTIAL_PORT if the Go code implements some but not all tools
- MISMATCH if the Go code doesn't match the original server at all
- STUB_ONLY if it's just a skeleton/framework

Then on the next line, briefly explain why (max 100 chars)."""

    result = query_llm(prompt, model="free-llm", max_tokens=200)

    # Parse the result
    lines = result.strip().split("\n")
    status = lines[0].strip() if lines else "UNKNOWN"
    reason = lines[1].strip() if len(lines) > 1 else ""

    # Clean up status
    for valid in ["CLEAN_PORT", "PARTIAL_PORT", "MISMATCH", "STUB_ONLY"]:
        if valid in status.upper():
            status = valid
            break
    else:
        if "LLM_ERROR" in result:
            status = "LLM_ERROR"
            reason = result[:100]
        else:
            status = "UNKNOWN"
            reason = result[:100]

    log(f"  -> {name}: {status} ({reason})")
    return name, status, reason


def main():
    db = sqlite3.connect(DB_PATH)
    servers = db.execute(
        "SELECT name, github_url FROM mcp_servers WHERE status='implemented' ORDER BY RANDOM() LIMIT 100"
    ).fetchall()
    db.close()

    log(f"Starting port validation for {len(servers)} servers...")
    log(f"Worker pool: {MAX_WORKERS}")
    log(f"Proxy: {PROXY_URL}")
    log("-" * 60)

    results = {
        "CLEAN_PORT": 0,
        "PARTIAL_PORT": 0,
        "MISMATCH": 0,
        "STUB": 0,
        "STUB_ONLY": 0,
        "MISSING": 0,
        "EMPTY_STUB": 0,
        "NO_ORIGINAL": 0,
        "LLM_ERROR": 0,
        "INVALID_STUB": 0,
        "UNKNOWN": 0,
    }

    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as pool:
        futs = {pool.submit(validate_server, name, url): name for name, url in servers}
        for f in as_completed(futs):
            try:
                name, status, reason = f.result()
                results[status] = results.get(status, 0) + 1
            except Exception as e:
                log(f"Error: {futs[f]}: {e}", "ERROR")
                results["UNKNOWN"] = results.get("UNKNOWN", 0) + 1

    log("=" * 60)
    log("VALIDATION SUMMARY")
    log("=" * 60)
    for status, count in sorted(results.items()):
        if count > 0:
            log(f"  {status}: {count}")
    log(f"  TOTAL: {sum(results.values())}")
    log("=" * 60)


if __name__ == "__main__":
    main()
