import { confirm, select } from '@inquirer/prompts';
import chalk from 'chalk';
import { getConfigPath, loadConfig, saveConfig } from '../config/loader.js';
import { MvmtConfig } from '../config/schema.js';
import { getConnectorSetupDefinition, getSetupRegistry } from '../connectors/setup-registry.js';
import { resolveSetupPath } from '../connectors/setup-paths.js';

export interface ConnectorCommandOptions {
  config?: string;
}

export interface ConnectorSetupStatus {
  name: 'filesystem';
  displayName: string;
  configured: boolean;
  addable: boolean;
  note: string;
}

export function getConnectorSetupStatuses(config: MvmtConfig): ConnectorSetupStatus[] {
  return getSetupRegistry().map((definition) => ({
    name: definition.id,
    displayName: definition.displayName,
    configured: definition.isConfigured(config),
    addable: definition.isAddable,
    note: definition.isAddable
      ? `can be added with mvmt connectors add ${definition.id}`
      : 'configured by mvmt config setup',
  }));
}

export async function listConnectors(options: ConnectorCommandOptions = {}): Promise<void> {
  const config = loadConfig(options.config);
  const statuses = getConnectorSetupStatuses(config);

  console.log(chalk.bold('Connector setups'));
  for (const status of statuses) {
    const state = status.configured ? chalk.green('configured') : chalk.dim('not configured');
    console.log(`  ${status.displayName.padEnd(12)} ${state}  ${chalk.dim(status.note)}`);
  }
}

export async function addConnector(name?: string, options: ConnectorCommandOptions = {}): Promise<void> {
  const configPath = options.config ? resolveSetupPath(options.config) : getConfigPath();
  const config = loadConfig(configPath);
  const connectorName = name || (await promptForConnectorToAdd(config));
  const definition = getConnectorSetupDefinition(connectorName);

  if (!definition) {
    console.log(chalk.red(`Unknown connector setup: ${connectorName}`));
    return;
  }

  if (!definition.isAddable) {
    console.log(chalk.yellow(`Connector setup is not supported by this command yet: ${connectorName}`));
    console.log(chalk.dim('Run mvmt config setup to configure filesystem folders.'));
    return;
  }

  if (definition.isConfigured(config)) {
    const overwrite = await confirm({
      message: `${definition.displayName} is already configured. Replace its config?`,
      default: false,
    });
    if (!overwrite) {
      console.log(chalk.yellow(`${definition.displayName} config unchanged.`));
      return;
    }
  }

  const detected = await definition.detect();
  const input = await definition.prompt(detected);
  if (input === undefined) {
    console.log(chalk.yellow(`${definition.displayName} config unchanged.`));
    return;
  }

  const nextConfig = definition.apply(config, input);

  await saveConfig(configPath, nextConfig);
  console.log(chalk.green(`${definition.displayName} connector saved to ${configPath}`));
  console.log(chalk.dim('Restart mvmt for the new connector to load.'));
}
async function promptForConnectorToAdd(config: MvmtConfig): Promise<string> {
  const addable = getConnectorSetupStatuses(config).filter((status) => status.addable && !status.configured);
  if (addable.length === 0) {
    console.log(chalk.yellow('No supported unconfigured connector setups are available.'));
    console.log(chalk.dim('Run mvmt connectors list to see current connector setup status.'));
    process.exit(0);
  }

  return select({
    message: 'Which connector should mvmt add?',
    choices: addable.map((status) => ({
      name: status.displayName,
      value: status.name,
    })),
  });
}
