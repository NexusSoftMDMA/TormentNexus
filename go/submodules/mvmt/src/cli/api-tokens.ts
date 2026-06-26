import crypto from 'crypto';
import fs from 'fs';
import path from 'path';
import { confirm, input, select } from '@inquirer/prompts';
import chalk from 'chalk';
import { getConfigPath, loadConfig, saveConfig, withConfigLock } from '../config/loader.js';
import {
  ClientConfig,
  ConfigSchema,
  LocalFolderMountConfig,
  MvmtConfig,
  PermissionConfig,
  resolveProxySourceId,
} from '../config/schema.js';
import { resolveSetupPath } from '../connectors/setup-paths.js';
import { AUDIT_LOG_PATH, createAuditLogger, type AuditEntry } from '../utils/audit.js';
import { hashApiToken } from '../utils/api-token-hash.js';
import { formatTokenExpiry, parseTokenTtl } from './token-ttl.js';

const API_TOKEN_ID_RE = /^[a-z0-9][a-z0-9_-]*$/;
const TOKEN_PREFIX = 'mvmt_t_';

export interface ApiTokensCommandOptions {
  config?: string;
  json?: boolean;
}

export interface AddApiTokenOptions extends ApiTokensCommandOptions {
  name?: string;
  description?: string;
  expires?: string;
  ttl?: string;
  client?: string;
  scope?: string[];
  read?: string[];
  write?: string[];
}

export interface EditApiTokenOptions extends AddApiTokenOptions {
  permissions?: boolean;
}

export interface RemoveApiTokenOptions extends ApiTokensCommandOptions {
  yes?: boolean;
}

export interface RotateApiTokenOptions extends ApiTokensCommandOptions {
  expires?: string;
  ttl?: string;
  yes?: boolean;
}

export type ApiTokenPermissionMode = 'read' | 'write';

export interface ApiTokenPermissionInput {
  source: string;
  mode: ApiTokenPermissionMode;
}

export interface ApiTokenInput {
  id: string;
  name?: string;
  description?: string;
  expires?: string;
  now?: number;
  clientBinding?: string | null;
  permissions: ApiTokenPermissionInput[];
  // Pre-resolved permission entries, used by callers (the dashboard) that
  // scope grants to files/folders rather than whole mounts. When present,
  // `permissions` is ignored and these are stored verbatim.
  resolvedPermissions?: PermissionConfig[];
  replacePermissions?: boolean;
  plaintextToken?: string;
  // Exposure boundary for the new grant. Omit to leave it grandfathered
  // as published; pass false for a capability-only grant.
  published?: boolean;
}

export interface ApiTokenUpdateResult {
  config: MvmtConfig;
  created: boolean;
  plaintextToken?: string;
  client: ClientConfig;
  credentialVersionChanged?: boolean;
  permissionsReplaced?: boolean;
}

interface PolicySource {
  id: string;
  label: string;
  path: string;
  writeAccess: boolean;
}

export async function listApiTokens(options: ApiTokensCommandOptions = {}): Promise<void> {
  const configPath = resolveApiTokensConfigPath(options.config);
  const config = loadConfig(configPath);
  printApiTokens(config, {
    json: Boolean(options.json),
    auditLogPath: auditLogPathForConfig(configPath),
  });
}

export function printApiTokens(
  config: MvmtConfig,
  options: { json?: boolean; auditLogPath?: string } = {},
): void {
  const tokens = tokenClients(config);
  const lastUsed = loadTokenLastUsed(options.auditLogPath ?? AUDIT_LOG_PATH);

  if (options.json) {
    console.log(JSON.stringify({
      tokens: tokens.map((client) => toApiTokenSummary(config, client, lastUsed.get(client.id))),
    }, null, 2));
    return;
  }

  if (tokens.length === 0) {
    console.log('No tokens configured.');
    console.log('');
    console.log('Create one with: mvmt token add <name>');
    return;
  }

  const rows = tokens.map((client) => ({
    NAME: client.id,
    SCOPE: formatClientScope(config, client),
    CLIENT: client.clientBinding ?? '(any)',
    CREATED: formatPastTime(client.createdAt),
    'LAST USED': formatLastUsed(lastUsed.get(client.id) ?? client.lastUsedAt),
    EXPIRES: formatExpiryRelative(client.expiresAt),
  }));
  printTable(rows, ['NAME', 'SCOPE', 'CLIENT', 'CREATED', 'LAST USED', 'EXPIRES']);
}

export async function addApiToken(id: string | undefined, options: AddApiTokenOptions = {}): Promise<void> {
  try {
    const configPath = resolveApiTokensConfigPath(options.config);
    const config = loadConfig(configPath);
    const inputValue = hasAddOptions(options)
      ? apiTokenInputFromOptions(config, id, options)
      : await promptForApiTokenInput(config, id, options);

    const result = await withConfigLock(configPath, async () => {
      const latest = loadConfig(configPath);
      const update = addApiTokenToConfig(latest, inputValue);
      await saveConfig(configPath, update.config);
      return update;
    });
    recordTokenLifecycle(configPath, 'token.add', result.config, result.client);
    printApiTokenSaved(configPath, result);
    console.log(chalk.dim('Running mvmt processes load API-token changes on the next auth request.'));
  } catch (err) {
    printApiTokenCommandError(err);
  }
}

