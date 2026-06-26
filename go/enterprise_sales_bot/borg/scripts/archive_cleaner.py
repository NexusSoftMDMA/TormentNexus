#!/usr/bin/env python3
import os
import shutil
import glob
from pathlib import Path

ROOT = Path(".")
ARCHIVE = ROOT / "archive"

# Define destination subfolders
SWARM_DIR = ARCHIVE / "old_swarm_versions"
SCRATCH_DIR = ARCHIVE / "scratch_scripts"
CHATS_DIR = ARCHIVE / "chat_exports"
LOGS_DIR = ARCHIVE / "text_logs"
DB_DIR = ARCHIVE / "old_databases"

def ensure_dirs():
    for d in [SWARM_DIR, SCRATCH_DIR, CHATS_DIR, LOGS_DIR, DB_DIR]:
        d.mkdir(parents=True, exist_ok=True)

def safe_move(src_path: Path, dst_dir: Path):
    if not src_path.exists():
        return
    dst_path = dst_dir / src_path.name
    print(f"Moving: {src_path} -> {dst_path}")
    try:
        shutil.move(str(src_path), str(dst_path))
    except Exception as e:
        print(f"Error moving {src_path}: {e}")

def main():
    print("=== TormentNexus Workspace Archiving & Cleanup ===")
    ensure_dirs()

    # 1. Old Swarm Scripts
    old_swarms = ["swarm.py", "swarm_v2.py", "swarm_v3.py", "swarm_v4.py", "swarm_v5.py", "swarm_v6.py"]
    for f in old_swarms:
        safe_move(ROOT / f, SWARM_DIR)

    # 2. Scratch/Temporary Scripts
    scratch_files = glob.glob("tmp_*.*") + glob.glob("go_*.py") + ["fix_build3.py"]
    for f in scratch_files:
        safe_move(ROOT / f, SCRATCH_DIR)

    # 3. Chat Exports
    chat_exports = glob.glob("chat_export_*.md")
    for f in chat_exports:
        safe_move(ROOT / f, CHATS_DIR)

    # 4. Text Logs/Lists
    logs_files = ["broken_list.txt", "broken_list2.txt", "batch_001_mcp.txt", "current_broken.txt"]
    for f in logs_files:
        safe_move(ROOT / f, LOGS_DIR)

    # 5. Restored/Old Databases
    old_dbs = glob.glob("db_v*.db") + ["tormentnexus_restored.db", "mydatabase.db", "old_db_restored.db"]
    for f in old_dbs:
        safe_move(ROOT / f, DB_DIR)

    print("Cleanup completed successfully!")

if __name__ == "__main__":
    main()
