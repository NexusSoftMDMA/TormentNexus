#!/usr/bin/env node
import { spawn } from 'node:child_process';
import fs from 'node:fs';
import http from 'node:http';
import { createServer } from 'node:net';
import os from 'node:os';
import path from 'node:path';

const args = new Set(process.argv.slice(2));
const includeSmoke = !args.has('--no-smoke');
const includeAudit = args.has('--audit');
const npmCache = path.join(os.tmpdir(), 'mvmt-npm-cache');

const checks = [
  {
    name: 'build',
    command: process.platform === 'win32' ? 'npm.cmd' : 'npm',
    args: ['run', 'build'],
  },
  {
    name: 'test',
    command: process.platform === 'win32' ? 'npm.cmd' : 'npm',
    args: ['test'],
  },
  {
    name: 'diff whitespace',
    command: 'git',
    args: ['diff', '--check'],
  },
  {
    name: 'package dry-run',
    command: process.platform === 'win32' ? 'npm.cmd' : 'npm',
    args: ['pack', '--dry-run'],
    env: { npm_config_cache: npmCache },
  },
];

if (includeAudit) {
  checks.push({
    name: 'production dependency audit',
    command: process.platform === 'win32' ? 'npm.cmd' : 'npm',
    args: ['audit', '--omit=dev'],
  });
}

if (includeSmoke) {
  checks.push({
    name: 'runtime smoke',
    run: runSmokeTest,
  });
}

main().catch((err) => {
  console.error(`\nverify failed: ${err instanceof Error ? err.message : String(err)}`);
  process.exit(1);
});

async function main() {
  console.log(`mvmt verify${includeSmoke ? '' : ' --no-smoke'}${includeAudit ? ' --audit' : ''}\n`);

  for (const check of checks) {
    console.log(`==> ${check.name}`);
    if ('run' in check) {
      await check.run();
    } else {
      await runCommand(check.command, check.args, check.env);
    }
    console.log(`ok: ${check.name}\n`);
  }

  console.log('verify passed');
}

function runCommand(command, commandArgs, extraEnv = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, commandArgs, {
      stdio: 'inherit',
      env: { ...process.env, ...extraEnv },
    });

    child.on('error', reject);
    child.on('exit', (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }
      reject(new Error(`${command} ${commandArgs.join(' ')} exited with ${code ?? signal}`));
    });
  });
}

async function runSmokeTest() {
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-verify-'));
  const home = path.join(tmp, 'home');
  const mountRoot = path.join(tmp, 'mount');
  const configPath = path.join(tmp, 'config.yaml');
  const configuredPort = await findFreePort();
  let child;

  try {
    fs.mkdirSync(home, { recursive: true });
    fs.cpSync(path.join(process.cwd(), 'fixtures', 'sample-vault'), mountRoot, { recursive: true });
    fs.writeFileSync(
      configPath,
      [
        'version: 1',
        'server:',
        `  port: ${configuredPort}`,
        '  allowedOrigins: []',
        '  access: local',
        'proxy: []',
        'mounts:',
        '  - name: smoke',
        '    type: local_folder',
        '    path: /smoke',
        `    root: ${JSON.stringify(mountRoot)}`,
        '    writeAccess: false',
        'plugins: []',
        '',
      ].join('\n'),
    );

    child = spawn(process.execPath, ['dist/bin/mvmt.js', 'start', '--config', configPath, '--port', String(configuredPort)], {
      env: { ...process.env, HOME: home },
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    const { port } = await waitForStartup(child);
    if (port !== configuredPort) {
      throw new Error(`mvmt started on ${port}, expected ${configuredPort}`);
    }
    const token = fs.readFileSync(path.join(home, '.mvmt', '.session-token'), 'utf-8').trim();
    const health = await httpJson(`http://127.0.0.1:${port}/health`, token);
    if (health.status !== 'ok' || health.tools !== 5) {
      throw new Error(`unexpected health response: ${JSON.stringify(health)}`);
    }

    await stopChild(child);
    child = undefined;
    await assertPortClosed(port);
  } finally {
    if (child) await stopChild(child).catch(() => undefined);
    fs.rmSync(tmp, { recursive: true, force: true });
  }
}

function waitForStartup(child) {
  return new Promise((resolve, reject) => {
    let output = '';
    const timer = setTimeout(() => {
      cleanup();
      reject(new Error(`runtime smoke timed out waiting for startup\n${output}`));
    }, 10_000);
    timer.unref();

    const onData = (chunk) => {
      output += chunk.toString('utf-8');
      const match = output.match(/mvmt running -> http:\/\/127\.0\.0\.1:(\d+)\/dashboard/)
        ?? output.match(/mvmt running -> http:\/\/127\.0\.0\.1:(\d+)\/mcp/);
      if (match) {
        cleanup();
        resolve({ port: Number(match[1]) });
      }
    };

    const onExit = (code, signal) => {
      cleanup();
      reject(new Error(`mvmt exited before startup (${code ?? signal})\n${output}`));
    };

    const cleanup = () => {
      clearTimeout(timer);
      child.stdout.off('data', onData);
      child.stderr.off('data', onData);
      child.off('exit', onExit);
    };

    child.stdout.on('data', onData);
    child.stderr.on('data', onData);
    child.once('exit', onExit);
  });
}

function httpJson(url, token) {
  return new Promise((resolve, reject) => {
    const req = http.get(
      url,
      { headers: { Authorization: `Bearer ${token}` } },
      (res) => {
        let body = '';
        res.setEncoding('utf-8');
        res.on('data', (chunk) => {
          body += chunk;
        });
        res.on('end', () => {
          if (res.statusCode !== 200) {
            reject(new Error(`GET ${url} returned ${res.statusCode}: ${body}`));
            return;
          }
          try {
            resolve(JSON.parse(body));
          } catch (err) {
            reject(err);
          }
        });
      },
    );
    req.on('error', reject);
  });
}

function stopChild(child) {
  return new Promise((resolve, reject) => {
    if (child.exitCode !== null || child.signalCode !== null) {
      resolve();
      return;
    }

    const timer = setTimeout(() => {
      child.kill('SIGKILL');
      reject(new Error('mvmt did not exit after SIGTERM'));
    }, 6_000);
    timer.unref();

    child.once('exit', (code) => {
      clearTimeout(timer);
      if (code === 0 || code === null) {
        resolve();
      } else {
        reject(new Error(`mvmt exited with ${code}`));
      }
    });
    child.kill('SIGTERM');
  });
}

function assertPortClosed(port) {
  return new Promise((resolve, reject) => {
    const req = http.get(`http://127.0.0.1:${port}/health`, () => {
      reject(new Error(`port ${port} still accepted connections after shutdown`));
    });
    req.on('error', () => resolve());
    req.setTimeout(1000, () => {
      req.destroy();
      resolve();
    });
  });
}

function findFreePort() {
  return new Promise((resolve, reject) => {
    const server = createServer();
    server.once('error', reject);
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      const port = typeof address === 'object' && address !== null ? address.port : undefined;
      server.close(() => {
        if (port) resolve(port);
        else reject(new Error('Unable to allocate a local smoke-test port'));
      });
    });
  });
}
