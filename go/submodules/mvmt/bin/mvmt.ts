#!/usr/bin/env node

import { Command } from 'commander';
import { addApiToken, editApiToken, listApiTokens, removeApiToken, rotateApiToken, setApiTokenPublished } from '../src/cli/api-tokens.js';
import { withInheritedConfig } from '../src/cli/command-options.js';
import { addConnector, listConnectors } from '../src/cli/connectors.js';
import { runConfigSetup, showConfig } from '../src/cli/config.js';
import { doctor } from '../src/cli/doctor.js';
import { init } from '../src/cli/init.js';
import { addPathsToLease, createFolderLease, listFolderLeases, revokeFolderLease, setFolderLeasePublished } from '../src/cli/lease.js';
import { reindex } from '../src/cli/reindex.js';
import { addMount, editMount, listMounts, removeMount } from '../src/cli/mounts.js';
import { start } from '../src/cli/start.js';
import {
  configureTunnel,
  disableTunnelAccess,
  refreshTunnelCommand,
  showTunnel,
  showTunnelLogs,
  startTunnelCommand,
  stopTunnelCommand,
  streamTunnelLogs,
} from '../src/cli/tunnel.js';
import { rotateToken, showToken, showTokenSummary } from '../src/cli/token.js';
import { addPrivilegedUserCommand, listPrivilegedUsersCommand, removePrivilegedUserCommand, setPrivilegedUserAdminCommand, setPrivilegedUserPasswordCommand } from '../src/cli/users.js';
import { maybePrintUpdateNotice, readPackageInfo } from '../src/utils/version.js';

const packageInfo = readPackageInfo();
const argv = normalizeHelpArgs(process.argv.slice(2));

if (argv.includes('--version') || argv.includes('-V')) {
  console.log(`mvmt ${packageInfo.version}`);
  if (!argv.includes('--no-update-check')) {
    await maybePrintUpdateNotice({ packageInfo });
  }
  process.exit(0);
}

const program = new Command();

program
  .name('mvmt')
  .description('Share local files and folders through dashboard links, leases, and MCP')
  .option('-V, --version', 'Print mvmt version and check for updates')
  .option('--no-update-check', 'Skip automatic npm update checks');
program.showHelpAfterError();
program.showSuggestionAfterError();
program.addHelpText('after', examples([
  ['mvmt serve -i', 'start the local dashboard and control prompt'],
  ['mvmt users add owner --admin', 'create a local dashboard login that can manage sources'],
  ['mvmt lease create ~/Taxes ~/Receipts --label "Sarah - tax docs"', 'create one 24h shared link for multiple paths'],
  ['mvmt tunnel config', 'configure public access for dashboard and shared links'],
  ['mvmt serve --path ~/Documents', 'temporarily expose one read-only source for this run'],
  ['mvmt token add codex --scope notes:read', 'advanced: create a scoped MCP/API token'],
  ['mvmt doctor', 'validate config, sources, and runtime health'],
]));

program.hook('preAction', async (thisCommand, actionCommand) => {
  const globalOptions = thisCommand.opts<{ updateCheck?: boolean }>();
  if (globalOptions.updateCheck === false) return;
  if (actionCommand.name() === 'doctor') return;
  if (isTokenCommand(actionCommand) || isTokenShortcutCommand(actionCommand)) return;

  const actionOptions = actionCommand.opts<{ stdio?: boolean }>();
  if ((actionCommand.name() === 'serve' || actionCommand.name() === 'start') && actionOptions.stdio) return;

  await maybePrintUpdateNotice({ packageInfo });
});

