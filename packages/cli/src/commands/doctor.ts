/**
 * `tormentnexus doctor` — Diagnose common issues and suggest fixes
 */
import type { Command } from 'commander';

export function registerDoctorCommand(program: Command): void {
  program
    .command('doctor')
    .description('Diagnose common issues and suggest fixes')
    .action(async () => {
      const chalk = (await import('chalk')).default;
      const { existsSync, readFileSync } = await import('fs');
      const { resolve } = await import('path');
      const { homedir } = await import('os');
      const home = homedir();

      console.log(chalk.bold.cyan('\n  TormentNexus Doctor — Diagnostics\n'));

      let issues = 0;
      let checks = 0;

      // Check 1: Server running
      checks++;
      try {
        const res = await fetch('http://127.0.0.1:4100/health', { signal: AbortSignal.timeout(2000) });
        if (res.ok) {
          console.log(chalk.green('  ✓') + ' Server is running');
        } else {
          issues++;
          console.log(chalk.red('  ✗') + ' Server returned ' + res.status);
          console.log(chalk.dim('    Fix: Run `tormentnexus start`'));
        }
      } catch {
        issues++;
        console.log(chalk.red('  ✗') + ' Server not running');
        console.log(chalk.dim('    Fix: Run `tormentnexus start`'));
      }

      // Check 2: Go sidecar
      checks++;
      try {
        const res = await fetch('http://127.0.0.1:4300/health', { signal: AbortSignal.timeout(2000) });
        if (res.ok) {
          console.log(chalk.green('  ✓') + ' Go sidecar is running');
        } else {
          issues++;
          console.log(chalk.yellow('  ⚠') + ' Go sidecar returned ' + res.status);
        }
      } catch {
        issues++;
        console.log(chalk.yellow('  ⚠') + ' Go sidecar not running (optional)');
      }

      // Check 3: At least one provider
      checks++;
      const providerKeys = ['OPENAI_API_KEY', 'ANTHROPIC_API_KEY', 'GOOGLE_API_KEY', 'DEEPSEEK_API_KEY'];
      const hasProvider = providerKeys.some(k => process.env[k]);
      if (hasProvider) {
        console.log(chalk.green('  ✓') + ' At least one provider API key configured');
      } else {
        issues++;
        console.log(chalk.red('  ✗') + ' No provider API keys detected');
        console.log(chalk.dim('    Fix: Set OPENAI_API_KEY or ANTHROPIC_API_KEY'));
      }

      // Check 4: Node.js version
      checks++;
      const nodeVersion = process.version;
      if (nodeVersion >= 'v20') {
        console.log(chalk.green('  ✓') + ` Node.js ${nodeVersion} (>= 20)`);
      } else {
        issues++;
        console.log(chalk.yellow('  ⚠') + ` Node.js ${nodeVersion} (recommend v20+)`);
      }

      // Check 5: pnpm
      checks++;
      try {
        const { execSync } = await import('child_process');
        const pnpmVersion = execSync('pnpm --version', { encoding: 'utf8' }).trim();
        console.log(chalk.green('  ✓') + ` pnpm v${pnpmVersion}`);
      } catch {
        issues++;
        console.log(chalk.red('  ✗') + ' pnpm not found');
        console.log(chalk.dim('    Fix: npm install -g pnpm'));
      }

      // Check 6: Config directory
      checks++;
      const configDir = resolve(home, '.tormentnexus');
      if (existsSync(configDir)) {
        console.log(chalk.green('  ✓') + ` Config dir exists (${configDir})`);
      } else {
        issues++;
        console.log(chalk.yellow('  ⚠') + ' Config dir missing');
        console.log(chalk.dim('    Fix: Run `tormentnexus config init --global`'));
      }

      // Check 7: mcp.jsonc
      checks++;
      const mcpJsonc = resolve(process.cwd(), 'mcp.jsonc');
      if (existsSync(mcpJsonc)) {
        console.log(chalk.green('  ✓') + ' mcp.jsonc found');
      } else {
        issues++;
        console.log(chalk.yellow('  ⚠') + ' mcp.jsonc not found in cwd');
        console.log(chalk.dim('    Fix: Run from the tormentnexus workspace root'));
      }

      // Check 8: VERSION file
      checks++;
      if (existsSync(resolve(process.cwd(), 'VERSION'))) {
        const version = readFileSync(resolve(process.cwd(), 'VERSION'), 'utf8').trim();
        console.log(chalk.green('  ✓') + ` VERSION file: ${version}`);
      } else {
        issues++;
        console.log(chalk.yellow('  ⚠') + ' VERSION file not found');
      }

      // Check 9: Go binary
      checks++;
      const goBin = resolve(process.cwd(), 'go', 'tormentnexus.exe');
      if (existsSync(goBin)) {
        console.log(chalk.green('  ✓') + ' Go binary exists');
      } else {
        const goBinAlt = resolve(process.cwd(), 'go', 'tormentnexus');
        if (existsSync(goBinAlt)) {
          console.log(chalk.green('  ✓') + ' Go binary exists');
        } else {
          issues++;
          console.log(chalk.yellow('  ⚠') + ' Go binary not built');
          console.log(chalk.dim('    Fix: cd go && go build ./cmd/tormentnexus'));
        }
      }

      // Check 10: MCP Fleet
      try {
        const { readdirSync: readDir, readFileSync: readFile } = await import('fs');
        const { join: joinPath } = await import('path');
        const { homedir: getHome } = await import('os');
        const pidDir = joinPath(getHome(), '.tormentnexus', 'mcp-pids');
        if (existsSync(pidDir)) {
          const pidFiles = readDir(pidDir).filter(f => f.endsWith('.pid'));
          let alive = 0;
          for (const pf of pidFiles) {
            try {
              const pid = parseInt(readFile(joinPath(pidDir, pf), 'utf8').trim());
              process.kill(pid, 0);
              alive++;
            } catch {}
          }
          checks++;
          if (alive > 0) {
            console.log(chalk.green('  ✓') + ` MCP Fleet: ${alive}/${pidFiles.length} servers alive`);
          } else {
            console.log(chalk.dim('  ○') + ' MCP Fleet: no servers spawned (use tormentnexus mcp connect-all)');
          }
        }
      } catch {}

      // Summary
      console.log(chalk.bold(`\n  ${checks} checks, ${issues} issues found`));
      if (issues === 0) {
        console.log(chalk.green('  All systems healthy!\n'));
      } else {
        console.log(chalk.yellow(`  Run the suggested fixes above to resolve.\n`));
      }
    });
}
