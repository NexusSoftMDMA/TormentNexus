#!/usr/bin/env python3
"""Borg Research Worker v3.0 - Deep Processing Pass

Focused on:
1. 787 new URLs from incoming_resources.txt not yet in atlas
2. 1,500 entries missing main_features (high-signal first)
3. 178 entries with short descriptions
4. Uses larger LM Studio models for better quality

Usage: python _research_worker_v3.py [--dry-run] [--limit N]
"""
import os, re, json, sqlite3, requests, time, logging, sys, argparse
from datetime import datetime
from bs4 import BeautifulSoup
from urllib.parse import urlparse

sys.stdout.reconfigure(encoding='utf-8')
os.makedirs('logs', exist_ok=True)

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(message)s',
    handlers=[
        logging.FileHandler('logs/research_v3.log', mode='a', encoding='utf-8'),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger(__name__)

ATLAS_DB = 'atlas.db'

LMSTUDIO_URL = "http://localhost:1234/v1/chat/completions"

# Models ordered by speed/quality tradeoff
# Use smaller models for speed, bigger for high-value entries
MODELS_FAST = [
    ("liquid/lfm2.5-1.2b", 60),
    ("gemma-4-e2b-uncensored-hauhaucs-aggressive", 90),
]
MODELS_QUALITY = [
    ("gemma-4-e2b-uncensored-hauhaucs-aggressive", 90),
    ("gemma-4-e4b-uncensored-hauhaucs-aggressive", 120),
    ("liquid/lfm2.5-1.2b", 60),
]

BORG_TAXONOMY = [
    "Agent Orchestration & Workflow",
    "Context Engineering & Isolation",
    "Memory & Persistence Architecture",
    "Interface & Developer UX",
    "Connectivity / MCP / A2A",
    "Infrastructure & Proxy Layers",
    "Guides & Industry Trends",
    "Coding Harness Tools",
    "AI Agents & Frameworks",
    "Search & Discovery",
    "Coding Tools & IDEs",
    "Developer Workflow & Tools",
    "Vector Databases & Embeddings",
    "Security & Red Teaming",
]

CAT_MAP = {
    'Connectivity & Interoperability (MCP/A2A)': 'Connectivity / MCP / A2A',
    'Connectivity & Interoperability': 'Connectivity / MCP / A2A',
    'Development Tools & Libraries': 'Coding Tools & IDEs',
    'Vector Databases & Search': 'Vector Databases & Embeddings',
    'Developer Workflow': 'Developer Workflow & Tools',
    'Guides & Articles': 'Guides & Industry Trends',
    'Infrastructure': 'Infrastructure & Proxy Layers',
    'Other': None,
    'Software Development': 'Developer Workflow & Tools',
    'Software/Developer Tools': 'Coding Tools & IDEs',
    'Software/Application Development Tool': 'Developer Workflow & Tools',
}

FIELD_NAMES = ['CATEGORY', 'SHORT_DESCRIPTION', 'LONG_DESCRIPTION', 'MAIN_FEATURES', 'INNOVATION_SCORE', 'TAGS']

# --- Stats ---
stats = {
    'accepted': 0, 'rejected': 0, 'failed': 0, 'skipped': 0,
    'metadata_only': 0, 'fetched': 0, 'new_ingested': 0,
    'features_enriched': 0, 'desc_enriched': 0,
    'llm_calls': 0, 'llm_failures': 0,
}

# --- Helpers ---

def stringify(v):
    if v is None:
        return ''
    if isinstance(v, str):
        return v
    if isinstance(v, (list, tuple)):
        return ', '.join(str(x) for x in v)
    return str(v)

def call_llm(prompt, prefer_big=False):
    """Call LM Studio with model fallback. If prefer_big, use quality models."""
    models = MODELS_QUALITY if prefer_big else MODELS_FAST
    
    for model, tout in models:
        try:
            stats['llm_calls'] += 1
            resp = requests.post(LMSTUDIO_URL, json={
                "model": model,
                "messages": [{"role": "user", "content": prompt}],
                "temperature": 0.2,
                "max_tokens": 400,
            }, timeout=tout)
            if resp.status_code == 200:
                data = resp.json()
                text = data.get('choices', [{}])[0].get('message', {}).get('content', '')
                if text and len(text) > 30:
                    return text, model
            else:
                logger.warning(f"  HTTP {resp.status_code} from {model}")
        except requests.exceptions.Timeout:
            logger.warning(f"  Timeout ({tout}s) on {model}")
        except Exception as e:
            logger.warning(f"  Error on {model}: {str(e)[:80]}")
        time.sleep(2)  # Let GPU cool between model switches
    
    stats['llm_failures'] += 1
    return None, None

def parse_llm_response(raw):
    if not raw:
        return None
    text = raw.strip()
    
    # 1. Direct JSON
    try:
        return json.loads(text)
    except (json.JSONDecodeError, ValueError):
        pass
    
    # 2. JSON in code blocks
    for delim in ["```json", "```"]:
        if delim in text:
            parts = text.split(delim)
            for part in parts[1:]:
                end = part.find("```")
                block = part[:end].strip() if end >= 0 else part.strip()
                try:
                    return json.loads(block)
                except (json.JSONDecodeError, ValueError):
                    continue
    
    # 3. Extract key-value pairs
    result = {}
    for field in FIELD_NAMES:
        # Try "FIELD": "value"
        patterns = [
            rf'"{field}"\s*:\s*"((?:[^"\\]|\\.)*)"',
            rf'"{field}"\s*:\s*(\d+)',
            rf'{field}\s*:\s*"((?:[^"\\]|\\.)*)"',
            rf'{field}\s*:\s*(\d+)',
        ]
        for pat in patterns:
            m = re.search(pat, text)
            if m:
                result[field] = m.group(1)
                break
    
    return result if len(result) >= 3 else None

def fetch_content(url, timeout=15):
    """Fetch and extract readable content from a URL"""
    try:
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
            'Accept': 'text/html,application/xhtml+xml',
        }
        resp = requests.get(url, headers=headers, timeout=timeout, allow_redirects=True)
        if resp.status_code != 200:
            return None
        return resp.text
    except Exception:
        return None