export async function editApiToken(id: string | undefined, options: EditApiTokenOptions = {}): Promise<void> {
  try {
    const configPath = resolveApiTokensConfigPath(options.config);
    const config = loadConfig(configPath);
    const tokenId = id ?? await promptForApiTokenId(config, 'Edit which API token?');
    const inputValue = hasEditOptions(options)
      ? apiTokenInputFromOptions(config, tokenId, options, { requirePermissions: false })
      : await promptForApiTokenEditInput(config, tokenId, options);
    if (!inputValue) {
      console.log(chalk.yellow('Token config unchanged.'));
      return;
    }

    const result = await withConfigLock(configPath, async () => {
      const latest = loadConfig(configPath);
      const update = editApiTokenInConfig(latest, tokenId, inputValue);
      await saveConfig(configPath, update.config);
      return update;
    });
    recordTokenLifecycle(configPath, 'token.edit', result.config, result.client);
    printApiTokenSaved(configPath, result);
    console.log(chalk.dim('Running mvmt processes load API-token changes on the next auth request.'));
  } catch (err) {
    printApiTokenCommandError(err);
  }
}

export async function removeApiToken(id: string | undefined, options: RemoveApiTokenOptions = {}): Promise<void> {
  try {
    const configPath = resolveApiTokensConfigPath(options.config);
    const config = loadConfig(configPath);
    const tokenId = id ?? await promptForApiTokenId(config, 'Remove which API token?');
    const ok = options.yes ? true : await confirmDestructive(
      `This will permanently delete '${tokenId}' and disconnect clients using it.`,
    );
    if (!ok) {
      console.log(chalk.yellow('Token config unchanged.'));
      return;
    }

    const update = await withConfigLock(configPath, async () => {
      const latest = loadConfig(configPath);
      const latestExisting = findTokenClient(latest, tokenId);
      const nextConfig = removeApiTokenFromConfig(latest, tokenId);
      await saveConfig(configPath, nextConfig);
      return { config: latest, client: latestExisting };
    });
    recordTokenLifecycle(configPath, 'token.remove', update.config, update.client);
    console.log(chalk.green(`Token '${tokenId}' removed.`));
    console.log(chalk.dim('Running mvmt processes unload API-token changes on the next auth request.'));
  } catch (err) {
    printApiTokenCommandError(err);
  }
}

export async function setApiTokenPublished(
  id: string | undefined,
  published: boolean,
  options: ApiTokensCommandOptions = {},
): Promise<void> {
  try {
    const configPath = resolveApiTokensConfigPath(options.config);
    const config = loadConfig(configPath);
    const tokenId = id ?? await promptForApiTokenId(
      config,
      published ? 'Publish which API token?' : 'Unpublish which API token?',
    );

    const result = await withConfigLock(configPath, async () => {
      const latest = loadConfig(configPath);
      const update = setApiTokenPublishedInConfig(latest, tokenId, published);
      await saveConfig(configPath, update.config);
      return update;
    });
    recordTokenLifecycle(configPath, 'token.edit', result.config, result.client);
    if (published) {
      console.log(chalk.green(`Token '${tokenId}' published - it can now be used through a public tunnel.`));
    } else {
      console.log(chalk.green(`Token '${tokenId}' unpublished - apps on this machine keep access; public tunnels are blocked.`));
    }
    console.log(chalk.dim('Running mvmt processes load API-token changes on the next auth request.'));
  } catch (err) {
    printApiTokenCommandError(err);
  }
}

export async function rotateApiToken(id: string | undefined, options: RotateApiTokenOptions = {}): Promise<void> {
  try {
    const configPath = resolveApiTokensConfigPath(options.config);
    const config = loadConfig(configPath);
    const tokenId = id ?? await promptForApiTokenId(config, 'Rotate which API token?');
    const ok = options.yes ? true : await confirmDestructive(
      `This will invalidate the current '${tokenId}' token. Connected clients lose access until reconfigured.`,
    );
    if (!ok) {
      console.log(chalk.yellow('Token config unchanged.'));
      return;
    }

    const result = await withConfigLock(configPath, async () => {
      const latest = loadConfig(configPath);
      const update = rotateApiTokenInConfig(latest, tokenId, undefined, {
        expires: options.expires ?? options.ttl,
      });
      await saveConfig(configPath, update.config);
      return update;
    });
    recordTokenLifecycle(configPath, 'token.rotate', result.config, result.client);
    printApiTokenSaved(configPath, result);
    console.log(chalk.dim('Running mvmt processes load API-token changes on the next auth request.'));
  } catch (err) {
    printApiTokenCommandError(err);
  }
}

