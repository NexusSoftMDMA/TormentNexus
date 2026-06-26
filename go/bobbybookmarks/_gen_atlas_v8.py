#!/usr/bin/env python3
"""
Borg Intelligence Atlas v8.0 — Markdown Layer Exporter
Regenerates all layer markdown files from atlas.db
"""
import sqlite3, sys, os, time
from collections import defaultdict
sys.stdout.reconfigure(encoding='utf-8')

DB = 'atlas.db'
OUTDIR = '.'
DATE = time.strftime('%Y-%m-%d')

# Layer emoji/icon mapping
LAYER_ICONS = {
    'Agent Orchestration & Workflow': ('🧠', 'AGENT_ORCHESTRATION_WORKFLOW'),
    'Context Engineering & Isolation': ('👁', 'CONTEXT_ENGINEERING_ISOLATION'),
    'Memory & Persistence Architecture': ('🧬', 'MEMORY_PERSISTENCE_ARCHITECTURE'),
    'Interface & Developer UX': ('🤳', 'INTERFACE_DEVELOPER_UX'),
    'Connectivity / MCP / A2A': ('⚡', 'CONNECTIVITY_MCP_A2A'),
    'Infrastructure & Proxy Layers': ('🦴', 'INFRASTRUCTURE_PROXY_LAYERS'),
    'Guides & Industry Trends': ('🗺', 'GUIDES_INDUSTRY_TRENDS'),
    'Coding Harness Tools': ('🛠', 'CODING_HARNESS_TOOLS'),
    'AI Agents & Frameworks': ('🤖', 'AI_AGENTS_FRAMEWORKS'),
    'Search & Discovery': ('🔍', 'SEARCH_DISCOVERY'),
    'Coding Tools & IDEs': ('💻', 'CODING_TOOLS_IDES'),
    'Developer Workflow & Tools': ('🔧', 'DEVELOPER_WORKFLOW_TOOLS'),
    'Vector Databases & Embeddings': ('📐', 'VECTOR_DATABASES_EMBEDDINGS'),
    'Security & Red Teaming': ('🛡', 'SECURITY_RED_TEAMING'),
}

LAYER_DESCRIPTIONS = {
    'Agent Orchestration & Workflow': 'Multi-agent swarms, workflows, planning, loops, verification',
    'Context Engineering & Isolation': 'Context compression, codebase indexing, RAG, isolation, ingestion',
    'Memory & Persistence Architecture': 'Graph memory, episodic, semantic, MCP memory, second brain, memory OS',
    'Interface & Developer UX': 'Computer-use agents, terminal UIs, IDEs, web dashboards, voice, canvas',
    'Connectivity / MCP / A2A': 'MCP infrastructure, A2A, gateways, tool discovery, registries',
    'Infrastructure & Proxy Layers': 'AI OSes, inference engines, sandboxes, security, deployment, LLM routers',
    'Guides & Industry Trends': 'Awesome lists, tutorials, architecture patterns, benchmarks',
    'Coding Harness Tools': 'Agent harnesses, skills, governance, spec-driven dev, bridges',
    'AI Agents & Frameworks': 'Coding agents, GUI agents, research agents, AI OS, security agents',
    'Search & Discovery': 'Semantic search, web APIs, code search, MCP registries',
    'Coding Tools & IDEs': 'AI editors, autocomplete, code review, refactoring, testing',
    'Developer Workflow & Tools': 'Git, CI/CD, project management, documentation',
    'Vector Databases & Embeddings': 'Vector DBs, embedding models, ANN indexes, RAG frameworks',
    'Security & Red Teaming': 'AI guardrails, LLM red teaming, vulnerability scanning, pentesting',
}


def signal_badge(sig):
    if sig >= 85:
        return '🏆 World-class'
    elif sig >= 70:
        return '⭐ Excellent'
    elif sig >= 50:
        return '✓ Solid'
    elif sig >= 30:
        return '○ Adequate'
    else:
        return '⚠ Thin'


def signal_icon(sig):
    if sig >= 85:
        return '🏆'
    elif sig >= 70:
        return '⭐'
    elif sig >= 50:
        return '✓'
    elif sig >= 30:
        return '○'
    else:
        return '⚠'


def bar(count, max_count, width=20):
    filled = int(count / max(max_count, 1) * width)
    return '█' * filled + '░' * (width - filled)


