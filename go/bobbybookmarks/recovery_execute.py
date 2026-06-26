#!/usr/bin/env python3
"""
Borg Conversation Data Loss — Revised Recovery Execution
Phase 2+: Re-ingest archived session transcripts to restore memories
"""

import sqlite3
import json
import gzip
import uuid as uuid_mod
import time
from pathlib import Path

PROJECT = Path(r"C:\Users\hyper\workspace\bobbybookmarks")
DB_PATH = PROJECT / "tormentnexus.db"
ARCHIVE = PROJECT / ".tormentnexus/imported_sessions/archive/sessions"


def get_existing_hashes(conn):
    """Get set of transcript hashes already in DB."""
    cur = conn.cursor()
    cur.execute("SELECT transcript_hash FROM imported_sessions")
    return {row[0] for row in cur.fetchall()}


def ingest_archived_sessions():
    """Re-ingest archived session transcripts not yet in DB."""
    conn = sqlite3.connect(str(DB_PATH))
    existing = get_existing_hashes(conn)

    txt_files = sorted(ARCHIVE.glob("*.txt.gz"))
    print(f"Found {len(txt_files)} archived transcript files")

    ingested = 0
    for tf in txt_files:
        # Hash is filename without .txt.gz
        stem = tf.stem  # e.g., abc123.txt -> abc123
        if stem.endswith(".txt"):
            stem = stem[:-4]

        if stem in existing:
            continue

        # Read the transcript content
        try:
            with gzip.open(tf, "rt", encoding="utf-8", errors="replace") as f:
                transcript = f.read()
        except Exception as e:
            print(f"  Error reading {tf.name}: {e}")
            continue

        # Find corresponding meta file
        meta_path = tf.with_suffix(".meta.json.gz")
        metadata = {}
        if meta_path.exists():
            try:
                with gzip.open(
                    meta_path, "rt", encoding="utf-8", errors="replace"
                ) as f:
                    metadata = json.load(f)
            except:
                pass

        # Generate UUID and insert
        uuid_str = str(uuid_mod.uuid4())
        source_tool = metadata.get("sourceTool", "archive-restore")
        title = metadata.get("title", "")
        session_format = metadata.get("sessionFormat", "txt")
        source_path = metadata.get("sourcePath", str(tf))
        content_len = len(transcript)
        now_ms = int(time.time() * 1000)

        try:
            conn.execute(
                """
                INSERT INTO imported_sessions 
                (uuid, source_tool, source_path, source_size, source_mtime, 
                 title, session_format, transcript, excerpt, 
                 transcript_hash, transcript_archive_path,
                 transcript_metadata_archive_path, transcript_archive_format,
                 transcript_stored_bytes, normalized_session, metadata,
                 discovered_at, imported_at, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
                (
                    uuid_str,
                    source_tool,
                    source_path,
                    content_len,
                    now_ms,
                    title,
                    session_format,
                    transcript,
                    transcript[:200],
                    stem,
                    str(tf.relative_to(PROJECT)),
                    str(meta_path.relative_to(PROJECT)) if meta_path.exists() else "",
                    "gzip-text-v1",
                    content_len,
                    "{}",
                    "{}",
                    now_ms,
                    now_ms,
                    now_ms,
                    now_ms,
                ),
            )

            # Extract simple memories from the transcript (heuristic)
            # Split on lines and look for meaningful content
            lines = [l.strip() for l in transcript.split("\n") if l.strip()]
            memory_idx = 0
            for i, line in enumerate(lines):
                if len(line) > 30:  # Skip short lines
                    mem_uuid = str(uuid_mod.uuid4())
                    conn.execute(
                        """
                        INSERT INTO imported_session_memories
                        (uuid, imported_session_uuid, memory_index, kind, content, tags, source, metadata, created_at)
                        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                    """,
                        (
                            mem_uuid,
                            uuid_str,
                            memory_idx,
                            "memory",
                            line[:500],
                            '["heuristic","archive-restore"]',
                            "heuristic",
                            json.dumps(
                                {
                                    "extraction": "heuristic",
                                    "sourceTool": source_tool,
                                    "path": str(tf),
                                }
                            ),
                            now_ms,
                        ),
                    )
                    memory_idx += 1

            ingested += 1
            if ingested % 20 == 0:
                print(
                    f"  Ingested {ingested} sessions ({memory_idx} memories from latest)"
                )
                conn.commit()

        except Exception as e:
            print(f"  Error inserting {stem}: {e}")
            continue

    conn.commit()
    print(f"\nIngested {ingested} new sessions from archive")

    # Final verification
    cur = conn.cursor()
    cur.execute("SELECT COUNT(*) FROM imported_sessions")
    total_sessions = cur.fetchone()[0]
    cur.execute("SELECT COUNT(*) FROM imported_session_memories")
    total_memories = cur.fetchone()[0]
    print(f"Total sessions: {total_sessions}")
    print(f"Total memories: {total_memories}")

    conn.close()
    return ingested


if __name__ == "__main__":
    print("=" * 60)
    print("Phase 2: Re-ingesting Archived Session Transcripts")
    print("=" * 60)
    count = ingest_archived_sessions()
    print(f"\nPhase 2 complete: {count} sessions re-ingested")
