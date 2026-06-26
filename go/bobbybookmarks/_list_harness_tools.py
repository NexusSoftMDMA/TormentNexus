#!/usr/bin/env python3
"""Curated list of AI CLI/TUI coding harness tools from the Atlas database."""
import sqlite3, sys
sys.stdout.reconfigure(encoding='utf-8')

atl = sqlite3.connect('atlas.db')
a = atl.cursor()

# DEFINITIVE curated list - only actual standalone AI coding tools
# owner/repo -> (category, tier)
TOOLS = {
    # === TIER 1: The Big Three + Major Standalones ===
    ('anthropics', 'claude-code'): ('Full CLI/TUI Harness', 1),
    ('openai', 'codex'): ('Full CLI/TUI Harness', 1),
    ('google-gemini', 'gemini-cli'): ('Full CLI/TUI Harness', 1),
    ('sst', 'opencode'): ('Full CLI/TUI Harness', 1),
    ('block', 'goose'): ('Full CLI/TUI Harness', 1),
    ('plandex-ai', 'plandex'): ('Full CLI/TUI Harness', 1),
    ('badlogic', 'pi-mono'): ('Full CLI/TUI Harness', 1),
    
    # === TIER 2: Well-known CLI coding tools ===
    ('aider-ai', 'aider'): ('Full CLI/TUI Harness', 2),
    ('paul-gauthier', 'aider'): ('Full CLI/TUI Harness', 2),
    ('OpenInterpreter', 'open-interpreter'): ('AI Interpreter/REPL', 2),
    ('KillianLucas', 'open-interpreter'): ('AI Interpreter/REPL', 2),
    ('danielmiessler', 'fabric'): ('Full CLI/TUI Harness', 2),
    ('TheR1D', 'shell-gpt'): ('Full CLI/TUI Harness', 2),
    ('QwenLM', 'Qwen3-Coder'): ('Full CLI/TUI Harness', 2),
    ('smol-ai', 'GodMode'): ('AI Interpreter/REPL', 2),
    
    # === TIER 2: Major coding agents ===
    ('All-Hands-AI', 'OpenHands'): ('Autonomous Coding Agent', 2),
    ('OpenHands', 'OpenHands'): ('Autonomous Coding Agent', 2),
    ('SWE-agent', 'SWE-agent'): ('Autonomous Coding Agent', 2),
    ('OpenCodeInterpreter', 'OpenCodeInterpreter'): ('Autonomous Coding Agent', 2),
    
    # === TIER 3: Harness orchestrators / meta-tools ===
    ('2mawi2', 'schaltwerk'): ('Harness Orchestrator', 3),
    ('nyldn', 'claude-octopus'): ('Harness Orchestrator', 3),
    ('cluesmith', 'codev'): ('Harness Orchestrator', 3),
    ('SuperClaude-Org', 'SuperClaude_Framework'): ('Skill/Framework', 3),
    ('railyard-dev', 'railguard'): ('Safety Runtime', 3),
    ('neiii', 'bridle'): ('Config Manager', 3),
    ('twaldin', 'harness'): ('Multi-Harness CLI', 3),
    ('mcpware', 'cross-code-organizer'): ('Cross-Harness Dashboard', 3),
    ('kaitranntt', 'ccs'): ('Proxy/Profile Manager', 3),
    ('yitianlian', 'harnessbridge'): ('Config Migrator', 3),
    ('hahaxiang27', 'openHarness'): ('Harness Orchestrator', 3),
    ('HKUDS', 'OpenHarness'): ('Harness Orchestrator', 3),
    ('Cluster444', 'agentic'): ('Harness Framework', 3),
    ('mvschwarz', 'openrig'): ('Multi-Agent Harness', 3),
    ('muqiao215', 'ControlMesh'): ('Control Plane', 3),
    ('runtm-ai', 'runtm-coding-agent-runtime-control-plane'): ('Control Plane', 3),
    ('iOfficeAI', 'AionUi'): ('Unified Agent GUI', 3),
    ('coleam00', 'Archon'): ('Harness Builder', 3),
    ('BloopAI', 'vibe-kanban'): ('Orchestrator', 3),
    ('Chachamaru127', 'claude-code-harness'): ('Harness Framework', 3),
    ('mindfold-ai', 'Trellis'): ('Harness Framework', 3),
    ('BA-CalderonMorales', 'agent-harness'): ('Go Harness', 3),
    ('aayoawoyemi', 'Aries-cli'): ('CLI Harness', 3),
    ('RealZST', 'HarnessKit'): ('Harness Toolkit', 3),
    ('hyspacex', 'harness-cli'): ('Harness CLI', 3),
    ('keli-wen', 'agentic-harness-patterns-skill'): ('Harness Patterns', 3),
    ('gotalab', 'cc-sdd'): ('SDD Harness', 3),
    ('starbased-co', 'ccproxy'): ('Proxy Layer', 3),
    ('Biajin-PKU', 'research-harness'): ('Research Harness', 3),
    ('wangrenzhu-ola', 'GaleHarnessCodingCLI'): ('Harness CLI', 3),
    
    # === TIER 3: Extensions, plugins, frameworks ===
    ('anthropics', 'skills'): ('Skill Framework', 3),
    ('anthropics', 'mcpb'): ('MCP Bundler', 3),
    ('openai', 'codex-plugin-cc'): ('Codex-Claude Plugin', 3),
    ('openai', 'symphony'): ('Agent Pipeline', 3),
    ('github', 'spec-kit'): ('Spec Framework', 3),
    ('liberzon', 'claude-hooks'): ('Hooks Framework', 3),
    ('AnandChowdhary', 'continuous-claude'): ('Autonomous PR', 3),
    ('badrisnarayanan', 'antigravity-claude-proxy'): ('API Proxy', 3),
    ('shekohex', 'opencode-pty'): ('PTY Manager', 3),
    ('Opencode-DCP', 'opencode-dynamic-context-pruning'): ('Context Pruner', 3),
    ('zeddy89', 'Context-Engine'): ('Context Engine', 3),
    ('dagger', 'container-use'): ('Container Sandbox', 3),
    ('kitfunso', 'hippo-memory'): ('Memory System', 3),
    ('cyberchitta', 'llm-context.py'): ('Context Manager', 3),
    ('manuelschipper', 'nah'): ('Permission Guard', 3),
    ('goodfylink', 'Claude-Code'): ('Source Study', 3),
    ('bmad-code-org', 'BMAD-METHOD'): ('Workflow/Method', 3),
    ('theredsix', 'agent-browser-protocol'): ('Browser Protocol', 3),
    ('anomalyco', 'opencode'): ('OpenCode Fork', 3),
    ('jaehongpark-agent', 'claude-code-spinner-verbs'): ('UI Customizer', 3),
    ('NeuralNomadsAI', 'CodeNomad'): ('Multi-Instance Cockpit', 3),
    
    # === Gemini CLI extensions (important ecosystem) ===
    ('gemini-cli-extensions', 'conductor'): ('Gemini CLI Extension', 3),
    ('Jasonzhangf', 'gemini-cli-router'): ('Gemini CLI Proxy', 3),
    
    # === IDE-based AI tools (important context) ===
    ('RooCodeInc', 'Roo-Code'): ('IDE Plugin (VS Code)', 3),
}

