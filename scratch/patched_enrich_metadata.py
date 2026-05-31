"""
Comprehensive MCP Metadata Enricher + New Directory Scraper

Three major tasks:
1. SCRAPE: Official MCP Registry (cursor-paginated) + new directories not yet scraped
2. ENRICH: Fetch GitHub repo metadata (stars, language, topics, package.json, .env.example)
           and enrich existing servers with auth_model, required_secrets, real recipes
3. ENRICH: Smithery detailed server info (configSchema / environment variables)
"""

import sqlite3, uuid, json, re, time, urllib.request, urllib.error, os

BORG_DB_PATH = r"c:\Users\hyper\workspace\borg\tormentnexus.db"

def http_get(url, timeout=20, accept="application/json"):
    req = urllib.request.Request(url, headers={
        "User-Agent": "Mozilla/5.0 (compatible; HypercodeBot/1.0)",
        "Accept": accept,
    })
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            return r.read().decode("utf-8", errors="replace")
    except Exception as e:
        return None

def http_get_json(url, timeout=20):
    d = http_get(url, timeout)
    if d:
        try: return json.loads(d)
        except: return None
    return None

def extract_github(url):
    m = re.search(r'github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)', url or "")
    if m: return m.group(1), m.group(2)
    return None, None

# ═══════════════════════════════════════════════════════════════════════════
# PART 1: Scrape Official MCP Registry (correct cursor pagination)
# ═══════════════════════════════════════════════════════════════════════════
def scrape_official_registry():
    print("\n[Official MCP Registry] Scraping registry.modelcontextprotocol.io...")
    conn = sqlite3.connect(BORG_DB_PATH)
    c = conn.cursor()
    
    added = updated = 0
    cursor = None
    page = 0
    latest_only = set()  # track server names we've seen latest version for
    
    while True:
        url = "https://registry.modelcontextprotocol.io/v0.1/servers?limit=100"
        if cursor:
            url += f"&cursor={urllib.request.quote(cursor)}"
        
        data = http_get_json(url)
        if not data or not data.get("servers"):
            break
        
        servers = data["servers"]
        for entry in servers:
            srv = entry.get("server", {})
            meta = entry.get("_meta", {}).get("io.modelcontextprotocol.registry/official", {})
            
            # Skip non-latest versions (registry has multiple versions per server)
            is_latest = meta.get("isLatest", True)
            srv_name = srv.get("name", "")
            if srv_name in latest_only:
                continue
            if is_latest:
                latest_only.add(srv_name)
            
            # Build canonical ID
            title = srv.get("title", "") or srv.get("name", "")
            desc = srv.get("description", "") or ""
            repo_url = (srv.get("repository") or {}).get("url", "")
            homepage = srv.get("homepage", "") or ""
            
            # Extract packages for install info
            packages = srv.get("packages", []) or []
            remotes = srv.get("remotes", []) or []
            
            install_method = "unknown"
            transport = "stdio"
            remote_url = ""
            pkg_id = ""
            required_env = {}
            
            for pkg in packages:
                reg_type = pkg.get("registryType", "").lower()
                pkg_id = pkg.get("identifier", "")
                pkg_transport = pkg.get("transport", {}) or {}
                t_type = pkg_transport.get("type", "stdio")
                t_url = pkg_transport.get("url", "")
                
                if reg_type in ("npm", "npmjs"):
                    install_method = "npm"
                elif reg_type == "pypi":
                    install_method = "pip"
                elif reg_type in ("docker", "dockerhub"):
                    install_method = "docker"
                elif reg_type == "cargo":
                    install_method = "cargo"
                
                transport = t_type
                if t_url:
                    remote_url = t_url
                
                # Extract runtime env requirements
                runtime = pkg.get("runtime", {}) or {}
                env_vars = runtime.get("env", {}) or {}
                required_env.update(env_vars)
            
            if remotes and not packages:
                transport = remotes[0].get("type", "streamable-http")
                remote_url = remotes[0].get("url", "")
                install_method = "remote"
            
            # Build recipe template
            if install_method == "npm" and pkg_id:
                recipe_template = {"type": "stdio", "command": "npx", "args": ["-y", pkg_id], "env": {}}
            elif install_method == "pip" and pkg_id:
                recipe_template = {"type": "stdio", "command": "uvx", "args": [pkg_id], "env": {}}
            elif install_method == "remote" and remote_url:
                recipe_template = {"type": transport, "url": remote_url}
            elif install_method == "docker" and pkg_id:
                recipe_template = {"type": "stdio", "command": "docker", "args": ["run", "--rm", "-i", pkg_id], "env": {}}
            else:
                recipe_template = {"type": transport or "stdio", "command": "npx", "args": ["-y", pkg_id or srv_name], "env": {}}
            
            # Determine required secrets from env vars
            required_secrets_list = [k for k in required_env.keys() 
                                     if any(w in k.upper() for w in ["KEY", "TOKEN", "SECRET", "PASSWORD", "AUTH", "API"])]
            
            # Build canonical ID
            owner, repo = extract_github(repo_url)
            if owner and repo:
                cid = f"github/{owner.lower()}/{repo.lower()}"
            else:
                cid_clean = re.sub(r'[^a-z0-9/_.-]', '-', srv_name.lower())
                cid = f"mcp-registry/{cid_clean}"
            
            # Determine auth model
            auth_model = "none"
            if required_secrets_list:
                auth_model = "api_key"
            elif transport in ("streamable-http", "http", "sse"):
                auth_model = "http"
            
            tags = ["official-registry", "modelcontextprotocol", install_method]
            if transport in ("streamable-http", "http"):
                tags.append("remote")
            
            # Upsert server
            c.execute("SELECT uuid, description, tags FROM published_mcp_servers WHERE canonical_id=?", (cid,))
            row = c.fetchone()
            if row:
                server_uuid = row[0]
                merged_desc = desc if len(desc) > len(row[1] or "") else (row[1] or "")
                try: ex_t = json.loads(row[2] or "[]")
                except: ex_t = []
                merged_tags = json.dumps(list(set(ex_t + tags)))
                c.execute("""UPDATE published_mcp_servers 
                    SET description=?, tags=?, auth_model=?, install_method=?, transport=?,
                        homepage_url=?, repository_url=?, updated_at=strftime('%s','now')
                    WHERE canonical_id=?""",
                    (merged_desc, merged_tags, auth_model, install_method, transport,
                     homepage or row[1] or "", repo_url or "", cid))
                updated += 1
            else:
                server_uuid = str(uuid.uuid4())
                c.execute("""INSERT INTO published_mcp_servers
                    (uuid, canonical_id, display_name, description, repository_url, homepage_url,
                     transport, install_method, auth_model, tags, categories, status, created_at, updated_at)
                    VALUES (?,?,?,?,?,?,?,?,?,?,?,'active',strftime('%s','now'),strftime('%s','now'))""",
                    (server_uuid, cid, title[:200], desc, repo_url, homepage or remote_url,
                     transport, install_method, auth_model,
                     json.dumps(tags), '["mcp-official-registry"]'))
                added += 1
            
            # Upsert source provenance
            src_uuid = str(uuid.uuid4())
            c.execute("""INSERT INTO published_mcp_server_sources (uuid,server_uuid,source_name,source_url,raw_payload)
                VALUES (?,?,?,?,?)
                ON CONFLICT(server_uuid,source_name) DO UPDATE SET
                    raw_payload=excluded.raw_payload, last_seen_at=strftime('%s','now')""",
                (src_uuid, server_uuid, "registry.modelcontextprotocol.io",
                 f"https://registry.modelcontextprotocol.io/v0.1/servers/{srv_name}", json.dumps(entry)))
            
            # Upsert recipe with real data
            c.execute("SELECT uuid FROM published_mcp_config_recipes WHERE server_uuid=?", (server_uuid,))
            existing_recipe = c.fetchone()
            req_secrets = json.dumps(required_secrets_list)
            req_env = json.dumps(required_env)
            confidence = 85 if required_env else 70  # Official registry = high confidence
            explanation = f"Official MCP registry entry. Transport: {transport}. Registry: {install_method}."
            
            if existing_recipe:
                c.execute("""UPDATE published_mcp_config_recipes 
                    SET template=?, required_secrets=?, required_env=?, confidence=?, explanation=?,
                        generated_by='OfficialRegistry', updated_at=strftime('%s','now')
                    WHERE server_uuid=?""",
                    (json.dumps(recipe_template), req_secrets, req_env, confidence, explanation, server_uuid))
            else:
                r_uuid = str(uuid.uuid4())
                c.execute("""INSERT INTO published_mcp_config_recipes
                    (uuid,server_uuid,recipe_version,template,required_secrets,required_env,
                     confidence,explanation,is_active,generated_by,created_at,updated_at)
                    VALUES (?,?,1,?,?,?,?,?,1,'OfficialRegistry',strftime('%s','now'),strftime('%s','now'))""",
                    (r_uuid, server_uuid, json.dumps(recipe_template), req_secrets, req_env,
                     confidence, explanation))
        
        page += 1
        cursor = data["metadata"].get("nextCursor")
        print(f"  Page {page}: {len(servers)} entries (added={added}, updated={updated}, cursor={cursor[:40] if cursor else 'END'})")
        
        if not cursor:
            break
        time.sleep(0.3)
    
    conn.commit()
    conn.close()
    print(f"  [Official Registry] Done: added={added}, updated={updated}")

