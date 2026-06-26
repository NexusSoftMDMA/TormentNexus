#!/usr/bin/env python3
"""Borg Link Scraper — scrape Reddit & HackerNews discussion pages for embedded URLs"""
import sys
import io
import os
import re
import json
import sqlite3
import time
import logging
from collections import OrderedDict
from urllib.parse import urlparse, urljoin

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

os.makedirs('logs', exist_ok=True)
logging.basicConfig(
    level=logging.INFO, format='%(asctime)s %(levelname)s %(message)s',
    handlers=[logging.FileHandler('logs/scrape_reddit_hn.log', mode='a', encoding='utf-8'), logging.StreamHandler()],
)
logger = logging.getLogger(__name__)

# Config
REQUEST_DELAY = 1.5  # seconds between requests
MAX_URLS = 2000       # limit per run to be practical

# We'll import requests/bs4 optionally
try:
    import requests
    from bs4 import BeautifulSoup
except ImportError:
    logger.error("Missing dependencies: pip install requests beautifulsoup4")
    sys.exit(1)

ATLAS_DB = 'atlas.db'
QUEUE_FILE = 'incoming_resources.txt'
SOURCE_URLS_FILE = 'reddit_hn_urls.txt'

# Noise patterns — skip these if found in extracted links
NOISE_DOMAINS = {
    'reddit.com', 'www.reddit.com', 'old.reddit.com', 'new.reddit.com',
    'ycombinator.com', 'news.ycombinator.com',
    'twitter.com', 'x.com', 'youtube.com', 'youtu.be', 'tiktok.com',
    'instagram.com', 'facebook.com', 'linkedin.com',
    'amazon.com', 'amzn.to', 'ebay.com', 'etsy.com', 'shop.',
    'paypal.com', 'stripe.com',
    'google.com/search', 'googleadservices.com',
    'doubleclick.net', 'googlesyndication.com',
    'wikipedia.org', 'wikidata.org',
    'medium.com', 'substack.com',
    'github.com/settings', 'github.com/login',
}

NOISE_PATTERNS = [
    r'mailto:', r'tel:', r'javascript:',
    r'\.pdf$', r'\.zip$', r'\.tar\.gz$', r'\.exe$',
    r'reddit\.com/r/\w+/comments/\w+/\w+/[a-z0-9]+/$',
    r'reddit\.com/r/\w+/submit\?',
    r'reddit\.com/login',
    r'reddit\.com/register',
    r'news\.ycombinator\.com/submitted',
    r'news\.ycombinator\.com/threads',
    r'news\.ycombinator\.com/vote',
    r'news\.ycombinator\.com/reply',
]

def norm_url(u):
    """Normalize URL for dedup"""
    u = u.strip().rstrip('/')
    u = u.replace('http://', 'https://')
    nu = u.lower().replace('www.', '')
    if '#' in nu:
        nu = nu.split('#')[0]
    return nu

def is_noise(url):
    """Check if URL is noise"""
    url_lower = url.lower()
    parsed = urlparse(url_lower)
    domain = parsed.netloc
    
    for nd in NOISE_DOMAINS:
        if nd in domain or nd in url_lower:
            return True
    
    for p in NOISE_PATTERNS:
        if re.search(p, url_lower):
            return True
    
    return False

def is_tool_url(url):
    """Check if URL is relevant for the atlas"""
    url_lower = url.lower()
    parsed = urlparse(url_lower)
    domain = parsed.netloc
    
    # Must have scheme
    if parsed.scheme not in ('http', 'https'):
        return False
    
    # GitHub repos
    if 'github.com' in domain:
        path_parts = parsed.path.strip('/').split('/')
        if len(path_parts) >= 2 and path_parts[0] and path_parts[1]:
            # It's a repo (user/repo)
            if not any(x in path_parts[1].lower() for x in ['issues', 'pull', 'releases', 'wiki', 'settings']):
                return True
    
    # Known tool/documentation domains
    TOOL_DOMAINS = [
        'pypi.org', 'npmjs.com', 'pypi.python.org',
        'arxiv.org', 'arxiv-vanity.com', 'huggingface.co',
        'paperswithcode.com', 'openai.com', 'anthropic.com',
        'deepmind.com', 'ai.google', 'ai.meta.com',
        'mistral.ai', 'cohere.com', 'replicate.com',
        'llamaindex.ai', 'langchain.com',
        'modelcontextprotocol.io',
    ]
    for td in TOOL_DOMAINS:
        if td in domain:
            return True
    
    # Pages that look like tool docs or blog posts about AI tools
    if any(kw in url_lower for kw in [
        '/blog/', '/docs/', '/documentation', '/getting-started',
        '/quickstart', '/tutorial', '/guide', '/api/',
        'mcp', 'ai-agent', 'llm', 'gpt', 'claude', 'gemini',
    ]):
        return True
    
    return False

