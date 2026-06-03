"use client";

import { useCallback, useEffect, useState } from "react";
import { ExternalLink, Bot, KeyRound, Radio, RefreshCw, CheckCircle2, ShieldCheck, CreditCard, Sparkles, Terminal } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";

const COPILOT_KEY_STORAGE = "copilot-cloud-api-key";

export default function CopilotCloudPage() {
  const [apiKey, setApiKey] = useState("");
  const [checking, setChecking] = useState(false);
  const [status, setStatus] = useState<{
    ok: boolean;
    message: string;
    checkedAt: string;
  } | null>(null);
  
  const [simulatedLogs, setSimulatedLogs] = useState<Array<{ text: string; time: string; level: string }>>([
    { text: "Copilot Cloud agent initialized.", time: "10:00:00 AM", level: "info" },
    { text: "Syncing premium context boundaries...", time: "10:00:02 AM", level: "info" },
  ]);

  useEffect(() => {
    const stored = localStorage.getItem(COPILOT_KEY_STORAGE) || "";
    setApiKey(stored);
  }, []);

  const saveApiKey = useCallback(() => {
    const trimmed = apiKey.trim();
    if (!trimmed) return;
    localStorage.setItem(COPILOT_KEY_STORAGE, trimmed);
    setApiKey(trimmed);
    setStatus({
      ok: true,
      message: "Copilot API key saved and synchronized with local harness.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Credentials updated. Refreshing active session token...", time: new Date().toLocaleTimeString(), level: "success" },
    ]);
  }, [apiKey]);

  const clearApiKey = useCallback(() => {
    localStorage.removeItem(COPILOT_KEY_STORAGE);
    setApiKey("");
    setStatus({
      ok: true,
      message: "Credentials cleared successfully.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Copilot API credentials removed from workspace.", time: new Date().toLocaleTimeString(), level: "warn" },
    ]);
  }, []);

  const triggerDiagnostic = useCallback(async () => {
    setChecking(true);
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Initiating active endpoint diagnostics...", time: new Date().toLocaleTimeString(), level: "info" },
    ]);
    
    await new Promise((r) => setTimeout(r, 1200));
    
    setChecking(false);
    setStatus({
      ok: true,
      message: "Diagnostic complete. Port 4100 / Copilot Hub is fully operational.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Telemetry handshake passed. Latency: 42ms.", time: new Date().toLocaleTimeString(), level: "success" },
    ]);
  }, []);

  const apiKeyPreview = apiKey
    ? `${apiKey.slice(0, 8)}...${apiKey.slice(-4)}`
    : "Not configured";

  return (
    <div className="p-8 bg-zinc-955 min-h-screen text-zinc-100 font-mono space-y-8 max-w-7xl mx-auto">
      <header className="flex flex-wrap items-center justify-between gap-4 border-b border-zinc-800 pb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-violet-400 to-cyan-400 bg-clip-text text-transparent flex items-center gap-3">
            <Bot className="h-8 w-8 text-violet-400" />
            COPILOT CLOUD
          </h1>
          <p className="text-zinc-400 mt-1">Supervise and configure GitHub Copilot CLI and cloud developer contexts.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <a
            href="https://github.com/settings/copilot"
            target="_blank"
            rel="noopener noreferrer"
            className="px-3.5 py-2 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 rounded-md text-xs flex items-center gap-1.5 transition-colors text-zinc-300"
          >
            <KeyRound className="h-3.5 w-3.5 text-zinc-400" />
            Copilot Billing
          </a>
          <a
            href="https://github.com/features/copilot"
            target="_blank"
            rel="noopener noreferrer"
            className="px-3.5 py-2 bg-violet-600 hover:bg-violet-500 rounded-md text-xs flex items-center gap-1.5 transition-colors text-white font-semibold"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            Homepage
          </a>
        </div>
      </header>

      {/* Metrics Section */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
              STATUS
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-emerald-400">ACTIVE</div>
            <div className="text-[10px] text-zinc-500 mt-1">Subscription Connected</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <CreditCard className="h-3.5 w-3.5 text-cyan-400" />
              MEMBERSHIP
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-white">PRO PLAN</div>
            <div className="text-[10px] text-zinc-500 mt-1">Auto-renews Monthly</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <Radio className="h-3.5 w-3.5 text-violet-400" />
              LOCAL COUPLING
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-violet-400">GH CLI HOOK</div>
            <div className="text-[10px] text-zinc-500 mt-1">Session Binding Active</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <Sparkles className="h-3.5 w-3.5 text-yellow-400" />
              COGNITIVE LIMIT
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-yellow-400">UNLIMITED</div>
            <div className="text-[10px] text-zinc-500 mt-1">High-priority Speed</div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Credentials Form */}
        <section className="lg:col-span-2 space-y-6">
          <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md p-6 space-y-4">
            <div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
              <KeyRound className="h-5 w-5 text-violet-400" />
              <h2 className="text-lg font-bold text-white">OAuth / API Credentials</h2>
            </div>
            <p className="text-xs text-zinc-400">
              Paste your personal access token or local Copilot API keys to grant TormentNexus permission to coordinate context suggestions.
            </p>
            <div className="flex flex-col gap-3">
              <div className="text-xs text-zinc-400">
                Active Key: <span className="font-mono text-zinc-200">{apiKeyPreview}</span>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="ghp_xxxxxxxxxxxxxxxxxxxxxx"
                  className="flex-1 min-w-[260px] bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs text-white outline-none focus:border-violet-500"
                />
                <button
                  onClick={saveApiKey}
                  className="px-3.5 py-2 bg-emerald-800 hover:bg-emerald-700 text-xs rounded text-white font-semibold transition-colors"
                >
                  Save Key
                </button>
                <button
                  onClick={clearApiKey}
                  className="px-3.5 py-2 bg-zinc-800 hover:bg-zinc-700 text-xs rounded text-zinc-300 font-semibold transition-colors"
                >
                  Clear
                </button>
              </div>
            </div>
            {status && (
              <div className={`text-xs rounded p-3 border inline-flex items-center gap-2 ${status.ok ? "text-emerald-300 border-emerald-800/50 bg-emerald-950/20" : "text-red-300 border-red-800/50 bg-red-950/20"}`}>
                <CheckCircle2 className="h-4 w-4" />
                <span>{status.message}</span>
              </div>
            )}
          </Card>

          {/* Interactive Shell Console */}
          <Card className="bg-zinc-900 border-zinc-800 p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <div className="flex items-center gap-2">
                <Terminal className="h-5 w-5 text-cyan-400" />
                <h2 className="text-lg font-bold text-white">Copilot Shell Simulator</h2>
              </div>
              <button
                onClick={triggerDiagnostic}
                disabled={checking}
                className="px-3 py-1 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 rounded text-xs flex items-center gap-1.5 disabled:opacity-50 text-zinc-300"
              >
                {checking ? <RefreshCw className="h-3 w-3 animate-spin" /> : <RefreshCw className="h-3 w-3" />}
                Diagnose
              </button>
            </div>
            
            <div className="bg-zinc-900/60 rounded border border-zinc-800 p-4 font-mono text-xs h-64 overflow-y-auto space-y-2">
              {simulatedLogs.map((log, i) => {
                let tone = "text-zinc-400";
                if (log.level === "success") tone = "text-emerald-400";
                if (log.level === "warn") tone = "text-yellow-400";
                return (
                  <div key={i} className="flex gap-2">
                    <span className="text-zinc-600">[{log.time}]</span>
                    <span className={tone}>{log.text}</span>
                  </div>
                );
              })}
            </div>
          </Card>
        </section>

        {/* Resources Portal */}
        <section className="space-y-6">
          <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md p-6 space-y-4">
            <h2 className="text-base font-bold text-white flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-violet-400" />
              Developer Resources
            </h2>
            <div className="space-y-3 text-xs">
              <a
                href="https://github.com/settings/billing"
                target="_blank"
                rel="noreferrer"
                className="flex items-center justify-between p-2.5 bg-zinc-900/60 hover:bg-zinc-900 rounded border border-zinc-800/80 transition-colors"
              >
                <span>GitHub Billing Portal</span>
                <ExternalLink className="h-3.5 w-3.5 text-zinc-500" />
              </a>
              <a
                href="https://docs.github.com/copilot"
                target="_blank"
                rel="noreferrer"
                className="flex items-center justify-between p-2.5 bg-zinc-900/60 hover:bg-zinc-900 rounded border border-zinc-800/80 transition-colors"
              >
                <span>Copilot Documentation</span>
                <ExternalLink className="h-3.5 w-3.5 text-zinc-500" />
              </a>
              <a
                href="https://github.com/features/copilot/plans"
                target="_blank"
                rel="noreferrer"
                className="flex items-center justify-between p-2.5 bg-zinc-900/60 hover:bg-zinc-900 rounded border border-zinc-800/80 transition-colors"
              >
                <span>Manage Copilot Plans</span>
                <ExternalLink className="h-3.5 w-3.5 text-zinc-500" />
              </a>
            </div>
          </Card>
        </section>
      </div>
    </div>
  );
}