export async function promptAndAddApiToken(config: MvmtConfig): Promise<ApiTokenUpdateResult | undefined> {
  const inputValue = await promptForApiTokenInput(config, undefined, {});
  return addApiTokenToConfig(config, inputValue);
}

export async function promptAndEditApiToken(config: MvmtConfig): Promise<ApiTokenUpdateResult | undefined> {
  const tokenId = await promptForApiTokenId(config, 'Edit which API token?');
  const inputValue = await promptForApiTokenEditInput(config, tokenId, {});
  if (!inputValue) return undefined;
  return editApiTokenInConfig(config, tokenId, inputValue);
}

export async function promptAndRemoveApiToken(config: MvmtConfig): Promise<MvmtConfig | undefined> {
  const tokenId = await promptForApiTokenId(config, 'Remove which API token?');
  const ok = await confirm({ message: `Remove token ${tokenId}?`, default: false });
  return ok ? removeApiTokenFromConfig(config, tokenId) : undefined;
}

export function addApiTokenToConfig(config: MvmtConfig, inputValue: ApiTokenInput): ApiTokenUpdateResult {
  const id = normalizeApiTokenId(inputValue.id);
  const existing = config.clients?.find((client) => client.id === id);
  if (existing) {
    if (existing.auth.type !== 'token') throw new Error(`Client id ${id} already exists and is not an API token.`);
    throw new Error(`API token ${id} already exists. Use mvmt token edit ${id} to change policy or mvmt token rotate ${id} to replace the token value.`);
  }

  const name = (inputValue.name ?? id).trim();
  if (!name) throw new Error('API token display name cannot be empty.');
  const description = (inputValue.description ?? '').trim();
  const expiresAt = resolveExpiresAt(inputValue);
  const permissions = inputValue.resolvedPermissions
    ?? applyPermissionInputs(config, [], inputValue.permissions);
  const plaintextToken = inputValue.plaintextToken ?? generateApiToken();
  const createdAt = new Date(inputValue.now ?? Date.now()).toISOString();
  const clientBinding = normalizeClientBinding(inputValue.clientBinding);
  const nextClient: ClientConfig = {
    id,
    name,
    description,
    createdAt,
    credentialVersion: 1,
    ...(expiresAt ? { expiresAt } : {}),
    ...(clientBinding ? { clientBinding } : {}),
    auth: {
      type: 'token',
      tokenHash: hashApiToken(plaintextToken),
    },
    rawToolsEnabled: false,
    permissions,
    ...(inputValue.published === undefined ? {} : { published: inputValue.published }),
  };
  const clients = [
    ...(config.clients ?? []),
    nextClient,
  ];
  const nextConfig = ConfigSchema.parse({ ...config, clients });
  return { config: nextConfig, created: true, plaintextToken, client: nextClient };
}

export function editApiTokenInConfig(
  config: MvmtConfig,
  id: string,
  inputValue: ApiTokenInput,
): ApiTokenUpdateResult {
  const tokenId = normalizeApiTokenId(id);
  const existing = findTokenClient(config, tokenId);
  const name = (inputValue.name ?? existing.name).trim();
  if (!name) throw new Error('API token display name cannot be empty.');
  const description = (inputValue.description ?? existing.description ?? '').trim();
  const expiresAt = inputValue.expires !== undefined
    ? resolveExpiresAt(inputValue)
    : existing.expiresAt;
  const replacePermissions = inputValue.replacePermissions ?? inputValue.permissions.length > 0;
  const permissions = replacePermissions
    ? applyPermissionInputs(config, [], inputValue.permissions, { allowEmpty: true })
    : existing.permissions;
  const clientBinding = inputValue.clientBinding !== undefined
    ? normalizeClientBinding(inputValue.clientBinding)
    : existing.clientBinding;
  const shouldBumpCredentialVersion = inputValue.clientBinding !== undefined && clientBinding !== existing.clientBinding;
  const { expiresAt: _oldExpiresAt, clientBinding: _oldClientBinding, ...existingWithoutOptionalPolicy } = existing;
  const nextClient: ClientConfig = {
    ...existingWithoutOptionalPolicy,
    id: tokenId,
    name,
    description,
    ...(expiresAt ? { expiresAt } : {}),
    ...(shouldBumpCredentialVersion ? { credentialVersion: (existing.credentialVersion ?? 1) + 1 } : {}),
    ...(clientBinding ? { clientBinding } : {}),
    auth: existing.auth,
    rawToolsEnabled: existing.rawToolsEnabled,
    permissions,
  };
  const clients = [
    ...(config.clients ?? []).filter((client) => client.id !== tokenId),
    nextClient,
  ];
  const nextConfig = ConfigSchema.parse({ ...config, clients });
  return {
    config: nextConfig,
    created: false,
    client: nextClient,
    credentialVersionChanged: shouldBumpCredentialVersion,
    permissionsReplaced: replacePermissions,
  };
}

