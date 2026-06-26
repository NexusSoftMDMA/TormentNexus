import readline from 'node:readline';
import { confirm, input, select } from '@inquirer/prompts';
import chalk from 'chalk';
import { MvmtConfig } from '../config/schema.js';
import { AuditEntry, AuditLogger, AUDIT_LOG_PATH } from '../utils/audit.js';
import { HttpRequestLogEntry } from '../server/index.js';
import { formatDashboardPublicUrl, formatMcpPublicUrl } from '../utils/tunnel.js';
import { printApiTokenSaved, printApiTokens, promptAndAddApiToken, promptAndEditApiToken, promptAndRemoveApiToken, rotateApiTokenInConfig } from './api-tokens.js';
import { printConfigSummary } from './config.js';
import {
  findMount,
  MOUNT_LOAD_NOTICE,
  MOUNT_UNLOAD_NOTICE,
  printMountBasePermission,
  printMounts,
  promptAndAddMount,
  promptAndEditMount,
  promptAndRemoveMount,
} from './mounts.js';
import { applyTunnelConfig, printTunnelEnabledWithNoTokens, promptForTunnelConfig } from './tunnel.js';
import { LoadedConnector } from './connector-loader.js';
import { TunnelController } from './tunnel-controller.js';
import { addPathsToLease, createFolderLease, listFolderLeases, revokeFolderLease } from './lease.js';
import { DEFAULT_LEASE_TTL, defaultLeasesPath, leaseUnavailableReason, listLeases } from '../core/leases/store.js';
import { loadConfig } from '../config/loader.js';
import { addPrivilegedUserCommand, listPrivilegedUsersCommand, removePrivilegedUserCommand, setPrivilegedUserAdminCommand } from './users.js';

const SIGINT_EXIT_WINDOW_MS = 2_000;

export class InteractiveAuditLogger implements AuditLogger {
  private liveLogs = true;
  private writer: ((message: string) => void) | undefined;

  constructor(private readonly inner: AuditLogger) {}

  record(entry: AuditEntry): void {
    this.inner.record(entry);
    if (this.liveLogs && this.writer) {
      this.writer(formatAuditEntry(entry));
    }
  }

  recordHttp(entry: HttpRequestLogEntry): void {
    this.inner.recordHttp(entry);
    if (this.liveLogs && this.writer) {
      this.writer(formatHttpRequestEntry(entry));
    }
  }

  setLiveLogs(enabled: boolean): void {
    this.liveLogs = enabled;
  }

  getLiveLogs(): boolean {
    return this.liveLogs;
  }

  setWriter(writer: (message: string) => void): void {
    this.writer = writer;
  }
}

export interface InteractivePromptState {
  config: MvmtConfig;
  configPath: string;
  port: number;
  tunnel: TunnelController;
  loaded: LoadedConnector[];
  totalTools: number;
  audit: InteractiveAuditLogger;
  shutdown: () => Promise<void>;
  persistConfig: () => Promise<void>;
}

