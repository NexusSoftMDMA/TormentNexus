"use client";

import React, { useEffect, useState, useCallback } from 'react';
import ReactFlow, { 
  Background, 
  Controls, 
  MiniMap, 
  useNodesState, 
  useEdgesState,
  MarkerType,
  type Node,
  type Edge
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../ui/card';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { 
  Layout, 
  Play, 
  Search, 
  GitBranch, 
  Activity, 
  CheckCircle2, 
  AlertCircle,
  Terminal,
  Code2
} from 'lucide-react';

export function ArchitectDashboard() {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [loading, setLoading] = useState(true);
  const [task, setTask] = useState('');
  const [session, setSession] = useState<any>(null);

  useEffect(() => {
    fetchGraph();
  }, []);

  const fetchGraph = async () => {
    try {
      const res = await fetch('/api/harness/graph');
      const data = await res.json();
      
      const flowNodes: Node[] = data.nodes.map((n: any, i: number) => ({
        id: n.id,
        data: { label: n.label },
        position: { x: (i % 5) * 200, y: Math.floor(i / 5) * 150 },
        style: { 
          background: n.type === 'typescript' ? '#1e3a8a' : '#1e293b',
          color: '#fff',
          border: '1px solid #334155',
          borderRadius: '8px',
          fontSize: '10px',
          width: 150
        }
      }));

      const flowEdges: Edge[] = data.edges.map((e: any) => ({
        id: `${e.source}-${e.target}`,
        source: e.source,
        target: e.target,
        animated: true,
        label: e.type,
        labelStyle: { fill: '#64748b', fontSize: 8 },
        markerEnd: { type: MarkerType.ArrowClosed, color: '#64748b' },
        style: { stroke: '#334155' }
      }));

      setNodes(flowNodes);
      setEdges(flowEdges);
      setLoading(false);
    } catch (err) {
      console.error('Failed to fetch repo graph:', err);
      setLoading(false);
    }
  };

  const handleStartSession = async () => {
    if (!task) return;
    setLoading(true);
    try {
      const res = await fetch('/api/architect/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ task })
      });
      const data = await res.json();
      setSession(data.session);
      setLoading(false);
    } catch (err) {
      console.error('Failed to start architect session:', err);
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col h-full space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-[800px]">
        {/* Left Column: Input & Plan */}
        <div className="lg:col-span-1 flex flex-col space-y-6">
          <Card className="bg-slate-900 border-slate-800">
            <CardHeader className="pb-4">
              <CardTitle className="text-lg font-bold text-slate-50 flex items-center gap-2">
                <Terminal className="h-5 w-5 text-emerald-400" />
                SuperAI Command
              </CardTitle>
              <CardDescription className="text-slate-400">Architect Mode (Reasoning + Editing)</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <textarea 
                className="w-full h-32 bg-slate-950 border border-slate-800 rounded-md p-3 text-sm text-slate-200 focus:ring-1 focus:ring-emerald-500 outline-none resize-none"
                placeholder="e.g. Implement a new user registration flow with email verification..."
                value={task}
                onChange={e => setTask(e.target.value)}
              />
              <Button 
                onClick={handleStartSession} 
                disabled={loading || !task}
                className="w-full bg-emerald-600 hover:bg-emerald-500 font-bold"
              >
                {loading ? 'Thinking...' : 'Start Reasoning'}
                <Play className="h-4 w-4 ml-2" />
              </Button>
            </CardContent>
          </Card>

          {session && (
            <Card className="bg-slate-900 border-slate-800 flex-1 overflow-hidden flex flex-col">
              <CardHeader className="pb-2 border-b border-slate-800">
                <div className="flex justify-between items-center">
                  <CardTitle className="text-sm font-bold text-slate-50">Implementation Plan</CardTitle>
                  <Badge variant="outline" className="bg-blue-500/10 text-blue-400 uppercase text-[9px]">
                    {session.status}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="flex-1 overflow-y-auto p-4 space-y-4">
                {session.plan ? (
                  <div className="space-y-4">
                    <p className="text-xs text-slate-400 leading-relaxed">{session.plan.description}</p>
                    <div className="space-y-2">
                      <div className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Affected Files</div>
                      {session.plan.files.map((f: any) => (
                        <div key={f.path} className="flex items-center gap-2 p-2 rounded bg-slate-950 border border-slate-800">
                          <Code2 className="h-3 w-3 text-blue-400" />
                          <span className="text-[10px] text-slate-300 font-mono truncate">{f.path}</span>
                          <Badge className="ml-auto text-[8px] bg-slate-800 text-slate-400 border-none">{f.action}</Badge>
                        </div>
                      ))}
                    </div>
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center h-full text-slate-600">
                    <Activity className="h-8 w-8 mb-2 animate-pulse" />
                    <p className="text-xs">Reasoning in progress...</p>
                  </div>
                )}
              </CardContent>
              <div className="p-4 border-t border-slate-800 flex gap-2">
                <Button variant="outline" className="flex-1 text-xs border-slate-700 h-8">Revise</Button>
                <Button className="flex-1 text-xs bg-blue-600 hover:bg-blue-500 h-8">Approve & Edit</Button>
              </div>
            </Card>
          )}
        </div>

        {/* Right Column: Repo Visualization */}
        <div className="lg:col-span-2 flex flex-col">
          <Card className="bg-slate-900 border-slate-800 flex-1 overflow-hidden flex flex-col">
            <CardHeader className="pb-2 flex flex-row items-center justify-between border-b border-slate-800">
              <div>
                <CardTitle className="text-lg font-bold text-slate-50 flex items-center gap-2">
                  <Layout className="h-5 w-5 text-blue-400" />
                  Codebase Graph
                </CardTitle>
                <CardDescription className="text-xs text-slate-500">Live dependency and symbol map of the repository.</CardDescription>
              </div>
              <div className="flex items-center gap-4">
                <div className="relative">
                  <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-500" />
                  <input className="bg-slate-950 border border-slate-800 rounded pl-7 pr-3 py-1 text-[10px] text-slate-300 outline-none w-48" placeholder="Search symbols..." />
                </div>
                <Button variant="ghost" size="icon" onClick={fetchGraph} className="h-8 w-8 text-slate-400 hover:text-white">
                  <Activity className="h-4 w-4" />
                </Button>
              </div>
            </CardHeader>
            <CardContent className="flex-1 p-0 relative">
              <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                fitView
              >
                <Background color="#1e293b" gap={20} />
                <Controls className="bg-slate-900 border-slate-800 fill-slate-400" />
                <MiniMap className="bg-slate-900 border-slate-800" nodeColor={(n) => n.style?.background as string} />
              </ReactFlow>
              
              {/* Overlay for LSP Diagnostics */}
              <div className="absolute top-4 right-4 space-y-2 pointer-events-none">
                <Badge className="bg-red-500/20 text-red-400 border-red-500/20 backdrop-blur-md flex items-center gap-1.5 text-[9px]">
                  <AlertCircle className="h-3 w-3" />
                  3 LSP ERRORS
                </Badge>
                <Badge className="bg-emerald-500/20 text-emerald-400 border-emerald-500/20 backdrop-blur-md flex items-center gap-1.5 text-[9px]">
                  <CheckCircle2 className="h-3 w-3" />
                  BUILD READY
                </Badge>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
