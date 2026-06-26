import { describe, test, expect, beforeEach } from 'vitest';

type CLIType = 'opencode' | 'claude' | 'aider' | 'cursor' | 'continue' | 'cody' | 'copilot' | 'custom';

interface EnvironmentConfig {
  inherit: boolean;
  variables: Record<string, string>;
  secrets: string[];
  passthrough: string[];
}

class TestEnvironmentManager {
  private sessions: Map<string, {
    sessionId: string;
    cliType: CLIType;
    config: EnvironmentConfig;
    resolved: Record<string, string>;
  }> = new Map();
  private globalOverrides: Record<string, string> = {};
  private globalSecrets: Set<string> = new Set();

  setGlobalOverride(key: string, value: string): void {
    this.globalOverrides[key] = value;
  }

  removeGlobalOverride(key: string): void {
    delete this.globalOverrides[key];
  }

  addGlobalSecret(key: string): void {
    this.globalSecrets.add(key);
  }

  createSessionEnvironment(
    sessionId: string,
    cliType: CLIType,
    config?: Partial<EnvironmentConfig>
  ): Record<string, string> {
    const envConfig: EnvironmentConfig = {
      inherit: config?.inherit ?? true,
      variables: config?.variables ?? {},
      secrets: config?.secrets ?? [],
      passthrough: config?.passthrough ?? [],
    };

    const resolved = this.resolveEnvironment(cliType, envConfig);

    this.sessions.set(sessionId, {
      sessionId,
      cliType,
      config: envConfig,
      resolved,
    });

    return resolved;
  }

  private resolveEnvironment(cliType: CLIType, config: EnvironmentConfig): Record<string, string> {
    const result: Record<string, string> = {};

    for (const [key, value] of Object.entries(this.globalOverrides)) {
      result[key] = value;
    }

    for (const [key, value] of Object.entries(config.variables)) {
      result[key] = this.expandVariables(value, result);
    }

    return result;
  }

  private expandVariables(value: string, env: Record<string, string>): string {
    return value.replace(/\$\{(\w+)\}/g, (_, varName) => {
      return env[varName] || process.env[varName] || '';
    });
  }

  getSessionEnvironment(sessionId: string): Record<string, string> | undefined {
    return this.sessions.get(sessionId)?.resolved;
  }

  updateSessionVariable(sessionId: string, key: string, value: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    session.config.variables[key] = value;
    session.resolved[key] = this.expandVariables(value, session.resolved);
  }

  removeSessionVariable(sessionId: string, key: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    delete session.config.variables[key];
    delete session.resolved[key];
  }

  deleteSessionEnvironment(sessionId: string): void {
    this.sessions.delete(sessionId);
  }

  getSanitizedEnvironment(sessionId: string): Record<string, string> {
    const session = this.sessions.get(sessionId);
    if (!session) return {};

    const result: Record<string, string> = {};
    const secretKeys = new Set([...session.config.secrets, ...this.globalSecrets]);

    for (const [key, value] of Object.entries(session.resolved)) {
      if (secretKeys.has(key) || this.looksLikeSecret(key)) {
        result[key] = '***REDACTED***';
      } else {
        result[key] = value;
      }
    }

    return result;
  }

  private looksLikeSecret(key: string): boolean {
    const secretPatterns = [
      /api[_-]?key/i,
      /secret/i,
      /password/i,
      /token/i,
      /credential/i,
      /auth/i,
      /private[_-]?key/i,
    ];
    return secretPatterns.some(pattern => pattern.test(key));
  }

  getRequiredVarsForCLI(cliType: CLIType): string[] {
    const required: Record<CLIType, string[]> = {
      opencode: [],
      claude: ['ANTHROPIC_API_KEY'],
      aider: ['OPENAI_API_KEY'],
      cursor: [],
      continue: [],
      cody: ['SRC_ACCESS_TOKEN'],
      copilot: ['GITHUB_TOKEN'],
      custom: [],
    };
    return required[cliType];
  }

