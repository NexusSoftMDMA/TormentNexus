/**
 * `tormentnexus plan` — Plan mode and diff sandbox management
 */
import type { Command } from 'commander';

const GO_URL = 'http://127.0.0.1:4300';

export function registerPlanCommand(program: Command): void {
  const plan = program
    .command('plan')
    .description('Plan — manage plan mode, diffs, checkpoints, and rollback');

  plan
    .command('status')
    .description('Show current plan mode and diff summary')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      let mode: any = {};
      let summary: string = '';
      try {
        const [modeRes, sumRes] = await Promise.all([
          fetch(`${GO_URL}/api/plan/mode`, { signal: AbortSignal.timeout(3000) }),
          fetch(`${GO_URL}/api/plan/summary`, { signal: AbortSignal.timeout(3000) }),
        ]);
        if (modeRes.ok) mode = (await modeRes.json()).data ?? {};
        if (sumRes.ok) summary = (await sumRes.json()).data ?? '';
      } catch {}

      console.log(chalk.bold.cyan('\n  Plan Mode\n'));
      console.log(chalk.dim('  Mode: ') + chalk.bold(mode.mode ?? 'PLAN'));

      if (summary) {
        console.log(chalk.dim('\n  ') + summary.split('\n').join('\n  '));
      }
      console.log('');
    });

  plan
    .command('diffs')
    .description('List pending diffs in the sandbox')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let diffs: any[] = [];
      try {
        const res = await fetch(`${GO_URL}/api/plan/diffs`, { signal: AbortSignal.timeout(3000) });
        if (res.ok) diffs = (await res.json()).data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ diffs }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Pending Diffs (${diffs.length})\n`));
      if (diffs.length === 0) {
        console.log(chalk.dim('  No pending diffs.\n'));
        return;
      }

      for (const d of diffs) {
        const status = d.approved ? chalk.green('✓ Approved') : d.rejected ? chalk.red('✗ Rejected') : chalk.yellow('○ Pending');
        console.log(`  ${status} ${d.file ?? d.path ?? '-'} ${chalk.dim(`(${d.type ?? 'change'})`)}`);
      }
      console.log('');
    });

  plan
    .command('approve <id>')
    .description('Approve a pending diff')
    .action(async (id) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/plan/approve-diff`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ diffId: id }),
          signal: AbortSignal.timeout(5000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Diff '${id}' approved`));
        } else {
          console.log(chalk.yellow(`  ⚠ Approve returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  plan
    .command('reject <id>')
    .description('Reject a pending diff')
    .action(async (id) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/plan/reject-diff`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ diffId: id }),
          signal: AbortSignal.timeout(5000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Diff '${id}' rejected`));
        } else {
          console.log(chalk.yellow(`  ⚠ Reject returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  plan
    .command('apply-all')
    .description('Apply all approved diffs')
    .action(async () => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/plan/apply-all`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          signal: AbortSignal.timeout(10000),
        });
        if (res.ok) {
          const data = (await res.json()).data;
          console.log(chalk.green(`  ✓ All approved diffs applied`));
          if (data) console.log(chalk.dim(`    ${JSON.stringify(data).substring(0, 100)}`));
        } else {
          console.log(chalk.yellow(`  ⚠ Apply returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  plan
    .command('checkpoints')
    .description('List plan checkpoints')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let checkpoints: any[] = [];
      try {
        const res = await fetch(`${GO_URL}/api/plan/checkpoints`, { signal: AbortSignal.timeout(3000) });
        if (res.ok) checkpoints = (await res.json()).data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ checkpoints }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Plan Checkpoints (${checkpoints.length})\n`));
      if (checkpoints.length === 0) {
        console.log(chalk.dim('  No checkpoints. Create one with `tormentnexus plan checkpoint`.\n'));
        return;
      }

      for (const c of checkpoints) {
        console.log(`  ${chalk.green('●')} ${c.id ?? c.name ?? '-'} ${chalk.dim(c.createdAt ?? '')}`);
      }
      console.log('');
    });

  plan
    .command('checkpoint <name>')
    .description('Create a new plan checkpoint')
    .action(async (name) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/plan/create-checkpoint`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name }),
          signal: AbortSignal.timeout(5000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Checkpoint '${name}' created`));
        } else {
          console.log(chalk.yellow(`  ⚠ Checkpoint returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  plan
    .command('rollback <checkpoint>')
    .description('Rollback to a plan checkpoint')
    .action(async (checkpoint) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/plan/rollback`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ checkpointId: checkpoint }),
          signal: AbortSignal.timeout(10000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Rolled back to '${checkpoint}'`));
        } else {
          console.log(chalk.yellow(`  ⚠ Rollback returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  plan
    .command('clear')
    .description('Clear all pending diffs and reset plan')
    .action(async () => {
      const chalk = (await import('chalk')).default;
      try {
        await fetch(`${GO_URL}/api/plan/clear`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          signal: AbortSignal.timeout(5000),
        });
        console.log(chalk.green('  ✓ Plan cleared'));
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });
}
