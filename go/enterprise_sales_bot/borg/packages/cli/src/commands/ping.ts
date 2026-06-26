/**
 * `tormentnexus ping` — Test connectivity to TS server and Go sidecar
 */
import type { Command } from 'commander';

export function registerPingCommand(program: Command): void {
  program
    .command('ping')
    .description('Test connectivity and latency to TS server and Go sidecar')
    .action(async () => {
      const chalk = (await import('chalk')).default;

      console.log(chalk.bold.cyan('\n  TormentNexus Connectivity Test\n'));

      // Ping TS server
      const tsStart = Date.now();
      try {
        const res = await fetch('http://127.0.0.1:4100/health', { signal: AbortSignal.timeout(3000) });
        const tsLatency = Date.now() - tsStart;
        if (res.ok) {
          const data = await res.json();
          console.log(`  TS Server:  ${chalk.green('● OK')}  ${chalk.dim(`${tsLatency}ms`)}  ${chalk.dim(`uptime: ${Math.floor(data.uptime / 60)}m`)}`);
        } else {
          console.log(`  TS Server:  ${chalk.red('✗ ' + res.status)}  ${chalk.dim(`${tsLatency}ms`)}`);
        }
      } catch (e: any) {
        const tsLatency = Date.now() - tsStart;
        console.log(`  TS Server:  ${chalk.red('✗ unreachable')}  ${chalk.dim(`${tsLatency}ms`)}`);
      }

      // Ping Go sidecar
      const goStart = Date.now();
      try {
        const res = await fetch('http://127.0.0.1:4300/health', { signal: AbortSignal.timeout(3000) });
        const goLatency = Date.now() - goStart;
        if (res.ok) {
          const data = await res.json();
          console.log(`  Go Sidecar: ${chalk.green('● OK')}  ${chalk.dim(`${goLatency}ms`)}  ${chalk.dim(`v${data.version}`)}`);
        } else {
          console.log(`  Go Sidecar: ${chalk.red('✗ ' + res.status)}  ${chalk.dim(`${goLatency}ms`)}`);
        }
      } catch {
        const goLatency = Date.now() - goStart;
        console.log(`  Go Sidecar: ${chalk.red('✗ unreachable')}  ${chalk.dim(`${goLatency}ms`)}`);
      }

      // Ping MCP data endpoint
      const mcpStart = Date.now();
      try {
        const res = await fetch('http://127.0.0.1:4100/trpc/mcp.getStatus', { signal: AbortSignal.timeout(3000) });
        const mcpLatency = Date.now() - mcpStart;
        if (res.ok) {
          const json = await res.json();
          const d = json?.result?.data;
          console.log(`  MCP Data:   ${chalk.green('● OK')}  ${chalk.dim(`${mcpLatency}ms`)}  ${chalk.dim(`${d?.serverCount ?? 0} servers, ${d?.toolCount ?? 0} tools`)}`);
        }
      } catch {
        console.log(`  MCP Data:   ${chalk.red('✗ timeout')}`);
      }

      console.log('');
    });
}
