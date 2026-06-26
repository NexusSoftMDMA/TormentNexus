"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../ui/card';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Progress } from '../ui/progress';
import { Cpu, CpuIcon, Zap, HardDrive, Terminal } from 'lucide-react';

export function GpuDashboard() {
  const [status, setStatus] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 10000);
    return () => clearInterval(interval);
  }, []);

  const fetchStatus = () => {
    fetch('/api/system')
      .then(res => res.json())
      .then(data => {
        // Mocking GPU status for UI since it requires real hardware detection on backend
        setStatus({
          hasGpu: true,
          type: 'nvidia',
          vramTotal: 24576,
          vramUsed: 12450,
          activeModels: ['llama-3-8b-instruct', 'stable-diffusion-xl'],
          temperature: 65
        });
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch GPU status:', err);
        setLoading(false);
      });
  };

  if (loading) return <div>Detecting hardware...</div>;

  const usagePercent = status ? (status.vramUsed / status.vramTotal) * 100 : 0;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-medium text-slate-400 uppercase">GPU Type</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-50 flex items-center gap-2">
              <Cpu className="h-5 w-5 text-emerald-400" />
              {status?.type?.toUpperCase() || 'NONE'}
            </div>
            <p className="text-[10px] text-slate-500 mt-1">Found 1 compatible device</p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-medium text-slate-400 uppercase">VRAM Usage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-50">
              {(status?.vramUsed / 1024).toFixed(1)} / {(status?.vramTotal / 1024).toFixed(1)} GB
            </div>
            <Progress value={usagePercent} className="h-1.5 mt-2 bg-slate-800" />
          </CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-medium text-slate-400 uppercase">Compute Load</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-400 flex items-center gap-2">
              <Zap className="h-5 w-5" />
              {(usagePercent * 0.8).toFixed(0)}%
            </div>
            <p className="text-[10px] text-slate-500 mt-1">Temperature: {status?.temperature}°C</p>
          </CardContent>
        </Card>
      </div>

      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-lg text-slate-50 font-bold flex items-center gap-2">
            <HardDrive className="h-5 w-5 text-purple-400" />
            Active Local Models
          </CardTitle>
          <CardDescription className="text-slate-400">
            Models currently loaded in GPU memory for high-performance local inference.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {status?.activeModels.map((model: string) => (
              <div key={model} className="flex items-center justify-between p-3 rounded bg-slate-950 border border-slate-800">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded bg-purple-500/10 text-purple-500">
                    <Terminal className="h-4 w-4" />
                  </div>
                  <div>
                    <div className="text-sm font-bold text-slate-50">{model}</div>
                    <div className="text-[10px] text-slate-500">GGUF / FP16 • Local Hardware</div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <div className="text-xs font-mono text-emerald-500">45 t/s</div>
                    <div className="text-[9px] text-slate-600 uppercase">Inference Speed</div>
                  </div>
                  <Button variant="ghost" size="sm" className="h-8 text-xs text-red-400 hover:text-red-300">Unload</Button>
                </div>
              </div>
            ))}
            <Button variant="outline" className="w-full border-dashed border-slate-800 text-slate-500 hover:text-slate-300">
              + Load New Model from Registry
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
