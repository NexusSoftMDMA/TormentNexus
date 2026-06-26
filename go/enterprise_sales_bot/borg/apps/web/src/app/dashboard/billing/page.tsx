"use client";

import React, { useState } from 'react';
import { trpc } from '@/utils/trpc';
import { Card, CardHeader, CardTitle, CardContent, Button, Badge } from '@tormentnexus/ui';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, ResponsiveContainer } from 'recharts';
import { Loader2, DollarSign, Activity, Settings, Key, Zap, AlertCircle, Database, Shield, ExternalLink, WalletCards } from 'lucide-react';
import { toast } from 'sonner';
import {
    formatRoutingStrategyLabel,
    formatTaskRoutingLabel,
    getPortalBadgeClasses,
    getProviderPortalCards,
    getProviderQuickAccessSections,
    getRoutingStrategyBadgeClasses,
    ROUTING_STRATEGY_OPTIONS,
    type BillingRoutingStrategy,
    type BillingTaskRoutingRuleSummary,
    type BillingProviderQuotaSummary,
} from './billing-portal-data';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@tormentnexus/ui';
import { Input } from '@tormentnexus/ui';
import {
    getBillingUsageSummary,
    getDefaultRoutingStrategy,
    getFallbackTaskType,
    normalizeBillingPricingModels,
    normalizeBillingQuotaRows,
    normalizeFallbackChain,
    normalizeTaskRoutingRules,
} from './billing-page-normalizers';

const FALLBACK_TASK_OPTIONS: BillingTaskRoutingRuleSummary['taskType'][] = ['general', 'coding', 'planning', 'research', 'worker', 'supervisor'];

