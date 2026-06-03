"use client";

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@tormentnexus/ui";
import { Badge } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Activity, Server, AlertTriangle, RefreshCcw, HardDrive, Cpu, Network, Radio } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';
import { ComponentType, useState } from 'react';
import type { DashboardStartupStatus } from '../dashboard-home-view';
import { buildSystemEnvironmentRows, buildSystemStartupNotice } from '../mcp/system/system-status-helpers';
import { getEventBusMetric, getMcpRouterMetric } from './health-metrics';
import { getConnectedServerKeys, normalizeHealthServers } from './health-server-list';
import { buildHealthStartupViewModel } from './health-startup-view-model';

export default function HealthDashboard() {
    const [isRefreshing, setIsRefreshing] = useState(false);
    const utils = trpc.useUtils();
    const toolsClient = trpc.tools as any;

    const { data: mcpStatus, refetch: refetchMcpStatus } = trpc.mcp.getStatus.useQuery();
    const { data: startupStatus, refetch: refetchStartup } = trpc.startupStatus.useQuery(undefined, { refetchInterval: 5000 });
    const { data: servers, refetch: refetchServers } = trpc.mcpServers.list.useQuery();
    const installArtifactsQuery = toolsClient?.detectInstallSurfaces?.useQuery
        ? toolsClient.detectInstallSurfaces.useQuery(undefined, { refetchInterval: 10000 })
        : ({ data: null, refetch: async () => undefined } as { data: null; refetch: () => Promise<unknown> });
    
    // We will query health for each server via a separate component or handle it manually if we need bulk
    // For simplicity, we just leverage TRPC queries directly where we render individual servers

    const handleRefresh = async () => {
        setIsRefreshing(true);
        try {
            await Promise.all([
                refetchMcpStatus(),
                refetchStartup(),
                refetchServers(),
                installArtifactsQuery.refetch(),
                utils.serverHealth.check.invalidate(),
            ]);
            toast.success("Health data refreshed");
        } finally {
            setIsRefreshing(false);
        }
    };

    const startupSnapshot = startupStatus as DashboardStartupStatus | undefined;
    const startupViewModel = buildHealthStartupViewModel(
        startupSnapshot,
        Boolean(mcpStatus?.initialized),
        installArtifactsQuery.data,
    );
    const startupChecks = startupViewModel.startupChecks;
    const environmentRows = buildSystemEnvironmentRows(startupSnapshot);
    const startupNotice = buildSystemStartupNotice(startupSnapshot);
    const statusCards = startupViewModel.statusCards;
    const eventBusMetric = getEventBusMetric(startupSnapshot);
    const mcpRouterMetric = getMcpRouterMetric(startupSnapshot, Boolean(mcpStatus?.initialized));
    const connectedServers = getConnectedServerKeys(mcpStatus);
    const normalizedServers = normalizeHealthServers(servers);

    return (
        <div className="p-8 space-y-8 h-full overflow-y-auto">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <Activity className="h-8 w-8 text-green-500" />
                        System Health
                    </h1>
                    <p className="text-zinc-500 mt-1">
                        Monitor TormentNexus infrastructure status, component uptime, and server crash rates.
                    </p>
                </div>
                <Button 
                    onClick={handleRefresh} 
                    disabled={isRefreshing}
                    variant="outline" 
                    className="border-zinc-700 hover:bg-zinc-800"
                >
                    <RefreshCcw className={`mr-2 h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} /> 
                    Refresh Health
                </Button>
            </div>

            {/* Top Level System Metrics (Imported logic from System Status) */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <MetricCard
                    title="MCP Router"
                    status={mcpRouterMetric.status}
                    icon={Server}
                    color={mcpRouterMetric.color}
                    detail={mcpRouterMetric.detail}
                />
                <MetricCard
                    title="Database"
                    status="Connected"
                    icon={HardDrive}
                    color="text-green-500"
                    detail="SQLite (local)"
                />
                <MetricCard
                    title="Event Bus"
                    status={eventBusMetric.status}
                    icon={Cpu}
                    color={eventBusMetric.color}
                    detail={eventBusMetric.detail}
                />
                <MetricCard
                    title="Startup Readiness"
                    status={statusCards.startupReadiness.status}
                    icon={Radio}
                    color={statusCards.startupReadiness.status === 'Ready' ? 'text-green-500' : statusCards.startupReadiness.status === 'Degraded' ? 'text-amber-500' : 'text-yellow-500'}
                    detail={statusCards.startupReadiness.detail}
                />
            </div>

            {startupNotice ? (
                <Card className={`border ${startupNotice.tone === 'warning' ? 'bg-amber-950/10 border-amber-900/30' : 'bg-cyan-950/10 border-cyan-900/30'}`}>
                    <CardHeader className="pb-3">
                        <CardTitle className={`text-base font-medium ${startupNotice.tone === 'warning' ? 'text-amber-300' : 'text-cyan-300'}`}>{startupNotice.title}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm text-zinc-300">{startupNotice.detail}</p>
                    </CardContent>
                </Card>
            ) : null}

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 text-sm">
                
                {/* Left Column: Server Health Grid */}
                <div className="lg:col-span-2 space-y-6">
                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="pb-4">
                            <CardTitle className="text-lg font-medium text-white flex items-center gap-2">
                                <Server className="h-5 w-5 text-zinc-400" />
                                MCP Server Health
                            </CardTitle>
                            <CardDescription>
                                Individual status for each configured server. Servers that crash repeatedly are isolated.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-3">
                                {normalizedServers.length === 0 ? (
                                    <div className="text-zinc-500 text-center py-8 bg-zinc-950/50 rounded border border-zinc-800/50 border-dashed">
                                        No MCP servers configured or detected.
                                    </div>
                                ) : (
                                    normalizedServers.map((server) => (
                                        <ServerHealthRow 
                                            key={server.uuid} 
                                            server={server} 
                                            isConnected={connectedServers.includes(server.configKey)} 
                                        />
                                    ))
                                )}
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Right Column: Key Details */}
                <div className="space-y-6">
                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="pb-4">
                            <CardTitle className="text-lg font-medium text-white">Environment</CardTitle>
                        </CardHeader>
                        <CardContent>
                             <div className="space-y-2 font-mono text-xs text-zinc-400">
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

                    <Card className="bg-zinc-900 border-zinc-800 bg-amber-950/10 border-amber-900/20">
                        <CardHeader className="pb-4">
                            <CardTitle className="text-lg font-medium text-amber-500 flex items-center gap-2">
                                <AlertTriangle className="h-5 w-5" />
                                Incident Reporting
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="text-xs text-zinc-400 space-y-4">
                            <p>
                                If an MCP server exceeds its maximum crash attempts, it will enter an <strong className="text-red-400">ERROR</strong> state and be removed from the active routing pool.
                            </p>
                            <p>
                                You can manually clear error states using the <strong>Reset Health</strong> action. The supervisor will attempt to reconnect to the server during the next polling cycle.
                            </p>
                        </CardContent>
                    </Card>
                </div>

            </div>
        </div>
    );
}

function MetricCard({ title, status, icon: Icon, color, detail }: { title: string; status: string; icon: ComponentType<{ className?: string }>; color: string; detail?: string }) {
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

function ServerHealthRow({ server, isConnected }: { server: any, isConnected: boolean }) {
    const { data: health, refetch } = trpc.serverHealth.check.useQuery({ serverUuid: server.uuid }, { refetchInterval: 5000 });
    const resetHealth = trpc.serverHealth.reset.useMutation({
        onSuccess: () => {
            toast.success(`Reset health state for ${server.configKey}`);
            void refetch();
        },
        onError: (err) => {
            toast.error(`Failed to reset health: ${err.message}`);
        }
    });

    const isError = health?.status === 'ERROR';
    const isHealthy = health?.status === 'HEALTHY' && isConnected;
    const isOffline = health?.status === 'HEALTHY' && !isConnected;

    return (
        <div className={`flex items-center justify-between p-4 rounded border ${isError ? 'bg-red-950/10 border-red-900/30' : 'bg-zinc-950 border-zinc-800'}`}>
            <div className="flex flex-col gap-1">
                <div className="flex items-center gap-2">
                    <span className="text-zinc-200 font-medium">{server.name || server.configKey}</span>
                    {isError && <Badge variant="destructive" className="h-5 text-[10px]">ERROR</Badge>}
                    {isHealthy && <Badge variant="outline" className="h-5 text-[10px] border-green-500/30 text-green-400 bg-green-500/10">CONNECTED</Badge>}
                    {isOffline && <Badge variant="outline" className="h-5 text-[10px] border-zinc-700 text-zinc-400">OFFLINE</Badge>}
                </div>
                <div className="text-xs font-mono text-zinc-500 truncate max-w-[300px]">
                    {server.configKey} ({server.transportType})
                </div>
            </div>

            <div className="flex items-center gap-6">
                <div className="flex flex-col items-end gap-1">
                    <span className="text-xs text-zinc-500">Crashes</span>
                    <span className={`text-sm font-mono ${health?.crashCount && health.crashCount > 0 ? 'text-yellow-500' : 'text-zinc-400'}`}>
                        {health?.crashCount ?? 0} / {health?.maxAttempts ?? 3}
                    </span>
                </div>
                
                <Button 
                    size="sm" 
                    variant="outline" 
                    className={`h-8 text-xs ${isError ? 'border-red-500/50 text-red-400 hover:bg-red-950 hover:text-red-300' : 'border-zinc-700'}`}
                    disabled={!isError || resetHealth.isPending}
                    onClick={() => resetHealth.mutate({ serverUuid: server.uuid })}
                >
                    Reset Health
                </Button>
            </div>
        </div>
    );
}

