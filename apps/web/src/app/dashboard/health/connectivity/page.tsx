"use client";

import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@tormentnexus/ui";
import { Activity, AlertTriangle, CheckCircle2, Loader2, Server, Wifi, WifiOff } from "lucide-react";

type ServiceStatus = {
  status: string;
  reachable: boolean;
  url?: string;
  baseURL?: string;
  name?: string;
  activeBaseURL?: string;
  statusCode?: number;
  error?: string;
  tried?: string[];
};

type ConnectivityData = {
  success: boolean;
  allHealthy: boolean;
  services: {
    trpcUpstream: ServiceStatus;
    dashboard: ServiceStatus;
    bridge: ServiceStatus;
    goSidecar: {
      status: string;
      port: number;
      baseURL: string;
      reachable: boolean;
    };
  };
  discovery: {
    goSidecarPort: number;
    trpcUpstreams: string[];
    bridgePort: number;
    dashboardPort: number;
    dashboardHost: string;
  };
};

function StatusIndicator({ reachable, status }: { reachable: boolean; status: string }) {
  if (reachable) {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2.5 py-1 text-xs font-medium text-emerald-300">
        <CheckCircle2 className="h-3.5 w-3.5" /> Online
      </span>
    );
  }
  if (status === 'error') {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-full border border-red-500/20 bg-red-500/10 px-2.5 py-1 text-xs font-medium text-red-300">
        <AlertTriangle className="h-3.5 w-3.5" /> Error
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1.5 rounded-full border border-amber-500/20 bg-amber-500/10 px-2.5 py-1 text-xs font-medium text-amber-300">
      <WifiOff className="h-3.5 w-3.5" /> Offline
    </span>
  );
}

function ServiceCard({ title, icon: Icon, service, details }: {
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  service: ServiceStatus;
  details?: Record<string, string | number | boolean | string[] | undefined>;
}) {
  return (
    <Card className="bg-zinc-900 border-zinc-800">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Icon className="h-4 w-4 text-blue-400" />
            <CardTitle className="text-sm text-white">{title}</CardTitle>
          </div>
          <StatusIndicator reachable={service.reachable} status={service.status} />
        </div>
      </CardHeader>
      <CardContent className="space-y-2 text-xs">
        {service.activeBaseURL && (
          <div className="flex justify-between">
            <span className="text-zinc-500">Active URL</span>
            <span className="font-mono text-emerald-300 break-all">{service.activeBaseURL}</span>
          </div>
        )}
        {service.baseURL && (
          <div className="flex justify-between">
            <span className="text-zinc-500">Base URL</span>
            <span className="font-mono text-zinc-300 break-all">{service.baseURL}</span>
          </div>
        )}
        {service.url && (
          <div className="flex justify-between">
            <span className="text-zinc-500">Probe URL</span>
            <span className="font-mono text-zinc-400 break-all">{service.url}</span>
          </div>
        )}
        {service.statusCode && (
          <div className="flex justify-between">
            <span className="text-zinc-500">Status Code</span>
            <span className="text-zinc-300">{service.statusCode}</span>
          </div>
        )}
        {service.error && (
          <div className="rounded border border-red-500/20 bg-red-500/10 p-2 text-red-200 break-words">
            {service.error}
          </div>
        )}
        {service.tried && service.tried.length > 0 && (
          <div>
            <span className="text-zinc-500">Attempted URLs:</span>
            <div className="mt-1 space-y-0.5">
              {service.tried.map((url) => (
                <div key={url} className="font-mono text-zinc-500 break-all">{url}</div>
              ))}
            </div>
          </div>
        )}
        {details && Object.entries(details).map(([key, value]) => (
          value !== undefined && value !== '' ? (
            <div key={key} className="flex justify-between">
              <span className="text-zinc-500">{key}</span>
              <span className="text-zinc-300">
                {Array.isArray(value) ? value.join(', ') : String(value)}
              </span>
            </div>
          ) : null
        ))}
      </CardContent>
    </Card>
  );
}

