import { confirm, input, select } from '@inquirer/prompts';
import chalk from 'chalk';
import { getConfigPath, saveConfig } from '../config/loader.js';
import {
  MvmtConfig,
  TunnelConfig,
} from '../config/schema.js';
import { getSetupRegistry } from '../connectors/setup-registry.js';
import { pathExists, resolveSetupPath } from '../connectors/setup-paths.js';
import { promptForTunnelConfig } from './tunnel.js';

export interface SetupConfigOptions {
  config?: string;
  promptOnOverwrite?: boolean;
  printNextStep?: boolean;
}

export async function setupConfig(
  options: SetupConfigOptions = {},
): Promise<{ config: MvmtConfig; configPath: string } | undefined> {
  printBanner();

  const configPath = options.config ? resolveSetupPath(options.config) : getConfigPath();
  if (options.promptOnOverwrite !== false && await pathExists(configPath)) {
    const overwrite = await confirm({
      message: `Config already exists at ${configPath}. Overwrite?`,
      default: false,
    });

    if (!overwrite) {
      console.log(chalk.yellow('Aborted. Existing config unchanged.'));
      return;
    }
  }

  console.log('Configure the local data mvmt is allowed to expose.\n');

  console.log('Checking available setup options...\n');
  const setupRegistry = getSetupRegistry();
  const detections = new Map(
    await Promise.all(
      setupRegistry.map(async (definition) => [definition.id, await definition.detect()] as const),
    ),
  );

  printAvailableConnectors();

  const applySelections: Array<(config: MvmtConfig) => MvmtConfig> = [];
  for (const definition of setupRegistry) {
    const detected = detections.get(definition.id);
    const input = await definition.prompt(detected);
    if (input !== undefined) {
      applySelections.push((config) => definition.apply(config, input));
    }
  }

  const portAnswer = await input({
    message: 'Server port',
    default: '4141',
    validate: (value) => {
      const port = Number(value);
      return Number.isInteger(port) && port > 0 && port <= 65535 ? true : 'Enter a port from 1 to 65535';
    },
  });
  const port = Number(portAnswer);
  const access = await promptForAccess(port);

  let config = buildConfig({ port, access });
  for (const applySelection of applySelections) {
    config = applySelection(config);
  }

  await saveConfig(configPath, config);

  console.log(chalk.green(`\nConfig saved to ${configPath}`));
  if (options.printNextStep !== false) {
    console.log(`\nNext: run ${chalk.cyan('mvmt serve')}\n`);
  } else {
    console.log('');
  }

  return { config, configPath };
}

export async function init(options: SetupConfigOptions = {}): Promise<void> {
  await setupConfig(options);
}

export interface BuildConfigInput {
  port: number;
  access?: { access: 'local' | 'tunnel'; tunnel?: TunnelConfig };
}

export function buildConfig(input: BuildConfigInput): MvmtConfig {
  const access = input.access ?? { access: 'local' };
  return {
    version: 1,
    server: {
      port: input.port,
      allowedOrigins: [],
      access: access.access,
      ...(access.tunnel ? { tunnel: access.tunnel } : {}),
    },
    proxy: [],
    mounts: [],
  };
}

async function promptForAccess(port: number): Promise<{ access: 'local' | 'tunnel'; tunnel?: TunnelConfig }> {
  const access = await select<'local' | 'tunnel'>({
    message: 'How should mvmt be accessible?',
    choices: [
      { name: `Local only (localhost:${port})`, value: 'local' },
      { name: 'Tunnel (get a public URL)', value: 'tunnel' },
    ],
  });

  if (access === 'local') return { access: 'local' };

  const tunnel = await promptForTunnelConfig(port);
  return {
    access: 'tunnel',
    tunnel,
  };
}

function printBanner(): void {
  console.log(`\n${chalk.bold('mvmt')} - local MCP hub\n`);
}

function printAvailableConnectors(): void {
  console.log('Available setup options:');
  console.log('  - Local folder mounts');
  console.log('');
}
