"use client";

import { useState } from 'react';
import { Button } from "@tormentnexus/ui";
import { Loader2, Save, X, Server, Terminal, Globe, Plus, Trash2 } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

interface EditMcpServerProps {
    server?: any; // If provided, edit mode. If null, create mode.
    onSuccess: () => void;
    onCancel: () => void;
}

export function EditMcpServer({ server, onSuccess, onCancel }: EditMcpServerProps) {
    const isEdit = !!server;
    const [formData, setFormData] = useState({
        name: server?.name || '',
        command: server?.config?.command || '',
        args: (server?.config?.args || []).join(' '), // simple space separated for UI
        envJson: JSON.stringify(server?.config?.env || {}, null, 2),
    });

    const createMutation = trpc.mcpServers.create.useMutation({
        onSuccess: () => {
            toast.success(`Server ${formData.name} created`);
            onSuccess();
        },
        onError: (err) => toast.error(`Failed to create: ${err.message}`)
    });

    const updateMutation = trpc.mcpServers.update.useMutation({ // Assuming update endpoint exists or we use addServer to overwrite
        onSuccess: () => {
            toast.success(`Server ${formData.name} updated`);
            onSuccess();
        },
        onError: (err) => toast.error(`Failed to update: ${err.message}`)
    });

    // Note: mcpServersRouter might strictly use 'create' (addServer) for both? 
    // Let's assume create works for upsert if logic allows, or we handled it in the router.
    // Checking back: router has create/update/delete.

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        try {
            const env = formData.envJson ? JSON.parse(formData.envJson) : {};
            const args = formData.args.split(' ').filter(Boolean);

            const payload = {
                name: formData.name,
                transport: {
                    type: 'STDIO' as const,
                    command: formData.command,
                    args: args,
                    env: env
                },
                enabled: true
            };

            if (isEdit) {
                updateMutation.mutate({
                    uuid: server.uuid,
                    ...payload
                } as any);
            } else {
                createMutation.mutate(payload as any);
            }

        } catch (e: any) {
            toast.error(`Invalid JSON in Environment: ${e.message}`);
        }
    };

    return (
        <div className="bg-zinc-900 border border-zinc-700 rounded-lg p-6 w-full max-w-2xl mx-auto shadow-xl">
            <div className="flex justify-between items-center mb-6 border-b border-zinc-800 pb-4">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                    <Server className="h-5 w-5 text-blue-500" />
                    {isEdit ? 'Edit Server' : 'Add New Server'}
                </h2>
                <Button variant="ghost" size="sm" onClick={onCancel}>
                    <X className="h-5 w-5" />
                </Button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
                <div className="grid grid-cols-2 gap-6">
                    <div className="space-y-2">
                        <label className="text-xs font-bold text-zinc-500 uppercase">Server Name</label>
                        <input
                            required
                            disabled={isEdit} // Often names are IDs
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-blue-500 outline-none disabled:opacity-50"
                            placeholder="e.g. filesystem"
                        />
                    </div>
                    <div className="space-y-2">
                        <label className="text-xs font-bold text-zinc-500 uppercase">Command</label>
                        <input
                            required
                            value={formData.command}
                            onChange={(e) => setFormData({ ...formData, command: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-blue-500 outline-none font-mono"
                            placeholder="e.g. npx"
                        />
                    </div>
                </div>

                <div className="space-y-2">
                    <label className="text-xs font-bold text-zinc-500 uppercase">Arguments (Space separated)</label>
                    <input
                        value={formData.args}
                        onChange={(e) => setFormData({ ...formData, args: e.target.value })}
                        className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-blue-500 outline-none font-mono"
                        placeholder="-y @modelcontextprotocol/server-filesystem /path/to/allow"
                    />
                </div>

                <div className="space-y-2">
                    <label className="text-xs font-bold text-zinc-500 uppercase">Environment Variables (JSON)</label>
                    <textarea
                        value={formData.envJson}
                        onChange={(e) => setFormData({ ...formData, envJson: e.target.value })}
                        className="w-full h-32 bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-blue-500 outline-none font-mono text-sm"
                        placeholder={'{\n  "KEY": "VALUE"\n}'}
                    />
                </div>

                <div className="flex justify-end gap-3 pt-4 border-t border-zinc-800">
                    <Button type="button" variant="outline" onClick={onCancel}>Cancel</Button>
                    <Button
                        type="submit"
                        disabled={createMutation.isPending || updateMutation.isPending}
                        className="bg-blue-600 hover:bg-blue-500"
                    >
                        {(createMutation.isPending || updateMutation.isPending) && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        {isEdit ? 'Save Changes' : 'Create Server'}
                    </Button>
                </div>
            </form>
        </div>
    );
}
