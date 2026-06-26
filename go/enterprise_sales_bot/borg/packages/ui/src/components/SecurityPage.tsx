"use client";

import React, { useState } from 'react';
import { Shield, ShieldAlert, ShieldCheck, Lock, Unlock, Plus, Trash2, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { trpc } from '../utils/trpc'; // Mock import, will be provided by context or generic

// Mock types for UI dev
interface PolicyRule {
    action: string;
    resource: string;
    effect: "ALLOW" | "DENY" | "ASK";
    reason?: string;
}

export const SecurityPage = () => {
    // @ts-ignore
    const rulesQuery = trpc.policy.getRules.useQuery();
    // @ts-ignore
    const updateRulesMutation = trpc.policy.updateRules.useMutation();
    // @ts-ignore
    const lockdownMutation = trpc.policy.lockdown.useMutation();
    // @ts-ignore
    const unlockMutation = trpc.policy.unlock.useMutation();
    // @ts-ignore
    const autonomyQuery = trpc.autonomy.getLevel.useQuery();
    // @ts-ignore
    const auditQuery = trpc.audit.query.useQuery({ level: "WARN", limit: 10 });

    const [newRule, setNewRule] = useState<PolicyRule>({ action: 'execute', resource: '', effect: 'DENY', reason: '' });

    const isLocked = rulesQuery.data?.[0]?.reason === 'SYSTEM LOCKDOWN';

    const handleLockdown = async () => {
        if (isLocked) {
            await unlockMutation.mutateAsync();
        } else {
            await lockdownMutation.mutateAsync();
        }
        rulesQuery.refetch();
        autonomyQuery.refetch();
    };

    const handleAddRule = async () => {
        if (!newRule.resource) return;
        const currentRules = rulesQuery.data || [];
        // Insert at top (but below lockdown if locked)
        const insertIndex = isLocked ? 1 : 0;
        const newRules = [...currentRules];
        newRules.splice(insertIndex, 0, newRule);

        await updateRulesMutation.mutateAsync({ rules: newRules });
        rulesQuery.refetch();
        setNewRule({ action: 'execute', resource: '', effect: 'DENY', reason: '' });
    };

    const handleDeleteRule = async (index: number) => {
        const currentRules = rulesQuery.data || [];
        const newRules = currentRules.filter((_: any, i: number) => i !== index);
        await updateRulesMutation.mutateAsync({ rules: newRules });
        rulesQuery.refetch();
    };

    return (
        <div className="p-8 space-y-8 max-w-6xl mx-auto text-zinc-900 dark:text-zinc-100">
            <header className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold flex items-center gap-3">
                        {isLocked ? <ShieldAlert className="h-8 w-8 text-red-500" /> : <ShieldCheck className="h-8 w-8 text-green-500" />}
                        Security & Policy
                    </h1>
                    <p className="text-zinc-500 dark:text-zinc-400 mt-2">
                        Manage system permissions, active guardrails, and emergency controls.
                    </p>
                </div>
                <div className="flex items-center gap-4">
                    <div className="flex flex-col items-end">
                        <span className="text-xs font-mono uppercase text-zinc-500">Autonomy Level</span>
                        <Badge variant={autonomyQuery.data === 'high' ? 'destructive' : 'outline'} className="text-lg">
                            {autonomyQuery.data?.toUpperCase() || 'UNKNOWN'}
                        </Badge>
                    </div>
                    <Button
                        size="lg"
                        variant={isLocked ? "outline" : "destructive"}
                        onClick={handleLockdown}
                        className={isLocked ? "border-red-500 text-red-500 hover:bg-red-500/10" : "bg-red-600 hover:bg-red-700 text-white"}
                    >
                        {isLocked ? <><Unlock className="mr-2 h-5 w-5" /> DISENGAGE LOCKDOWN</> : <><Lock className="mr-2 h-5 w-5" /> LOCKDOWN SYSTEM</>}
                    </Button>
                </div>
            </header>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Active Policies */}
                <Card className="col-span-2 shadow-sm border-zinc-200 dark:border-zinc-800">
                    <CardHeader>
                        <CardTitle>Active Guardrails</CardTitle>
                        <CardDescription>Rules are evaluated from top to bottom. First match wins.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {/* Policy Editor / List */}
                        <div className="space-y-2">
                            {rulesQuery.data?.map((rule: typeof rulesQuery.data[number], i: number) => (
                                <div key={i} className={`flex items-center justify-between p-3 rounded-md border ${rule.effect === 'DENY' ? 'bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-900' : 'bg-green-50 border-green-200 dark:bg-green-900/20 dark:border-green-900'}`}>
                                    <div className="flex items-center gap-4">
                                        <Badge variant={rule.effect === 'DENY' ? 'destructive' : 'default'} className="w-16 justify-center">
                                            {rule.effect}
                                        </Badge>
                                        <div className="flex flex-col">
                                            <span className="font-mono text-sm font-bold">{rule.action} <span className="text-zinc-400 mx-1">➜</span> {rule.resource}</span>
                                            {rule.reason && <span className="text-xs text-zinc-500 italic">"{rule.reason}"</span>}
                                        </div>
                                    </div>
                                    {rule.reason !== 'SYSTEM LOCKDOWN' ? (
                                        <Button variant="ghost" size="icon" onClick={() => handleDeleteRule(i)} className="text-zinc-400 hover:text-red-500">
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    ) : (
                                        <Lock className="h-4 w-4 text-red-500" />
                                    )}
                                </div>
                            ))}
                            {(!rulesQuery.data || rulesQuery.data.length === 0) && (
                                <div className="text-center p-8 text-zinc-400 border-2 border-dashed rounded-lg">
                                    No active policies. System relies on Permission Manager.
                                </div>
                            )}
                        </div>

                        {/* Add Rule Form */}
                        <div className="border-t pt-4 mt-6">
                            <h4 className="text-sm font-medium mb-3">Add New Guardrail</h4>
                            <div className="flex gap-2 items-end">
                                <div className="space-y-1 w-32">
                                    <Label className="text-xs">Action</Label>
                                    <Input value={newRule.action} onChange={e => setNewRule({ ...newRule, action: e.target.value })} placeholder="execute" className="font-mono text-xs" />
                                </div>
                                <div className="space-y-1 flex-1">
                                    <Label className="text-xs">Resource (Glob)</Label>
                                    <Input value={newRule.resource} onChange={e => setNewRule({ ...newRule, resource: e.target.value })} placeholder="rm -rf *" className="font-mono text-xs" />
                                </div>
                                <div className="space-y-1 w-24">
                                    <Label className="text-xs">Effect</Label>
                                    <select
                                        className="flex h-9 w-full items-center justify-between rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                                        value={newRule.effect}
                                        onChange={e => setNewRule({ ...newRule, effect: e.target.value as any })}
                                    >
                                        <option value="DENY">DENY</option>
                                        <option value="ALLOW">ALLOW</option>
                                        <option value="ASK">ASK</option>
                                    </select>
                                </div>
                                <Button onClick={handleAddRule} disabled={!newRule.resource}>
                                    <Plus className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Right Column: Alerts & Status */}
                <div className="space-y-6">
                    <Card className="border-zinc-200 dark:border-zinc-800">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <AlertTriangle className="h-5 w-5 text-yellow-500" />
                                Recent Violations
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-3">
                                {auditQuery.data?.map((log: any, i: number) => (
                                    <div key={i} className="text-sm border-l-2 border-yellow-500 pl-3 py-1">
                                        <div className="font-medium">{log.event}</div>
                                        <div className="text-xs text-zinc-500 font-mono truncate">{JSON.stringify(log.details)}</div>
                                        <div className="text-[10px] text-zinc-400 mt-1">{new Date(log.timestamp).toLocaleTimeString()}</div>
                                    </div>
                                ))}
                                {(!auditQuery.data || auditQuery.data.length === 0) && (
                                    <div className="text-zinc-500 text-sm">No recent security alerts.</div>
                                )}
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-zinc-50 dark:bg-zinc-900 border-zinc-200 dark:border-zinc-800">
                        <CardHeader>
                            <CardTitle>System Status</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-2 text-sm">
                            <div className="flex justify-between">
                                <span className="text-zinc-500">Policy Engine</span>
                                <span className="text-green-500 font-bold">ACTIVE</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-zinc-500">Rules Loaded</span>
                                <span>{rulesQuery.data?.length || 0}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-zinc-500">Permission Check</span>
                                <span>Risk-Based</span>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
};
