'use client';

import { useState, useEffect } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  Terminal, Play, Square, RotateCw, Pause, CheckCircle2,
  AlertCircle, Activity, Clock, HardDrive, FileText, MoreHorizontal,
  Plus, Search, RefreshCw, Zap, Shield, TrendingUp, Circle
} from 'lucide-react';
import { io } from 'socket.io-client';
import { resolveCliApiBaseUrl } from '@/lib/endpoints';

export interface CLIDashboardBadge {
  label: string;
  variant: 'default' | 'secondary' | 'destructive' | 'outline' | null | undefined;
}

export interface CLIDashboardInstance {
  id: string;
  cliType: string;
  name: string;
  command: string;
  status: 'idle' | 'starting' | 'running' | 'paused' | 'stopped' | 'error' | 'crashed';
  pid?: number;
  port?: number;
  workingDirectory?: string;
  directoryList?: string[];
  createdAt: number;
  startedAt?: number;
  lastActivity?: number;
  healthStatus?: 'healthy' | 'degraded' | 'unhealthy';
  healthCheckCount: number;
  consecutiveHealthFailures: number;
  restartCount: number;
  lastHealthCheck?: number;
  logs: string[];
  supervisor?: string;
  tags: string[];
}

export interface CLIDashboardStats {
  total: number;
  running: number;
  idle: number;
  error: number;
  crashed: number;
}

const STATUS_COLORS = {
  idle: 'bg-gray-500',
  starting: 'bg-blue-500',
  running: 'bg-green-500',
  paused: 'bg-yellow-500',
  stopped: 'bg-slate-500',
  error: 'bg-red-500',
  crashed: 'bg-red-700'
};

const STATUS_BADGES = {
  idle: { label: 'Idle', variant: 'secondary' as const },
  starting: { label: 'Starting', variant: 'default' as const },
  running: { label: 'Running', variant: 'default' as const },
  paused: { label: 'Paused', variant: 'secondary' as const },
  stopped: { label: 'Stopped', variant: 'secondary' as const },
  error: { label: 'Error', variant: 'destructive' as const },
  crashed: { label: 'Crashed', variant: 'destructive' as const }
};

