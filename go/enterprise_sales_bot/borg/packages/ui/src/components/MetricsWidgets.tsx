import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Activity, CheckCircle, AlertOctagon, Clock } from "lucide-react";
// We might need a charting library. If not installed, we can make simple SVG charts.
// Assuming we don't have recharts yet, let's build simple SVGs.

interface MetricsWidgetsProps {
    stats: any; // We'll type this properly later or infer from tRPC
}

export const ActivityPulse: React.FC<{ series: any[] }> = ({ series }) => {
    // Simple SVG Line Chart
    if (!series || series.length === 0) return <div className="text-zinc-500 text-sm">No activity</div>;

    const height = 60;
    const width = 300;
    const maxVal = Math.max(...series.map(s => s.count), 10);

    // Polyline points
    const points = series.map((s, i) => {
        const x = (i / (series.length - 1)) * width;
        const y = height - (s.count / maxVal) * height;
        return `${x},${y}`;
    }).join(' ');

    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Activity Pulse</CardTitle>
                <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold">{series.reduce((a, b) => a + b.count, 0)} <span className="text-xs font-normal text-muted-foreground">events/hr</span></div>
                <div className="h-[60px] w-full mt-4">
                    <svg width="100%" height="100%" viewBox={`0 0 ${width} ${height}`} preserveAspectRatio="none">
                        <polyline
                            fill="none"
                            stroke="#10b981"
                            strokeWidth="2"
                            points={points}
                        />
                    </svg>
                </div>
            </CardContent>
        </Card>
    );
};

export const SystemHealth: React.FC<{ counts: Record<string, number> }> = ({ counts }) => {
    const success = counts['tool_call'] || 0; // approximate total
    // We didn't separate success vs fail in 'type' easily yet, unless we query by tags.
    // Wait, getStats aggregates by TYPE.
    // In MCPServer we tracked: 'tool_call' (count) and 'tool_error'.
    const errors = counts['tool_error'] || 0;
    const total = success + (counts['tool_call_failed'] || 0); // Need to improve tracking logic if shared type. 
    // Actually, we tracked 'tool_call' for BOTH success and fail, but with tags. 
    // getStats currently aggregates by TYPE. So 'tool_call' is total calls. 'tool_error' is errors.

    const errorRate = success > 0 ? (errors / success) * 100 : 0;
    const health = errorRate < 5 ? "Healthy" : errorRate < 15 ? "Degraded" : "Critical";
    const color = errorRate < 5 ? "text-green-500" : errorRate < 15 ? "text-yellow-500" : "text-red-500";

    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">System Health</CardTitle>
                {health === "Healthy" ? <CheckCircle className="h-4 w-4 text-green-500" /> : <AlertOctagon className="h-4 w-4 text-red-500" />}
            </CardHeader>
            <CardContent>
                <div className={`text-2xl font-bold ${color}`}>{health}</div>
                <p className="text-xs text-muted-foreground">
                    Error Rate: {errorRate.toFixed(1)}% ({errors} errors)
                </p>
            </CardContent>
        </Card>
    );
}

export const LatencyMonitor: React.FC<{ averages: Record<string, number> }> = ({ averages }) => {
    const avg = averages['tool_execution'] || 0;
    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold">{Math.round(avg)} ms</div>
                <p className="text-xs text-muted-foreground">
                    Per tool execution
                </p>
            </CardContent>
        </Card>
    )
}
