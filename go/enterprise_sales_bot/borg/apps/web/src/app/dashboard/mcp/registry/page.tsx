"use client";

import { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Globe, Download, ExternalLink, Database } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

type RegistryItem = {
    id?: string;
    name: string;
    description: string;
    author?: string;
    command?: string;
    args?: string[];
    env?: Record<string, string>;
    tags: string[];
    url?: string;
    category?: string;
};

// Fallback install templates if live registry has no install metadata
const QUICK_INSTALL_TEMPLATES: RegistryItem[] = [
    {
        name: "filesystem",
        description: "Standard filesystem operations (read/write/list)",
        author: "ModelContextProtocol",
        command: "npx",
        args: ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/dir"],
        tags: ["official", "core"]
    },
    {
        name: "memory",
        description: "Knowledge graph memory server",
        author: "ModelContextProtocol",
        command: "npx",
        args: ["-y", "@modelcontextprotocol/server-memory"],
        tags: ["official", "core", "memory"]
    },
    {
        name: "brave-search",
        description: "Web search using Brave API",
        author: "ModelContextProtocol",
        command: "npx",
        args: ["-y", "@modelcontextprotocol/server-brave-search"],
        env: { "BRAVE_API_KEY": "YOUR_KEY_HERE" },
        tags: ["official", "search"]
    },
    {
        name: "github",
        description: "GitHub repository management and issue tracking",
        author: "ModelContextProtocol",
        command: "npx",
        args: ["-y", "@modelcontextprotocol/server-github"],
        env: { "GITHUB_PERSONAL_ACCESS_TOKEN": "YOUR_TOKEN" },
        tags: ["official", "dev"]
    },
    {
        name: "postgres",
        description: "Read-only database inspection",
        author: "ModelContextProtocol",
        command: "npx",
        args: ["-y", "@modelcontextprotocol/server-postgres", "postgresql://user:password@localhost/db"],
        tags: ["official", "database"]
    }
];

export default function RegistryDashboard() {
    const [filter, setFilter] = useState('');

    // We can use this to check which are already installed
    const { data: installedServers } = trpc.mcpServers.list.useQuery();
    const { data: registry, isLoading: loadingRegistry } = trpc.mcpServers.registrySnapshot.useQuery();

    const liveRegistry: RegistryItem[] = (registry || []).map((item: any) => ({
        id: item.id,
        name: item.name,
        description: item.description,
        tags: Array.isArray(item.tags) ? item.tags : [],
        url: item.url,
        category: item.category,
    }));

    const source = liveRegistry.length > 0 ? liveRegistry : QUICK_INSTALL_TEMPLATES;
    const filtered = source.filter((item) =>
        item.name.toLowerCase().includes(filter.toLowerCase()) ||
        item.description.toLowerCase().includes(filter.toLowerCase())
    );

    return (
        <div className="p-8 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Public Registry</h1>
                    <p className="text-zinc-500">
                        Discover and install community MCP servers
                    </p>
                </div>
                <div className="text-xs text-zinc-500 flex items-center gap-2">
                    <Database className="h-4 w-4" />
                    {liveRegistry.length > 0 ? `Live index (${liveRegistry.length})` : 'Fallback templates'}
                </div>
            </div>

            <div className="relative">
                <input
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    placeholder="Search registry..."
                    className="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-md p-3 pl-4 text-sm text-white focus:ring-1 focus:ring-blue-500 outline-none"
                />
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {loadingRegistry ? (
                    <div className="col-span-full flex items-center justify-center py-16 text-zinc-500">
                        <Loader2 className="h-5 w-5 animate-spin mr-2" /> Loading live registry...
                    </div>
                ) : filtered.map((item) => (
                    <RegistryCard
                        key={item.id || item.name}
                        item={item}
                        isInstalled={!!installedServers?.some((s: any) => s.name === item.name)}
                    />
                ))}
            </div>
        </div>
    );
}

function RegistryCard({ item, isInstalled }: { item: RegistryItem; isInstalled: boolean }) {
    const installMutation = trpc.mcpServers.create.useMutation({
        onSuccess: () => {
            toast.success(`Installed ${item.name}`);
        },
        onError: (err) => {
            toast.error(`Install failed: ${err.message}`);
        }
    });

    const handleInstall = () => {
        if (!item.command || !item.args) return;
        // Prepare config, prompting for ENV if needed (simplified here)
        installMutation.mutate({
            name: item.name,
            type: 'STDIO',
            command: item.command,
            args: item.args,
            env: item.env || {},
        });
    };

    return (
        <Card className="bg-zinc-900 border-zinc-800 hover:border-zinc-700 transition-colors flex flex-col">
            <CardHeader className="pb-2">
                <CardTitle className="text-lg font-medium text-zinc-200 flex items-center justify-between">
                    <span className="flex items-center gap-2">
                        <Globe className="h-4 w-4 text-blue-400" />
                        {item.name}
                    </span>
                    {isInstalled && (
                        <span className="text-[10px] bg-green-500/10 text-green-500 border border-green-500/20 px-2 py-0.5 rounded">
                            INSTALLED
                        </span>
                    )}
                </CardTitle>
                <div className="text-xs text-zinc-500">by {item.author}</div>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col justify-between space-y-4">
                <p className="text-sm text-zinc-400">
                    {item.description}
                </p>

                {item.url ? (
                    <a
                        href={item.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-xs text-blue-400 hover:text-blue-300 inline-flex items-center gap-1"
                    >
                        <ExternalLink className="h-3 w-3" /> Open Source
                    </a>
                ) : null}

                <div className="flex flex-wrap gap-1">
                    {item.tags.map((tag: string) => (
                        <span key={tag} className="px-1.5 py-0.5 bg-zinc-800 rounded text-[10px] text-zinc-500">
                            #{tag}
                        </span>
                    ))}
                </div>

                <Button
                    onClick={handleInstall}
                    disabled={isInstalled || installMutation.isPending || !item.command || !item.args}
                    className="w-full bg-blue-600 hover:bg-blue-500 disabled:bg-zinc-800 disabled:text-zinc-500"
                >
                    {installMutation.isPending ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : isInstalled ? (
                        "Installed"
                    ) : !item.command || !item.args ? (
                        "Manual Setup"
                    ) : (
                        <>
                            <Download className="mr-2 h-4 w-4" /> Install Server
                        </>
                    )}
                </Button>
            </CardContent>
        </Card>
    );
}