export function startInteractivePrompt(state: InteractivePromptState): void {
  let handlingCommand = false;
  let lastSigintAt = 0;
  let promptSuspended = false;
  let rl: readline.Interface | undefined;

  const writeAbovePrompt = (message: string) => {
    if (promptSuspended || !rl) {
      console.log(message);
      return;
    }
    readline.clearLine(process.stdout, 0);
    readline.cursorTo(process.stdout, 0);
    console.log(message);
    rl.prompt(true);
  };

  state.audit.setWriter(writeAbovePrompt);

  const suspendPrompt = () => {
    promptSuspended = true;
    rl?.close();
    rl = undefined;
  };

  const resumePrompt = () => {
    promptSuspended = false;
    rl = createInteractiveReadline();
    attachReadlineHandlers(rl);
    rl.prompt();
  };

  const attachReadlineHandlers = (activeRl: readline.Interface) => {
    activeRl.on('line', async (line) => {
      if (handlingCommand) return;
      handlingCommand = true;
      lastSigintAt = 0;
      const command = line.trim().toLowerCase();
      const shouldContinue = command !== 'quit' && command !== 'exit';
      if (!shouldContinue) {
        activeRl.close();
        return;
      }
      suspendPrompt();
      try {
        await handleInteractiveCommand(command, state, activeRl);
      } catch (err) {
        if (isPromptCancelError(err)) {
          console.log(chalk.yellow('Canceled.'));
        } else {
          console.log(chalk.red(err instanceof Error ? err.message : 'Command failed'));
        }
      } finally {
        handlingCommand = false;
      }
      if (shouldContinue) {
        resumePrompt();
      }
    });

    activeRl.on('SIGINT', async () => {
      if (handlingCommand) return;
      const now = Date.now();
      if (shouldShutdownOnSigint(lastSigintAt, now)) {
        await state.shutdown();
        return;
      }
      lastSigintAt = now;
      console.log(chalk.yellow('Press Ctrl-C again to stop mvmt, or type "quit".'));
      activeRl.prompt();
    });

    activeRl.on('close', async () => {
      if (promptSuspended) return;
      await state.shutdown();
    });
  };

  resumePrompt();
}

function createInteractiveReadline(): readline.Interface {
  return readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: chalk.cyan('> '),
  });
}

export function shouldShutdownOnSigint(lastSigintAt: number, now: number): boolean {
  return lastSigintAt > 0 && now - lastSigintAt <= SIGINT_EXIT_WINDOW_MS;
}

export function isPromptCancelError(err: unknown): boolean {
  return err instanceof Error
    && (
      err.name === 'ExitPromptError'
      || err.message.includes('User force closed the prompt')
      || err.message.includes('force closed')
    );
}

