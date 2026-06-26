import chalk from 'chalk';
import fsPromises from 'fs/promises';
import os from 'os';
import path from 'path';
import { MvmtConfig } from '../config/schema.js';
import { configExists, expandHome, readConfig, resolveConfigPath, saveConfig } from '../config/loader.js';
import { filesystemSetupDefinition } from '../connectors/filesystem-setup.js';
import { buildConfig, SetupConfigOptions, setupConfig } from './init.js';
import { formatDashboardPublicUrl, formatMcpPublicUrl } from '../utils/tunnel.js';

export interface ConfigCommandOptions {
  config?: string;
}

export interface TemporaryFilesystemConfigOptions {
  paths: string[];
  port?: number;
}

export interface TemporaryFilesystemConfig {
  config: MvmtConfig;
  configPath: string;
  cleanup(): Promise<void>;
}

export interface ConfigSummaryRuntime {
  tunnel?: {
    configured: boolean;
    running: boolean;
    command?: string;
    publicUrl?: string;
  };
}

export async function runConfigSetup(options: SetupConfigOptions = {}): Promise<void> {
  await setupConfig({
    ...options,
    promptOnOverwrite: options.promptOnOverwrite ?? true,
    printNextStep: options.printNextStep ?? true,
  });
}

export async function showConfig(options: ConfigCommandOptions = {}): Promise<void> {
  const configPath = resolveConfigPath(options.config);
  if (!configExists(configPath)) {
    console.error(`Config not found at ${configPath}`);
    console.error('Run `mvmt config setup` to create one.');
    process.exitCode = 1;
    return;
  }

  try {
    const config = readConfig(configPath);
    printConfigSummary(config, configPath);
  } catch (err) {
    console.error(err instanceof Error ? err.message : 'Invalid config.');
    process.exitCode = 1;
  }
}

export async function createTemporaryFilesystemConfig(
  options: TemporaryFilesystemConfigOptions,
): Promise<TemporaryFilesystemConfig> {
  const normalizedPaths = Array.from(new Set(options.paths.map((entry) => path.resolve(expandHome(entry)))));
  if (normalizedPaths.length === 0) {
    throw new Error('At least one folder is required for `mvmt serve --path`.');
  }

  const base = buildConfig({ port: options.port ?? 4141 });
  const config = filesystemSetupDefinition.apply(base, {
    paths: normalizedPaths,
    writeAccess: false,
  });
  const tempDir = await fsPromises.mkdtemp(path.join(os.tmpdir(), 'mvmt-serve-'));
  const configPath = path.join(tempDir, 'config.yaml');
  await saveConfig(configPath, config);

  return {
    config,
    configPath,
    async cleanup() {
      await fsPromises.rm(tempDir, { recursive: true, force: true });
    },
  };
}

export function printConfigSummary(
  config: MvmtConfig,
  configPath: string,
  runtime?: ConfigSummaryRuntime,
): void {
  console.log('mvmt config\n');
  console.log('Config');
  console.log(`  path: ${configPath}`);

  console.log('\nServer');
  console.log(`  port: ${config.server.port}`);
  console.log(`  dashboard: http://127.0.0.1:${config.server.port}/dashboard`);
  console.log(`  MCP: http://127.0.0.1:${config.server.port}/mcp`);
  console.log(`  access: ${config.server.access}`);

  if (config.server.access === 'tunnel') {
    const tunnelStatus = runtime?.tunnel;
    const tunnel = config.server.tunnel;
    console.log(`  provider: ${tunnel?.provider ?? 'custom'}`);
    if (tunnelStatus) {
      console.log(`  tunnel: ${tunnelStatus.running ? 'running' : 'stopped'}`);
    }
    const configuredCommand = tunnel && tunnel.provider !== 'relay' ? tunnel.command : undefined;
    if (tunnel?.provider === 'relay') {
      console.log(`  relay: ${tunnel.relayUrl}`);
      console.log(`  workspace: ${tunnel.workspaceSlug}`);
    }
    if (tunnelStatus?.command ?? configuredCommand) {
      console.log(`  command: ${tunnelStatus?.command ?? configuredCommand}`);
    }
    if (tunnelStatus?.publicUrl ?? tunnel?.url) {
      const publicUrl = tunnelStatus?.publicUrl ?? tunnel?.url ?? '';
      console.log(`  public dashboard: ${formatDashboardPublicUrl(publicUrl)}`);
      console.log(`  public MCP: ${formatMcpPublicUrl(publicUrl)}`);
    }
  }

  printMountSummary(config);
}

function printMountSummary(config: MvmtConfig): void {
  console.log('\nMounts');
  const mounts = config.mounts.filter((mount) => mount.enabled !== false);
  if (mounts.length === 0) {
    console.log(`  ${chalk.dim('not configured')}`);
    return;
  }

  for (const mount of mounts) {
    console.log(`  ${mount.name}: ${mount.path} -> ${mount.root}  ${mount.writeAccess ? 'writable' : 'read-only'}`);
    if (mount.description) console.log(`    description: ${mount.description}`);
    if (mount.guidance) console.log(`    guidance: ${mount.guidance}`);
  }
}
