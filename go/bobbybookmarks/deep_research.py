import os
import re
import json
import sqlite3
import requests
import time
import logging
from datetime import datetime, timezone
from bs4 import BeautifulSoup, Comment
import hashlib
from urllib.parse import urlparse, urlunparse

from llm_pool import LLMPool, stringify_field
from deduplicator import normalize_url

from borg_memory import TieredMemory
from borg_selfhealing import SelfHealingEngine, ExtractionValidator
from borg_skills import SkillEvolutionEngine
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
FLIGHT_LOG_DIR = os.path.join('logs', 'flight_recorder')
os.makedirs(FLIGHT_LOG_DIR, exist_ok=True)

llm_pool = LLMPool(logger=logger)
LLM_MODELS = [f"{b}/{m}" for b, m in llm_pool.all_backends]

# Phase 2 systems initialized inside main() — see below


BORG_TAXONOMY = [
    "Agent Orchestration & Workflow",
    "Context Engineering & Isolation",
    "Memory & Persistence Architecture",
    "Interface & Developer UX",
    "Connectivity & Interoperability (MCP/A2A)",
    "Infrastructure & Proxy Layers",
    "Guides & Industry Trends"
]

# =====================================================================
# UPGRADE 1: GARBAGE FILTER - Rejects known boilerplate patterns
# =====================================================================
GARBAGE_PATTERNS = [
    "automated discovery, heuristic mapping, other integration",
    "automated discovery, heuristic mapping, development tools",
    "automated discovery, heuristic mapping, ai agents",
    "automated discovery, heuristic mapping, search & discovery",
    "automated discovery (heuristic)",
    "automated discovery, heuristic mapping, guides",
    "automated discovery, heuristic mapping, mcp integration",
    "automated discovery, heuristic mapping, infrastructure",
    "automated discovery, heuristic mapping, fonts",
    "automated discovery, heuristic mapping, awesome lists",
]

GENERIC_DESCRIPTION_PATTERNS = [
    "a comprehensive resource detailing",
    "a powerful ai-powered",
    "sign in to continue",
]

SPAM_DOMAINS = [
    "temu.com", "pharmacy2home.com", "inawera.com",
    "vaporesso.com", "smoktech.com",
]


def is_garbage_extraction(rdata, url):
    """Returns (is_garbage, reason). If True, reject the extraction."""
    features = stringify_field(rdata.get('MAIN_FEATURES', '')).lower().strip()
    desc = stringify_field(rdata.get('SHORT_DESCRIPTION', '')).lower().strip()
    for pattern in GARBAGE_PATTERNS:
        if pattern in features:
            return True, "boilerplate:" + pattern[:50]
    for pattern in GENERIC_DESCRIPTION_PATTERNS:
        if pattern in desc:
            return True, "generic_desc:" + pattern[:40]
    parsed = urlparse(url)
    domain = (parsed.hostname or '').lower()
    for spam_domain in SPAM_DOMAINS:
        if spam_domain in domain:
            return True, "spam:" + spam_domain
    if not features or len(features) < 10:
        return True, "empty_features"
    if not desc or len(desc) < 15:
        return True, "empty_description"
    if features.startswith("unknown") or "unable to determine" in features:
        return True, "unknown_features"
    innovation = rdata.get('INNOVATION_SCORE', 0)
    try:
        innovation = int(innovation)
    except (ValueError, TypeError):
        innovation = 0
    if innovation <= 2 and len(features) < 50:
        return True, "low_innov_generic:score=" + str(innovation)
    return False, None


