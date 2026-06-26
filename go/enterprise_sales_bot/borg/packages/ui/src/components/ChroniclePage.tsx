"use client";

import React, { useMemo, useState } from 'react';
import { trpc } from '../utils/trpc';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { Button } from './ui/button';
import { History, GitCommit, AlertTriangle, Info, Shield, RefreshCw, Undo2 } from 'lucide-react';

type AuditLogLike = {
    timestamp?: number;
    level?: string;
    event?: string;
    action?: string;
    agentId?: string;
};

type GitCommitLike = {
    date?: string | number;
    hash: string;
    message?: string;
    author?: string;
};

type GitStatusLike = {
    branch: string;
    clean: boolean;
    modified: string[];
};

function normalizeAuditLogs(value: unknown): AuditLogLike[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value.filter((item): item is AuditLogLike => typeof item === 'object' && item !== null);
}

function normalizeGitCommits(value: unknown): GitCommitLike[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value
        .filter((item): item is { hash: string; date?: unknown; message?: unknown; author?: unknown } => {
            return typeof item === 'object' && item !== null && typeof item.hash === 'string';
        })
        .map((item) => ({
            hash: item.hash,
            date: typeof item.date === 'string' || typeof item.date === 'number' ? item.date : Date.now(),
            message: typeof item.message === 'string' ? item.message : '',
            author: typeof item.author === 'string' ? item.author : '',
        }));
}

function normalizeGitStatus(value: unknown): GitStatusLike {
    if (typeof value !== 'object' || value === null) {
        return { branch: '...', clean: true, modified: [] };
    }

    const branch = (value as { branch?: unknown }).branch;
    const clean = (value as { clean?: unknown }).clean;
    const modified = (value as { modified?: unknown }).modified;

    return {
        branch: typeof branch === 'string' ? branch : '...',
        clean: typeof clean === 'boolean' ? clean : true,
        modified: Array.isArray(modified) ? modified.filter((entry): entry is string => typeof entry === 'string') : [],
    };
}

