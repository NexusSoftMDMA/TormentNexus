#!/usr/bin/env python3
"""Final curated report: AI CLI/TUI Coding Harness Tools"""
import sqlite3, sys
sys.stdout.reconfigure(encoding='utf-8')

atl = sqlite3.connect('atlas.db')
a = atl.cursor()

a.execute('SELECT COUNT(*) FROM entries'); total = a.fetchone()[0]
a.execute("SELECT COUNT(*) FROM entries WHERE short_description IS NOT NULL AND short_description != ''"); enriched = a.fetchone()[0]

print('=' * 80)
print('  AI CLI/TUI CODING HARNESS TOOLS - COMPREHENSIVE INDEX')
print(f'  Borg Intelligence Atlas | {total:,} entries, {enriched:,} enriched')
print('=' * 80)

def get_sig(owner, repo):
    a.execute('SELECT e.signal, e.quality, e.innovation FROM entries e WHERE LOWER(e.owner)=? AND LOWER(e.repo)=? ORDER BY e.signal DESC LIMIT 1', (owner.lower(), repo.lower()))
    row = a.fetchone()
    if row:
        return f'Sig:{row[0]:.0f} Q:{row[1]:.2f} I:{row[2]:.0f}'
    return '(not in atlas)'

# TIER 1
print()
print('  TIER 1: MAJOR STANDALONE CLI/TUI AI CODING TOOLS')
print('  (Install and run in a terminal as your AI coding partner)')
print('  ' + '-' * 72)
print()

tier1 = [
    ('anthropics', 'claude-code', 'Anthropic official CLI agent. Pair programming with Claude in terminal. MCP, skills, plugins, hooks. Gold standard.'),
    ('openai', 'codex', 'OpenAI official CLI agent. Rust-based, terminal-native, MCP integration, sandboxed execution, multi-model.'),
    ('google-gemini', 'gemini-cli', 'Google official CLI agent. Full terminal coding with Gemini, extensible via extensions/MCP/hooks. 1M context.'),
    ('sst', 'opencode', 'Open-source provider-agnostic TUI coding agent. Client/server, built-in plan/build agents, LSP support. Community champion.'),
    ('block', 'goose', 'Open-source Rust-based autonomous AI agent. MCP-native, multi-model, local execution. Backed by Block (Square).'),
    ('badlogic', 'pi-mono', 'Full coding agent framework. Custom provider registration, multi-model routing, secure auth, mono-repo.'),
    ('plandex-ai', 'plandex', 'Context agent: 2M token effective context, 20M+ repo indexing, diff review sandbox, multi-model pipeline.'),
]

for i, (o, r, desc) in enumerate(tier1, 1):
    sig = get_sig(o, r)
    print(f'  {i}. [{o}/{r}](https://github.com/{o}/{r})')
    print(f'     {desc}')
    print(f'     {sig}')
    print()

# TIER 2
print()
print('  TIER 2: WELL-KNOWN CLI TOOLS & AUTONOMOUS CODING AGENTS')
print('  ' + '-' * 72)
print()

print('  -- AI Pair Programming CLIs --')
print()

tier2_cli = [
    ('paul-gauthier', 'aider', 'AI pair programming in terminal. Git-integrated, multi-model, 50+ LLMs. The OG AI coding CLI.'),
    ('danielmiessler', 'fabric', 'AI pattern framework for content extraction/summarization/transformation. Widely used, not a coding agent per se.'),
    ('TheR1D', 'shell-gpt', 'GPT CLI for shell commands, code generation, chat. Lightweight and fast. sgpt command.'),
    ('QwenLM', 'Qwen3-Coder', 'Alibaba coding agent. 80B/3B active params, 1M token context, execution-guided RL training.'),
]
for i, (o, r, desc) in enumerate(tier2_cli, 8):
    sig = get_sig(o, r)
    print(f'  {i}. [{o}/{r}](https://github.com/{o}/{r})')
    print(f'     {desc}')
    print(f'     {sig}')
    print()

print('  -- AI Interpreters / REPLs --')
print()

tier2_interp = [
    ('OpenInterpreter', 'open-interpreter', 'Open-source Code Interpreter. LLMs execute code locally across Python/JS/Ruby/etc. Full system access.'),
    ('smol-ai', 'GodMode', 'Multi-model GUI for simultaneous prompting. Native web features, PromptCritic. More GUI than CLI.'),
]
for i, (o, r, desc) in enumerate(tier2_interp, 12):
    sig = get_sig(o, r)
    print(f'  {i}. [{o}/{r}](https://github.com/{o}/{r})')
    print(f'     {desc}')
    print(f'     {sig}')
    print()

print('  -- Autonomous Coding Agents --')
print()

