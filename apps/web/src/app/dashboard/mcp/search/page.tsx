"use client";

import Link from 'next/link';
import { Suspense, useEffect, useState } from 'react';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { Button, Card, CardContent, CardHeader, CardTitle } from "@tormentnexus/ui";
import { Loader2, Search, Zap, Code, Layers, ExternalLink, Activity, Database, ArrowDownToLine, Sparkles, Trash2, SlidersHorizontal, History } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

type SearchResult = {
    name: string;
    description: string;
    server: string;
    serverDisplayName?: string;
    serverTags?: string[];
    toolTags?: string[];
    semanticGroup?: string;
    semanticGroupLabel?: string;
    advertisedName?: string;
    keywords?: string[];
    alwaysOn?: boolean;
    originalName?: string | null;
    loaded?: boolean;
    hydrated?: boolean;
    deferred?: boolean;
    requiresSchemaHydration?: boolean;
    matchReason?: string;
    score?: number;
    rank?: number;
    important?: boolean;
    alwaysShow?: boolean;
    alwaysLoaded?: boolean;
    inputSchema: Record<string, unknown> | null;
};

type WorkingSetTool = {
    name: string;
    hydrated: boolean;
    lastLoadedAt: number;
    lastHydratedAt: number | null;
};

type ToolSearchProfile = 'web-research' | 'repo-coding' | 'browser-automation' | 'local-ops' | 'database';

type ToolSelectionTelemetryEvent = {
    id: string;
    type: 'search' | 'load' | 'hydrate' | 'unload';
    timestamp: number;
    query?: string;
    profile?: string;
    source?: 'runtime-search' | 'cached-ranking' | 'live-aggregator';
    resultCount?: number;
    topResultName?: string;
    topMatchReason?: string;
    topScore?: number;
    secondResultName?: string;
    secondMatchReason?: string;
    secondScore?: number;
    scoreGap?: number;
    toolName?: string;
    status: 'success' | 'error';
    message?: string;
    evictedTools?: string[];
    latencyMs?: number;
    autoLoadReason?: string;
    autoLoadConfidence?: number;
    autoLoadEvaluated?: boolean;
    autoLoadOutcome?: 'loaded' | 'skipped' | 'not-applicable';
    autoLoadSkipReason?: string;
    autoLoadMinConfidence?: number;
    autoLoadExecutionStatus?: 'success' | 'error' | 'not-attempted';
    autoLoadExecutionError?: string;
};

type WorkingSetEvictionEvent = {
    toolName: string;
    timestamp: number;
    tier: 'loaded' | 'hydrated';
};

type ToolPreferences = {
    importantTools: string[];
    alwaysLoadedTools: string[];
    autoLoadMinConfidence: number;
    maxLoadedTools: number;
    maxHydratedSchemas: number;
};

type ToolPreferenceMutationInput = {
    importantTools?: string[];
    alwaysLoadedTools?: string[];
    autoLoadMinConfidence?: number;
    maxLoadedTools?: number;
    maxHydratedSchemas?: number;
};

type TelemetryWindowPreset = 'all' | '5m' | '15m' | '1h' | '24h';
type TelemetrySourceFilter = 'all' | 'runtime-search' | 'cached-ranking' | 'live-aggregator';
type TelemetryTriagePreset = 'errors-now' | 'runtime-failures' | 'load-incidents' | 'hydration-failures' | 'live-aggregator-focus';

const TELEMETRY_FILTERS_STORAGE_KEY = 'tormentnexus.mcp.search.telemetryFilters.v1';
const TELEMETRY_TYPE_QUERY_KEY = 'telemetryType';
const TELEMETRY_STATUS_QUERY_KEY = 'telemetryStatus';
const TELEMETRY_WINDOW_QUERY_KEY = 'telemetryWindow';
const TELEMETRY_SOURCE_QUERY_KEY = 'telemetrySource';

type TelemetryTrendBucket = {
    start: number;
    end: number;
    label: string;
};

function resolveTelemetryWindowStart(windowPreset: TelemetryWindowPreset): number | null {
    const now = Date.now();

    if (windowPreset === '5m') {
        return now - (5 * 60 * 1000);
    }

    if (windowPreset === '15m') {
        return now - (15 * 60 * 1000);
    }

    if (windowPreset === '1h') {
        return now - (60 * 60 * 1000);
    }

    if (windowPreset === '24h') {
        return now - (24 * 60 * 60 * 1000);
    }

    return null;
}

function buildTelemetryTrendBuckets(options: {
    windowPreset: TelemetryWindowPreset;
    windowStart: number | null;
    events: ToolSelectionTelemetryEvent[];
}): TelemetryTrendBucket[] {
    if (options.events.length === 0) {
        return [];
    }

    const now = Date.now();
    const earliestEventTimestamp = Math.min(...options.events.map((event) => event.timestamp));
    const computedStart = options.windowStart ?? earliestEventTimestamp;
    const start = Math.min(computedStart, now - 1000);

    const targetBucketCount = options.windowPreset === 'all'
        ? 6
        : options.windowPreset === '24h'
            ? 6
            : 5;
    const totalWindowMs = Math.max(60_000, now - start);
    const bucketSizeMs = Math.max(1, Math.ceil(totalWindowMs / targetBucketCount));

    return Array.from({ length: targetBucketCount }, (_value, index) => {
        const bucketStart = start + (index * bucketSizeMs);
        const bucketEnd = index === targetBucketCount - 1 ? now : Math.min(now, bucketStart + bucketSizeMs);
        const label = new Date(bucketEnd).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

        return {
            start: bucketStart,
            end: bucketEnd,
            label,
        };
    });
}

function formatRelativeTimestamp(timestamp: number | null): string {
    if (!timestamp) {
        return '—';
    }

    const deltaSeconds = Math.max(0, Math.round((Date.now() - timestamp) / 1000));
    if (deltaSeconds < 60) {
        return `${deltaSeconds}s ago`;
    }

    const deltaMinutes = Math.round(deltaSeconds / 60);
    if (deltaMinutes < 60) {
        return `${deltaMinutes}m ago`;
    }

    const deltaHours = Math.round(deltaMinutes / 60);
    return `${deltaHours}h ago`;
}

export default function SearchDashboardPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center h-screen"><Loader2 className="h-8 w-8 animate-spin text-zinc-400" /></div>}>
            <SearchDashboard />
        </Suspense>
    );
}

