/**
 * `tormentnexus health` — Detailed subsystem readiness check
 * Shows startup status, readiness checks, and blocking reasons
 */
import type { Command } from 'commander';

interface CheckResult {
  ready: boolean;
  [key: string]: any;
}

interface StartupStatus {
  status: string;
  ready: boolean;
  uptime: number;
  blockingReasons: { code: string; detail: string }[];
  checks: Record<string, CheckResult>;
}

export function registerHealthCommand(program: Command): void {
  program
    .command('health')
    .description('Show detailed subsystem readiness checks and blocking reasons')
    .option('--json', 'Output raw JSON from startupStatus')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;

      try {
        const res = await fetch('http://127.0.0.1:4100/trpc/startupStatus', {
          signal: AbortSignal.timeout(5000),
        });

        if (!res.ok) {
          console.log(chalk.red(`  ✗ Server returned ${res.status}`));
          return;
        }

        const json = await res.json();
        const data: StartupStatus = json?.result?.data;

        if (!data) {
          console.log(chalk.red('  ✗ No data returned from server'));
          return;
        }

        if (opts.json) {
          console.log(JSON.stringify(data, null, 2));
          return;
        }

        const h = Math.floor(data.uptime / 3600);
        const m = Math.floor((data.uptime % 3600) / 60);
        const s = Math.floor(data.uptime % 60);

        console.log(chalk.bold.cyan(`\n  ⬡ TormentNexus TORMENTNEXUS — Health Check\n`));
        console.log(`  Overall: ${data.ready ? chalk.green('● Ready') : chalk.yellow('◐ Partial')}`);
        console.log(`  Uptime:  ${h}h ${m}m ${s}s`);

        // Readiness checks
        console.log(chalk.bold('\n  Subsystems:\n'));
        const checkNames: Record<string, string> = {
          mcpAggregator: 'MCP Aggregator',
          configSync: 'Config Sync',
          memory: 'Memory Manager',
          claudeMem: 'Claude Memory',
          browser: 'Browser Control',
          sessionSupervisor: 'Session Supervisor',
          extensionBridge: 'Extension Bridge',
          executionEnvironment: 'Execution Environment',
        };

        let ready = 0, total = 0;
        for (const [key, label] of Object.entries(checkNames)) {
          const check = data.checks[key];
          if (!check) continue;
          total++;
          if (check.ready) {
            ready++;
            console.log(`  ${chalk.green('✓')} ${label}`);
          } else {
            console.log(`  ${chalk.red('✗')} ${label} ${chalk.dim('— not ready')}`);
          }
        }
        console.log(chalk.dim(`\n  ${ready}/${total} subsystems ready`));

        // Execution environment details
        const ee = data.checks.executionEnvironment;
        if (ee && ee.ready) {
          console.log(chalk.bold('\n  Execution Environment:\n'));
          console.log(`  Shells:    ${ee.verifiedShellCount}/${ee.shellCount} verified`);
          console.log(`  Harnesses: ${ee.verifiedHarnessCount}/${ee.harnessCount} verified`);
          console.log(`  Tools:     ${ee.verifiedToolCount}/${ee.toolCount} verified`);
          console.log(`  Preferred: ${ee.preferredShellLabel}`);
          if (ee.notes?.length) {
            for (const note of ee.notes) {
              console.log(chalk.dim(`  Note: ${note}`));
            }
          }
        }

        // Blocking reasons
        if (data.blockingReasons?.length > 0) {
          console.log(chalk.bold('\n  Blocking Reasons:\n'));
          for (const r of data.blockingReasons) {
            console.log(`  ${chalk.yellow('⚠')} ${r.code}`);
            console.log(chalk.dim(`    ${r.detail}`));
          }
        }

        console.log('');

      } catch (e: any) {
        if (e.name === 'AbortError' || e.code === 'ECONNREFUSED') {
          console.log(chalk.red('\n  ✗ Server not running\n'));
          console.log(chalk.dim('  Use tormentnexus start to launch the server\n'));
        } else {
          console.log(chalk.red(`\n  ✗ Error: ${e.message}\n`));
        }
      }
    });
}
