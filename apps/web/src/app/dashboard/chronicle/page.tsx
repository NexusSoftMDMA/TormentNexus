"use client";

import { useState } from 'react';
import { Card, CardContent, Badge, Button } from "@tormentnexus/ui";
import { ScrollText, GitCommit, GitBranch, Loader2, RefreshCw, AlertCircle, GitMerge, Plus, Minus } from "lucide-react";
import { trpc } from '@/utils/trpc';

type CommitEntry = {
    hash: string;
    message: string;
    author?: string;
    date?: string;
};

type StatusEntry = {
    path: string;
    status: string;
};

function normalizeCommits(data: unknown): CommitEntry[] {
    if (!Array.isArray(data)) return [];
    return data.map((entry: unknown) => {
        if (!entry || typeof entry !== 'object') return { hash: '', message: '' };
        const e = entry as Record<string, unknown>;
        return {
            hash: typeof e['hash'] === 'string' ? e['hash'] : '',
            message: typeof e['message'] === 'string' ? e['message'] : '',
            author: typeof e['author'] === 'string' ? e['author'] : undefined,
            date: typeof e['date'] === 'string' ? e['date'] : undefined,
        };
    });
}

function normalizeStatus(data: unknown): StatusEntry[] {
    if (!Array.isArray(data)) return [];
    return data.map((entry: unknown) => {
        if (!entry || typeof entry !== 'object') return { path: '', status: '' };
        const e = entry as Record<string, unknown>;
        return {
            path: typeof e['path'] === 'string' ? e['path'] : '',
            status: typeof e['status'] === 'string' ? e['status'] : '?',
        };
    });
}

function statusColor(status: string): string {
    switch (status.toUpperCase()) {
        case 'M': return 'text-yellow-400';
        case 'A': return 'text-green-400';
        case 'D': return 'text-red-400';
        case 'R': return 'text-blue-400';
        case '??': return 'text-zinc-500';
        default: return 'text-zinc-400';
    }
}

function statusLabel(status: string): string {
    switch (status.toUpperCase()) {
        case 'M': return 'Modified';
        case 'A': return 'Added';
        case 'D': return 'Deleted';
        case 'R': return 'Renamed';
        case '??': return 'Untracked';
        default: return status;
    }
}

function shortHash(hash: string): string {
    return hash?.slice(0, 8) ?? '—';
}

