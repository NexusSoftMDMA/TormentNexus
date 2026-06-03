/**
 * `tormentnexus browser` — Browser automation control
 */
import type { Command } from 'commander';

const GO_URL = 'http://127.0.0.1:4300';

export function registerBrowserCommand(program: Command): void {
  const browser = program
    .command('browser')
    .description('Browser — control browser automation (open, close, navigate)');

  browser
    .command('status')
    .description('Show browser automation status')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      let status: any = {};
      try {
        const res = await fetch(`${GO_URL}/api/browser/status`, { signal: AbortSignal.timeout(3000) });
        if (res.ok) status = (await res.json()).data ?? {};
      } catch {}

      console.log(chalk.bold.cyan('\n  Browser Automation\n'));
      console.log(chalk.dim('  Active:    ') + (status.active ? chalk.green('● Yes') : chalk.dim('○ No')));
      console.log(chalk.dim('  Available: ') + (status.available ? chalk.green('● Yes') : chalk.dim('○ No')));
      console.log(chalk.dim('  Pages:     ') + String(status.pageCount ?? 0));
      if (status.pageIds?.length > 0) {
        for (const id of status.pageIds) {
          console.log(chalk.dim(`    ${id}`));
        }
      }
      console.log('');
    });

  browser
    .command('close-all')
    .description('Close all browser pages')
    .action(async () => {
      const chalk = (await import('chalk')).default;
      try {
        await fetch(`${GO_URL}/api/browser/close-all`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          signal: AbortSignal.timeout(5000),
        });
        console.log(chalk.green('  ✓ All browser pages closed'));
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });
}
