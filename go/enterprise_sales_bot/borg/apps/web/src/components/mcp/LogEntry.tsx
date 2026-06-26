"use client";

import { cn } from "@/lib/utils";
import { format } from "date-fns";

interface LogEntryProps {
    entry: {
        timestamp: Date | string;
        level: 'INFO' | 'WARN' | 'ERROR' | 'DEBUG';
        component: string;
        message: string;
        data?: any;
    };
    className?: string;
}

export function LogEntry({ entry, className }: LogEntryProps) {
    const levelColors = {
        INFO: "text-blue-400",
        WARN: "text-yellow-400",
        ERROR: "text-red-400",
        DEBUG: "text-zinc-500",
    };

    return (
        <div className={cn("font-mono text-xs py-1 border-b border-zinc-800/50 last:border-0 hover:bg-zinc-900/50 transition-colors", className)}>
            <div className="flex gap-4">
                <span className="text-zinc-500 shrink-0 w-32">
                    {format(new Date(entry.timestamp), 'HH:mm:ss.SSS')}
                </span>
                <span className={cn("font-bold shrink-0 w-16", levelColors[entry.level] || "text-zinc-400")}>
                    {entry.level}
                </span>
                <span className="text-zinc-400 shrink-0 w-24 truncate" title={entry.component}>
                    [{entry.component}]
                </span>
                <span className="text-zinc-300 break-all">
                    {entry.message}
                    {entry.data && (
                        <pre className="mt-1 text-[10px] text-zinc-500 overflow-x-auto">
                            {JSON.stringify(entry.data, null, 2)}
                        </pre>
                    )}
                </span>
            </div>
        </div>
    );
}
