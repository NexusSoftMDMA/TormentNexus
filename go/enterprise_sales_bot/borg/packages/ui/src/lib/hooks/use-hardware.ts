'use client';

import { useState, useEffect, useCallback } from 'react';

export interface SerialPortInfo {
  path: string;
  manufacturer?: string;
  serialNumber?: string;
  vendorId?: string;
  productId?: string;
  pnpId?: string;
}

export interface ConnectedPort extends SerialPortInfo {
  baudRate: number;
  isConnected: boolean;
  connectedAt: string;
}

export interface CpuInfo {
  model: string;
  cores: number;
  speed: string;
  usage: number;
}

export interface MemoryInfo {
  total: string;
  used: string;
  free: string;
  usagePercent: number;
}

export interface DiskInfo {
  filesystem: string;
  size: string;
  used: string;
  available: string;
  usagePercent: number;
  mount: string;
}

export interface GpuInfo {
  model: string;
  vendor: string;
  vram?: string;
  driver?: string;
}

export interface SystemInfo {
  os: {
    platform: string;
    distro: string;
    release: string;
    arch: string;
    hostname: string;
  };
  cpu: CpuInfo;
  memory: MemoryInfo;
  disks: DiskInfo[];
  gpus: GpuInfo[];
  uptime: number;
}

export interface ActivityDataPoint {
  timestamp: string;
  cpuUsage: number;
  memoryUsage: number;
  networkIn: number;
  networkOut: number;
}

export interface ActivityStats {
  avgCpuUsage: number;
  avgMemoryUsage: number;
  peakCpuUsage: number;
  peakMemoryUsage: number;
  totalNetworkIn: string;
  totalNetworkOut: string;
}

export interface ActivityData {
  history: ActivityDataPoint[];
  stats: ActivityStats;
}

export interface MiningStats {
  hashrate: string;
  hashrateUnit: string;
  shares: {
    accepted: number;
    rejected: number;
    total: number;
  };
  uptime: number;
  algorithm: string;
  pool?: string;
}

export interface MiningStatus {
  isRunning: boolean;
  startedAt?: string;
  stats?: MiningStats;
}

export interface EconomyBalance {
  available: number;
  pending: number;
  total: number;
  currency: string;
  lastUpdated: string;
}

const API_BASE = '';

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function fetchSerialPorts(): Promise<{ ports: SerialPortInfo[] }> {
  return fetchJSON<{ ports: SerialPortInfo[] }>('/api/hardware/ports');
}

export async function connectSerialPort(path: string, baudRate: number): Promise<ConnectedPort> {
  return fetchJSON<ConnectedPort>('/api/hardware/connect', {
    method: 'POST',
    body: JSON.stringify({ path, baudRate }),
  });
}

export async function disconnectSerialPort(path: string): Promise<void> {
  await fetchJSON<void>('/api/hardware/disconnect', {
    method: 'POST',
    body: JSON.stringify({ path }),
  });
}

export async function fetchSystemInfo(): Promise<SystemInfo> {
  return fetchJSON<SystemInfo>('/api/hardware/system');
}

export async function fetchActivityData(): Promise<ActivityData> {
  return fetchJSON<ActivityData>('/api/hardware/activity');
}

export async function fetchMiningStatus(): Promise<MiningStatus> {
  return fetchJSON<MiningStatus>('/api/mining/status');
}

export async function startMining(): Promise<MiningStatus> {
  return fetchJSON<MiningStatus>('/api/mining/start', { method: 'POST' });
}

export async function stopMining(): Promise<MiningStatus> {
  return fetchJSON<MiningStatus>('/api/mining/stop', { method: 'POST' });
}

export async function fetchEconomyBalance(): Promise<EconomyBalance> {
  return fetchJSON<EconomyBalance>('/api/economy/balance');
}

export interface UseHardwareReturn {
  ports: SerialPortInfo[];
  connectedPorts: ConnectedPort[];
  portsLoading: boolean;
  portsError: string | null;
  refreshPorts: () => Promise<void>;
  connect: (path: string, baudRate: number) => Promise<void>;
  disconnect: (path: string) => Promise<void>;

  systemInfo: SystemInfo | null;
  systemLoading: boolean;
  systemError: string | null;
  refreshSystem: () => Promise<void>;

  activity: ActivityData | null;
  activityLoading: boolean;
  activityError: string | null;
  refreshActivity: () => Promise<void>;

  miningStatus: MiningStatus | null;
  miningLoading: boolean;
  miningError: string | null;
  startMiningOperation: () => Promise<void>;
  stopMiningOperation: () => Promise<void>;

  balance: EconomyBalance | null;
  balanceLoading: boolean;
  balanceError: string | null;
  refreshBalance: () => Promise<void>;

  refreshAll: () => Promise<void>;
}

