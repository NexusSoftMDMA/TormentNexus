'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { ScrollArea } from './ui/scroll-area';
import { Badge } from './ui/badge';
import { X, Plus, Trash2, FileText, Pin } from 'lucide-react';
import { trpc } from '../utils/trpc';

function normalizePinnedFiles(value: unknown): string[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value.filter((entry): entry is string => typeof entry === 'string');
}

export function ContextPanel() {
    const [newFile, setNewFile] = useState('');

    const contextQuery = trpc.tormentnexusContext.list.useQuery(undefined, {
        refetchInterval: 5000
    });
    const { data: rawPinnedFiles, refetch } = contextQuery;
    const pinnedFiles = normalizePinnedFiles(rawPinnedFiles);

    const addMutation = trpc.tormentnexusContext.add.useMutation({
        onSuccess: () => {
            setNewFile('');
            refetch();
        }
    });

    const removeMutation = trpc.tormentnexusContext.remove.useMutation({
        onSuccess: () => refetch()
    });

    const clearMutation = trpc.tormentnexusContext.clear.useMutation({
        onSuccess: () => refetch()
    });

    const handleAdd = () => {
        if (newFile.trim()) {
            addMutation.mutate({ filePath: newFile.trim() });
        }
    };

    const handleRemove = (filePath: string) => {
        removeMutation.mutate({ filePath });
    };

    const handleClear = () => {
        clearMutation.mutate();
    };

    return (
        <Card className="h-full flex flex-col">
            <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-medium flex items-center gap-2">
                        <Pin className="w-4 h-4" />
                        Pinned Context
                    </CardTitle>
                    <Badge variant="secondary" className="text-xs">
                        {pinnedFiles.length} files
                    </Badge>
                </div>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col gap-2 overflow-hidden">
                {/* Add file input */}
                <div className="flex gap-2">
                    <Input
                        placeholder="Add file path..."
                        value={newFile}
                        onChange={(e) => setNewFile(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
                        className="flex-1 h-8 text-xs"
                    />
                    <Button size="sm" onClick={handleAdd} disabled={!newFile.trim()}>
                        <Plus className="w-3 h-3" />
                    </Button>
                </div>

                {/* Pinned files list */}
                <ScrollArea className="flex-1">
                    <div className="space-y-1">
                        {pinnedFiles.length === 0 ? (
                            <p className="text-xs text-muted-foreground text-center py-4">
                                No files pinned. Add files to include them in context.
                            </p>
                        ) : (
                            pinnedFiles.map((file) => (
                                <div
                                    key={file}
                                    className="flex items-center justify-between p-2 bg-muted/50 rounded-md group"
                                >
                                    <div className="flex items-center gap-2 overflow-hidden">
                                        <FileText className="w-3 h-3 shrink-0 text-muted-foreground" />
                                        <span className="text-xs truncate" title={file}>
                                            {file.split(/[/\\]/).pop()}
                                        </span>
                                    </div>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                                        onClick={() => handleRemove(file)}
                                    >
                                        <X className="w-3 h-3" />
                                    </Button>
                                </div>
                            ))
                        )}
                    </div>
                </ScrollArea>

                {/* Clear all button */}
                {pinnedFiles.length > 0 && (
                    <Button
                        variant="outline"
                        size="sm"
                        className="w-full"
                        onClick={handleClear}
                    >
                        <Trash2 className="w-3 h-3 mr-1" />
                        Clear All
                    </Button>
                )}
            </CardContent>
        </Card>
    );
}
