"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Plus, FileCode, Trash2, Play, Edit2 } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';
import { normalizeSavedScripts } from './scripts-page-normalizers';

export default function ScriptsDashboard() {
    const { data: scripts, isLoading, refetch } = trpc.savedScripts.list.useQuery();
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const normalizedScripts = normalizeSavedScripts(scripts);

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Saved Scripts</h1>
                    <p className="text-zinc-500">
                        Manage and execute reusable automation scripts
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button onClick={() => setIsCreateOpen(!isCreateOpen)} className="bg-blue-600 hover:bg-blue-500">
                        <Plus className="mr-2 h-4 w-4" /> New Script
                    </Button>
                </div>
            </div>

            {isCreateOpen && (
                <CreateScriptForm onSuccess={() => { setIsCreateOpen(false); refetch(); }} />
            )}

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 flex justify-center p-12">
                        <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                    </div>
                ) : normalizedScripts.length === 0 ? (
                    <div className="col-span-3 text-center p-12 text-zinc-500 bg-zinc-900/50 rounded-lg border border-zinc-800 border-dashed">
                        <FileCode className="h-12 w-12 mx-auto mb-4 opacity-30" />
                        <p className="text-lg font-medium">No Scripts Saved</p>
                        <p className="text-sm mt-1">Save your common automation tasks here.</p>
                    </div>
                ) : normalizedScripts.map((script) => (
                    <ScriptCard key={script.uuid} script={script} onUpdate={refetch} />
                ))}
            </div>
        </div>
    );
}

function ScriptCard({ script, onUpdate }: { script: any; onUpdate: () => void }) {
    const deleteMutation = trpc.savedScripts.delete.useMutation({
        onSuccess: () => {
            toast.success("Script deleted");
            onUpdate();
        },
        onError: (err) => {
            toast.error(`Failed to delete: ${err.message}`);
        }
    });

    const runMutation = trpc.savedScripts.execute.useMutation({
        onSuccess: (result) => {
            toast.success("Script executed");
            console.log("Script Result:", result);
            // Could show a modal result dialog here
        },
        onError: (err) => {
            toast.error(`Execution failed: ${err.message}`);
        }
    });

    return (
        <Card className="bg-zinc-900 border-zinc-800 hover:border-zinc-700 transition-colors group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-lg font-medium text-zinc-200 flex items-center gap-2 truncate">
                    <FileCode className="h-5 w-5 text-yellow-500" />
                    <span className="truncate">{script.name}</span>
                </CardTitle>
                <div className="flex items-center gap-2">
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            runMutation.mutate({ uuid: script.uuid });
                        }}
                        disabled={runMutation.isPending}
                        className="text-zinc-600 hover:text-green-400 transition-colors"
                        title="Run Script"
                    >
                        {runMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
                    </button>
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            if (confirm(`Delete script "${script.name}"?`)) {
                                deleteMutation.mutate({ uuid: script.uuid });
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
                    <p className="text-sm text-zinc-400 min-h-[40px] line-clamp-2">
                        {script.description || "No description."}
                    </p>
                    <div className="bg-black/30 p-2 rounded border border-zinc-800">
                        <pre className="text-[10px] text-zinc-500 font-mono h-20 overflow-hidden opacity-70">
                            {script.code}
                        </pre>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function CreateScriptForm({ onSuccess }: { onSuccess: () => void }) {
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        code: `// Write your script here
// Use 'await mcp.toolName({...})' to call tools

console.log("Hello World");
`,
    });

    const createMutation = trpc.savedScripts.create.useMutation({
        onSuccess: () => {
            toast.success("Script saved");
            onSuccess();
        },
        onError: (err) => {
            toast.error(`Error saving script: ${err.message}`);
        }
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        createMutation.mutate(formData);
    };

    return (
        <Card className="bg-zinc-900 border-zinc-700 mb-8 border-l-4 border-l-yellow-600 shadow-xl">
            <CardContent className="pt-6">
                <div className="flex justify-between items-start mb-6">
                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium bg-yellow-500/10 text-yellow-400 border border-yellow-500/20">
                        <Plus className="h-3 w-3" /> New Script
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
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-yellow-500 outline-none"
                            placeholder="e.g. daily-cleanup"
                        />
                    </div>
                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Description</label>
                        <input
                            value={formData.description}
                            onChange={e => setFormData({ ...formData, description: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white focus:ring-1 focus:ring-yellow-500 outline-none"
                            placeholder="Optional description"
                        />
                    </div>
                    <div>
                        <label className="text-xs text-zinc-500 uppercase font-bold mb-1.5 block">Code (JS/TS)</label>
                        <textarea
                            required
                            value={formData.code}
                            onChange={e => setFormData({ ...formData, code: e.target.value })}
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 text-sm text-white font-mono h-64 focus:ring-1 focus:ring-yellow-500 outline-none"
                        />
                    </div>

                    <div className="flex justify-end pt-2">
                        <Button type="submit" disabled={createMutation.isPending} className="bg-yellow-600 hover:bg-yellow-500 text-white font-medium">
                            {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Save Script
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}