def extract_fit_markdown(html, url=""):
    """Extract main content from HTML, stripping nav/footer/noise"""
    if not html:
        return ""
    try:
        soup = BeautifulSoup(html, 'html.parser')
        # Remove noise
        for tag in soup.find_all(['script', 'style', 'nav', 'footer', 'header', 'aside']):
            tag.decompose()
        for div in soup.find_all(attrs={'class': re.compile(r'sidebar|nav|footer|header|comment', re.I)}):
            div.decompose()
        
        # Try <main> or <article> first
        main = soup.find('main') or soup.find('article') or soup.find(attrs={'role': 'main'})
        if main:
            text = main.get_text(separator='\n', strip=True)
        else:
            text = soup.get_text(separator='\n', strip=True)
        
        # Truncate to fit in prompt
        lines = [l.strip() for l in text.split('\n') if l.strip()]
        return '\n'.join(lines[:150])[:3000]
    except Exception:
        return ""

def extract_gh_meta(url, html):
    """Extract GitHub-specific metadata"""
    meta = {}
    m = re.match(r'https://github\.com/([^/]+)/([^/?#]+)', url)
    if m:
        meta['owner'] = m.group(1)
        meta['repo'] = m.group(2)
    
    if html:
        try:
            soup = BeautifulSoup(html, 'html.parser')
            # GitHub about/description
            about = soup.find(attrs={'class': re.compile(r'RepositoryDescription|about', re.I)})
            if about:
                meta['desc'] = about.get_text(strip=True)[:200]
            # Topics/tags
            topics = soup.find_all(attrs={'class': re.compile(r'topic-tag', re.I)})
            if topics:
                meta['topics'] = ', '.join(t.get_text(strip=True) for t in topics[:10])
            # Stars
            stars = soup.find(attrs={'class': re.compile(r'stargazers|star', re.I)})
            if stars:
                star_text = stars.get_text(strip=True)
                meta['stars'] = star_text
            # Language
            lang = soup.find(attrs={'class': re.compile(r'programming-language|language-color', re.I)})
            if lang:
                meta['language'] = lang.get_text(strip=True)
        except Exception:
            pass
    
    return meta if meta else None

def extract_reddit_context(url):
    """Extract context from Reddit URL structure"""
    m = re.search(r'reddit\.com/r/([^/]+)/comments/([^/]+)/([^/?]+)', url)
    if m:
        return {
            'subreddit': m.group(1),
            'post_slug': m.group(3).replace('_', ' '),
        }
    return None

