import sqlite3
import os
import json
import time
import sys
from datetime import datetime

DB_PATH = 'bookmarks.db'

def pulse(agent_name, current_task, metadata={}):
    try:
        conn = sqlite3.connect(DB_PATH)
        cur = conn.cursor()
        cur.execute('''
            INSERT INTO agent_heartbeats (agent_name, last_pulse, current_task, status_metadata)
            VALUES (?, CURRENT_TIMESTAMP, ?, ?)
            ON CONFLICT(agent_name) DO UPDATE SET
                last_pulse=CURRENT_TIMESTAMP,
                current_task=excluded.current_task,
                status_metadata=excluded.status_metadata
        ''', (agent_name, current_task, json.dumps(metadata)))
        conn.commit()
        conn.close()
    except Exception as e:
        print(f"Heartbeat failed for {agent_name}: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python worker_wrapper.py <agent_name> <task_desc>")
        sys.exit(1)
    
    agent_name = sys.argv[1]
    task_desc = sys.argv[2]
    
    # Single pulse and exit (for integration into other loops)
    pulse(agent_name, task_desc, {"pid": os.getpid()})
