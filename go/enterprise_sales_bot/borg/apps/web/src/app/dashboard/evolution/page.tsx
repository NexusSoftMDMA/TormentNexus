
'use client';

import React, { useState } from 'react';
import { trpc } from '@/utils/trpc';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@tormentnexus/ui';
import { Badge } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';
import { Input } from '@tormentnexus/ui';
import { Textarea } from '@tormentnexus/ui';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@tormentnexus/ui";
import { ScrollArea } from '@tormentnexus/ui';
import { RefreshCcw, Dna, FlaskConical, Play, CheckCircle, XCircle } from 'lucide-react';
import { useToast } from "@tormentnexus/ui";

export default function EvolutionPage() {
    const { toast } = useToast();
    const [prompt, setPrompt] = useState('');
    const [goal, setGoal] = useState('');
    const [task, setTask] = useState('');
    const [selectedMutation, setSelectedMutation] = useState<string | null>(null);

    const { data: status, refetch } = trpc.darwin.getStatus.useQuery();

    const { mutate: mutateIdea, isPending: isMutating } = trpc.darwin.evolve.useMutation({
        onSuccess: (data: any) => {
            toast({
                title: "Mutation Successful",
                description: `Created Variant: ${data.id}`,
                variant: "success",
            });
            refetch(); // Added refetch here to update status after mutation
        },
        onError: (error) => {
            toast({
                title: "Mutation Failed",
                description: error.message,
                variant: "destructive",
            });
        }
    });

    const { mutate: evaluateExperiment, isPending: isEvaluating } = trpc.darwin.experiment.useMutation({
        onSuccess: (data: any) => {
            toast({
                title: "Experiment Evaluated",
                description: `Experiment ${data.experimentId} processed.`,
            });
            refetch(); // Added refetch here to update status after experiment
        },
        onError: (err) => {
            toast({ title: "Experiment Failed", description: err.message, variant: "destructive" });
        }
    });

    const handleEvolve = () => {
        if (!prompt || !goal) return;
        mutateIdea({ prompt, goal });
    };

    const handleExperiment = () => {
        if (!selectedMutation || !task) return;
        evaluateExperiment({ mutationId: selectedMutation, task });
    };

    return (
        <div className="p-6 space-y-6">
            <header className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Evolution Engine</h1>
                    <p className="text-muted-foreground">The Darwin Protocol: Mutation and Natural Selection of Agents.</p>
                </div>
                <Button variant="outline" onClick={() => refetch()}><RefreshCcw className="w-4 h-4 mr-2" /> Refresh</Button>
            </header>

            <Tabs defaultValue="mutations" className="w-full">
                <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="mutations">Mutations</TabsTrigger>
                    <TabsTrigger value="experiments">Experiments</TabsTrigger>
                </TabsList>

                <TabsContent value="mutations" className="space-y-4">
                    <Card>
                        <CardHeader>
                            <CardTitle>Propose New Mutation</CardTitle>
                            <CardDescription>Use LLM to evolve a system prompt towards a specific goal.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">Original Prompt</label>
                                <Textarea
                                    placeholder="You are a helpful assistant..."
                                    className="min-h-[100px]"
                                    value={prompt}
                                    onChange={(e) => setPrompt(e.target.value)}
                                />
                            </div>
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">Evolutionary Goal</label>
                                <Input
                                    placeholder="Make it more concise and strict..."
                                    value={goal}
                                    onChange={(e) => setGoal(e.target.value)}
                                />
                            </div>
                            <Button onClick={handleEvolve} disabled={isMutating}>
                                {isMutating ? <RefreshCcw className="w-4 h-4 animate-spin mr-2" /> : <Dna className="w-4 h-4 mr-2" />}
                                Evolve
                            </Button>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Mutation History</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <ScrollArea className="h-[400px]">
                                {(status as any)?.mutations?.map((m: any) => (
                                    <div key={m.id} className="border-b p-4 last:border-0 hover:bg-muted/50 cursor-pointer" onClick={() => setSelectedMutation(m.id)}>
                                        <div className="flex justify-between items-center mb-2">
                                            <Badge variant="outline" className="font-mono">{m.id}</Badge>
                                            <span className="text-xs text-muted-foreground">{new Date(m.timestamp).toLocaleString()}</span>
                                        </div>
                                        <div className="bg-muted p-2 rounded text-xs font-mono mb-2">
                                            {m.reasoning}
                                        </div>
                                        <div className="grid grid-cols-2 gap-2 text-xs">
                                            <div className="border p-2 rounded">
                                                <div className="font-semibold mb-1">Before</div>
                                                <div className="line-clamp-3">{m.originalPrompt}</div>
                                            </div>
                                            <div className="border p-2 rounded bg-green-950/10">
                                                <div className="font-semibold mb-1">After</div>
                                                <div className="line-clamp-3">{m.mutatedPrompt}</div>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </ScrollArea>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="experiments" className="space-y-4">
                    <Card>
                        <CardHeader>
                            <CardTitle>Run A/B Experiment</CardTitle>
                            <CardDescription>Compare original vs mutated prompt on a specific task.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">Using Mutation ID</label>
                                <Input
                                    placeholder="Select a mutation from the list..."
                                    value={selectedMutation || ''}
                                    onChange={(e) => setSelectedMutation(e.target.value)}
                                />
                            </div>
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">Test Task</label>
                                <Textarea
                                    placeholder="Describe a task to test both agents..."
                                    value={task}
                                    onChange={(e) => setTask(e.target.value)}
                                />
                            </div>
                            <Button onClick={handleExperiment} disabled={isEvaluating}>
                                {isEvaluating ? <RefreshCcw className="w-4 h-4 animate-spin mr-2" /> : <FlaskConical className="w-4 h-4 mr-2" />}
                                Start Experiment
                            </Button>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Active Experiments</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <ScrollArea className="h-[400px]">
                                {(status as any)?.experiments?.map((e: any) => (
                                    <div key={e.id} className="border-b p-4 last:border-0">
                                        <div className="flex justify-between items-center mb-2">
                                            <div className="flex items-center gap-2">
                                                <Badge variant={e.status === 'COMPLETED' ? 'default' : 'secondary'}>{e.status}</Badge>
                                                <span className="font-mono text-sm">{e.id}</span>
                                            </div>
                                            {e.status === 'COMPLETED' && (
                                                <Badge variant={e.winner === 'B' ? 'default' : e.winner === 'TIE' ? 'outline' : 'secondary'}>
                                                    Winner: {e.winner}
                                                </Badge>
                                            )}
                                        </div>
                                        <div className="text-sm mb-2 font-medium">{e.task}</div>
                                        {e.status === 'COMPLETED' && (
                                            <div className="bg-muted p-3 rounded text-sm">
                                                <p className="font-semibold mb-1">Judge's Verdict:</p>
                                                <p>{e.judgeReasoning}</p>
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </ScrollArea>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