program
  .command('serve')
  .description('Start mvmt with the dashboard, shared links, and MCP endpoint')
  .option('-p, --port <number>', 'Override port')
  .option('-c, --config <path>', 'Config file path')
  .option('--path <dir>', 'Temporarily expose a filesystem folder as read-only for this run only (repeatable)', collectValues)
  .option('--stdio', 'Use stdio transport')
  .option('-i, --interactive', 'Start an interactive control prompt')
  .option('--relay-url <url>', 'Connect to an mvmt relay WebSocket, such as ws://localhost:8080/connect')
  .option('--relay-workspace <slug>', 'Workspace slug to claim on the relay')
  .option('--relay-token <token>', 'Agent token for the relay workspace')
  .option('-v, --verbose', 'Verbose logging')
  .addHelpText('after', examples([
    ['mvmt serve -i', 'start the dashboard with the interactive control prompt'],
    ['mvmt serve --path ~/Documents -i', 'temporarily expose one read-only source and open the prompt'],
    ['mvmt serve --stdio', 'advanced: start stdio MCP mode for a client that launches mvmt'],
  ]))
  .action(async (options: { port?: string; config?: string; path?: string[]; stdio?: boolean; interactive?: boolean; verbose?: boolean; relayUrl?: string; relayWorkspace?: string; relayToken?: string }) => {
    await start(options);
  });

program
  .command('doctor')
  .description('Validate config, sources, and runtime health')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .option('--timeout-ms <number>', 'Per-connector health-check timeout in milliseconds')
  .action(async (options: { config?: string; json?: boolean; timeoutMs?: string }) => {
    const globalOptions = program.opts<{ updateCheck?: boolean }>();
    await doctor({ ...options, updateCheck: globalOptions.updateCheck !== false });
  });

program
  .command('reindex')
  .description('Rebuild the local text search index')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }) => {
    await reindex(options);
  });

const mountsCommand = program
  .command('mounts')
  .description('Manage local sources (advanced)')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }) => {
    await listMounts(options);
  });

mountsCommand
  .command('list')
  .description('List configured local sources')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }, command: Command) => {
    await listMounts(withInheritedConfig(options, command));
  });

mountsCommand
  .command('add [name] [root]')
  .description('Add a local file or folder source')
  .option('-c, --config <path>', 'Config file path')
  .option('--mount-path <path>', 'Virtual mount path, such as /notes')
  .option('--write', 'Allow write/remove tools for this mount')
  .option('--read-only', 'Keep this mount read-only')
  .option('--description <text>', 'Short description shown to clients when listing mounts')
  .option('--guidance <text>', 'Mount-specific instructions shown to clients when listing mounts')
  .option('--exclude <pattern>', 'Exclude glob pattern (repeatable)', collectValues)
  .option('--protect <pattern>', 'Protected write/remove glob pattern (repeatable)', collectValues)
  .option('--disabled', 'Add the mount disabled')
  .addHelpText('after', examples([
    ['mvmt mounts add notes ~/notes --mount-path /notes --read-only', 'add a read-only notes mount'],
    ['mvmt mounts add report ~/report.pdf --mount-path /report.pdf --read-only', 'add a single-file mount'],
    ['mvmt mounts add workspace ~/code/mvmt --mount-path /workspace --write', 'add a writable project mount'],
  ]))
  .action(async (name: string | undefined, root: string | undefined, options: { config?: string; mountPath?: string; write?: boolean; readOnly?: boolean; description?: string; guidance?: string; exclude?: string[]; protect?: string[]; disabled?: boolean }, command: Command) => {
    await addMount(name, root, withInheritedConfig(options, command));
  });

mountsCommand
  .command('edit [name]')
  .description('Edit a local source')
  .option('-c, --config <path>', 'Config file path')
  .option('--root <path>', 'New local folder root')
  .option('--mount-path <path>', 'New virtual mount path, such as /notes')
  .option('--write', 'Allow write/remove tools for this mount')
  .option('--read-only', 'Make this mount read-only')
  .option('--description <text>', 'Replace the mount description shown to clients')
  .option('--guidance <text>', 'Replace the mount-specific instructions shown to clients')
  .option('--exclude <pattern>', 'Replace exclude glob patterns (repeatable)', collectValues)
  .option('--protect <pattern>', 'Replace protected write/remove glob patterns (repeatable)', collectValues)
  .option('--enable', 'Enable this mount')
  .option('--disable', 'Disable this mount')
  .action(async (name: string | undefined, options: { config?: string; root?: string; mountPath?: string; write?: boolean; readOnly?: boolean; description?: string; guidance?: string; exclude?: string[]; protect?: string[]; enable?: boolean; disable?: boolean }, command: Command) => {
    await editMount(name, withInheritedConfig(options, command));
  });

