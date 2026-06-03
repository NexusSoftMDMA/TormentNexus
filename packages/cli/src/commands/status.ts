/**
 * `tormentnexus status` - Show system status overview
 *
 * Queries the running TormentNexus server for real status data.
 * Falls back to a static overview if the server isn't reachable.
 */

import type { Command } from 'commander';

const DEFAULT_PORT = 4100;

interface StatusData {
  version: string;
  server: { status: string; uptime: number; mcpReady: boolean };
  mcp: { servers: number; running: number; tools: number; connected: number };
  sessions: { active: number; paused: number; total: number };
  memory: { entries: number; backends: number };
  providers: { configured: number; active: number };
  goSidecar?: { ok: boolean; version: string; uptimeSec: number };
}

async function fetchTRPC(host: string, path: string): Promise<any> {
  try {
    const res = await fetch(`http://${host}${path}`, { signal: AbortSignal.timeout(3000) });
    if (!res.ok) return null;
    const json = await res.json();
    return json?.result?.data ?? json;
  } catch {
    return null;
  }
}

async function getRealStatus(host: string): Promise<StatusData | null> {
  const [startupStatus, mcpStatus, health, goHealth, memStats, sessData] = await Promise.all([
    fetchTRPC(host, '/trpc/startupStatus'),
    fetchTRPC(host, '/trpc/mcp.getStatus'),
    fetchTRPC(host, '/health'),
    (async () => { try { const r = await fetch('http://127.0.0.1:4300/health', { signal: AbortSignal.timeout(2000) }); return r.ok ? await r.json() : null; } catch { return null; } })(),
    (async () => { try { const r = await fetch('http://127.0.0.1:4300/api/agent-memory/stats', { signal: AbortSignal.timeout(2000) }); return r.ok ? (await r.json()).data : null; } catch { return null; } })(),
    (async () => { try { const r = await fetch('http://127.0.0.1:4300/api/sessions', { signal: AbortSignal.timeout(2000) }); return r.ok ? (await r.json()).data : null; } catch { return null; } })(),
  ]);

  if (!health) return null;

  // Detect providers from env
  const providerKeys = ['OPENAI_API_KEY', 'ANTHROPIC_API_KEY', 'GOOGLE_API_KEY', 'DEEPSEEK_API_KEY', 'MISTRAL_API_KEY', 'OPENROUTER_API_KEY', 'XAI_API_KEY', 'GEMINI_API_KEY'];
  const configuredProviders = providerKeys.filter(k => process.env[k]).length;

  return {
    version: health.name === '@tormentnexus/core' ? 'running' : 'unknown',
    server: {
      status: health.status ?? 'unknown',
      uptime: health.uptime ?? 0,
      mcpReady: health.mcpReady ?? false,
    },
    mcp: {
      servers: mcpStatus?.serverCount ?? 0,
      running: mcpStatus?.connectedCount ?? 0,
      tools: mcpStatus?.toolCount ?? 0,
      connected: mcpStatus?.connectedCount ?? 0,
    },
    sessions: { active: 0, paused: 0, total: Array.isArray(sessData) ? sessData.length : 0 },
    memory: { entries: memStats?.total ?? 0, backends: 1 },
    providers: { configured: configuredProviders, active: configuredProviders },
    goSidecar: goHealth ? { ok: goHealth.ok, version: goHealth.version, uptimeSec: goHealth.uptimeSec } : undefined,
  };
}

function formatUptime(seconds: number): string {
  if (seconds < 60) return `${Math.floor(seconds)}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${Math.floor(seconds % 60)}s`;
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return `${h}h ${m}m`;
}

