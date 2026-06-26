'use client';

import React from 'react';
import { trpc } from '@/utils/trpc';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@tormentnexus/ui';
import { Badge } from "@tormentnexus/ui";
import { Activity, File, Terminal, Zap, Cpu, Clock } from 'lucide-react';

export default function EventsPage() {
    // Fetch real events from EventBus via pulse router
    const { data: events = [], isLoading } = trpc.pulse.getLatestEvents.useQuery(
        { limit: 50 },
        { refetchInterval: 3000 } // Poll every 3s for real-time feel
    );
    const { data: systemStatus } = trpc.pulse.getSystemStatus.useQuery(undefined, {
        refetchInterval: 5000,
    });

    // Categorize events by type
    const eventTypeColors: Record<string, string> = {
        'file:change': 'text-blue-400',
        'file:create': 'text-green-400',
        'file:delete': 'text-red-400',
        'agent:heartbeat': 'text-purple-400',
        'terminal:output': 'text-yellow-400',
        'heal:start': 'text-orange-400',
        'heal:complete': 'text-green-400',
        'error': 'text-red-400',
    };

    const getEventColor = (type: string) => {
        return eventTypeColors[type] || 'text-zinc-400';
    };

    const formatTime = (ts: number) => {
        const d = new Date(ts);
        return d.toLocaleTimeString();
    };

    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Nervous System</h1>
                    <p className="text-muted-foreground">Real-time event stream from sensors and reactors.</p>
                </div>
                <div className="flex gap-3 items-center">
                    {systemStatus && (
                        <Badge variant="outline" className="border-zinc-700">
                            <Clock className="w-3 h-3 mr-1" />
                            {Math.round((systemStatus as any).uptime ?? 0)}s uptime
                        </Badge>
                    )}
                    <Badge variant="outline" className={events.length > 0 ? "text-green-400 border-green-900 bg-green-950/30 animate-pulse" : "text-zinc-500 border-zinc-700"}>
                        <Activity className="w-3 h-3 mr-2" /> {events.length > 0 ? 'Live' : 'Idle'}
                    </Badge>
                </div>
            </div>

            {/* Stats Row */}
            <div className="grid grid-cols-4 gap-4">
                <Card className="border-zinc-800 bg-zinc-950/50">
                    <CardContent className="pt-4 text-center">
                        <div className="text-2xl font-bold text-white">{events.length}</div>
                        <div className="text-xs text-muted-foreground">Total Events</div>
                    </CardContent>
                </Card>
                <Card className="border-zinc-800 bg-zinc-950/50">
                    <CardContent className="pt-4 text-center">
                        <div className="text-2xl font-bold text-blue-400">
                            {events.filter((e: any) => e.type?.startsWith('file:')).length}
                        </div>
                        <div className="text-xs text-muted-foreground">File Events</div>
                    </CardContent>
                </Card>
                <Card className="border-zinc-800 bg-zinc-950/50">
                    <CardContent className="pt-4 text-center">
                        <div className="text-2xl font-bold text-purple-400">
                            {events.filter((e: any) => e.type?.startsWith('agent:')).length}
                        </div>
                        <div className="text-xs text-muted-foreground">Agent Events</div>
                    </CardContent>
                </Card>
                <Card className="border-zinc-800 bg-zinc-950/50">
                    <CardContent className="pt-4 text-center">
                        <div className="text-2xl font-bold text-green-400">
                            {(systemStatus as any)?.agents?.length ?? 0}
                        </div>
                        <div className="text-xs text-muted-foreground">Active Agents</div>
                    </CardContent>
                </Card>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Active Sensors + System Status */}
                <Card className="col-span-1 border-zinc-800 bg-zinc-950/50">
                    <CardHeader>
                        <CardTitle className="text-sm font-medium uppercase text-muted-foreground">System Status</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <Cpu className="w-5 h-5 text-green-400" />
                                <span>TormentNexus Core</span>
                            </div>
                            <Badge className={systemStatus?.status === 'online' ? "bg-green-600" : "bg-red-600"}>
                                {systemStatus?.status === 'online' ? 'Online' : 'Offline'}
                            </Badge>
                        </div>
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <File className="w-5 h-5 text-blue-400" />
                                <span>File Sensor</span>
                            </div>
                            <Badge className="bg-green-600">Active</Badge>
                        </div>
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <Terminal className="w-5 h-5 text-yellow-400" />
                                <span>Terminal Sensor</span>
                            </div>
                            <Badge className="bg-green-600">Active</Badge>
                        </div>

                        {/* Active Agents */}
                        {(systemStatus as any)?.agents && (systemStatus as any).agents.length > 0 && (
                            <div className="mt-4 pt-4 border-t border-zinc-800">
                                <div className="text-xs text-muted-foreground mb-2 uppercase">Active Agents</div>
                                {(systemStatus as any).agents.map((agent: string, i: number) => (
                                    <div key={i} className="text-sm font-mono text-zinc-300 mb-1">{agent}</div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>

                {/* Event Stream */}
                <Card className="col-span-1 lg:col-span-2 border-zinc-800 bg-black">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><Zap className="w-5 h-5 text-purple-400" /> Event Stream</CardTitle>
                        <CardDescription>{events.length} events in buffer</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="h-[500px] overflow-y-auto pr-4 space-y-2">
                            {isLoading ? (
                                <div className="text-zinc-500 text-center py-10">Loading events...</div>
                            ) : events.length === 0 ? (
                                <div className="text-zinc-500 text-center py-10 italic">
                                    No events recorded yet. Events will appear as the system operates.
                                </div>
                            ) : (
                                events.slice().reverse().map((event: any, i: number) => (
                                    <div key={i} className="flex flex-col p-3 rounded-md border border-zinc-900 bg-zinc-900/30 text-sm">
                                        <div className="flex justify-between items-start mb-1">
                                            <span className={`font-mono ${getEventColor(event.type)}`}>
                                                {event.type || 'unknown'}
                                            </span>
                                            <span className="text-xs text-zinc-500">
                                                {event.timestamp ? formatTime(event.timestamp) : '—'}
                                            </span>
                                        </div>
                                        {event.source && (
                                            <div className="text-zinc-400 text-xs">Source: {event.source}</div>
                                        )}
                                        {event.data && (
                                            <div className="font-mono text-xs text-zinc-600 mt-1 truncate max-w-full">
                                                {typeof event.data === 'string' ? event.data : JSON.stringify(event.data).slice(0, 200)}
                                            </div>
                                        )}
                                    </div>
                                ))
                            )}
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
