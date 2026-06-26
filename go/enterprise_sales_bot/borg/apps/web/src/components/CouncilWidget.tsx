"use client";
import React, { useState } from 'react';
import { trpc } from '../utils/trpc';
import { motion } from 'framer-motion';

interface TranscriptEntry {
    speaker: string;
    text: string;
}

function normalizeTranscripts(value: unknown): TranscriptEntry[] {
    if (typeof value !== 'object' || value === null) {
        return [];
    }

    const transcripts = (value as { transcripts?: unknown }).transcripts;
    if (!Array.isArray(transcripts)) {
        return [];
    }

    return transcripts
        .filter((item): item is { speaker: string; text: string } => {
            return (
                typeof item === 'object' &&
                item !== null &&
                typeof (item as { speaker?: unknown }).speaker === 'string' &&
                typeof (item as { text?: unknown }).text === 'string'
            );
        })
        .map((item) => ({ speaker: item.speaker, text: item.text }));
}

export const CouncilWidget: React.FC = () => {
    const [topic, setTopic] = useState('');
    const [isDebating, setIsDebating] = useState(false);

    const { data: latestSession, refetch } = trpc.council.getLatestSession.useQuery(undefined, {
        enabled: true,
        refetchInterval: isDebating ? 1000 : 5000
    });

    const runSessionMutation = trpc.council.runSession.useMutation({
        onSuccess: () => {
            setIsDebating(false);
            refetch();
        },
        onError: () => {
            setIsDebating(false);
        }
    });

    const handleStartDebate = () => {
        if (!topic) return;
        setIsDebating(true);
        runSessionMutation.mutate({ proposal: topic });
    };

    const transcriptEntries = normalizeTranscripts(latestSession);

    return (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 flex flex-col h-[500px]">
            <div className="flex justify-between items-center mb-4 border-b border-zinc-800 pb-2">
                <h2 className="text-lg font-bold text-white flex items-center gap-2">
                    🏛️ AI Council Consensus
                </h2>
                {isDebating && <span className="text-yellow-500 text-xs animate-pulse">Session in Progress...</span>}
            </div>

            <div className="flex-1 overflow-y-auto space-y-4 mb-4 pr-1 scrollbar-thin scrollbar-thumb-zinc-700">
                {!latestSession && !isDebating && (
                    <div className="text-center text-zinc-500 mt-20">
                        <p>No consensus session active.</p>
                        <p className="text-sm">Propose a topic to convene the Council.</p>
                    </div>
                )}

                {latestSession && (
                    <div className="space-y-4">
                        {transcriptEntries.map((entry, idx) => (
                            <motion.div
                                key={idx}
                                initial={{ opacity: 0, x: -10 }}
                                animate={{ opacity: 1, x: 0 }}
                                transition={{ delay: idx * 0.1 }}
                                className={`p-3 rounded-lg text-sm border ${entry.speaker === 'Product Manager' ? 'bg-blue-900/20 border-blue-800/50' :
                                    entry.speaker === 'The Architect' ? 'bg-purple-900/20 border-purple-800/50' :
                                        entry.speaker === 'The Critic' ? 'bg-red-900/20 border-red-800/50' :
                                            'bg-green-900/20 border-green-800/50 font-bold'
                                    }`}
                            >
                                <div className="font-bold text-xs mb-1 opacity-80 uppercase tracking-wider flex items-center gap-2">
                                    {entry.speaker === 'Product Manager' && '👔'}
                                    {entry.speaker === 'The Architect' && '📐'}
                                    {entry.speaker === 'The Critic' && '🛡️'}
                                    {entry.speaker}
                                </div>
                                <div className="text-zinc-300 leading-relaxed whitespace-pre-wrap">
                                    {entry.text}
                                </div>
                            </motion.div>
                        ))}
                    </div>
                )}
            </div>

            <div className="border-t border-zinc-800 pt-3">
                <div className="flex gap-2">
                    <input
                        type="text"
                        value={topic}
                        onChange={(e) => setTopic(e.target.value)}
                        placeholder="Enter a strategic topic for debate..."
                        className="flex-1 bg-black border border-zinc-700 rounded px-3 py-2 text-white text-sm focus:border-blue-500 focus:outline-none"
                        onKeyDown={(e) => e.key === 'Enter' && handleStartDebate()}
                        disabled={isDebating}
                    />
                    <button
                        onClick={handleStartDebate}
                        disabled={!topic || isDebating}
                        className="bg-blue-700 hover:bg-blue-600 disabled:opacity-50 text-white px-4 py-2 rounded text-sm font-semibold transition-colors"
                    >
                        {isDebating ? 'Convening...' : 'Convene'}
                    </button>
                </div>
            </div>
        </div>
    );
};