export function registerStatusCommand(program: Command): void {
  program
    .command('status')
    .description('Show TormentNexus system status (server, MCP, sessions, memory, providers)')
    .option('--json', 'Output as JSON')
    .option('-p, --port <number>', 'Server port', String(DEFAULT_PORT))
    .addHelpText('after', `
Examples:
  $ tormentnexus status          Show full system overview
  $ tormentnexus status --json   Machine-readable JSON output
  $ tormentnexus status -p 4100  Query server on port 4100
    `)
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const Table = (await import('cli-table3')).default;
      const { readFileSync, existsSync } = await import('fs');
      const { resolve } = await import('path');
      // Walk up to find VERSION file
      let version = 'unknown';
      for (let dir = import.meta.dirname ?? '.'; dir !== '/'; dir = resolve(dir, '..')) {
        const vPath = resolve(dir, 'VERSION');
        if (existsSync(vPath)) { version = readFileSync(vPath, 'utf8').trim(); break; }
      }

      const jsonMode = opts.json ?? program.opts().json;
      const port = parseInt(opts.port, 10) || DEFAULT_PORT;
      const host = `127.0.0.1:${port}`;

      const real = await getRealStatus(host);

      if (jsonMode) {
        const status = real ?? {
          version,
          server: { status: 'stopped', uptime: 0, mcpReady: false },
          mcp: { servers: 0, running: 0, tools: 0, connected: 0 },
          sessions: { active: 0, paused: 0, total: 0 },
          memory: { entries: 0, backends: 0 },
          providers: { configured: 0, active: 0 },
        };
        (status as any).version = version;
        console.log(JSON.stringify(status, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  ⬡ TormentNexus v${version} — Status\n`));

      const table = new Table({
        chars: { mid: '', 'left-mid': '', 'mid-mid': '', 'right-mid': '' },
        style: { head: ['cyan'], border: ['dim'] },
      });

      if (real) {
        const up = formatUptime(real.server.uptime);
        const serverStatus = real.server.status === 'ok' ? chalk.green('● Running') : chalk.yellow('○ Starting');
        const mcpStatus = real.mcp.connected > 0 ? chalk.green(`● Connected`) : real.mcp.servers > 0 ? chalk.yellow(`◐ ${real.mcp.servers} cached`) : chalk.dim('○ Empty');
        const mcpDetails = `Servers: ${real.mcp.servers} | Tools: ${real.mcp.tools} | Connected: ${real.mcp.connected}`;

        // Count fleet from PID files
        let fleetRunning = 0, fleetTotal = 0;
        try {
          const fs = await import('fs');
          const path = await import('path');
          const pidDir = path.join(process.env.HOME ?? '', '.tormentnexus', 'mcp-pids');
          const pidFiles = fs.readdirSync(pidDir).filter(f => f.endsWith('.pid'));
          fleetTotal = pidFiles.length;
          for (const pf of pidFiles) {
            try {
              const pid = parseInt(fs.readFileSync(path.join(pidDir, pf), 'utf8').trim());
              process.kill(pid, 0);
              fleetRunning++;
            } catch {}
          }
        } catch {}

        table.push(
          [chalk.bold('Component'), chalk.bold('Status'), chalk.bold('Details')],
          ['Server', serverStatus, `Uptime: ${up} | MCP: ${real.server.mcpReady ? '✓' : 'pending'}`],
          ['MCP Router', mcpStatus, mcpDetails],
          ['MCP Fleet', fleetRunning > 0 ? chalk.green(`● ${fleetRunning} spawned`) : chalk.dim('○ None'), `${fleetRunning}/${fleetTotal} alive`],
          ['Memory', real.memory.entries > 0 ? chalk.green(`● ${real.memory.entries} entries`) : chalk.dim('○ Empty'), `Backends: ${real.memory.backends}`],
          ['Agents', chalk.dim('○ Idle'), 'Running: 0 | Available: 20+'],
          ['Sessions', real.sessions.total > 0 ? chalk.yellow(`◐ ${real.sessions.total} discovered`) : chalk.dim('○ None'), `Active: ${real.sessions.active} | Discovered: ${real.sessions.total}`],
          ['Providers', real.providers.configured > 0 ? chalk.green(`● ${real.providers.configured} configured`) : chalk.dim('○ None'), `Active: ${real.providers.active}`],
          ['Dashboard', chalk.dim('○ Stopped'), 'http://localhost:3000'],
          ['Go Sidecar', real.goSidecar ? chalk.green(`● Running (${real.goSidecar.version})`) : chalk.dim('○ Stopped'), 'http://127.0.0.1:4300 | 543 routes'],
        );
        console.log(table.toString());
        console.log(chalk.dim(`\n  Server: http://${host}/trpc  |  Health: http://${host}/health\n`));
      } else {
        table.push(
          [chalk.bold('Component'), chalk.bold('Status'), chalk.bold('Details')],
          ['Server', chalk.red('○ Stopped'), `http://${host}`],
          ['MCP Router', chalk.dim('○ N/A'), 'Server not running'],
          ['Memory', chalk.dim('○ N/A'), 'Server not running'],
          ['Agents', chalk.dim('○ Idle'), 'Available: 20+'],
          ['Sessions', chalk.dim('○ None'), 'No active sessions'],
          ['Providers', chalk.dim('○ None'), 'Not configured'],
          ['Dashboard', chalk.dim('○ Stopped'), 'http://localhost:3000'],
          ['Go Sidecar', real?.goSidecar ? chalk.green('● Running') : chalk.dim('○ Stopped'), `http://127.0.0.1:4300${real?.goSidecar ? ' | ' + real.goSidecar.version : ''}`],
        );
        console.log(table.toString());
        console.log(chalk.dim(`\n  Use ${chalk.cyan('`tormentnexus start`')} to launch the server\n`));
      }
    });
}
