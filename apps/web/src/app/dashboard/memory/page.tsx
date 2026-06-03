"use client";

import { useEffect, useMemo, useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent, Button, Badge, ScrollArea } from "@tormentnexus/ui";
import { Loader2, Brain, Search, Database, History, Zap, Filter, Plus, Save, Download, RefreshCw, ChevronRight } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';
import {
    filterMemoryRecords,
    getMemoryBadgeLabel,
    getMemoryDetailSections,
    getMemoryModeHint,
    getMemoryPivotSections,
    getMemoryPreview,
    getMemoryProvenance,
    getMemoryRecordKey,
    getMemorySessionId,
    getRelatedMemoryRecords,
    getMemoryTimestamp,
    getMemoryTitle,
    groupMemoryWindowAroundAnchor,
    groupMemoryRecordsByDay,
    MEMORY_MODEL_PILLARS,
    MEMORY_SEARCH_MODES,
    sortMemoryRecordsByTimestamp,
    type MemoryRecord,
    type MemoryPivotAction,
    type RelatedMemoryRecord,
    type MemorySearchMode,
} from './memory-dashboard-utils';

type MemoryInterchangeFormat = 'json' | 'csv' | 'jsonl' | 'json-provider' | 'tormentnexus-store';

const MEMORY_FORMAT_OPTIONS: Array<{ value: MemoryInterchangeFormat; label: string }> = [
    { value: 'json', label: 'Canonical JSON' },
    { value: 'csv', label: 'Canonical CSV' },
    { value: 'jsonl', label: 'Canonical JSONL' },
    { value: 'json-provider', label: 'TormentNexus JSON Provider' },
    { value: 'tormentnexus-store', label: 'tormentnexus Store' },
];

