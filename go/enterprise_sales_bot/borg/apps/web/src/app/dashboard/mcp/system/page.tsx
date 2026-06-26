"use client";

import { Card, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Activity, Server, Cpu, HardDrive, Network, Globe, Radio, Puzzle } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';
import type { ComponentType } from 'react';
import type { DashboardStartupStatus } from '../../dashboard-home-view';
import { buildSystemComponentHealthRows, buildSystemEnvironmentRows, buildSystemStartupChecks, buildSystemStartupNotice, buildSystemStatusCards } from './system-status-helpers';

function getStatusCardColor(status: string): string {
    if (status === 'Healthy' || status === 'Ready' || status === 'Listening') {
        return 'text-green-500';
    }

    if (status === 'Connecting') {
        return 'text-cyan-400';
    }

    return 'text-yellow-500';
}

export default function SystemStatusDashboard() {
    const { data: status, refetch } = trpc.mcp.getStatus.useQuery();
    const toolsClient = trpc.tools as any;
    const { data: startupStatus, refetch: refetchStartup } = trpc.startupStatus.useQuery(undefined, { refetchInterval: 5000 });
    const { data: browserStatus, refetch: refetchBrowser } = trpc.browser.status.useQuery(undefined, { refetchInterval: 5000 });
    const installArtifactsQuery = toolsClient?.detectInstallSurfaces?.useQuery
        ? toolsClient.detectInstallSurfaces.useQuery(undefined, { refetchInterval: 10000 })
        : ({ data: null, refetch: async () => undefined } as { data: null; refetch: () => Promise<unknown> });

    const closeAllPages = trpc.browser.closeAll.useMutation({
        onSuccess: () => {
            toast.success('Closed all browser pages.');
            void refetchBrowser();
        },
        onError: (err) => toast.error(`Failed to close pages: ${err.message}`),
    });

    const handleRefresh = () => {
        void refetch();
        void refetchStartup();
        void refetchBrowser();
        void installArtifactsQuery.refetch();
    };

    const startupSnapshot = startupStatus as DashboardStartupStatus | undefined;
    const startupChecks = startupSnapshot ? buildSystemStartupChecks(startupSnapshot, installArtifactsQuery.data) : [];
    const componentHealthRows = buildSystemComponentHealthRows(startupSnapshot, browserStatus ?? undefined, installArtifactsQuery.data);
    const environmentRows = buildSystemEnvironmentRows(startupSnapshot);
    const startupNotice = buildSystemStartupNotice(startupSnapshot);
    const statusCards = buildSystemStatusCards(startupSnapshot, Boolean(status?.initialized), installArtifactsQuery.data);

    return (
        <div className="p-8 space-y-8 h-full overflow-y-auto">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">System Status</h1>
                    <p className="text-zinc-500">
                        Infrastructure health and resource usage
                    </p>
                </div>
                <Button onClick={handleRefresh} variant="outline" className="border-zinc-700 hover:bg-zinc-800">
                    <Activity className="mr-2 h-4 w-4" /> Refresh Status
                </Button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatusCard
                    title="MCP Server"
                    status={statusCards.mcpServer.status}
                    icon={Server}
                    color={getStatusCardColor(statusCards.mcpServer.status)}
                    detail={statusCards.mcpServer.detail}
                />
                <StatusCard
                    title="Cached inventory"
                    status={statusCards.cachedInventory.status}
                    icon={HardDrive}
                    color={getStatusCardColor(statusCards.cachedInventory.status)}
                    detail={statusCards.cachedInventory.detail}
                />
                <StatusCard
                    title="Extension bridge"
                    status={statusCards.extensionBridge.status}
                    icon={Cpu}
                    color={getStatusCardColor(statusCards.extensionBridge.status)}
                    detail={statusCards.extensionBridge.detail}
                />
                <StatusCard
                    title="Extension artifacts"
                    status={statusCards.extensionArtifacts.status}
                    icon={Puzzle}
                    color={getStatusCardColor(statusCards.extensionArtifacts.status)}
                    detail={statusCards.extensionArtifacts.detail}
                />
                <StatusCard
                    title="Network"
                    status="Active"
                    icon={Network}
                    color="text-blue-500"
                    detail="Port 3000 / 3001"
                />
                <StatusCard
                    title="Startup Readiness"
                    status={statusCards.startupReadiness.status}
                    icon={Radio}
                    color={getStatusCardColor(statusCards.startupReadiness.status)}
                    detail={statusCards.startupReadiness.detail}
                />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {startupNotice ? (
                    <Card className={`md:col-span-2 border ${startupNotice.tone === 'warning' ? 'border-amber-900/30 bg-amber-950/10' : 'border-cyan-900/30 bg-cyan-950/10'}`}>
                        <CardContent className="p-6">
                            <div className={`text-sm font-semibold ${startupNotice.tone === 'warning' ? 'text-amber-300' : 'text-cyan-300'}`}>{startupNotice.title}</div>
                            <p className="mt-2 text-sm text-zinc-300">{startupNotice.detail}</p>
                        </CardContent>
                    </Card>
                ) : null}

                <Card className="bg-zinc-900 border-zinc-800">
                    <CardContent className="p-6 space-y-4">
                        <div className="flex items-center justify-between">
                            <div>
                                <h3 className="text-lg font-medium text-white">Browser Runtime</h3>
                                <p className="text-xs text-zinc-500 mt-1">Live status from `browserRouter`</p>
                            </div>
                            <Globe className="h-5 w-5 text-cyan-400" />
                        </div>

                        <div className="grid grid-cols-3 gap-2 text-sm">
                            <div className="bg-zinc-950 border border-zinc-800 rounded p-3">
                                <div className="text-zinc-500 text-xs">Available</div>
                                <div className="text-white font-semibold">{browserStatus?.available ? 'Yes' : 'No'}</div>
                            </div>
                            <div className="bg-zinc-950 border border-zinc-800 rounded p-3">
                                <div className="text-zinc-500 text-xs">Active</div>
                                <div className="text-white font-semibold">{browserStatus?.active ? 'Yes' : 'No'}</div>
                            </div>
                            <div className="bg-zinc-950 border border-zinc-800 rounded p-3">
                                <div className="text-zinc-500 text-xs">Pages</div>
                                <div className="text-white font-semibold">{browserStatus?.pageCount ?? 0}</div>
                            </div>
                        </div>

                        <div className="flex justify-end">
                            <Button
                                onClick={() => closeAllPages.mutate()}
                                disabled={!browserStatus?.available || closeAllPages.isPending}
                                variant="outline"
                                className="border-zinc-700 hover:bg-zinc-800"
                            >
                                Close All Pages
                            </Button>
                        </div>
                    </CardContent>
                </Card>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <Card className="bg-zinc-900 border-zinc-800">
                    <CardContent className="p-6">
                        <h3 className="text-lg font-medium text-white mb-4">Component Health</h3>
                        <div className="space-y-4">
                            {componentHealthRows.map((row) => (
                                <HealthRow
                                    key={row.name}
                                    name={row.name}
                                    status={row.status}
                                    latency={row.latency}
                                    detail={row.detail}
                                />
                            ))}
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-zinc-900 border-zinc-800">
                    <CardContent className="p-6">
                        <h3 className="text-lg font-medium text-white mb-4">Startup Checks</h3>
                        <div className="space-y-4">
                            {startupChecks.length === 0 ? (
                                <div className="rounded border border-zinc-800 bg-zinc-950 p-4 text-sm text-zinc-500">
                                    Connecting to live startup telemetry from TormentNexus Core…
                                </div>
                            ) : startupChecks.map((check) => (
                                <HealthRow key={check.name} name={check.name} status={check.status} latency={check.latency} detail={check.detail} />
                            ))}
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-zinc-900 border-zinc-800">
                    <CardContent className="p-6">
                        <h3 className="text-lg font-medium text-white mb-4">Environment</h3>
                        <div className="space-y-2 font-mono text-sm text-zinc-400">
                            {environmentRows.map((row, index) => (
                                <div
                                    key={row.label}
                                    className={`flex justify-between ${index < environmentRows.length - 1 ? 'border-b border-zinc-800 pb-2' : 'pt-2'} ${index > 0 && index < environmentRows.length - 1 ? 'pt-2' : ''}`}
                                >
                                    <span>{row.label}</span>
                                    <span className={row.accent ? 'text-blue-400' : 'text-white'}>{row.value}</span>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}

function StatusCard({ title, status, icon: Icon, color, detail }: { title: string; status: string; icon: ComponentType<{ className?: string }>; color: string; detail?: string }) {
    return (
        <Card className="bg-zinc-900 border-zinc-800">
            <CardContent className="p-6">
                <div className="flex items-center justify-between mb-2">
                    <span className="text-zinc-400 font-medium">{title}</span>
                    <Icon className={`h-5 w-5 ${color}`} />
                </div>
                <div className="text-2xl font-bold text-white mb-1">{status}</div>
                {detail && <div className="text-xs text-zinc-500">{detail}</div>}
            </CardContent>
        </Card>
    );
}

function HealthRow({ name, status, latency, detail }: { name: string; status: string; latency: string; detail?: string }) {
    const isHealthy = status === 'Operational' || status === 'Healthy' || status === 'Active';
    return (
        <div className="flex items-center justify-between p-3 bg-zinc-950 rounded border border-zinc-800">
            <div className="flex items-center gap-3">
                <div className={`h-2 w-2 rounded-full ${isHealthy ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]' : 'bg-yellow-500 shadow-[0_0_8px_rgba(234,179,8,0.5)]'}`} />
                <div>
                    <span className="text-zinc-200 font-medium">{name}</span>
                    {detail ? <div className="text-xs text-zinc-500 mt-1">{detail}</div> : null}
                </div>
            </div>
            <div className="flex items-center gap-4 text-sm">
                <span className="text-zinc-500">{latency}</span>
                <span className={isHealthy ? 'text-green-400' : 'text-yellow-400'}>{status}</span>
            </div>
        </div>
    );
}
