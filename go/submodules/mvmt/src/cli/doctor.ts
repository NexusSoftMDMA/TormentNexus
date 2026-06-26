import fs from 'fs';
import yaml from 'yaml';
import { expandHome, getConfigPath, parseConfig } from '../config/loader.js';
import { LocalFolderMountConfig, MvmtConfig } from '../config/schema.js';
import { checkForUpdate, PackageInfo, readPackageInfo, UpdateCheckResult } from '../utils/version.js';

export interface DoctorOptions {
  config?: string;
  json?: boolean;
  updateCheck?: boolean;
  packageInfo?: PackageInfo;
  fetchImpl?: typeof fetch;
  env?: NodeJS.ProcessEnv;
}

export type DoctorStatus = 'ok' | 'warn' | 'fail' | 'skip';

export interface DoctorConnectorReport {
  name: string;
  id?: string;
  kind: 'mount';
  enabled: boolean;
  status: DoctorStatus;
  message: string;
  toolCount?: number;
  durationMs?: number;
}

export interface DoctorReport {
  ok: boolean;
  version: {
    packageName: string;
    currentVersion: string;
    update: UpdateCheckResult;
  };
  config: {
    path: string;
    exists: boolean;
    valid: boolean;
    permissions?: string;
    permissionStatus?: DoctorStatus;
    server?: {
      port: number;
      allowedOrigins: string[];
    };
    errors: string[];
  };
  connectors: DoctorConnectorReport[];
  summary: {
    ok: number;
    warn: number;
    fail: number;
    skip: number;
  };
}

export async function doctor(options: DoctorOptions = {}): Promise<void> {
  const report = await collectDoctorReport(options);

  if (options.json) {
    console.log(JSON.stringify(report, null, 2));
  } else {
    printDoctorReport(report);
  }

  process.exitCode = report.ok ? 0 : 1;
}

export async function collectDoctorReport(options: DoctorOptions = {}): Promise<DoctorReport> {
  const packageInfo = options.packageInfo ?? readPackageInfo();
  const update = await checkForUpdate({
    packageInfo,
    timeoutMs: 1000,
    fetchImpl: options.fetchImpl,
    env: options.updateCheck === false ? { ...(options.env ?? process.env), MVMT_NO_UPDATE_CHECK: '1' } : options.env,
  });
  const configPath = expandHome(options.config ?? getConfigPath());
  const report: DoctorReport = {
    ok: true,
    version: {
      packageName: packageInfo.name,
      currentVersion: packageInfo.version,
      update,
    },
    config: {
      path: configPath,
      exists: false,
      valid: false,
      errors: [],
    },
    connectors: [],
    summary: { ok: 0, warn: 0, fail: 0, skip: 0 },
  };

  if (!fs.existsSync(configPath)) {
    report.config.errors.push('Config file does not exist. Run `mvmt config setup` or `mvmt serve`.');
    report.ok = false;
    summarize(report);
    return report;
  }

  report.config.exists = true;
  report.config.permissions = readMode(configPath);
  report.config.permissionStatus = checkConfigPermissions(configPath);

  let config: MvmtConfig;
  try {
    const raw = fs.readFileSync(configPath, 'utf-8');
    config = parseConfig(yaml.parse(raw));
  } catch (err) {
    report.config.errors.push(err instanceof Error ? err.message : 'Invalid config.');
    report.ok = false;
    summarize(report);
    return report;
  }

  report.config.valid = true;
  report.config.server = {
    port: config.server.port,
    allowedOrigins: config.server.allowedOrigins,
  };

  if (report.config.permissionStatus === 'warn') {
    report.config.errors.push('Config file is readable or writable by group/others. Recommended mode is 600.');
  }

  for (const mount of config.mounts) {
    report.connectors.push(checkMount(mount));
  }

  if (report.connectors.length === 0 || report.connectors.every((connector) => connector.status === 'skip')) {
    report.connectors.push({
      name: 'mounts',
      kind: 'mount',
      enabled: false,
      status: 'fail',
      message: 'no enabled mounts found',
    });
  }

  summarize(report);
  report.ok = report.config.valid && report.summary.fail === 0;
  return report;
}

