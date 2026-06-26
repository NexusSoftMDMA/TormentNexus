'use client';

import React, { useState } from 'react';
import { trpc } from '@/utils/trpc';
import VoiceInput from './VoiceInput';

export default function CommandCenter() {
    const [command, setCommand] = useState('');
    const [lastResponse, setLastResponse] = useState<string | null>(null);

    const chatMutation = trpc.director.chat.useMutation({
        onSuccess: (result) => {
            setLastResponse(JSON.stringify(result, null, 2));
            setCommand('');
        }
    });

    const handleSubmit = (e?: React.FormEvent) => {
        e?.preventDefault();
        if (!command.trim()) return;

        chatMutation.mutate({ message: command });
    };

    const handleVoice = (text: string) => {
        setCommand(text);
        // Auto-submit voice commands for fluidity?
        // Let's delay slighty to allow user to verify, or just submit.
        // For "Jarvis" feel, auto-submit is better.
        chatMutation.mutate({ message: text });
    };

    return (
        <div className="w-full max-w-4xl mx-auto mb-8">
            <div className="bg-zinc-900/80 backdrop-blur border border-indigo-500/30 rounded-2xl p-4 shadow-2xl relative overflow-hidden group">
                {/* Glow Effect */}
                <div className="absolute inset-0 bg-gradient-to-r from-indigo-500/10 via-purple-500/10 to-blue-500/10 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />

                <form onSubmit={handleSubmit} className="relative z-10 flex flex-col md:flex-row items-center gap-3">
                    <VoiceInput onTranscript={handleVoice} isProcessing={chatMutation.isPending} />

                    <div className="flex-1 relative">
                        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-indigo-400 font-mono text-lg">{'>'}</span>
                        <input
                            type="text"
                            value={command}
                            onChange={(e) => setCommand(e.target.value)}
                            placeholder="Command the Director... (e.g. 'Start Squad on feature/login', 'Refactor MCPServer')"
                            className="w-full bg-black/40 border border-zinc-700/50 rounded-xl py-3 pl-10 pr-4 text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 font-mono transition-all"
                            disabled={chatMutation.isPending}
                        />
                        {chatMutation.isPending && (
                            <div className="absolute right-3 top-1/2 -translate-y-1/2">
                                <span className="animate-spin block w-4 h-4 border-2 border-indigo-500 border-t-transparent rounded-full" />
                            </div>
                        )}
                    </div>

                    <button
                        type="submit"
                        disabled={chatMutation.isPending || !command}
                        className="px-6 py-3 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl font-bold shadow-lg shadow-indigo-900/20 transition-all"
                    >
                        EXECUTE
                    </button>
                </form>

                {/* Response / Status Area */}
                {lastResponse && (
                    <div className="mt-4 p-4 bg-black/40 rounded-xl border border-zinc-800 font-mono text-xs text-zinc-300 max-h-40 overflow-y-auto">
                        <div className="text-zinc-500 mb-1">LAST EXECUTION RESULT:</div>
                        <pre className="whitespace-pre-wrap">{lastResponse}</pre>
                    </div>
                )}
            </div>
        </div>
    );
}