# ═══════════════════════════════════════════════════════════════════════════
# PART 2: GitHub Metadata Enrichment for github/owner/repo entries
# Fetches: stars, language, topics, description, package.json, .env.example
# ═══════════════════════════════════════════════════════════════════════════
def enrich_github_metadata():
    """Enrich GitHub-sourced servers with real metadata from GitHub API."""
    print("\n[GitHub Enrichment] Enriching github/* entries with real metadata...")
    
    conn = sqlite3.connect(BORG_DB_PATH)
    c = conn.cursor()
    
    # Get github/* entries that have no stars yet or poor metadata
    c.execute("""SELECT uuid, canonical_id, repository_url, stars, description, install_method
        FROM published_mcp_servers 
        WHERE canonical_id LIKE 'github/%'
        AND (stars = 0 OR stars IS NULL OR install_method = 'unknown')
        ORDER BY uuid
        LIMIT 50""")
    rows = c.fetchall()
    print(f"  Enriching {len(rows):,} GitHub entries...")
    
    enriched = 0
    rate_limited = 0
    
    for server_uuid, cid, repo_url, stars, desc, install_method in rows:
        parts = cid.split("/")
        if len(parts) < 3:
            continue
        owner, repo = parts[1], parts[2]
        
        # GitHub API for repo metadata
        api_url = f"https://api.github.com/repos/{owner}/{repo}"
        data = http_get_json(api_url)
        
        if not data:
            time.sleep(0.2)
            continue
        
        if "message" in data and "rate limit" in str(data.get("message", "")).lower():
            rate_limited += 1
            print(f"  Rate limited after {enriched} enriched. Waiting 60s...")
            time.sleep(60)
            continue
        
        if "message" in data:  # 404 etc
            continue
        
        real_stars = data.get("stargazers_count", 0)
        language = data.get("language", "") or ""
        real_desc = data.get("description", "") or desc or ""
        homepage = data.get("homepage", "") or ""
        topics = data.get("topics", []) or []
        default_branch = data.get("default_branch", "main")
        is_archived = data.get("archived", False)
        
        # Determine install method from language and topics
        new_install = install_method
        if new_install == "unknown":
            lang_lower = language.lower()
            topic_str = " ".join(topics).lower()
            if "python" in lang_lower or "python" in topic_str:
                new_install = "pip"
            elif language in ("TypeScript", "JavaScript"):
                new_install = "npm"
            elif language == "Go":
                new_install = "go"
            elif language == "Rust":
                new_install = "cargo"
            elif language == "C#":
                new_install = "dotnet"
            elif language in ("Java", "Kotlin"):
                new_install = "mvn"
            elif language == "Ruby":
                new_install = "gem"
        
        # Determine auth_model from topics
        auth_model = "unknown"
        topic_str = " ".join(topics).lower()
        if any(t in topic_str for t in ["oauth", "oauth2"]):
            auth_model = "oauth2"
        elif any(t in topic_str for t in ["api-key", "apikey", "api_key"]):
            auth_model = "api_key"
        elif any(t in topic_str for t in ["no-auth", "noauth", "open"]):
            auth_model = "none"
        
        # Check package.json for main entry / scripts
        pkg_json = None
        for fname in ["package.json"]:
            raw = http_get(f"https://raw.githubusercontent.com/{owner}/{repo}/{default_branch}/{fname}")
            if raw:
                try:
                    pkg_json = json.loads(raw)
                    break
                except:
                    pass
        
        # Try to find .env.example or README for env var hints
        env_hints = []
        env_example = http_get(f"https://raw.githubusercontent.com/{owner}/{repo}/{default_branch}/.env.example")
        if not env_example:
            env_example = http_get(f"https://raw.githubusercontent.com/{owner}/{repo}/{default_branch}/.env.sample")
        if env_example:
            # Extract KEY=... lines
            for line in env_example.split("\n"):
                line = line.strip()
                if "=" in line and not line.startswith("#"):
                    key = line.split("=")[0].strip()
                    if key and re.match(r'^[A-Z][A-Z0-9_]+$', key):
                        env_hints.append(key)
                elif line.startswith("#") and ("API_KEY" in line or "TOKEN" in line or "SECRET" in line):
                    m = re.search(r'([A-Z][A-Z0-9_]+(?:_KEY|_TOKEN|_SECRET|_API))', line)
                    if m:
                        env_hints.append(m.group(1))
        
        # Determine required secrets
        required_secrets = [k for k in env_hints if any(w in k for w in ["KEY", "TOKEN", "SECRET", "PASSWORD", "AUTH", "API"])]
        if required_secrets and auth_model == "unknown":
            auth_model = "api_key"
        
        # Build better recipe from package.json
        new_recipe = None
        if pkg_json:
            bin_entry = pkg_json.get("bin", {})
            pkg_name = pkg_json.get("name", "")
            scripts = pkg_json.get("scripts", {})
            
            if isinstance(bin_entry, dict) and bin_entry:
                first_bin = next(iter(bin_entry.keys()))
                new_recipe = {"type": "stdio", "command": "npx", "args": ["-y", pkg_name or repo], "env": {}}
            elif pkg_name and "mcp" in pkg_name.lower():
                new_recipe = {"type": "stdio", "command": "npx", "args": ["-y", pkg_name], "env": {}}
            
            if new_install == "unknown" and pkg_name:
                new_install = "npm"
        
        # Build all_tags
        new_tags_additions = topics + ([language.lower()] if language else [])
        
        # Update server record
        c.execute("""SELECT tags FROM published_mcp_servers WHERE uuid=?""", (server_uuid,))
        tag_row = c.fetchone()
        try: ex_tags = json.loads(tag_row[0] or "[]")
        except: ex_tags = []
        merged_tags = json.dumps(list(set(ex_tags + new_tags_additions)))
        
        c.execute("""UPDATE published_mcp_servers SET
            stars=?, description=?, homepage_url=COALESCE(NULLIF(homepage_url,''), ?),
            auth_model=?, install_method=?, tags=?,
            status=CASE WHEN ? THEN 'archived' ELSE status END,
            updated_at=strftime('%s','now')
            WHERE uuid=?""",
            (real_stars, real_desc[:500], homepage,
             auth_model, new_install, merged_tags,
             is_archived, server_uuid))
        
        # Update recipe if we have better data
        if env_hints or new_recipe:
            c.execute("SELECT uuid, template FROM published_mcp_config_recipes WHERE server_uuid=?", (server_uuid,))
            recipe_row = c.fetchone()
            
            if new_recipe is None:
                # Use existing template, just update env hints
                if recipe_row:
                    try: existing_tmpl = json.loads(recipe_row[1] or "{}")
                    except: existing_tmpl = {}
                    new_recipe = existing_tmpl
            
            if new_recipe:
                req_secrets = json.dumps(required_secrets)
                req_env_dict = {k: f"<your_{k.lower()}>" for k in env_hints[:20]}
                req_env = json.dumps(req_env_dict)
                
                if recipe_row:
                    c.execute("""UPDATE published_mcp_config_recipes SET
                        template=?, required_secrets=?, required_env=?,
                        confidence=?, explanation=?, generated_by='GitHubEnricher',
                        updated_at=strftime('%s','now')
                        WHERE server_uuid=?""",
                        (json.dumps(new_recipe), req_secrets, req_env,
                         60 if env_hints else 35,
                         f"GitHub enriched: {language}, {real_stars} stars, env vars: {len(env_hints)}",
                         server_uuid))
                else:
                    r_uuid = str(uuid.uuid4())
                    c.execute("""INSERT INTO published_mcp_config_recipes
                        (uuid,server_uuid,recipe_version,template,required_secrets,required_env,
                         confidence,explanation,is_active,generated_by,created_at,updated_at)
                        VALUES (?,?,1,?,?,?,?,?,1,'GitHubEnricher',strftime('%s','now'),strftime('%s','now'))""",
                        (r_uuid, server_uuid, json.dumps(new_recipe), req_secrets, req_env,
                         60 if env_hints else 30,
                         f"GitHub enriched: {language}, {real_stars} stars"))
        
        enriched += 1
        if enriched % 50 == 0:
            conn.commit()
            print(f"  Enriched {enriched}/{len(rows)} (rate_limited={rate_limited})")
        
        time.sleep(0.15)  # ~6 req/sec, GitHub allows 60/min unauthenticated
    
    conn.commit()
    conn.close()
    print(f"  [GitHub Enrichment] Done: {enriched} enriched, {rate_limited} rate limited")

