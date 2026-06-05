import sqlite3

conn = sqlite3.connect('catalog.db')
cursor = conn.cursor()

# Get verified servers with database keywords
cursor.execute("""
    SELECT display_name, repository_url, description 
    FROM published_mcp_servers 
    WHERE status = 'verified' 
      AND (display_name LIKE '%sqlite%' 
       OR display_name LIKE '%postgres%' 
       OR display_name LIKE '%mysql%'
       OR display_name LIKE '%redis%'
       OR display_name LIKE '%database%'
       OR description LIKE '%database%'
       OR description LIKE '%postgresql%')
""")

with open('scratch/db_repos_list.txt', 'w', encoding='utf-8') as f:
    for row in cursor.fetchall():
        f.write(f"Name: {row[0]} | Repo: {row[1]}\n")

conn.close()
print("Wrote database repos list!")
