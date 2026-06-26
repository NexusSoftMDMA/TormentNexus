import fs from 'fs';
import fsp from 'fs/promises';
import type { FileHandle } from 'fs/promises';
import os from 'os';
import path from 'path';
import { randomUUID } from 'crypto';
import yaml from 'yaml';
import { ConfigSchema, MvmtConfig } from './schema.js';

export async function saveConfig(configPath: string, config: MvmtConfig): Promise<void> {
  const configDir = path.dirname(configPath);
  await fsp.mkdir(configDir, { recursive: true });
  const tempPath = path.join(configDir, `.${path.basename(configPath)}.${process.pid}.${randomUUID()}.tmp`);

  try {
    const handle = await fsp.open(tempPath, 'w', 0o600);
    try {
      await handle.writeFile(yaml.stringify(config), { encoding: 'utf-8' });
      await handle.sync();
    } finally {
      await handle.close();
    }
    if (process.platform !== 'win32') {
      await fsp.chmod(tempPath, 0o600);
    }
    await fsp.rename(tempPath, configPath);
    if (process.platform !== 'win32') {
      await fsp.chmod(configPath, 0o600);
      await fsyncDirectory(configDir);
    }
  } catch (err) {
    await fsp.rm(tempPath, { force: true }).catch(() => undefined);
    throw err;
  }
}

export async function withConfigLock<T>(
  configPath: string,
  fn: () => Promise<T>,
  options: { waitMs?: number; staleMs?: number } = {},
): Promise<T> {
  const lockPath = `${configPath}.lock`;
  await fsp.mkdir(path.dirname(configPath), { recursive: true });
  const handle = await acquireConfigLock(lockPath, {
    waitMs: options.waitMs ?? 10_000,
    staleMs: options.staleMs ?? 30_000,
  });
  try {
    return await fn();
  } finally {
    await handle.close().catch(() => undefined);
    await fsp.rm(lockPath, { force: true }).catch(() => undefined);
  }
}

export function getConfigPath(): string {
  return path.join(os.homedir(), '.mvmt', 'config.yaml');
}

export function parseConfig(rawConfig: unknown): MvmtConfig {
  const result = ConfigSchema.safeParse(rawConfig);
  if (!result.success) {
    const details = result.error.issues
      .map((issue) => `  ${issue.path.join('.') || '(root)'}: ${issue.message}`)
      .join('\n');
    throw new Error(`Invalid config:\n${details}`);
  }
  return result.data;
}

export function resolveConfigPath(overridePath?: string): string {
  return overridePath ? expandHome(overridePath) : getConfigPath();
}

export function configExists(overridePath?: string): boolean {
  return fs.existsSync(resolveConfigPath(overridePath));
}

export function readConfig(configPath: string): MvmtConfig {
  const raw = fs.readFileSync(configPath, 'utf-8');
  return parseConfig(yaml.parse(raw));
}

export function loadConfig(overridePath?: string): MvmtConfig {
  const configPath = resolveConfigPath(overridePath);

  if (!fs.existsSync(configPath)) {
    console.error(`Config not found at ${configPath}`);
    console.error('Run `mvmt config setup` to create a config, or `mvmt serve` to set up and start mvmt.');
    process.exit(1);
  }

  try {
    return readConfig(configPath);
  } catch (err) {
    if (err instanceof Error) {
      console.error(err.message);
    } else {
      console.error('Invalid config.');
    }
    process.exit(1);
  }
}

export function expandHome(inputPath: string): string {
  if (inputPath === '~') return os.homedir();
  if (inputPath.startsWith(`~${path.sep}`)) {
    return path.join(os.homedir(), inputPath.slice(2));
  }
  return inputPath;
}

async function acquireConfigLock(
  lockPath: string,
  options: { waitMs: number; staleMs: number },
): Promise<FileHandle> {
  const startedAt = Date.now();
  while (true) {
    try {
      const handle = await fsp.open(lockPath, 'wx', 0o600);
      await handle.writeFile(JSON.stringify({ pid: process.pid, createdAt: new Date().toISOString() }));
      return handle;
    } catch (err) {
      const code = (err as NodeJS.ErrnoException).code;
      if (code !== 'EEXIST') throw err;
      await removeStaleLock(lockPath, options.staleMs);
      if (Date.now() - startedAt >= options.waitMs) {
        throw new Error(`Timed out waiting for config lock at ${lockPath}`);
      }
      await sleep(50);
    }
  }
}

async function removeStaleLock(lockPath: string, staleMs: number): Promise<void> {
  try {
    const stat = await fsp.stat(lockPath);
    if (Date.now() - stat.mtimeMs > staleMs) {
      await fsp.rm(lockPath, { force: true });
    }
  } catch {
    // If the lock disappeared between open attempts, retry acquisition.
  }
}

async function fsyncDirectory(dir: string): Promise<void> {
  let handle: FileHandle | undefined;
  try {
    handle = await fsp.open(dir, 'r');
    await handle.sync();
  } catch {
    // Directory fsync is best-effort across filesystems.
  } finally {
    await handle?.close().catch(() => undefined);
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