def build_prompt(url, fit_text, gh_meta=None, reddit_ctx=None, existing_sd=None):
    """Build the LLM enrichment prompt"""
    context_parts = [f"URL: {url}"]
    
    if gh_meta:
        if 'owner' in gh_meta and 'repo' in gh_meta:
            context_parts.append(f"GitHub: {gh_meta['owner']}/{gh_meta['repo']}")
        if 'desc' in gh_meta:
            context_parts.append(f"About: {gh_meta['desc']}")
        if 'topics' in gh_meta:
            context_parts.append(f"Topics: {gh_meta['topics']}")
        if 'language' in gh_meta:
            context_parts.append(f"Language: {gh_meta['language']}")
    
    if reddit_ctx:
        if 'subreddit' in reddit_ctx:
            context_parts.append(f"Subreddit: r/{reddit_ctx['subreddit']}")
        if 'post_slug' in reddit_ctx:
            context_parts.append(f"Post: {reddit_ctx['post_slug']}")
    
    if existing_sd and len(existing_sd) > 5:
        context_parts.append(f"Existing description: {existing_sd}")
    
    if fit_text and len(fit_text) > 50:
        # Truncate page content for prompt
        context_parts.append(f"Page content:\n{fit_text[:2000]}")
    
    context = '\n'.join(context_parts)
    
    prompt = f"""Analyze this AI/developer tool or resource. Return a JSON object with these fields:

- CATEGORY: one of [{', '.join(BORG_TAXONOMY)}]
- SHORT_DESCRIPTION: 1-sentence description (max 150 chars)
- LONG_DESCRIPTION: 2-3 sentence detailed description
- MAIN_FEATURES: comma-separated list of 3-6 key features
- INNOVATION_SCORE: 1-10 (10=paradigm shift, 5=incremental, 1=marginal)
- TAGS: comma-separated lowercase tags (5-8 tags)

Context:
{context}

Return ONLY the JSON object, no other text."""

    return prompt

def is_garbage(rdata):
    """Check if parsed data is garbage/noise"""
    sd = stringify(rdata.get('SHORT_DESCRIPTION', ''))
    if not sd or len(sd) < 8:
        return True, "short description too short"
    if sd.lower() in ['n/a', 'none', 'no description', 'placeholder', 'tbd']:
        return True, "placeholder description"
    ld = stringify(rdata.get('LONG_DESCRIPTION', ''))
    if len(ld) < 10:
        return True, "long description too short"
    return False, ""

def reclassify_entry(a, eid, mapped):
    """Update layer_membership for an entry"""
    if mapped and mapped in BORG_TAXONOMY:
        a.execute("DELETE FROM layer_membership WHERE entry_id=?", (eid,))
        a.execute("INSERT INTO layer_membership (entry_id, layer, is_primary) VALUES (?, ?, 1)", (eid, mapped))

def compute_scores(rdata, is_gh, owner, page_title, existing_innovation=0):
    """Compute quality, signal, innovation scores"""
    long_desc = stringify(rdata.get('LONG_DESCRIPTION', ''))
    features = stringify(rdata.get('MAIN_FEATURES', ''))
    tags_str = stringify(rdata.get('TAGS', ''))
    innovation = rdata.get('INNOVATION_SCORE', 8)
    
    try:
        innovation = int(innovation)
    except (ValueError, TypeError):
        innovation = 8
    innovation = max(1, min(10, innovation))
    new_innov = max(innovation, existing_innovation)
    
    # Normalize tags
    tags = [t.strip().lower().replace(' ', '-').replace('_', '-') for t in tags_str.split(',') if t.strip()]
    seen_tags = set()
    clean_tags = []
    for t in tags:
        if t not in seen_tags:
            seen_tags.add(t)
            clean_tags.append(t)
    
    # Quality score
    score = 0.0
    ld_len = len(long_desc)
    if ld_len > 500: score += 30
    elif ld_len > 300: score += 25
    elif ld_len > 150: score += 20
    elif ld_len > 50: score += 12
    elif ld_len > 10: score += 6
    
    feat_count = len([x.strip() for x in features.split(',') if x.strip() and len(x.strip()) > 3])
    if feat_count >= 5: score += 25
    elif feat_count >= 4: score += 22
    elif feat_count >= 3: score += 18
    elif feat_count >= 2: score += 12
    elif feat_count >= 1: score += 6
    
    tag_count = len(clean_tags)
    if tag_count >= 6: score += 15
    elif tag_count >= 4: score += 12
    elif tag_count >= 2: score += 8
    elif tag_count >= 1: score += 4
    
    if page_title and len(page_title) > 5: score += 10
    if owner and len(owner) > 1: score += 10
    quality = min(1.0, score / 100)
    
    feat_score = min(15, feat_count * 3)
    desc_score = min(10, len(long_desc) / 50.0)
    gh_bonus = 5 if is_gh else 0
    signal = min(100, max(0, int(round(
        (new_innov * 4) + (quality * 30) + feat_score + desc_score + gh_bonus
    ))))
    is_standout = 1 if new_innov >= 9 and quality >= 0.8 else 0
    
    return {
        'quality': quality, 'signal': signal, 'innovation': new_innov,
        'is_standout': is_standout, 'tags': clean_tags, 'feat_count': feat_count
    }

