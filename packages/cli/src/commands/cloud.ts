/**
 * `tormentnexus cloud` — Manage cloud development sessions (Jules, Codex, Devin, Copilot)
 */
import type { Command } from 'commander';

const GO_URL = 'http://127.0.0.1:4300';
const TS_URL = 'http://127.0.0.1:4100/trpc';

export function registerCloudCommand(program: Command): void {
  const cloud = program
    .command('cloud')
    .description('Cloud development — manage Jules, Codex, Devin, Copilot Workspace sessions');

  cloud
    .command('providers')
    .description('List cloud development providers and their status')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const Table = (await import('cli-table3')).default;

      let providers: any[] = [];
      try {
        const res = await fetch(`${TS_URL}/cloudDev.listProviders`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) providers = (await res.json())?.result?.data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ providers }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Cloud Development Providers (${providers.length})\n`));

      if (providers.length === 0) {
        console.log(chalk.dim('  No cloud providers configured.\n'));
        return;
      }

      const table = new Table({
        head: ['Provider', 'API Key', 'Enabled', 'Env Var'],
        style: { head: ['cyan'] },
      });

      for (const p of providers) {
        const key = p.hasApiKey ? chalk.green('✓') : chalk.red('✗');
        const enabled = p.enabled ? chalk.green('●') : chalk.dim('○');
        table.push([p.name, key, enabled, p.apiKeyEnvVar]);
      }

      console.log(table.toString());
      console.log('');
    });

  cloud
    .command('sessions')
    .description('List cloud development sessions')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let sessions: any[] = [];
      try {
        const res = await fetch(`${TS_URL}/cloudDev.listSessions?input=${encodeURIComponent(JSON.stringify({}))}`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) sessions = (await res.json())?.result?.data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ sessions }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Cloud Sessions (${sessions.length})\n`));
      if (sessions.length === 0) {
        console.log(chalk.dim('  No cloud sessions. Create one with `tormentnexus cloud create`.\n'));
        return;
      }

      for (const s of sessions) {
        const status = s.status === 'running' ? chalk.green('●') : chalk.dim('○');
        console.log(`  ${status} ${s.id ?? s.sessionId ?? '-'} ${chalk.dim(`(${s.provider ?? 'unknown'})`)}`);
      }
      console.log('');
    });

  cloud
    .command('stats')
    .description('Show cloud development statistics')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      let stats: any = {};
      try {
        const res = await fetch(`${TS_URL}/cloudDev.stats`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) stats = (await res.json())?.result?.data ?? {};
      } catch {}

      console.log(chalk.bold.cyan('\n  Cloud Development Statistics\n'));
      console.log(chalk.dim('  Providers:       ') + `${stats.providers ?? 0} (${stats.enabledProviders ?? 0} enabled)`);
      console.log(chalk.dim('  Total sessions:  ') + String(stats.totalSessions ?? 0));
      console.log(chalk.dim('  Total messages:  ') + String(stats.totalMessages ?? 0));
      console.log(chalk.dim('  Total logs:      ') + String(stats.totalLogs ?? 0));
      console.log('');
    });

  cloud
    .command('create <provider>')
    .description('Create a new cloud development session')
    .option('-d, --description <desc>', 'Session description')
    .option('-r, --repo <url>', 'Repository URL')
    .action(async (provider, opts) => {
      const chalk = (await import('chalk')).default;

      try {
        const res = await fetch(`${TS_URL}/cloudDev.createSession`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ json: { provider, description: opts.description, repoUrl: opts.repo } }),
          signal: AbortSignal.timeout(10000),
        });
        if (res.ok) {
          const json = await res.json();
          const session = json?.result?.data ?? {};
          console.log(chalk.green(`  ✓ Cloud session created`));
          console.log(chalk.dim(`    ID:       ${session.id ?? session.sessionId ?? 'pending'}`));
          console.log(chalk.dim(`    Provider: ${provider}`));
        } else {
          console.log(chalk.yellow(`  ⚠ Create returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  cloud
    .command('loops')
    .description('Show auto-dev loops (agentic coding cycles)')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let loops: any[] = [];
      try {
        const res = await fetch(`${TS_URL}/autoDev.getLoops`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) loops = (await res.json())?.result?.data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ loops }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Auto-Dev Loops (${loops.length})\n`));
      if (loops.length === 0) {
        console.log(chalk.dim('  No auto-dev loops running.\n'));
        return;
      }

      for (const l of loops) {
        const status = l.status === 'running' ? chalk.green('●') : l.status === 'completed' ? chalk.green('✓') : chalk.dim('○');
        console.log(`  ${status} ${l.id ?? '-'} ${chalk.dim(`(${l.status ?? 'unknown'})`)}`);
      }
      console.log('');
    });
}
