import sqlite3
conn = sqlite3.connect('tormentnexus.db')
for table in ['imported_sessions', 'imported_session_memories']:
    print(table + ":")
    for r in conn.execute(f"PRAGMA table_info({table})"):
        print("  ", r)
conn.close()
