/**
 * `tormentnexus git` — Git repository status and integration
 */
import type { Command } from 'commander';

const GO_URL = 'http://127.0.0.1:4300';

export function registerGitCommand(program: Command): void {
  const git = program
    .command('git')
    .description('Git — repository status, branches, and recent commits');

  git
    .command('status')
    .description('Show git repository status')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      let status: any = {};
      try {
        const res = await fetch(`${GO_URL}/api/git/status`, { signal: AbortSignal.timeout(3000) });
        if (res.ok) status = (await res.json()).data ?? {};
      } catch {}

      console.log(chalk.bold.cyan('\n  Git Status\n'));
      console.log(chalk.dim('  Branch:   ') + chalk.bold(status.branch ?? 'unknown'));
      console.log(chalk.dim('  Clean:    ') + (status.clean ? chalk.green('✓ Yes') : chalk.yellow('⚠ No')));

      if (status.modified?.length > 0) {
        console.log(chalk.dim('  Modified: ') + status.modified.join(', '));
      }
      if (status.staged?.length > 0) {
        console.log(chalk.dim('  Staged:   ') + status.staged.join(', '));
      }
      if (status.aheadBehind) {
        console.log(chalk.dim('  Ahead:    ') + `${status.aheadBehind.ahead ?? 0} commits`);
        console.log(chalk.dim('  Behind:   ') + `${status.aheadBehind.behind ?? 0} commits`);
      }
      console.log('');
    });

  git
    .command('log')
    .description('Show recent commits')
    .option('-n, --limit <count>', 'Number of commits', '10')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let log: any[] = [];
      try {
        const res = await fetch(`${GO_URL}/api/git/log?limit=${parseInt(opts.limit) || 10}`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) log = (await res.json()).data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ log }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Recent Commits (${log.length})\n`));
      if (log.length === 0) {
        console.log(chalk.dim('  No commits found.\n'));
        return;
      }

      for (const c of log) {
        const hash = (c.hash ?? '').substring(0, 8);
        const date = c.date ? new Date(c.date).toLocaleDateString() : '';
        const msg = (c.message ?? '').substring(0, 70);
        console.log(`  ${chalk.yellow(hash)} ${chalk.dim(date)} ${msg}`);
        console.log(chalk.dim(`    ${c.author ?? ''}`));
      }
      console.log('');
    });

  git
    .command('submodules')
    .description('Show submodule status')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      let subs: any[] = [];
      try {
        const res = await fetch(`${GO_URL}/api/submodules/status`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) subs = (await res.json()).data ?? [];
      } catch {}

      console.log(chalk.bold.cyan(`\n  Submodules (${subs.length})\n`));
      if (subs.length === 0) {
        console.log(chalk.dim('  No submodules found.\n'));
        return;
      }

      for (const s of subs) {
        const status = s.dirty ? chalk.yellow('⚠ dirty') : s.modified ? chalk.yellow('○ modified') : chalk.green('● clean');
        console.log(`  ${status} ${chalk.bold(s.name ?? s.path ?? '-')} ${chalk.dim(`(${s.ref ?? '-'})`)}`);
      }
      console.log('');
    });
}
