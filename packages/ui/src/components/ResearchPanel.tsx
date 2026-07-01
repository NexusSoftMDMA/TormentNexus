
'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Card } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { ScrollArea } from './ui/scroll-area';
import { resolveCoreSseUrl } from '../lib/endpoints';
import { createReconnectPolicy, getReconnectDelayMs, normalizeNumericInput, shouldRetryReconnect } from '../lib/connection-policy';

export default function ResearchPanel() {
    const [topic, setTopic] = useState('');
    const [depth, setDepth] = useState(3);
    const [depthInput, setDepthInput] = useState('3');
    const [logs, setLogs] = useState<any[]>([]);
    const [status, setStatus] = useState('idle'); // idle, researching, complete
    const [progress, setProgress] = useState(0);
    const scrollRef = useRef<HTMLDivElement>(null);
    const socketRef = useRef<WebSocket | null>(null);
    const reconnectAttemptsRef = useRef(0);
    const reconnectTimerRef = useRef<number | null>(null);
    const reconnectPolicy = createReconnectPolicy();

    useEffect(() => {
        const wsUrl = resolveCoreSseUrl(process.env.NEXT_PUBLIC_CORE_WS_URL);

        const connect = () => {
            const ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                reconnectAttemptsRef.current = 0;
                setLogs(prev => [...prev, { type: 'system', message: 'Connected to TormentNexus Core' }]);
            };

            ws.onmessage = (event: MessageEvent) => {
                try {
                    const data = JSON.parse(event.data);

                    if (data.type === 'RESEARCH_UPDATE') {
                        const payload = data.payload;
                        setLogs(prev => [...prev, payload]);
                        if (payload.progress) setProgress(payload.progress);
                        if (scrollRef.current) {
                            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
                        }
                    }

                    if (data.type === 'RESEARCH_COMPLETE') {
                        setStatus('complete');
                        setLogs(prev => [...prev, { type: 'success', message: 'Research Complete!', report: data.payload.report }]);
                        setProgress(100);
                    }
                } catch (e) { }
            };

            ws.onerror = () => {
                ws.close();
            };

            ws.onclose = () => {
                socketRef.current = null;
                if (shouldRetryReconnect(reconnectAttemptsRef.current, reconnectPolicy)) {
                    reconnectAttemptsRef.current += 1;
                    const delayMs = getReconnectDelayMs(reconnectAttemptsRef.current, reconnectPolicy);
                    reconnectTimerRef.current = window.setTimeout(connect, delayMs);
                }
            };

            socketRef.current = ws;
        };

        connect();

        return () => {
            if (reconnectTimerRef.current !== null) {
                window.clearTimeout(reconnectTimerRef.current);
            }
            socketRef.current?.close();
        };
    }, []);

    const startResearch = () => {
        if (!topic) return;

        const normalizedDepth = normalizeNumericInput(depthInput, 3, 1, 10);

        setDepth(normalizedDepth);
        setDepthInput(String(normalizedDepth));
        setStatus('researching');
        setLogs([]);
        setProgress(0);

        if (socketRef.current?.readyState === WebSocket.OPEN) {
            socketRef.current.send(JSON.stringify({
                jsonrpc: '2.0',
                method: 'research',
                params: { topic, depth: normalizedDepth },
                id: Date.now()
            }));
        }
    };

    return (
        <div className="p-6 max-w-4xl mx-auto space-y-6">
            <h1 className="text-3xl font-bold tracking-tight text-primary">Deep Research Console 🧠</h1>

            <Card className="p-6 space-y-4">
                <div className="flex gap-4">
                    <Input
                        placeholder="Research Topic (e.g. 'Advanced React Patterns')"
                        value={topic}
                        onChange={e => setTopic(e.target.value)}
                        className="flex-1"
                        disabled={status === 'researching'}
                    />
                    <Input
                        type="number"
                        value={depthInput}
                        onChange={(e) => {
                            const nextValue = e.target.value;
                            setDepthInput(nextValue);
                            setDepth(normalizeNumericInput(nextValue, 3, 1, 10));
                        }}
                        className="w-24"
                        min={1}
                        max={10}
                        disabled={status === 'researching'}
                    />
                    <Button onClick={startResearch} disabled={status === 'researching'}>
                        {status === 'researching' ? 'Researching...' : 'Start Deep Dive'}
                    </Button>
                </div>

                {status === 'researching' && (
                    <div className="w-full bg-secondary h-2 rounded-full overflow-hidden">
                        <div
                            className="bg-primary h-full transition-all duration-500"
                            style={{ width: `${progress}%` }}
                        />
                    </div>
                )}
            </Card>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Card className="p-4 h-[500px] flex flex-col">
                    <h3 className="font-semibold mb-2">Live Logs</h3>
                    <ScrollArea className="flex-1 rounded border p-2 bg-black/90 font-mono text-xs text-green-400" ref={scrollRef}>
                        {logs.map((log, i) => (
                            <div key={i} className="mb-1">
                                {log.status === 'reading' && <span className="text-blue-400">[READ] </span>}
                                {log.status === 'memorized' && <span className="text-green-400">[MEM] </span>}
                                {log.status === 'error' && <span className="text-red-500">[ERR] </span>}
                                {log.message || log.target || JSON.stringify(log)}
                            </div>
                        ))}
                    </ScrollArea>
                </Card>

                <Card className="p-4 h-[500px] flex flex-col">
                    <h3 className="font-semibold mb-2">Report Preview</h3>
                    <ScrollArea className="flex-1 rounded border p-4 bg-secondary/20">
                        {status === 'complete' ? (
                            <div className="prose prose-sm dark:prose-invert">
                                {logs.find(l => l.report)?.report?.split('\n').map((line: string, i: number) => (
                                    <p key={i}>{line}</p>
                                ))}
                            </div>
                        ) : (
                            <div className="flex items-center justify-center h-full text-muted-foreground">
                                Report will appear here...
                            </div>
                        )}
                    </ScrollArea>
                </Card>
            </div>
        </div>
    );
}