export function rotateApiTokenInConfig(
  config: MvmtConfig,
  id: string,
  plaintextToken = generateApiToken(),
  options: { expires?: string; ttl?: string; now?: number } = {},
): ApiTokenUpdateResult {
  const tokenId = normalizeApiTokenId(id);
  const existing = findTokenClient(config, tokenId);
  const expiresInput = options.expires ?? options.ttl;
  const expiresAt = expiresInput !== undefined
    ? parseTokenTtl(expiresInput, options.now).expiresAt
    : existing.expiresAt;
  const { expiresAt: _previousExpiresAt, ...existingWithoutExpiry } = existing;
  const nextClient: ClientConfig = {
    ...existingWithoutExpiry,
    ...(expiresAt ? { expiresAt } : {}),
    credentialVersion: (existing.credentialVersion ?? 1) + 1,
    auth: {
      type: 'token',
      tokenHash: hashApiToken(plaintextToken),
    },
  };
  const clients = [
    ...(config.clients ?? []).filter((client) => client.id !== tokenId),
    nextClient,
  ];
  const nextConfig = ConfigSchema.parse({ ...config, clients });
  return { config: nextConfig, created: false, plaintextToken, client: nextClient };
}

// Sets the exposure boundary for an API-token grant. `published: false`
// makes it capability-only: usable only by apps on this machine, never
// through a public tunnel. `published: true` allows remote access.
// Distinct from removeApiTokenFromConfig, which deletes the grant.
export function setApiTokenPublishedInConfig(
  config: MvmtConfig,
  id: string,
  published: boolean,
): ApiTokenUpdateResult {
  const tokenId = normalizeApiTokenId(id);
  const existing = findTokenClient(config, tokenId);
  const nextClient: ClientConfig = { ...existing, published };
  const clients = [
    ...(config.clients ?? []).filter((client) => client.id !== tokenId),
    nextClient,
  ];
  const nextConfig = ConfigSchema.parse({ ...config, clients });
  return { config: nextConfig, created: false, client: nextClient };
}

export function removeApiTokenFromConfig(config: MvmtConfig, id: string): MvmtConfig {
  const tokenId = normalizeApiTokenId(id);
  findTokenClient(config, tokenId);
  const clients = (config.clients ?? []).filter((client) => client.id !== tokenId);
  return ConfigSchema.parse({
    ...config,
    ...(clients.length > 0 ? { clients } : { clients: undefined }),
  });
}

function apiTokenInputFromOptions(
  config: MvmtConfig,
  id: string | undefined,
  options: AddApiTokenOptions | EditApiTokenOptions,
  parseOptions: { requirePermissions?: boolean } = {},
): ApiTokenInput {
  if (!id) throw new Error('Token name is required when using non-interactive token options.');
  const permissions = parsePermissionInputsFromOptions(config, options);
  const noPermissions = 'permissions' in options && options.permissions === false;
  if (noPermissions && permissions.length > 0) {
    throw new Error('Use either --no-permissions or --scope/--read/--write, not both.');
  }
  if (permissions.length === 0 && parseOptions.requirePermissions !== false) {
    throw new Error('Add at least one scope with --scope <scope> (for example all:read).');
  }
  const expires = options.expires ?? options.ttl;
  if (expires !== undefined) parseTokenTtl(expires);
  return {
    id,
    name: options.name,
    description: options.description,
    expires,
    clientBinding: options.client !== undefined ? normalizeClientBindingOption(options.client) : undefined,
    permissions,
    replacePermissions: permissions.length > 0 || noPermissions,
  };
}

function parsePermissionInputsFromOptions(config: MvmtConfig, options: AddApiTokenOptions): ApiTokenPermissionInput[] {
  const inputs: ApiTokenPermissionInput[] = [];
  for (const scope of splitScopeValues(options.scope ?? [])) {
    inputs.push(parseScope(config, scope));
  }
  inputs.push(...(options.read ?? []).map((source) => ({ source, mode: 'read' as const })));
  inputs.push(...(options.write ?? []).map((source) => ({ source, mode: 'write' as const })));
  return inputs;
}

async function promptForApiTokenInput(
  config: MvmtConfig,
  id: string | undefined,
  options: AddApiTokenOptions,
): Promise<ApiTokenInput> {
  const tokenId = id ? normalizeApiTokenId(id) : await input({
    message: 'Token name:',
    validate: validateApiTokenId,
  });
  if (config.clients?.some((client) => client.id === tokenId)) {
    throw new Error(`API token ${tokenId} already exists. Use token edit or token rotate.`);
  }
  const name = options.name ?? await input({
    message: 'Display name:',
    default: tokenId,
    validate: (value) => value.trim().length > 0 ? true : 'Enter a display name',
  });
  const description = options.description ?? await input({
    message: 'Description (optional):',
    default: '',
  });
  const permissions = await promptForScopeInputs(config);
  const clientBinding = await promptForClientBinding();
  const expires = await promptForExpires();
  return { id: tokenId, name, description, expires, clientBinding, permissions };
}

