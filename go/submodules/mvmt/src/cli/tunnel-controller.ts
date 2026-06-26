import { MvmtConfig, TunnelConfig } from '../config/schema.js';
import { Logger } from '../utils/logger.js';
import { startRelayClient } from '../utils/relay-client.js';
import { formatDashboardPublicUrl, missingTunnelDependency, RunningTunnel, startTunnel } from '../utils/tunnel.js';
import { withTerminalProgress } from './progress.js';
import { printMissingTunnelDependencyWarning } from './tunnel.js';

export interface TunnelSnapshot {
  configured: boolean;
  running: boolean;
  command?: string;
  publicUrl?: string;
  recentLogs: string[];
  lastError?: string;
}

export class TunnelController {
  private current: RunningTunnel | undefined;
  private readonly logs: string[] = [];
  private readonly listeners = new Set<(line: string) => void>();
  private lastError: string | undefined;

  constructor(
    // This intentionally aliases the caller-owned server config so tunnel
    // changes stay in sync with the config object that gets persisted.
    private readonly serverConfig: MvmtConfig['server'],
    private readonly port: number,
    private readonly logger: Logger,
  ) {}

  get configured(): boolean {
    return Boolean(this.serverConfig.tunnel);
  }

  get running(): boolean {
    return Boolean(this.current);
  }

  get publicUrl(): string | undefined {
    return this.current?.url;
  }

  get command(): string | undefined {
    const tunnel = this.serverConfig.tunnel;
    if (!tunnel) return undefined;
    if (tunnel.provider === 'relay') return `relay ${tunnel.relayUrl} (${tunnel.workspaceSlug})`;
    return tunnel.command.replaceAll('{port}', String(this.port));
  }

  configure(tunnel: TunnelConfig): void {
    this.serverConfig.access = 'tunnel';
    this.serverConfig.tunnel = tunnel;
    this.logs.length = 0;
    this.lastError = undefined;
  }

  recentLogs(): string[] {
    return [...this.logs];
  }

  snapshot(): TunnelSnapshot {
    return {
      configured: this.configured,
      running: this.running,
      command: this.command,
      publicUrl: this.publicUrl,
      recentLogs: this.recentLogs(),
      lastError: this.lastError,
    };
  }

  subscribeLogs(listener: (line: string) => void): () => void {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  async start(options: { enable?: boolean } = {}): Promise<RunningTunnel | undefined> {
    if (this.current) return this.current;

    if (this.serverConfig.access !== 'tunnel') {
      if (!options.enable || !this.serverConfig.tunnel) return undefined;
      this.serverConfig.access = 'tunnel';
    }

    if (!this.serverConfig.tunnel) {
      this.logger.warn('Tunnel access is enabled, but no tunnel command is configured.');
      return undefined;
    }

    const missingDependency = missingTunnelDependency(this.serverConfig.tunnel);
    if (missingDependency) {
      this.addLog(`${missingDependency}: command not found`);
      this.lastError = `${missingDependency}: command not found`;
      printMissingTunnelDependencyWarning(missingDependency, (line) => this.logger.warn(line));
      return undefined;
    }

    const tunnelConfig = this.serverConfig.tunnel;
    if (tunnelConfig.provider === 'relay') {
      this.logger.info(`Starting relay: ${tunnelConfig.workspaceSlug} -> ${tunnelConfig.relayUrl}`);
      try {
        const relay = await startRelayClient({
          relayUrl: tunnelConfig.relayUrl,
          workspaceSlug: tunnelConfig.workspaceSlug,
          agentToken: tunnelConfig.agentToken,
          localPort: this.port,
          logger: this.logger,
        });
        const running: RunningTunnel = {
          command: this.command ?? 'relay',
          url: tunnelConfig.url,
          stop: () => relay.close(),
        };
        this.current = running;
        this.lastError = undefined;
        if (running.url) this.logger.info(`Dashboard URL: ${formatDashboardPublicUrl(running.url)}`);
        return running;
      } catch (err) {
        this.lastError = err instanceof Error ? err.message : 'Unknown error';
        this.logger.warn(`Relay failed to start: ${this.lastError}`);
        this.logger.warn('mvmt is still running locally.');
        return undefined;
      }
    }

    this.logger.info(`Starting tunnel: ${this.command}`);
    try {
      const tunnel = await withTerminalProgress('Waiting for tunnel URL', () =>
        startTunnel(tunnelConfig.command, this.port, {
          onOutput: (line) => this.addLog(line),
        }),
      );
      tunnel.url = tunnel.url || this.serverConfig.tunnel.url;
      this.current = tunnel;
      this.lastError = undefined;
      if (!tunnel.url) {
        this.logger.warn('Tunnel process started, but mvmt could not detect a public URL from its output yet.');
        return tunnel;
      }
      this.logger.info(`Dashboard URL: ${formatDashboardPublicUrl(tunnel.url)}`);
      return tunnel;
    } catch (err) {
      this.lastError = err instanceof Error ? err.message : 'Unknown error';
      this.logger.warn(`Tunnel failed to start: ${this.lastError}`);
      this.logger.warn('mvmt is still running locally.');
      return undefined;
    }
  }

  async refresh(options: { enable?: boolean } = {}): Promise<RunningTunnel | undefined> {
    await this.stop();
    return this.start(options);
  }

  async stop(): Promise<void> {
    if (!this.current) return;
    const tunnel = this.current;
    this.current = undefined;
    await tunnel.stop();
  }

  private addLog(line: string): void {
    this.logs.push(line);
    if (this.logs.length > 100) this.logs.shift();
    for (const listener of this.listeners) {
      listener(line);
    }
  }
}
