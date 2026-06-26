import type { CLIType } from '../../types.js';
import { CLI_DEFINITIONS, splitCommandSpec } from '../cli-registry.js';
import { describe, test, expect, beforeEach } from 'vitest';

// Create a testable version of the CLI registry
class TestCLIRegistry {
  private tools: Map<string, any> = new Map();
  private customTools: Map<string, any> = new Map();

  registerCustomTool(tool: any): void {
    this.customTools.set(tool.name, tool);
  }

  unregisterCustomTool(name: string): boolean {
    return this.customTools.delete(name);
  }

  getTool(type: string): any | undefined {
    return this.tools.get(type);
  }

  getAvailableTools(): any[] {
    return [...this.tools.values(), ...this.customTools.values()].filter(t => t.available);
  }

  getAllTools(): any[] {
    return [...this.tools.values(), ...this.customTools.values()];
  }

  // Simulate setting a detected tool
  setTool(type: string, tool: any): void {
    this.tools.set(type, tool);
  }

  getServeCommand(type: string, port: number): { command: string; args: string[] } | null {
    const tool = this.tools.get(type);
    if (!tool || !tool.available) return null;

    const args = [...tool.args];
    const portArgIndex = args.findIndex((a: string) => a.includes('port'));
    if (portArgIndex >= 0) {
      args.splice(portArgIndex + 1, 0, String(port));
    } else {
      args.push(String(port));
    }

    return { command: tool.command, args };
  }

  getHealthEndpoint(type: string): string {
    const tool = this.tools.get(type);
    return tool?.healthEndpoint || '/health';
  }

  parseVersion(output: string): string {
    const match = output.match(/v?(\d+\.\d+\.\d+)/);
    return match ? match[1] : 'unknown';
  }
}

