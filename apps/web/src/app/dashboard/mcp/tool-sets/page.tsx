"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Plus, Box, Trash2, Layers, Wrench, Check } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

export default function ToolSetsDashboard() {
    const { data: toolSets, isLoading, refetch } = trpc.toolSets.list.useQuery();
    const { data: tools } = trpc.tools.list.useQuery(); // For selection in creation
    const [isCreateOpen, setIsCreateOpen] = useState(false);

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Tool Sets</h1>
                    <p className="text-zinc-500">
                        Group multiple tools into reusable logical sets
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button onClick={() => setIsCreateOpen(!isCreateOpen)} className="bg-blue-600 hover:bg-blue-500">
                        <Plus className="mr-2 h-4 w-4" /> Create Tool Set
                    </Button>
                </div>
            </div>

            {isCreateOpen && (
                <CreateToolSetForm
                    tools={tools || []}
                    onSuccess={() => { setIsCreateOpen(false); refetch(); }}
                />
            )}

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 flex justify-center p-12">
                        <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                    </div>
                ) : (toolSets?.length ?? 0) === 0 ? (
                    <div className="col-span-3 text-center p-12 text-zinc-500 bg-zinc-900/50 rounded-lg border border-zinc-800 border-dashed">
                        <Layers className="h-12 w-12 mx-auto mb-4 opacity-30" />
                        <p className="text-lg font-medium">No Tool Sets</p>
                        <p className="text-sm mt-1">Combine tools into a set for easier assignment.</p>
                    </div>
                ) : toolSets?.map((ts: any) => (
                    <ToolSetCard key={ts.uuid} toolSet={ts} onUpdate={refetch} />
                ))}
            </div>
        </div>
    );
}

function ToolSetCard({ toolSet, onUpdate }: { toolSet: any; onUpdate: () => void }) {
    const deleteMutation = trpc.toolSets.delete.useMutation({
        onSuccess: () => {
            toast.success("Tool set deleted");
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
                    <Layers className="h-5 w-5 text-indigo-500" />
                    {toolSet.name}
                </CardTitle>
                <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            if (confirm(`Delete tool set "${toolSet.name}"?`)) {
                                deleteMutation.mutate({ uuid: toolSet.uuid });
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
                        {toolSet.description || "No description."}
                    </p>
                    <div className="flex gap-2 text-xs">
                        <div className="flex items-center gap-1 bg-zinc-800 px-2 py-1 rounded text-zinc-300">
                            <Wrench className="h-3 w-3" />
                            <span>{toolSet.tools?.length ?? 0} Tools</span>
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function CreateToolSetForm({ tools, onSuccess }: { tools: any[]; onSuccess: () => void }) {
    const [formData, setFormData] = useState<{
        name: string;
        description: string;
        selectedTools: string[];
    }>({
        name: '',
        description: '',
        selectedTools: [],
    });

    const createMutation = trpc.toolSets.create.useMutation({
        onSuccess: () => {
            toast.success("Tool set created");
            onSuccess();
        },
        onError: (err) => {
            toast.error(`Error creating tool set: ${err.message}`);
        }
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        createMutation.mutate({
            name: formData.name,
            description: formData.description,
            tools: formData.selectedTools,
        });
    };

    const toggleTool = (uuid: string) => {
        setFormData(prev => ({
            ...prev,
            selectedTools: prev.selectedTools.includes(uuid)
                ? prev.selectedTools.filter(id => id !== uuid)
                : [...prev.selectedTools, uuid]
        }));
    };

    return (
        <Card className="bg-zinc-900 border-zinc-700 mb-8 border-l-4 border-l-indigo-600 shadow-xl">
            <CardContent className="pt-6">
                <div className="flex justify-between items-start mb-6">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                        <Plus className="h-3 w-3" /> New Tool Set
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
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-indigo-500 outline-none"
                            placeholder="e.g. data-analysis-tools"
                        />
                    </div>
                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Description</label>
                        <input
                            value={formData.description}
                            onChange={e => setFormData({ ...formData, description: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-indigo-500 outline-none"
                            placeholder="Optional description"
                        />
                    </div>

                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Select Tools ({formData.selectedTools.length})</label>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 max-h-60 overflow-y-auto p-2 bg-zinc-950 border border-zinc-800 rounded-md">
                            {tools.map(tool => (
                                <div
                                    key={tool.uuid}
                                    onClick={() => toggleTool(tool.uuid)}
                                    className={`p-2 rounded cursor-pointer border text-xs flex items-center justify-between ${formData.selectedTools.includes(tool.uuid)
                                        ? 'bg-indigo-500/20 border-indigo-500/50 text-white'
                                        : 'bg-zinc-900 border-zinc-800 text-zinc-400 hover:bg-zinc-800'
                                        }`}
                                >
                                    <span className="truncate mr-2" title={tool.name}>{tool.name}</span>
                                    {formData.selectedTools.includes(tool.uuid) && <Check className="h-3 w-3 text-indigo-400" />}
                                </div>
                            ))}
                            {tools.length === 0 && (
                                <div className="col-span-3 text-center text-zinc-500 py-4">No tools available to select.</div>
                            )}
                        </div>
                    </div>

                    <div className="flex justify-end pt-2">
                        <Button type="submit" disabled={createMutation.isPending} className="bg-indigo-600 hover:bg-indigo-500 text-white font-medium">
                            {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Create Tool Set
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}
