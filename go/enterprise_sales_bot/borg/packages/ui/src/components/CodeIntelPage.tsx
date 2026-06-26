
import React from 'react';
import { trpc } from '../utils/trpc';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { DependencyGraphWidget } from '../widgets/DependencyGraphWidget';
import { Loader2, RefreshCw, Network } from 'lucide-react';
import { Button } from './ui/button';

export function CodeIntelPage() {
    const { data: graph, isLoading, refetch, isRefetching } = trpc.graph.get.useQuery(undefined, {
        staleTime: 60000 // Cache for 1 min
    });

    return (
        <div className="p-6 space-y-6 max-w-[1600px] mx-auto text-white h-full flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">Code Intelligence</h1>
                    <p className="text-white/40">Dependency Graph & Codebase Analysis</p>
                </div>
                <div className="flex gap-2">
                    <Button
                        variant="outline"
                        onClick={() => refetch()}
                        disabled={isLoading || isRefetching}
                        className="bg-zinc-900 border-zinc-700 hover:bg-zinc-800"
                    >
                        {isRefetching ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <RefreshCw className="w-4 h-4 mr-2" />}
                        Refresh Graph
                    </Button>
                </div>
            </div>

            <Card className="flex-1 bg-zinc-900 border-zinc-800 flex flex-col overflow-hidden">
                <CardHeader className="border-b border-zinc-800 bg-zinc-950/50">
                    <CardTitle className="flex items-center gap-2">
                        <Network className="w-5 h-5 text-indigo-400" />
                        Deep Dependency Graph
                    </CardTitle>
                    <CardDescription>
                        Visualizing {graph?.nodes.length || 0} files and {graph?.links.length || 0} dependencies
                    </CardDescription>
                </CardHeader>
                <CardContent className="flex-1 p-0 relative bg-zinc-950">
                    {isLoading ? (
                        <div className="absolute inset-0 flex items-center justify-center text-zinc-500 gap-2">
                            <Loader2 className="w-6 h-6 animate-spin" />
                            Parsing Codebase AST...
                        </div>
                    ) : (
                        <DependencyGraphWidget data={graph || { nodes: [], links: [] }} />
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