# ═══════════════════════════════════════════════════════════════════════════
# PART 3: Smithery Enrichment - fetch configSchema / environment per server
# ═══════════════════════════════════════════════════════════════════════════
def enrich_smithery():
    """Re-fetch Smithery server details with full config schema."""
    print("\n[Smithery Enrichment] Fetching full config schemas from Smithery...")
    
    conn = sqlite3.connect(BORG_DB_PATH)
    c = conn.cursor()
    
    # Get smithery-sourced servers
    c.execute("""SELECT DISTINCT s.uuid, s.canonical_id, src.raw_payload
        FROM published_mcp_servers s
        JOIN published_mcp_server_sources src ON src.server_uuid = s.uuid
        WHERE src.source_name = 'smithery.ai'""")
    rows = c.fetchall()
    print(f"  Processing {len(rows)} Smithery entries...")
    
    enriched = 0
    for server_uuid, cid, raw_payload_str in rows:
        try:
            payload = json.loads(raw_payload_str or "{}")
        except:
            continue
        
        slug = payload.get("qualifiedName", "") or payload.get("slug", "")
        if not slug:
            continue
        
        # Fetch full server detail from Smithery
        # The slug format is like "@namespace/server-name"
        encoded_slug = urllib.request.quote(slug, safe='@/')
        detail_url = f"https://registry.smithery.ai/servers/{encoded_slug}"
        detail = http_get_json(detail_url)
        
        if not detail:
            # Try alternative endpoint
            detail_url2 = f"https://smithery.ai/api/servers/{encoded_slug}"
            detail = http_get_json(detail_url2)
        
        if not detail:
            time.sleep(0.1)
            continue
        
        # Extract enhanced metadata
        config_schema = detail.get("configSchema", {}) or detail.get("configuration", {}) or {}
        connections = detail.get("connections", []) or []
        verified = detail.get("verified", False)
        use_count = detail.get("useCount", 0)
        icon_url = detail.get("iconUrl", "")
        
        # Extract env vars from configSchema
        properties = (config_schema.get("properties", {}) if isinstance(config_schema, dict) else {}) or {}
        required_list = (config_schema.get("required", []) if isinstance(config_schema, dict) else []) or []
        
        required_secrets = [k for k in required_list 
                           if any(w in k.upper() for w in ["KEY", "TOKEN", "SECRET", "PASSWORD", "AUTH"])]
        
        all_env = {k: {"description": v.get("description", ""), "required": k in required_list}
                   for k, v in properties.items() if isinstance(v, dict)}
        
        # Build proper recipe from connections
        recipe_template = None
        transport = "stdio"
        for conn_entry in connections:
            t = conn_entry.get("type", "")
            if t == "stdio":
                cmd_args = conn_entry.get("commandFunction", {}) or {}
                config_cmd = conn_entry.get("stdioFunction", {}) or {}
                # Use the config schema to build the template
                recipe_template = {
                    "type": "stdio",
                    "command": "npx",
                    "args": ["-y", f"@smithery/cli@latest", "run", slug, "--key", "<smithery_api_key>"],
                    "env": {}
                }
                transport = "stdio"
                break
            elif t in ("streamable-http", "http", "sse"):
                url = conn_entry.get("url", "")
                recipe_template = {"type": t, "url": url}
                transport = t
                break
        
        confidence = 80 if verified else 60
        if properties:
            confidence += 10
        
        # Update server record  
        update_fields = []
        if icon_url:
            c.execute("UPDATE published_mcp_servers SET icon_url=?, updated_at=strftime('%s','now') WHERE uuid=?",
                (icon_url, server_uuid))
        
        # Update source with full payload
        c.execute("""UPDATE published_mcp_server_sources SET 
            raw_payload=?, last_seen_at=strftime('%s','now')
            WHERE server_uuid=? AND source_name='smithery.ai'""",
            (json.dumps(detail), server_uuid))
        
        # Update recipe
        if recipe_template or properties:
            c.execute("SELECT uuid FROM published_mcp_config_recipes WHERE server_uuid=?", (server_uuid,))
            recipe_row = c.fetchone()
            req_secrets_json = json.dumps(required_secrets)
            req_env_json = json.dumps(all_env)
            if recipe_row and (recipe_template or properties):
                c.execute("""UPDATE published_mcp_config_recipes SET
                    template=COALESCE(?, template),
                    required_secrets=?, required_env=?,
                    confidence=?, explanation=?,
                    generated_by='SmitheryEnricher',
                    updated_at=strftime('%s','now')
                    WHERE server_uuid=?""",
                    (json.dumps(recipe_template) if recipe_template else None,
                     req_secrets_json, req_env_json,
                     confidence, f"Smithery verified={verified}, useCount={use_count}, envVars={len(properties)}",
                     server_uuid))
        
        enriched += 1
        if enriched % 20 == 0:
            conn.commit()
            print(f"  Smithery enriched: {enriched}/{len(rows)}")
        time.sleep(0.5)
    
    conn.commit()
    conn.close()
    print(f"  [Smithery Enrichment] Done: {enriched} enriched")

