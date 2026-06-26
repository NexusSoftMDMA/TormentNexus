"use client";

import React, { useEffect, useRef, useState } from 'react';
import { trpc } from '../utils/trpc';

export function AuditLogViewer() {
    // @ts-ignore
    const { data: logs, refetch } = trpc.audit.query.useQuery({ limit: 100 }, {
        refetchInterval: 5000
    });

    // Auto-scroll logic could go here, but simple list is fine for now

    return (
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 h-96 flex flex-col">
            <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-mono font-semibold text-blue-400">🛡️ Audit Log</h3>
                <span className="text-xs text-gray-500">Auto-refresh: 5s</span>
            </div>

            <div className="flex-1 overflow-y-auto space-y-1 font-mono text-xs pr-2">
                {logs && logs.length > 0 ? (
                    logs.map((log: any, i: number) => (
                        <div key={i} className="flex gap-2 border-b border-gray-800/50 pb-1 last:border-0 hover:bg-gray-800/30 px-1 rounded">
                            <span className="text-gray-500 shrink-0 w-32">{new Date(log.timestamp).toLocaleTimeString()}</span>
                            <span className={`shrink-0 w-12 font-bold ${log.level === 'ERROR' ? 'text-red-500' :
                                log.level === 'WARN' ? 'text-yellow-500' :
                                    'text-blue-300'
                                }`}>{log.level}</span>
                            <span className="text-gray-300 break-all">
                                <span className="text-purple-400 font-semibold">{log.event}</span>
                                {log.details && (
                                    <span className="text-gray-500 ml-2">
                                        {JSON.stringify(log.details).substring(0, 100)}
                                        {JSON.stringify(log.details).length > 100 && "..."}
                                    </span>
                                )}
                            </span>
                        </div>
                    ))
                ) : (
                    <div className="text-gray-600 italic">No audit events recorded. System is quiet.</div>
                )}
            </div>
        </div>
    );
}
