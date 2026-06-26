/**
 * `tormentnexus dashboard` - Open the web dashboard
 *
 * Launches the TormentNexus WebUI dashboard in the default browser.
 * If the server isn't running, optionally starts it first.
 *
 * @example
 *   tormentnexus dashboard            # Open dashboard in browser
 *   tormentnexus dashboard --port 8080
 */

import type { Command } from 'commander';
import { readFileSync } from 'fs';
import { resolve } from 'path';

async function openBrowserUrl(url: string): Promise<boolean> {
  const { spawn } = await import('child_process');

  try {
    if (process.platform === 'win32') {
      const child = spawn('cmd', ['/c', 'start', '', url], {
        stdio: 'ignore',
        detached: true,
      });
      child.unref();
      return true;
    }

    const opener = process.platform === 'darwin' ? 'open' : 'xdg-open';
    const child = spawn(opener, [url], {
      stdio: 'ignore',
      detached: true,
    });
    child.unref();
    return true;
  } catch {
    return false;
  }
}

async function isCoreReachable(trpcUrl: string): Promise<boolean> {
  const healthUrl = trpcUrl.replace(/\/trpc\/?$/, '/health');
  try {
    const response = await fetch(healthUrl, { signal: AbortSignal.timeout(1500) });
    return response.ok;
  } catch {
    return false;
  }
}

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

export function registerDashboardCommand(program: Command): void {
  program
    .command('dashboard')
    .alias('ui')
    .description('Open the TormentNexus WebUI dashboard in the default browser')
    .option('-p, --port <number>', 'Dashboard port', '3000')
    .option('-H, --host <address>', 'Dashboard host', 'localhost')
    .option('--no-open', 'Start dashboard server without opening browser')
    .option('--dev', 'Start in development mode with hot reload')
    .addHelpText('after', `
The dashboard provides a comprehensive visual interface to all TormentNexus subsystems:
  - System overview with health metrics
  - MCP Router management (servers, tools, traffic, config, directory)
  - Memory browser and search
  - Agent management and chat
  - Session tracking and control
  - Provider quota and billing dashboard
  - Tool browser with semantic search
  - Configuration editor
  - Submodule dashboard

Examples:
  $ tormentnexus dashboard                  Open in browser at localhost:3000
  $ tormentnexus dashboard --port 8080      Custom port
  $ tormentnexus dashboard --dev            Development mode with HMR
  $ tormentnexus dashboard --no-open        Start without opening browser
    `)
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const { spawn } = await import('child_process');
      const webDir = resolve(process.cwd(), 'apps/web');
      const url = `http://${opts.host}:${opts.port}`;
      const upstreamTrpc = process.env.TORMENTNEXUS_TRPC_UPSTREAM?.trim() || 'http://127.0.0.1:4100/trpc';
      const scriptPath = resolve(webDir, 'scripts', opts.dev ? 'dev.mjs' : 'start.mjs');
      const coreReachable = await isCoreReachable(upstreamTrpc);

      console.log(chalk.bold.cyan('\n  ⬡ TormentNexus Dashboard\n'));
      console.log(chalk.dim(`  URL: ${url}`));
      console.log(chalk.dim(`  Mode: ${opts.dev ? 'development' : 'production'}`));
      console.log(chalk.dim(`  Core: ${upstreamTrpc}`));
      console.log('');

      if (!coreReachable) {
        console.log(chalk.yellow('  ⚠ Core control plane is not responding yet.'));
        console.log(chalk.dim('    Start it with: tormentnexus start --port 4100'));
        console.log('');
      }

      console.log(chalk.yellow(`  Starting dashboard ${opts.dev ? 'dev' : 'standalone'} server...`));
      const child = spawn(process.execPath, [scriptPath, '--port', String(opts.port), '--host', String(opts.host)], {
        stdio: 'inherit',
        cwd: webDir,
        env: {
          ...process.env,
          TORMENTNEXUS_TRPC_UPSTREAM: upstreamTrpc,
        },
      });

      if (opts.open !== false) {
        const openDelayMs = opts.dev ? 3500 : 1500;
        setTimeout(() => {
          void openBrowserUrl(url).then((opened) => {
            if (opened) {
              console.log(chalk.green(`  ✓ Opening ${url}`));
            } else {
              console.log(chalk.yellow(`  ⚠ Could not open browser automatically. Visit ${url} manually.`));
            }
          });
        }, openDelayMs);
      }

      console.log(chalk.dim('\n  Press Ctrl+C to stop\n'));
      child.on('exit', (code) => process.exit(code ?? 0));
    });

  // About command (bonus)
  program
    .command('about')
    .description('Show TormentNexus TORMENTNEXUS version, project info, and submodule status')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      if (opts.json) {
        console.log(JSON.stringify({
          name: 'TormentNexus',
          subtitle: 'The Neural Operating System',
          version: getVersion(),
          codename: 'TORMENTNEXUS',
          packages: ['@tormentnexus/core', '@tormentnexus/cli', '@tormentnexus/types', '@tormentnexus/ai', '@tormentnexus/agents', '@tormentnexus/tools', '@tormentnexus/search', '@tormentnexus/memory', '@tormentnexus/adk'],
          repository: 'https://github.com/NexusSoftMDMA/TormentNexus',
        }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan('\n  ⬡ TormentNexus — The Neural Operating System'));
      console.log(chalk.dim(`  Version: ${getVersion()} | Codename: TORMENTNEXUS\n`));
      console.log(chalk.dim('  "The Ultimate AI Tool Dashboard & Development Orchestrator"\n'));

      console.log(chalk.bold('  Packages:'));
      const pkgs = [
        ['@tormentnexus/core', 'Backend server, MCP router, orchestrator'],
        ['@tormentnexus/cli', 'Command-line interface'],
        ['@tormentnexus/types', 'Shared TypeScript types & Zod schemas'],
        ['@tormentnexus/ai', 'LLM service, model selector'],
        ['@tormentnexus/agents', 'Director, Council, Supervisor'],
        ['@tormentnexus/tools', 'File, terminal, browser, chain executor'],
        ['@tormentnexus/search', 'Semantic & text search service'],
        ['@tormentnexus/memory', 'Multi-backend memory system'],
        ['@tormentnexus/adk', 'Agent Development Kit'],
      ];

      for (const [name, desc] of pkgs) {
        console.log(chalk.cyan(`    ${name.padEnd(20)}`) + chalk.dim(desc));
      }

      console.log(chalk.dim('\n  Repository: https://github.com/NexusSoftMDMA/TormentNexus'));
      console.log(chalk.dim('  License: MIT'));

      // Quick-start hints
      console.log(chalk.bold.cyan('\n  Quick Start:'));
      console.log(chalk.dim('    tormentnexus start                # Launch the control plane'));
      console.log(chalk.dim('    tormentnexus info                 # System overview'));
      console.log(chalk.dim('    tormentnexus provider test openai  # Verify API keys'));
      console.log(chalk.dim('    tormentnexus catalog search memory # Browse MCP servers'));
      console.log(chalk.dim('    tormentnexus dashboard --dev       # Launch Web UI\n'));
    });
}