# ═══════════════════════════════════════════════════════════════════════════
# PART 4: Scrape more MCP directories we haven't hit yet
# ═══════════════════════════════════════════════════════════════════════════
def scrape_more_directories():
    """Scrape additional MCP directories not yet covered."""
    print("\n[More Directories] Scraping additional MCP directories...")
    
    conn = sqlite3.connect(BORG_DB_PATH)
    c = conn.cursor()
    added = 0
    gh_pattern = re.compile(r'github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)')
    SKIP = {'topics', 'trending', 'features', 'pricing', 'login', 'join', 'explore',
            'marketplace', 'collections', 'notifications', 'settings', 'orgs', 'github'}
    
    sources = [
        # New directories from search results
        ("https://raw.githubusercontent.com/korchasa/awesome-mcp/main/README.md", "korchasa"),
        ("https://raw.githubusercontent.com/ever-works/awesome-mcp-servers/main/README.md", "ever-works"),
        # Community compiled lists
        ("https://raw.githubusercontent.com/josephluck/mcp-servers/main/README.md", "josephluck"),
        ("https://raw.githubusercontent.com/anthonybudd/mcp-server/main/README.md", "anthonybudd"),
        # GitHub search via API for more specific terms
    ]
    
    # GitHub API search for even more specific MCP server repos
    github_queries = [
        "mcp+server+claude+in:name,description",
        "model+context+protocol+server+in:name",
        "mcp+tools+server+in:name",
        "anthropic+mcp+server+in:name",
        "mcp+integration+server+in:name",
        "mcp+fastmcp+server+in:name",
        "mcp+stdio+server+in:description",
        "mcp+npx+server+in:description",
        "mcp+python+server+in:name",
        "mcp+typescript+server+in:name",
        "mcp+rust+server+in:name",
        "mcp+go+server+in:name",
        "mcp+database+server+in:name",
        "mcp+api+server+in:name",
        "mcp+automation+server+in:name",
    ]
    
    print("  Scraping GitHub READMEs...")
    for url, label in sources:
        raw = http_get(url)
        if not raw:
            continue
        n = 0
        for owner, repo in gh_pattern.findall(raw):
            ol, rl = owner.lower(), repo.lower()
            if ol in SKIP or len(rl) < 2:
                continue
            cid = f"github/{ol}/{rl}"
            c.execute("SELECT uuid FROM published_mcp_servers WHERE canonical_id=?", (cid,))
            if not c.fetchone():
                srv_uuid = str(uuid.uuid4())
                c.execute("""INSERT INTO published_mcp_servers
                    (uuid, canonical_id, display_name, description, repository_url, homepage_url,
                     transport, install_method, tags, categories, status, created_at, updated_at)
                    VALUES (?,?,?,?,?,?,'stdio','unknown',?,?,'discovered',strftime('%s','now'),strftime('%s','now'))""",
                    (srv_uuid, cid, repo.replace("-"," ").title()[:200],
                     f"From {label}", f"https://github.com/{owner}/{repo}", f"https://github.com/{owner}/{repo}",
                     json.dumps(["mcp", label, ol]), '["more-directories"]'))
                src_uuid = str(uuid.uuid4())
                c.execute("INSERT INTO published_mcp_server_sources (uuid,server_uuid,source_name,source_url,raw_payload) VALUES (?,?,?,?,?)",
                    (src_uuid, srv_uuid, f"more-dirs/{label}", f"https://github.com/{owner}/{repo}", "{}"))
                n += 1
                added += 1
        if n:
            print(f"  +{n} from {label}")
        time.sleep(0.2)
    
    print(f"  Scraping GitHub API searches...")
    for query in github_queries:
        for page in range(1, 4):
            url = f"https://api.github.com/search/repositories?q={query}&sort=stars&per_page=100&page={page}"
            data = http_get_json(url)
            if not data:
                time.sleep(2)
                break
            if "message" in (data or {}) and "rate limit" in str(data.get("message","")).lower():
                print(f"  Rate limited on GitHub search. Pausing 60s...")
                time.sleep(60)
                break
            items = (data or {}).get("items", [])
            n = 0
            for item in items:
                owner = item.get("owner", {}).get("login", "").lower()
                repo = item.get("name", "").lower()
                desc = item.get("description", "") or ""
                stars = item.get("stargazers_count", 0)
                lang = item.get("language", "") or ""
                if not owner or not repo:
                    continue
                cid = f"github/{owner}/{repo}"
                c.execute("SELECT uuid FROM published_mcp_servers WHERE canonical_id=?", (cid,))
                if not c.fetchone():
                    srv_uuid = str(uuid.uuid4())
                    c.execute("""INSERT INTO published_mcp_servers
                        (uuid, canonical_id, display_name, description, repository_url, homepage_url,
                         transport, install_method, stars, tags, categories, status, created_at, updated_at)
                        VALUES (?,?,?,?,?,?,'stdio','unknown',?,?,?,'discovered',strftime('%s','now'),strftime('%s','now'))""",
                        (srv_uuid, cid, item.get("name","").replace("-"," ").title()[:200], desc[:500],
                         f"https://github.com/{owner}/{repo}",
                         item.get("homepage","") or f"https://github.com/{owner}/{repo}",
                         stars, json.dumps(["mcp", "github-search-new", owner, lang.lower() if lang else "unknown"]),
                         '["more-directories"]'))
                    src_uuid = str(uuid.uuid4())
                    c.execute("INSERT INTO published_mcp_server_sources (uuid,server_uuid,source_name,source_url,raw_payload) VALUES (?,?,?,?,?)",
                        (src_uuid, srv_uuid, "github-search-new", f"https://github.com/{owner}/{repo}", json.dumps({"stars": stars, "lang": lang, "query": query})))
                    n += 1
                    added += 1
            if n:
                print(f"  '{query}' p{page}: +{n}")
            if len(items) < 100:
                break
            time.sleep(1.5)
    
    conn.commit()
    c.execute("SELECT COUNT(*) FROM published_mcp_servers")
    total = c.fetchone()[0]
    conn.close()
    print(f"  [More Directories] +{added} new servers. TOTAL: {total:,}")