# =====================================================================
# UPGRADE 2: FIT MARKDOWN FILTERING
# =====================================================================
def extract_fit_markdown(html_content, url=""):
    """Extract meaningful content from HTML, filtering out noise."""
    soup = BeautifulSoup(html_content, 'html.parser')
    for tag in soup.find_all(['script', 'style', 'nav', 'footer', 'header',
                               'aside', 'iframe', 'noscript', 'svg', 'button',
                               'form', 'input', 'select', 'textarea']):
        tag.decompose()
    for comment in soup.find_all(string=lambda text: isinstance(text, Comment)):
        comment.extract()
    noise_re = re.compile(
        r'nav|footer|header|sidebar|comment|social|share|cookie|banner|'
        r'advertisement|ad-|popup|modal|overlay|tooltip|notification|'
        r'related|recommend|suggested|newsletter|subscribe|signup|login',
        re.IGNORECASE)
    for tag in soup.find_all(True, attrs={'class': noise_re}):
        tag.decompose()
    for tag in soup.find_all(True, attrs={'id': noise_re}):
        tag.decompose()
    main_content = None
    for selector in ['main', 'article', '[role="main"]', '.content', '#content',
                      '.post-body', '.article-body', '.markdown-body',
                      '.readme', '#readme', '.entry-content']:
        main_content = soup.select_one(selector)
        if main_content:
            break
    if not main_content:
        main_content = soup.find('body') or soup
    text = main_content.get_text(separator='\n', strip=True)
    lines = [line.strip() for line in text.splitlines() if line.strip()]
    text = '\n'.join(lines)
    max_chars = 8000
    if len(text) > max_chars:
        text = text[:max_chars] + "\n[...content truncated...]"
    return text


def extract_github_metadata(url, html_content):
    """For GitHub URLs, extract structured metadata from the HTML."""
    metadata = {}
    soup = BeautifulSoup(html_content, 'html.parser')
    about = soup.find('p', class_='f4 my-3')
    if about:
        metadata['repo_description'] = about.get_text(strip=True)
    topics = soup.find_all('a', class_='topic-tag')
    if topics:
        metadata['repo_topics'] = [t.get_text(strip=True) for t in topics]
    lang_item = soup.find('span', class_='color-fg-default text-bold mr-1')
    if lang_item:
        metadata['primary_language'] = lang_item.get_text(strip=True)
    stars = soup.find('span', id='repo-stars-counter-star')
    if stars:
        metadata['stars'] = stars.get_text(strip=True)
    forks = soup.find('span', id='repo-network-counter')
    if forks:
        metadata['forks'] = forks.get_text(strip=True)
    readme = soup.find('div', class_='markdown-body')
    if readme:
        metadata['readme_excerpt'] = readme.get_text(separator='\n', strip=True)[:3000]
    return metadata


# =====================================================================
# UPGRADE 3: FLIGHT RECORDER
# =====================================================================
def write_flight_receipt(url, rdata, model_used, raw_response, decision, reason=None):
    """Write a flight recorder receipt for an extraction attempt."""
    receipt = {
        "timestamp": iso_now(),
        "url": url,
        "url_hash": hashlib.md5(url.encode()).hexdigest()[:12],
        "model": model_used,
        "decision": decision,
        "reason": reason,
        "raw_response_preview": raw_response[:500] if raw_response else None,
    }
    if rdata:
        receipt["extracted"] = {
            "category": stringify_field(rdata.get('CATEGORY', ''))[:100],
            "innovation_score": rdata.get('INNOVATION_SCORE'),
            "features_preview": stringify_field(rdata.get('MAIN_FEATURES', ''))[:200],
        }
    date_str = datetime.now(timezone.utc).strftime('%Y-%m-%d')
    log_path = os.path.join(FLIGHT_LOG_DIR, 'flight_' + date_str + '.jsonl')
    try:
        with open(log_path, 'a', encoding='utf-8') as f:
            f.write(json.dumps(receipt, ensure_ascii=False) + '\n')
    except Exception as e:
        logger.warning("Flight recorder write failed: %s", e)


# =====================================================================
# UPGRADE 4: TIERED ROUTING
# =====================================================================
SIMPLE_DOMAINS = {
    'reddit.com', 'old.reddit.com', 'www.reddit.com',
    'news.ycombinator.com', 'youtube.com', 'youtu.be',
    'x.com', 'twitter.com',
}
COMPLEX_DOMAINS = {
    'github.com', 'gitlab.com', 'bitbucket.org',
    'arxiv.org', 'paperswithcode.com', 'huggingface.co',
    'docs.python.org', 'docs.rs', 'pkg.go.dev',
    'medium.com', 'substack.com',
    'blog.cloudflare.com', 'openai.com', 'anthropic.com',
}


def classify_url_complexity(url):
    """Returns 'simple', 'medium', or 'complex'."""
    try:
        parsed = urlparse(url)
        domain = (parsed.hostname or '').lower().replace('www.', '')
    except Exception:
        return 'medium'
    if any(d in domain for d in SIMPLE_DOMAINS):
        return 'simple'
    if any(d in domain for d in COMPLEX_DOMAINS):
        return 'complex'
    path = parsed.path.lower()
    if '/wiki/' in path or '/docs/' in path or '/api/' in path:
        return 'complex'
    return 'medium'