# Fetch data from atlas
all_tools = []
seen = set()
for (owner, repo), (cat, tier) in TOOLS.items():
    lookup_key = f'{owner.lower()}/{repo.lower()}'
    if lookup_key in seen:
        continue
    seen.add(lookup_key)
    
    # Try exact match
    a.execute("""
        SELECT e.url, e.page_title, e.short_description, e.main_features, 
               e.innovation, e.quality, e.signal, e.owner, e.repo
        FROM entries e
        WHERE LOWER(e.owner) = ? AND LOWER(e.repo) = ?
        ORDER BY e.signal DESC LIMIT 1
    """, (owner.lower(), repo.lower()))
    row = a.fetchone()
    
    if not row:
        # Try just repo
        a.execute("""
            SELECT e.url, e.page_title, e.short_description, e.main_features, 
                   e.innovation, e.quality, e.signal, e.owner, e.repo
            FROM entries e
            WHERE LOWER(e.repo) = ?
            ORDER BY e.signal DESC LIMIT 1
        """, (repo.lower(),))
        row = a.fetchone()
    
    if row:
        url, pt, sd, feats, innov, qual, sig, o, r = row
        all_tools.append((tier, cat, o, r, url, sd, feats, innov, qual, sig, True))
    else:
        url = f'https://github.com/{owner}/{repo}'
        all_tools.append((tier, cat, owner, repo, url, '', '', 0, 0, 0, False))

