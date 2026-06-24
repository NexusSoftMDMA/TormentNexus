#!/usr/bin/env python3
import pathlib
import sqlite3
import os

DB_PATH = "data/assimilation_state.db"
TOOLS_DIR = pathlib.Path("go/internal/tools")

def main():
    if not TOOLS_DIR.exists():
        print(f"ERROR: Tools directory {TOOLS_DIR} does not exist")
        return

    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()

    # Find all .go files in TOOLS_DIR
    go_files = list(TOOLS_DIR.glob("*.go"))
    print(f"Scanning {len(go_files)} Go files in {TOOLS_DIR}...")

    reset_count = 0
    deleted_files = 0

    for fpath in go_files:
        # Ignore registry.go, parity.go, etc. which are core files
        if fpath.name in ("registry.go", "parity.go", "server.go", "factory.go", "basic_memory.go", "filesystem.go", "web_fetch.go", "sqlite.go", "bash.go", "glob.go", "apply_patch.go", "multi_edit.go", "git_ingest.go"):
            continue

        size = fpath.stat().st_size
        
        # We consider a file empty or a stub if it is less than 150 bytes
        is_empty = False
        content = ""
        if size < 150:
            is_empty = True
        else:
            try:
                content = fpath.read_text(encoding="utf-8", errors="ignore").strip()
                # If it only contains package declaration or placeholder comments
                if len(content) < 150 or "placeholder" in content.lower() or "stub" in content.lower():
                    is_empty = True
            except Exception as e:
                print(f"Error reading {fpath.name}: {e}")
                is_empty = True

        if is_empty:
            server_name = fpath.stem
            print(f"Resetting empty/stub tool: {fpath.name} ({size} bytes)")
            
            # Reset database status to pending
            c.execute(
                "UPDATE mcp_servers SET status='pending', notes='reset empty tool', go_file=NULL, tools_exposed=NULL WHERE name=? OR package_name=?",
                (server_name, server_name)
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
    print(f"Empty Go files deleted:     {deleted_files}")

if __name__ == "__main__":
    main()