def gen_layer_md(atl, layer_name):
    """Generate markdown for a single layer."""
    a = atl.cursor()
    icon, fname = LAYER_ICONS.get(layer_name, ('📄', layer_name.replace(' ', '_').replace('/', '_').upper()))
    desc = LAYER_DESCRIPTIONS.get(layer_name, '')
    
    # Get all entries in this layer
    a.execute('''
        SELECT e.id, e.url, e.page_title, e.short_description, e.long_description,
            e.main_features, e.innovation, e.quality, e.signal,
            e.owner, e.repo, e.tags, e.is_standout, e.verdict,
            (SELECT GROUP_CONCAT(lm2.layer, '|') FROM layer_membership lm2 WHERE lm2.entry_id = e.id AND lm2.is_primary = 0) as other_layers
        FROM entries e
        JOIN layer_membership lm ON lm.entry_id = e.id AND lm.is_primary = 1
        WHERE lm.layer = ?
        ORDER BY e.signal DESC
    ''', (layer_name,))
    entries = a.fetchall()
    
    if not entries:
        return None
    
    # Stats
    total = len(entries)
    standout_count = sum(1 for e in entries if e[12])
    avg_signal = sum(e[8] for e in entries) / total
    # Aggregate innovation across ALL entries (not per-subcat)
    inn_dist = defaultdict(int)
    for e in entries:
        inn_rounded = int(e[6])
        inn_dist[inn_rounded] += 1
    max_inn = max(inn_dist.values()) if inn_dist else 1
    
    # Smarter subcategorization: derive from tool content, not tags
    import json
    
    SUBCAT_RULES = [
        # (pattern_in_name_or_desc, subcategory_name)
        ('memory|persist|memsearch|hippo', 'Memory & Context Systems'),
        ('govern|guard|rail|permission|policy|safety|sandbox|isolat', 'Governance & Safety'),
        ('skill|slash-command|command', 'Skill Systems'),
        ('hook|lifecycle|trigger|event', 'Hooks & Lifecycle'),
        ('spec|sdd|plan|constitution|specify', 'Spec-Driven Development'),
        ('orchestrat|coordinat|dispatch|swarm|multi-agent|parallel', 'Orchestration'),
        ('monitor|analyt|dashboard|metrics|log|trace|observ', 'Monitoring & Analytics'),
        ('proxy|rout|bridge|translat|relay|gateway|mcp-server', 'Bridges & Proxies'),
        ('harness|framework|agent-harness|toolkit|builder|archit', 'Harness Frameworks'),
        ('config|settings|profile|manage|bridle', 'Config & Profile Management'),
        ('context|prune|compact|compress|engine', 'Context Engineering'),
        ('test|verif|bench|review|quality|lint', 'Verification & Testing'),
        ('code-agent|coding-agent|cli-agent|terminal', 'CLI Coding Agents'),
        ('claude-code|codex|opencode|gemini|goose|aider|copilot', 'Major Harness Integrations'),
        ('browser|web|scrape|crawl', 'Browser & Web Tools'),
    ]
    
    import re
    tag_subcats = defaultdict(list)
    for e in entries:
        text = f'{(e[9] or "")}/{(e[10] or "")} {(e[3] or "")} {(e[4] or "")} {(e[5] or "")}'.lower()
        matched = False
        for pattern, subcat_name in SUBCAT_RULES:
            if re.search(pattern, text):
                tag_subcats[subcat_name].append(e)
                matched = True
                break
        if not matched:
            tag_subcats['Other Tools'].append(e)
    
    # Merge small subcats (< 2 entries) into 'Other Tools'
    subcats = {}
    other_entries = []
    for tag, sub_entries in tag_subcats.items():
        if len(sub_entries) >= 2 and tag != 'Other Tools':
            subcats[tag] = sub_entries
        elif tag == 'Other Tools':
            other_entries.extend(sub_entries)  # keep these for later
        else:
            other_entries.extend(sub_entries)
    # Combine any existing 'Other Tools' entries too
    if other_entries:
        subcats['Other Tools'] = sorted(other_entries, key=lambda x: -x[8])
    
    # Sort subcats by count desc
    subcats_sorted = sorted(subcats.items(), key=lambda x: -len(x[1]))
    
    # Build markdown
    lines = []
    lines.append(f'# {icon} {layer_name}')
    lines.append(f'> Borg Intelligence Atlas v8 · {DATE} · {total:,} tools')
    lines.append(f'> {desc}')
    lines.append('')
    
    # Stats table
    lines.append('| Metric | Value |')
    lines.append('|--------|-------|')
    lines.append(f'| Total tools | **{total:,}** |')
    lines.append(f'| Standout 🏆⭐ | {standout_count} |')
    lines.append(f'| Avg Signal | ⚡{avg_signal:.0f} |')
    
    for inn in sorted(inn_dist.keys(), reverse=True):
        cnt = inn_dist[inn]
        lines.append(f'| Innovation {inn:.0f} | {cnt} {bar(cnt, max_inn, 20)} |')
    
    lines.append('')
    lines.append('---')
    lines.append('')
    
    # Top 20
    lines.append('## 🏆 Top 20 by Signal Strength')
    lines.append('')
    for i, e in enumerate(entries[:20], 1):
        eid, url, pt, sd, ld, feats, inn, qual, sig, owner, repo, tags, is_so, verdict, other_layers = e
        name = f'{owner}/{repo}' if owner and repo else (pt or url.split('/')[-1])
        badge = signal_badge(sig)
        desc_short = (sd[:100] + '...') if sd and len(sd) > 100 else (sd or 'No description')
        lines.append(f'{i}. **[{name}]({url})** ⚡{sig:.0f} · {badge} — {desc_short}')
    
    lines.append('')
    lines.append('---')
    lines.append('')
    
    # Subcategory TOC
    lines.append('## Contents')
    lines.append('')
    for subcat_name, subcat_entries in subcats_sorted:
        sc_avg = sum(e[8] for e in subcat_entries) / len(subcat_entries)
        anchor = subcat_name.lower().replace(' ', '-').replace('&', '').replace('/', '-')
        lines.append(f'- [{subcat_name.title()}](#{anchor}) — {len(subcat_entries)} tools · ⚡{sc_avg:.0f}')
    
    lines.append('')
    lines.append('---')
    lines.append('')
    
    # Full entries by subcategory
    for subcat_name, subcat_entries in subcats_sorted:
        sc_avg = sum(e[8] for e in subcat_entries) / len(subcat_entries)
        anchor = subcat_name.lower().replace(' ', '-').replace('&', '').replace('/', '-')
        lines.append(f'## {subcat_name.title()}')
        lines.append(f'> {len(subcat_entries)} tools · avg signal ⚡{sc_avg:.0f}')
        lines.append('')
        
        for i, e in enumerate(subcat_entries, 1):
            eid, url, pt, sd, ld, feats, inn, qual, sig, owner, repo, tags, is_so, verdict, other_layers = e
            name = f'{owner}/{repo}' if owner and repo else (pt or url.split('/')[-1])
            badge = signal_badge(sig)
            q_display = f'Q{qual:.1f}' if qual >= 0.1 else 'Q0.0'
            other_count = len((other_layers or '').split('|')) if other_layers else 0
            
            # Header
            lines.append(f'### {i}. [{name}]({url})')
            lines.append(f'`{inn:.1f}` {"★★★" if inn >= 10 else "★★" if inn >= 9 else "★"} ⚡{sig:.0f} {q_display}{signal_icon(sig)} {badge}')
            if other_count > 0:
                lines.append(f'↗{other_count + 1} layers')
            lines.append('')
            
            # Description
            if ld:
                desc_text = ld[:500] + ('...' if len(ld) > 500 else '')
            elif sd:
                desc_text = sd
            else:
                desc_text = 'No description available.'
            lines.append(f'**{desc_text}**')
            lines.append('')
            
            # Features
            if feats:
                feat_list = [f.strip() for f in feats.split(',') if f.strip()]
                lines.append('**Features:**')
                for feat in feat_list[:12]:
                    lines.append(f'- {feat}')
                if len(feat_list) > 12:
                    lines.append(f'- ... and {len(feat_list) - 12} more')
                lines.append('')
            
            # Tags
            tag_list_raw = e[11] or '[]'
            try:
                tag_list = json.loads(tag_list_raw) if tag_list_raw.startswith('[') else [t.strip() for t in tag_list_raw.split(',') if t.strip()]
            except (json.JSONDecodeError, TypeError):
                tag_list = [t.strip() for t in tag_list_raw.split(',') if t.strip()]
            if tag_list:
                lines.append(f'*Tags: {", ".join(tag_list[:8])}{"..." if len(tag_list) > 8 else ""}*')
                lines.append('')
            
            lines.append('---')
            lines.append('')
    
    return '\n'.join(lines), fname, total, standout_count, avg_signal