export default function CLIDashboard() {
  const [instances, setInstances] = useState<CLIDashboardInstance[]>([]);
  const [stats, setStats] = useState<CLIDashboardStats | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedInstance, setSelectedInstance] = useState<string | null>(null);
  const [logs, setLogs] = useState<string[]>([]);
  const [isConnected, setIsConnected] = useState(false);

  const getApiBaseUrl = () => resolveCliApiBaseUrl(process.env.NEXT_PUBLIC_API_URL);

  useEffect(() => {
    const baseUrl = getApiBaseUrl();
    const socket = io(baseUrl);

    socket.on('connect', () => {
      setIsConnected(true);
      fetchInstances();
    });

    socket.on('disconnect', () => {
      setIsConnected(false);
    });

    socket.on('cli:instances', (data: CLIDashboardInstance[]) => {
      setInstances(data);
      updateStats(data);
    });

    socket.on('cli:instance:updated', (instance: CLIDashboardInstance) => {
      setInstances(prev => prev.map(i => i.id === instance.id ? instance : i));
      updateStats(instances.map(i => i.id === instance.id ? instance : i));
    });

    socket.on('cli:instance:created', (instance: CLIDashboardInstance) => {
      setInstances(prev => [...prev, instance]);
      updateStats([...instances, instance]);
    });

    socket.on('cli:instance:deleted', (instanceId: string) => {
      setInstances(prev => prev.filter(i => i.id !== instanceId));
      updateStats(instances.filter(i => i.id !== instanceId));
      if (selectedInstance === instanceId) {
        setSelectedInstance(null);
      }
    });

    socket.on('cli:instance:logs', (data: { instanceId: string, logs: string[] }) => {
      if (data.instanceId === selectedInstance) {
        setLogs(data.logs);
      }
    });

    return () => {
      socket.disconnect();
    };
  }, [selectedInstance]);

  const fetchInstances = async () => {
    try {
      const res = await fetch(`${getApiBaseUrl()}/api/cli-supervisor/instances`);
      if (res.ok) {
        const data = await res.json();
        setInstances(data.instances || []);
        updateStats(data.instances || []);
      }
    } catch (error) {
      console.error('Failed to fetch instances:', error);
    }
  };

  const updateStats = (instances: CLIDashboardInstance[]) => {
    const newStats: CLIDashboardStats = {
      total: instances.length,
      running: instances.filter(i => i.status === 'running').length,
      idle: instances.filter(i => i.status === 'idle').length,
      error: instances.filter(i => i.status === 'error').length,
      crashed: instances.filter(i => i.status === 'crashed').length
    };
    setStats(newStats);
  };

  const handleInstanceAction = async (instanceId: string, action: 'start' | 'stop' | 'restart' | 'pause' | 'resume' | 'delete') => {
    try {
      if (action === 'delete') {
        await fetch(`${getApiBaseUrl()}/api/cli-supervisor/instances/${instanceId}`, {
          method: 'DELETE'
        });
      } else {
        await fetch(`${getApiBaseUrl()}/api/cli-supervisor/instances/${instanceId}/control`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ action })
        });
      }
    } catch (error) {
      console.error('Failed to execute action:', error);
    }
  };

  const fetchLogs = async (instanceId: string) => {
    try {
      const res = await fetch(`${getApiBaseUrl()}/api/cli-supervisor/instances/${instanceId}/logs`);
      if (res.ok) {
        const data = await res.json();
        setLogs(data.logs || []);
      }
    } catch (error) {
      console.error('Failed to fetch logs:', error);
    }
  };

  const filteredInstances = instances.filter(instance =>
    instance.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    instance.cliType.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return <Play className="h-4 w-4" />;
      case 'paused': return <Pause className="h-4 w-4" />;
      case 'error': return <AlertCircle className="h-4 w-4" />;
      case 'crashed': return <AlertCircle className="h-4 w-4" />;
      default: return <Square className="h-4 w-4" />;
    }
  };

  const formatDuration = (ms: number) => {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) return `${hours}h ${minutes % 60}m`;
    if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
    return `${seconds}s`;
  };

  return (
    <div className="container mx-auto p-6 h-screen flex flex-col gap-6 bg-slate-950 text-slate-50">
      <div className="flex justify-between items-center border-b border-slate-800 pb-4">
        <div className="flex items-center gap-4">
          <Terminal className="h-8 w-8 text-purple-400" />
          <h1 className="text-2xl font-bold">CLI Supervisor</h1>
          <Badge variant={isConnected ? 'default' : 'secondary'}>
            {isConnected ? 'Connected' : 'Disconnected'}
          </Badge>
        </div>
        <Button onClick={fetchInstances} variant="outline" size="sm">
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      <div className="flex flex-1 gap-6 overflow-hidden">
        <div className="flex-1 flex flex-col gap-4">
          <Card className="bg-slate-900 border-slate-800">
            <div className="flex items-center justify-between p-4 border-b border-slate-800">
              <h2 className="text-lg font-semibold">Instances ({filteredInstances.length})</h2>
              <div className="flex gap-2">
                <Input
                  placeholder="Search instances..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-64 bg-slate-950 border-slate-700"
                />
                <Button size="sm">
                  <Plus className="h-4 w-4 mr-2" />
                  New Instance
                </Button>
              </div>
            </div>
            <ScrollArea className="flex-1">
              <div className="p-4 space-y-3">
                {filteredInstances.map(instance => (
                  <Card
                    key={instance.id}
                    className={`cursor-pointer transition-all hover:scale-[1.02] ${
                      selectedInstance === instance.id ? 'ring-2 ring-purple-500' : ''
                    }`}
                    onClick={() => {
                      setSelectedInstance(instance.id);
                      fetchLogs(instance.id);
                    }}
                  >
                    <div className="flex items-start justify-between p-3">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          {getStatusIcon(instance.status)}
                          <h3 className="font-semibold text-lg">{instance.name}</h3>
                          <Badge variant={STATUS_BADGES[instance.status as keyof typeof STATUS_BADGES].variant}>
                            {STATUS_BADGES[instance.status as keyof typeof STATUS_BADGES].label}
                          </Badge>
                        </div>
                        <div className="flex gap-4 text-sm text-slate-400">
                          <span>Type: {instance.cliType}</span>
                          <span>Command: {instance.command}</span>
                          {instance.workingDirectory && (
                            <span>Dir: {instance.workingDirectory}</span>
                          )}
                        </div>
                        <div className="flex items-center gap-4 text-sm mt-2">
                          {instance.startedAt && (
                            <span className="flex items-center gap-1">
                              <Clock className="h-3 w-3" />
                              {formatDuration(Date.now() - instance.startedAt)}
                            </span>
                          )}
                          <span className="flex items-center gap-1">
                            <Activity className="h-3 w-3" />
                            {instance.logs.length} logs
                          </span>
                          {instance.restartCount > 0 && (
                            <span className="flex items-center gap-1">
                              <RotateCw className="h-3 w-3" />
                              Restarts: {instance.restartCount}
                            </span>
                          )}
                        </div>
                        {instance.tags.length > 0 && (
                          <div className="flex gap-1 flex-wrap mt-2">
                            {instance.tags.map(tag => (
                              <Badge key={tag} variant="outline" className="text-xs">
                                {tag}
                              </Badge>
                            ))}
                          </div>
                        )}
                      </div>
                      <div className="flex gap-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleInstanceAction(instance.id, instance.status === 'running' ? 'pause' : 'resume');
                          }}
                        >
                          {instance.status === 'running' ? (
                            <Pause className="h-4 w-4" />
                          ) : (
                            <Play className="h-4 w-4" />
                          )}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleInstanceAction(instance.id, 'restart');
                          }}
                        >
                          <RotateCw className="h-4 w-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleInstanceAction(instance.id, 'delete');
                          }}
                          className="text-red-400 hover:text-red-300"
                        >
                          <Circle className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  </Card>
                ))}
                {filteredInstances.length === 0 && (
                  <div className="text-center py-12 text-slate-400">
                    <Terminal className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p>No instances found</p>
                  </div>
                )}
              </div>
            </ScrollArea>
          </Card>
        </div>

        <div className="w-96 flex flex-col gap-4">
          {stats && (
            <Card className="bg-slate-900 border-slate-800">
              <div className="p-4 border-b border-slate-800">
                <h2 className="text-lg font-semibold flex items-center gap-2">
                  <TrendingUp className="h-5 w-5 text-green-400" />
                  Statistics
                </h2>
              </div>
              <div className="p-4 space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-slate-400">Total Instances</span>
                  <span className="text-2xl font-bold text-slate-100">{stats.total}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-slate-400">Running</span>
                  <span className="text-2xl font-bold text-green-400">{stats.running}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-slate-400">Idle</span>
                  <span className="text-2xl font-bold text-gray-400">{stats.idle}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-slate-400">Errors</span>
                  <span className="text-2xl font-bold text-red-400">{stats.error}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-slate-400">Crashed</span>
                  <span className="text-2xl font-bold text-red-600">{stats.crashed}</span>
                </div>
              </div>
            </Card>
          )}

          {selectedInstance && (
            <Card className="flex-1 bg-slate-900 border-slate-800">
              <div className="flex items-center justify-between p-4 border-b border-slate-800">
                <h2 className="text-lg font-semibold flex items-center gap-2">
                  <FileText className="h-5 w-5 text-blue-400" />
                  Instance Logs
                </h2>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => fetchLogs(selectedInstance)}
                >
                  <RefreshCw className="h-4 w-4" />
                </Button>
              </div>
              <ScrollArea className="flex-1">
                <div className="p-4">
                  {logs.length === 0 ? (
                    <div className="text-center py-8 text-slate-400">
                      No logs available
                    </div>
                  ) : (
                    <div className="font-mono text-sm space-y-1">
                      {logs.map((log, index) => (
                        <div key={index} className="p-2 hover:bg-slate-800 rounded text-slate-300">
                          {log}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </ScrollArea>
            </Card>
          )}

          {!selectedInstance && (
            <Card className="bg-slate-900 border-slate-800">
              <div className="p-6 text-center">
                <Terminal className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p className="text-slate-400">Select an instance to view logs</p>
              </div>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
