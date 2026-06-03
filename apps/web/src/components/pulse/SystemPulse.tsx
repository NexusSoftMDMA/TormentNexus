
"use client";


import { trpc } from '@/utils/trpc';
import { Card, CardHeader, CardTitle, CardContent, Badge } from '@tormentnexus/ui';

export function SystemPulse() {
    // Polling System Pulse
    const { data } = trpc.pulse.getSystemStatus.useQuery(undefined, {
        refetchInterval: 5000
    });
    const status = data as any;

    return (
        <div className="space-y-4">
            <Card>
                <CardHeader>
                    <CardTitle className="flex justify-between items-center">
                        <span>System Status</span>
                        {status?.status === 'online' ? (
                            <Badge variant="default" className="animate-pulse">Online</Badge>
                        ) : (
                            <Badge variant="destructive">Offline</Badge>
                        )}
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div className="bg-zinc-900 p-4 rounded-lg">
                            <div className="text-zinc-400 text-sm">Active Agents</div>
                            <div className="text-2xl font-mono text-cyan-400">
                                {status?.agents?.length || 0}
                            </div>
                        </div>
                        <div className="bg-zinc-900 p-4 rounded-lg">
                            <div className="text-zinc-400 text-sm">Uptime</div>
                            <div className="text-xl font-mono text-zinc-300">
                                {Math.floor((status?.uptime || 0) / 60)}m
                            </div>
                        </div>
                        <div className="bg-zinc-900 p-4 rounded-lg">
                            <div className="text-zinc-400 text-sm">Memory Core</div>
                            <div className="text-xl font-mono text-zinc-300">
                                {status?.memoryInitialized ? 'Active' : 'Standby'}
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Card className="h-[500px] flex flex-col">
                <CardHeader>
                    <CardTitle>Live Activity Stream</CardTitle>
                </CardHeader>
                <CardContent className="flex-1 bg-zinc-950 p-4 overflow-y-auto font-mono text-sm space-y-2">
                    <div className="text-zinc-500 italic">Waiting for events... (Event stream integration pending)</div>
                </CardContent>
            </Card>
        </div>
    );
}