export default function ServiceConnectivityPage() {
  const [data, setData] = useState<ConnectivityData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function poll() {
      try {
        // Use Next.js API proxy to reach the Go sidecar
        const endpoints = [
          '/api/go/service/connectivity',
          '/api/go/api/service/connectivity',
        ];

        for (const endpoint of endpoints) {
          try {
            const resp = await fetch(endpoint, { signal: AbortSignal.timeout(3000) });
            if (resp.ok) {
              const json = await resp.json();
              if (!cancelled) {
                setData(json as ConnectivityData);
                setError(null);
              }
              return;
            }
          } catch {
            continue;
          }
        }

        if (!cancelled) {
          setError('Could not reach Go sidecar or dashboard proxy');
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    poll();
    const interval = setInterval(poll, 5000);
    return () => { cancelled = true; clearInterval(interval); };
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-zinc-400" />
      </div>
    );
  }

  if (error && !data) {
    return (
      <div className="p-8">
        <h1 className="text-3xl font-bold tracking-tight text-white mb-4">Service Connectivity</h1>
        <Card className="bg-zinc-900 border-red-500/30">
          <CardContent className="p-6 text-red-300">
            <AlertTriangle className="h-6 w-6 mb-2" />
            <p>{error}</p>
            <p className="mt-2 text-sm text-zinc-500">
              Ensure the Go sidecar is running: <code className="bg-zinc-800 px-1.5 py-0.5 rounded">tormentnexus serve</code>
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const trpc = data?.services?.trpcUpstream;
  const dashboard = data?.services?.dashboard;
  const bridge = data?.services?.bridge;
  const goSidecar = data?.services?.goSidecar;
  const discovery = data?.discovery;

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white">Service Connectivity</h1>
          <p className="text-zinc-500 mt-1">
            Real-time health of the TormentNexus multi-service architecture
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Activity className={`h-5 w-5 ${data?.allHealthy ? 'text-emerald-400' : 'text-amber-400'}`} />
          <span className={`text-sm font-medium ${data?.allHealthy ? 'text-emerald-300' : 'text-amber-300'}`}>
            {data?.allHealthy ? 'All services online' : 'Some services offline'}
          </span>
        </div>
      </div>

      {/* Architecture Overview */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardContent className="p-4">
          <div className="flex items-center justify-center gap-4 text-sm text-zinc-400 flex-wrap">
            <span className="flex items-center gap-1.5">
              <Wifi className="h-4 w-4 text-blue-400" />
              tRPC Core (:{discovery?.trpcUpstreams?.[0]?.split(':')?.pop()?.replace('/trpc', '') ?? '4100'})
            </span>
            <span>→</span>
            <span className="flex items-center gap-1.5">
              <Server className="h-4 w-4 text-emerald-400" />
              Go Sidecar (:{discovery?.goSidecarPort ?? 4300})
            </span>
            <span>→</span>
            <span className="flex items-center gap-1.5">
              <Activity className="h-4 w-4 text-purple-400" />
              Dashboard (:{discovery?.dashboardPort ?? 3000})
            </span>
            <span>|</span>
            <span className="flex items-center gap-1.5">
              <Wifi className="h-4 w-4 text-cyan-400" />
              Bridge (:{discovery?.bridgePort ?? 3001})
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Service Cards Grid */}
      <div className="grid gap-4 md:grid-cols-2">
        {trpc && (
          <ServiceCard
            title="tRPC Upstream (TypeScript Core)"
            icon={Wifi}
            service={trpc}
            details={{
              'Default ports': '4100, 4000, 3847',
            }}
          />
        )}
        {dashboard && (
          <ServiceCard
            title="Next.js Dashboard"
            icon={Activity}
            service={dashboard}
          />
        )}
        {bridge && (
          <ServiceCard
            title="SSE/WebSocket Bridge"
            icon={Wifi}
            service={bridge}
          />
        )}
        {goSidecar && (
          <Card className="bg-zinc-900 border-zinc-800">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Server className="h-4 w-4 text-blue-400" />
                  <CardTitle className="text-sm text-white">Go Sidecar</CardTitle>
                </div>
                <StatusIndicator reachable={goSidecar.reachable} status={goSidecar.status} />
              </div>
            </CardHeader>
            <CardContent className="space-y-2 text-xs">
              <div className="flex justify-between">
                <span className="text-zinc-500">Port</span>
                <span className="text-zinc-300">{goSidecar.port}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-zinc-500">Base URL</span>
                <span className="font-mono text-zinc-300 break-all">{goSidecar.baseURL}</span>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Discovery Configuration */}
      {discovery && (
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm text-white">Service Discovery Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="text-xs text-zinc-400 overflow-x-auto whitespace-pre-wrap">
{JSON.stringify(discovery, null, 2)}
            </pre>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
