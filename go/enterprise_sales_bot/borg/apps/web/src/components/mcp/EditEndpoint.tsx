"use client";

import { useState } from 'react';
import { Button } from "@tormentnexus/ui";
import { Loader2, Save, X, Globe, Shield, Activity } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

type NamespaceOption = {
    uuid: string;
    name: string;
};

type EndpointLike = {
    uuid: string;
    name: string;
    path?: string;
    method?: string;
    namespace?: string;
    namespace_uuid?: string;
    authRequired?: boolean;
    rateLimit?: number | string;
};

type EndpointMutationPayload = {
    name: string;
    description: string | null;
    namespace_uuid: string;
    enable_api_key_auth: boolean;
    enable_oauth: boolean;
    enable_max_rate: boolean;
    enable_client_max_rate: boolean;
    max_rate: number;
    max_rate_seconds: number;
    client_max_rate: number;
    client_max_rate_seconds: number;
    use_query_param_auth: boolean;
};

interface EditEndpointProps {
    endpoint?: EndpointLike;
    onSuccess: () => void;
    onCancel: () => void;
}

export function EditEndpoint({ endpoint, onSuccess, onCancel }: EditEndpointProps) {
    const isEdit = !!endpoint;
    const [formData, setFormData] = useState({
        name: endpoint?.name || '',
        path: endpoint?.path || '',
        method: endpoint?.method || 'POST',
        namespace: endpoint?.namespace || 'default',
        authRequired: endpoint?.authRequired ?? true,
        rateLimit: endpoint?.rateLimit?.toString() || '60',
    });

    const { data: namespaces } = trpc.namespaces.list.useQuery();

    const createMutation = trpc.endpoints.create.useMutation({
        onSuccess: () => {
            toast.success(`Endpoint ${formData.name} created`);
            onSuccess();
        },
        onError: (err) => toast.error(`Failed to create: ${err.message}`)
    });

    const updateMutation = trpc.endpoints.update.useMutation({
        onSuccess: () => {
            toast.success(`Endpoint ${formData.name} updated`);
            onSuccess();
        },
        onError: (err) => toast.error(`Failed to update: ${err.message}`)
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        const namespaceOptions = (namespaces ?? []) as NamespaceOption[];
        const selectedNamespace = namespaceOptions.find(
            (ns) => ns.name === formData.namespace,
        );
        const namespaceUuid = selectedNamespace?.uuid || endpoint?.namespace_uuid;

        if (!namespaceUuid) {
            toast.error('Please select a valid namespace');
            return;
        }

        const parsedRateLimit = Number.parseInt(formData.rateLimit, 10);
        const safeRateLimit = Number.isFinite(parsedRateLimit)
            ? parsedRateLimit
            : 60;

        const payload: EndpointMutationPayload = {
            name: formData.name,
            description: null,
            namespace_uuid: namespaceUuid,
            enable_api_key_auth: formData.authRequired,
            enable_oauth: false,
            enable_max_rate: true,
            enable_client_max_rate: true,
            max_rate: safeRateLimit,
            max_rate_seconds: 60,
            client_max_rate: safeRateLimit,
            client_max_rate_seconds: 60,
            use_query_param_auth: false,
        };

        if (isEdit) {
            updateMutation.mutate({
                uuid: endpoint.uuid,
                ...payload,
            });
        } else {
            createMutation.mutate({
                ...payload,
            });
        }
    };

    return (
        <div className="bg-zinc-900 border border-zinc-700 rounded-lg p-6 w-full max-w-2xl mx-auto shadow-xl">
            <div className="flex justify-between items-center mb-6 border-b border-zinc-800 pb-4">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                    <Globe className="h-5 w-5 text-cyan-500" />
                    {isEdit ? 'Edit Endpoint' : 'Create Endpoint'}
                </h2>
                <Button variant="ghost" size="sm" onClick={onCancel}>
                    <X className="h-5 w-5" />
                </Button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
                <div className="grid grid-cols-2 gap-6">
                    <div className="space-y-2">
                        <label className="text-xs font-bold text-zinc-500 uppercase">Friendly Name</label>
                        <input
                            required
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-cyan-500 outline-none"
                            placeholder="e.g. Process File"
                        />
                    </div>
                    <div className="space-y-2">
                        <label className="text-xs font-bold text-zinc-500 uppercase">Namespace</label>
                        <select
                            value={formData.namespace}
                            onChange={(e) => setFormData({ ...formData, namespace: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-cyan-500 outline-none"
                        >
                            <option value="default">default</option>
                            {(namespaces as NamespaceOption[] | undefined)?.map((ns) => (
                                <option key={ns.name} value={ns.name}>{ns.name}</option>
                            ))}
                        </select>
                    </div>
                </div>

                <div className="grid grid-cols-3 gap-6">
                    <div className="col-span-2 space-y-2">
                        <label className="text-xs font-bold text-zinc-500 uppercase">Path</label>
                        <div className="flex items-center">
                            <span className="bg-zinc-800 border border-zinc-700 border-r-0 rounded-l p-2 text-zinc-400 font-mono text-sm">/api/mcp/</span>
                            <input
                                required
                                value={formData.path}
                                onChange={(e) => setFormData({ ...formData, path: e.target.value })}
                                className="w-full bg-zinc-950 border border-zinc-800 rounded-r p-2 text-white focus:border-cyan-500 outline-none font-mono"
                                placeholder="process/file"
                            />
                        </div>
                    </div>
                    <div className="space-y-2">
                        <label className="text-xs font-bold text-zinc-500 uppercase">Method</label>
                        <select
                            value={formData.method}
                            onChange={(e) => setFormData({ ...formData, method: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-cyan-500 outline-none font-mono"
                        >
                            <option value="POST">POST</option>
                            <option value="GET">GET</option>
                            <option value="PUT">PUT</option>
                            <option value="DELETE">DELETE</option>
                        </select>
                    </div>
                </div>

                <div className="grid grid-cols-2 gap-6 pt-4 border-t border-zinc-800/50">
                    <div className="flex items-center gap-3">
                        <input
                            type="checkbox"
                            checked={formData.authRequired}
                            onChange={(e) => setFormData({ ...formData, authRequired: e.target.checked })}
                            className="w-4 h-4 rounded border-zinc-700 bg-zinc-900 text-cyan-600 focus:ring-cyan-500"
                        />
                        <div>
                            <label className="block text-sm font-medium text-white">Require Authentication</label>
                            <p className="text-xs text-zinc-500">Must provide valid API key</p>
                        </div>
                    </div>
                    <div className="space-y-1">
                        <label className="text-xs font-bold text-zinc-500 uppercase">Rate Limit (req/min)</label>
                        <input
                            type="number"
                            value={formData.rateLimit}
                            onChange={(e) => setFormData({ ...formData, rateLimit: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-cyan-500 outline-none"
                        />
                    </div>
                </div>

                <div className="flex justify-end gap-3 pt-4 border-t border-zinc-800">
                    <Button type="button" variant="outline" onClick={onCancel}>Cancel</Button>
                    <Button
                        type="submit"
                        disabled={createMutation.isPending || updateMutation.isPending}
                        className="bg-cyan-600 hover:bg-cyan-500"
                    >
                        {(createMutation.isPending || updateMutation.isPending) && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        {isEdit ? 'Save Changes' : 'Create Endpoint'}
                    </Button>
                </div>
            </form>
        </div>
    );
}