  validateEnvironmentForCLI(cliType: CLIType, env: Record<string, string>): { valid: boolean; missing: string[] } {
    const required = this.getRequiredVarsForCLI(cliType);
    const missing = required.filter(key => !env[key]);
    return { valid: missing.length === 0, missing };
  }
}

describe('EnvironmentManager', () => {
  let manager: TestEnvironmentManager;

  beforeEach(() => {
    manager = new TestEnvironmentManager();
  });

  describe('global overrides', () => {
    test('sets global override', () => {
      manager.setGlobalOverride('MY_VAR', 'my-value');
      
      const env = manager.createSessionEnvironment('session-1', 'opencode');
      
      expect(env.MY_VAR).toBe('my-value');
    });

    test('removes global override', () => {
      manager.setGlobalOverride('MY_VAR', 'my-value');
      manager.removeGlobalOverride('MY_VAR');
      
      const env = manager.createSessionEnvironment('session-1', 'opencode');
      
      expect(env.MY_VAR).toBeUndefined();
    });

    test('applies to all new sessions', () => {
      manager.setGlobalOverride('GLOBAL', 'value');
      
      manager.createSessionEnvironment('session-1', 'opencode');
      manager.createSessionEnvironment('session-2', 'claude');
      
      expect(manager.getSessionEnvironment('session-1')?.GLOBAL).toBe('value');
      expect(manager.getSessionEnvironment('session-2')?.GLOBAL).toBe('value');
    });
  });

  describe('createSessionEnvironment', () => {
    test('creates environment with default config', () => {
      const env = manager.createSessionEnvironment('session-1', 'opencode');
      
      expect(env).toBeDefined();
      expect(manager.getSessionEnvironment('session-1')).toEqual(env);
    });

    test('applies session-specific variables', () => {
      const env = manager.createSessionEnvironment('session-1', 'opencode', {
        variables: {
          CUSTOM_VAR: 'custom-value',
        },
      });
      
      expect(env.CUSTOM_VAR).toBe('custom-value');
    });

    test('expands variable references', () => {
      manager.setGlobalOverride('BASE_PATH', '/base');
      
      const env = manager.createSessionEnvironment('session-1', 'opencode', {
        variables: {
          FULL_PATH: '${BASE_PATH}/subdir',
        },
      });
      
      expect(env.FULL_PATH).toBe('/base/subdir');
    });
  });

  describe('getSessionEnvironment', () => {
    test('returns undefined for non-existent session', () => {
      expect(manager.getSessionEnvironment('nonexistent')).toBeUndefined();
    });

    test('returns resolved environment', () => {
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { FOO: 'bar' },
      });
      
      expect(manager.getSessionEnvironment('session-1')?.FOO).toBe('bar');
    });
  });

  describe('updateSessionVariable', () => {
    test('updates existing variable', () => {
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { VAR: 'old' },
      });
      
      manager.updateSessionVariable('session-1', 'VAR', 'new');
      
      expect(manager.getSessionEnvironment('session-1')?.VAR).toBe('new');
    });

    test('adds new variable', () => {
      manager.createSessionEnvironment('session-1', 'opencode');
      
      manager.updateSessionVariable('session-1', 'NEW_VAR', 'value');
      
      expect(manager.getSessionEnvironment('session-1')?.NEW_VAR).toBe('value');
    });

    test('does nothing for non-existent session', () => {
      manager.updateSessionVariable('nonexistent', 'VAR', 'value');
      
      expect(manager.getSessionEnvironment('nonexistent')).toBeUndefined();
    });
  });

  describe('removeSessionVariable', () => {
    test('removes variable', () => {
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { VAR: 'value' },
      });
      
      manager.removeSessionVariable('session-1', 'VAR');
      
      expect(manager.getSessionEnvironment('session-1')?.VAR).toBeUndefined();
    });

    test('does nothing for non-existent session', () => {
      manager.removeSessionVariable('nonexistent', 'VAR');
    });
  });

  describe('deleteSessionEnvironment', () => {
    test('deletes session environment', () => {
      manager.createSessionEnvironment('session-1', 'opencode');
      
      manager.deleteSessionEnvironment('session-1');
      
      expect(manager.getSessionEnvironment('session-1')).toBeUndefined();
    });
  });

  describe('getSanitizedEnvironment', () => {
    test('redacts secrets by name', () => {
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { MY_SECRET: 'secret-value' },
        secrets: ['MY_SECRET'],
      });
      
      const sanitized = manager.getSanitizedEnvironment('session-1');
      
      expect(sanitized.MY_SECRET).toBe('***REDACTED***');
    });

    test('redacts global secrets', () => {
      manager.addGlobalSecret('GLOBAL_SECRET');
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { GLOBAL_SECRET: 'secret' },
      });
      
      const sanitized = manager.getSanitizedEnvironment('session-1');
      
      expect(sanitized.GLOBAL_SECRET).toBe('***REDACTED***');
    });

    test('auto-redacts API keys', () => {
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { OPENAI_API_KEY: 'sk-xxx' },
      });
      
      const sanitized = manager.getSanitizedEnvironment('session-1');
      
      expect(sanitized.OPENAI_API_KEY).toBe('***REDACTED***');
    });

    test('auto-redacts tokens', () => {
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { ACCESS_TOKEN: 'token-xxx' },
      });
      
      const sanitized = manager.getSanitizedEnvironment('session-1');
      
      expect(sanitized.ACCESS_TOKEN).toBe('***REDACTED***');
    });

    test('auto-redacts passwords', () => {
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { DB_PASSWORD: 'password123' },
      });
      
      const sanitized = manager.getSanitizedEnvironment('session-1');
      
      expect(sanitized.DB_PASSWORD).toBe('***REDACTED***');
    });

    test('preserves non-secret values', () => {
      manager.createSessionEnvironment('session-1', 'opencode', {
        variables: { NORMAL_VAR: 'normal-value' },
      });
      
      const sanitized = manager.getSanitizedEnvironment('session-1');
      
      expect(sanitized.NORMAL_VAR).toBe('normal-value');
    });

    test('returns empty object for non-existent session', () => {
      expect(manager.getSanitizedEnvironment('nonexistent')).toEqual({});
    });
  });

  describe('getRequiredVarsForCLI', () => {
    test('returns required vars for claude', () => {
      expect(manager.getRequiredVarsForCLI('claude')).toContain('ANTHROPIC_API_KEY');
    });

    test('returns required vars for aider', () => {
      expect(manager.getRequiredVarsForCLI('aider')).toContain('OPENAI_API_KEY');
    });

    test('returns required vars for cody', () => {
      expect(manager.getRequiredVarsForCLI('cody')).toContain('SRC_ACCESS_TOKEN');
    });

    test('returns required vars for copilot', () => {
      expect(manager.getRequiredVarsForCLI('copilot')).toContain('GITHUB_TOKEN');
    });

    test('returns empty array for opencode', () => {
      expect(manager.getRequiredVarsForCLI('opencode')).toEqual([]);
    });
  });

  describe('validateEnvironmentForCLI', () => {
    test('returns valid when all required vars present', () => {
      const result = manager.validateEnvironmentForCLI('claude', {
        ANTHROPIC_API_KEY: 'sk-xxx',
      });
      
      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
    });

    test('returns invalid with missing vars', () => {
      const result = manager.validateEnvironmentForCLI('claude', {});
      
      expect(result.valid).toBe(false);
      expect(result.missing).toContain('ANTHROPIC_API_KEY');
    });

    test('returns valid for CLI with no requirements', () => {
      const result = manager.validateEnvironmentForCLI('opencode', {});
      
      expect(result.valid).toBe(true);
      expect(result.missing).toEqual([]);
    });
  });
});
