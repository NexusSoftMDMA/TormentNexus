"""
TormentNexus Trends Analyzer
============================
Correlates bobbybookmarks data with assimilated MCP servers to identify:
- Trending topics from bookmark data (10K+ bookmarks)
- Gaps between what's trending vs what's implemented
- Hot features to implement next
- Areas for improvement
- Uses freellm for AI-powered analysis
"""

import json
import os
import re
import sqlite3
import time
import urllib.request
import urllib.error
from collections import Counter
from datetime import datetime
from pathlib import Path

WORKSPACE = Path(__file__).resolve().parent.parent
TRENDS_DB = WORKSPACE / "data" / "trends.db"
LOG_PATH = WORKSPACE / "data" / "trends_analysis.log"
ANALYSIS_INTERVAL = 21600  # 6 hours
LLM_PROXY = "http://localhost:4000/v1/chat/completions"

def log(msg, level="INFO"):
    ts = datetime.now().strftime("%H:%M:%S")
    line = f"[{ts}][{level}] {msg}"
    with open(LOG_PATH, "a", encoding="utf-8") as f:
        f.write(line + "\n")
    print(line)

def llm_call(prompt, max_tokens=500):
    payload = json.dumps({
        "model": "free-llm",
        "messages": [{"role": "user", "content": prompt}],
        "max_tokens": max_tokens,
        "temperature": 0.3,
    }).encode()
    req = urllib.request.Request(LLM_PROXY, data=payload, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=3600) as r:
            return json.loads(r.read())["choices"][0]["message"]["content"]
    except Exception as e:
        log(f"LLM error: {e}", "WARN")
        return None

def extract_topics():
    """Extract trending topics from bookmarks."""
    conn = sqlite3.connect(str(WORKSPACE / "tormentnexus.db"))
    urls = conn.execute("SELECT url, title, tags FROM bobby_bookmarks").fetchall()
    conn.close()
    
    domains = Counter()
    keywords = Counter()
    
    for url, title, tags in urls:
        # Domain analysis
        m = re.search(r'https?://([^/]+)', url or "")
        if m:
            domains[m.group(1)] += 1
        
        # Keyword extraction from title
        if title:
            words = re.findall(r'[A-Z][a-z]+|[A-Z]+(?=[A-Z]|$|[a-z])|\w+', title)
            for w in words:
                wl = w.lower()
                if len(wl) > 2 and wl not in ("the", "and", "for", "that", "this", "with", "from"):
                    keywords[wl] += 1
    
    return domains, keywords

def analyze_mcp_gaps():
    """Find gaps between trending topics and implemented MCP servers."""
    conn = sqlite3.connect(str(WORKSPACE / "data" / "assimilation_state.db"))
    
    # What's implemented
    implemented = conn.execute("SELECT name, github_url FROM mcp_servers WHERE status='implemented'").fetchall()
    failed = conn.execute("SELECT name, github_url FROM mcp_servers WHERE status='failed'").fetchall()
    
    # What categories exist
    cat_counts = conn.execute("SELECT COALESCE(notes,'unknown'), COUNT(*) FROM mcp_servers WHERE status='implemented' GROUP BY notes ORDER BY COUNT(*) DESC LIMIT 30").fetchall()
    
    with_github = conn.execute("SELECT COUNT(*) FROM mcp_servers WHERE status='implemented' AND github_url != ''").fetchone()[0]
    
    conn.close()
    return implemented, failed, cat_counts, with_github

