"use client";
import React from 'react';
import Link from 'next/link';

export default function GuidePage() {
    return (
        <div className="min-h-screen bg-zinc-50 dark:bg-black">
            <header className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 px-6 py-4">
                <div className="max-w-4xl mx-auto flex items-center justify-between">
                    <h1 className="text-2xl font-bold text-zinc-900 dark:text-white">📚 TormentNexus User Guide</h1>
                    <Link href="/" className="text-sm text-blue-500 hover:text-blue-400">← Dashboard</Link>
                </div>
            </header>

            <main className="max-w-4xl mx-auto px-6 py-8 prose prose-invert max-w-none">

                {/* Quick Start */}
                <section className="mb-12 p-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                    <h2 className="text-xl font-bold mb-4">🚀 Quick Start</h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="p-4 bg-zinc-50 dark:bg-zinc-800 rounded-lg">
                            <h3 className="font-bold text-sm mb-2">Access Keys</h3>
                            <ul className="text-sm text-zinc-600 dark:text-zinc-400 space-y-1">
                                <li><code className="bg-zinc-200 dark:bg-zinc-700 px-1 rounded">admin</code> - Full access</li>
                                <li><code className="bg-zinc-200 dark:bg-zinc-700 px-1 rounded">tormentnexus</code> - Full access</li>
                                <li><span className="text-zinc-500">Press Enter</span> - Dev access</li>
                            </ul>
                        </div>
                        <div className="p-4 bg-zinc-50 dark:bg-zinc-800 rounded-lg">
                            <h3 className="font-bold text-sm mb-2">Keyboard Shortcuts</h3>
                            <ul className="text-sm text-zinc-600 dark:text-zinc-400 space-y-1">
                                <li><kbd className="bg-zinc-200 dark:bg-zinc-700 px-1 rounded">Ctrl+K</kbd> - Global search</li>
                                <li><kbd className="bg-zinc-200 dark:bg-zinc-700 px-1 rounded">Drag</kbd> - Rearrange widgets</li>
                                <li><kbd className="bg-zinc-200 dark:bg-zinc-700 px-1 rounded">❓</kbd> - Widget help</li>
                            </ul>
                        </div>
                    </div>
                </section>

                {/* AI Agents */}
                <section className="mb-12">
                    <h2 className="text-xl font-bold mb-4">🤖 AI Agents</h2>

                    <div className="space-y-4">
                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold">Director Chat</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-2">Your autonomous AI assistant that writes code, runs commands, and manages tasks.</p>
                            <div className="text-xs text-zinc-500">
                                <strong>How:</strong> Type in chat widget → Enter → Watch execution
                            </div>
                        </div>

                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold">Council Debate</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-2">Three AI personas (Architect, Product, Critic) debate decisions.</p>
                            <div className="text-xs text-zinc-500">
                                <strong>How:</strong> Enter topic → Start Debate → Watch → Review consensus
                            </div>
                        </div>

                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold">Squad Workers</h3>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-2">Parallel agents in isolated git worktrees.</p>
                            <div className="text-xs text-zinc-500">
                                <strong>How:</strong> Spawn Worker → Assign task → Merge when done
                            </div>
                        </div>
                    </div>
                </section>

                {/* Development Tools */}
                <section className="mb-12">
                    <h2 className="text-xl font-bold mb-4">🛠️ Development Tools</h2>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold text-sm">Code Sandbox</h3>
                            <p className="text-xs text-zinc-500">Execute Python/Node in Docker</p>
                        </div>
                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold text-sm">Auto-Test Watcher</h3>
                            <p className="text-xs text-zinc-500">Tests run on file save</p>
                        </div>
                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold text-sm">Knowledge Graph</h3>
                            <p className="text-xs text-zinc-500">Click nodes → opens in VS Code</p>
                        </div>
                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold text-sm">Global Search</h3>
                            <p className="text-xs text-zinc-500">Ctrl+K → semantic search</p>
                        </div>
                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold text-sm">Shell History</h3>
                            <p className="text-xs text-zinc-500">Browse command history</p>
                        </div>
                        <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                            <h3 className="font-bold text-sm">Command Runner</h3>
                            <p className="text-xs text-zinc-500">Execute commands directly</p>
                        </div>
                    </div>
                </section>

                {/* Slash Commands */}
                <section className="mb-12">
                    <h2 className="text-xl font-bold mb-4">⌨️ Slash Commands</h2>
                    <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                        <div className="grid grid-cols-2 gap-2 text-sm font-mono">
                            <div><code>/help</code> <span className="text-zinc-500">List commands</span></div>
                            <div><code>/director status</code> <span className="text-zinc-500">Agent info</span></div>
                            <div><code>/git status</code> <span className="text-zinc-500">Repo status</span></div>
                            <div><code>/context add [file]</code> <span className="text-zinc-500">Pin file</span></div>
                            <div><code>/council debate [topic]</code> <span className="text-zinc-500">Start debate</span></div>
                            <div><code>/squad spawn [task]</code> <span className="text-zinc-500">Create worker</span></div>
                            <div><code>/test run</code> <span className="text-zinc-500">Run tests</span></div>
                            <div><code>/heal</code> <span className="text-zinc-500">Auto-fix errors</span></div>
                        </div>
                    </div>
                </section>

                {/* Security */}
                <section className="mb-12">
                    <h2 className="text-xl font-bold mb-4">🔐 Security</h2>
                    <div className="p-4 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                        <h3 className="font-bold text-sm mb-2">Autonomy Levels</h3>
                        <table className="w-full text-sm">
                            <tbody>
                                <tr className="border-b border-zinc-200 dark:border-zinc-800">
                                    <td className="py-2 font-bold text-yellow-500">LOW</td>
                                    <td className="py-2 text-zinc-500">All actions require approval</td>
                                </tr>
                                <tr className="border-b border-zinc-200 dark:border-zinc-800">
                                    <td className="py-2 font-bold text-blue-500">MEDIUM</td>
                                    <td className="py-2 text-zinc-500">Safe actions auto-approved</td>
                                </tr>
                                <tr>
                                    <td className="py-2 font-bold text-red-500">HIGH</td>
                                    <td className="py-2 text-zinc-500">Most actions auto-approved</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </section>

                {/* Footer */}
                <footer className="mt-12 pt-8 border-t border-zinc-200 dark:border-zinc-800 text-center text-sm text-zinc-500">
                    <p>TormentNexus Mission Control • Complete Documentation</p>
                    <div className="mt-4 flex justify-center gap-4">
                        <Link href="/" className="text-blue-500 hover:text-blue-400">Dashboard</Link>
                        <Link href="/docs" className="text-blue-500 hover:text-blue-400">Feature Docs</Link>
                    </div>
                </footer>
            </main>
        </div>
    );
}
