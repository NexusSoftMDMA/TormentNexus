import { spawn, spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, mkdirSync, copyFileSync, readdirSync, rmSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const REPO_ROOT = path.resolve(import.meta.dirname, '..', '..');
const IMAGE_TAG = 'mvmt-e2e:test';
const CONTAINER_PORT = 4141;
const READY_TIMEOUT_MS = 15_000;

let imagePromise: Promise<string> | undefined;

export function buildMvmtImage(): Promise<string> {
  imagePromise ??= buildMvmtImageOnce();
  return imagePromise;
}

async function buildMvmtImageOnce(): Promise<string> {
  await runHost('npm', ['run', 'build'], { cwd: REPO_ROOT });
  const buildCtx = mkdtempSync(path.join(os.tmpdir(), 'mvmt-e2e-build-'));
  try {
    await runHost('npm', ['pack', '--silent', '--pack-destination', buildCtx], { cwd: REPO_ROOT });
    const tarball = readdirSync(buildCtx).find((f) => f.endsWith('.tgz'));
    if (!tarball) throw new Error('npm pack did not produce a tarball');
    copyFileSync(path.join(buildCtx, tarball), path.join(buildCtx, 'mvmt.tgz'));
    copyFileSync(path.join(REPO_ROOT, 'tests', 'e2e', 'Dockerfile'), path.join(buildCtx, 'Dockerfile'));
    await runHost('docker', ['build', '--quiet', '-t', IMAGE_TAG, buildCtx]);
    return IMAGE_TAG;
  } finally {
    rmSync(buildCtx, { recursive: true, force: true });
  }
}

export interface ContainerBind {
  host: string;
  container: string;
  readonly?: boolean;
}

export interface RunContainerOptions {
  binds?: ContainerBind[];
}

export interface ExecResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

export interface HttpRequest {
  method: string;
  path: string;
  token?: string;
  body?: unknown;
  sessionId?: string;
  protocolVersion?: string;
}

export interface HttpResponse {
  status: number;
  headers: Record<string, string>;
  body: string;
}

export interface Container {
  id: string;
  exec(args: string[]): Promise<ExecResult>;
  run(args: string[]): Promise<ExecResult>;
  http(req: HttpRequest): Promise<HttpResponse>;
  startServer(): Promise<void>;
  stop(): Promise<void>;
}

export async function runMvmtContainer(opts: RunContainerOptions = {}): Promise<Container> {
  const image = await buildMvmtImage();
  const bindArgs: string[] = [];
  for (const bind of opts.binds ?? []) {
    if (!existsSync(bind.host)) mkdirSync(bind.host, { recursive: true });
    bindArgs.push('-v', `${bind.host}:${bind.container}${bind.readonly ? ':ro' : ''}`);
  }
  const { stdout } = await runHost('docker', ['run', '-d', '--rm', ...bindArgs, image]);
  const id = stdout.trim();
  if (!id) throw new Error('docker run did not return a container id');

  const container: Container = {
    id,
    exec: (args) => execInContainer(id, ['mvmt', ...args]),
    run: (args) => execInContainer(id, args),
    http: (req) => httpInContainer(id, req),
    startServer: () => startServerInContainer(id),
    stop: async () => {
      await runHost('docker', ['stop', '-t', '1', id], { ignoreExit: true });
    },
  };
  return container;
}

async function execInContainer(id: string, command: string[]): Promise<ExecResult> {
  return runHost('docker', ['exec', id, ...command], { ignoreExit: true });
}

async function httpInContainer(id: string, req: HttpRequest): Promise<HttpResponse> {
  const script = `
    (async () => {
      const r = JSON.parse(process.argv[1]);
      const headers = { Accept: 'application/json, text/event-stream' };
      if (r.token) headers.Authorization = 'Bearer ' + r.token;
      if (r.sessionId) headers['Mcp-Session-Id'] = r.sessionId;
      if (r.protocolVersion) headers['Mcp-Protocol-Version'] = r.protocolVersion;
      if (r.body !== undefined) headers['Content-Type'] = 'application/json';
      const res = await fetch('http://127.0.0.1:${CONTAINER_PORT}' + r.path, {
        method: r.method,
        headers,
        body: r.body !== undefined ? JSON.stringify(r.body) : undefined,
      });
      const out = { status: res.status, headers: {}, body: await res.text() };
      for (const [k, v] of res.headers) out.headers[k] = v;
      process.stdout.write(JSON.stringify(out));
    })().catch((err) => { process.stderr.write(String(err)); process.exit(1); });
  `;
  const result = await runHost('docker', ['exec', id, 'node', '-e', script, JSON.stringify(req)]);
  return JSON.parse(result.stdout);
}

async function startServerInContainer(id: string): Promise<void> {
  // Spawn `mvmt serve` detached inside the container; it will keep running
  // until the container is stopped. We poll /health until it answers.
  await runHost('docker', ['exec', '-d', id, 'sh', '-c',
    `mvmt serve --no-update-check >/workdir/serve.log 2>&1`,
  ]);
  // /.well-known/oauth-authorization-server is unauthenticated, unlike /health.
  const deadline = Date.now() + READY_TIMEOUT_MS;
  while (Date.now() < deadline) {
    const probe = await execInContainer(id, ['wget', '-qO-',
      `http://127.0.0.1:${CONTAINER_PORT}/.well-known/oauth-authorization-server`,
    ]);
    if (probe.exitCode === 0 && probe.stdout.includes('issuer')) return;
    await sleep(200);
  }
  const log = await execInContainer(id, ['cat', '/workdir/serve.log']);
  throw new Error(`mvmt serve did not become ready within ${READY_TIMEOUT_MS}ms\n${log.stdout}`);
}

function runHost(command: string, args: string[], opts: { cwd?: string; ignoreExit?: boolean } = {}): Promise<ExecResult> {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, { cwd: opts.cwd, stdio: ['ignore', 'pipe', 'pipe'] });
    let stdout = '';
    let stderr = '';
    child.stdout?.on('data', (chunk) => { stdout += chunk.toString(); });
    child.stderr?.on('data', (chunk) => { stderr += chunk.toString(); });
    child.on('error', reject);
    child.on('close', (code) => {
      const exitCode = code ?? 0;
      if (exitCode !== 0 && !opts.ignoreExit) {
        reject(new Error(`${command} ${args.join(' ')} exited ${exitCode}\nstdout: ${stdout}\nstderr: ${stderr}`));
        return;
      }
      resolve({ stdout, stderr, exitCode });
    });
  });
}

export function dockerAvailable(): boolean {
  const probe = spawnSync('docker', ['version', '--format', '{{.Server.Version}}'], { stdio: 'ignore' });
  return probe.status === 0;
}

export function parseMcpResponse(body: string): { jsonrpc: string; id: number; result?: any; error?: any } {
  const trimmed = body.trimStart();
  if (trimmed.startsWith('{')) return JSON.parse(body);
  const dataLine = body.split('\n').find((line) => line.startsWith('data: '));
  if (!dataLine) throw new Error(`could not parse MCP response: ${body}`);
  return JSON.parse(dataLine.slice('data: '.length));
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => { setTimeout(resolve, ms); });
}
