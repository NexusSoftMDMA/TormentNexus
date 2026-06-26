'use client';

import { trpc } from '@/utils/trpc';

export default function SystemStatus() {
    const snapshotQuery = trpc.metrics.systemSnapshot.useQuery(undefined, {
        refetchInterval: 5000,
        refetchOnWindowFocus: false,
    });

    const load1m = snapshotQuery.data?.system.loadAvg?.[0] ?? 0;
    const memPercent = snapshotQuery.data?.system.memoryUsagePercent ?? 0;
    const freeRamGb = snapshotQuery.data ? snapshotQuery.data.system.freeMemory / (1024 ** 3) : 0;
    const uptimeHours = snapshotQuery.data ? snapshotQuery.data.process.uptimeSeconds / 3600 : 0;

    return (
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 space-y-4">
            <h2 className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-green-400 to-emerald-400">
                System Status
            </h2>
            <div className="grid grid-cols-2 gap-4">
                <Stat label="CPU Load (1m)" value={snapshotQuery.isPending ? '…' : load1m.toFixed(2)} unit="" />
                <Stat label="Memory Usage" value={snapshotQuery.isPending ? '…' : String(memPercent)} unit="%" />
                <Stat label="Free RAM" value={snapshotQuery.isPending ? '…' : freeRamGb.toFixed(2)} unit="GB" />
                <Stat label="Uptime" value={snapshotQuery.isPending ? '…' : uptimeHours.toFixed(2)} unit="h" />
            </div>
            {snapshotQuery.error && (
                <div className="text-xs text-red-400 font-mono mt-2">
                    Failed to load system snapshot: {snapshotQuery.error.message}
                </div>
            )}
        </div>
    );
}

function Stat({ label, value, unit }: { label: string; value: string; unit: string }) {
    return (
        <div className="bg-gray-800/40 p-3 rounded-lg border border-gray-800/50">
            <div className="text-gray-400 text-xs font-medium uppercase tracking-wider">{label}</div>
            <div className="text-emerald-300 font-mono text-xl font-bold mt-1">
                {value}<span className="text-xs text-gray-500 ml-1 font-normal">{unit}</span>
            </div>
        </div>
    )
}
