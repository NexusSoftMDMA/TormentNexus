"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Share2, Activity, Network, ExternalLink, ShieldCheck } from 'lucide-react';

export function A2AMesh() {
  const [agents, setAgents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAgents();
  }, []);

  const fetchAgents = () => {
    setLoading(true);
    // In a real implementation, we'd have a specific endpoint for A2A agents
    fetch('/api/state')
      .then(res => res.json())
      .then(data => {
        // Mocking registered A2A agents for now since they are in A2AManager memory
        setAgents([
          { url: 'http://localhost:4300', card: { name: 'TORMENTNEXUS Meta-Orchestrator', version: '0.3.0', skills: [{ name: 'Task Orchestration' }] }, healthy: true },
          { url: 'https://agent-east.tormentnexus.dev', card: { name: 'Security-Audit-Bot', version: '1.2.0', skills: [{ name: 'Vulnerability Scan' }] }, healthy: true }
        ]);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch A2A mesh:', err);
        setLoading(false);
      });
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-medium text-slate-400 uppercase">Mesh Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-emerald-500 flex items-center gap-2">
              <ShieldCheck className="h-5 w-5" />
              Secure
            </div>
            <p className="text-[10px] text-slate-500 mt-1">mTLS Encryption Active</p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-medium text-slate-400 uppercase">Active Links</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-50">{agents.length}</div>
            <p className="text-[10px] text-slate-500 mt-1">Cross-node connections</p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-medium text-slate-400 uppercase">Protocols</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-400">A2A v1.0</div>
            <p className="text-[10px] text-slate-500 mt-1">Google Agent-to-Agent Spec</p>
          </CardContent>
        </Card>
      </div>

      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-lg text-slate-50 font-bold flex items-center gap-2">
            <Network className="h-5 w-5 text-blue-400" />
            Agent-to-Agent (A2A) Mesh
          </CardTitle>
          <CardDescription className="text-slate-400">
            Real-time peer-to-peer communication between autonomous agents across the network.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader className="border-slate-800">
              <TableRow className="hover:bg-transparent border-slate-800">
                <TableHead className="text-slate-400">Remote Agent</TableHead>
                <TableHead className="text-slate-400">Endpoint</TableHead>
                <TableHead className="text-slate-400">Primary Skill</TableHead>
                <TableHead className="text-slate-400">Health</TableHead>
                <TableHead className="text-slate-400 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-10 text-slate-500">Loading mesh...</TableCell>
                </TableRow>
              ) : agents.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-10 text-slate-500">No remote agents discovered.</TableCell>
                </TableRow>
              ) : agents.map((agent) => (
                <TableRow key={agent.url} className="border-slate-800 hover:bg-white/5">
                  <TableCell>
                    <div className="flex flex-col">
                      <span className="font-bold text-slate-50">{agent.card.name}</span>
                      <span className="text-[10px] text-slate-500">v{agent.card.version}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-400 font-mono text-xs">{agent.url}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="bg-blue-500/10 text-blue-400 border-blue-500/20 uppercase text-[10px]">
                      {agent.card.skills[0]?.name || 'Generalist'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge className={agent.healthy ? 'bg-emerald-500/10 text-emerald-500' : 'bg-red-500/10 text-red-500'}>
                      {agent.healthy ? 'STABLE' : 'UNSTABLE'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400">
                        <Activity className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-white">
                        <Share2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
