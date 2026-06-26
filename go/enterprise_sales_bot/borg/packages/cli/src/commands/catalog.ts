/**
 * `tormentnexus catalog` — Browse the MCP server catalog
 * Queries the Go sidecar's local catalog database
 */
import type { Command } from 'commander';

export function registerCatalogCommand(program: Command): void {
  const catalog = program
    .command('catalog')
    .description('Browse the MCP server catalog (311+ servers from smithery.ai and GitHub)');

  catalog
    .command('list')
    .description('List all catalog entries')
    .option('-n, --limit <count>', 'Max entries to show', '20')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const isJson = opts.json === true;

      try {
        const limit = parseInt(opts.limit) || 20;
        const res = await fetch(`http://127.0.0.1:4300/api/catalog?limit=${limit}`, {
          signal: AbortSignal.timeout(8000),
        });
        if (!res.ok) {
          console.log(chalk.red(`  ✗ Go sidecar returned ${res.status}`));
          return;
        }
        const json = await res.json();
        const servers = json?.data?.servers ?? [];
        const total = json?.data?.total ?? servers.length;
        const shown = servers.slice(0, limit);

        if (isJson) {
          console.log(JSON.stringify(shown, null, 2));
          return;
        }

        console.log(chalk.bold.cyan(`\n  MCP Server Catalog (${total} entries, showing ${shown.length})\n`));

        for (const s of shown) {
          const name = s.display_name ?? s.canonical_id?.split('/').pop() ?? 'unknown';
          const desc = (s.description ?? '').substring(0, 60);
          const stars = s.stars ? `★ ${s.stars}` : '';
          console.log(`  ${chalk.cyan(name)} ${chalk.dim(stars)}`);
          if (desc) console.log(chalk.dim(`    ${desc}`));
        }

        if (total > limit) {
          console.log(chalk.dim(`\n  ... and ${total - limit} more (use --limit to show more)`));
        }
        console.log('');
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
        console.log(chalk.dim('    Is the Go sidecar running? Use tormentnexus start'));
      }
    });

  catalog
    .command('stats')
    .description('Show catalog statistics')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      try {
        const res = await fetch('http://127.0.0.1:4300/api/catalog/stats', {
          signal: AbortSignal.timeout(8000),
        });
        if (!res.ok) {
          console.log(chalk.red(`  ✗ Go sidecar returned ${res.status}`));
          return;
        }
        const json = await res.json();
        const data = json?.data ?? {};

        console.log(chalk.bold.cyan('\n  Catalog Statistics\n'));
        console.log(`  Total entries:     ${data.total ?? 'N/A'}`);
        console.log(`  Recently updated:  ${data.recentlyUpdated ?? 0}`);
        console.log(`  Broken:            ${data.broken ?? 0}`);

        if (data.statusCounts?.length) {
          console.log('\n  By status:');
          for (const s of data.statusCounts) {
            console.log(`    ${s.status}: ${s.count}`);
          }
        }
        console.log('');
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  catalog
    .command('search <query>')
    .description('Search the catalog by name, description, or category')
    .option('-n, --limit <count>', 'Max results', '10')
    .action(async (query, opts) => {
      const chalk = (await import('chalk')).default;

      try {
        const res = await fetch('http://127.0.0.1:4300/api/catalog?limit=340', {
          signal: AbortSignal.timeout(8000),
        });
        if (!res.ok) {
          console.log(chalk.red(`  ✗ Go sidecar returned ${res.status}`));
          return;
        }
        const json = await res.json();
        const servers = json?.data?.servers ?? [];
        const q = query.toLowerCase();

        const results = servers.filter((s: any) =>
          (s.display_name ?? '').toLowerCase().includes(q) ||
          (s.description ?? '').toLowerCase().includes(q) ||
          (s.author ?? '').toLowerCase().includes(q) ||
          (s.canonical_id ?? '').toLowerCase().includes(q) ||
          (s.tags ?? []).some((c: string) => c.toLowerCase().includes(q)) ||
          (s.categories ?? []).some((c: string) => c.toLowerCase().includes(q))
        );

        const limit = parseInt(opts.limit) || 10;
        const shown = results.slice(0, limit);

        console.log(chalk.bold.cyan(`\n  Catalog Search: "${query}" (${shown.length}/${results.length} results)\n`));

        if (shown.length === 0) {
          console.log(chalk.dim('  No matching entries found.\n'));
          return;
        }

        for (const s of shown) {
          const name = s.display_name ?? s.canonical_id?.split('/').pop() ?? 'unknown';
          const desc = (s.description ?? '').substring(0, 70);
          const stars = s.stars ? chalk.dim(`★ ${s.stars}`) : '';
          console.log(`  ${chalk.cyan(name)} ${stars}`);
          if (desc) console.log(chalk.dim(`    ${desc}`));
          if (s.repository_url) console.log(chalk.dim(`    Source: ${s.repository_url}`));
        }
        console.log('');
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  catalog
    .command('install <name>')
    .description('Install an MCP server from the catalog')
    .option('--yes', 'Skip confirmation')
    .action(async (name, opts) => {
      const chalk = (await import('chalk')).default;

      console.log(chalk.yellow(`  Installing '${name}' from catalog...`));

      try {
        const res = await fetch('http://127.0.0.1:4300/api/catalog/install', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, autoConfirm: opts.yes }),
          signal: AbortSignal.timeout(30000),
        });
        if (res.ok) {
          const json = await res.json();
          const data = json.data ?? {};
          console.log(chalk.green(`  ✓ '${name}' installed`));
          if (data.serverName) console.log(chalk.dim(`    Server: ${data.serverName}`));
          if (data.command) console.log(chalk.dim(`    Command: ${data.command}`));
          console.log(chalk.dim(`    Use 'tormentnexus mcp start ${name}' to connect`));
        } else {
          const json = await res.json().catch(() => ({}));
          console.log(chalk.yellow(`  ⚠ Install returned ${res.status}: ${json.error ?? 'unknown'}`));
          console.log(chalk.dim(`    Try: tormentnexus mcp install @modelcontextprotocol/server-${name}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });
}
