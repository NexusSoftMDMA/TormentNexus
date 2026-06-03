'use client';

import { trpc } from '@/utils/trpc';
import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';
import { Badge } from '@tormentnexus/ui';
import { ScrollArea } from '@tormentnexus/ui';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@tormentnexus/ui";
import { GitBranch, GitCommit, Play, Rewind, CheckCircle, XCircle, AlertCircle, FileCode } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from "@tormentnexus/ui";

export default function PlansDashboard() {
    const [activeTab, setActiveTab] = useState('diffs');

    const { data: modeData, refetch: refetchMode } = trpc.planService.getMode.useQuery();
    const { data: diffs, refetch: refetchDiffs } = trpc.planService.getDiffs.useQuery();
    const { data: checkpoints, refetch: refetchCheckpoints } = trpc.planService.getCheckpoints.useQuery();

    const setModeMutation = trpc.planService.setMode.useMutation({
        onSuccess: () => refetchMode()
    });

    const approveMutation = trpc.planService.approveDiff.useMutation({ onSuccess: () => refetchDiffs() });
    const rejectMutation = trpc.planService.rejectDiff.useMutation({ onSuccess: () => refetchDiffs() });
    const applyAllMutation = trpc.planService.applyAll.useMutation({ onSuccess: () => refetchDiffs() });
    const rollbackMutation = trpc.planService.rollback.useMutation({ onSuccess: () => { refetchDiffs(); refetchCheckpoints(); } });
    const createCheckpointMutation = trpc.planService.createCheckpoint.useMutation({ onSuccess: () => refetchCheckpoints() });

    const currentMode = modeData?.mode || 'PLAN';

    const handleModeSwitch = (mode: 'PLAN' | 'BUILD') => {
        setModeMutation.mutate({ mode });
    };

    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Plan & Build</h1>
                    <p className="text-muted-foreground">Detailed diff sandbox and checkpoint management.</p>
                </div>
                <div className="flex items-center gap-4">
                    <div className="flex items-center bg-zinc-900 rounded-lg p-1 border border-zinc-800">
                        <Button
                            variant={currentMode === 'PLAN' ? 'secondary' : 'ghost'}
                            size="sm"
                            onClick={() => handleModeSwitch('PLAN')}
                            className={currentMode === 'PLAN' ? 'bg-blue-900/50 text-blue-200' : ''}
                        >
                            <BrainIcon className="w-4 h-4 mr-2" /> PLAN
                        </Button>
                        <Button
                            variant={currentMode === 'BUILD' ? 'secondary' : 'ghost'}
                            size="sm"
                            onClick={() => handleModeSwitch('BUILD')}
                            className={currentMode === 'BUILD' ? 'bg-red-900/50 text-red-200' : ''}
                        >
                            <HammerIcon className="w-4 h-4 mr-2" /> BUILD
                        </Button>
                    </div>
                </div>
            </div>

            {currentMode === 'PLAN' && (
                <Alert className="bg-blue-950/20 border-blue-900/50 text-blue-200">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Planning Mode</AlertTitle>
                    <AlertDescription>
                        Changes are explored but not applied. Switch to BUILD mode to execute changes.
                    </AlertDescription>
                </Alert>
            )}

            <Tabs defaultValue="diffs" className="space-y-4">
                <TabsList>
                    <TabsTrigger value="diffs">Pending Changes ({diffs?.length || 0})</TabsTrigger>
                    <TabsTrigger value="checkpoints">Checkpoints ({checkpoints?.length || 0})</TabsTrigger>
                </TabsList>

                <TabsContent value="diffs" className="space-y-4">
                    <div className="flex justify-end gap-2">
                        <Button
                            variant="default"
                            disabled={currentMode !== 'BUILD' || !diffs?.length}
                            onClick={() => applyAllMutation.mutate()}
                        >
                            <Play className="w-4 h-4 mr-2" /> Apply All Approved
                        </Button>
                    </div>

                    <ScrollArea className="h-[600px]">
                        <div className="space-y-4">
                            {diffs?.map((diff: any) => (
                                <Card key={diff.id} className="border-zinc-800 bg-zinc-950/50">
                                    <CardHeader className="py-3">
                                        <div className="flex justify-between items-start">
                                            <div className="flex items-center gap-2">
                                                <FileCode className="w-4 h-4 text-zinc-400" />
                                                <span className="font-mono text-sm">{diff.filePath}</span>
                                                <Badge variant={diff.status === 'approved' ? 'default' : 'secondary'}>
                                                    {diff.status}
                                                </Badge>
                                            </div>
                                            <div className="flex gap-2">
                                                <Button size="icon" variant="ghost" onClick={() => approveMutation.mutate({ diffId: diff.id })}>
                                                    <CheckCircle className="w-4 h-4 text-green-500" />
                                                </Button>
                                                <Button size="icon" variant="ghost" onClick={() => rejectMutation.mutate({ diffId: diff.id })}>
                                                    <XCircle className="w-4 h-4 text-red-500" />
                                                </Button>
                                            </div>
                                        </div>
                                    </CardHeader>
                                    <CardContent className="py-0 pb-3">
                                        <pre className="bg-black/50 p-2 rounded text-xs font-mono overflow-x-auto text-zinc-300">
                                            {diff.proposedContent}
                                        </pre>
                                    </CardContent>
                                </Card>
                            ))}
                            {diffs?.length === 0 && (
                                <div className="text-center p-12 text-zinc-500">
                                    No pending changes in sandbox.
                                </div>
                            )}
                        </div>
                    </ScrollArea>
                </TabsContent>

                <TabsContent value="checkpoints" className="space-y-4">
                    <div className="flex justify-end gap-2">
                        <Button
                            variant="outline"
                            onClick={() => {
                                const name = prompt("Checkpoint Name:");
                                if (name) createCheckpointMutation.mutate({ name });
                            }}
                        >
                            <GitCommit className="w-4 h-4 mr-2" /> Create Checkpoint
                        </Button>
                    </div>
                    <ScrollArea className="h-[600px]">
                        <div className="space-y-4">
                            {checkpoints?.slice().reverse().map((cp: any) => (
                                <Card key={cp.id} className="border-zinc-800 bg-zinc-950/50">
                                    <CardHeader className="py-3">
                                        <div className="flex justify-between items-center">
                                            <div className="flex items-center gap-2">
                                                <GitBranch className="w-4 h-4 text-purple-400" />
                                                <span className="font-semibold">{cp.name}</span>
                                                <span className="text-xs text-muted-foreground">{new Date(cp.timestamp).toLocaleString()}</span>
                                            </div>
                                            <Button
                                                variant="destructive"
                                                size="sm"
                                                onClick={() => {
                                                    if (confirm(`Rollback to ${cp.name}? This will undo changes after this point.`)) {
                                                        rollbackMutation.mutate({ checkpointId: cp.id });
                                                    }
                                                }}
                                            >
                                                <Rewind className="w-4 h-4 mr-2" /> Rollback
                                            </Button>
                                        </div>
                                        {cp.description && <CardDescription>{cp.description}</CardDescription>}
                                    </CardHeader>
                                </Card>
                            ))}
                        </div>
                    </ScrollArea>
                </TabsContent>
            </Tabs>
        </div>
    );
}

