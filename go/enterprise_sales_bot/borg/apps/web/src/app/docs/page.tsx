"use client";
import React from 'react';
import Link from 'next/link';

const SECTIONS = [
    {
        id: 'agents',
        title: '🤖 AI Agents',
        features: [
            {
                name: 'Director',
                widget: 'director_chat',
                description: 'The Director is the autonomous AI agent that manages your development workflow. It can execute tasks, write code, and coordinate with other agents.',
                usage: ['Type messages in the chat to interact', 'Use /director status to check agent state', 'Enable Auto-Drive for fully autonomous operation'],
                related: ['Autonomy Controls', 'Council'],
            },
            {
                name: 'Council',
                widget: 'council',
                description: 'Multi-agent debate system where three AI personas (Architect, Product Manager, Critic) debate decisions to reach consensus.',
                usage: ['Start a debate by entering a topic', 'Watch the debate unfold in real-time', 'Review the final consensus'],
                related: ['Director', 'Config'],
            },
            {
                name: 'Squad',
                widget: 'squad',
                description: 'Parallel agent workers that operate in isolated git worktrees. Spawn workers for concurrent task execution.',
                usage: ['Click "Spawn" to create a new worker', 'Each worker operates in its own branch', 'Merge results back to main when complete'],
                related: ['Director', 'Git'],
            },
        ],
    },
    {
        id: 'monitoring',
        title: '📊 Monitoring',
        features: [
            {
                name: 'System Health',
                widget: 'system_health',
                description: 'Real-time monitoring of CPU, memory, and system resources.',
                usage: ['View current resource utilization', 'Watch for performance issues'],
                related: ['Activity Pulse', 'Latency'],
            },
            {
                name: 'Audit Logs',
                widget: 'audit',
                description: 'Complete audit trail of all agent actions for compliance and debugging.',
                usage: ['Review all tool executions', 'Filter by action type', 'Track who did what and when'],
                related: ['Security', 'Director'],
            },
            {
                name: 'Self-Healing',
                widget: 'healer',
                description: 'Automatic error detection and repair. The Healer agent diagnoses issues and applies fixes.',
                usage: ['Watch for error events', 'Review applied fixes', 'Check success/failure status'],
                related: ['Tests', 'Director'],
            },
        ],
    },
    {
        id: 'development',
        title: '🛠️ Development',
        features: [
            {
                name: 'Code Sandbox',
                widget: 'sandbox',
                description: 'Execute Python or Node.js code safely in Docker containers.',
                usage: ['Select language', 'Enter code in the editor', 'Click Run to execute', 'View output below'],
                related: ['Tests', 'Shell'],
            },
            {
                name: 'Auto-Test Watcher',
                widget: 'tests',
                description: 'Automatically runs tests when files change. Shows pass/fail status in real-time.',
                usage: ['Click Start to begin watching', 'Tests run on file save', 'View results and error output'],
                related: ['Healer', 'Sandbox'],
            },
            {
                name: 'Knowledge Graph',
                widget: 'graph_1',
                description: 'Interactive visualization of codebase structure. Shows files, dependencies, and relationships.',
                usage: ['Drag to pan, scroll to zoom', 'Click a node to open file in VS Code', 'Hover for file details'],
                related: ['Indexing', 'Search'],
            },
        ],
    },
    {
        id: 'security',
        title: '🔐 Security',
        features: [
            {
                name: 'Autonomy Control',
                widget: 'autonomy',
                description: 'Configure agent permission levels. Controls what actions agents can take without approval.',
                usage: ['Low: All actions require approval', 'Medium: Safe actions auto-approved', 'High: Most actions auto-approved'],
                related: ['Security Shield', 'Director'],
            },
            {
                name: 'Security Shield',
                widget: 'security',
                description: 'Policy-based security controls. Restrict file access, rate limits, and dangerous operations.',
                usage: ['View active policies', 'Toggle lockdown mode', 'Review blocked actions'],
                related: ['Autonomy', 'Audit'],
            },
        ],
    },
    {
        id: 'productivity',
        title: '⚡ Productivity',
        features: [
            {
                name: 'Global Search',
                widget: 'search',
                description: 'Semantic search across your entire codebase. Find code, documentation, and files.',
                usage: ['Use the search bar in the header', 'Click results to open in VS Code', 'Supports natural language queries'],
                related: ['Graph', 'Context'],
            },
            {
                name: 'Context Management',
                widget: 'context',
                description: 'Manage pinned files and active context for AI conversations.',
                usage: ['Pin important files', 'Add/remove from context', 'Clear context to start fresh'],
                related: ['Director', 'Search'],
            },
            {
                name: 'Suggestions',
                widget: 'suggestions',
                description: 'AI-generated proactive recommendations based on your current context.',
                usage: ['Review AI suggestions', 'Click Approve to execute', 'Dismiss irrelevant suggestions'],
                related: ['Director', 'Context'],
            },
        ],
    },
];