mountsCommand
  .command('remove [name]')
  .description('Remove a local source')
  .option('-c, --config <path>', 'Config file path')
  .option('-y, --yes', 'Remove without prompting for confirmation')
  .addHelpText('after', examples([
    ['mvmt mounts remove notes', 'prompt before removing the notes mount'],
    ['mvmt mounts remove notes --yes', 'remove the notes mount without an interactive prompt'],
  ]))
  .action(async (name: string | undefined, options: { config?: string; yes?: boolean }, command: Command) => {
    await removeMount(name, withInheritedConfig(options, command));
  });

const leaseCommand = program
  .command('lease')
  .description('Create and manage shared links')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }) => {
    await listFolderLeases(options);
  });

leaseCommand
  .command('list')
  .description('List shared links')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }, command: Command) => {
    await listFolderLeases(withInheritedConfig(options, command));
  });

leaseCommand
  .command('create <paths...>')
  .alias('add')
  .description('Create one shared link for one or more files/folders')
  .option('-c, --config <path>', 'Config file path')
  .requiredOption('--label <text>', 'Required lease label, such as "Sarah - tax docs"')
  .option('--mode <mode>', 'Lease mode: read, upload, two-way, or write', 'read')
  .option('--upload', 'Shortcut for --mode upload')
  .option('--expires <duration>', 'Lease lifetime, such as 24h, 7d, or never')
  .option('--ttl <duration>', 'Alias for --expires')
  .option('--downloads <count>', 'Maximum successful file downloads; use -1 for unlimited')
  .option('--json', 'Output as JSON')
  .addHelpText('after', examples([
    ['mvmt lease create ~/Documents/Taxes ~/Receipts --label "Sarah - tax docs"', 'create a 24h shared link for multiple paths'],
    ['mvmt lease create ~/Downloads/report.pdf --label "Report"', 'create a 24h shared link for one file'],
    ['mvmt lease create ~/Documents/Shared --label "Writable folder" --mode write', 'create a writable shared link'],
    ['mvmt lease create ~/DropBox --label "Sarah uploads" --mode upload', 'create an upload-only folder link'],
    ['mvmt lease create ~/DropBox --label "Sarah exchange" --mode two-way', 'create a browse + upload folder link'],
    ['mvmt lease create ~/secret.txt --label "Invite" --downloads 1 --expires 1h', 'create a one-time download link'],
    ['mvmt lease create ~/Photos --label "Family photos" --downloads -1', 'share with unlimited downloads'],
    ['mvmt lease create ~/Photos --label "Family photos" --expires never', 'share until revoked'],
  ]))
  .action(async (paths: string[], options: { config?: string; label?: string; mode?: string; upload?: boolean; expires?: string; ttl?: string; downloads?: string; json?: boolean }, command: Command) => {
    await createFolderLease(paths, withInheritedConfig(options, command));
  });

leaseCommand
  .command('add-path <id> <paths...>')
  .alias('add-paths')
  .description('Add files/folders to an existing read link')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .addHelpText('after', examples([
    ['mvmt lease add-path bz200Pmis1xCT1-8 ~/Documents/Receipts', 'add another folder to an existing lease'],
  ]))
  .action(async (id: string, paths: string[], options: { config?: string; json?: boolean }, command: Command) => {
    await addPathsToLease(id, paths, withInheritedConfig(options, command));
  });

leaseCommand
  .command('revoke <id>')
  .alias('remove')
  .alias('rm')
  .description('Revoke a shared link')
  .option('-c, --config <path>', 'Config file path')
  .action(async (id: string, options: { config?: string }, command: Command) => {
    await revokeFolderLease(id, withInheritedConfig(options, command));
  });

