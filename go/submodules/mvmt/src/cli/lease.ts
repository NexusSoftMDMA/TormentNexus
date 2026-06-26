import chalk from 'chalk';
import fsp from 'fs/promises';
import path from 'path';
import { configExists, getConfigPath, loadConfig, parseConfig, saveConfig } from '../config/loader.js';
import { MvmtConfig } from '../config/schema.js';
import { resolveSetupPath } from '../connectors/setup-paths.js';
import { MountRegistry, normalizeVirtualPath } from '../context/mount-registry.js';
import { addMountToConfig, MountInput } from './mounts.js';
import {
  addLeaseResources,
  createLease,
  DEFAULT_LEASE_TTL,
  defaultLeasesPath,
  findLease,
  LeasePermission,
  LeaseResource,
  leaseResources,
  leaseUnavailableReason,
  listLeases,
  revokeLease,
  setLeasePublished,
} from '../core/leases/store.js';
import { leaseSecretsPathForLeaseStore, removeLeaseSecret, saveLeaseSecret } from '../core/leases/secrets.js';
import { parseTokenTtl } from '../utils/token-ttl.js';
import { normalizeTunnelBaseUrl } from '../utils/tunnel.js';

export interface LeaseCommandOptions {
  config?: string;
  json?: boolean;
  leaseStorePath?: string;
}

export interface CreateLeaseOptions extends LeaseCommandOptions {
  label?: string;
  expires?: string;
  ttl?: string;
  mode?: string;
  upload?: boolean;
  downloads?: string;
}

export async function listFolderLeases(options: LeaseCommandOptions = {}): Promise<void> {
  const leases = listLeases(resolveLeaseStorePath(options));
  if (options.json) {
    console.log(JSON.stringify({ leases }, null, 2));
    return;
  }

  console.log(chalk.bold('Leases'));
  if (leases.length === 0) {
    console.log(`  ${chalk.dim('none')}`);
    return;
  }
  for (const lease of leases) {
    const reason = leaseUnavailableReason(lease);
    const state = reason ? chalk.yellow(reason) : chalk.green('active');
    const activity = lease.permissions.includes('upload')
      ? `uploads ${lease.uploadCount ?? 0}`
      : `downloads ${lease.downloadCount ?? 0}`;
    const limit = lease.maxDownloads === undefined ? '' : ` limit ${lease.maxDownloads}`;
    const resourceText = leaseResources(lease).map((resource) => resource.path).join(',');
    console.log(`  ${lease.id.padEnd(16)} ${state}  ${lease.label}  ${resourceText}  expires ${lease.expiresAt ?? 'never'}  ${activity}${limit}`);
  }
}

export async function createFolderLease(inputPath: string | string[] | undefined, options: CreateLeaseOptions = {}): Promise<void> {
  const inputPaths = Array.isArray(inputPath) ? inputPath : inputPath ? [inputPath] : [];
  if (inputPaths.length === 0) throw new Error('Path is required. Example: mvmt lease create ~/Documents/Taxes --label "Sarah - tax docs"');
  const label = options.label?.trim();
  if (!label) throw new Error('Label is required. Example: --label "Sarah - tax docs"');

  const configPath = options.config ? resolveSetupPath(options.config) : getConfigPath();
  const config = configExists(configPath) ? loadConfig(configPath) : parseConfig({ version: 1 });
  const permissions = permissionsFromOptions(options);
  const maxDownloads = maxDownloadsFromOptions(options);
  if (maxDownloads !== undefined && !permissions.includes('read')) {
    throw new Error('Download limits require a read or two-way lease.');
  }
  const prepared = await prepareLeasePaths(config, inputPaths, label, permissions);
  if (prepared.config !== config) await saveConfig(configPath, prepared.config);

  const ttl = parseTokenTtl(options.expires ?? options.ttl ?? DEFAULT_LEASE_TTL);
  const leaseStorePath = resolveLeaseStorePath(options);
  const created = createLease(leaseStorePath, {
    label,
    path: prepared.resources[0]!.sourcePath,
    resources: prepared.resources,
    expiresAt: ttl.expiresAt,
    permissions,
    ...(maxDownloads === undefined ? {} : { maxDownloads }),
  });
  saveLeaseSecret(leaseSecretsPathForLeaseStore(leaseStorePath), created.record.id, created.token);
  const url = leaseUrl(prepared.config, created.record.id, created.token);

  if (options.json) {
    console.log(JSON.stringify({ lease: { ...created.record, url } }, null, 2));
    return;
  }

  console.log(chalk.green('Lease created'));
  console.log(`  Label: ${created.record.label}`);
  console.log(`  Paths: ${leaseResources(created.record).map((resource) => resource.path).join(', ')}`);
  console.log(`  Mode: ${formatPermissions(created.record.permissions)}`);
  console.log(`  Expires: ${created.record.expiresAt ?? 'never'}${options.expires || options.ttl ? '' : ` (${DEFAULT_LEASE_TTL} default)`}`);
  if (created.record.maxDownloads !== undefined) console.log(`  Downloads: ${created.record.maxDownloads}`);
  for (const mount of prepared.addedMounts) {
    const access = mount.writeAccess ? 'upload-capable' : 'read-only';
    console.log(chalk.dim(`  Added ${access} mount: ${mount.name} -> ${mount.root}`));
  }
  console.log(`  URL: ${url}`);
}