# ═══════════════════════════════════════════════════════════════════════════
# PART 5: Add schema column for mcp_server_json if not present
# ═══════════════════════════════════════════════════════════════════════════
def ensure_metadata_columns():
    """Add extra metadata columns for testing if not present."""
    conn = sqlite3.connect(BORG_DB_PATH)
    c = conn.cursor()
    
    # Check and add columns
    c.execute("PRAGMA table_info(published_mcp_servers)")
    existing_cols = {row[1] for row in c.fetchall()}
    
    new_cols = [
        ("language", "TEXT DEFAULT ''"),
        ("mcp_server_json", "TEXT DEFAULT ''"),  # Raw server.json content if found
        ("env_vars_found", "TEXT DEFAULT '[]'"),  # All env var names discovered
        ("has_env_file", "INTEGER DEFAULT 0"),    # Whether .env.example exists
        ("github_topics", "TEXT DEFAULT '[]'"),   # GitHub topic tags
        ("package_name", "TEXT DEFAULT ''"),       # npm/pypi package name
    ]
    
    for col_name, col_def in new_cols:
        if col_name not in existing_cols:
            try:
                c.execute(f"ALTER TABLE published_mcp_servers ADD COLUMN {col_name} {col_def}")
                print(f"  Added column: {col_name}")
            except Exception as e:
                print(f"  Column {col_name} already exists or error: {e}")
    
    conn.commit()
    conn.close()
    print("  Schema migration complete")