async function handleInteractiveCommand(
  command: string,
  state: InteractivePromptState,
  rl: readline.Interface,
): Promise<void> {
  if (command === 'quit' || command === 'exit') {
    rl.close();
    return;
  }

  if (command === 'clear') {
    console.clear();
    return;
  }

  if (command === 'help' || command === '?') {
    printInteractiveHelp();
    return;
  }

  switch (command) {
    case '':
      return;
    case 'lease':
    case 'lease list':
    case 'leases':
    case 'leases list':
      await listFolderLeases();
      return;
    case 'lease create':
    case 'lease add':
    case 'leases create':
    case 'leases add':
      await handleLeaseCreate(state);
      return;
    case 'lease add-path':
    case 'lease add-paths':
    case 'leases add-path':
    case 'leases add-paths':
      await handleLeaseAddPath(state);
      return;
    case 'lease revoke':
    case 'lease remove':
    case 'lease rm':
    case 'leases revoke':
    case 'leases remove':
    case 'leases rm':
      await handleLeaseRevoke();
      return;
    case 'users':
    case 'users list':
      await listPrivilegedUsersCommand();
      return;
    case 'users add':
      await handleUsersAdd();
      return;
    case 'users grant':
      await handleUsersAdmin(true);
      return;
    case 'users revoke':
      await handleUsersAdmin(false);
      return;
    case 'users remove':
    case 'users delete':
    case 'users rm':
      await handleUsersRemove();
      return;
    case 'advanced':
      printAdvancedHelp();
      return;
    case 'config':
    case 'advanced config':
      printConfigSummary(state.config, state.configPath, {
        tunnel: state.tunnel.snapshot(),
      });
      return;
    case 'config setup':
    case 'advanced config setup':
      console.log(chalk.dim('Run `mvmt config setup` in another shell, then restart `mvmt serve` to apply the new config.'));
      return;
    case 'token':
    case 'token list':
    case 'token show':
    case 'advanced token':
    case 'advanced token list':
    case 'advanced token show':
      printApiTokens(state.config);
      return;
    case 'token add':
    case 'advanced token add':
      await handleTokensAdd(state);
      return;
    case 'token edit':
    case 'advanced token edit':
      await handleTokensEdit(state);
      return;
    case 'token rotate':
    case 'advanced token rotate':
      await handleTokensRotate(state);
      return;
    case 'token remove':
    case 'advanced token remove':
      await handleTokensRemove(state);
      return;
    case 'tokens':
    case 'tokens list':
    case 'advanced tokens':
    case 'advanced tokens list':
      printApiTokens(state.config);
      return;
    case 'tokens add':
    case 'advanced tokens add':
      await handleTokensAdd(state);
      return;
    case 'tokens edit':
    case 'advanced tokens edit':
      await handleTokensEdit(state);
      return;
    case 'tokens rotate':
    case 'advanced tokens rotate':
      await handleTokensRotate(state);
      return;
    case 'tokens remove':
    case 'advanced tokens remove':
      await handleTokensRemove(state);
      return;
    case 'logs':
      printLiveLogState(state);
      return;
    case 'logs on':
      state.audit.setLiveLogs(true);
      console.log(chalk.green('live logs on'));
      return;
    case 'logs off':
      state.audit.setLiveLogs(false);
      console.log(chalk.dim('live logs off'));
      return;
    case 'status':
      printInteractiveStatus(state);
      return;
    case 'dashboard':
    case 'dashboard url':
    case 'dashboard urls':
    case 'url':
    case 'urls':
      printInteractiveUrls(state);
      return;
    case 'tunnel':
      printTunnelStatus(state);
      return;
    case 'tunnel config':
      await handleTunnelConfig(state);
      return;
    case 'tunnel start':
      await handleTunnelStart(state);
      return;
    case 'tunnel refresh':
      await handleTunnelRefresh(state);
      return;
    case 'tunnel stop':
      await handleTunnelStop(state);
      return;
    case 'tunnel logs':
      printTunnelLogs(state);
      return;
    case 'tunnel logs stream':
      console.log(chalk.dim('Use `mvmt tunnel logs stream` from another shell to follow tunnel output continuously.'));
      return;
    case 'connectors':
    case 'advanced connectors':
      printConnectorSummary(state.loaded, state.totalTools);
      return;
    case 'mounts':
    case 'advanced mounts':
      printMounts(state.config);
      return;
    case 'mounts add':
    case 'advanced mounts add':
      await handleMountsAdd(state);
      return;
    case 'mounts edit':
    case 'advanced mounts edit':
      await handleMountsEdit(state);
      return;
    case 'mounts remove':
    case 'advanced mounts remove':
      await handleMountsRemove(state);
      return;
    default:
      console.log(chalk.yellow(`Unknown command: ${command}`));
      console.log(chalk.dim('Type "help" for commands.'));
  }
}

export function printInteractiveHelp(): void {
  console.log('');
  console.log(chalk.bold('Commands'));
  console.log('  dashboard           show local/public dashboard URLs');
  console.log('  lease               list shared links');
  console.log('  lease create        create a shared link for a file or folder');
  console.log('  lease add-path      add files/folders to a shared link');
  console.log('  lease revoke        turn off a shared link');
  console.log('  users               list dashboard users');
  console.log('  users add           create dashboard user');
  console.log('  users grant/revoke  allow local source management');
  console.log('  users remove        delete dashboard user');
  console.log('  tunnel              show public access status');
  console.log('  tunnel config       configure relay or Cloudflare');
  console.log('  tunnel start        start public access');
  console.log('  tunnel refresh      restart public access and print the new URL');
  console.log('  tunnel stop         stop public access');
  console.log('  logs                show live activity log state');
  console.log('  logs on/off         toggle live activity logs');
  console.log('  status              show server, tunnel, and sharing status');
  console.log('  url                 show dashboard and MCP URLs');
  console.log('  advanced            show source, token, and MCP commands');
  console.log('  clear               clear the terminal');
  console.log('  quit                stop mvmt');
  console.log('');
}

