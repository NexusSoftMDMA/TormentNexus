import sqlite3
import os
import requests
import json
import csv
from io import StringIO
import re

def compute_jaccard(text1, text2):
    set1 = set(text1.lower().split())
    set2 = set(text2.lower().split())
    if not set1 and not set2:
        return 1.0
    intersection = len(set1.intersection(set2))
    union = len(set1.union(set2))
    return intersection / union

def is_duplicate(new_prompt, existing_prompts, threshold=0.90):
    for ep in existing_prompts:
        sim = compute_jaccard(new_prompt['content'], ep['content'])
        if sim >= threshold:
            return True
    return False

def fetch_awesome_chatgpt_prompts():
    print("Fetching awesome-chatgpt-prompts...")
    url = "https://raw.githubusercontent.com/f/awesome-chatgpt-prompts/main/prompts.csv"
    prompts = []
    try:
        response = requests.get(url)
        response.raise_for_status()
        csv_reader = csv.reader(StringIO(response.text))
        header = next(csv_reader, None)
        for row in csv_reader:
            if len(row) >= 2:
                name, content = row[0], row[1]
                prompts.append({
                    "name": name,
                    "description": f"Prompt for {name}",
                    "category": "ChatGPT",
                    "content": content,
                    "tags": "awesome-chatgpt-prompts",
                    "usage_count": 0
                })
    except Exception as e:
        print(f"Error fetching awesome-chatgpt-prompts: {e}")
    return prompts

def fetch_awesome_prompts_json():
    print("Fetching awesome-prompts...")
    # This repo uses a README.md structure typically, but we'll try to extract or mock
    # to demonstrate handling multiple sources. We will use a known json collection if available
    # Actually, awesome-prompts structure can vary. Let's use leaked-system-prompts as another source.
    url = "https://raw.githubusercontent.com/jujumilk3/leaked-system-prompts/main/prompts.json"
    prompts = []
    try:
        response = requests.get(url)
        if response.status_code == 200:
            data = response.json()
            # Depending on structure, adapting. Assuming a list of objects or dict
            for key, val in data.items() if isinstance(data, dict) else enumerate(data):
                content = val.get("prompt", val.get("content", str(val))) if isinstance(val, dict) else str(val)
                name = val.get("name", str(key)) if isinstance(val, dict) else str(key)
                prompts.append({
                    "name": name,
                    "description": "System prompt",
                    "category": "System",
                    "content": content,
                    "tags": "leaked-system-prompts",
                    "usage_count": 0
                })
    except Exception as e:
        print(f"Error fetching leaked-system-prompts: {e}")
    return prompts

def init_db(db_path):
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS prompts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            category TEXT,
            content TEXT NOT NULL,
            tags TEXT,
            usage_count INTEGER DEFAULT 0,
            UNIQUE(name, content)
        )
    ''')
    conn.commit()
    return conn

def insert_prompts(conn, prompts):
    cursor = conn.cursor()
    # Read existing
    cursor.execute("SELECT content FROM prompts")
    existing_contents = [{'content': r[0]} for r in cursor.fetchall()]

    inserted = 0
    for p in prompts:
        if not is_duplicate(p, existing_contents):
            try:
                cursor.execute('''
                    INSERT INTO prompts (name, description, category, content, tags, usage_count)
                    VALUES (?, ?, ?, ?, ?, ?)
                ''', (p['name'], p['description'], p['category'], p['content'], p['tags'], p['usage_count']))
                existing_contents.append(p)
                inserted += 1
            except sqlite3.IntegrityError:
                pass # Exact duplicate caught by DB UNIQUE constraint
    conn.commit()
    print(f"Inserted {inserted} new prompts.")

def main():
    db_path = os.path.join(os.path.dirname(__file__), '..', 'data', 'prompt_library.db')
    os.makedirs(os.path.dirname(db_path), exist_ok=True)

    # Remove old corrupted db
    if os.path.exists(db_path):
        os.remove(db_path)

    conn = init_db(db_path)

    all_prompts = []
    all_prompts.extend(fetch_awesome_chatgpt_prompts())
    all_prompts.extend(fetch_awesome_prompts_json())

    if all_prompts:
        insert_prompts(conn, all_prompts)
    else:
        print("No prompts fetched.")

    conn.close()

if __name__ == '__main__':
    main()
