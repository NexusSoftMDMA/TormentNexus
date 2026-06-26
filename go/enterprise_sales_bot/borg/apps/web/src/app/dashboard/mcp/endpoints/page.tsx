"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Plus, Globe, Trash2, Key, Shield } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

export default function EndpointsDashboard() {
    const { data: endpoints, isLoading, refetch } = trpc.endpoints.list.useQuery();
    const { data: namespaces } = trpc.namespaces.list.useQuery(); // Needed for creation
    const [isCreateOpen, setIsCreateOpen] = useState(false);

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Endpoints</h1>
                    <p className="text-zinc-500">
                        Expose namespaces as public or private API endpoints
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button onClick={() => setIsCreateOpen(!isCreateOpen)} className="bg-blue-600 hover:bg-blue-500">
                        <Plus className="mr-2 h-4 w-4" /> Create Endpoint
                    </Button>
                </div>
            </div>

            {isCreateOpen && (
                <CreateEndpointForm
                    namespaces={namespaces || []}
                    onSuccess={() => { setIsCreateOpen(false); refetch(); }}
                />
            )}

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 flex justify-center p-12">
                        <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                    </div>
                ) : (endpoints?.length ?? 0) === 0 ? (
                    <div className="col-span-3 text-center p-12 text-zinc-500 bg-zinc-900/50 rounded-lg border border-zinc-800 border-dashed">
                        <Globe className="h-12 w-12 mx-auto mb-4 opacity-30" />
                        <p className="text-lg font-medium">No Endpoints Configured</p>
                        <p className="text-sm mt-1">Expose a namespace to allow external access.</p>
                    </div>
                ) : endpoints?.map((ep: any) => (
                    <EndpointCard key={ep.uuid} endpoint={ep} onUpdate={refetch} />
                ))}
            </div>
        </div>
    );
}

function EndpointCard({ endpoint, onUpdate }: { endpoint: any; onUpdate: () => void }) {
    const deleteMutation = trpc.endpoints.delete.useMutation({
        onSuccess: () => {
            toast.success("Endpoint deleted");
            onUpdate();
        },
        onError: (err) => {
            toast.error(`Failed to delete: ${err.message}`);
        }
    });

    return (
        <Card className="bg-zinc-900 border-zinc-800 hover:border-zinc-700 transition-colors group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-lg font-medium text-zinc-200 flex items-center gap-2 truncate">
                    <Globe className="h-5 w-5 text-green-500" />
                    <span className="truncate">/{endpoint.path}</span>
                </CardTitle>
                <div className="flex items-center gap-2">
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            if (confirm(`Delete endpoint "${endpoint.path}"?`)) {
                                deleteMutation.mutate({ uuid: endpoint.uuid });
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
                        <span className="text-zinc-500">Method</span>
                        <span className="bg-zinc-800 px-2 py-0.5 rounded text-zinc-300 font-mono text-xs">{endpoint.method}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                        <span className="text-zinc-500">Namespace</span>
                        <span className="text-purple-400 truncate max-w-[150px]">{endpoint.namespace_uuid}</span>
                    </div>

                    <div className="grid grid-cols-2 gap-2 text-xs mt-2">
                        <div className={`p-2 rounded flex flex-col items-center ${endpoint.auth_config?.enabled !== false ? 'bg-green-500/10 text-green-400 border border-green-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'}`}>
                            <Shield className="h-3 w-3 mb-1" />
                            <span>{endpoint.auth_config?.enabled !== false ? 'Auth On' : 'No Auth'}</span>
                        </div>
                        <div className="bg-zinc-800/50 p-2 rounded flex flex-col items-center text-zinc-400">
                            <Key className="h-3 w-3 mb-1" />
                            <span>{endpoint.auth_config?.type || 'N/A'}</span>
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function CreateEndpointForm({ namespaces, onSuccess }: { namespaces: any[]; onSuccess: () => void }) {
    const [formData, setFormData] = useState({
        path: '',
        method: 'POST',
        namespaceUuid: namespaces[0]?.uuid || '',
        authType: 'api-key', // default
    });

    const createMutation = trpc.endpoints.create.useMutation({
        onSuccess: () => {
            toast.success("Endpoint created");
            onSuccess();
        },
        onError: (err) => {
            toast.error(`Error creating endpoint: ${err.message}`);
        }
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        createMutation.mutate({
            name: formData.path,
            namespace_uuid: formData.namespaceUuid,
            enable_api_key_auth: formData.authType === 'api-key',
            enable_oauth: formData.authType === 'oauth2',
            enable_max_rate: false,
            enable_client_max_rate: false,
        });
    };

    return (
        <Card className="bg-zinc-900 border-zinc-700 mb-8 border-l-4 border-l-green-600 shadow-xl">
            <CardContent className="pt-6">
                <div className="flex justify-between items-start mb-6">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium bg-green-500/10 text-green-400 border border-green-500/20">
                        <Plus className="h-3 w-3" /> New Endpoint
                    </div>
                    <Button variant="ghost" size="sm" onClick={onSuccess} className="text-zinc-500 hover:text-white h-6 w-6 p-0 rounded-full">
                        X
                    </Button>
                </div>

                <form onSubmit={handleSubmit} className="space-y-5">
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Path</label>
                            <div className="flex">
                                <span className="bg-zinc-800 text-zinc-500 p-2.5 rounded-l-md text-sm border border-zinc-800 border-r-0">/</span>
                                <input
                                    required
                                    value={formData.path}
                                    onChange={e => setFormData({ ...formData, path: e.target.value.replace(/^\//, '') })}
                                    className="w-full bg-zinc-950 border border-zinc-800 rounded-r-md p-2.5 text-sm text-white focus:ring-1 focus:ring-green-500 outline-none"
                                    placeholder="my-tool"
                                />
                            </div>
                        </div>
                        <div>
                            <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Method</label>
                            <select
                                value={formData.method}
                                onChange={e => setFormData({ ...formData, method: e.target.value })}
                                className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-green-500 outline-none"
                            >
                                <option value="POST">POST</option>
                                <option value="GET">GET</option>
                            </select>
                        </div>
                    </div>

                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Target Namespace</label>
                        <select
                            required
                            value={formData.namespaceUuid}
                            onChange={e => setFormData({ ...formData, namespaceUuid: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-green-500 outline-none"
                        >
                            <option value="" disabled>Select a namespace</option>
                            {namespaces.map(ns => (
                                <option key={ns.uuid} value={ns.uuid}>{ns.name}</option>
                            ))}
                        </select>
                    </div>

                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Authentication</label>
                        <select
                            value={formData.authType}
                            onChange={e => setFormData({ ...formData, authType: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-green-500 outline-none"
                        >
                            <option value="api-key">API Key</option>
                            <option value="oauth2">OAuth 2.0 (Coming Soon)</option>
                            <option value="none">None (Public)</option>
                        </select>
                    </div>

                    <div className="flex justify-end pt-2">
                        <Button type="submit" disabled={createMutation.isPending} className="bg-green-600 hover:bg-green-500 text-white font-medium">
                            {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Create Endpoint
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}
