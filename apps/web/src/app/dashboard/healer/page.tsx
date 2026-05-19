'use client';

import { useHealerStream } from '@borg/ui';
import { trpc } from '@/utils/trpc';

export default function HealerDashboard() {
    const { events, isLoading } = useHealerStream();

    // @ts-ignore
    const { data: vaultRecords, isLoading: vaultLoading } = trpc.healer.vaultRecords.useQuery(undefined, {
        refetchOnWindowFocus: false,
        refetchInterval: 30000, // Refresh every 30s
    });

    // Derive active infections from failed heal attempts
    const history = events;
    const activeInfections = history.filter((e: any) => !e.success);
    const resolvedCount = history.filter((e: any) => e.success).length;
    const successRate = history.length > 0 ? Math.round((resolvedCount / history.length) * 100) : 100;
    const lastHealTime = history.length > 0 ? new Date(history[history.length - 1]?.timestamp).toLocaleString() : 'Never';

    return (
        <div className="p-8 bg-gray-900 min-h-screen text-gray-100 font-mono">
            <header className="mb-8 border-b border-gray-700 pb-4 flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-green-400">THE IMMUNE SYSTEM</h1>
                    <p className="text-gray-400">Self-Healing &amp; Auto-Correction</p>
                </div>
                <div className="text-right">
                    <div className="text-sm text-gray-500">STATUS</div>
                    <div className={activeInfections.length > 0 ? "text-red-500 font-bold" : "text-green-500 font-bold"}>
                        {activeInfections.length > 0 ? `${activeInfections.length} ACTIVE` : 'HEALTHY'}
                    </div>
                </div>
            </header>

            {/* Stats Row */}
            <div className="grid grid-cols-4 gap-4 mb-8">
                <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 text-center">
                    <div className="text-2xl font-bold text-white">{history.length}</div>
                    <div className="text-xs text-gray-400">Total Events</div>
                </div>
                <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 text-center">
                    <div className="text-2xl font-bold text-green-400">{resolvedCount}</div>
                    <div className="text-xs text-gray-400">Neutralized</div>
                </div>
                <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 text-center">
                    <div className="text-2xl font-bold text-yellow-400">{successRate}%</div>
                    <div className="text-xs text-gray-400">Success Rate</div>
                </div>
                <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 text-center">
                    <div className="text-2xl font-bold text-blue-400">{vaultRecords?.length || 0}</div>
                    <div className="text-xs text-gray-400">L2 Vault Records</div>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
                {/* Active Infections */}
                <section className="bg-gray-800 rounded-lg border border-gray-700 p-6">
                    <h2 className="text-xl font-bold mb-4 text-red-400 flex items-center">
                        <span className={activeInfections.length > 0 ? "animate-pulse mr-2" : "mr-2"}>●</span> ACTIVE INFECTIONS ({activeInfections.length})
                    </h2>
                    {activeInfections.length === 0 ? (
                        <div className="text-gray-500 italic text-center py-10">
                            No active pathogens detected. System healthy.
                        </div>
                    ) : (
                        <div className="space-y-4 max-h-[400px] overflow-y-auto">
                            {activeInfections.slice().reverse().map((inf: any, i: number) => (
                                <div key={i} className="bg-red-900/20 border border-red-500/50 p-4 rounded">
                                    <div className="flex justify-between text-xs text-gray-400 mb-2">
                                        <span>{new Date(inf.timestamp).toLocaleString()}</span>
                                        <span className="text-red-400 font-bold">UNRESOLVED</span>
                                    </div>
                                    <div className="text-sm font-semibold text-white mb-2">
                                        {(inf.error || 'Unknown error').slice(0, 150)}
                                    </div>
                                    {inf.fix?.diagnosis && (
                                        <div className="bg-black/50 p-2 rounded text-xs mt-2">
                                            <div className="text-yellow-300 mb-1">Type: {inf.fix.diagnosis.errorType}</div>
                                            <div className="text-blue-300">File: {inf.fix.diagnosis.file || 'unknown'}</div>
                                            <div className="text-gray-400 mt-1">{inf.fix.diagnosis.description}</div>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    )}
                </section>

                {/* Immune History */}
                <section className="bg-gray-800 rounded-lg border border-gray-700 p-6">
                    <h2 className="text-xl font-bold mb-4 text-blue-400">IMMUNE HISTORY ({resolvedCount})</h2>
                    <div className="space-y-4 max-h-[400px] overflow-y-auto">
                        {history.length === 0 ? (
                            <div className="text-gray-500 italic text-center py-10">
                                No history available.
                            </div>
                        ) : (
                            history.filter((e: any) => e.success).slice().reverse().map((entry: any, i: number) => (
                                <div key={i} className="bg-gray-700/50 border border-gray-600 p-4 rounded">
                                    <div className="flex justify-between text-xs text-gray-400 mb-2">
                                        <span>{new Date(entry.timestamp).toLocaleString()}</span>
                                        <span className="text-green-400">NEUTRALIZED</span>
                                    </div>
                                    <div className="text-sm font-semibold text-white mb-2">
                                        {(entry.error || '').slice(0, 100)}...
                                    </div>
                                    {entry.fix && (
                                        <div className="bg-black/50 p-2 rounded text-xs overflow-x-auto">
                                            <div className="text-blue-300 mb-1">Fixed File: {entry.fix.diagnosis?.file}</div>
                                            <pre className="text-green-300">{entry.fix.diagnosis?.suggestedFix}</pre>
                                        </div>
                                    )}
                                </div>
                            ))
                        )}
                    </div>
                </section>
            </div>

            {/* Persistent Memory L2 Vault */}
            <section className="bg-gray-800 rounded-lg border border-gray-700 p-6 mb-8">
                <h2 className="text-xl font-bold mb-4 text-purple-400">L2 VAULT: HEALING INTELLIGENCE ({vaultRecords?.length || 0})</h2>
                {vaultLoading ? (
                    <div className="text-gray-500 text-center py-10">Accessing Vault...</div>
                ) : (!vaultRecords || vaultRecords.length === 0) ? (
                    <div className="text-gray-500 italic text-center py-10">
                        No persistent healing records found in the L2 Vault.
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-h-[600px] overflow-y-auto">
                        {vaultRecords.map((record: any, i: number) => (
                            <div key={i} className="bg-gray-900 border border-purple-500/30 p-4 rounded text-xs">
                                <div className="flex justify-between text-gray-400 border-b border-gray-700 pb-2 mb-2">
                                    <span>ID: {record.id?.slice(0,8) || 'unknown'}</span>
                                    <span className="text-purple-300">Score: {record.heatScore || record.importance || 'N/A'}</span>
                                </div>
                                <div className="text-white font-mono break-words whitespace-pre-wrap">
                                    {record.content || record.description || record.text || 'Empty record'}
                                </div>
                                <div className="mt-3 text-gray-500 flex justify-between">
                                    <span>{new Date(record.createdAt || record.timestamp || Date.now()).toLocaleDateString()}</span>
                                    <span className="bg-purple-900/30 text-purple-300 px-2 py-0.5 rounded">
                                        {record.type || record.source || 'healer'}
                                    </span>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </section>
        </div>
    );
}
