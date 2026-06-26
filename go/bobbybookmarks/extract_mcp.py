import sqlite3

atl = sqlite3.connect("atlas.db")
c = atl.cursor()

output = []
seen = set()


def add_entry(row):
    if row[1] in seen:
        return
    seen.add(row[1])
    title = (row[2] or "n/a")[:60].encode("ascii", "replace").decode()
    feats = (row[4] or "n/a")[:80].encode("ascii", "replace").decode()
    url = row[1][:80]
    output.append("sig={:5.0f} | {}".format(row[3], title))
    output.append("  URL: " + url)
    output.append("  " + feats)
    output.append("")


# SECTION 1: DIRECTORIES / REGISTRIES / LISTS
output.append("=" * 80)
output.append("1. MCP DIRECTORIES, REGISTRIES, LISTS & HUBS")
output.append("=" * 80)
output.append("")

dir_patterns = [
    "awesome-mcp-servers",
    "chatmcp/mcpso",
    "mcpservers.org",
    "mcp-marketplace",
    "glama.ai/mcp",
    "hub.docker.com/mcp",
    "playbooks.com/mcp",
    "mcppedia.org",
    "mcphub",
    "mcpm.sh",
    "mcp-context-forge",
    "mcp-hub-mcp",
    "infoseek.ai/mcp",
    "composio",
    "mcp-registry",
    "smithery",
    "mcp.run",
    "loadoutz",
    "mcp-marketplace-zeta",
    "toolsdk",
    "redhat-ai-tools",
    "awesome-crypto-mcp",
    "awesome-mcp-enterprise",
]

for name in dir_patterns:
    c.execute(
        "SELECT id, url, page_title, signal, main_features, short_description FROM entries WHERE url LIKE ? ORDER BY signal DESC LIMIT 3",
        ("%{}%".format(name),),
    )
    for row in c.fetchall():
        add_entry(row)

# Also find by title
c.execute("""SELECT id, url, page_title, signal, main_features, short_description 
    FROM entries
    WHERE (page_title LIKE '%MCP Director%' OR page_title LIKE '%MCP Registr%' 
           OR page_title LIKE '%MCP Hub%' OR page_title LIKE '%MCP Market%'
           OR page_title LIKE '%MCP Catalog%' OR page_title LIKE '%awesome mcp%'
           OR page_title LIKE '%Awesome MCP%' OR page_title LIKE '%MCP server list%')
    AND url NOT LIKE '%reddit%'
    ORDER BY signal DESC LIMIT 20""")
for row in c.fetchall():
    add_entry(row)

# SECTION 2: MCP INFRASTRUCTURE
output.append("=" * 80)
output.append("2. MCP INFRASTRUCTURE (Gateways, Routers, Proxies, Protocol Libraries)")
output.append("=" * 80)
output.append("")

c.execute("""
    SELECT DISTINCT e.id, e.url, e.page_title, e.signal, e.main_features
    FROM entries e
    JOIN layer_membership lm ON e.id = lm.entry_id
    WHERE (lm.subcategory LIKE '%MCP Infra%' OR lm.subcategory LIKE '%MCP Server Orch%'
           OR lm.subcategory LIKE '%MCP Discovery%' OR lm.subcategory LIKE '%Tool Discovery%')
    AND lm.is_primary = 1
    ORDER BY e.signal DESC
""")
for row in c.fetchall():
    add_entry(row)

c.execute("""
    SELECT id, url, page_title, signal, main_features
    FROM entries WHERE (tags LIKE '%mcp-gateway%' OR tags LIKE '%mcp-proxy%' 
                        OR tags LIKE '%mcp-router%' OR tags LIKE '%mcp-protocol%'
                        OR tags LIKE '%mcp-infrastructure%' OR tags LIKE '%mcp-bridge%')
    AND url NOT LIKE '%reddit%' ORDER BY signal DESC LIMIT 30
""")
for row in c.fetchall():
    add_entry(row)

# SECTION 3: MCP PROTOCOL IMPLEMENTATIONS
output.append("=" * 80)
output.append("3. MCP PROTOCOL IMPLEMENTATIONS (SDKs, Reference Impls, Specification)")
output.append("=" * 80)
output.append("")

c.execute("""
    SELECT id, url, page_title, signal, main_features
    FROM entries WHERE (url LIKE '%modelcontextprotocol%' OR url LIKE '%mcp-sdk%' 
                        OR url LIKE '%mcp-server-ts%' OR url LIKE '%mcp-python%'
                        OR tags LIKE '%mcp-protocol%' OR tags LIKE '%mcp-sdk%')
    AND url NOT LIKE '%reddit%' ORDER BY signal DESC LIMIT 30
""")
for row in c.fetchall():
    add_entry(row)

# SECTION 4: INDIVIDUAL MCP SERVERS BY FUNCTION
output.append("=" * 80)
output.append("4. INDIVIDUAL MCP SERVERS BY FUNCTION")
output.append("=" * 80)
output.append("")

