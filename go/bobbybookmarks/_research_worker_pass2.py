#!/usr/bin/env python3
"""Borg Research Worker v2.6 - Enrich remaining entries, including fetch-failed ones using metadata"""
import os
import re
import json
import sqlite3
import requests
import time
import logging
import sys
from bs4 import BeautifulSoup, Comment
sys.stdout.reconfigure(encoding='utf-8')

os.makedirs('logs', exist_ok=True)
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(message)s',
    handlers=[
        logging.FileHandler('logs/research_run2.log', mode='a', encoding='utf-8'),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger(__name__)

ATLAS_DB = 'atlas.db'
LMSTUDIO_URL = "http://localhost:1234/v1/chat/completions"

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

FIELD_NAMES = ['CATEGORY', 'SHORT_DESCRIPTION', 'LONG_DESCRIPTION',
               'MAIN_FEATURES', 'INNOVATION_SCORE', 'TAGS']

MODELS = [
    ("gemma-4-e4b-it-qat-unquantized-heretic", 150),  # Smallest loaded model, needs ~120s
]


def stringify(v):
    if v is None:
        return ''
    if isinstance(v, str):
        return v
    if isinstance(v, (list, tuple)):
        return ', '.join(str(x) for x in v)
    return str(v)


def call_llm(prompt):
    for model, tout in MODELS:
        try:
            resp = requests.post(LMSTUDIO_URL, json={
                "model": model,
                "messages": [{"role": "user", "content": prompt}],
                "temperature": 0.2,
                "max_tokens": 350,
            }, timeout=tout)
            if resp.status_code == 200:
                data = resp.json()
                text = data.get('choices', [{}])[0].get('message', {}).get('content', '')
                if text and len(text) > 30:
                    return text, model
        except requests.exceptions.Timeout:
            logger.warning(f"  Timeout ({tout}s) on {model}")
        except Exception as e:
            logger.warning(f"  Error on {model}: {str(e)[:60]}")
        time.sleep(1)
    return None, None


def extract_fields_from_jsonish(text):
    result = {}
    for field in FIELD_NAMES:
        m = re.search(rf'"{field}"\s*:\s*(\d+)', text)
        if m:
            result[field] = int(m.group(1))
            continue
        m = re.search(rf'"{field}"\s*:\s*"', text)
        if m:
            start = m.end()
            remaining = text[start:]
            next_pos = len(remaining)
            for other in FIELD_NAMES:
                if other == field:
                    continue
                pos = remaining.find(f'"{other}"')
                if pos > 0 and pos < next_pos:
                    next_pos = pos
            val = remaining[:next_pos].strip()
            val = val.rstrip(', \n\t}')
            val = val.replace('"', '')
            val = val.replace('\n', ' ')
            val = re.sub(r'\s*,\s*', ', ', val)
            val = val.strip().rstrip(',').strip()
            result[field] = val
    return result if len(result) >= 3 else None


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
                except Exception:
                    pass
                fixed = re.sub(r',\s*([}\]])', r'\1', block)
                try:
                    return json.loads(fixed)
                except Exception:
                    pass
                result = extract_fields_from_jsonish(fixed)
                if result:
                    return result

    # 3. Balanced braces
    start = text.find('{')
    if start >= 0:
        depth = 0
        for i in range(start, len(text)):
            if text[i] == '{':
                depth += 1
            elif text[i] == '}':
                depth -= 1
            if depth == 0:
                candidate = text[start:i + 1]
                try:
                    return json.loads(candidate)
                except Exception:
                    result = extract_fields_from_jsonish(candidate)
                    if result:
                        return result
                    break

    # 4. Markdown / plain-text key-value
    result = {}
    norm = text.replace('\\_', '_')
    stripped = re.sub(
        r'^.*?(?=\n[-*\s]*(?:\*\*)?(?:Resource\s+)?(?:CATEGORY|Classification|SHORT_DESCRIPTION|LONG_DESCRIPTION|MAIN_FEATURES))',
        '', norm, flags=re.DOTALL | re.IGNORECASE
    )
    if not re.search(r'CATEGORY|Classification|SHORT_DESC|LONG_DESC|MAIN_FEATURE', stripped, re.IGNORECASE):
        stripped = norm

    # Classification patterns
    for pat in [
        r'\*\*Resource\s+Classification:\*\*\s*\*\*([^*]+)\*\*',
        r'\*\*Primary\s+Classification:\*\*\s*\*\*([^*]+)\*\*',
    ]:
        cls = re.search(pat, stripped, re.IGNORECASE)
        if cls:
            result['CATEGORY'] = cls.group(1).strip()
            break

    kv_patterns = {
        'CATEGORY': r'(?:\*\*)?(?:Resource\s+)?CATEGORY(?:\*\*)?[\s:]*([^\n]+)',
        'SHORT_DESCRIPTION': r'(?:\*\*)?SHORT_DESCRIPTION(?:\*\*)?[\s:]*([^\n]+(?:\n(?!\*\*[A-Z]|[-*]\s*[A-Z_]+:)[^\n]+)*)',
        'LONG_DESCRIPTION': r'(?:\*\*)?LONG_DESCRIPTION(?:\*\*)?[\s:]*([^\n]+(?:\n(?!\*\*[A-Z]|[-*]\s*[A-Z_]+:)[^\n]+)*)',
        'MAIN_FEATURES': r'(?:\*\*)?MAIN_FEATURES(?:\*\*)?[\s:]*((?:[^\n]+|\n[-*]\s+[^\n]+)*)',
        'INNOVATION_SCORE': r'(?:\*\*)?INNOVATION_SCORE(?:\*\*)?[\s:]*(\d+)',
        'TAGS': r'(?:\*\*)?TAGS(?:\*\*)?[\s:]*((?:[^\n]+|\n[-*]\s+[^\n]+)*)',
    }
    for key, pat in kv_patterns.items():
        if key in result:
            continue
        m = re.search(pat, stripped, re.IGNORECASE | re.DOTALL)
        if m:
            val = m.group(1).strip()
            val = re.sub(r'^[\s:*-]+', '', val).strip()
            val = re.sub(r'\n[-*]\s+', ', ', val)
            val = re.sub(r'\n+', ' ', val)
            val = val.replace('**', '')
            if key == 'INNOVATION_SCORE':
                try:
                    result[key] = int(val)
                except (ValueError, TypeError):
                    result[key] = 8
            else:
                result[key] = val

    if 'CATEGORY' not in result:
        # Check **Resource Classification:** header style
        cls = re.search(r'\*\*Primary\s+Classification:\*\*\s*\*\*([^*]+)\*\*', stripped, re.IGNORECASE)
        if cls:
            result['CATEGORY'] = cls.group(1).strip()
        else:
            for cat in BORG_TAXONOMY:
                if cat.lower() in stripped.lower():
                    result['CATEGORY'] = cat
                    break

    if 'SHORT_DESCRIPTION' not in result:
        desc_match = re.search(
            r'(?:Key\s+Functionality|Detailed\s+Analysis|Project\s+Focus|Core\s+function)[\s:*]+(.+?)(?:\n\n|\n\*\*)',
            stripped, re.IGNORECASE | re.DOTALL
        )
        if desc_match:
            result['SHORT_DESCRIPTION'] = desc_match.group(1).strip().replace('**', '')[:200]

    if len(result) >= 3:
        result.setdefault('CATEGORY', 'Guides & Industry Trends')
        result.setdefault('SHORT_DESCRIPTION', result.get('LONG_DESCRIPTION', '')[:100])
        result.setdefault('LONG_DESCRIPTION', result.get('SHORT_DESCRIPTION', ''))
        result.setdefault('MAIN_FEATURES', '')
        result.setdefault('TAGS', '')
        result.setdefault('INNOVATION_SCORE', 8)
        return result

    # 5. Last resort
    return extract_fields_from_jsonish(text)


def fetch_content(url):
    headers = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'}
    try:
        resp = requests.get(url, headers=headers, timeout=12, allow_redirects=True)
        if resp.status_code == 200:
            return resp.text
    except Exception:
        pass
    return None


def extract_fit_markdown(html, url=""):
    soup = BeautifulSoup(html, 'html.parser')
    for tag in soup.find_all(['script', 'style', 'nav', 'footer', 'header', 'aside',
                               'iframe', 'noscript', 'svg', 'button', 'form']):
        tag.decompose()
    for comment in soup.find_all(string=lambda t: isinstance(t, Comment)):
        comment.extract()
    main = None
    for sel in ['main', 'article', '[role="main"]', '.content', '#content',
                '.post-body', '.markdown-body', '.readme', '#readme', '.entry-content']:
        main = soup.select_one(sel)
        if main:
            break
    if not main:
        main = soup.find('body') or soup
    text = main.get_text(separator='\n', strip=True)
    lines = [l.strip() for l in text.splitlines() if l.strip()]
    text = '\n'.join(lines)
    return text[:2000]  # Reduced from 3000 to help with GPU memory pressure


def extract_gh_meta(url, html):
    meta = {}
    soup = BeautifulSoup(html, 'html.parser')
    about = soup.find('p', class_='f4 my-3')
    if about:
        meta['desc'] = about.get_text(strip=True)
    topics = soup.find_all('a', class_='topic-tag')
    if topics:
        meta['topics'] = [t.get_text(strip=True) for t in topics[:8]]
    readme = soup.find('div', class_='markdown-body')
    if readme:
        meta['readme'] = readme.get_text(separator='\n', strip=True)[:2000]
    return meta


def extract_reddit_context(url):
    """Extract subreddit and title from Reddit URL for context."""
    ctx = {}
    sub_m = re.search(r'reddit\.com/r/(\w+)/', url, re.IGNORECASE)
    if sub_m:
        ctx['subreddit'] = sub_m.group(1)
    title_m = re.search(r'reddit\.com/r/\w+/comments/\w+/([^/?]+)/?', url, re.IGNORECASE)
    if title_m:
        ctx['title'] = title_m.group(1).replace('_', ' ').replace('-', ' ').title()
    return ctx


def build_prompt(url, fit_text, gh_meta=None, reddit_ctx=None, existing_sd=None):
    """Build prompt with available context, even if fetch failed."""
    prompt = "Classify this resource. URL: " + url + "\n"

    if gh_meta:
        if 'desc' in gh_meta:
            prompt += "Repo: " + gh_meta['desc'] + "\n"
        if 'topics' in gh_meta:
            prompt += "Topics: " + ", ".join(gh_meta['topics']) + "\n"

    if reddit_ctx:
        if 'subreddit' in reddit_ctx:
            prompt += "Subreddit: r/" + reddit_ctx['subreddit'] + "\n"
        if 'title' in reddit_ctx:
            prompt += "Post title: " + reddit_ctx['title'] + "\n"

    if existing_sd and len(existing_sd) > 20:
        prompt += "Known description: " + existing_sd + "\n"

    if fit_text and len(fit_text) > 50:
        prompt += "\nContent:\n" + fit_text + "\n\n"
    else:
        prompt += "\n(Note: Page content could not be fetched. Classify based on URL and metadata above.)\n\n"

    prompt += "Categories: " + ", ".join(BORG_TAXONOMY) + "\n\n"
    prompt += (
        "Return JSON with these fields:\n"
        "- CATEGORY: one of the above categories\n"
        "- SHORT_DESCRIPTION: 1 specific sentence\n"
        "- LONG_DESCRIPTION: 2-3 detailed sentences\n"
        "- MAIN_FEATURES: 3-5 specific features (comma separated)\n"
        "- INNOVATION_SCORE: 1-10\n"
        "- TAGS: 8-12 lowercase tags (comma separated)\n"
    )
    return prompt


def is_garbage(rdata):
    desc = stringify(rdata.get('SHORT_DESCRIPTION', '')).lower().strip()
    feats = stringify(rdata.get('MAIN_FEATURES', '')).lower().strip()
    if not feats or len(feats) < 10:
        return True, "empty_features"
    if not desc or len(desc) < 15:
        return True, "empty_description"
    for p in ['automated discovery', 'heuristic mapping', 'unable to determine']:
        if p in feats:
            return True, f"boilerplate:{p[:30]}"
    for p in ['sign in to continue', 'a comprehensive resource detailing']:
        if p in desc:
            return True, f"generic_desc:{p[:30]}"
    try:
        innov = int(rdata.get('INNOVATION_SCORE', 0))
        if innov <= 2 and len(feats) < 50:
            return True, f"low_innov:{innov}"
    except (ValueError, TypeError):
        pass
    return False, None


def reclassify_entry(a, eid, mapped):
    a.execute("SELECT layer FROM layer_membership WHERE entry_id=? AND is_primary=1", (eid,))
    current = a.fetchone()
    if not current or current[0] == mapped:
        return
    old_layer = current[0]
    a.execute("SELECT COUNT(*) FROM layer_membership WHERE entry_id=? AND layer=?", (eid, mapped))
    if a.fetchone()[0] > 0:
        a.execute("DELETE FROM layer_membership WHERE entry_id=? AND is_primary=1", (eid,))
        a.execute("UPDATE layer_membership SET is_primary=1 WHERE entry_id=? AND layer=?", (eid, mapped))
    else:
        a.execute("UPDATE layer_membership SET layer=? WHERE entry_id=? AND is_primary=1", (mapped, eid))
    logger.info(f"  Reclassified: {old_layer} -> {mapped}")


def main():
    atl = sqlite3.connect(ATLAS_DB)
    a = atl.cursor()

    # Find entries that need enrichment
    a.execute("""SELECT e.id, e.url, e.short_description, e.is_github, e.innovation,
        e.page_title, e.owner, e.repo
        FROM entries e
        WHERE (e.long_description = e.short_description OR LENGTH(e.long_description) < 50)
        ORDER BY e.is_github DESC, e.id DESC""")
    all_entries = a.fetchall()

    logger.info("Borg Research Worker v2.6 starting")
    logger.info(f"Entries to research: {len(all_entries):,}")

    accepted = 0
    rejected = 0
    failed = 0
    skipped = 0
    metadata_only = 0  # Enriched using metadata only (no page fetch)

    for idx, (eid, url, sd, is_gh, innov, pt, owner, repo) in enumerate(all_entries):
        logger.info(f"[{idx+1}/{len(all_entries)}] Researching: {url[:80]}")

        # Try fetching content first
        content = fetch_content(url)
        fit_text = ""
        gh_meta = None
        reddit_ctx = None
        used_metadata_only = False

        if content:
            fit_text = extract_fit_markdown(content, url)
            if is_gh:
                gh_meta = extract_gh_meta(url, content)
        else:
            # Fetch failed - use URL structure and existing metadata as context
            if 'reddit.com' in url.lower():
                reddit_ctx = extract_reddit_context(url)
                used_metadata_only = True
            elif is_gh and owner and repo:
                # GitHub fetch failed, but we have owner/repo
                gh_meta = {'desc': f"GitHub repository {owner}/{repo}"}
                used_metadata_only = True
            else:
                # Try with just URL and existing short_description
                used_metadata_only = True

        # If fit_text is too thin, we can still proceed with metadata
        if len(fit_text) < 50 and not used_metadata_only:
            skipped += 1
            continue

        prompt = build_prompt(url, fit_text, gh_meta, reddit_ctx, existing_sd=sd)
        raw, model = call_llm(prompt)

        if not raw:
            logger.warning("  LLM failed (all models)")
            failed += 1
            continue

        rdata = parse_llm_response(raw)
        if not rdata:
            dump_path = os.path.join('logs', 'parse_failures.jsonl')
            with open(dump_path, 'a', encoding='utf-8') as df:
                df.write(json.dumps({"url": url, "model": model, "raw": raw[:2000]}) + '\n')
            logger.warning(f"  Parse failed (model={model})")
            failed += 1
            continue

        garbage, reason = is_garbage(rdata)
        if garbage:
            rejected += 1
            continue

        category = stringify(rdata.get('CATEGORY', '')).strip()
        short_desc = stringify(rdata.get('SHORT_DESCRIPTION', '')).strip()
        long_desc = stringify(rdata.get('LONG_DESCRIPTION', '')).strip() or short_desc
        features = stringify(rdata.get('MAIN_FEATURES', '')).strip()
        tags_str = stringify(rdata.get('TAGS', '')).strip()
        innovation = rdata.get('INNOVATION_SCORE', 8)
        try:
            innovation = int(innovation)
        except (ValueError, TypeError):
            innovation = 8
        innovation = max(1, min(10, innovation))

        # Normalize tags
        tags = [t.strip().lower().replace(' ', '-').replace('_', '-')
                for t in tags_str.split(',') if t.strip()]
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

        feat_count = len([x.strip() for x in features.split(',')
                         if x.strip() and len(x.strip()) > 3])
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

        if pt and len(pt) > 5: score += 10
        if owner and len(owner) > 1: score += 10

        quality = min(1.0, score / 100)
        new_innov = max(innov, innovation)

        feat_score = min(15, feat_count * 3)
        desc_score = min(10, len(long_desc) / 50.0)
        gh_bonus = 5 if is_gh else 0
        signal = min(100, max(0, int(round(
            (new_innov * 4) + (quality * 30) + feat_score + desc_score + gh_bonus
        ))))
        is_standout = 1 if new_innov >= 9 and quality >= 0.8 else 0

        # Use the better short_description
        final_sd = short_desc if len(short_desc) > len(sd) else sd

        a.execute("""UPDATE entries SET
            short_description=?, long_description=?, main_features=?,
            tags=?, innovation=?, quality=?, signal=?, is_standout=?, verdict=?
            WHERE id=?""",
            (final_sd, long_desc, features, json.dumps(clean_tags),
             new_innov, quality, signal, is_standout, '', eid))

        mapped = CAT_MAP.get(category, category)
        if mapped and mapped in BORG_TAXONOMY:
            reclassify_entry(a, eid, mapped)

        atl.commit()
        accepted += 1
        if used_metadata_only:
            metadata_only += 1
        mshort = model[:15] if model else '?'
        mode = "META" if used_metadata_only else "FETCH"
        logger.info(f"  OK [{mshort}] [{mode}] I{new_innov} Q{quality:.2f} S{signal}: {final_sd[:60]}")

        time.sleep(0.5)

    logger.info("=" * 60)
    logger.info(f"Research complete: {accepted} enriched ({metadata_only} metadata-only), "
                f"{rejected} rejected, {failed} failed, {skipped} skipped")
    logger.info("=" * 60)

    a.execute("SELECT COUNT(*) FROM entries")
    total = a.fetchone()[0]
    a.execute("SELECT COUNT(*) FROM entries WHERE is_standout=1")
    standout = a.fetchone()[0]
    a.execute("SELECT COUNT(*) FROM entries WHERE LENGTH(long_description) > 50 AND long_description != short_description")
    enriched = a.fetchone()[0]
    logger.info(f"Atlas: {total:,} entries, {enriched:,} enriched, {standout:,} standout")

    atl.close()


if __name__ == "__main__":
    main()