function SearchDashboard() {
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();
    const [query, setQuery] = useState('');
    const [profile, setProfile] = useState<ToolSearchProfile | 'default'>('default');
    const [autoLoadMinConfidenceDraft, setAutoLoadMinConfidenceDraft] = useState(0.85);
    const [maxLoadedToolsDraft, setMaxLoadedToolsDraft] = useState(16);
    const [maxHydratedSchemasDraft, setMaxHydratedSchemasDraft] = useState(8);
    const [jsoncDraft, setJsoncDraft] = useState('');
    const [telemetryTypeFilter, setTelemetryTypeFilter] = useState<'all' | ToolSelectionTelemetryEvent['type']>('all');
    const [telemetryStatusFilter, setTelemetryStatusFilter] = useState<'all' | ToolSelectionTelemetryEvent['status']>('all');
    const [telemetryWindowFilter, setTelemetryWindowFilter] = useState<TelemetryWindowPreset>('15m');
    const [telemetrySourceFilter, setTelemetrySourceFilter] = useState<TelemetrySourceFilter>('all');
    const utils = trpc.useUtils();
    const searchQuery = trpc.mcp.searchTools.useQuery(
        { query, profile: profile === 'default' ? undefined : profile },
        { enabled: query.trim().length > 0 },
    );
    const workingSetQuery = trpc.mcp.getWorkingSet.useQuery(undefined, { refetchInterval: 4000 });
    const evictionHistoryQuery = trpc.mcp.getWorkingSetEvictionHistory.useQuery(undefined, { refetchInterval: 8000 });
    const telemetryQuery = trpc.mcp.getToolSelectionTelemetry.useQuery(undefined, { refetchInterval: 4000 });
    const preferencesQuery = trpc.mcp.getToolPreferences.useQuery();
    const jsoncEditorQuery = trpc.mcp.getJsoncEditor.useQuery();
    const clearTelemetryMutation = trpc.mcp.clearToolSelectionTelemetry.useMutation({
        onSuccess: async () => {
            toast.success('Telemetry history cleared');
            await telemetryQuery.refetch();
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const clearEvictionHistoryMutation = trpc.mcp.clearWorkingSetEvictionHistory.useMutation({
        onSuccess: async (data) => {
            toast.success(data?.message || 'Eviction history cleared');
            await evictionHistoryQuery.refetch();
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const loadMutation = trpc.mcp.loadTool.useMutation({
        onSuccess: async (data) => {
            toast.success(data.message || 'Tool loaded');
            await Promise.all([
                utils.mcp.getWorkingSet.invalidate(),
                utils.mcp.searchTools.invalidate(),
            ]);
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const unloadMutation = trpc.mcp.unloadTool.useMutation({
        onSuccess: async (data) => {
            toast.success(data.message || 'Tool unloaded');
            await Promise.all([
                utils.mcp.getWorkingSet.invalidate(),
                utils.mcp.searchTools.invalidate(),
                utils.mcp.getToolSelectionTelemetry.invalidate(),
            ]);
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const hydrateMutation = trpc.mcp.getToolSchema.useMutation({
        onSuccess: async (_data, variables) => {
            const toolName = (variables as { name?: string } | undefined)?.name ?? 'tool';
            toast.success(`Schema hydrated for ${toolName}`);
            await Promise.all([
                utils.mcp.getWorkingSet.invalidate(),
                utils.mcp.searchTools.invalidate(),
                utils.mcp.getToolSelectionTelemetry.invalidate(),
            ]);
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const setPreferencesMutation = trpc.mcp.setToolPreferences.useMutation({
        onSuccess: async () => {
            toast.success('Important tools updated');
            await Promise.all([
                utils.mcp.getToolPreferences.invalidate(),
                utils.mcp.searchTools.invalidate(),
            ]);
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const saveJsoncMutation = trpc.mcp.saveJsoncEditor.useMutation({
        onSuccess: async () => {
            toast.success('mcp.jsonc saved');
            await Promise.all([
                utils.mcp.getJsoncEditor.invalidate(),
                utils.mcp.getToolPreferences.invalidate(),
                utils.mcp.searchTools.invalidate(),
            ]);
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const callToolMutation = trpc.mcp.callTool.useMutation({
        onSuccess: () => {
            toast.success('Tool invoked');
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const results = (searchQuery.data || []) as SearchResult[];
    const isLoading = searchQuery.isLoading;
    const workingSet = ((workingSetQuery.data?.tools as WorkingSetTool[] | undefined) ?? []);
    const allToolsQuery = trpc.mcp.listTools.useQuery(undefined, { refetchInterval: 15000 });
    const allKnownTools = (allToolsQuery.data as SearchResult[] | undefined) ?? [];
    const telemetryEvents = ((telemetryQuery.data as ToolSelectionTelemetryEvent[] | undefined) ?? []);
    const telemetryWindowStart = resolveTelemetryWindowStart(telemetryWindowFilter);
    const telemetryEventsPreStatusFilter = telemetryEvents
        .filter((event) => telemetryWindowStart == null || event.timestamp >= telemetryWindowStart)
        .filter((event) => telemetryTypeFilter === 'all' || event.type === telemetryTypeFilter)
        .filter((event) => telemetrySourceFilter === 'all' || event.source === telemetrySourceFilter);
    const filteredTelemetryEvents = telemetryEventsPreStatusFilter
        .filter((event) => telemetryStatusFilter === 'all' || event.status === telemetryStatusFilter);
    const telemetry = filteredTelemetryEvents.slice(0, 12);
    const telemetryFiltersAtDefault = telemetryTypeFilter === 'all'
        && telemetryStatusFilter === 'all'
        && telemetryWindowFilter === '15m'
        && telemetrySourceFilter === 'all';
    const telemetrySummary = {
        total: filteredTelemetryEvents.length,
        success: filteredTelemetryEvents.filter((event) => event.status === 'success').length,
        error: filteredTelemetryEvents.filter((event) => event.status === 'error').length,
    };
    const telemetryTrendBuckets = buildTelemetryTrendBuckets({
        windowPreset: telemetryWindowFilter,
        windowStart: telemetryWindowStart,
        events: telemetryEventsPreStatusFilter,
    });
    const telemetryStatusTrend = telemetryTrendBuckets.map((bucket) => {
        const bucketEvents = telemetryEventsPreStatusFilter.filter((event) => event.timestamp >= bucket.start && event.timestamp < bucket.end);
        const successCount = bucketEvents.filter((event) => event.status === 'success').length;
        const errorCount = bucketEvents.filter((event) => event.status === 'error').length;

        return {
            label: bucket.label,
            total: bucketEvents.length,
            successCount,
            errorCount,
        };
    });
    const telemetrySourceStats = (['runtime-search', 'cached-ranking', 'live-aggregator'] as const)
        .map((source) => {
            const sourceEvents = filteredTelemetryEvents.filter((event) => event.source === source);
            const avgLatencyMs = sourceEvents.length > 0
                ? Math.round(sourceEvents.reduce((sum, event) => sum + (event.latencyMs ?? 0), 0) / sourceEvents.length)
                : 0;

            return {
                source,
                count: sourceEvents.length,
                success: sourceEvents.filter((event) => event.status === 'success').length,
                error: sourceEvents.filter((event) => event.status === 'error').length,
                avgLatencyMs,
                trend: telemetryTrendBuckets.map((bucket) => {
                    const bucketEvents = sourceEvents.filter((event) => event.timestamp >= bucket.start && event.timestamp < bucket.end);
                    const bucketErrors = bucketEvents.filter((event) => event.status === 'error').length;

                    return {
                        label: bucket.label,
                        count: bucketEvents.length,
                        errorCount: bucketErrors,
                    };
                }),
            };
        });
    const maxTelemetrySourceCount = telemetrySourceStats.reduce((max, item) => Math.max(max, item.count), 0);
    const maxTelemetryTrendBucketCount = telemetrySourceStats.reduce((max, source) => {
        const sourceMax = source.trend.reduce((bucketMax, bucket) => Math.max(bucketMax, bucket.count), 0);
        return Math.max(max, sourceMax);
    }, 0);
    const recentEvictions = (evictionHistoryQuery.data as WorkingSetEvictionEvent[] | undefined) ?? [];
    const preferences = (preferencesQuery.data as ToolPreferences | undefined) ?? {
        importantTools: [],
        alwaysLoadedTools: [],
        autoLoadMinConfidence: 0.85,
        maxLoadedTools: 16,
        maxHydratedSchemas: 8,
    };
    const importantTools = new Set(preferences.importantTools);
    const alwaysLoadedTools = new Set(preferences.alwaysLoadedTools);
    const loadedToolNames = new Set(workingSet.map((tool) => tool.name));
    const alwaysOnAdvertisedNames = new Set(
        allKnownTools
            .filter((tool) => Boolean(tool.alwaysOn))
            .map((tool) => tool.name),
    );
    const alwaysOnWorkingSet = workingSet.filter((tool) => alwaysOnAdvertisedNames.has(tool.name));
    const keepWarmWorkingSet = workingSet.filter((tool) => alwaysLoadedTools.has(tool.name) && !alwaysOnAdvertisedNames.has(tool.name));
    const dynamicWorkingSet = workingSet.filter((tool) => !alwaysLoadedTools.has(tool.name) && !alwaysOnAdvertisedNames.has(tool.name));
    const hydratedCount = workingSet.filter((tool) => tool.hydrated).length;

    useEffect(() => {
        if (jsoncEditorQuery.data?.content && jsoncDraft.length === 0) {
            setJsoncDraft(jsoncEditorQuery.data.content);
        }
    }, [jsoncDraft.length, jsoncEditorQuery.data?.content]);

    useEffect(() => {
        const normalized = Math.max(0.5, Math.min(0.99, preferences.autoLoadMinConfidence ?? 0.85));
        setAutoLoadMinConfidenceDraft(normalized);
    }, [preferences.autoLoadMinConfidence]);

    useEffect(() => {
        setMaxLoadedToolsDraft(preferences.maxLoadedTools ?? 16);
        setMaxHydratedSchemasDraft(preferences.maxHydratedSchemas ?? 8);
    }, [preferences.maxLoadedTools, preferences.maxHydratedSchemas]);

    useEffect(() => {
        let hasHydratedFromUrl = false;

        const urlType = searchParams.get(TELEMETRY_TYPE_QUERY_KEY);
        const urlStatus = searchParams.get(TELEMETRY_STATUS_QUERY_KEY);
        const urlWindow = searchParams.get(TELEMETRY_WINDOW_QUERY_KEY);
        const urlSource = searchParams.get(TELEMETRY_SOURCE_QUERY_KEY);

        if (urlType && ['all', 'search', 'load', 'hydrate', 'unload'].includes(urlType)) {
            setTelemetryTypeFilter(urlType as 'all' | ToolSelectionTelemetryEvent['type']);
            hasHydratedFromUrl = true;
        }

        if (urlStatus && ['all', 'success', 'error'].includes(urlStatus)) {
            setTelemetryStatusFilter(urlStatus as 'all' | ToolSelectionTelemetryEvent['status']);
            hasHydratedFromUrl = true;
        }

        if (urlWindow && ['all', '5m', '15m', '1h', '24h'].includes(urlWindow)) {
            setTelemetryWindowFilter(urlWindow as TelemetryWindowPreset);
            hasHydratedFromUrl = true;
        }

        if (urlSource && ['all', 'runtime-search', 'cached-ranking', 'live-aggregator'].includes(urlSource)) {
            setTelemetrySourceFilter(urlSource as TelemetrySourceFilter);
            hasHydratedFromUrl = true;
        }

        if (hasHydratedFromUrl) {
            return;
        }

        try {
            const raw = window.localStorage.getItem(TELEMETRY_FILTERS_STORAGE_KEY);
            if (!raw) {
                return;
            }

            const parsed = JSON.parse(raw) as {
                type?: string;
                status?: string;
                window?: string;
                source?: string;
            };

            if (parsed.type && ['all', 'search', 'load', 'hydrate', 'unload'].includes(parsed.type)) {
                setTelemetryTypeFilter(parsed.type as 'all' | ToolSelectionTelemetryEvent['type']);
            }

            if (parsed.status && ['all', 'success', 'error'].includes(parsed.status)) {
                setTelemetryStatusFilter(parsed.status as 'all' | ToolSelectionTelemetryEvent['status']);
            }

            if (parsed.window && ['all', '5m', '15m', '1h', '24h'].includes(parsed.window)) {
                setTelemetryWindowFilter(parsed.window as TelemetryWindowPreset);
            }

            if (parsed.source && ['all', 'runtime-search', 'cached-ranking', 'live-aggregator'].includes(parsed.source)) {
                setTelemetrySourceFilter(parsed.source as TelemetrySourceFilter);
            }
        } catch {
            // Ignore invalid persisted filter payloads and continue with defaults.
        }
    }, [searchParams]);

    useEffect(() => {
        try {
            window.localStorage.setItem(
                TELEMETRY_FILTERS_STORAGE_KEY,
                JSON.stringify({
                    type: telemetryTypeFilter,
                    status: telemetryStatusFilter,
                    window: telemetryWindowFilter,
                    source: telemetrySourceFilter,
                }),
            );
        } catch {
            // Ignore storage write failures (private mode/quota) and keep UI functional.
        }
    }, [telemetrySourceFilter, telemetryStatusFilter, telemetryTypeFilter, telemetryWindowFilter]);

    useEffect(() => {
        const nextParams = new URLSearchParams(searchParams.toString());

        if (telemetryTypeFilter === 'all') {
            nextParams.delete(TELEMETRY_TYPE_QUERY_KEY);
        } else {
            nextParams.set(TELEMETRY_TYPE_QUERY_KEY, telemetryTypeFilter);
        }

        if (telemetryStatusFilter === 'all') {
            nextParams.delete(TELEMETRY_STATUS_QUERY_KEY);
        } else {
            nextParams.set(TELEMETRY_STATUS_QUERY_KEY, telemetryStatusFilter);
        }

        if (telemetryWindowFilter === '15m') {
            nextParams.delete(TELEMETRY_WINDOW_QUERY_KEY);
        } else {
            nextParams.set(TELEMETRY_WINDOW_QUERY_KEY, telemetryWindowFilter);
        }

        if (telemetrySourceFilter === 'all') {
            nextParams.delete(TELEMETRY_SOURCE_QUERY_KEY);
        } else {
            nextParams.set(TELEMETRY_SOURCE_QUERY_KEY, telemetrySourceFilter);
        }

        const currentQuery = searchParams.toString();
        const nextQuery = nextParams.toString();
        if (currentQuery === nextQuery) {
            return;
        }

        router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname, { scroll: false });
    }, [pathname, router, searchParams, telemetrySourceFilter, telemetryStatusFilter, telemetryTypeFilter, telemetryWindowFilter]);

    const updateToolPreferences = (next: ToolPreferenceMutationInput) => {
        setPreferencesMutation.mutate(next as never);
    };

    const toggleImportant = (toolName: string) => {
        const next = new Set(importantTools);
        if (next.has(toolName)) {
            next.delete(toolName);
        } else {
            next.add(toolName);
        }

        updateToolPreferences({
            importantTools: Array.from(next),
            alwaysLoadedTools: Array.from(alwaysLoadedTools),
            autoLoadMinConfidence: preferences.autoLoadMinConfidence,
            maxLoadedTools: preferences.maxLoadedTools,
            maxHydratedSchemas: preferences.maxHydratedSchemas,
        });
    };

    const toggleAlwaysLoaded = (toolName: string) => {
        const next = new Set(alwaysLoadedTools);
        if (next.has(toolName)) {
            next.delete(toolName);
        } else {
            next.add(toolName);
        }

        updateToolPreferences({
            importantTools: Array.from(importantTools),
            alwaysLoadedTools: Array.from(next),
            autoLoadMinConfidence: preferences.autoLoadMinConfidence,
            maxLoadedTools: preferences.maxLoadedTools,
            maxHydratedSchemas: preferences.maxHydratedSchemas,
        });
    };

    const saveAutoLoadMinConfidence = () => {
        const normalized = Math.max(0.5, Math.min(0.99, Number(autoLoadMinConfidenceDraft)));
        setAutoLoadMinConfidenceDraft(normalized);

        if (Math.abs(normalized - preferences.autoLoadMinConfidence) < 0.0001) {
            return;
        }

        updateToolPreferences({
            importantTools: Array.from(importantTools),
            alwaysLoadedTools: Array.from(alwaysLoadedTools),
            autoLoadMinConfidence: normalized,
            maxLoadedTools: preferences.maxLoadedTools,
            maxHydratedSchemas: preferences.maxHydratedSchemas,
        });
    };

    const saveCapacity = () => {
        const nextMax = Math.max(4, Math.min(64, Math.round(maxLoadedToolsDraft)));
        const nextHydrated = Math.max(2, Math.min(32, Math.round(maxHydratedSchemasDraft)));
        setMaxLoadedToolsDraft(nextMax);
        setMaxHydratedSchemasDraft(nextHydrated);

        if (nextMax === preferences.maxLoadedTools && nextHydrated === preferences.maxHydratedSchemas) {
            return;
        }

        updateToolPreferences({
            importantTools: Array.from(importantTools),
            alwaysLoadedTools: Array.from(alwaysLoadedTools),
            autoLoadMinConfidence: preferences.autoLoadMinConfidence,
            maxLoadedTools: nextMax,
            maxHydratedSchemas: nextHydrated,
        });
    };

    const resetTelemetryFilters = () => {
        setTelemetryTypeFilter('all');
        setTelemetryStatusFilter('all');
        setTelemetryWindowFilter('15m');
        setTelemetrySourceFilter('all');

        try {
            window.localStorage.removeItem(TELEMETRY_FILTERS_STORAGE_KEY);
        } catch {
            // Ignore local storage cleanup errors.
        }
    };

    const applyTelemetryPreset = (preset: TelemetryTriagePreset) => {
        if (preset === 'errors-now') {
            setTelemetryTypeFilter('all');
            setTelemetryStatusFilter('error');
            setTelemetryWindowFilter('15m');
            setTelemetrySourceFilter('all');
            return;
        }

        if (preset === 'runtime-failures') {
            setTelemetryTypeFilter('all');
            setTelemetryStatusFilter('error');
            setTelemetryWindowFilter('1h');
            setTelemetrySourceFilter('runtime-search');
            return;
        }

        if (preset === 'load-incidents') {
            setTelemetryTypeFilter('load');
            setTelemetryStatusFilter('error');
            setTelemetryWindowFilter('1h');
            setTelemetrySourceFilter('all');
            return;
        }

        if (preset === 'hydration-failures') {
            setTelemetryTypeFilter('hydrate');
            setTelemetryStatusFilter('error');
            setTelemetryWindowFilter('24h');
            setTelemetrySourceFilter('all');
            return;
        }

        setTelemetryTypeFilter('all');
        setTelemetryStatusFilter('all');
        setTelemetryWindowFilter('15m');
        setTelemetrySourceFilter('live-aggregator');
    };

    const copyTelemetryShareLink = async () => {
        const nextParams = new URLSearchParams(searchParams.toString());
        const shareUrl = `${window.location.origin}${pathname}${nextParams.toString() ? `?${nextParams.toString()}` : ''}`;

        try {
            await navigator.clipboard.writeText(shareUrl);
            toast.success('Share link copied');
        } catch {
            toast.error('Failed to copy share link');
        }
    };

    return (
        <div className="p-8 space-y-8 h-full flex flex-col">
            <div className="flex justify-between items-center shrink-0">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Semantic Search</h1>
                    <p className="text-zinc-500">
                        Search, load, and triage tools without dumping the entire catalog into the model’s face
                    </p>
                </div>
            </div>

            <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_380px] min-h-0 flex-1">
                <div className="space-y-6 min-h-0 flex flex-col">
                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="pb-4">
                            <CardTitle className="text-white flex items-center gap-2">
                                <Search className="h-5 w-5 text-blue-400" />
                                Search tools by intent
                            </CardTitle>
                            <p className="text-sm text-zinc-500">
                                Inspired by the better inspector palettes: search should get you to the right tool fast, not ask you to babysit a giant list.
                            </p>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="relative">
                                <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-zinc-500" />
                                <input
                                    value={query}
                                    onChange={(e) => setQuery(e.target.value)}
                                    placeholder="What do you want to achieve? e.g. process csv files"
                                    title="Search the aggregated MCP tool catalog by intent, capability, server name, and tool metadata"
                                    aria-label="Search MCP tools by intent"
                                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl p-4 pl-12 text-base text-white focus:ring-2 focus:ring-blue-500 outline-none"
                                    autoFocus
                                />
                            </div>

                            <div className="space-y-2">
                                <div className="text-xs uppercase tracking-wider text-zinc-500">Task profile</div>
                                <div className="flex flex-wrap gap-2">
                                    {[
                                        { value: 'default', label: 'Default' },
                                        { value: 'repo-coding', label: 'Repo coding' },
                                        { value: 'web-research', label: 'Web research' },
                                        { value: 'browser-automation', label: 'Browser automation' },
                                        { value: 'local-ops', label: 'Local ops' },
                                        { value: 'database', label: 'Database' },
                                    ].map((option) => {
                                        const isActive = profile === option.value;

                                        return (
                                            <button
                                                key={option.value}
                                                type="button"
                                                onClick={() => setProfile(option.value as ToolSearchProfile | 'default')}
                                                className={`rounded-md border px-3 py-1.5 text-xs transition-colors ${isActive
                                                    ? 'border-blue-500/50 bg-blue-500/15 text-blue-200'
                                                    : 'border-zinc-700 bg-zinc-950/70 text-zinc-300 hover:bg-zinc-800'
                                                    }`}
                                                title={`Bias ranking toward ${option.label.toLowerCase()} workflows`}
                                                aria-label={`Use ${option.label} task profile`}
                                            >
                                                {option.label}
                                            </button>
                                        );
                                    })}
                                </div>
                            </div>

                            <p className="text-xs text-zinc-500">
                                Tip: describe the outcome you want (for example, “sync issues from github repo” or “extract text from pdf”).
                                Ranking uses match reason + metadata confidence so the best candidates surface first.
                                {profile !== 'default' ? ` Active profile: ${profile}.` : ''}
                            </p>

                            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                                <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                    <div className="text-xs uppercase tracking-wider text-zinc-500">Matches</div>
                                    <div className="mt-1 text-2xl font-semibold text-white">{results.length}</div>
                                </div>
                                <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                    <div className="text-xs uppercase tracking-wider text-zinc-500">Loaded</div>
                                    <div className="mt-1 text-2xl font-semibold text-white">{workingSet.length}</div>
                                </div>
                                <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                    <div className="text-xs uppercase tracking-wider text-zinc-500">Hydrated schemas</div>
                                    <div className="mt-1 text-2xl font-semibold text-white">{hydratedCount}</div>
                                </div>
                                <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3 md:col-span-3">
                                    <div className="text-xs uppercase tracking-wider text-zinc-500">Always-on tools</div>
                                    <div className="mt-1 text-2xl font-semibold text-white">{alwaysLoadedTools.size}</div>
                                    <div className="mt-1 text-xs text-zinc-500">Pinned warm tools auto-load into the session working set when MCP state refreshes.</div>
                                </div>
                                <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3 md:col-span-3 space-y-3">
                                    <div className="flex items-center justify-between gap-3">
                                        <div>
                                            <div className="text-xs uppercase tracking-wider text-zinc-500">Auto-load confidence floor</div>
                                            <div className="mt-1 text-2xl font-semibold text-white">{Math.round((preferences.autoLoadMinConfidence ?? 0.85) * 100)}%</div>
                                        </div>
                                        <div className="text-xs text-zinc-500 text-right max-w-xs">
                                            Cached ranking auto-loads only when confidence is above this threshold.
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-3">
                                        <input
                                            type="range"
                                            min={0.5}
                                            max={0.99}
                                            step={0.01}
                                            value={autoLoadMinConfidenceDraft}
                                            onChange={(event) => setAutoLoadMinConfidenceDraft(Number(event.target.value))}
                                            className="w-full"
                                            title="Set minimum confidence required before TormentNexus auto-loads the top ranked tool"
                                            aria-label="Auto-load confidence threshold"
                                        />
                                        <input
                                            type="number"
                                            min={0.5}
                                            max={0.99}
                                            step={0.01}
                                            value={autoLoadMinConfidenceDraft.toFixed(2)}
                                            onChange={(event) => setAutoLoadMinConfidenceDraft(Number(event.target.value))}
                                            className="w-24 rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm text-zinc-100"
                                            title="Numeric confidence threshold between 0.50 and 0.99"
                                            aria-label="Auto-load confidence threshold numeric input"
                                        />
                                        <Button
                                            type="button"
                                            variant="outline"
                                            className="border-zinc-700 text-zinc-200 hover:bg-zinc-800"
                                            onClick={saveAutoLoadMinConfidence}
                                            disabled={setPreferencesMutation.isPending}
                                            title="Save auto-load confidence threshold"
                                            aria-label="Save auto-load confidence threshold"
                                        >
                                            Save
                                        </Button>
                                    </div>
                                </div>
                            </div>

                            {!query && (
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-3 opacity-80">
                                    <div className="p-4 rounded border border-dashed border-zinc-800 text-sm text-zinc-500 flex items-center gap-3">
                                        <Zap className="h-4 w-4" /> “memory store tools”
                                    </div>
                                    <div className="p-4 rounded border border-dashed border-zinc-800 text-sm text-zinc-500 flex items-center gap-3">
                                        <Code className="h-4 w-4" /> “github issue search”
                                    </div>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800 min-h-0 flex-1 flex flex-col">
                        <CardHeader className="pb-3 border-b border-zinc-800">
                            <CardTitle className="text-white text-base">Search results</CardTitle>
                        </CardHeader>
                        <CardContent className="p-0 overflow-y-auto flex-1">
                            {isLoading ? (
                                <div className="flex justify-center p-8">
                                    <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
                                </div>
                            ) : results.length > 0 ? (
                                <div className="divide-y divide-zinc-800/80">
                                    {results.map((tool) => {
                                        const isLoaded = tool.loaded ?? loadedToolNames.has(tool.name);

                                        return (
                                            <div key={tool.name} className="p-5 hover:bg-zinc-950/60 transition-colors space-y-4">
                                                <div className="flex items-start justify-between gap-4">
                                                    <div className="min-w-0">
                                                        <div className="font-mono text-blue-400 font-medium text-lg mb-1 break-all">{tool.name}</div>
                                                        {tool.advertisedName && tool.advertisedName !== tool.name ? (
                                                            <div className="text-xs text-cyan-300 mb-2 break-all">
                                                                advertised as <span className="font-mono">{tool.advertisedName}</span>
                                                            </div>
                                                        ) : null}
                                                        <div className="flex flex-wrap gap-2 mb-2">
                                                            {tool.rank ? (
                                                                <span className="text-[10px] bg-blue-500/10 border border-blue-500/20 px-2 py-0.5 rounded text-blue-300 uppercase tracking-wider">
                                                                    rank #{tool.rank}
                                                                </span>
                                                            ) : null}
                                                            <span className="text-[10px] bg-zinc-800 px-2 py-0.5 rounded text-zinc-400 uppercase tracking-wider">
                                                                {tool.serverDisplayName || tool.server}
                                                            </span>
                                                            {tool.semanticGroupLabel ? (
                                                                <span className="text-[10px] bg-indigo-500/10 border border-indigo-500/20 px-2 py-0.5 rounded text-indigo-300 uppercase tracking-wider">
                                                                    {tool.semanticGroupLabel}
                                                                </span>
                                                            ) : null}
                                                            {isLoaded ? (
                                                                <span className="text-[10px] bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded text-emerald-300 uppercase tracking-wider">
                                                                    loaded
                                                                </span>
                                                            ) : null}
                                                            {tool.hydrated ? (
                                                                <span className="text-[10px] bg-purple-500/10 border border-purple-500/20 px-2 py-0.5 rounded text-purple-300 uppercase tracking-wider">
                                                                    schema ready
                                                                </span>
                                                            ) : null}
                                                            {tool.requiresSchemaHydration ? (
                                                                <span className="text-[10px] bg-amber-500/10 border border-amber-500/20 px-2 py-0.5 rounded text-amber-300 uppercase tracking-wider">
                                                                    metadata only
                                                                </span>
                                                            ) : null}
                                                            {(tool.alwaysOn || alwaysOnAdvertisedNames.has(tool.name)) ? (
                                                                <span className="text-[10px] bg-sky-500/10 border border-sky-500/20 px-2 py-0.5 rounded text-sky-300 uppercase tracking-wider">
                                                                    server always-on
                                                                </span>
                                                            ) : null}
                                                            {(tool.alwaysLoaded || alwaysLoadedTools.has(tool.name)) ? (
                                                                <span className="text-[10px] bg-cyan-500/10 border border-cyan-500/20 px-2 py-0.5 rounded text-cyan-300 uppercase tracking-wider">
                                                                    keep warm profile
                                                                </span>
                                                            ) : null}
                                                            {(tool.important || importantTools.has(tool.name)) ? (
                                                                <span className="text-[10px] bg-fuchsia-500/10 border border-fuchsia-500/20 px-2 py-0.5 rounded text-fuchsia-300 uppercase tracking-wider">
                                                                    always show
                                                                </span>
                                                            ) : null}
                                                        </div>
                                                        <p className="text-zinc-400 text-sm">{tool.description || 'No description available.'}</p>
                                                        <div className="mt-3 grid gap-2 text-xs text-zinc-500 md:grid-cols-2">
                                                            <div>
                                                                <span className="uppercase tracking-wider text-zinc-600">Why it matched</span>
                                                                <div className="mt-1 text-zinc-300">{tool.matchReason ?? 'matched available tool metadata'}</div>
                                                            </div>
                                                            <div>
                                                                <span className="uppercase tracking-wider text-zinc-600">Original tool</span>
                                                                <div className="mt-1 font-mono text-zinc-300 break-all">{tool.originalName || 'n/a'}</div>
                                                            </div>
                                                        </div>
                                                        {(tool.serverTags?.length || tool.toolTags?.length) ? (
                                                            <div className="mt-3 grid gap-2 text-xs text-zinc-500 md:grid-cols-2">
                                                                <div>
                                                                    <span className="uppercase tracking-wider text-zinc-600">Server tags</span>
                                                                    <div className="mt-1 text-zinc-300">{(tool.serverTags ?? []).join(', ') || 'n/a'}</div>
                                                                </div>
                                                                <div>
                                                                    <span className="uppercase tracking-wider text-zinc-600">Tool tags</span>
                                                                    <div className="mt-1 text-zinc-300">{(tool.toolTags ?? []).join(', ') || 'n/a'}</div>
                                                                </div>
                                                            </div>
                                                        ) : null}
                                                    </div>
                                                    <Link
                                                        href={`/dashboard/mcp/inspector?tool=${encodeURIComponent(tool.name)}`}
                                                        title="Open the MCP inspector with this tool preselected"
                                                        aria-label={`Inspect tool ${tool.name}`}
                                                        className="inline-flex items-center gap-1 text-xs text-zinc-400 hover:text-white shrink-0"
                                                    >
                                                        Inspect
                                                        <ExternalLink className="h-3.5 w-3.5" />
                                                    </Link>
                                                </div>

                                                <div className="flex flex-wrap gap-2">
                                                    <Button
                                                        onClick={() => loadMutation.mutate({ name: tool.name })}
                                                        disabled={loadMutation.isPending}
                                                        title="Load this tool into the active working set so it is immediately callable"
                                                        aria-label={`Load tool ${tool.name}`}
                                                        className="bg-blue-600 hover:bg-blue-500 text-white"
                                                    >
                                                        Load tool
                                                    </Button>
                                                    <Button
                                                        onClick={() => toggleImportant(tool.name)}
                                                        disabled={setPreferencesMutation.isPending}
                                                        title="Pin this tool so it is always shown in search results"
                                                        aria-label={`${(tool.important || importantTools.has(tool.name)) ? 'Unmark' : 'Mark'} tool ${tool.name} as important`}
                                                        variant="outline"
                                                        className="border-fuchsia-700 text-fuchsia-200 hover:bg-fuchsia-950/30"
                                                    >
                                                        {(tool.important || importantTools.has(tool.name)) ? 'Unmark important' : 'Mark important'}
                                                    </Button>
                                                    <Button
                                                        onClick={() => toggleAlwaysLoaded(tool.name)}
                                                        disabled={setPreferencesMutation.isPending}
                                                        title="Keep this tool warm so it auto-loads into the active working set"
                                                        aria-label={`${(tool.alwaysLoaded || alwaysLoadedTools.has(tool.name)) ? 'Stop keeping' : 'Keep'} tool ${tool.name} always loaded`}
                                                        variant="outline"
                                                        className="border-cyan-700 text-cyan-200 hover:bg-cyan-950/30"
                                                    >
                                                        {(tool.alwaysLoaded || alwaysLoadedTools.has(tool.name)) ? 'Disable always-on' : 'Keep warm'}
                                                    </Button>
                                                    <Button
                                                        onClick={() => callToolMutation.mutate({ name: tool.name, args: {} })}
                                                        disabled={callToolMutation.isPending}
                                                        title="Invoke this tool immediately with an empty/default argument payload"
                                                        aria-label={`Call tool ${tool.name} now`}
                                                        variant="outline"
                                                        className="border-emerald-700 text-emerald-200 hover:bg-emerald-950/30"
                                                    >
                                                        Call now
                                                    </Button>
                                                    <Button
                                                        onClick={() => hydrateMutation.mutate({ name: tool.name })}
                                                        disabled={hydrateMutation.isPending || !isLoaded || Boolean(tool.hydrated)}
                                                        title="Hydrate this tool's schema into the active working set"
                                                        aria-label={`Hydrate schema for tool ${tool.name}`}
                                                        variant="outline"
                                                        className="border-purple-700 text-purple-200 hover:bg-purple-950/30"
                                                    >
                                                        Hydrate schema
                                                    </Button>
                                                    <Button
                                                        onClick={() => unloadMutation.mutate({ name: tool.name })}
                                                        disabled={unloadMutation.isPending || !isLoaded}
                                                        title="Unload this tool from the current working set"
                                                        aria-label={`Unload tool ${tool.name}`}
                                                        variant="outline"
                                                        className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                                    >
                                                        Unload
                                                    </Button>
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            ) : query ? (
                                <div className="text-center text-zinc-500 py-12 px-6">
                                    No tools found matching “{query}”.
                                </div>
                            ) : (
                                <div className="text-center text-zinc-500 py-12 px-6">
                                    Start typing to search across available MCP capabilities.
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>

                <div className="space-y-6">
                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="pb-3 border-b border-zinc-800">
                            <CardTitle className="text-white flex items-center gap-2 text-base">
                                <Layers className="h-4 w-4 text-indigo-400" />
                                Session working set
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="p-4 space-y-4">
                            <div className="grid grid-cols-2 gap-3 text-sm">
                                <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                    <div className="text-xs uppercase tracking-wider text-zinc-500">Loaded cap</div>
                                    <div className="mt-1 text-xl font-semibold text-white">{workingSetQuery.data?.limits?.maxLoadedTools ?? 0}</div>
                                </div>
                                <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                    <div className="text-xs uppercase tracking-wider text-zinc-500">Schema cap</div>
                                    <div className="mt-1 text-xl font-semibold text-white">{workingSetQuery.data?.limits?.maxHydratedSchemas ?? 0}</div>
                                </div>
                            </div>

                            <div className="space-y-3 max-h-[420px] overflow-y-auto">
                                {workingSet.length > 0 ? (
                                    <>
                                        {[
                                            { label: 'Server always-on', tone: 'text-sky-300', tools: alwaysOnWorkingSet },
                                            { label: 'Keep warm profile', tone: 'text-cyan-300', tools: keepWarmWorkingSet },
                                            { label: 'Dynamic loaded', tone: 'text-zinc-300', tools: dynamicWorkingSet },
                                        ].map((section) => (
                                            <div key={section.label} className="space-y-2">
                                                <div className={`text-[10px] uppercase tracking-wider ${section.tone}`}>
                                                    {section.label} ({section.tools.length})
                                                </div>
                                                {section.tools.length > 0 ? section.tools.map((tool) => (
                                                    <div key={tool.name} className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-3 space-y-2">
                                                        <div className="flex items-start justify-between gap-3">
                                                            <div className="min-w-0">
                                                                <div className="font-mono text-sm text-zinc-100 break-all">{tool.name}</div>
                                                                <div className="text-xs text-zinc-500 mt-1">
                                                                    loaded {formatRelativeTimestamp(tool.lastLoadedAt)}
                                                                </div>
                                                            </div>
                                                            {tool.hydrated ? (
                                                                <span className="text-[10px] bg-purple-500/10 border border-purple-500/20 px-2 py-0.5 rounded text-purple-300 uppercase tracking-wider">
                                                                    schema ready
                                                                </span>
                                                            ) : (
                                                                <span className="text-[10px] bg-zinc-800 px-2 py-0.5 rounded text-zinc-400 uppercase tracking-wider">
                                                                    metadata only
                                                                </span>
                                                            )}
                                                        </div>
                                                        <div className="grid grid-cols-3 gap-2">
                                                            <Link
                                                                href={`/dashboard/mcp/inspector?tool=${encodeURIComponent(tool.name)}`}
                                                                title="Inspect this loaded tool"
                                                                aria-label={`Inspect loaded tool ${tool.name}`}
                                                                className="inline-flex items-center justify-center rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-800"
                                                            >
                                                                Inspect
                                                            </Link>
                                                            <Button
                                                                onClick={() => hydrateMutation.mutate({ name: tool.name })}
                                                                disabled={hydrateMutation.isPending || tool.hydrated}
                                                                variant="outline"
                                                                title="Hydrate this loaded tool schema"
                                                                aria-label={`Hydrate loaded tool ${tool.name}`}
                                                                className="w-full border-purple-700 text-purple-200 hover:bg-purple-950/30"
                                                            >
                                                                Hydrate
                                                            </Button>
                                                            <Button
                                                                onClick={() => unloadMutation.mutate({ name: tool.name })}
                                                                variant="outline"
                                                                title="Remove this loaded tool from the active session"
                                                                aria-label={`Unload loaded tool ${tool.name}`}
                                                                className="w-full border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                                            >
                                                                Unload
                                                            </Button>
                                                        </div>
                                                    </div>
                                                )) : (
                                                    <div className="rounded-lg border border-dashed border-zinc-800 p-3 text-xs text-zinc-500 text-center">
                                                        none
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </>
                                ) : (
                                    <div className="rounded-lg border border-dashed border-zinc-800 p-6 text-sm text-zinc-500 text-center">
                                        No tools currently loaded.
                                    </div>
                                )}
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="pb-3 border-b border-zinc-800">
                            <CardTitle className="text-white flex items-center gap-2 text-base">
                                <SlidersHorizontal className="h-4 w-4 text-violet-400" />
                                Working-set capacity
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="p-4 space-y-4">
                            <p className="text-xs text-zinc-500">
                                Controls how many tools and schemas the session keeps warm before LRU eviction.
                            </p>

                            {/* maxLoadedTools slider */}
                            <div className="space-y-2">
                                <div className="flex items-center justify-between">
                                    <span className="text-xs uppercase tracking-wider text-zinc-400">Loaded tools cap</span>
                                    <span className="text-sm font-semibold text-white">{maxLoadedToolsDraft}</span>
                                </div>
                                <div className="flex items-center gap-3">
                                    <input
                                        type="range"
                                        min={4}
                                        max={64}
                                        step={1}
                                        value={maxLoadedToolsDraft}
                                        onChange={(e) => setMaxLoadedToolsDraft(Number(e.target.value))}
                                        className="w-full"
                                        title="Maximum number of tools loaded simultaneously before LRU eviction (4–64)"
                                        aria-label="Maximum loaded tools"
                                    />
                                    <input
                                        type="number"
                                        min={4}
                                        max={64}
                                        step={1}
                                        value={maxLoadedToolsDraft}
                                        onChange={(e) => setMaxLoadedToolsDraft(Number(e.target.value))}
                                        className="w-16 rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm text-zinc-100"
                                        title="Loaded tools cap (4–64)"
                                        aria-label="Loaded tools cap numeric input"
                                    />
                                </div>
                            </div>

                            {/* maxHydratedSchemas slider */}
                            <div className="space-y-2">
                                <div className="flex items-center justify-between">
                                    <span className="text-xs uppercase tracking-wider text-zinc-400">Hydrated schemas cap</span>
                                    <span className="text-sm font-semibold text-white">{maxHydratedSchemasDraft}</span>
                                </div>
                                <div className="flex items-center gap-3">
                                    <input
                                        type="range"
                                        min={2}
                                        max={32}
                                        step={1}
                                        value={maxHydratedSchemasDraft}
                                        onChange={(e) => setMaxHydratedSchemasDraft(Number(e.target.value))}
                                        className="w-full"
                                        title="Maximum number of hydrated schemas kept warm simultaneously before LRU eviction (2–32)"
                                        aria-label="Maximum hydrated schemas"
                                    />
                                    <input
                                        type="number"
                                        min={2}
                                        max={32}
                                        step={1}
                                        value={maxHydratedSchemasDraft}
                                        onChange={(e) => setMaxHydratedSchemasDraft(Number(e.target.value))}
                                        className="w-16 rounded-md border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm text-zinc-100"
                                        title="Hydrated schemas cap (2–32)"
                                        aria-label="Hydrated schemas cap numeric input"
                                    />
                                </div>
                            </div>

                            <Button
                                type="button"
                                variant="outline"
                                className="w-full border-zinc-700 text-zinc-200 hover:bg-zinc-800"
                                onClick={saveCapacity}
                                disabled={setPreferencesMutation.isPending}
                                title="Save working-set capacity limits and apply them to the live session"
                                aria-label="Save working-set capacity limits"
                            >
                                Apply capacity
                            </Button>
                        </CardContent>
                    </Card>

                    {recentEvictions.length > 0 && (
                        <Card className="bg-zinc-900 border-zinc-800">
                            <CardHeader className="pb-3 border-b border-zinc-800 flex flex-row items-center justify-between gap-3">
                                <CardTitle className="text-white flex items-center gap-2 text-base">
                                    <History className="h-4 w-4 text-amber-400" />
                                    Recent evictions
                                </CardTitle>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    disabled={clearEvictionHistoryMutation.isPending || recentEvictions.length === 0}
                                    onClick={() => clearEvictionHistoryMutation.mutate()}
                                    title="Clear the recent working-set eviction history"
                                    aria-label="Clear working-set eviction history"
                                    className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                >
                                    {clearEvictionHistoryMutation.isPending ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <Trash2 className="mr-2 h-3.5 w-3.5" />}
                                    Clear
                                </Button>
                            </CardHeader>
                            <CardContent className="p-4">
                                <div className="space-y-2 max-h-[200px] overflow-y-auto">
                                    {recentEvictions.slice(0, 10).map((event, index) => (
                                        // eslint-disable-next-line react/no-array-index-key
                                        <div key={`${event.toolName}-${event.timestamp}-${index}`} className="flex items-center justify-between gap-3 rounded-lg border border-zinc-800 bg-zinc-950/60 px-3 py-2">
                                            <span className="font-mono text-xs text-zinc-200 break-all min-w-0">{event.toolName}</span>
                                            <div className="flex items-center gap-2 shrink-0">
                                                <span className={`text-[10px] px-1.5 py-0.5 rounded border ${event.tier === 'loaded' ? 'border-red-500/20 bg-red-500/10 text-red-300' : 'border-amber-500/20 bg-amber-500/10 text-amber-300'}`}>
                                                    {event.tier}
                                                </span>
                                                <span className="text-[10px] text-zinc-500">{formatRelativeTimestamp(event.timestamp)}</span>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    )}

                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="pb-3 border-b border-zinc-800 flex flex-row items-center justify-between gap-3">
                            <CardTitle className="text-white flex items-center gap-2 text-base">
                                <Activity className="h-4 w-4 text-emerald-400" />
                                Search & loading telemetry
                            </CardTitle>
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                disabled={clearTelemetryMutation.isPending || telemetry.length === 0}
                                onClick={() => clearTelemetryMutation.mutate()}
                                title="Clear the recent search/load telemetry timeline shown in this panel"
                                aria-label="Clear MCP search telemetry history"
                                className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                            >
                                {clearTelemetryMutation.isPending ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : <Trash2 className="mr-2 h-3.5 w-3.5" />}
                                Clear
                            </Button>
                        </CardHeader>
                        <CardContent className="p-4">
                            <div className="mb-3 grid gap-2">
                                <div className="flex flex-wrap items-center gap-2 text-xs">
                                    <span className="text-zinc-500 uppercase tracking-wider">Presets</span>
                                    {([
                                        { value: 'errors-now', label: 'Errors now' },
                                        { value: 'runtime-failures', label: 'Runtime failures' },
                                        { value: 'load-incidents', label: 'Load incidents' },
                                        { value: 'hydration-failures', label: 'Hydration failures' },
                                        { value: 'live-aggregator-focus', label: 'Live aggregator' },
                                    ] as const).map((preset) => (
                                        <button
                                            key={`telemetry-preset-${preset.value}`}
                                            type="button"
                                            onClick={() => applyTelemetryPreset(preset.value)}
                                            className="rounded-md border border-zinc-700 bg-zinc-950/70 px-2 py-1 text-zinc-300 transition-colors hover:bg-zinc-800"
                                            title={`Apply ${preset.label.toLowerCase()} telemetry triage preset`}
                                            aria-label={`Apply ${preset.label} telemetry triage preset`}
                                        >
                                            {preset.label}
                                        </button>
                                    ))}
                                </div>

                                <div className="flex flex-wrap items-center gap-2 text-xs">
                                    <span className="text-zinc-500 uppercase tracking-wider">Active filters</span>
                                    {telemetryTypeFilter !== 'all' ? (
                                        <button
                                            type="button"
                                            onClick={() => setTelemetryTypeFilter('all')}
                                            className="rounded-md border border-blue-500/30 bg-blue-500/10 px-2 py-1 text-blue-200 transition-colors hover:bg-blue-500/20"
                                            title="Clear telemetry type filter"
                                            aria-label="Clear telemetry type filter"
                                        >
                                            type: {telemetryTypeFilter} ×
                                        </button>
                                    ) : null}
                                    {telemetryStatusFilter !== 'all' ? (
                                        <button
                                            type="button"
                                            onClick={() => setTelemetryStatusFilter('all')}
                                            className="rounded-md border border-cyan-500/30 bg-cyan-500/10 px-2 py-1 text-cyan-200 transition-colors hover:bg-cyan-500/20"
                                            title="Clear telemetry status filter"
                                            aria-label="Clear telemetry status filter"
                                        >
                                            status: {telemetryStatusFilter} ×
                                        </button>
                                    ) : null}
                                    {telemetryWindowFilter !== '15m' ? (
                                        <button
                                            type="button"
                                            onClick={() => setTelemetryWindowFilter('15m')}
                                            className="rounded-md border border-violet-500/30 bg-violet-500/10 px-2 py-1 text-violet-200 transition-colors hover:bg-violet-500/20"
                                            title="Clear telemetry window filter"
                                            aria-label="Clear telemetry window filter"
                                        >
                                            window: {telemetryWindowFilter} ×
                                        </button>
                                    ) : null}
                                    {telemetrySourceFilter !== 'all' ? (
                                        <button
                                            type="button"
                                            onClick={() => setTelemetrySourceFilter('all')}
                                            className="rounded-md border border-amber-500/30 bg-amber-500/10 px-2 py-1 text-amber-200 transition-colors hover:bg-amber-500/20"
                                            title="Clear telemetry source filter"
                                            aria-label="Clear telemetry source filter"
                                        >
                                            source: {telemetrySourceFilter} ×
                                        </button>
                                    ) : null}
                                    {telemetryFiltersAtDefault ? (
                                        <span className="rounded-md border border-zinc-700 bg-zinc-950/70 px-2 py-1 text-zinc-500">
                                            default scope
                                        </span>
                                    ) : null}
                                </div>

                                <div className="flex flex-wrap items-center gap-2 text-xs">
                                    <span className="rounded-md border border-zinc-700 bg-zinc-950/70 px-2 py-1 text-zinc-300">
                                        total: {telemetrySummary.total}
                                    </span>
                                    <span className="rounded-md border border-emerald-500/30 bg-emerald-500/10 px-2 py-1 text-emerald-300">
                                        success: {telemetrySummary.success}
                                    </span>
                                    <span className="rounded-md border border-red-500/30 bg-red-500/10 px-2 py-1 text-red-300">
                                        errors: {telemetrySummary.error}
                                    </span>
                                </div>

                                <div className="flex flex-wrap items-center gap-2 text-xs">
                                    <span className="text-zinc-500 uppercase tracking-wider">Type</span>
                                    {(['all', 'search', 'load', 'hydrate', 'unload'] as const).map((option) => {
                                        const active = telemetryTypeFilter === option;
                                        return (
                                            <button
                                                key={`telemetry-type-${option}`}
                                                type="button"
                                                onClick={() => setTelemetryTypeFilter(option)}
                                                className={`rounded-md border px-2 py-1 transition-colors ${active
                                                    ? 'border-blue-500/50 bg-blue-500/15 text-blue-200'
                                                    : 'border-zinc-700 bg-zinc-950/70 text-zinc-300 hover:bg-zinc-800'
                                                    }`}
                                                title={`Filter telemetry by ${option} events`}
                                                aria-label={`Filter telemetry by ${option} events`}
                                            >
                                                {option}
                                            </button>
                                        );
                                    })}
                                </div>

                                <div className="flex flex-wrap items-center gap-2 text-xs">
                                    <span className="text-zinc-500 uppercase tracking-wider">Status</span>
                                    {(['all', 'success', 'error'] as const).map((option) => {
                                        const active = telemetryStatusFilter === option;
                                        return (
                                            <button
                                                key={`telemetry-status-${option}`}
                                                type="button"
                                                onClick={() => setTelemetryStatusFilter(option)}
                                                className={`rounded-md border px-2 py-1 transition-colors ${active
                                                    ? 'border-cyan-500/50 bg-cyan-500/15 text-cyan-200'
                                                    : 'border-zinc-700 bg-zinc-950/70 text-zinc-300 hover:bg-zinc-800'
                                                    }`}
                                                title={`Filter telemetry by ${option} status`}
                                                aria-label={`Filter telemetry by ${option} status`}
                                            >
                                                {option}
                                            </button>
                                        );
                                    })}
                                </div>

                                <div className="flex flex-wrap items-center gap-2 text-xs">
                                    <span className="text-zinc-500 uppercase tracking-wider">Window</span>
                                    {([
                                        { value: 'all', label: 'All' },
                                        { value: '5m', label: '5m' },
                                        { value: '15m', label: '15m' },
                                        { value: '1h', label: '1h' },
                                        { value: '24h', label: '24h' },
                                    ] as const).map((option) => {
                                        const active = telemetryWindowFilter === option.value;
                                        return (
                                            <button
                                                key={`telemetry-window-${option.value}`}
                                                type="button"
                                                onClick={() => setTelemetryWindowFilter(option.value)}
                                                className={`rounded-md border px-2 py-1 transition-colors ${active
                                                    ? 'border-violet-500/50 bg-violet-500/15 text-violet-200'
                                                    : 'border-zinc-700 bg-zinc-950/70 text-zinc-300 hover:bg-zinc-800'
                                                    }`}
                                                title={`Filter telemetry to ${option.label} window`}
                                                aria-label={`Filter telemetry to ${option.label} window`}
                                            >
                                                {option.label}
                                            </button>
                                        );
                                    })}
                                </div>

                                <div className="flex flex-wrap items-center gap-2 text-xs">
                                    <span className="text-zinc-500 uppercase tracking-wider">Source</span>
                                    {([
                                        { value: 'all', label: 'All' },
                                        { value: 'runtime-search', label: 'Runtime' },
                                        { value: 'cached-ranking', label: 'Cached' },
                                        { value: 'live-aggregator', label: 'Live' },
                                    ] as const).map((option) => {
                                        const active = telemetrySourceFilter === option.value;
                                        return (
                                            <button
                                                key={`telemetry-source-filter-${option.value}`}
                                                type="button"
                                                onClick={() => setTelemetrySourceFilter(option.value)}
                                                className={`rounded-md border px-2 py-1 transition-colors ${active
                                                    ? 'border-amber-500/50 bg-amber-500/15 text-amber-200'
                                                    : 'border-zinc-700 bg-zinc-950/70 text-zinc-300 hover:bg-zinc-800'
                                                    }`}
                                                title={`Filter telemetry to ${option.label} source`}
                                                aria-label={`Filter telemetry to ${option.label} source`}
                                            >
                                                {option.label}
                                            </button>
                                        );
                                    })}

                                    <button
                                        type="button"
                                        onClick={resetTelemetryFilters}
                                        disabled={telemetryFiltersAtDefault}
                                        className="ml-auto rounded-md border border-zinc-700 bg-zinc-950/70 px-2 py-1 text-zinc-300 transition-colors hover:bg-zinc-800"
                                        title="Reset telemetry type/status/window/source filters to defaults"
                                        aria-label="Reset telemetry filters"
                                    >
                                        Reset filters
                                    </button>

                                    <button
                                        type="button"
                                        onClick={copyTelemetryShareLink}
                                        className="rounded-md border border-zinc-700 bg-zinc-950/70 px-2 py-1 text-zinc-300 transition-colors hover:bg-zinc-800"
                                        title="Copy URL with current telemetry filters"
                                        aria-label="Copy telemetry share link"
                                    >
                                        Copy link
                                    </button>
                                </div>
                            </div>

                            <div className="mb-4 space-y-2 rounded-lg border border-zinc-800 bg-zinc-950/50 p-3">
                                <div className="text-[10px] uppercase tracking-wider text-zinc-500">Status trend ({telemetryWindowFilter})</div>
                                {telemetryStatusTrend.some((bucket) => bucket.total > 0) ? (
                                    <div className="grid grid-cols-6 gap-1">
                                        {telemetryStatusTrend.map((bucket) => {
                                            const successWidth = bucket.total > 0 ? Math.round((bucket.successCount / bucket.total) * 100) : 0;
                                            const errorWidth = bucket.total > 0 ? Math.round((bucket.errorCount / bucket.total) * 100) : 0;

                                            return (
                                                <div key={`status-trend-${bucket.label}`} className="space-y-1" title={`${bucket.label} • ${bucket.successCount} ok / ${bucket.errorCount} err`}>
                                                    <div className="h-2 rounded border border-zinc-800/80 bg-zinc-900/80 overflow-hidden flex">
                                                        <div className="h-full bg-emerald-500/70" style={{ width: `${successWidth}%` }} />
                                                        <div className="h-full bg-red-500/75" style={{ width: `${errorWidth}%` }} />
                                                    </div>
                                                    <div className="text-[9px] text-zinc-500 text-center">{bucket.label}</div>
                                                </div>
                                            );
                                        })}
                                    </div>
                                ) : (
                                    <div className="text-xs text-zinc-500">
                                        No status trend data in the selected scope.
                                    </div>
                                )}
                            </div>

                            <div className="mb-4 space-y-2 rounded-lg border border-zinc-800 bg-zinc-950/50 p-3">
                                <div className="text-[10px] uppercase tracking-wider text-zinc-500">Per-source breakdown</div>
                                {telemetrySourceStats.some((item) => item.count > 0) ? (
                                    <div className="space-y-2">
                                        {telemetrySourceStats.map((item) => {
                                            const widthPercent = maxTelemetrySourceCount > 0
                                                ? Math.max(6, Math.round((item.count / maxTelemetrySourceCount) * 100))
                                                : 0;

                                            return (
                                                <div key={`telemetry-source-${item.source}`} className="space-y-2">
                                                    <div className="flex items-center justify-between gap-3 text-xs">
                                                        <span className="font-mono text-zinc-300">{item.source}</span>
                                                        <div className="flex items-center gap-2">
                                                            <span className="text-zinc-500">
                                                                {item.count} events • {item.success} ok / {item.error} err • avg {item.avgLatencyMs}ms
                                                            </span>
                                                            <button
                                                                type="button"
                                                                onClick={() => {
                                                                    setTelemetrySourceFilter(item.source);
                                                                    setTelemetryStatusFilter('error');
                                                                }}
                                                                disabled={item.error === 0}
                                                                className="rounded border border-red-500/30 bg-red-500/10 px-2 py-0.5 text-[10px] text-red-200 disabled:opacity-40"
                                                                title="Focus this source and show only error events"
                                                                aria-label={`Focus failing events for ${item.source}`}
                                                            >
                                                                Focus failures
                                                            </button>
                                                        </div>
                                                    </div>
                                                    <div className="h-1.5 w-full rounded bg-zinc-800/80">
                                                        <div
                                                            className="h-1.5 rounded bg-gradient-to-r from-cyan-500/70 to-blue-500/70"
                                                            style={{ width: `${widthPercent}%` }}
                                                        />
                                                    </div>

                                                    {item.trend.length > 0 ? (
                                                        <div className="space-y-1">
                                                            <div className="text-[10px] uppercase tracking-wider text-zinc-500">trend ({telemetryWindowFilter})</div>
                                                            <div className="grid grid-cols-6 gap-1">
                                                                {item.trend.map((bucket) => {
                                                                    const intensity = maxTelemetryTrendBucketCount > 0
                                                                        ? Math.max(0, Math.min(1, bucket.count / maxTelemetryTrendBucketCount))
                                                                        : 0;
                                                                    const errorRatio = bucket.count > 0
                                                                        ? bucket.errorCount / bucket.count
                                                                        : 0;

                                                                    return (
                                                                        <div
                                                                            key={`${item.source}-${bucket.label}`}
                                                                            className="space-y-1"
                                                                            title={`${bucket.label} • ${bucket.count} events • ${bucket.errorCount} errors`}
                                                                        >
                                                                            <div className="h-2 rounded border border-zinc-800/80 bg-zinc-900/80 overflow-hidden">
                                                                                <div
                                                                                    className="h-full bg-cyan-500/80"
                                                                                    style={{ width: `${Math.round(intensity * 100)}%` }}
                                                                                />
                                                                                {errorRatio > 0 ? (
                                                                                    <div
                                                                                        className="-mt-2 h-full bg-red-500/70"
                                                                                        style={{ width: `${Math.round(errorRatio * 100)}%` }}
                                                                                    />
                                                                                ) : null}
                                                                            </div>
                                                                            <div className="text-[9px] text-zinc-500 text-center">{bucket.label}</div>
                                                                        </div>
                                                                    );
                                                                })}
                                                            </div>
                                                        </div>
                                                    ) : null}
                                                </div>
                                            );
                                        })}
                                    </div>
                                ) : (
                                    <div className="text-xs text-zinc-500">
                                        No source telemetry in the selected filter window.
                                    </div>
                                )}
                            </div>

                            <div className="space-y-3 max-h-[420px] overflow-y-auto">
                                {telemetry.length > 0 ? (
                                    telemetry.map((event) => (
                                        <div key={event.id} className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-3 space-y-2">
                                            <div className="flex items-center justify-between gap-3">
                                                <div className="flex items-center gap-2 min-w-0">
                                                    {event.type === 'search' ? <Sparkles className="h-4 w-4 text-blue-400" /> : null}
                                                    {event.type === 'load' ? <ArrowDownToLine className="h-4 w-4 text-emerald-400" /> : null}
                                                    {event.type === 'hydrate' ? <Database className="h-4 w-4 text-purple-400" /> : null}
                                                    {event.type === 'unload' ? <Layers className="h-4 w-4 text-zinc-400" /> : null}
                                                    <span className="text-xs uppercase tracking-wider text-zinc-300">{event.type}</span>
                                                    <span className={`text-[10px] px-2 py-0.5 rounded border ${event.status === 'success' ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300' : 'border-red-500/20 bg-red-500/10 text-red-300'}`}>
                                                        {event.status}
                                                    </span>
                                                </div>
                                                <span className="text-[10px] text-zinc-500">{formatRelativeTimestamp(event.timestamp)}</span>
                                            </div>

                                            {event.query ? <div className="text-xs text-zinc-400 break-all">query: <span className="text-zinc-200">{event.query}</span></div> : null}
                                            {event.profile ? <div className="text-xs text-zinc-400 break-all">profile: <span className="text-zinc-200">{event.profile}</span></div> : null}
                                            {event.toolName ? <div className="text-xs text-zinc-400 break-all">tool: <span className="font-mono text-zinc-200">{event.toolName}</span></div> : null}
                                            {typeof event.resultCount === 'number' ? <div className="text-xs text-zinc-400">results: <span className="text-zinc-200">{event.resultCount}</span></div> : null}
                                            {event.topResultName ? <div className="text-xs text-zinc-400 break-all">top result: <span className="font-mono text-zinc-200">{event.topResultName}</span></div> : null}
                                            {event.topMatchReason ? <div className="text-xs text-zinc-400">why: <span className="text-zinc-200">{event.topMatchReason}</span></div> : null}
                                            {typeof event.topScore === 'number' ? <div className="text-xs text-zinc-400">top score: <span className="text-zinc-200">{event.topScore.toFixed(1)}</span></div> : null}
                                            {event.secondResultName ? <div className="text-xs text-zinc-400 break-all">second result: <span className="font-mono text-zinc-200">{event.secondResultName}</span></div> : null}
                                            {event.secondMatchReason ? <div className="text-xs text-zinc-500">second why: <span className="text-zinc-300">{event.secondMatchReason}</span></div> : null}
                                            {typeof event.secondScore === 'number' ? <div className="text-xs text-zinc-500">second score: <span className="text-zinc-300">{event.secondScore.toFixed(1)}</span></div> : null}
                                            {typeof event.scoreGap === 'number' ? <div className="text-xs text-zinc-400">score gap: <span className="text-zinc-200">{event.scoreGap.toFixed(1)}</span></div> : null}
                                            {typeof event.autoLoadConfidence === 'number' ? (
                                                <div className="text-xs text-cyan-300">confidence: {(event.autoLoadConfidence * 100).toFixed(0)}%</div>
                                            ) : null}
                                            {typeof event.autoLoadMinConfidence === 'number' ? (
                                                <div className="text-xs text-zinc-400">confidence floor: {(event.autoLoadMinConfidence * 100).toFixed(0)}%</div>
                                            ) : null}
                                            {event.autoLoadEvaluated ? (
                                                <div className="text-xs text-zinc-400">auto-load evaluated: <span className="text-zinc-200">yes</span></div>
                                            ) : null}
                                            {event.autoLoadOutcome ? (
                                                <div className="text-xs text-zinc-400">auto-load outcome: <span className="text-zinc-200">{event.autoLoadOutcome}</span></div>
                                            ) : null}
                                            {event.autoLoadExecutionStatus ? (
                                                <div className="text-xs text-zinc-400">auto-load execution: <span className="text-zinc-200">{event.autoLoadExecutionStatus}</span></div>
                                            ) : null}
                                            {event.autoLoadReason ? <div className="text-xs text-cyan-300 break-all">auto-load: {event.autoLoadReason}</div> : null}
                                            {event.autoLoadSkipReason ? <div className="text-xs text-amber-300 break-all">auto-load skipped: {event.autoLoadSkipReason}</div> : null}
                                            {event.autoLoadExecutionError ? <div className="text-xs text-red-300 break-all">auto-load failed: {event.autoLoadExecutionError}</div> : null}
                                            {typeof event.latencyMs === 'number' ? <div className="text-xs text-zinc-500">latency: {event.latencyMs}ms</div> : null}
                                            {event.source ? <div className="text-xs text-zinc-500">source: {event.source}</div> : null}
                                            {event.evictedTools && event.evictedTools.length > 0 ? (
                                                <div className="text-xs text-amber-300 break-all">evicted: {event.evictedTools.join(', ')}</div>
                                            ) : null}
                                            {event.message ? <div className="text-xs text-zinc-500 break-all">{event.message}</div> : null}
                                        </div>
                                    ))
                                ) : (
                                    <div className="rounded-lg border border-dashed border-zinc-800 p-6 text-sm text-zinc-500 text-center">
                                        Telemetry will appear as searches, loads, hydrations, and evictions happen.
                                    </div>
                                )}
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="pb-3 border-b border-zinc-800">
                            <CardTitle className="text-white text-base">mcp.jsonc editor</CardTitle>
                        </CardHeader>
                        <CardContent className="p-4 space-y-3">
                            <div className="text-xs text-zinc-500 break-all">{jsoncEditorQuery.data?.path ?? 'mcp.jsonc'}</div>
                            <textarea
                                value={jsoncDraft}
                                onChange={(event) => setJsoncDraft(event.target.value)}
                                title="Edit the TormentNexus MCP JSONC configuration. Changes are saved to the TormentNexus config mcp.jsonc file (typically ~/.tormentnexus/mcp.jsonc)."
                                aria-label="MCP JSONC configuration editor"
                                className="w-full h-48 bg-zinc-950 border border-zinc-800 rounded-md p-3 font-mono text-xs text-zinc-200 outline-none"
                                spellCheck={false}
                            />
                            <div className="flex gap-2">
                                <Button
                                    onClick={() => saveJsoncMutation.mutate({ content: jsoncDraft })}
                                    disabled={saveJsoncMutation.isPending || jsoncDraft.trim().length < 2}
                                    title="Save JSONC changes and refresh MCP config-dependent views"
                                    aria-label="Save MCP JSONC configuration"
                                    className="bg-indigo-600 hover:bg-indigo-500 text-white"
                                >
                                    Save JSONC
                                </Button>
                                <Button
                                    variant="outline"
                                    className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
                                    onClick={() => setJsoncDraft(jsoncEditorQuery.data?.content ?? '')}
                                    title="Discard unsaved edits and restore the latest loaded JSONC content"
                                    aria-label="Reset MCP JSONC editor to loaded content"
                                >
                                    Reset
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
