'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { ScrollArea } from "./ui/scroll-area";
import { Search, Database, RefreshCw, HardDrive, Trash2 } from 'lucide-react';
import { toast } from "sonner";
import { trpc } from '@/utils/trpc';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "./ui/dialog";

function getContextContent(value: unknown): string {
    if (!value || typeof value !== 'object') {
        return '';
    }

    const record = value as Record<string, unknown>;
    return typeof record.content === 'string' ? record.content : '';
}

export default function MemoryPage() {
    const [searchQuery, setSearchQuery] = useState('');
    const [activeTab, setActiveTab] = useState('contexts');

    // Queries
    const search = trpc.memory.query.useQuery({ query: searchQuery }, { enabled: false });
    const contexts = trpc.memory.listContexts.useQuery(undefined, { refetchOnWindowFocus: true });
    const deleteContext = trpc.memory.deleteContext.useMutation({
        onSuccess: () => {
            contexts.refetch();
            toast.success("Context deleted");
        }
    });

    // View Modal State
    const [selectedContextId, setSelectedContextId] = useState<string | null>(null);
    const contextDetail = trpc.memory.getContext.useQuery({ id: selectedContextId! }, { enabled: !!selectedContextId });

    const handleSearch = () => {
        search.refetch();
    };

    const handleDelete = (id: string) => {
        if (confirm('Are you sure you want to delete this context?')) {
            deleteContext.mutate({ id });
        }
    };

    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-3xl font-bold">Memory Management</h1>
                <Button onClick={() => contexts.refetch()} variant="outline">
                    <RefreshCw className="mr-2 h-4 w-4" /> Refresh
                </Button>
            </div>

            <Tabs defaultValue="contexts" value={activeTab} onValueChange={setActiveTab}>
                <TabsList className="mb-4">
                    <TabsTrigger value="contexts">Exported Contexts</TabsTrigger>
                    <TabsTrigger value="search">Search DB</TabsTrigger>
                </TabsList>

                <TabsContent value="contexts">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center justify-between">
                                <span>Saved Contexts</span>
                                <Badge variant="secondary">{contexts.data?.length || 0} items</Badge>
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <ScrollArea className="h-[600px]">
                                {contexts.isLoading ? (
                                    <div className="p-10 text-center">Loading...</div>
                                ) : contexts.data?.length === 0 ? (
                                    <div className="text-center text-muted-foreground py-10">
                                        No exported contexts found. Use the "Save Context" tool to add some.
                                    </div>
                                ) : (
                                    <div className="space-y-3">
                                        {contexts.data?.map((ctx: any) => (
                                            <div key={ctx.id} className="p-4 border rounded-lg hover:bg-accent/50 transition-colors flex justify-between items-start">
                                                <div>
                                                    <div className="font-medium">{ctx.title || 'Untitled'}</div>
                                                    <div className="text-sm text-muted-foreground mb-1">
                                                        ID: {ctx.id}
                                                    </div>
                                                    <div className="flex gap-2">
                                                        <Badge variant="outline">{ctx.source}</Badge>
                                                        <span className="text-xs text-muted-foreground pt-1">
                                                            {new Date(ctx.createdAt || Date.now()).toLocaleDateString()}
                                                        </span>
                                                    </div>
                                                </div>
                                                <div className="flex gap-2">
                                                    <Button size="sm" variant="outline" onClick={() => setSelectedContextId(ctx.id)}>View</Button>
                                                    <Button size="sm" variant="destructive" onClick={() => handleDelete(ctx.id)}>
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </ScrollArea>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="search">
                    <Card>
                        <CardHeader>
                            <CardTitle>Semantic Search</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex gap-2 mb-4">
                                <Input
                                    placeholder="Search vector database..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                                />
                                <Button onClick={handleSearch} disabled={search.isFetching}>
                                    <Search className="h-4 w-4" />
                                </Button>
                            </div>

                            <ScrollArea className="h-[500px]">
                                {search.data?.length === 0 ? (
                                    <div className="text-center text-muted-foreground py-10">
                                        No matches found.
                                    </div>
                                ) : (
                                    <div className="space-y-3">
                                        {search.data?.map((m: any) => (
                                            <div key={m.id} className="p-3 border rounded-lg hover:bg-accent/50 transition-colors">
                                                <div className="text-sm font-mono bg-muted p-2 rounded mb-2 overflow-x-auto">
                                                    {m.content.substring(0, 300)}...
                                                </div>
                                                <div className="flex items-center justify-between">
                                                    <div className="text-xs text-muted-foreground">
                                                        File: {m.metadata?.file_path || m.id}
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </ScrollArea>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>

            <Dialog open={!!selectedContextId} onOpenChange={(open) => !open && setSelectedContextId(null)}>
                <DialogContent className="max-w-3xl max-h-[80vh]">
                    <DialogHeader>
                        <DialogTitle>Context Viewer</DialogTitle>
                        <DialogDescription>
                            ID: {selectedContextId}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="mt-4">
                        {contextDetail.isLoading ? (
                            <div className="flex justify-center p-8">Loading content...</div>
                        ) : contextDetail.data ? (
                            <ScrollArea className="h-[500px] w-full rounded-md border p-4 bg-muted/50">
                                <pre className="text-xs font-mono whitespace-pre-wrap break-all">
                                    {getContextContent(contextDetail.data)}
                                </pre>
                                {/* <div className="mt-4 border-t pt-4">
                            <h4 className="font-semibold mb-2">Metadata</h4>
                            <pre className="text-xs">{JSON.stringify(contextDetail.data.metadata, null, 2)}</pre>
                        </div> */}
                            </ScrollArea>
                        ) : (
                            <div className="text-center p-4 text-red-500">Failed to load context.</div>
                        )}
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
}
