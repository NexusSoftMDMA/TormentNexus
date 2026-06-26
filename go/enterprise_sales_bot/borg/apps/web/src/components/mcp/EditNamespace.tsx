"use client";

import { useState } from 'react';
import { Button } from "@tormentnexus/ui";
import { Loader2, Save, X, Box, Shield, Layers, FileText } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

interface NamespaceInput {
    uuid: string;
    name: string;
    description?: string | null;
}

interface EditNamespaceProps {
    namespace?: NamespaceInput;
    onSuccess: () => void;
    onCancel: () => void;
}

export function EditNamespace({ namespace, onSuccess, onCancel }: EditNamespaceProps) {
    const isEdit = !!namespace;
    const [formData, setFormData] = useState({
        name: namespace?.name || '',
        description: namespace?.description || '',
    });

    // Assuming namespacesRouter has create/update
    const createMutation = trpc.namespaces.create.useMutation({
        onSuccess: () => {
            toast.success(`Namespace ${formData.name} created`);
            onSuccess();
        },
        onError: (err) => toast.error(`Failed to create: ${err.message}`)
    });

    const updateMutation = trpc.namespaces.update.useMutation({
        onSuccess: () => {
            toast.success(`Namespace ${formData.name} updated`);
            onSuccess();
        },
        onError: (err) => toast.error(`Failed to update: ${err.message}`)
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (isEdit) {
            updateMutation.mutate({
                uuid: namespace.uuid,
                name: namespace.name,
                description: formData.description
            });
        } else {
            createMutation.mutate({
                name: formData.name,
                description: formData.description
            });
        }
    };

    return (
        <div className="bg-zinc-900 border border-zinc-700 rounded-lg p-6 w-full max-w-lg mx-auto shadow-xl">
            <div className="flex justify-between items-center mb-6 border-b border-zinc-800 pb-4">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                    <Box className="h-5 w-5 text-purple-500" />
                    {isEdit ? 'Edit Namespace' : 'Create Namespace'}
                </h2>
                <Button variant="ghost" size="sm" onClick={onCancel}>
                    <X className="h-5 w-5" />
                </Button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
                <div className="space-y-2">
                    <label className="text-xs font-bold text-zinc-500 uppercase">Name</label>
                    <input
                        required
                        disabled={isEdit}
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                        className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-purple-500 outline-none disabled:opacity-50"
                        placeholder="e.g. data-processing"
                    />
                </div>

                <div className="space-y-2">
                    <label className="text-xs font-bold text-zinc-500 uppercase">Description</label>
                    <textarea
                        value={formData.description}
                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                        className="w-full h-24 bg-zinc-950 border border-zinc-800 rounded p-2 text-white focus:border-purple-500 outline-none resize-none"
                        placeholder="Purpose of this namespace..."
                    />
                </div>

                <div className="flex justify-end gap-3 pt-4 border-t border-zinc-800">
                    <Button type="button" variant="outline" onClick={onCancel}>Cancel</Button>
                    <Button
                        type="submit"
                        disabled={createMutation.isPending || updateMutation.isPending}
                        className="bg-purple-600 hover:bg-purple-500"
                    >
                        {(createMutation.isPending || updateMutation.isPending) && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        {isEdit ? 'Save Changes' : 'Create Namespace'}
                    </Button>
                </div>
            </form>
        </div>
    );
}
