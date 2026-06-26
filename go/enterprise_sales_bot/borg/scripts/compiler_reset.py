#!/usr/bin/env python3
import pathlib
import sqlite3
import os
import subprocess
import re

DB_PATH = "data/assimilation_state.db"
GO_DIR = pathlib.Path("go")
TOOLS_DIR = GO_DIR / "internal" / "tools"

PROTECTED_FILES = {
    "registry.go", "parity.go", "factory.go", "basic_memory.go",
    "filesystem.go", "web_fetch.go", "sqlite.go", "bash.go", "glob.go",
    "apply_patch.go", "multi_edit.go", "git_ingest.go",
    # Core handler files
    "ddg_search.go", "gitingest.go", "search_tools.go",
    "skill_registry.go", "harnesses.go", "fetch.go", "ollama.go",
    "tts.go", "nws_tools.go", "nws_weather.go", "playwright_browser.go",
    "semgrep.go", "puppeteer.go", "prompt_library.go",
    "ripgrep.go"
}

def run_go_build():
    print("Running go build in", GO_DIR)
    r = subprocess.run(
        ["go", "build", "-buildvcs=false", "./cmd/tormentnexus"],
        cwd=GO_DIR,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    return r.returncode == 0, r.stdout + "\n" + r.stderr

def main():
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()

    iteration = 0
    while True:
        iteration += 1
        print(f"\n--- Iteration {iteration} ---")
        success, output = run_go_build()
        if success:
            print("SUCCESS: Go compilation succeeded cleanly!")
            break

        # Find all files with compilation errors in internal/tools/
        err_files = set()
        for line in output.splitlines():
            # Match internal/tools/filename.go or internal\tools\filename.go followed by colon and line number
            match = re.search(r"internal[\\/]tools[\\/]([a-zA-Z0-9_]+\.go):", line)
            if match:
                err_files.add(match.group(1))

        if not err_files:
            print("Compilation failed, but no errors matched internal/tools/ files:")
            # Print first 30 lines of output to help debug
            lines = output.strip().split("\n")
            for line in lines[:30]:
                print(line)
            break

        print(f"Found {len(err_files)} files with compilation errors: {err_files}")

        reset_count = 0
        deleted_count = 0

        for fname in err_files:
            if fname in PROTECTED_FILES:
                print(f"Skipping protected file: {fname}")
                continue

            fpath = TOOLS_DIR / fname
            # Reset database record
            c.execute(
                "UPDATE mcp_servers SET status='pending', notes=?, go_file=NULL, tools_exposed=NULL WHERE go_file=?",
                ("reset broken tool (compiler error)", fname)
            )
            rows = c.rowcount
            if rows > 0:
                reset_count += rows
            else:
                # Also try to match by name (stem) just in case
                stem = fpath.stem
                c.execute(
                    "UPDATE mcp_servers SET status='pending', notes=?, go_file=NULL, tools_exposed=NULL WHERE name=? OR package_name=?",
                    ("reset broken tool (compiler error)", stem, stem)
                )
                reset_count += c.rowcount

            if fpath.exists():
                try:
                    os.remove(fpath)
                    print(f"Deleted: {fname}")
                    deleted_count += 1
                except Exception as e:
                    print(f"Failed to delete {fname}: {e}")

        conn.commit()
        print(f"Reset {reset_count} DB records to pending. Deleted {deleted_count} files.")
        if deleted_count == 0:
            print("No files deleted in this round. Breaking to prevent infinite loop.")
            break

    conn.close()

if __name__ == "__main__":
    main()