async function promptForApiTokenEditInput(
  config: MvmtConfig,
  id: string,
  options: AddApiTokenOptions,
): Promise<ApiTokenInput | undefined> {
  const existing = findTokenClient(config, id);
  console.log(`Current settings for '${existing.id}':`);
  console.log(`  Scope: ${formatClientScope(config, existing)}`);
  console.log(`  Client: ${existing.clientBinding ?? '(any)'}`);
  console.log(`  Expires: ${formatTokenExpiry(existing.expiresAt)}`);
  console.log('');

  const change = await select<'scope' | 'client' | 'expiration' | 'cancel'>({
    message: 'What would you like to change?',
    choices: [
      { name: 'Scope', value: 'scope' },
      { name: 'Client binding', value: 'client' },
      { name: 'Expiration', value: 'expiration' },
      { name: 'Cancel', value: 'cancel' },
    ],
  });
  if (change === 'cancel') return undefined;
  if (change === 'scope') {
    return {
      id,
      name: options.name,
      description: options.description,
      permissions: await promptForScopeInputs(config),
    };
  }
  if (change === 'client') {
    return {
      id,
      name: options.name,
      description: options.description,
      clientBinding: await promptForClientBinding(existing.clientBinding),
      permissions: [],
    };
  }
  return {
    id,
    name: options.name,
    description: options.description,
    expires: await promptForExpires(existing.expiresAt),
    permissions: [],
  };
}

async function promptForScopeInputs(config: MvmtConfig): Promise<ApiTokenPermissionInput[]> {
  const access = await select<'specific' | 'all'>({
    message: 'What can this token access?',
    choices: [
      { name: 'Specific connectors only', value: 'specific' },
      { name: 'All connectors', value: 'all' },
    ],
    default: 'specific',
  });
  const mode = await select<ApiTokenPermissionMode>({
    message: 'Permission level?',
    choices: [
      { name: 'Read only', value: 'read' },
      { name: 'Read and write', value: 'write' },
    ],
    default: 'read',
  });
  if (access === 'all') return [{ source: 'all', mode }];
  return promptForPermissionInputs(config, mode);
}

async function promptForPermissionInputs(
  config: MvmtConfig,
  mode: ApiTokenPermissionMode,
): Promise<ApiTokenPermissionInput[]> {
  const sources = policySources(config);
  if (sources.length === 0) {
    throw new Error('No enabled connectors or mounts. Add a mount with `mvmt mounts add <name> <folder>` first.');
  }
  if (mode === 'write' && sources.every((source) => !source.writeAccess)) {
    throw new Error('No writable connectors or mounts. Enable write access first with `mvmt mounts edit <name> --write`.');
  }
  const inputs: ApiTokenPermissionInput[] = [];
  let addMore = true;
  while (addMore) {
    const source = await select<PolicySource>({
      message: 'Grant access to which connector?',
      choices: sources.map((candidate) => ({
        name: `${candidate.id} (${candidate.path}, ${candidate.writeAccess ? 'read/write' : 'read-only'})`,
        value: candidate,
        ...(mode === 'write' && !candidate.writeAccess
          ? { disabled: 'read-only; enable write access on the mount first' }
          : {}),
      })),
    });
    inputs.push({ source: source.id, mode });
    addMore = await confirm({ message: 'Add another connector permission?', default: false });
  }
  if (inputs.length === 0) throw new Error('API token needs at least one scope.');
  return inputs;
}

async function promptForClientBinding(current?: string): Promise<string | null> {
  const value = await input({
    message: 'Bind to a specific client identity? (optional)',
    default: current ?? '',
  });
  return normalizeClientBindingOption(value);
}

async function promptForExpires(currentExpiresAt?: string): Promise<string> {
  const choice = await select<'never' | '1h' | '24h' | '7d' | 'custom'>({
    message: 'Expires?',
    choices: [
      { name: 'Never', value: 'never' },
      { name: 'In 1 hour', value: '1h' },
      { name: 'In 24 hours', value: '24h' },
      { name: 'In 7 days', value: '7d' },
      { name: 'Custom', value: 'custom' },
    ],
    default: currentExpiresAt ? '7d' : 'never',
  });
  if (choice !== 'custom') return choice;
  return input({
    message: 'Expiration (examples: 30m, 7d, 30d, never):',
    default: '30d',
    validate: (value) => {
      try {
        parseTokenTtl(value);
        return true;
      } catch (err) {
        return err instanceof Error ? err.message : 'Invalid expiration';
      }
    },
  });
}

