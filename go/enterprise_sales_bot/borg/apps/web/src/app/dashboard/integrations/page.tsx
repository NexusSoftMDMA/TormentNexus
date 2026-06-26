"use client";

import { useState } from 'react';
import type { ComponentType } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardHeader, CardTitle } from '@tormentnexus/ui';
import { Bot, Cable, Check, Copy, ExternalLink, FolderCode, Globe, Loader2, Puzzle, Settings2, Sparkles, TerminalSquare } from 'lucide-react';
import { toast } from 'sonner';

import { trpc } from '@/utils/trpc';
import {
    getExternalClientRows,
    getConnectedBridgeClientRows,
    getBridgeClientEmptyStateMessage,
    getBridgeClientStatDetail,
    getIntegrationOverview,
    getInstallSurfaceRows,
    getStatusBadgeClasses,
    type StartupStatusSummary,
} from './integration-catalog';

function StatCard({
    title,
    value,
    detail,
    icon: Icon,
    tone,
}: {
    title: string;
    value: string;
    detail: string;
    icon: ComponentType<{ className?: string }>;
    tone: string;
}) {
    return (
        <Card className="bg-zinc-900 border-zinc-800">
            <CardContent className="p-5 flex items-start justify-between gap-4">
                <div>
                    <div className="text-xs uppercase tracking-wide text-zinc-500">{title}</div>
                    <div className="mt-2 text-3xl font-semibold text-white">{value}</div>
                    <div className="mt-1 text-sm text-zinc-400">{detail}</div>
                </div>
                <div className={`rounded-full border border-zinc-800 bg-zinc-950 p-3 ${tone}`}>
                    <Icon className="h-5 w-5" />
                </div>
            </CardContent>
        </Card>
    );
}

