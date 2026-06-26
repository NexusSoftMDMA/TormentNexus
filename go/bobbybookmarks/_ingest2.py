#!/usr/bin/env python3
"""Ingest new URLs from incoming_resources.txt into atlas.db - v2 with smarter filtering"""
import sqlite3, sys, re, json
sys.stdout.reconfigure(encoding='utf-8')

DB = 'atlas.db'
atl = sqlite3.connect(DB)
a = atl.cursor()

# Build existing URL set
a.execute('SELECT url FROM entries')
existing = set()
for (u,) in a.fetchall():
    nu = u.lower().strip().rstrip('/').replace('http://','https://').replace('www.','')
    if '#' in nu: nu = nu.split('#')[0]
    if nu.endswith('/'): nu = nu[:-1]
    existing.add(nu)

a.execute("SELECT MAX(id) FROM entries")
max_id = a.fetchone()[0] or 0

def norm_url(u):
    nu = u.strip().rstrip('/')
    nu = nu.replace('http://','https://')
    if nu.endswith('/'): nu = nu[:-1]
    return nu

# Load and dedupe incoming
with open('incoming_resources.txt', 'r') as f:
    raw = [line.strip() for line in f if line.strip()]

seen = set()
unique = []
for u in raw:
    n = norm_url(u).lower().replace('www.','')
    if '#' in n: n = n.split('#')[0]
    if n not in seen:
        seen.add(n)
        unique.append(u)

# Smart filtering
HARD_REJECT_SUBS = {
    'ufo', 'ufob', 'ufos', 'businessowners', 'productivitycafe', 'work',
    'domainflipping', 'wtfisai', 'darkpsychology101', 'oddlysatisfying',
    'funny', 'pics', 'askreddit', 'todayilearned', 'showerthoughts',
    'explainlikeimfive', 'lifeprotips', 'dataisbeautiful', 'diwhy',
    'science', 'worldnews', 'politics', 'conspiracy', 'television',
    'gaming', 'movies', 'books', 'music', 'food', 'travel', 'fitness',
    'sports', 'relationship_advice', 'aihub',
}

TOOL_SUBS = {
    'mcp', 'mcpServers', 'localalla', 'claudeai', 'chatgpt', 'openai',
    'anthropicai', 'llmdevs', 'accelerate', 'promptengineering',
    'cursor', 'cline', 'githubcopilot', 'devops', 'programming',
    'kiroide', 'claudeoctopus', 'conductorbuild', 'geminicli',
    'localllama', 'aicode', 'aiagents', 'agentframework',
    'codingagent', 'aiprogramming', 'llmcoding', 'netsec',
    'singularity', 'machinelearning', 'deeplearning', 'artificial',
    'stablediffusion', 'locallyllama', 'chatgptpro', 'hermesagent',
    'vibecoding', 'sysadmin', 'patterns', 'designtecture', 'bbs',
    'claudeplayspokemon', 'ai_trading', 'trading',
}

TOOL_INDICATORS = [
    'mcp', 'agent', 'tool', 'ai', 'llm', 'model', 'coding',
    'prompt', 'framework', 'server', 'open source', 'github', 'api',
    'sdk', 'library', 'security', 'hack', 'vulnerab', 'deploy', 'automat',
    'claude', 'gpt', 'build', 'memory', 'workflow', 'autonomous',
    'persistent memory', 'local model', 'provider', 'tokens',
    'algorithm', 'sovereign', 'os-level', 'neon pulse',
]

HARD_REJECT_DOMAINS = {'temu.com', 'pharmacy2home.com'}

HARD_REJECT_PATTERNS = [
    r'^https?://(www\.)?reddit\.com/?$',
    r'google\.com/search\?',
]