def ingest_new_url(a, atl, url, rdata, is_gh, owner, repo, page_title):
    """Insert a brand new entry into the atlas"""
    scores = compute_scores(rdata, is_gh, owner, page_title)
    
    short_desc = stringify(rdata.get('SHORT_DESCRIPTION', ''))
    long_desc = stringify(rdata.get('LONG_DESCRIPTION', '')) or short_desc
    features = stringify(rdata.get('MAIN_FEATURES', ''))
    
    a.execute("""INSERT INTO entries 
        (url, page_title, short_description, long_description, main_features,
         tags, owner, repo, is_github, innovation, quality, signal, is_standout, verdict, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
        (url, page_title or '', short_desc, long_desc, features,
         json.dumps(scores['tags']), owner or '', repo or '', 1 if is_gh else 0,
         scores['innovation'], scores['quality'], scores['signal'],
         scores['is_standout'], '', datetime.now().isoformat()))
    
    eid = a.execute("SELECT last_insert_rowid()").fetchone()[0]
    
    category = stringify(rdata.get('CATEGORY', ''))
    mapped = CAT_MAP.get(category, category)
    reclassify_entry(a, eid, mapped)
    atl.commit()
    
    stats['new_ingested'] += 1
    return eid

def enrich_existing(a, atl, eid, rdata, is_gh, owner, page_title, existing_sd, existing_innov):
    """Update an existing entry with enriched data"""
    scores = compute_scores(rdata, is_gh, owner, page_title, existing_innov)
    
    short_desc = stringify(rdata.get('SHORT_DESCRIPTION', ''))
    long_desc = stringify(rdata.get('LONG_DESCRIPTION', '')) or short_desc
    features = stringify(rdata.get('MAIN_FEATURES', ''))
    
    # Keep the better short description
    final_sd = short_desc if len(short_desc) > len(existing_sd or '') else (existing_sd or short_desc)
    
    a.execute("""UPDATE entries SET 
        short_description=?, long_description=?, main_features=?,
        tags=?, innovation=?, quality=?, signal=?, is_standout=?, verdict=?
        WHERE id=?""",
        (final_sd, long_desc, features,
         json.dumps(scores['tags']), scores['innovation'], scores['quality'],
         scores['signal'], scores['is_standout'], '', eid))
    
    category = stringify(rdata.get('CATEGORY', ''))
    mapped = CAT_MAP.get(category, category)
    reclassify_entry(a, eid, mapped)
    atl.commit()
    
    if len(features) > 5:
        stats['features_enriched'] += 1
    if len(long_desc) > len(existing_sd or ''):
        stats['desc_enriched'] += 1

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--dry-run', action='store_true', help="Don't write to DB")
    parser.add_argument('--limit', type=int, default=0, help="Max entries to process (0=all)")
    parser.add_argument('--phase', choices=['new', 'enrich', 'both'], default='both',
                        help="Which phase to run: new URLs, enrichment, or both")
    args = parser.parse_args()
    
    atl = sqlite3.connect(ATLAS_DB)
    a = atl.cursor()
    
    logger.info("=" * 60)
    logger.info("Borg Research Worker v3.0 - Deep Processing Pass")
    logger.info(f"Phase: {args.phase}, Limit: {args.limit or 'unlimited'}, Dry-run: {args.dry_run}")
    logger.info("=" * 60)
    
    # ---- PHASE 1: INGEST NEW URLS FROM INCOMING_RESOURCES ----
    if args.phase in ('new', 'both'):
        logger.info("\n--- PHASE 1: Ingesting new URLs from incoming_resources.txt ---")
        
        # Load existing URLs
        a.execute('SELECT url FROM entries')
        existing_urls = set(r[0] for r in a.fetchall())
        a.execute('SELECT LOWER(owner)||"/"||LOWER(repo) FROM entries WHERE owner IS NOT NULL AND repo IS NOT NULL')
        existing_repos = set(r[0] for r in a.fetchall())
        
        # Parse incoming resources
        new_urls = []
        with open('incoming_resources.txt', 'r', encoding='utf-8') as f:
            for line in f:
                url = line.strip()
                if not url:
                    continue
                if url in existing_urls:
                    continue
                # Check GitHub repos too
                m = re.match(r'https://github\.com/([^/]+)/([^/?#\s]+)', url)
                if m:
                    repo_key = f'{m.group(1).lower()}/{m.group(2).lower()}'
                    if repo_key in existing_repos:
                        continue
                new_urls.append(url)
        
        logger.info(f"New URLs to ingest: {len(new_urls)}")
        
        processed = 0
        for url in new_urls:
            if args.limit and processed >= args.limit:
                break
            processed += 1
            
            logger.info(f"[NEW {processed}/{len(new_urls)}] {url[:80]}")
            
            is_gh = 'github.com' in url
            owner, repo = None, None
            if is_gh:
                m = re.match(r'https://github\.com/([^/]+)/([^/?#\s]+)', url)
                if m:
                    owner, repo = m.group(1), m.group(2)
            
            # Fetch content
            content = fetch_content(url)
            fit_text = ""
            gh_meta = None
            reddit_ctx = None
            used_metadata = False
            
            if content:
                stats['fetched'] += 1
                fit_text = extract_fit_markdown(content, url)
                if is_gh:
                    gh_meta = extract_gh_meta(url, content)
            else:
                if 'reddit.com' in url.lower():
                    reddit_ctx = extract_reddit_context(url)
                elif is_gh and owner and repo:
                    gh_meta = {'desc': f"GitHub repository {owner}/{repo}"}
                used_metadata = True
                stats['metadata_only'] += 1
            
            # Skip HN/Reddit discussion-only URLs with no extractable content
            if len(fit_text) < 30 and not gh_meta and not reddit_ctx:
                # For non-GitHub, non-Reddit URLs with no content, try metadata-only
                if not used_metadata:
                    used_metadata = True
                else:
                    logger.info(f"  SKIP: no extractable content")
                    stats['skipped'] += 1
                    continue
            
            # Call LLM for enrichment
            prompt = build_prompt(url, fit_text, gh_meta, reddit_ctx)
            raw, model = call_llm(prompt, prefer_big=True)
            
            if not raw:
                logger.warning(f"  LLM failed for {url[:60]}")
                stats['failed'] += 1
                continue
            
            rdata = parse_llm_response(raw)
            if not rdata:
                logger.warning(f"  Parse failed (model={model})")
                with open('logs/parse_failures_v3.jsonl', 'a', encoding='utf-8') as df:
                    df.write(json.dumps({"url": url, "model": model, "raw": raw[:2000]}) + '\n')
                stats['failed'] += 1
                continue
            
            garbage, reason = is_garbage(rdata)
            if garbage:
                logger.info(f"  REJECTED: {reason}")
                stats['rejected'] += 1
                continue
            
            # Determine page title
            page_title = ''
            if content:
                try:
                    soup = BeautifulSoup(content, 'html.parser')
                    title_tag = soup.find('title')
                    if title_tag:
                        page_title = title_tag.get_text(strip=True)[:200]
                except:
                    pass
            
            if not args.dry_run:
                ingest_new_url(a, atl, url, rdata, is_gh, owner, repo, page_title)
            
            mode = "META" if used_metadata else "FETCH"
            mshort = model[:20] if model else '?'
            sd = stringify(rdata.get('SHORT_DESCRIPTION', ''))[:60]
            logger.info(f"  INGESTED [{mshort}] [{mode}]: {sd}")
            stats['accepted'] += 1
            time.sleep(0.3)
    
    # ---- PHASE 2: ENRICH EXISTING ENTRIES MISSING FEATURES ----
    if args.phase in ('enrich', 'both'):
        logger.info("\n--- PHASE 2: Enriching existing entries missing features ---")
        
        # Priority: high-signal entries first, then by whether they have features
        a.execute("""
            SELECT e.id, e.url, e.short_description, e.is_github, 
                   e.innovation, e.page_title, e.owner, e.repo,
                   e.main_features, e.long_description
            FROM entries e
            WHERE (e.main_features IS NULL OR e.main_features = ''
                   OR LENGTH(e.short_description) < 20)
            ORDER BY e.signal DESC, e.innovation DESC
        """)
        enrich_entries = a.fetchall()
        
        logger.info(f"Entries to enrich: {len(enrich_entries)}")
        
        processed = 0
        for eid, url, sd, is_gh, innov, pt, owner, repo, existing_features, existing_ld in enrich_entries:
            if args.limit and processed >= args.limit:
                break
            processed += 1
            
            logger.info(f"[ENRICH {processed}/{len(enrich_entries)}] {owner}/{repo or url[:50]}")
            
            content = fetch_content(url)
            fit_text = ""
            gh_meta = None
            used_metadata = False
            
            if content:
                stats['fetched'] += 1
                fit_text = extract_fit_markdown(content, url)
                if is_gh:
                    gh_meta = extract_gh_meta(url, content)
            else:
                if is_gh and owner and repo:
                    gh_meta = {'desc': f"GitHub repository {owner}/{repo}"}
                used_metadata = True
                stats['metadata_only'] += 1
            
            if len(fit_text) < 20 and not gh_meta:
                logger.info(f"  SKIP: no content")
                stats['skipped'] += 1
                continue
            
            prompt = build_prompt(url, fit_text, gh_meta, existing_sd=sd)
            raw, model = call_llm(prompt, prefer_big=(innov >= 8))
            
            if not raw:
                stats['failed'] += 1
                continue
            
            rdata = parse_llm_response(raw)
            if not rdata:
                with open('logs/parse_failures_v3.jsonl', 'a', encoding='utf-8') as df:
                    df.write(json.dumps({"url": url, "model": model, "raw": raw[:2000]}) + '\n')
                stats['failed'] += 1
                continue
            
            garbage, reason = is_garbage(rdata)
            if garbage:
                stats['rejected'] += 1
                continue
            
            if not args.dry_run:
                enrich_existing(a, atl, eid, rdata, is_gh, owner, pt, sd, innov)
            
            mode = "META" if used_metadata else "FETCH"
            mshort = model[:20] if model else '?'
            new_sd = stringify(rdata.get('SHORT_DESCRIPTION', ''))[:60]
            logger.info(f"  ENRICHED [{mshort}] [{mode}]: {new_sd}")
            stats['accepted'] += 1
            time.sleep(0.2)
    
    # ---- FINAL REPORT ----
    logger.info("\n" + "=" * 60)
    logger.info("DEEP PROCESSING COMPLETE")
    logger.info(f"  New URLs ingested:    {stats['new_ingested']}")
    logger.info(f"  Features enriched:    {stats['features_enriched']}")
    logger.info(f"  Descriptions updated: {stats['desc_enriched']}")
    logger.info(f"  Total accepted:       {stats['accepted']}")
    logger.info(f"  Rejected (garbage):   {stats['rejected']}")
    logger.info(f"  Failed (LLM/parse):   {stats['failed']}")
    logger.info(f"  Skipped (no content): {stats['skipped']}")
    logger.info(f"  Metadata-only:        {stats['metadata_only']}")
    logger.info(f"  Page fetches:         {stats['fetched']}")
    logger.info(f"  LLM calls:            {stats['llm_calls']}")
    logger.info(f"  LLM failures:         {stats['llm_failures']}")
    logger.info("=" * 60)
    
    # DB stats
    a.execute("SELECT COUNT(*) FROM entries")
    total = a.fetchone()[0]
    a.execute("SELECT COUNT(*) FROM entries WHERE main_features IS NOT NULL AND main_features != ''")
    with_features = a.fetchone()[0]
    a.execute("SELECT COUNT(*) FROM entries WHERE is_standout=1")
    standout = a.fetchone()[0]
    a.execute("SELECT COUNT(*) FROM entries WHERE signal >= 85")
    high_signal = a.fetchone()[0]
    logger.info(f"Atlas: {total:,} entries | {with_features:,} with features | {high_signal:,} high-signal | {standout:,} standout")
    
    atl.close()

if __name__ == "__main__":
    main()