tier2_agents = [
    ('All-Hands-AI', 'OpenHands', 'Autonomous software engineering agent. Recursive delegation, Docker sandboxing. Formerly OpenDevin.'),
    ('SWE-agent', 'SWE-agent', 'Princeton NLP autonomous bug fixer. 57.5% SWE-bench. Specialized search subagent. Research-grade.'),
    ('OpenCodeInterpreter', 'OpenCodeInterpreter', 'Open-source code generation suite. Iterative refinement with compiler diagnostics. 33B flagship model.'),
    ('AbanteAI', 'mentat', 'AI coding agent working directly in codebase. GitHub Issue integration, code review, autonomous PRs.'),
]
for i, (o, r, desc) in enumerate(tier2_agents, 14):
    sig = get_sig(o, r)
    print(f'  {i}. [{o}/{r}](https://github.com/{o}/{r})')
    print(f'     {desc}')
    print(f'     {sig}')
    print()

# TIER 3
print()
print('  TIER 3: ECOSYSTEM - ORCHESTRATORS, FRAMEWORKS, PLUGINS')
print('  (Meta-tools that augment, manage, or connect Tier 1/2 tools)')
print('  ' + '-' * 72)
print()

tier3 = [
    ('2mawi2', 'schaltwerk', 'Harness Orchestrator', 'IDE-like TUI for orchestrating multiple AI agents with spec-driven dev. Terminal native.'),
    ('cluesmith', 'codev', 'Harness Orchestrator', 'AI-native OS for dev: Issues -> tested PRs. SPIR Protocol, Agent Farm for parallel builds.'),
    ('nyldn', 'claude-octopus', 'Harness Orchestrator', 'Multi-AI orchestration, consensus gates, persistent memory, Dark Factory autonomous mode.'),
    ('mcpware', 'cross-code-organizer', 'Cross-Harness Dashboard', 'Cross-harness config dashboard for Claude Code, Codex, MCP servers, skills.'),
    ('railyard-dev', 'railguard', 'Safety Runtime', 'Safe runtime for Claude Code: tool call interception, memory safety, behavioral blocks.'),
    ('SuperClaude-Org', 'SuperClaude_Framework', 'Skill Framework', '30 slash commands, 16 cognitive agents, behavioral instructions for Claude Code.'),
    ('kaitranntt', 'ccs', 'Proxy/Profile Mgr', 'Multi-account isolation, OAuth proxy, Anthropic API translation, local LLM support.'),
    ('starbased-co', 'ccproxy', 'Proxy Layer', 'LiteLLM proxy intercepting Claude Code requests for dynamic model routing.'),
    ('RealZST', 'HarnessKit', 'Harness Toolkit', 'Manage skills/MCP/plugins/hooks/CLIs/configs/memory across all AI coding agents.'),
    ('yitianlian', 'harnessbridge', 'Config Migrator', 'Convert rules/skills/hooks/memory/MCP configs between Claude Code/Cursor/Windsurf.'),
    ('twaldin', 'harness', 'Multi-Harness CLI', 'Unified Python interface: claude-code/opencode/codex/gemini/aider as subprocesses.'),
    ('neiii', 'bridle', 'Config Manager', 'TUI/CLI config manager for: Amp, Claude Code, Opencode, Goose, Copilot CLI, Crush, Droid.'),
    ('hahaxiang27', 'openHarness', 'Harness Orchestrator', 'SDD spec-driven + multi-agent across OpenCode/Claude Code/Codex.'),
    ('Cluster444', 'agentic', 'Harness Framework', 'Structured /thoughts, phased loops, specialized subagent delegation.'),
    ('muqiao215', 'ControlMesh', 'Control Plane', 'Runtime-first harness for coding CLIs, chat transports, background tasks.'),
    ('runtm-ai', 'runtm-coding-agent-runtime-control-plane', 'Control Plane', 'Ephemeral app lifecycle, human-in-the-loop infra approvals.'),
    ('iOfficeAI', 'AionUi', 'Unified Agent GUI', 'Multi-agent mode (auto-detects CLIs), zero-setup engine, full filesystem ops.'),
    ('anthropics', 'skills', 'Skill Framework', 'Official Anthropic modular skills. YAML-based discovery, dynamic instruction loading.'),
    ('anthropics', 'mcpb', 'MCP Bundler', 'Official MCPB format: bundle MCP servers into portable zips for one-click install.'),
    ('openai', 'codex-plugin-cc', 'Codex-Claude Plugin', 'Codex integration in Claude Code. Review mode, adversarial review, rescue tasks.'),
    ('github', 'spec-kit', 'Spec Framework', 'Executable specs with /specify and /plan, Project Constitution enforcement.'),
    ('liberzon', 'claude-hooks', 'Hooks Framework', 'Lifecycle triggers (BeforeCommit/PostCompact), context re-injection on compaction.'),
    ('dagger', 'container-use', 'Container Sandbox', 'Containerized environments for coding agents with preferred stacks.'),
    ('AnandChowdhary', 'continuous-claude', 'Autonomous PR', 'PR lifecycle: persistent state, Git worktree parallelism, CI failure recovery.'),
    ('BA-CalderonMorales', 'agent-harness', 'Go Harness', 'Clean-room Go impl of agentic harness patterns from production architectures.'),
    ('aayoawoyemi', 'Aries-cli', 'CLI Harness', 'Agentic coding harness with persistent memory and REPL body (Ori Mnemos).'),
    ('shekohex', 'opencode-pty', 'PTY Manager', 'Background process control, regex terminal filtering, persistent sessions.'),
    ('Opencode-DCP', 'opencode-dynamic-context-pruning', 'Context Pruner', 'Tool-call dedup, stale error removal, agent-driven context discarding.'),
    ('zeddy89', 'Context-Engine', 'Context Engine', 'Four-layer context architecture, automated state restoration, /compact hooks.'),
    ('kitfunso', 'hippo-memory', 'Memory System', 'Biologically-inspired: decay, retrieval strengthening, consolidation, cross-tool memory.'),
    ('cyberchitta', 'llm-context.py', 'Context Manager', 'Smart file selection, context formatting, AI agent context generation.'),
    ('manuelschipper', 'nah', 'Permission Guard', 'Deterministic action classifier, .env blocking, LLM-as-a-judge escalation.'),
    ('bmad-code-org', 'BMAD-METHOD', 'Workflow/Method', '12+ personas, atomic story sharding, YAML agent defs, scale-adaptive flows.'),
    ('BloopAI', 'vibe-kanban', 'Orchestrator', 'Parallel agent execution, isolated worktrees, inline diff review, browser preview.'),
    ('coleam00', 'Archon', 'Harness Builder', 'First open-source harness builder. Make AI coding deterministic and repeatable.'),
    ('NeuralNomadsAI', 'CodeNomad', 'Multi-Instance Cockpit', 'Desktop cockpit UI for managing parallel coding agent instances.'),
    ('gemini-cli-extensions', 'conductor', 'Gemini CLI Extension', 'Orchestrates dev lifecycle: spec -> plan -> implement -> review for Gemini CLI.'),
    ('Jasonzhangf', 'gemini-cli-router', 'Gemini CLI Proxy', 'Route Gemini CLI requests to third-party providers (OpenAI, Claude, etc.).'),
    ('RooCodeInc', 'Roo-Code', 'IDE Plugin (VS Code)', 'AI agent team in VS Code. Multi-mode: Code, Architect, Ask, Debug.'),
]

