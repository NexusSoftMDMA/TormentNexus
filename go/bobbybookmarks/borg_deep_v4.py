#!/usr/bin/env python3
"""Borg Deep Processor v4 - Crawl + Ingest + Enrich"""

import os
import re
import json
import sqlite3
import requests
import time
import logging
import argparse
from datetime import datetime
from urllib.parse import urljoin, urlparse
from bs4 import BeautifulSoup

os.makedirs("logs", exist_ok=True)
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    handlers=[
        logging.FileHandler("logs/v4_deep.log", mode="a", encoding="utf-8"),
        logging.StreamHandler(),
    ],
)
log = logging.getLogger(__name__)

ATLAS_DB = "atlas.db"
INCOMING = "incoming_resources.txt"
LMSTUDIO_URL = "http://localhost:1234/v1/chat/completions"
MODELS = [
    ("gemma-4-e2b-uncensored-hauhaucs-aggressive", 90),
    ("gemma-4-e4b-uncensored-hauhaucs-aggressive", 120),
    ("qwen3.6-27b-uncensored-hauhaucs-aggressive", 150),
    ("gemma-4-26b-a4b-it-heretic-ara", 180),
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
    "Connectivity & Interoperability (MCP/A2A)": "Connectivity / MCP / A2A",
    "Development Tools & Libraries": "Coding Tools & IDEs",
    "Developer Workflow": "Developer Workflow & Tools",
    "Guides & Articles": "Guides & Industry Trends",
    "Infrastructure": "Infrastructure & Proxy Layers",
    "Other": None,
    "Software Development": "Developer Workflow & Tools",
    "Software/Developer Tools": "Coding Tools & IDEs",
}
FIELD_NAMES = [
    "CATEGORY",
    "SHORT_DESCRIPTION",
    "LONG_DESCRIPTION",
    "MAIN_FEATURES",
    "INNOVATION_SCORE",
    "TAGS",
]
EMPTY = chr(39) + chr(39)

stats = {
    "crawled": 0,
    "links_found": 0,
    "new_links": 0,
    "ingested": 0,
    "enriched": 0,
    "failed": 0,
    "skipped": 0,
    "rejected": 0,
    "llm_calls": 0,
    "llm_fails": 0,
}


def stringify(v):
    if v is None:
        return ""
    if isinstance(v, str):
        return v
    if isinstance(v, (list, tuple)):
        return ", ".join(str(x) for x in v)
    return str(v)


def call_llm(prompt, timeout_override=None):
    for model, tout in MODELS:
        t = timeout_override or tout
        try:
            stats["llm_calls"] += 1
            resp = requests.post(
                LMSTUDIO_URL,
                json={
                    "model": model,
                    "messages": [{"role": "user", "content": prompt}],
                    "temperature": 0.2,
                    "max_tokens": 350,
                },
                timeout=t,
            )
            if resp.status_code == 200:
                choices = resp.json().get("choices", [{}])
                text = choices[0].get("message", {}).get("content", "")
                if text and len(text) > 30:
                    return text, model
            else:
                log.warning("  HTTP %d from %s", resp.status_code, model)
        except requests.exceptions.Timeout:
            log.warning("  Timeout (%ds) on %s", t, model)
        except Exception as exc:
            log.warning("  Error on %s: %s", model, str(exc)[:80])
        time.sleep(1)
    stats["llm_fails"] += 1
    return None, None


def parse_llm_response(raw):
    if not raw:
        return None
    text = raw.strip()
    try:
        return json.loads(text)
    except Exception:
        pass
    b3 = chr(96) * 3
    for delim in [b3 + "json", b3]:
        if delim in text:
            for part in text.split(delim)[1:]:
                end = part.find(b3)
                block = part[:end].strip() if end >= 0 else part.strip()
                try:
                    return json.loads(block)
                except Exception:
                    continue
    result = {}
    for field in FIELD_NAMES:
        # Pattern 1: "FIELD": "value" (string value)
        pat1 = '"' + field + '"' + r'\s*:\s*"' + r'((?:[^"\\]|\\.)*)' + '"'
        # Pattern 2: "FIELD": number
        pat2 = '"' + field + '"' + r'\s*:\s*(\d+)'
        # Pattern 3: FIELD: "value" (without quotes on key)
        pat3 = field + r'\s*:\s*"' + r'((?:[^"\\]|\\.)*)' + '"'
        for pat in [pat1, pat2, pat3]:
            try:
                m = re.search(pat, text)
                if m:
                    result[field] = m.group(1)
                    break
            except re.error:
                continue
    return result if len(result) >= 3 else None


