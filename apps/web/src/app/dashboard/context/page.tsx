"use client";

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, Button, Badge, ScrollArea } from "@tormentnexus/ui";
import { Layers, Plus, Trash2, Loader2, RefreshCw, FileText, Code2, Copy, Check } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

function normalizeContextFiles(data: unknown): string[] {
    if (!Array.isArray(data)) return [];
    return data.filter((f): f is string => typeof f === 'string');
}

export default function ContextDashboard() {
    const [newFile, setNewFile] = useState('');
    const [copied, setCopied] = useState(false);

    const utils = trpc.useUtils();
    const filesQuery = trpc.tormentnexusContext.list.useQuery();
    const promptQuery = trpc.tormentnexusContext.getPrompt.useQuery();

    const addMutation = trpc.tormentnexusContext.add.useMutation({
        onSuccess: async () => {
            toast.success('File added to context');
            setNewFile('');
            await utils.tormentnexusContext.list.invalidate();
            await utils.tormentnexusContext.getPrompt.invalidate();
        },
        onError: err => toast.error(`Failed to add: ${err.message}`),
    });

    const removeMutation = trpc.tormentnexusContext.remove.useMutation({
        onSuccess: async () => {
            toast.success('File removed from context');
            await utils.tormentnexusContext.list.invalidate();
            await utils.tormentnexusContext.getPrompt.invalidate();
        },
        onError: err => toast.error(`Failed to remove: ${err.message}`),
    });

    const clearMutation = trpc.tormentnexusContext.clear.useMutation({
        onSuccess: async () => {
            toast.success('Context cleared');
            await utils.tormentnexusContext.list.invalidate();
            await utils.tormentnexusContext.getPrompt.invalidate();
        },
        onError: err => toast.error(`Failed to clear: ${err.message}`),
    });

    const contextFiles = normalizeContextFiles(filesQuery.data);
    const promptText = typeof promptQuery.data === 'string' ? promptQuery.data : '';

    const handleAdd = () => {
        const trimmed = newFile.trim();
        if (!trimmed) return;
        addMutation.mutate({ filePath: trimmed });
    };

    const handleCopyPrompt = async () => {
        if (!promptText) return;
        await navigator.clipboard.writeText(promptText);
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
    };

    return (
        <div className="p-8 space-y-8">
            {/* Header */}
            <div className="flex items-start justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <Layers className="h-8 w-8 text-sky-500" />
                        Context Manager
                    </h1>
                    <p className="text-zinc-500 mt-2">
                        Manage the set of files that are injected into the TormentNexus context prompt for active AI sessions.
                    </p>
                </div>
                <Button
                    variant="outline"
                    size="sm"
                    className="border-zinc-700 text-zinc-400 hover:text-white"
                    onClick={() => { filesQuery.refetch(); promptQuery.refetch(); }}
                    disabled={filesQuery.isFetching || promptQuery.isFetching}
                >
                    {(filesQuery.isFetching || promptQuery.isFetching) ? (
                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ) : (
                        <RefreshCw className="h-4 w-4 mr-2" />
                    )}
                    Refresh
                </Button>
            </div>

            <div className="grid gap-8 xl:grid-cols-2">
                {/* Left: File List */}
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <h2 className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                            <FileText className="h-4 w-4" />
                            Context Files
                            {contextFiles.length > 0 && (
                                <Badge variant="secondary" className="ml-1 text-xs">
                                    {contextFiles.length}
                                </Badge>
                            )}
                        </h2>
                        {contextFiles.length > 0 && (
                            <Button
                                variant="ghost"
                                size="sm"
                                className="text-zinc-600 hover:text-red-400 h-7 px-2"
                                onClick={() => { if (confirm('Remove all context files?')) clearMutation.mutate(); }}
                                disabled={clearMutation.isPending}
                            >
                                {clearMutation.isPending ? (
                                    <Loader2 className="h-3.5 w-3.5 animate-spin mr-1" />
                                ) : (
                                    <Trash2 className="h-3.5 w-3.5 mr-1" />
                                )}
                                Clear All
                            </Button>
                        )}
                    </div>

                    {/* Add file input */}
                    <div className="flex gap-2">
                        <input
                            type="text"
                            value={newFile}
                            onChange={e => setNewFile(e.target.value)}
                            onKeyDown={e => e.key === 'Enter' && handleAdd()}
                            placeholder="Path to file (e.g. packages/core/src/trpc.ts)"
                            className="flex-1 bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-2 text-sm font-mono text-white placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500/50"
                        />
                        <Button
                            onClick={handleAdd}
                            disabled={!newFile.trim() || addMutation.isPending}
                            className="bg-sky-600 hover:bg-sky-500 text-white px-3 shrink-0"
                        >
                            {addMutation.isPending ? (
                                <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                                <Plus className="h-4 w-4" />
                            )}
                        </Button>
                    </div>

                    {/* List */}
                    {filesQuery.isLoading ? (
                        <div className="flex justify-center p-10">
                            <Loader2 className="h-6 w-6 animate-spin text-zinc-600" />
                        </div>
                    ) : contextFiles.length === 0 ? (
                        <div className="text-center p-10 text-zinc-600 text-sm border border-dashed border-zinc-800 rounded-lg">
                            <Layers className="h-8 w-8 mx-auto mb-3 opacity-30" />
                            No files in context. Add a file above.
                        </div>
                    ) : (
                        <Card className="bg-zinc-900 border-zinc-800">
                            <CardContent className="p-0">
                                <div className="divide-y divide-zinc-800">
                                    {contextFiles.map(file => (
                                        <div key={file} className="flex items-center gap-3 px-4 py-3 group hover:bg-zinc-800/40 transition-colors">
                                            <FileText className="h-4 w-4 text-sky-500 shrink-0" />
                                            <span className="font-mono text-xs text-zinc-300 flex-1 truncate" title={file}>
                                                {file}
                                            </span>
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-7 w-7 p-0 text-zinc-600 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                                                onClick={() => removeMutation.mutate({ filePath: file })}
                                                disabled={removeMutation.isPending && (removeMutation.variables as { filePath?: string } | undefined)?.filePath === file}
                                                title="Remove from context"
                                            >
                                                {removeMutation.isPending && (removeMutation.variables as { filePath?: string } | undefined)?.filePath === file ? (
                                                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                                                ) : (
                                                    <Trash2 className="h-3.5 w-3.5" />
                                                )}
                                            </Button>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    )}
                </div>

                {/* Right: Assembled Prompt */}
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <h2 className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                            <Code2 className="h-4 w-4" />
                            Assembled Context Prompt
                        </h2>
                        {promptText && (
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-7 px-2 text-zinc-500 hover:text-sky-400"
                                onClick={handleCopyPrompt}
                            >
                                {copied ? (
                                    <><Check className="h-3.5 w-3.5 mr-1 text-green-400" />Copied</>
                                ) : (
                                    <><Copy className="h-3.5 w-3.5 mr-1" />Copy</>
                                )}
                            </Button>
                        )}
                    </div>

                    <Card className="bg-zinc-950 border-zinc-800">
                        <CardHeader className="pb-2 border-b border-zinc-800">
                            <CardTitle className="text-xs text-zinc-500 font-mono">context_prompt.txt</CardTitle>
                        </CardHeader>
                        <CardContent className="p-0">
                            {promptQuery.isLoading ? (
                                <div className="flex justify-center p-10">
                                    <Loader2 className="h-6 w-6 animate-spin text-zinc-600" />
                                </div>
                            ) : !promptText ? (
                                <div className="text-center p-10 text-zinc-600 text-sm">
                                    No context files — prompt is empty.
                                </div>
                            ) : (
                                <ScrollArea className="h-[480px]">
                                    <pre className="p-4 text-xs font-mono text-zinc-300 whitespace-pre-wrap break-words leading-relaxed">
                                        {promptText}
                                    </pre>
                                </ScrollArea>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