describe('CLIRegistry', () => {
  let registry: TestCLIRegistry;

  beforeEach(() => {
    registry = new TestCLIRegistry();
  });

  describe('custom tools', () => {
    test('registers custom tool', () => {
      const customTool = {
        type: 'custom',
        name: 'MyCustomCLI',
        command: 'my-cli',
        args: ['serve', '--port'],
        healthEndpoint: '/health',
        available: true,
        capabilities: ['chat'],
      };

      registry.registerCustomTool(customTool);
      
      expect(registry.getAllTools()).toHaveLength(1);
      expect(registry.getAvailableTools()).toHaveLength(1);
    });

    test('unregisters custom tool', () => {
      registry.registerCustomTool({
        type: 'custom',
        name: 'ToRemove',
        available: true,
      });

      expect(registry.getAllTools()).toHaveLength(1);
      
      const removed = registry.unregisterCustomTool('ToRemove');
      
      expect(removed).toBe(true);
      expect(registry.getAllTools()).toHaveLength(0);
    });

    test('returns false when unregistering non-existent tool', () => {
      const removed = registry.unregisterCustomTool('NonExistent');
      expect(removed).toBe(false);
    });
  });

  describe('getTool', () => {
    test('returns undefined for non-existent tool', () => {
      expect(registry.getTool('opencode')).toBeUndefined();
    });

    test('returns tool when exists', () => {
      const tool = { type: 'opencode', name: 'OpenCode', available: true };
      registry.setTool('opencode', tool);
      
      expect(registry.getTool('opencode')).toEqual(tool);
    });
  });

  describe('getAvailableTools', () => {
    test('filters out unavailable tools', () => {
      registry.setTool('opencode', { type: 'opencode', available: true });
      registry.setTool('claude', { type: 'claude', available: false });
      registry.setTool('aider', { type: 'aider', available: true });

      const available = registry.getAvailableTools();
      
      expect(available).toHaveLength(2);
      expect(available.every(t => t.available)).toBe(true);
    });

    test('includes available custom tools', () => {
      registry.registerCustomTool({ name: 'Custom1', available: true });
      registry.registerCustomTool({ name: 'Custom2', available: false });

      const available = registry.getAvailableTools();
      
      expect(available).toHaveLength(1);
    });
  });

  describe('getServeCommand', () => {
    test('returns null for unavailable tool', () => {
      registry.setTool('opencode', {
        type: 'opencode',
        command: 'opencode',
        args: ['serve', '--port'],
        available: false,
      });

      expect(registry.getServeCommand('opencode', 4000)).toBeNull();
    });

    test('returns command with port inserted', () => {
      registry.setTool('opencode', {
        type: 'opencode',
        command: 'opencode',
        args: ['serve', '--port'],
        available: true,
      });

      const result = registry.getServeCommand('opencode', 4000);
      
      expect(result).not.toBeNull();
      expect(result!.command).toBe('opencode');
      expect(result!.args).toContain('4000');
    });

    test('appends port when no port placeholder', () => {
      registry.setTool('test', {
        type: 'test',
        command: 'test-cli',
        args: ['serve'],
        available: true,
      });

      const result = registry.getServeCommand('test', 5000);
      
      expect(result!.args[result!.args.length - 1]).toBe('5000');
    });
  });

  describe('getHealthEndpoint', () => {
    test('returns tool health endpoint', () => {
      registry.setTool('opencode', {
        type: 'opencode',
        healthEndpoint: '/api/health',
      });

      expect(registry.getHealthEndpoint('opencode')).toBe('/api/health');
    });

    test('returns default /health for non-existent tool', () => {
      expect(registry.getHealthEndpoint('nonexistent')).toBe('/health');
    });
  });

  describe('parseVersion', () => {
    test('parses version with v prefix', () => {
      expect(registry.parseVersion('v1.2.3')).toBe('1.2.3');
    });

    test('parses version without v prefix', () => {
      expect(registry.parseVersion('1.2.3')).toBe('1.2.3');
    });

    test('extracts version from longer string', () => {
      expect(registry.parseVersion('opencode version v2.5.10 (2024)')).toBe('2.5.10');
    });

    test('returns unknown for invalid version', () => {
      expect(registry.parseVersion('no version here')).toBe('unknown');
    });
  });

  describe('splitCommandSpec', () => {
    test('splits simple single-word commands', () => {
      expect(splitCommandSpec('codex')).toEqual({ command: 'codex', args: [] });
    });

    test('splits command specs with subcommands', () => {
      expect(splitCommandSpec('gh copilot')).toEqual({ command: 'gh', args: ['copilot'] });
      expect(splitCommandSpec('python -m aider')).toEqual({ command: 'python', args: ['-m', 'aider'] });
    });

    test('preserves quoted executable paths', () => {
      expect(splitCommandSpec('"C:\\Program Files\\Tool\\tool.exe" serve')).toEqual({
        command: 'C:\\Program Files\\Tool\\tool.exe',
        args: ['serve'],
      });
    });
  });

  describe('definition parity', () => {
    test('covers every supported CLI type except custom', () => {
      const expectedTypes: CLIType[] = [
        'tormentnexus',
        'opencode',
        'antigravity',
        'claude',
        'aider',
        'cursor',
        'continue',
        'cody',
        'copilot',
        'adrenaline',
        'amazon-q',
        'amazon-q-developer',
        'amp-code',
        'auggie',
        'azure-openai',
        'bito',
        'byterover',
        'claude-code',
        'code-codex',
        'codebuff',
        'codemachine',
        'codex',
        'crush',
        'dolt',
        'factory',
        'factory-droid',
        'gemini',
        'goose',
        'grok',
        'jules',
        'kilo-code',
        'kimi',
        'llm',
        'litellm',
        'llamafile',
        'manus',
        'mistral-vibe',
        'ollama',
        'open-interpreter',
        'pi',
        'qwen-code',
        'rowboatx',
        'rovo',
        'shell-pilot',
        'smithery',
        'superai-cli',
        'trae',
        'warp',
      ];

      expect(CLI_DEFINITIONS.map((definition) => definition.type).sort()).toEqual(expectedTypes.sort());
    });
  });
});
