"use client";

import { trpc } from "@/utils/trpc";
import { useState } from "react";
import {
    Search, Download, CheckCircle, Smartphone,
    Box, Cpu, Globe, Share2
} from "lucide-react";
import { toast } from "sonner";

export default function MarketplacePage() {
    const [filter, setFilter] = useState("");
    const { data: entries, isLoading, refetch } = trpc.marketplace.list.useQuery({ filter });
    const utils = trpc.useContext();

    const installMutation = trpc.marketplace.install.useMutation({
        onSuccess: () => {
            toast.success("Installation successful");
            utils.marketplace.list.invalidate();
        },
        onError: (error) => {
            toast.error("Installation failed: " + error.message);
        }
    });

    const isInstalling = installMutation.isPending;

    const handleInstall = (id: string) => {
        installMutation.mutate({ id });
    };

    return (
        <div className="p-8 max-w-7xl mx-auto space-y-8">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white mb-2">Marketplace</h1>
                    <p className="text-zinc-400">Discover and install AI agents, tools, and skills.</p>
                </div>
                <div className="flex gap-2">
                    <button className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded-md text-sm font-medium transition-colors border border-zinc-700 flex items-center gap-2">
                        <Share2 className="w-4 h-4" />
                        Publish
                    </button>
                </div>
            </div>

            {/* Search Bar */}
            <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
                <input
                    className="w-full h-12 bg-zinc-900/50 border border-zinc-800 rounded-lg pl-10 pr-4 text-zinc-200 focus:outline-none focus:border-indigo-500/50 transition-colors placeholder:text-zinc-600"
                    placeholder="Search for agents, tools, or skills..."
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                />
            </div>

            {/* Grid */}
            {isLoading ? (
                <div className="text-zinc-500">Loading marketplace...</div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {entries?.map((entry) => (
                        <div
                            key={entry.id}
                            className="group relative bg-zinc-900/40 border border-zinc-800 hover:border-zinc-700 hover:bg-zinc-900/60 rounded-xl p-6 transition-all duration-300"
                        >
                            <div className="flex items-start justify-between mb-4">
                                <div className="flex items-center gap-3">
                                    <div className={`w-10 h-10 rounded-lg flex items-center justify-center border border-zinc-800 ${entry.type === 'agent' ? 'bg-indigo-500/10 text-indigo-400' :
                                            entry.type === 'tool' ? 'bg-emerald-500/10 text-emerald-400' :
                                                'bg-amber-500/10 text-amber-400'
                                        }`}>
                                        {entry.type === 'agent' && <BotIcon />}
                                        {entry.type === 'tool' && <WrenchIcon />}
                                        {entry.type === 'skill' && <BookIcon />}
                                    </div>
                                    <div>
                                        <h3 className="font-semibold text-zinc-100">{entry.name}</h3>
                                        <p className="text-xs text-zinc-500 capitalize">{entry.type} • {entry.source}</p>
                                    </div>
                                </div>
                                {entry.installed ? (
                                    <span className="bg-emerald-500/10 text-emerald-400 text-xs px-2 py-1 rounded-full flex items-center gap-1 border border-emerald-500/20">
                                        <CheckCircle className="w-3 h-3" />
                                        Installed
                                    </span>
                                ) : (
                                    <button
                                        onClick={() => handleInstall(entry.id)}
                                        disabled={isInstalling}
                                        className="bg-zinc-100 text-zinc-900 hover:bg-white disabled:opacity-50 text-xs px-3 py-1.5 rounded-full font-medium transition-colors flex items-center gap-1.5"
                                    >
                                        <Download className="w-3 h-3" />
                                        Install
                                    </button>
                                )}
                            </div>

                            <p className="text-sm text-zinc-400 line-clamp-2 mb-4 h-10">
                                {entry.description}
                            </p>

                            <div className="flex items-center gap-2 flex-wrap">
                                {entry.tags?.slice(0, 3).map(tag => (
                                    <span key={tag} className="text-[10px] bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded border border-zinc-700/50">
                                        {tag}
                                    </span>
                                ))}
                                {entry.verified && (
                                    <span className="text-[10px] bg-blue-500/10 text-blue-400 px-2 py-0.5 rounded border border-blue-500/20 ml-auto">
                                        Verified
                                    </span>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

function BotIcon() {
    return <Cpu className="w-5 h-5" />;
}

function WrenchIcon() {
    return <Box className="w-5 h-5" />;
}

function BookIcon() {
    return <Globe className="w-5 h-5" />;
}