leaseCommand
  .command('publish <id>')
  .description('Allow a shared link to be used through a public tunnel')
  .option('-c, --config <path>', 'Config file path')
  .action(async (id: string, options: { config?: string }, command: Command) => {
    await setFolderLeasePublished(id, true, withInheritedConfig(options, command));
  });

leaseCommand
  .command('unpublish <id>')
  .description('Restrict a shared link to apps on this machine; block public tunnels')
  .option('-c, --config <path>', 'Config file path')
  .action(async (id: string, options: { config?: string }, command: Command) => {
    await setFolderLeasePublished(id, false, withInheritedConfig(options, command));
  });

const usersCommand = program
  .command('users')
  .description('Manage dashboard users')
  .option('--json', 'Output as JSON')
  .action(async (options: { json?: boolean }) => {
    await listPrivilegedUsersCommand(options);
  });

usersCommand
  .command('list')
  .description('List dashboard users')
  .option('--json', 'Output as JSON')
  .action(async (options: { json?: boolean }) => {
    await listPrivilegedUsersCommand(options);
  });

usersCommand
  .command('add <username>')
  .description('Create a dashboard user')
  .option('--password <password>', 'Password for non-interactive setup')
  .option('--admin', 'Grant admin (can manage sources from dashboard)')
  .option('--json', 'Output as JSON')
  .addHelpText('after', examples([
    ['mvmt users add sarah', 'prompt for a password and create sarah'],
    ['mvmt users add sarah --password "change-me-now"', 'non-interactive setup'],
    ['mvmt users add owner --admin', 'create an admin who can manage dashboard sources'],
  ]))
  .action(async (username: string | undefined, _options: { password?: string; json?: boolean; admin?: boolean }, command: Command) => {
    await addPrivilegedUserCommand(username, {
      password: command.getOptionValue('password') as string | undefined,
      admin: Boolean(command.getOptionValue('admin')),
      json: Boolean(command.getOptionValue('json') || command.parent?.getOptionValue('json')),
    });
  });

usersCommand
  .command('password <username>')
  .description('Rotate a dashboard user password')
  .option('--password <password>', 'New password for non-interactive setup')
  .option('--json', 'Output as JSON')
  .addHelpText('after', examples([
    ['mvmt users password sarah', 'prompt for a new password'],
    ['mvmt users password sarah --password "new-long-password"', 'rotate non-interactively'],
  ]))
  .action(async (username: string | undefined, _options: { password?: string; json?: boolean }, command: Command) => {
    await setPrivilegedUserPasswordCommand(username, {
      password: command.getOptionValue('password') as string | undefined,
      json: Boolean(command.getOptionValue('json') || command.parent?.getOptionValue('json')),
    });
  });

usersCommand
  .command('remove <username>')
  .alias('delete')
  .description('Remove a dashboard user')
  .option('-y, --yes', 'Remove without prompting for confirmation')
  .option('--json', 'Output as JSON')
  .addHelpText('after', examples([
    ['mvmt users remove sarah', 'prompt before removing sarah'],
    ['mvmt users remove sarah --yes', 'remove without an interactive prompt'],
  ]))
  .action(async (username: string | undefined, _options: { yes?: boolean; json?: boolean }, command: Command) => {
    await removePrivilegedUserCommand(username, {
      yes: Boolean(command.getOptionValue('yes')),
      json: Boolean(command.getOptionValue('json') || command.parent?.getOptionValue('json')),
    });
  });

usersCommand
  .command('grant <username>')
  .description('Allow a dashboard user to manage local sources')
  .option('--json', 'Output as JSON')
  .action(async (username: string | undefined, _options: { json?: boolean }, command: Command) => {
    await setPrivilegedUserAdminCommand(username, true, {
      json: Boolean(command.getOptionValue('json') || command.parent?.getOptionValue('json')),
    });
  });