export function printAdvancedHelp(): void {
  console.log('');
  console.log(chalk.bold('Advanced commands'));
  console.log('  advanced config              show saved mvmt config');
  console.log('  advanced config setup        rerun guided setup in another shell');
  console.log('  advanced mounts              list local sources');
  console.log('  advanced mounts add/edit/remove');
  console.log('  advanced token               list scoped MCP/API tokens');
  console.log('  advanced token add/edit      manage scoped MCP/API tokens');
  console.log('  advanced token rotate/remove');
  console.log('  advanced connectors          list compatibility MCP connectors');
  console.log('  tunnel logs                  show recent tunnel output');
  console.log('  tunnel logs stream           stream tunnel output from another shell');
  console.log('');
}

function printLiveLogState(state: InteractivePromptState): void {
  console.log(`live logs: ${state.audit.getLiveLogs() ? chalk.green('on') : chalk.dim('off')}`);
}

function printInteractiveStatus(state: InteractivePromptState): void {
  console.log('');
  console.log(chalk.bold('Status'));
  printInteractiveUrls(state);
  console.log(`shared links  ${chalk.cyan('lease')}`);
  console.log(`activity log  ${AUDIT_LOG_PATH}`);
  console.log(`live logs     ${state.audit.getLiveLogs() ? chalk.green('on') : chalk.dim('off')}`);
  printTunnelStatus(state);
  console.log(chalk.dim('Type `advanced` for source, token, connector, and MCP commands.'));
}

function printInteractiveUrls(state: InteractivePromptState): void {
  console.log(chalk.bold('dashboard'));
  console.log(`  local   ${chalk.cyan(localDashboardUrl(state.port))}`);
  if (state.tunnel.publicUrl) {
    console.log(`  public  ${chalk.yellow(formatDashboardPublicUrl(state.tunnel.publicUrl))}`);
  }
  console.log(chalk.bold('MCP endpoint'));
  console.log(`  local   ${chalk.dim(localMcpUrl(state.port))}`);
  if (state.tunnel.publicUrl) {
    console.log(`  public  ${chalk.dim(formatMcpPublicUrl(state.tunnel.publicUrl))}`);
  }
}

function printTunnelStatus(state: InteractivePromptState): void {
  console.log(chalk.bold('tunnel'));
  if (!state.tunnel.configured) {
    console.log(`  status      ${chalk.dim('not configured')}`);
    return;
  }
  console.log(`  status      ${state.tunnel.running ? chalk.green('running') : chalk.dim('stopped')}`);
  if (state.tunnel.command) console.log(`  command     ${chalk.dim(state.tunnel.command)}`);
  if (state.tunnel.publicUrl) {
    console.log(`  dashboard   ${chalk.yellow(formatDashboardPublicUrl(state.tunnel.publicUrl))}`);
    console.log(`  MCP         ${chalk.dim(formatMcpPublicUrl(state.tunnel.publicUrl))}`);
  }
}

async function handleTunnelStart(state: InteractivePromptState): Promise<void> {
  if (!state.tunnel.configured) {
    console.log(chalk.yellow('No tunnel is configured. Run `tunnel config` or `mvmt tunnel config`.'));
    return;
  }
  if (state.tunnel.running) {
    console.log(chalk.dim('Tunnel is already running.'));
    printTunnelStatus(state);
    return;
  }
  const wasLocalOnly = state.config.server.access !== 'tunnel';
  const tunnel = await state.tunnel.start({ enable: true });
  if (wasLocalOnly && state.config.server.access === 'tunnel') {
    await state.persistConfig();
  }
  if (tunnel?.url) {
    printPublicTunnelUrls(tunnel.url);
  }
  if ((state.config.clients?.length ?? 0) === 0 && state.config.server.access === 'tunnel') {
    console.log(chalk.dim('Lease links and MCP can use the same lease token.'));
  }
}