export default function DocsPage() {
    return (
        <div className="min-h-screen bg-zinc-50 dark:bg-black">
            <header className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 px-6 py-4">
                <div className="max-w-4xl mx-auto flex items-center justify-between">
                    <h1 className="text-2xl font-bold text-zinc-900 dark:text-white">📖 Feature Documentation</h1>
                    <div className="flex items-center gap-4 text-sm">
                        <Link href="/docs/tools" className="text-zinc-500 hover:text-blue-500">Tools Reference</Link>
                        <Link href="/docs/api" className="text-zinc-500 hover:text-blue-500">API Reference</Link>
                        <Link href="/docs/architecture" className="text-zinc-500 hover:text-blue-500">Architecture</Link>
                        <Link href="/" className="text-blue-500 hover:text-blue-400 font-medium">← Back to Dashboard</Link>
                    </div>
                </div>
            </header>

            <main className="max-w-4xl mx-auto px-6 py-8">
                {/* Quick Nav */}
                <nav className="mb-8 p-4 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg">
                    <h2 className="text-sm font-bold text-zinc-500 uppercase mb-3">Quick Navigation</h2>
                    <div className="flex flex-wrap gap-2">
                        {SECTIONS.map((section) => (
                            <a
                                key={section.id}
                                href={`#${section.id}`}
                                className="px-3 py-1 bg-zinc-100 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300 rounded-full text-sm hover:bg-blue-100 dark:hover:bg-blue-900 hover:text-blue-600 transition-colors"
                            >
                                {section.title}
                            </a>
                        ))}
                    </div>
                </nav>

                {/* Sections */}
                {SECTIONS.map((section) => (
                    <section key={section.id} id={section.id} className="mb-12">
                        <h2 className="text-xl font-bold text-zinc-900 dark:text-white mb-4">{section.title}</h2>
                        <div className="space-y-4">
                            {section.features.map((feature) => (
                                <article
                                    key={feature.name}
                                    id={feature.widget}
                                    className="p-5 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg"
                                >
                                    <h3 className="text-lg font-bold text-zinc-900 dark:text-white mb-2">{feature.name}</h3>
                                    <p className="text-zinc-600 dark:text-zinc-400 mb-4">{feature.description}</p>

                                    <div className="mb-4">
                                        <h4 className="text-sm font-bold text-zinc-500 uppercase mb-2">How to Use</h4>
                                        <ul className="list-disc list-inside text-sm text-zinc-600 dark:text-zinc-400 space-y-1">
                                            {feature.usage.map((step, i) => (
                                                <li key={i}>{step}</li>
                                            ))}
                                        </ul>
                                    </div>

                                    <div className="flex items-center gap-2 text-xs">
                                        <span className="text-zinc-500">Related:</span>
                                        {feature.related.map((rel) => (
                                            <span key={rel} className="px-2 py-0.5 bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 rounded">
                                                {rel}
                                            </span>
                                        ))}
                                    </div>
                                </article>
                            ))}
                        </div>
                    </section>
                ))}


                <section className="mb-16 pt-8 border-t border-zinc-200 dark:border-zinc-800">
                    <h2 className="text-xl font-bold text-zinc-900 dark:text-white mb-6">📚 Technical Deep Dives</h2>
                    <div className="grid md:grid-cols-3 gap-6">
                        <Link href="/docs/tools" className="block p-6 bg-zinc-50 dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 hover:border-green-500 transition-colors group">
                            <h3 className="text-lg font-bold text-zinc-900 dark:text-white mb-2 group-hover:text-green-500">🛠️ MCP Tools Reference</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">
                                Complete reference for 30 MCP tools including LSP, Plan/Build, Memory, Workflow, and Code Mode.
                            </p>
                        </Link>
                        <Link href="/docs/api" className="block p-6 bg-zinc-50 dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 hover:border-blue-500 transition-colors group">
                            <h3 className="text-lg font-bold text-zinc-900 dark:text-white mb-2 group-hover:text-blue-500">API & Technical Reference</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">
                                Complete specification of MCP Tools, tRPC endpoints, configuration schemas, and slash commands.
                            </p>
                        </Link>
                        <Link href="/docs/architecture" className="block p-6 bg-zinc-50 dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 hover:border-purple-500 transition-colors group">
                            <h3 className="text-lg font-bold text-zinc-900 dark:text-white mb-2 group-hover:text-purple-500">System Architecture</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">
                                Visual diagrams of system components, data flow, indexing pipelines, and deployment topology.
                            </p>
                        </Link>
                    </div>
                </section>

                {/* Footer */}
                <footer className="mt-12 pt-8 border-t border-zinc-200 dark:border-zinc-800 text-center text-sm text-zinc-500">
                    <p>TormentNexus Mission Control • All features documented</p>
                    <p className="mt-2">
                        <Link href="/" className="text-blue-500 hover:text-blue-400">Return to Dashboard</Link>
                    </p>
                </footer>
            </main>
        </div>
    );
}
