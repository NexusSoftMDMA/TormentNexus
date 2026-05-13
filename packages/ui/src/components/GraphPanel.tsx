"use client";

import React, { useRef, useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import { trpc } from '../utils/trpc';
import { Card } from './ui/card';
import { Button } from './ui/button';
import { Loader2, RefreshCw, ZoomIn, ZoomOut } from 'lucide-react';
import { useResizeObserver } from '../hooks/use-resize-observer';

const ForceGraph2D = dynamic(() => import('react-force-graph-2d'), { ssr: false });

export function GraphPanel() {
    const { data: graphData, isLoading, refetch } = trpc.graph.get.useQuery(undefined, {
        refetchOnWindowFocus: false,
        staleTime: 60000 // Cache for 1 min
    });

    const [showSymbols, setShowSymbols] = useState(false);
    const { data: symbolData } = trpc.graph.getSymbolsGraph.useQuery(undefined, {
        enabled: showSymbols,
        staleTime: 300000 // Cache longer
    });

    const displayGraph = React.useMemo(() => {
        if (!graphData) return { nodes: [], links: [] };
        if (!showSymbols || !symbolData) return graphData;

        // Merge
        return {
            nodes: [...graphData.nodes, ...symbolData.nodes],
            links: [...graphData.links, ...symbolData.links]
        };
    }, [graphData, symbolData, showSymbols]);

    const executeTool = trpc.executeTool.useMutation();
    const fgRef = useRef<any>();
    const containerRef = useRef<HTMLDivElement>(null);
    const { width, height } = useResizeObserver(containerRef);

    // Auto-resize / re-center when data or container size changes
    useEffect(() => {
        if (fgRef.current && displayGraph.nodes.length > 0) {
            fgRef.current.zoomToFit(400);
        }
    }, [displayGraph, width, height]);

    const handleNodeClick = (node: any) => {
        if (node.id) {
            // Convert relative path to absolute for VS Code
            // We assume backend returns relative to root, but vscode needs absolute.
            // Actually, executeTool('vscode_open', ...) handles paths?
            // Let's use 'vscode_execute_command' with 'vscode.open'.
            // But we need absolute path. 'view_file' takes absolute.
            // Let's use `view_file` logic or similar.
            // Actually, we can just send the ID (path) via `vscode_execute_command` relative to workspace?
            // Usually `vscode.open` needs a Uri.
            // Easier: interactive.open?
            // Let's use `vscode_execute_command` with `vscode.open` and a constructed URI if possible, 
            // or use our `view_file` logic on backend.
            // Better: `executeTool('vscode_execute_command', { command: 'vscode.open', args: [URI] })`
            // But formatting URI is hard on frontend.
            // Let's call `executeTool('read_file', { path: ... })` which triggers suggestionService logic too.
            // OR best: `executeTool('native_input', ...)` is bad.
            // Let's try `view_file` which we know opens/reads.

            // Wait, standard way to open in editor:
            executeTool.mutate({
                name: 'vscode_execute_command',
                args: {
                    command: 'vscode.open',
                    args: [node.id] // Start with path, hopefully VS Code resolves workspace relative
                }
            });
        }
    };

    if (isLoading) {
        return <div className="flex h-full items-center justify-center"><Loader2 className="animate-spin h-8 w-8 text-neutral-500" /></div>;
    }

    return (
        <div className="flex flex-col h-full w-full bg-neutral-950 text-white relative">
            <div className="absolute top-4 right-4 z-10 flex gap-2">
                <Button
                    variant={showSymbols ? "secondary" : "outline"}
                    size="sm"
                    onClick={() => setShowSymbols(!showSymbols)}
                    className="text-xs"
                >
                    {showSymbols ? "Hide Symbols" : "Show Symbols"}
                </Button>
                <Button variant="outline" size="icon" onClick={() => refetch()}><RefreshCw className="h-4 w-4" /></Button>
                <Button variant="outline" size="icon" onClick={() => fgRef.current?.zoomToFit(400)}><ZoomOut className="h-4 w-4" /></Button>
            </div>

            <div ref={containerRef} className="flex-1 w-full h-full overflow-hidden">
                {graphData && (
                    <ForceGraph2D
                        ref={fgRef}
                        width={width}
                        height={height}
                        graphData={displayGraph}
                        nodeLabel="id"
                        nodeColor={(node: any) => {
                            if (node.group === 'symbol') {
                                return node.kind === 'class' ? '#facc15' : '#3b82f6'; // Yellow Class, Blue Function
                            }
                            if (node.group === 'packages') return '#ef4444'; // Red for packages
                            if (node.group === 'apps') return '#3b82f6'; // Blue for apps
                            return '#10b981'; // Green for others
                        }}
                        nodeVal={(node: any) => node.group === 'symbol' ? 2 : (node.val || 4)}
                        nodeRelSize={6}
                        linkColor={() => 'rgba(255,255,255,0.2)'}
                        backgroundColor="#0a0a0a"
                        onNodeClick={handleNodeClick}
                        cooldownTicks={100}
                    />
                )}
            </div>
        </div>
    );
}