async function handleTunnelRefresh(state: InteractivePromptState): Promise<void> {
  if (!state.tunnel.configured) {
    console.log(chalk.yellow('No tunnel is configured. Run `tunnel config` or `mvmt tunnel config`.'));
    return;
  }
  console.log('Refreshing tunnel...');
  const wasLocalOnly = state.config.server.access !== 'tunnel';
  const tunnel = await state.tunnel.refresh({ enable: true });
  if (wasLocalOnly && state.config.server.access === 'tunnel') {
    await state.persistConfig();
  }
  if (tunnel?.url) {
    printPublicTunnelUrls(tunnel.url);
  } else {
    console.log(chalk.yellow('Tunnel refreshed, but no public URL was detected yet.'));
  }
}

function localDashboardUrl(port: number): string {
  return `http://127.0.0.1:${port}/dashboard`;
}

function localMcpUrl(port: number): string {
  return `http://127.0.0.1:${port}/mcp`;
}

async function handleTunnelConfig(state: InteractivePromptState): Promise<void> {
  const tunnel = await promptForTunnelConfig(state.port);
  const applied = applyTunnelConfig(state.config, tunnel);
  const wasRunning = state.tunnel.running;

  if (wasRunning) {
    await state.tunnel.stop();
  }

  state.config.server.access = applied.config.server.access;
  state.config.server.tunnel = tunnel;
  await state.persistConfig();

  console.log(chalk.green(`Tunnel config saved to ${state.configPath}`));
  if (applied.warning) {
    printTunnelEnabledWithNoTokens(applied.warning);
  }

  state.tunnel.configure(tunnel);
  if (!wasRunning) {
    console.log(chalk.dim('Run `tunnel start` or `tunnel refresh` to launch it.'));
    return;
  }

  console.log('Restarting tunnel with new config...');
  const restarted = await state.tunnel.start();
  if (restarted?.url) {
    printPublicTunnelUrls(restarted.url);
  }
}

function printPublicTunnelUrls(publicUrl: string): void {
  console.log(`dashboard   ${chalk.yellow(formatDashboardPublicUrl(publicUrl))}`);
  console.log(`MCP         ${chalk.dim(formatMcpPublicUrl(publicUrl))}`);
}

async function handleTunnelStop(state: InteractivePromptState): Promise<void> {
  if (!state.tunnel.running) {
    console.log(chalk.dim('Tunnel is not running.'));
    return;
  }
  await state.tunnel.stop();
  console.log(chalk.dim('Tunnel stopped.'));
}

function printTunnelLogs(state: InteractivePromptState): void {
  const logs = state.tunnel.recentLogs();
  console.log(chalk.bold('tunnel logs'));
  if (logs.length === 0) {
    console.log(`  ${chalk.dim('no tunnel output captured')}`);
    return;
  }
  for (const line of logs.slice(-20)) {
    console.log(`  ${line}`);
  }
}

function printConnectorSummary(loaded: LoadedConnector[], totalTools: number): void {
  console.log(chalk.bold('connectors'));
  for (const entry of loaded) {
    console.log(`  ${chalk.green('ok')} ${entry.connector.id.padEnd(22)} ${String(entry.toolCount).padStart(3)} tools`);
  }
  console.log(`  ${chalk.dim('total'.padEnd(25))} ${String(totalTools).padStart(3)} tools`);
}

