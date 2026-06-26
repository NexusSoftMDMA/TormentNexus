import { input, password, select } from '@inquirer/prompts';
import chalk from 'chalk';
import fsp from 'fs/promises';
import yaml from 'yaml';
import { configExists, expandHome, getConfigPath, readConfig, resolveConfigPath, saveConfig } from '../config/loader.js';
import { MvmtConfig, TunnelConfig } from '../config/schema.js';
import {
  cloudflareNamedTunnelCommand,
  defaultTunnelCommand,
  formatDashboardPublicUrl,
  formatMcpPublicUrl,
  missingTunnelDependency,
  normalizeTunnelBaseUrl,
} from '../utils/tunnel.js';
import {
  ControlAuthError,
  ControlUnavailableError,
  getControlSocketPath,
  sendJsonControlRequest,
  streamJsonControl,
} from '../utils/control.js';
import { readSessionToken } from '../utils/token.js';
import { tunnelLegacyAccessWarning } from './tunnel-safety.js';
import { promptForExistingFile } from './folder-prompt.js';

export interface TunnelCommandOptions {
  config?: string;
  quick?: boolean;
  cloudflareConfig?: string;
  relay?: boolean;
  relayUrl?: string;
  relayWorkspace?: string;
  relayToken?: string;
  publicUrl?: string;
}

export interface TunnelRuntimeStatus {
  configured: boolean;
  running: boolean;
  command?: string;
  publicUrl?: string;
  recentLogs?: string[];
}

export interface ApplyTunnelConfigResult {
  config: MvmtConfig;
  enabled: boolean;
  warning?: string;
}

const DEFAULT_RELAY_URL = 'wss://relay.mvmt.net/connect';

export async function showTunnel(options: TunnelCommandOptions = {}): Promise<void> {
  const loaded = loadConfigSummary(options.config);
  if (!loaded) return;

  const runtime = await readRuntimeStatus(loaded.configPath);
  printTunnelSummary(loaded.config, runtime);
}

function requireControlToken(): string {
  const token = readSessionToken();
  if (!token) {
    console.error('No session token found. Run `mvmt serve` once before using tunnel commands.');
    process.exitCode = 1;
    throw new Error('missing session token');
  }
  return token;
}

export async function configureTunnel(options: TunnelCommandOptions = {}): Promise<void> {
  const loaded = loadConfigSummary(options.config);
  if (!loaded) return;

  const usesRelay = Boolean(options.relay || options.relayUrl || options.relayWorkspace || options.relayToken);
  const configuredModes = [options.quick, options.cloudflareConfig, usesRelay].filter(Boolean).length;
  if (configuredModes > 1) {
    throw new Error('Use only one tunnel config mode.');
  }
  const tunnel = options.quick
    ? { provider: 'cloudflare-quick' as const, command: defaultTunnelCommand('cloudflare-quick') }
    : options.cloudflareConfig
      ? await cloudflareConfigTunnel(options.cloudflareConfig)
      : usesRelay
        ? relayTunnelFromOptions(options)
      : await promptForTunnelConfig(loaded.config.server.port);
  const missingDependency = missingTunnelDependency(tunnel);
  if (missingDependency) {
    printMissingTunnelDependencyWarning(missingDependency, (line) => console.log(chalk.yellow(line)));
    process.exitCode = 1;
    return;
  }
  const applied = applyTunnelConfig(loaded.config, tunnel);
  await saveConfig(loaded.configPath, applied.config);

  console.log(chalk.green(`Tunnel config saved to ${loaded.configPath}`));
  if (applied.warning) {
    printTunnelEnabledWithNoTokens(applied.warning);
  }
  if (tunnel.url) {
    printPublicTunnelUrls(tunnel.url);
  }

  try {
    const token = readSessionToken();
    if (!token) {
      console.log(chalk.dim('Config saved. Start mvmt with `mvmt serve` to launch the tunnel.'));
      return;
    }
    const runtime = await sendJsonControlRequest<TunnelRuntimeStatus>(
      getControlSocketPath(loaded.configPath),
      { type: 'tunnel.config', tunnel },
      token,
    );
    if (runtime.publicUrl) {
      printPublicTunnelUrls(runtime.publicUrl);
    } else {
      console.log(chalk.dim('Tunnel config applied. mvmt is still running locally.'));
    }
  } catch (err) {
    if (err instanceof ControlUnavailableError) {
      console.log(chalk.dim('Config saved. Start mvmt with `mvmt serve` to launch the tunnel.'));
      return;
    }
    if (err instanceof ControlAuthError) {
      console.error('Control socket rejected the session token. The running mvmt instance may be using a different token.');
      process.exitCode = 1;
      return;
    }
    throw err;
  }
}

