"use client";

import { trpc } from '@/utils/trpc';
import { Card, CardHeader, CardTitle, CardContent, ScrollArea, Badge } from '@borg/ui';
import { MessageSquare, Cpu, Brain, Activity } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

export function SwarmTranscript() {
    const { data: transcript, isLoading } = trpc.agent.getSwarmTranscript.useQuery(undefined, {
        refetchInterval: 3000
    });

    if (isLoading) return <div className="p-8 text-center text-zinc-500 italic">Connecting to Swarm Neural Bridge...</div>;

    return (
        <Card className="bg-zinc-950 border-zinc-800 h-full flex flex-col shadow-2xl">
            <CardHeader className="flex flex-row items-center justify-between border-b border-zinc-800 pb-4">
                <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                    <Activity className="h-4 w-4 text-cyan-400" />
                    Neural Transcript (Live)
                </CardTitle>
                <Badge variant="outline" className="border-cyan-500/20 text-cyan-400 bg-cyan-500/5">
                    {transcript?.length || 0} Turns
                </Badge>
            </CardHeader>
            <CardContent className="flex-1 min-h-0 p-0">
                <ScrollArea className="h-[600px] p-6">
                    <div className="space-y-6">
                        <AnimatePresence initial={false}>
                            {transcript && transcript.length > 0 ? (
                                transcript.map((entry: string, i: number) => {
                                    const [role, ...contentParts] = entry.split(':');
                                    const content = contentParts.join(':').trim();
                                    const isSystem = role === 'Collective Goal';
                                    
                                    return (
                                        <motion.div 
                                            key={i}
                                            initial={{ opacity: 0, x: -10 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            className={`flex flex-col gap-2 ${isSystem ? 'border-b border-zinc-800 pb-4 mb-4' : ''}`}
                                        >
                                            <div className="flex items-center gap-2">
                                                <span className={`text-[10px] font-bold uppercase tracking-tighter ${
                                                    role.includes('PLANNER') ? 'text-indigo-400' :
                                                    role.includes('IMPLEMENTER') ? 'text-amber-400' :
                                                    role.includes('TESTER') ? 'text-emerald-400' :
                                                    role.includes('CRITIC') ? 'text-rose-400' :
                                                    'text-zinc-500'
                                                }`}>
                                                    {role}
                                                </span>
                                            </div>
                                            <div className={`text-sm leading-relaxed ${isSystem ? 'text-zinc-400 font-medium italic' : 'text-zinc-300'}`}>
                                                {content || entry}
                                            </div>
                                        </motion.div>
                                    );
                                })
                            ) : (
                                <div className="h-full flex flex-col items-center justify-center text-zinc-600 py-20">
                                    <Brain className="h-12 w-12 mb-4 opacity-10" />
                                    <p className="text-sm italic">Neural bridge silent. Start a swarm session to begin.</p>
                                </div>
                            )}
                        </AnimatePresence>
                    </div>
                </ScrollArea>
            </CardContent>
        </Card>
    );
}
