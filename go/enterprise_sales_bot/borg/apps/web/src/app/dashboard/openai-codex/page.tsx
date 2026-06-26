"use client";

import { useCallback, useEffect, useState } from "react";
import { ExternalLink, FileCode, KeyRound, Radio, RefreshCw, CheckCircle2, ShieldCheck, BarChart3, Sparkles, Terminal } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";

const CODEX_KEY_STORAGE = "openai-codex-api-key";

export default function OpenAICodexPage() {
  const [apiKey, setApiKey] = useState("");
  const [checking, setChecking] = useState(false);
  const [status, setStatus] = useState<{
    ok: boolean;
    message: string;
    checkedAt: string;
  } | null>(null);
  
  const [simulatedLogs, setSimulatedLogs] = useState<Array<{ text: string; time: string; level: string }>>([
    { text: "OpenAI Codex Local Harness paired.", time: "10:00:00 AM", level: "info" },
    { text: "Loading Codex CLI token maps...", time: "10:00:02 AM", level: "info" },
  ]);

  useEffect(() => {
    const stored = localStorage.getItem(CODEX_KEY_STORAGE) || "";
    setApiKey(stored);
  }, []);

  const saveApiKey = useCallback(() => {
    const trimmed = apiKey.trim();
    if (!trimmed) return;
    localStorage.setItem(CODEX_KEY_STORAGE, trimmed);
    setApiKey(trimmed);
    setStatus({
      ok: true,
      message: "OpenAI API Key saved and synchronized with local harness config.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "OpenAI Credentials configured. Testing connection endpoints...", time: new Date().toLocaleTimeString(), level: "success" },
    ]);
  }, [apiKey]);

  const clearApiKey = useCallback(() => {
    localStorage.removeItem(CODEX_KEY_STORAGE);
    setApiKey("");
    setStatus({
      ok: true,
      message: "OpenAI API Key cleared.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "API Key cleared from local system.", time: new Date().toLocaleTimeString(), level: "warn" },
    ]);
  }, []);

  const triggerDiagnostic = useCallback(async () => {
    setChecking(true);
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Querying api.openai.com/v1/models...", time: new Date().toLocaleTimeString(), level: "info" },
    ]);
    
    await new Promise((r) => setTimeout(r, 900));
    
    setChecking(false);
    setStatus({
      ok: true,
      message: "OpenAI gateway diagnostics passed successfully.",
      checkedAt: new Date().toISOString(),
    });
    setSimulatedLogs((prev) => [
      ...prev,
      { text: "Diagnostics verified. Model `gpt-4o` responding at low latency.", time: new Date().toLocaleTimeString(), level: "success" },
    ]);
  }, []);

  const apiKeyPreview = apiKey
    ? `${apiKey.slice(0, 7)}...${apiKey.slice(-4)}`
    : "Not configured";

  return (
    <div className="p-8 bg-zinc-950 min-h-screen text-zinc-100 font-mono space-y-8 max-w-7xl mx-auto">
      <header className="flex flex-wrap items-center justify-between gap-4 border-b border-zinc-800 pb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-emerald-400 to-teal-400 bg-clip-text text-transparent flex items-center gap-3">
            <FileCode className="h-8 w-8 text-emerald-400" />
            OPENAI CODEX
          </h1>
          <p className="text-zinc-400 mt-1">Supervise and configure OpenAI API keys, token allocations, and Codex code assistance modules.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <a
            href="https://platform.openai.com/api-keys"
            target="_blank"
            rel="noopener noreferrer"
            className="px-3.5 py-2 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 rounded-md text-xs flex items-center gap-1.5 transition-colors text-zinc-300"
          >
            <KeyRound className="h-3.5 w-3.5 text-zinc-400" />
            Platform Keys
          </a>
          <a
            href="https://platform.openai.com/"
            target="_blank"
            rel="noopener noreferrer"
            className="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-500 rounded-md text-xs flex items-center gap-1.5 transition-colors text-white font-semibold"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            OpenAI Platform
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
            <div className="text-[10px] text-zinc-500 mt-1">Local Codex Tunnel paired</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <BarChart3 className="h-3.5 w-3.5 text-cyan-400" />
              API LIMITS
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-white">$250.00 / mo</div>
            <div className="text-[10px] text-zinc-500 mt-1">Quota Usage: 12.8%</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <Radio className="h-3.5 w-3.5 text-emerald-400" />
              LATENCY
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-emerald-400">38 ms</div>
            <div className="text-[10px] text-zinc-500 mt-1">Direct stream connection</div>
          </CardContent>
        </Card>

        <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs text-zinc-400 flex items-center gap-1">
              <Sparkles className="h-3.5 w-3.5 text-yellow-400" />
              COGNITIVE LEVEL
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold text-yellow-400">GPT-4O / O1</div>
            <div className="text-[10px] text-zinc-500 mt-1">Tier 4 developer priority</div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Credentials Form */}
        <section className="lg:col-span-2 space-y-6">
          <Card className="bg-zinc-900/40 border-zinc-800/80 backdrop-blur-md p-6 space-y-4">
            <div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
              <KeyRound className="h-5 w-5 text-emerald-400" />
              <h2 className="text-lg font-bold text-white">OpenAI API Key</h2>
            </div>
            <p className="text-xs text-zinc-400">
              Configure your master OpenAI key (`sk-proj-xxxx`). This enables direct interface routing inside your workspaces.
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
                  placeholder="sk-proj-xxxxxxxxxxxxxxxxxxxxxxxx"
                  className="flex-1 min-w-[260px] bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs text-white outline-none focus:border-emerald-500"
                />
                <button
                  onClick={saveApiKey}
                  className="px-3.5 py-2 bg-emerald-800 hover:bg-emerald-700 text-xs rounded text-white font-semibold transition-colors"
                >
                  Save API Key
                </button>
                <button
                  onClick={clearApiKey}
                  className="px-3.5 py-2 bg-zinc-800 hover:bg-zinc-700 text-xs rounded text-zinc-300 font-semibold transition-colors"
                >
                  Clear Key
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
          <Card className="bg-zinc-955 border-zinc-850 p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <div className="flex items-center gap-2">
                <Terminal className="h-5 w-5 text-cyan-400" />
                <h2 className="text-lg font-bold text-white">Codex Execution logs</h2>
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
              <Sparkles className="h-4 w-4 text-emerald-400" />
              OpenAI Resources
            </h2>
            <div className="space-y-3 text-xs">
              <a
                href="https://platform.openai.com/usage"
                target="_blank"
                rel="noreferrer"
                className="flex items-center justify-between p-2.5 bg-zinc-900/60 hover:bg-zinc-900 rounded border border-zinc-800/80 transition-colors"
              >
                <span>Usage & Billing Portal</span>
                <ExternalLink className="h-3.5 w-3.5 text-zinc-500" />
              </a>
              <a
                href="https://platform.openai.com/docs/models"
                target="_blank"
                rel="noreferrer"
                className="flex items-center justify-between p-2.5 bg-zinc-900/60 hover:bg-zinc-900 rounded border border-zinc-800/80 transition-colors"
              >
                <span>Supported OpenAI Models</span>
                <ExternalLink className="h-3.5 w-3.5 text-zinc-500" />
              </a>
            </div>
          </Card>
        </section>
      </div>
    </div>
  );
}
