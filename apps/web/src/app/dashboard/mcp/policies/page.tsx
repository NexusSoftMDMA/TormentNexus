"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Plus, Shield, Trash2, Edit2 } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

export default function PoliciesDashboard() {
    const { data: policies, isLoading, refetch } = trpc.policies.list.useQuery();
    const [isCreateOpen, setIsCreateOpen] = useState(false);

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Access Policies</h1>
                    <p className="text-zinc-500">
                        Define Allow/Deny rules for MCP tools and resources
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button onClick={() => setIsCreateOpen(!isCreateOpen)} className="bg-blue-600 hover:bg-blue-500">
                        <Plus className="mr-2 h-4 w-4" /> Create Policy
                    </Button>
                </div>
            </div>

            {isCreateOpen && (
                <CreatePolicyForm onSuccess={() => { setIsCreateOpen(false); refetch(); }} />
            )}

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 flex justify-center p-12">
                        <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                    </div>
                ) : (policies?.length ?? 0) === 0 ? (
                    <div className="col-span-3 text-center p-12 text-zinc-500 bg-zinc-900/50 rounded-lg border border-zinc-800 border-dashed">
                        <Shield className="h-12 w-12 mx-auto mb-4 opacity-30" />
                        <p className="text-lg font-medium">No Policies Defined</p>
                        <p className="text-sm mt-1">Create a policy to restrict access to sensitive tools.</p>
                    </div>
                ) : policies?.map((policy: any) => (
                    <PolicyCard key={policy.uuid} policy={policy} onUpdate={refetch} />
                ))}
            </div>
        </div>
    );
}

function PolicyCard({ policy, onUpdate }: { policy: any; onUpdate: () => void }) {
    const deleteMutation = trpc.policies.delete.useMutation({
        onSuccess: () => {
            toast.success("Policy deleted");
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
                    <Shield className="h-5 w-5 text-red-500" />
                    {policy.name}
                </CardTitle>
                <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            if (confirm(`Delete policy "${policy.name}"?`)) {
                                deleteMutation.mutate({ uuid: policy.uuid });
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
                        {policy.description || "No description."}
                    </p>
                    <div className="grid grid-cols-2 gap-2 text-xs">
                        <div className="bg-green-500/10 border border-green-500/20 text-green-400 p-2 rounded">
                            <span className="font-bold block mb-1">ALLOW ({policy.rules.allow?.length || 0})</span>
                            <div className="truncate opacity-70">
                                {policy.rules.allow?.join(', ') || 'None'}
                            </div>
                        </div>
                        <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-2 rounded">
                            <span className="font-bold block mb-1">DENY ({policy.rules.deny?.length || 0})</span>
                            <div className="truncate opacity-70">
                                {policy.rules.deny?.join(', ') || 'None'}
                            </div>
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function CreatePolicyForm({ onSuccess }: { onSuccess: () => void }) {
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        allow: '',
        deny: '',
    });

    const createMutation = trpc.policies.create.useMutation({
        onSuccess: () => {
            toast.success("Policy created");
            onSuccess();
        },
        onError: (err) => {
            toast.error(`Error creating policy: ${err.message}`);
        }
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        createMutation.mutate({
            name: formData.name,
            description: formData.description,
            rules: {
                allow: formData.allow.split(',').map(s => s.trim()).filter(Boolean),
                deny: formData.deny.split(',').map(s => s.trim()).filter(Boolean),
            }
        });
    };

    return (
        <Card className="bg-zinc-900 border-zinc-700 mb-8 border-l-4 border-l-red-600 shadow-xl">
            <CardContent className="pt-6">
                <div className="flex justify-between items-start mb-6">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                        <Plus className="h-3 w-3" /> New Policy
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
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-red-500 outline-none"
                            placeholder="e.g. read-only-access"
                        />
                    </div>
                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Description</label>
                        <input
                            value={formData.description}
                            onChange={e => setFormData({ ...formData, description: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-red-500 outline-none"
                            placeholder="Optional description"
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="text-xs text-green-500 uppercase font-bold mb-1.5 block">Allow Patterns (comma separated)</label>
                            <textarea
                                value={formData.allow}
                                onChange={e => setFormData({ ...formData, allow: e.target.value })}
                                className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white font-mono h-24 focus:ring-1 focus:ring-green-500 outline-none"
                                placeholder="*"
                            />
                        </div>
                        <div>
                            <label className="text-xs text-red-500 uppercase font-bold mb-1.5 block">Deny Patterns (comma separated)</label>
                            <textarea
                                value={formData.deny}
                                onChange={e => setFormData({ ...formData, deny: e.target.value })}
                                className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white font-mono h-24 focus:ring-1 focus:ring-red-500 outline-none"
                                placeholder="filesystem:write_*, shell:*"
                            />
                        </div>
                    </div>

                    <div className="flex justify-end pt-2">
                        <Button type="submit" disabled={createMutation.isPending} className="bg-red-600 hover:bg-red-500 text-white font-medium">
                            {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Create Policy
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}
