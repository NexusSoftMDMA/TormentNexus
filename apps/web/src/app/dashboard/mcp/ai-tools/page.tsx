"use client";

import { useMemo, useState } from 'react';
import Link from 'next/link';
import { Button, Card, CardContent, CardHeader, CardTitle } from '@tormentnexus/ui';
import { Bot, CheckCircle2, Database, ExternalLink, KeyRound, Loader2, RefreshCw, Search, Server, TerminalSquare, Wrench, XCircle } from 'lucide-react';
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

import { getCliHarnessCards, getProviderDirectoryCards, getStatusBadgeClasses } from './ai-tool-directory';
import { getPortalBadgeClasses } from '../../billing/billing-portal-data';

export default function AIToolsDashboard() {
    const [query, setQuery] = useState('');
    const [healthServerUuid, setHealthServerUuid] = useState('');
    const mcpServersClient = trpc.mcpServers as any;
    const toolsClient = trpc.tools as any;
    const hasCliDetectionQuery = typeof toolsClient?.detectCliHarnesses?.useQuery === 'function';
    const hasExecutionEnvironmentQuery = typeof toolsClient?.detectExecutionEnvironment?.useQuery === 'function';

    const toolsQuery = trpc.tools.list.useQuery();
    const serversQuery = trpc.mcpServers.list.useQuery();
    const apiKeysQuery = trpc.apiKeys.list.useQuery();
    const cliDetectionsQuery = hasCliDetectionQuery
        ? toolsClient.detectCliHarnesses.useQuery()
        : {
            data: null,
            isLoading: false,
            refetch: async () => undefined,
        };
    const executionEnvironmentQuery = hasExecutionEnvironmentQuery
        ? toolsClient.detectExecutionEnvironment.useQuery()
        : {
            data: null,
            isLoading: false,
            refetch: async () => undefined,
        };
    const providerQuotasQuery = trpc.billing.getProviderQuotas.useQuery();
    const sessionsQuery = trpc.session.list.useQuery();
    const { data: tools, isLoading: loadingTools } = toolsQuery;
    const { data: servers, isLoading: loadingServers } = serversQuery;
    const { data: apiKeys, isLoading: loadingKeys } = apiKeysQuery;
    const { data: cliDetections, isLoading: loadingCliDetections } = cliDetectionsQuery;
    const { data: providerQuotas } = providerQuotasQuery;
    const { data: sessions } = sessionsQuery;
    const { data: expertStatus } = trpc.expert.getStatus.useQuery();
    const { data: sessionState } = trpc.session.getState.useQuery();
    const { data: memoryStats } = trpc.agentMemory.stats.useQuery();
    const { data: shellHistory } = trpc.shell.getSystemHistory.useQuery({ limit: 8 });
    const { data: serverHealth } = trpc.serverHealth.check.useQuery(
        { serverUuid: healthServerUuid },
        { enabled: healthServerUuid.trim().length > 0 }
    );
    const reloadMetadataMutation = mcpServersClient.reloadMetadata.useMutation({
        onSuccess: async (result: any) => {
            toast.success(`Reloaded metadata for ${result.server.name} from ${result.metadata.metadataSource ?? 'metadata cache'}.`);
            await Promise.all([
                toolsQuery.refetch(),
                serversQuery.refetch(),
            ]);
        },
        onError: (error: any) => {
            toast.error(error.message);
        },
    });
    const clearMetadataCacheMutation = mcpServersClient.clearMetadataCache.useMutation({
        onSuccess: async (result: any) => {
            toast.success(`Cleared cached metadata for ${result.server.name}. Auto mode will rediscover from the binary next time.`);
            await Promise.all([
                toolsQuery.refetch(),
                serversQuery.refetch(),
            ]);
        },
        onError: (error: any) => {
            toast.error(error.message);
        },
    });

    const normalized = query.trim().toLowerCase();

    const filteredTools = useMemo(() => {
        const source = tools ?? [];
        if (!normalized) {
            return source;
        }

        return source.filter((tool: any) => {
            const name = String(tool?.name ?? '').toLowerCase();
            const description = String(tool?.description ?? '').toLowerCase();
            const server = String(tool?.server ?? '').toLowerCase();
            return name.includes(normalized) || description.includes(normalized) || server.includes(normalized);
        });
    }, [normalized, tools]);

    const activeServers = useMemo(() => {
        return (servers ?? []).filter((server: any) => server?.error_status === 'NONE');
    }, [servers]);

    const firstServerUuid = useMemo(() => {
        const first = (servers ?? [])[0] as any;
        return typeof first?.uuid === 'string' ? first.uuid : '';
    }, [servers]);

    const effectiveHealthServerUuid = healthServerUuid || firstServerUuid;

    const activeKeys = useMemo(() => {
        return (apiKeys ?? []).filter((key: any) => Boolean(key?.is_active));
    }, [apiKeys]);

    const normalizedSessions = useMemo(() => {
        return (Array.isArray(sessions) ? sessions : [])
            .filter((session: any) => session && typeof session === 'object' && typeof session.cliType === 'string')
            .map((session: any) => ({
                cliType: String(session.cliType),
                status: String(session.status ?? 'unknown'),
            }));
    }, [sessions]);

    const normalizedCliDetections = useMemo(() => {
        return (Array.isArray(cliDetections) ? cliDetections : [])
            .filter((detection: any) => detection && typeof detection === 'object')
            .map((detection: any, index: number) => ({
                id: typeof detection.id === 'string' && detection.id.trim().length > 0 ? detection.id : `cli-${index}`,
                name: typeof detection.name === 'string' && detection.name.trim().length > 0 ? detection.name : 'Unknown harness',
                command: typeof detection.command === 'string' ? detection.command : '',
                homepage: typeof detection.homepage === 'string' && detection.homepage.trim().length > 0 ? detection.homepage : '#',
                docsUrl: typeof detection.docsUrl === 'string' && detection.docsUrl.trim().length > 0 ? detection.docsUrl : '#',
                installHint: typeof detection.installHint === 'string' ? detection.installHint : 'Installation instructions unavailable',
                sessionCapable: Boolean(detection.sessionCapable),
                installed: Boolean(detection.installed),
                resolvedPath: typeof detection.resolvedPath === 'string' && detection.resolvedPath.trim().length > 0 ? detection.resolvedPath : null,
                version: typeof detection.version === 'string' && detection.version.trim().length > 0 ? detection.version : null,
                detectionError: typeof detection.detectionError === 'string' && detection.detectionError.trim().length > 0 ? detection.detectionError : null,
            }));
    }, [cliDetections]);

    const normalizedProviderQuotas = useMemo(() => {
        return (Array.isArray(providerQuotas) ? providerQuotas : []).filter((quota: any) => quota && typeof quota === 'object');
    }, [providerQuotas]);

    const cliHarnessCards = useMemo(() => getCliHarnessCards(normalizedCliDetections, normalizedSessions), [normalizedCliDetections, normalizedSessions]);
    const providerDirectoryCards = useMemo(() => getProviderDirectoryCards(normalizedProviderQuotas), [normalizedProviderQuotas]);
    const connectedProviders = useMemo(() => providerDirectoryCards.filter((card) => card.statusTone === 'success'), [providerDirectoryCards]);
    const detectedHarnesses = useMemo(() => cliHarnessCards.filter((card) => card.installed), [cliHarnessCards]);
    const executionEnvironmentData = useMemo(() => {
        const raw = executionEnvironmentQuery.data as any;
        if (!raw || typeof raw !== 'object') {
            return null;
        }

        const summary = raw.summary && typeof raw.summary === 'object' ? raw.summary : {};
        const shells = (Array.isArray(raw.shells) ? raw.shells : []).map((shell: any, index: number) => {
            const candidate = shell && typeof shell === 'object' ? shell : {};
            return {
                id: typeof candidate.id === 'string' && candidate.id.trim().length > 0 ? candidate.id : `shell-${index}`,
                name: typeof candidate.name === 'string' && candidate.name.trim().length > 0 ? candidate.name : 'Unknown shell',
                verified: Boolean(candidate.verified),
                installed: Boolean(candidate.installed),
                preferred: Boolean(candidate.preferred),
                resolvedPath: typeof candidate.resolvedPath === 'string' && candidate.resolvedPath.trim().length > 0 ? candidate.resolvedPath : null,
                family: typeof candidate.family === 'string' && candidate.family.trim().length > 0 ? candidate.family : 'unknown',
                version: typeof candidate.version === 'string' && candidate.version.trim().length > 0 ? candidate.version : null,
            };
        });
        const tools = (Array.isArray(raw.tools) ? raw.tools : []).map((tool: any, index: number) => {
            const candidate = tool && typeof tool === 'object' ? tool : {};
            return {
                id: typeof candidate.id === 'string' && candidate.id.trim().length > 0 ? candidate.id : `tool-${index}`,
                name: typeof candidate.name === 'string' && candidate.name.trim().length > 0 ? candidate.name : 'Unknown tool',
                verified: Boolean(candidate.verified),
                installed: Boolean(candidate.installed),
                resolvedPath: typeof candidate.resolvedPath === 'string' && candidate.resolvedPath.trim().length > 0 ? candidate.resolvedPath : null,
                version: typeof candidate.version === 'string' && candidate.version.trim().length > 0 ? candidate.version : null,
                capabilities: Array.isArray(candidate.capabilities)
                    ? candidate.capabilities.filter((capability: unknown): capability is string => typeof capability === 'string' && capability.trim().length > 0)
                    : [],
            };
        });

        return {
            summary: {
                preferredShellLabel: typeof summary.preferredShellLabel === 'string' ? summary.preferredShellLabel : null,
                verifiedShellCount: typeof summary.verifiedShellCount === 'number' ? summary.verifiedShellCount : 0,
                shellCount: typeof summary.shellCount === 'number' ? summary.shellCount : shells.length,
                verifiedToolCount: typeof summary.verifiedToolCount === 'number' ? summary.verifiedToolCount : 0,
                toolCount: typeof summary.toolCount === 'number' ? summary.toolCount : tools.length,
                ready: Boolean(summary.ready),
            },
            shells,
            tools,
        };
    }, [executionEnvironmentQuery.data]);

    const loading = loadingTools || loadingServers || loadingKeys || loadingCliDetections || executionEnvironmentQuery.isLoading;

    return (
        <div className="p-8 space-y-8 h-full overflow-y-auto">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">AI Tools</h1>
                    <p className="text-zinc-500">Unified operational view of CLI harness detection, provider connectivity, tool inventory, and MCP readiness.</p>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                    <Link href="/dashboard/billing" className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:border-zinc-600 hover:bg-zinc-800">
                        Provider billing
                        <ExternalLink className="h-4 w-4" />
                    </Link>
                    <Link href="/dashboard/session" className="inline-flex items-center gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 hover:border-zinc-600 hover:bg-zinc-800">
                        Session control
                        <ExternalLink className="h-4 w-4" />
                    </Link>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-4">
                <StatCard title="Tools Indexed" value={String(tools?.length ?? 0)} icon={Wrench} tone="text-blue-400" />
                <StatCard title="Active Servers" value={`${activeServers.length}/${servers?.length ?? 0}`} icon={Server} tone="text-emerald-400" />
                <StatCard title="Active API Keys" value={`${activeKeys.length}/${apiKeys?.length ?? 0}`} icon={KeyRound} tone="text-yellow-400" />
                <StatCard title="Detected Harnesses" value={`${detectedHarnesses.length}/${cliHarnessCards.length}`} icon={TerminalSquare} tone="text-violet-400" />
                <StatCard title="Connected Providers" value={`${connectedProviders.length}/${providerDirectoryCards.length}`} icon={Bot} tone="text-cyan-400" />
            </div>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-white">Execution Environment</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    {!executionEnvironmentData ? (
                        <div className="rounded border border-dashed border-zinc-800 bg-zinc-950/40 p-6 text-sm text-zinc-500">
                            Execution environment details are still loading.
                        </div>
                    ) : (
                        <>
                            <div className="grid grid-cols-1 md:grid-cols-4 gap-3 text-xs text-zinc-300">
                                <div className="rounded border border-zinc-800 bg-zinc-950/50 p-3">
                                    <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Preferred shell</div>
                                    <div className="mt-2 text-sm text-white">{executionEnvironmentData.summary.preferredShellLabel ?? 'None verified'}</div>
                                </div>
                                <div className="rounded border border-zinc-800 bg-zinc-950/50 p-3">
                                    <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Verified shells</div>
                                    <div className="mt-2 text-sm text-white">{executionEnvironmentData.summary.verifiedShellCount}/{executionEnvironmentData.summary.shellCount}</div>
                                </div>
                                <div className="rounded border border-zinc-800 bg-zinc-950/50 p-3">
                                    <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Verified tools</div>
                                    <div className="mt-2 text-sm text-white">{executionEnvironmentData.summary.verifiedToolCount}/{executionEnvironmentData.summary.toolCount}</div>
                                </div>
                                <div className="rounded border border-zinc-800 bg-zinc-950/50 p-3">
                                    <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Execution posture</div>
                                    <div className="mt-2 text-sm text-white">{executionEnvironmentData.summary.ready ? 'Ready' : 'Partial'}</div>
                                </div>
                            </div>

                            <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
                                <div className="space-y-3">
                                    <div className="text-[10px] uppercase tracking-wide text-zinc-500">Detected shells</div>
                                    {executionEnvironmentData.shells.map((shell: any) => (
                                        <div key={shell.id} className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 space-y-2">
                                            <div className="flex items-center justify-between gap-3">
                                                <div>
                                                    <div className="flex items-center gap-2">
                                                        <span className="text-sm font-semibold text-white">{shell.name}</span>
                                                        <span className={`rounded-full border px-2 py-0.5 text-[10px] uppercase tracking-wide ${getStatusBadgeClasses(shell.verified ? 'success' : shell.installed ? 'warning' : 'muted')}`}>
                                                            {shell.verified ? 'verified' : shell.installed ? 'detected' : 'missing'}
                                                        </span>
                                                        {shell.preferred ? (
                                                            <span className="rounded-full border border-cyan-500/30 bg-cyan-500/10 px-2 py-0.5 text-[10px] uppercase tracking-wide text-cyan-300">
                                                                preferred
                                                            </span>
                                                        ) : null}
                                                    </div>
                                                    <div className="mt-1 text-xs text-zinc-400 break-all">{shell.resolvedPath ?? 'Not detected'}</div>
                                                </div>
                                                <div className="text-xs text-zinc-500 uppercase tracking-wide">{shell.family}</div>
                                            </div>
                                            <div className="text-xs text-zinc-300">{shell.version ?? 'Version unavailable'}</div>
                                        </div>
                                    ))}
                                </div>

                                <div className="space-y-3">
                                    <div className="text-[10px] uppercase tracking-wide text-zinc-500">Common local tools</div>
                                    {executionEnvironmentData.tools.map((tool: any) => (
                                        <div key={tool.id} className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 space-y-2">
                                            <div className="flex items-center justify-between gap-3">
                                                <div>
                                                    <div className="flex items-center gap-2">
                                                        <span className="text-sm font-semibold text-white">{tool.name}</span>
                                                        <span className={`rounded-full border px-2 py-0.5 text-[10px] uppercase tracking-wide ${getStatusBadgeClasses(tool.verified ? 'success' : tool.installed ? 'warning' : 'muted')}`}>
                                                            {tool.verified ? 'verified' : tool.installed ? 'detected' : 'missing'}
                                                        </span>
                                                    </div>
                                                    <div className="mt-1 text-xs text-zinc-400 break-all">{tool.resolvedPath ?? 'Not detected'}</div>
                                                </div>
                                                <div className="text-xs text-zinc-500">{tool.version ?? '—'}</div>
                                            </div>
                                            <div className="flex flex-wrap gap-2">
                                                {(tool.capabilities ?? []).map((capability: string) => (
                                                    <span key={capability} className="rounded-full border border-zinc-700 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300">
                                                        {capability}
                                                    </span>
                                                ))}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </>
                    )}
                </CardContent>
            </Card>

            <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
                <Card className="bg-zinc-900 border-zinc-800">
                    <CardHeader>
                        <CardTitle className="text-white">CLI Harness Directory</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3">
                        {cliHarnessCards.length === 0 ? (
                            <div className="rounded border border-dashed border-zinc-800 bg-zinc-950/40 p-6 text-sm text-zinc-500">
                                No CLI harness detections available yet.
                            </div>
                        ) : (
                            cliHarnessCards.map((card) => (
                                <div key={card.id} className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 space-y-3">
                                    <div className="flex items-start justify-between gap-3">
                                        <div>
                                            <div className="flex items-center gap-2">
                                                <span className="text-sm font-semibold text-white">{card.name}</span>
                                                <span className={`rounded-full border px-2 py-0.5 text-[10px] uppercase tracking-wide ${getStatusBadgeClasses(card.statusTone)}`}>
                                                    {card.statusLabel}
                                                </span>
                                            </div>
                                            <div className="mt-1 text-xs text-zinc-400 font-mono">{card.command}</div>
                                        </div>
                                        <div className="flex gap-2 text-xs">
                                            <a href={card.homepage} target="_blank" rel="noreferrer" className="text-blue-300 hover:text-blue-200">Homepage</a>
                                            <a href={card.docsUrl} target="_blank" rel="noreferrer" className="text-zinc-300 hover:text-white">Docs</a>
                                        </div>
                                    </div>
                                    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs text-zinc-300">
                                        <div>
                                            <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Version</div>
                                            <div className="mt-1 break-all">{card.version ?? 'Not detected'}</div>
                                        </div>
                                        <div>
                                            <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Sessions</div>
                                            <div className="mt-1">{card.runningSessions} running / {card.activeSessions} total</div>
                                        </div>
                                        <div>
                                            <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Binary</div>
                                            <div className="mt-1 break-all">{card.resolvedPath ?? card.installHint}</div>
                                        </div>
                                    </div>
                                </div>
                            ))
                        )}
                    </CardContent>
                </Card>

                <Card className="bg-zinc-900 border-zinc-800">
                    <CardHeader>
                        <CardTitle className="text-white">Provider Directory</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3">
                        {providerDirectoryCards.length === 0 ? (
                            <div className="rounded border border-dashed border-zinc-800 bg-zinc-950/40 p-6 text-sm text-zinc-500">
                                Provider quota data has not been discovered yet.
                            </div>
                        ) : (
                            providerDirectoryCards.map((card) => (
                                <div key={card.provider} className="rounded-lg border border-zinc-800 bg-zinc-950/50 p-4 space-y-3">
                                    <div className="flex items-start justify-between gap-3">
                                        <div>
                                            <div className="flex items-center gap-2">
                                                <span className="text-sm font-semibold text-white">{card.label}</span>
                                                <span className={`rounded-full border px-2 py-0.5 text-[10px] uppercase tracking-wide ${getPortalBadgeClasses(card.statusTone)}`}>
                                                    {card.statusLabel}
                                                </span>
                                            </div>
                                            <div className="mt-1 text-xs text-zinc-400">{card.authLabel} • {card.availabilityLabel}</div>
                                        </div>
                                        <a href={card.href} target="_blank" rel="noreferrer" className="inline-flex items-center gap-1 text-xs text-blue-300 hover:text-blue-200">
                                            Open portal
                                            <ExternalLink className="h-3.5 w-3.5" />
                                        </a>
                                    </div>
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs text-zinc-300">
                                        <div>
                                            <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Usage</div>
                                            <div className="mt-1">{card.usageLabel}</div>
                                        </div>
                                        <div>
                                            <div className="text-zinc-500 uppercase tracking-wide text-[10px]">Reset</div>
                                            <div className="mt-1 break-all">{card.resetLabel}</div>
                                        </div>
                                    </div>
                                </div>
                            ))
                        )}
                    </CardContent>
                </Card>
            </div>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-white">Search Tool Inventory</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
                        <input
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            placeholder="Filter by tool name, description, or server"
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2.5 pl-9 text-sm text-white focus:ring-1 focus:ring-blue-500 outline-none"
                        />
                    </div>
                </CardContent>
            </Card>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-white">Tool List ({filteredTools.length})</CardTitle>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="flex items-center justify-center p-12 text-zinc-500 gap-2">
                            <Loader2 className="h-5 w-5 animate-spin" />
                            Loading AI tools dashboard...
                        </div>
                    ) : filteredTools.length === 0 ? (
                        <div className="text-center p-10 text-zinc-500 border border-zinc-800 border-dashed rounded-lg bg-zinc-950/40">
                            <Bot className="h-10 w-10 mx-auto mb-3 opacity-40" />
                            <p className="text-sm">No tools match your filter.</p>
                        </div>
                    ) : (
                        <div className="space-y-2">
                            {filteredTools.map((tool: any) => (
                                <div
                                    key={tool.uuid ?? `${tool.name}-${tool.server}`}
                                    className="rounded-md border border-zinc-800 bg-zinc-950/60 p-3 flex items-start justify-between gap-3"
                                >
                                    <div className="min-w-0">
                                        <div className="text-sm font-mono text-blue-300 truncate">{tool.name}</div>
                                        <div className="text-xs text-zinc-400 mt-1 line-clamp-2">{tool.description || 'No description'}</div>
                                    </div>
                                    <div className="text-[10px] px-2 py-0.5 rounded border border-zinc-700 text-zinc-400 bg-zinc-900 whitespace-nowrap">
                                        {tool.server || 'unknown-server'}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <ReadinessCard
                    title="Server Readiness"
                    healthyCount={activeServers.length}
                    totalCount={servers?.length ?? 0}
                    healthyLabel="Connected"
                    unhealthyLabel="Issues"
                />
                <ReadinessCard
                    title="API Key Readiness"
                    healthyCount={activeKeys.length}
                    totalCount={apiKeys?.length ?? 0}
                    healthyLabel="Active"
                    unhealthyLabel="Inactive"
                />
            </div>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-white">MCP metadata cache</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                    {(servers ?? []).length === 0 ? (
                        <div className="rounded border border-dashed border-zinc-800 bg-zinc-950/40 p-6 text-sm text-zinc-500">
                            No MCP servers are available yet.
                        </div>
                    ) : (
                        (servers ?? []).map((server: any) => {
                            const metadata = server?._meta;
                            const pending = reloadMetadataMutation.isPending && reloadMetadataMutation.variables?.uuid === server.uuid;

                            return (
                                <div key={server.uuid} className="rounded border border-zinc-800 bg-zinc-950/50 p-3">
                                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                                        <div className="min-w-0">
                                            <div className="flex items-center gap-2 text-white">
                                                <Database className="h-4 w-4 text-cyan-400" />
                                                <span className="font-medium">{server.name}</span>
                                            </div>
                                            <div className="mt-2 grid gap-1 text-xs text-zinc-400">
                                                <div>cache status: {String(metadata?.status ?? 'pending')}</div>
                                                <div>source: {String(metadata?.metadataSource ?? 'none')}</div>
                                                <div>tools cached: {String(metadata?.toolCount ?? 0)}</div>
                                                <div className="break-all">last binary load: {String(metadata?.lastSuccessfulBinaryLoadAt ?? 'never')}</div>
                                            </div>
                                        </div>
                                        <div className="flex flex-wrap gap-2">
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                disabled={pending || clearMetadataCacheMutation.isPending}
                                                onClick={() => clearMetadataCacheMutation.mutate({ uuid: server.uuid })}
                                            >
                                                {clearMetadataCacheMutation.isPending && clearMetadataCacheMutation.variables?.uuid === server.uuid ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <Database className="mr-2 h-3.5 w-3.5" />}
                                                Clear cache
                                            </Button>
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                disabled={pending}
                                                onClick={() => reloadMetadataMutation.mutate({ uuid: server.uuid, mode: 'cache' })}
                                            >
                                                {pending ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <Database className="mr-2 h-3.5 w-3.5" />}
                                                Reload cache
                                            </Button>
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                disabled={pending}
                                                onClick={() => reloadMetadataMutation.mutate({ uuid: server.uuid, mode: 'binary' })}
                                            >
                                                {pending ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="mr-2 h-3.5 w-3.5" />}
                                                Refresh binary
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            );
                        })
                    )}
                </CardContent>
            </Card>

            <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader>
                    <CardTitle className="text-white">Operational Coverage (Live)</CardTitle>
                </CardHeader>
                <CardContent className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
                    <CoverageCard
                        title="expert.getStatus"
                        lines={[
                            `researcher: ${String(expertStatus?.researcher ?? 'unknown')}`,
                            `coder: ${String(expertStatus?.coder ?? 'unknown')}`,
                        ]}
                    />

                    <CoverageCard
                        title="session.getState"
                        lines={[
                            `autoDrive: ${String((sessionState as any)?.isAutoDriveActive ?? false)}`,
                            `activeGoal: ${String((sessionState as any)?.activeGoal ?? 'none')}`,
                        ]}
                    />

                    <CoverageCard
                        title="agentMemory.stats"
                        lines={[
                            `session: ${String(memoryStats?.session ?? 0)}`,
                            `working: ${String(memoryStats?.working ?? 0)}`,
                            `longTerm: ${String(memoryStats?.longTerm ?? 0)}`,
                            `total: ${String(memoryStats?.total ?? 0)}`,
                        ]}
                    />

                    <div className="rounded border border-zinc-800 bg-zinc-950/50 p-3 xl:col-span-2">
                        <div className="text-sm text-zinc-200 font-medium mb-2">serverHealth.check</div>
                        <div className="flex items-center gap-2 mb-2">
                            <input
                                value={healthServerUuid}
                                onChange={(e) => setHealthServerUuid(e.target.value)}
                                placeholder={firstServerUuid ? `default: ${firstServerUuid}` : 'enter server UUID'}
                                className="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500"
                            />
                        </div>
                        {effectiveHealthServerUuid ? (
                            <div className="text-xs text-zinc-300 space-y-1">
                                <div>status: {String(serverHealth?.status ?? 'loading')}</div>
                                <div>crashCount: {String(serverHealth?.crashCount ?? 0)}</div>
                                <div>maxAttempts: {String(serverHealth?.maxAttempts ?? 0)}</div>
                                <div className="text-zinc-500 break-all">uuid: {effectiveHealthServerUuid}</div>
                            </div>
                        ) : (
                            <div className="text-xs text-zinc-500">No server UUID available yet.</div>
                        )}
                    </div>

                    <CoverageCard
                        title="shell.getSystemHistory"
                        lines={[
                            `entries: ${String(shellHistory?.length ?? 0)}`,
                            ...(shellHistory && shellHistory.length > 0
                                ? [`last: ${String((shellHistory[0] as any)?.command ?? 'n/a').slice(0, 60)}`]
                                : ['last: n/a']),
                        ]}
                    />
                </CardContent>
            </Card>
        </div>
    );
}

function CoverageCard({ title, lines }: { title: string; lines: string[] }) {
    return (
        <div className="rounded border border-zinc-800 bg-zinc-950/50 p-3">
            <div className="text-sm text-zinc-200 font-medium mb-2">{title}</div>
            <div className="space-y-1 text-xs text-zinc-300">
                {lines.map((line) => (
                    <div key={line} className="break-all">{line}</div>
                ))}
            </div>
        </div>
    );
}

function StatCard({
    title,
    value,
    icon: Icon,
    tone,
}: {
    title: string;
    value: string;
    icon: any;
    tone: string;
}) {
    return (
        <Card className="bg-zinc-900 border-zinc-800">
            <CardContent className="p-5">
                <div className="flex items-center justify-between mb-1">
                    <span className="text-zinc-500 text-sm">{title}</span>
                    <Icon className={`h-4 w-4 ${tone}`} />
                </div>
                <div className="text-2xl font-bold text-white">{value}</div>
            </CardContent>
        </Card>
    );
}

function ReadinessCard({
    title,
    healthyCount,
    totalCount,
    healthyLabel,
    unhealthyLabel,
}: {
    title: string;
    healthyCount: number;
    totalCount: number;
    healthyLabel: string;
    unhealthyLabel: string;
}) {
    const unhealthy = Math.max(0, totalCount - healthyCount);

    return (
        <Card className="bg-zinc-900 border-zinc-800">
            <CardHeader>
                <CardTitle className="text-white text-base">{title}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
                <div className="flex items-center justify-between rounded border border-zinc-800 p-2 bg-zinc-950/50">
                    <span className="text-zinc-400 text-sm">{healthyLabel}</span>
                    <span className="inline-flex items-center gap-1 text-emerald-400 text-sm">
                        <CheckCircle2 className="h-4 w-4" /> {healthyCount}
                    </span>
                </div>
                <div className="flex items-center justify-between rounded border border-zinc-800 p-2 bg-zinc-950/50">
                    <span className="text-zinc-400 text-sm">{unhealthyLabel}</span>
                    <span className="inline-flex items-center gap-1 text-red-400 text-sm">
                        <XCircle className="h-4 w-4" /> {unhealthy}
                    </span>
                </div>
            </CardContent>
        </Card>
    );
}
