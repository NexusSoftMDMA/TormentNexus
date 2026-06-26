
"use client";

import React, { useState } from 'react';
import { trpc } from '@/utils/trpc';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Code, Search, FileCode, Braces, Hash, Variable, Zap } from 'lucide-react';

const typeIcons: Record<string, React.ReactNode> = {
    function: <Code className="w-4 h-4 text-blue-400" />,
    class: <Braces className="w-4 h-4 text-yellow-400" />,
    method: <Hash className="w-4 h-4 text-purple-400" />,
    variable: <Variable className="w-4 h-4 text-green-400" />,
    interface: <FileCode className="w-4 h-4 text-orange-400" />,
};

export default function SymbolsPage() {
    const [query, setQuery] = useState("");
    const [debouncedQuery, setDebouncedQuery] = useState("");

    // Debounce
    React.useEffect(() => {
        const timer = setTimeout(() => setDebouncedQuery(query), 500);
        return () => clearTimeout(timer);
    }, [query]);

    const { data: results, isPending } = trpc.symbols.find.useQuery(
        { query: debouncedQuery, limit: 20 },
        { enabled: debouncedQuery.length > 2 }
    );

    const pinMutation = trpc.symbols.pin.useMutation();

    return (
        <div className="container mx-auto p-4 space-y-6 max-w-5xl h-screen flex flex-col">
            <header>
                <h1 className="text-3xl font-bold tracking-tight mb-2 flex items-center gap-2">
                    <Zap className="h-8 w-8 text-yellow-500" />
                    Deep Symbol Intelligence
                </h1>
                <p className="text-muted-foreground">
                    Semantic Search for Functions, Classes, and Interfaces across the codebase.
                </p>
            </header>

            <Card className="flex-1 flex flex-col overflow-hidden">
                <CardHeader className="pb-3 border-b border-border/50">
                    <div className="relative">
                        <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Search symbols (e.g. 'LLMService', 'executeTool', 'vector search')..."
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            className="pl-10 h-10 font-mono"
                        />
                    </div>
                </CardHeader>
                <CardContent className="flex-1 p-0 flex flex-col overflow-hidden">
                    <ScrollArea className="flex-1">
                        <div className="flex flex-col">
                            {isPending && (
                                <div className="p-8 text-center text-muted-foreground animate-pulse">
                                    Analyzing Codebase...
                                </div>
                            )}

                            {!isPending && results && results.length === 0 && debouncedQuery.length > 2 && (
                                <div className="p-8 text-center text-muted-foreground">
                                    No symbols found matching "{debouncedQuery}".
                                </div>
                            )}

                            {results?.map((res: any) => {
                                const meta = res.metadata || {};
                                const kind = meta.kind || 'unknown';
                                const Icon = typeIcons[kind] || <Code className="w-4 h-4" />;

                                return (
                                    <div key={res.id} className="p-4 border-b border-border/50 hover:bg-muted/30 transition-colors group">
                                        <div className="flex items-start justify-between mb-2">
                                            <div className="flex items-center gap-3">
                                                {Icon}
                                                <div>
                                                    <div className="font-mono font-semibold text-sm flex items-center gap-2">
                                                        {meta.name || res.id}
                                                        <Badge variant="outline" className="text-[10px] uppercase">{kind}</Badge>
                                                    </div>
                                                    <div className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                                                        <span>{meta.file_path}</span>
                                                        {meta.line && <span>:L{meta.line}</span>}
                                                    </div>
                                                </div>
                                            </div>
                                            <Button
                                                variant="secondary"
                                                size="sm"
                                                className="h-6 text-[10px] px-2 opacity-0 group-hover:opacity-100 transition-opacity"
                                                onClick={() => pinMutation.mutate({
                                                    name: meta.name || res.id,
                                                    file: meta.file_path || 'unknown',
                                                    type: (kind as any) || 'function'
                                                })}
                                            >
                                                Pin to Context
                                            </Button>
                                        </div>
                                        <div className="pl-7">
                                            {meta.docblock && (
                                                <p className="text-xs text-muted-foreground mb-2 italic">
                                                    {meta.docblock}
                                                </p>
                                            )}
                                            <pre className="text-[10px] bg-black/40 p-2 rounded border border-white/5 font-mono overflow-x-auto text-neutral-300">
                                                {meta.signature || res.content.substring(0, 200)}
                                            </pre>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    </ScrollArea>
                </CardContent>
            </Card>
        </div>
    );
}