export default function MemoryDashboard() {
    const utils = trpc.useUtils();
    const [searchQuery, setSearchQuery] = useState('');
    const [memoryType, setMemoryType] = useState<'session' | 'working' | 'long_term'>('working');
    const [searchMode, setSearchMode] = useState<MemorySearchMode>('all');
    const [newFact, setNewFact] = useState('');
    const [exportFormat, setExportFormat] = useState<MemoryInterchangeFormat>('json');
    const [convertToFormat, setConvertToFormat] = useState<MemoryInterchangeFormat>('tormentnexus-store');
    const [importing, setImporting] = useState(false);
    const [converting, setConverting] = useState(false);
    const [selectedRecordKey, setSelectedRecordKey] = useState<string | null>(null);
    const [activePivot, setActivePivot] = useState<MemoryPivotAction | null>(null);
    const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null);
    const trimmedSearchQuery = searchQuery.trim();
    const hasSearchQuery = trimmedSearchQuery.length > 0;

    const { data: stats } = trpc.memory.getAgentStats.useQuery(undefined, { refetchInterval: 10000 });
    const recentObservationsQuery = trpc.memory.getRecentObservations.useQuery({
        limit: 6,
        namespace: 'project',
    }, { refetchInterval: 10000 });
    const recentPromptsQuery = trpc.memory.getRecentUserPrompts.useQuery({
        limit: 5,
    }, { refetchInterval: 10000 });
    const recentSessionSummariesQuery = trpc.memory.getRecentSessionSummaries.useQuery({
        limit: 4,
    }, { refetchInterval: 10000 });
    const genericSearchQuery = trpc.memory.searchAgentMemory.useQuery({
        query: trimmedSearchQuery,
        type: memoryType,
        limit: 20
    }, { enabled: searchMode === 'all' || searchMode === 'facts' });
    const observationSearchQuery = trpc.memory.searchObservations.useQuery({
        query: trimmedSearchQuery,
        limit: 20,
        namespace: 'project',
    }, { enabled: (searchMode === 'observations' || searchMode === 'all') && hasSearchQuery });
    const promptSearchQuery = trpc.memory.searchUserPrompts.useQuery({
        query: trimmedSearchQuery,
        limit: 20,
    }, { enabled: (searchMode === 'prompts' || searchMode === 'all') && hasSearchQuery });
    const sessionSummarySearchQuery = trpc.memory.searchSessionSummaries.useQuery({
        query: trimmedSearchQuery,
        limit: 20,
    }, { enabled: (searchMode === 'session_summaries' || searchMode === 'all') && hasSearchQuery });
    const pivotSearchQuery = trpc.memory.searchMemoryPivot.useQuery({
        pivot: activePivot?.group ?? 'session',
        value: activePivot?.query ?? '',
        limit: 20,
    }, { enabled: activePivot !== null });

    const mergeMemoryRecords = (groups: Array<MemoryRecord[] | undefined>): MemoryRecord[] => {
        const merged = new Map<string, MemoryRecord>();

        for (const group of groups) {
            for (const memory of group ?? []) {
                merged.set(getMemoryRecordKey(memory), memory);
            }
        }

        return Array.from(merged.values());
    };

    const refreshMemoryViews = async () => {
        await Promise.all([
            utils.memory.getAgentStats.invalidate(),
            utils.memory.searchAgentMemory.invalidate(),
            utils.memory.searchObservations.invalidate(),
            utils.memory.searchUserPrompts.invalidate(),
            utils.memory.searchSessionSummaries.invalidate(),
            utils.memory.searchMemoryPivot.invalidate(),
            utils.memory.getMemoryTimelineWindow.invalidate(),
            utils.memory.getCrossSessionMemoryLinks.invalidate(),
            utils.memory.getRecentObservations.invalidate(),
            utils.memory.getRecentUserPrompts.invalidate(),
            utils.memory.getRecentSessionSummaries.invalidate(),
        ]);
    };

    const activeResults = useMemo<MemoryRecord[]>(() => {
        if (activePivot) {
            return filterMemoryRecords(((pivotSearchQuery.data as MemoryRecord[] | undefined) ?? []), searchMode);
        }

        if (searchMode === 'all') {
            return mergeMemoryRecords(hasSearchQuery
                ? [
                    (genericSearchQuery.data as MemoryRecord[] | undefined),
                    (observationSearchQuery.data as MemoryRecord[] | undefined),
                    (promptSearchQuery.data as MemoryRecord[] | undefined),
                    (sessionSummarySearchQuery.data as MemoryRecord[] | undefined),
                ]
                : [
                    (genericSearchQuery.data as MemoryRecord[] | undefined),
                    (recentObservationsQuery.data as MemoryRecord[] | undefined),
                    (recentPromptsQuery.data as MemoryRecord[] | undefined),
                    (recentSessionSummariesQuery.data as MemoryRecord[] | undefined),
                ]);
        }

        if (searchMode === 'observations') {
            return hasSearchQuery
                ? ((observationSearchQuery.data as MemoryRecord[] | undefined) ?? [])
                : ((recentObservationsQuery.data as MemoryRecord[] | undefined) ?? []);
        }

        if (searchMode === 'prompts') {
            return hasSearchQuery
                ? ((promptSearchQuery.data as MemoryRecord[] | undefined) ?? [])
                : ((recentPromptsQuery.data as MemoryRecord[] | undefined) ?? []);
        }

        if (searchMode === 'session_summaries') {
            return hasSearchQuery
                ? ((sessionSummarySearchQuery.data as MemoryRecord[] | undefined) ?? [])
                : ((recentSessionSummariesQuery.data as MemoryRecord[] | undefined) ?? []);
        }

        return filterMemoryRecords(((genericSearchQuery.data as MemoryRecord[] | undefined) ?? []), searchMode);
    }, [
        genericSearchQuery.data,
        hasSearchQuery,
        observationSearchQuery.data,
        promptSearchQuery.data,
        recentObservationsQuery.data,
        recentPromptsQuery.data,
        recentSessionSummariesQuery.data,
        activePivot,
        pivotSearchQuery.data,
        searchMode,
        sessionSummarySearchQuery.data,
    ]);

    const timelineRecords = useMemo(() => sortMemoryRecordsByTimestamp(activeResults), [activeResults]);
    const timelineGroups = useMemo(() => groupMemoryRecordsByDay(timelineRecords), [timelineRecords]);
    const selectedMemory = useMemo(() => {
        if (!timelineRecords.length) {
            return null;
        }

        return timelineRecords.find((memory) => getMemoryRecordKey(memory) === selectedRecordKey) ?? timelineRecords[0];
    }, [selectedRecordKey, timelineRecords]);
    const timelineWindowQuery = trpc.memory.getMemoryTimelineWindow.useQuery({
        sessionId: selectedSessionId ?? '',
        anchorTimestamp: selectedMemory ? getMemoryTimestamp(selectedMemory) : 0,
        before: 3,
        after: 3,
    }, {
        enabled: Boolean(selectedMemory && selectedSessionId),
    });
    const crossSessionLinksQuery = trpc.memory.getCrossSessionMemoryLinks.useQuery({
        memoryId: selectedMemory?.id ?? '',
        limit: 4,
    }, {
        enabled: Boolean(selectedMemory?.id),
    });
    const relatedRecords = useMemo(() => {
        if (!selectedMemory) {
            return [];
        }

        return getRelatedMemoryRecords(selectedMemory, timelineRecords);
    }, [selectedMemory, timelineRecords]);
    const pivotSections = useMemo(() => {
        if (!selectedMemory) {
            return [];
        }

        return getMemoryPivotSections(selectedMemory);
    }, [selectedMemory]);
    const sessionWindowRecords = useMemo(() => {
        if (!selectedMemory) {
            return [];
        }

        return (((timelineWindowQuery.data as MemoryRecord[] | undefined) ?? [])
            .filter((memory) => getMemoryRecordKey(memory) !== getMemoryRecordKey(selectedMemory)));
    }, [selectedMemory, timelineWindowQuery.data]);
    const sessionWindowGroups = useMemo(() => {
        if (!selectedMemory) {
            return [];
        }

        return groupMemoryWindowAroundAnchor(selectedMemory, sessionWindowRecords);
    }, [selectedMemory, sessionWindowRecords]);
    const crossSessionLinks = useMemo(() => {
        return (crossSessionLinksQuery.data as RelatedMemoryRecord[] | undefined) ?? [];
    }, [crossSessionLinksQuery.data]);

    const handlePivotAction = (action: MemoryPivotAction) => {
        setActivePivot(action);
        setSearchMode(action.mode);
        setSearchQuery(action.query);
        setSelectedRecordKey(null);
    };

    useEffect(() => {
        if (!timelineRecords.length) {
            if (selectedRecordKey !== null) {
                setSelectedRecordKey(null);
            }
            return;
        }

        const hasCurrentSelection = selectedRecordKey
            ? timelineRecords.some((memory) => getMemoryRecordKey(memory) === selectedRecordKey)
            : false;

        if (!hasCurrentSelection) {
            setSelectedRecordKey(getMemoryRecordKey(timelineRecords[0]));
        }
    }, [selectedRecordKey, timelineRecords]);

    useEffect(() => {
        setSelectedSessionId(selectedMemory ? getMemorySessionId(selectedMemory) : null);
    }, [selectedMemory]);

    const activeLoading =
        activePivot
            ? pivotSearchQuery.isLoading
            : searchMode === 'all'
            ? (hasSearchQuery
                ? genericSearchQuery.isLoading || observationSearchQuery.isLoading || promptSearchQuery.isLoading || sessionSummarySearchQuery.isLoading
                : genericSearchQuery.isLoading || recentObservationsQuery.isLoading || recentPromptsQuery.isLoading || recentSessionSummariesQuery.isLoading)
            : searchMode === 'observations'
            ? (hasSearchQuery ? observationSearchQuery.isLoading : recentObservationsQuery.isLoading)
            : searchMode === 'prompts'
                ? (hasSearchQuery ? promptSearchQuery.isLoading : recentPromptsQuery.isLoading)
                : searchMode === 'session_summaries'
                    ? (hasSearchQuery ? sessionSummarySearchQuery.isLoading : recentSessionSummariesQuery.isLoading)
                    : genericSearchQuery.isLoading;

    const activeEmptyMessage = activePivot
        ? `No related memory records found for this ${activePivot.group} pivot yet.`
        : hasSearchQuery
        ? `No matching ${searchMode === 'all' ? 'memory records' : searchMode.replace('_', ' ')} found right now.`
        : searchMode === 'all' || searchMode === 'facts'
            ? `No matching memories found in the ${memoryType} tier.`
            : `No ${searchMode.replace('_', ' ')} have been captured yet.`;

    const addFactMutation = trpc.memory.addFact.useMutation({
        onSuccess: () => {
            toast.success("Fact added to memory");
            setNewFact('');
            void refreshMemoryViews();
        },
        onError: (err) => {
            toast.error(`Failed to add fact: ${err.message}`);
        }
    });

    const handleAddFact = (e: React.FormEvent) => {
        e.preventDefault();
        if (!newFact.trim()) return;
        addFactMutation.mutate({ content: newFact, type: memoryType === 'long_term' ? 'long_term' : 'working' });
    };

    return (
        <div className="p-8 space-y-8 h-full flex flex-col">
            <div className="flex justify-between items-center shrink-0">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <Brain className="h-8 w-8 text-pink-500" />
                        TormentNexus Memory Control
                    </h1>
                    <p className="text-zinc-500 mt-2">
                        Search and inspect TormentNexus-native facts, observations, prompts, session summaries, and adapter exports from one control surface.
                    </p>
                </div>
                <div className="flex gap-4">
                    <StatCard label="Session" value={(stats as any)?.sessionCount || 0} icon={<History className="h-3 w-3" />} />
                    <StatCard label="Working" value={(stats as any)?.workingCount || 0} icon={<Zap className="h-3 w-3" />} />
                    <StatCard label="Long Term" value={(stats as any)?.longTermCount || 0} icon={<Database className="h-3 w-3" />} />
                    <StatCard label="Observations" value={(stats as any)?.observationCount || 0} icon={<RefreshCw className="h-3 w-3" />} />
                    <StatCard label="Summaries" value={(stats as any)?.sessionSummaryCount || 0} icon={<History className="h-3 w-3" />} />
                    <StatCard label="Prompts" value={(stats as any)?.promptCount || 0} icon={<Brain className="h-3 w-3" />} />
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-4 gap-8 flex-1 min-h-0">
                {/* Sidebar Controls */}
                <div className="lg:col-span-1 space-y-6">
                    <Card className="bg-zinc-900 border-zinc-800">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Filter className="h-4 w-4" />
                                Memory Filters
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <label className="text-[10px] font-bold text-zinc-600 uppercase">Storage Tier</label>
                                <div className="flex flex-col gap-1">
                                    <TierButton active={memoryType === 'session'} onClick={() => setMemoryType('session')} label="Session" description="Transient context" />
                                    <TierButton active={memoryType === 'working'} onClick={() => setMemoryType('working')} label="Working" description="Active task data" />
                                    <TierButton active={memoryType === 'long_term'} onClick={() => setMemoryType('long_term')} label="Long Term" description="Persistent facts" />
                                </div>
                                <p className="text-[11px] leading-relaxed text-zinc-500">
                                    {getMemoryModeHint(searchMode, memoryType)}
                                </p>
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800 border-l-4 border-l-amber-500">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Brain className="h-4 w-4" />
                                TormentNexus Memory Model
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-3 text-xs text-zinc-300">
                            {MEMORY_MODEL_PILLARS.map((pillar) => (
                                <div key={pillar.title} className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                    <p className="text-sm font-medium text-white">{pillar.title}</p>
                                    <p className="mt-1 text-[11px] leading-relaxed text-zinc-400">{pillar.description}</p>
                                </div>
                            ))}
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800 border-l-4 border-l-pink-600">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Plus className="h-4 w-4" />
                                Inject Fact
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <form onSubmit={handleAddFact} className="space-y-3">
                                <textarea
                                    value={newFact}
                                    onChange={e => setNewFact(e.target.value)}
                                    className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2 text-xs text-white h-20 focus:ring-1 focus:ring-pink-500 outline-none resize-none"
                                    placeholder="Manually add a fact to current tier..."
                                />
                                <Button
                                    type="submit"
                                    size="sm"
                                    disabled={addFactMutation.isPending || !newFact.trim()}
                                    className="w-full bg-pink-600 hover:bg-pink-500 text-white text-xs"
                                >
                                    {addFactMutation.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Save className="h-3 w-3 mr-2" />}
                                    Store Fact
                                </Button>
                            </form>
                        </CardContent>
                    </Card>

                    {/* Export / Import Controls (Phase 70) */}
                    <Card className="bg-zinc-900 border-zinc-800 border-l-4 border-l-cyan-600">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Download className="h-4 w-4" />
                                Export / Import
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-3">
                            <div className="space-y-2">
                                <label className="text-[10px] font-bold text-zinc-600 uppercase">Format</label>
                                <select
                                    value={exportFormat}
                                    onChange={e => setExportFormat(e.target.value as MemoryInterchangeFormat)}
                                    className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2 text-xs text-white focus:ring-1 focus:ring-cyan-500 outline-none"
                                >
                                    {MEMORY_FORMAT_OPTIONS.map((option) => (
                                        <option key={option.value} value={option.value}>{option.label}</option>
                                    ))}
                                </select>
                            </div>
                            <Button
                                size="sm"
                                className="w-full bg-cyan-700 hover:bg-cyan-600 text-white text-xs"
                                onClick={async () => {
                                    try {
                                        const res = await fetch(`/api/trpc/memory.exportMemories?input=${encodeURIComponent(JSON.stringify({ userId: 'default', format: exportFormat }))}`);
                                        const json = await res.json();
                                        const content = json?.result?.data?.data || '';
                                        const extension = exportFormat === 'csv' ? 'csv' : exportFormat === 'jsonl' ? 'jsonl' : 'json';
                                        const blob = new Blob([content], { type: extension === 'csv' ? 'text/csv' : 'application/json' });
                                        const url = URL.createObjectURL(blob);
                                        const a = document.createElement('a');
                                        a.href = url;
                                        a.download = `tormentnexus-memories.${extension}`;
                                        a.click();
                                        URL.revokeObjectURL(url);
                                        toast.success(`Exported as ${MEMORY_FORMAT_OPTIONS.find(option => option.value === exportFormat)?.label || exportFormat}`);
                                    } catch (err: any) {
                                        toast.error(`Export failed: ${err.message}`);
                                    }
                                }}
                            >
                                <Download className="h-3 w-3 mr-2" />
                                Download Export
                            </Button>
                            <div className="border-t border-zinc-800 pt-3">
                                <label className="text-[10px] font-bold text-zinc-600 uppercase block mb-2">Import File</label>
                                <input
                                    type="file"
                                    accept=".json,.csv,.jsonl"
                                    className="w-full text-[10px] text-zinc-400 file:mr-2 file:py-1 file:px-2 file:rounded file:border-0 file:text-[10px] file:bg-zinc-800 file:text-zinc-300 hover:file:bg-zinc-700"
                                    onChange={async (e) => {
                                        const file = e.target.files?.[0];
                                        if (!file) return;
                                        setImporting(true);
                                        try {
                                            const text = await file.text();
                                            const res = await fetch('/api/trpc/memory.importMemories', {
                                                method: 'POST',
                                                headers: { 'Content-Type': 'application/json' },
                                                body: JSON.stringify({ userId: 'default', format: exportFormat, data: text })
                                            });
                                            const result = await res.json();
                                            toast.success(`Imported ${result?.result?.data?.imported || 0} memories`);
                                            await refreshMemoryViews();
                                        } catch (err: any) {
                                            toast.error(`Import failed: ${err.message}`);
                                        } finally {
                                            setImporting(false);
                                        }
                                    }}
                                />
                                {importing && (
                                    <div className="flex items-center gap-2 mt-2 text-xs text-cyan-400">
                                        <Loader2 className="h-3 w-3 animate-spin" />
                                        Importing...
                                    </div>
                                )}
                            </div>
                            <div className="border-t border-zinc-800 pt-3 space-y-2">
                                <label className="text-[10px] font-bold text-zinc-600 uppercase block">Convert Export</label>
                                <select
                                    value={convertToFormat}
                                    onChange={e => setConvertToFormat(e.target.value as MemoryInterchangeFormat)}
                                    className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2 text-xs text-white focus:ring-1 focus:ring-cyan-500 outline-none"
                                >
                                    {MEMORY_FORMAT_OPTIONS.filter((option) => option.value !== exportFormat).map((option) => (
                                        <option key={option.value} value={option.value}>{option.label}</option>
                                    ))}
                                </select>
                                <Button
                                    size="sm"
                                    variant="outline"
                                    disabled={converting || convertToFormat === exportFormat}
                                    className="w-full border-zinc-700 hover:bg-zinc-800 text-cyan-300 text-xs"
                                    onClick={async () => {
                                        setConverting(true);
                                        try {
                                            const exportResponse = await fetch(`/api/trpc/memory.exportMemories?input=${encodeURIComponent(JSON.stringify({ userId: 'default', format: exportFormat }))}`);
                                            const exportJson = await exportResponse.json();
                                            const sourceData = exportJson?.result?.data?.data || '';

                                            const convertResponse = await fetch('/api/trpc/memory.convertMemories', {
                                                method: 'POST',
                                                headers: { 'Content-Type': 'application/json' },
                                                body: JSON.stringify({
                                                    userId: 'default',
                                                    fromFormat: exportFormat,
                                                    toFormat: convertToFormat,
                                                    data: sourceData,
                                                }),
                                            });
                                            const convertJson = await convertResponse.json();
                                            const convertedData = convertJson?.result?.data?.data || '';
                                            const extension = convertToFormat === 'csv' ? 'csv' : convertToFormat === 'jsonl' ? 'jsonl' : 'json';
                                            const blob = new Blob([convertedData], { type: extension === 'csv' ? 'text/csv' : 'application/json' });
                                            const url = URL.createObjectURL(blob);
                                            const a = document.createElement('a');
                                            a.href = url;
                                            a.download = `tormentnexus-memory-converted.${extension}`;
                                            a.click();
                                            URL.revokeObjectURL(url);
                                            toast.success(`Converted ${exportFormat} → ${convertToFormat}`);
                                        } catch (err: any) {
                                            toast.error(`Conversion failed: ${err.message}`);
                                        } finally {
                                            setConverting(false);
                                        }
                                    }}
                                >
                                    {converting ? <Loader2 className="h-3 w-3 animate-spin mr-2" /> : <RefreshCw className="h-3 w-3 mr-2" />}
                                    Convert & Download
                                </Button>
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800 border-l-4 border-l-emerald-600">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <RefreshCw className="h-4 w-4" />
                                Runtime Observations
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-3 text-xs text-zinc-300">
                            {recentObservationsQuery.isLoading ? (
                                <div className="flex items-center gap-2 text-zinc-500">
                                    <Loader2 className="h-3 w-3 animate-spin" />
                                    Loading observations…
                                </div>
                            ) : !(recentObservationsQuery.data ?? []).length ? (
                                <p className="text-zinc-500">No runtime observations have been captured yet.</p>
                            ) : (
                                <div className="space-y-2">
                                    {(recentObservationsQuery.data as any[]).map((memory, index) => {
                                        const observation = memory.metadata?.structuredObservation;
                                        return (
                                            <div key={memory.id ?? `observation-${index}`} className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                                <div className="mb-1 flex items-center justify-between gap-2">
                                                    <Badge variant="outline" className="border-emerald-500/30 text-emerald-300">
                                                        {observation?.type ?? 'observation'}
                                                    </Badge>
                                                    {observation?.toolName ? (
                                                        <span className="font-mono text-[10px] text-zinc-500">{observation.toolName}</span>
                                                    ) : null}
                                                </div>
                                                <p className="text-sm font-medium text-white">{observation?.title ?? 'Untitled observation'}</p>
                                                <p className="mt-1 line-clamp-3 whitespace-pre-wrap break-words text-[11px] text-zinc-400">
                                                    {observation?.narrative ?? memory.content}
                                                </p>
                                                <div className="mt-2 flex flex-wrap gap-2 text-[10px] text-zinc-500">
                                                    {memory.metadata?.source ? <span>source: {String(memory.metadata.source)}</span> : null}
                                                    {observation?.facts?.length ? <span>{observation.facts.length} fact{observation.facts.length === 1 ? '' : 's'}</span> : null}
                                                    {observation?.filesRead?.length ? <span>{observation.filesRead.length} read</span> : null}
                                                    {observation?.filesModified?.length ? <span>{observation.filesModified.length} modified</span> : null}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800 border-l-4 border-l-sky-600">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <History className="h-4 w-4" />
                                Session Summaries
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-3 text-xs text-zinc-300">
                            {recentSessionSummariesQuery.isLoading ? (
                                <div className="flex items-center gap-2 text-zinc-500">
                                    <Loader2 className="h-3 w-3 animate-spin" />
                                    Loading session summaries…
                                </div>
                            ) : !(recentSessionSummariesQuery.data ?? []).length ? (
                                <p className="text-zinc-500">No session summaries have been captured yet.</p>
                            ) : (
                                <div className="space-y-2">
                                    {(recentSessionSummariesQuery.data as any[]).map((memory, index) => {
                                        const summary = memory.metadata?.structuredSessionSummary;
                                        return (
                                            <div key={memory.id ?? `summary-${index}`} className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                                <div className="mb-1 flex items-center justify-between gap-2">
                                                    <Badge variant="outline" className="border-sky-500/30 text-sky-300">
                                                        {summary?.status ?? 'summary'}
                                                    </Badge>
                                                    {summary?.cliType ? (
                                                        <span className="font-mono text-[10px] text-zinc-500">{summary.cliType}</span>
                                                    ) : null}
                                                </div>
                                                <p className="text-sm font-medium text-white">{summary?.name ?? summary?.sessionId ?? 'Unnamed session'}</p>
                                                <p className="mt-1 line-clamp-3 whitespace-pre-wrap break-words text-[11px] text-zinc-400">
                                                    {summary?.activeGoal ?? summary?.lastObjective ?? memory.content}
                                                </p>
                                                <div className="mt-2 flex flex-wrap gap-2 text-[10px] text-zinc-500">
                                                    {summary?.sessionId ? <span>session: {summary.sessionId}</span> : null}
                                                    {memory.metadata?.source ? <span>source: {String(memory.metadata.source)}</span> : null}
                                                    {typeof summary?.restartCount === 'number' && summary.restartCount > 0 ? <span>{summary.restartCount} restart{summary.restartCount === 1 ? '' : 's'}</span> : null}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800 border-l-4 border-l-violet-600">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Brain className="h-4 w-4" />
                                Captured Prompts & Goals
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-3 text-xs text-zinc-300">
                            {recentPromptsQuery.isLoading ? (
                                <div className="flex items-center gap-2 text-zinc-500">
                                    <Loader2 className="h-3 w-3 animate-spin" />
                                    Loading prompt history…
                                </div>
                            ) : !(recentPromptsQuery.data ?? []).length ? (
                                <p className="text-zinc-500">No prompt captures have been recorded yet.</p>
                            ) : (
                                <div className="space-y-2">
                                    {(recentPromptsQuery.data as any[]).map((memory, index) => {
                                        const prompt = memory.metadata?.structuredUserPrompt;
                                        return (
                                            <div key={memory.id ?? `prompt-${index}`} className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                                                <div className="mb-1 flex items-center justify-between gap-2">
                                                    <Badge variant="outline" className="border-violet-500/30 text-violet-300">
                                                        {prompt?.role ?? 'prompt'}
                                                    </Badge>
                                                    <span className="text-[10px] text-zinc-500">#{prompt?.promptNumber ?? '?'}</span>
                                                </div>
                                                <p className="line-clamp-3 whitespace-pre-wrap break-words text-[11px] text-zinc-300">
                                                    {prompt?.content ?? memory.content}
                                                </p>
                                                <div className="mt-2 flex flex-wrap gap-2 text-[10px] text-zinc-500">
                                                    {prompt?.sessionId ? <span>session: {prompt.sessionId}</span> : null}
                                                    {memory.metadata?.source ? <span>source: {String(memory.metadata.source)}</span> : null}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>

                {/* Main Results Area */}
                <Card className="lg:col-span-3 bg-zinc-900 border-zinc-800 flex flex-col shadow-2xl overflow-hidden">
                    <CardHeader className="border-b border-white/5 bg-black/20 pb-4">
                        <div className="space-y-4">
                            <div className="flex flex-wrap gap-2">
                                {MEMORY_SEARCH_MODES.map((mode) => (
                                    <button
                                        key={mode.value}
                                        type="button"
                                        onClick={() => setSearchMode(mode.value)}
                                        className={`rounded-full border px-3 py-1.5 text-[11px] font-semibold transition-colors ${searchMode === mode.value
                                            ? 'border-pink-500/60 bg-pink-500/10 text-pink-200'
                                            : 'border-zinc-800 bg-zinc-950 text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'
                                            }`}
                                        title={mode.description}
                                    >
                                        {mode.label}
                                    </button>
                                ))}
                            </div>
                            <p className="text-xs text-zinc-500">
                                {MEMORY_SEARCH_MODES.find((mode) => mode.value === searchMode)?.description}
                            </p>
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
                                <input
                                    type="text"
                                    value={searchQuery}
                                    onChange={e => {
                                        const nextValue = e.target.value;
                                        setSearchQuery(nextValue);

                                        if (activePivot && nextValue.trim() !== activePivot.query) {
                                            setActivePivot(null);
                                        }
                                    }}
                                    placeholder={`Search ${searchMode === 'all' ? `${memoryType} memory records` : searchMode.replace('_', ' ')}...`}
                                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg pl-10 pr-4 py-2.5 text-sm text-white focus:ring-1 focus:ring-pink-500 outline-none transition-all"
                                />
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="flex-1 p-0 overflow-hidden flex flex-col">
                        <ScrollArea className="flex-1">
                            {activeLoading ? (
                                <div className="p-12 flex flex-col items-center justify-center text-zinc-500 gap-3">
                                    <Loader2 className="h-8 w-8 animate-spin" />
                                    <p className="text-sm font-mono uppercase tracking-widest">Accessing Synapses...</p>
                                </div>
                            ) : !activeResults || activeResults.length === 0 ? (
                                <div className="p-20 text-center text-zinc-600">
                                    <Brain className="h-12 w-12 mx-auto mb-4 opacity-10" />
                                    <p className="text-lg font-medium">Tabula Rasa</p>
                                    <p className="text-sm mt-1">{activeEmptyMessage}</p>
                                </div>
                            ) : (
                                <div className="grid min-h-full lg:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
                                    <div className="border-r border-white/5">
                                        {timelineGroups.map((group) => (
                                            <div key={group.key} className="border-b border-white/5 last:border-b-0">
                                                <div className="sticky top-0 z-10 border-b border-white/5 bg-zinc-950/95 px-4 py-2 text-[10px] font-bold uppercase tracking-[0.2em] text-zinc-500 backdrop-blur">
                                                    {group.label}
                                                </div>
                                                <div className="divide-y divide-white/5">
                                                    {group.items.map((memory) => {
                                                        const recordKey = getMemoryRecordKey(memory);
                                                        const provenance = getMemoryProvenance(memory);
                                                        const preview = getMemoryPreview(memory);
                                                        const isSelected = selectedMemory ? getMemoryRecordKey(selectedMemory) === recordKey : false;

                                                        return (
                                                            <button
                                                                key={recordKey}
                                                                type="button"
                                                                onClick={() => setSelectedRecordKey(recordKey)}
                                                                className={`group w-full px-4 py-4 text-left transition-colors ${isSelected
                                                                    ? 'bg-pink-500/10'
                                                                    : 'hover:bg-white/[0.02]'
                                                                    }`}
                                                            >
                                                                <div className="flex items-start justify-between gap-3">
                                                                    <div className="min-w-0 flex-1">
                                                                        <div className="mb-2 flex flex-wrap items-center gap-2">
                                                                            <Badge variant="outline" className={`text-[9px] uppercase font-bold tracking-tighter ${isSelected ? 'border-pink-500/60 text-pink-200' : 'border-zinc-700 text-zinc-500'}`}>
                                                                                {getMemoryBadgeLabel(memory)}
                                                                            </Badge>
                                                                            <span className="text-[10px] font-mono text-zinc-600">
                                                                                {new Date(getMemoryTimestamp(memory)).toLocaleTimeString([], {
                                                                                    hour: '2-digit',
                                                                                    minute: '2-digit',
                                                                                })}
                                                                            </span>
                                                                        </div>
                                                                        <p className="truncate text-sm font-semibold text-white">
                                                                            {getMemoryTitle(memory)}
                                                                        </p>
                                                                        <p className="mt-2 line-clamp-3 whitespace-pre-wrap break-words text-xs leading-relaxed text-zinc-400">
                                                                            {preview}
                                                                        </p>
                                                                        {provenance.length ? (
                                                                            <div className="mt-2 flex flex-wrap items-center gap-2 text-[10px] text-zinc-500 font-mono">
                                                                                {provenance.map((token) => (
                                                                                    <span key={`${recordKey}-${token}`}>{token}</span>
                                                                                ))}
                                                                            </div>
                                                                        ) : null}
                                                                    </div>
                                                                    <div className="flex shrink-0 flex-col items-end gap-2">
                                                                        <ChevronRight className={`h-4 w-4 transition-colors ${isSelected ? 'text-pink-300' : 'text-zinc-700 group-hover:text-zinc-400'}`} />
                                                                        <Badge variant="secondary" className="text-[9px] text-zinc-300">
                                                                            {(memory.score || 1).toFixed(2)}
                                                                        </Badge>
                                                                    </div>
                                                                </div>
                                                            </button>
                                                        );
                                                    })}
                                                </div>
                                            </div>
                                        ))}
                                    </div>

                                    <div className="bg-black/10">
                                        {selectedMemory ? (
                                            <div className="flex h-full flex-col">
                                                <div className="border-b border-white/5 px-5 py-4">
                                                    <div className="flex flex-wrap items-center gap-2">
                                                        <Badge variant="outline" className="border-pink-500/40 text-pink-200 text-[9px] uppercase tracking-tighter">
                                                            {getMemoryBadgeLabel(selectedMemory)}
                                                        </Badge>
                                                        <span className="text-[10px] font-mono text-zinc-600">
                                                            {new Date(getMemoryTimestamp(selectedMemory)).toLocaleString()}
                                                        </span>
                                                    </div>
                                                    <h2 className="mt-3 text-lg font-semibold text-white">
                                                        {getMemoryTitle(selectedMemory)}
                                                    </h2>
                                                    <p className="mt-2 whitespace-pre-wrap text-sm leading-relaxed text-zinc-300">
                                                        {getMemoryPreview(selectedMemory)}
                                                    </p>
                                                    {getMemoryProvenance(selectedMemory).length ? (
                                                        <div className="mt-3 flex flex-wrap gap-2">
                                                            {getMemoryProvenance(selectedMemory).map((token) => (
                                                                <Badge key={`${getMemoryRecordKey(selectedMemory)}-${token}`} variant="secondary" className="bg-zinc-800 text-[10px] text-zinc-300">
                                                                    {token}
                                                                </Badge>
                                                            ))}
                                                        </div>
                                                    ) : null}
                                                </div>

                                                <div className="flex-1 space-y-4 p-5">
                                                    {pivotSections.length ? (
                                                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-4">
                                                            <h3 className="text-xs font-bold uppercase tracking-widest text-zinc-500">
                                                                Pivot this record
                                                            </h3>
                                                            <div className="mt-3 space-y-4">
                                                                {pivotSections.map((section) => (
                                                                    <div key={`${getMemoryRecordKey(selectedMemory)}-${section.title}`}>
                                                                        <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-zinc-600">
                                                                            {section.title}
                                                                        </p>
                                                                        <div className="mt-2 flex flex-wrap gap-2">
                                                                            {section.actions.map((action) => (
                                                                                <button
                                                                                    key={action.key}
                                                                                    type="button"
                                                                                    onClick={() => handlePivotAction(action)}
                                                                                    className="max-w-full rounded-md border border-zinc-700 bg-black/20 px-3 py-2 text-left text-xs text-zinc-200 transition-colors hover:border-pink-500/50 hover:bg-pink-500/10"
                                                                                    title={action.description}
                                                                                >
                                                                                    <span className="block truncate font-mono">{action.label}</span>
                                                                                    <span className="mt-1 block text-[10px] uppercase tracking-widest text-zinc-500">
                                                                                        {action.mode === 'all' ? 'all records' : action.mode.replace('_', ' ')}
                                                                                    </span>
                                                                                </button>
                                                                            ))}
                                                                        </div>
                                                                    </div>
                                                                ))}
                                                            </div>
                                                        </div>
                                                    ) : null}

                                                    {getMemoryDetailSections(selectedMemory).map((section) => (
                                                        <div key={`${getMemoryRecordKey(selectedMemory)}-${section.title}`} className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-4">
                                                            <h3 className="text-xs font-bold uppercase tracking-widest text-zinc-500">
                                                                {section.title}
                                                            </h3>
                                                            {section.body ? (
                                                                <p className="mt-3 whitespace-pre-wrap break-words text-sm leading-relaxed text-zinc-300">
                                                                    {section.body}
                                                                </p>
                                                            ) : null}
                                                            {section.items?.length ? (
                                                                <ul className="mt-3 space-y-2 text-sm text-zinc-300">
                                                                    {section.items.map((item) => (
                                                                        <li key={`${section.title}-${item}`} className="rounded-md border border-zinc-800 bg-black/20 px-3 py-2 font-mono text-xs text-zinc-300">
                                                                            {item}
                                                                        </li>
                                                                    ))}
                                                                </ul>
                                                            ) : null}
                                                        </div>
                                                    ))}

                                                    {relatedRecords.length ? (
                                                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-4">
                                                            <h3 className="text-xs font-bold uppercase tracking-widest text-zinc-500">
                                                                Related records
                                                            </h3>
                                                            <div className="mt-3 space-y-3">
                                                                {relatedRecords.map((related) => {
                                                                    const recordKey = getMemoryRecordKey(related.memory);
                                                                    return (
                                                                        <button
                                                                            key={recordKey}
                                                                            type="button"
                                                                            onClick={() => setSelectedRecordKey(recordKey)}
                                                                            className="w-full rounded-lg border border-zinc-800 bg-black/20 px-3 py-3 text-left transition-colors hover:bg-zinc-900"
                                                                        >
                                                                            <div className="flex items-start justify-between gap-3">
                                                                                <div className="min-w-0 flex-1">
                                                                                    <div className="flex flex-wrap items-center gap-2">
                                                                                        <Badge variant="outline" className="border-zinc-700 text-[9px] uppercase tracking-tighter text-zinc-400">
                                                                                            {getMemoryBadgeLabel(related.memory)}
                                                                                        </Badge>
                                                                                        <span className="text-[10px] font-mono text-zinc-600">
                                                                                            {new Date(getMemoryTimestamp(related.memory)).toLocaleString()}
                                                                                        </span>
                                                                                    </div>
                                                                                    <p className="mt-2 truncate text-sm font-medium text-white">
                                                                                        {getMemoryTitle(related.memory)}
                                                                                    </p>
                                                                                    <p className="mt-2 line-clamp-2 whitespace-pre-wrap break-words text-xs text-zinc-400">
                                                                                        {getMemoryPreview(related.memory)}
                                                                                    </p>
                                                                                    <div className="mt-2 flex flex-wrap gap-2">
                                                                                        {related.reasons.map((reason) => (
                                                                                            <Badge key={`${recordKey}-${reason}`} variant="secondary" className="bg-zinc-800 text-[10px] text-zinc-300">
                                                                                                {reason}
                                                                                            </Badge>
                                                                                        ))}
                                                                                    </div>
                                                                                </div>
                                                                                <Badge variant="secondary" className="text-[9px] text-zinc-300">
                                                                                    {related.score}
                                                                                </Badge>
                                                                            </div>
                                                                        </button>
                                                                    );
                                                                })}
                                                            </div>
                                                        </div>
                                                    ) : null}

                                                    {selectedSessionId ? (
                                                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-4">
                                                            <h3 className="text-xs font-bold uppercase tracking-widest text-zinc-500">
                                                                Session window
                                                            </h3>
                                                            <p className="mt-2 text-[11px] leading-relaxed text-zinc-500">
                                                                Nearby same-session records around the currently selected memory anchor.
                                                            </p>

                                                            {timelineWindowQuery.isLoading ? (
                                                                <div className="mt-3 flex items-center gap-2 text-xs text-zinc-500">
                                                                    <Loader2 className="h-3 w-3 animate-spin" />
                                                                    Loading session timeline…
                                                                </div>
                                                            ) : sessionWindowGroups.length ? (
                                                                <div className="mt-3 space-y-4">
                                                                    {sessionWindowGroups.map((group) => (
                                                                        <div key={group.key}>
                                                                            <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-zinc-600">
                                                                                {group.label}
                                                                            </p>
                                                                            <div className="mt-2 space-y-3">
                                                                                {group.items.map((memory) => {
                                                                                    const recordKey = getMemoryRecordKey(memory);

                                                                                    return (
                                                                                        <button
                                                                                            key={recordKey}
                                                                                            type="button"
                                                                                            onClick={() => setSelectedRecordKey(recordKey)}
                                                                                            className="w-full rounded-lg border border-zinc-800 bg-black/20 px-3 py-3 text-left transition-colors hover:bg-zinc-900"
                                                                                        >
                                                                                            <div className="flex items-start justify-between gap-3">
                                                                                                <div className="min-w-0 flex-1">
                                                                                                    <div className="flex flex-wrap items-center gap-2">
                                                                                                        <Badge variant="outline" className="border-zinc-700 text-[9px] uppercase tracking-tighter text-zinc-400">
                                                                                                            {getMemoryBadgeLabel(memory)}
                                                                                                        </Badge>
                                                                                                        <span className="text-[10px] font-mono text-zinc-600">
                                                                                                            {new Date(getMemoryTimestamp(memory)).toLocaleString()}
                                                                                                        </span>
                                                                                                    </div>
                                                                                                    <p className="mt-2 truncate text-sm font-medium text-white">
                                                                                                        {getMemoryTitle(memory)}
                                                                                                    </p>
                                                                                                    <p className="mt-2 line-clamp-2 whitespace-pre-wrap break-words text-xs text-zinc-400">
                                                                                                        {getMemoryPreview(memory)}
                                                                                                    </p>
                                                                                                </div>
                                                                                            </div>
                                                                                        </button>
                                                                                    );
                                                                                })}
                                                                            </div>
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                            ) : (
                                                                <p className="mt-3 text-xs text-zinc-500">
                                                                    No nearby same-session records were found around this memory yet.
                                                                </p>
                                                            )}
                                                        </div>
                                                    ) : null}

                                                    {selectedMemory?.id ? (
                                                        <div className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-4">
                                                            <h3 className="text-xs font-bold uppercase tracking-widest text-zinc-500">
                                                                Cross-session links
                                                            </h3>
                                                            <p className="mt-2 text-[11px] leading-relaxed text-zinc-500">
                                                                Related records from other sessions that overlap by concepts, files, tools, or source.
                                                            </p>

                                                            {crossSessionLinksQuery.isLoading ? (
                                                                <div className="mt-3 flex items-center gap-2 text-xs text-zinc-500">
                                                                    <Loader2 className="h-3 w-3 animate-spin" />
                                                                    Finding related sessions…
                                                                </div>
                                                            ) : crossSessionLinks.length ? (
                                                                <div className="mt-3 space-y-3">
                                                                    {crossSessionLinks.map((related) => {
                                                                        const recordKey = getMemoryRecordKey(related.memory);
                                                                        return (
                                                                            <button
                                                                                key={recordKey}
                                                                                type="button"
                                                                                onClick={() => setSelectedRecordKey(recordKey)}
                                                                                className="w-full rounded-lg border border-zinc-800 bg-black/20 px-3 py-3 text-left transition-colors hover:bg-zinc-900"
                                                                            >
                                                                                <div className="flex items-start justify-between gap-3">
                                                                                    <div className="min-w-0 flex-1">
                                                                                        <div className="flex flex-wrap items-center gap-2">
                                                                                            <Badge variant="outline" className="border-zinc-700 text-[9px] uppercase tracking-tighter text-zinc-400">
                                                                                                {getMemoryBadgeLabel(related.memory)}
                                                                                            </Badge>
                                                                                            <span className="text-[10px] font-mono text-zinc-600">
                                                                                                {new Date(getMemoryTimestamp(related.memory)).toLocaleString()}
                                                                                            </span>
                                                                                        </div>
                                                                                        <p className="mt-2 truncate text-sm font-medium text-white">
                                                                                            {getMemoryTitle(related.memory)}
                                                                                        </p>
                                                                                        <p className="mt-2 line-clamp-2 whitespace-pre-wrap break-words text-xs text-zinc-400">
                                                                                            {getMemoryPreview(related.memory)}
                                                                                        </p>
                                                                                        <div className="mt-2 flex flex-wrap gap-2">
                                                                                            {related.reasons.map((reason) => (
                                                                                                <Badge key={`${recordKey}-${reason}`} variant="secondary" className="bg-zinc-800 text-[10px] text-zinc-300">
                                                                                                    {reason}
                                                                                                </Badge>
                                                                                            ))}
                                                                                        </div>
                                                                                    </div>
                                                                                    <Badge variant="secondary" className="text-[9px] text-zinc-300">
                                                                                        {related.score}
                                                                                    </Badge>
                                                                                </div>
                                                                            </button>
                                                                        );
                                                                    })}
                                                                </div>
                                                            ) : (
                                                                <p className="mt-3 text-xs text-zinc-500">
                                                                    No cross-session links were found for this memory yet.
                                                                </p>
                                                            )}
                                                        </div>
                                                    ) : null}
                                                </div>
                                            </div>
                                        ) : null}
                                    </div>
                                </div>
                            )}
                        </ScrollArea>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}

function StatCard({ label, value, icon }: { label: string, value: number, icon: React.ReactNode }) {
    return (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-2 flex flex-col items-end min-w-[100px]">
            <div className="flex items-center gap-1.5 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">
                {icon} {label}
            </div>
            <div className="text-xl font-bold text-white tabular-nums">{value}</div>
        </div>
    );
}

function TierButton({ active, onClick, label, description }: { active: boolean, onClick: () => void, label: string, description: string }) {
    return (
        <button
            onClick={onClick}
            className={`w-full text-left p-3 rounded-lg border transition-all ${active
                    ? 'bg-pink-500/10 border-pink-500/50 text-pink-400 shadow-[0_0_15px_rgba(236,72,153,0.05)]'
                    : 'bg-transparent border-transparent text-zinc-500 hover:bg-white/5'
                }`}
        >
            <div className="text-sm font-bold">{label}</div>
            <div className={`text-[10px] mt-0.5 ${active ? 'text-pink-400/60' : 'text-zinc-600'}`}>{description}</div>
        </button>
    );
}