CATEGORY_KEYWORDS = {
    'Agent Orchestration & Workflow': [
        'agent', 'orchestrat', 'workflow', 'langgraph', 'crewai', 'autogen',
        'agentic', 'multi-agent', 'pipeline', 'hermes',
    ],
    'Context Engineering & Isolation': [
        'rag', 'context', 'embed', 'chunk', 'retrieval', 'prompt', 'template',
        'isolat', 'contextium',
    ],
    'Memory & Persistence Architecture': [
        'memory', 'persist', 'graph', 'knowledge', 'recall', 'mem0', 'memgpt', 'cache',
    ],
    'Interface & Developer UX': [
        'browser', 'desktop', 'gui', 'chat', 'interface', 'dashboard', 'ux',
        'voice', 'terminal', 'cli', 'waterfox',
    ],
    'Connectivity / MCP / A2A': [
        'mcp', 'model context protocol', 'a2a', 'connect', 'interoperab',
        'bridge', 'gateway', 'proxy', 'adapter',
    ],
    'Infrastructure & Proxy Layers': [
        'infra', 'deploy', 'cloud', 'docker', 'kubernetes', 'server', 'runtime',
        'sandbox', 'inference', 'ollama', 'database', 'anchorgrid',
    ],
    'Guides & Industry Trends': [
        'guide', 'tutorial', 'awesome', 'benchmark', 'trend', 'report',
        'survey', 'analysis', 'strategy', 'lesswrong', 'ycombinator', 'wsj',
        'chinatalk', 'arxiv', 'sigasi', 'briancarpio', 'movementparties',
    ],
    'Coding Harness Tools': [
        'harness', 'codex', 'claude code', 'skill', 'hook', 'governance',
        'task-master', 'taskmaster', 'coding agent', 'opencode', 'codingfont',
    ],
    'AI Agents & Frameworks': [
        'ai agent', 'framework', 'autonomous', 'manus', 'devin', 'swe-agent',
        'research agent', 'browser agent',
    ],
    'Search & Discovery': [
        'search', 'tavily', 'serp', 'brave search', 'code search', 'scrape',
        'crawl', 'fetch', 'discovery',
    ],
    'Coding Tools & IDEs': [
        'copilot', 'autocomplete', 'ide', 'editor', 'kiro', 'cursor', 'cline',
        'jetbrains', 'vscode', 'gemini cli', 'vibecoding', 'cherri',
    ],
    'Developer Workflow & Tools': [
        'git', 'github', 'ci/cd', 'pipeline', 'deploy', 'documentation',
        'testing', 'monitor', 'devops', 'sysadmin',
    ],
    'Vector Databases & Embeddings': [
        'vector', 'embedding', 'pgvector', 'chromadb', 'pinecone', 'qdrant',
        'weaviate',
    ],
    'Security & Red Teaming': [
        'security', 'vulnerab', 'guardrail', 'firewall', 'audit', 'scan',
        'pentest', 'threat', 'red team', 'hack', 'netsec',
    ],
}

SUBCAT_MAP = {
    'Agent Orchestration & Workflow': {
        'Agentic Frameworks': ['agent', 'framework', 'orchestrat', 'agentic', 'multi-agent', 'hermes'],
        'Workflow & Pipeline Engines': ['workflow', 'pipeline', 'n8n', 'temporal'],
        'Tool Discovery & Routing': ['tool', 'mcp', 'bridge', 'gateway', 'routing'],
        'MCP Servers': ['mcp server', 'mcp-server'],
    },
    'Context Engineering & Isolation': {
        'RAG Pipelines': ['rag', 'retrieval', 'document'],
        'Prompt Engineering': ['prompt', 'template'],
        'Embedding & Indexing': ['embed', 'index', 'semantic'],
        'MCP Servers': ['mcp server', 'mcp-server'],
    },
    'Memory & Persistence Architecture': {
        'Memory Systems': ['memory', 'mem0', 'memgpt', 'recall', 'knowledge graph'],
        'Storage Engines': ['storage', 'persist', 'cache'],
        'MCP Servers': ['mcp server', 'mcp-server'],
    },
    'Interface & Developer UX': {
        'Chat Interfaces': ['chat', 'assistant'],
        'Browser Automation': ['browser', 'puppeteer', 'playwright'],
        'Terminal & CLI': ['terminal', 'cli', 'command'],
        'MCP Servers': ['mcp server', 'mcp-server'],
    },
    'Connectivity / MCP / A2A': {
        'MCP Servers': ['mcp server', 'mcp-server', 'mcp'],
        'MCP Clients': ['mcp client', 'mcp registry'],
        'Protocol Bridges': ['bridge', 'gateway', 'proxy', 'adapter', 'a2a'],
        'API Integrations': ['api', 'rest', 'webhook', 'integration'],
    },
    'Infrastructure & Proxy Layers': {
        'Model Serving': ['inference', 'serving', 'vllm', 'ollama'],
        'Cloud Infrastructure': ['cloud', 'aws', 'gcp', 'docker', 'kubernetes'],
        'Database Connectors': ['database', 'sql', 'postgres', 'redis'],
        'Sandboxing & Execution': ['sandbox', 'runtime', 'execution'],
    },
    'Guides & Industry Trends': {
        'Tutorials & Courses': ['tutorial', 'course', 'guide', 'learn'],
        'Industry Analysis': ['trend', 'report', 'survey', 'landscape', 'wsj'],
        'Curated Lists': ['awesome', 'curated', 'collection'],
    },
    'Coding Harness Tools': {
        'Spec-Driven Development': ['spec', 'harness', 'governance'],
        'Coding Agents': ['coding agent', 'codex', 'claude code'],
        'Task Management': ['task', 'task-master', 'taskmaster'],
        'MCP Servers': ['mcp server', 'mcp-server'],
    },
    'AI Agents & Frameworks': {
        'Autonomous Agents': ['autonomous', 'agent', 'manus', 'devin'],
        'Research Agents': ['research agent', 'deep research'],
        'GUI Agents': ['gui agent', 'computer use'],
    },
    'Search & Discovery': {
        'Web Search': ['web search', 'tavily', 'serp'],
        'Code Search': ['code search', 'codebase search'],
        'Web Scraping': ['scrape', 'crawl', 'fetch'],
        'MCP Servers': ['mcp server', 'mcp-server'],
    },
    'Coding Tools & IDEs': {
        'AI Code Completion': ['copilot', 'autocomplete', 'suggest'],
        'IDE Extensions': ['vscode', 'extension', 'jetbrains', 'kiro', 'cursor', 'cline'],
        'AI Editors & IDEs': ['ide', 'editor', 'gemini cli', 'vibe'],
    },
    'Developer Workflow & Tools': {
        'CI/CD & Automation': ['ci/cd', 'pipeline', 'deploy', 'github actions'],
        'Git & Version Control': ['git', 'github', 'version control'],
        'Documentation': ['document', 'readme', 'swagger'],
        'Monitoring & Observability': ['monitor', 'observab', 'sentry'],
    },
    'Vector Databases & Embeddings': {
        'Vector Databases': ['vector', 'pgvector', 'chromadb', 'pinecone', 'qdrant'],
        'Embedding Models': ['embed', 'sentence-transformer'],
    },
    'Security & Red Teaming': {
        'AI Safety': ['safety', 'alignment', 'guardrail'],
        'Red Teaming': ['red team', 'pentest', 'adversarial', 'jailbreak'],
        'Vulnerability & Scanning': ['security', 'vulnerab', 'scan', 'hack'],
    },
}


