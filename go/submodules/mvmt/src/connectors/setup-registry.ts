import { MvmtConfig } from '../config/schema.js';
import { filesystemSetupDefinition } from './filesystem-setup.js';

export type ConnectorSetupId = 'filesystem';

export interface ConnectorSetupDefinition<TDetected, TInput, TId extends ConnectorSetupId = ConnectorSetupId> {
  id: TId;
  displayName: string;
  isAddable: boolean;
  detect(): Promise<TDetected>;
  prompt(detected: TDetected): Promise<TInput | undefined>;
  isConfigured(config: MvmtConfig): boolean;
  apply(config: MvmtConfig, input: TInput): MvmtConfig;
}

export type AnyConnectorSetupDefinition = ConnectorSetupDefinition<unknown, unknown>;

// Registry order controls the guided setup prompt order and connector status
// display order shown by the CLI.
const setupRegistry = [
  filesystemSetupDefinition,
] satisfies AnyConnectorSetupDefinition[];

export function getSetupRegistry(): AnyConnectorSetupDefinition[] {
  return [...setupRegistry];
}

export function getConnectorSetupDefinition(id: string): AnyConnectorSetupDefinition | undefined {
  return setupRegistry.find((definition) => definition.id === id);
}
