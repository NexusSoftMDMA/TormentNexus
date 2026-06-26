import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

export interface PackageInfo {
  name: string;
  version: string;
  packagePath: string;
}

export type UpdateCheckStatus = 'available' | 'current' | 'unknown';

export interface UpdateCheckResult {
  status: UpdateCheckStatus;
  packageName: string;
  currentVersion: string;
  latestVersion?: string;
  reason?: string;
}

export interface UpdateCheckOptions {
  packageInfo?: PackageInfo;
  registryUrl?: string;
  timeoutMs?: number;
  fetchImpl?: typeof fetch;
  env?: NodeJS.ProcessEnv;
}

export function readPackageInfo(startUrl: string = import.meta.url): PackageInfo {
  let dir = fileURLToPath(new URL('.', startUrl));

  for (;;) {
    const packagePath = path.join(dir, 'package.json');
    if (fs.existsSync(packagePath)) {
      const parsed = JSON.parse(fs.readFileSync(packagePath, 'utf-8')) as { name?: unknown; version?: unknown };
      return {
        name: typeof parsed.name === 'string' && parsed.name.length > 0 ? parsed.name : 'mvmt',
        version: typeof parsed.version === 'string' && parsed.version.length > 0 ? parsed.version : '0.0.0',
        packagePath,
      };
    }

    const parent = path.dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }

  return { name: 'mvmt', version: '0.0.0', packagePath: '' };
}

export async function checkForUpdate(options: UpdateCheckOptions = {}): Promise<UpdateCheckResult> {
  const packageInfo = options.packageInfo ?? readPackageInfo();
  const env = options.env ?? process.env;

  if (shouldSkipUpdateCheck(env)) {
    return {
      status: 'unknown',
      packageName: packageInfo.name,
      currentVersion: packageInfo.version,
      reason: 'disabled',
    };
  }

  const fetchImpl = options.fetchImpl ?? globalThis.fetch;
  if (!fetchImpl) {
    return {
      status: 'unknown',
      packageName: packageInfo.name,
      currentVersion: packageInfo.version,
      reason: 'fetch unavailable',
    };
  }

  const registryUrl = options.registryUrl ?? `https://registry.npmjs.org/${encodeURIComponent(packageInfo.name)}/latest`;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), options.timeoutMs ?? 1500);

  try {
    const response = await fetchImpl(registryUrl, {
      signal: controller.signal,
      headers: {
        accept: 'application/json',
        'user-agent': `mvmt/${packageInfo.version}`,
      },
    });

    if (!response.ok) {
      return {
        status: 'unknown',
        packageName: packageInfo.name,
        currentVersion: packageInfo.version,
        reason: `registry returned ${response.status}`,
      };
    }

    const body = (await response.json()) as { version?: unknown };
    if (typeof body.version !== 'string' || body.version.length === 0) {
      return {
        status: 'unknown',
        packageName: packageInfo.name,
        currentVersion: packageInfo.version,
        reason: 'registry response missing version',
      };
    }

    return {
      status: isVersionGreater(body.version, packageInfo.version) ? 'available' : 'current',
      packageName: packageInfo.name,
      currentVersion: packageInfo.version,
      latestVersion: body.version,
    };
  } catch (err) {
    return {
      status: 'unknown',
      packageName: packageInfo.name,
      currentVersion: packageInfo.version,
      reason: err instanceof Error ? err.message : 'update check failed',
    };
  } finally {
    clearTimeout(timeout);
  }
}

export async function maybePrintUpdateNotice(
  options: UpdateCheckOptions & { stream?: NodeJS.WriteStream } = {},
): Promise<UpdateCheckResult> {
  const result = await checkForUpdate(options);
  const notice = formatUpdateNotice(result);
  if (notice) {
    (options.stream ?? process.stderr).write(`${notice}\n`);
  }
  return result;
}

export function formatUpdateNotice(result: UpdateCheckResult): string | undefined {
  if (result.status !== 'available' || !result.latestVersion) return undefined;
  return [
    `Update available: ${result.packageName} ${result.currentVersion} -> ${result.latestVersion}`,
    `Run: npm install -g ${result.packageName}`,
  ].join('\n');
}

export function shouldSkipUpdateCheck(env: NodeJS.ProcessEnv = process.env): boolean {
  return env.MVMT_NO_UPDATE_CHECK === '1' || env.MVMT_NO_UPDATE_CHECK === 'true' || Boolean(env.CI);
}

export function isVersionGreater(candidate: string, current: string): boolean {
  const candidateVersion = parseVersion(candidate);
  const currentVersion = parseVersion(current);
  if (!candidateVersion || !currentVersion) return false;

  for (let i = 0; i < 3; i++) {
    if (candidateVersion.parts[i] > currentVersion.parts[i]) return true;
    if (candidateVersion.parts[i] < currentVersion.parts[i]) return false;
  }

  return currentVersion.prerelease && !candidateVersion.prerelease;
}

function parseVersion(version: string): { parts: [number, number, number]; prerelease: boolean } | undefined {
  const [core, prerelease] = version.trim().replace(/^v/, '').split('-', 2);
  const pieces = core.split('.');
  if (pieces.length < 1 || pieces.length > 3) return undefined;

  const parts = pieces.map((piece) => {
    if (!/^\d+$/.test(piece)) return Number.NaN;
    return Number(piece);
  });

  if (parts.some((part) => !Number.isSafeInteger(part))) return undefined;

  return {
    parts: [parts[0] ?? 0, parts[1] ?? 0, parts[2] ?? 0],
    prerelease: Boolean(prerelease),
  };
}
