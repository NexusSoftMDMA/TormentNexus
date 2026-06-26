"""
Session Import — Fix Data Format
==================================
Reads candidate session files and wraps them in the correct
TormentNexus ExportPackage format for the TS tRPC import endpoint.

Expected format:
  {"version":"1.0","sessions":[{"id","name","cliType","createdAt","workingDirectory","metadata","logs","memories"},...]}

Usage:
    python scripts/import_sessions.py
    python scripts/import_sessions.py --dry-run
"""

import json
import os
import sys
import urllib.request
import urllib.error

GO_API = "http://localhost:4300"
CANDIDATE_FILE = os.path.expanduser("~/.claude/history.jsonl")


def build_export_package(raw_jsonl: str) -> dict:
    """Convert a JSONL file into TormentNexus ExportPackage format."""
    sessions = []
    for i, line in enumerate(raw_jsonl.strip().split("\n")):
        line = line.strip()
        if not line:
            continue
        try:
            entry = json.loads(line)
        except json.JSONDecodeError:
            continue

        session = {
            "id": entry.get("uuid", entry.get("id", f"claude_{i}")),
            "name": entry.get("title", entry.get("summary", f"Session {i}")),
            "cliType": "claude-code",
            "status": "stopped",
            "createdAt": entry.get("created_at", entry.get("startTime", 0)),
            "workingDirectory": os.getcwd(),
            "metadata": {
                k: v for k, v in entry.items() if k not in ("messages", "transcript")
            },
            "logs": [],
            "memories": [],
        }

        # Include transcript in metadata
        transcript = entry.get("transcript", "") or ""
        messages = entry.get("messages", [])
        if isinstance(messages, list) and len(messages) > 0:
            transcript = json.dumps(messages)
        if transcript:
            session["metadata"]["transcriptSnippet"] = transcript[:5000]
            session["logs"] = [
                {"timestamp": 0, "level": "info", "message": transcript[:50000]}
            ]

        sessions.append(session)

    return {
        "version": "1.0",
        "exportedAt": int(__import__("time").time() * 1000),
        "tormentnexusVersion": "0.90.5",
        "environment": "windows",
        "sessionCount": len(sessions),
        "sessions": sessions,
        "globalMemories": [],
    }


def import_export(pkg: dict, dry_run: bool = False):
    """Send properly formatted export package to the import endpoint."""
    payload = json.dumps(
        {
            "data": json.dumps(pkg),
            "merge": True,
            "dryRun": dry_run,
        }
    ).encode()

    req = urllib.request.Request(
        f"{GO_API}/api/session-export/import",
        data=payload,
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as r:
            return json.loads(r.read())
    except urllib.error.HTTPError as e:
        return {"error": e.read().decode()[:300]}
    except Exception as e:
        return {"error": str(e)}


def main():
    dry_run = "--dry-run" in sys.argv
    print(f"Session Import — Dry Run: {dry_run}\n")

    if not os.path.exists(CANDIDATE_FILE):
        print(f"Candidate file not found: {CANDIDATE_FILE}")
        return

    with open(CANDIDATE_FILE, "r", encoding="utf-8", errors="replace") as f:
        raw = f.read()

    print(f"Reading: {CANDIDATE_FILE} ({len(raw):,} bytes)")

    pkg = build_export_package(raw)
    print(f"Built export package: {pkg['sessionCount']} sessions\n")

    total_imported = 0
    total_merged = 0
    total_skipped = 0
    total_errors = []
    
    chunk_size = 10
    sessions = pkg["sessions"]
    num_chunks = (len(sessions) - 1) // chunk_size + 1
    
    for idx in range(0, len(sessions), chunk_size):
        chunk = sessions[idx:idx + chunk_size]
        sub_pkg = {
            "version": pkg["version"],
            "exportedAt": pkg["exportedAt"],
            "tormentnexusVersion": pkg["tormentnexusVersion"],
            "environment": pkg["environment"],
            "sessionCount": len(chunk),
            "sessions": chunk,
            "globalMemories": pkg["globalMemories"],
        }
        print(f"Importing chunk {idx // chunk_size + 1}/{num_chunks} (sessions {idx} to {idx + len(chunk) - 1})...")
        result = import_export(sub_pkg, dry_run=dry_run)
        
        if "error" in result:
            print(f"  Chunk Error: {result['error'][:200]}")
            total_errors.append(f"Chunk starting at {idx} failed: {result['error']}")
            continue
            
        report = result.get("data", {})
        total_imported += report.get('imported', 0)
        total_merged += report.get('merged', 0)
        total_skipped += report.get('skipped', 0)
        total_errors.extend(report.get('errors', []))

    print("\nFinal Result:")
    print(f"  Imported: {total_imported}")
    print(f"  Merged:  {total_merged}")
    print(f"  Skipped: {total_skipped}")
    print(f"  Errors:  {len(total_errors)}")
    for e in total_errors[:10]:
        print(f"    - {e[:150]}")


if __name__ == "__main__":
    main()
