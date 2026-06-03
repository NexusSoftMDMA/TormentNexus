"use client";

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Input } from "@tormentnexus/ui";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@tormentnexus/ui";
import { FileText, Search, RefreshCcw } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

function formatTime(timestamp: number | string) {
    const d = new Date(timestamp);
    if (isNaN(d.getTime())) return String(timestamp);
    return d.toLocaleString([], { 
        month: 'short', 
        day: '2-digit',
        hour12: false, 
        hour: '2-digit', 
        minute: '2-digit', 
        second: '2-digit' 
    });
}

export default function AuditDashboard() {
    const [searchQuery, setSearchQuery] = useState("");
    const [limit, setLimit] = useState(100);

    // Fetch audit entries
    const { data: auditLogs, refetch, isFetching } = trpc.audit.list.useQuery({ limit });

    const handleRefresh = async () => {
        await refetch();
        toast.success("Audit trail refreshed");
    };

    // Filter audit logs client-side
    const typedAuditLogs = (auditLogs as any[] | undefined) ?? [];
    const filteredLogs = typedAuditLogs.filter((log: any) => {
        if (!searchQuery) return true;
        const query = searchQuery.toLowerCase();
        return (
            (log.action && log.action.toLowerCase().includes(query)) ||
            (log.agentId && log.agentId.toLowerCase().includes(query)) ||
            (log.level && log.level.toLowerCase().includes(query)) ||
            (log.details && JSON.stringify(log.details).toLowerCase().includes(query))
        );
    });

    return (
        <div className="p-8 space-y-8 h-full overflow-y-auto w-full max-w-[1200px] mx-auto">
            {/* Header */}
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <FileText className="h-8 w-8 text-indigo-500" />
                        System Audit
                    </h1>
                    <p className="text-zinc-500 mt-1">
                        Cryptographic and system-level audit trail of critical operational events.
                    </p>
                </div>
                <Button 
                    onClick={handleRefresh} 
                    disabled={isFetching}
                    variant="outline" 
                    className="border-zinc-700 hover:bg-zinc-800"
                >
                    <RefreshCcw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} /> 
                    Refresh
                </Button>
            </div>

            {/* Main Content Area */}
            <Card className="bg-zinc-900 border-zinc-800 overflow-hidden">
                 <CardHeader className="pb-4 bg-zinc-950/50 border-b border-zinc-800">
                    <div className="flex items-center gap-4">
                        <div className="relative flex-1 max-w-md">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
                            <Input 
                                placeholder="Filter by action, level, or agent..." 
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="w-full pl-9 bg-zinc-900 border-zinc-800 text-white"
                            />
                        </div>
                        <div className="flex items-center gap-2 text-sm text-zinc-500 ml-auto">
                            <span>Showing {filteredLogs.length} events</span>
                        </div>
                    </div>
                </CardHeader>
                <div className="overflow-x-auto">
                    <Table>
                        <TableHeader className="bg-zinc-950">
                            <TableRow className="border-zinc-800 hover:bg-transparent">
                                <TableHead className="w-[180px] text-zinc-400">Time</TableHead>
                                <TableHead className="w-[100px] text-zinc-400">Level</TableHead>
                                <TableHead className="w-[150px] text-zinc-400">Agent/Source</TableHead>
                                <TableHead className="w-[200px] text-zinc-400">Action</TableHead>
                                <TableHead className="text-zinc-400">Details</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredLogs.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="h-32 text-center text-zinc-500">
                                        No audit records found matching your criteria.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredLogs.map((log: any, index: number) => {
                                    // Make level colored based on severity
                                    const levelColors: Record<string, string> = {
                                        info: 'text-blue-400',
                                        warn: 'text-yellow-400',
                                        error: 'text-red-400',
                                        critical: 'text-red-500 font-bold',
                                        security: 'text-purple-400 font-bold'
                                    };
                                    const levelStyle = levelColors[log.level?.toLowerCase()] || 'text-zinc-400';

                                    // Format details
                                    let detailsText = '';
                                    if (typeof log.details === 'string') {
                                        detailsText = log.details;
                                    } else if (log.details) {
                                        detailsText = JSON.stringify(log.details);
                                        if (detailsText === '{}') detailsText = '';
                                    }

                                    return (
                                        <TableRow key={log.id || index} className="border-zinc-800/50 hover:bg-zinc-800/30">
                                            <TableCell className="font-mono text-xs text-zinc-500 whitespace-nowrap">
                                                {formatTime(log.timestamp)}
                                            </TableCell>
                                            <TableCell className={`font-mono text-xs uppercase ${levelStyle}`}>
                                                {log.level || 'INFO'}
                                            </TableCell>
                                            <TableCell className="font-mono text-xs text-zinc-300 truncate max-w-[150px]">
                                                {log.agentId || 'system'}
                                            </TableCell>
                                            <TableCell className="font-medium text-xs text-zinc-200 truncate max-w-[200px]" title={log.action}>
                                                {log.action}
                                            </TableCell>
                                            <TableCell className="text-xs text-zinc-400 truncate max-w-[400px]" title={detailsText}>
                                                {detailsText || <span className="text-zinc-700 italic">No additional details</span>}
                                            </TableCell>
                                        </TableRow>
                                    );
                                })
                            )}
                        </TableBody>
                    </Table>
                </div>
            </Card>
        </div>
    );
}