export async function revokeFolderLease(id: string | undefined, options: LeaseCommandOptions = {}): Promise<void> {
  if (!id) throw new Error('Lease id is required.');
  const leaseStorePath = resolveLeaseStorePath(options);
  if (!revokeLease(leaseStorePath, id)) throw new Error(`Unknown lease: ${id}`);
  removeLeaseSecret(leaseSecretsPathForLeaseStore(leaseStorePath), id);
  console.log(chalk.green(`Lease ${id} revoked`));
}

export async function setFolderLeasePublished(
  id: string | undefined,
  published: boolean,
  options: LeaseCommandOptions = {},
): Promise<void> {
  if (!id) throw new Error('Lease id is required.');
  const leaseStorePath = resolveLeaseStorePath(options);
  const updated = setLeasePublished(leaseStorePath, id, published);
  if (!updated) throw new Error(`Unknown lease: ${id}`);
  if (published) {
    console.log(chalk.green(`Lease ${id} published - it can now be reached through a public tunnel.`));
  } else {
    console.log(chalk.green(`Lease ${id} unpublished - apps on this machine keep access; public tunnels are blocked.`));
  }
}

export async function addPathsToLease(
  id: string | undefined,
  inputPath: string | string[] | undefined,
  options: LeaseCommandOptions = {},
): Promise<void> {
  const leaseId = id?.trim();
  if (!leaseId) throw new Error('Lease id is required.');
  const inputPaths = Array.isArray(inputPath) ? inputPath : inputPath ? [inputPath] : [];
  if (inputPaths.length === 0) throw new Error('Path is required. Example: mvmt lease add-path <lease-id> ~/Documents/Receipts');

  const storePath = resolveLeaseStorePath(options);
  const lease = findLease(storePath, leaseId);
  if (!lease) throw new Error(`Unknown lease: ${leaseId}`);
  const unavailable = leaseUnavailableReason(lease);
  if (unavailable) throw new Error(`Lease ${leaseId} is ${unavailable}; create a new lease instead.`);
  if (lease.permissions.includes('upload')) throw new Error('Upload leases currently support one folder. Create a new upload lease instead.');

  const configPath = options.config ? resolveSetupPath(options.config) : getConfigPath();
  const config = configExists(configPath) ? loadConfig(configPath) : parseConfig({ version: 1 });
  const prepared = await prepareLeasePaths(config, inputPaths, lease.label, lease.permissions);
  if (prepared.config !== config) await saveConfig(configPath, prepared.config);
  const before = new Set(leaseResources(lease).map((resource) => `${resource.sourcePath}:${resource.type}`));
  const added = prepared.resources.filter((resource) => !before.has(`${resource.sourcePath}:${resource.type}`));
  const updated = addLeaseResources(storePath, leaseId, prepared.resources);
  if (!updated) throw new Error(`Unknown lease: ${leaseId}`);

  if (options.json) {
    console.log(JSON.stringify({ lease: updated, added }, null, 2));
    return;
  }

  console.log(chalk.green('Lease updated'));
  if (added.length === 0) {
    console.log(chalk.dim('  No new paths added.'));
  } else {
    console.log(`  Added paths: ${added.map((resource) => resource.path).join(', ')}`);
  }
  console.log(`  Paths: ${leaseResources(updated).map((resource) => resource.path).join(', ')}`);
  console.log(chalk.dim('  Existing lease token and URL now include these paths.'));
  for (const mount of prepared.addedMounts) {
    const access = mount.writeAccess ? 'upload-capable' : 'read-only';
    console.log(chalk.dim(`  Added ${access} mount: ${mount.name} -> ${mount.root}`));
  }
}

