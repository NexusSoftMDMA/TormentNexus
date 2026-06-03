"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Plus, Box, Shield, Trash2, Edit2 } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

export default function NamespacesDashboard() {
    const { data: namespaces, isLoading, refetch } = trpc.namespaces.list.useQuery();
    const [isCreateOpen, setIsCreateOpen] = useState(false);

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Namespaces</h1>
                    <p className="text-zinc-500">
                        Isolate and organize MCP servers into logical groups
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button onClick={() => setIsCreateOpen(!isCreateOpen)} className="bg-blue-600 hover:bg-blue-500">
                        <Plus className="mr-2 h-4 w-4" /> Create Namespace
                    </Button>
                </div>
            </div>

            {isCreateOpen && (
                <CreateNamespaceForm onSuccess={() => { setIsCreateOpen(false); refetch(); }} />
            )}

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 flex justify-center p-12">
                        <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                    </div>
                ) : (namespaces?.length ?? 0) === 0 ? (
                    <div className="col-span-3 text-center p-12 text-zinc-500 bg-zinc-900/50 rounded-lg border border-zinc-800 border-dashed">
                        <Box className="h-12 w-12 mx-auto mb-4 opacity-30" />
                        <p className="text-lg font-medium">No Namespaces Defined</p>
                        <p className="text-sm mt-1">Create a namespace to group your tools and servers.</p>
                    </div>
                ) : namespaces?.map((ns: any) => (
                    <NamespaceCard key={ns.uuid} namespace={ns} onUpdate={refetch} />
                ))}
            </div>
        </div>
    );
}

function NamespaceCard({ namespace, onUpdate }: { namespace: any; onUpdate: () => void }) {
    const deleteMutation = trpc.namespaces.delete.useMutation({
        onSuccess: () => {
            toast.success("Namespace deleted");
            onUpdate();
        },
        onError: (err) => {
            toast.error(`Failed to delete: ${err.message}`);
        }
    });

    return (
        <Card className="bg-zinc-900 border-zinc-800 hover:border-zinc-700 transition-colors group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-lg font-medium text-zinc-200 flex items-center gap-2">
                    <Shield className="h-5 w-5 text-purple-500" />
                    {namespace.name}
                </CardTitle>
                <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            if (confirm(`Delete namespace "${namespace.name}"?`)) {
                                deleteMutation.mutate({ uuid: namespace.uuid });
                            }
                        }}
                        className="text-zinc-600 hover:text-red-400 transition-colors"
                    >
                        <Trash2 className="h-4 w-4" />
                    </button>
                </div>
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
                    <p className="text-sm text-zinc-400 min-h-[40px]">
                        {namespace.description || "No description provided."}
                    </p>
                    {/* Placeholder for stats - could add server count if available in list response */}
                    <div className="flex gap-2 text-xs">
                        <span className="bg-zinc-800 px-2 py-1 rounded text-zinc-400">
                            ID: <span className="font-mono">{namespace.uuid.slice(0, 8)}...</span>
                        </span>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function CreateNamespaceForm({ onSuccess }: { onSuccess: () => void }) {
    const [formData, setFormData] = useState({
        name: '',
        description: '',
    });

    const createMutation = trpc.namespaces.create.useMutation({
        onSuccess: () => {
            toast.success("Namespace created");
            onSuccess();
        },
        onError: (err) => {
            toast.error(`Error creating namespace: ${err.message}`);
        }
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        createMutation.mutate(formData);
    };

    return (
        <Card className="bg-zinc-900 border-zinc-700 mb-8 border-l-4 border-l-purple-600 shadow-xl">
            <CardContent className="pt-6">
                <div className="flex justify-between items-start mb-6">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium bg-purple-500/10 text-purple-400 border border-purple-500/20">
                        <Plus className="h-3 w-3" /> New Namespace
                    </div>
                    <Button variant="ghost" size="sm" onClick={onSuccess} className="text-zinc-500 hover:text-white h-6 w-6 p-0 rounded-full">
                        X
                    </Button>
                </div>

                <form onSubmit={handleSubmit} className="space-y-5">
                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Name</label>
                        <input
                            required
                            value={formData.name}
                            onChange={e => setFormData({ ...formData, name: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-purple-500 outline-none"
                            placeholder="e.g. production-tools"
                        />
                    </div>
                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Description</label>
                        <textarea
                            value={formData.description}
                            onChange={e => setFormData({ ...formData, description: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-purple-500 outline-none h-20"
                            placeholder="Optional description..."
                        />
                    </div>

                    <div className="flex justify-end pt-2">
                        <Button type="submit" disabled={createMutation.isPending} className="bg-purple-600 hover:bg-purple-500 text-white font-medium">
                            {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Create Namespace
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}