export async function startTunnelCommand(options: TunnelCommandOptions = {}): Promise<void> {
  const runtime = await sendTunnelRequest(options.config, 'tunnel.start');
  printTunnelActionResult(runtime);
  if (runtime.configured) printApiTokenWarningForConfig(options.config);
}

export async function refreshTunnelCommand(options: TunnelCommandOptions = {}): Promise<void> {
  const runtime = await sendTunnelRequest(options.config, 'tunnel.refresh');
  printTunnelActionResult(runtime);
  if (runtime.configured) printApiTokenWarningForConfig(options.config);
}

export async function stopTunnelCommand(options: TunnelCommandOptions = {}): Promise<void> {
  const runtime = await sendTunnelRequest(options.config, 'tunnel.stop');
  if (!runtime.running) {
    console.log(chalk.dim('Tunnel stopped. mvmt is still serving locally.'));
    return;
  }
  printTunnelActionResult(runtime);
}

export async function disableTunnelAccess(options: TunnelCommandOptions = {}): Promise<void> {
  const loaded = loadConfigSummary(options.config);
  if (!loaded) return;

  if (loaded.config.server.access !== 'tunnel') {
    console.log(chalk.dim('Tunnel access is already disabled.'));
    return;
  }

  loaded.config.server.access = 'local';
  await saveConfig(loaded.configPath, loaded.config);
  console.log(chalk.green(`Tunnel access disabled in ${loaded.configPath}`));
  console.log(chalk.dim('Saved tunnel details were kept. Restart mvmt to serve locally.'));
}

export async function showTunnelLogs(options: TunnelCommandOptions = {}): Promise<void> {
  const runtime = await sendTunnelRequest(options.config, 'tunnel.logs');
  const logs = runtime.recentLogs ?? [];
  console.log('mvmt tunnel logs\n');
  if (logs.length === 0) {
    console.log(chalk.dim('No tunnel output captured.'));
    return;
  }

  for (const line of logs.slice(-20)) {
    console.log(line);
  }
}

export async function streamTunnelLogs(options: TunnelCommandOptions = {}): Promise<void> {
  const configPath = resolveConfigPath(options.config);
  const socketPath = getControlSocketPath(configPath);
  const token = requireControlToken();
  let finish: (() => void) | undefined;
  let handleSigint: (() => void) | undefined;
  const stop = await streamJsonControl(socketPath, { type: 'tunnel.logs.stream' }, (message) => {
    if (message?.kind === 'ready') {
      console.log(chalk.dim('Streaming tunnel logs. Press Ctrl+C to stop.'));
      for (const line of message.logs ?? []) {
        console.log(line);
      }
      return;
    }
    if (message?.kind === 'log' && typeof message.line === 'string') {
      console.log(message.line);
      return;
    }
    if (message?.kind === 'error') {
      console.error(message.error ?? 'Tunnel log stream failed.');
      finish?.();
      return;
    }
    if (message?.kind === 'end') {
      finish?.();
    }
  }, token);

  await new Promise<void>((resolve) => {
    finish = () => {
      if (handleSigint) {
        process.off('SIGINT', handleSigint);
      }
      resolve();
    };
    const sigintHandler = () => {
      process.off('SIGINT', sigintHandler);
      stop();
      resolve();
    };
    handleSigint = sigintHandler;
    process.on('SIGINT', handleSigint);
  });
}

