"use client";

import { useEffect, useState, useRef } from 'react';
import { createReconnectPolicy, getReconnectDelayMs, resolveCoreSseUrl, shouldRetryReconnect } from '@tormentnexus/ui';
import { motion, AnimatePresence } from 'framer-motion';

export function MirrorView() {
    const [screenshot, setScreenshot] = useState<string | null>(null);
    const [isMirroring, setIsMirroring] = useState(false);
    const [isConnected, setIsConnected] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectAttemptsRef = useRef(0);
    const reconnectPolicy = createReconnectPolicy();
    const wsUrlRef = useRef<string | null>(null);

    if (!wsUrlRef.current && typeof window !== 'undefined') {
        wsUrlRef.current = resolveCoreSseUrl(process.env.NEXT_PUBLIC_CORE_WS_URL);
    }

    const connect = () => {
        if (!wsUrlRef.current) {
            return;
        }

        const ws = new WebSocket(wsUrlRef.current);

        ws.onopen = () => {
            setIsConnected(true);
            reconnectAttemptsRef.current = 0;
            // If mirroring was previously active (e.g. on reconnect), re-enable
            if (isMirroring) {
                ws.send(JSON.stringify({ type: 'SET_MIRROR_ACTIVE', active: true }));
            }
        };

        ws.onclose = () => {
            setIsConnected(false);
            wsRef.current = null;
            if (isMirroring && shouldRetryReconnect(reconnectAttemptsRef.current, reconnectPolicy)) {
                reconnectAttemptsRef.current += 1;
                const delayMs = getReconnectDelayMs(reconnectAttemptsRef.current, reconnectPolicy);
                setTimeout(connect, delayMs);
            }
        };

        ws.onerror = () => {
            ws.close();
        };

        ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                if (msg.type === 'BROWSER_MIRROR_UPDATE') {
                    setScreenshot(msg.screenshot);
                }
            } catch (e) { }
        };

        wsRef.current = ws;
    };

    useEffect(() => {
        return () => {
            if (wsRef.current?.readyState === WebSocket.OPEN) {
                wsRef.current.send(JSON.stringify({ type: 'SET_MIRROR_ACTIVE', active: false }));
            }
            wsRef.current?.close();
        };
    }, []);

    useEffect(() => {
        if (isMirroring && !wsRef.current) {
            connect();
        }

        if (!isMirroring && wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({ type: 'SET_MIRROR_ACTIVE', active: false }));
        }
    }, [isMirroring]);

    const toggleMirror = () => {
        const nextState = !isMirroring;
        setIsMirroring(nextState);

        if (nextState && !wsRef.current) {
            connect();
            return;
        }

        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({
                type: 'SET_MIRROR_ACTIVE',
                active: nextState,
                interval: 3000 // 3s for smoother mirroring
            }));
        }
    };

    return (
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden flex flex-col h-full min-h-[400px]">
            {/* Header */}
            <div className="p-3 border-b border-zinc-800 flex justify-between items-center bg-zinc-900/80 backdrop-blur-sm z-10">
                <div className="flex items-center gap-2">
                    <span className="text-xl">📺</span>
                    <h2 className="font-bold text-zinc-200">Live Tab Mirror</h2>
                    {isMirroring && isConnected && (
                        <motion.div
                            animate={{ opacity: [1, 0.5, 1] }}
                            transition={{ repeat: Infinity, duration: 1.5 }}
                            className="text-[10px] bg-red-500/20 text-red-500 px-2 py-0.5 rounded border border-red-500/50 font-bold"
                        >
                            LIVE
                        </motion.div>
                    )}
                </div>
                <button
                    onClick={toggleMirror}
                    className={`px-3 py-1 rounded-lg text-sm font-medium transition-all ${isMirroring
                            ? 'bg-red-600 hover:bg-red-700 text-white shadow-lg shadow-red-900/20'
                            : 'bg-zinc-800 hover:bg-zinc-700 text-zinc-300'
                        }`}
                >
                    {isMirroring ? 'Stop Mirror' : 'Start Mirror'}
                </button>
            </div>

            {/* Viewport */}
            <div className="flex-1 relative bg-black flex items-center justify-center overflow-hidden">
                {!isConnected ? (
                    <div className="text-zinc-600 flex flex-col items-center gap-2">
                        <div className="w-8 h-8 border-2 border-zinc-800 border-t-zinc-500 rounded-full animate-spin" />
                        <span>Connecting to Core...</span>
                    </div>
                ) : !isMirroring ? (
                    <div className="text-zinc-500 text-center p-8">
                        <div className="text-4xl mb-4 opacity-20">📡</div>
                        <p>Tab Mirroring is currently inactive.</p>
                        <p className="text-sm mt-2 opacity-60">Enable mirroring to see exactly what the agent sees.</p>
                    </div>
                ) : !screenshot ? (
                    <div className="text-zinc-600 animate-pulse">Waiting for first frame...</div>
                ) : (
                    <AnimatePresence mode="wait">
                        <motion.img
                            key={screenshot}
                            initial={{ opacity: 0.8 }}
                            animate={{ opacity: 1 }}
                            className="w-full h-full object-contain"
                            src={screenshot}
                            alt="Browser Mirror"
                        />
                    </AnimatePresence>
                )}

                {/* Overlay Status */}
                <div className="absolute bottom-2 left-2 flex gap-2">
                    <div className="bg-black/60 backdrop-blur text-[10px] text-zinc-400 px-2 py-1 rounded">
                        {isConnected ? '🟢 Core Connected' : '🔴 Core Offline'}
                    </div>
                </div>
            </div>
        </div>
    );
}