export default function ChronicleDashboard() {
    const [limit, setLimit] = useState(30);

    const logQuery = trpc.git.getLog.useQuery({ limit });
    const statusQuery = trpc.git.getStatus.useQuery();

    const commits = normalizeCommits(logQuery.data);
    const statusFiles = normalizeStatus(statusQuery.data);

    return (
        <div className="p-8 space-y-8">
            <div className="flex items-start justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <ScrollText className="h-8 w-8 text-violet-500" />
                        Chronicle
                    </h1>
                    <p className="text-zinc-500 mt-2">
                        Git commit log and working-tree status for the active TormentNexus workspace.
                    </p>
                </div>
                <Button
                    variant="outline"
                    size="sm"
                    className="border-zinc-700 text-zinc-400 hover:text-white"
                    onClick={() => { logQuery.refetch(); statusQuery.refetch(); }}
                    disabled={logQuery.isFetching || statusQuery.isFetching}
                >
                    {(logQuery.isFetching || statusQuery.isFetching) ? (
                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ) : (
                        <RefreshCw className="h-4 w-4 mr-2" />
                    )}
                    Refresh
                </Button>
            </div>

            <div className="grid gap-8 xl:grid-cols-3">
                {/* Commit Log – takes 2 columns */}
                <div className="xl:col-span-2 space-y-4">
                    <div className="flex items-center justify-between">
                        <h2 className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                            <GitCommit className="h-4 w-4" />
                            Commit History
                        </h2>
                        <select
                            value={limit}
                            onChange={e => setLimit(Number(e.target.value))}
                            className="bg-zinc-900 border border-zinc-800 text-zinc-400 text-xs rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-violet-500"
                        >
                            {[20, 30, 50, 100].map(n => (
                                <option key={n} value={n}>Last {n}</option>
                            ))}
                        </select>
                    </div>

                    {logQuery.isLoading ? (
                        <div className="flex justify-center p-12">
                            <Loader2 className="h-7 w-7 animate-spin text-zinc-500" />
                        </div>
                    ) : logQuery.isError ? (
                        <div className="flex items-center gap-3 text-red-400 bg-red-900/10 border border-red-900/40 rounded-lg p-4">
                            <AlertCircle className="h-5 w-5 shrink-0" />
                            <span className="text-sm">Failed to load git log: {logQuery.error.message}</span>
                        </div>
                    ) : commits.length === 0 ? (
                        <div className="text-center p-10 text-zinc-600 text-sm border border-dashed border-zinc-800 rounded-lg">
                            No commits found.
                        </div>
                    ) : (
                        <div className="space-y-0 border border-zinc-800 rounded-lg overflow-hidden">
                            {commits.map((commit, i) => (
                                <div
                                    key={commit.hash || i}
                                    className="flex items-start gap-4 px-4 py-3 border-b border-zinc-800/50 last:border-b-0 hover:bg-zinc-900/50 transition-colors group"
                                >
                                    <GitCommit className="h-4 w-4 text-violet-500 mt-0.5 shrink-0" />
                                    <div className="flex-1 min-w-0">
                                        <p className="text-sm text-zinc-200 leading-snug truncate group-hover:text-white">
                                            {commit.message || '(no message)'}
                                        </p>
                                        <div className="flex items-center gap-3 mt-1">
                                            <span className="font-mono text-[11px] text-violet-400">
                                                {shortHash(commit.hash)}
                                            </span>
                                            {commit.author && (
                                                <span className="text-[11px] text-zinc-500">{commit.author}</span>
                                            )}
                                            {commit.date && (
                                                <span className="text-[11px] text-zinc-600">{commit.date}</span>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* Working Tree Status */}
                <div className="space-y-4">
                    <h2 className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                        <GitBranch className="h-4 w-4" />
                        Working Tree
                        {statusFiles.length > 0 && (
                            <Badge variant="secondary" className="ml-auto text-xs">
                                {statusFiles.length} changed
                            </Badge>
                        )}
                    </h2>

                    {statusQuery.isLoading ? (
                        <div className="flex justify-center p-8">
                            <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
                        </div>
                    ) : statusQuery.isError ? (
                        <div className="flex items-center gap-2 text-red-400 text-sm p-3 bg-red-900/10 rounded-lg border border-red-900/30">
                            <AlertCircle className="h-4 w-4 shrink-0" />
                            {statusQuery.error.message}
                        </div>
                    ) : statusFiles.length === 0 ? (
                        <div className="text-center p-8 text-zinc-600 text-sm border border-dashed border-zinc-800 rounded-lg">
                            Working tree clean
                        </div>
                    ) : (
                        <Card className="bg-zinc-900 border-zinc-800">
                            <CardContent className="p-0">
                                <div className="divide-y divide-zinc-800">
                                    {statusFiles.map((file, i) => (
                                        <div key={file.path || i} className="flex items-center gap-3 px-4 py-2.5">
                                            <span className={`font-mono text-xs font-bold w-6 text-center shrink-0 ${statusColor(file.status)}`}>
                                                {file.status}
                                            </span>
                                            <span className="font-mono text-xs text-zinc-300 truncate flex-1" title={file.path}>
                                                {file.path}
                                            </span>
                                            <span className={`text-[10px] shrink-0 ${statusColor(file.status)}`}>
                                                {statusLabel(file.status)}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    )}
                </div>
            </div>
        </div>
    );
}
