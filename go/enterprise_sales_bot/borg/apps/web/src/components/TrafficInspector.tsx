"use client";

import { useEffect, useState, useRef } from 'react';
import { createReconnectPolicy, getReconnectDelayMs, resolveCoreWsUrl, shouldRetryReconnect } from '@tormentnexus/ui';
import { trpc } from '@/utils/trpc';

interface Packet {
    id: string;
    type: 'TOOL_CALL_START' | 'TOOL_CALL_END' | 'LOG';
    tool?: string;
    args?: any;
    result?: string;
    duration?: number;
    success?: boolean;
    timestamp: number;
}

export function TrafficInspector() {
    const [packets, setPackets] = useState<Packet[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const [customWsUrl, setCustomWsUrl] = useState<string>('');
    const [connectTrigger, setConnectTrigger] = useState(0); // Used to force reconnect
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectAttemptsRef = useRef(0);
    const reconnectPolicy = createReconnectPolicy();

    useEffect(() => {
        // Prefer custom URL if set, otherwise fallback to env
        const targetUrl = customWsUrl || process.env.NEXT_PUBLIC_CORE_WS_URL || 'ws://localhost:3000';
        const wsUrl = resolveCoreWsUrl(targetUrl);

        console.log(`[TrafficInspector] Connecting to: ${wsUrl}`);

        const connect = () => {
            if (wsRef.current?.readyState === WebSocket.OPEN) {
                wsRef.current.close();
            }

            // Connect to TormentNexus Core Bridge
            const ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                setIsConnected(true);
                reconnectAttemptsRef.current = 0;
            };

            ws.onclose = () => {
                setIsConnected(false);
                if (shouldRetryReconnect(reconnectAttemptsRef.current, reconnectPolicy)) {
                    reconnectAttemptsRef.current += 1;
                    const delayMs = getReconnectDelayMs(reconnectAttemptsRef.current, reconnectPolicy);
                    setTimeout(connect, delayMs); // Reconnect with capped backoff
                }
            };

            ws.onerror = () => {
                // Let onclose handle capped retries.
                ws.close();
            };

            ws.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data);
                    // Filter for interesting events
                    if (msg.type === 'TOOL_CALL_START' || msg.type === 'TOOL_CALL_END') {
                        addPacket({
                            ...msg,
                            timestamp: Date.now()
                        });
                    }
                } catch (e) { }
            };

            wsRef.current = ws;
        };

        connect();

        return () => {
            wsRef.current?.close();
        };
    }, [connectTrigger]);

    const addPacket = (packet: Packet) => {
        setPackets(prev => {
            // Avoid duplicates
            if (prev.some(p => p.id === packet.id && p.type === packet.type)) return prev;
            const newPackets = [packet, ...prev].slice(0, 50);
            return newPackets;
        });
    };

    // Replay Logic
    const utils = trpc.useContext();
    const handleReplay = async () => {
        // logs router is not active — replay not available
        console.warn('[TrafficInspector] Log replay not available — logs router is disabled');
    };

    return (
        <div className="bg-black/80 rounded-xl border border-zinc-800 overflow-hidden flex flex-col h-[600px]">
            {/* Header */}
            <div className="p-4 border-b border-zinc-800 bg-zinc-900/50 flex flex-col gap-2">
                <div className="flex justify-between items-center">
                    <div className="flex items-center gap-2">
                        <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
                        <h2 className="font-mono font-bold text-zinc-300">NETWORK TRAFFIC (MCP)</h2>
                    </div>
                    <div className="flex gap-2">
                        <button
                            onClick={handleReplay}
                            className="text-xs px-2 py-1 bg-blue-900/30 text-blue-400 hover:bg-blue-900/50 rounded"
                        >
                            REPLAY LOGS
                        </button>
                        <button
                            onClick={() => setPackets([])}
                            className="text-xs text-zinc-500 hover:text-white"
                        >
                            CLEAR
                        </button>
                    </div>
                </div>

                {/* Connection Config Bar */}
                <div className="flex gap-2 items-center">
                    <input
                        type="text"
                        value={customWsUrl}
                        onChange={(e) => setCustomWsUrl(e.target.value)}
                        placeholder="ws://localhost:3000"
                        className="flex-1 bg-black/50 border border-zinc-800 rounded px-2 py-1 text-xs font-mono text-zinc-400 focus:border-blue-500 outline-none"
                    />
                    <button
                        onClick={() => setConnectTrigger(prev => prev + 1)}
                        className="text-xs px-3 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 rounded font-mono"
                    >
                        CONNECT
                    </button>
                </div>
            </div>

            {/* Packet Stream */}
            <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-sm">
                {packets.length === 0 && (
                    <div className="text-zinc-600 text-center mt-20">Waiting for traffic...</div>
                )}
                {packets.map((p, i) => (
                    <PacketRow key={`${p.id}-${p.type}-${i}`} packet={p} />
                ))}
            </div>
        </div>
    );
}

function PacketRow({ packet }: { packet: Packet }) {
    const isStart = packet.type === 'TOOL_CALL_START';

    // Color coding
    const borderColor = isStart ? 'border-blue-900/30' : (packet.success ? 'border-green-900/30' : 'border-red-900/30');
    const bgColor = isStart ? 'bg-blue-900/10' : (packet.success ? 'bg-green-900/10' : 'bg-red-900/10');
    const icon = isStart ? '→' : (packet.success ? '✓' : '✗');
    const iconColor = isStart ? 'text-blue-400' : (packet.success ? 'text-green-400' : 'text-red-400');

    return (
        <div className={`p-3 rounded border ${borderColor} ${bgColor} transition-all hover:bg-zinc-800/50`}>
            <div className="flex justify-between items-start">
                <div className="flex items-center gap-3">
                    <span className={`font-bold ${iconColor}`}>{icon}</span>
                    <span className="text-zinc-300 font-bold">{packet.tool}</span>
                    <span className="text-xs text-zinc-600">#{packet.id.substring(0, 4)}</span>
                </div>
                <span className="text-xs text-zinc-600">
                    {new Date(packet.timestamp).toLocaleTimeString().split(' ')[0]}
                </span>
            </div>

            {/* Details */}
            <div className="mt-2 pl-6">
                {isStart ? (
                    <div className="text-zinc-400 break-all text-xs">
                        Args: <span className="text-blue-300/80">{JSON.stringify(packet.args).substring(0, 200)}</span>
                    </div>
                ) : (
                    <div className="text-zinc-400 break-all text-xs">
                        Result: <span className="text-zinc-500">{packet.result}</span>
                        <div className="mt-1 text-zinc-600">
                            Duration: {packet.duration}ms
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