def gen_index_md(layer_stats):
    """Generate the master index markdown."""
    lines = []
    total_all = sum(s[1] for s in layer_stats)
    standout_all = sum(s[2] for s in layer_stats)
    
    lines.append(f'# Borg Intelligence Atlas v8 — Master Index')
    lines.append(f'> **{total_all:,}** tools · **{standout_all:,}** standout 🏆⭐ · Domain-classified · Signal-scored · {DATE}')
    lines.append('')
    lines.append('## The 7 Borg Categories')
    lines.append('')
    lines.append('| # | Layer | Tools | Standout | ⚡ | Description |')
    lines.append('|---|-------|-------|----------|-----|-------------|')
    
    # Primary 7
    primary_layers = [
        'Agent Orchestration & Workflow',
        'Context Engineering & Isolation',
        'Memory & Persistence Architecture',
        'Interface & Developer UX',
        'Connectivity / MCP / A2A',
        'Infrastructure & Proxy Layers',
        'Guides & Industry Trends',
    ]
    
    for i, layer_name in enumerate(primary_layers, 1):
        icon, fname = LAYER_ICONS.get(layer_name, ('📄', layer_name))
        desc = LAYER_DESCRIPTIONS.get(layer_name, '')
        stats = next((s for s in layer_stats if s[0] == layer_name), None)
        if stats:
            _, count, standout, avg_sig = stats
            lines.append(f'| {i} | {icon} [{layer_name}]({fname}.md) | **{count:,}** | {standout} | ⚡{avg_sig:.0f} | {desc} |')
    
    lines.append('')
    lines.append('## Cross-Cutting Domains')
    lines.append('')
    lines.append('| # | Domain | Tools | Standout | ⚡ | Description |')
    lines.append('|---|--------|-------|----------|-----|-------------|')
    
    cross_layers = [
        'Coding Harness Tools',
        'AI Agents & Frameworks',
        'Search & Discovery',
        'Coding Tools & IDEs',
        'Developer Workflow & Tools',
        'Vector Databases & Embeddings',
        'Security & Red Teaming',
    ]
    
    for idx, layer_name in enumerate(cross_layers):
        i = idx + 8
        icon, fname = LAYER_ICONS.get(layer_name, ('📄', layer_name))
        desc = LAYER_DESCRIPTIONS.get(layer_name, '')
        stats = next((s for s in layer_stats if s[0] == layer_name), None)
        if stats:
            _, count, standout, avg_sig = stats
            lines.append(f'| {i} | {icon} [{layer_name}]({fname}.md) | **{count:,}** | {standout} | ⚡{avg_sig:.0f} | {desc} |')
    
    lines.append('')
    lines.append('---')
    lines.append('')
    lines.append('## Signal Strength (⚡)')
    lines.append('')
    lines.append('**Signal** (0-100) answers: *"Is this tool actually worth my time?"*')
    lines.append('')
    lines.append('| Component | Weight | Measures |')
    lines.append('|-----------|--------|----------|')
    lines.append('| Innovation × 4 | 0-40 | Raw innovation from LLM analysis |')
    lines.append('| Quality × 30 | 0-30 | Description depth, feature count, tags, verdict, owner |')
    lines.append('| Feature richness | 0-15 | Concrete features listed |')
    lines.append('| Description depth | 0-10 | How detailed the description is |')
    lines.append('| GitHub trust | 0-5 | Open-source repo bonus |')
    lines.append('')
    lines.append('| Range | Meaning |')
    lines.append('|-------|----------|')
    lines.append('| ⚡85+ | 🏆 Must-know — world-class |')
    lines.append('| ⚡70-84 | ⭐ Excellent — highly recommended |')
    lines.append('| ⚡50-69 | ✓ Solid — worth exploring |')
    lines.append('| ⚡30-49 | ○ Adequate — has useful features |')
    lines.append('| ⚡0-29 | ⚠ Thin — limited data |')
    
    return '\n'.join(lines)


