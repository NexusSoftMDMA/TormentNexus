'use client';

import React, { useState, useEffect, useRef } from 'react';
import { trpc } from '../utils/trpc';
import { motion, AnimatePresence } from 'framer-motion';

export default function SuggestionsPanel() {
    const utils = trpc.useContext();
    // @ts-ignore
    const suggestionsQuery = trpc.suggestions.list.useQuery(undefined, {
        refetchInterval: 2000 // Poll every 2s
    });

    const [isMuted, setIsMuted] = useState(false);
    const lastSpeakId = useRef<string | null>(null);

    // TTS Logic
    useEffect(() => {
        if (!suggestionsQuery.data || suggestionsQuery.data.length === 0 || isMuted) return;

        // Get most recent suggestion
        const latest = suggestionsQuery.data[0];

        // Speak if new
        if (latest.id !== lastSpeakId.current) {
            lastSpeakId.current = latest.id;
            const text = `I have a suggestion: ${latest.title}`;

            // Invoke Browser TTS
            if ('speechSynthesis' in window) {
                const utterance = new SpeechSynthesisUtterance(text);
                // Try to find a good voice
                const voices = window.speechSynthesis.getVoices();
                const preferred = voices.find(v => v.name.includes('Google US English') || v.name.includes('Microsoft David'));
                if (preferred) utterance.voice = preferred;

                window.speechSynthesis.speak(utterance);
            }
        }
    }, [suggestionsQuery.data, isMuted]);

    // @ts-ignore
    const resolveMutation = trpc.suggestions.resolve.useMutation({
        onSuccess: () => {
            // @ts-ignore
            utils.suggestions.list.invalidate();
        }
    });

    // @ts-ignore
    const clearAllMutation = trpc.suggestions.clearAll.useMutation({
        onSuccess: () => {
            // @ts-ignore
            utils.suggestions.list.invalidate();
        }
    });

    if (!suggestionsQuery.data || suggestionsQuery.data.length === 0) {
        return null; // Hidden if empty
    }

    return (
        <div className="w-full max-w-4xl mx-auto mb-8 relative">
            <div className="flex items-center justify-between mb-3 px-1">
                <h3 className="text-sm font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
                    Pending Authorizations
                </h3>
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => setIsMuted(!isMuted)}
                        className={`text-xs uppercase font-bold tracking-wider transition-colors ${isMuted ? 'text-red-500/80 hover:text-red-400' : 'text-zinc-600 hover:text-zinc-400'}`}
                    >
                        {isMuted ? 'Muted 🔇' : 'Voice On 🗣️'}
                    </button>
                    <button
                        onClick={() => clearAllMutation.mutate()}
                        className="text-xs text-zinc-600 hover:text-zinc-400 transition-colors uppercase font-bold tracking-wider"
                    >
                        Clear All
                    </button>
                </div>
            </div>

            <div className="space-y-3">
                <AnimatePresence>
                    {suggestionsQuery.data.map((s: any) => (
                        <motion.div
                            key={s.id}
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, x: -20 }}
                            className="bg-zinc-900/60 border border-amber-500/30 rounded-xl p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 backdrop-blur hover:bg-zinc-900/80 transition-colors"
                        >
                            <div className="flex-1">
                                <div className="flex items-center gap-2 mb-1">
                                    <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-zinc-800 text-zinc-400 border border-zinc-700">
                                        {s.source}
                                    </span>
                                    <span className="text-xs text-zinc-500">
                                        {new Date(s.timestamp).toLocaleTimeString()}
                                    </span>
                                </div>
                                <h4 className="text-zinc-200 font-semibold">{s.title}</h4>
                                <p className="text-zinc-400 text-sm mt-1">{s.description}</p>
                                {s.payload?.tool && (
                                    <div className="mt-2 p-2 bg-black/40 rounded border border-zinc-800 font-mono text-xs text-amber-200/80">
                                        $ {s.payload.tool} {JSON.stringify(s.payload.args)}
                                    </div>
                                )}
                            </div>

                            <div className="flex items-center gap-2 w-full sm:w-auto">
                                <button
                                    onClick={() => resolveMutation.mutate({ id: s.id, status: 'REJECTED' })}
                                    className="flex-1 sm:flex-none px-4 py-2 bg-zinc-800 hover:bg-red-900/30 hover:text-red-400 text-zinc-400 rounded-lg text-sm font-medium transition-colors border border-transparent hover:border-red-500/30"
                                >
                                    Dismiss
                                </button>
                                <button
                                    onClick={() => resolveMutation.mutate({ id: s.id, status: 'APPROVED' })}
                                    className="flex-1 sm:flex-none px-4 py-2 bg-amber-600 hover:bg-amber-500 text-white rounded-lg text-sm font-medium shadow-lg shadow-amber-900/20 transition-all border border-amber-500/50"
                                >
                                    Approve
                                </button>
                            </div>
                        </motion.div>
                    ))}
                </AnimatePresence>
            </div>
        </div>
    );
}
