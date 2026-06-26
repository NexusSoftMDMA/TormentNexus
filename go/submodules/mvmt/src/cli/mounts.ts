import { confirm, input, select } from '@inquirer/prompts';
import chalk from 'chalk';
import fsp from 'fs/promises';
import { configExists, getConfigPath, loadConfig, parseConfig, saveConfig } from '../config/loader.js';
import {
  DEFAULT_MOUNT_EXCLUDE_PATTERNS,
  DEFAULT_MOUNT_PROTECT_PATTERNS,
  LocalFolderMountConfig,
  LocalFolderMountSchema,
  MvmtConfig,
} from '../config/schema.js';
import { resolveSetupPath } from '../connectors/setup-paths.js';
import { normalizePathSeparators, stripTrailingSlashes } from '../context/mount-registry.js';

export interface MountCommandOptions {
  config?: string;
  json?: boolean;
}

export interface AddMountOptions extends MountCommandOptions {
  mountPath?: string;
  write?: boolean;
  readOnly?: boolean;
  description?: string;
  guidance?: string;
  exclude?: string[];
  protect?: string[];
  disabled?: boolean;
}

export interface EditMountOptions extends MountCommandOptions {
  root?: string;
  mountPath?: string;
  write?: boolean;
  readOnly?: boolean;
  description?: string;
  guidance?: string;
  exclude?: string[];
  protect?: string[];
  enable?: boolean;
  disable?: boolean;
}

export interface RemoveMountOptions extends MountCommandOptions {
  yes?: boolean;
}

export interface MountInput {
  name: string;
  root: string;
  path?: string;
  writeAccess: boolean;
  description?: string;
  guidance?: string;
  exclude?: string[];
  protect?: string[];
  enabled?: boolean;
}

export const MOUNT_LOAD_NOTICE = 'Running mvmt loads mount changes on the next request. Run `mvmt reindex` for search.';
export const MOUNT_UNLOAD_NOTICE = 'Running mvmt unloads mount changes on the next request. Run `mvmt reindex` for search.';

export async function listMounts(options: MountCommandOptions = {}): Promise<void> {
  const config = loadConfig(resolveMountsConfigPath(options.config));
  printMounts(config, { json: Boolean(options.json) });
}

export async function addMount(name: string | undefined, root: string | undefined, options: AddMountOptions = {}): Promise<void> {
  assertWriteFlags(options);
  const configPath = resolveMountsConfigPath(options.config);
  const config = configExists(configPath) ? loadConfig(configPath) : parseConfig({ version: 1 });
  const inputValue = name && root
    ? mountInputFromOptions(name, root, options)
    : await promptForMountInput(config);
  if (!await validateMountRootForCommand(inputValue.root)) return;

  const nextConfig = addMountToConfig(config, inputValue);
  await saveConfig(configPath, nextConfig);
  console.log(chalk.green(`Mount ${inputValue.name} saved to ${configPath}`));
  printMountBasePermission(inputValue.writeAccess);
  console.log(chalk.dim(MOUNT_LOAD_NOTICE));
}

export async function editMount(name: string | undefined, options: EditMountOptions = {}): Promise<void> {
  assertWriteFlags(options);
  assertEnableFlags(options);
  const configPath = resolveMountsConfigPath(options.config);
  const config = loadConfig(configPath);
  const mountName = name ?? await promptForMountName(config, 'Edit which mount?');
  const patch = hasEditOptions(options)
    ? mountPatchFromOptions(options)
    : await promptForMountPatch(config, mountName);
  if (patch.root && !await validateMountRootForCommand(patch.root)) return;

  const nextConfig = editMountInConfig(config, mountName, patch);
  const updatedMount = findMount(nextConfig, mountName);
  await saveConfig(configPath, nextConfig);
  console.log(chalk.green(`Mount ${mountName} updated in ${configPath}`));
  printMountBasePermission(updatedMount.writeAccess);
  console.log(chalk.dim(MOUNT_LOAD_NOTICE));
}