function BrainIcon(props: any) {
    return (
        <svg
            {...props}
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <path d="M9.5 2A2.5 2.5 0 0 1 12 4.5v15a2.5 2.5 0 0 1-4.96.44 2.5 2.5 0 0 1-2.96-3.08 3 3 0 0 1-.34-5.58 2.5 2.5 0 0 1 1.32-4.24 2.5 2.5 0 0 1 1.98-3A2.5 2.5 0 0 1 9.5 2Z" />
            <path d="M14.5 2A2.5 2.5 0 0 0 12 4.5v15a2.5 2.5 0 0 0 4.96.44 2.5 2.5 0 0 0 2.96-3.08 3 3 0 0 0 .34-5.58 2.5 2.5 0 0 0-1.32-4.24 2.5 2.5 0 0 0-1.98-3A2.5 2.5 0 0 0 14.5 2Z" />
        </svg>
    )
}

function HammerIcon(props: any) {
    return (
        <svg
            {...props}
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <path d="m15 12-8.5 8.5c-.83.83-2.17.83-3 0 0 0 0 0 0 0a2.12 2.12 0 0 1 0-3L12 9" />
            <path d="M17.64 15 22 10.64" />
            <path d="m20.91 11.7-1.25-1.25c-.6-.6-.93-1.4-.93-2.25V7.86c0-.55-.45-1-1-1H16.4c-.84 0-1.65-.33-2.25-.93L12.9 4.7a1.001 1.001 0 0 0-1.41 0l-1.6 1.6a1.001 1.001 0 0 0 0 1.41l5.14 5.14" />
        </svg>
    )
}
