"use client";

import { useState } from "react";
import { trpc } from "../utils/trpc";

export function AutonomyControl() {
    const utils = trpc.useUtils();

    // Queries
    const getAutonomyQuery = trpc.autonomy.getLevel.useQuery(undefined, {
        refetchInterval: 5000,
    });

    const setAutonomyMutation = trpc.autonomy.setLevel.useMutation({
        onSuccess: (data) => {
            utils.autonomy.getLevel.invalidate();
        }
    });

    const activateMutation = trpc.autonomy.activateFullAutonomy.useMutation({
        onSuccess: () => {
            utils.autonomy.getLevel.invalidate();
        }
    });

    // Use server data or default to 'low'
    // @ts-ignore
    const level = getAutonomyQuery.data || 'low';

    const handleSetLevel = (newLevel: 'low' | 'medium' | 'high') => {
        setAutonomyMutation.mutate({ level: newLevel });
    };

    return (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 shadow-xl">
            <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                    🤖 Autopilot Control
                </h2>
                <span className={`px-2 py-1 rounded text-xs font-bold uppercase ${level === 'high' ? 'bg-red-900 text-red-100 animate-pulse' :
                    level === 'medium' ? 'bg-yellow-900 text-yellow-100' :
                        'bg-green-900 text-green-100'
                    }`}>
                    {level} Autonomy
                </span>
            </div>

            <p className="text-zinc-400 text-sm mb-6">
                Control how much freedom the Director Agent has to execute tools without approval.
            </p>

            <div className="grid grid-cols-3 gap-4">
                <button
                    onClick={() => handleSetLevel('low')}
                    className={`p-4 rounded-lg border text-center transition-all ${level === 'low'
                        ? 'bg-green-500/20 border-green-500 text-green-200'
                        : 'bg-zinc-800 border-zinc-700 text-zinc-400 hover:bg-zinc-700'
                        }`}
                >
                    <div className="font-bold mb-1">Low</div>
                    <div className="text-xs opacity-70">Ask for almost everything. Safe.</div>
                </button>

                <button
                    onClick={() => handleSetLevel('medium')}
                    className={`p-4 rounded-lg border text-center transition-all ${level === 'medium'
                        ? 'bg-yellow-500/20 border-yellow-500 text-yellow-200'
                        : 'bg-zinc-800 border-zinc-700 text-zinc-400 hover:bg-zinc-700'
                        }`}
                >
                    <div className="font-bold mb-1">Medium</div>
                    <div className="text-xs opacity-70">Approve reads, ask for writes.</div>
                </button>

                <button
                    onClick={() => handleSetLevel('high')}
                    className={`p-4 rounded-lg border text-center transition-all ${level === 'high'
                        ? 'bg-red-500/20 border-red-500 text-red-200 shadow-[0_0_15px_rgba(239,68,68,0.5)]'
                        : 'bg-zinc-800 border-zinc-700 text-zinc-400 hover:bg-zinc-700'
                        }`}
                >
                    <div className="font-bold mb-1">High (Autopilot)</div>
                    <div className="text-xs opacity-70">Auto-approve EVERYTHING.</div>
                </button>
            </div>

            {level === 'high' && (
                <div className="mt-4 p-3 bg-red-950/50 border border-red-900/50 rounded text-red-200 text-xs flex gap-2 items-center animate-pulse">
                    <span className="text-lg">⚠️</span>
                    <strong>Supervisor Active:</strong> The agent will execute file writes and commands autonomously.
                </div>
            )}

            <div className="mt-6 border-t border-zinc-800 pt-4">
                <h3 className="text-sm font-bold text-white mb-2">🚀 One-Click Activation</h3>
                <button
                    onClick={() => activateMutation.mutate()}
                    disabled={activateMutation.isPending}
                    className="w-full py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 text-white font-bold rounded-lg shadow-lg flex items-center justify-center gap-2 transition-all disabled:opacity-50"
                >
                    {activateMutation.isPending ? "Activating..." : "Enable Intelligent Supervisor (Full Autonomy)"}
                </button>
                <p className="text-zinc-500 text-xs mt-2 text-center">
                    System will auto-accept tools, watch the IDE chat, and monitor process health.
                </p>
                {activateMutation.data && (
                    <div className="mt-2 text-green-400 text-xs text-center font-bold">
                        ✅ {activateMutation.data}
                    </div>
                )}
            </div>
        </div>
    );
}