if __name__ == "__main__":
    print("=== MCP Metadata Enricher + New Directory Scraper ===")
    ensure_metadata_columns()
    scrape_official_registry()
    scrape_more_directories()
    enrich_smithery()
    # GitHub enrichment is rate-limited but run it for the first 5000
    enrich_github_metadata()
    
    # Final stats
    conn = sqlite3.connect(BORG_DB_PATH)
    c = conn.cursor()
    c.execute("SELECT COUNT(*) FROM published_mcp_servers")
    total = c.fetchone()[0]
    c.execute("SELECT COUNT(*) FROM published_mcp_config_recipes WHERE confidence >= 50")
    high_conf = c.fetchone()[0]
    c.execute("SELECT COUNT(*) FROM published_mcp_servers WHERE auth_model NOT IN ('', 'unknown') AND auth_model IS NOT NULL")
    has_auth = c.fetchone()[0]
    c.execute("SELECT COUNT(*) FROM published_mcp_servers WHERE stars > 0")
    has_stars = c.fetchone()[0]
    c.execute("SELECT COUNT(*) FROM published_mcp_config_recipes WHERE required_secrets != '[]' AND required_secrets IS NOT NULL AND required_secrets != ''")
    has_secrets = c.fetchone()[0]
    conn.close()
    
    print(f"\n{'='*55}")
    print(f"ENRICHMENT COMPLETE")
    print(f"  Total servers: {total:,}")
    print(f"  High-confidence recipes (>=50): {high_conf:,}")
    print(f"  Servers with auth_model: {has_auth:,}")
    print(f"  Servers with stars: {has_stars:,}")
    print(f"  Recipes with required_secrets: {has_secrets:,}")
