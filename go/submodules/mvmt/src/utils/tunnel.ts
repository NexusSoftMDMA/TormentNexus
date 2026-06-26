import { spawn, spawnSync, ChildProcessByStdio } from 'child_process';
import { Readable } from 'stream';
import { TunnelConfig } from '../config/schema.js';

export interface RunningTunnel {
  command: string;
  url?: string;
  stop(): Promise<void>;
}

interface StartTunnelOptions {
  timeoutMs?: number;
  onOutput?(line: string): void;
}

const HTTPS_URL_PATTERN = /https:\/\/[^\s"'<>),]+/gi;
const HOST_PATTERN = /\b[a-z0-9][a-z0-9-]*\.(?:trycloudflare\.com|lhr\.life|lhr\.rocks|loca\.lt|a\.pinggy\.io|pinggy\.io|pinggy\.link)\b/i;
// localhost.run is intentionally excluded: real tunnels now use .lhr.life,
// while .localhost.run subdomains (e.g. admin.localhost.run) are account
// management URLs that localhost.run hands out when no SSH key is registered.
const TUNNEL_HOST_SUFFIXES = [
  'trycloudflare.com',
  'lhr.life',
  'lhr.rocks',
  'loca.lt',
  'a.pinggy.io',
  'pinggy.io',
  'pinggy.link',
];

type TunnelChild = ChildProcessByStdio<null, Readable, Readable>;

export async function startTunnel(
  commandTemplate: string,
  port: number,
  options: StartTunnelOptions = {},
): Promise<RunningTunnel> {
  const command = renderTunnelCommand(commandTemplate, port);
  const child = spawn(command, {
    shell: true,
    stdio: ['ignore', 'pipe', 'pipe'],
  });

  const url = await waitForPublicUrl(child, options);

  return {
    command,
    url,
    stop: () => stopTunnel(child),
  };
}

export function renderTunnelCommand(commandTemplate: string, port: number): string {
  return commandTemplate.replaceAll('{port}', String(port));
}

export function formatMcpPublicUrl(publicUrl: string): string {
  return formatPublicPath(publicUrl, '/mcp');
}

export function formatDashboardPublicUrl(publicUrl: string): string {
  return formatPublicPath(publicUrl, '/dashboard');
}

function formatPublicPath(publicUrl: string, route: string): string {
  return `${trimTrailingSlashes(publicUrl)}${route}`;
}

export function normalizeTunnelBaseUrl(value: string): string {
  const parsed = new URL(value.trim());
  parsed.hash = '';
  parsed.search = '';
  if (parsed.pathname === '/mcp' || parsed.pathname === '/') {
    parsed.pathname = '';
  }
  return trimTrailingSlashes(parsed.toString());
}

export function cloudflareNamedTunnelCommand(configPath: string): string {
  return `cloudflared tunnel --config ${shellQuote(configPath)} run`;
}

export function defaultTunnelCommand(provider: Exclude<TunnelConfig['provider'], 'custom' | 'relay'>): string {
  const sshFlags = [
    '-T',
    '-o StrictHostKeyChecking=accept-new',
    '-o UserKnownHostsFile=/dev/null',
    '-o ExitOnForwardFailure=yes',
    '-o ServerAliveInterval=60',
    '-o ServerAliveCountMax=3',
  ];
  switch (provider) {
    case 'cloudflare-quick':
      return 'cloudflared tunnel --url http://127.0.0.1:{port}';
    case 'pinggy':
      return ['ssh', '-p 443', ...sshFlags, '-R0:localhost:{port}', 'a.pinggy.io'].join(' ');
    case 'localhost-run':
      return ['ssh', ...sshFlags, '-R 80:localhost:{port}', 'nokey@localhost.run'].join(' ');
  }
}

export function missingTunnelDependency(
  tunnel: TunnelConfig | undefined,
  isAvailable: (command: string) => boolean = isCommandAvailable,
): string | undefined {
  if (!tunnel || tunnel.provider === 'relay') return undefined;
  if (tunnel?.provider === 'cloudflare-quick' && !isAvailable('cloudflared')) {
    return 'cloudflared';
  }
  if (tunnel?.provider === 'custom' && commandStartsWith(tunnel.command, 'cloudflared') && !isAvailable('cloudflared')) {
    return 'cloudflared';
  }
  return undefined;
}

export function extractPublicUrl(text: string): string | undefined {
  for (const match of text.matchAll(HTTPS_URL_PATTERN)) {
    const candidate = trimTrailingPunctuation(match[0]);
    if (isKnownTunnelUrl(candidate)) return candidate;
  }

  const hostMatch = text.match(HOST_PATTERN);
  if (hostMatch) return `https://${trimTrailingPunctuation(hostMatch[0])}`;

  return undefined;
}

function isCommandAvailable(command: string): boolean {
  const result = spawnSync(command, ['--version'], { stdio: 'ignore' });
  return !result.error;
}

function commandStartsWith(command: string, executable: string): boolean {
  const trimmed = command.trim();
  return trimmed === executable || trimmed.startsWith(`${executable} `);
}

function shellQuote(value: string): string {
  if (/^[A-Za-z0-9_/:=.,@%+-]+$/.test(value)) return value;
  return `'${value.replace(/'/g, "'\\''")}'`;
}

function waitForPublicUrl(
  child: TunnelChild,
  options: StartTunnelOptions,
): Promise<string | undefined> {
  const timeoutMs = options.timeoutMs ?? 15_000;

  return new Promise((resolve, reject) => {
    let settled = false;
    let buffer = '';

    const drain = (chunk: Buffer) => {
      const text = chunk.toString('utf-8');
      for (const line of text.split(/\r?\n/).filter(Boolean)) {
        options.onOutput?.(line);
      }
    };

    const settle = (fn: () => void) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      child.stdout.off('data', onData);
      child.stderr.off('data', onData);
      child.off('error', onError);
      child.off('exit', onExit);
      // Keep stdio flowing so the child doesn't stall on full pipe buffers.
      child.stdout.on('data', drain);
      child.stderr.on('data', drain);
      fn();
    };

    const onData = (chunk: Buffer) => {
      const text = chunk.toString('utf-8');
      buffer += text;
      for (const line of text.split(/\r?\n/).filter(Boolean)) {
        options.onOutput?.(line);
      }

      const url = extractPublicUrl(buffer);
      if (url) settle(() => resolve(url));
    };

    const onError = (error: Error) => {
      settle(() => reject(error));
    };

    const onExit = (code: number | null, signal: NodeJS.Signals | null) => {
      settle(() => reject(new Error(`Tunnel command exited before a public URL was detected (code ${code ?? signal ?? 'unknown'})`)));
    };

    const timer = setTimeout(() => {
      settle(() => resolve(undefined));
    }, timeoutMs);
    timer.unref();

    child.stdout.on('data', onData);
    child.stderr.on('data', onData);
    child.once('error', onError);
    child.once('exit', onExit);
  });
}