function applyPermissionInputs(
  config: MvmtConfig,
  currentPermissions: readonly PermissionConfig[],
  inputs: readonly ApiTokenPermissionInput[],
  options: { allowEmpty?: boolean } = {},
): PermissionConfig[] {
  if (inputs.length === 0) {
    if (options.allowEmpty) return [];
    throw new Error('API token needs at least one scope.');
  }
  const next = [...currentPermissions];
  for (const inputValue of inputs) {
    const permission = permissionForInput(config, inputValue);
    const existingIndex = next.findIndex((candidate) => candidate.path === permission.path);
    if (existingIndex >= 0) {
      next[existingIndex] = permission;
    } else {
      next.push(permission);
    }
  }
  return next;
}

function permissionForInput(config: MvmtConfig, inputValue: ApiTokenPermissionInput): PermissionConfig {
  const source = inputValue.source.trim();
  if (source === 'all' || source === '*') {
    return {
      path: '/**',
      actions: actionsForMode(inputValue.mode),
    };
  }

  const target = resolvePolicySource(config, source);
  if (inputValue.mode === 'write' && !target.writeAccess) {
    throw new Error(`${target.label} is read-only; API token cannot be granted write/remove.`);
  }
  return {
    path: `${target.path}/**`,
    actions: actionsForMode(inputValue.mode),
  };
}

function actionsForMode(mode: ApiTokenPermissionMode): PermissionConfig['actions'] {
  return mode === 'write' ? ['search', 'read', 'write'] : ['search', 'read'];
}

function parseScope(config: MvmtConfig, scope: string): ApiTokenPermissionInput {
  const trimmed = scope.trim();
  if (!trimmed) throw new Error('Scope cannot be empty.');
  const parts = trimmed.split(':');
  if (parts.length > 2) throw new Error(`Invalid scope: ${scope}`);
  const source = parts[0]?.trim();
  const permission = parts[1]?.trim() || (source === 'all' ? 'write' : 'read');
  if (!source) throw new Error(`Invalid scope: ${scope}`);
  const mode = permissionToMode(permission, scope);
  const inputValue = { source, mode };
  permissionForInput(config, inputValue);
  return inputValue;
}

function permissionToMode(permission: string, originalScope: string): ApiTokenPermissionMode {
  if (permission === 'read') return 'read';
  if (permission === 'write' || permission === '*') return 'write';
  throw new Error(`Invalid scope permission in ${originalScope}; use read, write, or *.`);
}

function resolvePolicySource(config: MvmtConfig, sourceRef: string): PolicySource {
  const normalizedRef = normalizePolicyPath(sourceRef);
  const source = policySources(config).find((candidate) => (
    candidate.id === sourceRef
    || candidate.path === normalizedRef
  ));
  if (!source) throw new Error(`Unknown enabled connector or mount: ${sourceRef}`);
  return source;
}

function policySources(config: MvmtConfig): PolicySource[] {
  const mounts = enabledMounts(config).map((mount) => ({
    id: mount.name,
    label: `Mount ${mount.path}`,
    path: normalizePolicyPath(mount.path),
    writeAccess: mount.writeAccess,
  }));
  const proxies = config.proxy
    .filter((proxy) => proxy.enabled !== false)
    .map((proxy) => {
      const sourceId = resolveProxySourceId(proxy);
      return {
        id: sourceId,
        label: `Connector ${sourceId}`,
        path: `/${sourceId}`,
        writeAccess: Boolean(proxy.writeAccess),
      };
    });
  return [...mounts, ...proxies];
}

function enabledMounts(config: MvmtConfig): LocalFolderMountConfig[] {
  return config.mounts.filter((mount) => mount.enabled !== false);
}

function findTokenClient(config: MvmtConfig, id: string): ClientConfig {
  const tokenId = normalizeApiTokenId(id);
  const existing = config.clients?.find((client) => client.id === tokenId);
  if (!existing || existing.auth.type !== 'token') throw new Error(`Unknown API token: ${tokenId}`);
  return existing;
}

function promptForApiTokenId(config: MvmtConfig, message: string): Promise<string> {
  const tokens = tokenClients(config);
  if (tokens.length === 0) throw new Error('No API tokens configured.');
  return select({
    message,
    choices: tokens.map((client) => ({
      name: `${client.id} (${client.name})`,
      value: client.id,
    })),
  });
}

function tokenClients(config: MvmtConfig): ClientConfig[] {
  return (config.clients ?? []).filter((client) => client.auth.type === 'token');
}

function toApiTokenSummary(
  config: MvmtConfig,
  client: ClientConfig,
  lastUsedAt?: string,
): Record<string, unknown> {
  return {
    name: client.id,
    scope: formatClientScope(config, client),
    client: client.clientBinding ?? null,
    createdAt: client.createdAt ?? null,
    lastUsedAt: lastUsedAt ?? client.lastUsedAt ?? null,
    expiresAt: client.expiresAt ?? null,
  };
}

