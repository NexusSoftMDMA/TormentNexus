/**
 * `tormentnexus metrics` — System metrics, performance, and observability
 */
import type { Command } from 'commander';

const GO_URL = 'http://127.0.0.1:4300';

export function registerMetricsCommand(program: Command): void {
  const metrics = program
    .command('metrics')
    .description('Metrics — system performance, provider stats, routing history');

  metrics
    .command('system')
    .description('Show real-time system resource snapshot')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let snapshot: any = {};
      try {
        const res = await fetch(`${GO_URL}/api/metrics/system-snapshot`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) snapshot = (await res.json()).data ?? {};
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify(snapshot, null, 2));
        return;
      }

      const sys = snapshot.system ?? {};
      const proc = snapshot.process ?? {};

      console.log(chalk.bold.cyan('\n  System Resources\n'));
      console.log(chalk.dim('  Host:      ') + `${sys.hostname ?? '-'} (${sys.platform ?? '-'} ${sys.arch ?? '-'})`);
      console.log(chalk.dim('  CPUs:      ') + `${sys.cpuCount ?? '-'} cores`);
      console.log(chalk.dim('  PID:       ') + String(proc.pid ?? '-'));
      console.log(chalk.dim('  Uptime:    ') + `${Math.floor((proc.uptimeSeconds ?? 0) / 3600)}h ${Math.floor(((proc.uptimeSeconds ?? 0) % 3600) / 60)}m`);
      if (proc.heapTotal) {
        const heapGB = (proc.heapTotal / 1024 / 1024 / 1024).toFixed(1);
        const usedMB = (proc.heapUsed / 1024 / 1024).toFixed(0);
        console.log(chalk.dim('  Heap:      ') + `${usedMB} MB used / ${heapGB} GB total`);
      }
      console.log('');
    });

  metrics
    .command('stats')
    .description('Show aggregated metrics statistics')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let stats: any = {};
      try {
        const res = await fetch(`${GO_URL}/api/metrics/stats`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) stats = (await res.json()).data ?? {};
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify(stats, null, 2));
        return;
      }

      console.log(chalk.bold.cyan('\n  Metrics Statistics\n'));
      console.log(chalk.dim('  Total events: ') + String(stats.totalEvents ?? 0));
      console.log(chalk.dim('  Window:       ') + `${Math.floor((stats.windowMs ?? 0) / 60000)} minutes`);

      if (stats.counts && Object.keys(stats.counts).length > 0) {
        console.log(chalk.dim('\n  Counts:'));
        for (const [key, value] of Object.entries(stats.counts)) {
          console.log(chalk.dim(`    ${key}: ${value}`));
        }
      }

      if (stats.averages && Object.keys(stats.averages).length > 0) {
        console.log(chalk.dim('\n  Averages:'));
        for (const [key, value] of Object.entries(stats.averages)) {
          console.log(chalk.dim(`    ${key}: ${typeof value === 'number' ? value.toFixed(2) : value}`));
        }
      }
      console.log('');
    });

  metrics
    .command('providers')
    .description('Show provider request, latency, and cost breakdowns')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let data: any = {};
      try {
        const res = await fetch(`${GO_URL}/api/metrics/provider-breakdown`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) data = (await res.json()).data ?? {};
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify(data, null, 2));
        return;
      }

      console.log(chalk.bold.cyan('\n  Provider Breakdown\n'));
      const providers = Array.isArray(data) ? data : data.providers ?? [];
      if (providers.length === 0) {
        console.log(chalk.dim('  No provider metrics recorded yet.\n'));
        return;
      }

      for (const p of providers) {
        console.log(`  ${chalk.bold(p.provider ?? p.name ?? '-')} — ${p.requests ?? 0} requests`);
        if (p.avgLatencyMs) console.log(chalk.dim(`    Avg latency: ${p.avgLatencyMs}ms`));
        if (p.cost) console.log(chalk.dim(`    Cost: $${p.cost}`));
      }
      console.log('');
    });

  metrics
    .command('routing')
    .description('Show recent LLM routing and failover decisions')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      let history: any[] = [];
      try {
        const res = await fetch(`${GO_URL}/api/metrics/routing-history`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) history = (await res.json()).data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ routingHistory: history }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Routing History (${history.length})\n`));
      if (history.length === 0) {
        console.log(chalk.dim('  No routing decisions recorded yet.\n'));
        return;
      }

      for (const h of history.slice(0, 20)) {
        const success = h.success ? chalk.green('✓') : chalk.red('✗');
        console.log(`  ${success} ${h.provider ?? '-'} → ${h.model ?? '-'} ${chalk.dim(`(${h.latencyMs ?? 0}ms)`)}`);
      }
      console.log('');
    });
}
