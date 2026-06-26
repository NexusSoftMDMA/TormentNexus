"use client";
import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { trpc } from '@/utils/trpc';

type SearchResult = {
    file: string;
    snippet: string;
    line?: number;
    character?: number;
    uri?: string;
};

type SymbolLike = {
    name?: unknown;
    containerName?: unknown;
    location?: {
        uri?: unknown;
        range?: {
            start?: {
                line?: unknown;
                character?: unknown;
            };
        };
    };
};

function normalizeSymbols(value: unknown): SymbolLike[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value.filter((item): item is SymbolLike => typeof item === 'object' && item !== null);
}

export const GlobalSearch: React.FC = () => {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState<SearchResult[]>([]);
    const [isOpen, setIsOpen] = useState(false);
    const [isSearching, setIsSearching] = useState(false);
    const searchQuery = trpc.lsp.searchSymbols.useQuery(
        { query },
        { enabled: false, refetchOnWindowFocus: false }
    );

    const handleSearch = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!query.trim()) return;
        setIsOpen(true);
        setIsSearching(true);

        try {
            const { data: symbols } = await searchQuery.refetch();
            const normalizedSymbols = normalizeSymbols(symbols);
            const mapped: SearchResult[] = normalizedSymbols.slice(0, 50).map((s) => {
                const uri: string = String(s?.location?.uri ?? '');
                const file = uri.startsWith('file://') ? decodeURIComponent(uri.replace('file://', '')) : uri;
                const line = Number(s?.location?.range?.start?.line ?? 0);
                const character = Number(s?.location?.range?.start?.character ?? 0);
                return {
                    file,
                    line,
                    character,
                    uri,
                    snippet: `${s?.name ?? 'symbol'} (${s?.containerName ?? 'global'})`,
                };
            });

            setResults(mapped.length > 0 ? mapped : [{ file: 'No results', snippet: 'No matching symbols found in LSP index.' }]);
        } catch (error: unknown) {
            const message = error instanceof Error ? error.message : 'Unable to query symbol index.';
            setResults([{ file: 'Search failed', snippet: message }]);
        } finally {
            setIsSearching(false);
        }
    };

    const handleOpenFile = async (res: SearchResult) => {
        if (!res.file || res.file === 'No results' || res.file === 'Search failed') {
            return;
        }

        const line = (res.line ?? 0) + 1;
        const col = (res.character ?? 0) + 1;
        const vscodeUrl = `vscode://file/${encodeURIComponent(res.file)}:${line}:${col}`;
        window.open(vscodeUrl, '_blank');

        try {
            await navigator.clipboard.writeText(`${res.file}:${line}:${col}`);
        } catch {
            // Ignore clipboard failures in restricted browsers.
        }

        setIsOpen(false);
        setQuery('');
    };

    return (
        <div className="relative z-50">
            <form onSubmit={handleSearch} className="relative">
                <input
                    type="text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="Search codebase..."
                    className="w-64 bg-zinc-100 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-full px-4 py-1.5 text-sm text-zinc-800 dark:text-zinc-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:w-96 transition-all"
                />
                <button type="submit" className="absolute right-3 top-1.5 text-zinc-400 hover:text-blue-500">
                    🔍
                </button>
            </form>

            <AnimatePresence>
                {isOpen && (results.length > 0 || isSearching) && (
                    <motion.div
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0 }}
                        className="absolute right-0 top-12 w-96 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-2xl overflow-hidden"
                    >
                        {isSearching ? (
                            <div className="p-4 text-center text-zinc-500 text-sm">Searching vector index...</div>
                        ) : (
                            <div className="max-h-96 overflow-y-auto">
                                {results.map((res, i) => (
                                    <button
                                        key={i}
                                        onClick={() => handleOpenFile(res)}
                                        className="w-full text-left p-3 hover:bg-zinc-100 dark:hover:bg-zinc-800 border-b border-zinc-100 dark:border-zinc-800 last:border-0 transition-colors"
                                    >
                                        <div className="text-xs font-bold text-blue-500 break-all">{res.file}</div>
                                        <div className="text-xs text-zinc-500 mt-1 line-clamp-2 font-mono bg-zinc-50 dark:bg-black p-1 rounded">
                                            {res.snippet}
                                        </div>
                                    </button>
                                ))}
                            </div>
                        )}
                        <div className="bg-zinc-50 dark:bg-black p-2 text-[10px] text-center text-zinc-400 border-t border-zinc-200 dark:border-zinc-800">
                            Press ESC to close
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>

            {isOpen && (
                <div
                    className="fixed inset-0 z-[-1]"
                    onClick={() => setIsOpen(false)}
                />
            )}
        </div>
    );
};