export async function removeMount(name: string | undefined, options: RemoveMountOptions = {}): Promise<void> {
  const configPath = resolveMountsConfigPath(options.config);
  const config = loadConfig(configPath);
  const mountName = name ?? await promptForMountName(config, 'Remove which mount?');
  const ok = options.yes ? true : await confirm({
    message: `Remove mount ${mountName}? Client permissions that reference it will become invalid.`,
    default: false,
  });
  if (!ok) {
    console.log(chalk.yellow('Mount config unchanged.'));
    return;
  }

  const nextConfig = removeMountFromConfig(config, mountName);
  await saveConfig(configPath, nextConfig);
  console.log(chalk.green(`Mount ${mountName} removed from ${configPath}`));
  console.log(chalk.dim(MOUNT_UNLOAD_NOTICE));
}

export function printMounts(config: MvmtConfig, options: { json?: boolean } = {}): void {
  if (options.json) {
    console.log(JSON.stringify({ mounts: config.mounts.map(toMountSummary) }, null, 2));
    return;
  }

  console.log(chalk.bold('Mounts'));
  if (config.mounts.length === 0) {
    console.log(`  ${chalk.dim('none')}`);
    return;
  }

  for (const mount of config.mounts) {
    const state = mount.enabled === false ? chalk.dim('disabled') : chalk.green('enabled');
    const access = mount.writeAccess ? chalk.yellow('read/write') : chalk.dim('read-only');
    console.log(`  ${mount.name.padEnd(16)} ${state}  ${access}  ${mount.path} -> ${mount.root}`);
    if (mount.description) console.log(`    description: ${mount.description}`);
    if (mount.guidance) console.log(`    guidance: ${mount.guidance}`);
    if (mount.exclude.length > 0) console.log(`    exclude: ${mount.exclude.join(', ')}`);
    if (mount.protect.length > 0) console.log(`    protect: ${mount.protect.join(', ')}`);
  }
}

function toMountSummary(mount: LocalFolderMountConfig): Record<string, unknown> {
  return {
    name: mount.name,
    path: mount.path,
    root: mount.root,
    enabled: mount.enabled !== false,
    writeAccess: mount.writeAccess,
    description: mount.description,
    guidance: mount.guidance,
    exclude: mount.exclude,
    protect: mount.protect,
  };
}

export function printMountBasePermission(writeAccess: boolean): void {
  const access = writeAccess ? 'read/write' : 'read-only';
  const message = `Base permission: ${access}. Tokens cannot exceed it.`;
  console.log(writeAccess ? chalk.yellow(message) : chalk.dim(message));
}

export function addMountToConfig(config: MvmtConfig, mountInput: MountInput): MvmtConfig {
  if (config.mounts.some((mount) => mount.name === mountInput.name)) {
    throw new Error(`Mount already exists: ${mountInput.name}`);
  }
  const mount = normalizeMountInput(mountInput);
  if (config.mounts.some((candidate) => candidate.path === mount.path)) {
    throw new Error(`Mount path already exists: ${mount.path}`);
  }
  return { ...config, mounts: [...config.mounts, mount] };
}

export function editMountInConfig(config: MvmtConfig, name: string, patch: Partial<MountInput>): MvmtConfig {
  const index = config.mounts.findIndex((mount) => mount.name === name);
  if (index < 0) throw new Error(`Unknown mount: ${name}`);
  const current = config.mounts[index];
  const updated = normalizeMountInput({
    name: current.name,
    root: patch.root ?? current.root,
    path: patch.path ?? current.path,
    writeAccess: patch.writeAccess ?? current.writeAccess,
    description: patch.description ?? current.description,
    guidance: patch.guidance ?? current.guidance,
    exclude: patch.exclude ?? current.exclude,
    protect: patch.protect ?? current.protect,
    enabled: patch.enabled ?? current.enabled,
  });
  if (config.mounts.some((mount, mountIndex) => mountIndex !== index && mount.path === updated.path)) {
    throw new Error(`Mount path already exists: ${updated.path}`);
  }
  return {
    ...config,
    mounts: [
      ...config.mounts.slice(0, index),
      updated,
      ...config.mounts.slice(index + 1),
    ],
  };
}

