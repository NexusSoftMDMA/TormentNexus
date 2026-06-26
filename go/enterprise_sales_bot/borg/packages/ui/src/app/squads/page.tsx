
"use client";

import { useQuery, useMutation } from '@tanstack/react-query';
import { trpc } from '@/utils/trpc';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loader2, Users, Terminal, GitBranch, Trash2, Plus } from 'lucide-react';
import { useState } from 'react';

export default function SquadsPage() {
    const utils = trpc.useContext();
    const squads = trpc.squad.list.useQuery(undefined, {
        refetchInterval: 2000 // Real-time updates
    });

    const spawnMutation = trpc.squad.spawn.useMutation({
        onSuccess: () => utils.squad.list.invalidate()
    });

    const killMutation = trpc.squad.kill.useMutation({
        onSuccess: () => utils.squad.list.invalidate()
    });

    // Indexer Hooks
    const { data: indexerStatus, refetch: refetchIndexer } = trpc.squad.getIndexerStatus.useQuery(undefined, {
        refetchInterval: 5000
    });

    const indexerMutation = trpc.squad.toggleIndexer.useMutation({
        onSuccess: () => refetchIndexer()
    });

    const [branch, setBranch] = useState("chore/feature-x");
    const [goal, setGoal] = useState("Implement Feature X");

    const handleSpawn = () => {
        spawnMutation.mutate({ branch, goal });
    };

    return (
        <div className="container mx-auto p-4 space-y-6 max-w-4xl">
            <header className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight mb-2 flex items-center gap-2">
                        <Users className="h-8 w-8 text-indigo-500" />
                        Squad Command
                    </h1>
                    <p className="text-muted-foreground">
                        Manage autonomous agents working on parallel git worktrees.
                    </p>
                </div>
            </header>



            {/* Continuous Indexer Card */}
            <Card className="bg-muted/30">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-lg font-medium flex items-center gap-2">
                        <div className={`w-2 h-2 rounded-full ${indexerStatus?.running ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
                        Continuous Indexing Squad
                    </CardTitle>
                    <Button
                        variant={indexerStatus?.running ? "destructive" : "default"}
                        size="sm"
                        onClick={() => indexerMutation.mutate({ enabled: !indexerStatus?.running })}
                        disabled={indexerMutation.isPending}
                    >
                        {indexerStatus?.running ? "Stop Indexer" : "Start Indexer"}
                    </Button>
                </CardHeader>
                <CardContent>
                    <p className="text-sm text-muted-foreground mb-2">
                        Automatically scans codebase every 5 minutes to keep Symbols and Graph up-to-date.
                    </p>
                    {indexerStatus?.indexing && (
                        <div className="flex items-center gap-2 text-xs text-blue-400">
                            <Loader2 className="h-3 w-3 animate-spin" />
                            Currently Indexing...
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* Active Squads List */}
            <div className="grid gap-4">
                {squads.isLoading ? (
                    <div className="flex items-center justify-center p-8">
                        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                    </div>
                ) : squads.data?.length === 0 ? (
                    <Card className="border-dashed">
                        <CardContent className="flex flex-col items-center justify-center p-8 text-center">
                            <Users className="h-12 w-12 text-muted-foreground mb-4" />
                            <h3 className="text-lg font-medium">No Squad Members Active</h3>
                            <p className="text-sm text-muted-foreground mb-4">
                                Spawn an agent to offload tasks to a background process.
                            </p>
                        </CardContent>
                    </Card>
                ) : (
                    squads.data?.map((member: any) => (
                        <Card key={member.id} className="overflow-hidden border-l-4 border-l-indigo-500">
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-base font-mono flex items-center gap-2">
                                    <Terminal className="h-4 w-4" />
                                    {member.id}
                                </CardTitle>
                                <Badge variant={member.status === 'busy' ? "default" : "secondary"}>
                                    {member.status.toUpperCase()}
                                </Badge>
                            </CardHeader>
                            <CardContent>
                                <div className="grid gap-2 text-sm">
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <GitBranch className="h-4 w-4" />
                                        <span className="font-mono text-xs bg-muted px-2 py-0.5 rounded">
                                            {member.branch}
                                        </span>
                                    </div>
                                    <div className="flex items-center justify-between mt-4">
                                        <div className="text-xs text-muted-foreground">
                                            Active: {member.active ? 'YES' : 'NO'}
                                        </div>
                                        <Button
                                            variant="destructive"
                                            size="sm"
                                            onClick={() => killMutation.mutate({ branch: member.branch })}
                                            disabled={killMutation.isPending}
                                        >
                                            <Trash2 className="h-4 w-4 mr-2" />
                                            Terminate
                                        </Button>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    ))
                )}
            </div>

            {/* Spawn New (Simple Interface) */}
            <Card>
                <CardHeader>
                    <CardTitle>Deploy New Agent</CardTitle>
                    <CardDescription>Launch a new autonomous director on a fresh worktree.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid gap-2">
                        <label className="text-sm font-medium">Branch Name</label>
                        <input
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            value={branch}
                            onChange={(e) => setBranch(e.target.value)}
                        />
                    </div>
                    <div className="grid gap-2">
                        <label className="text-sm font-medium">Mission Goal</label>
                        <textarea
                            className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            value={goal}
                            onChange={(e) => setGoal(e.target.value)}
                        />
                    </div>
                    <Button onClick={handleSpawn} disabled={spawnMutation.isPending} className="w-full">
                        {spawnMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Plus className="mr-2 h-4 w-4" />}
                        Deploy Agent
                    </Button>
                </CardContent>
            </Card>
        </div >
    );
}
