/**
 * `tormentnexus knowledge` — Knowledge graph and resources
 */
import type { Command } from 'commander';

const TS_URL = 'http://127.0.0.1:4100/trpc';

export function registerKnowledgeCommand(program: Command): void {
  const knowledge = program
    .command('knowledge')
    .description('Knowledge — manage the knowledge graph and resources');

  knowledge
    .command('search <query>')
    .description('Search the knowledge graph')
    .option('-d, --depth <n>', 'Traversal depth', '1')
    .option('--json', 'Output as JSON')
    .action(async (query, opts) => {
      const chalk = (await import('chalk')).default;

      let result: any = {};
      try {
        const res = await fetch(`${TS_URL}/knowledge.query?input=${encodeURIComponent(JSON.stringify({ query, depth: parseInt(opts.depth) || 1 }))}`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) result = (await res.json())?.result?.data ?? {};
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify(result, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Knowledge Search: "${query}"\n`));
      const nodes = result.nodes ?? result.results ?? [];
      if (nodes.length === 0) {
        console.log(chalk.dim('  No results found.\n'));
        return;
      }
      for (const n of nodes) {
        console.log(`  ${chalk.bold(n.label ?? n.id ?? '-')} ${chalk.dim(n.type ?? '')}`);
        if (n.properties) {
          for (const [k, v] of Object.entries(n.properties).slice(0, 3)) {
            console.log(chalk.dim(`    ${k}: ${String(v).substring(0, 60)}`));
          }
        }
      }
      console.log('');
    });

  knowledge
    .command('stats')
    .description('Show knowledge graph statistics')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let stats: any = {};
      try {
        const res = await fetch(`${TS_URL}/knowledge.getStats`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) stats = (await res.json())?.result?.data ?? {};

        if (!stats.nodes && !stats.count) {
          const goRes = await fetch(`http://127.0.0.1:4300/api/knowledge/stats`, { signal: AbortSignal.timeout(3000) });
          if (goRes.ok) stats = (await goRes.json()).data ?? {};
        }
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify(stats, null, 2));
        return;
      }

      console.log(chalk.bold.cyan('\n  Knowledge Graph\n'));
      console.log(chalk.dim('  Nodes:    ') + String(stats.nodes ?? stats.nodeCount ?? stats.count ?? 0));
      console.log(chalk.dim('  Edges:    ') + String(stats.edges ?? stats.edgeCount ?? 0));
      console.log(chalk.dim('  Types:    ') + String(stats.types ?? 0));
      console.log(chalk.dim('  Size:     ') + String(stats.size ?? '0 KB'));
      console.log('');
    });

  knowledge
    .command('resources')
    .description('List knowledge resources')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let resources: any[] = [];
      try {
        const res = await fetch(`${TS_URL}/knowledge.getResources`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) {
          const raw = (await res.json())?.result?.data;
          resources = Array.isArray(raw) ? raw : [];
        }
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ resources }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Knowledge Resources (${resources.length})\n`));
      if (resources.length === 0) {
        console.log(chalk.dim('  No resources registered.\n'));
        return;
      }
      for (const r of resources) {
        console.log(`  ${chalk.bold(r.name ?? r.id ?? '-')} ${chalk.dim(r.type ?? '')}`);
      }
      console.log('');
    });
}
