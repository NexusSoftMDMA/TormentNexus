/**
 * `tormentnexus scripts` — Manage saved automation scripts
 */
import type { Command } from 'commander';

const GO_URL = 'http://127.0.0.1:4300';

export function registerScriptsCommand(program: Command): void {
  const scripts = program
    .command('scripts')
    .description('Scripts — manage saved automation scripts');

  scripts
    .command('list')
    .description('List all saved scripts')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let scriptList: any[] = [];
      try {
        const res = await fetch(`${GO_URL}/api/scripts`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) scriptList = (await res.json()).data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ scripts: scriptList }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Saved Scripts (${scriptList.length})\n`));
      if (scriptList.length === 0) {
        console.log(chalk.dim('  No scripts saved. Create one with `tormentnexus scripts create`.\n'));
        return;
      }

      for (const s of scriptList) {
        const name = s.name ?? s.id ?? '-';
        const lang = s.language ?? s.type ?? 'shell';
        console.log(`  ${chalk.cyan(name)} ${chalk.dim(`(${lang})`)}`);
        if (s.description) console.log(chalk.dim(`    ${s.description}`));
      }
      console.log('');
    });

  scripts
    .command('create <name>')
    .description('Create a new saved script')
    .option('-d, --description <desc>', 'Script description')
    .option('-l, --language <lang>', 'Script language', 'shell')
    .option('-c, --content <code>', 'Script content')
    .action(async (name, opts) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/scripts/create`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, description: opts.description, language: opts.language, content: opts.content ?? '' }),
          signal: AbortSignal.timeout(5000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Script '${name}' created`));
        } else {
          console.log(chalk.yellow(`  ⚠ Create returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  scripts
    .command('run <name>')
    .description('Execute a saved script')
    .option('--args <args...>', 'Arguments to pass to the script')
    .action(async (name, opts) => {
      const chalk = (await import('chalk')).default;
      console.log(chalk.yellow(`  Running script '${name}'...`));
      try {
        const res = await fetch(`${GO_URL}/api/scripts/execute`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, args: opts.args ?? [] }),
          signal: AbortSignal.timeout(30000),
        });
        if (res.ok) {
          const json = await res.json();
          const data = json.data ?? {};
          if (data.output) {
            console.log(chalk.dim('\n  Output:'));
            console.log(chalk.white(data.output));
          }
          if (data.exitCode !== undefined) {
            const exitIcon = data.exitCode === 0 ? chalk.green('✓') : chalk.red('✗');
            console.log(`  ${exitIcon} Exit code: ${data.exitCode}`);
          }
        } else {
          console.log(chalk.yellow(`  ⚠ Execute returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  scripts
    .command('delete <name>')
    .description('Delete a saved script')
    .action(async (name) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/scripts/delete`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name }),
          signal: AbortSignal.timeout(5000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Script '${name}' deleted`));
        } else {
          console.log(chalk.yellow(`  ⚠ Delete returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });
}