for i, (o, r, cat, desc) in enumerate(tier3, 18):
    sig = get_sig(o, r)
    print(f'  {i:2d}. [{o}/{r}](https://github.com/{o}/{r}) [{cat}]')
    print(f'     {desc}')
    print(f'     {sig}')
    print()

# COMMERCIAL
print()
print('  NOTABLE COMMERCIAL TOOLS (not open source / not in atlas)')
print('  ' + '-' * 72)
print()

commercial = [
    ('Cline (VS Code)', 'AI coding agent in VS Code. Formerly Claude Dev. Autonomous file editing, browser, MCP.'),
    ('Cursor IDE', 'AI-native VS Code fork. Tab completion, chat, multi-file edits, background agents. Most popular AI IDE.'),
    ('Windsurf (Codeium)', 'AI IDE with Cascade agentic flow. Multi-step reasoning, context awareness.'),
    ('Kiro (Amazon)', 'Spec-driven AI IDE. CLI + IDE modes, spec generation, steering docs. AWS.'),
    ('Trae (ByteDance)', 'AI IDE with Builder/Switcher modes. Free tier with Claude/GPT.'),
    ('Augment Code', 'AI assistant with real-time codebase indexing. Context Engine MCP. Enterprise.'),
    ('Amazon Q Developer', 'AWS AI coding companion. Code suggestions, security, CLI agent.'),
    ('Cody (Sourcegraph)', 'AI coding with code graph context. Multi-model, enterprise code search.'),
    ('Supermaven', 'Ultra-fast AI completion. 1M token context, 300ms latency.'),
    ('Devin (Cognition)', 'Autonomous AI software engineer. Full-stack, long-running tasks.'),
    ('Lovable', 'AI app builder. Natural language to full-stack web apps. Vibe coding.'),
    ('Blackbox AI', 'AI code generation. Code from videos, autocomplete, chat.'),
]
for i, (name, desc) in enumerate(commercial, 1):
    print(f'  {i:2d}. {name}')
    print(f'     {desc}')
    print()

# MISSING
print()
print('  MISSING FROM ATLAS (added to incoming_resources.txt)')
print('  ' + '-' * 72)
print()
missing = [
    'paul-gauthier/aider - The original AI pair programming CLI',
    'danielmiessler/fabric - AI pattern framework',
    'TheR1D/shell-gpt - Lightweight GPT CLI',
    'saoudrizwan/claude-dev (Cline) - VS Code AI agent',
    'openai/symphony - OpenAI agent pipeline',
    'aws/amazon-q-developer-cli - AWS CLI agent',
    'AbanteAI/mentat - Autonomous coding agent',
    'AutoCodeRover - Program structure-guided bug fixing',
]
for m in missing:
    print(f'  - {m}')

print()
print('=' * 80)
print(f'  TOTALS: 7 Tier 1 | 10 Tier 2 | 39 Tier 3 ecosystem | 12 commercial | 8 pending ingestion')
print(f'  Atlas: {total:,} entries ({enriched:,} enriched)')
print('=' * 80)

atl.close()
