/**
 * `borg dashboard` - Open the web dashboard
 *
 * Launches the Borg WebUI dashboard in the default browser.
 * If the server isn't running, optionally starts it first.
 *
 * @example
 *   borg dashboard            # Open dashboard in browser
 *   borg dashboard --port 8080
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

export function registerDashboardCommand(program: Command): void {
  program
    .command('dashboard')
    .alias('ui')
    .description('Open the Borg WebUI dashboard in the default browser')
    .option('-p, --port <number>', 'Dashboard port', '3000')
    .option('-H, --host <address>', 'Dashboard host', 'localhost')
    .option('--no-open', 'Start dashboard server without opening browser')
    .option('--dev', 'Start in development mode with hot reload')
    .addHelpText('after', `
The dashboard provides a comprehensive visual interface to all Borg subsystems:
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
  $ borg dashboard                  Open in browser at localhost:3000
  $ borg dashboard --port 8080      Custom port
  $ borg dashboard --dev            Development mode with HMR
  $ borg dashboard --no-open        Start without opening browser
    `)
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const url = `http://${opts.host}:${opts.port}`;

      console.log(chalk.bold.cyan('\n  ⬡ Borg Dashboard\n'));
      console.log(chalk.dim(`  URL: ${url}`));
      console.log(chalk.dim(`  Mode: ${opts.dev ? 'development' : 'production'}`));
      console.log('');

      if (opts.dev) {
        // Start the Next.js dev server
        console.log(chalk.yellow('  Starting Next.js dev server...'));
        const { spawn } = await import('child_process');
        const { resolve } = await import('path');
        const webDir = resolve(process.cwd(), 'apps/web');
        const child = spawn('pnpm', ['dev', '--port', String(opts.port)], {
          stdio: 'inherit',
          cwd: webDir,
          shell: true,
          env: { ...process.env, BORG_TRPC_UPSTREAM: `http://127.0.0.1:4100/trpc` },
        });
        child.on('exit', (code) => process.exit(code ?? 0));
        return;
      }

      if (opts.open !== false) {
        try {
          const open = (await import('open')).default;
          await open(url);
          console.log(chalk.green('  ✓ Dashboard opened in browser'));
        } catch {
          console.log(chalk.yellow(`  ⚠ Could not open browser. Visit ${url} manually.`));
        }
      }

      console.log(chalk.dim('\n  Press Ctrl+C to stop\n'));
    });

  // About command (bonus)
  program
    .command('about')
    .description('Show Borg AIOS version, project info, and submodule status')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      if (opts.json) {
        console.log(JSON.stringify({
          name: 'Borg',
          subtitle: 'The Neural Operating System',
          version: getVersion(),
          codename: 'AIOS',
          packages: ['@borg/core', '@borg/cli', '@borg/types', '@borg/ai', '@borg/agents', '@borg/tools', '@borg/search', '@borg/memory', '@borg/adk'],
          repository: 'https://github.com/robertpelloni/borg',
        }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan('\n  ⬡ Borg — The Neural Operating System'));
      console.log(chalk.dim(`  Version: ${getVersion()} | Codename: AIOS\n`));
      console.log(chalk.dim('  "The Ultimate AI Tool Dashboard & Development Orchestrator"\n'));

      console.log(chalk.bold('  Packages:'));
      const pkgs = [
        ['@borg/core', 'Backend server, MCP router, orchestrator'],
        ['@borg/cli', 'Command-line interface'],
        ['@borg/types', 'Shared TypeScript types & Zod schemas'],
        ['@borg/ai', 'LLM service, model selector'],
        ['@borg/agents', 'Director, Council, Supervisor'],
        ['@borg/tools', 'File, terminal, browser, chain executor'],
        ['@borg/search', 'Semantic & text search service'],
        ['@borg/memory', 'Multi-backend memory system'],
        ['@borg/adk', 'Agent Development Kit'],
      ];

      for (const [name, desc] of pkgs) {
        console.log(chalk.cyan(`    ${name.padEnd(20)}`) + chalk.dim(desc));
      }

      console.log(chalk.dim('\n  Repository: https://github.com/robertpelloni/borg'));
      console.log(chalk.dim('  License: MIT'));

      // Quick-start hints
      console.log(chalk.bold.cyan('\n  Quick Start:'));
      console.log(chalk.dim('    borg start                # Launch the control plane'));
      console.log(chalk.dim('    borg info                 # System overview'));
      console.log(chalk.dim('    borg provider test openai  # Verify API keys'));
      console.log(chalk.dim('    borg catalog search memory # Browse MCP servers'));
      console.log(chalk.dim('    borg dashboard --dev       # Launch Web UI\n'));
    });
}