export default function IntegrationsDashboard() {
    const [copiedActionId, setCopiedActionId] = useState<string | null>(null);
    const mcpServersClient = trpc.mcpServers as any;
    const toolsClient = trpc.tools as any;

    const startupStatusQuery = trpc.startupStatus.useQuery(undefined, { refetchInterval: 10000 });
    const browserStatusQuery = trpc.browser.status.useQuery(undefined, { refetchInterval: 5000 });
    const syncTargetsQuery = mcpServersClient.syncTargets.useQuery();
    const cliDetectionsQuery = toolsClient?.detectCliHarnesses?.useQuery
        ? toolsClient.detectCliHarnesses.useQuery()
        : ({ data: [], isLoading: false } as { data: []; isLoading: boolean });
    const installArtifactsQuery = toolsClient?.detectInstallSurfaces?.useQuery
        ? toolsClient.detectInstallSurfaces.useQuery(undefined, { refetchInterval: 10000 })
        : ({ data: [], isLoading: false } as { data: []; isLoading: boolean });
    const startupStatus: StartupStatusSummary | null = (startupStatusQuery.data ?? null) as StartupStatusSummary | null;

    const overview = getIntegrationOverview(
        startupStatus,
        browserStatusQuery.data,
        syncTargetsQuery.data,
        cliDetectionsQuery.data,
    );
    const clientRows = getExternalClientRows(syncTargetsQuery.data);
    const connectedBridgeClients = getConnectedBridgeClientRows(startupStatus);
    const installSurfaceRows = getInstallSurfaceRows(installArtifactsQuery.data);

    const isLoading = startupStatusQuery.isLoading || browserStatusQuery.isLoading || syncTargetsQuery.isLoading || cliDetectionsQuery.isLoading || installArtifactsQuery.isLoading;

    const handleCopyOperatorAction = async (surfaceId: string, value: string, successLabel: string) => {
        if (typeof navigator === 'undefined' || !navigator.clipboard) {
            toast.error('Clipboard unavailable in this browser');
            return;
        }

        try {
            await navigator.clipboard.writeText(value);
            setCopiedActionId(surfaceId);
            toast.success(`${successLabel} copied`);
            window.setTimeout(() => {
                setCopiedActionId((current) => (current === surfaceId ? null : current));
            }, 1500);
        } catch (error) {
            toast.error(`Copy failed: ${error instanceof Error ? error.message : 'Clipboard unavailable'}`);
        }
    };

    return (
        <div className="p-8 space-y-8 h-full overflow-y-auto">
            <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Integration Hub</h1>
                    <p className="mt-2 max-w-3xl text-zinc-500">
                        Install TormentNexus into the environments you actually use: browser bridges, VS Code, and MCP-aware clients.
                        This page centralizes extension package locations, supported MCP client sync targets, and live bridge readiness so setup is less treasure hunt, more control plane.
                    </p>
                </div>

                <div className="flex flex-wrap gap-2">
                    <Link href="/dashboard/browser" className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:border-zinc-600 hover:bg-zinc-800">
                        Browser runtime
                        <ExternalLink className="h-4 w-4" />
                    </Link>
                    <Link href="/dashboard/mcp/settings" className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:border-zinc-600 hover:bg-zinc-800">
                        MCP client sync
                        <ExternalLink className="h-4 w-4" />
                    </Link>
                    <Link href="/dashboard/mcp/ai-tools" className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:border-zinc-600 hover:bg-zinc-800">
                        AI tools directory
                        <ExternalLink className="h-4 w-4" />
                    </Link>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-4">
                <StatCard
                    title="Extension bridge clients"
                    value={String(overview.extensionClientCount)}
                    detail={getBridgeClientStatDetail(overview)}
                    icon={Cable}
                    tone="text-cyan-400"
                />
                <StatCard
                    title="Browser runtime"
                    value={overview.browserRuntimeReady ? 'Ready' : 'Offline'}
                    detail={`${overview.browserPageCount} active page${overview.browserPageCount === 1 ? '' : 's'} tracked`}
                    icon={Globe}
                    tone="text-emerald-400"
                />
                <StatCard
                    title="Synced MCP clients"
                    value={String(overview.syncedClientCount)}
                    detail="Detected config targets with existing TormentNexus-ready files"
                    icon={Settings2}
                    tone="text-violet-400"
                />
                <StatCard
                    title="Installed CLI harnesses"
                    value={`${overview.installedHarnessCount}/${overview.totalHarnessCount}`}
                    detail="Local coding harnesses discovered on PATH"
                    icon={Bot}
                    tone="text-amber-400"
                />
                <StatCard
                    title="Execution environment"
                    value={overview.executionPreferredShell ?? (overview.executionEnvironmentReady ? 'Ready' : 'Pending')}
                    detail={`${overview.verifiedExecutionToolCount} verified tools${overview.supportsPosixShell ? ' · POSIX available' : ''}`}
                    icon={TerminalSquare}
                    tone="text-emerald-400"
                />
            </div>

            <div className="grid grid-cols-1 xl:grid-cols-[1.2fr_0.8fr] gap-4">
                <Card className="bg-zinc-900 border-zinc-800">
                    <CardHeader>
                        <CardTitle className="text-white">Installable TormentNexus surfaces</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {installSurfaceRows.map((surface) => (
                            <div key={surface.id} className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 space-y-3">
                                <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                                    <div>
                                        <div className="flex items-center gap-2">
                                            <span className="text-sm font-semibold text-white">{surface.title}</span>
                                            <span className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-[10px] uppercase tracking-wide text-zinc-300">
                                                {surface.platforms}
                                            </span>
                                            <span className={`rounded-full border px-2 py-0.5 text-[10px] uppercase tracking-wide ${getStatusBadgeClasses(surface.statusTone)}`}>
                                                {surface.statusLabel}
                                            </span>
                                        </div>
                                        <div className="mt-2 text-xs text-zinc-400">Repo path</div>
                                        <div className="mt-1 font-mono text-xs text-zinc-300">{surface.repoPath}</div>
                                    </div>

                                    <Link href={surface.managementHref} className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:border-zinc-600 hover:bg-zinc-800">
                                        {surface.managementLabel}
                                        <ExternalLink className="h-4 w-4" />
                                    </Link>
                                </div>

                                <div className="grid gap-3 md:grid-cols-2">
                                    <div>
                                        <div className="text-[10px] uppercase tracking-wide text-zinc-500">Build / package</div>
                                        <div className="mt-1 rounded border border-zinc-800 bg-black/20 px-3 py-2 font-mono text-xs text-zinc-300">
                                            {surface.buildHint}
                                        </div>
                                        <div className="mt-3 text-[10px] uppercase tracking-wide text-zinc-500">Detected artifact</div>
                                        <div className="mt-1 rounded border border-zinc-800 bg-black/20 px-3 py-2 font-mono text-xs text-zinc-300">
                                            {surface.artifactStatus.artifactPath ?? 'Not detected yet'}
                                        </div>
                                        <div className="mt-2 text-xs text-zinc-400">{surface.artifactStatus.detail}</div>
                                        <div className="mt-3 text-[10px] uppercase tracking-wide text-zinc-500">Artifact metadata</div>
                                        <div className="mt-1 flex flex-wrap gap-2">
                                            <span className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300">
                                                {surface.artifactVersionLabel}
                                            </span>
                                            <span className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300">
                                                {surface.artifactKindLabel}
                                            </span>
                                            <span className={`rounded-full border px-2 py-1 text-[11px] ${getStatusBadgeClasses(surface.artifactFreshnessTone)}`}>
                                                {surface.artifactFreshnessLabel}
                                            </span>
                                        </div>
                                        <div className="mt-2 text-xs text-zinc-400">{surface.artifactUpdatedLabel}</div>
                                        <div className="mt-1 text-xs text-zinc-500">{surface.artifactTimestampLabel}</div>
                                    </div>
                                    <div>
                                        <div className="text-[10px] uppercase tracking-wide text-zinc-500">Install hint</div>
                                        <div className="mt-1 rounded border border-zinc-800 bg-black/20 px-3 py-2 text-xs text-zinc-300">
                                            {surface.installHint}
                                        </div>
                                        <div className="mt-3 text-[10px] uppercase tracking-wide text-zinc-500">Next step</div>
                                        <div className="mt-1 rounded border border-zinc-800 bg-black/20 px-3 py-2 text-xs text-zinc-300">
                                            <span className="font-medium text-white">{surface.nextStepLabel}</span>
                                            <div className="mt-1 text-zinc-400">{surface.nextStepDetail}</div>
                                        </div>
                                        <div className="mt-3 text-[10px] uppercase tracking-wide text-zinc-500">Operator action</div>
                                        <div className="mt-1 rounded border border-zinc-800 bg-black/20 px-3 py-2 text-xs text-zinc-300">
                                            <div className="flex items-start justify-between gap-3">
                                                <span className="font-medium text-white">{surface.operatorActionLabel}</span>
                                                <button
                                                    type="button"
                                                    className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300 transition hover:border-zinc-600 hover:bg-zinc-800 hover:text-white"
                                                    onClick={() => handleCopyOperatorAction(surface.id, surface.operatorActionValue, surface.operatorActionCopyLabel)}
                                                >
                                                    {copiedActionId === surface.id ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                                                    {copiedActionId === surface.id ? 'Copied' : surface.operatorActionCopyLabel}
                                                </button>
                                            </div>
                                            <div className="mt-1 font-mono text-[11px] text-zinc-300 break-all">{surface.operatorActionValue}</div>
                                            <div className="mt-1 text-zinc-400">{surface.operatorActionDetail}</div>
                                        </div>
                                    </div>
                                </div>

                                <div>
                                    <div className="text-[10px] uppercase tracking-wide text-zinc-500">Exposed capabilities</div>
                                    <ul className="mt-2 grid gap-2 md:grid-cols-2">
                                        {surface.capabilities.map((capability) => (
                                            <li key={capability} className="rounded border border-zinc-800 bg-black/20 px-3 py-2 text-xs text-zinc-300">
                                                {capability}
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            </div>
                        ))}
                    </CardContent>
                </Card>

                <Card className="bg-zinc-900 border-zinc-800">
                    <CardHeader>
                        <CardTitle className="text-white">Quick routing</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3 text-sm text-zinc-300">
                        <Link href="/dashboard/browser" className="flex items-start gap-3 rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 hover:border-zinc-700 hover:bg-zinc-950">
                            <Globe className="mt-0.5 h-4 w-4 text-cyan-400" />
                            <div>
                                <div className="font-medium text-white">Browser bridge & telemetry</div>
                                <div className="mt-1 text-xs text-zinc-400">History search, screenshots, proxy fetch, CDP debug, memory capture, and page-to-RAG ingestion.</div>
                            </div>
                        </Link>

                        <Link href="/dashboard/mcp/settings" className="flex items-start gap-3 rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 hover:border-zinc-700 hover:bg-zinc-950">
                            <Settings2 className="mt-0.5 h-4 w-4 text-violet-400" />
                            <div>
                                <div className="font-medium text-white">Client config sync</div>
                                <div className="mt-1 text-xs text-zinc-400">Preview and write TormentNexus-managed MCP configs for Claude Desktop, Cursor, and VS Code.</div>
                            </div>
                        </Link>

                        <Link href="/dashboard/mcp/ai-tools" className="flex items-start gap-3 rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 hover:border-zinc-700 hover:bg-zinc-950">
                            <Sparkles className="mt-0.5 h-4 w-4 text-amber-400" />
                            <div>
                                <div className="font-medium text-white">CLI harness directory</div>
                                <div className="mt-1 text-xs text-zinc-400">See which local harnesses are installed, how many sessions are running, and which providers are connected.</div>
                            </div>
                        </Link>
                    </CardContent>
                </Card>
            </div>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-white">Connected bridge clients</CardTitle>
                </CardHeader>
                <CardContent>
                    {connectedBridgeClients.length === 0 ? (
                        <div className="rounded-lg border border-dashed border-zinc-800 bg-zinc-950/50 p-4 text-sm text-zinc-400">
                            {getBridgeClientEmptyStateMessage(overview)}
                        </div>
                    ) : (
                        <div className="space-y-3">
                            {connectedBridgeClients.map((client) => (
                                <div key={client.clientId} className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 space-y-3">
                                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                                        <div>
                                            <div className="flex items-center gap-2">
                                                <span className="text-sm font-semibold text-white">{client.clientName}</span>
                                                <span className="rounded-full border border-cyan-500/30 bg-cyan-500/10 px-2 py-0.5 text-[10px] uppercase tracking-wide text-cyan-300">
                                                    {client.clientType}
                                                </span>
                                            </div>
                                            <div className="mt-2 text-xs text-zinc-400">
                                                {client.platform ?? 'Unknown platform'}{client.version ? ` · v${client.version}` : ''}
                                            </div>
                                        </div>
                                        <div className="text-xs text-zinc-500">Last seen {client.lastSeenLabel}</div>
                                    </div>

                                    <div className="grid gap-3 md:grid-cols-2">
                                        <div>
                                            <div className="text-[10px] uppercase tracking-wide text-zinc-500">Non-MCP capabilities</div>
                                            <div className="mt-2 flex flex-wrap gap-2">
                                                {client.capabilities.map((capability) => (
                                                    <span key={capability} className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300">
                                                        {capability}
                                                    </span>
                                                ))}
                                            </div>
                                        </div>
                                        <div>
                                            <div className="text-[10px] uppercase tracking-wide text-zinc-500">Advertised hook phases</div>
                                            <div className="mt-2 flex flex-wrap gap-2">
                                                {client.hookPhases.map((phase) => (
                                                    <span key={phase} className="rounded-full border border-violet-500/30 bg-violet-500/10 px-2 py-1 text-[11px] text-violet-200">
                                                        {phase}
                                                    </span>
                                                ))}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-white">Known MCP / extension client targets</CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="flex items-center justify-center gap-2 p-12 text-zinc-500">
                            <Loader2 className="h-5 w-5 animate-spin" />
                            Checking integration targets…
                        </div>
                    ) : (
                        <div className="space-y-3">
                            {clientRows.map((row) => (
                                <div key={row.id} className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4">
                                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                                        <div className="min-w-0">
                                            <div className="flex items-center gap-2">
                                                <span className="text-sm font-semibold text-white">{row.label}</span>
                                                <span className={`rounded-full border px-2 py-0.5 text-[10px] uppercase tracking-wide ${getStatusBadgeClasses(row.statusTone)}`}>
                                                    {row.statusLabel}
                                                </span>
                                                {row.autoSyncSupported ? (
                                                    <span className="rounded-full border border-blue-500/30 bg-blue-500/10 px-2 py-0.5 text-[10px] uppercase tracking-wide text-blue-300">
                                                        auto-sync supported
                                                    </span>
                                                ) : null}
                                            </div>
                                            <div className="mt-2 text-xs text-zinc-500">Windows config path</div>
                                            <div className="mt-1 break-all font-mono text-xs text-zinc-300">{row.resolvedPath}</div>
                                            <div className="mt-2 text-xs text-zinc-400">{row.notes}</div>
                                        </div>
                                        <div className="flex items-center gap-2 text-xs text-zinc-400">
                                            <FolderCode className="h-4 w-4" />
                                            {row.detected ? 'Detected on this machine' : 'Not detected from TormentNexus yet'}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-white">What this page covers now</CardTitle>
                </CardHeader>
                <CardContent className="grid gap-3 md:grid-cols-3 text-sm text-zinc-300">
                    <div className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4">
                        <div className="flex items-center gap-2 font-medium text-white"><Puzzle className="h-4 w-4 text-cyan-400" /> Browser & editor surfaces</div>
                        <div className="mt-2 text-xs text-zinc-400">Install roots, packaging hints, and management links for the browser bridge and VS Code extension.</div>
                    </div>
                    <div className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4">
                        <div className="flex items-center gap-2 font-medium text-white"><Cable className="h-4 w-4 text-violet-400" /> Live readiness</div>
                        <div className="mt-2 text-xs text-zinc-400">Extension bridge clients, browser runtime readiness, MCP target detection, and local harness install state.</div>
                    </div>
                    <div className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4">
                        <div className="flex items-center gap-2 font-medium text-white"><Settings2 className="h-4 w-4 text-amber-400" /> Next connection steps</div>
                        <div className="mt-2 text-xs text-zinc-400">Direct routes into browser runtime controls, MCP config sync, and the AI tools/provider directory.</div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}