# Sort: tier first, then by signal within tier
all_tools.sort(key=lambda x: (x[0], -x[9]))

# Print report
print('=' * 80)
print('  AI CLI/TUI CODING HARNESS TOOLS - COMPREHENSIVE INDEX')
print('  From the Borg Intelligence Atlas (atlas.db)')
print('=' * 80)

tier_names = {1: 'TIER 1 - Major Standalone CLI/TUI AI Coding Tools',
              2: 'TIER 2 - Well-Known CLI Tools & Autonomous Agents',
              3: 'TIER 3 - Ecosystem: Orchestrators, Frameworks, Plugins'}

current_tier = None
current_cat = None
i = 0

for tier, cat, o, r, url, sd, feats, innov, qual, sig, in_atlas in all_tools:
    if tier != current_tier:
        current_tier = tier
        current_cat = None
        tier_count = sum(1 for t in all_tools if t[0] == tier)
        print()
        print()
        print('#' * 80)
        print(f'  {tier_names[tier]} ({tier_count} tools)')
        print('#' * 80)
    
    if cat != current_cat:
        current_cat = cat
        cat_count = sum(1 for t in all_tools if t[0] == tier and t[1] == cat)
        print()
        print(f'  == {cat} ({cat_count}) ==')
        print()
    
    i += 1
    name = f'{o}/{r}'
    desc = (sd[:120] + '...') if sd and len(sd) > 120 else (sd or 'No description available')
    feat_str = (feats[:100] + '...') if feats and len(feats) > 100 else (feats or '')
    
    if in_atlas:
        print(f'  {i:3d}. [{name}]({url}) | I:{innov:.0f} Q:{qual:.2f} Sig:{sig:.0f}')
    else:
        print(f'  {i:3d}. [{name}]({url}) | *(not in atlas - add to incoming_resources.txt)*')
    print(f'      {desc}')
    if feat_str:
        print(f'      Features: {feat_str}')
    print()

# Summary
print()
print('=' * 80)
print('  SUMMARY')
print('=' * 80)
t1 = sum(1 for t in all_tools if t[0] == 1)
t2 = sum(1 for t in all_tools if t[0] == 2)
t3 = sum(1 for t in all_tools if t[0] == 3)
in_db = sum(1 for t in all_tools if t[10])
missing = sum(1 for t in all_tools if not t[10])
print(f'  Tier 1 (Major Standalones): {t1}')
print(f'  Tier 2 (Well-Known/Agents): {t2}')
print(f'  Tier 3 (Ecosystem/Plugins): {t3}')
print(f'  Total: {t1+t2+t3}')
print(f'  In Atlas: {in_db}')
print(f'  Missing from Atlas: {missing}')

# List missing tools
print()
print('  Missing tools to add to incoming_resources.txt:')
for tier, cat, o, r, url, sd, feats, innov, qual, sig, in_atlas in all_tools:
    if not in_atlas:
        print(f'    https://github.com/{o}/{r}')

atl.close()