export function printApiTokenSaved(configPath: string, result: ApiTokenUpdateResult): void {
  const action = result.created ? 'created' : result.plaintextToken ? 'rotated' : 'updated';
  console.log(chalk.green(`Token ${action}.`));
  console.log('');
  console.log(`  Name:    ${result.client.id}`);
  console.log(`  Scope:   ${formatClientScope(result.config, result.client)}`);
  console.log(`  Client:  ${result.client.clientBinding ?? '(any)'}`);
  console.log(`  Expires: ${formatTokenExpiry(result.client.expiresAt)}`);
  if (result.plaintextToken) {
    console.log(`  Token:   ${result.plaintextToken}`);
    console.log('');
    console.log(chalk.yellow('This is the only time the token will be shown. Store it now.'));
    console.log('');
    console.log(`  HTTP MCP endpoint${result.client.clientBinding ? ` (client: ${result.client.clientBinding})` : ''}:`);
    console.log('    URL:    http://127.0.0.1:4141/mcp');
    console.log(`    Header: Authorization: Bearer ${result.plaintextToken}`);
    console.log('');
    console.log(chalk.dim('Use these values in any HTTP MCP client. See docs/client-setup.md for client-specific setup.'));
    console.log(chalk.dim('For OAuth clients, paste this token into the mvmt approval page when asked.'));
  } else {
    console.log('');
    console.log(chalk.dim('Existing token value was not changed.'));
    if (result.credentialVersionChanged) {
      console.log(chalk.dim('Existing OAuth grants must reauthorize because this edit changed client binding.'));
      if (result.permissionsReplaced) {
        console.log(chalk.dim('Permission edit applies immediately to API tokens, and to OAuth grants after reauthorization.'));
      }
    } else if (result.permissionsReplaced) {
      console.log(chalk.dim('Permission edit applies to existing API tokens and OAuth grants on the next request.'));
    } else {
      console.log(chalk.dim('Existing OAuth grants keep working.'));
    }
  }
  console.log(chalk.dim(`Config: ${configPath}`));
}

function hasAddOptions(options: AddApiTokenOptions): boolean {
  return Boolean(
    (options.scope?.length ?? 0) > 0
    || (options.read?.length ?? 0) > 0
    || (options.write?.length ?? 0) > 0
    || options.name !== undefined
    || options.description !== undefined
    || options.expires !== undefined
    || options.ttl !== undefined
    || options.client !== undefined
  );
}

function hasEditOptions(options: EditApiTokenOptions): boolean {
  return hasAddOptions(options) || options.permissions === false;
}

function resolveExpiresAt(inputValue: ApiTokenInput): string | undefined {
  if (inputValue.expires !== undefined) {
    return parseTokenTtl(inputValue.expires, inputValue.now).expiresAt;
  }
  return undefined;
}

function normalizeApiTokenId(id: string): string {
  const trimmed = id.trim();
  const valid = validateApiTokenId(trimmed);
  if (valid !== true) throw new Error(String(valid));
  return trimmed;
}

function validateApiTokenId(value: string): true | string {
  const trimmed = value.trim();
  if (!trimmed) return 'Enter a token name';
  if (!API_TOKEN_ID_RE.test(trimmed)) {
    return 'Use lowercase letters, numbers, dash, or underscore';
  }
  return true;
}

function generateApiToken(): string {
  return `${TOKEN_PREFIX}${crypto.randomBytes(32).toString('base64url')}`;
}

function resolveApiTokensConfigPath(configPath?: string): string {
  return configPath ? resolveSetupPath(configPath) : getConfigPath();
}

function auditLogPathForConfig(configPath: string): string {
  return path.join(path.dirname(configPath), 'audit.log');
}

function recordTokenLifecycle(
  configPath: string,
  event: Extract<AuditEntry['event'], 'token.add' | 'token.edit' | 'token.rotate' | 'token.remove'>,
  config: MvmtConfig,
  client: ClientConfig,
): void {
  createAuditLogger(auditLogPathForConfig(configPath)).record({
    ts: new Date().toISOString(),
    event,
    connectorId: 'mvmt',
    tool: event,
    clientId: client.id,
    name: client.id,
    scope: formatClientScope(config, client),
    client: client.clientBinding ?? '(any)',
    expires: client.expiresAt ?? null,
    argKeys: ['name'],
    argPreview: JSON.stringify({ name: client.id }),
    isError: false,
    durationMs: 0,
  });
}

function loadTokenLastUsed(logPath: string): Map<string, string> {
  const lastUsed = new Map<string, string>();
  let raw: string;
  try {
    raw = fs.readFileSync(logPath, 'utf-8');
  } catch {
    return lastUsed;
  }
  for (const line of raw.split('\n')) {
    if (!line.trim()) continue;
    try {
      const entry = JSON.parse(line) as Partial<AuditEntry>;
      if (entry.event !== 'token.use') continue;
      const name = entry.name ?? entry.clientId;
      if (!name || !entry.ts) continue;
      if (!lastUsed.has(name) || Date.parse(entry.ts) > Date.parse(lastUsed.get(name)!)) {
        lastUsed.set(name, entry.ts);
      }
    } catch {
      // Ignore malformed historical audit rows.
    }
  }
  return lastUsed;
}