usersCommand
  .command('revoke <username>')
  .description('Remove local source management from a dashboard user')
  .option('--json', 'Output as JSON')
  .action(async (username: string | undefined, _options: { json?: boolean }, command: Command) => {
    await setPrivilegedUserAdminCommand(username, false, {
      json: Boolean(command.getOptionValue('json') || command.parent?.getOptionValue('json')),
    });
  });

const configCommand = program
  .command('config')
  .description('Show saved mvmt config (advanced)')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }) => {
    await showConfig(options);
  });

configCommand
  .command('setup')
  .description('Run guided setup and save mvmt config')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }, command: Command) => {
    await runConfigSetup({ config: withInheritedConfig(options, command).config });
  });

const tokenCommand = program
  .command('token')
  .description('Manage scoped MCP/API tokens (advanced)')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }) => {
    await listApiTokens(options);
  });

tokenCommand
  .command('list')
  .description('List scoped MCP/API tokens')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }, command: Command) => {
    await listApiTokens(withInheritedConfig(options, command));
  });

tokenCommand
  .command('show')
  .description('List scoped MCP/API tokens')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }, command: Command) => {
    await listApiTokens(withInheritedConfig(options, command));
  });

tokenCommand
  .command('add [id]')
  .alias('create')
  .description('Create a scoped MCP/API token and print it once')
  .option('-c, --config <path>', 'Config file path')
  .option('--name <text>', 'Display name')
  .option('--description <text>', 'Optional description')
  .option('--scope <scope>', 'Grant scope such as all:read or notes:write (repeatable, comma-separated)', collectValues)
  .option('--client <identity>', 'Bind token to a client identity label')
  .option('--expires <duration>', 'Token lifetime, such as 30m, 7d, 30d, or never')
  .option('--ttl <duration>', 'Alias for --expires')
  .option('--read <mount>', 'Grant search/read on a mount name or path (repeatable)', collectValues)
  .option('--write <mount>', 'Grant search/read/write/remove on a writable mount (repeatable)', collectValues)
  .addHelpText('after', examples([
    ['mvmt token add codex --scope notes:read --expires 7d', 'create a token that can search/read notes for 7 days'],
    ['mvmt token add laptop-claude --scope all:read --client claude-desktop --expires never', 'create a read-only token for Claude Desktop'],
    ['mvmt token add', 'interactive token setup'],
  ]))
  .action(async (id: string | undefined, options: { config?: string; name?: string; description?: string; scope?: string[]; client?: string; expires?: string; ttl?: string; read?: string[]; write?: string[] }, command: Command) => {
    await addApiToken(id, withInheritedConfig(options, command));
  });

tokenCommand
  .command('edit [id]')
  .description('Edit a scoped MCP/API token')
  .option('-c, --config <path>', 'Config file path')
  .option('--name <text>', 'Display name')
  .option('--description <text>', 'Optional description')
  .option('--scope <scope>', 'Replace scopes, such as all:read or notes:write (repeatable, comma-separated)', collectValues)
  .option('--no-permissions', 'Replace scopes with no access')
  .option('--client <identity>', 'Replace client identity binding, or "any" to clear')
  .option('--expires <duration>', 'Token lifetime, such as 30m, 7d, 30d, or never')
  .option('--ttl <duration>', 'Alias for --expires')
  .option('--read <mount>', 'Grant search/read on a mount name or path (repeatable)', collectValues)
  .option('--write <mount>', 'Grant search/read/write/remove on a writable mount (repeatable)', collectValues)
  .action(async (id: string | undefined, options: { config?: string; name?: string; description?: string; scope?: string[]; permissions?: boolean; client?: string; expires?: string; ttl?: string; read?: string[]; write?: string[] }, command: Command) => {
    await editApiToken(id, withInheritedConfig(options, command));
  });

