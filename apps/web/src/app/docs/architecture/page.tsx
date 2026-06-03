"use client";
import React from 'react';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import Mermaid from '@/components/Mermaid';

export default function ArchitectureDocsPage() {
    return (
        <div className="min-h-screen bg-zinc-50 dark:bg-black">
            <header className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 px-6 py-4 sticky top-0 z-20">
                <div className="max-w-5xl mx-auto flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <Link href="/docs" className="text-zinc-500 hover:text-blue-500 transition-colors">
                            <ArrowLeft size={20} />
                        </Link>
                        <h1 className="text-xl font-bold text-zinc-900 dark:text-white">System Architecture</h1>
                    </div>
                </div>
            </header>

            <main className="max-w-5xl mx-auto px-6 py-8 prose prose-zinc dark:prose-invert max-w-none">

                <section className="mb-16">
                    <h2>System Overview</h2>
                    <p>
                        TormentNexus is architected as a modular monorepo using Turborepo. It separates concerns between the
                        Core MCP Server (Backend), the Next.js Dashboard (Frontend), and the Browser Extension (Bridge).
                    </p>

                    <div className="my-8 p-4 bg-white dark:bg-zinc-900 rounded border border-zinc-200 dark:border-zinc-800">
                        <Mermaid chart={`
graph TB
    CLI[apps/cli] --> Core[packages/core]
    Web[apps/web] --> Core
    Web --> UI[packages/ui]
    Core --> AI[packages/ai]
    Extension[apps/extension] --> Core
    
    subgraph External
        OpenAI[OpenAI API]
        Anthropic[Anthropic API]
        Docker[Docker]
    end
    
    AI --> OpenAI
    AI --> Anthropic
    Core --> Docker
                        `} />
                        <p className="text-center text-xs text-zinc-500 mt-2">Package Dependency Graph</p>
                    </div>
                </section>

                <section className="mb-16">
                    <h2>Data Flow & Communication</h2>
                    <p>
                        The system uses a mix of HTTP, WebSocket, and tRPC for robust real-time communication.
                    </p>

                    <div className="grid md:grid-cols-2 gap-8">
                        <div>
                            <h3>Message Processing</h3>
                            <div className="p-4 bg-white dark:bg-zinc-900 rounded border border-zinc-200 dark:border-zinc-800">
                                <Mermaid chart={`
graph TD
    User[User Input] --> Chat[Director Chat]
    Chat -- tRPC --> Router
    Router --> Director[Director Agent]
    Director -- Plan --> Policy[Policy Check]
    Policy -- execute --> Tool[Tool Exec]
    Tool --> Audit[Audit Log]
    Audit --> Response
                                `} />
                            </div>
                        </div>

                        <div>
                            <h3>Indexing Pipeline</h3>
                            <div className="p-4 bg-white dark:bg-zinc-900 rounded border border-zinc-200 dark:border-zinc-800">
                                <Mermaid chart={`
graph TD
    File[Source Files] --> Watcher[File Watcher]
    Watcher --> Chunker
    Chunker --> Embedder[OpenAI Embeddings]
    Embedder --> VectorDB[Vector Store]
                                `} />
                            </div>
                        </div>
                    </div>
                </section>

                <section className="mb-16">
                    <h2>Deployment Architecture</h2>
                    <div className="p-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800">
                        <h4>Local Development Stack</h4>
                        <ul className="list-disc pl-5 text-sm space-y-2 mt-4">
                            <li><strong>CLI (Entry)</strong>: Launches Core server and orchestrates processes.</li>
                            <li><strong>Core (Port 3001)</strong>: MCP Server handling agents, tools, and DBs.</li>
                            <li><strong>Web (Port 3000)</strong>: Next.js app for visualization and control.</li>
                            <li><strong>Runtime Data</strong>: Stored in <code>.tormentnexus/</code> (indexes, logs, config).</li>
                            <li><strong>Sandboxes</strong>: Docker containers for safe code execution.</li>
                        </ul>
                    </div>
                </section>

                <footer className="mt-12 pt-8 border-t border-zinc-200 dark:border-zinc-800 text-center text-sm text-zinc-500">
                    <p>Generated from system architecture definitions.</p>
                </footer>
            </main>
        </div>
    );
}
