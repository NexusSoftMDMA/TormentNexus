'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Badge } from './ui/badge';
import { ScrollArea } from './ui/scroll-area';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { X, Plus, Code, Hash, Braces, Variable, FileCode, GripVertical } from 'lucide-react';
import { trpc } from '../utils/trpc';

const typeIcons: Record<string, React.ReactNode> = {
    function: <Code className="w-3 h-3" />,
    class: <Braces className="w-3 h-3" />,
    method: <Hash className="w-3 h-3" />,
    variable: <Variable className="w-3 h-3" />,
    interface: <FileCode className="w-3 h-3" />,
};

export function SymbolPinPanel() {
    const [newSymbol, setNewSymbol] = useState({ name: '', file: '', type: 'function' as const });

    const { data: symbols = [], refetch } = trpc.symbols.list.useQuery(undefined, {
        refetchInterval: 5000,
    });

    const pinMutation = trpc.symbols.pin.useMutation({
        onSuccess: () => {
            setNewSymbol({ name: '', file: '', type: 'function' });
            refetch();
        }
    });

    const unpinMutation = trpc.symbols.unpin.useMutation({
        onSuccess: () => refetch()
    });

    const clearMutation = trpc.symbols.clear.useMutation({
        onSuccess: () => refetch()
    });

    const handlePin = () => {
        if (newSymbol.name && newSymbol.file) {
            pinMutation.mutate(newSymbol);
        }
    };

    return (
        <Card className="h-full flex flex-col">
            <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-medium flex items-center gap-2">
                        <Code className="w-4 h-4" />
                        Symbol Pins
                    </CardTitle>
                    <Badge variant="secondary" className="text-xs">
                        {symbols.length} pinned
                    </Badge>
                </div>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col gap-3 overflow-hidden">
                {/* Add symbol form */}
                <div className="space-y-2">
                    <Input
                        placeholder="Symbol name (e.g. MCPServer.executeTool)"
                        value={newSymbol.name}
                        onChange={(e) => setNewSymbol(s => ({ ...s, name: e.target.value }))}
                        className="h-8 text-xs"
                    />
                    <div className="flex gap-2">
                        <Input
                            placeholder="File path"
                            value={newSymbol.file}
                            onChange={(e) => setNewSymbol(s => ({ ...s, file: e.target.value }))}
                            className="flex-1 h-8 text-xs"
                        />
                        <Select
                            value={newSymbol.type}
                            onValueChange={(v: any) => setNewSymbol(s => ({ ...s, type: v }))}
                        >
                            <SelectTrigger className="w-24 h-8 text-xs">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="function">Function</SelectItem>
                                <SelectItem value="class">Class</SelectItem>
                                <SelectItem value="method">Method</SelectItem>
                                <SelectItem value="variable">Variable</SelectItem>
                                <SelectItem value="interface">Interface</SelectItem>
                            </SelectContent>
                        </Select>
                        <Button size="sm" onClick={handlePin} disabled={!newSymbol.name || !newSymbol.file}>
                            <Plus className="w-3 h-3" />
                        </Button>
                    </div>
                </div>

                {/* Pinned symbols list */}
                <ScrollArea className="flex-1">
                    <div className="space-y-1">
                        {symbols.length === 0 ? (
                            <p className="text-xs text-muted-foreground text-center py-4">
                                Pin symbols to prioritize them in context.
                            </p>
                        ) : (
                            symbols.map((sym: any) => (
                                <div
                                    key={sym.id}
                                    className="flex items-center gap-2 p-2 bg-muted/50 rounded-md group"
                                >
                                    <GripVertical className="w-3 h-3 text-muted-foreground cursor-grab" />
                                    {typeIcons[sym.type]}
                                    <div className="flex-1 overflow-hidden">
                                        <div className="text-xs font-medium truncate">{sym.name}</div>
                                        <div className="text-[10px] text-muted-foreground truncate">
                                            {sym.file.split(/[/\\]/).pop()}
                                        </div>
                                    </div>
                                    <Badge variant="outline" className="text-[10px]">
                                        #{sym.priority}
                                    </Badge>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 opacity-0 group-hover:opacity-100"
                                        onClick={() => unpinMutation.mutate({ id: sym.id })}
                                    >
                                        <X className="w-3 h-3" />
                                    </Button>
                                </div>
                            ))
                        )}
                    </div>
                </ScrollArea>

                {symbols.length > 0 && (
                    <Button variant="outline" size="sm" className="w-full" onClick={() => clearMutation.mutate()}>
                        Clear All
                    </Button>
                )}
            </CardContent>
        </Card>
    );
}
