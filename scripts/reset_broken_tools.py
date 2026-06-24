#!/usr/bin/env python3
import pathlib
import sqlite3
import os
import subprocess

DB_PATH = "data/assimilation_state.db"
TOOLS_DIR = pathlib.Path("go/internal/tools")

PROTECTED_FILES = {
    "registry.go", "parity.go", "server.go", "factory.go", "basic_memory.go",
    "filesystem.go", "web_fetch.go", "sqlite.go", "bash.go", "glob.go",
    "apply_patch.go", "multi_edit.go", "git_ingest.go"
}

def gofmt_ok(p: pathlib.Path) -> bool:
    r = subprocess.run(
        ["gofmt", "-e", str(p)],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    return r.returncode == 0

def main():
    if not TOOLS_DIR.exists():
        print(f"ERROR: Tools directory {TOOLS_DIR} does not exist")
        return

    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()

    # Find all .go files in TOOLS_DIR
    go_files = list(TOOLS_DIR.glob("*.go"))
    print(f"Scanning {len(go_files)} Go files in {TOOLS_DIR} for malformed content...")

    reset_count = 0
    deleted_files = 0

    for fpath in go_files:
        if fpath.name in PROTECTED_FILES:
            continue

        size = fpath.stat().st_size
        if size == 0:
            continue  # Already handled by reset_empty_tools.py

        is_broken = False
        reason = ""

        # Check 1: Does it start with "package tools"
        try:
            content = fpath.read_text(encoding="utf-8", errors="ignore").strip()
            # Remove leading comments
            lines = [l.strip() for l in content.splitlines() if l.strip()]
            first_code_line = ""
            for line in lines:
                if not line.startswith("//") and not line.startswith("/*"):
                    first_code_line = line
                    break
            
            if not first_code_line.startswith("package tools"):
                is_broken = True
                reason = "does not start with 'package tools'"
        except Exception as e:
            is_broken = True
            reason = f"error reading file: {e}"

        # Check 2: Run gofmt -e
        if not is_broken and not gofmt_ok(fpath):
            is_broken = True
            reason = "fails gofmt compilation"

        if is_broken:
            server_name = fpath.stem
            print(f"Resetting broken tool: {fpath.name} ({reason})")
            
            # Reset database status to pending
            c.execute(
                "UPDATE mcp_servers SET status='pending', notes=?, go_file=NULL, tools_exposed=NULL WHERE name=? OR package_name=?",
                (f"reset broken tool: {reason}", server_name, server_name)
            )
            if c.rowcount > 0:
                reset_count += c.rowcount
            
            # Delete the file
            try:
                os.remove(fpath)
                deleted_files += 1
            except Exception as e:
                print(f"Failed to delete {fpath.name}: {e}")

    conn.commit()
    conn.close()

    print("\n=== Sanitization Summary ===")
    print(f"DB records reset to pending: {reset_count}")
    print(f"Broken Go files deleted:     {deleted_files}")

if __name__ == "__main__":
    main()
