
'use client';

import React from 'react';
import { KnowledgeGraph } from '@tormentnexus/ui';
import { trpc } from '@/utils/trpc';

export default function BrainPage() {
    const graphQuery = trpc.graph.getSymbolsGraph.useQuery();
    const { nodes: rawNodes = [], links: rawLinks = [] } = graphQuery.data || {};

    const nodes = rawNodes.map((node: any) => ({
        id: node.id,
        label: node.name,
        type: 'concept' as const,
        val: node.val || 1
    }));

    const links = rawLinks.map((link: any) => ({
        source: link.source,
        target: link.target,
        value: 1
    }));

    return (
        <div className="h-full w-full p-6 flex flex-col">
            <header className="mb-6">
                <h1 className="text-3xl font-bold bg-gradient-to-r from-purple-500 to-pink-500 bg-clip-text text-transparent">
                    The TormentNexus Brain
                </h1>
                <p className="text-zinc-500 dark:text-zinc-400">
                    Visualizing the collective knowledge graph ({nodes.length} nodes).
                </p>
            </header>

            <div className="flex-1 min-h-0 bg-zinc-100 dark:bg-zinc-900 rounded-2xl overflow-hidden shadow-inner relative">
                <div className="absolute inset-0">
                    <KnowledgeGraph
                        nodes={nodes}
                        links={links}
                        loading={graphQuery.isLoading}
                    />
                </div>
            </div>
        </div>
    );
}
