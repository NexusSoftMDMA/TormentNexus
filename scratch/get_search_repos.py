import sqlite3

conn = sqlite3.connect('catalog.db')
cursor = conn.cursor()

# Get verified servers with search, scrape, fetch, brave, or web keywords
cursor.execute("""
    SELECT display_name, repository_url, description 
    FROM published_mcp_servers 
    WHERE status = 'verified' 
      AND (display_name LIKE '%search%' 
       OR display_name LIKE '%scrape%' 
       OR display_name LIKE '%fetch%'
       OR display_name LIKE '%brave%'
       OR display_name LIKE '%tavily%'
       OR description LIKE '%search%'
       OR description LIKE '%scrape%'
       OR description LIKE '%fetch%')
""")

with open('scratch/search_repos_list.txt', 'w', encoding='utf-8') as f:
    for row in cursor.fetchall():
        f.write(f"Name: {row[0]} | Repo: {row[1]}\n")

conn.close()
print("Wrote search repos list!")
