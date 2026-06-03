"use client";

import { useState } from 'react';
import { Card } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Calendar, User, Search, FileText } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

export default function AuditDashboard() {
    const [limit, setLimit] = useState(50);
    const { data: logs, isLoading } = trpc.audit.list.useQuery({ limit });

    return (
        <div className="p-8 space-y-8 h-full flex flex-col">
            <div className="flex justify-between items-center shrink-0">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">System Audit</h1>
                    <p className="text-zinc-500">
                        Immutable record of system actions and security events
                    </p>
                </div>
            </div>

            <Card className="bg-zinc-900 border-zinc-800 flex-1 flex flex-col overflow-hidden">
                <div className="p-4 border-b border-zinc-800 flex justify-between items-center">
                    <div className="flex items-center gap-2 text-sm text-zinc-400">
                        <FileText className="h-4 w-4" />
                        Showing last {limit} events
                    </div>
                </div>

                <div className="flex-1 overflow-auto">
                    {isLoading ? (
                        <div className="flex justify-center p-12">
                            <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
                        </div>
                    ) : ((logs as any)?.length ?? 0) === 0 ? (
                        <div className="text-center p-12 text-zinc-500">
                            <FileText className="h-12 w-12 mx-auto mb-4 opacity-30" />
                            <p className="text-lg font-medium">No Audit Logs</p>
                        </div>
                    ) : (
                        <table className="w-full text-left text-sm">
                            <thead className="bg-zinc-950 text-zinc-400 font-medium">
                                <tr>
                                    <th className="p-4 border-b border-zinc-800">Time</th>
                                    <th className="p-4 border-b border-zinc-800">Action</th>
                                    <th className="p-4 border-b border-zinc-800">Actor</th>
                                    <th className="p-4 border-b border-zinc-800">Resource</th>
                                    <th className="p-4 border-b border-zinc-800">Details</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-zinc-800">
                                {(logs as any)?.map((log: any) => (
                                    <tr key={log.id} className="hover:bg-zinc-800/50 transition-colors">
                                        <td className="p-4 text-zinc-500 whitespace-nowrap font-mono text-xs">
                                            {new Date(log.timestamp).toLocaleString()}
                                        </td>
                                        <td className="p-4 font-medium text-white">
                                            {log.action}
                                        </td>
                                        <td className="p-4 text-purple-400">
                                            {log.actor || 'System'}
                                        </td>
                                        <td className="p-4 text-zinc-300">
                                            {log.resource}
                                        </td>
                                        <td className="p-4 text-zinc-500 max-w-md truncate">
                                            {JSON.stringify(log.details)}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </div>
            </Card>
        </div>
    );
}
