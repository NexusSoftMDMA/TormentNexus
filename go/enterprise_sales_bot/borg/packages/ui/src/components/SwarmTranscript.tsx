
'use client';

import React, { useEffect, useRef, useState } from 'react';
import { trpc } from '../utils/trpc';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { ScrollArea } from './ui/scroll-area';
import { MessageSquare, Users, Zap, CheckCircle2 } from 'lucide-react';

interface SwarmEvent {
    event: string;
    data: any;
    timestamp: number;
}

export function SwarmTranscript() {
    const [events, setEvents] = useState<SwarmEvent[]>([]);
    const [lastEventId, setLastEventId] = useState<number | undefined>(undefined);
    const scrollRef = useRef<HTMLDivElement>(null);

    // Subscribe to swarm events
    trpc.agent.swarmEvents.useSubscription({ lastEventId }, {
        onData: (data: SwarmEvent) => {
            setEvents((prev) => {
                // Prevent duplicates if history replay overlaps
                if (prev.some(e => e.timestamp === data.timestamp && e.event === data.event)) {
                    return prev;
                }
                return [...prev, data].sort((a, b) => a.timestamp - b.timestamp);
            });
            setLastEventId(data.timestamp);
        },
        onError: (err: any) => {
            console.error('Swarm subscription error:', err);
            // Re-subscription happens automatically via react-query/trpc,
            // and using lastEventId will trigger history replay.
        }
    });

    // Auto-scroll to bottom
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [events]);

    const getEventIcon = (type: string) => {
        if (type.includes('turn')) return <Zap className="w-4 h-4 text-yellow-500" />;
        if (type.includes('consensus')) return <CheckCircle2 className="w-4 h-4 text-green-500" />;
        if (type.includes('signal')) return <Users className="w-4 h-4 text-blue-500" />;
        return <MessageSquare className="w-4 h-4 text-zinc-500" />;
    };

    return (
        <Card className="h-full flex flex-col bg-zinc-50 dark:bg-black border-zinc-200 dark:border-zinc-800 shadow-xl overflow-hidden">
            <CardHeader className="border-b border-zinc-200 dark:border-zinc-800 py-3">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <Users className="w-4 h-4" />
                    Neural Swarm Transcript
                </CardTitle>
            </CardHeader>
            <CardContent className="flex-1 p-0 overflow-hidden relative">
                <ScrollArea className="h-full p-4">
                    <div className="space-y-4">
                        {events.length === 0 && (
                            <div className="flex flex-col items-center justify-center h-40 text-zinc-500 gap-2 opacity-50">
                                <Zap className="w-8 h-8 animate-pulse" />
                                <p className="text-xs">Awaiting swarm signals...</p>
                            </div>
                        )}

                        {events.map((ev, i) => (
                            <div key={i} className="flex flex-col gap-1 animate-in fade-in slide-in-from-bottom-2 duration-300">
                                <div className="flex items-center gap-2">
                                    {getEventIcon(ev.event)}
                                    <span className="text-[10px] font-mono text-zinc-500">
                                        {new Date(ev.timestamp).toLocaleTimeString()}
                                    </span>
                                    <Badge variant="outline" className="text-[10px] uppercase py-0 px-1 border-zinc-200 dark:border-zinc-800">
                                        {ev.event.replace('swarm:', '')}
                                    </Badge>
                                </div>
                                <div className="pl-6">
                                    <div className="text-sm text-zinc-800 dark:text-zinc-200 bg-white dark:bg-zinc-900 p-3 rounded-lg border border-zinc-200 dark:border-zinc-800 shadow-sm">
                                        {typeof ev.data === 'string' ? ev.data : (
                                            <pre className="text-xs font-mono overflow-x-auto">
                                                {JSON.stringify(ev.data, null, 2)}
                                            </pre>
                                        )}
                                    </div>
                                </div>
                            </div>
                        ))}
                        <div ref={scrollRef} />
                    </div>
                </ScrollArea>

                <div className="absolute bottom-0 left-0 right-0 h-8 bg-gradient-to-t from-zinc-50 dark:from-black to-transparent pointer-events-none" />
            </CardContent>
        </Card>
    );
}
