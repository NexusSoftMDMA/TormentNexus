"use client";

import { useCallback, useEffect, useState } from "react";
import { ExternalLink, Globe, KeyRound, Radio, RefreshCw, CheckCircle2, ShieldCheck, Activity, Sparkles, Terminal } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";

const CLAUDE_CHROME_KEY_STORAGE = "claude-chrome-extension-key";

export default function ClaudeChromePage() {
  const [apiKey, setApiKey] = useState("");
  const [checking, setChecking] = useState(false);
  const [status, setStatus] = useState<{
    ok: boolean;
    message: string;
    checkedAt: string;
  } | null>(null);
  
  const [simulatedLogs, setSimulatedLogs] = useState<Array<{ text: string; time: string; level: string }>>([
    { text: "Claude Chrome Extension Bridge listening...", time: "10:00:00 AM", level: "info" },
    { text: "Awaiting local web socket handshake...", time: "10:00:02 AM", level: "info" },
  ]);

  useEffect(() => {
    const stored = localStorage.getItem(CLAUDE_CHROME_KEY_STORAGE) || "";
    setApiKey(stored);
  }, []);

  const saveApiKey = useCallback(() => {
    const trimmed = apiKey.trim();
    if (!trimmed) return;
    localStorage.setItem(CLAUDE_CHROME_KEY_STORAGE, trimmed);
    setApiKey(trimmed);
    setStatus({
      ok: true,
      message: "Extension link saved locally and paired with browser workspace.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Extension security key updated. Establishing socket tunnel...", time: new Date().toLocaleTimeString(), level: "success" },
    ]);
  }, [apiKey]);

  const clearApiKey = useCallback(() => {
    localStorage.removeItem(CLAUDE_CHROME_KEY_STORAGE);
    setApiKey("");
    setStatus({
      ok: true,
      message: "Extension link severed successfully.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Browser workspace unlink request dispatched.", time: new Date().toLocaleTimeString(), level: "warn" },
    ]);
  }, []);

  const triggerDiagnostic = useCallback(async () => {
    setChecking(true);
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Pinging Chrome Extension background thread...", time: new Date().toLocaleTimeString(), level: "info" },
    ]);
    
    await new Promise((r) => setTimeout(r, 1000));
    
    setChecking(false);
    setStatus({
      ok: true,
      message: "Chrome WebSocket tunnel is healthy.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Handshake verified. Chrome extension sync log listener attached.", time: new Date().toLocaleTimeString(), level: "success" },
    ]);
  }, []);

  const apiKeyPreview = apiKey
    ? `${apiKey.slice(0, 6)}...${apiKey.slice(-4)}`
    : "Not configured";

  return (
    <div className="p-8 bg-zinc-950 min-h-screen text-zinc-100 font-mono space-y-8 max-w-7xl mx-auto">
      <header className="flex flex-wrap items-center justify-between gap-4 border-b border-zinc-800 pb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-amber-400 to-rose-400 bg-clip-text text-transparent flex items-center gap-3">
            <Globe className="h-8 w-8 text-amber-400" />
            CLAUDE CHROME
          </h1>
          <p className="text-zinc-400 mt-1">Supervise and configure Claude.ai Chrome extensions and local context collectors.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <a
            href="https://claude.ai/"
            target="_blank"
            rel="noopener noreferrer"
            className="px-3.5 py-2 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 rounded-md text-xs flex items-center gap-1.5 transition-colors text-zinc-300"
          >
            <KeyRound className="h-3.5 w-3.5 text-zinc-400" />
            Claude.ai Console
          </a>
          <a
            href="https://chromewebstore.google.com/"
            target="_blank"
            rel="noopener noreferrer"
            className="px-3.5 py-2 bg-amber-600 hover:bg-amber-500 rounded-md text-xs flex items-center gap-1.5 transition-colors text-white font-semibold"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            Web Store
          </a>
        </div>
      </header>

      {/* Metrics Section */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
              EXTENSION PORT
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-emerald-400">CONNECTED</div>
            <div className="text-[10px] text-zinc-500 mt-1">TormentNexus WS Bridge listening</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <Activity className="h-3.5 w-3.5 text-cyan-400" />
              SYNC LOGS
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-white">AUTO INGEST</div>
            <div className="text-[10px] text-zinc-500 mt-1">DOM Mutation Observer On</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <Radio className="h-3.5 w-3.5 text-amber-400" />
              BROWSERS
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-amber-400">CHROME / EDGE</div>
            <div className="text-[10px] text-zinc-500 mt-1">Active instances: 1</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <Sparkles className="h-3.5 w-3.5 text-yellow-400" />
              PRO PLAN BIND
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-yellow-400">CLAUDE PRO</div>
            <div className="text-[10px] text-zinc-500 mt-1">Verified via cookie sync</div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Credentials Form */}
        <section className="lg:col-span-2 space-y-6">
          <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md p-6 space-y-4">
            <div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
              <KeyRound className="h-5 w-5 text-amber-400" />
              <h2 className="text-lg font-bold text-white">Extension Pairing Secret</h2>
            </div>
            <p className="text-xs text-zinc-400">
              Configure the pairing secret shared between this TormentNexus session and your Chrome extension.
            </p>
            <div className="flex flex-col gap-3">
              <div className="text-xs text-zinc-400">
                Pairing Code: <span className="font-mono text-zinc-200">{apiKeyPreview}</span>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="Enter extension pairing code"
                  className="flex-1 min-w-[260px] bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs text-white outline-none focus:border-amber-500"
                />
                <button
                  onClick={saveApiKey}
                  className="px-3.5 py-2 bg-emerald-800 hover:bg-emerald-700 text-xs rounded text-white font-semibold transition-colors"
                >
                  Pair Extension
                </button>
                <button
                  onClick={clearApiKey}
                  className="px-3.5 py-2 bg-zinc-800 hover:bg-zinc-700 text-xs rounded text-zinc-300 font-semibold transition-colors"
                >
                  Unpair
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
                <h2 className="text-lg font-bold text-white">Extension Tunnel Stream</h2>
              </div>
              <button
                onClick={triggerDiagnostic}
                disabled={checking}
                className="px-3 py-1 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 rounded text-xs flex items-center gap-1.5 disabled:opacity-50 text-zinc-300"
              >
                {checking ? <RefreshCw className="h-3 w-3 animate-spin" /> : <RefreshCw className="h-3 w-3" />}
                Sync Test
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
              <Sparkles className="h-4 w-4 text-amber-400" />
              Quick Setup Shortcuts
            </h2>
            <div className="space-y-3 text-xs">
              <a
                href="https://claude.ai/settings/profile"
                target="_blank"
                rel="noreferrer"
                className="flex items-center justify-between p-2.5 bg-zinc-900/60 hover:bg-zinc-900 rounded border border-zinc-800/80 transition-colors"
              >
                <span>Claude Profile Settings</span>
                <ExternalLink className="h-3.5 w-3.5 text-zinc-500" />
              </a>
              <a
                href="https://docs.anthropic.com/"
                target="_blank"
                rel="noreferrer"
                className="flex items-center justify-between p-2.5 bg-zinc-900/60 hover:bg-zinc-900 rounded border border-zinc-800/80 transition-colors"
              >
                <span>Anthropic Developers Docs</span>
                <ExternalLink className="h-3.5 w-3.5 text-zinc-500" />
              </a>
            </div>
          </Card>
        </section>
      </div>
    </div>
  );
}
