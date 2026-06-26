import os
import re
import json
import sqlite3
import requests
import time
import logging
from datetime import datetime, timezone
from bs4 import BeautifulSoup
from urllib.parse import urlparse, urlunparse

from llm_pool import LLMPool, stringify_field
from deduplicator import normalize_url
logging.basicConfig(
    level=logging.INFO, 
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(os.path.join('logs', 'borg_research.log'), mode='a'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# File paths
BOOKMARKS_FILE = 'bookmarks.txt'
DB_PATH = 'bookmarks.db'
STATUS_PATH = 'deep_research_status.json'

llm_pool = LLMPool(logger=logger)
LLM_MODELS = [f"{b}/{m}" for b, m in llm_pool.all_backends]

BORG_TAXONOMY = [
    "Agent Orchestration & Workflow",
    "Context Engineering & Isolation",
    "Memory & Persistence Architecture",
    "Interface & Developer UX",
    "Connectivity & Interoperability (MCP/A2A)",
    "Infrastructure & Proxy Layers",
    "Guides & Industry Trends"
]

def init_db():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS bookmarks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            url TEXT UNIQUE,
            category TEXT,
            short_description TEXT,
            long_description TEXT,
            tags TEXT,
            main_features TEXT,
            research_level TEXT,
            innovation_score INTEGER,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    try:
        cursor.execute("ALTER TABLE bookmarks ADD COLUMN research_level TEXT DEFAULT 'heuristic'")
        cursor.execute("ALTER TABLE bookmarks ADD COLUMN innovation_score INTEGER DEFAULT 0")
    except sqlite3.OperationalError: pass 
    conn.commit()
    return conn

def fetch_content(url):
    headers = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'}
    try:
        resp = requests.get(url, headers=headers, timeout=10, allow_redirects=True)
        if resp.status_code == 200: return resp.text
    except Exception: pass
    return None

def iso_now():
    return datetime.now(timezone.utc).isoformat()

def write_status(status):
    payload = dict(status)
    payload['updated_at'] = iso_now()
    # Robust write: try atomic replace, fall back to direct write
    temp_path = f"{STATUS_PATH}.tmp"
    try:
        with open(temp_path, 'w', encoding='utf-8') as f:
            json.dump(payload, f, indent=2, sort_keys=True)
        try:
            os.replace(temp_path, STATUS_PATH)
        except (PermissionError, OSError):
            # Windows sometimes locks the file; fall back to direct write
            with open(STATUS_PATH, 'w', encoding='utf-8') as f:
                json.dump(payload, f, indent=2, sort_keys=True)
            try:
                os.unlink(temp_path)
            except OSError:
                pass
    except (PermissionError, OSError):
        # Last resort: direct write
        try:
            with open(STATUS_PATH, 'w', encoding='utf-8') as f:
                json.dump(payload, f, indent=2, sort_keys=True)
        except OSError:
            pass  # Non-critical; keep processing

def write_feed(message, type="info"):
    feed_path = os.path.join('logs', 'live_feed.json')
    entry = {
        "timestamp": iso_now(),
        "type": type,
        "message": message
    }
    try:
        # Keep only the last 100 entries for efficiency
        entries = []
        if os.path.exists(feed_path):
            with open(feed_path, 'r', encoding='utf-8') as f:
                entries = json.load(f)
        entries.append(entry)
        with open(feed_path, 'w', encoding='utf-8') as f:
            json.dump(entries[-100:], f, indent=2)
    except Exception: pass

def borg_research_url(url, content, status):
    soup = BeautifulSoup(content, 'html.parser')
    for s in soup(['script', 'style']): s.decompose()
    text = re.sub(r'\s+', ' ', soup.get_text())[:10000] 
    
    prompt = f"""
    Analyze the following technical resource for inclusion in the 'Borg' Project intelligence database.
    URL: {url}
    Content: {text}
    
    Categorize this into EXACTLY ONE of these Borg Categories: {', '.join(BORG_TAXONOMY)}.
    
    Return a strict JSON object:
    - CATEGORY: The chosen category.
    - SHORT_DESCRIPTION: 1 sentence.
    - LONG_DESCRIPTION: Detailed breakdown of the technical approach.
    - MAIN_FEATURES: List of features that Borg should consider implementing (comma separated).
    - INNOVATION_SCORE: 1-10 rating of how unique this project's approach is.
    - TAGS: 8-12 technical tags (lowercase).
    """
    
    max_retries = 3
    for attempt in range(max_retries):
        try:
            status.update({
                'state': 'researching',
                'active_url': url,
                'sleep_seconds': None,
                'last_error': None,
                'attempt': attempt + 1,
            })
            write_status(status)
            res_text, model_used = llm_pool.generate_content(prompt, f"researching {url}")
            if res_text is None:
                status.update({
                    'state': 'backing_off',
                    'active_url': url,
                    'sleep_seconds': llm_pool.last_backoff_seconds,
                    'last_error': llm_pool.last_error_summary,
                    'last_model': llm_pool.active_model_name,
                })
                write_status(status)
                continue
            # Clean response - strip markdown code fences
            res_text = res_text.strip()
            if "```json" in res_text:
                res_text = res_text.split("```json")[1].split("```")[0].strip()
            elif "```" in res_text:
                res_text = res_text.split("```")[1].split("```")[0].strip()
            # Try to find JSON object in response
            json_match = re.search(r'\{[^{}]*\}', res_text, re.DOTALL)
            if json_match:
                res_text = json_match.group(0)
            return json.loads(res_text)
        except json.JSONDecodeError as e:
            logger.warning(f"JSON decode error (attempt {attempt+1}/{max_retries}) for {url}: {e}")
            if attempt == max_retries - 1:
                logger.error(f"Failed to decode LLM response for {url} after {max_retries} attempts")
                status.update({
                    'state': 'decode_error',
                    'active_url': url,
                    'last_error': str(e),
                })
                write_status(status)
                return None
            time.sleep(5)
        except Exception as e:
            logger.error(f"Unexpected error researching {url}: {e}")
            status.update({
                'state': 'error',
                'active_url': url,
                'last_error': str(e),
            })
            write_status(status)
            return None
    return None

def main():
    logger.info(f"Using LLM backends: {', '.join(LLM_MODELS)}")
    conn = init_db()
    cursor = conn.cursor()
    cursor.execute("SELECT url FROM bookmarks WHERE research_level = 'borg'")
    processed = {normalize_url(row[0]) for row in cursor.fetchall()}
    cursor.execute("SELECT COUNT(*) FROM bookmarks WHERE research_level = 'borg'")
    existing_borg_rows = cursor.fetchone()[0]
    
    urls = []
    with open(BOOKMARKS_FILE, 'r', encoding='utf-8', errors='ignore') as f:
        for line in f:
            u = line.strip()
            if u.startswith('http') and normalize_url(u) not in processed:
                urls.append(u)

    status = {
        'worker_pid': os.getpid(),
        'models': LLM_MODELS,
        'backend': 'lmstudio -> openrouter/free',
        'state': 'starting',
        'active_url': None,
        'last_extracted_url': None,
        'last_error': None,
        'sleep_seconds': None,
        'remaining_urls': len(urls),
        'borg_rows': existing_borg_rows,
    }
    write_status(status)
    logger.info(f"Borg Intelligence Phase: {len(urls)} links remaining.")
    
    for index, url in enumerate(urls):
        status.update({
            'worker_pid': os.getpid(),
            'state': 'fetching',
            'active_url': url,
            'sleep_seconds': None,
            'remaining_urls': len(urls) - index,
        })
        write_status(status)
        write_feed(f"Fetching content for: {url}", "fetch")
        content = fetch_content(url)
        if not content:
            write_feed(f"Fetch failed for: {url}", "error")
            status.update({
                'state': 'fetch_failed',
                'active_url': url,
                'last_error': 'fetch_failed',
            })
            write_status(status)
            continue
        
        write_feed(f"Starting Borg research on: {url}", "research")
        rdata = borg_research_url(url, content, status)
        if rdata:
            write_feed(f"Finalizing extraction for: {url}", "process")
            from worker_wrapper import pulse
            pulse("Borg Research Worker", f"Assimilating: {url}", {"remaining": len(urls) - index, "borg_rows": status.get('borg_rows')})
            try:
                cursor.execute('''
                    INSERT INTO bookmarks (url, category, short_description, long_description, tags, main_features, research_level, innovation_score)
                    VALUES (?, ?, ?, ?, ?, ?, 'borg', ?)
                    ON CONFLICT(url) DO UPDATE SET
                        category=excluded.category,
                        short_description=excluded.short_description,
                        long_description=excluded.long_description,
                        tags=excluded.tags,
                        main_features=excluded.main_features,
                        research_level='borg',
                        innovation_score=excluded.innovation_score
                ''', (
                    url,
                    stringify_field(rdata.get('CATEGORY')) or "Other",
                    stringify_field(rdata.get('SHORT_DESCRIPTION')),
                    stringify_field(rdata.get('LONG_DESCRIPTION')),
                    stringify_field(rdata.get('TAGS')),
                    stringify_field(rdata.get('MAIN_FEATURES')),
                    rdata.get('INNOVATION_SCORE', 0),
                ))
                conn.commit()
                logger.info(f"Borg Intelligence Extracted: {url}")
                status.update({
                    'state': 'processing',
                    'active_url': url,
                    'last_extracted_url': url,
                    'last_error': None,
                    'sleep_seconds': None,
                    'remaining_urls': len(urls) - index - 1,
                    'borg_rows': status.get('borg_rows', len(processed)) + 1,
                    'last_model': llm_pool.active_model_name,
                })
                write_status(status)
            except Exception as e:
                logger.error(f"DB Error: {e}")
                status.update({
                    'state': 'db_error',
                    'active_url': url,
                    'last_error': str(e),
                })
                write_status(status)
        
        time.sleep(5)  # Faster with local LLM - no API rate limits

if __name__ == "__main__":
    main()