def classify_url(url, title=''):
    text = f"{url} {title}".lower()
    scores = {}
    for layer, kws in CATEGORY_KEYWORDS.items():
        scores[layer] = sum(1 if kw in text else 0 for kw in kws)
    if max(scores.values()) > 0:
        best_layer = max(scores, key=scores.get)
    else:
        best_layer = 'Guides & Industry Trends'

    subcats = SUBCAT_MAP.get(best_layer, {})
    best_sub = 'Unclassified'
    best_ss = 0
    for sub, kws in subcats.items():
        s = sum(1 if kw in text else 0 for kw in kws)
        if s > best_ss:
            best_ss = s
            best_sub = sub
    return best_layer, best_sub


def extract_reddit_title(url):
    m = re.search(r'reddit\.com/r/\w+/comments/\w+/([^/?]+)/?', url, re.IGNORECASE)
    if m:
        return m.group(1).replace('_', ' ').replace('-', ' ').title()
    return None


def extract_gh(url):
    m = re.match(r'https?://github\.com/([^/]+)/([^/\?#]+)', url, re.IGNORECASE)
    if m:
        owner, repo = m.group(1), m.group(2)
        if repo.endswith('.git'):
            repo = repo[:-1]
        if repo.lower() in ('issues', 'pulls', 'releases', 'wiki', 'actions', 'blob', 'tree', 'commit'):
            return None, None
        return owner, repo
    return None, None


inserted = 0
skipped_noise = 0
skipped_dup = 0

