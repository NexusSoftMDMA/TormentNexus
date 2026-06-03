/**
 * `tormentnexus top` - Live system monitor
 * Shows updating status of all TormentNexus subsystems
 */
import type { Command } from 'commander';
import { readFileSync } from 'fs';
import { resolve } from 'path';

function getVersion(): string {
  try {
    let dir = process.cwd();
    for (let i = 0; i < 20; i++) {
      try { return readFileSync(resolve(dir, 'VERSION'), 'utf8').trim(); } catch {}
      const parent = resolve(dir, '..');
      if (parent === dir) break;
      dir = parent;
    }
  } catch {}
  return 'dev';
}

export function registerTopCommand(program: Command): void {
  program
    .command('top')
    .description('Live system monitor — auto-refreshing status dashboard')
    .option('-i, --interval <ms>', 'Refresh interval in milliseconds', '3000')
    .option('-n, --iterations <count>', 'Number of refreshes before exiting', '20')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const interval = parseInt(opts.interval) || 3000;
      const maxIterations = parseInt(opts.iterations) || 20;
      const version = getVersion();

      let iteration = 0;

      const update = async () => {
        // Clear screen
        process.stdout.write('\x1B[2J\x1B[H');

        console.log(chalk.bold.cyan(`  ⬡ TormentNexus TORMENTNEXUS v${version} — System Monitor`));
        console.log(chalk.dim(`  Refresh: ${interval}ms | Iteration: ${iteration + 1}/${maxIterations} | Ctrl+C to stop\n`));

        // Fetch all data in parallel
        const [health, mcpStatus, mcpServers, goHealth] = await Promise.all([
          fetchJSON('http://127.0.0.1:4100/health'),
          fetchTRPC('http://127.0.0.1:4100/trpc/mcp.getStatus'),
          fetchTRPC('http://127.0.0.1:4100/trpc/mcp.listServers'),
          fetchJSON('http://127.0.0.1:4300/health'),
        ]);

        // Server status
        if (health) {
          const up = health.uptime ?? 0;
          const h = Math.floor(up / 3600);
          const m = Math.floor((up % 3600) / 60);
          const s = Math.floor(up % 60);
          console.log(chalk.bold('  Server: ') + chalk.green(`● Running (${h}h ${m}m ${s}s)`));
        } else {
          console.log(chalk.bold('  Server: ') + chalk.red('○ Stopped'));
        }

        // MCP status
        if (mcpStatus) {
          console.log(chalk.bold('  MCP:    ') + chalk.yellow(`◐ ${mcpStatus.serverCount ?? 0} servers, ${mcpStatus.toolCount ?? 0} tools`));
        } else {
          console.log(chalk.bold('  MCP:    ') + chalk.dim('○ Not initialized'));
        }

        // Connected servers
        if (mcpServers && Array.isArray(mcpServers)) {
          const connected = mcpServers.filter((s: any) => s.runtimeConnected).length;
          const topServers = mcpServers
            .sort((a: any, b: any) => (b.toolCount ?? 0) - (a.toolCount ?? 0))
            .slice(0, 5);
          console.log(chalk.bold('  Online: ') + `${connected} connected`);
          for (const s of topServers) {
            const status = s.runtimeConnected ? chalk.green('●') : chalk.dim('○');
            console.log(`    ${status} ${s.name} (${s.toolCount ?? 0} tools)`);
          }
        }

        // Go sidecar
        if (goHealth && goHealth.ok) {
          console.log(chalk.bold('  Go:     ') + chalk.green(`● Running v${goHealth.version} (${Math.floor(goHealth.uptimeSec ?? 0)}s)`));
        } else {
          console.log(chalk.bold('  Go:     ') + chalk.dim('○ Stopped'));
        }

        // Memory
        const memStats = await fetchJSON('http://127.0.0.1:4300/api/agent-memory/stats');
        if (memStats?.data) {
          const m = memStats.data;
          console.log(chalk.bold('  Memory: ') + `${m.total ?? 0} entries (${m.longTerm ?? 0} long-term, ${m.working ?? 0} working)`);
        }

        // Providers
        const providerKeys = ['OPENAI_API_KEY', 'ANTHROPIC_API_KEY', 'GOOGLE_API_KEY', 'DEEPSEEK_API_KEY', 'MISTRAL_API_KEY', 'OPENROUTER_API_KEY', 'XAI_API_KEY'];
        const providers = providerKeys.filter(k => process.env[k]);
        if (providers.length > 0) {
          console.log(chalk.bold('  Providers:') + ` ${providers.length} active`);
        }

        // Sessions
        const sessRes = await fetchJSON('http://127.0.0.1:4300/api/sessions');
        if (sessRes?.data) {
          const sessions = sessRes.data;
          const active = sessions.filter((s: any) => s.status === 'active').length;
          console.log(chalk.bold('  Sessions:') + ` ${sessions.length} discovered` + (active > 0 ? `, ${active} active` : ''));
        }

        // Fleet
        try {
          const fs = await import('fs');
          const path = await import('path');
          const pidDir = path.join(process.env.HOME ?? '', '.tormentnexus', 'mcp-pids');
          const pidFiles = fs.readdirSync(pidDir).filter(f => f.endsWith('.pid'));
          let fleetAlive = 0;
          for (const pf of pidFiles) {
            try {
              const pid = parseInt(fs.readFileSync(path.join(pidDir, pf), 'utf8').trim());
              process.kill(pid, 0);
              fleetAlive++;
            } catch {}
          }
          if (pidFiles.length > 0) {
            console.log(chalk.bold('  Fleet:') + ` ${fleetAlive}/${pidFiles.length} spawned alive`);
          }
        } catch {}

        console.log('');
        iteration++;
        if (iteration < maxIterations) {
          setTimeout(update, interval);
        }
      };

      update();
    });
}

async function fetchJSON(url: string): Promise<any> {
  try {
    const res = await fetch(url, { signal: AbortSignal.timeout(2000) });
    if (res.ok) return await res.json();
  } catch {}
  return null;
}

async function fetchTRPC(url: string): Promise<any> {
  try {
    const res = await fetch(url, { signal: AbortSignal.timeout(2000) });
    if (res.ok) {
      const json = await res.json();
      return json?.result?.data ?? null;
    }
  } catch {}
  return null;
}