def main():
    atl = sqlite3.connect(DB)
    
    # Get all layers
    a = atl.cursor()
    a.execute("SELECT DISTINCT layer FROM layer_membership ORDER BY layer")
    layers = [row[0] for row in a.fetchall()]
    
    print(f'Regenerating {len(layers)} layer markdown files from {DB}...')
    print()
    
    layer_stats = []
    
    for layer_name in layers:
        result = gen_layer_md(atl, layer_name)
        if result is None:
            print(f'  SKIP: {layer_name} (no entries)')
            continue
        
        md_content, fname, total, standout, avg_signal = result
        filepath = os.path.join(OUTDIR, f'{fname}.md')
        
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(md_content)
        
        layer_stats.append((layer_name, total, standout, avg_signal))
        size_kb = len(md_content.encode('utf-8')) / 1024
        print(f'  {fname}.md: {total:,} tools, {standout} standout, avg signal {avg_signal:.0f}, {size_kb:.0f}KB')
    
    # Sort layer stats by count desc for index
    layer_stats.sort(key=lambda x: -x[1])
    
    # Generate master index
    index_md = gen_index_md(layer_stats)
    index_path = os.path.join(OUTDIR, 'BORG_ATLAS_INDEX.md')
    with open(index_path, 'w', encoding='utf-8') as f:
        f.write(index_md)
    print(f'\n  BORG_ATLAS_INDEX.md: {sum(s[1] for s in layer_stats):,} total tools across {len(layer_stats)} layers')
    
    atl.close()
    print('\nDone!')


if __name__ == '__main__':
    main()