for url in unique:
    nurl = norm_url(url).lower().replace('www.', '')
    if '#' in nurl:
        nurl = nurl.split('#')[0]

    if nurl in existing:
        skipped_dup += 1
        continue

    # Hard reject patterns
    skip = False
    for pat in HARD_REJECT_PATTERNS:
        if re.search(pat, url, re.IGNORECASE):
            skip = True
            break
    if skip:
        skipped_noise += 1
        continue

    # Hard reject domains
    parsed = re.search(r'://(?:www\.)?([^/]+)', url)
    domain = parsed.group(1).lower() if parsed else ''
    if any(d in domain for d in HARD_REJECT_DOMAINS):
        skipped_noise += 1
        continue

    # Reddit filtering
    reddit_m = re.search(r'reddit\.com/r/(\w+)/', url, re.IGNORECASE)
    if reddit_m:
        sub = reddit_m.group(1).lower()
        if sub in HARD_REJECT_SUBS:
            skipped_noise += 1
            continue
        if sub not in TOOL_SUBS:
            title_check = extract_reddit_title(url) or ''
            if not any(ind in title_check.lower() for ind in TOOL_INDICATORS):
                skipped_noise += 1
                continue

    # Parse metadata
    title = ''
    owner, repo, is_gh = None, None, 0
    sd = ''

    if 'github.com' in url.lower():
        owner, repo = extract_gh(url)
        if owner and repo:
            is_gh = 1
            title = f"{owner}/{repo}"
            sd = f"GitHub repository: {owner}/{repo}"
        else:
            title = url[:70]
            sd = url
    elif 'reddit.com' in url.lower():
        title = extract_reddit_title(url) or 'Reddit Discussion'
        sd = title
    elif 'news.ycombinator.com' in url.lower():
        title = 'Hacker News Discussion'
        sd = f'HN: {url.split("id=")[-1] if "id=" in url else ""}'
    elif 'x.com' in url.lower() or 'twitter.com' in url.lower():
        title = 'X/Twitter Post'
        sd = url
    elif 'arxiv.org' in url.lower():
        title = 'arXiv Paper'
        sd = f'arXiv: {url.split("/")[-1]}'
    elif 'lesswrong.com' in url.lower():
        title = 'LessWrong Post'
        sd = url
    elif 'youtube.com' in url.lower() or 'youtu.be' in url.lower():
        title = 'YouTube Video'
        sd = url
    else:
        m = re.search(r'://(?:www\.)?([^/]+)', url)
        title = m.group(1) if m else url[:70]
        sd = title

    # Classify
    layer, subcat = classify_url(url, title)

    # Score
    if is_gh:
        innovation, quality = 9, 0.7
    elif 'reddit.com' in url.lower():
        innovation, quality = 7, 0.5
    elif 'news.ycombinator' in url.lower():
        innovation, quality = 7, 0.5
    elif 'arxiv.org' in url.lower():
        innovation, quality = 8, 0.6
    elif 'lesswrong.com' in url.lower():
        innovation, quality = 7, 0.5
    elif 'youtube.com' in url.lower():
        innovation, quality = 6, 0.4
    else:
        innovation, quality = 8, 0.6

    desc_score = min(10, len(sd) / 50.0)
    gh_bonus = 5 if is_gh else 0
    signal = min(100, max(0, int(round((innovation * 4) + (quality * 30) + desc_score + gh_bonus))))
    is_standout = 1 if innovation >= 9 and quality >= 0.8 else 0

    max_id += 1
    eid = max_id

    a.execute(
        """INSERT INTO entries (id, url, page_title, short_description, long_description,
           main_features, tags, owner, repo, is_github, innovation, quality, signal, is_standout, verdict)
           VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)""",
        (eid, url, title, sd, sd, '', '[]', owner, repo, is_gh,
         innovation, quality, signal, is_standout, ''),
    )

    # Handle UNIQUE constraint on layer_membership
    a.execute(
        "SELECT COUNT(*) FROM layer_membership WHERE entry_id=? AND layer=?",
        (eid, layer),
    )
    if a.fetchone()[0] == 0:
        a.execute(
            "INSERT INTO layer_membership (entry_id, layer, subcategory, is_primary, match_score) VALUES (?,?,?,?,?)",
            (eid, layer, subcat, 1, 3),
        )

    existing.add(nurl)
    inserted += 1

atl.commit()

a.execute("SELECT COUNT(*) FROM entries")
total = a.fetchone()[0]

print(f"Incoming: {len(unique):,} unique URLs")
print(f"  Skipped (duplicate): {skipped_dup:,}")
print(f"  Skipped (noise): {skipped_noise:,}")
print(f"  Inserted: {inserted:,}")
print(f"\nAtlas total: {total:,} entries")

# Layer distribution for new entries
a.execute(
    """SELECT lm.layer, COUNT(*) FROM entries e
       JOIN layer_membership lm ON e.id = lm.entry_id AND lm.is_primary=1
       WHERE e.id > ? GROUP BY lm.layer ORDER BY 2 DESC""",
    (max_id - inserted,),
)
print(f"\nNew entries by layer:")
for layer, cnt in a.fetchall():
    print(f"  {layer:50s}: {cnt:4d}")

atl.close()
print("\nDone.")