export default function ProviderAuthBillingMatrix() {
    const [historyDays, setHistoryDays] = useState(30);
    const [fallbackTaskType, setFallbackTaskType] = useState<BillingTaskRoutingRuleSummary['taskType']>('general');
    
    // Key update dialog state
    const [activePortalId, setActivePortalId] = useState<string | null>(null);
    const [activePortalName, setActivePortalName] = useState<string>('');
    const [newKeyValue, setNewKeyValue] = useState<string>('');
    
    const utils = trpc.useUtils();

    const { data: status, isLoading: isStatusLoading } = trpc.billing.getStatus.useQuery();
    const { data: quotas, isLoading: isQuotasLoading } = trpc.billing.getProviderQuotas.useQuery();
    const { data: costHistory, isLoading: isHistoryLoading } = trpc.billing.getCostHistory.useQuery({ days: historyDays });
    const { data: pricing, isLoading: isPricingLoading } = trpc.billing.getModelPricing.useQuery();
    const { data: fallback, isLoading: isFallbackLoading } = trpc.billing.getFallbackChain.useQuery({ taskType: fallbackTaskType });
    const { data: taskRouting, isLoading: isTaskRoutingLoading } = trpc.billing.getTaskRoutingRules.useQuery();
    const setRoutingStrategyMutation = trpc.billing.setRoutingStrategy.useMutation({
        onSuccess: async () => {
            toast.success('Default provider routing updated');
            await utils.billing.getTaskRoutingRules.invalidate();
        },
        onError: (error) => {
            toast.error(`Routing update failed: ${error.message}`);
        },
    });
    const setTaskRoutingRuleMutation = trpc.billing.setTaskRoutingRule.useMutation({
        onSuccess: async (_result, variables) => {
            if (variables) {
                toast.success(`${formatTaskRoutingLabel(variables.taskType)} routing updated`);
            }
            await utils.billing.getTaskRoutingRules.invalidate();
        },
        onError: (error) => {
            toast.error(`Task routing update failed: ${error.message}`);
        },
    });

    const updateKeyMutation = trpc.settings.updateProviderKey.useMutation({
        onSuccess: async (result) => {
            toast.success(`Key updated securely to ${result.updatedKey}`);
            // Also attempt to immediately test the connection
            if (activePortalId) {
                testConnectionMutation.mutate({ provider: activePortalId });
            }
            await utils.billing.getProviderQuotas.invalidate();
            setActivePortalId(null);
            setNewKeyValue('');
        },
        onError: (error) => {
            toast.error(`Failed to update key: ${error.message}`);
        }
    });

    const testConnectionMutation = trpc.settings.testConnection.useMutation({
        onSuccess: async (result) => {
            if (result.success) {
                toast.success(`Connection test successful! (${result.latencyMs}ms)`);
            } else {
                toast.error(`Connection failed: ${result.error}`);
            }
            // invalidate quotas to refresh connected status badge
            await utils.billing.getProviderQuotas.invalidate();
        }
    });

    const providerPortalCards = getProviderPortalCards(quotas as BillingProviderQuotaSummary[] | undefined);
    const providerQuickAccessSections = getProviderQuickAccessSections(quotas as BillingProviderQuotaSummary[] | undefined);
    const usageSummary = getBillingUsageSummary(status);
    const quotaRows = normalizeBillingQuotaRows(quotas);
    const fallbackChain = normalizeFallbackChain(fallback);
    const fallbackSelectedTaskType = getFallbackTaskType(fallback, fallbackTaskType);
    const defaultRoutingStrategy = getDefaultRoutingStrategy(taskRouting);
    const routingRules = normalizeTaskRoutingRules(taskRouting);
    const pricingModels = normalizeBillingPricingModels(pricing);
    const activeRoutingMutationTask = setTaskRoutingRuleMutation.variables && 'taskType' in setTaskRoutingRuleMutation.variables
        ? setTaskRoutingRuleMutation.variables.taskType
        : undefined;

    const handleDefaultStrategyChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        setRoutingStrategyMutation.mutate({ strategy: event.target.value as BillingRoutingStrategy });
    };

    const handleTaskStrategyChange = (taskType: BillingTaskRoutingRuleSummary['taskType'], event: React.ChangeEvent<HTMLSelectElement>) => {
        const nextValue = event.target.value as BillingRoutingStrategy | 'default';
        setTaskRoutingRuleMutation.mutate({
            taskType,
            strategy: nextValue === 'default' ? null : nextValue,
        });
    };

    const handleSaveKey = () => {
        if (!activePortalId || !newKeyValue.trim()) return;
        updateKeyMutation.mutate({ provider: activePortalId, key: newKeyValue });
    };

    const renderCostChart = () => {
        if (isHistoryLoading) return <div className="h-48 flex items-center justify-center"><Loader2 className="w-6 h-6 animate-spin text-zinc-500" /></div>;
        if (!costHistory?.history || costHistory.history.length === 0) return <div className="h-48 flex items-center justify-center text-zinc-500">No cost history data.</div>;

        return (
            <div className="h-64 w-full mt-4">
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={costHistory.history} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                        <defs>
                            <linearGradient id="colorCost" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                                <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#ffffff10" />
                        <XAxis dataKey="date" stroke="#ffffff50" fontSize={10} tickMargin={10} />
                        <YAxis stroke="#ffffff50" fontSize={10} tickFormatter={(val) => `$${val}`} />
                        <RechartsTooltip
                            contentStyle={{ backgroundColor: '#18181b', borderColor: '#27272a', borderRadius: '8px', fontSize: '12px' }}
                            itemStyle={{ color: '#10b981' }}
                            formatter={(value: number) => [`$${value.toFixed(4)}`, 'Estimated Cost']}
                        />
                        <Area type="monotone" dataKey="cost" stroke="#10b981" strokeWidth={2} fillOpacity={1} fill="url(#colorCost)" />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        );
    };

    return (
        <div className="p-8 max-w-[1600px] mx-auto space-y-8">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <Database className="h-8 w-8 text-emerald-500" />
                        Provider Auth & Billing Matrix
                    </h1>
                    <p className="text-zinc-500 mt-2">
                        Comprehensive overview of AI model quotas, pricing, and system authentication keys.
                    </p>
                </div>
                <div className="flex gap-3">
                    <Button
                        variant="outline"
                        className="border-emerald-500/20 text-emerald-400 bg-emerald-500/5 hover:bg-emerald-500/10 transition-colors"
                        onClick={() => document.getElementById('provider-portals')?.scrollIntoView({ behavior: 'smooth', block: 'start' })}
                    >
                        <DollarSign className="w-4 h-4 mr-2" />
                        Open Provider Portals
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
                {/* Left Column - Financial Overview & Fallback */}
                <div className="xl:col-span-1 space-y-6">
                    {/* Current Usage Card */}
                    <Card className="bg-zinc-900 border-zinc-800 shadow-xl relative overflow-hidden">
                        <div className="absolute top-0 right-0 w-32 h-32 bg-emerald-500/5 blur-3xl -mr-10 -mt-10 rounded-full" />
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Activity className="h-4 w-4 text-emerald-400" />
                                Current Sprint Usage
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-end gap-2 mb-2">
                                <span className="text-4xl font-mono text-white font-bold">
                                    ${isStatusLoading ? '0.00' : usageSummary.currentMonth.toFixed(2)}
                                </span>
                                <span className="text-sm text-zinc-500 mb-1 font-mono">
                                    / ${isStatusLoading ? '0.00' : usageSummary.limit.toFixed(2)} Limit
                                </span>
                            </div>

                            {/* Simple usage bar */}
                            <div className="w-full h-2 bg-zinc-950 rounded-full overflow-hidden mt-4">
                                <div
                                    className="h-full bg-emerald-500 transition-all duration-1000"
                                    style={{ width: `${Math.min(100, ((usageSummary.currentMonth / (usageSummary.limit || 100)) * 100))}%` }}
                                />
                            </div>

                            <div className="mt-6 space-y-3">
                                <div className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2">Cost Breakdown</div>
                                {usageSummary.breakdown.map((item, i: number) => (
                                    <div key={i} className="flex justify-between items-center text-sm">
                                        <span className="text-zinc-300 capitalize flex items-center gap-2">
                                            {item.provider}
                                        </span>
                                        <div className="flex items-center gap-4">
                                            <span className="text-xs text-zinc-500 font-mono">{item.requests} reqs</span>
                                            <span className="font-mono text-emerald-400">${item.cost.toFixed(4)}</span>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </CardContent>
                    </Card>

                    {/* Routing Fallback Chain */}
                    <Card className="bg-zinc-900 border-zinc-800 shadow-xl">
                        <CardHeader className="pb-2">
                            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                                <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                    <Zap className="h-4 w-4 text-amber-500" />
                                    Execution Fallback Chain
                                </CardTitle>
                                <select
                                    value={fallbackTaskType}
                                    onChange={(event) => setFallbackTaskType(event.target.value as BillingTaskRoutingRuleSummary['taskType'])}
                                    className="rounded-md border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-xs text-zinc-200 outline-none focus:border-amber-500"
                                    aria-label="Inspect fallback chain for task type"
                                >
                                    {FALLBACK_TASK_OPTIONS.map((taskType) => (
                                        <option key={taskType} value={taskType}>{formatTaskRoutingLabel(taskType)}</option>
                                    ))}
                                </select>
                            </div>
                        </CardHeader>
                        <CardContent className="pt-2">
                            {isFallbackLoading ? (
                                <div className="flex justify-center py-6"><Loader2 className="w-6 h-6 animate-spin text-zinc-500" /></div>
                            ) : (
                                <div className="space-y-3">
                                    <div className="rounded-lg border border-zinc-800/60 bg-black/30 px-3 py-2 text-[11px] text-zinc-500">
                                        Ranked providers for <span className="font-semibold text-zinc-300">{formatTaskRoutingLabel(fallbackSelectedTaskType)}</span> work.
                                    </div>
                                    {fallbackChain.length ? fallbackChain.map((link, idx: number) => (
                                        <div key={idx} className="flex items-center gap-3 p-3 rounded-lg bg-black/40 border border-zinc-800/50">
                                            <div className="w-6 h-6 rounded-full bg-amber-500/10 text-amber-500 flex items-center justify-center font-bold text-xs shrink-0 border border-amber-500/20">
                                                {link.priority}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-center gap-2">
                                                    <span className="font-bold text-zinc-200 capitalize text-sm truncate">{link.provider}</span>
                                                    {link.model ? <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-zinc-800 text-zinc-400 border-zinc-700 truncate">{link.model}</Badge> : null}
                                                </div>
                                                <div className="text-xs text-zinc-500 mt-0.5 truncate">{link.reason}</div>
                                            </div>
                                        </div>
                                    )) : (
                                        <div className="rounded-lg border border-dashed border-zinc-800 bg-black/20 px-3 py-4 text-sm text-zinc-500">
                                            No ranked providers are currently available for {formatTaskRoutingLabel(fallbackTaskType).toLowerCase()} work.
                                        </div>
                                    )}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800 shadow-xl">
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Settings className="h-4 w-4 text-cyan-400" />
                                Task Routing Matrix
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="pt-2">
                            {isTaskRoutingLoading ? (
                                <div className="flex justify-center py-6"><Loader2 className="w-6 h-6 animate-spin text-zinc-500" /></div>
                            ) : (
                                <div className="space-y-3">
                                    <div className="rounded-lg border border-zinc-800/60 bg-black/30 px-3 py-3 text-xs text-zinc-500 space-y-3">
                                        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                                            <div>
                                                Default routing strategy: <span className="font-semibold text-zinc-300">{formatRoutingStrategyLabel(taskRouting?.defaultStrategy ?? 'best')}</span>
                                            </div>
                                            <div className="flex items-center gap-2">
                                                {setRoutingStrategyMutation.isPending ? <Loader2 className="h-3.5 w-3.5 animate-spin text-cyan-400" /> : null}
                                                <select
                                                    value={defaultRoutingStrategy}
                                                    onChange={handleDefaultStrategyChange}
                                                    disabled={setRoutingStrategyMutation.isPending || setTaskRoutingRuleMutation.isPending}
                                                    className="rounded-md border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-xs text-zinc-200 outline-none focus:border-cyan-500"
                                                    aria-label="Default provider routing strategy"
                                                >
                                                    {ROUTING_STRATEGY_OPTIONS.map((option) => (
                                                        <option key={option.value} value={option.value}>{option.label}</option>
                                                    ))}
                                                </select>
                                            </div>
                                        </div>
                                        <div className="text-[11px] text-zinc-500">
                                            Changes apply to the next model-selection decision immediately, so you can tune cost vs quality without restarting TormentNexus.
                                        </div>
                                    </div>
                                    {routingRules.map((rule) => (
                                        <div key={rule.taskType} className="rounded-lg border border-zinc-800/50 bg-black/40 p-3">
                                            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                                                <div className="flex items-center justify-between gap-3">
                                                    <span className="font-semibold text-zinc-200">{formatTaskRoutingLabel(rule.taskType)}</span>
                                                    <Badge variant="outline" className={`text-[10px] capitalize ${getRoutingStrategyBadgeClasses(rule.strategy)}`}>
                                                        {rule.strategy}
                                                    </Badge>
                                                </div>
                                                <div className="flex items-center gap-2">
                                                    {setTaskRoutingRuleMutation.isPending && activeRoutingMutationTask === rule.taskType ? <Loader2 className="h-3.5 w-3.5 animate-spin text-cyan-400" /> : null}
                                                    <select
                                                        value={rule.strategy}
                                                        onChange={(event) => handleTaskStrategyChange(rule.taskType, event)}
                                                        disabled={setRoutingStrategyMutation.isPending || setTaskRoutingRuleMutation.isPending}
                                                        className="rounded-md border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-xs text-zinc-200 outline-none focus:border-cyan-500"
                                                        aria-label={`${formatTaskRoutingLabel(rule.taskType)} routing strategy`}
                                                    >
                                                        {ROUTING_STRATEGY_OPTIONS.map((option) => (
                                                            <option key={`${rule.taskType}-${option.value}`} value={option.value}>{option.label}</option>
                                                        ))}
                                                    </select>
                                                </div>
                                            </div>
                                            <div className="mt-3 flex flex-wrap gap-2">
                                                {rule.fallbackPreview.length > 0 ? rule.fallbackPreview.map((candidate, index) => (
                                                    <div key={`${rule.taskType}-${candidate.provider}-${candidate.model ?? index}`} className="rounded-md border border-zinc-800 bg-zinc-950/80 px-2.5 py-2 text-xs text-zinc-300">
                                                        <div className="font-medium capitalize">{candidate.provider}</div>
                                                        {candidate.model ? <div className="mt-0.5 font-mono text-[10px] text-zinc-500">{candidate.model}</div> : null}
                                                        {candidate.reason ? <div className="mt-1 text-[10px] uppercase tracking-wide text-zinc-500">{candidate.reason.replace(/_/g, ' ')}</div> : null}
                                                    </div>
                                                )) : (
                                                    <span className="text-xs text-zinc-500">No ranked providers available for this task yet.</span>
                                                )}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>

                {/* Middle/Right Columns - Charts & Matrices */}
                <div className="xl:col-span-2 space-y-6">
                    {/* Cost History */}
                    <Card className="bg-zinc-900 border-zinc-800 shadow-xl">
                        <CardHeader className="pb-0 flex flex-row items-center justify-between">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Activity className="h-4 w-4 text-emerald-500" />
                                30-Day Cost Trend
                            </CardTitle>
                            <div className="flex gap-2">
                                {[7, 14, 30].map(days => (
                                    <button
                                        key={days}
                                        onClick={() => setHistoryDays(days)}
                                        className={`text-xs px-2 py-1 rounded font-medium transition-colors ${historyDays === days ? 'bg-emerald-500/20 text-emerald-400' : 'bg-zinc-800 text-zinc-500 hover:text-zinc-300'}`}
                                    >
                                        {days}D
                                    </button>
                                ))}
                            </div>
                        </CardHeader>
                        <CardContent>
                            {renderCostChart()}
                        </CardContent>
                    </Card>

                    {/* Unified Auth & Quota Matrix */}
                    <Card className="bg-zinc-900 border-zinc-800 shadow-xl overflow-hidden">
                        <CardHeader className="bg-black/20 border-b border-white/5 pb-4">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Shield className="h-4 w-4 text-blue-400" />
                                Provider Capabilities & Limits
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="p-0 overflow-x-auto">
                            <table className="w-full text-sm text-left whitespace-nowrap">
                                <thead className="text-xs text-zinc-500 uppercase bg-black/40 border-b border-zinc-800">
                                    <tr>
                                        <th className="px-6 py-4 font-bold tracking-wider">Provider</th>
                                        <th className="px-6 py-4 font-bold tracking-wider text-center">Auth Status</th>
                                        <th className="px-6 py-4 font-bold tracking-wider">Tier</th>
                                        <th className="px-6 py-4 font-bold tracking-wider text-right">Quota Used</th>
                                        <th className="px-6 py-4 font-bold tracking-wider text-right">Rate Limit</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-zinc-800/50">
                                    {isQuotasLoading ? (
                                        <tr><td colSpan={5} className="px-6 py-12 text-center text-zinc-500"><Loader2 className="w-6 h-6 animate-spin mx-auto" /></td></tr>
                                    ) : quotaRows.map((q) => (
                                        <tr key={q.provider} className="hover:bg-white/[0.02] transition-colors">
                                            <td className="px-6 py-4 font-medium text-zinc-200 capitalize">
                                                <div>
                                                    <div>{q.name}</div>
                                                    <div className="mt-1 text-[10px] uppercase tracking-wide text-zinc-500">{(q.availability ?? 'unknown').replace(/_/g, ' ')}</div>
                                                    {q.lastError ? (
                                                        <div className="mt-1 flex items-center gap-1 text-[10px] text-amber-400">
                                                            <AlertCircle className="h-3 w-3" />
                                                            <span className="truncate max-w-[18rem]">{q.lastError}</span>
                                                        </div>
                                                    ) : null}
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 text-center">
                                                <div className="flex flex-col items-center gap-1">
                                                    {q.authenticated ? (
                                                        <Badge variant="outline" className="bg-emerald-500/10 text-emerald-400 border-emerald-500/20 text-[10px]">CONNECTED</Badge>
                                                    ) : q.configured ? (
                                                        <Badge variant="outline" className="bg-amber-500/10 text-amber-300 border-amber-500/20 text-[10px]">CONFIGURED</Badge>
                                                    ) : (
                                                        <Badge variant="outline" className="bg-zinc-800 text-zinc-500 border-zinc-700 text-[10px]">MISSING AUTH</Badge>
                                                    )}
                                                    <span className="text-[10px] uppercase tracking-wide text-zinc-500">{(q.authMethod ?? 'none').replace(/_/g, ' ')}</span>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <span className={`text-xs px-2 py-0.5 rounded capitalize ${q.tier === 'free' ? 'text-zinc-400 bg-zinc-800' :
                                                    q.tier === 'high' ? 'text-fuchsia-400 bg-fuchsia-900/30' :
                                                        'text-blue-400 bg-blue-900/30'
                                                    }`}>
                                                    {q.tier}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 text-right font-mono">
                                                {q.limit ? (
                                                    <div className="flex flex-col items-end gap-1">
                                                        <span className={q.used >= q.limit ? 'text-red-400' : 'text-zinc-300'}>
                                                            ${q.used.toFixed(2)} / ${q.limit.toFixed(2)}
                                                        </span>
                                                        <div className="w-24 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                                                            <div className={`h-full ${q.used >= q.limit ? 'bg-red-500' : 'bg-emerald-500'}`} style={{ width: `${Math.min(100, (q.used / q.limit) * 100)}%` }} />
                                                        </div>
                                                    </div>
                                                ) : <span className="text-zinc-500">Unlimited</span>}
                                            </td>
                                            <td className="px-6 py-4 text-right font-mono text-zinc-400 text-xs">
                                                {q.rateLimitRpm ? `${q.rateLimitRpm} RPM` : '-'}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </CardContent>
                    </Card>

                    <Card id="provider-portals" className="bg-zinc-900 border-zinc-800 shadow-xl overflow-hidden">
                        <CardHeader className="bg-black/20 border-b border-white/5 pb-4">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <ExternalLink className="h-4 w-4 text-emerald-400" />
                                Quick Setup Shortcuts
                            </CardTitle>
                            <p className="text-sm text-zinc-500 mt-2">
                                Curated one-click links for the setup chores operators reach for most: credentials, plans, billing, and cloud consoles.
                            </p>
                        </CardHeader>
                        <CardContent className="p-6 border-b border-white/5">
                            <div className="grid gap-4 xl:grid-cols-3">
                                {providerQuickAccessSections.map((section) => (
                                    <div key={section.id} className="rounded-xl border border-zinc-800 bg-black/30 p-4 shadow-sm">
                                        <div className="flex items-start justify-between gap-3">
                                            <div>
                                                <h3 className="text-sm font-semibold text-zinc-100">{section.title}</h3>
                                                <p className="mt-1 text-xs text-zinc-500">{section.description}</p>
                                            </div>
                                            <Badge variant="outline" className="text-[10px] bg-zinc-800 text-zinc-400 border-zinc-700">
                                                {section.links.length} links
                                            </Badge>
                                        </div>

                                        <div className="mt-4 space-y-2">
                                            {section.links.map((link) => (
                                                <a
                                                    key={`${section.id}-${link.providerId}-${link.actionLabel}`}
                                                    href={link.href}
                                                    target="_blank"
                                                    rel="noreferrer"
                                                    className="flex items-center gap-3 rounded-lg border border-zinc-800/80 bg-zinc-950/70 px-3 py-2.5 transition hover:border-emerald-500/30 hover:text-emerald-200"
                                                >
                                                    <div className="min-w-0 flex-1">
                                                        <div className="truncate text-sm font-medium text-zinc-100">{link.providerLabel}</div>
                                                        <div className="text-[11px] text-zinc-500">{link.actionLabel}</div>
                                                    </div>
                                                    <Badge variant="outline" className={`shrink-0 text-[10px] ${getPortalBadgeClasses(link.statusTone)}`}>
                                                        {link.statusLabel}
                                                    </Badge>
                                                    <ExternalLink className="h-3.5 w-3.5 shrink-0 text-zinc-500" />
                                                </a>
                                            ))}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-900 border-zinc-800 shadow-xl overflow-hidden">
                        <CardHeader className="bg-black/20 border-b border-white/5 pb-4">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <WalletCards className="h-4 w-4 text-cyan-400" />
                                Provider Portals & Subscriptions
                            </CardTitle>
                            <p className="text-sm text-zinc-500 mt-2">
                                Jump straight to API keys, usage dashboards, billing consoles, and plan-management pages for the providers TormentNexus knows about.
                            </p>
                        </CardHeader>
                        <CardContent className="p-6">
                            <div className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3">
                                {providerPortalCards.map((portal) => (
                                    <div key={portal.id} className="rounded-xl border border-zinc-800 bg-black/30 p-4 shadow-sm">
                                        <div className="flex items-start justify-between gap-3">
                                            <div>
                                                <h3 className="text-sm font-semibold text-zinc-100">{portal.label}</h3>
                                                <p className="mt-1 text-xs text-zinc-500">{portal.notes}</p>
                                            </div>
                                            <Badge variant="outline" className={`text-[10px] ${getPortalBadgeClasses(portal.statusTone)}`}>
                                                {portal.statusLabel}
                                            </Badge>
                                        </div>

                                        <div className="mt-3 grid gap-1 text-[11px] text-zinc-400">
                                            <div>
                                                <span className="text-zinc-500">Auth:</span> {portal.authLabel}
                                            </div>
                                            <div>
                                                <span className="text-zinc-500">Availability:</span> {portal.availabilityLabel}
                                            </div>
                                            {portal.errorLabel ? (
                                                <div className="flex items-start gap-1 text-amber-400">
                                                    <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                                                    <span>{portal.errorLabel}</span>
                                                </div>
                                            ) : null}
                                        </div>

                                        <div className="mt-4 flex flex-wrap gap-2">
                                            <Button 
                                                variant="outline"
                                                size="sm"
                                                className="h-7 text-[10px] bg-zinc-800/80 text-zinc-300 border-zinc-700 hover:bg-zinc-700"
                                                onClick={() => {
                                                    setActivePortalId(portal.id);
                                                    setActivePortalName(portal.label);
                                                    setNewKeyValue('');
                                                }}
                                            >
                                                <Key className="h-3 w-3 mr-1.5" />
                                                Update Key
                                            </Button>
                                            {portal.actions.map((action) => (
                                                <a
                                                    key={`${portal.id}-${action.label}`}
                                                    href={action.href}
                                                    target="_blank"
                                                    rel="noreferrer"
                                                    className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-800/60 px-2.5 py-1.5 text-xs font-medium text-zinc-200 transition hover:border-cyan-500/40 hover:text-cyan-200"
                                                >
                                                    {action.label}
                                                    <ExternalLink className="h-3 w-3" />
                                                </a>
                                            ))}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </CardContent>
                    </Card>

                    {/* Model Pricing Dictionary */}
                    <Card className="bg-zinc-900 border-zinc-800 shadow-xl overflow-hidden">
                        <CardHeader className="bg-black/20 border-b border-white/5 pb-4">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Key className="h-4 w-4 text-indigo-400" />
                                Model Pricing Dictionary
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="p-0 overflow-x-auto max-h-[400px]">
                            <table className="w-full text-sm text-left whitespace-nowrap">
                                <thead className="text-[10px] text-zinc-500 uppercase bg-zinc-950 sticky top-0 z-10 border-b border-zinc-800">
                                    <tr>
                                        <th className="px-6 py-3 font-bold tracking-wider">Model ID</th>
                                        <th className="px-6 py-3 font-bold tracking-wider">Context Window</th>
                                        <th className="px-6 py-3 font-bold tracking-wider text-right">Input/1MT</th>
                                        <th className="px-6 py-3 font-bold tracking-wider text-right">Output/1MT</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-zinc-800/50">
                                    {isPricingLoading ? (
                                        <tr><td colSpan={4} className="px-6 py-12 text-center text-zinc-500"><Loader2 className="w-6 h-6 animate-spin mx-auto" /></td></tr>
                                    ) : pricingModels.filter((m) => m.inputPrice !== null).map((m) => (
                                        <tr key={m.id} className="hover:bg-white/[0.02] transition-colors">
                                            <td className="px-6 py-3">
                                                <div className="flex items-center gap-2">
                                                    <span className="font-mono text-zinc-300 text-xs">{m.id}</span>
                                                    {m.recommended && <Badge variant="outline" className="text-[9px] px-1 py-0 h-4 bg-indigo-500/10 text-indigo-400 border-indigo-500/30">RECOMMENDED</Badge>}
                                                </div>
                                            </td>
                                            <td className="px-6 py-3 font-mono text-zinc-400 text-xs">
                                                {m.contextWindow ? `${(m.contextWindow / 1000).toFixed(0)}k` : 'Auto'}
                                            </td>
                                            <td className="px-6 py-3 text-right font-mono text-emerald-400/80 text-xs">
                                                ${(m.inputPricePer1k * 1000).toFixed(2)}
                                            </td>
                                            <td className="px-6 py-3 text-right font-mono text-blue-400/80 text-xs">
                                                ${(m.outputPricePer1k * 1000).toFixed(2)}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </CardContent>
                    </Card>

                    {/* OAuth Client Integrations Card Scaffold */}
                    <Card className="bg-zinc-900 border-zinc-800 shadow-xl overflow-hidden">
                        <CardHeader className="bg-black/20 border-b border-white/5 pb-4 flex flex-row items-center justify-between">
                            <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Shield className="h-4 w-4 text-purple-400" />
                                OAuth App Integrations
                            </CardTitle>
                            <Button variant="outline" size="sm" className="h-7 text-[10px] border-purple-500/20 text-purple-400 bg-purple-500/5 hover:bg-purple-500/10">
                                Register New Client
                            </Button>
                        </CardHeader>
                        <CardContent className="p-6 text-center">
                            <div className="flex flex-col items-center justify-center text-zinc-500 gap-2">
                                <Shield className="h-8 w-8 opacity-20 mb-2" />
                                <p className="text-sm">No global OAuth clients registered.</p>
                                <p className="text-xs max-w-sm mx-auto">
                                    OAuth flows for specific MCP endpoints are managed per-environment or dynamically requested via the Broker during Agent tool execution.
                                </p>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>

            <Dialog open={!!activePortalId} onOpenChange={(open) => !open && setActivePortalId(null)}>
                <DialogContent className="sm:max-w-md bg-zinc-950 border-zinc-800 text-zinc-200">
                    <DialogHeader>
                        <DialogTitle>Update {activePortalName} Credentials</DialogTitle>
                        <DialogDescription className="text-zinc-400">
                            Enter your new API key, Personal Access Token (PAT), or OAuth token.
                            This will be written to `.env` immediately.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="py-4">
                        <Input
                            autoFocus
                            placeholder="Enter credential string..."
                            type="password"
                            value={newKeyValue}
                            onChange={(e) => setNewKeyValue(e.target.value)}
                            className="bg-black/50 border-zinc-800 focus-visible:ring-cyan-500/50"
                        />
                    </div>
                    <DialogFooter>
                        <Button 
                            variant="outline" 
                            onClick={() => setActivePortalId(null)}
                            className="bg-zinc-900 border-zinc-800 hover:bg-zinc-800 text-zinc-300"
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={handleSaveKey}
                            disabled={!newKeyValue.trim() || updateKeyMutation.isPending}
                            className="bg-cyan-600 hover:bg-cyan-500 text-white border-transparent"
                        >
                            {updateKeyMutation.isPending ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <Key className="w-4 h-4 mr-2" />}
                            Save & Test Connection
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

        </div>
    );
}
