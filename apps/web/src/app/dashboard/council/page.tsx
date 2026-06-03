
'use client';

import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Badge } from "@tormentnexus/ui";
import { ScrollArea } from "@tormentnexus/ui";
import { useState, useEffect } from "react";
import { Loader2, MessageSquare, Gavel, User, Play, RefreshCw, ChevronRight } from "lucide-react";
import { Input } from "@tormentnexus/ui";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@tormentnexus/ui";
import { trpc } from "@/utils/trpc";
import { DebateVisualizer } from "@/components/council/DebateVisualizer";
import { normalizeCouncilSessions, type CouncilSessionRow } from './council-page-normalizers';

export default function CouncilPage() {
    const [sessions, setSessions] = useState<CouncilSessionRow[]>([]);
    const [selectedSession, setSelectedSession] = useState<CouncilSessionRow | null>(null);
    const [newTopic, setNewTopic] = useState("");

    // TRPC Queries
    const listQuery = trpc.council.listSessions.useQuery(undefined, {
        refetchInterval: 5000 // Poll every 5s for live updates
    });
    const runMutation = trpc.council.runSession.useMutation();

    useEffect(() => {
        if (listQuery.data) {
            const data = normalizeCouncilSessions(listQuery.data);
            setSessions(data);
            // Auto-select most recent if none selected
            if (!selectedSession && data.length > 0) {
                setSelectedSession(data[0]);
            }
            // Update selected session object from list if it exists (live updates)
            if (selectedSession) {
                const updated = data.find(s => s.id === selectedSession.id);
                if (updated) setSelectedSession(updated);
            }
        }
    }, [listQuery.data, selectedSession]);

    const handleCreateSession = async () => {
        if (!newTopic) return;
        try {
            await runMutation.mutateAsync({ proposal: newTopic });
            setNewTopic("");
            listQuery.refetch();
        } catch (e) {
            console.error("Failed to start session:", e);
        }
    };

    return (
        <div className="container mx-auto p-6 space-y-6 max-w-7xl">
            <div className="flex justify-between items-center border-b pb-4 mb-4">
                <div className="flex items-center gap-3">
                    <div className="h-12 w-12 bg-indigo-900/20 rounded-lg flex items-center justify-center border border-indigo-500/30">
                        <Gavel className="h-6 w-6 text-indigo-400" />
                    </div>
                    <div>
                        <div className="flex items-center gap-3">
                            <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-indigo-400 to-purple-400 bg-clip-text text-transparent">
                                The Council
                            </h1>
                            <Badge variant="outline" className="border-purple-500/30 text-purple-400 bg-purple-500/10">Labs</Badge>
                        </div>
                        <p className="text-muted-foreground text-sm mt-1">Multi-Agent Consensus & Debate System (Experimental)</p>
                    </div>
                </div>
                <div className="flex gap-2">
                    <Input
                        placeholder="Propose a topic for debate..."
                        value={newTopic}
                        onChange={(e) => setNewTopic(e.target.value)}
                        className="w-80 bg-background/50"
                    />
                    <Button onClick={handleCreateSession} className="bg-indigo-600 hover:bg-indigo-700">
                        <Play className="mr-2 h-4 w-4" />
                        Convene Session
                    </Button>
                </div>
            </div>

            <Tabs defaultValue="sessions" className="w-full">
                <TabsList className="bg-background/20 border mb-4">
                    <TabsTrigger value="sessions">Active Debates</TabsTrigger>
                    <TabsTrigger value="history">History</TabsTrigger>
                    <TabsTrigger value="members">Council Members</TabsTrigger>
                </TabsList>

                <TabsContent value="sessions" className="space-y-4">
                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                        {/* Session List */}
                        <div className="space-y-4">
                            <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Active & Recent</h3>
                            {listQuery.isPending ? (
                                <div className="flex justify-center p-8">
                                    <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                                </div>
                            ) : sessions.length === 0 ? (
                                <div className="text-center p-8 border border-dashed rounded-lg">
                                    <p className="text-muted-foreground">No sessions found.</p>
                                </div>
                            ) : (
                                <div className="space-y-2">
                                    {sessions.map(session => (
                                        <div
                                            key={session.id}
                                            onClick={() => setSelectedSession(session)}
                                            className={`p-4 rounded-lg border cursor-pointer transition-colors ${selectedSession?.id === session.id
                                                ? 'bg-indigo-500/10 border-indigo-500/50'
                                                : 'bg-card hover:bg-accent/50 border-border'
                                                }`}
                                        >
                                            <div className="flex justify-between items-start mb-2">
                                                <Badge variant={session.status === 'active' ? 'default' : 'secondary'} className={session.status === 'active' ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30' : ''}>
                                                    {session.status}
                                                </Badge>
                                                <span className="text-xs text-muted-foreground">
                                                    {new Date(session.createdAt).toLocaleTimeString()}
                                                </span>
                                            </div>
                                            <h4 className="font-medium text-sm line-clamp-2 mb-2">{session.topic}</h4>
                                            <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                                <MessageSquare className="h-3 w-3" />
                                                <span>{session.opinions.length} Opinions</span>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Visualizer Area */}
                        <div className="lg:col-span-2">
                            {selectedSession ? (
                                <DebateVisualizer
                                    topic={selectedSession.topic}
                                    transcripts={selectedSession.opinions.map(o => ({
                                        speaker: o.agentId, // Assuming agentId is name for now
                                        text: o.content,
                                        round: o.round,
                                        vote: undefined // logic to map votes to opinions is complex, maybe just list votes?
                                    }))}
                                    config={{
                                        rounds: selectedSession.round,
                                        status: selectedSession.status,
                                        result: selectedSession.status === 'concluded' ? "Session Concluded" : undefined
                                    }}
                                />
                            ) : (
                                <div className="h-[600px] border border-dashed rounded-lg flex items-center justify-center text-muted-foreground">
                                    Select a session to view details
                                </div>
                            )}
                        </div>
                    </div>
                </TabsContent>


                <TabsContent value="members">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {/* Architect */}
                        <Card className="border-blue-500/20 bg-gradient-to-br from-blue-950/10 to-transparent">
                            <CardHeader className="flex flex-row items-center gap-4 pb-2">
                                <div className="h-12 w-12 rounded-full bg-blue-900/30 border border-blue-500/40 flex items-center justify-center text-2xl">
                                    🏛️
                                </div>
                                <div>
                                    <CardTitle className="text-lg text-blue-300">The Architect</CardTitle>
                                    <p className="text-xs text-muted-foreground">System Design & Architecture</p>
                                </div>
                                <Badge variant="outline" className="ml-auto text-green-500 border-green-900/50">Active</Badge>
                            </CardHeader>
                            <CardContent className="space-y-3 pt-2">
                                <p className="text-sm text-muted-foreground leading-relaxed">
                                    Evaluates proposals from a systems architecture perspective. Focuses on scalability,
                                    modularity, and long-term maintainability. Advocates for clean abstractions and
                                    separation of concerns.
                                </p>
                                <div className="flex gap-2 flex-wrap">
                                    <Badge variant="secondary" className="text-xs">Architecture</Badge>
                                    <Badge variant="secondary" className="text-xs">Scalability</Badge>
                                    <Badge variant="secondary" className="text-xs">Design Patterns</Badge>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Critic */}
                        <Card className="border-red-500/20 bg-gradient-to-br from-red-950/10 to-transparent">
                            <CardHeader className="flex flex-row items-center gap-4 pb-2">
                                <div className="h-12 w-12 rounded-full bg-red-900/30 border border-red-500/40 flex items-center justify-center text-2xl">
                                    🔍
                                </div>
                                <div>
                                    <CardTitle className="text-lg text-red-300">The Critic</CardTitle>
                                    <p className="text-xs text-muted-foreground">Quality Assurance & Risk Analysis</p>
                                </div>
                                <Badge variant="outline" className="ml-auto text-green-500 border-green-900/50">Active</Badge>
                            </CardHeader>
                            <CardContent className="space-y-3 pt-2">
                                <p className="text-sm text-muted-foreground leading-relaxed">
                                    Stress-tests proposals for edge cases, security vulnerabilities, and performance
                                    bottlenecks. Plays devil&apos;s advocate to ensure robustness. Identifies risks
                                    others may overlook.
                                </p>
                                <div className="flex gap-2 flex-wrap">
                                    <Badge variant="secondary" className="text-xs">Security</Badge>
                                    <Badge variant="secondary" className="text-xs">Edge Cases</Badge>
                                    <Badge variant="secondary" className="text-xs">Risk Analysis</Badge>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Product */}
                        <Card className="border-green-500/20 bg-gradient-to-br from-green-950/10 to-transparent">
                            <CardHeader className="flex flex-row items-center gap-4 pb-2">
                                <div className="h-12 w-12 rounded-full bg-green-900/30 border border-green-500/40 flex items-center justify-center text-2xl">
                                    📦
                                </div>
                                <div>
                                    <CardTitle className="text-lg text-green-300">The Product Lead</CardTitle>
                                    <p className="text-xs text-muted-foreground">User Experience & Feature Prioritization</p>
                                </div>
                                <Badge variant="outline" className="ml-auto text-green-500 border-green-900/50">Active</Badge>
                            </CardHeader>
                            <CardContent className="space-y-3 pt-2">
                                <p className="text-sm text-muted-foreground leading-relaxed">
                                    Champions the end-user perspective. Evaluates proposals based on user impact,
                                    development velocity, and alignment with the product roadmap. Prioritizes features
                                    that deliver maximum value.
                                </p>
                                <div className="flex gap-2 flex-wrap">
                                    <Badge variant="secondary" className="text-xs">UX</Badge>
                                    <Badge variant="secondary" className="text-xs">Roadmap</Badge>
                                    <Badge variant="secondary" className="text-xs">Prioritization</Badge>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Meta-Architect */}
                        <Card className="border-purple-500/20 bg-gradient-to-br from-purple-950/10 to-transparent">
                            <CardHeader className="flex flex-row items-center gap-4 pb-2">
                                <div className="h-12 w-12 rounded-full bg-purple-900/30 border border-purple-500/40 flex items-center justify-center text-2xl">
                                    🧠
                                </div>
                                <div>
                                    <CardTitle className="text-lg text-purple-300">The Meta-Architect</CardTitle>
                                    <p className="text-xs text-muted-foreground">Cross-System Integration & Strategy</p>
                                </div>
                                <Badge variant="outline" className="ml-auto text-yellow-500 border-yellow-900/50">Reserved</Badge>
                            </CardHeader>
                            <CardContent className="space-y-3 pt-2">
                                <p className="text-sm text-muted-foreground leading-relaxed">
                                    Activated for cross-cutting architectural decisions that span multiple subsystems.
                                    Synthesizes perspectives from all other council members and resolves conflicts at
                                    the highest abstraction level.
                                </p>
                                <div className="flex gap-2 flex-wrap">
                                    <Badge variant="secondary" className="text-xs">Strategy</Badge>
                                    <Badge variant="secondary" className="text-xs">Integration</Badge>
                                    <Badge variant="secondary" className="text-xs">Synthesis</Badge>
                                </div>
                            </CardContent>
                        </Card>
                    </div>

                    {/* Consensus Modes */}
                    <Card className="mt-6 border-indigo-900/30">
                        <CardHeader>
                            <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                                Available Consensus Modes
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                <div className="p-3 bg-muted/20 rounded-lg border text-center">
                                    <div className="font-bold text-sm mb-1">Majority</div>
                                    <p className="text-xs text-muted-foreground">&gt;50% approval</p>
                                </div>
                                <div className="p-3 bg-muted/20 rounded-lg border text-center">
                                    <div className="font-bold text-sm mb-1">Supermajority</div>
                                    <p className="text-xs text-muted-foreground">&gt;66% approval</p>
                                </div>
                                <div className="p-3 bg-muted/20 rounded-lg border text-center">
                                    <div className="font-bold text-sm mb-1">Unanimous</div>
                                    <p className="text-xs text-muted-foreground">100% approval</p>
                                </div>
                                <div className="p-3 bg-muted/20 rounded-lg border text-center">
                                    <div className="font-bold text-sm mb-1">Single Approval</div>
                                    <p className="text-xs text-muted-foreground">Any 1 member</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs >
        </div >
    );
}
