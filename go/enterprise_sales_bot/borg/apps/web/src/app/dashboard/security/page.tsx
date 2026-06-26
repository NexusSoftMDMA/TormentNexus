
'use client';

import React, { useMemo, useState } from 'react';
import { trpc } from '@/utils/trpc';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@tormentnexus/ui';
import { Badge } from '@tormentnexus/ui';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';
import { ScrollArea } from '@tormentnexus/ui';
import { Shield, Lock, Rocket, RefreshCw } from 'lucide-react';
import { toast } from 'sonner';
import { OidcConfig, RbacManager } from '@tormentnexus/enterprise';

interface AuditLogEntry {
    timestamp: number;
    action: string;
    params: unknown;
    level: string;
}

function normalizeAuditLogs(value: unknown): AuditLogEntry[] {
    if (!Array.isArray(value)) return [];

    return value.filter((entry): entry is AuditLogEntry => {
        if (!entry || typeof entry !== 'object') return false;
        const log = entry as Partial<AuditLogEntry>;
        return (
            typeof log.timestamp === 'number' &&
            typeof log.action === 'string' &&
            typeof log.level === 'string'
        );
    });
}

export default function SecurityPage() {
    const [auditLimit] = useState(50);
    const [levelFilter, setLevelFilter] = useState<'ALL' | 'INFO' | 'WARN' | 'ERROR'>('ALL');
    const [searchTerm, setSearchTerm] = useState('');
    const utils = trpc.useUtils();
    const { data: rawAuditLogs, isLoading: loadingLogs, error: logsError, refetch: refetchLogs } = trpc.audit.list.useQuery(
        { limit: auditLimit },
        { refetchInterval: 5000 }
    );
    const auditLogs = normalizeAuditLogs(rawAuditLogs);
    const { data: autonomyLevel } = trpc.autonomy.getLevel.useQuery();
    const setLevelMutation = trpc.autonomy.setLevel.useMutation({
        onSuccess: async () => {
            toast.success('Autonomy level updated.');
            await utils.autonomy.getLevel.invalidate();
        },
        onError: (error) => {
            toast.error(error.message || 'Failed to update autonomy level.');
        }
    });
    const fullAutonomyMutation = trpc.autonomy.activateFullAutonomy.useMutation({
        onSuccess: async (message) => {
            toast.success(typeof message === 'string' ? message : 'Full autonomy activated.');
            await utils.autonomy.getLevel.invalidate();
        },
        onError: (error) => {
            toast.error(error.message || 'Failed to activate full autonomy.');
        }
    });

    const filteredLogs = useMemo(() => {
        return auditLogs.filter((log) => {
            const levelMatch = levelFilter === 'ALL' || log.level.toUpperCase() === levelFilter;
            const search = searchTerm.trim().toLowerCase();
            const action = log.action.toLowerCase();
            const paramsText = typeof log.params === 'string'
                ? log.params.toLowerCase()
                : JSON.stringify(log.params ?? {}).toLowerCase();
            const textMatch = !search || action.includes(search) || paramsText.includes(search);
            return levelMatch && textMatch;
        });
    }, [auditLogs, levelFilter, searchTerm]);

    const policies = [
        {
            toolName: 'Autonomy Guardrail',
            description: autonomyLevel === 'high'
                ? 'System currently in HIGH autonomy. Lockdown will reduce to LOW.'
                : 'System is in LOW autonomy lockdown mode.',
            state: autonomyLevel === 'high' ? 'Elevated' : 'Locked',
        },
        {
            toolName: 'Audit Logging',
            description: 'All tool executions and governance actions are persisted in audit logs.',
            state: 'Active',
        },
    ];


    return (
        <div className="p-6 space-y-6">
            <header className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Security & Governance</h1>
                    <p className="text-muted-foreground">Manage permissions, policies, and view audit logs.</p>
                </div>
                <div className="flex items-center gap-4">
                    <Badge variant={autonomyLevel === 'high' ? 'destructive' : 'secondary'} className="text-md px-3 py-1">
                        Autonomy: {autonomyLevel?.toUpperCase()}
                    </Badge>
                    <Button
                        variant="outline"
                        disabled={setLevelMutation.isPending || autonomyLevel === 'medium'}
                        onClick={() => setLevelMutation.mutate({ level: 'medium' })}
                        title="Set autonomy to MEDIUM"
                    >
                        MEDIUM
                    </Button>
                    <Button
                        variant="outline"
                        disabled={setLevelMutation.isPending || autonomyLevel === 'high'}
                        onClick={() => setLevelMutation.mutate({ level: 'high' })}
                        title="Set autonomy to HIGH"
                    >
                        HIGH
                    </Button>
                    <Button
                        variant="destructive"
                        disabled={setLevelMutation.isPending || autonomyLevel === 'low'}
                        onClick={() => setLevelMutation.mutate({ level: 'low' })}
                        title={autonomyLevel === 'low' ? 'System already in lockdown mode' : 'Set autonomy to LOW immediately'}
                    >
                        <Lock className="w-4 h-4 mr-2" />
                        {setLevelMutation.isPending ? 'Applying...' : autonomyLevel === 'low' ? 'LOCKDOWN ACTIVE' : 'SYSTEM LOCKDOWN'}
                    </Button>
                    <Button
                        variant="default"
                        disabled={fullAutonomyMutation.isPending}
                        onClick={() => fullAutonomyMutation.mutate()}
                        title="Enable full autonomous supervisor mode"
                    >
                        <Rocket className="w-4 h-4 mr-2" />
                        {fullAutonomyMutation.isPending ? 'Activating...' : 'ACTIVATE FULL AUTONOMY'}
                    </Button>
                </div>
            </header>

            <Tabs defaultValue="audit" className="w-full">
                <TabsList>
                    <TabsTrigger value="audit">Audit Logs</TabsTrigger>
                    <TabsTrigger value="policies">Policies</TabsTrigger>
                    <TabsTrigger value="oidc">Identity Providers (SSO)</TabsTrigger>
                    <TabsTrigger value="rbac">RBAC Manager</TabsTrigger>
                </TabsList>

                {/* AUDIT LOGS TAB */}
                <TabsContent value="audit" className="space-y-4">
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between">
                            <div>
                                <CardTitle>System Audit Log</CardTitle>
                                <CardDescription>Real-time record of all agent actions and tool executions.</CardDescription>
                            </div>
                            <div className="flex gap-2 items-center">
                                <select
                                    value={levelFilter}
                                    onChange={(e) => setLevelFilter(e.target.value as 'ALL' | 'INFO' | 'WARN' | 'ERROR')}
                                    className="h-9 rounded-md border bg-background px-2 text-xs"
                                >
                                    <option value="ALL">All levels</option>
                                    <option value="INFO">INFO</option>
                                    <option value="WARN">WARN</option>
                                    <option value="ERROR">ERROR</option>
                                </select>
                                <input
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    placeholder="Filter action/params..."
                                    className="h-9 rounded-md border bg-background px-2 text-xs w-44"
                                />
                                <Button variant="outline" size="sm" onClick={() => refetchLogs()}>
                                    <RefreshCw className="w-4 h-4 mr-1" /> Refresh
                                </Button>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <ScrollArea className="h-[600px] w-full rounded-md border p-4">
                                {loadingLogs ? <div>Loading logs...</div> : logsError ? (
                                    <div className="text-sm text-red-500">Failed to load audit logs. Try refresh.</div>
                                ) : (
                                    <div className="space-y-4">
                                        {filteredLogs.map((log, i: number) => (
                                            <div key={i} className="flex flex-col gap-1 border-b pb-2 last:border-0">
                                                <div className="flex justify-between items-start">
                                                    <div className="flex items-center gap-2">
                                                        <span className="text-xs text-muted-foreground font-mono">
                                                            {new Date(log.timestamp).toLocaleTimeString()}
                                                        </span>
                                                        <Badge variant={
                                                            log.level === 'WARN' ? 'secondary' :
                                                                log.level === 'ERROR' ? 'destructive' : 'outline'
                                                        }>
                                                            {log.level}
                                                        </Badge>
                                                        <span className="font-semibold text-sm">{log.action}</span>
                                                    </div>
                                                </div>
                                                <pre className="text-xs bg-muted/50 p-2 rounded overflow-x-auto">
                                                    {typeof log.params === 'string' ? log.params : JSON.stringify(log.params, null, 2)}
                                                </pre>
                                            </div>
                                        ))}
                                        {filteredLogs.length === 0 && (
                                            <div className="text-xs text-muted-foreground">No audit logs match the current filters.</div>
                                        )}
                                    </div>
                                )}
                            </ScrollArea>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* POLICIES TAB */}
                <TabsContent value="policies" className="space-y-4">
                    <Card>
                        <CardHeader>
                            <CardTitle>Active Policies</CardTitle>
                            <CardDescription>Rules governing tool execution and resource access.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="grid gap-4">
                                {policies.map((policy, i: number) => (
                                    <div key={i} className="flex items-center justify-between p-4 border rounded-lg bg-card">
                                        <div className="flex items-center gap-4">
                                            <Shield className="w-5 h-5 text-primary" />
                                            <div>
                                                <h3 className="font-medium">{policy.toolName}</h3>
                                                <p className="text-sm text-muted-foreground">{policy.description}</p>
                                            </div>
                                        </div>
                                        <Badge variant="outline">{policy.state}</Badge>
                                    </div>
                                ))}
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* OIDC TAB */}
                <TabsContent value="oidc" className="space-y-4">
                    <OidcConfig />
                </TabsContent>

                {/* RBAC TAB */}
                <TabsContent value="rbac" className="space-y-4">
                    <RbacManager />
                </TabsContent>
            </Tabs>
        </div>
    );
}
