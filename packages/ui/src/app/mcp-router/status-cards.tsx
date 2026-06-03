/**
 * TORMENTNEXUS MCP Router - Status Cards Component
 *
 * Displays key statistics for registry, sessions, and system health.
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useState, useEffect } from 'react';

interface StatusCardProps {
    title: string;
    value: string | number;
    description?: string;
    trend?: { value: number; label: string };
    icon?: string;
    color?: string;
}

type StatusCardColor = 'green' | 'yellow' | 'red' | 'blue';

function StatusCard({ title, value, description, trend, icon, color = 'green' }: StatusCardProps) {
    const getColorStyles = (color: StatusCardColor) => {
        const styles = {
            green: 'bg-green-500/20 text-green-900',
            yellow: 'bg-yellow-500/20 text-yellow-900',
            red: 'bg-red-500/20 text-red-900',
            blue: 'bg-blue-500/20 text-blue-900'
        };
        return styles[color];
    };

    const getBadgeVariant = (color: StatusCardColor): 'default' | 'secondary' | 'destructive' => {
        const variantMap: Record<StatusCardColor, 'default' | 'secondary' | 'destructive'> = {
            green: 'default',
            yellow: 'secondary',
            red: 'destructive',
            blue: 'default'
        };
        return variantMap[color];
    };

    const typedColor = color as StatusCardColor;
    const bgColor = getColorStyles(typedColor).split(' ')[0];

        return (
        <Card className="hover:shadow-lg transition-shadow">
            <CardContent>
                <div className="flex items-center justify-between">
                    <div className="flex-1">
                        <CardHeader className="text-sm font-medium text-gray-600">
                            {title}
                        </CardHeader>
                        {icon && <span className="text-2xl ml-2">{icon}</span>}
                        <CardTitle className="text-4xl font-bold">
                            {typeof value === 'number' ? value.toLocaleString() : value}
                        </CardTitle>
                        {trend && (
                            <div className="flex items-center text-sm text-gray-500 ml-4">
                                <span className={trend.value >= 0 ? 'text-green-500' : 'text-red-500'}>
                                    {trend.value >= 0 ? '↑' : '↓'}
                                </span>
                                <span className="ml-1">{trend.label}</span>
                            </div>
                        )}
                    </div>
                    <Badge variant={getBadgeVariant(typedColor)}>
                        {typeof value === 'number' ? value : 'OK'}
                    </Badge>
                </div>
                {description && (
                    <CardDescription className="text-gray-600">
                        {description}
                    </CardDescription>
                )}
            </CardContent>
        </Card>
    );
}

interface MCPRouterStatsProps {
    registryStats?: {
        totalServers: number;
        installedServers: number;
        categories: number;
    };
    sessionStats?: {
        totalSessions: number;
        running: number;
        stopped: number;
        error: number;
        totalClients: number;
    };
    healthStatus?: {
        status: 'healthy' | 'degraded' | 'unhealthy';
        uptime: number;
    };
    loading?: boolean;
}

export function MCPRouterStats({ registryStats, sessionStats, healthStatus, loading = false }: MCPRouterStatsProps) {
    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Registry Stats */}
            <StatusCard
                title="Total Servers"
                value={registryStats?.totalServers || 0}
                description={`From ${registryStats?.categories || 0} registries`}
                icon="📦"
                color="blue"
            />
            <StatusCard
                title="Installed"
                value={registryStats?.installedServers || 0}
                description="Ready to use"
                icon="✓"
                color="green"
            />
            <StatusCard
                title="Running Sessions"
                value={sessionStats?.running ?? 0}
                description={sessionStats ? `${sessionStats.stopped} stopped, ${sessionStats.error} errors` : undefined}
                icon="▶️"
                color={sessionStats?.running ?? 0 > 0 ? 'green' : 'yellow'}
            />
            <StatusCard
                title="Total Clients"
                value={sessionStats?.totalClients ?? 0}
                description="Active connections"
                icon="👥"
                color="blue"
            />

            {/* Health Status */}
            <StatusCard
                title="System Health"
                value={healthStatus?.status === 'healthy' ? 'OK' : (healthStatus?.status || 'unknown').toUpperCase()}
                description={`Uptime: ${healthStatus?.uptime ? Math.floor(healthStatus.uptime / 60).toFixed(1) + 'm' : 'N/A'}`}
                icon={healthStatus?.status === 'healthy' ? '✓' : '⚠️'}
                color={healthStatus?.status === 'healthy' ? 'green' : healthStatus?.status === 'degraded' ? 'yellow' : 'red'}
                trend={healthStatus?.uptime ? { value: healthStatus.uptime - (healthStatus.uptime * 0.95), label: '1h ago' } : undefined}
            />
        </div>
    );
}

export default MCPRouterStats;