export function printTunnelSummary(config: MvmtConfig, runtime?: TunnelRuntimeStatus): void {
  console.log('mvmt tunnel\n');
  console.log('Configured');
  console.log(`  access: ${config.server.access}`);
  if (!config.server.tunnel) {
    console.log(`  ${chalk.dim('Tunnel is not configured.')}`);
    return;
  }

  console.log(`  provider: ${config.server.tunnel.provider}`);
  if (config.server.tunnel.provider === 'relay') {
    console.log(`  relay: ${config.server.tunnel.relayUrl}`);
    console.log(`  workspace: ${config.server.tunnel.workspaceSlug}`);
  } else {
    console.log(`  command: ${config.server.tunnel.command}`);
  }
  if (config.server.tunnel.url) {
    console.log(`  dashboard: ${formatDashboardPublicUrl(config.server.tunnel.url)}`);
    console.log(`  MCP: ${formatMcpPublicUrl(config.server.tunnel.url)}`);
  }
  if (config.server.access !== 'tunnel') {
    console.log(`  ${chalk.dim('Tunnel details are saved, but tunnel access is disabled.')}`);
    return;
  }

  console.log('\nLive');
  if (!runtime) {
    console.log(`  ${chalk.dim('mvmt is not running.')}`);
    return;
  }
  console.log(`  status: ${runtime.running ? 'running' : 'stopped'}`);
  if (runtime.command) console.log(`  command: ${runtime.command}`);
  if (runtime.publicUrl) {
    console.log(`  dashboard: ${formatDashboardPublicUrl(runtime.publicUrl)}`);
    console.log(`  MCP: ${formatMcpPublicUrl(runtime.publicUrl)}`);
  }
}

export function applyTunnelConfig(
  config: MvmtConfig,
  tunnel: TunnelConfig,
  env: NodeJS.ProcessEnv = process.env,
): ApplyTunnelConfigResult {
  const enabledConfig: MvmtConfig = {
    ...config,
    server: {
      ...config.server,
      access: 'tunnel',
      tunnel,
    },
  };
  return {
    config: enabledConfig,
    enabled: true,
    warning: tunnelLegacyAccessWarning(enabledConfig, env),
  };
}

export function printTunnelEnabledWithNoTokens(warning: string): void {
  console.log(chalk.yellow(warning));
  console.log(chalk.dim('The public URL can be reachable before any API token can read data.'));
  console.log(chalk.dim('Next step: mvmt token add'));
}

export async function promptForTunnelConfig(port: number): Promise<TunnelConfig> {
  while (true) {
    const provider = await select<'relay' | 'custom-relay' | 'cloudflare-quick' | 'cloudflare-named' | 'localhost-run' | 'custom'>({
      message: 'Which tunnel?',
      choices: [
        { name: 'MVMT Relay (stable URL, no cloudflared)', value: 'relay' },
        { name: 'Custom MVMT Relay (self-hosted relay URL)', value: 'custom-relay' },
        { name: 'Cloudflare Quick Tunnel (temporary URL, requires cloudflared)', value: 'cloudflare-quick' },
        { name: 'Cloudflare Named Tunnel (stable domain, requires existing cloudflared config)', value: 'cloudflare-named' },
        { name: 'localhost.run (fallback, less stable)', value: 'localhost-run' },
        { name: 'Custom tunnel command', value: 'custom' },
      ],
    });

    const tunnel = await promptForTunnelDetails(provider);
    const missingDependency = missingTunnelDependency(tunnel);
    if (missingDependency) {
      printMissingTunnelDependencyWarning(missingDependency, (line) => console.log(chalk.yellow(line)));
      console.log(chalk.dim('Choose another tunnel provider, or press Ctrl+C and install the missing command first.'));
      continue;
    }
    if (tunnel.provider !== 'relay' && !tunnel.command.includes('{port}') && provider !== 'cloudflare-named') {
      console.log(chalk.yellow(`The configured tunnel command does not include ${chalk.cyan('{port}')}.`));
      console.log(chalk.dim(`mvmt expects tunnel commands to reference the local port ${port}.`));
    }
    return tunnel;
  }
}

export function printMissingTunnelDependencyWarning(command: string, write: (line: string) => void): void {
  if (command === 'cloudflared') {
    write('Cloudflare Quick Tunnel requires `cloudflared`, but it is not installed or not on PATH.');
    write('Install it with `brew install cloudflared`, or run `mvmt tunnel config` and choose another provider.');
    return;
  }
  write(`Tunnel dependency is missing: ${command}`);
}

