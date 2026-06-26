"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Save, RotateCcw } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

type ConfigItem = { key: string; value: string; description?: string };

export default function MCPSettings() {
    const { data: rawConfig, isLoading, refetch } = trpc.config.list.useQuery();
    const config = rawConfig as ConfigItem[] | undefined;
    const [editing, setEditing] = useState<Record<string, string>>({});

    const updateMutation = trpc.config.update.useMutation({
        onSuccess: () => {
            toast.success("Configuration updated");
            setEditing({});
            refetch();
        },
        onError: (err) => {
            toast.error(`Update failed: ${err.message}`);
        }
    });

    const handleSave = (key: string) => {
        if (editing[key] === undefined) return;
        updateMutation.mutate({ key, value: editing[key] });
    };

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Global Configuration</h1>
                    <p className="text-zinc-500">
                        System-wide settings and feature flags
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button onClick={() => refetch()} variant="outline" className="border-zinc-700 hover:bg-zinc-800">
                        <RotateCcw className="mr-2 h-4 w-4" /> Refresh
                    </Button>
                </div>
            </div>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-lg font-medium text-zinc-200">System Parameters</CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="flex justify-center p-12">
                            <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                        </div>
                    ) : (config?.length ?? 0) === 0 ? (
                        <div className="text-center p-8 text-zinc-500">
                            No configuration parameters defined.
                        </div>
                    ) : (
                        <div className="divide-y divide-zinc-800">
                            {config?.map((item: any) => (
                                <div key={item.key} className="py-4 flex items-center justify-between group">
                                    <div className="flex-1 pr-8">
                                        <div className="font-mono text-sm text-blue-400 font-medium mb-1">{item.key}</div>
                                        <p className="text-xs text-zinc-500">{item.description || "No description provided."}</p>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <input
                                            value={editing[item.key] !== undefined ? editing[item.key] : item.value}
                                            onChange={(e) => setEditing({ ...editing, [item.key]: e.target.value })}
                                            className={`bg-zinc-950 border rounded px-3 py-1.5 text-sm text-white font-mono min-w-[300px] outline-none ${editing[item.key] !== undefined && editing[item.key] !== item.value
                                                ? 'border-yellow-500/50 ring-1 ring-yellow-500/20'
                                                : 'border-zinc-800 focus:border-blue-500'
                                                }`}
                                        />
                                        {editing[item.key] !== undefined && editing[item.key] !== item.value && (
                                            <Button
                                                size="sm"
                                                onClick={() => handleSave(item.key)}
                                                disabled={updateMutation.isPending}
                                                className="bg-green-600 hover:bg-green-500 h-8"
                                            >
                                                {updateMutation.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Save className="h-3 w-3" />}
                                            </Button>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
