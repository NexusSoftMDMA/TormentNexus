"use client";

import React, { useState, useEffect } from 'react';
import { trpc } from '@/utils/trpc';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Loader2, Play, AlertCircle, CheckCircle, XCircle, RefreshCw, Terminal } from 'lucide-react';

export default function AutoDevPage() {
    const [target, setTarget] = useState('');
    const [type, setType] = useState<'test' | 'lint'>('test');

    // Poll loops
    const { data: loops, refetch } = trpc.autoDev.getLoops.useQuery(undefined, {
        refetchInterval: 2000
    });

    const startMutation = trpc.autoDev.startLoop.useMutation({
        onSuccess: () => refetch()
    });

    const cancelMutation = trpc.autoDev.cancelLoop.useMutation({
        onSuccess: () => refetch()
    });

    const clearMutation = trpc.autoDev.clearCompleted.useMutation({
        onSuccess: () => refetch()
    });

    const handleStart = () => {
        if (!target) return;
        startMutation.mutate({
            type,
            target,
            maxAttempts: 5
        });
        setTarget('');
    };

    return (
        <div className="h-screen flex flex-col bg-neutral-950 text-neutral-200">
            <header className="h-14 border-b border-neutral-800 flex items-center px-6 bg-neutral-950 shrink-0">
                <h1 className="font-semibold text-lg text-neutral-200">Auto-Dev Loops</h1>
                <span className="ml-4 text-xs text-neutral-500 px-2 py-1 rounded bg-neutral-900 border border-neutral-800">
                    Autonomous Repair
                </span>
                <div className="ml-auto">
                    <Button variant="ghost" size="sm" onClick={() => clearMutation.mutate()}>Clear Completed</Button>
                </div>
            </header>

            <div className="flex-1 p-6 flex gap-6 overflow-hidden">
                {/* Control Panel */}
                <div className="w-80 shrink-0 flex flex-col gap-6">
                    <Card className="bg-neutral-900 border-neutral-800">
                        <CardHeader>
                            <CardTitle>Start Loop</CardTitle>
                            <CardDescription>Run autonomous fix-until-pass</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <Tabs value={type} onValueChange={(v: any) => setType(v)}>
                                <TabsList className="w-full">
                                    <TabsTrigger value="test" className="flex-1">Test</TabsTrigger>
                                    <TabsTrigger value="lint" className="flex-1">Lint</TabsTrigger>
                                </TabsList>
                            </Tabs>

                            <div className="space-y-2">
                                <label className="text-xs font-medium text-neutral-400">Target (File/Pattern)</label>
                                <Input
                                    placeholder={type === 'test' ? "src/foo.test.ts" : "src/"}
                                    value={target}
                                    onChange={(e) => setTarget(e.target.value)}
                                    className="bg-neutral-950 font-mono text-xs"
                                />
                            </div>

                            <Button
                                className="w-full"
                                disabled={!target || startMutation.isPending}
                                onClick={handleStart}
                            >
                                {startMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                Start Loop
                            </Button>
                        </CardContent>
                    </Card>

                    <div className="text-xs text-neutral-500">
                        <p>TormentNexus will:</p>
                        <ol className="list-decimal list-inside mt-2 space-y-1">
                            <li>Run the command</li>
                            <li>Analyze failure output</li>
                            <li>Attempt to modify code</li>
                            <li>Repeat until pass (max 5)</li>
                        </ol>
                    </div>
                </div>

                {/* Active Loops Stream */}
                <div className="flex-1 flex flex-col gap-4 overflow-hidden">
                    <h2 className="text-sm font-medium text-neutral-400 mb-2">Active Sessions</h2>
                    <ScrollArea className="flex-1 pr-4">
                        <div className="space-y-4 pb-20">
                            {loops?.map((loop: any) => (
                                <Card key={loop.id} className="bg-neutral-900 border-neutral-800 overflow-hidden">
                                    <div className="flex items-center justify-between p-4 border-b border-neutral-800 bg-neutral-900/50">
                                        <div className="flex items-center gap-3">
                                            <StatusBadge status={loop.status} />
                                            <span className="font-mono text-sm font-semibold">{loop.config.type.toUpperCase()}</span>
                                            <span className="text-sm text-neutral-400">{loop.config.target}</span>
                                        </div>
                                        <div className="flex items-center gap-4">
                                            <div className="text-xs text-neutral-500">
                                                Attempt {loop.currentAttempt}/{loop.config.maxAttempts}
                                            </div>
                                            {loop.status === 'running' && (
                                                <Button variant="ghost" size="sm" onClick={() => cancelMutation.mutate({ loopId: loop.id })}>
                                                    Cancel
                                                </Button>
                                            )}
                                        </div>
                                    </div>

                                    <div className="p-0">
                                        <div className="bg-black p-4 font-mono text-xs text-neutral-300 max-h-60 overflow-y-auto whitespace-pre-wrap">
                                            {loop.lastOutput || "Waiting for output..."}
                                        </div>
                                    </div>
                                </Card>
                            ))}
                            {loops?.length === 0 && (
                                <div className="text-center py-20 text-neutral-600">
                                    No active development loops.
                                </div>
                            )}
                        </div>
                    </ScrollArea>
                </div>
            </div>
        </div>
    );
}

function StatusBadge({ status }: { status: string }) {
    if (status === 'running') return <Badge variant="secondary" className="bg-blue-900/50 text-blue-400 hover:bg-blue-900/70"><RefreshCw className="w-3 h-3 mr-1 animate-spin" /> Running</Badge>;
    if (status === 'success') return <Badge variant="secondary" className="bg-green-900/50 text-green-400 hover:bg-green-900/70"><CheckCircle className="w-3 h-3 mr-1" /> Passed</Badge>;
    if (status === 'failed') return <Badge variant="secondary" className="bg-red-900/50 text-red-400 hover:bg-red-900/70"><XCircle className="w-3 h-3 mr-1" /> Failed</Badge>;
    return <Badge variant="outline">{status}</Badge>;
}
