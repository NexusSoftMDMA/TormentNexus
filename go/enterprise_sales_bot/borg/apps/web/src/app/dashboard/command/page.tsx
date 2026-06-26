"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent, Badge, Button, ScrollArea } from "@tormentnexus/ui";
import { Terminal, Play, Loader2, ChevronRight, Search, Trash2 } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

type CommandResult = {
    input: string;
    output: string | null;
    error: unknown;
    handled: boolean;
    timestamp: number;
};

export default function CommandDashboard() {
    const [input, setInput] = useState('');
    const [history, setHistory] = useState<CommandResult[]>([]);
    const [historyIndex, setHistoryIndex] = useState(-1);
    const [filter, setFilter] = useState('');

    const commandsQuery = trpc.commands.list.useQuery();
    const executeMutation = trpc.commands.execute.useMutation({
        onSuccess: (result) => {
            setHistory(prev => [{
                input: currentInput,
                output: result.output ?? null,
                error: result.error ?? null,
                handled: result.handled,
                timestamp: Date.now(),
            }, ...prev]);
        },
        onError: (err) => {
            toast.error(`Command failed: ${err.message}`);
        },
    });

    // Track the input at time of submission for the history entry.
    const [currentInput, setCurrentInput] = useState('');

    const handleExecute = () => {
        const trimmed = input.trim();
        if (!trimmed) return;
        setCurrentInput(trimmed);
        setHistoryIndex(-1);
        executeMutation.mutate({ input: trimmed });
        setInput('');
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter') {
            handleExecute();
            return;
        }
        // Navigate command history with arrow keys.
        if (e.key === 'ArrowUp') {
            e.preventDefault();
            const next = Math.min(historyIndex + 1, history.length - 1);
            setHistoryIndex(next);
            if (history[next]) setInput(history[next].input);
        }
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            const next = Math.max(historyIndex - 1, -1);
            setHistoryIndex(next);
            setInput(next === -1 ? '' : (history[next]?.input ?? ''));
        }
    };

    const filteredCommands = (commandsQuery.data ?? []).filter(
        cmd => !filter || cmd.name.toLowerCase().includes(filter.toLowerCase()) || cmd.description?.toLowerCase().includes(filter.toLowerCase())
    );

    return (
        <div className="p-8 space-y-8">
            <div>
                <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                    <Terminal className="h-8 w-8 text-emerald-500" />
                    Command Center
                </h1>
                <p className="text-zinc-500 mt-2">
                    Execute slash commands and inspect available command handlers registered with TormentNexus Core.
                </p>
            </div>

            <div className="grid gap-8 lg:grid-cols-2">
                {/* Left: Available Commands */}
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <h2 className="text-sm font-bold text-zinc-400 uppercase tracking-widest">Available Commands</h2>
                        <Badge variant="secondary" className="text-xs">
                            {filteredCommands.length} of {commandsQuery.data?.length ?? 0}
                        </Badge>
                    </div>

                    {/* Filter */}
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500 pointer-events-none" />
                        <input
                            type="text"
                            value={filter}
                            onChange={e => setFilter(e.target.value)}
                            placeholder="Filter commands…"
                            className="w-full bg-zinc-900 border border-zinc-800 rounded-md pl-9 pr-3 py-2 text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                        />
                    </div>

                    <ScrollArea className="h-[420px]">
                        {commandsQuery.isLoading ? (
                            <div className="flex justify-center p-8">
                                <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
                            </div>
                        ) : filteredCommands.length === 0 ? (
                            <div className="text-center p-8 text-zinc-600 text-sm border border-dashed border-zinc-800 rounded-lg">
                                {filter ? 'No commands match the filter.' : 'No commands registered.'}
                            </div>
                        ) : (
                            <div className="space-y-2 pr-2">
                                {filteredCommands.map(cmd => (
                                    <button
                                        key={cmd.name}
                                        onClick={() => setInput(`/${cmd.name} `)}
                                        className="w-full text-left p-3 rounded-lg bg-zinc-900 border border-zinc-800 hover:border-emerald-500/50 hover:bg-zinc-800 transition-colors group"
                                    >
                                        <div className="flex items-center justify-between">
                                            <span className="font-mono text-sm text-emerald-400 group-hover:text-emerald-300">
                                                /{cmd.name}
                                            </span>
                                            <ChevronRight className="h-3.5 w-3.5 text-zinc-600 group-hover:text-zinc-400" />
                                        </div>
                                        {cmd.description && (
                                            <p className="text-xs text-zinc-500 mt-1">{cmd.description}</p>
                                        )}
                                    </button>
                                ))}
                            </div>
                        )}
                    </ScrollArea>
                </div>

                {/* Right: REPL */}
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <h2 className="text-sm font-bold text-zinc-400 uppercase tracking-widest">Command REPL</h2>
                        {history.length > 0 && (
                            <Button
                                variant="ghost"
                                size="sm"
                                className="text-zinc-600 hover:text-red-400 h-7 px-2"
                                onClick={() => setHistory([])}
                            >
                                <Trash2 className="h-3.5 w-3.5 mr-1" />
                                Clear
                            </Button>
                        )}
                    </div>

                    {/* Input row */}
                    <div className="flex gap-2">
                        <div className="flex-1 flex items-center bg-zinc-900 border border-zinc-800 rounded-lg px-3 focus-within:ring-1 focus-within:ring-emerald-500 focus-within:border-emerald-500/50">
                            <span className="text-emerald-500 text-sm font-mono mr-2 select-none">$</span>
                            <input
                                type="text"
                                value={input}
                                onChange={e => setInput(e.target.value)}
                                onKeyDown={handleKeyDown}
                                placeholder="Type a command (e.g. /help) or free text…"
                                className="flex-1 bg-transparent py-2.5 text-sm font-mono text-white placeholder:text-zinc-600 outline-none"
                                autoComplete="off"
                                spellCheck={false}
                            />
                        </div>
                        <Button
                            onClick={handleExecute}
                            disabled={!input.trim() || executeMutation.isPending}
                            className="bg-emerald-600 hover:bg-emerald-500 text-white px-4 shrink-0"
                        >
                            {executeMutation.isPending ? (
                                <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                                <Play className="h-4 w-4" />
                            )}
                        </Button>
                    </div>

                    {/* Output history */}
                    <ScrollArea className="h-[360px]">
                        {history.length === 0 ? (
                            <div className="flex flex-col items-center justify-center h-32 text-zinc-600 text-sm border border-dashed border-zinc-800 rounded-lg">
                                <Terminal className="h-8 w-8 mb-2 opacity-30" />
                                Command output will appear here
                            </div>
                        ) : (
                            <div className="space-y-3 pr-2">
                                {history.map((entry, i) => (
                                    <Card key={i} className={`bg-zinc-950 border ${entry.handled ? 'border-zinc-800' : 'border-yellow-900/50'}`}>
                                        <CardContent className="p-3 space-y-2">
                                            {/* Input echo */}
                                            <div className="flex items-center gap-2">
                                                <span className="text-emerald-500 font-mono text-xs select-none">$</span>
                                                <span className="font-mono text-sm text-zinc-300">{entry.input}</span>
                                                {!entry.handled && (
                                                    <Badge variant="outline" className="ml-auto text-yellow-500 border-yellow-700 text-[10px]">
                                                        Unhandled
                                                    </Badge>
                                                )}
                                            </div>
                                            {/* Output */}
                                            {entry.output != null && (
                                                <pre className="text-xs font-mono text-zinc-400 whitespace-pre-wrap break-words bg-black/40 rounded p-2 max-h-48 overflow-y-auto">
                                                    {entry.output}
                                                </pre>
                                            )}
                                            {/* Error */}
                                            {entry.error != null && (
                                                <pre className="text-xs font-mono text-red-400 whitespace-pre-wrap bg-red-950/20 rounded p-2">
                                                    {String(entry.error)}
                                                </pre>
                                            )}
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>
                        )}
                    </ScrollArea>
                </div>
            </div>
        </div>
    );
}
