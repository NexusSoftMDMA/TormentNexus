'use client';

import { trpc } from '@/utils/trpc';
import { useState } from 'react';
import { Card } from '@tormentnexus/ui';
import { Input } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';
import { ScrollArea } from '@tormentnexus/ui';

export default function CodeDashboard() {
    const [filePath, setFilePath] = useState('packages/core/src/MCPServer.ts');
    const [query, setQuery] = useState('');

    // Determine which query to run based on input state
    const symbolsQuery = trpc.lsp.getSymbols.useQuery(
        { filePath },
        { enabled: !!filePath && !query }
    );

    const searchQuery = trpc.lsp.searchSymbols.useQuery(
        { query },
        { enabled: !!query }
    );

    const indexMutation = trpc.lsp.indexProject.useMutation();

    const results = query ? searchQuery.data : symbolsQuery.data;
    const isPending = query ? searchQuery.isPending : symbolsQuery.isPending;

    return (
        <div className="p-6 space-y-6 h-full flex flex-col">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-blue-400">Code Intelligence</h1>
                    <p className="text-muted-foreground">LSP Symbol Navigation & Search</p>
                </div>
                <Button
                    onClick={() => indexMutation.mutate()}
                    disabled={indexMutation.isPending}
                    variant="outline"
                >
                    {indexMutation.isPending ? 'Indexing...' : 'Re-Index Project'}
                </Button>
            </div>

            <Card className="p-4 flex gap-4 bg-gray-800 border-gray-700">
                <div className="flex-1">
                    <label className="text-xs text-gray-400 mb-1 block">File Path (Relative to Root)</label>
                    <Input
                        value={filePath}
                        onChange={(e) => { setFilePath(e.target.value); setQuery(''); }}
                        className="bg-gray-900 border-gray-600 font-mono text-sm"
                    />
                </div>
                <div className="flex-1">
                    <label className="text-xs text-gray-400 mb-1 block">Search Symbols (Global)</label>
                    <Input
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        className="bg-gray-900 border-gray-600 font-mono text-sm"
                        placeholder="e.g. MCPServer"
                    />
                </div>
            </Card>

            <Card className="flex-1 min-h-0 bg-gray-800 border-gray-700 overflow-hidden flex flex-col">
                <div className="p-3 border-b border-gray-700 bg-gray-900/50">
                    <h3 className="font-semibold text-sm text-gray-300">
                        {query ? `Search Results for "${query}"` : `Symbols in ${filePath}`}
                    </h3>
                </div>

                <ScrollArea className="flex-1 p-4">
                    {isPending && <div className="text-gray-500 animate-pulse">Loading symbols...</div>}

                    {!isPending && (!results || (Array.isArray(results) && results.length === 0)) && (
                        <div className="text-gray-500 italic">No symbols found. Try indexing the project.</div>
                    )}

                    <div className="space-y-2">
                        {Array.isArray(results) && results.map((symbol: any, idx: number) => (
                            <div key={idx} className="flex items-center justify-between p-2 rounded hover:bg-gray-700/50 group border border-transparent hover:border-gray-600 transition-colors">
                                <div className="flex items-center gap-3">
                                    <span className={`text-xs px-1.5 py-0.5 rounded border ${symbol.kind === 6 ? 'bg-blue-900/30 text-blue-400 border-blue-800' : // Method
                                        symbol.kind === 5 ? 'bg-yellow-900/30 text-yellow-400 border-yellow-800' : // Class
                                            symbol.kind === 13 ? 'bg-purple-900/30 text-purple-400 border-purple-800' : // Variable
                                                'bg-gray-800 text-gray-400 border-gray-700'
                                        }`}>
                                        {getSymbolKindName(symbol.kind)}
                                    </span>
                                    <span className="font-mono text-sm text-gray-200">
                                        {symbol.containerName ? <span className="text-gray-500">{symbol.containerName}.</span> : ''}
                                        {symbol.name}
                                    </span>
                                </div>
                                <div className="text-xs text-gray-500 font-mono opacity-50 group-hover:opacity-100">
                                    {symbol.location?.uri.split('/').pop()}:{symbol.location?.range?.start?.line + 1}
                                </div>
                            </div>
                        ))}
                    </div>
                </ScrollArea>
            </Card>
        </div>
    );
}

function getSymbolKindName(kind: number): string {
    const kinds: Record<number, string> = {
        1: 'File', 2: 'Module', 3: 'Namespace', 4: 'Package', 5: 'Class',
        6: 'Method', 7: 'Property', 8: 'Field', 9: 'Constructor', 10: 'Enum',
        11: 'Interface', 12: 'Function', 13: 'Variable', 14: 'Constant', 15: 'String'
    };
    return kinds[kind] || `Kind(${kind})`;
}
