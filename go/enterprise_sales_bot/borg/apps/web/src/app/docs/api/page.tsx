"use client";
import React from 'react';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';

export default function ApiDocsPage() {
    return (
        <div className="min-h-screen bg-zinc-50 dark:bg-black">
            <header className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 px-6 py-4 sticky top-0 z-20">
                <div className="max-w-5xl mx-auto flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <Link href="/docs" className="text-zinc-500 hover:text-blue-500 transition-colors">
                            <ArrowLeft size={20} />
                        </Link>
                        <h1 className="text-xl font-bold text-zinc-900 dark:text-white">Technical Reference & API</h1>
                    </div>
                    <div className="text-xs font-mono text-zinc-500">v1.0.0</div>
                </div>
            </header>

            <main className="max-w-5xl mx-auto px-6 py-8 prose prose-zinc dark:prose-invert max-w-none">
                <nav className="mb-12 p-4 bg-zinc-100 dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                    <h3 className="text-sm font-bold uppercase text-zinc-500 mb-3">Table of Contents</h3>
                    <ul className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm not-prose">
                        <li><a href="#mcp-tools" className="text-blue-500 hover:underline">MCP Tools</a></li>
                        <li><a href="#trpc-endpoints" className="text-blue-500 hover:underline">tRPC Endpoints</a></li>
                        <li><a href="#agent-config" className="text-blue-500 hover:underline">Agent Config</a></li>
                        <li><a href="#slash-commands" className="text-blue-500 hover:underline">Slash Commands</a></li>
                    </ul>
                </nav>

                <section id="mcp-tools" className="mb-16">
                    <h2>MCP Tools API</h2>
                    <p className="lead">The Model Context Protocol (MCP) exposes these primitive tools to AI agents.</p>

                    <div className="grid gap-6">
                        <div className="p-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="mt-0 font-mono text-blue-500">read_file</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">Read contents of a file relative to workspace root.</p>
                            <pre className="bg-zinc-100 dark:bg-black p-4 rounded text-xs font-mono">
                                {`{
  "path": "/absolute/path/to/file",
  "startLine": 1, // optional
  "endLine": 100 // optional
}`}
                            </pre>
                        </div>

                        <div className="p-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="mt-0 font-mono text-blue-500">write_file</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">Write content to a file (creates if not exists).</p>
                            <pre className="bg-zinc-100 dark:bg-black p-4 rounded text-xs font-mono">
                                {`{
  "path": "/absolute/path/to/file",
  "content": "file contents here"
}`}
                            </pre>
                        </div>

                        <div className="p-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="mt-0 font-mono text-blue-500">search_codebase</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">Semantic search using vector embeddings.</p>
                            <pre className="bg-zinc-100 dark:bg-black p-4 rounded text-xs font-mono">
                                {`{
  "query": "authentication logic",
  "limit": 10
}`}
                            </pre>
                        </div>

                        <div className="p-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="mt-0 font-mono text-blue-500">run_command</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">Execute a shell command via child_process.</p>
                            <pre className="bg-zinc-100 dark:bg-black p-4 rounded text-xs font-mono">
                                {`{
  "command": "pnpm test",
  "cwd": "/working/directory",
  "timeout": 30000
}`}
                            </pre>
                        </div>
                    </div>
                </section>

                <section id="trpc-endpoints" className="mb-16">
                    <h2>tRPC Endpoints</h2>
                    <p>Backend API routes accessible via <code>trpc/server</code>.</p>

                    <div className="overflow-x-auto">
                        <table className="w-full text-sm text-left">
                            <thead className="text-xs uppercase bg-zinc-100 dark:bg-zinc-800">
                                <tr>
                                    <th className="px-6 py-3">Router</th>
                                    <th className="px-6 py-3">Procedure</th>
                                    <th className="px-6 py-3">Type</th>
                                    <th className="px-6 py-3">Description</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-zinc-200 dark:divide-zinc-800">
                                <tr>
                                    <td className="px-6 py-4 font-mono">director</td>
                                    <td className="px-6 py-4 font-mono">sendMessage</td>
                                    <td className="px-6 py-4"><span className="px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-600 text-xs rounded">Mutation</span></td>
                                    <td className="px-6 py-4">Send message to AI agent</td>
                                </tr>
                                <tr>
                                    <td className="px-6 py-4 font-mono">health</td>
                                    <td className="px-6 py-4 font-mono">check</td>
                                    <td className="px-6 py-4"><span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-600 text-xs rounded">Query</span></td>
                                    <td className="px-6 py-4">Service status ping</td>
                                </tr>
                                <tr>
                                    <td className="px-6 py-4 font-mono">indexing</td>
                                    <td className="px-6 py-4 font-mono">status</td>
                                    <td className="px-6 py-4"><span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-600 text-xs rounded">Query</span></td>
                                    <td className="px-6 py-4">Get vector indexing progress</td>
                                </tr>
                                <tr>
                                    <td className="px-6 py-4 font-mono">council</td>
                                    <td className="px-6 py-4 font-mono">startDebate</td>
                                    <td className="px-6 py-4"><span className="px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-600 text-xs rounded">Mutation</span></td>
                                    <td className="px-6 py-4">Trigger multi-agent debate</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </section>

                <section id="slash-commands" className="mb-16">
                    <h2>Slash Commands</h2>
                    <p>Available commands in the Director Chat interface.</p>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {[
                            { cmd: '/help', desc: 'List all commands' },
                            { cmd: '/director status', desc: 'Show agent internal state' },
                            { cmd: '/director pause', desc: 'Suspend autonomous loop' },
                            { cmd: '/git status', desc: 'Show repository status' },
                            { cmd: '/context add [file]', desc: 'Pin file to context window' },
                            { cmd: '/council debate [topic]', desc: 'Start generic debate' },
                            { cmd: '/squad spawn [task]', desc: 'Create worker in new worktree' },
                            { cmd: '/test run', desc: 'Execute test suite' },
                            { cmd: '/search [query]', desc: 'Direct semantic search' },
                            { cmd: '/metrics', desc: 'Show system resource stats' }
                        ].map(cmd => (
                            <div key={cmd.cmd} className="flex items-center justify-between p-3 bg-zinc-100 dark:bg-zinc-800 rounded border border-zinc-200 dark:border-zinc-700">
                                <code className="text-sm font-bold text-blue-500">{cmd.cmd}</code>
                                <span className="text-xs text-zinc-500">{cmd.desc}</span>
                            </div>
                        ))}
                    </div>
                </section>
            </main>
        </div>
    );
}
