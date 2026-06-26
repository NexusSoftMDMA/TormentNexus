"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Loader2, Wrench, Search, ArrowUpRight } from "lucide-react";
import { trpc } from '@/utils/trpc';

export default function CatalogDashboard() {
    const { data: tools, isLoading } = trpc.tools.list.useQuery();
    const [filter, setFilter] = useState('');

    const filteredTools = tools?.filter((tool: any) =>
        tool.name.toLowerCase().includes(filter.toLowerCase()) ||
        (tool.description || '').toLowerCase().includes(filter.toLowerCase()) ||
        tool.server.toLowerCase().includes(filter.toLowerCase())
    ) || [];

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Tool Catalog</h1>
                    <p className="text-zinc-500">
                        Searchable index of all capabilities available to the system
                    </p>
                </div>
            </div>

            <div className="relative">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
                <input
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    placeholder="Search tools by name, description, or server..."
                    className="w-full bg-zinc-900 border border-zinc-800 rounded-md p-3 pl-10 text-sm text-white focus:ring-1 focus:ring-purple-500 outline-none"
                />
            </div>

            {isLoading ? (
                <div className="flex justify-center p-12">
                    <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                </div>
            ) : filteredTools.length === 0 ? (
                <div className="text-center p-12 text-zinc-500 bg-zinc-900/50 rounded-lg border border-zinc-800 border-dashed">
                    <Wrench className="h-12 w-12 mx-auto mb-4 opacity-30" />
                    <p className="text-lg font-medium">No Tools Found</p>
                    <p className="text-sm mt-1">Try adjusting your search terms.</p>
                </div>
            ) : (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {filteredTools.map((tool: any) => (
                        <Card key={tool.uuid} className="bg-zinc-900 border-zinc-800 hover:border-zinc-700 transition-colors group">
                            <CardHeader className="pb-2">
                                <CardTitle className="text-base font-medium text-zinc-200 flex items-center justify-between">
                                    <span className="font-mono text-blue-400">{tool.name}</span>
                                    <span className="text-[10px] bg-zinc-800 px-2 py-0.5 rounded text-zinc-500">
                                        {tool.server}
                                    </span>
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="text-sm text-zinc-400 line-clamp-3 min-h-[60px]">
                                    {tool.description || "No description provided."}
                                </p>
                                <div className="mt-4 pt-4 border-t border-zinc-800 flex justify-between items-center text-xs text-zinc-500">
                                    <span>
                                        Schemas: {tool.inputSchema ? Object.keys(tool.inputSchema.properties || {}).length : 0} params
                                    </span>
                                    <ArrowUpRight className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}
        </div>
    );
}
