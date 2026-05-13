'use client';

import React, { useState } from 'react';
import { trpc } from '../utils/trpc';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { Loader2, Book, Plus, Save, Edit, Code, Zap, ExternalLink, Trash2 } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from './ui/dialog';
import { Badge } from './ui/badge';

function getSkillContentText(value: unknown): string | null {
    if (!value || typeof value !== 'object') {
        return null;
    }

    const record = value as Record<string, unknown>;
    if (!Array.isArray(record.content) || record.content.length === 0) {
        return null;
    }

    const first = record.content[0];
    if (!first || typeof first !== 'object') {
        return null;
    }

    const firstRecord = first as Record<string, unknown>;
    return typeof firstRecord.text === 'string' ? firstRecord.text : null;
}

export function SkillLibrary() {
    const { data: skills, isLoading, refetch } = trpc.skills.list.useQuery();
    const { data: loadedSkills, refetch: refetchLoaded } = trpc.skills.listLoaded.useQuery();

    const createSkill = trpc.skills.create.useMutation();
    const saveSkill = trpc.skills.save.useMutation();
    const loadSkill = trpc.skills.load.useMutation();
    const unloadSkill = trpc.skills.unload.useMutation();

    const [selectedSkill, setSelectedSkill] = useState<any>(null);
    const [skillContent, setSkillContent] = useState('');
    const [isEditing, setIsEditing] = useState(false);

    const { data: skillData } = trpc.skills.read.useQuery(
        { name: selectedSkill?.id || '' },
        { enabled: !!selectedSkill }
    );

    React.useEffect(() => {
        const nextSkillContent = getSkillContentText(skillData);
        if (nextSkillContent) {
            setSkillContent(nextSkillContent);
        } else if (selectedSkill && !skillData) {
            setSkillContent('Loading...');
        }
    }, [skillData, selectedSkill]);

    // Create Dialog State
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [newSkillId, setNewSkillId] = useState('');
    const [newSkillName, setNewSkillName] = useState('');
    const [newSkillDesc, setNewSkillDesc] = useState('');

    const handleSelectSkill = (skill: any) => {
        setSelectedSkill(skill);
        setIsEditing(false);
        setSkillContent('Loading...');
    };

    const handleLoadToggle = async (id: string, isLoaded: boolean) => {
        try {
            if (isLoaded) {
                await unloadSkill.mutateAsync({ id });
            } else {
                await loadSkill.mutateAsync({ id });
            }
            refetchLoaded();
        } catch (e) {
            console.error(e);
        }
    };

    const handleSave = async () => {
        if (!selectedSkill) return;
        try {
            await saveSkill.mutateAsync({ id: selectedSkill.id, content: skillContent });
            setIsEditing(false);
            refetch();
        } catch (e) {
            console.error(e);
        }
    };

    const handleCreate = async () => {
        if (!newSkillId || !newSkillName) return;
        try {
            await createSkill.mutateAsync({
                id: newSkillId,
                name: newSkillName,
                description: newSkillDesc
            });
            setIsCreateOpen(false);
            setNewSkillId('');
            setNewSkillName('');
            setNewSkillDesc('');
            refetch();
        } catch (e) {
            console.error(e);
        }
    };

    return (
        <div className="flex h-full gap-4 p-4 text-white overflow-hidden">
            {/* Sidebar */}
            <div className="w-80 flex flex-col gap-4">
                <Card className="bg-zinc-900 border-zinc-800 flex-1 flex flex-col overflow-hidden shadow-lg">
                    <CardHeader className="p-4 border-b border-zinc-800 bg-zinc-950/50 flex flex-row items-center justify-between">
                        <div>
                            <CardTitle className="text-sm font-bold text-emerald-400 flex items-center gap-2">
                                <Book className="w-4 h-4" />
                                SKILLS
                            </CardTitle>
                            <CardDescription className="text-xs text-zinc-500">
                                Autonomous Capabilities
                            </CardDescription>
                        </div>
                        <Button size="icon" variant="ghost" className="h-8 w-8 hover:bg-zinc-800" onClick={() => setIsCreateOpen(true)}>
                            <Plus className="w-4 h-4 text-zinc-400" />
                        </Button>
                    </CardHeader>
                    <CardContent className="p-0 overflow-y-auto flex-1 custom-scrollbar">
                        {isLoading ? (
                            <div className="flex justify-center p-4"><Loader2 className="animate-spin w-6 h-6 text-zinc-600" /></div>
                        ) : skills && skills.length > 0 ? (
                            skills.map((s: any) => {
                                const isLoaded = loadedSkills?.some((ls: any) => ls.id === s.id);
                                return (
                                    <div
                                        key={s.id}
                                        onClick={() => handleSelectSkill(s)}
                                        className={`p-3 border-b border-zinc-800/50 cursor-pointer transition-colors hover:bg-zinc-800/40 group relative ${selectedSkill?.id === s.id ? 'bg-zinc-800/60 border-l-2 border-l-emerald-500' : 'border-l-2 border-l-transparent'}`}
                                    >
                                        <div className="flex items-center justify-between">
                                            <div className="font-medium text-sm text-zinc-200">{s.name}</div>
                                            {isLoaded && (
                                                <Badge variant="outline" className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20 text-[10px] py-0 px-1">
                                                    LOADED
                                                </Badge>
                                            )}
                                        </div>
                                        <div className="text-xs text-zinc-500 line-clamp-1 mt-1">{s.description}</div>
                                        <div className="flex items-center gap-2 mt-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <Button
                                                size="sm"
                                                variant="ghost"
                                                className={`h-6 px-2 text-[10px] ${isLoaded ? 'text-amber-400 hover:text-amber-300' : 'text-emerald-400 hover:text-emerald-300'}`}
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleLoadToggle(s.id, !!isLoaded);
                                                }}
                                            >
                                                <Zap className={`w-3 h-3 mr-1 ${isLoaded ? 'fill-current' : ''}`} />
                                                {isLoaded ? 'Unload' : 'Load JIT'}
                                            </Button>
                                        </div>
                                    </div>
                                );
                            })
                        ) : (
                            <div className="p-4 text-zinc-500 text-sm text-center">No skills found.</div>
                        )}
                    </CardContent>
                </Card>
            </div>

            {/* Main Area */}
            <div className="flex-1 flex flex-col gap-4 overflow-hidden">
                {selectedSkill ? (
                    <Card className="flex-1 bg-zinc-900 border-zinc-800 flex flex-col overflow-hidden shadow-xl">
                        <div className="flex justify-between items-center p-3 border-b border-zinc-800 bg-zinc-950/50">
                            <div className="flex items-center gap-3">
                                <div className="p-1.5 bg-emerald-500/10 rounded-md">
                                    <Code className="w-4 h-4 text-emerald-500" />
                                </div>
                                <div>
                                    <div className="font-bold text-zinc-200 text-sm">{selectedSkill.name}</div>
                                    <div className="text-xs text-zinc-500">SKILL.md</div>
                                </div>
                            </div>
                            <div className="flex gap-2">
                                {isEditing ? (
                                    <>
                                        <Button variant="ghost" size="sm" onClick={() => setIsEditing(false)} className="text-zinc-400 hover:text-white">Cancel</Button>
                                        <Button variant="default" size="sm" onClick={handleSave} disabled={saveSkill.isPending} className="bg-emerald-600 hover:bg-emerald-700">
                                            {saveSkill.isPending ? <Loader2 className="w-3 h-3 animate-spin mr-2" /> : <Save className="w-3 h-3 mr-2" />}
                                            Save Changes
                                        </Button>
                                    </>
                                ) : (
                                    <Button variant="outline" size="sm" onClick={() => setIsEditing(true)} className="border-zinc-700 text-zinc-300 hover:bg-zinc-800">
                                        <Edit className="w-3 h-3 mr-2" />
                                        Edit Skill
                                    </Button>
                                )}
                            </div>
                        </div>

                        <div className="flex-1 relative bg-zinc-950">
                            {isEditing ? (
                                <textarea
                                    value={skillContent}
                                    onChange={(e) => setSkillContent(e.target.value)}
                                    className="absolute inset-0 w-full h-full bg-zinc-950 text-zinc-300 font-mono text-sm p-4 resize-none focus:outline-none"
                                    spellCheck={false}
                                />
                            ) : (
                                <div className="absolute inset-0 w-full h-full overflow-y-auto p-6">
                                    <pre className="whitespace-pre-wrap font-mono text-sm text-zinc-400">
                                        {skillContent}
                                    </pre>
                                </div>
                            )}
                        </div>
                    </Card>
                ) : (
                    <div className="flex-1 flex items-center justify-center flex-col gap-4 border-2 border-dashed border-zinc-800 rounded-lg bg-zinc-900/50">
                        <div className="p-6 rounded-full bg-zinc-900">
                            <Book className="w-16 h-16 opacity-20" />
                        </div>
                        <p className="text-zinc-500">Select a skill to view instructions.</p>
                    </div>
                )}
            </div>

            {/* Create Modal */}
            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                <DialogContent className="bg-zinc-900 border-zinc-800 text-white">
                    <DialogHeader>
                        <DialogTitle>Create New Skill</DialogTitle>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-zinc-400">Skill ID (Folder Name)</label>
                            <Input value={newSkillId} onChange={e => setNewSkillId(e.target.value.toLowerCase().replace(/\s+/g, '-'))} placeholder="e.g. web-research" className="bg-zinc-950 border-zinc-700" />
                        </div>
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-zinc-400">Display Name</label>
                            <Input value={newSkillName} onChange={e => setNewSkillName(e.target.value)} placeholder="e.g. Web Research" className="bg-zinc-950 border-zinc-700" />
                        </div>
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-zinc-400">Description</label>
                            <Input value={newSkillDesc} onChange={e => setNewSkillDesc(e.target.value)} placeholder="What does this skill do?" className="bg-zinc-950 border-zinc-700" />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="ghost" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                        <Button variant="default" onClick={handleCreate} disabled={createSkill.isPending || !newSkillId} className="bg-emerald-600 hover:bg-emerald-700">
                            Create Skill
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
