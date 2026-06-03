/**
 * `tormentnexus info` — Show comprehensive system information
 * Combines version, status, providers, catalog, and uptime in one view
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

export function registerInfoCommand(program: Command): void {
  program
    .command('info')
    .description('Show comprehensive system information (version, status, providers, catalog)')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const version = getVersion();

      if (opts.json) {
        const info: Record<string, any> = { version, codename: 'TORMENTNEXUS' };

        try {
          const res = await fetch('http://127.0.0.1:4100/health', { signal: AbortSignal.timeout(3000) });
          if (res.ok) info.server = await res.json();
        } catch { info.server = { status: 'stopped' }; }

        try {
          const res = await fetch('http://127.0.0.1:4100/trpc/mcp.getStatus', { signal: AbortSignal.timeout(3000) });
          if (res.ok) info.mcp = (await res.json())?.result?.data;
        } catch {}

        try {
          const res = await fetch('http://127.0.0.1:4300/health', { signal: AbortSignal.timeout(3000) });
          if (res.ok) info.goSidecar = await res.json();
        } catch {}

        console.log(JSON.stringify(info, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  ⬡ TormentNexus TORMENTNEXUS v${version} — System Info\n`));

      // Parallel fetch all data
      const [health, mcpStatus, goHealth, catalogStats] = await Promise.all([
        fetch('http://127.0.0.1:4100/health', { signal: AbortSignal.timeout(3000) }).then(r => r.ok ? r.json() : null).catch(() => null),
        fetch('http://127.0.0.1:4100/trpc/mcp.getStatus', { signal: AbortSignal.timeout(3000) }).then(r => r.ok ? r.json() : null).catch(() => null),
        fetch('http://127.0.0.1:4300/health', { signal: AbortSignal.timeout(3000) }).then(r => r.ok ? r.json() : null).catch(() => null),
        fetch('http://127.0.0.1:4300/api/catalog/stats', { signal: AbortSignal.timeout(3000) }).then(r => r.ok ? r.json() : null).catch(() => null),
      ]);

      // Server
      if (health) {
        const up = Math.floor(health.uptime / 60);
        const h = Math.floor(up / 60);
        const m = up % 60;
        console.log(chalk.bold('  Server: ') + chalk.green(`● Running (${h}h ${m}m)`));
      } else {
        console.log(chalk.bold('  Server: ') + chalk.red('○ Stopped'));
      }

      // Go Sidecar
      if (goHealth) {
        console.log(chalk.bold('  Go:     ') + chalk.green(`● Running v${goHealth.version} (543 routes)`));
      } else {
        console.log(chalk.bold('  Go:     ') + chalk.dim('○ Stopped'));
      }

      // MCP
      const mcp = mcpStatus?.result?.data;
      if (mcp) {
        console.log(chalk.bold('  MCP:    ') + chalk.yellow(`◐ ${mcp.serverCount} servers, ${mcp.toolCount} tools`));
      }

      // Catalog
      const catData = catalogStats?.data;
      if (catData) {
        console.log(chalk.bold('  Catalog:') + chalk.cyan(` ${catData.recentlyUpdated ?? 0} entries`));
      }

      // Providers (count from env)
      const envKeys = ['OPENAI_API_KEY', 'ANTHROPIC_API_KEY', 'GOOGLE_API_KEY', 'GEMINI_API_KEY', 'XAI_API_KEY', 'DEEPSEEK_API_KEY', 'MISTRAL_API_KEY', 'OPENROUTER_API_KEY'];
      const providerCount = envKeys.filter(k => process.env[k]).length;
      console.log(chalk.bold('  Providers:') + ` ${providerCount} active`);

      // AI Tools
      const { existsSync } = await import('fs');
      const { homedir } = await import('os');
      const home = homedir();
      const tools = [];
      if (existsSync(resolve(home, 'AppData/Roaming/Claude/claude_desktop_config.json'))) tools.push('Claude');
      if (existsSync(resolve(home, '.cursor/mcp.json'))) tools.push('Cursor');
      if (existsSync(resolve(home, 'AppData/Roaming/Code/User/settings.json'))) tools.push('VS Code');
      console.log(chalk.bold('  AI Tools:') + ` ${tools.length > 0 ? tools.join(', ') : 'none detected'}`);

      // Memory
      try {
        const memRes = await fetch('http://127.0.0.1:4300/api/agent-memory/stats', { signal: AbortSignal.timeout(2000) });
        if (memRes.ok) {
          const memData = (await memRes.json()).data;
          if (memData) console.log(chalk.bold('  Memory:  ') + `${memData.total ?? 0} entries (${memData.longTerm ?? 0} long-term, ${memData.working ?? 0} working)`);
        }
      } catch {}

      // Cloud providers
      try {
        const cloudRes = await fetch('http://127.0.0.1:4100/trpc/cloudDev.listProviders', { signal: AbortSignal.timeout(2000) });
        if (cloudRes.ok) {
          const cloudProviders = (await cloudRes.json())?.result?.data ?? [];
          if (cloudProviders.length > 0) {
            const names = cloudProviders.filter((p: any) => p.enabled).map((p: any) => p.name);
            console.log(chalk.bold('  Cloud:   ') + `${cloudProviders.length} providers` + (names.length > 0 ? ` (${names.join(', ')} active)` : ''));
          }
        }
      } catch {}

      // Harnesses
      try {
        const hRes = await fetch('http://127.0.0.1:4100/trpc/tools.detectCliHarnesses', { signal: AbortSignal.timeout(2000) });
        if (hRes.ok) {
          const harnesses = (await hRes.json())?.result?.data ?? [];
          const installed = harnesses.filter((h: any) => h.installed !== false);
          if (installed.length > 0) {
            const names = installed.slice(0, 5).map((h: any) => h.name ?? h);
            console.log(chalk.bold('  Harness:') + ` ${installed.length}/${harnesses.length} (${names.join(', ')}...)`);
          }
        }
      } catch {}

      // Sessions
      try {
        const sessRes = await fetch('http://127.0.0.1:4300/api/sessions', { signal: AbortSignal.timeout(2000) });
        if (sessRes.ok) {
          const sessions = (await sessRes.json())?.data ?? [];
          if (sessions.length > 0) console.log(chalk.bold('  Sessions:') + ` ${sessions.length} discovered`);

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
          console.log(chalk.bold('  Fleet:') + ` ${fleetAlive}/${pidFiles.length} spawned alive`);
        } catch {}
        }
      } catch {}

      console.log('');
    });
}