tokenCommand
  .command('rotate [id]')
  .description('Rotate a scoped MCP/API token and print the replacement once')
  .option('-c, --config <path>', 'Config file path')
  .option('--expires <duration>', 'Replacement lifetime, such as 30m, 7d, 30d, or never')
  .option('--ttl <duration>', 'Alias for --expires')
  .option('-y, --yes', 'Rotate without prompting for confirmation')
  .action(async (id: string | undefined, options: { config?: string; expires?: string; ttl?: string; yes?: boolean }, command: Command) => {
    await rotateApiToken(id, withInheritedConfig(options, command));
  });

tokenCommand
  .command('remove [id]')
  .description('Remove a scoped MCP/API token')
  .option('-c, --config <path>', 'Config file path')
  .option('-y, --yes', 'Remove without prompting for confirmation')
  .action(async (id: string | undefined, options: { config?: string; yes?: boolean }, command: Command) => {
    await removeApiToken(id, withInheritedConfig(options, command));
  });

tokenCommand
  .command('publish [id]')
  .description('Allow a scoped MCP/API token to be used through a public tunnel')
  .option('-c, --config <path>', 'Config file path')
  .action(async (id: string | undefined, options: { config?: string }, command: Command) => {
    await setApiTokenPublished(id, true, withInheritedConfig(options, command));
  });

tokenCommand
  .command('unpublish [id]')
  .description('Restrict a token to apps on this machine; block public tunnels')
  .option('-c, --config <path>', 'Config file path')
  .action(async (id: string | undefined, options: { config?: string }, command: Command) => {
    await setApiTokenPublished(id, false, withInheritedConfig(options, command));
  });

tokenCommand
  .command('session', { hidden: true })
  .description('Show the internal legacy session bearer token summary')
  .action(async () => {
    await showTokenSummary();
  });

tokenCommand
  .command('session-raw', { hidden: true })
  .description('Print the internal legacy session bearer token')
  .action(async () => {
    await showToken();
  });

tokenCommand
  .command('session-rotate', { hidden: true })
  .description('Regenerate and print the internal legacy session bearer token')
  .action(async () => {
    await rotateToken();
  });

const apiTokensCommand = program
  .command('tokens')
  .description('Manage scoped MCP/API tokens (advanced)')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }) => {
    await listApiTokens(options);
  });

apiTokensCommand
  .command('list')
  .description('List scoped MCP/API tokens')
  .option('-c, --config <path>', 'Config file path')
  .option('--json', 'Output as JSON')
  .action(async (options: { config?: string; json?: boolean }, command: Command) => {
    await listApiTokens(withInheritedConfig(options, command));
  });

apiTokensCommand
  .command('add [id]')
  .alias('create')
  .description('Create a scoped MCP/API token and print it once')
  .option('-c, --config <path>', 'Config file path')
  .option('--name <text>', 'Display name')
  .option('--description <text>', 'Optional description')
  .option('--scope <scope>', 'Grant scope such as all:read or notes:write (repeatable, comma-separated)', collectValues)
  .option('--client <identity>', 'Bind token to a client identity label')
  .option('--expires <duration>', 'Token lifetime, such as 30m, 7d, 30d, or never')
  .option('--ttl <duration>', 'Alias for --expires')
  .option('--read <mount>', 'Grant search/read on a mount name or path (repeatable)', collectValues)
  .option('--write <mount>', 'Grant search/read/write/remove on a writable mount (repeatable)', collectValues)
  .addHelpText('after', examples([
    ['mvmt token add codex --scope notes:read --expires 7d', 'create a token that can search/read notes for 7 days'],
    ['mvmt token add laptop-claude --scope all:read --client claude-desktop --expires never', 'create a read-only token for Claude Desktop'],
    ['mvmt token add', 'interactive token setup'],
  ]))
  .action(async (id: string | undefined, options: { config?: string; name?: string; description?: string; scope?: string[]; client?: string; expires?: string; ttl?: string; read?: string[]; write?: string[] }, command: Command) => {
    await addApiToken(id, withInheritedConfig(options, command));
  });