def generate_report(domains, keywords, cat_counts, with_github):
    """Use freellm to generate an actionable trends report."""
    
    top_domains = "\n".join(f"  {d}: {c}" for d, c in domains.most_common(15))
    top_keywords = "\n".join(f"  {k}: {c}" for k, c in keywords.most_common(25))
    top_cats = "\n".join(f"  {c or 'uncategorized'}: {n}" for c, n in cat_counts[:15])
    
    prompt = f"""You are TormentNexus's trend analysis engine. Analyze this data and provide actionable insights.

=== TRENDING SOURCES (10,820 bookmarks) ===
Top domains by bookmark count:
{top_domains}

Top topics from bookmark titles:
{top_keywords}

=== IMPLEMENTED MCP SERVERS ===
Servers with GitHub URLs: {with_github}
Top categories implemented:
{top_cats}

=== TASK ===
Identify:
1. TOP 3 HOTTEST FEATURES to implement next — what's trending that we don't have yet
2. TOP 3 AREAS FOR IMPROVEMENT in current implementation
3. EMERGING TRENDS from the bookmark data (what are people bookmarking most?)
4. CORRELATION INSIGHTS — what do the bookmark trends tell us about MCP server demand?

Format your response as structured JSON with these keys:
- hot_features: [{{"name": "...", "reason": "...", "evidence": "..."}}]
- improvements: [{{"area": "...", "issue": "...", "suggestion": "..."}}]
- emerging_trends: [{{"trend": "...", "signal_strength": "high/medium/low", "description": "..."}}]
- correlations: [{{"insight": "...", "supporting_data": "..."}}]"""
    
    result = llm_call(prompt, max_tokens=1500)
    if result:
        # Extract JSON from response
        json_match = re.search(r'\{.*\}', result, re.DOTALL)
        if json_match:
            try:
                return json.loads(json_match.group())
            except json.JSONDecodeError:
                pass
        return {"raw": result}
    return {"error": "LLM call failed"}

def save_report(report):
    """Save trend report to trends.db for dashboard consumption."""
    os.makedirs(os.path.dirname(TRENDS_DB), exist_ok=True)
    conn = sqlite3.connect(str(TRENDS_DB))
    
    conn.execute("""
        CREATE TABLE IF NOT EXISTS trend_reports (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            generated_at TEXT NOT NULL,
            report_json TEXT NOT NULL,
            summary TEXT
        )
    """)
    
    conn.execute("""
        CREATE TABLE IF NOT EXISTS trending_keywords (
            keyword TEXT PRIMARY KEY,
            count INTEGER,
            last_seen TEXT
        )
    """)
    
    report_json = json.dumps(report, indent=2)
    summary = ""
    if "hot_features" in report:
        summary = "Hot: " + ", ".join(f["name"] for f in report["hot_features"][:3])
    
    conn.execute(
        "INSERT INTO trend_reports (generated_at, report_json, summary) VALUES (?, ?, ?)",
        (datetime.now().isoformat(), report_json, summary)
    )
    conn.commit()
    conn.close()

def main():
    log("=" * 60)
    log("TRENDS ANALYZER STARTED")
    log("Analyzing 10,820 bookmarks + MCP server data")
    log(f"Interval: {ANALYSIS_INTERVAL}s")
    log("=" * 60)
    
    while True:
        start = time.time()
        log(f"--- Analysis cycle {datetime.now().isoformat()} ---")
        
        # Step 1: Extract topics from bookmarks
        domains, keywords = extract_topics()
        log(f"Extracted {len(domains)} domains, {len(keywords)} keywords")
        
        # Step 2: Analyze MCP gaps
        implemented, failed, cat_counts, with_github = analyze_mcp_gaps()
        log(f"MCP servers: {len(implemented)} implemented, {len(failed)} failed, {with_github} with GitHub URLs")
        
        # Step 3: Generate AI report
        log("Generating trend report via freellm...")
        report = generate_report(domains, keywords, cat_counts, with_github)
        
        # Step 4: Save report
        save_report(report)
        log(f"Report saved to {TRENDS_DB}")
        
        if "hot_features" in report:
            log(f"Top features: {[f['name'] for f in report['hot_features'][:3]]}")
        if "improvements" in report:
            log(f"Improvements: {[i['area'] for i in report['improvements'][:3]]}")
        if "emerging_trends" in report:
            log(f"Trends: {[t['trend'] for t in report['emerging_trends'][:3]]}")
        if "raw" in report:
            log(f"Raw report: {report['raw'][:200]}")
        
        elapsed = time.time() - start
        log(f"Cycle complete: {elapsed:.1f}s")
        log(f"Next analysis in {ANALYSIS_INTERVAL}s")
        log("")
        
        time.sleep(ANALYSIS_INTERVAL)

if __name__ == "__main__":
    main()
