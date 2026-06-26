import { Command } from 'commander';
import { describe, expect, it } from 'vitest';
import { withInheritedConfig } from '../src/cli/command-options.js';

describe('withInheritedConfig', () => {
  it('keeps an explicit command config', () => {
    const command = new Command('child');

    expect(withInheritedConfig({ config: 'child.yaml' }, command)).toEqual({ config: 'child.yaml' });
  });

  it('inherits config from an ancestor command', () => {
    const root = new Command('mvmt');
    root.option('-c, --config <path>');
    const parent = root.command('tunnel');
    const child = parent.command('logs');
    root.parse(['node', 'mvmt', '--config', 'root.yaml', 'tunnel', 'logs'], { from: 'node' });

    expect(withInheritedConfig({}, child)).toEqual({ config: 'root.yaml' });
  });
});
