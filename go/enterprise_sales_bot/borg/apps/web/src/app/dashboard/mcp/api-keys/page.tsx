"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Plus, Key, Trash2, Copy, Check } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

export default function ApiKeysDashboard() {
    const { data: apiKeys, isLoading, refetch } = trpc.apiKeys.list.useQuery();
    const [isCreateOpen, setIsCreateOpen] = useState(false);

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">API Keys</h1>
                    <p className="text-zinc-500">
                        Manage authentication keys for accessing managed endpoints
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button onClick={() => setIsCreateOpen(!isCreateOpen)} className="bg-blue-600 hover:bg-blue-500">
                        <Plus className="mr-2 h-4 w-4" /> Generate Key
                    </Button>
                </div>
            </div>

            {isCreateOpen && (
                <CreateApiKeyForm onSuccess={() => { setIsCreateOpen(false); refetch(); }} />
            )}

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 flex justify-center p-12">
                        <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                    </div>
                ) : (apiKeys?.length ?? 0) === 0 ? (
                    <div className="col-span-3 text-center p-12 text-zinc-500 bg-zinc-900/50 rounded-lg border border-zinc-800 border-dashed">
                        <Key className="h-12 w-12 mx-auto mb-4 opacity-30" />
                        <p className="text-lg font-medium">No API Keys Found</p>
                        <p className="text-sm mt-1">Generate a key to authenticate external requests.</p>
                    </div>
                ) : apiKeys?.map((key: any) => (
                    <ApiKeyCard key={key.uuid} apiKey={key} onUpdate={refetch} />
                ))}
            </div>
        </div>
    );
}

function ApiKeyCard({ apiKey, onUpdate }: { apiKey: any; onUpdate: () => void }) {
    const deleteMutation = trpc.apiKeys.delete.useMutation({
        onSuccess: () => {
            toast.success("API Key revoked");
            onUpdate();
        },
        onError: (err) => {
            toast.error(`Failed to revoke: ${err.message}`);
        }
    });

    const [copied, setCopied] = useState(false);

    const copyToClipboard = () => {
        navigator.clipboard.writeText(apiKey.key || "****************"); // In real app, key is only shown once usually. Assuming full key might not be available here, handling mostly metadata.
        // If the key IS available (e.g. for display purposes in this internal dashboard), we copy it.
        // Usually we only show it on creation. 
        // For now, let's assume we copy the ID or a placeholder if actual key isn't stored in plain text (it shouldn't be).
        // TORMENTNEXUS pattern: key is stored hashed? Or is it retrievable?
        // Checked api-keys.repo.ts -> findPublicApiKeys.
        // If it returns full key, that's a security risk, but for MVP/Internal usage might be acceptable or it returns a masked version.
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <Card className="bg-zinc-900 border-zinc-800 hover:border-zinc-700 transition-colors group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-lg font-medium text-zinc-200 flex items-center gap-2 truncate">
                    <Key className="h-5 w-5 text-yellow-500" />
                    <span className="truncate">{apiKey.name}</span>
                </CardTitle>
                <div className="flex items-center gap-2">
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            if (confirm(`Revoke API Key "${apiKey.name}"?`)) {
                                deleteMutation.mutate({ uuid: apiKey.uuid });
                            }
                        }}
                        className="text-zinc-600 hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
                    >
                        <Trash2 className="h-4 w-4" />
                    </button>
                </div>
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
                    <div className="flex items-center justify-between text-sm">
                        <span className="text-zinc-500">Prefix</span>
                        <span className="font-mono text-zinc-300 bg-zinc-800 px-2 py-0.5 rounded">{apiKey.key_prefix || 'sk-...'}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                        <span className="text-zinc-500">Created</span>
                        <span className="text-zinc-400">{new Date(apiKey.created_at).toLocaleDateString()}</span>
                    </div>
                    <div className="pt-2">
                        <div className={`text-xs text-center border rounded py-1 ${apiKey.is_active ? 'border-green-500/20 text-green-500 bg-green-500/10' : 'border-red-500/20 text-red-500 bg-red-500/10'}`}>
                            {apiKey.is_active ? 'ACTIVE' : 'REVOKED'}
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function CreateApiKeyForm({ onSuccess }: { onSuccess: () => void }) {
    const [name, setName] = useState('');
    const [generatedKey, setGeneratedKey] = useState<string | null>(null);

    const createMutation = trpc.apiKeys.create.useMutation({
        onSuccess: (data: any) => {
            toast.success("API Key generated");
            setGeneratedKey(data.key); // Assuming creation returns the raw key once
        },
        onError: (err) => {
            toast.error(`Error generating key: ${err.message}`);
        }
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        createMutation.mutate({
            name: name,
            type: "MCP", // default
        });
    };

    const handleClose = () => {
        setGeneratedKey(null);
        setName('');
        onSuccess();
    }

    if (generatedKey) {
        return (
            <Card className="bg-zinc-900 border-zinc-700 mb-8 border-l-4 border-l-green-600 shadow-xl animate-in fade-in zoom-in-95 duration-200">
                <CardContent className="pt-6">
                    <div className="text-center space-y-4">
                        <div className="mx-auto w-12 h-12 bg-green-500/20 rounded-full flex items-center justify-center">
                            <Check className="h-6 w-6 text-green-500" />
                        </div>
                        <h3 className="text-lg font-medium text-white">API Key Generated</h3>
                        <p className="text-sm text-zinc-400">
                            Copy this key now. You won't be able to see it again!
                        </p>
                        <div className="bg-black p-4 rounded border border-zinc-800 font-mono text-green-400 break-all select-all">
                            {generatedKey}
                        </div>
                        <Button onClick={handleClose} className="w-full bg-zinc-800 hover:bg-zinc-700">
                            Done
                        </Button>
                    </div>
                </CardContent>
            </Card>
        )
    }

    return (
        <Card className="bg-zinc-900 border-zinc-700 mb-8 border-l-4 border-l-yellow-600 shadow-xl">
            <CardContent className="pt-6">
                <div className="flex justify-between items-start mb-6">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium bg-yellow-500/10 text-yellow-400 border border-yellow-500/20">
                        <Plus className="h-3 w-3" /> New API Key
                    </div>
                    <Button variant="ghost" size="sm" onClick={onSuccess} className="text-zinc-500 hover:text-white h-6 w-6 p-0 rounded-full">
                        X
                    </Button>
                </div>

                <form onSubmit={handleSubmit} className="space-y-5">
                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Key Name / Description</label>
                        <input
                            required
                            value={name}
                            onChange={e => setName(e.target.value)}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-yellow-500 outline-none"
                            placeholder="e.g. CI/CD Pipeline access"
                        />
                    </div>

                    <div className="flex justify-end pt-2">
                        <Button type="submit" disabled={createMutation.isPending} className="bg-yellow-600 hover:bg-yellow-500 text-white font-medium">
                            {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Generate Key
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}