async function prepareLeasePaths(
  config: MvmtConfig,
  inputPaths: string[],
  label: string,
  permissions: LeasePermission[],
): Promise<{ config: MvmtConfig; resources: LeaseResource[]; addedMounts: MountInput[] }> {
  if (permissions.includes('upload') && inputPaths.length > 1) {
    throw new Error('Upload leases currently support one folder.');
  }
  let nextConfig = config;
  const resources: LeaseResource[] = [];
  const addedMounts: MountInput[] = [];
  for (const inputPath of inputPaths) {
    const prepared = await prepareLeasePath(nextConfig, inputPath, label, permissions);
    nextConfig = prepared.config;
    resources.push(prepared.resource);
    if (prepared.addedMount) addedMounts.push(prepared.addedMount);
  }
  return { config: nextConfig, resources: uniqueLeaseResources(resources), addedMounts };
}

async function prepareLeasePath(
  config: MvmtConfig,
  inputPath: string,
  label: string,
  permissions: LeasePermission[],
): Promise<{ config: MvmtConfig; resource: LeaseResource; addedMount?: MountInput }> {
  const resolvedLocalPath = resolveSetupPath(inputPath);
  const localStat = await statIfExists(resolvedLocalPath);
  if (localStat?.isDirectory()) return prepareLocalFolderLeasePath(config, resolvedLocalPath, label, permissions);
  if (localStat?.isFile()) return prepareLocalFileLeasePath(config, resolvedLocalPath, label, permissions);
  if (inputPath.trim().startsWith('/')) {
    const leasePath = normalizeVirtualPath(inputPath);
    const resolved = new MountRegistry(config.mounts).resolve(leasePath);
    const stat = await fsp.stat(resolved.realPath);
    if (!stat.isDirectory() && !stat.isFile()) throw new Error(`${leasePath} is not a mounted file or folder`);
    if ((permissions.includes('upload') || permissions.includes('write')) && !resolved.mount.config.writeAccess) {
      throw new Error(`${leasePath} is read-only; create the lease from a local path so mvmt can add a writable mount.`);
    }
    if (permissions.includes('upload') && !stat.isDirectory()) {
      throw new Error('Upload leases require a folder.');
    }
    return {
      config,
      resource: {
        path: resourcePathForVirtualPath(leasePath),
        sourcePath: leasePath,
        type: stat.isFile() ? 'file' : 'folder',
      },
    };
  }
  throw new Error(`Path not found: ${resolvedLocalPath}`);
}

async function prepareLocalFolderLeasePath(
  config: MvmtConfig,
  root: string,
  label: string,
  permissions: LeasePermission[],
): Promise<{ config: MvmtConfig; resource: LeaseResource; addedMount?: MountInput }> {
  const realRoot = await fsp.realpath(root);
  const needsWritableMount = permissions.includes('upload') || permissions.includes('write');
  for (const mount of config.mounts) {
    try {
      const mountRoot = await fsp.realpath(resolveSetupPath(mount.root));
      if (mount.enabled !== false && mountRoot === realRoot && (!needsWritableMount || mount.writeAccess)) {
        return { config, resource: { path: mount.path, sourcePath: mount.path, type: 'folder' } };
      }
    } catch {
      // Ignore stale mount roots while creating a lease for a valid folder.
    }
  }

  const prefix = needsWritableMount ? 'lease-write' : 'lease';
  const name = uniqueMountName(config, `${prefix}-${slugFromPath(realRoot)}`);
  const mountPath = uniqueMountPath(config, resourcePathForLocalPath(realRoot));
  const mount: MountInput = {
    name,
    root,
    path: mountPath,
    writeAccess: needsWritableMount,
    description: `Folder lease source: ${label}`,
    guidance: '',
    enabled: true,
  };
  return { config: addMountToConfig(config, mount), resource: { path: mountPath, sourcePath: mountPath, type: 'folder' }, addedMount: mount };
}

async function prepareLocalFileLeasePath(
  config: MvmtConfig,
  root: string,
  label: string,
  permissions: LeasePermission[],
): Promise<{ config: MvmtConfig; resource: LeaseResource; addedMount?: MountInput }> {
  if (permissions.includes('upload')) throw new Error('Upload leases require a folder.');
  const realRoot = await fsp.realpath(root);
  const needsWritableMount = permissions.includes('write');
  for (const mount of config.mounts) {
    try {
      const mountRoot = await fsp.realpath(resolveSetupPath(mount.root));
      if (mount.enabled !== false && mountRoot === realRoot && (!needsWritableMount || mount.writeAccess)) {
        return { config, resource: { path: mount.path, sourcePath: mount.path, type: 'file' } };
      }
    } catch {
      // Ignore stale mount roots while creating a lease for a valid file.
    }
  }

  const name = uniqueMountName(config, `lease-file-${slugFromPath(realRoot)}`);
  const mountPath = uniqueMountPath(config, resourcePathForLocalPath(realRoot));
  const mount: MountInput = {
    name,
    root,
    path: mountPath,
    writeAccess: needsWritableMount,
    description: `Lease source: ${label}`,
    guidance: '',
    enabled: true,
  };
  return { config: addMountToConfig(config, mount), resource: { path: mountPath, sourcePath: mountPath, type: 'file' }, addedMount: mount };
}

