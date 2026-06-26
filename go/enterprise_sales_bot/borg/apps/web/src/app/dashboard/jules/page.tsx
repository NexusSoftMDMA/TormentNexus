"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { ExternalLink, Rocket, KeyRound, PlugZap, RefreshCw, CheckCircle2, XCircle, Trash2 } from "lucide-react";

const JULES_API_KEY_STORAGE = "jules-api-key";
const JULES_SYNC_LOG_STORAGE_KEY = 'jules-session-sync-log-v1';

type SessionSyncLogEntry = {
  sessionId: string;
  targetStatus?: 'active' | 'completed' | 'failed' | 'paused' | 'awaiting_approval';
  outcome: 'success' | 'fallback' | 'error';
  message: string;
  timestamp: string;
};

export default function JulesDashboardPage() {
  const embeddedUrl = useMemo(
    () => process.env.NEXT_PUBLIC_JULES_DASHBOARD_URL || "http://localhost:3002/jules/autopilot",
    [],
  );

  const [apiKey, setApiKey] = useState("");
  const [checking, setChecking] = useState(false);
  const [status, setStatus] = useState<{
    ok: boolean;
    message: string;
    checkedAt: string;
  } | null>(null);
  const [syncLogs, setSyncLogs] = useState<SessionSyncLogEntry[]>([]);

  useEffect(() => {
    const stored = localStorage.getItem(JULES_API_KEY_STORAGE) || "";
    setApiKey(stored);
  }, []);

  const refreshSyncLogs = useCallback(() => {
    try {
      const raw = localStorage.getItem(JULES_SYNC_LOG_STORAGE_KEY);
      const parsed: SessionSyncLogEntry[] = raw ? JSON.parse(raw) : [];
      setSyncLogs(Array.isArray(parsed) ? parsed : []);
    } catch {
      setSyncLogs([]);
    }
  }, []);

  useEffect(() => {
    refreshSyncLogs();
  }, [refreshSyncLogs]);

  const saveApiKey = useCallback(() => {
    const trimmed = apiKey.trim();
    if (!trimmed) return;
    localStorage.setItem(JULES_API_KEY_STORAGE, trimmed);
    setApiKey(trimmed);
    setStatus({
      ok: true,
      message: "API key saved locally for Jules dashboard usage.",
      checkedAt: new Date().toISOString(),
    });
  }, [apiKey]);

  const clearApiKey = useCallback(() => {
    localStorage.removeItem(JULES_API_KEY_STORAGE);
    setApiKey("");
    setStatus({
      ok: true,
      message: "API key cleared.",
      checkedAt: new Date().toISOString(),
    });
  }, []);

  const clearSyncLogs = useCallback(() => {
    localStorage.removeItem(JULES_SYNC_LOG_STORAGE_KEY);
    setSyncLogs([]);
  }, []);

  const testProxy = useCallback(async () => {
    const key = apiKey.trim();
    if (!key) {
      setStatus({
        ok: false,
        message: "Enter and save a Jules API key before testing connectivity.",
        checkedAt: new Date().toISOString(),
      });
      return;
    }

    setChecking(true);
    try {
      const path = encodeURIComponent("/sources?pageSize=1");
      const res = await fetch(`/api/jules?path=${path}`, {
        headers: {
          "x-jules-api-key": key,
        },
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        const errMessage = typeof data?.error === "string" ? data.error : `HTTP ${res.status}`;
        throw new Error(errMessage);
      }

      const sourceCount = Array.isArray(data?.sources) ? data.sources.length : 0;
      setStatus({
        ok: true,
        message: `Connected successfully via /api/jules (sample sources: ${sourceCount}).`,
        checkedAt: new Date().toISOString(),
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      setStatus({
        ok: false,
        message: `Connectivity check failed: ${message}`,
        checkedAt: new Date().toISOString(),
      });
    } finally {
      setChecking(false);
    }
  }, [apiKey]);

  const apiKeyPreview = apiKey
    ? `${apiKey.slice(0, 4)}...${apiKey.slice(-4)}`
    : "Not set";

  return (
    <div className="w-full h-full flex flex-col bg-black text-white">
      <div className="p-4 border-b border-zinc-800 flex flex-wrap items-center justify-between gap-3 bg-zinc-900">
        <div>
          <h1 className="text-xl font-bold">Jules Autopilot</h1>
          <p className="text-zinc-400 text-sm">Run and supervise Jules sessions directly from TormentNexus.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <a
            href="https://jules.google.com/settings"
            target="_blank"
            rel="noopener noreferrer"
            className="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 rounded text-xs flex items-center gap-1.5"
          >
            <KeyRound className="h-3.5 w-3.5" />
            Jules API Settings
          </a>
          <a
            href={embeddedUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 rounded text-xs flex items-center gap-1.5"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            Open Standalone
          </a>
        </div>
      </div>

      <div className="px-4 py-2 border-b border-zinc-800 text-xs text-zinc-400 bg-zinc-950 flex items-center gap-2">
        <Rocket className="h-3.5 w-3.5 text-cyan-400" />
        Embedded URL: <span className="text-zinc-300 font-mono">{embeddedUrl}</span>
      </div>

      <div className="px-4 py-3 border-b border-zinc-800 bg-zinc-950/80">
        <div className="flex flex-wrap items-center gap-2 mb-2">
          <PlugZap className="h-4 w-4 text-cyan-400" />
          <h2 className="text-sm font-semibold">Jules API Connectivity</h2>
          <span className="text-[11px] text-zinc-400">Current key: <span className="font-mono text-zinc-300">{apiKeyPreview}</span></span>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder="Paste Jules API key"
            className="min-w-[280px] flex-1 max-w-[520px] bg-zinc-900 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white outline-none focus:border-cyan-500"
          />
          <button
            onClick={saveApiKey}
            className="px-3 py-1.5 bg-emerald-700 hover:bg-emerald-600 rounded text-xs"
          >
            Save Key
          </button>
          <button
            onClick={clearApiKey}
            className="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 rounded text-xs flex items-center gap-1"
          >
            <Trash2 className="h-3 w-3" />
            Clear
          </button>
          <button
            onClick={testProxy}
            disabled={checking}
            className="px-3 py-1.5 bg-blue-700 hover:bg-blue-600 disabled:opacity-50 rounded text-xs flex items-center gap-1"
          >
            {checking ? <RefreshCw className="h-3 w-3 animate-spin" /> : <PlugZap className="h-3 w-3" />}
            Test Proxy
          </button>
        </div>
        {status ? (
          <div className={`mt-2 text-xs rounded px-2.5 py-1.5 border inline-flex items-center gap-1.5 ${status.ok ? "text-emerald-300 border-emerald-700/60 bg-emerald-950/30" : "text-red-300 border-red-700/60 bg-red-950/30"}`}>
            {status.ok ? <CheckCircle2 className="h-3.5 w-3.5" /> : <XCircle className="h-3.5 w-3.5" />}
            <span>{status.message}</span>
            <span className="text-zinc-500 ml-2">{new Date(status.checkedAt).toLocaleTimeString()}</span>
          </div>
        ) : null}
      </div>

      <div className="px-4 py-3 border-b border-zinc-800 bg-zinc-950/70">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-2">
          <h2 className="text-sm font-semibold">Last Sync Results</h2>
          <div className="flex items-center gap-2">
            <button
              onClick={refreshSyncLogs}
              className="px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 rounded text-[11px] flex items-center gap-1"
            >
              <RefreshCw className="h-3 w-3" />
              Refresh
            </button>
            <button
              onClick={clearSyncLogs}
              className="px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 rounded text-[11px] flex items-center gap-1"
            >
              <Trash2 className="h-3 w-3" />
              Clear Logs
            </button>
          </div>
        </div>

        {syncLogs.length === 0 ? (
          <p className="text-xs text-zinc-500">No sync attempts recorded yet. Drag/drop status changes in Jules Kanban will appear here.</p>
        ) : (
          <div className="max-h-36 overflow-auto space-y-1.5 pr-1">
            {syncLogs.slice(0, 12).map((entry, idx) => {
              const tone = entry.outcome === 'success'
                ? 'text-emerald-300 border-emerald-700/50 bg-emerald-950/20'
                : entry.outcome === 'fallback'
                  ? 'text-yellow-300 border-yellow-700/50 bg-yellow-950/20'
                  : 'text-red-300 border-red-700/50 bg-red-950/20';

              return (
                <div key={`${entry.timestamp}-${entry.sessionId}-${idx}`} className={`text-[11px] border rounded px-2 py-1 ${tone}`}>
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-mono">{entry.sessionId.slice(0, 12)}{entry.sessionId.length > 12 ? '…' : ''}</span>
                    <span className="text-zinc-400">{new Date(entry.timestamp).toLocaleTimeString()}</span>
                  </div>
                  <div className="text-zinc-200">
                    {entry.outcome.toUpperCase()}
                    {entry.targetStatus ? ` → ${entry.targetStatus}` : ''}
                    {` • ${entry.message}`}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div className="flex-1 relative bg-black min-h-[70vh]">
        <iframe
          src={embeddedUrl}
          title="Jules Autopilot"
          className="w-full h-full border-0"
          allow="clipboard-read; clipboard-write"
        />
      </div>
    </div>
  );
}
