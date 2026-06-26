import { confirm, input, select } from '@inquirer/prompts';
import chalk from 'chalk';
import { getConfigPath, loadConfig } from '../config/loader.js';
import { resolveSetupPath } from '../connectors/setup-paths.js';
import { normalizeVirtualPath } from '../context/mount-registry.js';
import { resolveShareFileTarget } from '../share/files.js';
import { createShare, DEFAULT_SHARE_TTL, defaultSharesPath, listShares, removeShare, shareUnavailableReason } from '../share/store.js';
import { parseTokenTtl } from '../utils/token-ttl.js';
import { normalizeTunnelBaseUrl } from '../utils/tunnel.js';

export interface ShareCommandOptions {
  config?: string;
  json?: boolean;
  shareStorePath?: string;
}

export interface AddShareOptions extends ShareCommandOptions {
  expires?: string;
  ttl?: string;
}

export interface RemoveShareOptions extends ShareCommandOptions {
  yes?: boolean;
}

export async function listShareLinks(options: ShareCommandOptions = {}): Promise<void> {
  const shares = listShares(resolveShareStorePath(options));
  if (options.json) {
    console.log(JSON.stringify({ shares }, null, 2));
    return;
  }

  console.log(chalk.bold('Shares'));
  if (shares.length === 0) {
    console.log(`  ${chalk.dim('none')}`);
    return;
  }
  for (const share of shares) {
    const reason = shareUnavailableReason(share);
    const state = reason ? chalk.yellow(reason) : chalk.green('active');
    console.log(`  ${share.id.padEnd(16)} ${state}  ${share.path}  expires ${share.expiresAt ?? 'never'}  downloads ${share.downloadCount}`);
  }
}

export async function addShareLink(inputPath: string | undefined, options: AddShareOptions = {}): Promise<void> {
  if (!inputPath) throw new Error('Path is required. Example: mvmt share add /books/pg100.txt');
  const configPath = options.config ? resolveSetupPath(options.config) : getConfigPath();
  const config = loadConfig(configPath);
  const virtualPath = normalizeVirtualPath(inputPath);
  await resolveShareFileTarget(config.mounts, virtualPath);

  const ttl = parseTokenTtl(options.expires ?? options.ttl ?? DEFAULT_SHARE_TTL);
  const created = createShare(resolveShareStorePath(options), { path: virtualPath, expiresAt: ttl.expiresAt });
  const url = shareUrl(config, created.record.id, created.token);

  if (options.json) {
    console.log(JSON.stringify({ share: { ...created.record, url } }, null, 2));
    return;
  }

  console.log(chalk.green('Share created'));
  console.log(`  Path: ${created.record.path}`);
  console.log(`  Expires: ${created.record.expiresAt ?? 'never'}${options.expires || options.ttl ? '' : ` (${DEFAULT_SHARE_TTL} default)`}`);
  console.log(`  URL: ${url}`);
}

export async function removeShareLink(id: string | undefined, _options: RemoveShareOptions = {}): Promise<void> {
  if (!id) throw new Error('Share id is required.');
  if (!removeShare(resolveShareStorePath(_options), id)) throw new Error(`Unknown share: ${id}`);
  console.log(chalk.green(`Share ${id} removed`));
}

export async function promptAndAddShareLink(configPath: string): Promise<void> {
  const inputPath = await input({
    message: 'Mounted file path to share:',
    validate: (value) => value.trim().startsWith('/') ? true : 'Enter a mounted path such as /books/pg100.txt',
  });
  const expires = await input({
    message: 'Expires after:',
    default: DEFAULT_SHARE_TTL,
  });
  await addShareLink(inputPath.trim(), {
    config: configPath,
    expires: expires.trim() || DEFAULT_SHARE_TTL,
  });
}

export async function promptAndRemoveShareLink(options: ShareCommandOptions = {}): Promise<void> {
  const shares = listShares(resolveShareStorePath(options));
  if (shares.length === 0) {
    console.log(chalk.yellow('No shares configured.'));
    return;
  }
  const id = await select({
    message: 'Remove which share?',
    choices: shares.map((share) => ({
      name: `${share.id} (${share.path})`,
      value: share.id,
    })),
  });
  const ok = await confirm({
    message: `Remove share ${id}? Existing links will stop working.`,
    default: false,
  });
  if (!ok) {
    console.log(chalk.yellow('Share config unchanged.'));
    return;
  }
  await removeShareLink(id, options);
}

function resolveShareStorePath(options: ShareCommandOptions): string {
  return options.shareStorePath ?? defaultSharesPath();
}

function shareUrl(config: ReturnType<typeof loadConfig>, id: string, token: string): string {
  const base = config.server.access === 'tunnel' && config.server.tunnel?.url
    ? normalizeTunnelBaseUrl(config.server.tunnel.url)
    : `http://127.0.0.1:${config.server.port}`;
  const url = new URL(`/share/${id}`, base);
  url.searchParams.set('token', token);
  return url.toString();
}