def extract_links_from_reddit(soup, url):
    """Extract links from a Reddit page"""
    links = set()
    
    # Old Reddit — look for links in comments and posts
    # New Reddit — different selectors
    
    # All <a> tags
    for a in soup.find_all('a', href=True):
        href = a['href'].strip()
        if not href or href.startswith('#'):
            continue
        
        # Resolve relative URLs
        absolute = urljoin(url, href)
        
        # Skip reddit internal links
        parsed = urlparse(absolute)
        if 'reddit.com' in parsed.netloc.lower():
            continue
        
        links.add(absolute)
    
    return links

def extract_links_from_hn(soup, url):
    """Extract links from HackerNews page"""
    links = set()
    
    for a in soup.find_all('a', href=True):
        href = a['href'].strip()
        if not href or href.startswith('#'):
            continue
        
        absolute = urljoin(url, href)
        
        # Skip HN internal links
        parsed = urlparse(absolute)
        if 'ycombinator.com' in parsed.netloc.lower():
            continue
        
        links.add(absolute)
    
    return links

def load_existing_urls():
    """Load all existing URLs from DB and queue for dedup"""
    existing = set()
    
    # From DB
    try:
        conn = sqlite3.connect(ATLAS_DB)
        cursor = conn.cursor()
        cursor.execute('SELECT url FROM entries')
        for (url,) in cursor.fetchall():
            existing.add(norm_url(url))
        conn.close()
    except Exception as e:
        logger.warning(f"Error loading DB URLs: {e}")
    
    # From queue
    try:
        with open(QUEUE_FILE, 'r', encoding='utf-8') as f:
            for line in f:
                u = line.strip()
                if u:
                    existing.add(norm_url(u))
    except FileNotFoundError:
        pass
    
    return existing

def main():
    logger.info("Borg Link Scraper starting")
    
    # Load source URLs
    try:
        with open(SOURCE_URLS_FILE, 'r', encoding='utf-8') as f:
            source_urls = [l.strip() for l in f if l.strip()]
    except FileNotFoundError:
        logger.error(f"{SOURCE_URLS_FILE} not found")
        return
    
    logger.info(f"Source URLs to process: {len(source_urls)}")
    
    # Load existing URLs for dedup
    existing = load_existing_urls()
    logger.info(f"Existing URLs (DB+queue): {len(existing)}")
    
    # Track newly discovered links
    discovered = OrderedDict()
    processed = 0
    errors = 0
    skipped_src = 0
    
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
    }
    
    for idx, src_url in enumerate(source_urls):
        if len(discovered) >= MAX_URLS:
            logger.info(f"Reached max new URLs ({MAX_URLS}), stopping")
            break
        
        # Skip Reddit URLs that are comments
        if 'reddit.com' in src_url.lower() and re.search(r'/comments/\w+/\w+/\w+/(?!$)', src_url):
            skipped_src += 1
            continue
        
        processed += 1
        
        try:
            resp = requests.get(src_url, headers=headers, timeout=20)
            resp.raise_for_status()
        except Exception as e:
            errors += 1
            if errors % 10 == 0:
                logger.info(f"  [{idx+1}/{len(source_urls)}] Error {src_url[:80]}: {str(e)[:40]}")
            continue
        
        soup = BeautifulSoup(resp.text, 'html.parser')
        
        if 'reddit.com' in src_url.lower():
            raw_links = extract_links_from_reddit(soup, src_url)
        else:
            raw_links = extract_links_from_hn(soup, src_url)
        
        added = 0
        for link in raw_links:
            if len(discovered) >= MAX_URLS:
                break
            
            n = norm_url(link)
            if n in existing or n in discovered:
                continue
            if is_noise(link):
                continue
            if not is_tool_url(link):
                continue
            
            discovered[n] = link
            added += 1
        
        if added > 0:
            logger.info(f"  [{idx+1}/{len(source_urls)}] {src_url[:70]}... → {added} new URLs")
        
        time.sleep(REQUEST_DELAY)
    
    logger.info(f"\nSummary: processed {processed} pages, {errors} errors, {skipped_src} skipped comments")
    logger.info(f"New URLs discovered: {len(discovered)}")
    
    # Append new URLs to queue
    if discovered:
        with open(QUEUE_FILE, 'a', encoding='utf-8') as f:
            for url in discovered.values():
                f.write(url + '\n')
        logger.info(f"Appended {len(discovered)} new URLs to {QUEUE_FILE}")
    
    # Save discovered list for inspection
    discovered_path = 'logs/discovered_links.json'
    with open(discovered_path, 'w', encoding='utf-8') as f:
        json.dump(list(discovered.values()), f, indent=2)
    logger.info(f"Saved discovered links to {discovered_path}")

if __name__ == '__main__':
    main()