async function handleLeaseCreate(state: InteractivePromptState): Promise<void> {
  const folder = await input({
    message: 'Path to lease:',
    validate: (value) => value.trim().length > 0 ? true : 'Enter a file or folder path',
  });
  const label = await input({
    message: 'Lease label:',
    validate: (value) => value.trim().length > 0 ? true : 'Enter a label such as "Sarah - tax docs"',
  });
  const mode = await select({
    message: 'Lease mode:',
    choices: [
      { name: 'Browse/download', value: 'read' },
      { name: 'Upload only', value: 'upload' },
    ],
  });
  const expires = await input({
    message: 'Expires after:',
    default: DEFAULT_LEASE_TTL,
  });
  await createFolderLease(folder.trim(), {
    config: state.configPath,
    label: label.trim(),
    mode,
    expires: expires.trim() || DEFAULT_LEASE_TTL,
  });
  state.config = loadConfig(state.configPath);
}

async function handleLeaseAddPath(state: InteractivePromptState): Promise<void> {
  const leases = listLeases(defaultLeasesPath())
    .filter((lease) => !leaseUnavailableReason(lease) && !lease.permissions.includes('upload'));
  if (leases.length === 0) {
    console.log(chalk.yellow('No active read leases configured.'));
    return;
  }
  const id = await select({
    message: 'Add paths to which lease?',
    choices: leases.map((lease) => ({
      name: `${lease.label} (${lease.resources.map((resource) => resource.path).join(', ')})`,
      value: lease.id,
    })),
  });
  const rawPaths = await input({
    message: 'Path(s) to add, comma-separated:',
    validate: (value) => value.trim().length > 0 ? true : 'Enter at least one file or folder path',
  });
  const paths = rawPaths.split(',').map((value) => value.trim()).filter(Boolean);
  await addPathsToLease(id, paths, { config: state.configPath });
  state.config = loadConfig(state.configPath);
}

async function handleLeaseRevoke(): Promise<void> {
  const leases = listLeases(defaultLeasesPath());
  if (leases.length === 0) {
    console.log(chalk.yellow('No leases configured.'));
    return;
  }
  const id = await select({
    message: 'Revoke which lease?',
    choices: leases.map((lease) => ({
      name: `${lease.label} (${lease.resources.map((resource) => resource.path).join(', ')})`,
      value: lease.id,
    })),
  });
  const ok = await confirm({
    message: `Revoke lease ${id}? Existing links will stop working.`,
    default: false,
  });
  if (!ok) {
    console.log(chalk.yellow('Lease unchanged.'));
    return;
  }
  await revokeFolderLease(id);
}

async function handleUsersAdd(): Promise<void> {
  const username = await input({
    message: 'Username:',
    validate: (value) => value.trim().length > 0 ? true : 'Enter a username',
  });
  const admin = await confirm({
    message: 'Allow this user to manage sources?',
    default: false,
  });
  await addPrivilegedUserCommand(username.trim(), { admin });
}

async function handleUsersAdmin(admin: boolean): Promise<void> {
  const username = await input({
    message: admin ? 'Allow local source management for:' : 'Remove local source management from:',
    validate: (value) => value.trim().length > 0 ? true : 'Enter a username',
  });
  await setPrivilegedUserAdminCommand(username.trim(), admin);
}

async function handleUsersRemove(): Promise<void> {
  const username = await input({
    message: 'Remove dashboard user:',
    validate: (value) => value.trim().length > 0 ? true : 'Enter a username',
  });
  await removePrivilegedUserCommand(username.trim());
}

async function handleMountsAdd(state: InteractivePromptState): Promise<void> {
  const result = await promptAndAddMount(state.config);
  if (!result) return;
  state.config.mounts = result.config.mounts;
  await state.persistConfig();
  console.log(chalk.green(`Mounts saved to ${state.configPath}`));
  printMountBasePermission(findMount(result.config, result.name).writeAccess);
  console.log(chalk.dim(MOUNT_LOAD_NOTICE));
}

async function handleMountsEdit(state: InteractivePromptState): Promise<void> {
  const result = await promptAndEditMount(state.config);
  if (!result) return;
  state.config.mounts = result.config.mounts;
  await state.persistConfig();
  console.log(chalk.green(`Mounts saved to ${state.configPath}`));
  printMountBasePermission(findMount(result.config, result.name).writeAccess);
  console.log(chalk.dim(MOUNT_LOAD_NOTICE));
}

