/**
 * `tormentnexus inventory` — Full system inventory from Go sidecar
 * Shows tools, harnesses, and detected capabilities
 */
import type { Command } from 'commander';

export function registerInventoryCommand(program: Command): void {
  program
    .command('inventory')
    .description('Full system inventory — tools, harnesses, and detected capabilities')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      console.log(chalk.bold.cyan('\n  TormentNexus System Inventory\n'));
      console.log(chalk.yellow('  Querying Go sidecar (may take 30s)...\n'));

      try {
        const res = await fetch('http://127.0.0.1:4300/api/cli/summary', {
          signal: AbortSignal.timeout(60000),
        });
        if (!res.ok) {
          console.log(chalk.red(`  ✗ Go sidecar returned ${res.status}`));
          return;
        }
        const json = await res.json();
        const data = json.data;

        if (opts.json) {
          console.log(JSON.stringify(data, null, 2));
          return;
        }

        // Tools
        console.log(chalk.bold('  Tools:'));
        console.log(`    Detected:  ${data.toolCount ?? 0}`);
        console.log(`    Available: ${data.availableToolCount ?? 0}`);
        if (data.availableTools?.length > 0) {
          console.log(chalk.dim('    Top tools:'));
          for (const t of data.availableTools.slice(0, 10)) {
            console.log(chalk.dim(`      ${t.name} ${t.version ?? ''} (${t.type ?? 'unknown'})`));
          }
        }

        // Harnesses
        console.log(chalk.bold('\n  Harnesses:'));
        console.log(`    Detected:  ${data.harnessCount ?? 0}`);
        console.log(`    Installed: ${data.installedHarnessCount ?? 0}`);
        console.log(`    Primary:   ${data.primaryHarness ?? 'none'}`);
        if (data.installedHarnesses?.length > 0) {
          console.log(chalk.dim('    Installed:'));
          for (const h of data.installedHarnesses.slice(0, 15)) {
            const name = typeof h === 'string' ? h : h.id ?? h.name ?? 'unknown';
            const desc = h.description ? chalk.dim(` — ${h.description}`) : '';
            console.log(chalk.dim(`      ✓ ${name}${desc}`));
          }
        }

        // MCP
        if (data.mcpServers !== undefined) {
          console.log(chalk.bold('\n  MCP:'));
          console.log(`    Servers: ${data.mcpServers ?? 0}`);
        }

        console.log('');
      } catch (e: any) {
        if (e.name === 'TimeoutError') {
          console.log(chalk.red('  ✗ Go sidecar timed out (60s)'));
        } else {
          console.log(chalk.red(`  ✗ Error: ${e.message}`));
        }
      }
    });
}