function stopTunnel(child: TunnelChild): Promise<void> {
  if (child.exitCode !== null || child.killed) return Promise.resolve();

  return new Promise((resolve) => {
    const timer = setTimeout(() => {
      child.kill('SIGKILL');
      resolve();
    }, 2_000);
    timer.unref();

    child.once('exit', () => {
      clearTimeout(timer);
      resolve();
    });
    child.kill('SIGTERM');
  });
}

function trimTrailingPunctuation(value: string): string {
  let end = value.length;
  while (end > 0) {
    const ch = value.charCodeAt(end - 1);
    // '.' ',' ';' ':'
    if (ch === 46 || ch === 44 || ch === 59 || ch === 58) end -= 1;
    else break;
  }
  return end === value.length ? value : value.slice(0, end);
}

function trimTrailingSlashes(value: string): string {
  let end = value.length;
  while (end > 0 && value.charCodeAt(end - 1) === 47 /* '/' */) end -= 1;
  return end === value.length ? value : value.slice(0, end);
}

function isKnownTunnelUrl(value: string): boolean {
  try {
    const hostname = new URL(value).hostname.toLowerCase();
    return TUNNEL_HOST_SUFFIXES.some((suffix) => hostname.endsWith(`.${suffix}`));
  } catch {
    return false;
  }
}
