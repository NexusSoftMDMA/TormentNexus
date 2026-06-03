"use client";

import Link from 'next/link';
import { Card, CardHeader, CardTitle, CardContent, Button, Badge } from '@tormentnexus/ui';
import { trpc } from '@/utils/trpc';
import { Terminal, Globe, Wrench, Layers, ExternalLink, ShieldCheck, Cpu, Box, Clock } from 'lucide-react';
import { normalizeShellHistory } from './tools-page-normalizers';

export default function ToolsRegistryDashboard() {
    const shellHistoryQuery = trpc.shell.getSystemHistory.useQuery({ limit: 10 }, { refetchInterval: 5000 });
    const { data: shellHistory, isLoading } = shellHistoryQuery;
    const normalizedShellHistory = normalizeShellHistory(shellHistory);
    const shellHistoryUnavailable = shellHistoryQuery.isError || (shellHistory != null && !Array.isArray(shellHistory));

    return (
        <div className="p-8 space-y-8 h-full flex flex-col">
            <div className="flex justify-between items-center shrink-0">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <Cpu className="h-8 w-8 text-fuchsia-500" />
                        Tools & Extensions
                    </h1>
                    <p className="text-zinc-500 mt-2">
                        Unified governance hub for Agent capabilities and external environment access
                    </p>
                </div>
            </div>

            {/* Quick Links Row */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 shrink-0">
                <Link
                    href="/dashboard/mcp/tool-sets"
                    title="Open Tool Sets to manage curated capability bundles per agent persona"
                    aria-label="Open MCP Tool Sets dashboard"
                    className="group"
                >
                    <Card className="bg-zinc-900 border-zinc-800 hover:border-indigo-500/50 transition-colors h-full overflow-hidden relative">
                        <div className="absolute inset-0 bg-indigo-500/5 opacity-0 group-hover:opacity-100 transition-opacity" />
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-bold text-zinc-300 uppercase tracking-widest flex items-center gap-2">
                                <Layers className="h-4 w-4 text-indigo-400 group-hover:scale-110 transition-transform" />
                                Tool Sets
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p className="text-xs text-zinc-500">Manage curated collections of capabilities assigned to specific Agent personas.</p>
                            <div className="mt-4 flex items-center text-xs text-indigo-400 font-medium">
                                Configure <ExternalLink className="h-3 w-3 ml-1" />
                            </div>
                        </CardContent>
                    </Card>
                </Link>

                <Link
                    href="/dashboard/browser"
                    title="Open Semantic Browser service controls and active viewport sessions"
                    aria-label="Open Semantic Browser dashboard"
                    className="group"
                >
                    <Card className="bg-zinc-900 border-zinc-800 hover:border-blue-500/50 transition-colors h-full overflow-hidden relative">
                        <div className="absolute inset-0 bg-blue-500/5 opacity-0 group-hover:opacity-100 transition-opacity" />
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-bold text-zinc-300 uppercase tracking-widest flex items-center gap-2">
                                <Globe className="h-4 w-4 text-blue-400 group-hover:scale-110 transition-transform" />
                                Semantic Browser
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p className="text-xs text-zinc-500">Monitor active headless Viewports granting Agents read/write access to the Web.</p>
                            <div className="mt-4 flex items-center text-xs text-blue-400 font-medium">
                                Access Service <ExternalLink className="h-3 w-3 ml-1" />
                            </div>
                        </CardContent>
                    </Card>
                </Link>

                <Link
                    href="/dashboard/marketplace"
                    title="Open Extensions Marketplace to discover and audit installable tools"
                    aria-label="Open Extensions Marketplace dashboard"
                    className="group"
                >
                    <Card className="bg-zinc-900 border-zinc-800 hover:border-emerald-500/50 transition-colors h-full overflow-hidden relative">
                        <div className="absolute inset-0 bg-emerald-500/5 opacity-0 group-hover:opacity-100 transition-opacity" />
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-bold text-zinc-300 uppercase tracking-widest flex items-center gap-2">
                                <Box className="h-4 w-4 text-emerald-400 group-hover:scale-110 transition-transform" />
                                Extensions Marketplace
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p className="text-xs text-zinc-500">Discover, install, and audit verified tools from the community and official repos.</p>
                            <div className="mt-4 flex items-center text-xs text-emerald-400 font-medium">
                                Browse <ExternalLink className="h-3 w-3 ml-1" />
                            </div>
                        </CardContent>
                    </Card>
                </Link>
            </div>

            {/* Shell Router History Component */}
            <Card className="flex-1 bg-zinc-900 border-zinc-800 flex flex-col shadow-xl min-h-0">
                <CardHeader className="bg-black/20 border-b border-white/5 pb-4 shrink-0">
                    <div className="flex justify-between items-center">
                        <div>
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Terminal className="h-5 w-5 text-fuchsia-400" />
                                Host Environment Terminal History
                            </CardTitle>
                            <p className="text-xs text-zinc-500 mt-1">
                                Audit log of POSIX/pwsh shell commands executed by Agent tools
                            </p>
                        </div>
                        <Badge variant="outline" className="border-fuchsia-500/30 text-fuchsia-400 bg-fuchsia-500/10 flex gap-2 items-center">
                            <ShieldCheck className="h-3 w-3" />
                            Secure Mode
                        </Badge>
                    </div>
                </CardHeader>
                <CardContent className="flex-1 overflow-auto p-0 bg-[#0c0c0c] font-mono text-[11px] lg:text-xs">
                    {isLoading ? (
                        <div className="p-8 text-zinc-600 animate-pulse flex items-center gap-2">
                            <span className="w-2 h-4 bg-fuchsia-500 animate-bounce"></span>
                            Loading shell history...
                        </div>
                    ) : shellHistoryUnavailable ? (
                        <div className="p-8 text-red-300">
                            Shell history unavailable{shellHistoryQuery.isError ? `: ${shellHistoryQuery.error.message}` : ' due to malformed data'}.
                        </div>
                    ) : normalizedShellHistory.length === 0 ? (
                        <div className="p-8 text-zinc-600">
                            <span className="text-emerald-500">server@tormentnexus</span><span className="text-zinc-400">:</span><span className="text-blue-500">~</span>$ No commands logged in recent history.
                        </div>
                    ) : (
                        <ul className="divide-y divide-white/5">
                            {normalizedShellHistory.map((entry, i: number) => (
                                <li key={entry.id} className="p-3 hover:bg-white/[0.02] transition-colors group">
                                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-1">
                                        <div className="flex items-center gap-2 text-zinc-400">
                                            <span className="text-emerald-500">agent@tormentnexus</span>
                                            <span className="text-zinc-600">:</span>
                                            <span className="text-blue-400 truncate max-w-[200px]" title={entry.cwd}>{entry.cwd || '~'}</span>
                                            <span className="text-zinc-600">$</span>
                                            <span className="text-zinc-200 font-bold ml-1 break-all">{entry.command}</span>
                                        </div>
                                        <div className="flex items-center gap-3 shrink-0">
                                            {typeof entry.duration === 'number' && (
                                                <span className="text-zinc-600 flex items-center gap-1 text-[10px]">
                                                    <Clock className="w-3 h-3" />
                                                    {entry.duration}ms
                                                </span>
                                            )}
                                            {entry.exitCode !== undefined && (
                                                <Badge variant="outline" className={`rounded-sm px-1.5 py-0 items-center justify-center font-bold tracking-tighter text-[9px]
                                                    ${entry.exitCode === 0 ? 'bg-green-500/10 text-green-500 border-green-500/20' : 'bg-red-500/10 text-red-500 border-red-500/20'}
                                                `}>
                                                    EXIT {entry.exitCode}
                                                </Badge>
                                            )}
                                        </div>
                                    </div>
                                    {entry.outputSnippet && (
                                        <pre className="mt-2 p-2 bg-black rounded border border-white/5 text-zinc-400 overflow-x-auto whitespace-pre-wrap max-h-[100px] text-[10px] leading-relaxed group-hover:border-white/10 transition-colors">
                                            {entry.outputSnippet}
                                        </pre>
                                    )}
                                </li>
                            ))}
                        </ul>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
