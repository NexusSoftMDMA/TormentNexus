"use client";

import React, { useState } from 'react';
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { trpc } from '../utils/trpc';
import { toast } from 'sonner';
import { Loader2, Save, Trash2, Edit2, X } from 'lucide-react';

interface Prompt {
    id: string;
    version: number;
    description: string;
    template: string;
    updatedAt: string;
}

export function PromptLibrary() {
    const [selected, setSelected] = useState<Prompt | null>(null);
    const [editMode, setEditMode] = useState(false);
    const [editedTemplate, setEditedTemplate] = useState("");
    const [search, setSearch] = useState('');

    const utils = trpc.useUtils();
    const { data: prompts = [], isLoading: loading, error: fetchError } = trpc.prompts.list.useQuery(undefined, {
        refetchInterval: 10000,
    });

    const saveMutation = trpc.prompts.save.useMutation({
        onSuccess: () => {
            toast.success('Prompt saved successfully');
            setEditMode(false);
            utils.prompts.list.invalidate();
        },
        onError: (err: any) => {
            toast.error(`Failed to save prompt: ${err.message}`);
        }
    });

    const deleteMutation = trpc.prompts.delete.useMutation({
        onSuccess: () => {
            toast.success('Prompt deleted');
            setSelected(null);
            utils.prompts.list.invalidate();
        },
        onError: (err: any) => {
            toast.error(`Failed to delete prompt: ${err.message}`);
        }
    });

    const error = fetchError ? fetchError.message : null;

    const handleSave = () => {
        if (!selected) return;
        saveMutation.mutate({
            id: selected.id,
            description: selected.description,
            template: editedTemplate
        });
    };

    const handleDelete = () => {
        if (!selected) return;
        if (confirm(`Are you sure you want to delete the prompt '${selected.id}'?`)) {
            deleteMutation.mutate({ id: selected.id });
        }
    };

    const filteredPrompts = prompts.filter((p: Prompt) => {
        const q = search.trim().toLowerCase();
        if (!q) return true;
        return p.id.toLowerCase().includes(q) || p.description.toLowerCase().includes(q);
    });

    return (
        <div className="flex h-[80vh] border border-zinc-800 rounded-xl overflow-hidden bg-black/40">
            <div className="w-1/3 border-r border-zinc-800 bg-black/20 flex flex-col h-full min-h-0">
                <div className="p-4 border-b border-zinc-800 shrink-0">
                    <div className="mt-1 space-y-2">
                        <Input
                            value={search}
                            onChange={(e: any) => setSearch(e.target.value)}
                            placeholder="Search prompts..."
                            className="bg-black/30 border-zinc-800"
                        />
                        {error ? (
                            <div className="text-xs text-red-400 flex items-center justify-between">
                                <span>Load failed: {error}</span>
                                <Button size="sm" variant="ghost" onClick={() => utils.prompts.list.invalidate()}>Retry</Button>
                            </div>
                        ) : null}
                    </div>
                </div>
                <div className="overflow-y-auto flex-1">
                    {loading ? (
                        <div className="p-4 text-sm text-zinc-400 flex items-center gap-2">
                            <Loader2 className="w-4 h-4 animate-spin" /> Loading prompts...
                        </div>
                    ) : filteredPrompts.map((p: Prompt) => (
                        <div
                            key={p.id}
                            onClick={() => { setSelected(p); setEditedTemplate(p.template); setEditMode(false); }}
                            className={`p-4 border-b border-zinc-800/50 cursor-pointer hover:bg-zinc-800/30 transition-colors ${selected?.id === p.id ? 'bg-indigo-500/10 border-l-2 border-l-indigo-400' : ''}`}
                        >
                            <div className="font-mono font-bold text-sm text-indigo-300">{p.id}</div>
                            <div className="text-xs text-zinc-400 truncate mt-1">{p.description}</div>
                            <div className="text-[10px] text-zinc-600 mt-2 font-mono uppercase tracking-wider">v{p.version} • {new Date(p.updatedAt).toLocaleDateString()}</div>
                        </div>
                    ))}
                    {!loading && !error && filteredPrompts.length === 0 ? (
                        <div className="p-4 text-xs text-zinc-500">No prompts match current filter.</div>
                    ) : null}
                </div>
            </div>

            <div className="flex-1 flex flex-col bg-black/40 min-h-0">
                {selected ? (
                    <>
                        <div className="p-4 border-b border-zinc-800 flex justify-between items-start bg-zinc-950/50 shrink-0">
                            <div>
                                <h3 className="text-lg font-bold text-white flex items-center gap-2">
                                    {selected.id}
                                    <span className="text-[10px] bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded uppercase font-mono">v{selected.version}</span>
                                </h3>
                                <p className="text-xs text-zinc-400 mt-1">{selected.description}</p>
                            </div>
                            <div className="flex items-center gap-2">
                                {editMode ? (
                                    <>
                                        <Button variant="ghost" size="sm" onClick={() => setEditMode(false)}>
                                            <X className="w-4 h-4 mr-2" /> Cancel
                                        </Button>
                                        <Button size="sm" className="bg-emerald-600 hover:bg-emerald-500 text-white" onClick={handleSave} disabled={saveMutation.isPending}>
                                            {saveMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Save className="w-4 h-4 mr-2" />}
                                            Save Template
                                        </Button>
                                    </>
                                ) : (
                                    <>
                                        <Button variant="outline" size="sm" className="border-zinc-700 hover:bg-zinc-800 text-zinc-300" onClick={() => setEditMode(true)}>
                                            <Edit2 className="w-4 h-4 mr-2" /> Edit
                                        </Button>
                                        <Button variant="outline" size="sm" className="border-red-900/50 text-red-400 hover:bg-red-950/30 hover:text-red-300" onClick={handleDelete} disabled={deleteMutation.isPending}>
                                            {deleteMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Trash2 className="w-4 h-4 mr-2" />}
                                            Delete
                                        </Button>
                                    </>
                                )}
                            </div>
                        </div>
                        <div className="flex-1 p-4 overflow-hidden relative">
                            {editMode ? (
                                <textarea
                                    className="w-full h-full bg-zinc-950/80 border border-zinc-800 rounded-lg p-4 font-mono text-sm text-emerald-300 focus:outline-none focus:ring-1 focus:ring-emerald-500 resize-none"
                                    value={editedTemplate}
                                    onChange={(e: any) => setEditedTemplate(e.target.value)}
                                    spellCheck={false}
                                />
                            ) : (
                                <pre className="w-full h-full bg-black/40 border border-zinc-800 rounded-lg p-6 font-mono text-sm text-zinc-300 overflow-auto whitespace-pre-wrap selection:bg-indigo-500/30">
                                    {selected.template}
                                </pre>
                            )}
                        </div>
                    </>
                ) : (
                    <div className="flex-1 flex flex-col items-center justify-center text-zinc-600 p-8 text-center">
                        <div className="w-16 h-16 border-2 border-dashed border-zinc-800 rounded-full flex items-center justify-center mb-4">
                            <Edit2 className="w-6 h-6" />
                        </div>
                        <h3 className="text-zinc-400 font-medium">No Prompt Selected</h3>
                        <p className="text-xs max-w-xs mt-2 italic leading-relaxed">
                            Choose a prompt from the library to view or edit its implementation details.
                        </p>
                    </div>
                )}
            </div>
        </div>
    );
}
