'use client';

import React from 'react';
import { trpc } from '../utils/trpc';
import { Badge } from './ui/badge';
import { Wifi, WifiOff, Loader2, Zap } from 'lucide-react';
import { useState, useEffect } from 'react';

export function StreamStatus() {
  const [sidecarOnline, setSidecarOnline] = useState<boolean | null>(null);

  const healthQuery = trpc.health.useQuery(undefined, {
    refetchInterval: 5000,
    retry: true,
  });

  useEffect(() => {
    const checkSidecar = async () => {
      try {
        const res = await fetch('http://127.0.0.1:4300/health', { signal: AbortSignal.timeout(2000) });
        setSidecarOnline(res.ok);
      } catch {
        setSidecarOnline(false);
      }
    };

    checkSidecar();
    const timer = setInterval(checkSidecar, 10000);
    return () => clearInterval(timer);
  }, []);

  const tsLive = healthQuery.data?.status === 'running';
  const isError = healthQuery.isError;

  return (
    <div className="flex items-center gap-2">
      {/* TS Core Status */}
      {isError ? (
        <Badge variant="destructive" className="flex items-center gap-1 py-0.5 px-2 text-[10px]">
          <WifiOff className="h-3 w-3" />
          CORE OFFLINE
        </Badge>
      ) : tsLive ? (
        <Badge variant="outline" className="bg-green-500/10 text-green-500 border-green-500/20 flex items-center gap-1 py-0.5 px-2 text-[10px]">
          <Wifi className="h-3 w-3" />
          STREAM LIVE
        </Badge>
      ) : (
        <Badge variant="secondary" className="flex items-center gap-1 py-0.5 px-2 text-[10px]">
          <Loader2 className="h-3 w-3 animate-spin" />
          RECONNECTING
        </Badge>
      )}

      {/* Go Sidecar Status */}
      {sidecarOnline === true && (
        <Badge variant="outline" className="bg-blue-500/10 text-blue-400 border-blue-500/20 flex items-center gap-1 py-0.5 px-2 text-[10px]">
          <Zap className="h-3 w-3 fill-current" />
          SIDECAR
        </Badge>
      )}
    </div>
  );
}
