"use client";

import { useState, useEffect, useCallback } from "react";
import { Activity, Server, Radio, RefreshCw, Loader2, Users, Signal, Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@tormentnexus/ui";

interface PeerInfo {
  nodeId: string;
  capabilities?: string[];
  role?: string;
  load?: number;
  lastSeen?: string;
}

export default function MeshPage() {
  const [status, setStatus] = useState<{ nodeId: string; peersCount: number } | null>(null);
  const [peers, setPeers] = useState<PeerInfo[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      const [statusRes, peersRes] = await Promise.all([
        fetch("/api/go/api/mesh/status"),
        fetch("/api/go/api/mesh/peers"),
      ]);
      const statusData = await statusRes.json();
      const peersData = await peersRes.json();
      setStatus(statusData.data ?? statusData);
      // Peers returns array of node IDs or array of PeerInfo objects
      const rawPeers = peersData.data ?? peersData ?? [];
      if (rawPeers.length > 0 && typeof rawPeers[0] === "string") {
        const enriched = await Promise.all(
          rawPeers.map(async (nodeId: string) => {
            try {
              const capRes = await fetch(`/api/go/api/mesh/query-capabilities?nodeId=${nodeId}`);
              const capData = await capRes.json();
              return { nodeId, capabilities: capData.data?.capabilities ?? [] };
            } catch {
              return { nodeId };
            }
          })
        );
        setPeers(enriched);
      } else {
        setPeers(rawPeers);
      }
    } catch {
      // Best-effort
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Radio className="w-5 h-5 text-cyan-400" />
          <div>
            <div className="flex items-center gap-2"><h1 className="text-lg font-semibold text-white">P2P Fleet-Wide Intelligence</h1><TooltipProvider><Tooltip><TooltipTrigger><Info className="w-4 h-4 text-zinc-400" /></TooltipTrigger><TooltipContent><p>A decentralized network for cross-machine AI memory sharing.</p></TooltipContent></Tooltip></TooltipProvider></div>
            <p className="text-xs text-zinc-500 mt-0.5">
              Encrypted mesh for cross-machine memory sharing via UDP gossip protocol
            </p>
          </div>
        </div>
        <button
          onClick={fetchAll}
          disabled={loading}
          className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50 flex items-center gap-1.5"
          title="Refresh mesh status and peer list"
        >
          {loading ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
          Refresh
        </button>
      </div>

      {/* Status Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
          <div className="flex items-center gap-2 text-zinc-500 text-xs mb-2">
            <Server className="w-3.5 h-3.5" />
            Node ID
          </div>
          <p className="text-sm font-mono text-zinc-200 truncate" title={status?.nodeId}>
            {status?.nodeId || (loading ? "..." : "Not connected")}
          </p>
        </div>
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
          <div className="flex items-center gap-2 text-zinc-500 text-xs mb-2">
            <Users className="w-3.5 h-3.5" />
            Connected Peers<TooltipProvider><Tooltip><TooltipTrigger><Info className="w-3 h-3 text-zinc-500" /></TooltipTrigger><TooltipContent><p>Active agents in the A2A network protocol.</p></TooltipContent></Tooltip></TooltipProvider>
          </div>
          <p className="text-2xl font-bold text-cyan-400">{status?.peersCount ?? 0}</p>
        </div>
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
          <div className="flex items-center gap-2 text-zinc-500 text-xs mb-2">
            <Activity className="w-3.5 h-3.5" />
            Mesh Status
          </div>
          <p className="text-sm text-emerald-400 font-medium">
            {status ? "Connected" : loading ? "Connecting..." : "Disconnected"}
          </p>
        </div>
      </div>

      {/* Peers List */}
      <div>
        <h2 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-2">
          <Signal className="w-4 h-4" />
          Peers ({peers.length})
        </h2>
        {peers.length === 0 && !loading && (
          <div className="text-center py-12 text-zinc-600 bg-zinc-900/30 border border-zinc-800 rounded-lg">
            <Radio className="w-10 h-10 mx-auto mb-3 opacity-30" />
            <p className="font-medium">No peers found</p>
            <p className="text-xs mt-1 max-w-md mx-auto">
              The mesh is listening for other TormentNexus nodes on the network. Start another
              instance or check your network configuration.
            </p>
          </div>
        )}
        <div className="space-y-2">
          {peers.map((peer) => (
            <div key={peer.nodeId} className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-mono text-zinc-300 truncate" title={peer.nodeId}>
                  {peer.nodeId.slice(0, 48)}...
                </span>
                <span className="text-xs text-zinc-500">{peer.role || "peer"}</span>
              </div>
              {peer.capabilities && peer.capabilities.length > 0 && (
                <div className="flex gap-1 flex-wrap">
                  {peer.capabilities.map((cap) => (
                    <span key={cap} className="px-1.5 py-0.5 bg-zinc-800 rounded text-2xs text-zinc-400">
                      {cap}
 Charter
                    </span>
                  ))}
                </div>
              )}
              {peer.load !== undefined && (
                <div className="mt-2 flex items-center gap-2 text-xs text-zinc-500">
                  <span>Load</span>
                  <div className="flex-1 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-cyan-500 rounded-full transition-all"
                      style={{ width: `${Math.min(peer.load * 100, 100)}%` }}
                    />
                  </div>
                  <span className="text-zinc-400">{(peer.load * 100).toFixed(0)}%</span>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
