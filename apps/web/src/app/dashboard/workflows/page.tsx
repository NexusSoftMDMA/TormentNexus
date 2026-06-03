'use client';

import React, { useState } from 'react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Badge } from "@tormentnexus/ui";
import { WorkflowVisualizer } from "@tormentnexus/ui";
import { trpc } from '@/utils/trpc';
import { Play, Pause, RotateCcw, Activity } from 'lucide-react';
import { toast } from 'sonner';

interface WorkflowListItem {
    id: string;
    name: string;
}

interface WorkflowExecution {
    id: string;
    status: string;
    currentNode: string;
}

interface WorkflowGraphData {
    nodes: unknown[];
    edges: unknown[];
}

interface StartWorkflowResult {
    id: string;
}

function normalizeWorkflows(value: unknown): WorkflowListItem[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value
        .filter((item): item is { id: string; name?: string } => {
            return typeof item === 'object' && item !== null && typeof (item as { id?: unknown }).id === 'string';
        })
        .map((item) => ({
            id: item.id,
            name: typeof item.name === 'string' ? item.name : item.id,
        }));
}

function normalizeExecutions(value: unknown): WorkflowExecution[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value
        .filter((item): item is { id: string; status?: string; currentNode?: string } => {
            return typeof item === 'object' && item !== null && typeof (item as { id?: unknown }).id === 'string';
        })
        .map((item) => ({
            id: item.id,
            status: typeof item.status === 'string' ? item.status : 'unknown',
            currentNode: typeof item.currentNode === 'string' ? item.currentNode : 'unknown',
        }));
}

function normalizeGraphData(value: unknown): WorkflowGraphData | null {
    if (typeof value !== 'object' || value === null) {
        return null;
    }

    const nodes = (value as { nodes?: unknown }).nodes;
    const edges = (value as { edges?: unknown }).edges;

    if (!Array.isArray(nodes) || !Array.isArray(edges)) {
        return null;
    }

    return { nodes, edges };
}

function getStartResultId(value: unknown): string | null {
    if (typeof value !== 'object' || value === null) {
        return null;
    }

    const id = (value as { id?: unknown }).id;
    return typeof id === 'string' ? id : null;
}

export default function WorkflowsPage() {
    const [selectedWorkflowId, setSelectedWorkflowId] = useState<string | null>(null);
    const [activeExecutionId, setActiveExecutionId] = useState<string | null>(null);

    // 1. List Workflows (Mocked for now as engine returns empty)
    // Actually, let's try to fetch graph for hardcoded IDs if list is empty?
    // Or assume we have some.
    // Let's rely on trpc.workflow.list
    const { data: workflows } = trpc.workflow.list.useQuery();

    // 2. Get Graph
    const { data: graphData } = trpc.workflow.getGraph.useQuery(
        { workflowId: selectedWorkflowId! },
        { enabled: !!selectedWorkflowId }
    );

    // 3. List Executions
    const { data: executions, refetch: refetchExecutions } = trpc.workflow.listExecutions.useQuery(undefined, {
        refetchInterval: 2000 // Poll for updates
    });

    const workflowList = normalizeWorkflows(workflows);
    const executionList = normalizeExecutions(executions);
    const normalizedGraphData = normalizeGraphData(graphData);

    const startMutation = trpc.workflow.start.useMutation({
        onSuccess: (data) => {
            const executionId = getStartResultId(data);
            if (!executionId) {
                toast.error("Workflow started, but execution ID was missing in response.");
                return;
            }

            toast.success("Workflow started: " + executionId);
            setActiveExecutionId(executionId);
            refetchExecutions();
        }
    });

    const resumeMutation = trpc.workflow.resume.useMutation();
    const pauseMutation = trpc.workflow.pause.useMutation();

    // Determine active node from active execution
    const activeExecution = executionList.find((execution) => execution.id === activeExecutionId);

    return (
        <div className="p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Workflows</h1>
                    <p className="text-muted-foreground">Manage and visualize agentic workflows.</p>
                </div>
                <div className="flex gap-2">
                    {/* Toolbar */}
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                {/* Sidebar: Workflows List */}
                <Card className="md:col-span-1">
                    <CardHeader>
                        <CardTitle>Library</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2">
                        {workflowList.length === 0 && (
                            <div className="text-sm text-muted-foreground">No workflows found.</div>
                        )}
                        {workflowList.map((workflow) => (
                            <Button
                                key={workflow.id}
                                variant={selectedWorkflowId === workflow.id ? "secondary" : "ghost"}
                                className="w-full justify-start"
                                onClick={() => setSelectedWorkflowId(workflow.id)}
                            >
                                <Activity className="w-4 h-4 mr-2" />
                                {workflow.name}
                            </Button>
                        ))}

                        {/* Default / Test Item */}
                        <Button
                            variant={selectedWorkflowId === 'test-workflow' ? "secondary" : "ghost"}
                            className="w-full justify-start"
                            onClick={() => setSelectedWorkflowId('test-workflow')}
                        >
                            <Activity className="w-4 h-4 mr-2" />
                            Test Workflow
                        </Button>
                    </CardContent>
                </Card>

                {/* Main: Visualization & Controls */}
                <div className="md:col-span-3 space-y-6">
                    <Card className="h-[600px] flex flex-col">
                        <CardHeader className="flex flex-row items-center justify-between">
                            <div>
                                <CardTitle>Visualizer</CardTitle>
                                <CardDescription>{selectedWorkflowId || "Select a workflow"}</CardDescription>
                            </div>
                            <div className="flex gap-2">
                                <Button
                                    size="sm"
                                    onClick={() => selectedWorkflowId && startMutation.mutate({ workflowId: selectedWorkflowId })}
                                    disabled={!selectedWorkflowId}
                                >
                                    <Play className="w-4 h-4 mr-2" /> Run
                                </Button>
                            </div>
                        </CardHeader>
                        <CardContent className="flex-1 p-0 relative">
                            {normalizedGraphData ? (
                                <WorkflowVisualizer
                                    data={normalizedGraphData}
                                    activeNodeId={activeExecution?.currentNode}
                                    className="h-full border-0"
                                />
                            ) : (
                                <div className="flex items-center justify-center h-full text-muted-foreground">
                                    No graph data available.
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Active Executions */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Active Executions</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-2">
                                {executionList.map((exec) => (
                                    <div key={exec.id} className="flex items-center justify-between p-2 border rounded hover:bg-muted/50 cursor-pointer" onClick={() => setActiveExecutionId(exec.id)}>
                                        <div className="flex items-center gap-2">
                                            <Badge variant={exec.status === 'running' ? 'default' : 'secondary'}>
                                                {exec.status}
                                            </Badge>
                                            <span className="font-mono text-sm">{exec.id}</span>
                                            <span className="text-sm text-muted-foreground">Node: {exec.currentNode}</span>
                                        </div>
                                        <div className="flex gap-1">
                                            {exec.status === 'running' && (
                                                <Button size="icon" variant="ghost" onClick={(e) => { e.stopPropagation(); pauseMutation.mutate({ executionId: exec.id }); }}>
                                                    <Pause className="w-4 h-4" />
                                                </Button>
                                            )}
                                            {exec.status === 'paused' && (
                                                <Button size="icon" variant="ghost" onClick={(e) => { e.stopPropagation(); resumeMutation.mutate({ executionId: exec.id }); }}>
                                                    <Play className="w-4 h-4" />
                                                </Button>
                                            )}
                                        </div>
                                    </div>
                                ))}
                                {executionList.length === 0 && <div className="text-sm text-muted-foreground">No active executions.</div>}
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
