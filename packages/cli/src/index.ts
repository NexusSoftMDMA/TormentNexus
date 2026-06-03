#!/usr/bin/env node
/**
 * TormentNexus CLI - The Neural Operating System Command Interface
 * @module @tormentnexus/cli
 * @version 2.5.0
 *
 * Main entry point for the `tormentnexus` command. Provides comprehensive CLI access
 * to all TORMENTNEXUS subsystems: MCP router, memory, agents, sessions, providers,
 * tools, skills, configuration, and the web dashboard.
 */

import { Command } from 'commander';
import { readFileSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

import { registerStartCommand } from './commands/start.js';
import { registerStatusCommand } from './commands/status.js';
import { registerMcpCommand } from './commands/mcp.js';
import { registerMemoryCommand } from './commands/memory.js';
import { registerAgentCommand } from './commands/agent.js';
import { registerSessionCommand } from './commands/session.js';
import { registerProviderCommand } from './commands/provider.js';
import { registerToolsCommand } from './commands/tools.js';
import { registerConfigCommand } from './commands/config.js';
import { registerDashboardCommand } from './commands/dashboard.js';
import { registerTopCommand } from './commands/top.js';
import { registerHealthCommand } from './commands/health.js';
import { registerPingCommand } from './commands/ping.js';
import { registerCatalogCommand } from './commands/catalog.js';
import { registerInfoCommand } from './commands/info.js';
import { registerDoctorCommand } from './commands/doctor.js';
import { registerInventoryCommand } from './commands/inventory.js';
import { registerCloudCommand } from './commands/cloud.js';
import { registerBillingCommand } from './commands/billing.js';
import { registerContextCommand } from './commands/context.js';
import { registerKnowledgeCommand } from './commands/knowledge.js';
import { registerSwarmCommand } from './commands/swarm.js';
import { registerMetricsCommand } from './commands/metrics.js';
import { registerSkillsCommand } from './commands/skills.js';
import { registerUpgradeCommand } from './commands/upgrade.js';
import { registerPlanCommand } from './commands/plan.js';
import { registerBrowserCommand } from './commands/browser.js';
import { registerGitCommand } from './commands/git.js';
import { registerScriptsCommand } from './commands/scripts.js';

const __dirname = dirname(fileURLToPath(import.meta.url));

/** Read version from VERSION.md (single source of truth) */
function getVersion(): string {
  try {
    const versionPath = resolve(__dirname, '..', '..', '..', '..', '..', 'VERSION');
    return readFileSync(versionPath, 'utf-8').trim();
  } catch {
    return '1.0.0-alpha.39';
  }
}

const version = getVersion();

const program = new Command();

program
  .name('tormentnexus')
  .description('TormentNexus — The Neural Operating System / TORMENTNEXUS\n\nThe ultimate AI tool dashboard & development orchestrator.\nManage MCP servers, memory, agents, sessions, providers, and more.')
  .version(version, '-v, --version', 'Display the current TormentNexus version')
  .option('--json', 'Output results as JSON (applies to list/status commands)')
  .option('--config <path>', 'Path to tormentnexus config file', '~/.tormentnexus/config.jsonc')
  .option('--log-level <level>', 'Log level: debug, info, warn, error', 'info')
  .option('--no-color', 'Disable colored output');

// Register all command groups
registerStartCommand(program);
registerStatusCommand(program);
registerMcpCommand(program);
registerMemoryCommand(program);
registerAgentCommand(program);
registerSessionCommand(program);
registerProviderCommand(program);
registerToolsCommand(program);
registerConfigCommand(program);
registerDashboardCommand(program);
registerTopCommand(program);
registerHealthCommand(program);
registerPingCommand(program);
registerCatalogCommand(program);
registerInfoCommand(program);
registerDoctorCommand(program);
registerInventoryCommand(program);
registerCloudCommand(program);
registerBillingCommand(program);
registerContextCommand(program);
registerKnowledgeCommand(program);
registerSwarmCommand(program);
registerMetricsCommand(program);
registerSkillsCommand(program);
registerUpgradeCommand(program);
registerPlanCommand(program);
registerBrowserCommand(program);
registerGitCommand(program);
registerScriptsCommand(program);

// Default action: show help if no command given
program.action(() => {
  program.help();
});

program.parse(process.argv);