async function cloudflareConfigTunnel(configPathInput: string): Promise<TunnelConfig> {
  const configPath = expandHome(configPathInput.trim());
  if (!configPath) throw new Error('Cloudflare config path is required.');
  const raw = await fsp.readFile(configPath, 'utf-8');
  const publicUrl = inferCloudflareConfigPublicUrl(raw);
  return {
    provider: 'custom',
    command: cloudflareNamedTunnelCommand(configPath),
    ...(publicUrl ? { url: publicUrl } : {}),
  };
}

function relayTunnelFromOptions(options: TunnelCommandOptions): TunnelConfig {
  const relayUrl = normalizeRelayUrlInput(options.relayUrl ?? DEFAULT_RELAY_URL);
  const workspaceSlug = options.relayWorkspace?.trim();
  const agentToken = options.relayToken?.trim();
  if (!workspaceSlug || !agentToken) {
    throw new Error('Relay tunnel config requires --relay-workspace and --relay-token.');
  }
  return {
    provider: 'relay',
    relayUrl,
    workspaceSlug,
    agentToken,
    ...(options.publicUrl ? { url: normalizeTunnelBaseUrl(options.publicUrl) } : {}),
  };
}

export function inferCloudflareConfigPublicUrl(rawConfig: string): string | undefined {
  const parsed = yaml.parse(rawConfig) as unknown;
  const hostname = firstCloudflareHostname(parsed);
  return hostname ? normalizeTunnelBaseUrl(`https://${hostname}`) : undefined;
}

function firstCloudflareHostname(value: unknown): string | undefined {
  if (!value || typeof value !== 'object') return undefined;
  const record = value as Record<string, unknown>;
  if (typeof record.hostname === 'string' && record.hostname.trim()) return record.hostname.trim();
  if (Array.isArray(record.ingress)) {
    for (const entry of record.ingress) {
      if (!entry || typeof entry !== 'object') continue;
      const hostname = (entry as Record<string, unknown>).hostname;
      if (typeof hostname === 'string' && hostname.trim()) return hostname.trim();
    }
  }
  return undefined;
}

async function promptForTunnelDetails(
  provider: 'relay' | 'custom-relay' | 'cloudflare-quick' | 'cloudflare-named' | 'localhost-run' | 'custom',
): Promise<TunnelConfig> {
  if (provider === 'relay' || provider === 'custom-relay') {
    const relayUrl = await input({
      message: 'Relay WebSocket URL',
      default: provider === 'relay' ? DEFAULT_RELAY_URL : '',
      validate: validateRelayUrlInput,
    });
    const workspaceSlug = await input({
      message: 'Workspace slug',
      validate: validateWorkspaceSlugInput,
    });
    const agentToken = await password({
      message: 'Relay agent token/secret',
      validate: (value) => (value.trim().length > 0 ? true : 'Enter the relay agent token/secret'),
    });
    const publicUrl = await input({
      message: 'Public base URL (optional, for printed lease/dashboard links)',
      default: '',
      validate: (value) => (value.trim().length === 0 ? true : validatePublicUrlInput(value)),
    });
    return {
      provider: 'relay',
      relayUrl: normalizeRelayUrlInput(relayUrl),
      workspaceSlug: workspaceSlug.trim(),
      agentToken: agentToken.trim(),
      ...(publicUrl.trim().length > 0 ? { url: normalizeTunnelBaseUrl(publicUrl) } : {}),
    };
  }

  if (provider === 'cloudflare-named') {
    console.log(chalk.dim('Use this after creating a Cloudflare named tunnel and DNS route.'));
    const configPath = await promptForExistingFile('Cloudflared config file:', {
      defaultValue: '~/.cloudflared/config.yml',
    });
    const publicUrl = await input({
      message: 'Public base URL',
      default: 'https://example.com',
      validate: validatePublicUrlInput,
    });
    return {
      provider: 'custom',
      command: cloudflareNamedTunnelCommand(expandHome(configPath.trim())),
      url: normalizeTunnelBaseUrl(publicUrl),
    };
  }

  if (provider === 'custom') {
    const command = await input({
      message: 'Tunnel command (use {port} where mvmt should insert the local port)',
      validate: (value) => (value.trim().length > 0 ? true : 'Enter a tunnel command'),
    });
    const publicUrl = await input({
      message: 'Public base URL (optional, recommended if the command does not print one)',
      default: '',
      validate: (value) => (value.trim().length === 0 ? true : validatePublicUrlInput(value)),
    });
    return {
      provider: 'custom',
      command: command.trim(),
      ...(publicUrl.trim().length > 0 ? { url: normalizeTunnelBaseUrl(publicUrl) } : {}),
    };
  }

  return {
    provider,
    command: defaultTunnelCommand(provider),
  };
}

