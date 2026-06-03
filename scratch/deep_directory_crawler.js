const { chromium } = require('playwright');
const Database = require('better-sqlite3');
const { v4: uuidv4 } = require('uuid');
const path = require('path');
const crypto = require('crypto');

const DB_PATH = path.join(__dirname, '..', 'tormentnexus.db');

const SEED_URLS = [
    'https://mcpservers.org',
    'https://mcp-marketplace-zeta.vercel.app',
    'http://mcp-marketplace.io',
    'https://glama.ai/mcp',
    'https://hub.docker.com/mcp',
    'https://playbooks.com/mcp',
    'https://mcppedia.org',
    'https://mcphubx.com',
    'https://Infoseek.ai/mcp',
    'https://mcp-registry-dh5.pages.dev',
    'http://loadoutz.io'
];

const MAX_DEPTH = 2; // Only go 2 levels deep to avoid infinite crawling
const MAX_PAGES_PER_DOMAIN = 100;

function extractGithubInfo(url) {
    try {
        const match = url.match(/github\.com\/([a-zA-Z0-9_.-]+)\/([a-zA-Z0-9_.-]+)/i);
        if (match) {
            const owner = match[1].toLowerCase().trim();
            const repo = match[2].toLowerCase().trim();
            const blacklist = ["features", "marketplace", "pricing", "collections", "topics", "trending", "notifications", "settings", "orgs", "issues", "pulls", "wiki", "releases", "actions", "projects"];
            if (blacklist.includes(owner) || blacklist.includes(repo)) {
                return null;
            }
            return { owner, repo, cid: `github/${owner}/${repo}`, url: `https://github.com/${owner}/${repo}` };
        }
    } catch (e) {}
    return null;
}

async function main() {
    console.log("=== STARTING DEEP DIRECTORY CRAWLER ===");
    const db = new Database(DB_PATH);
    
    // Load existing
    const existingCids = new Set(db.prepare("SELECT canonical_id FROM published_mcp_servers").all().map(r => r.canonical_id));
    console.log(`Loaded ${existingCids.size} existing servers from DB.`);

    const insertServer = db.prepare(`
        INSERT INTO published_mcp_servers (
            uuid, canonical_id, display_name, description, repository_url, homepage_url,
            transport, install_method, tags, categories, status, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, 'stdio', 'unknown', ?, ?, 'discovered', strftime('%s', 'now'), strftime('%s', 'now'))
    `);
    
    const insertSource = db.prepare(`
        INSERT INTO published_mcp_server_sources (uuid, server_uuid, source_name, source_url, raw_payload)
        VALUES (?, ?, ?, ?, ?)
    `);

    const browser = await chromium.launch({ headless: true });
    const context = await browser.newContext({
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
    });

    let newServersAdded = 0;
    
    for (const seedUrl of SEED_URLS) {
        console.log(`\n--- Crawling seed: ${seedUrl} ---`);
        const queue = [{ url: seedUrl, depth: 0 }];
        const visited = new Set();
        let pagesCrawled = 0;
        
        const seedDomain = new URL(seedUrl).hostname;

        while (queue.length > 0 && pagesCrawled < MAX_PAGES_PER_DOMAIN) {
            const { url, depth } = queue.shift();
            
            if (visited.has(url)) continue;
            visited.add(url);
            
            console.log(`[Depth ${depth}] Visiting: ${url}`);
            pagesCrawled++;
            
            const page = await context.newPage();
            try {
                // Navigate and wait for network idle to catch SPAs
                await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 15000 });
                // Briefly wait for any JS to render more links
                await page.waitForTimeout(2000); 

                // Extract all hrefs
                const links = await page.evaluate(() => {
                    return Array.from(document.querySelectorAll('a[href]')).map(a => a.href);
                });

                for (const link of links) {
                    // Check if it's a GitHub link
                    if (link.includes('github.com')) {
                        const info = extractGithubInfo(link);
                        if (info && !existingCids.has(info.cid)) {
                            console.log(`  ⭐ Found NEW server: ${info.cid}`);
                            
                            const serverUuid = uuidv4();
                            insertServer.run(
                                serverUuid, info.cid, info.repo.replace(/[-_]/g, ' '), 
                                `Discovered via crawler on ${seedDomain}`, info.url, info.url,
                                JSON.stringify(["github-repo", "crawler-discovered"]), JSON.stringify(["crawler-imported"])
                            );
                            insertSource.run(
                                uuidv4(), serverUuid, `crawler-${seedDomain}`, url, JSON.stringify({source: url, link: link})
                            );
                            
                            existingCids.add(info.cid);
                            newServersAdded++;
                        }
                    } else if (depth < MAX_DEPTH) {
                        // If it's an internal link, add to queue
                        try {
                            const linkDomain = new URL(link).hostname;
                            // Only follow links on the same domain or a subdomain
                            if (linkDomain.includes(seedDomain) || seedDomain.includes(linkDomain)) {
                                if (!visited.has(link)) {
                                    queue.push({ url: link, depth: depth + 1 });
                                }
                            }
                        } catch (e) {} // Ignore malformed URLs
                    }
                }
            } catch (err) {
                console.log(`  [!] Failed to crawl ${url}: ${err.message.split('\\n')[0]}`);
            } finally {
                await page.close();
            }
        }
    }

    await browser.close();
    console.log(`\n=== CRAWLER FINISHED ===`);
    console.log(`Total new servers added: ${newServersAdded}`);
    console.log(`Total catalog size: ${existingCids.size}`);
}

main().catch(err => {
    console.error(err);
    process.exit(1);
});
