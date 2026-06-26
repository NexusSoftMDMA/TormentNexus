'use client';

import { useState } from 'react';
import {
  Cpu,
  HardDrive,
  Activity,
  Usb,
  Play,
  Square,
  RefreshCw,
  Wifi,
  WifiOff,
  Monitor,
  MemoryStick,
  Clock,
  Coins,
  Zap,
  Server,
  ChevronDown,
} from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { ScrollArea } from './ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import {
  useHardware,
  type SerialPortInfo,
  type ConnectedPort,
  type DiskInfo,
  type GpuInfo,
  type ActivityDataPoint,
} from '../lib/hooks/use-hardware';

const BAUD_RATES = [9600, 19200, 38400, 57600, 115200] as const;

function formatUptime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}h ${minutes}m`;
}

function PortCard({
  port,
  isConnected,
  connectedPort,
  onConnect,
  onDisconnect,
}: {
  port: SerialPortInfo;
  isConnected: boolean;
  connectedPort?: ConnectedPort;
  onConnect: (baudRate: number) => void;
  onDisconnect: () => void;
}) {
  const [showBaudSelector, setShowBaudSelector] = useState(false);
  const [selectedBaud, setSelectedBaud] = useState<number>(115200);

  return (
    <div className="p-4 rounded-lg bg-black/40 border border-white/5 space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Usb className="h-4 w-4 text-purple-400" />
          <span className="font-bold text-sm text-white font-mono">{port.path}</span>
        </div>
        <Badge className={`${isConnected ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'} border-0`}>
          {isConnected ? 'Connected' : 'Available'}
        </Badge>
      </div>

      <div className="text-xs font-mono text-white/60 space-y-1">
        {port.manufacturer && (
          <div className="flex items-center gap-2">
            <span className="text-white/40">Manufacturer:</span>
            <span className="text-purple-300">{port.manufacturer}</span>
          </div>
        )}
        {port.serialNumber && (
          <div className="flex items-center gap-2">
            <span className="text-white/40">Serial:</span>
            <span>{port.serialNumber}</span>
          </div>
        )}
        {(port.vendorId || port.productId) && (
          <div className="flex items-center gap-2">
            <span className="text-white/40">VID:PID:</span>
            <span>{port.vendorId || '----'}:{port.productId || '----'}</span>
          </div>
        )}
        {connectedPort && (
          <div className="flex items-center gap-2">
            <span className="text-white/40">Baud Rate:</span>
            <span className="text-green-400">{connectedPort.baudRate}</span>
          </div>
        )}
      </div>

      <div className="flex gap-2">
        {isConnected ? (
          <Button
            variant="outline"
            size="sm"
            onClick={onDisconnect}
            className="flex-1 text-xs bg-transparent border-white/10 hover:bg-red-600/20 hover:text-red-400"
          >
            <WifiOff className="h-3 w-3 mr-1" />
            Disconnect
          </Button>
        ) : (
          <>
            <div className="relative flex-1">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowBaudSelector(!showBaudSelector)}
                className="w-full text-xs bg-transparent border-white/10 hover:bg-white/5"
              >
                <span className="font-mono">{selectedBaud}</span>
                <ChevronDown className="h-3 w-3 ml-1" />
              </Button>
              {showBaudSelector && (
                <div className="absolute left-0 top-full mt-1 w-full bg-zinc-800 border border-white/10 rounded-md shadow-lg z-10">
                  {BAUD_RATES.map((rate) => (
                    <button
                      key={rate}
                      onClick={() => { setSelectedBaud(rate); setShowBaudSelector(false); }}
                      className="block w-full text-left px-3 py-1.5 text-xs text-white/80 hover:bg-white/10 font-mono"
                    >
                      {rate}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <Button
              size="sm"
              onClick={() => onConnect(selectedBaud)}
              className="flex-1 bg-purple-600 hover:bg-purple-500 text-xs"
            >
              <Wifi className="h-3 w-3 mr-1" />
              Connect
            </Button>
          </>
        )}
      </div>
    </div>
  );
}

function DiskCard({ disk }: { disk: DiskInfo }) {
  return (
    <div className="p-3 bg-black/40 rounded-lg border border-white/5">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <HardDrive className="h-4 w-4 text-blue-400" />
          <span className="font-mono text-sm text-white">{disk.mount}</span>
        </div>
        <span className="text-xs text-white/40">{disk.filesystem}</span>
      </div>
      <div className="space-y-1">
        <div className="flex justify-between text-xs">
          <span className="text-white/40">{disk.used} / {disk.size}</span>
          <span className={`font-mono ${disk.usagePercent > 90 ? 'text-red-400' : disk.usagePercent > 75 ? 'text-yellow-400' : 'text-white/60'}`}>
            {disk.usagePercent.toFixed(1)}%
          </span>
        </div>
        <div className="h-1.5 bg-black/60 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${
              disk.usagePercent > 90 ? 'bg-red-500' : disk.usagePercent > 75 ? 'bg-yellow-500' : 'bg-blue-500'
            }`}
            style={{ width: `${disk.usagePercent}%` }}
          />
        </div>
      </div>
    </div>
  );
}

function GpuCard({ gpu }: { gpu: GpuInfo }) {
  return (
    <div className="p-3 bg-black/40 rounded-lg border border-white/5">
      <div className="flex items-center gap-2 mb-2">
        <Monitor className="h-4 w-4 text-green-400" />
        <span className="text-sm text-white font-medium truncate">{gpu.model}</span>
      </div>
      <div className="text-xs font-mono text-white/60 space-y-1">
        <div className="flex items-center gap-2">
          <span className="text-white/40">Vendor:</span>
          <span>{gpu.vendor}</span>
        </div>
        {gpu.vram && (
          <div className="flex items-center gap-2">
            <span className="text-white/40">VRAM:</span>
            <span className="text-green-400">{gpu.vram}</span>
          </div>
        )}
        {gpu.driver && (
          <div className="flex items-center gap-2">
            <span className="text-white/40">Driver:</span>
            <span>{gpu.driver}</span>
          </div>
        )}
      </div>
    </div>
  );
}

function ActivityBar({ dataPoint, maxValue }: { dataPoint: ActivityDataPoint; maxValue: number }) {
  const cpuHeight = Math.max(4, (dataPoint.cpuUsage / maxValue) * 100);
  const memHeight = Math.max(4, (dataPoint.memoryUsage / maxValue) * 100);

  return (
    <div className="flex flex-col items-center gap-1 w-2">
      <div className="h-16 w-full flex flex-col justify-end gap-0.5">
        <div
          className="w-full bg-purple-500 rounded-t"
          style={{ height: `${cpuHeight}%` }}
          title={`CPU: ${dataPoint.cpuUsage.toFixed(1)}%`}
        />
        <div
          className="w-full bg-blue-500 rounded-b"
          style={{ height: `${memHeight}%` }}
          title={`Memory: ${dataPoint.memoryUsage.toFixed(1)}%`}
        />
      </div>
    </div>
  );
}

export function HardwareDashboard() {
  const {
    ports,
    connectedPorts,
    portsLoading,
    refreshPorts,
    connect,
    disconnect,
    systemInfo,
    systemLoading,
    refreshSystem,
    activity,
    activityLoading,
    miningStatus,
    miningLoading,
    startMiningOperation,
    stopMiningOperation,
    balance,
    balanceLoading,
    refreshBalance,
    refreshAll,
  } = useHardware();

  const isPortConnected = (path: string) => connectedPorts.some(p => p.path === path);
  const getConnectedPort = (path: string) => connectedPorts.find(p => p.path === path);

  return (
    <div className="flex-1 overflow-y-auto bg-gray-900">
      <div className="p-6 space-y-6 max-w-7xl mx-auto">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white tracking-tight">Hardware Dashboard</h1>
            <p className="text-white/40 text-sm">Manage serial connections, monitor system resources, and control mining operations</p>
          </div>
          <div className="flex items-center gap-3">
            <Badge variant="outline" className="border-white/20 text-white/60">
              {connectedPorts.length} connected / {ports.length} available
            </Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={refreshAll}
              className="bg-transparent border-white/10 hover:bg-white/5"
            >
              <RefreshCw className="h-4 w-4 mr-1" />
              Refresh All
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <Tabs defaultValue="ports" className="w-full">
              <TabsList className="bg-zinc-900/50 border border-white/10">
                <TabsTrigger value="ports" className="data-[state=active]:bg-purple-600">
                  <Usb className="h-4 w-4 mr-2" />
                  Serial Ports
                </TabsTrigger>
                <TabsTrigger value="system" className="data-[state=active]:bg-purple-600">
                  <Server className="h-4 w-4 mr-2" />
                  System Specs
                </TabsTrigger>
                <TabsTrigger value="activity" className="data-[state=active]:bg-purple-600">
                  <Activity className="h-4 w-4 mr-2" />
                  Activity
                </TabsTrigger>
              </TabsList>

              <TabsContent value="ports" className="mt-4">
                <Card className="bg-zinc-900/50 border-white/10">
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-white">Serial Ports</CardTitle>
                        <CardDescription className="text-white/40">
                          Available serial port connections
                        </CardDescription>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={refreshPorts}
                        disabled={portsLoading}
                        className="bg-transparent border-white/10 hover:bg-white/5"
                      >
                        <RefreshCw className={`h-4 w-4 mr-1 ${portsLoading ? 'animate-spin' : ''}`} />
                        Scan Ports
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {portsLoading ? (
                      <div className="text-center text-white/40 py-8">Scanning for serial ports...</div>
                    ) : ports.length === 0 ? (
                      <div className="text-center text-white/40 py-8 border border-dashed border-white/10 rounded-lg">
                        <Usb className="h-8 w-8 mx-auto mb-2 opacity-40" />
                        <p>No serial ports detected</p>
                        <p className="text-xs mt-1">Connect a device and click &quot;Scan Ports&quot;</p>
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {ports.map((port) => (
                          <PortCard
                            key={port.path}
                            port={port}
                            isConnected={isPortConnected(port.path)}
                            connectedPort={getConnectedPort(port.path)}
                            onConnect={(baudRate) => connect(port.path, baudRate)}
                            onDisconnect={() => disconnect(port.path)}
                          />
                        ))}
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="system" className="mt-4">
                <Card className="bg-zinc-900/50 border-white/10">
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-white">System Specifications</CardTitle>
                        <CardDescription className="text-white/40">
                          Hardware and operating system information
                        </CardDescription>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={refreshSystem}
                        disabled={systemLoading}
                        className="bg-transparent border-white/10 hover:bg-white/5"
                      >
                        <RefreshCw className={`h-4 w-4 mr-1 ${systemLoading ? 'animate-spin' : ''}`} />
                        Refresh
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {systemLoading ? (
                      <div className="text-center text-white/40 py-8">Loading system info...</div>
                    ) : !systemInfo ? (
                      <div className="text-center text-red-400 py-8">Failed to load system info</div>
                    ) : (
                      <div className="space-y-6">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                          <div className="p-4 bg-black/40 rounded-lg border border-white/5">
                            <div className="flex items-center gap-2 mb-3">
                              <Server className="h-4 w-4 text-purple-400" />
                              <span className="text-sm font-medium text-white">Operating System</span>
                            </div>
                            <div className="text-xs font-mono text-white/60 space-y-1">
                              <div>{systemInfo.os.distro} {systemInfo.os.release}</div>
                              <div className="text-white/40">{systemInfo.os.platform} ({systemInfo.os.arch})</div>
                              <div className="text-white/40">Host: {systemInfo.os.hostname}</div>
                            </div>
                          </div>

                          <div className="p-4 bg-black/40 rounded-lg border border-white/5">
                            <div className="flex items-center gap-2 mb-3">
                              <Cpu className="h-4 w-4 text-orange-400" />
                              <span className="text-sm font-medium text-white">Processor</span>
                            </div>
                            <div className="text-xs font-mono text-white/60 space-y-1">
                              <div className="truncate">{systemInfo.cpu.model}</div>
                              <div>{systemInfo.cpu.cores} cores @ {systemInfo.cpu.speed}</div>
                              <div className="flex items-center gap-2">
                                <span>Usage:</span>
                                <span className={systemInfo.cpu.usage > 80 ? 'text-red-400' : 'text-green-400'}>
                                  {systemInfo.cpu.usage.toFixed(1)}%
                                </span>
                              </div>
                            </div>
                          </div>

                          <div className="p-4 bg-black/40 rounded-lg border border-white/5">
                            <div className="flex items-center gap-2 mb-3">
                              <MemoryStick className="h-4 w-4 text-blue-400" />
                              <span className="text-sm font-medium text-white">Memory</span>
                            </div>
                            <div className="space-y-2">
                              <div className="flex justify-between text-xs">
                                <span className="text-white/40">{systemInfo.memory.used} / {systemInfo.memory.total}</span>
                                <span className={`font-mono ${systemInfo.memory.usagePercent > 90 ? 'text-red-400' : 'text-white/60'}`}>
                                  {systemInfo.memory.usagePercent.toFixed(1)}%
                                </span>
                              </div>
                              <div className="h-2 bg-black/60 rounded-full overflow-hidden">
                                <div
                                  className={`h-full rounded-full transition-all ${
                                    systemInfo.memory.usagePercent > 90 ? 'bg-red-500' : 'bg-blue-500'
                                  }`}
                                  style={{ width: `${systemInfo.memory.usagePercent}%` }}
                                />
                              </div>
                            </div>
                          </div>

                          <div className="p-4 bg-black/40 rounded-lg border border-white/5">
                            <div className="flex items-center gap-2 mb-3">
                              <Clock className="h-4 w-4 text-green-400" />
                              <span className="text-sm font-medium text-white">Uptime</span>
                            </div>
                            <div className="text-2xl font-bold text-white font-mono">
                              {formatUptime(systemInfo.uptime)}
                            </div>
                          </div>
                        </div>

                        {systemInfo.disks.length > 0 && (
                          <div>
                            <h3 className="text-sm font-medium text-white mb-3 flex items-center gap-2">
                              <HardDrive className="h-4 w-4 text-blue-400" />
                              Storage
                            </h3>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                              {systemInfo.disks.map((disk, index) => (
                                <DiskCard key={`${disk.mount}-${index}`} disk={disk} />
                              ))}
                            </div>
                          </div>
                        )}

                        {systemInfo.gpus.length > 0 && (
                          <div>
                            <h3 className="text-sm font-medium text-white mb-3 flex items-center gap-2">
                              <Monitor className="h-4 w-4 text-green-400" />
                              Graphics
                            </h3>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                              {systemInfo.gpus.map((gpu, index) => (
                                <GpuCard key={`${gpu.model}-${index}`} gpu={gpu} />
                              ))}
                            </div>
                          </div>
                        )}
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="activity" className="mt-4">
                <Card className="bg-zinc-900/50 border-white/10">
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-white">Activity Monitor</CardTitle>
                        <CardDescription className="text-white/40">
                          Real-time system activity and resource usage
                        </CardDescription>
                      </div>
                      <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2 text-xs">
                          <div className="h-2 w-2 rounded-full bg-purple-500" />
                          <span className="text-white/40">CPU</span>
                          <div className="h-2 w-2 rounded-full bg-blue-500 ml-2" />
                          <span className="text-white/40">Memory</span>
                        </div>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {activityLoading && !activity ? (
                      <div className="text-center text-white/40 py-8">Loading activity data...</div>
                    ) : !activity ? (
                      <div className="text-center text-red-400 py-8">Failed to load activity data</div>
                    ) : (
                      <div className="space-y-6">
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                            <div className="text-2xl font-bold text-purple-400 font-mono">
                              {activity.stats.avgCpuUsage.toFixed(1)}%
                            </div>
                            <div className="text-xs text-white/40">Avg CPU</div>
                          </div>
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                            <div className="text-2xl font-bold text-blue-400 font-mono">
                              {activity.stats.avgMemoryUsage.toFixed(1)}%
                            </div>
                            <div className="text-xs text-white/40">Avg Memory</div>
                          </div>
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                            <div className="text-2xl font-bold text-green-400 font-mono">
                              {activity.stats.totalNetworkIn}
                            </div>
                            <div className="text-xs text-white/40">Network In</div>
                          </div>
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                            <div className="text-2xl font-bold text-orange-400 font-mono">
                              {activity.stats.totalNetworkOut}
                            </div>
                            <div className="text-xs text-white/40">Network Out</div>
                          </div>
                        </div>

                        <div className="p-4 bg-black/40 rounded-lg border border-white/5">
                          <div className="text-xs text-white/40 mb-2">Recent Activity</div>
                          <ScrollArea className="w-full">
                            <div className="flex items-end gap-1 min-w-[400px]">
                              {activity.history.slice(-60).map((point, index) => (
                                <ActivityBar
                                  key={`${point.timestamp}-${index}`}
                                  dataPoint={point}
                                  maxValue={100}
                                />
                              ))}
                            </div>
                          </ScrollArea>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5">
                            <div className="text-xs text-white/40 mb-1">Peak CPU Usage</div>
                            <div className="text-lg font-bold text-red-400 font-mono">
                              {activity.stats.peakCpuUsage.toFixed(1)}%
                            </div>
                          </div>
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5">
                            <div className="text-xs text-white/40 mb-1">Peak Memory Usage</div>
                            <div className="text-lg font-bold text-red-400 font-mono">
                              {activity.stats.peakMemoryUsage.toFixed(1)}%
                            </div>
                          </div>
                        </div>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </div>

          <div className="space-y-6">
            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Zap className="h-5 w-5 text-yellow-400" />
                  Mining Operations
                </CardTitle>
                <CardDescription className="text-white/40">
                  Control and monitor mining activities
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {miningLoading && !miningStatus ? (
                  <div className="text-center text-white/40 py-4">Loading mining status...</div>
                ) : !miningStatus ? (
                  <div className="text-center text-white/40 py-4">Mining service unavailable</div>
                ) : (
                  <>
                    <div className="flex items-center justify-between p-3 bg-black/40 rounded-lg border border-white/5">
                      <div className="flex items-center gap-2">
                        <div className={`h-3 w-3 rounded-full ${miningStatus.isRunning ? 'bg-green-500 animate-pulse' : 'bg-gray-500'}`} />
                        <span className="text-sm text-white">
                          {miningStatus.isRunning ? 'Mining Active' : 'Mining Stopped'}
                        </span>
                      </div>
                      <Badge className={`${miningStatus.isRunning ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'} border-0`}>
                        {miningStatus.isRunning ? 'Running' : 'Idle'}
                      </Badge>
                    </div>

                    {miningStatus.stats && (
                      <div className="space-y-3">
                        <div className="p-3 bg-black/40 rounded-lg border border-white/5">
                          <div className="flex items-center justify-between mb-2">
                            <span className="text-xs text-white/40">Hashrate</span>
                            <span className="text-xs text-white/40">{miningStatus.stats.algorithm}</span>
                          </div>
                          <div className="text-2xl font-bold text-purple-400 font-mono">
                            {miningStatus.stats.hashrate} <span className="text-sm">{miningStatus.stats.hashrateUnit}</span>
                          </div>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                            <div className="text-lg font-bold text-green-400 font-mono">
                              {miningStatus.stats.shares.accepted}
                            </div>
                            <div className="text-xs text-white/40">Accepted</div>
                          </div>
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                            <div className="text-lg font-bold text-red-400 font-mono">
                              {miningStatus.stats.shares.rejected}
                            </div>
                            <div className="text-xs text-white/40">Rejected</div>
                          </div>
                        </div>

                        <div className="p-3 bg-black/40 rounded-lg border border-white/5">
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-white/40">Uptime</span>
                            <span className="text-sm font-mono text-white">
                              {formatUptime(miningStatus.stats.uptime)}
                            </span>
                          </div>
                        </div>

                        {miningStatus.stats.pool && (
                          <div className="p-3 bg-black/40 rounded-lg border border-white/5">
                            <div className="text-xs text-white/40 mb-1">Pool</div>
                            <div className="text-sm font-mono text-white truncate">
                              {miningStatus.stats.pool}
                            </div>
                          </div>
                        )}
                      </div>
                    )}

                    <div className="flex gap-2">
                      {miningStatus.isRunning ? (
                        <Button
                          onClick={stopMiningOperation}
                          disabled={miningLoading}
                          className="flex-1 bg-red-600 hover:bg-red-500"
                        >
                          <Square className="h-4 w-4 mr-1" />
                          Stop Mining
                        </Button>
                      ) : (
                        <Button
                          onClick={startMiningOperation}
                          disabled={miningLoading}
                          className="flex-1 bg-purple-600 hover:bg-purple-500"
                        >
                          <Play className="h-4 w-4 mr-1" />
                          Start Mining
                        </Button>
                      )}
                    </div>
                  </>
                )}
              </CardContent>
            </Card>

            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-white flex items-center gap-2">
                    <Coins className="h-5 w-5 text-yellow-400" />
                    Earnings
                  </CardTitle>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={refreshBalance}
                    disabled={balanceLoading}
                    className="bg-transparent border-white/10 hover:bg-white/5"
                  >
                    <RefreshCw className={`h-3 w-3 ${balanceLoading ? 'animate-spin' : ''}`} />
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {balanceLoading && !balance ? (
                  <div className="text-center text-white/40 py-4">Loading balance...</div>
                ) : !balance ? (
                  <div className="text-center text-white/40 py-4">Balance unavailable</div>
                ) : (
                  <div className="space-y-4">
                    <div className="p-4 bg-black/40 rounded-lg border border-white/5 text-center">
                      <div className="text-xs text-white/40 mb-1">Total Balance</div>
                      <div className="text-3xl font-bold text-white font-mono">
                        {balance.total.toFixed(8)}
                      </div>
                      <div className="text-sm text-white/40">{balance.currency}</div>
                    </div>

                    <div className="grid grid-cols-2 gap-3">
                      <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                        <div className="text-xs text-white/40 mb-1">Available</div>
                        <div className="text-lg font-bold text-green-400 font-mono">
                          {balance.available.toFixed(8)}
                        </div>
                      </div>
                      <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                        <div className="text-xs text-white/40 mb-1">Pending</div>
                        <div className="text-lg font-bold text-yellow-400 font-mono">
                          {balance.pending.toFixed(8)}
                        </div>
                      </div>
                    </div>

                    <div className="text-xs text-white/40 text-center">
                      Last updated: {new Date(balance.lastUpdated).toLocaleString()}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Activity className="h-5 w-5 text-blue-400" />
                  Quick Stats
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Connected Ports</span>
                    <span className="text-sm font-bold text-green-400">{connectedPorts.length}</span>
                  </div>
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Available Ports</span>
                    <span className="text-sm font-bold text-purple-400">{ports.length}</span>
                  </div>
                  {systemInfo && (
                    <>
                      <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                        <span className="text-xs text-white/40">CPU Usage</span>
                        <span className={`text-sm font-bold font-mono ${systemInfo.cpu.usage > 80 ? 'text-red-400' : 'text-green-400'}`}>
                          {systemInfo.cpu.usage.toFixed(1)}%
                        </span>
                      </div>
                      <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                        <span className="text-xs text-white/40">Memory Usage</span>
                        <span className={`text-sm font-bold font-mono ${systemInfo.memory.usagePercent > 80 ? 'text-red-400' : 'text-blue-400'}`}>
                          {systemInfo.memory.usagePercent.toFixed(1)}%
                        </span>
                      </div>
                    </>
                  )}
                  {miningStatus?.isRunning && miningStatus.stats && (
                    <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                      <span className="text-xs text-white/40">Hashrate</span>
                      <span className="text-sm font-bold text-yellow-400 font-mono">
                        {miningStatus.stats.hashrate} {miningStatus.stats.hashrateUnit}
                      </span>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
