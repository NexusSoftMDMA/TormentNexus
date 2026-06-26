[de](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo?locale=de-DE#readme-md "German") [en](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo#readme-md "English") [es](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo?locale=es-ES#readme-md "Spanish") [ja](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo?locale=ja-JP#readme-md "Japanese") [ko](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo?locale=ko-KR#readme-md "Korean") [ru](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo?locale=ru-RU#readme-md "Russian") [zh](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo?locale=zh-CN#readme-md "Chinese")

Which integrations are available for this server?

Uses [DuckDuckGo](https://glama.ai/mcp/servers/integrations/duckduckgo) search to find pages mentioning a domain, discover guest post opportunities, and find competitor link sources.

# backlink-mcp

[**Full docs & install guide → sellonllm.com/backlink-mcp.html**](https://www.sellonllm.com/backlink-mcp.html)

> Automate backlink research, unlinked mention hunting, and outreach prep inside your AI assistant. Free, no API keys required.

Connect to **Claude**, **Cursor**, or any MCP-compatible AI assistant and let it find backlink opportunities, discover unlinked mentions, research prospects, and extract contact info for outreach — all for free.

Part of the [**SellOnLLM SEO MCP suite**](https://www.sellonllm.com/) — a hub of free MCP servers for SEO and AI visibility.

* * *

## Why this exists

Tools like Ahrefs and Moz cost hundreds of dollars a month. This MCP gives you backlink research capabilities directly inside your AI assistant at zero cost, using:

- **DuckDuckGo** — mention discovery and prospect finding

- **Wayback Machine CDX API** — historical link data

- **httpx + BeautifulSoup** — page scraping and link verification


* * *

## Tools

|     |     |
| --- | --- |
| Tool | Description |
| `find_mentions` | Find all pages mentioning your domain (linked or unlinked) |
| `find_prospects` | Discover guest post, resource page, and roundup opportunities by niche |
| `find_competitor_link_sources` | Find pages linking to a competitor — prime outreach targets |
| `verify_page_links` | Scrape a URL to check if it links to you and extract contact info |
| `extract_contact_info` | Pull emails, social handles, and contact pages from any site |
| `check_page_history` | Check Wayback Machine history — verify a page still exists |

* * *

## Quickstart

### 1\. Clone and install

```
git clone https://github.com/vipul510-web/mcp-backlink-for-seo.git
cd mcp-backlink-for-seo
python3 -m venv .venv
.venv/bin/pip install "mcp[cli]>=1.0.0" "ddgs>=9.0.0" httpx beautifulsoup4 lxml
```

### 2\. Connect to Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (Mac) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```
{
  "mcpServers": {
    "backlink-mcp": {
      "command": "/absolute/path/to/backlink-mcp/.venv/bin/python",
      "args": ["/absolute/path/to/backlink-mcp/server.py"]
    }
  }
}
```

Restart Claude Desktop. The backlink tools will appear automatically.

### 3\. Connect to Cursor

Add to `~/.cursor/mcp.json` (global) or `.cursor/mcp.json` (per project):

```
{
  "mcpServers": {
    "backlink-mcp": {
      "command": "/absolute/path/to/backlink-mcp/.venv/bin/python",
      "args": ["/absolute/path/to/backlink-mcp/server.py"]
    }
  }
}
```

### 4\. Connect to any MCP-compatible client

```
.venv/bin/python server.py
```

The server communicates over stdio, compatible with any MCP client.

* * *

## Usage examples

Once connected, just talk to your AI assistant:

**Find unlinked mentions (outreach opportunities):**

```
Find unlinked mentions of mybrand.com
```

**Discover guest post opportunities:**

```
Find guest post opportunities in the personal finance niche
```

**Research a competitor's backlinks:**

```
Who links to competitor.com? Find me 20 results.
```

**Verify and enrich a prospect:**

```
Check if techblog.com/article links to mybrand.com and find their contact email
```

**Full link building workflow:**

```
1. Find prospects in the SaaS marketing niche
2. Verify which ones don't already link to mysaas.com
3. Extract contact info for the top 5
4. Draft an outreach email for each
```

* * *

## Typical workflow

```
find_prospects / find_mentions / find_competitor_link_sources
                    ↓
              verify_page_links
          (linked or unlinked? contact info?)
                    ↓
           extract_contact_info
              (email, socials)
                    ↓
          outreach via Gmail MCP
```

* * *

## Requirements

- Python 3.10+

- No API keys needed

- No paid subscriptions


* * *

## Limitations

- DuckDuckGo returns a sample of results, not a complete link graph

- Rate limiting: built-in 1.5s delay between searches to avoid blocks

- The Wayback CDX endpoint can occasionally return 503 or time out; `check_page_history` retries automatically

- Common Crawl graph data (full inbound link index) is not yet integrated — contributions welcome


### Changelog (recent)

- **0.1.1** — Switched search from deprecated `duckduckgo-search` to the maintained `ddgs` package (same DuckDuckGo backend; fixes empty search results). Hardened Wayback CDX with HTTPS, longer timeouts, and retries.


* * *

## Contributing

PRs welcome. High-impact areas:

- Common Crawl graph API integration for true inbound link discovery

- Broken link detection (find dead pages on prospect sites)

- Bulk processing (run across a list of URLs)

- Output to CSV / Google Sheets


* * *

## Part of SellOnLLM

This MCP is part of the [SellOnLLM](https://www.sellonllm.com/) SEO MCP suite — free, open-source MCP servers for SEO and AI visibility built for Claude and Cursor.

- [Backlink MCP](https://www.sellonllm.com/backlink-mcp.html) — this tool

- [GA4 + Search Console MCP](https://www.sellonllm.com/) — query your traffic and rankings from your AI assistant

- [AI Visibility MCP](https://www.sellonllm.com/) — check if your site is cited by AI tools like Perplexity


* * *

## License

MIT

- [.gitignore](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/blob/2d97dd239bcb49baa8033d4d621a3f2f7bd3622f/.gitignore)

- [pyproject.toml](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/blob/2d97dd239bcb49baa8033d4d621a3f2f7bd3622f/pyproject.toml)

- [README.md](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/blob/2d97dd239bcb49baa8033d4d621a3f2f7bd3622f/README.md)

- [server.py](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/blob/2d97dd239bcb49baa8033d4d621a3f2f7bd3622f/server.py)

- [smithery.yaml](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/blob/2d97dd239bcb49baa8033d4d621a3f2f7bd3622f/smithery.yaml)


Install Server

This server [cannot be installed](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/score)

A

license - permissive license

-

quality - not tested

C

maintenance

[How are these scores calculated?](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/score)

#### Resources

- [GitHub Repository](https://github.com/vipul510-web/mcp-backlink-for-seo)
[Need Help?](https://glama.ai/mcp/discord) Report Issue [Related Servers](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/related-servers)

Unclaimed servers have limited discoverability.

#### Looking for Admin?

If you are the server author,claim this serverto access and configure the [admin panel](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/admin).

#### Related MCP Servers

[View all related MCP servers](https://glama.ai/mcp/servers/vipul510-web/mcp-backlink-for-seo/related-servers)

#### Related MCP Connectors

- [AiPayGen — 65+ AI Tools as an MCP Server](https://glama.ai/mcp/connectors/io.github.Damien829/aipaygen)





65+ AI tools as MCP: research, write, code, scrape, translate, RAG, agent memory, workflows

- [ScrapeGraphAI-scrapegraph-mcp](https://glama.ai/mcp/connectors/ai.smithery/ScrapeGraphAI-scrapegraph-mcp)





Enable language models to perform advanced AI-powered web scraping with enterprise-grade reliabili…

- [rankoracle](https://glama.ai/mcp/connectors/io.tooloracle/rankoracle)





SEO Intelligence MCP — 13 tools: keyword research, SERP, domain audits, competitors.


[View all MCP Connectors](https://glama.ai/mcp/connectors)

## Latest Blog Posts

- [Lightport: Open-Sourcing Glama's AI Gateway](https://glama.ai/blog/2026-04-27-open-source-ai-gateway)

By [punkpeye](https://github.com/punkpeye) onApril 27, 2026.





open source



OpenAI

- [Tool Definition Quality Score (TDQS)](https://glama.ai/blog/2026-04-03-tool-definition-quality-score-tdqs)

By [punkpeye](https://github.com/punkpeye) onApril 3, 2026.





mcp

- [The Hackers Who Tracked My Sleep Cycle](https://glama.ai/blog/2026-03-26-the-hackers-who-tracked-my-sleep-cycle)

By [punkpeye](https://github.com/punkpeye) onMarch 26, 2026.





security


#### MCP directory API

We provide all the information about MCP servers via our [MCP API](https://glama.ai/mcp/reference).

```
curl -X GET 'https://glama.ai/api/mcp/v1/servers/vipul510-web/mcp-backlink-for-seo'
```

If you have feedback or need assistance with the MCP directory API, please join our [Discord server](https://glama.ai/discord)