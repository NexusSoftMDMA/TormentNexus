'use client';

import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Badge } from "@tormentnexus/ui";
import { ScrollArea } from "@tormentnexus/ui";
import { useState, useEffect } from "react";
import { BrainCircuit, GitBranch, Shield, Zap } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { normalizeDirectorAutonomyLevel, normalizeDirectorPlan } from './director-page-normalizers';

export default function DirectorPage() {
    const { data: config } = trpc.directorConfig.get.useQuery();
    const { data: taskStatus } = trpc.getTaskStatus.useQuery({});
    const { data: autonomyLevel } = trpc.autonomy.getLevel.useQuery();

    const plan = normalizeDirectorPlan(config, taskStatus);
    const normalizedAutonomyLevel = normalizeDirectorAutonomyLevel(autonomyLevel);

    return (
        <div className="container mx-auto p-6 space-y-6 max-w-7xl">
            <div className="flex items-center gap-3 border-b pb-6">
                <div className="h-12 w-12 bg-amber-900/20 rounded-lg flex items-center justify-center border border-amber-500/30">
                    <BrainCircuit className="h-6 w-6 text-amber-400" />
                </div>
                <div>
                    <div className="flex items-center gap-3">
                        <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-amber-400 to-orange-500 bg-clip-text text-transparent">
                            The Director's Office
                        </h1>
                        <Badge variant="outline" className="border-amber-500/30 text-amber-500 bg-amber-500/10">Labs</Badge>
                    </div>
                    <p className="text-muted-foreground mt-1">High-Level Planning & Squad Orchestration (Experimental Autonomy Surface)</p>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Main Plan Visualizer */}
                <Card className="lg:col-span-2 border-amber-500/20 bg-amber-950/5">
                    <CardHeader>
                        <CardTitle className="flex items-center justify-between">
                            <span>Current Strategic Goal</span>
                            <Badge variant="outline" className={`border-amber-500 ${plan.status === 'IN_PROGRESS' ? 'text-amber-500 animate-pulse' : 'text-muted-foreground'}`}>
                                {plan.status}
                            </Badge>
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-xl font-medium mb-6 p-4 bg-background/50 rounded border">
                            "{plan.goal}"
                        </div>

                        <div className="relative border-l-2 border-muted ml-4 space-y-8 pl-8 py-2">
                            {plan.steps.length > 0 ? plan.steps.map((step) => (
                                <div key={step.id} className="relative">
                                    <div className={`absolute -left-[41px] h-4 w-4 rounded-full border-2 ${step.status === 'DONE' ? 'bg-green-500 border-green-500' :
                                        step.status === 'RUNNING' ? 'bg-amber-500 border-amber-500 animate-ping' :
                                            'bg-background border-muted'
                                        }`} />
                                    <div className="flex flex-col gap-1">
                                        <div className="font-mono text-sm text-muted-foreground uppercase">{step.action}</div>
                                        <div className="font-medium">{step.result}</div>
                                    </div>
                                </div>
                            )) : (
                                <div className="text-muted-foreground italic">No active tasks.</div>
                            )}
                        </div>
                    </CardContent>
                </Card>

                {/* Sidebar Stats */}
                <div className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-sm uppercase text-muted-foreground">Squad Status</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center justify-between p-3 border rounded bg-muted/20">
                                <div className="flex items-center gap-2">
                                    <GitBranch className="h-4 w-4 text-purple-400" />
                                    <span className="font-mono text-sm">feat/autonomous</span>
                                </div>
                                <Badge variant="secondary">Active</Badge>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-sm uppercase text-muted-foreground">Autonomy Level</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className={`flex items-center gap-2 ${normalizedAutonomyLevel === 'high' ? 'text-green-400' : 'text-yellow-400'}`}>
                                <Shield className="h-5 w-5" />
                                <span className="font-bold uppercase">{normalizedAutonomyLevel}</span>
                            </div>
                            <p className="text-xs text-muted-foreground mt-2">
                                Director is authorized to recruit squads and perform deep research without explicit approval.
                            </p>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
