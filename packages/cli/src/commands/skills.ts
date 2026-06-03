/**
 * `tormentnexus skills` — Manage and browse AI skills
 */
import type { Command } from 'commander';

const GO_URL = 'http://127.0.0.1:4300';
const TS_URL = 'http://127.0.0.1:4100/trpc';

export function registerSkillsCommand(program: Command): void {
  const skills = program
    .command('skills')
    .description('Skills — browse, read, and manage AI skill definitions');

  skills
    .command('list')
    .description('List all registered skills')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const Table = (await import('cli-table3')).default;

      let skillList: any[] = [];
      try {
        const res = await fetch(`${GO_URL}/api/skills/summary`, { signal: AbortSignal.timeout(5000) });
        if (res.ok) skillList = (await res.json()).data ?? [];
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ skills: skillList }, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  AI Skills (${skillList.length})\n`));
      if (skillList.length === 0) {
        console.log(chalk.dim('  No skills registered.\n'));
        return;
      }

      const table = new Table({
        head: ['Name', 'Folder', 'Description'],
        style: { head: ['cyan'] },
      });

      for (const s of skillList) {
        table.push([
          s.name ?? s.id ?? '-',
          s.folder ?? '-',
          (s.description ?? '').substring(0, 50),
        ]);
      }

      console.log(table.toString());
      console.log('');
    });

  skills
    .command('show <name>')
    .description('Show detailed information about a skill')
    .option('--json', 'Output as JSON')
    .action(async (name, opts) => {
      const chalk = (await import('chalk')).default;

      let skill: any = {};
      try {
        const res = await fetch(`${GO_URL}/api/skills/read?name=${encodeURIComponent(name)}`, {
          signal: AbortSignal.timeout(5000),
        });
        if (res.ok) skill = (await res.json()).data ?? {};
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify(skill, null, 2));
        return;
      }

      console.log(chalk.bold.cyan(`\n  Skill: ${name}\n`));
      if (!skill.content && !skill.name && !skill.id) {
        console.log(chalk.dim('  Skill not found.\n'));
        return;
      }

      if (skill.name || skill.id) {
        console.log(chalk.dim('  Name:        ') + (skill.name ?? skill.id ?? '-'));
        console.log(chalk.dim('  Folder:      ') + (skill.folder ?? '-'));
        if (skill.description) console.log(chalk.dim('  Description: ') + skill.description);
        if (skill.trigger) console.log(chalk.dim('  Trigger:     ') + skill.trigger);
      }

      if (Array.isArray(skill.content)) {
        console.log(chalk.dim('\n  Content:\n'));
        for (const part of skill.content) {
          if (part.type === 'text') {
            console.log(part.text);
          }
        }
      } else if (skill.instructions) {
        const instr = typeof skill.instructions === 'string' ? skill.instructions : JSON.stringify(skill.instructions);
        console.log(chalk.dim('\n  Instructions:\n'));
        console.log(instr);
      } else if (skill.content) {
        console.log(chalk.dim('\n  Content:\n'));
        console.log(skill.content);
      }
      console.log('');
    });

  skills
    .command('assimilate <name> <docs>')
    .description('Assimilate documentation into a skill')
    .action(async (name, docs) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/skills/assimilate`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ skillName: name, documentation: docs }),
          signal: AbortSignal.timeout(10000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Documentation assimilated into skill '${name}'`));
        } else {
          console.log(chalk.yellow(`  ⚠ Assimilation returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });

  skills
    .command('create <name>')
    .description('Create a new skill')
    .option('-d, --description <desc>', 'Skill description')
    .option('-t, --trigger <trigger>', 'Trigger pattern')
    .action(async (name, opts) => {
      const chalk = (await import('chalk')).default;
      try {
        const res = await fetch(`${GO_URL}/api/skills/create`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, description: opts.description, trigger: opts.trigger }),
          signal: AbortSignal.timeout(5000),
        });
        if (res.ok) {
          console.log(chalk.green(`  ✓ Skill '${name}' created`));
        } else {
          console.log(chalk.yellow(`  ⚠ Create returned ${res.status}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ Error: ${e.message}`));
      }
    });
}
