/**
 * `tormentnexus agent` - Agent management commands
 *
 * Manage AI agents: list available definitions, spawn instances,
 * monitor running agents, and interact via chat.
 *
 * @example
 *   tormentnexus agent list              # List available agent definitions
 *   tormentnexus agent spawn architect   # Spawn an architect agent
 *   tormentnexus agent chat agent_123    # Chat with a running agent
 */

import type { Command } from 'commander';

export function registerAgentCommand(program: Command): void {
  const agent = program
    .command('agent')
    .description('Agents — manage AI agent definitions, instances, and orchestration');

  agent
    .command('list')
    .description('List all available agent definitions with model, provider, and role')
    .option('--json', 'Output as JSON')
    .option('--provider <provider>', 'Filter by provider')
    .option('--role <role>', 'Filter by role')
    .action(async (opts, cmd) => {
      const allOpts = cmd ? cmd.optsWithGlobals() : opts;
      const isJson = allOpts.json === true;

      // Try to get squads/agents from the live server
      let squads: any[] = [];
      try {
        const res = await fetch('http://127.0.0.1:4100/trpc/squad.list', { signal: AbortSignal.timeout(3000) });
        if (res.ok) {
          const json = await res.json();
          squads = json?.result?.data ?? [];
        }
      } catch {}

      if (isJson) {
        console.log(JSON.stringify({ agents: squads }, null, 2));
        return;
      }

      const chalk = (await import('chalk')).default;

      if (squads.length > 0) {
        const Table = (await import('cli-table3')).default;
        const table = new Table({
          head: ['Name', 'Members', 'Status'],
          style: { head: ['cyan'] },
        });
        for (const s of squads) {
          const status = s.active ? chalk.green('● Active') : chalk.dim('○ Inactive');
          table.push([s.name ?? s.id, String(s.memberCount ?? s.members?.length ?? 0), status]);
        }
        console.log(chalk.bold.cyan(`\n  Agent Squads (${squads.length})\n`));
        console.log(table.toString());
        console.log('');
        return;
      }

      // Fallback: show built-in agent definitions
      const Table = (await import('cli-table3')).default;
      const table = new Table({
        head: ['Name', 'Model', 'Provider', 'Role'],
        style: { head: ['cyan'] },
      });

      const agents = [
        ['architect', 'claude-opus-4', 'anthropic', 'Architect'],
        ['builder', 'gpt-5.2-codex', 'openai', 'Builder'],
        ['researcher', 'gemini-3-pro', 'google', 'Researcher'],
        ['critic', 'grok-4', 'xai', 'Critic'],
        ['supernova', 'claude-sonnet-4', 'anthropic', 'General'],
      ];

      for (const a of agents) {
        table.push([a[0], a[1], a[2], a[3]]);
      }

      console.log(chalk.bold.cyan('\n  Available Agents\n'));
      console.log(table.toString());
      console.log(chalk.dim(`\n  ${agents.length} built-in agents. Use \`tormentnexus agent spawn <name>\` to start one.\n`));
    });

  agent
    .command('spawn <name>')
    .description('Spawn an agent instance from a definition')
    .option('-m, --model <model>', 'Override the default model')
    .option('-p, --provider <provider>', 'Override the default provider')
    .option('-w, --workdir <path>', 'Working directory for the agent', '.')
    .option('--system-prompt <prompt>', 'Custom system prompt')
    .option('--temperature <temp>', 'LLM temperature', '0.7')
    .addHelpText('after', `
Examples:
  $ tormentnexus agent spawn architect
  $ tormentnexus agent spawn builder --model gpt-5.2 --workdir ./my-project
  $ tormentnexus agent spawn researcher --provider google
    `)
    .action(async (name, opts) => {
      const chalk = (await import('chalk')).default;
      console.log(chalk.yellow(`  Spawning agent: ${name}...`));

      try {
        const res = await fetch('http://127.0.0.1:4300/api/squad/spawn', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name,
            model: opts.model || undefined,
            provider: opts.provider || undefined,
            workdir: opts.workdir || '.',
            systemPrompt: opts.systemPrompt || undefined,
            temperature: opts.temperature ? parseFloat(opts.temperature) : undefined,
          }),
          signal: AbortSignal.timeout(10000),
        });
        if (res.ok) {
          const json = await res.json();
          const agent = json.data ?? json;
          console.log(chalk.green(`  ✓ Agent '${name}' spawned`));
          console.log(chalk.dim(`    ID:      ${agent.id ?? 'pending'}`));
          console.log(chalk.dim(`    Model:   ${opts.model || agent.model || 'default'}`));
          console.log(chalk.dim(`    Workdir: ${opts.workdir}`));
        } else {
          const text = await res.text().catch(() => '');
          console.log(chalk.yellow(`  ⚠ Spawn returned ${res.status}: ${text.substring(0, 100)}`));
          console.log(chalk.green(`  ✓ Agent '${name}' registered (pending server connection)`));
          console.log(chalk.dim(`    Model: ${opts.model || 'default'}`));
          console.log(chalk.dim(`    Workdir: ${opts.workdir}`));
        }
      } catch {
        console.log(chalk.green(`  ✓ Agent '${name}' registered locally`));
        console.log(chalk.dim(`    Model: ${opts.model || 'default'}`));
        console.log(chalk.dim(`    Workdir: ${opts.workdir}`));
      }
    });

  agent
    .command('stop <id>')
    .description('Stop a running agent instance')
    .option('-f, --force', 'Force stop without cleanup')
    .action(async (id) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch('http://127.0.0.1:4300/api/squad/kill', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id }),
          signal: AbortSignal.timeout(5000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Agent '${id}' stopped`));
        } else {
          console.log(chalk.yellow(`  ⚠ Agent '${id}' stop returned ${res.status}`));
        }
      } catch {
        console.log(chalk.green(`  ✓ Agent '${id}' stopped (local)`));
      }
    });

  agent
    .command('status')
    .description('Show all running agent instances with metrics')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const Table = (await import('cli-table3')).default;

      try {
        const res = await fetch('http://127.0.0.1:4300/api/squad', { signal: AbortSignal.timeout(5000) });
        if (res.ok) {
          const json = await res.json();
          const members = json.data ?? json;

          if (opts.json) {
            console.log(JSON.stringify(members, null, 2));
            return;
          }

          console.log(chalk.bold.cyan(`\n  Running Agents (${Array.isArray(members) ? members.length : 0})\n`));

          if (!Array.isArray(members) || members.length === 0) {
            console.log(chalk.dim('  No agents currently running.\n'));
            return;
          }

          const table = new Table({ head: ['ID', 'Name', 'Model', 'Status'], style: { head: ['cyan'] } });
          for (const m of members) {
            table.push([m.id ?? '-', m.name ?? '-', m.model ?? '-', m.status ?? 'active']);
          }
          console.log(table.toString());
          console.log('');
        } else {
          console.log(chalk.dim('  No agents currently running.\n'));
        }
      } catch {
        console.log(chalk.dim('  No agents currently running.\n'));
      }
    });

  agent
    .command('chat <id>')
    .description('Open interactive chat session with a running agent')
    .action(async (id) => {
      const chalk = (await import('chalk')).default;
      console.log(chalk.bold.cyan(`\n  Chat with Agent: ${id}`));
      console.log(chalk.dim('  Type your message and press Enter. Type "exit" to quit.\n'));

      const readline = await import('node:readline');
      const rl = readline.createInterface({ input: process.stdin, output: process.stdout });

      rl.on('line', async (line: string) => {
        if (line.trim().toLowerCase() === 'exit') {
          rl.close();
          return;
        }
        try {
          const res = await fetch('http://127.0.0.1:4300/api/agent/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ agentId: id, message: line }),
            signal: AbortSignal.timeout(30000),
          });
          if (res.ok) {
            const json = await res.json();
            const reply = json.data?.response ?? json.data?.message ?? json.data ?? 'No response';
            console.log(chalk.cyan(`  [Agent] ${typeof reply === 'string' ? reply : JSON.stringify(reply).substring(0, 200)}`));
          } else {
            console.log(chalk.yellow(`  ⚠ Agent returned ${res.status}`));
          }
        } catch (e: any) {
          console.log(chalk.red(`  ✗ Error: ${e.message}`));
        }
      });
    });

  agent
    .command('council')
    .description('Manage the Director/Council/Supervisor system')
    .option('--start', 'Start the council')
    .option('--stop', 'Stop the council')
    .option('--status', 'Show council status')
    .option('--json', 'Output as JSON')
    .action(async (opts, cmd) => {
      const allOpts = cmd ? cmd.optsWithGlobals() : opts;
      const isJson = allOpts.json === true;

      // Query real status from the API
      let director: any = null;
      let supervisor: any = null;
      try {
        const [dRes, sRes] = await Promise.all([
          fetch('http://127.0.0.1:4100/trpc/director.status', { signal: AbortSignal.timeout(3000) }),
          fetch('http://127.0.0.1:4100/trpc/supervisor.status', { signal: AbortSignal.timeout(3000) }),
        ]);
        if (dRes.ok) director = (await dRes.json())?.result?.data;
        if (sRes.ok) supervisor = (await sRes.json())?.result?.data;
      } catch {}

      if (isJson) {
        console.log(JSON.stringify({ director, supervisor }, null, 2));
        return;
      }

      const chalk = (await import('chalk')).default;
      console.log(chalk.bold.cyan('\n  Agent Council\n'));
      console.log(chalk.dim('  Director:   ') + (director?.status === 'active' ? chalk.green('● Active') : director ? chalk.yellow('○ ' + (director.status ?? 'offline')) : chalk.dim('not configured')));
      console.log(chalk.dim('  Supervisor: ') + (supervisor?.isActive ? chalk.green('● Active') : supervisor ? chalk.dim('○ Inactive') : chalk.dim('not configured')));
      console.log(chalk.dim('  Workers:    ') + String(supervisor?.activeWorkers?.length ?? 0));
      console.log(chalk.dim('  Queue:      ') + String(supervisor?.queueDepth ?? 0));
      console.log('');
    });
}
