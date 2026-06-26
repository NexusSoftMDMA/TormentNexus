"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../ui/card';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Progress } from '../ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { 
  ShieldAlert, 
  Activity, 
  Network, 
  Database, 
  Zap, 
  Lock, 
  RefreshCcw, 
  Binary,
  Microscope
} from 'lucide-react';

export function InfrastructureDashboard() {
  const [loading, setLoading] = useState(true);
  const [stats, setStatus] = useState<any>(null);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchStats = () => {
    // In a real implementation, we'd have a specific endpoint for infra stats
    fetch('/api/system')
      .then(res => res.json())
      .then(data => {
        // Mocking Infra stats for Phase 14 visualization
        setStatus({
          p2p: { peers: 3, messages: 1284, status: 'stable' },
          consensus: { active: 0, total: 12, approved: 11 },
          replay: { recordings: 45, storageUsed: '124MB' },
          redis: { connected: true, keys: 856, memory: '42MB' },
          sandbox: { containers: 2, wasmIsolates: 0, status: 'ready' }
        });
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch infra stats:', err);
        setLoading(false);
      });
  };

  if (loading) return <div className="text-center py-20 text-slate-500">Initializing Infrastructure monitor...</div>;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-medium text-slate-400">P2P MESH</CardTitle>
            <Network className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-50">{stats.p2p.peers} Peers</div>
            <div className="flex items-center gap-2 mt-1">
              <Badge className="bg-emerald-500/10 text-emerald-500 border-none text-[9px] uppercase">{stats.p2p.status}</Badge>
              <span className="text-[10px] text-slate-500">{stats.p2p.messages} msgs</span>
            </div>
          </CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-medium text-slate-400">REDIS CONTEXT</CardTitle>
            <Database className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-50">{stats.redis.keys} Keys</div>
            <div className="flex items-center gap-2 mt-1">
              <Badge className="bg-blue-500/10 text-blue-500 border-none text-[9px] uppercase">Connected</Badge>
              <span className="text-[10px] text-slate-500">{stats.redis.memory}</span>
            </div>
          </CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-medium text-slate-400">CONSENSUS</CardTitle>
            <Lock className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-50">{stats.consensus.approved}/{stats.consensus.total}</div>
            <p className="text-[10px] text-slate-500 mt-1">Actions Approved</p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-medium text-slate-400">SANDBOX</CardTitle>
            <Binary className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-50">{stats.sandbox.containers} Active</div>
            <p className="text-[10px] text-slate-500 mt-1">Docker Containers</p>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="replay" className="space-y-4">
        <TabsList className="bg-slate-900 border-slate-800">
          <TabsTrigger value="replay">Deterministic Replay</TabsTrigger>
          <TabsTrigger value="policy">Policy-as-Code (Rego)</TabsTrigger>
          <TabsTrigger value="twin">Digital Twins</TabsTrigger>
          <TabsTrigger value="synthetic">Synthetic Data</TabsTrigger>
        </TabsList>

        <TabsContent value="replay" className="space-y-4">
          <Card className="bg-slate-950 border-slate-800">
            <CardHeader>
              <CardTitle className="text-slate-50 flex items-center gap-2">
                <RefreshCcw className="h-5 w-5 text-blue-400" />
                Trajectory Replay System
              </CardTitle>
              <CardDescription className="text-slate-400">
                Verify agent behavior by replaying recorded trajectories against policy constraints.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="p-4 rounded bg-slate-900 border border-slate-800 flex items-center justify-between">
                  <div>
                    <div className="text-sm font-bold text-slate-50">research-session-2026-01-14</div>
                    <div className="text-[10px] text-slate-500 font-mono uppercase">Agent: Researcher • 12 Steps</div>
                  </div>
                  <div className="flex gap-2">
                    <Button size="sm" variant="outline" className="h-8 text-[10px] border-slate-700">View Log</Button>
                    <Button size="sm" className="h-8 text-[10px] bg-blue-600 hover:bg-blue-500">Run Replay</Button>
                  </div>
                </div>
                <Button variant="ghost" className="w-full text-xs text-slate-500 border border-dashed border-slate-800">Load More Trajectories</Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="policy">
          <Card className="bg-slate-950 border-slate-800">
            <CardHeader>
              <CardTitle className="text-slate-50 flex items-center gap-2">
                <ShieldAlert className="h-5 w-5 text-red-400" />
                Open Policy Agent (Rego)
              </CardTitle>
              <CardDescription className="text-slate-400">
                Formal verification of agent plans and tool calls using declarative Rego policies.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="p-4 rounded bg-slate-900 border border-slate-800 font-mono text-[11px] text-emerald-400 h-64 overflow-y-auto">
                <pre>
{`package tormentnexus.agent.authz

import rego.v1

default allow := false

# Block dangerous tool calls
deny contains decision if {
    input.hook_event_name == "beforeToolExecution"
    dangerous_tools := ["execute_shell", "delete_file"]
    some tool in dangerous_tools
    input.tool_name == tool

    decision := {
        "rule_id": "TORMENTNEXUS-SECURITY-001",
        "reason": "Restricted tool access",
        "severity": "CRITICAL"
    }
}`}
                </pre>
              </div>
              <div className="mt-4 flex justify-between items-center">
                <Badge className="bg-blue-500/10 text-blue-400 border-none">OPA SERVER: ACTIVE</Badge>
                <Button size="sm" className="bg-slate-100 text-slate-900 hover:bg-white text-xs">Update Policies</Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="twin">
          <Card className="bg-slate-950 border-slate-800">
            <CardHeader>
              <CardTitle className="text-slate-50 flex items-center gap-2">
                <Microscope className="h-5 w-5 text-purple-400" />
                Digital Twin Simulation
              </CardTitle>
              <CardDescription className="text-slate-400">
                Test agents in isolated "Digital Twin" environments before live deployment.
              </CardDescription>
            </CardHeader>
            <CardContent className="py-10 text-center">
              <Activity className="h-12 w-12 text-slate-800 mx-auto mb-4" />
              <p className="text-sm text-slate-500">No active simulations. Start a twin to verify agent logic.</p>
              <Button size="sm" variant="secondary" className="mt-4">New Simulation</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="synthetic">
          <Card className="bg-slate-950 border-slate-800">
            <CardHeader>
              <CardTitle className="text-slate-50 flex items-center gap-2">
                <Binary className="h-5 w-5 text-emerald-400" />
                Synthetic Ecosystem
              </CardTitle>
              <CardDescription className="text-slate-400">
                Generate synthetic tasks and knowledge bases for agent training and benchmarking.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 rounded bg-slate-900 border border-slate-800">
                  <div className="text-xs font-bold text-slate-400 mb-2">GENERATION TASK</div>
                  <div className="text-sm text-slate-50 font-bold">Researcher Benchmarking Set</div>
                  <div className="text-[10px] text-slate-500 mt-1">100 Synthetic Tasks • READY</div>
                  <Button size="sm" variant="outline" className="w-full mt-4 h-8 text-[10px] border-slate-700">Export JSONL</Button>
                </div>
                <div className="p-4 rounded bg-slate-950 border border-slate-800 border-dashed flex flex-col items-center justify-center">
                  <Button size="sm" variant="ghost" className="text-slate-500 text-[10px]">+ Create Dataset</Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
