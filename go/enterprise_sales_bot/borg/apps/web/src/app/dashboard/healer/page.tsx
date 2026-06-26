'use client';

import React, { useState } from 'react';
import { useHealerStream } from '@tormentnexus/ui';
import { trpc } from '@/utils/trpc';
import { 
    Shield, 
    Heart, 
    Zap, 
    Thermometer, 
    Database, 
    Sparkles, 
    AlertTriangle, 
    CheckCircle, 
    RefreshCw, 
    Layers, 
    Search,
    BookOpen
} from 'lucide-react';

export default function HealerDashboard() {
    const { events, isLoading: isStreamLoading } = useHealerStream();
    const [limit, setLimit] = useState(30);

    // Fetch persistent L2 Vault records via tRPC
    const { data: vaultRecords, isLoading: isVaultLoading, refetch: refetchVault } = trpc.healer.vaultRecords.useQuery(
        { limit },
        { refetchInterval: 5000 } // Keep in sync every 5s
    );

    // Derive active streams
    const history = events || [];
    const activeInfections = history.filter((e: any) => !e.success);
    const resolvedCount = history.filter((e: any) => e.success).length;
    const successRate = history.length > 0 ? Math.round((resolvedCount / history.length) * 100) : 100;
    const lastHealTime = history.length > 0 ? new Date(history[history.length - 1]?.timestamp).toLocaleString() : 'Never';

    // Group or filter vault records
    const normalizedVault = Array.isArray(vaultRecords) ? vaultRecords : [];

    return (
        <div className="p-8 bg-gray-950 min-h-screen text-gray-100 font-mono selection:bg-green-500 selection:text-black">
            {/* Header section with pulsating live radar */}
            <header className="mb-8 border-b border-gray-800 pb-6 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-green-500/10 rounded-lg border border-green-500/30">
                            <Shield className="w-8 h-8 text-green-400 animate-pulse" />
                        </div>
                        <div>
                            <h1 className="text-3xl font-extrabold tracking-tight bg-gradient-to-r from-green-400 via-emerald-500 to-teal-400 bg-clip-text text-transparent">
                                THE IMMUNE SYSTEM
                            </h1>
                            <p className="text-gray-400 text-xs mt-1">Autonomous Self-Healing, Auto-Correction &amp; Persistent L2 Vault</p>
                        </div>
                    </div>
                </div>
                <div className="flex items-center gap-4">
                    <button 
                        onClick={() => refetchVault()}
                        className="flex items-center gap-2 px-3 py-1.5 text-xs bg-gray-900 border border-gray-800 hover:border-gray-700 rounded text-gray-300 transition-all active:scale-95"
                    >
                        <RefreshCw className="w-3.5 h-3.5" />
                        Re-sync DB
                    </button>
                    <div className="text-right">
                        <div className="text-[10px] text-gray-500 tracking-widest font-bold">RADAR STATUS</div>
                        <div className="flex items-center gap-2 mt-1">
                            <span className={`h-2.5 w-2.5 rounded-full ${activeInfections.length > 0 ? 'bg-red-500 animate-ping' : 'bg-green-500 animate-pulse'}`} />
                            <span className={activeInfections.length > 0 ? "text-red-400 font-extrabold text-sm" : "text-green-400 font-extrabold text-sm"}>
                                {activeInfections.length > 0 ? `${activeInfections.length} ACTIVE PATHOGENS` : 'SECURE & HEALTHY'}
                            </span>
                        </div>
                    </div>
                </div>
            </header>

            {/* Stats Row */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
                <div className="bg-gray-900/60 backdrop-blur-md border border-gray-800 rounded-xl p-5 hover:border-gray-700 transition-all flex items-center justify-between">
                    <div>
                        <div className="text-xs text-gray-400 uppercase tracking-wider">Live Pathogens</div>
                        <div className="text-3xl font-black text-white mt-1">{activeInfections.length}</div>
                    </div>
                    <div className="p-3 bg-red-500/10 rounded-lg text-red-400 border border-red-500/20">
                        <AlertTriangle className="w-5 h-5" />
                    </div>
                </div>
                <div className="bg-gray-900/60 backdrop-blur-md border border-gray-800 rounded-xl p-5 hover:border-gray-700 transition-all flex items-center justify-between">
                    <div>
                        <div className="text-xs text-gray-400 uppercase tracking-wider">Auto-Neutralized</div>
                        <div className="text-3xl font-black text-green-400 mt-1">{resolvedCount}</div>
                    </div>
                    <div className="p-3 bg-green-500/10 rounded-lg text-green-400 border border-green-500/20">
                        <CheckCircle className="w-5 h-5" />
                    </div>
                </div>
                <div className="bg-gray-900/60 backdrop-blur-md border border-gray-800 rounded-xl p-5 hover:border-gray-700 transition-all flex items-center justify-between">
                    <div>
                        <div className="text-xs text-gray-400 uppercase tracking-wider">Immune Efficacy</div>
                        <div className="text-3xl font-black text-yellow-400 mt-1">{successRate}%</div>
                    </div>
                    <div className="p-3 bg-yellow-500/10 rounded-lg text-yellow-400 border border-yellow-500/20">
                        <Zap className="w-5 h-5" />
                    </div>
                </div>
                <div className="bg-gray-900/60 backdrop-blur-md border border-gray-800 rounded-xl p-5 hover:border-gray-700 transition-all flex items-center justify-between">
                    <div>
                        <div className="text-xs text-gray-400 uppercase tracking-wider">Vault Records</div>
                        <div className="text-3xl font-black text-blue-400 mt-1">{normalizedVault.length}</div>
                    </div>
                    <div className="p-3 bg-blue-500/10 rounded-lg text-blue-400 border border-blue-500/20">
                        <Database className="w-5 h-5" />
                    </div>
                </div>
            </div>

            {/* Three column visual panel: Active, Live History, and L2 Vault */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                
                {/* Column 1: Pulsing Pathogens */}
                <section className="bg-gray-900/40 border border-gray-800 rounded-xl p-6 flex flex-col h-[650px]">
                    <div className="flex justify-between items-center mb-4 border-b border-gray-800 pb-3">
                        <h2 className="text-sm font-black uppercase tracking-widest text-red-400 flex items-center gap-2">
                            <span className="h-2 w-2 rounded-full bg-red-500 animate-ping" />
                            Active Pathogens ({activeInfections.length})
                        </h2>
                    </div>
                    
                    <div className="flex-1 overflow-y-auto space-y-4 pr-1 scrollbar-thin">
                        {activeInfections.length === 0 ? (
                            <div className="h-full flex flex-col items-center justify-center text-center text-gray-500 p-6">
                                <Shield className="w-12 h-12 text-green-500/20 mb-3" />
                                <div className="text-xs font-bold text-gray-400">RADAR SWEEP CLEAN</div>
                                <div className="text-[10px] text-gray-600 mt-1">No active system infections detected.</div>
                            </div>
                        ) : (
                            activeInfections.slice().reverse().map((inf: any, i: number) => (
                                <div key={i} className="bg-red-950/10 border border-red-900/50 p-4 rounded-lg hover:border-red-800 transition-all">
                                    <div className="flex justify-between text-[10px] text-red-400 font-bold mb-2">
                                        <span>{new Date(inf.timestamp).toLocaleTimeString()}</span>
                                        <span>ATTEMPT #{inf.attempts || 1}</span>
                                    </div>
                                    <div className="text-xs text-white break-words mb-2 leading-relaxed">
                                        {inf.error}
                                    </div>
                                    {inf.fix?.diagnosis && (
                                        <div className="bg-black/40 border border-red-950 p-2.5 rounded text-[10px] space-y-1">
                                            <div className="text-yellow-400 font-semibold">Diagnosis: {inf.fix.diagnosis.errorType}</div>
                                            <div className="text-blue-400 truncate">Culprit: {inf.fix.diagnosis.file}</div>
                                            <div className="text-gray-400">{inf.fix.diagnosis.description}</div>
                                        </div>
                                    )}
                                </div>
                            ))
                        )}
                    </div>
                </section>

                {/* Column 2: Live Immune Operations Stream */}
                <section className="bg-gray-900/40 border border-gray-800 rounded-xl p-6 flex flex-col h-[650px]">
                    <div className="flex justify-between items-center mb-4 border-b border-gray-800 pb-3">
                        <h2 className="text-sm font-black uppercase tracking-widest text-emerald-400 flex items-center gap-2">
                            <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                            Live Healer Stream ({resolvedCount})
                        </h2>
                    </div>

                    <div className="flex-1 overflow-y-auto space-y-4 pr-1 scrollbar-thin">
                        {history.filter((e: any) => e.success).length === 0 ? (
                            <div className="h-full flex flex-col items-center justify-center text-center text-gray-500 p-6">
                                <Heart className="w-12 h-12 text-emerald-500/20 mb-3" />
                                <div className="text-xs font-bold text-gray-400">IMMUNE RADAR IDLE</div>
                                <div className="text-[10px] text-gray-600 mt-1">Awaiting real-time healing events.</div>
                            </div>
                        ) : (
                            history.filter((e: any) => e.success).slice().reverse().map((entry: any, i: number) => (
                                <div key={i} className="bg-emerald-950/10 border border-emerald-900/50 p-4 rounded-lg hover:border-emerald-800 transition-all">
                                    <div className="flex justify-between text-[10px] text-emerald-400 font-bold mb-2">
                                        <span>{new Date(entry.timestamp).toLocaleTimeString()}</span>
                                        <span className="flex items-center gap-1">NEUTRALIZED <CheckCircle className="w-3 h-3" /></span>
                                    </div>
                                    <div className="text-xs text-white truncate mb-2">
                                        {entry.error}
                                    </div>
                                    {entry.fix && (
                                        <div className="bg-black/40 border border-emerald-950 p-2.5 rounded text-[10px]">
                                            <div className="text-blue-400 mb-1 font-semibold">Healed: {entry.fix.diagnosis?.file?.split('/').pop()}</div>
                                            <pre className="text-emerald-300 font-mono text-[9px] overflow-x-auto max-h-20 max-w-full">
                                                {entry.fix.diagnosis?.suggestedFix}
                                            </pre>
                                        </div>
                                    )}
                                </div>
                            ))
                        )}
                    </div>
                </section>

                {/* Column 3: Persistent L2 Vector Vault */}
                <section className="bg-gray-900/40 border border-gray-800 rounded-xl p-6 flex flex-col h-[650px] lg:col-span-1">
                    <div className="flex justify-between items-center mb-4 border-b border-gray-800 pb-3">
                        <h2 className="text-sm font-black uppercase tracking-widest text-blue-400 flex items-center gap-2">
                            <Layers className="w-4 h-4 text-blue-400" />
                            SQLite L2 Vector Vault ({normalizedVault.length})
                        </h2>
                        <select 
                            value={limit} 
                            onChange={(e) => setLimit(Number(e.target.value))}
                            className="bg-black text-xs border border-gray-800 rounded px-1 text-gray-300"
                        >
                            <option value={10}>10 records</option>
                            <option value={30}>30 records</option>
                            <option value={50}>50 records</option>
                        </select>
                    </div>

                    <div className="flex-1 overflow-y-auto space-y-4 pr-1 scrollbar-thin">
                        {isVaultLoading ? (
                            <div className="h-full flex flex-col items-center justify-center text-center text-gray-500">
                                <RefreshCw className="w-8 h-8 text-blue-400 animate-spin mb-2" />
                                <span className="text-[10px] text-gray-400">Loading vector logs...</span>
                            </div>
                        ) : normalizedVault.length === 0 ? (
                            <div className="h-full flex flex-col items-center justify-center text-center text-gray-500 p-6">
                                <Database className="w-12 h-12 text-blue-500/20 mb-3" />
                                <div className="text-xs font-bold text-gray-400">VAULT IS EMPTY</div>
                                <div className="text-[10px] text-gray-600 mt-1">No long-term memories committed yet.</div>
                            </div>
                        ) : (
                            normalizedVault.map((record: any, i: number) => {
                                const importancePercent = Math.min(100, Math.round(record.Importance * 100));
                                const heatScore = Math.round(record.HeatScore || 50);
                                
                                return (
                                    <div key={i} className="bg-gray-950/60 border border-gray-800 hover:border-blue-900/60 p-4 rounded-lg transition-all group">
                                        {/* Header */}
                                        <div className="flex justify-between items-center text-[10px] mb-2">
                                            <div className="flex items-center gap-1.5">
                                                <span className="px-1.5 py-0.5 rounded bg-blue-500/10 border border-blue-500/20 text-blue-400 font-bold uppercase tracking-wider text-[8px]">
                                                    {record.Type || 'Vault'}
                                                </span>
                                                <span className="text-gray-500">
                                                    {new Date(record.CreatedAt || Date.now()).toLocaleTimeString()}
                                                </span>
                                            </div>
                                            <div className="flex items-center gap-2">
                                                <div className="flex items-center gap-0.5 text-orange-400" title={`Heat Score: ${heatScore}%`}>
                                                    <Thermometer className="w-3 h-3" />
                                                    <span className="font-bold">{heatScore}</span>
                                                </div>
                                            </div>
                                        </div>

                                        {/* Content body */}
                                        <div className="text-xs text-gray-300 leading-relaxed font-sans mb-3 select-all bg-black/35 p-2 rounded border border-gray-900/50 break-words group-hover:text-white transition-colors">
                                            {record.Content}
                                        </div>

                                        {/* Importance & Metadata Indicators */}
                                        <div className="space-y-2 border-t border-gray-900/50 pt-2.5">
                                            <div className="flex justify-between items-center text-[9px] text-gray-500">
                                                <span>Importance Weight:</span>
                                                <span className="font-bold text-gray-300">{importancePercent}%</span>
                                            </div>
                                            <div className="w-full bg-gray-900 h-1.5 rounded-full overflow-hidden border border-gray-800/40">
                                                <div 
                                                    className="h-full bg-gradient-to-r from-blue-500 via-indigo-500 to-purple-500 rounded-full transition-all duration-500" 
                                                    style={{ width: `${importancePercent}%` }} 
                                                />
                                            </div>
                                            {record.SessionID && (
                                                <div className="flex items-center justify-between text-[9px] text-gray-600 font-mono mt-1">
                                                    <span>Session Attached:</span>
                                                    <span className="text-blue-500/80 truncate max-w-[150px]">{record.SessionID}</span>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                );
                            })
                        )}
                    </div>
                </section>
            </div>
        </div>
    );
}