async function handleMountsRemove(state: InteractivePromptState): Promise<void> {
  const nextConfig = await promptAndRemoveMount(state.config);
  if (!nextConfig) {
    console.log(chalk.yellow('Mount config unchanged.'));
    return;
  }
  state.config.mounts = nextConfig.mounts;
  await state.persistConfig();
  console.log(chalk.green(`Mounts saved to ${state.configPath}`));
  console.log(chalk.dim(MOUNT_UNLOAD_NOTICE));
}

async function handleTokensAdd(state: InteractivePromptState): Promise<void> {
  const result = await promptAndAddApiToken(state.config);
  if (!result) return;
  state.config.clients = result.config.clients;
  await state.persistConfig();
  printApiTokenSaved(state.configPath, result);
}

async function handleTokensEdit(state: InteractivePromptState): Promise<void> {
  const result = await promptAndEditApiToken(state.config);
  if (!result) return;
  state.config.clients = result.config.clients;
  await state.persistConfig();
  printApiTokenSaved(state.configPath, result);
}

async function handleTokensRotate(state: InteractivePromptState): Promise<void> {
  const tokenId = await promptForInteractiveTokenId(state.config, 'Rotate which API token?');
  const ok = await confirm({
    message: `Rotate token ${tokenId}? Connected clients using it will lose access until reconfigured.`,
    default: false,
  });
  if (!ok) {
    console.log(chalk.yellow('Token config unchanged.'));
    return;
  }
  const result = rotateApiTokenInConfig(state.config, tokenId);
  state.config.clients = result.config.clients;
  await state.persistConfig();
  printApiTokenSaved(state.configPath, result);
}

async function handleTokensRemove(state: InteractivePromptState): Promise<void> {
  const nextConfig = await promptAndRemoveApiToken(state.config);
  if (!nextConfig) {
    console.log(chalk.yellow('API token config unchanged.'));
    return;
  }
  state.config.clients = nextConfig.clients;
  await state.persistConfig();
  console.log(chalk.green(`API token config saved to ${state.configPath}`));
}

async function promptForInteractiveTokenId(config: MvmtConfig, message: string): Promise<string> {
  const tokens = (config.clients ?? []).filter((client) => client.auth.type === 'token');
  if (tokens.length === 0) throw new Error('No API tokens configured.');
  return select({
    message,
    choices: tokens.map((client) => ({
      name: `${client.id} (${client.name})`,
      value: client.id,
    })),
  });
}

export function formatAuditEntry(entry: AuditEntry): string {
  const status = entry.isError ? chalk.red('ERR') : chalk.green('OK ');
  const time = chalk.dim(new Date(entry.ts).toLocaleTimeString());
  const args = entry.argKeys.length > 0 ? chalk.dim(` args=${entry.argKeys.join(',')}`) : '';
  return `${time} ${status} ${chalk.cyan(entry.connectorId)} ${entry.tool}${args} ${chalk.dim(`${entry.durationMs}ms`)}`;
}

export function formatHttpRequestEntry(entry: HttpRequestLogEntry): string {
  const status =
    entry.status >= 500
      ? chalk.red(String(entry.status))
      : entry.status >= 400
        ? chalk.yellow(String(entry.status))
        : chalk.green(String(entry.status));
  const time = chalk.dim(new Date(entry.ts).toLocaleTimeString());
  const client = entry.clientId ? chalk.dim(` client=${entry.clientId}`) : '';
  const ip = entry.ip ? chalk.dim(` from=${entry.ip}`) : '';
  const detail = entry.detail ? chalk.dim(` ${entry.detail}`) : '';
  return `${time} ${status} ${chalk.magenta(entry.kind)} ${entry.method} ${entry.path}${client}${ip}${detail}`;
}
