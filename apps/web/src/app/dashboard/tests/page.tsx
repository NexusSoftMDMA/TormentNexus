"use client";

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, Button, Badge, ScrollArea } from "@tormentnexus/ui";
import { FlaskConical, Play, Square, Loader2, RefreshCw, CheckCircle2, XCircle, Clock, AlertCircle, RotateCcw, ChevronDown, ChevronRight } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

type TestResult = {
    file: string;
    status: 'pass' | 'fail' | 'running' | string;
    timestamp: number;
    output?: string;
};

function normalizeResults(data: unknown): TestResult[] {
    if (!Array.isArray(data)) return [];
    return data.map((r: unknown) => {
        if (!r || typeof r !== 'object') return { file: '', status: 'unknown', timestamp: 0 };
        const entry = r as Record<string, unknown>;
        return {
            file: typeof entry['file'] === 'string' ? entry['file'] : '',
            status: typeof entry['status'] === 'string' ? entry['status'] : 'unknown',
            timestamp: typeof entry['timestamp'] === 'number' ? entry['timestamp'] : 0,
            output: typeof entry['output'] === 'string' ? entry['output'] : undefined,
        };
    });
}

function normalizeStatus(data: unknown): { isRunning: boolean; results: Record<string, TestResult> } {
    if (!data || typeof data !== 'object') return { isRunning: false, results: {} };
    const s = data as Record<string, unknown>;
    return {
        isRunning: s['isRunning'] === true,
        results: (s['results'] && typeof s['results'] === 'object' && !Array.isArray(s['results']))
            ? s['results'] as Record<string, TestResult>
            : {},
    };
}

function statusIcon(status: string) {
    switch (status) {
        case 'pass': return <CheckCircle2 className="h-4 w-4 text-green-400 shrink-0" />;
        case 'fail': return <XCircle className="h-4 w-4 text-red-400 shrink-0" />;
        case 'running': return <Loader2 className="h-4 w-4 text-yellow-400 animate-spin shrink-0" />;
        default: return <AlertCircle className="h-4 w-4 text-zinc-500 shrink-0" />;
    }
}

function statusBadgeClass(status: string): string {
    switch (status) {
        case 'pass': return 'bg-green-500/10 text-green-400 border-green-800';
        case 'fail': return 'bg-red-500/10 text-red-400 border-red-800';
        case 'running': return 'bg-yellow-500/10 text-yellow-400 border-yellow-800';
        default: return 'bg-zinc-800 text-zinc-500 border-zinc-700';
    }
}

function formatRelativeTime(ts: number): string {
    if (!ts) return '—';
    const delta = Math.round((Date.now() - ts) / 1000);
    if (delta < 60) return `${delta}s ago`;
    const m = Math.round(delta / 60);
    if (m < 60) return `${m}m ago`;
    return `${Math.round(m / 60)}h ago`;
}

function TestResultCard({ result }: { result: TestResult }) {
    const [expanded, setExpanded] = useState(result.status === 'fail');

    return (
        <div className={`rounded-lg border ${result.status === 'fail' ? 'border-red-900/50 bg-zinc-950' : 'border-zinc-800 bg-zinc-900'}`}>
            <button
                className="w-full flex items-center gap-3 px-4 py-3 hover:bg-zinc-800/30 transition-colors text-left"
                onClick={() => result.output && setExpanded(e => !e)}
            >
                {statusIcon(result.status)}
                <span className="font-mono text-sm text-zinc-300 flex-1 truncate" title={result.file}>
                    {result.file}
                </span>
                <span className="text-xs text-zinc-600 shrink-0 ml-2">{formatRelativeTime(result.timestamp)}</span>
                <Badge
                    variant="outline"
                    className={`text-[10px] h-5 ml-2 shrink-0 ${statusBadgeClass(result.status)}`}
                >
                    {result.status}
                </Badge>
                {result.output && (
                    expanded
                        ? <ChevronDown className="h-3.5 w-3.5 text-zinc-600 shrink-0 ml-1" />
                        : <ChevronRight className="h-3.5 w-3.5 text-zinc-600 shrink-0 ml-1" />
                )}
            </button>
            {expanded && result.output && (
                <div className="border-t border-zinc-800 px-4 pb-4 pt-3">
                    <ScrollArea className="max-h-48">
                        <pre className="text-xs font-mono text-zinc-400 whitespace-pre-wrap break-words leading-relaxed">
                            {result.output}
                        </pre>
                    </ScrollArea>
                </div>
            )}
        </div>
    );
}

