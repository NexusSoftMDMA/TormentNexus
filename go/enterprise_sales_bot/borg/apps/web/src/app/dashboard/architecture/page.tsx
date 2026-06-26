"use client";

import React from 'react';
import { trpc } from '@/utils/trpc';

export default function ArchitecturePage() {
    const { data: submodules, isLoading } = trpc.git.getModules.useQuery();

    return (
        <div className="p-6">
            <h1 className="text-2xl font-bold mb-6">Project Architecture & Submodules</h1>

            <div className="space-y-6">
                {/* Project Structure */}
                <div className="p-6 border rounded-lg bg-zinc-900 border-zinc-700">
                    <h2 className="text-xl font-semibold mb-4 text-emerald-400">Directory Structure</h2>
                    <pre className="text-sm font-mono text-zinc-300 overflow-x-auto whitespace-pre">
                        {`
/apps
  /web              # Next.js Dashboard (Frontend)
  /extension        # Chrome Extension (Browser Client)
  
/packages
  /core             # Core Agentic Framework (MCPServer, Director, Memory)
  /cli              # CLI functionality (tormentnexus start, tormentnexus doctor)
  /vscode           # VS Code Extension (Editor Interface)
  
/dockers
  /tormentnexus-server      # Containerized Deployment
  
/docs               # Documentation & Research
`}
                    </pre>
                </div>

                {/* Submodules List from Git */}
                <div className="p-6 border rounded-lg bg-zinc-900 border-zinc-700">
                    <h2 className="text-xl font-semibold mb-4 text-blue-400">Integrated Submodules (Live)</h2>
                    <div className="overflow-x-auto">
                        <table className="min-w-full text-left text-sm">
                            <thead className="bg-zinc-800 text-zinc-400">
                                <tr>
                                    <th className="px-4 py-2">Name</th>
                                    <th className="px-4 py-2">Path</th>
                                    <th className="px-4 py-2">URL</th>
                                    <th className="px-4 py-2">Status</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-zinc-700 text-zinc-300">
                                {isLoading ? (
                                    <tr><td colSpan={4} className="px-4 py-4 text-center">Loading git modules...</td></tr>
                                ) : submodules?.map((mod: any) => (
                                    <tr key={mod.path}>
                                        <td className="px-4 py-2 font-medium">{mod.name}</td>
                                        <td className="px-4 py-2">{mod.path}</td>
                                        <td className="px-4 py-2 text-zinc-500 truncate max-w-xs">{mod.url}</td>
                                        <td className="px-4 py-2 text-emerald-500">Active</td>
                                    </tr>
                                ))}
                                {submodules && submodules.length === 0 && (
                                    <tr><td colSpan={4} className="px-4 py-4 text-center text-zinc-500">No submodules found (.gitmodules empty)</td></tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
                {/* Dependency Graph */}
                <div className="p-6 border rounded-lg bg-zinc-900 border-zinc-700">
                    <h2 className="text-xl font-semibold mb-4 text-purple-400">Deep Code Intelligence Graph</h2>

                    <GraphVisualizer />
                </div>

                {/* Auto-Test Health */}
                <div className="p-6 border rounded-lg bg-zinc-900 border-zinc-700">
                    <h2 className="text-xl font-semibold mb-4 text-rose-400">Auto-Test Health</h2>
                    <AutoTestWidget />
                </div>
            </div>
        </div>
    );
}

import Mermaid from '@/components/Mermaid';

function GraphVisualizer() {
    const { data: graph, isLoading } = trpc.graph.get.useQuery();
    const [mermaidSrc, setMermaidSrc] = React.useState('');

    React.useEffect(() => {
        if (!graph) return;

        let src = 'graph TD\n';
        // Limit nodes for performance/visual clarity
        const nodes = Object.keys(graph.dependencies).slice(0, 50); // Top 50 modules

        nodes.forEach(node => {
            const cleanNode = node.replace(/[^a-zA-Z0-9]/g, '_');
            src += `    ${cleanNode}["${node.split('/').pop()}"]\n`;

            const deps = graph.dependencies[node] || [];
            deps.forEach((dep: string) => {
                if (nodes.includes(dep)) {
                    const cleanDep = dep.replace(/[^a-zA-Z0-9]/g, '_');
                    src += `    ${cleanNode} --> ${cleanDep}\n`;
                }
            });
        });

        if (nodes.length === 0) src += '    Start --> End\n';

        setMermaidSrc(src);
    }, [graph]);

    if (isLoading) return <div className="text-zinc-500">Loading Intelligence Graph...</div>;

    return (
        <div className="w-full">
            <Mermaid chart={mermaidSrc} />
        </div>
    );
}

function AutoTestWidget() {
    return (
        <div className="space-y-2">
            <div className="text-zinc-500 italic text-sm">
                Auto-test service is currently disabled. Enable the <code className="bg-zinc-800 px-1 rounded">autoTest</code> router in <code className="bg-zinc-800 px-1 rounded">trpc.ts</code> to activate file-save-triggered testing.
            </div>
            <div className="grid grid-cols-3 gap-2 text-xs">
                <div className="bg-zinc-800 p-3 rounded text-center">
                    <div className="text-2xl font-bold text-zinc-400">—</div>
                    <div className="text-zinc-500 mt-1">Tests Run</div>
                </div>
                <div className="bg-zinc-800 p-3 rounded text-center">
                    <div className="text-2xl font-bold text-emerald-400">—</div>
                    <div className="text-zinc-500 mt-1">Passing</div>
                </div>
                <div className="bg-zinc-800 p-3 rounded text-center">
                    <div className="text-2xl font-bold text-rose-400">—</div>
                    <div className="text-zinc-500 mt-1">Failing</div>
                </div>
            </div>
        </div>
    );
}