export function ChroniclePage() {
    const [limit, setLimit] = useState(50);
    const { data: rawAuditLogs, isLoading: auditLoading, refetch: refetchAudit } = trpc.audit.list.useQuery({ limit });
    const { data: rawGitLog, isLoading: gitLoading, refetch: refetchGit } = trpc.git.getLog.useQuery({ limit });
    const { data: rawGitStatus } = trpc.git.getStatus.useQuery();
    const revertMutation = trpc.git.revert.useMutation();
    const auditLogs = normalizeAuditLogs(rawAuditLogs);
    const gitLog = normalizeGitCommits(rawGitLog);
    const gitStatus = normalizeGitStatus(rawGitStatus);

    type ChronicleEvent =
        | { type: 'audit'; date: Date; data: AuditLogLike }
        | { type: 'git'; date: Date; data: GitCommitLike };

    const events = useMemo<ChronicleEvent[]>(() => {
        const combined: ChronicleEvent[] = [];

        combined.push(
            ...auditLogs.map((log) => ({
                type: 'audit' as const,
                date: new Date(typeof log.timestamp === 'number' ? log.timestamp : Date.now()),
                data: log,
            })),
        );

        combined.push(
            ...gitLog.map((commit) => ({
                type: 'git' as const,
                date: new Date(commit.date ?? Date.now()),
                data: commit,
            })),
        );

        combined.sort((a, b) => b.date.getTime() - a.date.getTime());
        return combined;
    }, [rawAuditLogs, rawGitLog]);

    const handleRevert = async (hash: string) => {
        if (confirm(`Are you sure you want to revert commit ${hash.substring(0, 7)}? This will create a new commit undoing changes.`)) {
            try {
                await revertMutation.mutateAsync({ hash });
                refetchGit();
                alert('Revert successful.');
            } catch (e: any) {
                alert(`Revert failed: ${e.message}`);
            }
        }
    };

    return (
        <div className="flex flex-col h-full bg-black text-white p-6 gap-6 overflow-hidden">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-zinc-200 tracking-tight flex items-center gap-2">
                        <History className="w-6 h-6 text-purple-400" />
                        The Chronicle
                    </h1>
                    <p className="text-zinc-500 text-sm">System Timeline & Audit Log</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={() => { refetchAudit(); refetchGit(); }} className="border-zinc-800 text-zinc-400">
                        <RefreshCw className="w-4 h-4 mr-2" /> Refresh
                    </Button>
                </div>
            </div>

            <div className="flex gap-6 flex-1 overflow-hidden">
                {/* Timeline Main */}
                <div className="flex-1 flex flex-col gap-4 overflow-hidden">
                    <Card className="flex-1 bg-zinc-900 border-zinc-800 flex flex-col overflow-hidden">
                        <CardHeader className="p-4 border-b border-zinc-800 bg-zinc-950/50">
                            <CardTitle className="text-sm font-bold text-zinc-300">Timeline</CardTitle>
                        </CardHeader>
                        <CardContent className="p-0 overflow-y-auto flex-1 custom-scrollbar">
                            <div className="relative border-l border-zinc-800 ml-6 my-4 space-y-6">
                                {events.map((event, idx) => (
                                    <div key={idx} className="relative pl-6">
                                        <div className={`absolute -left-1.5 mt-1.5 w-3 h-3 rounded-full border-2 ${event.type === 'git' ? 'border-blue-500 bg-blue-900' : event.data.level === 'ERROR' ? 'border-red-500 bg-red-900' : 'border-zinc-500 bg-zinc-900'}`} />

                                        {event.type === 'git' ? (
                                            <div className="bg-zinc-950/50 border border-zinc-800 p-3 rounded-md hover:border-zinc-700 transition-colors">
                                                <div className="flex justify-between items-start mb-1">
                                                    <span className="text-xs font-mono text-blue-400 flex items-center gap-1">
                                                        <GitCommit className="w-3 h-3" />
                                                        {event.data.hash.substring(0, 7)}
                                                    </span>
                                                    <span className="text-xs text-zinc-500">{event.date.toLocaleString()}</span>
                                                </div>
                                                <div className="text-sm font-medium text-zinc-200">{event.data.message}</div>
                                                <div className="flex justify-between items-center mt-2">
                                                    <span className="text-xs text-zinc-500">Author: {event.data.author}</span>
                                                    <Button variant="ghost" size="sm" onClick={() => handleRevert(event.data.hash)} className="h-6 text-xs text-zinc-500 hover:text-red-400">
                                                        <Undo2 className="w-3 h-3 mr-1" /> Revert
                                                    </Button>
                                                </div>
                                            </div>
                                        ) : (
                                            <div className="bg-zinc-950/30 border border-zinc-900 p-3 rounded-md">
                                                <div className="flex justify-between items-start mb-1">
                                                    <span className={`text-xs font-bold flex items-center gap-1 ${event.data.level === 'ERROR' ? 'text-red-400' : 'text-zinc-400'}`}>
                                                        {event.data.level === 'ERROR' ? <AlertTriangle className="w-3 h-3" /> : <Info className="w-3 h-3" />}
                                                        {event.data.level}
                                                    </span>
                                                    <span className="text-xs text-zinc-500">{event.date.toLocaleString()}</span>
                                                </div>
                                                <div className="text-sm text-zinc-300">{event.data.event}</div>
                                                {event.data.agentId && <div className="text-xs text-zinc-500 mt-1 flex items-center gap-1"><Shield className="w-3 h-3" /> Agent: {event.data.agentId}</div>}
                                            </div>
                                        )}
                                    </div>
                                ))}
                                {events.length === 0 && <div className="ml-6 text-zinc-500 text-sm">No events found.</div>}
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Status Side Panel */}
                <div className="w-80 flex flex-col gap-4">
                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="p-4 border-b border-zinc-800">
                            <CardTitle className="text-sm font-bold text-zinc-200">System Status</CardTitle>
                        </CardHeader>
                        <CardContent className="p-4 space-y-4">
                            <div>
                                <div className="text-xs text-zinc-500 uppercase font-bold mb-1">Git Branch</div>
                                <div className="text-sm font-mono text-blue-400">{gitStatus?.branch || '...'}</div>
                            </div>
                            <div>
                                <div className="text-xs text-zinc-500 uppercase font-bold mb-1">Workspace State</div>
                                <div className={`text-sm font-medium ${gitStatus?.clean ? 'text-emerald-400' : 'text-amber-400'}`}>
                                    {gitStatus?.clean ? 'Clean' : 'Dirty (Uncommitted Changes)'}
                                </div>
                            </div>
                            {gitStatus?.modified && gitStatus.modified.length > 0 && (
                                <div>
                                    <div className="text-xs text-zinc-500 uppercase font-bold mb-1">Modified Files</div>
                                    <div className="text-xs text-zinc-400 font-mono space-y-1">
                                        {gitStatus.modified.slice(0, 5).map((f: string) => <div key={f} className="truncate">{f}</div>)}
                                        {gitStatus.modified.length > 5 && <div>+ {gitStatus.modified.length - 5} more</div>}
                                    </div>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
