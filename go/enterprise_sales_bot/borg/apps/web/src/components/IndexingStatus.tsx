"use client";
import { trpc } from "../utils/trpc";
import { motion } from 'framer-motion';

type IndexingStatusData = {
    status: string;
    filesIndexed: number;
    totalFiles: number;
};

function normalizeIndexingStatus(value: unknown): IndexingStatusData | null {
    if (typeof value !== 'object' || value === null) {
        return null;
    }

    const status = (value as { status?: unknown }).status;
    const filesIndexed = (value as { filesIndexed?: unknown }).filesIndexed;
    const totalFiles = (value as { totalFiles?: unknown }).totalFiles;

    return {
        status: typeof status === 'string' ? status : 'loading',
        filesIndexed: typeof filesIndexed === 'number' ? filesIndexed : 0,
        totalFiles: typeof totalFiles === 'number' ? totalFiles : 0,
    };
}

export default function IndexingStatus() {
    const status = trpc.indexingStatus.useQuery(undefined, { refetchInterval: 3000 });
    const statusData = normalizeIndexingStatus(status.data);

    const progress = statusData
        ? Math.round((statusData.filesIndexed / Math.max(statusData.totalFiles || 1, 1)) * 100)
        : 0;

    const isComplete = progress >= 100;

    return (
        <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-zinc-900 to-black border border-zinc-800 p-4">
            {/* Background Glow */}
            <div className={`absolute inset-0 opacity-10 ${isComplete ? 'bg-cyan-500' : 'bg-purple-500'} blur-3xl`} />

            <div className="relative z-10">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-bold text-white flex items-center gap-2">
                        <span className="text-2xl">🧠</span>
                        Deep Code Intelligence
                    </h2>
                    <span className={`px-2 py-1 rounded-full text-xs font-bold ${isComplete ? 'bg-cyan-500/20 text-cyan-400' : 'bg-purple-500/20 text-purple-400'
                        }`}>
                        {statusData?.status?.toUpperCase() || 'LOADING'}
                    </span>
                </div>

                {!statusData ? (
                    <div className="flex items-center gap-2 text-zinc-400">
                        <motion.div
                            animate={{ rotate: 360 }}
                            transition={{ repeat: Infinity, duration: 1, ease: "linear" }}
                            className="w-4 h-4 border-2 border-purple-500 border-t-transparent rounded-full"
                        />
                        Connecting to Indexer...
                    </div>
                ) : (
                    <div className="space-y-4">
                        {/* Progress Bar */}
                        <div className="relative">
                            <div className="h-3 bg-zinc-800 rounded-full overflow-hidden">
                                <motion.div
                                    initial={{ width: 0 }}
                                    animate={{ width: `${progress}%` }}
                                    transition={{ duration: 0.5 }}
                                    className={`h-full ${isComplete
                                        ? 'bg-gradient-to-r from-cyan-500 to-blue-500'
                                        : 'bg-gradient-to-r from-purple-500 to-pink-500'
                                        }`}
                                />
                            </div>
                            <div className="absolute inset-0 flex items-center justify-center">
                                <span className="text-[10px] font-bold text-white drop-shadow">{progress}%</span>
                            </div>
                        </div>

                        {/* Stats Grid */}
                        <div className="grid grid-cols-2 gap-3">
                            <div className="p-3 bg-zinc-800/50 rounded-lg text-center">
                                <p className="text-2xl font-bold text-white">{statusData.filesIndexed}</p>
                                <p className="text-[10px] text-zinc-500 uppercase">Files Indexed</p>
                            </div>
                            <div className="p-3 bg-zinc-800/50 rounded-lg text-center">
                                <p className="text-2xl font-bold text-white">{statusData.totalFiles || '—'}</p>
                                <p className="text-[10px] text-zinc-500 uppercase">Total Files</p>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