def build_tiered_prompt(url, fit_text, complexity, github_metadata=None):
    """Build an extraction prompt appropriate to the URL complexity."""
    prompt = "Analyze this technical resource for the Borg Intelligence database.\n\n"
    prompt += "URL: " + url + "\n"
    if github_metadata:
        if 'repo_description' in github_metadata:
            prompt += "\nRepo: " + github_metadata['repo_description'] + "\n"
        if 'repo_topics' in github_metadata:
            prompt += "Topics: " + ", ".join(github_metadata['repo_topics']) + "\n"
        if 'primary_language' in github_metadata:
            prompt += "Language: " + github_metadata['primary_language'] + "\n"
    prompt += "\nContent: " + fit_text + "\n\n"
    prompt += "Categorize into EXACTLY ONE: " + ", ".join(BORG_TAXONOMY) + "\n\n"
    prompt += "Return strict JSON:\n"
    prompt += "- CATEGORY: one of the above categories\n"
    prompt += "- SHORT_DESCRIPTION: 1 specific sentence about what this DOES\n"
    prompt += "- LONG_DESCRIPTION: detailed technical breakdown\n"
    prompt += "- MAIN_FEATURES: 3-5 SPECIFIC concrete features (comma separated)\n"
    prompt += "- INNOVATION_SCORE: 1-10 uniqueness rating\n"
    prompt += "- TAGS: 8-12 lowercase technical tags\n\n"
    prompt += "CRITICAL: MAIN_FEATURES must be SPECIFIC capabilities, NOT 'automated discovery' or 'heuristic mapping'.\n"
    return prompt


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
    """Upgraded v2.0: Fit Markdown + GitHub metadata + Tiered prompt + Garbage filter."""
    fit_text = extract_fit_markdown(content, url)
    github_metadata = None
    if 'github.com' in url:
        github_metadata = extract_github_metadata(url, content)
    complexity = classify_url_complexity(url)
    # Phase 2: Skill-enhanced extraction
    skill_engine = getattr(borg_research_url, '_skill_engine', None)
    if skill_engine:
        skill_name, skill_config = skill_engine.match_skill(url, content)
        prompt = skill_engine.build_skill_prompt(url, skill_name, fit_text, skill_config)
        status['_skill'] = skill_name
    else:
        prompt = build_tiered_prompt(url, fit_text, complexity, github_metadata)

    raw_response = None
    model_used = None
    max_retries = 3

    for attempt in range(max_retries):
        try:
            status.update({
                'state': 'researching',
                'active_url': url,
                'sleep_seconds': None,
                'last_error': None,
                'attempt': attempt + 1,
                'url_complexity': complexity,
            })
            write_status(status)
            res_text, model_used = llm_pool.generate_content(prompt, "researching " + url)
            raw_response = res_text
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
            res_text = res_text.strip()
            if "```json" in res_text:
                res_text = res_text.split("```json")[1].split("```")[0].strip()
            elif "```" in res_text:
                res_text = res_text.split("```")[1].split("```")[0].strip()
            json_match = re.search(r'\{[^{}]*\}', res_text, re.DOTALL)
            if json_match:
                res_text = json_match.group(0)
            rdata = json.loads(res_text)

            # Phase 2: Self-healing validation
            healing_engine = getattr(borg_research_url, "_healing_engine", None)
            if healing_engine:
                final_rdata, quality, decision_path = healing_engine.process_extraction(
                    url, rdata, content, raw_response)
                if final_rdata is None:
                    logger.info("Self-healing REJECTED %%s (path: %%s)", url, " -> ".join(decision_path))
                    write_flight_receipt(url, rdata, model_used, raw_response, "rejected", "selfhealing:" + str(decision_path))
                    return None
                rdata = final_rdata
                status["_quality"] = round(quality, 3)
                status["_decision_path"] = decision_path


            # UPGRADE 1: Garbage filter
            is_garbage, garbage_reason = is_garbage_extraction(rdata, url)
            if is_garbage:
                logger.warning("Garbage REJECTED %s: %s", url, garbage_reason)
                write_flight_receipt(url, rdata, model_used, raw_response, 'rejected', garbage_reason)
                write_feed("Rejected: " + url + " - " + garbage_reason, "reject")
                status.update({'state': 'garbage_rejected', 'active_url': url,
                               'last_error': 'garbage:' + garbage_reason})
                write_status(status)
                return None

            # UPGRADE 3: Flight receipt for accepted extraction
            write_flight_receipt(url, rdata, model_used, raw_response, 'accepted')
            return rdata

        except json.JSONDecodeError as e:
            logger.warning("JSON decode error (attempt %d/%d) for %s: %s", attempt+1, max_retries, url, e)
            if attempt == max_retries - 1:
                logger.error("Failed to decode LLM response for %s after %d attempts", url, max_retries)
                write_flight_receipt(url, None, model_used, raw_response, 'failed', 'decode_error:' + str(e))
                status.update({'state': 'decode_error', 'active_url': url, 'last_error': str(e)})
                write_status(status)
                return None
            time.sleep(5)
        except Exception as e:
            logger.error("Unexpected error researching %s: %s", url, e)
            write_flight_receipt(url, None, model_used, raw_response, 'failed', 'exception:' + str(e))
            status.update({'state': 'error', 'active_url': url, 'last_error': str(e)})
            write_status(status)
            return None
    return None


