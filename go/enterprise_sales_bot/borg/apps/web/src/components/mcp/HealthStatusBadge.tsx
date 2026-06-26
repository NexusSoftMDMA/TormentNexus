"use client";

import { cn } from "@/lib/utils";

interface HealthStatusBadgeProps {
    status: 'HEALTHY' | 'UNHEALTHY' | 'DEGRADED' | 'UNKNOWN' | 'ERROR';
    className?: string;
}

export function HealthStatusBadge({ status, className }: HealthStatusBadgeProps) {
    const variants = {
        HEALTHY: "bg-green-500/10 text-green-500 border-green-500/20",
        UNHEALTHY: "bg-red-500/10 text-red-500 border-red-500/20",
        DEGRADED: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
        UNKNOWN: "bg-zinc-500/10 text-zinc-500 border-zinc-500/20",
        ERROR: "bg-red-500/10 text-red-500 border-red-500/20",
    };

    return (
        <span className={cn(
            "px-2 py-0.5 rounded text-[10px] font-medium border uppercase tracking-wider",
            variants[status] || variants.UNKNOWN,
            className
        )}>
            {status}
        </span>
    );
}