export function findMount(config: MvmtConfig, name: string): LocalFolderMountConfig {
  const mount = config.mounts.find((candidate) => candidate.name === name);
  if (!mount) throw new Error(`Unknown mount: ${name}`);
  return mount;
}

export function removeMountFromConfig(config: MvmtConfig, name: string): MvmtConfig {
  if (!config.mounts.some((mount) => mount.name === name)) {
    throw new Error(`Unknown mount: ${name}`);
  }
  assertMountNotReferenced(config, name);
  return { ...config, mounts: config.mounts.filter((mount) => mount.name !== name) };
}

export interface MountPromptResult {
  config: MvmtConfig;
  name: string;
}

export async function promptAndAddMount(config: MvmtConfig): Promise<MountPromptResult | undefined> {
  const mount = await promptForMountInput(config);
  return { config: addMountToConfig(config, mount), name: mount.name };
}

export async function promptAndEditMount(config: MvmtConfig): Promise<MountPromptResult | undefined> {
  if (config.mounts.length === 0) {
    console.log(chalk.yellow('No mounts configured.'));
    return undefined;
  }
  const name = await promptForMountName(config, 'Edit which mount?');
  const patch = await promptForMountPatch(config, name);
  return { config: editMountInConfig(config, name, patch), name };
}

export async function promptAndRemoveMount(config: MvmtConfig): Promise<MvmtConfig | undefined> {
  if (config.mounts.length === 0) {
    console.log(chalk.yellow('No mounts configured.'));
    return undefined;
  }
  const name = await promptForMountName(config, 'Remove which mount?');
  const ok = await confirm({ message: `Remove mount ${name}?`, default: false });
  return ok ? removeMountFromConfig(config, name) : undefined;
}

function resolveMountsConfigPath(configPath?: string): string {
  return configPath ? resolveSetupPath(configPath) : getConfigPath();
}

function normalizeMountInput(inputValue: MountInput): LocalFolderMountConfig {
  return LocalFolderMountSchema.parse({
    name: inputValue.name,
    type: 'local_folder',
    path: normalizeMountPath(inputValue.path ?? `/${inputValue.name}`),
    root: resolveSetupPath(inputValue.root),
    ...(inputValue.description !== undefined ? { description: inputValue.description } : {}),
    ...(inputValue.guidance !== undefined ? { guidance: inputValue.guidance } : {}),
    ...(inputValue.exclude ? { exclude: inputValue.exclude } : {}),
    ...(inputValue.protect ? { protect: inputValue.protect } : {}),
    writeAccess: inputValue.writeAccess,
    enabled: inputValue.enabled ?? true,
  });
}

function mountInputFromOptions(name: string, root: string, options: AddMountOptions): MountInput {
  return {
    name,
    root,
    ...(options.mountPath !== undefined ? { path: options.mountPath } : {}),
    writeAccess: Boolean(options.write),
    ...(options.description !== undefined ? { description: options.description } : {}),
    ...(options.guidance !== undefined ? { guidance: options.guidance } : {}),
    ...(options.exclude ? { exclude: options.exclude } : {}),
    ...(options.protect ? { protect: options.protect } : {}),
    enabled: !options.disabled,
  };
}

function mountPatchFromOptions(options: EditMountOptions): Partial<MountInput> {
  return {
    ...(options.root ? { root: options.root } : {}),
    ...(options.mountPath !== undefined ? { path: options.mountPath } : {}),
    ...(options.write ? { writeAccess: true } : {}),
    ...(options.readOnly ? { writeAccess: false } : {}),
    ...(options.description !== undefined ? { description: options.description } : {}),
    ...(options.guidance !== undefined ? { guidance: options.guidance } : {}),
    ...(options.exclude ? { exclude: options.exclude } : {}),
    ...(options.protect ? { protect: options.protect } : {}),
    ...(options.enable ? { enabled: true } : {}),
    ...(options.disable ? { enabled: false } : {}),
  };
}