apiTokensCommand
  .command('edit [id]')
  .description('Edit a scoped MCP/API token')
  .option('-c, --config <path>', 'Config file path')
  .option('--name <text>', 'Display name')
  .option('--description <text>', 'Optional description')
  .option('--scope <scope>', 'Replace scopes, such as all:read or notes:write (repeatable, comma-separated)', collectValues)
  .option('--no-permissions', 'Replace scopes with no access')
  .option('--client <identity>', 'Replace client identity binding, or "any" to clear')
  .option('--expires <duration>', 'Token lifetime, such as 30m, 7d, 30d, or never')
  .option('--ttl <duration>', 'Alias for --expires')
  .option('--read <mount>', 'Grant search/read on a mount name or path (repeatable)', collectValues)
  .option('--write <mount>', 'Grant search/read/write/remove on a writable mount (repeatable)', collectValues)
  .action(async (id: string | undefined, options: { config?: string; name?: string; description?: string; scope?: string[]; permissions?: boolean; client?: string; expires?: string; ttl?: string; read?: string[]; write?: string[] }, command: Command) => {
    await editApiToken(id, withInheritedConfig(options, command));
  });

apiTokensCommand
  .command('rotate [id]')
  .description('Rotate a scoped MCP/API token and print the replacement once')
  .option('-c, --config <path>', 'Config file path')
  .option('--expires <duration>', 'Replacement lifetime, such as 30m, 7d, 30d, or never')
  .option('--ttl <duration>', 'Alias for --expires')
  .option('-y, --yes', 'Rotate without prompting for confirmation')
  .action(async (id: string | undefined, options: { config?: string; expires?: string; ttl?: string; yes?: boolean }, command: Command) => {
    await rotateApiToken(id, withInheritedConfig(options, command));
  });

apiTokensCommand
  .command('remove [id]')
  .description('Remove a scoped MCP/API token')
  .option('-c, --config <path>', 'Config file path')
  .option('-y, --yes', 'Remove without prompting for confirmation')
  .action(async (id: string | undefined, options: { config?: string; yes?: boolean }, command: Command) => {
    await removeApiToken(id, withInheritedConfig(options, command));
  });

apiTokensCommand
  .command('publish [id]')
  .description('Allow a scoped MCP/API token to be used through a public tunnel')
  .option('-c, --config <path>', 'Config file path')
  .action(async (id: string | undefined, options: { config?: string }, command: Command) => {
    await setApiTokenPublished(id, true, withInheritedConfig(options, command));
  });

apiTokensCommand
  .command('unpublish [id]')
  .description('Restrict a token to apps on this machine; block public tunnels')
  .option('-c, --config <path>', 'Config file path')
  .action(async (id: string | undefined, options: { config?: string }, command: Command) => {
    await setApiTokenPublished(id, false, withInheritedConfig(options, command));
  });

const tunnelCommand = program
  .command('tunnel')
  .description('Manage public access for dashboard and shared links')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }) => {
    await showTunnel(options);
  });

tunnelCommand
  .command('config')
  .description('Choose a different tunnel and save it to config')
  .option('-c, --config <path>', 'Config file path')
  .option('--quick', 'Configure Cloudflare Quick Tunnel without prompting')
  .option('--cloudflare-config <path>', 'Configure a Cloudflare named tunnel from a cloudflared config file')
  .option('--relay', 'Configure the default MVMT relay without prompting')
  .option('--relay-url <url>', 'Configure a custom relay WebSocket URL')
  .option('--relay-workspace <slug>', 'Relay workspace slug')
  .option('--relay-token <token>', 'Relay agent token/secret')
  .option('--public-url <url>', 'Public base URL for generated dashboard and lease links')
  .action(async (
    options: {
      config?: string;
      quick?: boolean;
      cloudflareConfig?: string;
      relay?: boolean;
      relayUrl?: string;
      relayWorkspace?: string;
      relayToken?: string;
      publicUrl?: string;
    },
    command: Command,
  ) => {
    await configureTunnel(withInheritedConfig(options, command));
  });

