"use client";

import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Brain, Database, Loader2, RefreshCw, Search, Zap, FileText, Settings, Cpu, Map } from "lucide-react";

type HydrationReport = {
  startedAt: string;
  completedAt: string;
  totalEntries: number;
  sections: string[];
  projectContextEntries: number;
  architectureEntries: number;
  agentInstructionsEntries: number;
  configEntries: number;
  repoGraphEntries: number;
  environmentEntries: number;
};

type HydrationStatus = {
  totalEntries: number;
  sections: string[];
  sectionCounts: Record<string, number>;
};

type HydrationEntry = {
  id: string;
  section: string;
  key: string;
  content: string;
  source: string;
  tags: string[];
  createdAt: string;
};

const SECTION_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  project_context: FileText,
  architecture: Map,
  agent_instructions: Brain,
  configuration: Settings,
  repo_graph: Database,
  environment: Cpu,
};

const SECTION_COLORS: Record<string, string> = {
  project_context: "text-blue-400",
  architecture: "text-purple-400",
  agent_instructions: "text-emerald-400",
  configuration: "text-amber-400",
  repo_graph: "text-cyan-400",
  environment: "text-rose-400",
};

export default function MemoryHydrationPage() {
  const [status, setStatus] = useState<HydrationStatus | null>(null);
  const [report, setReport] = useState<HydrationReport | null>(null);
  const [entries, setEntries] = useState<HydrationEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [hydrating, setHydrating] = useState(false);
  const [query, setQuery] = useState('');
  const [selectedSection, setSelectedSection] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = async () => {
    setLoading(true);
    try {
      const endpoints = [
        '/api/go/memory/hydration/status',
        '/api/go/api/memory/hydration/status',
      ];
      for (const endpoint of endpoints) {
        try {
          const resp = await fetch(endpoint, { signal: AbortSignal.timeout(3000) });
          if (resp.ok) {
            const data = await resp.json();
            if (data.success) {
              setStatus(data.data as HydrationStatus);
              setError(null);
              return;
            }
          }
        } catch { continue; }
      }
      setError('Could not reach hydration store');
    } finally {
      setLoading(false);
    }
  };

  const handleHydrate = async () => {
    setHydrating(true);
    setError(null);
    try {
      const endpoints = [
        '/api/go/memory/hydrate',
        '/api/go/api/memory/hydrate',
      ];
      for (const endpoint of endpoints) {
        try {
          const resp = await fetch(endpoint, {
            method: 'POST',
            signal: AbortSignal.timeout(30000),
          });
          if (resp.ok) {
            const data = await resp.json();
            if (data.success) {
              setReport(data.data as HydrationReport);
              await fetchStatus();
              return;
            }
          }
        } catch { continue; }
      }
      setError('Hydration failed — could not reach Go sidecar');
    } finally {
      setHydrating(false);
    }
  };

  const handleQuery = async () => {
    if (!query.trim()) return;
    setLoading(true);
    try {
      const params = new URLSearchParams({ query });
      if (selectedSection) params.set('section', selectedSection);
      const endpoints = [
        `/api/go/memory/hydration/query?${params}`,
        `/api/go/api/memory/hydration/query?${params}`,
      ];
      for (const endpoint of endpoints) {
        try {
          const resp = await fetch(endpoint, { signal: AbortSignal.timeout(5000) });
          if (resp.ok) {
            const data = await resp.json();
            if (data.success) {
              setEntries((data.data as HydrationEntry[]) || []);
              return;
            }
          }
        } catch { continue; }
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchStatus(); }, []);

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white">Memory Hydration</h1>
          <p className="text-zinc-500 mt-1">
            Bootstrap the Go sidecar context store with essential project knowledge for autonomous operation
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
            onClick={fetchStatus}
            disabled={loading}
          >
            <RefreshCw className="mr-2 h-4 w-4" /> Refresh
          </Button>
          <Button
            className="bg-emerald-600 hover:bg-emerald-500 text-white"
            onClick={handleHydrate}
            disabled={hydrating}
          >
            {hydrating ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Zap className="mr-2 h-4 w-4" />
            )}
            Hydrate Store
          </Button>
        </div>
      </div>

      {/* Hydration Report */}
      {report && (
        <Card className="bg-zinc-900 border-emerald-500/20">
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-emerald-300 font-semibold mb-3">
              <Zap className="h-4 w-4" /> Hydration Complete
            </div>
            <div className="grid grid-cols-3 md:grid-cols-6 gap-3 text-center">
              {[
                { label: "Project", value: report.projectContextEntries },
                { label: "Architecture", value: report.architectureEntries },
                { label: "Instructions", value: report.agentInstructionsEntries },
                { label: "Config", value: report.configEntries },
                { label: "Repo Graph", value: report.repoGraphEntries },
                { label: "Environment", value: report.environmentEntries },
              ].map(item => (
                <div key={item.label} className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3">
                  <div className="text-xs uppercase tracking-wider text-zinc-500">{item.label}</div>
                  <div className="mt-1 text-2xl font-semibold text-white">{item.value}</div>
                </div>
              ))}
            </div>
            <div className="mt-3 text-xs text-zinc-500">
              Total: {report.totalEntries} entries across {report.sections.length} sections · {report.startedAt} → {report.completedAt}
            </div>
          </CardContent>
        </Card>
      )}

      {error && (
        <Card className="bg-zinc-900 border-red-500/30">
          <CardContent className="p-4 text-red-300 text-sm">{error}</CardContent>
        </Card>
      )}

      {/* Status Overview */}
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-5">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-xs uppercase tracking-wider text-zinc-500">Total Entries</div>
                <div className="mt-1 text-3xl font-semibold text-white">{status?.totalEntries ?? 0}</div>
              </div>
              <Database className="h-5 w-5 text-blue-400" />
            </div>
          </CardContent>
        </Card>
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-5">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-xs uppercase tracking-wider text-zinc-500">Sections</div>
                <div className="mt-1 text-3xl font-semibold text-white">{status?.sections?.length ?? 0}</div>
              </div>
              <Brain className="h-5 w-5 text-purple-400" />
            </div>
          </CardContent>
        </Card>
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-5">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-xs uppercase tracking-wider text-zinc-500">Status</div>
                <div className="mt-1 text-lg font-semibold text-white">
                  {(status?.totalEntries ?? 0) > 0 ? "Hydrated" : "Empty"}
                </div>
              </div>
              {(status?.totalEntries ?? 0) > 0 ? (
                <Zap className="h-5 w-5 text-emerald-400" />
              ) : (
                <Loader2 className="h-5 w-5 text-zinc-500" />
              )}
            </div>
          </CardContent>
        </Card>
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-5">
            <div className="text-xs uppercase tracking-wider text-zinc-500 mb-2">Section Breakdown</div>
            <div className="space-y-1">
              {status?.sectionCounts && Object.entries(status.sectionCounts).map(([section, count]) => {
                const Icon = SECTION_ICONS[section] || Database;
                const color = SECTION_COLORS[section] || "text-zinc-400";
                return (
                  <button
                    key={section}
                    onClick={() => setSelectedSection(selectedSection === section ? null : section)}
                    className={`flex items-center gap-2 text-xs w-full text-left px-2 py-1 rounded transition-colors ${
                      selectedSection === section ? 'bg-zinc-800 text-white' : 'text-zinc-400 hover:text-zinc-200'
                    }`}
                  >
                    <Icon className={`h-3 w-3 ${color}`} />
                    <span className="flex-1">{section}</span>
                    <span className="font-mono">{count}</span>
                  </button>
                );
              })}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader className="pb-3">
          <CardTitle className="text-sm text-white flex items-center gap-2">
            <Search className="h-4 w-4 text-blue-400" /> Query Hydration Store
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search the hydration store..."
              className="flex-1 bg-zinc-950 border border-zinc-800 rounded-lg px-4 py-2.5 text-sm text-white focus:ring-2 focus:ring-blue-500 outline-none"
              onKeyDown={(e) => e.key === 'Enter' && handleQuery()}
            />
            <Button
              onClick={handleQuery}
              disabled={loading || !query.trim()}
              variant="outline"
              className="border-zinc-700 text-zinc-300 hover:bg-zinc-800"
            >
              Search
            </Button>
          </div>
          {selectedSection && (
            <div className="mt-2 text-xs text-zinc-500">
              Filtered to section: <span className="text-white">{selectedSection}</span>
              <button onClick={() => setSelectedSection(null)} className="ml-2 text-zinc-400 hover:text-white">✕ Clear</button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Query Results */}
      {entries.length > 0 && (
        <div className="space-y-3">
          {entries.map((entry) => {
            const Icon = SECTION_ICONS[entry.section] || Database;
            const color = SECTION_COLORS[entry.section] || "text-zinc-400";
            return (
              <Card key={entry.id} className="bg-zinc-900 border-zinc-800">
                <CardContent className="p-4">
                  <div className="flex items-start gap-3">
                    <Icon className={`h-4 w-4 mt-0.5 ${color}`} />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-mono text-sm text-blue-300">{entry.key}</span>
                        <span className="text-[10px] bg-zinc-800 px-2 py-0.5 rounded text-zinc-400 uppercase tracking-wider">
                          {entry.section}
                        </span>
                        <span className="text-[10px] bg-zinc-800 px-2 py-0.5 rounded text-zinc-500">
                          {entry.source}
                        </span>
                      </div>
                      <pre className="text-xs text-zinc-400 whitespace-pre-wrap break-words mt-2 max-h-32 overflow-auto">
                        {entry.content.length > 500 ? entry.content.slice(0, 500) + '...' : entry.content}
                      </pre>
                      {entry.tags && entry.tags.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-2">
                          {entry.tags.map(tag => (
                            <span key={tag} className="text-[10px] bg-zinc-950 border border-zinc-800 px-1.5 py-0.5 rounded text-zinc-500">
                              {tag}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