function formatClientScope(config: MvmtConfig, client: ClientConfig): string {
  if (client.permissions.length === 0) return '(none)';
  return client.permissions.map((permission) => permissionToScope(config, permission)).join(',');
}

function permissionToScope(config: MvmtConfig, permission: PermissionConfig): string {
  const mode = permission.actions.includes('write') ? 'write' : 'read';
  const normalizedPath = normalizePermissionPattern(permission.path);
  if (normalizedPath === '/**') return mode === 'write' ? 'all' : 'all:read';
  const base = normalizedPath.endsWith('/**') ? normalizedPath.slice(0, -3) : normalizedPath;
  const source = policySources(config).find((candidate) => candidate.path === base);
  const id = source?.id ?? base;
  return `${id}:${mode}`;
}

function normalizePermissionPattern(value: string): string {
  const trimmed = value.trim().replaceAll('\\', '/');
  if (trimmed === '/**') return trimmed;
  if (trimmed.endsWith('/**')) return `${normalizePolicyPath(trimmed.slice(0, -3))}/**`;
  return normalizePolicyPath(trimmed);
}

function normalizePolicyPath(value: string): string {
  const withSlash = value.startsWith('/') ? value : `/${value}`;
  const normalized = withSlash.replaceAll('\\', '/');
  let end = normalized.length;
  while (end > 1 && normalized[end - 1] === '/') end -= 1;
  return normalized.slice(0, end);
}

function normalizeClientBinding(value: string | null | undefined): string | undefined {
  if (value === null || value === undefined) return undefined;
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
}

function normalizeClientBindingOption(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed || trimmed === '(any)' || trimmed.toLowerCase() === 'any' || trimmed.toLowerCase() === 'none') {
    return null;
  }
  return trimmed;
}

function splitScopeValues(values: string[]): string[] {
  return values.flatMap((value) => value.split(',')).map((value) => value.trim()).filter(Boolean);
}

function formatPastTime(iso: string | undefined): string {
  return iso ? formatRelativeTime(Date.parse(iso), Date.now(), { pastSuffix: 'ago' }) : 'unknown';
}

function formatLastUsed(iso: string | undefined): string {
  return iso ? formatRelativeTime(Date.parse(iso), Date.now(), { pastSuffix: 'ago' }) : 'never';
}

function formatExpiryRelative(iso: string | undefined): string {
  if (!iso) return 'never';
  const target = Date.parse(iso);
  if (Number.isNaN(target)) return 'unknown';
  if (target <= Date.now()) return 'expired';
  return formatRelativeTime(target, Date.now(), { futurePrefix: 'in' });
}

function formatRelativeTime(
  target: number,
  now: number,
  labels: { futurePrefix?: string; pastSuffix?: string },
): string {
  if (Number.isNaN(target)) return 'unknown';
  const deltaMs = target - now;
  const absSeconds = Math.max(0, Math.round(Math.abs(deltaMs) / 1000));
  const units: Array<[seconds: number, singular: string]> = [
    [365 * 24 * 60 * 60, 'year'],
    [30 * 24 * 60 * 60, 'month'],
    [24 * 60 * 60, 'day'],
    [60 * 60, 'hour'],
    [60, 'minute'],
  ];
  const [unitSeconds, unit] = units.find(([seconds]) => absSeconds >= seconds) ?? [1, 'second'];
  const count = Math.max(1, Math.round(absSeconds / unitSeconds));
  const value = `${count} ${unit}${count === 1 ? '' : 's'}`;
  if (deltaMs >= 0) return labels.futurePrefix ? `${labels.futurePrefix} ${value}` : value;
  return labels.pastSuffix ? `${value} ${labels.pastSuffix}` : value;
}

function printTable(rows: Array<Record<string, string>>, headers: string[]): void {
  const widths = headers.map((header) => Math.max(
    header.length,
    ...rows.map((row) => row[header]?.length ?? 0),
  ));
  console.log(headers.map((header, index) => chalk.bold(header.padEnd(widths[index]))).join('  '));
  for (const row of rows) {
    console.log(headers.map((header, index) => (row[header] ?? '').padEnd(widths[index])).join('  '));
  }
}

async function confirmDestructive(message: string): Promise<boolean> {
  if (!process.stdin.isTTY || !process.stdout.isTTY) {
    throw new Error(`${message} Re-run with --yes to confirm non-interactively.`);
  }
  return confirm({
    message,
    default: false,
  });
}

function printApiTokenCommandError(err: unknown): void {
  console.error(err instanceof Error ? err.message : 'API token command failed.');
  process.exitCode = 1;
}
