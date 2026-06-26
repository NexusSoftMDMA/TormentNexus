"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { 
  Network, 
  ArrowRight, 
  ArrowLeft, 
  Clock, 
  Trash2, 
  Filter, 
  Activity,
  AlertCircle,
  Search
} from 'lucide-react';

interface TrafficFrame {
  id: string;
  timestamp: number;
  direction: 'request' | 'response' | 'notification';
  serverId: string;
  method: string;
  data: any;
  latency?: number;
  correlatedId?: string;
}

export function TrafficInspector() {
  const [frames, setFrames] = useState<TrafficFrame[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState({ serverId: '', method: '', direction: 'all' });
  const [selectedFrame, setSelectedFrame] = useState<TrafficFrame | null>(null);

  useEffect(() => {
    fetchFrames();
    const interval = setInterval(fetchFrames, 2000);
    return () => clearInterval(interval);
  }, []);

  const fetchFrames = async () => {
    try {
      const params = new URLSearchParams();
      if (filter.serverId) params.append('serverId', filter.serverId);
      if (filter.method) params.append('method', filter.method);
      if (filter.direction !== 'all') params.append('direction', filter.direction);

      const res = await fetch(`/api/traffic?${params.toString()}`);
      const data = await res.json();
      setFrames(data.frames || []);
      setLoading(false);
    } catch (err) {
      console.error('Failed to fetch traffic:', err);
      setLoading(false);
    }
  };

  const handleClear = async () => {
    await fetch('/api/traffic', { method: 'DELETE' });
    setFrames([]);
  };

  const getDirectionIcon = (direction: string) => {
    switch (direction) {
      case 'request': return <ArrowRight className="h-4 w-4 text-blue-500" />;
      case 'response': return <ArrowLeft className="h-4 w-4 text-emerald-500" />;
      case 'notification': return <Activity className="h-4 w-4 text-amber-500" />;
      default: return <AlertCircle className="h-4 w-4 text-slate-500" />;
    }
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-[800px]">
      <div className="lg:col-span-2 flex flex-col space-y-4">
        <Card className="bg-slate-900 border-slate-800 flex-1 flex flex-col overflow-hidden">
          <CardHeader className="pb-2 border-b border-slate-800 flex flex-row items-center justify-between">
            <div>
              <CardTitle className="text-lg font-bold text-slate-50 flex items-center gap-2">
                <Network className="h-5 w-5 text-blue-400" />
                MCP Traffic
              </CardTitle>
              <CardDescription className="text-xs text-slate-500">Real-time JSON-RPC inspection</CardDescription>
            </div>
            <div className="flex gap-2">
              <Button size="sm" variant="ghost" onClick={handleClear} className="h-8 w-8 p-0 text-slate-400 hover:text-red-400">
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </CardHeader>
          
          <div className="p-2 bg-slate-950 border-b border-slate-800 flex gap-2">
            <Input 
              placeholder="Server ID" 
              className="h-8 text-xs bg-slate-900 border-slate-800 w-32"
              value={filter.serverId}
              onChange={e => setFilter({...filter, serverId: e.target.value})}
            />
            <Input 
              placeholder="Method" 
              className="h-8 text-xs bg-slate-900 border-slate-800 w-32"
              value={filter.method}
              onChange={e => setFilter({...filter, method: e.target.value})}
            />
            <Select 
              value={filter.direction} 
              onValueChange={v => setFilter({...filter, direction: v})}
            >
              <SelectTrigger className="h-8 w-24 text-xs bg-slate-900 border-slate-800">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-slate-900 border-slate-800">
                <SelectItem value="all">All</SelectItem>
                <SelectItem value="request">Request</SelectItem>
                <SelectItem value="response">Response</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex-1 overflow-auto">
            <Table>
              <TableHeader className="bg-slate-950 sticky top-0 z-10">
                <TableRow className="border-slate-800 hover:bg-transparent">
                  <TableHead className="w-[50px] text-xs">Dir</TableHead>
                  <TableHead className="text-xs">Method</TableHead>
                  <TableHead className="text-xs">Server</TableHead>
                  <TableHead className="w-[80px] text-xs text-right">Latency</TableHead>
                  <TableHead className="w-[100px] text-xs text-right">Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {frames.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center py-10 text-slate-500 text-xs">No traffic recorded</TableCell>
                  </TableRow>
                ) : (
                  frames.slice().reverse().map(frame => (
                    <TableRow 
                      key={frame.id} 
                      className={`border-slate-800 cursor-pointer ${selectedFrame?.id === frame.id ? 'bg-blue-900/20' : 'hover:bg-slate-800/50'}`}
                      onClick={() => setSelectedFrame(frame)}
                    >
                      <TableCell className="py-2">{getDirectionIcon(frame.direction)}</TableCell>
                      <TableCell className="font-mono text-xs text-slate-300 py-2">{frame.method}</TableCell>
                      <TableCell className="text-xs text-slate-400 py-2">{frame.serverId}</TableCell>
                      <TableCell className="text-xs text-right font-mono text-slate-500 py-2">
                        {frame.latency ? `${frame.latency}ms` : '-'}
                      </TableCell>
                      <TableCell className="text-xs text-right text-slate-500 py-2">
                        {new Date(frame.timestamp).toLocaleTimeString()}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </Card>
      </div>

      <div className="lg:col-span-1 flex flex-col">
        <Card className="bg-slate-900 border-slate-800 h-full overflow-hidden flex flex-col">
          <CardHeader className="pb-2 border-b border-slate-800">
            <CardTitle className="text-sm font-bold text-slate-50">Payload Details</CardTitle>
          </CardHeader>
          <CardContent className="flex-1 p-0 overflow-auto bg-slate-950">
            {selectedFrame ? (
              <pre className="p-4 text-[10px] font-mono text-emerald-400 whitespace-pre-wrap break-all">
                {JSON.stringify(selectedFrame.data, null, 2)}
              </pre>
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-slate-600 p-8 text-center">
                <Search className="h-8 w-8 mb-2" />
                <p className="text-xs">Select a frame to inspect payload</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
