"use client";

import React, { useEffect, useState, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { ScrollArea } from "./ui/scroll-area";
import { Bot, User, BrainCircuit, ShieldAlert, Cpu } from "lucide-react";
import { resolveCouncilWsUrl } from '../lib/endpoints';
import { createReconnectPolicy, getReconnectDelayMs, shouldRetryReconnect } from '../lib/connection-policy';

interface Transcript {
    speaker: string;
    text: string;
}

export function CouncilDebateWidget() {
    const [topic, setTopic] = useState<string | null>(null);
    const [transcripts, setTranscripts] = useState<Transcript[]>([]);
    const [activeSpeaker, setActiveSpeaker] = useState<string | null>(null);
    const scrollRef = useRef<HTMLDivElement>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectAttemptsRef = useRef(0);
    const reconnectTimerRef = useRef<number | null>(null);
    const reconnectPolicy = createReconnectPolicy();

    useEffect(() => {
        const wsUrl = resolveCouncilWsUrl(process.env.NEXT_PUBLIC_COUNCIL_WS_URL);

        const connect = () => {
            const ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                reconnectAttemptsRef.current = 0;
            };

            ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);

                    // Handle COUNCIL events (Request/Response format might differ based on ADK)
                    // If broadcastRequest uses a specific envelope, handle it.
                    // Assuming direct payload for now or wrapped in 'type'.

                    if (data.type === 'COUNCIL_START') {
                        setTopic(data.topic);
                        setTranscripts([]);
                        setActiveSpeaker(null);
                    }
                    else if (data.type === 'COUNCIL_THINKING') {
                        setActiveSpeaker(data.speaker);
                    }
                    else if (data.type === 'COUNCIL_TRANSCRIPT') {
                        setTranscripts(prev => [...prev, { speaker: data.speaker, text: data.text }]);
                        setActiveSpeaker(null);
                    }
                    else if (data.type === 'COUNCIL_END') {
                        setActiveSpeaker(null);
                    }

                } catch (e) {
                    console.error("Council Widget WS Error:", e);
                }
            };

            ws.onerror = () => {
                ws.close();
            };

            ws.onclose = () => {
                wsRef.current = null;
                if (shouldRetryReconnect(reconnectAttemptsRef.current, reconnectPolicy)) {
                    reconnectAttemptsRef.current += 1;
                    const delayMs = getReconnectDelayMs(reconnectAttemptsRef.current, reconnectPolicy);
                    reconnectTimerRef.current = window.setTimeout(connect, delayMs);
                }
            };

            wsRef.current = ws;
        };

        connect();

        return () => {
            if (reconnectTimerRef.current !== null) {
                window.clearTimeout(reconnectTimerRef.current);
            }
            wsRef.current?.close();
        };
    }, []);

    // Auto-scroll
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [transcripts, activeSpeaker]);

    const getIcon = (speaker: string) => {
        if (speaker.includes("Architect")) return <Cpu className="w-5 h-5 text-blue-400" />;
        if (speaker.includes("Critic")) return <ShieldAlert className="w-5 h-5 text-red-400" />;
        if (speaker.includes("Product")) return <User className="w-5 h-5 text-green-400" />; // Or Robot
        return <Bot className="w-5 h-5 text-zinc-400" />;
    };

    return (
        <Card className="h-full bg-zinc-900 border-zinc-800 flex flex-col">
            <CardHeader className="py-3 px-4 border-b border-zinc-800 bg-zinc-950/50">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-medium text-purple-400 flex items-center gap-2">
                        <BrainCircuit className="w-4 h-4" />
                        THE COUNCIL
                    </CardTitle>
                    {topic && (
                        <Badge variant="outline" className="text-xs max-w-[200px] truncate bg-purple-900/20 text-purple-300 border-purple-800">
                            {topic}
                        </Badge>
                    )}
                </div>
            </CardHeader>
            <CardContent className="flex-1 p-0 overflow-hidden relative">
                <ScrollArea className="h-full p-4" ref={scrollRef}>
                    <div className="space-y-4">
                        {transcripts.length === 0 && !topic && (
                            <div className="flex flex-col items-center justify-center h-full text-zinc-600 mt-10">
                                <BrainCircuit className="w-12 h-12 mb-2 opacity-20" />
                                <p className="text-sm">Council in session standby...</p>
                            </div>
                        )}

                        {transcripts.map((t, i) => (
                            <div key={i} className="flex gap-3 animate-in fade-in slide-in-from-bottom-2 duration-300">
                                <div className="mt-1 flex-shrink-0">
                                    {getIcon(t.speaker)}
                                </div>
                                <div className="bg-zinc-800/50 rounded-lg p-3 text-sm text-zinc-300">
                                    <span className="font-bold text-zinc-100 block mb-1 text-xs uppercase tracking-wider opacity-70">
                                        {t.speaker}
                                    </span>
                                    {t.text}
                                </div>
                            </div>
                        ))}

                        {activeSpeaker && (
                            <div className="flex gap-3 animate-pulse">
                                <div className="mt-1 flex-shrink-0">
                                    {getIcon(activeSpeaker)}
                                </div>
                                <div className="bg-zinc-800/20 rounded-lg p-3 text-sm text-zinc-500 italic">
                                    {activeSpeaker} is thinking...
                                </div>
                            </div>
                        )}

                        {/* Spacer for scroll */}
                        <div className="h-4" />
                    </div>
                </ScrollArea>
            </CardContent>
        </Card>
    );
}