export function useHardware(): UseHardwareReturn {
  const [ports, setPorts] = useState<SerialPortInfo[]>([]);
  const [connectedPorts, setConnectedPorts] = useState<ConnectedPort[]>([]);
  const [portsLoading, setPortsLoading] = useState(true);
  const [portsError, setPortsError] = useState<string | null>(null);

  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [systemLoading, setSystemLoading] = useState(true);
  const [systemError, setSystemError] = useState<string | null>(null);

  const [activity, setActivity] = useState<ActivityData | null>(null);
  const [activityLoading, setActivityLoading] = useState(true);
  const [activityError, setActivityError] = useState<string | null>(null);

  const [miningStatus, setMiningStatus] = useState<MiningStatus | null>(null);
  const [miningLoading, setMiningLoading] = useState(true);
  const [miningError, setMiningError] = useState<string | null>(null);

  const [balance, setBalance] = useState<EconomyBalance | null>(null);
  const [balanceLoading, setBalanceLoading] = useState(true);
  const [balanceError, setBalanceError] = useState<string | null>(null);

  const loadPorts = useCallback(async () => {
    setPortsLoading(true);
    setPortsError(null);
    try {
      const data = await fetchSerialPorts();
      setPorts(data.ports);
    } catch (error) {
      setPortsError(error instanceof Error ? error.message : 'Failed to load serial ports');
    } finally {
      setPortsLoading(false);
    }
  }, []);

  const loadSystemInfo = useCallback(async () => {
    setSystemLoading(true);
    setSystemError(null);
    try {
      const data = await fetchSystemInfo();
      setSystemInfo(data);
    } catch (error) {
      setSystemError(error instanceof Error ? error.message : 'Failed to load system info');
    } finally {
      setSystemLoading(false);
    }
  }, []);

  const loadActivity = useCallback(async () => {
    setActivityLoading(true);
    setActivityError(null);
    try {
      const data = await fetchActivityData();
      setActivity(data);
    } catch (error) {
      setActivityError(error instanceof Error ? error.message : 'Failed to load activity data');
    } finally {
      setActivityLoading(false);
    }
  }, []);

  const loadMiningStatus = useCallback(async () => {
    setMiningLoading(true);
    setMiningError(null);
    try {
      const data = await fetchMiningStatus();
      setMiningStatus(data);
    } catch (error) {
      setMiningError(error instanceof Error ? error.message : 'Failed to load mining status');
    } finally {
      setMiningLoading(false);
    }
  }, []);

  const loadBalance = useCallback(async () => {
    setBalanceLoading(true);
    setBalanceError(null);
    try {
      const data = await fetchEconomyBalance();
      setBalance(data);
    } catch (error) {
      setBalanceError(error instanceof Error ? error.message : 'Failed to load balance');
    } finally {
      setBalanceLoading(false);
    }
  }, []);

  useEffect(() => {
    loadPorts();
    loadSystemInfo();
    loadActivity();
    loadMiningStatus();
    loadBalance();
  }, [loadPorts, loadSystemInfo, loadActivity, loadMiningStatus, loadBalance]);

  useEffect(() => {
    const interval = setInterval(() => {
      loadActivity();
    }, 5000);
    return () => clearInterval(interval);
  }, [loadActivity]);

  const connect = useCallback(async (path: string, baudRate: number) => {
    try {
      const connected = await connectSerialPort(path, baudRate);
      setConnectedPorts(prev => [...prev, connected]);
    } catch (error) {
      setPortsError(error instanceof Error ? error.message : 'Failed to connect');
      throw error;
    }
  }, []);

  const disconnect = useCallback(async (path: string) => {
    try {
      await disconnectSerialPort(path);
      setConnectedPorts(prev => prev.filter(p => p.path !== path));
    } catch (error) {
      setPortsError(error instanceof Error ? error.message : 'Failed to disconnect');
      throw error;
    }
  }, []);

  const startMiningOperation = useCallback(async () => {
    setMiningLoading(true);
    try {
      const status = await startMining();
      setMiningStatus(status);
    } catch (error) {
      setMiningError(error instanceof Error ? error.message : 'Failed to start mining');
      throw error;
    } finally {
      setMiningLoading(false);
    }
  }, []);

  const stopMiningOperation = useCallback(async () => {
    setMiningLoading(true);
    try {
      const status = await stopMining();
      setMiningStatus(status);
    } catch (error) {
      setMiningError(error instanceof Error ? error.message : 'Failed to stop mining');
      throw error;
    } finally {
      setMiningLoading(false);
    }
  }, []);

  const refreshAll = useCallback(async () => {
    await Promise.all([
      loadPorts(),
      loadSystemInfo(),
      loadActivity(),
      loadMiningStatus(),
      loadBalance(),
    ]);
  }, [loadPorts, loadSystemInfo, loadActivity, loadMiningStatus, loadBalance]);

  return {
    ports,
    connectedPorts,
    portsLoading,
    portsError,
    refreshPorts: loadPorts,
    connect,
    disconnect,

    systemInfo,
    systemLoading,
    systemError,
    refreshSystem: loadSystemInfo,

    activity,
    activityLoading,
    activityError,
    refreshActivity: loadActivity,

    miningStatus,
    miningLoading,
    miningError,
    startMiningOperation,
    stopMiningOperation,

    balance,
    balanceLoading,
    balanceError,
    refreshBalance: loadBalance,

    refreshAll,
  };
}
