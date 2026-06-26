"use client";
import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';

interface Stats {
    brain: {
        totalMemories: number;
        status: string;
    };
    ingestion: {
        lastBatch: string;
        status: string;
    };
}

export default function IngestionStatus() {
    const [stats, setStats] = useState<Stats | null>(null);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const res = await fetch('/api/monitoring/stats');
                if (res.ok) {
                    setStats(await res.json());
                }
            } catch (e) {
                console.error("Failed to fetch stats", e);
            }
        };

        fetchStats();
        const interval = setInterval(fetchStats, 5000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-zinc-900 to-black border border-zinc-800 p-4">
            {/* Background Glow */}
            <div className="absolute inset-0 opacity-10 bg-gradient-to-br from-blue-500 to-purple-500 blur-3xl" />

            <div className="relative z-10">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-bold text-white flex items-center gap-2">
                        <span className="text-2xl">📥</span>
                        Data Ingestion
                    </h2>
                </div>

                {!stats ? (
                    <div className="flex items-center gap-2 text-zinc-400">
                        <motion.div
                            animate={{ rotate: 360 }}
                            transition={{ repeat: Infinity, duration: 1, ease: "linear" }}
                            className="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full"
                        />
                        Loading stats...
                    </div>
                ) : (
                    <div className="grid grid-cols-2 gap-3">
                        <motion.div
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            className="relative overflow-hidden bg-gradient-to-br from-blue-500/20 to-blue-600/10 border border-blue-500/30 rounded-xl p-4 flex flex-col justify-center items-center"
                        >
                            <motion.div
                                animate={{ scale: [1, 1.05, 1] }}
                                transition={{ repeat: Infinity, duration: 2 }}
                                className="text-3xl font-bold text-blue-400"
                            >
                                {stats.brain.totalMemories}
                            </motion.div>
                            <div className="text-[10px] text-blue-300/70 uppercase tracking-widest mt-1">Memories</div>
                            <div className="absolute -top-4 -right-4 w-12 h-12 bg-blue-400/20 rounded-full blur-xl" />
                        </motion.div>

                        <motion.div
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            transition={{ delay: 0.1 }}
                            className="relative overflow-hidden bg-gradient-to-br from-purple-500/20 to-purple-600/10 border border-purple-500/30 rounded-xl p-4 flex flex-col justify-center items-center"
                        >
                            <div className={`text-lg font-bold ${stats.ingestion.status === 'idle' ? 'text-purple-400' : 'text-green-400'
                                }`}>
                                {stats.ingestion.status.toUpperCase()}
                            </div>
                            <div className="text-[10px] text-purple-300/70 uppercase tracking-widest mt-1">
                                {stats.ingestion.lastBatch || 'Batch Status'}
                            </div>
                            <div className="absolute -bottom-4 -left-4 w-12 h-12 bg-purple-400/20 rounded-full blur-xl" />
                        </motion.div>
                    </div>
                )}
            </div>
        </div>
    );
}
