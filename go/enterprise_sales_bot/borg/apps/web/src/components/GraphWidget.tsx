"use client";

import React, { useMemo } from 'react';
import { trpc } from '@/utils/trpc';
import { KnowledgeGraph } from '@tormentnexus/ui';
import { motion } from 'framer-motion';

export function GraphWidget() {
    const { data, isLoading, refetch } = trpc.graph.get.useQuery(undefined, {
        refetchInterval: 10000,
        refetchOnWindowFocus: false
    });

    const mappedData = useMemo(() => {
        if (!data) return { nodes: [], links: [] };

        return {
            nodes: data.nodes.map((n: any) => ({
                id: n.id,
                label: n.name || n.id.split('/').pop() || n.id,
                type: (n.group === 1 ? 'topic' : (n.group === 2 ? 'document' : 'concept')) as 'topic' | 'document' | 'concept',
                val: n.val || 5
            })),
            links: data.links.map((l: any) => ({
                source: l.source,
                target: l.target,
                value: l.value || 1
            }))
        };
    }, [data]);

    const openFile = async (rawPath: string) => {
        if (!rawPath) {
            return;
        }

        const normalized = rawPath.startsWith('file://')
            ? decodeURIComponent(rawPath.replace('file://', ''))
            : rawPath;

        const vscodeUrl = `vscode://file/${encodeURIComponent(normalized)}`;
        window.open(vscodeUrl, '_blank');

        try {
            await navigator.clipboard.writeText(normalized);
        } catch {
            // Ignore clipboard failures in restricted browsers.
        }
    };

    const nodeCount = mappedData.nodes.length;
    const linkCount = mappedData.links.length;

    return (
        <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-zinc-900 to-black border border-zinc-800 h-full w-full min-h-[300px] flex flex-col">
            {/* Background Glow */}
            <div className="absolute inset-0 opacity-10 bg-gradient-to-br from-indigo-500 to-violet-500 blur-3xl" />

            <div className="relative z-10 flex flex-col h-full">
                {/* Header */}
                <div className="flex items-center justify-between p-3 border-b border-zinc-800">
                    <div className="flex items-center gap-2">
                        <span className="text-xl">🕸️</span>
                        <span className="text-sm font-bold text-white">Knowledge Graph</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="text-[10px] px-2 py-1 bg-indigo-500/20 text-indigo-400 rounded-full font-mono">
                            {nodeCount} nodes
                        </span>
                        <span className="text-[10px] px-2 py-1 bg-violet-500/20 text-violet-400 rounded-full font-mono">
                            {linkCount} links
                        </span>
                        <button
                            onClick={() => refetch()}
                            className="text-xs px-2 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white rounded transition-all"
                        >
                            ↻
                        </button>
                    </div>
                </div>

                {/* Graph Container */}
                <div className="flex-1 relative min-h-0">
                    {isLoading && nodeCount === 0 ? (
                        <div className="absolute inset-0 flex items-center justify-center">
                            <motion.div
                                animate={{ rotate: 360 }}
                                transition={{ repeat: Infinity, duration: 2, ease: "linear" }}
                                className="w-12 h-12 border-4 border-indigo-500 border-t-transparent rounded-full"
                            />
                        </div>
                    ) : nodeCount === 0 ? (
                        <div className="absolute inset-0 flex flex-col items-center justify-center text-zinc-600">
                            <span className="text-5xl mb-3">🌐</span>
                            <p className="text-sm">No graph data</p>
                            <p className="text-[10px]">Run indexer to populate</p>
                        </div>
                    ) : (
                        <KnowledgeGraph
                            nodes={mappedData.nodes}
                            links={mappedData.links}
                            loading={isLoading}
                            onNodeClick={(node) => {
                                if (node.type === 'document' || (node.id && node.id.includes('.'))) {
                                    openFile(node.id);
                                }
                            }}
                        />
                    )}
                </div>

                {/* Footer */}
                <div className="px-3 py-2 border-t border-zinc-800 flex justify-between items-center">
                    <span className="text-[10px] text-zinc-600">Click nodes to open in VS Code</span>
                    <span className="text-[10px] text-zinc-600">Refresh: 10s</span>
                </div>
            </div>
        </div>
    );
}
