"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { 
  Wrench, 
  RefreshCcw, 
  Download, 
  CheckCircle2, 
  XCircle, 
  ExternalLink,
  Search,
  LayoutGrid,
  Activity
} from 'lucide-react';

export function ToolInventory() {
  const [tools, setTools] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [checking, setChecking] = useState<string | null>(null);

  useEffect(() => {
    fetchTools();
  }, []);

  const fetchTools = async () => {
    try {
      const res = await fetch('/api/inventory');
      const data = await res.json();
      setTools(data.tools || []);
      setLoading(false);
    } catch (err) {
      console.error('Failed to fetch tool inventory:', err);
      setLoading(false);
    }
  };

  const handleCheck = async (id?: string) => {
    setChecking(id || 'all');
    try {
      await fetch('/api/inventory/check', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id })
      });
      fetchTools();
    } catch (err) {
      console.error('Check failed:', err);
    } finally {
      setChecking(null);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'installed': return <CheckCircle2 className="h-4 w-4 text-emerald-500" />;
      case 'missing': return <XCircle className="h-4 w-4 text-red-500" />;
      default: return <RefreshCcw className="h-4 w-4 text-slate-500 animate-spin" />;
    }
  };

  return (
    <div className="space-y-6">
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-xl font-bold text-slate-50 flex items-center gap-2">
              <Wrench className="h-5 w-5 text-blue-400" />
              Local AI Tool Inventory
            </CardTitle>
            <CardDescription className="text-slate-400">
              Manage and verify the installation of CLI tools, background services, and MCP servers.
            </CardDescription>
          </div>
          <Button 
            variant="outline" 
            size="sm" 
            onClick={() => handleCheck()} 
            disabled={!!checking}
            className="border-slate-700 text-slate-300"
          >
            <RefreshCcw className={`h-4 w-4 mr-2 ${checking === 'all' ? 'animate-spin' : ''}`} />
            Re-scan All
          </Button>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader className="border-slate-800">
              <TableRow className="hover:bg-transparent border-slate-800">
                <TableHead className="text-slate-400">Tool Name</TableHead>
                <TableHead className="text-slate-400">Category</TableHead>
                <TableHead className="text-slate-400">Status</TableHead>
                <TableHead className="text-slate-400">Version</TableHead>
                <TableHead className="text-slate-400 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-10 text-slate-500 italic">Scanning system path...</TableCell>
                </TableRow>
              ) : tools.map((tool) => (
                <TableRow key={tool.id} className="border-slate-800 hover:bg-white/5">
                  <TableCell className="font-bold text-slate-50">{tool.name}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="bg-slate-800/50 text-slate-400 border-slate-700 text-[10px] uppercase">
                      {tool.category}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getStatusIcon(tool.status)}
                      <span className={`text-xs capitalize ${tool.status === 'installed' ? 'text-emerald-400' : 'text-slate-500'}`}>
                        {tool.status}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell className="font-mono text-[10px] text-slate-400">{tool.version || 'n/a'}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      {tool.status === 'missing' && tool.installCommand && (
                        <Button 
                          size="sm" 
                          className="h-7 text-[10px] bg-blue-600 hover:bg-blue-500"
                          onClick={() => {
                            navigator.clipboard.writeText(tool.installCommand);
                            alert(`Install command copied: ${tool.installCommand}`);
                          }}
                        >
                          <Download className="h-3 w-3 mr-1" /> Install
                        </Button>
                      )}
                      <Button 
                        variant="ghost" 
                        size="icon" 
                        className="h-7 w-7 text-slate-500 hover:text-white"
                        onClick={() => handleCheck(tool.id)}
                        disabled={checking === tool.id}
                      >
                        <RefreshCcw className={`h-3 w-3 ${checking === tool.id ? 'animate-spin' : ''}`} />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-bold text-slate-50 flex items-center gap-2">
              <LayoutGrid className="h-4 w-4 text-purple-400" />
              Environment & PATH
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-xs text-slate-400">Current TORMENTNEXUS service path includes 14 validated tool directories.</p>
            <div className="p-3 rounded bg-slate-950 border border-slate-800 font-mono text-[9px] text-slate-500 break-all overflow-hidden h-20">
              System Path: /usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin...
            </div>
            <Button variant="outline" size="sm" className="w-full border-slate-800 text-xs">Manage PATH Variables</Button>
          </CardContent>
        </Card>
        
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-bold text-slate-50 flex items-center gap-2">
              <Activity className="h-4 w-4 text-amber-400" />
              Process Guardian
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex justify-between items-center text-xs">
              <span className="text-slate-400">Background Services</span>
              <span className="text-slate-50">4 Active</span>
            </div>
            <div className="flex justify-between items-center text-xs">
              <span className="text-slate-400">Process Restarts (24h)</span>
              <span className="text-slate-50">0</span>
            </div>
            <div className="pt-2">
              <div className="flex justify-between text-[10px] mb-1">
                <span className="text-slate-500">Service Reliability</span>
                <span className="text-emerald-500">100%</span>
              </div>
              <div className="w-full bg-slate-800 h-1 rounded-full overflow-hidden">
                <div className="bg-emerald-500 h-full w-full" />
              </div>
            </div>
            <Button variant="outline" size="sm" className="w-full border-slate-800 text-xs">Open Process Monitor</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
