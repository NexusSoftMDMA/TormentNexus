'use client';

import React, { useState, useEffect } from 'react';
import { trpc } from '../utils/trpc';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Textarea } from './ui/textarea';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Loader2, Save, FileCode, Plus, RefreshCw, Box } from 'lucide-react';

export function WorkshopPage() {
    const [selectedFile, setSelectedFile] = useState<string | null>(null);
    const [code, setCode] = useState('');
    const [isSaving, setIsSaving] = useState(false);
    const [files, setFiles] = useState<string[]>([]);
    const [statusMsg, setStatusMsg] = useState('');

    // Generator State
    const [genName, setGenName] = useState('');
    const [genDesc, setGenDesc] = useState('');
    const [isGenerating, setIsGenerating] = useState(false);

    const executeToolMutation = trpc.executeTool.useMutation();

    const fetchFiles = async () => {
        try {
            const res = await executeToolMutation.mutateAsync({
                name: 'list_files',
                args: { path: 'packages/tools/src' }
            });
            // executeTool returns result.content[0].text
            // list_files returns a JSON string of file array
            const content = res;
            // The executeTool mutation in trpc.ts returns result.content[0].text directly.

            const fileList = JSON.parse(content).filter((f: any) => !f.isDirectory && f.name.endsWith('.ts'));
            // Sort files
            fileList.sort((a: any, b: any) => a.name.localeCompare(b.name));
            setFiles(fileList.map((f: any) => f.name));
        } catch (e) {
            console.error("Failed to list files:", e);
            setStatusMsg("Error listing files");
        }
    };

    useEffect(() => {
        fetchFiles();
    }, []);

    const loadFile = async (fileName: string) => {
        setSelectedFile(fileName);
        setStatusMsg("Loading...");
        try {
            const res = await executeToolMutation.mutateAsync({
                name: 'read_tool_source',
                args: { fileName }
            });
            setCode(res); // executeTool returns file content string
            setStatusMsg("");
        } catch (e) {
            console.error(e);
            setCode("");
            setStatusMsg("Error loading file");
        }
    };

    const saveFile = async () => {
        if (!selectedFile) return;
        setIsSaving(true);
        setStatusMsg("Saving...");
        try {
            await executeToolMutation.mutateAsync({
                name: 'update_tool_source',
                args: { fileName: selectedFile, content: code }
            });
            setStatusMsg("Saved!");
            setTimeout(() => setStatusMsg(""), 2000);
        } catch (e) {
            console.error(e);
            setStatusMsg("Failed to save");
        }
        setIsSaving(false);
    };

    const generateTool = async () => {
        if (!genName) return;
        setIsGenerating(true);
        const fileName = genName.endsWith('.ts') ? genName : `${genName}.ts`;
        const toolName = genName.replace('.ts', '');

        try {
            await executeToolMutation.mutateAsync({
                name: 'create_tool_scaffold',
                args: {
                    fileName,
                    toolName,
                    functionName: toolName.toLowerCase() + '_action',
                    description: genDesc || 'A new tool'
                }
            });
            setStatusMsg("Generated " + fileName);
            setGenName('');
            setGenDesc('');
            fetchFiles();
        } catch (e) {
            console.error(e);
            setStatusMsg("Failed to generate");
        }
        setIsGenerating(false);
    };

    return (
        <div className="flex h-full gap-4 p-4 text-white overflow-hidden">
            {/* Sidebar List */}
            <div className="w-72 flex flex-col gap-4">
                <Card className="bg-zinc-900 border-zinc-800 p-4 space-y-3 shadow-lg">
                    <div className="flex items-center gap-2 text-indigo-400">
                        <Box className="w-5 h-5" />
                        <span className="font-bold text-sm tracking-widest uppercase">Tool Generator</span>
                    </div>

                    <div className="space-y-2">
                        <Input
                            placeholder="MyNewTool"
                            value={genName}
                            onChange={e => setGenName(e.target.value)}
                            className="bg-zinc-950 border-zinc-700 h-9 text-sm"
                        />
                        <Input
                            placeholder="Description..."
                            value={genDesc}
                            onChange={e => setGenDesc(e.target.value)}
                            className="bg-zinc-950 border-zinc-700 h-9 text-sm"
                        />
                        <Button
                            className="w-full bg-indigo-600 hover:bg-indigo-700 h-9 text-sm"
                            onClick={generateTool}
                            disabled={isGenerating || !genName}
                        >
                            {isGenerating ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plus className="w-4 h-4 mr-2" />}
                            Scaffold Tool
                        </Button>
                    </div>
                </Card>

                <Card className="bg-zinc-900 border-zinc-800 flex-1 flex flex-col overflow-hidden shadow-lg">
                    <CardHeader className="p-3 border-b border-zinc-800 bg-zinc-950/50">
                        <CardTitle className="text-sm font-medium flex justify-between items-center text-zinc-300">
                            package/tools/src
                            <Button size="icon" variant="ghost" className="h-6 w-6 hover:bg-zinc-800" onClick={fetchFiles}>
                                <RefreshCw className="h-4 w-4 text-zinc-500" />
                            </Button>
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="p-0 overflow-y-auto flex-1 custom-scrollbar">
                        {files.map(f => (
                            <button
                                key={f}
                                onClick={() => loadFile(f)}
                                className={`w-full text-left px-4 py-2 text-sm border-l-2 transition-colors ${selectedFile === f
                                        ? 'bg-zinc-800/50 border-indigo-500 text-indigo-300'
                                        : 'border-transparent text-zinc-400 hover:bg-zinc-800/30 hover:text-zinc-200'
                                    }`}
                            >
                                {f}
                            </button>
                        ))}
                    </CardContent>
                </Card>
            </div>

            {/* Editor */}
            <div className="flex-1 flex flex-col gap-4 overflow-hidden">
                {selectedFile ? (
                    <Card className="flex-1 bg-zinc-900 border-zinc-800 flex flex-col overflow-hidden shadow-xl">
                        <div className="flex justify-between items-center p-3 border-b border-zinc-800 bg-zinc-950/50">
                            <div className="flex items-center gap-2">
                                <FileCode className="w-4 h-4 text-zinc-500" />
                                <span className="font-mono text-sm text-zinc-300 font-medium">{selectedFile}</span>
                                {statusMsg && <span className="text-xs text-yellow-500 ml-2 animate-pulse">{statusMsg}</span>}
                            </div>
                            <Button size="sm" onClick={saveFile} disabled={isSaving} className="bg-emerald-600 hover:bg-emerald-700 h-8 text-xs font-medium">
                                {isSaving ? <Loader2 className="w-3 h-3 animate-spin mr-2" /> : <Save className="w-3 h-3 mr-2" />}
                                Save Changes
                            </Button>
                        </div>
                        <div className="flex-1 relative">
                            <Textarea
                                value={code}
                                onChange={e => setCode(e.target.value)}
                                className="absolute inset-0 w-full h-full font-mono text-sm bg-zinc-950 text-zinc-300 border-none p-4 resize-none focus-visible:ring-0 leading-relaxed"
                                spellCheck={false}
                            />
                        </div>
                    </Card>
                ) : (
                    <div className="flex-1 flex items-center justify-center text-zinc-600 flex-col gap-4 border-2 border-dashed border-zinc-800 rounded-lg bg-zinc-900/50">
                        <div className="p-6 rounded-full bg-zinc-900">
                            <FileCode className="w-16 h-16 opacity-20" />
                        </div>
                        <div className="text-center">
                            <h3 className="text-lg font-medium text-zinc-400">The Workshop</h3>
                            <p className="text-sm text-zinc-600 max-w-md mt-2">
                                Select a tool from the sidebar to view or edit its source code.
                                Use the generator to scaffold new capabilities.
                            </p>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