export default function TestsDashboard() {
    const statusQuery = trpc.tests.status.useQuery(undefined, { refetchInterval: 3000 });
    const resultsQuery = trpc.tests.results.useQuery(undefined, { refetchInterval: 3000 });

    const startMutation = trpc.tests.start.useMutation({
        onSuccess: () => { toast.success('Auto-test watcher started'); statusQuery.refetch(); },
        onError: err => toast.error(`Failed to start: ${err.message}`),
    });
    const stopMutation = trpc.tests.stop.useMutation({
        onSuccess: () => { toast.success('Auto-test watcher stopped'); statusQuery.refetch(); },
        onError: err => toast.error(`Failed to stop: ${err.message}`),
    });

    const normalized = normalizeStatus(statusQuery.data);
    const results = normalizeResults(resultsQuery.data);

    const passing = results.filter(r => r.status === 'pass').length;
    const failing = results.filter(r => r.status === 'fail').length;
    const running = results.filter(r => r.status === 'running').length;

    return (
        <div className="p-8 space-y-8">
            {/* Header */}
            <div className="flex items-start justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <FlaskConical className="h-8 w-8 text-cyan-500" />
                        Auto-Test Runner
                    </h1>
                    <p className="text-zinc-500 mt-2 max-w-2xl">
                        File-watcher driven test runner. Watch mode detects source changes and automatically re-runs related test files using the repo dependency graph.
                    </p>
                </div>
                <div className="flex gap-2 shrink-0">
                    <Button
                        variant="outline"
                        size="sm"
                        className="border-zinc-700 text-zinc-400 hover:text-white"
                        onClick={() => { statusQuery.refetch(); resultsQuery.refetch(); }}
                        disabled={statusQuery.isFetching}
                    >
                        {statusQuery.isFetching ? (
                            <Loader2 className="h-4 w-4 animate-spin mr-2" />
                        ) : (
                            <RefreshCw className="h-4 w-4 mr-2" />
                        )}
                        Refresh
                    </Button>
                    {normalized.isRunning ? (
                        <Button
                            onClick={() => stopMutation.mutate()}
                            disabled={stopMutation.isPending}
                            className="bg-red-700 hover:bg-red-600 text-white"
                        >
                            {stopMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Square className="h-4 w-4 mr-2" />}
                            Stop Watcher
                        </Button>
                    ) : (
                        <Button
                            onClick={() => startMutation.mutate()}
                            disabled={startMutation.isPending}
                            className="bg-cyan-700 hover:bg-cyan-600 text-white"
                        >
                            {startMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Play className="h-4 w-4 mr-2" />}
                            Start Watcher
                        </Button>
                    )}
                </div>
            </div>

            {/* Stats row */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                <Card className={`border ${normalized.isRunning ? 'border-cyan-800 bg-cyan-950/20' : 'border-zinc-800 bg-zinc-900'}`}>
                    <CardContent className="p-4">
                        <div className="text-xs text-zinc-500 uppercase tracking-widest mb-1">Watcher</div>
                        <div className={`text-lg font-bold ${normalized.isRunning ? 'text-cyan-400' : 'text-zinc-500'}`}>
                            {normalized.isRunning ? 'Active' : 'Stopped'}
                        </div>
                    </CardContent>
                </Card>
                <Card className="border-green-900/40 bg-zinc-900">
                    <CardContent className="p-4">
                        <div className="text-xs text-zinc-500 uppercase tracking-widest mb-1">Passing</div>
                        <div className="text-lg font-bold text-green-400">{passing}</div>
                    </CardContent>
                </Card>
                <Card className={`${failing > 0 ? 'border-red-900/50' : 'border-zinc-800'} bg-zinc-900`}>
                    <CardContent className="p-4">
                        <div className="text-xs text-zinc-500 uppercase tracking-widest mb-1">Failing</div>
                        <div className={`text-lg font-bold ${failing > 0 ? 'text-red-400' : 'text-zinc-500'}`}>{failing}</div>
                    </CardContent>
                </Card>
                <Card className={`${running > 0 ? 'border-yellow-900/50' : 'border-zinc-800'} bg-zinc-900`}>
                    <CardContent className="p-4">
                        <div className="text-xs text-zinc-500 uppercase tracking-widest mb-1">Running</div>
                        <div className={`text-lg font-bold ${running > 0 ? 'text-yellow-400' : 'text-zinc-500'}`}>{running}</div>
                    </CardContent>
                </Card>
            </div>

            {/* Results list */}
            <div className="space-y-4">
                <div className="flex items-center justify-between">
                    <h2 className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                        <Clock className="h-4 w-4" />
                        Recent Results
                        {results.length > 0 && (
                            <Badge variant="secondary" className="ml-1 text-xs">{results.length}</Badge>
                        )}
                    </h2>
                </div>

                {resultsQuery.isLoading ? (
                    <div className="flex justify-center p-12">
                        <Loader2 className="h-7 w-7 animate-spin text-zinc-500" />
                    </div>
                ) : results.length === 0 ? (
                    <div className="text-center p-12 text-zinc-600 text-sm border border-dashed border-zinc-800 rounded-lg">
                        <FlaskConical className="h-10 w-10 mx-auto mb-3 opacity-25" />
                        <p>No test results yet.</p>
                        <p className="mt-1">Start the watcher and save a source file to trigger an automatic run.</p>
                    </div>
                ) : (
                    <div className="space-y-2">
                        {/* Failures first */}
                        {[...results.filter(r => r.status === 'fail'), ...results.filter(r => r.status === 'running'), ...results.filter(r => r.status === 'pass'), ...results.filter(r => r.status !== 'fail' && r.status !== 'running' && r.status !== 'pass')].map((r, i) => (
                            <TestResultCard key={r.file || i} result={r} />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
