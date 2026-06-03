'use client';

import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Input } from "@tormentnexus/ui";
import { Badge } from "@tormentnexus/ui";
import { ScrollArea } from "@tormentnexus/ui";
import { useEffect, useState } from "react";
import { Loader2, Search, BookOpen, GitBranch, ExternalLink, Network } from "lucide-react";
import { trpc } from '@/utils/trpc';

interface ResearchNode {
    topic: string;
    summary: string;
    sources: { title: string, url: string }[];
    relatedTopics: string[];
    subTopics?: ResearchNode[];
}

export default function ResearchPage() {
    const [topic, setTopic] = useState("");
    const [depth, setDepth] = useState(2);
    const [depthInput, setDepthInput] = useState("2");
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<ResearchNode | null>(null);
    const [queueMessage, setQueueMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
    const [lastQueueActionAt, setLastQueueActionAt] = useState<string | null>(null);

    const conductMutation = trpc.research.conduct.useMutation();
    const queueQuery = trpc.research.ingestionQueue.useQuery(undefined, {
        refetchInterval: 10000,
    });
    const retryMutation = trpc.research.retryFailed.useMutation({
        onSuccess: () => {
            setQueueMessage({ type: 'success', text: 'URL moved to pending queue.' });
            setLastQueueActionAt(new Date().toISOString());
            queueQuery.refetch();
        },
        onError: (error) => {
            setQueueMessage({ type: 'error', text: error.message || 'Failed to retry URL.' });
            setLastQueueActionAt(new Date().toISOString());
        }
    });
    const retryAllMutation = trpc.research.retryAllFailed.useMutation({
        onSuccess: (data) => {
            setQueueMessage({ type: 'success', text: data.message || 'Failed URLs moved to pending queue.' });
            setLastQueueActionAt(new Date().toISOString());
            queueQuery.refetch();
        },
        onError: (error) => {
            setQueueMessage({ type: 'error', text: error.message || 'Failed to retry all URLs.' });
            setLastQueueActionAt(new Date().toISOString());
        }
    });

    useEffect(() => {
        if (!queueMessage) return;
        const timer = window.setTimeout(() => {
            setQueueMessage(null);
        }, 5000);

        return () => window.clearTimeout(timer);
    }, [queueMessage]);

    const handleResearch = async () => {
        if (!topic) return;
        const parsedDepth = Number.parseInt(depthInput, 10);
        const normalizedDepth = Number.isFinite(parsedDepth)
            ? Math.min(5, Math.max(1, parsedDepth))
            : 2;

        setDepth(normalizedDepth);
        setDepthInput(String(normalizedDepth));
        setLoading(true);
        setResult(null);

        try {
            console.log(`Starting research: ${topic} (Depth: ${normalizedDepth})`);
            const response = await conductMutation.mutateAsync({ topic, depth: normalizedDepth });

            // Assuming response.report matches the ResearchNode structure loosely or is text.
            // If report is a string, we might need to parse it or display it simply.
            // For now, let's treat the root result as the node.
            // Adjust based on actual default return type of ResearchService if needed.
            if (response.report) {
                setResult(response.report as unknown as ResearchNode);
            }
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    const renderTree = (node: ResearchNode, level: number = 0) => {
        if (!node) return null;
        return (
            <div key={node.topic} className={`ml-${level * 4} border-l-2 border-indigo-500/30 pl-4 mb-4`}>
                <div className="flex items-center gap-2 mb-2">
                    <Badge variant="outline" className="border-indigo-500/50 text-indigo-400">
                        {level === 0 ? "ROOT" : `DEPTH ${level}`}
                    </Badge>
                    <h3 className="font-bold text-lg">{node.topic}</h3>
                </div>
                <p className="text-sm text-muted-foreground mb-3 leading-relaxed bg-muted/20 p-3 rounded-md">
                    {node.summary}
                </p>

                {node.sources && node.sources.length > 0 && (
                    <div className="flex flex-wrap gap-2 mb-3">
                        {node.sources.slice(0, 3).map((s, i) => (
                            <a key={i} href={s.url} target="_blank" rel="noopener noreferrer"
                                className="text-xs flex items-center gap-1 text-blue-400 hover:text-blue-300 bg-blue-950/30 px-2 py-1 rounded border border-blue-500/20 transition-colors">
                                <ExternalLink className="h-3 w-3" />
                                {s.title || 'Source'}
                            </a>
                        ))}
                    </div>
                )}

                {node.subTopics && node.subTopics.length > 0 && (
                    <div className="mt-4">
                        {node.subTopics.map(sub => renderTree(sub, level + 1))}
                    </div>
                )}
            </div>
        );
    };

    return (
        <div className="container mx-auto p-6 space-y-6 max-w-7xl">
            <div className="flex flex-col gap-4 border-b pb-6">
                <div className="flex items-center gap-3">
                    <div className="h-12 w-12 bg-cyan-900/20 rounded-lg flex items-center justify-center border border-cyan-500/30">
                        <Network className="h-6 w-6 text-cyan-400" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-cyan-400 to-blue-500 bg-clip-text text-transparent">
                            Deep Research
                        </h1>
                        <p className="text-muted-foreground">Recursive Knowledge Explorer & Graph Builder</p>
                    </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                    <div className="rounded-md border border-emerald-500/30 bg-emerald-950/20 px-3 py-2">
                        <div className="text-xs uppercase tracking-wide text-emerald-300/80">Processed</div>
                        <div className="text-2xl font-semibold text-emerald-300">{queueQuery.data?.totals.processed ?? 0}</div>
                    </div>
                    <div className="rounded-md border border-amber-500/30 bg-amber-950/20 px-3 py-2">
                        <div className="text-xs uppercase tracking-wide text-amber-300/80">Pending</div>
                        <div className="text-2xl font-semibold text-amber-300">{queueQuery.data?.totals.pending ?? 0}</div>
                    </div>
                    <div className="rounded-md border border-rose-500/30 bg-rose-950/20 px-3 py-2">
                        <div className="text-xs uppercase tracking-wide text-rose-300/80">Failed</div>
                        <div className="text-2xl font-semibold text-rose-300">{queueQuery.data?.totals.failed ?? 0}</div>
                    </div>
                </div>

                <div className="flex gap-4 items-end bg-muted/20 p-4 rounded-lg border border-border/50">
                    <div className="flex-1 space-y-2">
                        <label className="text-sm font-medium">Research Topic</label>
                        <Input
                            placeholder="e.g. Impact of Quantum Computing on Cryptography..."
                            value={topic}
                            onChange={(e) => setTopic(e.target.value)}
                            className="bg-background"
                        />
                    </div>
                    <div className="w-24 space-y-2">
                        <label className="text-sm font-medium">Depth</label>
                        <Input
                            type="number"
                            min={1}
                            max={5}
                            value={depthInput}
                            onChange={(e) => {
                                const nextValue = e.target.value;
                                setDepthInput(nextValue);

                                const parsed = Number.parseInt(nextValue, 10);
                                if (Number.isFinite(parsed)) {
                                    setDepth(Math.min(5, Math.max(1, parsed)));
                                }
                            }}
                            className="bg-background"
                        />
                    </div>
                    <Button onClick={handleResearch} disabled={loading || !topic} className="bg-cyan-600 hover:bg-cyan-700 w-40">
                        {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Search className="mr-2 h-4 w-4" />}
                        Start Research
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <Card className="lg:col-span-2 min-h-[500px]">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <BookOpen className="h-5 w-5 text-cyan-500" />
                            Research Report
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        {loading ? (
                            <div className="flex flex-col items-center justify-center h-64 text-muted-foreground animate-pulse">
                                <Search className="h-12 w-12 mb-4 opacity-50" />
                                <p>Exploring deep knowledge tree...</p>
                                <p className="text-xs mt-2">This may take a moment.</p>
                            </div>
                        ) : result ? (
                            <ScrollArea className="h-[600px] pr-4">
                                {renderTree(result)}
                            </ScrollArea>
                        ) : (
                            <div className="flex flex-col items-center justify-center h-64 text-muted-foreground border-2 border-dashed rounded-lg">
                                <Network className="h-12 w-12 mb-4 opacity-20" />
                                <p>Enter a topic above to begin a deep dive.</p>
                            </div>
                        )}
                    </CardContent>
                </Card>

                <div className="space-y-6">
                    <Card>
                        <CardHeader>
                            <div className="flex items-center justify-between gap-2">
                                <CardTitle className="text-sm font-medium uppercase tracking-wider text-muted-foreground">Ingestion Failures</CardTitle>
                                <Button
                                    size="sm"
                                    variant="outline"
                                    className="h-7 text-xs"
                                    disabled={(queueQuery.data?.queue.failed.length ?? 0) === 0 || retryAllMutation.isPending}
                                    onClick={() => retryAllMutation.mutate()}
                                >
                                    {retryAllMutation.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : 'Retry All'}
                                </Button>
                            </div>
                            {queueMessage ? (
                                <div
                                    className={`mt-2 rounded-md border px-2 py-1 text-xs ${queueMessage.type === 'success'
                                            ? 'border-emerald-500/30 bg-emerald-950/20 text-emerald-300'
                                            : 'border-rose-500/30 bg-rose-950/20 text-rose-300'
                                        }`}
                                >
                                    {queueMessage.text}
                                </div>
                            ) : null}
                        </CardHeader>
                        <CardContent>
                            {queueQuery.isLoading ? (
                                <div className="text-sm text-muted-foreground flex items-center gap-2 py-4">
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                    Loading queue...
                                </div>
                            ) : (queueQuery.data?.queue.failed.length ?? 0) === 0 ? (
                                <div className="text-sm text-muted-foreground text-center py-4">
                                    No failed URLs. Queue is healthy.
                                </div>
                            ) : (
                                <ScrollArea className="h-64 pr-2">
                                    <div className="space-y-3">
                                        {queueQuery.data?.queue.failed.slice(0, 20).map((item) => (
                                            <div key={item.url} className="rounded-md border border-border/50 p-2 bg-muted/20">
                                                <div className="text-xs font-medium text-foreground truncate" title={item.name}>{item.name}</div>
                                                <div className="text-[11px] text-muted-foreground truncate" title={item.url}>{item.url}</div>
                                                <div className="text-[11px] text-rose-300 mt-1 line-clamp-2">{item.error}</div>
                                                <div className="flex items-center justify-between mt-2">
                                                    <Badge variant="outline" className="text-[10px]">attempts: {item.attempts}</Badge>
                                                    <Button
                                                        size="sm"
                                                        variant="outline"
                                                        className="h-7 text-xs"
                                                        disabled={retryMutation.isPending}
                                                        onClick={() => retryMutation.mutate({ url: item.url })}
                                                    >
                                                        {retryMutation.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : 'Retry'}
                                                    </Button>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </ScrollArea>
                            )}
                        </CardContent>
                    </Card>

                    <Card className="bg-blue-950/10 border-blue-900/30">
                        <CardHeader>
                            <CardTitle className="text-sm font-medium text-blue-400">Queue Status</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-2">
                            <div className="flex justify-between text-sm">
                                <span>Indexer Sync</span>
                                <Badge variant="outline" className="text-green-500 border-green-900/50">READY</Badge>
                            </div>
                            <div className="flex justify-between text-xs text-muted-foreground">
                                <span>Last queue action</span>
                                <span>{lastQueueActionAt ? new Date(lastQueueActionAt).toLocaleTimeString() : '—'}</span>
                            </div>
                            <div className="flex justify-between text-xs text-muted-foreground">
                                <span>Last refresh</span>
                                <span>{queueQuery.data?.updatedAt ? new Date(queueQuery.data.updatedAt).toLocaleTimeString() : '—'}</span>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
