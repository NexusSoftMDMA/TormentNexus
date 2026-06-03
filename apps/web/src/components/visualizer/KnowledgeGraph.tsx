
"use client";

import dynamic from 'next/dynamic';
import React, { useRef } from 'react';
import { trpc } from '@/utils/trpc'; // Assuming standard TRPC hook location or adjust
import { Card, CardHeader, CardTitle, CardContent, useResizeObserver } from '@tormentnexus/ui';

// ForceGraph must be dynamically imported as it uses window
const ForceGraph2D = dynamic(() => import('react-force-graph-2d'), { ssr: false });

export function KnowledgeGraph() {
    // For now, we might fetch via a new TRPC procedure or a direct MCP call proxy.
    // Since Phase 29 didn't explicitly build a TRPC router yet, I'll mock the data fetch via "mcp_call" proxy or similar if available,
    // OR we can assume a `knowledge.getGraph` procedure exists.
    // Given previous patterns, I should probably add a TRPC router for this.
    // But for "Visual Grounding", let's visualize the "Knowledge Service" directly.

    // Use TRPC to fetch graph data
    const { data } = trpc.knowledge.getGraph.useQuery({ query: undefined, depth: 2 }, {
        refetchOnWindowFocus: false
    });

    const graphData = React.useMemo(() => {
        if (!data) return { nodes: [], links: [] };
        return {
            nodes: data.nodes || [],
            links: (data.edges || []).map((e: { source: string; target: string; value?: number }) => ({
                source: e.source,
                target: e.target,
                value: e.value || 1
            }))
        };
    }, [data]);

    const fgRef = useRef<any>();
    const containerRef = useRef<HTMLDivElement>(null);
    const { width, height } = useResizeObserver(containerRef);

    // Auto-resize / re-center when data or container size changes
    React.useEffect(() => {
        if (fgRef.current && graphData.nodes.length > 0) {
            fgRef.current.zoomToFit(400);
        }
    }, [graphData, width, height]);

    return (
        <Card className="h-[80vh] w-full flex flex-col">
            <CardHeader>
                <CardTitle>Memory Layout (Hippocampus)</CardTitle>
            </CardHeader>
            <CardContent ref={containerRef} className="flex-1 p-0 relative bg-zinc-950 overflow-hidden rounded-b-xl">
                <ForceGraph2D
                    ref={fgRef}
                    width={width}
                    height={height}
                    graphData={graphData}
                    nodeLabel="id"
                    nodeColor={node => (node as any).group === 1 ? '#ef4444' : '#3b82f6'}
                    linkColor={() => '#ffffff40'}
                    backgroundColor="#09090b" // Zinc-950
                />
            </CardContent>
        </Card>
    );
}
