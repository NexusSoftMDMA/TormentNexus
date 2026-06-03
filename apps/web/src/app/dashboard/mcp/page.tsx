"use client";

import Link from 'next/link';
import { useMemo, useState } from 'react';
import type { ComponentType, FormEvent } from 'react';
import { Button, Card, CardContent, CardHeader, CardTitle } from '@tormentnexus/ui';
import { trpc } from '@/utils/trpc';
import { buildBulkImportServers } from '@/lib/mcp-import';
import { toast } from 'sonner';
import { buildDashboardServerRecords, buildServerToolActionLinks, getBulkMetadataTargetUuids, getManagedServerDiscoverySummary, hasStaleReadyMetadata, isLocalCompatMetadataSource } from './mcp-dashboard-utils';
import { EditMcpServer } from '@/components/mcp/EditMcpServer';
import {
    Activity,
    AlertTriangle,
    Database,
    Eye,
    ExternalLink,
    HeartPulse,
    Layers,
    Loader2,
    Network,
    Pencil,
    Play,
    Plus,
    RefreshCw,
    Search,
    Server,
    Shield,
    Trash2,
    Upload,
    Wrench,
    Zap,
} from 'lucide-react';

type AggregatedServer = {
    uuid?: string;
    name: string;
    status: string;
    toolCount: number;
    runtimeState?: string;
    warmupState?: string;
    runtimeConnected?: boolean;
    advertisedToolCount?: number;
    advertisedSource?: string;
    lastConnectedAt?: string | null;
    lastError?: string | null;
    metadataStatus?: string;
    metadataSource?: string;
    metadataToolCount?: number;
    lastSuccessfulBinaryLoadAt?: string;
    config?: {
        command?: string;
        args?: string[];
        env?: string[];
    };
};

type ManagedServerMetadata = {
    uuid: string;
    name: string;
    type?: 'STDIO' | 'SSE' | 'STREAMABLE_HTTP';
    description?: string | null;
    command?: string | null;
    args?: string[];
    env?: Record<string, string>;
    url?: string | null;
    bearerToken?: string | null;
    headers?: Record<string, string>;
    always_on?: boolean;
    _meta?: {
        status?: string;
        metadataSource?: string;
        toolCount?: number;
        lastSuccessfulBinaryLoadAt?: string;
    } | null;
};

type BulkDiscoveryOperationState = {
    mode: 'all' | 'unresolved';
    completedCount: number;
    totalCount: number;
};

type AggregatedTool = {
    name: string;
    description: string;
    server: string;
};

type StatusSummary = {
    initialized: boolean;
    serverCount: number;
    toolCount: number;
    connectedCount: number;
};

type QuickLinkCardProps = {
    title: string;
    description: string;
    href: string;
    accentClass: string;
    icon: ComponentType<{ className?: string }>;
};

type ServerConfigInput = {
    name: string;
    type: 'STDIO' | 'SSE' | 'STREAMABLE_HTTP';
    command: string;
    args: string;
    url: string;
    bearerToken: string;
    headers: string;
    env: string;
};

type BulkImportMutationResult =
        | unknown[]
        | {
                imported?: number;
            };

type BulkImportClassification = {
    newNames: string[];
    updatingNames: string[];
};

type ManagedServerHealth = {
    status?: string;
    crashCount?: number;
    maxAttempts?: number;
};

function maskValue(value?: string | null): string {
    if (!value) {
        return 'none';
    }

    if (value.length <= 8) {
        return '•'.repeat(value.length);
    }

    return `${value.slice(0, 4)}${'•'.repeat(Math.min(Math.max(value.length - 8, 4), 12))}${value.slice(-4)}`;
}

function QuickLinkCard(props: QuickLinkCardProps): React.JSX.Element {
    const Icon = props.icon;

    return (
        <Link href={props.href} className="group">
            <Card className="h-full border-zinc-800 bg-zinc-900 transition-colors hover:border-zinc-700">
                <CardContent className="space-y-3 p-5">
                    <div className="flex items-center gap-2">
                        <Icon className={`h-4 w-4 ${props.accentClass}`} />
                        <div className="text-sm font-semibold text-white">{props.title}</div>
                    </div>
                    <p className="text-sm leading-relaxed text-zinc-500">{props.description}</p>
                    <div className="inline-flex items-center gap-1 text-xs text-zinc-400 group-hover:text-white">
                        Open
                        <ExternalLink className="h-3.5 w-3.5" />
                    </div>
                </CardContent>
            </Card>
        </Link>
    );
}

function StatusBadge({ status }: { status: string }): React.JSX.Element {
    const normalized = status.toLowerCase();
    let tone = 'border-zinc-700 bg-zinc-800 text-zinc-400';

    if (normalized === 'connected' || normalized === 'active' || normalized === 'ready' || normalized === 'configured') {
        tone = 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300';
    } else if (normalized.includes('pending') || normalized.includes('cold') || normalized.includes('starting')) {
        tone = 'border-amber-500/20 bg-amber-500/10 text-amber-300';
    } else if (normalized.includes('error') || normalized.includes('failed') || normalized.includes('dead')) {
        tone = 'border-red-500/20 bg-red-500/10 text-red-300';
    }

    return (
        <span className={`inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] uppercase tracking-wider ${tone}`}>
            {status}
        </span>
    );
}

