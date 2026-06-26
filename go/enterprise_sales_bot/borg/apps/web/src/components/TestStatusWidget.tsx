"use client";
import React from 'react';
import { trpc } from '@/utils/trpc';

export function TestStatusWidget() {
    const utils = trpc.useUtils();
    const statusQuery = trpc.tests.status.useQuery(undefined, {
        refetchInterval: 4000,
        refetchOnWindowFocus: false,
    });

    const startMutation = trpc.tests.start.useMutation({
        onSuccess: () => utils.tests.status.invalidate(),
    });

    const stopMutation = trpc.tests.stop.useMutation({
        onSuccess: () => utils.tests.status.invalidate(),
    });

    const results = statusQuery.data?.results ?? {};
    const entries = Object.values(results);
    const passed = entries.filter((r: any) => r.status === 'pass').length;
    const failed = entries.filter((r: any) => r.status === 'fail').length;
    const running = entries.filter((r: any) => r.status === 'running').length;
    const isRunning = Boolean(statusQuery.data?.isRunning);

    return (
        <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-zinc-900 to-black border border-zinc-800 p-4 h-full flex flex-col">
            <div className="absolute inset-0 opacity-10 blur-3xl bg-green-500" />
            <div className="relative z-10 flex flex-col h-full">
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-bold text-white flex items-center gap-2">
                        <span className="text-2xl">🧪</span>
                        Auto-Test Watcher
                    </h3>
                </div>

                <div className="grid grid-cols-3 gap-2 mb-4">
                    <div className="bg-green-500/10 border border-green-500/30 rounded-lg p-2 text-center">
                        <div className="text-xl font-bold text-green-400">{statusQuery.isPending ? '…' : passed}</div>
                        <div className="text-[10px] text-green-300/70 uppercase">Passed</div>
                    </div>
                    <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-2 text-center">
                        <div className="text-xl font-bold text-red-400">{statusQuery.isPending ? '…' : failed}</div>
                        <div className="text-[10px] text-red-300/70 uppercase">Failed</div>
                    </div>
                    <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-2 text-center">
                        <div className="text-xl font-bold text-blue-400">{statusQuery.isPending ? '…' : running}</div>
                        <div className="text-[10px] text-blue-300/70 uppercase">Running</div>
                    </div>
                </div>

                <div className="flex-1 flex flex-col items-center justify-center text-zinc-600 gap-2">
                    {statusQuery.error ? (
                        <>
                            <span className="text-4xl mb-2">⚠️</span>
                            <p className="text-sm text-red-400">Unable to fetch test status</p>
                            <p className="text-[10px] text-red-300">{statusQuery.error.message}</p>
                        </>
                    ) : (
                        <>
                            <span className="text-4xl mb-2">🧪</span>
                            <p className="text-sm">AutoTest service is {isRunning ? 'running' : 'stopped'}</p>
                            <button
                                onClick={() => (isRunning ? stopMutation.mutate() : startMutation.mutate())}
                                disabled={startMutation.isPending || stopMutation.isPending}
                                className="text-xs px-3 py-1.5 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 disabled:opacity-60"
                            >
                                {startMutation.isPending || stopMutation.isPending
                                    ? 'Updating...'
                                    : isRunning
                                        ? 'Stop Watcher'
                                        : 'Start Watcher'}
                            </button>
                        </>
                    )}
                </div>

                <div className="mt-3 pt-3 border-t border-zinc-800 flex justify-between items-center">
                    <span className={`text-[10px] px-2 py-1 rounded-full ${isRunning ? 'bg-green-900/40 text-green-300' : 'bg-zinc-800 text-zinc-500'}`}>
                        {isRunning ? '● RUNNING' : '○ STOPPED'}
                    </span>
                    <span className="text-[10px] text-zinc-600">{entries.length} test results</span>
                </div>
            </div>
        </div>
    );
}
