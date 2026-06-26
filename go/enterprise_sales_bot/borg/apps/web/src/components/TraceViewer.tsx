"use client";

import { useEffect, useRef, useState } from 'react';
import { trpc } from '@/utils/trpc';

type TraceLog = {
    timestamp?: number;
    level?: string;
    action?: string;
    message?: string;
    agentId?: string;
};

function normalizeTraceLogs(value: unknown): TraceLog[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value.filter((item): item is TraceLog => typeof item === 'object' && item !== null);
}

export function TraceViewer() {
    const [autoScroll, setAutoScroll] = useState(true);
    const bottomRef = useRef<HTMLDivElement>(null);

    const { data: rawLogs, isLoading } = trpc.audit.list.useQuery(
        { limit: 100 },
        { refetchInterval: 3000 } // Poll every 3s for live feel
    );
    const logs = normalizeTraceLogs(rawLogs);

    useEffect(() => {
        if (autoScroll && bottomRef.current) {
            bottomRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [logs, autoScroll]);

    return (
        <div className="p-6 bg-[#1e1e1e] rounded-xl border border-[#333] shadow-lg flex flex-col h-[500px]">
            <div className="flex justify-between items-center mb-4">
                <div>
                    <h2 className="text-xl font-bold text-white mb-1">Supervisor Trace</h2>
                    <p className="text-gray-400 text-sm">
                        {logs ? `${logs.length} audit entries` : 'Live loop activity and autonomous decisions'}
                    </p>
                </div>
                <div className="flex gap-2">
                    <button
                        onClick={() => setAutoScroll(!autoScroll)}
                        className={`px-3 py-1 text-sm rounded transition-colors ${autoScroll
                            ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20 hover:bg-blue-500/20'
                            : 'bg-[#333] text-gray-400 border border-[#444] hover:bg-[#444]'
                            }`}
                    >
                        {autoScroll ? '⬇ Locked' : '✋ Manual'}
                    </button>
                </div>
            </div>

            <div className="flex-1 bg-[#111] rounded p-4 overflow-y-auto font-mono text-sm text-gray-300 whitespace-pre-wrap">
                {isLoading ? (
                    <div className="flex flex-col items-center justify-center h-full text-gray-500">
                        <p className="text-sm">Loading audit logs...</p>
                    </div>
                ) : logs && logs.length > 0 ? (
                    <div className="space-y-1">
                        {logs.map((log, i: number) => (
                            <div key={i} className="flex gap-2">
                                <span className="text-gray-600 shrink-0">
                                    {log.timestamp ? new Date(log.timestamp).toLocaleTimeString() : '??:??:??'}
                                </span>
                                <span className={`shrink-0 ${log.level === 'error' ? 'text-red-400' :
                                        log.level === 'warn' ? 'text-yellow-400' :
                                            log.level === 'info' ? 'text-blue-400' : 'text-gray-400'
                                    }`}>
                                    [{log.level || 'info'}]
                                </span>
                                <span className="text-purple-400 shrink-0">{log.agentId || 'system'}</span>
                                <span className="text-gray-300">{log.action || log.message || JSON.stringify(log)}</span>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="flex flex-col items-center justify-center h-full text-gray-500">
                        <p className="text-lg font-medium">No Audit Entries</p>
                        <p className="text-sm mt-1">Traces will appear here when agents perform actions.</p>
                    </div>
                )}
                <div ref={bottomRef} />
            </div>
        </div>
    );
}