async function promptForMountInput(config: MvmtConfig): Promise<MountInput> {
  const root = await promptForExistingMountRoot();
  const defaultName = uniqueMountName(config, mountNameFromRoot(root));
  const name = await input({
    message: 'Mount id (optional stable id):',
    default: defaultName,
    validate: (value) => {
      const trimmed = value.trim();
      if (!/^[a-z0-9][a-z0-9_-]*$/.test(trimmed)) return 'Use lowercase letters, numbers, dash, or underscore';
      if (config.mounts.some((mount) => mount.name === trimmed)) return `Mount already exists: ${trimmed}`;
      return true;
    },
  });
  const mountPath = await input({
    message: 'Virtual path clients will use:',
    default: `/${name.trim()}`,
    validate: (value) => validateMountPathInput(value, config),
  });
  const writeAccess = await confirm({
    message: 'Allow write/remove? This is the base permission.',
    default: false,
  });
  const description = await input({
    message: 'Short description shown by list("/") (optional):',
    default: '',
  });
  const guidance = await input({
    message: 'Guidance shown by list("/") (optional):',
    default: '',
  });
  const customizePatterns = await confirm({
    message: 'Customize hidden/protected file patterns?',
    default: false,
  });
  const exclude = customizePatterns
    ? splitPatterns(await input({
      message: 'Hidden patterns, comma-separated:',
      default: DEFAULT_MOUNT_EXCLUDE_PATTERNS.join(', '),
    }))
    : [...DEFAULT_MOUNT_EXCLUDE_PATTERNS];
  const protect = customizePatterns
    ? splitPatterns(await input({
      message: 'Protected write/remove patterns, comma-separated:',
      default: DEFAULT_MOUNT_PROTECT_PATTERNS.join(', '),
    }))
    : [...DEFAULT_MOUNT_PROTECT_PATTERNS];
  return {
    name: name.trim(),
    root: root.trim(),
    path: mountPath.trim(),
    writeAccess,
    description: description.trim(),
    guidance: guidance.trim(),
    exclude,
    protect,
    enabled: true,
  };
}

async function promptForMountPatch(config: MvmtConfig, name: string): Promise<Partial<MountInput>> {
  const current = config.mounts.find((mount) => mount.name === name);
  if (!current) throw new Error(`Unknown mount: ${name}`);
  const root = await promptForExistingMountRoot('File or folder on this computer:', {
    defaultValue: current.root,
  });
  const mountPath = await input({
    message: 'Virtual path clients will use:',
    default: current.path,
    validate: (value) => validateMountPathInput(value, config, name),
  });
  const writeAccess = await confirm({
    message: 'Allow write/remove? This is the base permission.',
    default: current.writeAccess,
  });
  const enabled = await confirm({
    message: 'Enable this mount?',
    default: current.enabled,
  });
  const description = await input({
    message: 'Short description shown by list("/"):',
    default: current.description,
  });
  const guidance = await input({
    message: 'Guidance shown by list("/"):',
    default: current.guidance,
  });
  const exclude = splitPatterns(await input({
    message: 'Exclude patterns (comma-separated):',
    default: current.exclude.join(', '),
  }));
  const protect = splitPatterns(await input({
    message: 'Protected write/remove patterns (comma-separated):',
    default: current.protect.join(', '),
  }));
  return { root: root.trim(), path: mountPath.trim(), writeAccess, enabled, description: description.trim(), guidance: guidance.trim(), exclude, protect };
}

async function promptForMountName(config: MvmtConfig, message: string): Promise<string> {
  if (config.mounts.length === 0) throw new Error('No mounts configured.');
  return select({
    message,
    choices: config.mounts.map((mount) => ({
      name: `${mount.name} (${mount.path} -> ${mount.root})`,
      value: mount.name,
    })),
  });
}

function splitPatterns(value: string): string[] {
  return value.split(',').map((item) => item.trim()).filter(Boolean);
}