function checkMount(mount: LocalFolderMountConfig): DoctorConnectorReport {
  if (mount.enabled === false) {
    return {
      name: mount.name,
      id: mount.path,
      kind: 'mount',
      enabled: false,
      status: 'skip',
      message: 'disabled in config',
    };
  }

  const root = expandHome(mount.root);
  try {
    const stat = fs.statSync(root);
    if (!stat.isDirectory() && !stat.isFile()) {
      return {
        name: mount.name,
        id: mount.path,
        kind: 'mount',
        enabled: true,
        status: 'fail',
        message: `mount root is not a file or folder: ${root}`,
      };
    }
    return {
      name: mount.name,
      id: mount.path,
      kind: 'mount',
      enabled: true,
      status: 'ok',
      message: `mounted ${root}`,
    };
  } catch (err) {
    return {
      name: mount.name,
      id: mount.path,
      kind: 'mount',
      enabled: true,
      status: 'fail',
      message: err instanceof Error ? err.message : `cannot access mount root: ${root}`,
    };
  }
}

export function printDoctorReport(report: DoctorReport): void {
  console.log('mvmt doctor\n');

  console.log('Version');
  console.log(`  mvmt ${report.version.currentVersion}`);
  if (report.version.update.status === 'available') {
    console.log(`  update available: ${report.version.update.latestVersion}`);
  } else if (report.version.update.status === 'current') {
    console.log('  update check: current');
  } else {
    console.log(`  update check: unavailable${report.version.update.reason ? ` (${report.version.update.reason})` : ''}`);
  }

  console.log('\nConfig');
  console.log(`  path: ${report.config.path}`);
  console.log(`  exists: ${report.config.exists ? 'yes' : 'no'}`);
  console.log(`  valid: ${report.config.valid ? 'yes' : 'no'}`);
  if (report.config.permissions) {
    console.log(`  permissions: ${report.config.permissions}${report.config.permissionStatus === 'warn' ? ' (warning)' : ''}`);
  }
  if (report.config.server) {
    console.log(`  port: ${report.config.server.port}`);
    console.log(
      `  allowed origins: ${
        report.config.server.allowedOrigins.length > 0 ? report.config.server.allowedOrigins.join(', ') : '(localhost only)'
      }`,
    );
  }
  for (const error of report.config.errors) {
    console.log(`  warning: ${error}`);
  }

  console.log('\nMounts');
  for (const connector of report.connectors) {
    const label = connector.id ?? connector.name;
    const detail =
      connector.status === 'ok'
        ? `${connector.toolCount ?? 0} tools in ${connector.durationMs ?? 0}ms`
        : connector.message;
    console.log(`  ${connector.status.padEnd(4)} ${label.padEnd(24)} ${detail}`);
  }

  console.log('\nSummary');
  console.log(
    `  ok: ${report.summary.ok}, warnings: ${report.summary.warn}, failed: ${report.summary.fail}, skipped: ${report.summary.skip}`,
  );
}

function summarize(report: DoctorReport): void {
  report.summary = { ok: 0, warn: 0, fail: 0, skip: 0 };

  if (report.config.permissionStatus === 'warn') report.summary.warn += 1;

  for (const connector of report.connectors) {
    report.summary[connector.status] += 1;
  }
}

function readMode(configPath: string): string | undefined {
  if (process.platform === 'win32') return undefined;
  const mode = fs.statSync(configPath).mode & 0o777;
  return `0${mode.toString(8)}`;
}

function checkConfigPermissions(configPath: string): DoctorStatus | undefined {
  if (process.platform === 'win32') return undefined;
  const mode = fs.statSync(configPath).mode & 0o777;
  return mode & 0o077 ? 'warn' : 'ok';
}
