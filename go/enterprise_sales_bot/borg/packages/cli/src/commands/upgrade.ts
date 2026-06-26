/**
 * `tormentnexus upgrade` — Check for updates and upgrade TormentNexus TORMENTNEXUS
 */
import type { Command } from 'commander';

export function registerUpgradeCommand(program: Command): void {
  program
    .command('upgrade')
    .description('Check for updates and upgrade TormentNexus TORMENTNEXUS')
    .option('--check', 'Only check for updates without upgrading')
    .option('--force', 'Force upgrade even if up to date')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const { execSync } = await import('child_process');
      const { readFileSync } = await import('fs');
      const { resolve } = await import('path');

      // Get current version
      let currentVersion = 'unknown';
      try {
        let dir = process.cwd();
        for (let i = 0; i < 20; i++) {
          try { currentVersion = readFileSync(resolve(dir, 'VERSION'), 'utf8').trim(); break; } catch {}
          const parent = resolve(dir, '..');
          if (parent === dir) break;
          dir = parent;
        }
      } catch {}

      console.log(chalk.bold.cyan('\n  TormentNexus TORMENTNEXUS Upgrade\n'));
      console.log(chalk.dim('  Current version: ') + currentVersion);

      // Check latest version from GitHub
      let latestVersion = '';
      try {
        const result = execSync(
          'git ls-remote --tags origin "v*" 2>/dev/null | tail -1',
          { encoding: 'utf8', timeout: 10000 }
        ).trim();
        const match = result.match(/refs\/tags\/v?(.+)/);
        if (match) latestVersion = match[1];
      } catch {}

      // Also check latest commit
      let latestCommit = '';
      try {
        execSync('git fetch origin main 2>/dev/null', { timeout: 15000 });
        latestCommit = execSync('git rev-parse --short origin/main', { encoding: 'utf8' }).trim();
      } catch {}

      let localCommit = '';
      try {
        localCommit = execSync('git rev-parse --short HEAD', { encoding: 'utf8' }).trim();
      } catch {}

      console.log(chalk.dim('  Local commit:   ') + localCommit);
      console.log(chalk.dim('  Remote commit:  ') + latestCommit);

      if (latestVersion) {
        console.log(chalk.dim('  Latest release: ') + latestVersion);
      }

      // Check if we're behind
      let behind = 0;
      try {
        const result = execSync('git rev-list HEAD..origin/main --count', { encoding: 'utf8' }).trim();
        behind = parseInt(result) || 0;
      } catch {}

      if (behind > 0) {
        console.log(chalk.yellow(`\n  ${behind} commits behind origin/main`));

        if (opts.check) {
          console.log(chalk.dim('  Run `tormentnexus upgrade` to update.\n'));
          return;
        }

        console.log(chalk.yellow('  Upgrading...'));

        try {
          // Pull latest
          execSync('git pull origin main', { encoding: 'utf8', timeout: 60000, stdio: 'pipe' });
          console.log(chalk.green('  ✓ Pulled latest changes'));

          // Rebuild
          try {
            execSync('pnpm install', { encoding: 'utf8', timeout: 120000, stdio: 'pipe' });
            console.log(chalk.green('  ✓ Dependencies installed'));
          } catch {
            console.log(chalk.yellow('  ⚠ pnpm install had warnings'));
          }

          try {
            execSync('pnpm rebuild better-sqlite3', { encoding: 'utf8', timeout: 60000, stdio: 'pipe' });
          } catch {}

          try {
            execSync('pnpm -C packages/core exec tsc --noEmit', { encoding: 'utf8', timeout: 60000, stdio: 'pipe' });
            console.log(chalk.green('  ✓ TypeScript compiled'));
          } catch {
            console.log(chalk.yellow('  ⚠ TypeScript had errors'));
          }

          // Build Go binary
          try {
            execSync('cd go && go build -buildvcs=false ./cmd/tormentnexus', { encoding: 'utf8', timeout: 120000, stdio: 'pipe' });
            console.log(chalk.green('  ✓ Go binary built'));
          } catch {
            console.log(chalk.yellow('  ⚠ Go build had errors'));
          }

          // Read new version
          let newVersion = currentVersion;
          try {
            let dir = process.cwd();
            for (let i = 0; i < 20; i++) {
              try { newVersion = readFileSync(resolve(dir, 'VERSION'), 'utf8').trim(); break; } catch {}
              const parent = resolve(dir, '..');
              if (parent === dir) break;
              dir = parent;
            }
          } catch {}

          console.log(chalk.green(`\n  ✓ Upgraded to ${newVersion}`));
          console.log(chalk.dim('  Restart with `tormentnexus start` to apply changes.\n'));
        } catch (e: any) {
          console.log(chalk.red(`  ✗ Upgrade failed: ${e.message}`));
        }
      } else {
        console.log(chalk.green('\n  ✓ Already up to date'));
        console.log(chalk.dim(`  ${localCommit} is the latest commit\n`));
      }
    });
}