function loadConfigSummary(configOverride?: string): { configPath: string; config: MvmtConfig } | undefined {
  const configPath = resolveConfigPath(configOverride);
  if (!configExists(configPath)) {
    console.error(`Config not found at ${configPath}`);
    console.error('Run `mvmt config setup` to create one.');
    process.exitCode = 1;
    return undefined;
  }

  try {
    return { configPath, config: readConfig(configPath) };
  } catch (err) {
    console.error(err instanceof Error ? err.message : 'Invalid config.');
    process.exitCode = 1;
    return undefined;
  }
}

async function readRuntimeStatus(configPath: string): Promise<TunnelRuntimeStatus | undefined> {
  const token = readSessionToken();
  if (!token) return undefined;
  try {
    return await sendJsonControlRequest<TunnelRuntimeStatus>(
      getControlSocketPath(configPath),
      { type: 'tunnel.status' },
      token,
    );
  } catch (err) {
    if (err instanceof ControlUnavailableError) return undefined;
    if (err instanceof ControlAuthError) return undefined;
    throw err;
  }
}

async function sendTunnelRequest(configOverride: string | undefined, type: string): Promise<TunnelRuntimeStatus> {
  const configPath = resolveConfigPath(configOverride);
  try {
    const token = requireControlToken();
    return await sendJsonControlRequest<TunnelRuntimeStatus>(
      getControlSocketPath(configPath),
      { type },
      token,
    );
  } catch (err) {
    if (err instanceof ControlUnavailableError) {
      console.error('mvmt is not running. Start it first with `mvmt serve`.');
      process.exitCode = 1;
      return { configured: false, running: false };
    }
    if (err instanceof ControlAuthError) {
      console.error('Control socket rejected the session token. The running mvmt instance may be using a different token.');
      process.exitCode = 1;
      return { configured: false, running: false };
    }
    throw err;
  }
}

function printTunnelActionResult(runtime: TunnelRuntimeStatus): void {
  if (runtime.publicUrl) {
    printPublicTunnelUrls(runtime.publicUrl);
    return;
  }

  if (!runtime.configured) {
    console.log(chalk.yellow('No tunnel is configured. Run `mvmt tunnel config` first.'));
    return;
  }

  if (!runtime.running) {
    console.log(chalk.dim('Tunnel is configured but not running.'));
    return;
  }

  console.log(chalk.dim('Tunnel is running, but no public URL has been detected yet.'));
}

function printPublicTunnelUrls(publicUrl: string): void {
  console.log(`dashboard   ${chalk.yellow(formatDashboardPublicUrl(publicUrl))}`);
  console.log(`MCP         ${chalk.dim(formatMcpPublicUrl(publicUrl))}`);
}

function printApiTokenWarningForConfig(configOverride: string | undefined): void {
  const loaded = loadConfigSummary(configOverride);
  if (!loaded) return;
  const warning = tunnelLegacyAccessWarning(loaded.config);
  if (warning) printTunnelEnabledWithNoTokens(warning);
}

function validatePublicUrlInput(value: string): true | string {
  try {
    const url = new URL(normalizeTunnelBaseUrl(value));
    return url.protocol === 'https:' ? true : 'Enter an https:// URL';
  } catch {
    return 'Enter a valid public URL, for example https://pnee.gofrieda.org';
  }
}

function validateRelayUrlInput(value: string): true | string {
  try {
    normalizeRelayUrlInput(value);
    return true;
  } catch {
    return 'Enter a valid ws:// or wss:// relay URL';
  }
}

function validateWorkspaceSlugInput(value: string): true | string {
  return /^[a-z0-9][a-z0-9-]{0,62}$/.test(value.trim())
    ? true
    : 'Use lowercase letters, numbers, and dashes';
}

function normalizeRelayUrlInput(value: string): string {
  const url = new URL(value.trim());
  if (url.protocol !== 'ws:' && url.protocol !== 'wss:') throw new Error('Relay URL must use ws:// or wss://');
  url.hash = '';
  return url.toString();
}