function ServerInspectionPanel(
    {
        server,
        runtime,
        health,
        onClose,
    }: {
        server?: ManagedServerMetadata;
        runtime?: AggregatedServer;
        health?: ManagedServerHealth;
        onClose: () => void;
    },
): React.JSX.Element {
    return (
        <Card className="border-zinc-700 border-l-4 border-l-cyan-500 bg-zinc-900">
            <CardHeader className="flex flex-row items-start justify-between gap-4">
                <div>
                    <CardTitle className="text-base text-white">Inspect downstream MCP server</CardTitle>
                    <p className="mt-1 text-sm text-zinc-500">
                        Review the effective transport config, cached-vs-runtime posture, and current health counters without leaving the control plane.
                    </p>
                </div>
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={onClose}
                    title="Close server inspection panel"
                    aria-label="Close server inspection panel"
                    className="text-zinc-500 hover:text-white"
                >
                    Close
                </Button>
            </CardHeader>
            <CardContent>
                {!server ? (
                    <div className="rounded-lg border border-dashed border-zinc-800 p-6 text-sm text-zinc-500">
                        Loading server details…
                    </div>
                ) : (
                    <div className="space-y-4">
                        <div className="flex flex-wrap items-start justify-between gap-4">
                            <div>
                                <div className="text-lg font-semibold text-white">{server.name}</div>
                                <div className="mt-1 text-sm text-zinc-400">
                                    Transport <span className="font-medium text-white">{server.type ?? 'STDIO'}</span>
                                </div>
                            </div>
                            <div className="flex flex-wrap gap-2">
                                <StatusBadge status={server._meta?.status ?? 'pending'} />
                                {runtime ? <StatusBadge status={runtime.runtimeState ?? runtime.status} /> : null}
                            </div>
                        </div>

                        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4 text-sm">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Runtime</div>
                                <div className="mt-2 space-y-1 text-zinc-300">
                                    <div>Runtime state: <span className="font-semibold text-white">{runtime?.runtimeState ?? runtime?.status ?? 'unknown'}</span></div>
                                    <div>Warmup: <span className="font-semibold text-white">{runtime?.warmupState ?? 'idle'}</span></div>
                                    <div>Advertised source: <span className="font-semibold text-white">{runtime?.advertisedSource ?? 'unknown'}</span></div>
                                    <div>Advertised tools: <span className="font-semibold text-white">{String(runtime?.advertisedToolCount ?? runtime?.metadataToolCount ?? 0)}</span></div>
                                </div>
                            </div>
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4 text-sm">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Connection</div>
                                <div className="mt-2 space-y-1 text-zinc-300">
                                    <div>Command: <span className="font-mono text-xs text-white">{server.command ?? 'n/a'}</span></div>
                                    <div>URL: <span className="break-all font-mono text-xs text-white">{server.url ?? 'n/a'}</span></div>
                                    <div>Last runtime connect: <span className="font-semibold text-white">{runtime?.lastConnectedAt ?? 'never'}</span></div>
                                </div>
                            </div>
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4 text-sm">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Metadata</div>
                                <div className="mt-2 space-y-1 text-zinc-300">
                                    <div>Status: <span className="font-semibold text-white">{String(server._meta?.status ?? 'pending')}</span></div>
                                    <div>Source: <span className="font-semibold text-white">{String(server._meta?.metadataSource ?? 'none')}</span></div>
                                    <div>Cached tools: <span className="font-semibold text-white">{String(server._meta?.toolCount ?? 0)}</span></div>
                                    <div>Last binary load: <span className="font-semibold text-white">{String(server._meta?.lastSuccessfulBinaryLoadAt ?? 'never')}</span></div>
                                </div>
                            </div>
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4 text-sm">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Health</div>
                                <div className="mt-2 space-y-1 text-zinc-300">
                                    <div>Status: <span className="font-semibold text-white">{health?.status ?? 'unknown'}</span></div>
                                    <div>Crash count: <span className="font-semibold text-white">{health?.crashCount ?? 0}</span></div>
                                    <div>Max attempts: <span className="font-semibold text-white">{health?.maxAttempts ?? 0}</span></div>
                                </div>
                            </div>
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4 text-sm">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Description</div>
                                <div className="mt-2 text-zinc-300">{server.description ?? 'No description provided.'}</div>
                                {runtime?.lastError ? (
                                    <div className="mt-3 rounded border border-rose-500/20 bg-rose-500/10 p-3 text-xs text-rose-100">
                                        <div className="font-semibold text-white">Latest runtime error</div>
                                        <div className="mt-1 break-words">{runtime.lastError}</div>
                                    </div>
                                ) : null}
                            </div>
                        </div>

                        <div className="grid gap-4 xl:grid-cols-2">
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Args</div>
                                <pre className="mt-2 whitespace-pre-wrap break-all font-mono text-xs text-zinc-300">
                                    {JSON.stringify(server.args ?? [], null, 2)}
                                </pre>
                            </div>
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Environment</div>
                                <pre className="mt-2 whitespace-pre-wrap break-all font-mono text-xs text-zinc-300">
                                    {JSON.stringify(server.env ?? {}, null, 2)}
                                </pre>
                            </div>
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Headers</div>
                                <pre className="mt-2 whitespace-pre-wrap break-all font-mono text-xs text-zinc-300">
                                    {JSON.stringify(server.headers ?? {}, null, 2)}
                                </pre>
                            </div>
                            <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Sensitive fields</div>
                                <div className="mt-2 space-y-1 text-sm text-zinc-300">
                                    <div>Bearer token: <span className="font-mono text-xs text-white">{maskValue(server.bearerToken)}</span></div>
                                    <div>Description: <span className="text-white">{server.description ?? 'none'}</span></div>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}

function AddServerForm({ onDone }: { onDone: () => void }): React.JSX.Element {
    const mcpServersClient = trpc.mcpServers as any;
    const [formData, setFormData] = useState<ServerConfigInput>({
        name: '',
        type: 'STDIO',
        command: 'npx',
        args: '',
        url: '',
        bearerToken: '',
        headers: '',
        env: '',
    });

    const createMutation = mcpServersClient.create.useMutation({
        onSuccess: () => {
            toast.success('Server added successfully');
            onDone();
        },
        onError: (error: any) => {
            toast.error(error.message);
        },
    });

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        let parsedEnv: Record<string, string> = {};
        if (formData.env.trim()) {
            try {
                parsedEnv = JSON.parse(formData.env) as Record<string, string>;
            } catch {
                toast.error('Environment variables must be valid JSON');
                return;
            }
        }

        let parsedHeaders: Record<string, string> = {};
        if (formData.headers.trim()) {
            try {
                parsedHeaders = JSON.parse(formData.headers) as Record<string, string>;
            } catch {
                toast.error('Headers must be valid JSON');
                return;
            }
        }

        createMutation.mutate({
            name: formData.name,
            type: formData.type,
            command: formData.type === 'STDIO' ? formData.command : undefined,
            args: formData.type === 'STDIO' ? formData.args.split(' ').filter(Boolean) : undefined,
            url: formData.type === 'STDIO' ? undefined : formData.url.trim(),
            bearerToken: formData.type === 'STDIO' ? undefined : (formData.bearerToken.trim() || undefined),
            headers: formData.type === 'STDIO' ? undefined : parsedHeaders,
            env: parsedEnv,
            metadataStrategy: 'auto',
        });
    }

    return (
        <Card className="bg-zinc-900 border-zinc-700 border-l-4 border-l-blue-600">
            <CardHeader className="flex flex-row items-start justify-between">
                <div>
                    <CardTitle className="text-white text-base">Add downstream MCP server</CardTitle>
                    <p className="text-sm text-zinc-500 mt-1">Register another MCP endpoint under TormentNexus’s aggregated router.</p>
                </div>
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={onDone}
                    title="Close add-server form"
                    aria-label="Close add-server form"
                    className="text-zinc-500 hover:text-white"
                >
                    Close
                </Button>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-xs uppercase tracking-wider text-zinc-500 mb-1.5">Server name</label>
                        <input
                            required
                            value={formData.name}
                            onChange={(event) => setFormData((current) => ({ ...current, name: event.target.value }))}
                            title="Unique display name used for this downstream MCP server"
                            aria-label="MCP server name"
                            className="w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                            placeholder="github"
                        />
                    </div>

                    <div>
                        <label className="block text-xs uppercase tracking-wider text-zinc-500 mb-1.5">Connection type</label>
                        <select
                            value={formData.type}
                            onChange={(event) => setFormData((current) => ({ ...current, type: event.target.value as ServerConfigInput['type'] }))}
                            title="Transport used to connect to the downstream MCP server"
                            aria-label="MCP server connection type"
                            className="w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                        >
                            <option value="STDIO">STDIO</option>
                            <option value="SSE">SSE</option>
                            <option value="STREAMABLE_HTTP">STREAMABLE_HTTP</option>
                        </select>
                    </div>

                    {formData.type === 'STDIO' ? (
                        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                            <div>
                                <label className="block text-xs uppercase tracking-wider text-zinc-500 mb-1.5">Command</label>
                                <input
                                    required
                                    value={formData.command}
                                    onChange={(event) => setFormData((current) => ({ ...current, command: event.target.value }))}
                                    title="Executable command used to launch the downstream MCP server"
                                    aria-label="MCP server command"
                                    className="w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                                />
                            </div>
                            <div className="md:col-span-2">
                                <label className="block text-xs uppercase tracking-wider text-zinc-500 mb-1.5">Args</label>
                                <input
                                    value={formData.args}
                                    onChange={(event) => setFormData((current) => ({ ...current, args: event.target.value }))}
                                    title="Command arguments for the server process, separated by spaces"
                                    aria-label="MCP server command arguments"
                                    className="w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                                    placeholder="-y @modelcontextprotocol/server-memory"
                                />
                            </div>
                        </div>
                    ) : (
                        <>
                            <div>
                                <label className="block text-xs uppercase tracking-wider text-zinc-500 mb-1.5">URL</label>
                                <input
                                    required
                                    value={formData.url}
                                    onChange={(event) => setFormData((current) => ({ ...current, url: event.target.value }))}
                                    title="Remote MCP endpoint URL"
                                    aria-label="MCP server URL"
                                    className="w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                                    placeholder="https://example.com/mcp"
                                />
                            </div>

                            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                <div>
                                    <label className="block text-xs uppercase tracking-wider text-zinc-500 mb-1.5">Bearer token</label>
                                    <input
                                        value={formData.bearerToken}
                                        onChange={(event) => setFormData((current) => ({ ...current, bearerToken: event.target.value }))}
                                        title="Optional bearer token used for remote MCP authentication"
                                        aria-label="MCP server bearer token"
                                        className="w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                                        placeholder="Optional bearer token"
                                    />
                                </div>
                                <div>
                                    <label className="block text-xs uppercase tracking-wider text-zinc-500 mb-1.5">Headers</label>
                                    <textarea
                                        value={formData.headers}
                                        onChange={(event) => setFormData((current) => ({ ...current, headers: event.target.value }))}
                                        title="Optional request headers as a JSON object"
                                        aria-label="MCP server headers JSON"
                                        className="h-24 w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                                        placeholder='{"Authorization":"Bearer ..."}'
                                    />
                                </div>
                            </div>
                        </>
                    )}

                    <div>
                        <label className="block text-xs uppercase tracking-wider text-zinc-500 mb-1.5">Environment variables</label>
                        <textarea
                            value={formData.env}
                            onChange={(event) => setFormData((current) => ({ ...current, env: event.target.value }))}
                            title="Optional environment variables as a JSON object"
                            aria-label="MCP server environment variables JSON"
                            className="h-24 w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                            placeholder='{"API_KEY":"secret"}'
                        />
                    </div>

                    <div className="rounded-md border border-zinc-800 bg-zinc-950/60 p-3 text-xs text-zinc-400">
                        New servers use <span className="font-semibold text-white">Auto</span> discovery by default. If you ever need to force a clean rediscovery, use the <span className="font-semibold text-white">Clear cache</span> button on the server card.
                    </div>

                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            disabled={createMutation.isPending}
                            title="Register this downstream MCP server in TormentNexus"
                            aria-label="Add downstream MCP server"
                            className="bg-blue-600 hover:bg-blue-500 text-white"
                        >
                            {createMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                            Add Server
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}

function classifyImportNames(importedNames: string[], existingServerNames: string[]): BulkImportClassification {
    const existingNames = new Set(existingServerNames.map((name) => name.trim()).filter(Boolean));

    return importedNames.reduce<BulkImportClassification>((result, name) => {
        if (existingNames.has(name)) {
            result.updatingNames.push(name);
        } else {
            result.newNames.push(name);
        }

        return result;
    }, { newNames: [], updatingNames: [] });
}

function BulkImportForm({ onDone, existingServerNames }: { onDone: () => void; existingServerNames: string[] }): React.JSX.Element {
    const [jsonConfig, setJsonConfig] = useState('');
    const [lastImportedNames, setLastImportedNames] = useState<string[]>([]);

    const preview = useMemo(() => {
        if (!jsonConfig.trim()) {
            return null;
        }

        try {
            return {
                data: buildBulkImportServers(jsonConfig),
                error: null,
            } as const;
        } catch (error) {
            return {
                data: null,
                error: error instanceof Error ? error.message : 'Invalid JSON',
            } as const;
        }
    }, [jsonConfig]);

    const importClassification = useMemo(() => {
        if (!preview?.data) {
            return null;
        }

        return classifyImportNames(preview.data.importedNames, existingServerNames);
    }, [existingServerNames, preview]);

    const importMutation = trpc.mcpServers.bulkImport.useMutation({
        onSuccess: (data: BulkImportMutationResult) => {
            const importedCount = Array.isArray(data)
                ? data.length
                : typeof data === 'object' && data !== null && typeof data.imported === 'number'
                    ? data.imported
                    : 0;

            const importedPreview = lastImportedNames.length > 0
                ? ` ${lastImportedNames.slice(0, 3).join(', ')}${lastImportedNames.length > 3 ? `, +${lastImportedNames.length - 3} more` : ''}.`
                : '';

            toast.success(`Imported ${importedCount} server${importedCount === 1 ? '' : 's'}.${importedPreview}`);
            onDone();
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    function handleImport(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        try {
            const { servers, importedNames } = buildBulkImportServers(jsonConfig);

            setLastImportedNames(importedNames);
            importMutation.mutate(servers);
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Invalid JSON';
            toast.error(message);
        }
    }

    return (
        <Card className="bg-zinc-900 border-zinc-700 border-l-4 border-l-purple-600">
            <CardHeader className="flex flex-row items-start justify-between">
                <div>
                    <CardTitle className="text-white text-base">Bulk import MCP config</CardTitle>
                    <p className="text-sm text-zinc-500 mt-1">Import existing client configs and fold them into TormentNexus’s router.</p>
                </div>
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={onDone}
                    title="Close bulk import form"
                    aria-label="Close bulk import form"
                    className="text-zinc-500 hover:text-white"
                >
                    Close
                </Button>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleImport} className="space-y-4">
                    <textarea
                        value={jsonConfig}
                        onChange={(event) => setJsonConfig(event.target.value)}
                        title="Paste MCP JSON or JSONC config to preview and import server entries"
                        aria-label="Bulk MCP config input"
                        className="h-56 w-full rounded-md border border-zinc-800 bg-zinc-950 p-2.5 text-sm text-white outline-none focus:ring-1 focus:ring-purple-500"
                        placeholder='{ "mcpServers": { "memory": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-memory"] } } }'
                    />
                    <div className="rounded-md border border-zinc-800 bg-zinc-950/60 p-3 text-xs text-zinc-400">
                        Paste a full MCP config, JSONC with comments/trailing commas, or just the raw <span className="font-semibold text-white">mcpServers</span> object.
                    </div>
                    <div className="rounded-md border border-zinc-800 bg-zinc-950/60 p-3 text-xs text-zinc-400">
                        Imports currently <span className="font-semibold text-white">merge by server name</span>: matching names are updated, untouched servers stay in place.
                    </div>
                    {preview ? (
                        preview.error ? (
                            <div className="rounded-md border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-200">
                                {preview.error}
                            </div>
                        ) : (
                            <div className="rounded-md border border-emerald-500/20 bg-emerald-500/10 p-3 text-xs text-emerald-100 space-y-2">
                                <div>
                                    Previewing <span className="font-semibold text-white">{preview.data.servers.length}</span> server{preview.data.servers.length === 1 ? '' : 's'}.
                                </div>
                                {importClassification ? (
                                    <div className="space-y-2">
                                        {importClassification.newNames.length > 0 ? (
                                            <div className="space-y-1.5">
                                                <div className="text-[11px] uppercase tracking-wider text-emerald-200/80">
                                                    New servers · {importClassification.newNames.length}
                                                </div>
                                                <div className="flex flex-wrap gap-2">
                                                    {importClassification.newNames.slice(0, 8).map((name) => (
                                                        <span key={`new-${name}`} className="rounded-full border border-emerald-400/20 bg-zinc-950/50 px-2 py-1 text-[11px] text-emerald-100">
                                                            {name}
                                                        </span>
                                                    ))}
                                                    {importClassification.newNames.length > 8 ? (
                                                        <span className="rounded-full border border-zinc-700 bg-zinc-950/50 px-2 py-1 text-[11px] text-zinc-300">
                                                            +{importClassification.newNames.length - 8} more
                                                        </span>
                                                    ) : null}
                                                </div>
                                            </div>
                                        ) : null}
                                        {importClassification.updatingNames.length > 0 ? (
                                            <div className="space-y-1.5">
                                                <div className="text-[11px] uppercase tracking-wider text-amber-200/80">
                                                    Updating existing · {importClassification.updatingNames.length}
                                                </div>
                                                <div className="flex flex-wrap gap-2">
                                                    {importClassification.updatingNames.slice(0, 8).map((name) => (
                                                        <span key={`updating-${name}`} className="rounded-full border border-amber-400/20 bg-zinc-950/50 px-2 py-1 text-[11px] text-amber-100">
                                                            {name}
                                                        </span>
                                                    ))}
                                                    {importClassification.updatingNames.length > 8 ? (
                                                        <span className="rounded-full border border-zinc-700 bg-zinc-950/50 px-2 py-1 text-[11px] text-zinc-300">
                                                            +{importClassification.updatingNames.length - 8} more
                                                        </span>
                                                    ) : null}
                                                </div>
                                            </div>
                                        ) : null}
                                    </div>
                                ) : null}
                            </div>
                        )
                    ) : null}
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            disabled={importMutation.isPending || Boolean(preview?.error)}
                            title="Import all valid server definitions from this config into TormentNexus"
                            aria-label="Import MCP server configuration"
                            className="bg-purple-600 hover:bg-purple-500 text-white"
                        >
                            {importMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                            Import
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}

export default function MCPDashboard(): React.JSX.Element {
    const trpcUtils = trpc.useUtils();
    const { data: servers, isLoading: isLoadingServers, refetch: refetchServers } = trpc.mcp.listServers.useQuery();
    const mcpServersClient = trpc.mcpServers as any;
    const { data: managedServers, refetch: refetchManagedServers } = mcpServersClient.list.useQuery();
    const { data: tools, isLoading: isLoadingTools, refetch: refetchTools } = trpc.mcp.listTools.useQuery();
    const { data: status, refetch: refetchStatus } = trpc.mcp.getStatus.useQuery(undefined, { refetchInterval: 5000 });
    const [editingServerUuid, setEditingServerUuid] = useState<string | null>(null);
    const [inspectingServerUuid, setInspectingServerUuid] = useState<string | null>(null);
    const [deletingServerUuid, setDeletingServerUuid] = useState<string | null>(null);
    const [resettingServerUuid, setResettingServerUuid] = useState<string | null>(null);
    const [isAddOpen, setIsAddOpen] = useState(false);
    const [isImportOpen, setIsImportOpen] = useState(false);
    const [bulkRefreshState, setBulkRefreshState] = useState<BulkDiscoveryOperationState | null>(null);
    const reloadMetadataMutation = mcpServersClient.reloadMetadata.useMutation();
    const clearMetadataCacheMutation = mcpServersClient.clearMetadataCache.useMutation();
    const deleteServerMutation = mcpServersClient.delete.useMutation();
    const updateServerMutation = mcpServersClient.update.useMutation();
    const resetServerHealthMutation = trpc.serverHealth.reset.useMutation();
    const { data: editingServer } = mcpServersClient.get.useQuery(
        { uuid: editingServerUuid ?? '' },
        { enabled: Boolean(editingServerUuid) },
    );
    const { data: inspectingServer, refetch: refetchInspectingServer } = mcpServersClient.get.useQuery(
        { uuid: inspectingServerUuid ?? '' },
        { enabled: Boolean(inspectingServerUuid) },
    );
    const { data: inspectingServerHealth } = trpc.serverHealth.check.useQuery(
        { serverUuid: inspectingServerUuid ?? '' },
        { enabled: Boolean(inspectingServerUuid) },
    );

    const managedServerList = (managedServers || []) as ManagedServerMetadata[];
    const discoverySummary = getManagedServerDiscoverySummary(managedServerList);
    const unresolvedDiscoveryTargetUuids = getBulkMetadataTargetUuids(managedServerList, 'unresolved');
    const allDiscoveryTargetUuids = getBulkMetadataTargetUuids(managedServerList, 'all');
    const serverList = buildDashboardServerRecords((servers || []) as AggregatedServer[], managedServerList);
    const inspectingRuntimeServer = useMemo(() => serverList.find((server) => server.uuid === inspectingServerUuid), [inspectingServerUuid, serverList]);
    const toolList = (tools || []) as AggregatedTool[];
    const existingServerNames = useMemo(() => Array.from(new Set([
        ...serverList.map((server) => server.name),
        ...managedServerList.map((server) => server.name),
    ])).sort(), [managedServerList, serverList]);
    const summary = (status || {
        initialized: false,
        serverCount: 0,
        toolCount: 0,
        connectedCount: 0,
    }) as StatusSummary;

    const topTools = toolList.slice(0, 8);
    const bulkActionsDisabled = bulkRefreshState !== null || reloadMetadataMutation.isPending || clearMetadataCacheMutation.isPending;
    const unresolvedActionableCount = unresolvedDiscoveryTargetUuids.length;
    const allActionableCount = allDiscoveryTargetUuids.length;
    const unresolvedWithoutActionCount = Math.max(discoverySummary.unresolvedCount - unresolvedActionableCount, 0);
    const localCompatActive = discoverySummary.localCompatCount > 0;

    async function refreshDashboardQueries() {
        await Promise.all([
            refetchServers(),
            refetchManagedServers(),
            refetchTools(),
            refetchStatus(),
            inspectingServerUuid ? refetchInspectingServer() : Promise.resolve(undefined),
        ]);
    }

    async function handleDeleteServer(uuid: string, serverName: string) {
        const confirmed = typeof window === 'undefined'
            ? true
            : window.confirm(`Delete MCP server '${serverName}'? This removes the server from TormentNexus configuration.`);
        if (!confirmed) {
            return;
        }

        try {
            setDeletingServerUuid(uuid);
            await deleteServerMutation.mutateAsync({ uuid });
            if (editingServerUuid === uuid) {
                setEditingServerUuid(null);
            }
            if (inspectingServerUuid === uuid) {
                setInspectingServerUuid(null);
            }
            await refreshDashboardQueries();
            toast.success(`Deleted ${serverName}.`);
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to delete MCP server.';
            toast.error(message);
        } finally {
            setDeletingServerUuid(null);
        }
    }

    async function handleToggleAlwaysOn(serverUuid: string, serverName: string, currentValue: boolean) {
        try {
            await updateServerMutation.mutateAsync({
                uuid: serverUuid,
                always_on: !currentValue,
            });
            await refreshDashboardQueries();
            toast.success(`'Always On' setting updated for ${serverName}.`);
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to update setting.';
            toast.error(message);
        }
    }

    async function handleResetHealth(serverUuid: string, serverName: string) {
        try {
            setResettingServerUuid(serverUuid);
            await resetServerHealthMutation.mutateAsync({ serverUuid });
            if (inspectingServerUuid === serverUuid) {
                await refetchInspectingServer();
            }
            toast.success(`Reset health counters for ${serverName}.`);
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to reset server health.';
            toast.error(message);
        } finally {
            setResettingServerUuid(null);
        }
    }

    async function handleTestServer(serverUuid: string, serverName: string) {
        try {
            const result = await trpcUtils.serverHealth.check.fetch({ serverUuid });
            setInspectingServerUuid(serverUuid);
            toast.success(`${serverName} health: ${result.status} (${result.crashCount} crashes tracked).`);
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to test MCP server health.';
            toast.error(message);
        }
    }

    async function handleReloadMetadata(
        input: { uuid: string; mode: 'auto' | 'binary' | 'cache' },
        options?: { notify?: boolean; refresh?: boolean },
    ) {
        try {
            const result = await reloadMetadataMutation.mutateAsync(input);

            if (options?.refresh !== false) {
                await refreshDashboardQueries();
            }

            if (options?.notify !== false) {
                toast.success(`Reloaded metadata for ${result.server.name} from ${result.metadata.metadataSource ?? 'metadata cache'}.`);
            }

            return result;
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to reload MCP metadata.';
            toast.error(message);
            throw error;
        }
    }

    async function handleClearMetadataCache(uuid: string) {
        try {
            const result = await clearMetadataCacheMutation.mutateAsync({ uuid });
            await refreshDashboardQueries();
            toast.success(`Cleared cached metadata for ${result.server.name}. The next auto discovery will reload from the binary.`);
            return result;
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to clear MCP metadata cache.';
            toast.error(message);
            throw error;
        }
    }

    async function handleBulkBinaryRefresh(targetMode: 'all' | 'unresolved') {
        const targetUuids = targetMode === 'all'
            ? allDiscoveryTargetUuids
            : unresolvedDiscoveryTargetUuids;

        if (targetUuids.length === 0) {
            toast.info(targetMode === 'all'
                ? 'No managed MCP servers are available for discovery refresh.'
                : 'All managed MCP servers already have ready metadata.');
            return;
        }

        let completedCount = 0;
        setBulkRefreshState({
            mode: targetMode,
            completedCount,
            totalCount: targetUuids.length,
        });

        try {
            for (const uuid of targetUuids) {
                await handleReloadMetadata({ uuid, mode: 'binary' }, { notify: false, refresh: false });
                completedCount += 1;
                setBulkRefreshState({
                    mode: targetMode,
                    completedCount,
                    totalCount: targetUuids.length,
                });
            }

            await refreshDashboardQueries();
            toast.success(`Refreshed binary metadata for ${completedCount} MCP server${completedCount === 1 ? '' : 's'}.`);
        } catch (error) {
            await refreshDashboardQueries();
            const message = error instanceof Error ? error.message : 'Bulk discovery refresh failed.';
            toast.error(`Bulk refresh stopped after ${completedCount} of ${targetUuids.length} server${targetUuids.length === 1 ? '' : 's'}. ${message}`);
        } finally {
            setBulkRefreshState(null);
        }
    }

    async function handleLoadTestAndCacheServer(serverUuid: string, serverName: string) {
        try {
            await handleReloadMetadata({ uuid: serverUuid, mode: 'binary' }, { notify: false, refresh: false });
            const health = await trpcUtils.serverHealth.check.fetch({ serverUuid });
            await refreshDashboardQueries();
            setInspectingServerUuid(serverUuid);
            toast.success(`${serverName} loaded and cached. Health: ${health.status}.`);
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to load, test, and cache MCP server.';
            toast.error(message);
        }
    }

    return (
        <div className="p-4 sm:p-6 xl:p-8 space-y-8">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">MCP Router Control Plane</h1>
                    <p className="text-zinc-500 mt-2 max-w-3xl">
                        TormentNexus should read like the ultimate MCP aggregator/router first: one operator surface, many downstream servers, semantic search and grouping, lifecycle control, traffic visibility, and client config sync.
                    </p>
                </div>
                <div className="flex flex-wrap gap-2">
                    <Button
                        variant="outline"
                        className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                        onClick={() => setIsImportOpen((value) => !value)}
                        title="Open or close the bulk config import panel"
                        aria-label="Toggle MCP config import panel"
                    >
                        <Upload className="mr-2 h-4 w-4" /> Import Config
                    </Button>
                    <Button
                        className="bg-blue-600 hover:bg-blue-500 text-white"
                        onClick={() => setIsAddOpen((value) => !value)}
                        title="Open or close the add-server panel"
                        aria-label="Toggle add MCP server panel"
                    >
                        <Plus className="mr-2 h-4 w-4" /> Add Server
                    </Button>
                </div>
            </div>

            {isAddOpen ? <AddServerForm onDone={() => { setIsAddOpen(false); void refreshDashboardQueries(); }} /> : null}
            {isImportOpen ? <BulkImportForm existingServerNames={existingServerNames} onDone={() => { setIsImportOpen(false); void refreshDashboardQueries(); }} /> : null}
            {editingServerUuid && editingServer ? (
                <EditMcpServer
                    server={editingServer}
                    onCancel={() => setEditingServerUuid(null)}
                    onSuccess={() => {
                        setEditingServerUuid(null);
                        void refreshDashboardQueries();
                    }}
                />
            ) : null}
            {inspectingServerUuid ? (
                <ServerInspectionPanel
                    server={inspectingServer}
                    runtime={inspectingRuntimeServer}
                    health={inspectingServerHealth}
                    onClose={() => setInspectingServerUuid(null)}
                />
            ) : null}

            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <Card className="bg-zinc-900 border-zinc-800">
                    <CardContent className="p-5">
                        <div className="flex items-center justify-between">
                            <div>
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Configured servers</div>
                                <div className="mt-1 text-3xl font-semibold text-white">{summary.serverCount}</div>
                            </div>
                            <Server className="h-5 w-5 text-blue-400" />
                        </div>
                    </CardContent>
                </Card>
                <Card className="bg-zinc-900 border-zinc-800">
                    <CardContent className="p-5">
                        <div className="flex items-center justify-between">
                            <div>
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Connected peers</div>
                                <div className="mt-1 text-3xl font-semibold text-white">{summary.connectedCount}</div>
                            </div>
                            <Network className="h-5 w-5 text-emerald-400" />
                        </div>
                    </CardContent>
                </Card>
                <Card className="bg-zinc-900 border-zinc-800">
                    <CardContent className="p-5">
                        <div className="flex items-center justify-between">
                            <div>
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Aggregated tools</div>
                                <div className="mt-1 text-3xl font-semibold text-white">{summary.toolCount}</div>
                            </div>
                            <Wrench className="h-5 w-5 text-purple-400" />
                        </div>
                    </CardContent>
                </Card>
                <Card className="bg-zinc-900 border-zinc-800">
                    <CardContent className="p-5">
                        <div className="flex items-center justify-between">
                            <div>
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Router status</div>
                                <div className="mt-1 text-3xl font-semibold text-white">{summary.initialized ? 'Ready' : 'Cold'}</div>
                            </div>
                            <Zap className="h-5 w-5 text-yellow-400" />
                        </div>
                    </CardContent>
                </Card>
            </div>

            <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)]">
                <Card className="min-w-0 overflow-hidden bg-zinc-900 border-zinc-800">
                    <CardHeader>
                        <CardTitle className="text-white">Why this page exists</CardTitle>
                    </CardHeader>
                    <CardContent className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                            <div className="flex items-center gap-2 text-sm font-semibold text-white"><Server className="h-4 w-4 text-blue-400" /> Aggregation</div>
                            <p className="mt-2 text-sm text-zinc-500">One TormentNexus endpoint should make many downstream MCP servers feel like a coherent control plane, not a pile of loose wires.</p>
                        </div>
                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                            <div className="flex items-center gap-2 text-sm font-semibold text-white"><Search className="h-4 w-4 text-cyan-400" /> Semantic grouping</div>
                            <p className="mt-2 text-sm text-zinc-500">Tool collisions and overlap should be handled through search, ranking, and working-set grouping instead of extra namespace ceremony.</p>
                        </div>
                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                            <div className="flex items-center gap-2 text-sm font-semibold text-white"><Activity className="h-4 w-4 text-emerald-400" /> Lifecycle supervision</div>
                            <p className="mt-2 text-sm text-zinc-500">A dead child server should restart without knocking over its healthy neighbors. Drama belongs in logs, not uptime.</p>
                        </div>
                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                            <div className="flex items-center gap-2 text-sm font-semibold text-white"><Search className="h-4 w-4 text-purple-400" /> Discoverability</div>
                            <p className="mt-2 text-sm text-zinc-500">Search, load, and working-set management are router features too; a giant static tool dump is not a UX strategy.</p>
                        </div>
                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                            <div className="flex items-center gap-2 text-sm font-semibold text-white"><Zap className="h-4 w-4 text-yellow-400" /> Traffic visibility</div>
                            <p className="mt-2 text-sm text-zinc-500">Operators should see message flow, latency, and failures clearly enough to debug the router without reading tea leaves.</p>
                        </div>
                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                            <div className="flex items-center gap-2 text-sm font-semibold text-white"><Layers className="h-4 w-4 text-indigo-400" /> Client sync</div>
                            <p className="mt-2 text-sm text-zinc-500">The system should push usable config into clients like Claude Desktop, Cursor, and VS Code rather than making setup a scavenger hunt.</p>
                        </div>
                    </CardContent>
                </Card>

                <Card className="min-w-0 overflow-hidden bg-zinc-900 border-zinc-800">
                    <CardHeader>
                        <CardTitle className="text-white">Control-plane quick links</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3">
                        <QuickLinkCard title="Tool Search" description="Find tools semantically and shape the active working set without extra routing layers." href="/dashboard/mcp/search" accentClass="text-cyan-400" icon={Search} />
                        <QuickLinkCard title="Observability" description="Watch metrics, health, and live router state." href="/dashboard/mcp/observability" accentClass="text-yellow-400" icon={Zap} />
                        <QuickLinkCard title="Policies" description="Enforce tool access and governance rules." href="/dashboard/mcp/policies" accentClass="text-green-400" icon={Shield} />
                        <QuickLinkCard title="Testing Lab" description="Use inspector/search/playground surfaces without polluting the main control-plane view." href="/dashboard/mcp/testing" accentClass="text-fuchsia-400" icon={Wrench} />
                    </CardContent>
                </Card>
            </div>

            <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)]">
                <Card className="min-w-0 overflow-hidden bg-zinc-900 border-zinc-800">
                    <CardHeader className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                        <div>
                            <CardTitle className="text-white">Downstream runtime</CardTitle>
                            <p className="mt-1 text-sm text-zinc-500">
                                Managed-server discovery is now batch-operable too, so recovering a stale fleet no longer means click-click-click until your mouse files a complaint.
                            </p>
                        </div>
                        <div className="flex flex-wrap gap-2">
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                disabled={bulkActionsDisabled || unresolvedActionableCount === 0}
                                onClick={() => void handleBulkBinaryRefresh('unresolved')}
                                title={unresolvedActionableCount > 0
                                    ? `Repair ${unresolvedActionableCount} managed MCP server${unresolvedActionableCount === 1 ? '' : 's'} with unresolved metadata or stale ready zero-tool caches`
                                    : 'No unresolved or stale managed MCP servers are currently actionable for binary rediscovery'}
                                aria-label="Repair unresolved or stale MCP server discovery"
                                className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                            >
                                {bulkRefreshState?.mode === 'unresolved' ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <AlertTriangle className="mr-2 h-3.5 w-3.5" />}
                                Repair stale / unresolved {unresolvedActionableCount > 0 ? `(${unresolvedActionableCount})` : ''}
                            </Button>
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                disabled={bulkActionsDisabled || allActionableCount === 0}
                                onClick={() => void handleBulkBinaryRefresh('all')}
                                title={allActionableCount > 0
                                    ? `Refresh binary metadata across ${allActionableCount} managed MCP server${allActionableCount === 1 ? '' : 's'}`
                                    : 'No managed MCP servers are currently actionable for binary rediscovery'}
                                aria-label="Refresh binary metadata for all managed MCP servers"
                                className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                            >
                                {bulkRefreshState?.mode === 'all' ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="mr-2 h-3.5 w-3.5" />}
                                Refresh all binaries {allActionableCount > 0 ? `(${allActionableCount})` : ''}
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-3">
                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <div className="text-sm font-medium text-white">Fleet discovery summary</div>
                                    <div className="mt-2 flex flex-wrap gap-2 text-[11px] uppercase tracking-wider text-zinc-400">
                                        <span className="rounded border border-zinc-800 bg-zinc-900/60 px-2 py-1">Managed {discoverySummary.totalCount}</span>
                                        <span className="rounded border border-emerald-500/20 bg-emerald-500/10 px-2 py-1 text-emerald-200">Ready {discoverySummary.readyCount}</span>
                                        <span className="rounded border border-amber-500/20 bg-amber-500/10 px-2 py-1 text-amber-200">Unresolved {discoverySummary.unresolvedCount}</span>
                                        <span className="rounded border border-rose-500/20 bg-rose-500/10 px-2 py-1 text-rose-200">Stale ready {discoverySummary.staleReadyCount}</span>
                                        <span className="rounded border border-zinc-700 bg-zinc-900/60 px-2 py-1">Never loaded {discoverySummary.neverLoadedCount}</span>
                                        <span className="rounded border border-cyan-500/20 bg-cyan-500/10 px-2 py-1 text-cyan-200">Repairable {discoverySummary.repairableCount}</span>
                                        {localCompatActive ? (
                                            <span className="rounded border border-sky-500/20 bg-sky-500/10 px-2 py-1 text-sky-200">Local compat {discoverySummary.localCompatCount}</span>
                                        ) : null}
                                    </div>
                                </div>
                                <div className="max-w-md text-xs text-zinc-500 lg:text-right">
                                    {bulkRefreshState ? (
                                        <span className="text-zinc-300">
                                            Running {bulkRefreshState.mode === 'all' ? 'full fleet' : 'unresolved'} binary discovery: {bulkRefreshState.completedCount} / {bulkRefreshState.totalCount}
                                        </span>
                                    ) : localCompatActive ? (
                                        <span>
                                            Local compat fallback is active for {discoverySummary.localCompatCount} managed server{discoverySummary.localCompatCount === 1 ? '' : 's'}, so TormentNexus is surfacing config-backed records with stable local IDs and action links while live core telemetry is unavailable.
                                        </span>
                                    ) : discoverySummary.staleReadyCount > 0 ? (
                                        <span>
                                            {discoverySummary.staleReadyCount} server{discoverySummary.staleReadyCount === 1 ? '' : 's'} look <span className="font-semibold text-white">ready</span> but still have zero cached tools. Use <span className="font-semibold text-white">Repair stale / unresolved</span> to force fresh binary discovery and scrub the zombie cache state.
                                        </span>
                                    ) : unresolvedWithoutActionCount > 0 ? (
                                        <span>
                                            {unresolvedWithoutActionCount} unresolved server{unresolvedWithoutActionCount === 1 ? '' : 's'} are missing actionable card links in this view. The bulk buttons currently target {allActionableCount} managed server{allActionableCount === 1 ? '' : 's'} with stable identifiers.
                                        </span>
                                    ) : allActionableCount === 0 ? (
                                        <span>
                                            No managed servers are currently actionable for bulk rediscovery yet.
                                        </span>
                                    ) : (
                                        <span>
                                            Use <span className="font-semibold text-white">Retry unresolved</span> for failed or pending metadata, or <span className="font-semibold text-white">Refresh all binaries</span> after a broad config/tooling change.
                                        </span>
                                    )}
                                </div>
                            </div>
                        </div>
                        {isLoadingServers ? (
                            <div className="flex justify-center p-8"><Loader2 className="h-6 w-6 animate-spin text-zinc-500" /></div>
                        ) : serverList.length > 0 ? (
                            serverList.map((server, index) => {
                                const actionLinks = buildServerToolActionLinks(server.name);
                                const serverUuid = server.uuid;
                                const isLocalCompatServer = isLocalCompatMetadataSource(server.metadataSource);
                                const hasStaleReadyCache = hasStaleReadyMetadata({
                                    name: server.name,
                                    _meta: {
                                        status: server.metadataStatus,
                                        metadataSource: server.metadataSource,
                                        toolCount: server.metadataToolCount,
                                        lastSuccessfulBinaryLoadAt: server.lastSuccessfulBinaryLoadAt,
                                    },
                                });

                                return (
                                <div
                                    key={serverUuid ?? `${server.name}-${server.config?.command ?? 'na'}-${(server.config?.args ?? []).join(' ')}-${index}`}
                                    className="min-w-0 rounded-lg border border-zinc-800 bg-zinc-950/60 p-4"
                                >
                                    <div className="flex items-start justify-between gap-4">
                                        <div className="min-w-0">
                                            <div className="font-medium text-white">{server.name}</div>
                                            <div className="mt-1 text-xs text-zinc-500 font-mono break-all">
                                                {server.config?.command || 'n/a'} {(server.config?.args || []).join(' ')}
                                            </div>
                                            <div className="mt-2 flex flex-wrap gap-2 text-[10px] uppercase tracking-wider text-zinc-400">
                                                <span className="rounded border border-zinc-800 bg-zinc-900/60 px-2 py-1">
                                                    cache {server.metadataStatus ?? 'pending'}
                                                </span>
                                                <span className={`rounded border px-2 py-1 ${server.runtimeConnected ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-200' : 'border-zinc-800 bg-zinc-900/60'}`}>
                                                    runtime {server.runtimeState ?? server.status}
                                                </span>
                                                <span className={`rounded border px-2 py-1 ${server.warmupState === 'ready' ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-200' : server.warmupState === 'failed' ? 'border-rose-500/20 bg-rose-500/10 text-rose-200' : server.warmupState === 'warming' || server.warmupState === 'scheduled' ? 'border-amber-500/20 bg-amber-500/10 text-amber-200' : 'border-zinc-800 bg-zinc-900/60'}`}>
                                                    warmup {server.warmupState ?? 'idle'}
                                                </span>
                                                {server.always_on ? (
                                                    <span className="rounded border border-indigo-500/20 bg-indigo-500/10 px-2 py-1 text-indigo-200 font-bold">
                                                        Always On
                                                    </span>
                                                ) : null}
                                                <span className="rounded border border-zinc-800 bg-zinc-900/60 px-2 py-1">
                                                    source {server.metadataSource ?? 'none'}
                                                </span>
                                                {hasStaleReadyCache ? (
                                                    <span className="rounded border border-rose-500/20 bg-rose-500/10 px-2 py-1 text-rose-200">
                                                        stale ready cache
                                                    </span>
                                                ) : null}
                                                {isLocalCompatServer ? (
                                                    <span className="rounded border border-sky-500/20 bg-sky-500/10 px-2 py-1 text-sky-200">
                                                        local compat actions enabled
                                                    </span>
                                                ) : null}
                                            </div>
                                        </div>
                                        <StatusBadge status={server.status} />
                                    </div>
                                    <div className="mt-3 grid grid-cols-2 gap-3 text-sm">
                                        <div className="rounded border border-zinc-800 bg-zinc-900/60 p-2.5">
                                            <div className="text-xs uppercase tracking-wider text-zinc-500">Tools</div>
                                            <div className="mt-1 text-white font-semibold">{server.toolCount}</div>
                                        </div>
                                        <div className="rounded border border-zinc-800 bg-zinc-900/60 p-2.5">
                                            <div className="text-xs uppercase tracking-wider text-zinc-500">Advertised tools</div>
                                            <div className="mt-1 text-white font-semibold">{server.advertisedToolCount ?? server.metadataToolCount ?? 0}</div>
                                        </div>
                                    </div>
                                    <div className="mt-3 space-y-3">
                                        <div className="rounded border border-zinc-800 bg-zinc-900/60 p-2.5 text-xs text-zinc-400">
                                            <div>Cached tools: <span className="font-semibold text-white">{server.metadataToolCount ?? 0}</span></div>
                                            <div className="mt-1">Advertised source: <span className="font-semibold text-white">{server.advertisedSource ?? 'unknown'}</span></div>
                                            <div className="mt-1">Last runtime connect: <span className="font-semibold text-white">{server.lastConnectedAt ?? 'never'}</span></div>
                                            <div className="mt-1">Env keys: <span className="font-semibold text-white">{server.config?.env?.length ?? 0}</span></div>
                                            <div className="mt-1 break-all">Last binary load: {server.lastSuccessfulBinaryLoadAt ?? 'never'}</div>
                                        </div>
                                        {hasStaleReadyCache ? (
                                            <div className="rounded border border-rose-500/20 bg-rose-500/10 p-3 text-xs text-rose-100">
                                                <div className="flex items-center gap-2 font-semibold text-white">
                                                    <AlertTriangle className="h-3.5 w-3.5 text-rose-300" />
                                                    Ready cache looks stale
                                                </div>
                                                <p className="mt-1 text-rose-100/90">
                                                    This server is marked ready, but TormentNexus has zero cached tools for it. That usually means an older discovery failure got cached as success. Run a binary refresh to repair it.
                                                </p>
                                            </div>
                                        ) : null}
                                        {serverUuid ? (
                                            <div className="min-w-0 rounded border border-zinc-800 bg-zinc-900/60 p-3">
                                                <div className={`rounded-md px-3 py-2 ${isLocalCompatServer ? 'border border-sky-500/20 bg-sky-500/5' : 'border border-cyan-500/20 bg-cyan-500/5'}`}>
                                                    <div className="text-[11px] uppercase tracking-[0.24em] text-cyan-300">Server actions live here</div>
                                                    <p className="mt-1 text-xs text-zinc-400">
                                                        {isLocalCompatServer
                                                            ? 'This server is being surfaced through local compat fallback, so these controls act on the TormentNexus-managed local config record while upstream core telemetry is unavailable.'
                                                            : 'Keep the operator controls anchored on every server card so inspection, edits, cache warm-up, health tests, and logs stay one click away.'}
                                                    </p>
                                                </div>
                                                <div className="mt-3 text-[11px] uppercase tracking-wider text-zinc-500">Primary actions</div>
                                                <p className="mt-1 text-xs text-zinc-500">
                                                    The controls below stay visible on each server card so inspection, editing, cache init, testing, and logs are easy to reach even on narrow layouts.
                                                </p>
                                                <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2 2xl:grid-cols-3">
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        onClick={() => setInspectingServerUuid(serverUuid)}
                                                        title={`Inspect ${server.name} configuration and metadata`}
                                                        aria-label={`Inspect ${server.name}`}
                                                        className="justify-center border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                                    >
                                                        <Eye className="mr-2 h-3.5 w-3.5" />
                                                        Inspect
                                                    </Button>
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        onClick={() => setEditingServerUuid(serverUuid)}
                                                        title={`Edit ${server.name}`}
                                                        aria-label={`Edit ${server.name}`}
                                                        className="justify-center border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                                    >
                                                        <Pencil className="mr-2 h-3.5 w-3.5" />
                                                        Edit
                                                    </Button>
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        disabled={reloadMetadataMutation.isPending}
                                                        onClick={() => void handleReloadMetadata({ uuid: serverUuid, mode: 'binary' })}
                                                        title={hasStaleReadyCache
                                                            ? `Repair stale ready cache for ${server.name} by relaunching the binary and rediscovering tools`
                                                            : `Refresh metadata by launching ${server.name} binary and rediscovering tools`}
                                                        aria-label={hasStaleReadyCache
                                                            ? `Repair stale cache for ${server.name}`
                                                            : `Refresh binary metadata for ${server.name}`}
                                                        className={`justify-center ${hasStaleReadyCache ? 'border-rose-500/30 text-rose-200 hover:bg-rose-500/10' : 'border-zinc-700 text-zinc-300 hover:bg-zinc-800'}`}
                                                    >
                                                        {reloadMetadataMutation.isPending && reloadMetadataMutation.variables?.uuid === serverUuid ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="mr-2 h-3.5 w-3.5" />}
                                                        {hasStaleReadyCache ? 'Repair cache' : 'Refresh binary'}
                                                    </Button>
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        disabled={reloadMetadataMutation.isPending}
                                                        onClick={() => void handleLoadTestAndCacheServer(serverUuid, server.name)}
                                                        title={`Start ${server.name}, test its health, and initialize cached tools in one pass`}
                                                        aria-label={`Initialize cache for ${server.name}`}
                                                        className="justify-center border-emerald-500/30 text-emerald-200 hover:bg-emerald-500/10"
                                                    >
                                                        {reloadMetadataMutation.isPending && reloadMetadataMutation.variables?.uuid === serverUuid ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <Play className="mr-2 h-3.5 w-3.5" />}
                                                        Init cache
                                                    </Button>
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        onClick={() => void handleTestServer(serverUuid, server.name)}
                                                        title={`Run a health check for ${server.name}`}
                                                        aria-label={`Test ${server.name}`}
                                                        className="justify-center border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                                    >
                                                        <HeartPulse className="mr-2 h-3.5 w-3.5" />
                                                        Test
                                                    </Button>
                                                    <Link
                                                        href={actionLinks.logsHref}
                                                        title={`Open live logs while you test ${server.name}`}
                                                        aria-label={`Open logs for ${server.name}`}
                                                        className="inline-flex items-center justify-center rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-300 transition-colors hover:bg-zinc-800"
                                                    >
                                                        <Activity className="mr-2 h-3.5 w-3.5" />
                                                        Logs
                                                    </Link>
                                                        <Link
                                                            href={`/dashboard/mcp/testing/servers?target=${encodeURIComponent(server.name)}`}
                                                            title={`Open the interactive probe panel for ${server.name}`}
                                                            aria-label={`Open interactive test for ${server.name}`}
                                                            className="inline-flex items-center justify-center rounded-md border border-cyan-500/30 px-3 py-2 text-sm text-cyan-200 transition-colors hover:bg-cyan-500/10"
                                                        >
                                                            <Play className="mr-2 h-3.5 w-3.5" />
                                                            Interactive test
                                                        </Link>
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        disabled={reloadMetadataMutation.isPending}
                                                        onClick={() => void handleReloadMetadata({ uuid: serverUuid, mode: 'cache' })}
                                                        title={`Refresh cached metadata snapshot for ${server.name}`}
                                                        aria-label={`Refresh cache for ${server.name}`}
                                                        className="justify-center border-cyan-500/30 text-cyan-200 hover:bg-cyan-500/10"
                                                    >
                                                        {reloadMetadataMutation.isPending && reloadMetadataMutation.variables?.uuid === serverUuid ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="mr-2 h-3.5 w-3.5" />}
                                                        Refresh cache
                                                    </Button>
                                                </div>
                                                <div className="mt-3 border-t border-zinc-800 pt-3">
                                                    <div className="text-[11px] uppercase tracking-wider text-zinc-500">Secondary actions</div>
                                                    <div className="flex flex-wrap gap-2">
                                                        <Link
                                                            href={actionLinks.inspectToolsHref}
                                                            title={`Inspect tools discovered from ${server.name}`}
                                                            aria-label={`Inspect tools for ${server.name}`}
                                                            className="inline-flex items-center rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-300 transition-colors hover:bg-zinc-800"
                                                        >
                                                            <Wrench className="mr-2 h-3.5 w-3.5" />
                                                            Inspect tools
                                                        </Link>
                                                        <Link
                                                            href={actionLinks.editToolsHref}
                                                            title={`Edit working-set behavior for tools from ${server.name}`}
                                                            aria-label={`Edit tools for ${server.name}`}
                                                            className="inline-flex items-center rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-300 transition-colors hover:bg-zinc-800"
                                                        >
                                                            <Pencil className="mr-2 h-3.5 w-3.5" />
                                                            Edit tools
                                                        </Link>
                                                        <Button
                                                            type="button"
                                                            variant="outline"
                                                            size="sm"
                                                            disabled={updateServerMutation.isPending}
                                                            onClick={() => void handleToggleAlwaysOn(serverUuid, server.name, !!server.always_on)}
                                                            title={`Toggle Always On status for ${server.name}`}
                                                            aria-label={`Toggle Always On for ${server.name}`}
                                                            className={`border-zinc-700 hover:bg-zinc-800 ${server.always_on ? 'text-indigo-400 border-indigo-500/30 bg-indigo-500/5' : 'text-zinc-300'}`}
                                                        >
                                                            <Zap className="mr-2 h-3.5 w-3.5" />
                                                            Toggle Auto-Load Tools: {server.always_on ? 'ON' : 'OFF'}
                                                        </Button>
                                                        <Button
                                                            type="button"
                                                            variant="outline"
                                                            size="sm"
                                                            onClick={() => void handleResetHealth(serverUuid, server.name)}
                                                            disabled={resetServerHealthMutation.isPending}
                                                            title={`Reset tracked health state for ${server.name}`}
                                                            aria-label={`Reset health for ${server.name}`}
                                                            className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                                        >
                                                            {resetServerHealthMutation.isPending && resettingServerUuid === serverUuid ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <HeartPulse className="mr-2 h-3.5 w-3.5" />}
                                                            Reset health
                                                        </Button>
                                                        <Button
                                                            type="button"
                                                            variant="outline"
                                                            size="sm"
                                                            disabled={clearMetadataCacheMutation.isPending || reloadMetadataMutation.isPending}
                                                            onClick={() => void handleClearMetadataCache(serverUuid)}
                                                            title={`Clear cached metadata for ${server.name}`}
                                                            aria-label={`Clear metadata cache for ${server.name}`}
                                                            className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                                        >
                                                            {clearMetadataCacheMutation.isPending && clearMetadataCacheMutation.variables?.uuid === serverUuid ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <Database className="mr-2 h-3.5 w-3.5" />}
                                                            Clear cache
                                                        </Button>
                                                        <Button
                                                            type="button"
                                                            variant="outline"
                                                            size="sm"
                                                            disabled={reloadMetadataMutation.isPending}
                                                            onClick={() => void handleReloadMetadata({ uuid: serverUuid, mode: 'binary' })}
                                                            title={`Force a fresh binary rediscovery for ${server.name}`}
                                                            aria-label={`Force binary rediscovery for ${server.name}`}
                                                            className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                                        >
                                                            {reloadMetadataMutation.isPending && reloadMetadataMutation.variables?.uuid === serverUuid ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="mr-2 h-3.5 w-3.5" />}
                                                            Force rediscovery
                                                        </Button>
                                                        <Button
                                                            type="button"
                                                            variant="outline"
                                                            size="sm"
                                                            onClick={() => void handleDeleteServer(serverUuid, server.name)}
                                                            disabled={deleteServerMutation.isPending}
                                                            title={`Delete ${server.name} from TormentNexus configuration`}
                                                            aria-label={`Delete ${server.name}`}
                                                            className="border-red-500/30 text-red-200 hover:bg-red-500/10"
                                                        >
                                                            {deleteServerMutation.isPending && deletingServerUuid === serverUuid ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <Trash2 className="mr-2 h-3.5 w-3.5" />}
                                                            Delete
                                                        </Button>
                                                    </div>
                                                </div>
                                            </div>
                                        ) : null}
                                    </div>
                                </div>
                                );
                            })
                        ) : (
                            <div className="rounded-lg border border-dashed border-zinc-800 p-8 text-center text-zinc-500">
                                <div className="space-y-3">
                                    <div>No downstream servers are configured yet.</div>
                                    <p className="mx-auto max-w-xl text-sm text-zinc-500">
                                        Once a server is available, each server card exposes <span className="font-medium text-zinc-300">Inspect</span>, <span className="font-medium text-zinc-300">Edit</span>, <span className="font-medium text-zinc-300">Init cache</span>, <span className="font-medium text-zinc-300">Test</span>, <span className="font-medium text-zinc-300">Refresh cache</span>, and <span className="font-medium text-zinc-300">Logs</span> as primary actions.
                                    </p>
                                    <div className="flex flex-wrap items-center justify-center gap-2">
                                        <Button
                                            type="button"
                                            variant="outline"
                                            onClick={() => setIsAddOpen(true)}
                                            className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                        >
                                            <Plus className="mr-2 h-4 w-4" />
                                            Add Server
                                        </Button>
                                    </div>
                                </div>
                            </div>
                        )}
                    </CardContent>
                </Card>

                <Card className="min-w-0 overflow-hidden bg-zinc-900 border-zinc-800">
                    <CardHeader className="flex flex-row items-start justify-between gap-3">
                        <div>
                            <CardTitle className="text-white">Tool discovery snapshot</CardTitle>
                            <p className="text-sm text-zinc-500 mt-1">A compact view of the aggregated catalog. Use Search for deeper discovery and working-set actions.</p>
                        </div>
                        <Link
                            href="/dashboard/mcp/search"
                            title="Open semantic search and working-set management for MCP tools"
                            aria-label="Open MCP search dashboard"
                            className="inline-flex items-center gap-1 text-xs text-zinc-400 hover:text-white"
                        >
                            Open Search
                            <ExternalLink className="h-3.5 w-3.5" />
                        </Link>
                    </CardHeader>
                    <CardContent className="space-y-3">
                        {isLoadingTools ? (
                            <div className="flex justify-center p-8"><Loader2 className="h-6 w-6 animate-spin text-zinc-500" /></div>
                        ) : topTools.length > 0 ? (
                            topTools.map((tool) => (
                                <div key={`${tool.server}:${tool.name}`} className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4">
                                    <div className="flex items-center gap-2">
                                        <div className="font-mono text-sm text-blue-400 break-all">{tool.name}</div>
                                        <span className="rounded bg-zinc-800 px-2 py-0.5 text-[10px] uppercase tracking-wider text-zinc-400">{tool.server}</span>
                                    </div>
                                    <p className="mt-2 text-sm text-zinc-500">{tool.description || 'No description available.'}</p>
                                </div>
                            ))
                        ) : (
                            <div className="rounded-lg border border-dashed border-zinc-800 p-8 text-center text-zinc-500">
                                No aggregated tools available yet.
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