tunnelCommand
  .command('start')
  .description('Start the configured tunnel for the running mvmt process')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }, command: Command) => {
    await startTunnelCommand(withInheritedConfig(options, command));
  });

tunnelCommand
  .command('refresh')
  .description('Restart the configured tunnel and print the new URL')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }, command: Command) => {
    await refreshTunnelCommand(withInheritedConfig(options, command));
  });

tunnelCommand
  .command('stop')
  .description('Stop public tunnel exposure without stopping mvmt')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }, command: Command) => {
    await stopTunnelCommand(withInheritedConfig(options, command));
  });

tunnelCommand
  .command('disable')
  .description('Disable tunnel access in config and keep saved tunnel details')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }, command: Command) => {
    await disableTunnelAccess(withInheritedConfig(options, command));
  });

const tunnelLogsCommand = tunnelCommand
  .command('logs')
  .description('Show recent tunnel output')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }, command: Command) => {
    await showTunnelLogs(withInheritedConfig(options, command));
  });

tunnelLogsCommand
  .command('stream')
  .description('Stream live tunnel output from the running mvmt process')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }, command: Command) => {
    await streamTunnelLogs(withInheritedConfig(options, command));
  });

const connectorsCommand = program
  .command('connectors', { hidden: true })
  .description('Compatibility connector helpers');

connectorsCommand
  .command('list')
  .description('Show supported connector setup status')
  .option('-c, --config <path>', 'Config file path')
  .action(async (options: { config?: string }, command: Command) => {
    await listConnectors(withInheritedConfig(options, command));
  });

connectorsCommand
  .command('add [name]')
  .description('Add a supported connector setup to config')
  .option('-c, --config <path>', 'Config file path')
  .action(async (name: string | undefined, options: { config?: string }, command: Command) => {
    await addConnector(name, withInheritedConfig(options, command));
  });

program
  .command('init', { hidden: true })
  .description('Compatibility alias for `mvmt config setup`')
  .action(async () => {
    await init();
  });

program
  .command('start', { hidden: true })
  .description('Compatibility alias for `mvmt serve`')
  .option('-p, --port <number>', 'Override port')
  .option('-c, --config <path>', 'Config file path')
  .option('--path <dir>', 'Temporarily expose a filesystem folder as read-only for this run only (repeatable)', collectValues)
  .option('--stdio', 'Use stdio transport')
  .option('-i, --interactive', 'Start an interactive control prompt')
  .option('-v, --verbose', 'Verbose logging')
  .action(async (options: { port?: string; config?: string; path?: string[]; stdio?: boolean; interactive?: boolean; verbose?: boolean }) => {
    await start(options);
  });

program
  .command('show', { hidden: true })
  .description('Compatibility alias for raw token output')
  .action(async () => {
    await showToken();
  });

program
  .command('rotate', { hidden: true })
  .description('Compatibility alias for internal session-token rotation')
  .action(async () => {
    await rotateToken();
  });

await program.parseAsync(['node', process.argv[1] ?? 'mvmt', ...argv]);

function isTokenCommand(command: Command): boolean {
  let current: Command | null = command;
  while (current) {
    if (current.name() === 'token' || current.name() === 'tokens') return true;
    current = current.parent;
  }
  return false;
}

function isTokenShortcutCommand(command: Command): boolean {
  return command.name() === 'show' || command.name() === 'rotate';
}

function normalizeHelpArgs(args: string[]): string[] {
  if (args.length > 1 && args.at(-1) === 'help' && args[0] !== 'help' && !args.includes('--help') && !args.includes('-h')) {
    return [...args.slice(0, -1), '--help'];
  }
  return args;
}

function collectValues(value: string, previous?: string[]): string[] {
  return [...(previous ?? []), value];
}

function examples(rows: Array<[command: string, description: string]>): string {
  return [
    '',
    'Examples:',
    ...rows.flatMap(([command, description]) => [`  ${command}`, `      ${description}`]),
  ].join('\n');
}
