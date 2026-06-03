/**
 * `tormentnexus context` — Context harvesting and management
 */
import type { Command } from 'commander';

const TS_URL = 'http://127.0.0.1:4100/trpc';

export function registerContextCommand(program: Command): void {
  const ctx = program
    .command('context')
    .description('Context — manage codebase context harvesting for AI agents');

  ctx
    .command('harvest')
    .description('Run context harvest on the current workspace')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      try {
        const res = await fetch(`${TS_URL}/tormentnexusContext.harvest`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ json: {} }),
          signal: AbortSignal.timeout(30000),
        });
        if (res.ok) {
          const json = await res.json();
          const data = json?.result?.data ?? {};
          console.log(chalk.green('  ✓ Context harvested'));
          if (data.chunks) console.log(chalk.dim(`    Chunks: ${data.chunks}`));
        } else {
          console.log(chalk.yellow(`  ⚠ Harvest returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  ctx
    .command('stats')
    .description('Show context harvesting statistics')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let stats: any = {};
      try {
        const res = await fetch(`${TS_URL}/tormentnexusContext.getHarvestStats`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) stats = (await res.json())?.result?.data ?? {};
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify(stats, null, 2));
        return;
      }

      console.log(chalk.bold.cyan('\n  Context Harvesting\n'));
      console.log(chalk.dim('  Chunks:    ') + String(stats.chunks ?? stats.totalChunks ?? 0));
      console.log(chalk.dim('  Languages: ') + String(stats.languages ?? 0));
      console.log(chalk.dim('  Files:     ') + String(stats.files ?? 0));
      console.log(chalk.dim('  Size:      ') + String(stats.totalSize ?? '0 KB'));
      console.log('');
    });

  ctx
    .command('list')
    .description('List harvested context chunks')
    .option('-n, --limit <count>', 'Number of chunks to show', '20')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let chunks: any[] = [];
      try {
        const res = await fetch(`${TS_URL}/tormentnexusContext.getHarvestedContext`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) chunks = (await res.json())?.result?.data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ chunks }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Harvested Context (${chunks.length})\n`));
      if (chunks.length === 0) {
        console.log(chalk.dim('  No context harvested. Run `tormentnexus context harvest` first.\n'));
        return;
      }

      const limit = parseInt(opts.limit) || 20;
      for (const c of chunks.slice(0, limit)) {
        console.log(`  ${chalk.dim(c.type ?? 'chunk')} ${chalk.bold(c.path ?? c.id ?? '-')}`);
        if (c.summary) console.log(chalk.dim(`    ${c.summary.substring(0, 80)}`));
      }
      console.log('');
    });

  ctx
    .command('prompt')
    .description('Show the current system prompt with context')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      let prompt: string = '';
      try {
        const res = await fetch(`${TS_URL}/tormentnexusContext.getPrompt`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) prompt = (await res.json())?.result?.data ?? '';
      } catch {}

      console.log(chalk.bold.cyan('\n  System Prompt\n'));
      if (prompt) {
        console.log(prompt.substring(0, 500));
        if (prompt.length > 500) console.log(chalk.dim(`\n  ... (${prompt.length} chars total)`));
      } else {
        console.log(chalk.dim('  No prompt configured.\n'));
      }
      console.log('');
    });

  ctx
    .command('clear')
    .description('Clear all harvested context')
    .action(async () => {
      const chalk = (await import('chalk')).default;
      try {
        await fetch(`${TS_URL}/tormentnexusContext.clear`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ json: {} }),
          signal: AbortSignal.timeout(5000),
        });
        console.log(chalk.green('  ✓ Context cleared'));
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });
}
