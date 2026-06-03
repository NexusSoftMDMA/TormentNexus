/**
 * `tormentnexus swarm` — Multi-model swarm orchestration
 * Coordinate multiple AI models for debate, consensus, and task decomposition
 */
import type { Command } from 'commander';

const GO_URL = 'http://127.0.0.1:4300';
const TS_URL = 'http://127.0.0.1:4100/trpc';

export function registerSwarmCommand(program: Command): void {
  const swarm = program
    .command('swarm')
    .description('Swarm — multi-model orchestration (debate, consensus, tasks)');

  swarm
    .command('start <mission>')
    .description('Start a new swarm mission with a description')
    .option('-m, --models <models...>', 'Models to include in swarm')
    .option('-s, --strategy <strategy>', 'Orchestration strategy: debate, consensus, divide', 'debate')
    .action(async (mission, opts) => {
      const chalk = (await import('chalk')).default;
      console.log(chalk.yellow(`  Starting swarm mission...`));

      try {
        const res = await fetch(`${GO_URL}/api/swarm/start`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            mission,
            models: opts.models ?? [],
            strategy: opts.strategy,
          }),
          signal: AbortSignal.timeout(15000),
        });
        if (res.ok) {
          const json = await res.json();
          const data = json.data ?? json;
          console.log(chalk.green(`  ✓ Swarm mission started`));
          console.log(chalk.dim(`    ID:       ${data.id ?? data.missionId ?? 'pending'}`));
          console.log(chalk.dim(`    Strategy: ${opts.strategy}`));
          console.log(chalk.dim(`    Models:   ${(opts.models ?? []).join(', ') || 'auto'}`));
        } else {
          console.log(chalk.yellow(`  ⚠ Start returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  swarm
    .command('missions')
    .description('List swarm mission history')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const Table = (await import('cli-table3')).default;

      let missions: any[] = [];
      try {
        const res = await fetch(`${GO_URL}/api/swarm/missions`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) {
          const json = await res.json();
          const data = json.data ?? json;
          missions = Array.isArray(data) ? data : [];
        }
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ missions }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Swarm Missions (${missions.length})\n`));
      if (missions.length === 0) {
        console.log(chalk.dim('  No missions. Start one with `tormentnexus swarm start`.\n'));
        return;
      }

      for (const m of missions) {
        const status = m.status === 'active' || m.status === 'started' ? chalk.green('●') : m.status === 'completed' ? chalk.green('✓') : chalk.dim('○');
        const id = m.missionId ?? m.id ?? '-';
        const dateStr = m.createdAt ? chalk.dim(` [${new Date(m.createdAt).toLocaleTimeString()}]`) : '';
        console.log(`  ${status} ${id} ${chalk.dim(`(${m.status ?? 'unknown'})`)}${dateStr} ${(m.description ?? m.mission ?? '').substring(0, 60)}`);
      }
      console.log('');
    });

  swarm
    .command('debate <topic>')
    .description('Run a multi-model debate on a topic')
    .option('-r, --rounds <n>', 'Number of debate rounds', '3')
    .action(async (topic, opts) => {
      const chalk = (await import('chalk')).default;
      console.log(chalk.yellow(`  Starting swarm debate: "${topic.substring(0, 50)}..."`));

      try {
        const res = await fetch(`${GO_URL}/api/swarm/debate`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ topic, rounds: parseInt(opts.rounds) || 3 }),
          signal: AbortSignal.timeout(60000),
        });
        if (res.ok) {
          const json = await res.json();
          const data = json.data ?? json;
          console.log(chalk.green(`  ✓ Debate completed`));
          if (data.consensus) console.log(chalk.dim(`    Consensus: ${String(data.consensus).substring(0, 200)}`));
          if (data.participants) console.log(chalk.dim(`    Participants: ${data.participants}`));
        } else {
          console.log(chalk.yellow(`  ⚠ Debate returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  swarm
    .command('consensus <question>')
    .description('Seek multi-model consensus on a question')
    .action(async (question) => {
      const chalk = (await import('chalk')).default;
      console.log(chalk.yellow(`  Seeking consensus: "${question.substring(0, 50)}..."`));

      try {
        const res = await fetch(`${GO_URL}/api/swarm/consensus`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ question }),
          signal: AbortSignal.timeout(60000),
        });
        if (res.ok) {
          const json = await res.json();
          const data = json.data ?? json;
          console.log(chalk.green(`  ✓ Consensus reached`));
          if (data.answer) console.log(chalk.dim(`    Answer: ${String(data.answer).substring(0, 200)}`));
          if (data.confidence) console.log(chalk.dim(`    Confidence: ${data.confidence}`));
        } else {
          console.log(chalk.yellow(`  ⚠ Consensus returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  swarm
    .command('capabilities')
    .description('Show swarm mesh capabilities')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      let caps: any = {};
      try {
        const res = await fetch(`${GO_URL}/api/swarm/mesh-capabilities`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) caps = (await res.json()).data ?? {};
      } catch {}

      console.log(chalk.bold.cyan('\n  Swarm Mesh Capabilities\n'));
      const models = caps.models ?? caps.providers ?? [];
      if (Array.isArray(models) && models.length > 0) {
        for (const m of models) {
          const name = typeof m === 'string' ? m : m.name ?? m.model ?? '-';
          console.log(chalk.dim(`  ● ${name}`));
        }
      } else {
        console.log(chalk.dim('  No mesh capabilities configured.\n'));
      }
      console.log('');
    });

  swarm
    .command('risk')
    .description('Show swarm mission risk summary')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      let risk: any = {};
      try {
        const res = await fetch(`${GO_URL}/api/swarm/risk/summary`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) risk = (await res.json()).data ?? {};
      } catch {}

      console.log(chalk.bold.cyan('\n  Swarm Risk Summary\n'));
      console.log(chalk.dim('  Overall:  ') + String(risk.overall ?? 'N/A'));
      console.log(chalk.dim('  Missions: ') + String(risk.missionCount ?? 0));
      console.log(chalk.dim('  Risks:    ') + String(risk.riskCount ?? 0));
      console.log('');
    });
}
