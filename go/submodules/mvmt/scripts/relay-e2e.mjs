#!/usr/bin/env node
import { spawn } from 'node:child_process';
import { createHash } from 'node:crypto';
import fs from 'node:fs';
import http from 'node:http';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const relayRoot = process.env.MVMT_RELAY_ROOT
  ? path.resolve(process.env.MVMT_RELAY_ROOT)
  : path.resolve(root, '..', 'mvmt-relay');
const workspaceSlug = process.env.MVMT_RELAY_WORKSPACE ?? 'demo';
const agentToken = process.env.MVMT_RELAY_TOKEN ?? 'dev-agent-token';
const projectName = `mvmt-relay-e2e-${process.pid}`;

main().catch((err) => {
  console.error(`relay e2e failed: ${err instanceof Error ? err.message : String(err)}`);
  process.exit(1);
});

async function main() {
  if (!fs.existsSync(path.join(relayRoot, 'compose.yaml'))) {
    throw new Error(`mvmt-relay compose.yaml not found at ${relayRoot}. Set MVMT_RELAY_ROOT if it lives elsewhere.`);
  }

  const tmpBase = fs.existsSync('/tmp') ? '/tmp' : os.tmpdir();
  const tmp = fs.mkdtempSync(path.join(tmpBase, 'mvmt-relay-e2e-'));
  const home = path.join(tmp, 'home');
  const mountRoot = path.join(tmp, 'files');
  const configPath = path.join(tmp, 'config.yaml');
  const fileName = 'hello.txt';
  const fileBody = `relay e2e ${Date.now()}\n`;
  const expectedSha = sha256(Buffer.from(fileBody));
  const relayPort = await findFreePort();
  const mvmtPort = await findFreePort();
  let mvmt;
  let mvmtOutput = '';

  try {
    fs.mkdirSync(home, { recursive: true });
    fs.mkdirSync(mountRoot, { recursive: true });
    fs.writeFileSync(path.join(mountRoot, fileName), fileBody);
    fs.writeFileSync(configPath, minimalConfig(mvmtPort, mountRoot));

    console.log(`relay root: ${relayRoot}`);
    console.log(`workspace: ${workspaceSlug}`);
    console.log(`relay port: ${relayPort}`);
    console.log(`mvmt port: ${mvmtPort}`);

    await compose(['up', '-d', '--build', 'relay'], {
      MVMT_RELAY_PORT: String(relayPort),
      MVMT_WORKSPACE_TOKENS: `${workspaceSlug}=${agentToken}`,
      MVMT_ALLOWED_WORKSPACES: workspaceSlug,
    });
    await waitForRelayHealth(relayPort);

    await runMvmt([
      '--no-update-check',
      'tunnel',
      'config',
      '--config',
      configPath,
      '--relay-url',
      `ws://127.0.0.1:${relayPort}/connect`,
      '--relay-workspace',
      workspaceSlug,
      '--relay-token',
      agentToken,
    ], { HOME: home });

    mvmt = spawn(process.execPath, [
      path.join(root, 'dist', 'bin', 'mvmt.js'),
      '--no-update-check',
      'serve',
      '--config',
      configPath,
      '--port',
      String(mvmtPort),
    ], {
      cwd: root,
      env: { ...process.env, HOME: home },
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    mvmt.stdout.on('data', (chunk) => {
      mvmtOutput += chunk.toString('utf-8');
    });
    mvmt.stderr.on('data', (chunk) => {
      mvmtOutput += chunk.toString('utf-8');
    });

    await waitForMvmtRelayStartup(mvmt, workspaceSlug);
    await waitFor(async () => {
      const response = await httpRequest(`http://127.0.0.1:${relayPort}/t/${workspaceSlug}/health`);
      if (response.status === 401 || response.status === 200) return;
      throw new Error(`GET /health returned ${response.status}: ${response.body.toString('utf-8')}`);
    }, 'relay tunnel registration');

    const leaseOutput = await runMvmt([
      '--no-update-check',
      'lease',
      'create',
      mountRoot,
      '--config',
      configPath,
      '--label',
      'Relay E2E',
      '--expires',
      '1h',
    ], { HOME: home });
    const lease = parseLeaseOutput(leaseOutput);
    const downloaded = await httpGet(
      `http://127.0.0.1:${relayPort}/t/${workspaceSlug}/lease/${lease.id}/files/${encodeURIComponent(fileName)}?token=${encodeURIComponent(lease.token)}`,
    );
    const actualSha = sha256(downloaded);
    if (actualSha !== expectedSha) {
      throw new Error(`download checksum mismatch: expected ${expectedSha}, got ${actualSha}`);
    }

    console.log(`ok: relay lease download sha256 ${actualSha}`);
  } catch (err) {
    if (mvmtOutput) console.error(`\n--- mvmt output ---\n${mvmtOutput.trim()}\n--- end mvmt output ---`);
    const relayLogs = await compose(['logs', 'relay'], {}, { allowFailure: true }).catch(() => '');
    if (relayLogs) console.error(`\n--- relay logs ---\n${relayLogs.trim()}\n--- end relay logs ---`);
    throw err;
  } finally {
    if (mvmt) await stopChild(mvmt).catch((err) => console.warn(err.message));
    await compose(['down', '--remove-orphans'], {}, { allowFailure: true }).catch(() => undefined);
    fs.rmSync(tmp, { recursive: true, force: true });
  }
}

function minimalConfig(port, mountRoot) {
  return [
    'version: 1',
    'server:',
    `  port: ${port}`,
    '  allowedOrigins: []',
    '  access: local',
    'proxy: []',
    'mounts:',
    '  - name: files',
    '    type: local_folder',
    '    path: /files',
    `    root: ${JSON.stringify(mountRoot)}`,
    '    writeAccess: false',
    'plugins: []',
    '',
  ].join('\n');
}

function runMvmt(args, extraEnv = {}) {
  return run(process.execPath, [path.join(root, 'dist', 'bin', 'mvmt.js'), ...args], {
    cwd: root,
    env: { ...process.env, ...extraEnv },
  });
}

function compose(args, extraEnv = {}, options = {}) {
  return run('docker', ['compose', '-p', projectName, '-f', path.join(relayRoot, 'compose.yaml'), ...args], {
    cwd: relayRoot,
    env: { ...process.env, ...extraEnv },
    allowFailure: options.allowFailure,
  });
}

function run(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    let output = '';
    const child = spawn(command, args, {
      cwd: options.cwd,
      env: options.env,
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    child.stdout.on('data', (chunk) => {
      output += chunk.toString('utf-8');
    });
    child.stderr.on('data', (chunk) => {
      output += chunk.toString('utf-8');
    });
    child.on('error', reject);
    child.on('exit', (code, signal) => {
      if (code === 0 || options.allowFailure) {
        resolve(output);
        return;
      }
      reject(new Error(`${command} ${args.join(' ')} exited with ${code ?? signal}\n${output}`));
    });
  });
}

function waitForRelayHealth(port) {
  return waitFor(async () => {
    const body = await httpGet(`http://127.0.0.1:${port}/actuator/health`);
    const parsed = JSON.parse(body.toString('utf-8'));
    if (parsed.status !== 'UP') throw new Error(`relay health is ${parsed.status ?? 'unknown'}`);
  }, 'relay health');
}

function waitForMvmtRelayStartup(child, slug) {
  return new Promise((resolve, reject) => {
    let output = '';
    const timer = setTimeout(() => {
      cleanup();
      reject(new Error(`timed out waiting for mvmt relay startup\n${output}`));
    }, 15_000);
    timer.unref();

    const onData = (chunk) => {
      output += chunk.toString('utf-8');
      if (output.includes(`Relay connected: ${slug}`)) {
        cleanup();
        resolve();
      }
    };
    const onExit = (code, signal) => {
      cleanup();
      reject(new Error(`mvmt exited before relay startup (${code ?? signal})\n${output}`));
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

function waitFor(fn, label) {
  const deadline = Date.now() + 60_000;
  return new Promise((resolve, reject) => {
    const attempt = async () => {
      try {
        await fn();
        resolve();
      } catch (err) {
        if (Date.now() > deadline) {
          reject(new Error(`timed out waiting for ${label}: ${err instanceof Error ? err.message : String(err)}`));
          return;
        }
        setTimeout(attempt, 500);
      }
    };
    void attempt();
  });
}

function httpGet(url, headers = {}) {
  return httpRequest(url, headers).then((response) => {
    if (response.status < 200 || response.status >= 300) {
      throw new Error(`GET ${url} returned ${response.status}: ${response.body.toString('utf-8')}`);
    }
    return response.body;
  });
}

function httpRequest(url, headers = {}) {
  return new Promise((resolve, reject) => {
    const req = http.get(url, { headers }, (res) => {
      const chunks = [];
      res.on('data', (chunk) => chunks.push(chunk));
      res.on('end', () => {
        const body = Buffer.concat(chunks);
        resolve({ status: res.statusCode ?? 0, body });
      });
    });
    req.on('error', reject);
  });
}

function parseLeaseOutput(output) {
  const match = output.match(/\/lease\/([^?\s]+)\?token=([^\s]+)/);
  if (!match) throw new Error(`could not parse lease URL from output\n${output}`);
  return { id: match[1], token: match[2] };
}

function sha256(bytes) {
  return createHash('sha256').update(bytes).digest('hex');
}

function findFreePort() {
  return new Promise((resolve, reject) => {
    const server = http.createServer();
    server.once('error', reject);
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      const port = typeof address === 'object' && address ? address.port : undefined;
      server.close(() => {
        if (port) resolve(port);
        else reject(new Error('failed to allocate a free port'));
      });
    });
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

    child.once('exit', () => {
      clearTimeout(timer);
      resolve();
    });
    child.kill('SIGTERM');
  });
}