def is_garbage(rdata):
    sd = stringify(rdata.get("SHORT_DESCRIPTION", ""))
    if not sd or len(sd) < 8:
        return True, "short desc"
    if sd.lower() in ("n/a", "none", "no description", "placeholder", "tbd"):
        return True, "placeholder"
    ld = stringify(rdata.get("LONG_DESCRIPTION", ""))
    if len(ld) < 10:
        return True, "long desc too short"
    return False, ""


def fetch_page(url, timeout=18):
    fetch_url = url
    if "www.reddit.com" in url:
        fetch_url = url.replace("www.reddit.com", "old.reddit.com")
    elif "reddit.com" in url and "old.reddit.com" not in url:
        fetch_url = url.replace("https://reddit.com", "https://old.reddit.com")
    try:
        resp = requests.get(
            fetch_url,
            headers={"User-Agent": "Mozilla/5.0 (compatible; BorgIntel/4.0)"},
            timeout=timeout,
            allow_redirects=True,
        )
        if resp.status_code == 200:
            return resp.text
    except Exception:
        pass
    if fetch_url != url:
        try:
            resp = requests.get(
                url,
                headers={"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},
                timeout=timeout,
                allow_redirects=True,
            )
            if resp.status_code == 200:
                return resp.text
        except Exception:
            pass
    return None


def fetch_hn_story_url(hn_url):
    m = re.search(r"id=(\d+)", hn_url)
    if not m:
        return None
    try:
        item_url = "https://hacker-news.firebaseio.com/v0/item/" + m.group(1) + ".json"
        resp = requests.get(item_url, timeout=10)
        if resp.status_code == 200:
            return resp.json().get("url")
    except Exception:
        pass
    return None


_URL_RE = re.compile(r"https?://[^\s<>\x22\x27=]+")


def extract_links_from_html(html, base_url):
    if not html:
        return set()
    links = set()
    for u in _URL_RE.findall(html):
        u = u.rstrip(").,;:}").strip()
        if "#" in u:
            u = u[: u.index("#")]
        if len(u) > 250:
            continue
        links.add(u)
    try:
        soup = BeautifulSoup(html, "html.parser")
        for a_tag in soup.find_all("a", href=True):
            href = str(a_tag["href"]).strip()
            if not href or href.startswith("javascript:") or href.startswith("mailto:"):
                continue
            full = urljoin(base_url, href)
            parsed = urlparse(full)
            if not parsed.scheme or not parsed.netloc:
                continue
            links.add(parsed._replace(fragment="").geturl())
    except Exception:
        pass
    return links


def extract_readable(html, url=""):
    if not html:
        return ""
    try:
        soup = BeautifulSoup(html, "html.parser")
        for tag in soup.find_all(
            ["script", "style", "nav", "footer", "header", "aside"]
        ):
            tag.decompose()
        main_el = (
            soup.find("main")
            or soup.find("article")
            or soup.find(attrs={"role": "main"})
        )
        text = (main_el or soup).get_text(separator="\n", strip=True)
        lines = [ln.strip() for ln in text.split("\n") if ln.strip()]
        return "\n".join(lines[:120])[:2500]
    except Exception:
        return ""


def extract_gh_meta(url, html):
    meta = {}
    m = re.match(r"https://github\.com/([^/]+)/([^/?#]+)", url)
    if m:
        meta["owner"] = m.group(1)
        meta["repo"] = m.group(2)
    if html:
        try:
            soup = BeautifulSoup(html, "html.parser")
            about = soup.find(attrs={"class": re.compile(r"about", re.I)})
            if about:
                meta["desc"] = about.get_text(strip=True)[:200]
            topics = soup.find_all(attrs={"class": re.compile(r"topic-tag", re.I)})
            if topics:
                meta["topics"] = ", ".join(t.get_text(strip=True) for t in topics[:10])
        except Exception:
            pass
    return meta if meta else None


def build_prompt(url, fit_text, gh_meta=None, existing_sd=None):
    ctx = ["URL: " + url]
    if gh_meta:
        o = gh_meta.get("owner", "")
        r = gh_meta.get("repo", "")
        if o and r:
            ctx.append("GitHub: " + o + "/" + r)
        if "desc" in gh_meta:
            ctx.append("About: " + gh_meta["desc"])
        if "topics" in gh_meta:
            ctx.append("Topics: " + gh_meta["topics"])
    if existing_sd and len(existing_sd) > 5:
        ctx.append("Existing description: " + existing_sd)
    if fit_text and len(fit_text) > 50:
        ctx.append("Page content:\n" + fit_text[:1800])
    ctx_str = "\n".join(ctx)
    tax = ", ".join(BORG_TAXONOMY)
    parts = [
        "Analyze this AI/developer tool or resource. Return a JSON object with these fields:",
        "- CATEGORY: one of [" + tax + "]",
        "- SHORT_DESCRIPTION: 1-sentence description (max 150 chars)",
        "- LONG_DESCRIPTION: 2-3 sentence detailed description",
        "- MAIN_FEATURES: comma-separated list of 3-6 key features",
        "- INNOVATION_SCORE: 1-10 (10=paradigm shift, 5=incremental, 1=marginal)",
        "- TAGS: comma-separated lowercase tags (5-8 tags)",
        "Context:",
        ctx_str,
        "Return ONLY the JSON object, no other text.",
    ]
    return "\n".join(parts)


def compute_scores(rdata, is_gh, owner, page_title, existing_innovation=0):
    ld = stringify(rdata.get("LONG_DESCRIPTION", ""))
    features = stringify(rdata.get("MAIN_FEATURES", ""))
    tags_str = stringify(rdata.get("TAGS", ""))
    inn = rdata.get("INNOVATION_SCORE", 8)
    try:
        inn = int(inn)
    except (ValueError, TypeError):
        inn = 8
    inn = max(1, min(10, inn))
    new_inn = max(inn, existing_innovation)

    tags = [
        t.strip().lower().replace(" ", "-").replace("_", "-")
        for t in tags_str.split(",")
        if t.strip()
    ]
    seen, clean = set(), []
    for t in tags:
        if t not in seen:
            seen.add(t)
            clean.append(t)

    score = 0.0
    ld_len = len(ld)
    if ld_len > 500:
        score += 30
    elif ld_len > 300:
        score += 25
    elif ld_len > 150:
        score += 20
    elif ld_len > 50:
        score += 12
    elif ld_len > 10:
        score += 6

    feat_count = len(
        [x.strip() for x in features.split(",") if x.strip() and len(x.strip()) > 3]
    )
    if feat_count >= 5:
        score += 25
    elif feat_count >= 4:
        score += 22
    elif feat_count >= 3:
        score += 18
    elif feat_count >= 2:
        score += 12
    elif feat_count >= 1:
        score += 6

    tag_count = len(clean)
    if tag_count >= 6:
        score += 15
    elif tag_count >= 4:
        score += 12
    elif tag_count >= 2:
        score += 8
    elif tag_count >= 1:
        score += 4

    if page_title and len(page_title) > 5:
        score += 10
    if owner and len(owner) > 1:
        score += 10

    quality = min(1.0, score / 100)
    feat_score = min(15, feat_count * 3)
    desc_score = min(10, len(ld) / 50.0)
    gh_bonus = 5 if is_gh else 0
    signal = min(
        100,
        max(
            0,
            int(
                round(
                    (new_inn * 4) + (quality * 30) + feat_score + desc_score + gh_bonus
                )
            ),
        ),
    )
    is_standout = 1 if new_inn >= 9 and quality >= 0.8 else 0
    return {
        "quality": quality,
        "signal": signal,
        "innovation": new_inn,
        "is_standout": is_standout,
        "tags": clean,
        "feat_count": feat_count,
    }


def phase_crawl(limit=0):
    log.info("=" * 60)
    log.info("PHASE 1: CRAWL - Extract links from HN/Reddit pages")

    with open(INCOMING, "r", encoding="utf-8") as f:
        all_urls = set(l.strip() for l in f if l.strip())

    crawl_targets = [
        url for url in all_urls if "news.ycombinator.com" in url or "reddit.com" in url
    ]
    log.info("Crawl targets: %d", len(crawl_targets))
    if limit:
        crawl_targets = crawl_targets[:limit]

    new_links_total = set()
    for i, url in enumerate(crawl_targets, 1):
        log.info("[CRAWL %d/%d] %s", i, len(crawl_targets), url[:80])

        if "news.ycombinator.com" in url:
            ext_url = fetch_hn_story_url(url)
            if ext_url and ext_url not in all_urls:
                new_links_total.add(ext_url)
                log.info("  HN API -> %s", ext_url[:70])
            else:
                log.info("  HN: no new")
            stats["crawled"] += 1
            time.sleep(0.3)
            continue

        html = fetch_page(url, timeout=15)
        if not html:
            stats["skipped"] += 1
            continue

        links = extract_links_from_html(html, url)
        stats["crawled"] += 1
        stats["links_found"] += len(links)

        NOISE = [
            "reddit.com",
            "redd.it",
            "redditstatic.com",
            "redditmedia.com",
            "ycombinator.com",
            "hn.algolia.com",
            "google.com/search",
            "twitter.com",
            "facebook.com",
            "apple.com",
            "play.google.com",
            "amazon.com",
            "amzn.to",
            "redditblog.com",
            "redditinc.com",
            "reddithelp.com",
            "w3.org",
            "itunes.apple.com",
        ]
        kept = set()
        for link in links:
            if not link.startswith("http"):
                continue
            lower = link.lower()
            if any(nd in lower for nd in NOISE):
                continue
            if "reddit.com" in lower:
                continue
            if len(link) > 250:
                continue
            kept.add(link)

        new_links = kept - all_urls
        if new_links:
            new_links_total.update(new_links)
            log.info("  %d links, %d NEW", len(kept), len(new_links))
        else:
            log.info("  %d links, 0 new", len(kept))
        time.sleep(0.5)

    if new_links_total:
        with open(INCOMING, "a", encoding="utf-8") as f:
            for u in sorted(new_links_total):
                f.write(u + "\n")
        stats["new_links"] = len(new_links_total)
        log.info("Appended %d new URLs", len(new_links_total))


def phase_ingest(limit=0):
    log.info("=" * 60)
    log.info("PHASE 2: INGEST - Process new URLs into Atlas")
    atl = sqlite3.connect(ATLAS_DB)
    c = atl.cursor()
    c.execute("SELECT url FROM entries")
    existing_urls = set(r[0] for r in c.fetchall())
    with open(INCOMING, "r", encoding="utf-8") as f:
        incoming = [l.strip() for l in f if l.strip()]
    new_urls = [u for u in incoming if u not in existing_urls and u.startswith("http")]

    # Pre-filter noise URLs before applying limit
    noise_patterns = [
        "/login", "/signup", "/auth", "/oauth", "/settings",
        "/device/", "/billing", "/marketplace", "/notifications",
        "/blob/", "/tree/", "/raw/", "/blame/", "/edit/",
        "/issues/", "/pull/", "/commits/", "/compare/",
        "/releases/", "/actions", "/security", "/wiki/",
        "/releases/tag/",
        "/session/",
        "/skills/",
    ]
    noise_domains = [
        "0.0.0.0", "127.0.0.1", "localhost", "::1",
        "tumblr.com", "medium.com", "substack.com",
        "youtube.com", "youtu.be",
        "linkedin.com/", "facebook.com/", "twitter.com/",
        "instagram.com/", "tiktok.com/",
        "patreon.com/", "buymeacoffee.com/",
        "chrome.google.com/webstore",
        "apps.apple.com", "play.google.com/store",
        "microsoft.com/store",
        "bandcamp.com", "discogs.com",
        "giphy.com",
        "wolframalpha.com/input", "uapreporting.org",
        "djfindr.com", "deepvaluereports.com",
        "smartymeapp.com", "post.smzdm.com",
        "pmvhaven.com",
        "adoptium.net",
        "aim.applyists.net",
        "aimoprize.com",
        "theatlantic.com",
        "engadget.com",
        "madison-reed.com",
        "lookchem.com",
        "abc.net.au",
        "hey.com/dhh",
        "prepaidcompare.net",
        "public.com/better",
        "portal.mendfamily.com", "discord.com/invite",
        "googleapis.com/v1internal",
        "cloudcode-pa.googleapis.com",
        "kilosessions.ai", "rns.id/app",
        "discord.com/invite",
        "discord.gg/",
        "suntimes.com",
        "openmhz.com",
        "open.spotify.com",
        "spotify.com/",
        "vault.fbi.gov",
        "ufos.wiki",
        "telegram.me/",
        "t.me/",
        "chat.openai.com",
        "goo.gl/forms",
        "drive.proton.me",
        "pmc.ncbi.nlm.nih.gov",
        "gemsloot.com",
        "jnco.com",
        "spacetribe.com",
        "minifigures.space",
        "sotozen.com",
        "afu.info",
        "temu.com",
        "bidprowl.com",
        "govauctions.app",
        "ektoplazm.com",
        "on.soundcloud.com",
        "soundcloud.com/",
        "fractaltribe.org",
        "townbuzz.app",
        "thechilluminati.com",
        "retrogames.cc",
        "oau.bet",
        "youtubetime.com",
        "checkloadapp.com",
        "env.md",
        "shop.futurebit.io",
        "ebay.com",
        "wsj.com",
        "academic.oup.com",
        "bing.com/search",
        "ebay.com",
        "wsj.com",
        "temu.com",
        "psymedia.co.za",
        "bit.ly/",
        "api.context.dev",
        "api.mcp-assistant.in",
        "goo.gl/",
        "slack.com/",
        "ra.co/events/",
        "fxgears.com",
        "letmegooglethat.com",
        "formgrid.com",
        "icrl.org",
        "share.formgrid.com",
        "giphy.com/gifs/",
        "fandom.com/wiki/",
        "wikipedia.org/wiki/",
        "rateyourmusic.com",
        "soatok.blog",
        "stocktwits.com",
        "unitednuclear.com",
        "variety.com",
        "apnews.com",
        "bloomberg.com",
        "bbc.com/news",
        "theguardian.com",
        "economist.com",
        "rnz.co.nz",
        "adsabs.harvard.edu",
        "sciencedirect.com",
        "cloudflarestatus.com",
        "skool.com",
        "war.gov",
        "drive.google.com",
        "forum.bitcoin.com",
        "freedomhouse.org",
        "etsy.com",
        "movementfestival.com",
        "foodgonewrong.com",
        "portal.mendfamily.com",
        "smartymeapp.com",
        "rns.id/app",
        "post.smzdm.com",
        "foodgonewrong.com",
        "127.0.0.1",
        "photosort-production",
        "portal.mendfamily.com",
        "start.smartymeapp.com",
        "rns.id/app",
        "post.smzdm.com",
        "discogs.com",
        "gumroad.com/l/",
        "jules.google.com/session",
        "nytimes.com",
        "theglobeandmail.com",
        "macrumors.com",
        "404media.co",
        "semiengineering.com",
        "construct.net/en/blogs",
        "discogs.com",
        "gumroad.com/l/",
        "drive.google.com",
        "127.0.0.1",
        "war.gov",
        "freedomhouse.org",
    
    ]
    filtered_urls = []
    for u in new_urls:
        ul = u.lower()
        # Skip local/internal URLs
        if ul.startswith("http://0.") or ul.startswith("http://127.") or ul.startswith("http://localhost"):
            continue
        if ul.startswith("http://%") or "dockerdesktop" in ul:
            continue
        if not ul.startswith("https://") and not ul.startswith("http://"):
            continue
        if "gist.github.com" in ul:
            continue
        if "api.apis.guru" in ul:
            continue
        if "news.ycombinator.com/item" in ul:
            continue
        if "reddit.com/r/" in ul and "/comments/" in ul:
            continue
        if any(x in ul for x in noise_patterns):
            continue
        if any(x in ul for x in noise_domains):
            continue
        # Session/ephemeral URLs
        if "jules.google.com/session" in ul:
            continue
        if "gumroad.com/l/" in ul:
            continue
        if "photosort-production" in ul:
            continue
        # API endpoints (not browsable)
        if "googleapis.com/v1" in ul:
            continue
        # Individual tool endpoints on MCP directories
        if "glama.ai/mcp/servers/" in ul and "/tools/" in ul:
            continue
        # Landing pages with no substance
        if ul.rstrip("/") in ["http://claude.ai", "https://claude.ai"]:
            continue
        # Markdown link artifacts (URL contains ](  
        if "](" in u or "](" in ul:
            continue
        # Obvious garbage URLs
        if "xxxxxx" in ul:
            continue
        # HTML entities in URL (broken by markdown parsing)
        if "&quot;" in ul or "&amp;" in ul or "&lt;" in ul or "&gt;" in ul:
            continue
        # Standalone markdown filenames (not real URLs)
        if ul.endswith(".md") and "/" not in ul[7:]:
            continue
        # StackOverflow careers/jobs
        if "careers.stackoverflow.com" in ul or "stackoverflow.com/jobs" in ul:
            continue
        # URL shorteners and API endpoints
        if ul.startswith("http://bit.ly") or ul.startswith("https://bit.ly"):
            continue
        # Search engine result pages
        if "google.com/search" in ul or "bing.com/search" in ul:
            continue
        # URL-encoded duplicates (markdown artifacts)
        if "%5b" in ul or "%5d" in ul or ")[" in ul:
            continue
        # News sites with no AI relevance
        if "suntimes.com" in ul or "newrepublic.com" in ul:
            continue
        filtered_urls.append(u)
    log.info("After noise filter: %d (removed %d)", len(filtered_urls), len(new_urls) - len(filtered_urls))
    new_urls = filtered_urls

    log.info("New URLs to ingest: %d", len(new_urls))
    if limit:
        new_urls = new_urls[:limit]

    for i, url in enumerate(new_urls, 1):
        try:
            log.info("[INGEST %d/%d] %s", i, len(new_urls), url[:80])
            # Normalize URL
            url_lower = url.lower()
            if url_lower.startswith("http://github.com"):
                url = "https://github.com" + url[len("http://github.com"):]
                url_lower = url.lower()
            is_gh = "github.com" in url_lower
            is_gist = "gist.github.com" in url_lower
            owner, repo = None, None
            if is_gh and not is_gist:
                m = re.match(r"https?://github\.com/([^/]+)/([^/?#\s]+)", url, re.IGNORECASE)
                if m:
                    owner, repo = m.group(1), m.group(2)
            html = fetch_page(url)
            fit_text, gh_meta = "", None
            if html:
                fit_text = extract_readable(html, url)
                if is_gh:
                    gh_meta = extract_gh_meta(url, html)
            elif is_gh and owner and repo:
                gh_meta = {
                    "owner": owner,
                    "repo": repo,
                    "desc": "GitHub repository " + owner + "/" + repo,
                }
            if len(fit_text) < 20 and not gh_meta:
                stats["skipped"] += 1
                continue
            raw, model = call_llm(build_prompt(url, fit_text, gh_meta))
            if not raw:
                stats["failed"] += 1
                continue
            rdata = parse_llm_response(raw)
            # Retry with simplified prompt if parse fails
            if not rdata:
                nl = chr(10)
                simple_prompt = (
                    "Return JSON for this URL: " + url + nl
                    + "Fields: CATEGORY, SHORT_DESCRIPTION, LONG_DESCRIPTION, "
                    + "MAIN_FEATURES, INNOVATION_SCORE(1-10), TAGS" + nl
                    + "Categories: " + ", ".join(BORG_TAXONOMY) + nl
                    + "Return ONLY valid JSON."
                )
                raw2, model2 = call_llm(simple_prompt)
                if raw2:
                    rdata = parse_llm_response(raw2)
                    if rdata:
                        model = model2
                        log.info("  Retry parse succeeded")
                if not rdata:
                    log.warning("  Parse failed")
                    stats["failed"] += 1
                    continue

            garbage, reason = is_garbage(rdata)
            if garbage:
                stats["rejected"] += 1
                continue

            page_title = ""
            if html:
                try:
                    soup = BeautifulSoup(html, "html.parser")
                    tt = soup.find("title")
                    if tt:
                        page_title = tt.get_text(strip=True)[:200]
                except Exception:
                    pass

            scores = compute_scores(rdata, is_gh, owner, page_title)
            short_desc = stringify(rdata.get("SHORT_DESCRIPTION", ""))
            long_desc = stringify(rdata.get("LONG_DESCRIPTION", "")) or short_desc
            features = stringify(rdata.get("MAIN_FEATURES", ""))

            c.execute(
                """INSERT INTO entries
                (url, page_title, short_description, long_description,
                 main_features, tags, owner, repo, is_github,
                 innovation, quality, signal, is_standout, verdict, created_at)
                VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)""",
                (url, page_title or "", short_desc, long_desc, features,
                 json.dumps(scores["tags"]), owner or "", repo or "",
                 1 if is_gh else 0, scores["innovation"], scores["quality"],
                 scores["signal"], scores["is_standout"], "",
                 datetime.now().isoformat()),
            )
            eid = c.execute("SELECT last_insert_rowid()").fetchone()[0]

            category = stringify(rdata.get("CATEGORY", ""))
            mapped = CAT_MAP.get(category, category)
            if mapped and mapped in BORG_TAXONOMY:
                c.execute("DELETE FROM layer_membership WHERE entry_id=?", (eid,))
                c.execute(
                    "INSERT OR REPLACE INTO layer_membership (entry_id, layer, is_primary) VALUES (?, ?, 1)",
                    (eid, mapped),
                )

            atl.commit()
            ml = (model or "?")[:25]
            log.info(" INGESTED [%s]: %s", ml, short_desc[:60])
            stats["ingested"] += 1
            time.sleep(0.3)

        except Exception as e:
            log.warning("  Error processing %s: %s", url[:50], str(e)[:60])
            stats["failed"] += 1
            continue
    c.execute("SELECT COUNT(*) FROM entries")
    total = c.fetchone()[0]
    c.execute(
        "SELECT COUNT(*) FROM entries WHERE main_features IS NOT NULL AND main_features != "
        + EMPTY
    )
    wf = c.fetchone()[0]
    log.info("Atlas now: %d entries | %d with features", total, wf)
    atl.close()


def phase_enrich(limit=0):
    log.info("=" * 60)
    log.info("PHASE 3: ENRICH - Fill missing features")

    atl = sqlite3.connect(ATLAS_DB)
    c = atl.cursor()
    c.execute(
        "SELECT e.id, e.url, e.short_description, e.is_github, "
        "e.innovation, e.page_title, e.owner, e.repo, "
        "e.main_features, e.long_description, e.signal "
        "FROM entries e "
        "WHERE (e.main_features IS NULL OR e.main_features = " + EMPTY + " "
        "OR LENGTH(e.short_description) < 20) "
        "ORDER BY e.signal DESC"
    )
    rows = c.fetchall()
    log.info("Entries needing enrichment: %d", len(rows))
    if limit:
        rows = rows[:limit]

    for i, (eid, url, sd, is_gh, inn, pt, owner, repo, ef, eld, sig) in enumerate(
        rows, 1
    ):
        lbl = (owner + "/" + repo) if owner and repo else url[:50]
        log.info("[ENRICH %d/%d] %s", i, len(rows), lbl)

        html = fetch_page(url)
        fit_text, gh_meta = "", None
        if html:
            fit_text = extract_readable(html, url)
            if is_gh:
                gh_meta = extract_gh_meta(url, html)
        elif is_gh and owner and repo:
            gh_meta = {"owner": owner, "repo": repo}

        if len(fit_text) < 20 and not gh_meta:
            stats["skipped"] += 1
            continue

        tout = 90 if (inn or 0) >= 8 else None
        raw, model = call_llm(
            build_prompt(url, fit_text, gh_meta, existing_sd=sd),
            timeout_override=tout,
        )
        if not raw:
            stats["failed"] += 1
            continue

        rdata = parse_llm_response(raw)
        if not rdata:
            stats["failed"] += 1
            continue

        garbage, reason = is_garbage(rdata)
        if garbage:
            stats["rejected"] += 1
            continue

        scores = compute_scores(rdata, is_gh, owner, pt, existing_innovation=inn or 0)
        nsd = stringify(rdata.get("SHORT_DESCRIPTION", ""))
        nld = stringify(rdata.get("LONG_DESCRIPTION", "")) or nsd
        nfeat = stringify(rdata.get("MAIN_FEATURES", ""))
        fsd = nsd if len(nsd) > len(sd or "") else (sd or nsd)

        c.execute(
            """UPDATE entries
            SET short_description=?, long_description=?, main_features=?,
                tags=?, innovation=?, quality=?, signal=?, is_standout=?, verdict=?
            WHERE id=?""",
            (
                fsd,
                nld,
                nfeat,
                json.dumps(scores["tags"]),
                scores["innovation"],
                scores["quality"],
                scores["signal"],
                scores["is_standout"],
                "",
                eid,
            ),
        )

        category = stringify(rdata.get("CATEGORY", ""))
        mapped = CAT_MAP.get(category, category)
        if mapped and mapped in BORG_TAXONOMY:
            c.execute("DELETE FROM layer_membership WHERE entry_id=?", (eid,))
            c.execute(
                "INSERT OR REPLACE INTO layer_membership (entry_id, layer, is_primary) VALUES (?, ?, 1)",
                (eid, mapped),
            )

        atl.commit()
        log.info("  ENRICHED: %s", nsd[:60])
        stats["enriched"] += 1
        time.sleep(0.2)

    atl.close()


def main():
    parser = argparse.ArgumentParser(description="Borg Deep Processor v4")
    parser.add_argument(
        "--phase",
        choices=["crawl", "ingest", "enrich", "all"],
        default="all",
    )
    parser.add_argument("--limit", type=int, default=0)
    args = parser.parse_args()

    log.info("Borg Deep Processor v4 - Starting")
    log.info("Phase: %s, Limit: %s", args.phase, args.limit or "unlimited")

    if args.phase in ("crawl", "all"):
        phase_crawl(limit=args.limit)
    if args.phase in ("ingest", "all"):
        phase_ingest(limit=args.limit)
    if args.phase in ("enrich", "all"):
        phase_enrich(limit=args.limit)

    log.info("=" * 60)
    log.info("DEEP PROCESSING COMPLETE")
    for k in stats:
        log.info("  %s: %d", k, stats[k])

    atl = sqlite3.connect(ATLAS_DB)
    c = atl.cursor()
    c.execute("SELECT COUNT(*) FROM entries")
    total = c.fetchone()[0]
    c.execute(
        "SELECT COUNT(*) FROM entries WHERE main_features IS NOT NULL AND main_features != "
        + EMPTY
    )
    wf = c.fetchone()[0]
    c.execute("SELECT COUNT(*) FROM entries WHERE signal >= 85")
    hs = c.fetchone()[0]
    c.execute("SELECT COUNT(*) FROM entries WHERE is_standout=1")
    st = c.fetchone()[0]
    pct = wf / total * 100 if total else 0
    log.info(
        "FINAL: %d entries | %d features (%.1f%%) | %d high-sig | %d standout",
        total,
        wf,
        pct,
        hs,
        st,
    )
    atl.close()


if __name__ == "__main__":
    main()
