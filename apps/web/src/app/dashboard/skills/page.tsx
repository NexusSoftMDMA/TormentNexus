'use client';

import React, { useState } from 'react';
import { trpc } from '@/utils/trpc';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@tormentnexus/ui';
import { Input } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';
import { Badge } from "@tormentnexus/ui";
import { Hammer, BookOpen, Terminal, CheckCircle2, Loader2, AlertCircle } from 'lucide-react';

interface SkillListItem {
    id: string;
    name: string;
    description: string;
    content: string;
    path: string;
}

function normalizeSkills(value: unknown): SkillListItem[] {
    if (!Array.isArray(value)) return [];

    return value.filter((item): item is SkillListItem => {
        if (!item || typeof item !== 'object') return false;
        const skill = item as Partial<SkillListItem>;
        return (
            typeof skill.id === 'string' &&
            typeof skill.name === 'string' &&
            typeof skill.description === 'string' &&
            typeof skill.content === 'string' &&
            typeof skill.path === 'string'
        );
    });
}

export default function SkillsPage() {
    const [topic, setTopic] = useState('');
    const [logs, setLogs] = useState<string[]>([]);
    const [status, setStatus] = useState<'idle' | 'assimilating' | 'success' | 'error'>('idle');

    // List existing skills
    const { data: skills, refetch } = trpc.skills.list.useQuery();
    const skillList = normalizeSkills(skills);

    const assimilateMutation = trpc.skills.assimilate.useMutation({
        onMutate: () => {
            setStatus('assimilating');
            setLogs(['Initiating assimilation protocol...', ' contacting Deep Research...']);
        },
        onSuccess: (data) => {
            if (data.success) {
                setStatus('success');
                setLogs(prev => [...prev, ...data.logs, `Successfully assimilated: ${data.toolName}`]);
                refetch();
            } else {
                setStatus('error');
                setLogs(prev => [...prev, ...data.logs, 'Assimilation failed.']);
            }
        },
        onError: (e) => {
            setStatus('error');
            setLogs(prev => [...prev, `RPC Error: ${e.message}`]);
        }
    });

    const handleAssimilate = (e: React.FormEvent) => {
        e.preventDefault();
        if (!topic.trim()) return;
        assimilateMutation.mutate({ topic });
    };

    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Skill Acquisition</h1>
                    <p className="text-muted-foreground">Teach the agent new capabilities by pointing it to documentation.</p>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Acquisition Panel */}
                <Card className="border-blue-900/50 bg-blue-950/10">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><BookOpen className="w-5 h-5 text-blue-400" /> New Skill</CardTitle>
                        <CardDescription>Enter a topic or tool name (e.g., "Stripe API", "cowsay CLI")</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <form onSubmit={handleAssimilate} className="flex gap-2">
                            <Input
                                placeholder="What should I learn?"
                                value={topic}
                                onChange={(e) => setTopic(e.target.value)}
                                disabled={status === 'assimilating'}
                            />
                            <Button type="submit" disabled={status === 'assimilating'}>
                                {status === 'assimilating' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Hammer className="w-4 h-4" />}
                                <span className="ml-2">Assimilate</span>
                            </Button>
                        </form>

                        {/* Logs Console */}
                        <div className="bg-black/80 rounded-md border border-zinc-800 p-4 font-mono text-sm h-[300px] flex flex-col">
                            <div className="flex items-center gap-2 border-b border-zinc-800 pb-2 mb-2 text-zinc-400">
                                <Terminal className="w-4 h-4" /> Assimilation Log
                            </div>
                            <div className="flex-1 overflow-y-auto">
                                {logs.map((log, i) => (
                                    <div key={i} className="mb-1 text-zinc-300">
                                        <span className="text-zinc-600 mr-2">{'>'}</span>{log}
                                    </div>
                                ))}
                                {status === 'assimilating' && (
                                    <div className="animate-pulse text-blue-400">Processing...</div>
                                )}
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Library Panel */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><CheckCircle2 className="w-5 h-5 text-green-400" /> Skill Library</CardTitle>
                        <CardDescription>Available capabilities</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="h-[400px] overflow-y-auto">
                            {skillList.length === 0 && <p className="text-muted-foreground italic">No custom skills loaded.</p>}
                            <div className="grid grid-cols-1 gap-2">
                                {skillList.map((skill, i) => (
                                    <div key={i} className="flex items-center justify-between p-3 rounded-md border border-zinc-800 bg-zinc-900/50">
                                        <div className="font-medium">{skill.name}</div>
                                        <Badge variant="outline">Active</Badge>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