def main():
    logger.info("=" * 60)
    logger.info("Borg Intelligence Deep Research Worker v2.0")
    logger.info("Upgrades: Garbage Filter | Fit Markdown | Flight Recorder | Tiered Routing | Self-Healing | Skills")
    logger.info("=" * 60)
    logger.info("Using LLM backends: %s", ', '.join(LLM_MODELS))

    # Phase 2: Initialize intelligence systems
    borg_memory = TieredMemory()
    borg_healing = SelfHealingEngine(llm_pool=llm_pool, memory=borg_memory)
    borg_skills_engine = SkillEvolutionEngine(memory=borg_memory)
    borg_research_url._healing_engine = borg_healing
    borg_research_url._skill_engine = borg_skills_engine
    borg_research_url._memory = borg_memory
    logger.info("Phase 2 systems: Memory | Self-Healing | Skills initialized")
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
        'version': '2.0',
        'stats': {'accepted': 0, 'rejected': 0, 'fetch_failed': 0, 'decode_failed': 0},
    }
    write_status(status)
    complexity_counts = {'simple': 0, 'medium': 0, 'complex': 0}
    for u in urls:
        complexity_counts[classify_url_complexity(u)] += 1
    logger.info('Borg Intelligence Phase: %d links remaining.', len(urls))
    logger.info('  Complexity: %s', complexity_counts)
    
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
                logger.info("Borg Intelligence Extracted: %s", url)
                status['stats']['accepted'] += 1
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
        
        else:
            if status.get('state') == 'garbage_rejected':
                status['stats']['rejected'] += 1
            else:
                status['stats']['decode_failed'] += 1

        time.sleep(5)  # Faster with local LLM - no API rate limits

    # Final summary
    logger.info('=' * 60)
    logger.info('Borg Intelligence Run Complete')
    logger.info('  Accepted: %d', status['stats']['accepted'])
    logger.info('  Rejected (garbage): %d', status['stats']['rejected'])
    logger.info('  Fetch failed: %d', status['stats']['fetch_failed'])
    logger.info('  Decode failed: %d', status['stats']['decode_failed'])
    logger.info('=' * 60)

    # Phase 2: Final memory and stats dump
    mem = getattr(borg_research_url, '_memory', None)
    heal = getattr(borg_research_url, '_healing_engine', None)
    if mem:
        mem.flush_l1_to_l2()
        mem_stats = mem.get_memory_stats()
        logger.info("Memory: L1=%d L2=%d L3=%d Skills=%d Tools=%d",
                    mem_stats['l1_count'], mem_stats['l2_count'],
                    mem_stats['l3_count'], mem_stats['skills_count'],
                    mem_stats['tools_count'])
    if heal:
        hs = heal.get_stats()
        logger.info("Self-Healing: validated=%d corrected=%d cross_validated=%d rejected=%d avg_quality=%.2f",
                    hs['validated_first_pass'], hs['corrected'],
                    hs['cross_validated'], hs['rejected_after_all'], hs['avg_quality'])

if __name__ == "__main__":
    main()