async function validateMountRootForCommand(root: string): Promise<boolean> {
  const valid = await validateExistingMountRootPath(root);
  if (valid === true) return true;
  console.error(valid);
  process.exitCode = 1;
  return false;
}

async function promptForExistingMountRoot(
  message = 'File or folder on this computer:',
  options: { defaultValue?: string } = {},
): Promise<string> {
  return input({
    message,
    ...(options.defaultValue ? { default: options.defaultValue } : {}),
    validate: async (value) => validateExistingMountRootPath(value),
  });
}

async function validateExistingMountRootPath(value: string): Promise<true | string> {
  const resolved = resolveSetupPath(value);
  try {
    const stat = await fsp.stat(resolved);
    if (stat.isDirectory() || stat.isFile()) return true;
    return `Mount root is not a file or folder: ${resolved}`;
  } catch {
    return `File or folder not found: ${resolved}`;
  }
}

function assertWriteFlags(options: { write?: boolean; readOnly?: boolean }): void {
  if (options.write && options.readOnly) {
    throw new Error('Use either --write or --read-only, not both.');
  }
}

function assertEnableFlags(options: { enable?: boolean; disable?: boolean }): void {
  if (options.enable && options.disable) {
    throw new Error('Use either --enable or --disable, not both.');
  }
}

function hasEditOptions(options: EditMountOptions): boolean {
  return Boolean(
    options.root ||
    options.mountPath !== undefined ||
    options.write ||
    options.readOnly ||
    options.description !== undefined ||
    options.guidance !== undefined ||
    options.exclude ||
    options.protect ||
    options.enable ||
    options.disable
  );
}

function assertMountNotReferenced(config: MvmtConfig, name: string): void {
  const mount = config.mounts.find((candidate) => candidate.name === name);
  const mountPath = mount?.path;
  for (const client of config.clients ?? []) {
    if (mountPath && client.permissions.some((permission) => (
      permission.path === mountPath || permission.path.startsWith(`${mountPath}/`)
    ))) {
      throw new Error(`Mount ${name} is still referenced by client ${client.id}`);
    }
  }

  const semantic = config.semanticTools;
  if (semantic?.searchPersonalContext?.sourceIds.includes(name)) {
    throw new Error(`Mount ${name} is still referenced by semanticTools.searchPersonalContext`);
  }
  if (semantic?.readContextItem?.sourceIds.includes(name)) {
    throw new Error(`Mount ${name} is still referenced by semanticTools.readContextItem`);
  }
}

function normalizeMountPath(value: string): string {
  const trimmed = stripTrailingSlashes(normalizePathSeparators(value.trim()));
  const withSlash = trimmed.startsWith('/') ? trimmed : `/${trimmed}`;
  return withSlash === '' ? '/' : withSlash;
}

function validateMountPathInput(value: string, config: MvmtConfig, currentName?: string): true | string {
  const trimmed = value.trim();
  if (!trimmed) return 'Enter a mount path such as /notes';
  const normalized = normalizeMountPath(trimmed);
  if (!/^\/[A-Za-z0-9._~/-]+$/.test(normalized)) return 'Use an absolute path such as /notes or /desktop/projects';
  if (normalized === '/') return 'Mount path cannot be /';
  if (config.mounts.some((mount) => mount.name !== currentName && mount.path === normalized)) {
    return `Mount path already exists: ${normalized}`;
  }
  return true;
}

function mountNameFromRoot(root: string): string {
  const resolved = resolveSetupPath(root);
  const slug = resolved
    .replaceAll('\\', '/')
    .split('/')
    .filter(Boolean)
    .join('-')
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '');
  return slug || 'mount';
}

function uniqueMountName(config: MvmtConfig, baseName: string): string {
  if (!config.mounts.some((mount) => mount.name === baseName)) return baseName;
  for (let index = 2; ; index += 1) {
    const candidate = `${baseName}-${index}`;
    if (!config.mounts.some((mount) => mount.name === candidate)) return candidate;
  }
}
