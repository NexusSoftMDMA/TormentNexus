import { Command } from 'commander';

export function withInheritedConfig<T extends { config?: string }>(options: T, command: Command): T {
  const config = options.config ?? inheritedConfig(command);
  return config ? { ...options, config } : options;
}

function inheritedConfig(command: Command): string | undefined {
  let current = command.parent;
  while (current) {
    const config = current.opts<{ config?: string }>().config;
    if (config) return config;
    current = current.parent;
  }
  return undefined;
}