function permissionsFromOptions(options: CreateLeaseOptions): LeasePermission[] {
  const mode = options.upload ? 'upload' : (options.mode ?? 'read').trim().toLowerCase();
  if (mode === 'read' || mode === 'download') return ['read'];
  if (mode === 'write' || mode === 'writable') return ['read', 'write'];
  if (mode === 'upload') return ['upload'];
  if (mode === 'two-way' || mode === 'two_way' || mode === 'read+upload') return ['read', 'upload'];
  throw new Error('Invalid lease mode. Use --mode read, --mode upload, --mode two-way, or --mode write.');
}

function maxDownloadsFromOptions(options: CreateLeaseOptions): number | undefined {
  if (options.downloads === undefined) return undefined;
  const trimmed = options.downloads.trim();
  if (trimmed === '-1') return undefined;
  if (!/^[1-9]\d*$/.test(trimmed)) throw new Error('Invalid download limit. Use --downloads -1 for unlimited, or a positive integer.');
  const limit = Number.parseInt(trimmed, 10);
  if (!Number.isSafeInteger(limit)) throw new Error('Invalid download limit. Use a smaller positive integer.');
  return limit;
}

function formatPermissions(permissions: readonly LeasePermission[]): string {
  if (permissions.length === 1 && permissions[0] === 'read') return 'browse/download';
  if (permissions.includes('read') && permissions.includes('write')) return 'browse/download/write';
  if (permissions.includes('read') && permissions.includes('upload')) return 'browse/download/upload';
  if (permissions.length === 1 && permissions[0] === 'upload') return 'upload only';
  return permissions.join(', ');
}

async function statIfExists(inputPath: string): Promise<Awaited<ReturnType<typeof fsp.stat>> | undefined> {
  try {
    return await fsp.stat(inputPath);
  } catch {
    return undefined;
  }
}

function resolveLeaseStorePath(options: LeaseCommandOptions): string {
  return options.leaseStorePath ?? defaultLeasesPath();
}

function leaseUrl(config: MvmtConfig, id: string, token: string): string {
  const base = config.server.access === 'tunnel' && config.server.tunnel?.url
    ? normalizeTunnelBaseUrl(config.server.tunnel.url)
    : `http://127.0.0.1:${config.server.port}`;
  const url = new URL(`/lease/${id}`, base);
  url.searchParams.set('token', token);
  return url.toString();
}

function uniqueMountName(config: MvmtConfig, baseName: string): string {
  if (!config.mounts.some((mount) => mount.name === baseName)) return baseName;
  for (let index = 2; ; index += 1) {
    const candidate = `${baseName}-${index}`;
    if (!config.mounts.some((mount) => mount.name === candidate)) return candidate;
  }
}

function uniqueMountPath(config: MvmtConfig, basePath: string): string {
  if (!config.mounts.some((mount) => mount.path === basePath)) return basePath;
  for (let index = 2; ; index += 1) {
    const candidate = `${basePath}-${index}`;
    if (!config.mounts.some((mount) => mount.path === candidate)) return candidate;
  }
}

function slugFromPath(inputPath: string): string {
  const leaf = path.basename(inputPath) || 'folder';
  return leaf
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '') || 'folder';
}

function resourcePathForLocalPath(inputPath: string): string {
  const leaf = path.basename(inputPath) || 'folder';
  const segment = leaf
    .replace(/[^a-zA-Z0-9._-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '') || 'folder';
  return `/${segment}`;
}

function resourcePathForVirtualPath(inputPath: string): string {
  const relative = inputPath.split('/').filter(Boolean).join('-');
  return `/${relative || 'resource'}`;
}

function uniqueLeaseResources(resources: LeaseResource[]): LeaseResource[] {
  const seen = new Set<string>();
  const unique: LeaseResource[] = [];
  for (const resource of resources) {
    const key = `${resource.sourcePath}:${resource.type}`;
    if (seen.has(key)) continue;
    seen.add(key);
    unique.push(resource);
  }
  return unique;
}
