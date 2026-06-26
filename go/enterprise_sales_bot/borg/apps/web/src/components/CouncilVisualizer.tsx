'use client';

import { useState, useEffect } from 'react';
import { trpc } from '@/utils/trpc';

export default function CouncilVisualizer() {
    const [session, setSession] = useState<any>(null);

    // Poll for latest session
    const { data, refetch } = trpc.council.getLatestSession.useQuery(undefined, {
        refetchInterval: 5000
    });

    useEffect(() => {
        if (data) setSession(data);
    }, [data]);

    if (!session) {
        return (
            <div className="p-6 bg-slate-900/50 rounded-xl border border-slate-800 text-center">
                <h3 className="text-xl font-bold text-slate-300 mb-2">🏛️ The Council is Adjourned</h3>
                <p className="text-slate-500">Waiting for a Consensus Session to convene...</p>
            </div>
        );
    }

    return (
        <div className="w-full bg-slate-950 rounded-xl border border-indigo-900/50 overflow-hidden shadow-2xl">
            {/* Header */}
            <div className="bg-gradient-to-r from-indigo-900 to-slate-900 p-4 border-b border-indigo-800/50 flex justify-between items-center">
                <div>
                    <h2 className="text-lg font-bold text-white flex items-center gap-2">
                        <span>🏛️</span> Council Chamber
                    </h2>
                    <p className="text-xs text-indigo-300 ml-7">Consensus Protocol Active</p>
                </div>
                <div className="flex gap-2">
                    {['Product Manager', 'The Architect', 'The Critic'].map(role => (
                        <div key={role} className={`px-2 py-1 rounded-full text-xs font-mono border ${role.includes('Market') ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300' :
                                role.includes('Architect') ? 'border-sky-500/30 bg-sky-500/10 text-sky-300' :
                                    'border-rose-500/30 bg-rose-500/10 text-rose-300'
                            }`}>
                            {role.split(' ')[1] || role}
                        </div>
                    ))}
                </div>
            </div>

            {/* Transcript */}
            <div className="p-6 space-y-6 max-h-[500px] overflow-y-auto">
                {session.transcripts.map((t: any, i: number) => {
                    const isDirective = t.speaker === 'Final Directive';
                    const color = t.speaker.includes('Product') ? 'text-emerald-400 border-emerald-500/20 bg-emerald-950/30' :
                        t.speaker.includes('Architect') ? 'text-sky-400 border-sky-500/20 bg-sky-950/30' :
                            t.speaker.includes('Critic') ? 'text-rose-400 border-rose-500/20 bg-rose-950/30' :
                                'text-amber-400 border-amber-500/20 bg-amber-950/30 font-bold'; // Directive

                    return (
                        <div key={i} className={`flex flex-col gap-1 ${isDirective ? 'items-center mt-8' : 'items-start'}`}>
                            <span className={`text-xs font-bold uppercase tracking-wider opacity-70 ml-1 mb-1 ${t.speaker.includes('Product') ? 'text-emerald-500' :
                                    t.speaker.includes('Architect') ? 'text-sky-500' :
                                        t.speaker.includes('Critic') ? 'text-rose-500' : 'text-amber-500'
                                }`}>
                                {t.speaker}
                            </span>
                            <div className={`p-4 rounded-xl border ${color} ${isDirective ? 'w-full text-center text-lg' : 'max-w-[90%]'}`}>
                                {t.text}
                            </div>
                        </div>
                    );
                })}
            </div>

            {/* Footer Status */}
            <div className="bg-slate-900/50 p-3 text-center border-t border-slate-800">
                <span className="text-xs text-slate-500">Session Status: {session.approved ? '✅ CONSENSUS REACHED' : '❌ ADJOURNED'}</span>
            </div>
        </div>
    );
}