func_categories = [
    (
        "4a. Database and Storage",
        [
            "database",
            "sql",
            "postgres",
            "mysql",
            "sqlite",
            "vector",
            "qdrant",
            "chroma",
            "neo4j",
            "mongodb",
            "redis",
            "supabase",
        ],
    ),
    (
        "4b. Memory and Knowledge Graph",
        [
            "memory",
            "persistent-memory",
            "rag",
            "embedding",
            "knowledge-graph",
            "long-term-memory",
        ],
    ),
    (
        "4c. Browser Automation",
        [
            "browser",
            "playwright",
            "puppeteer",
            "selenium",
            "web-scraping",
            "browser-use",
        ],
    ),
    (
        "4d. Search and Discovery",
        [
            "search",
            "web-search",
            "tavily",
            "brave-search",
            "google-search",
            "perplexity",
        ],
    ),
    (
        "4e. Git and Version Control",
        ["git", "version-control", "github", "gitlab", "code-review"],
    ),
    (
        "4f. File and Document",
        ["file", "filesystem", "document", "pdf", "markdown", "obsidian"],
    ),
    (
        "4g. Cloud and Infrastructure",
        [
            "cloud",
            "aws",
            "azure",
            "gcp",
            "terraform",
            "kubernetes",
            "docker",
            "infrastructure",
        ],
    ),
    (
        "4h. Security and Auth",
        ["security", "vulnerability", "scanning", "audit", "auth", "oauth", "ssh"],
    ),
    (
        "4i. AI/LLM Integration",
        [
            "llm",
            "openai",
            "claude",
            "anthropic",
            "gemini",
            "inference",
            "model-routing",
        ],
    ),
    (
        "4j. Messaging and Communication",
        [
            "slack",
            "discord",
            "telegram",
            "email",
            "whatsapp",
            "messaging",
            "notification",
        ],
    ),
    (
        "4k. Coding Tools and IDEs",
        ["coding", "ide", "editor", "development", "debugging", "lsp"],
    ),
    (
        "4l. Finance and Trading",
        ["finance", "trading", "crypto", "payment", "stock", "bitcoin"],
    ),
    (
        "4m. Testing and Monitoring",
        ["testing", "qa", "monitoring", "observability", "logging", "metrics"],
    ),
    (
        "4n. Agent Orchestration",
        ["agent", "orchestration", "workflow", "multi-agent", "automation"],
    ),
    (
        "4o. Media Generation",
        ["image", "video", "audio", "media", "generation", "tts", "stt"],
    ),
]

for cat_name, keywords in func_categories:
    output.append("--- {} ---".format(cat_name))
    tag_conditions = " OR ".join(["e.tags LIKE '%{}%'".format(kw) for kw in keywords])
    c.execute(
        """
        SELECT e.id, e.url, e.page_title, e.signal, e.main_features
        FROM entries e
        WHERE ({})
        AND (e.tags LIKE '%mcp%' OR e.url LIKE '%mcp%')
        AND e.url NOT LIKE '%reddit%'
        AND e.signal >= 75
        ORDER BY e.signal DESC LIMIT 50
    """.format(tag_conditions)
    )
    rows = c.fetchall()
    output.append("  ({} shown, signal >= 75)".format(len(rows)))
    for row in rows:
        add_entry(row)

# SECTION 5: COMPLETE HIGH-SIGNAL INDEX
output.append("=" * 80)
output.append("5. COMPLETE MCP SERVER INDEX (signal >= 80, non-Reddit)")
output.append("=" * 80)
output.append("")

c.execute("""
    SELECT e.id, e.url, e.page_title, e.signal, e.main_features
    FROM entries e
    WHERE (e.tags LIKE '%mcp%' OR e.url LIKE '%mcp%')
    AND e.url NOT LIKE '%reddit%'
    AND e.signal >= 80
    ORDER BY e.signal DESC
""")
all_high = c.fetchall()
output.append("Total high-signal MCP entries: {}".format(len(all_high)))
output.append("")
for row in all_high:
    add_entry(row)

# SECTION 6: A2A PROTOCOL
output.append("=" * 80)
output.append("6. A2A (Agent-to-Agent) PROTOCOL")
output.append("=" * 80)
output.append("")

c.execute("""
    SELECT e.id, e.url, e.page_title, e.signal, e.main_features
    FROM entries e
    WHERE (e.tags LIKE '%a2a%' OR e.url LIKE '%a2a%' OR e.url LIKE '%agent-to-agent%')
    AND e.url NOT LIKE '%reddit%'
    ORDER BY e.signal DESC
""")
for row in c.fetchall():
    add_entry(row)

# Write
with open("research/mcp_comprehensive_list.md", "w", encoding="utf-8") as f:
    f.write("\n".join(output))

print("Written {} lines to research/mcp_comprehensive_list.md".format(len(output)))
print("Unique entries: {}".format(len(seen)))

atl.close()
