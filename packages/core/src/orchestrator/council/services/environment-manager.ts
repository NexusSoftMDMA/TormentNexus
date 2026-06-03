import type { CLIType } from './types.js';

interface EnvironmentConfig {
  inherit: boolean;
  variables: Record<string, string>;
  secrets: string[];
  passthrough: string[];
}

interface SessionEnvironment {
  sessionId: string;
  cliType: CLIType;
  config: EnvironmentConfig;
  resolved: Record<string, string>;
}

const DEFAULT_PASSTHROUGH_VARS = [
  'PATH',
  'HOME',
  'USER',
  'SHELL',
  'TERM',
  'LANG',
  'LC_ALL',
  'TMPDIR',
  'TMP',
  'TEMP',
  'NODE_ENV',
  'BUN_ENV',
];

const CLI_SPECIFIC_VARS: Record<CLIType, string[]> = {
  tormentnexus: [],
  opencode: ['OPENCODE_*', 'ANTHROPIC_API_KEY', 'OPENAI_API_KEY'],
  antigravity: ['ANTIGRAVITY_*', 'GOOGLE_API_KEY'],
  claude: ['ANTHROPIC_API_KEY', 'CLAUDE_*'],
  aider: ['AIDER_*', 'OPENAI_API_KEY', 'ANTHROPIC_API_KEY', 'GEMINI_API_KEY'],
  cursor: ['CURSOR_*', 'OPENAI_API_KEY'],
  continue: ['CONTINUE_*', 'OPENAI_API_KEY', 'ANTHROPIC_API_KEY'],
  cody: ['SRC_*', 'CODY_*'],
  copilot: ['GITHUB_*', 'GH_*'],
  custom: [],
  adrenaline: ['ADRENALINE_*', 'OPENAI_API_KEY'],
  'amazon-q': ['AWS_*', 'Q_*'],
  'amazon-q-developer': ['AWS_*', 'Q_*'],
  'amp-code': ['AMP_*'],
  auggie: ['AUGGIE_*'],
  'azure-openai': ['AZURE_*', 'OPENAI_*'],
  bito: ['BITO_*'],
  byterover: ['BYTEROVER_*'],
  'claude-code': ['ANTHROPIC_API_KEY', 'CLAUDE_*'],
  'code-codex': ['CODEX_*'],
  codebuff: ['CODEBUFF_*'],
  codemachine: ['CODEMACHINE_*'],
  codex: ['CODEX_*', 'OPENAI_API_KEY'],
  crush: ['CRUSH_*'],
  dolt: ['DOLT_*'],
  factory: ['FACTORY_*'],
  gemini: ['GEMINI_API_KEY', 'GOOGLE_API_KEY'],
  goose: ['GOOSE_*'],
  grok: ['GROK_API_KEY', 'XAI_API_KEY'],
  jules: ['JULES_*'],
  'kilo-code': ['KILO_*'],
  kimi: ['KIMI_API_KEY', 'MOONSHOT_API_KEY'],
  llm: ['LLM_*', 'OPENAI_API_KEY'],
  litellm: ['LITELLM_*', 'OPENAI_API_KEY'],
  llamafile: ['LLAMAFILE_*'],
  manus: ['MANUS_*'],
  'mistral-vibe': ['MISTRAL_API_KEY'],
  ollama: ['OLLAMA_*'],
  'open-interpreter': ['INTERPRETER_*', 'OPENAI_API_KEY'],
  'qwen-code': ['QWEN_API_KEY', 'DASHSCOPE_API_KEY'],
  rowboatx: ['ROWBOATX_*'],
  rovo: ['ROVO_*'],
  'shell-pilot': ['SHELL_PILOT_*'],
  smithery: ['SMITHERY_*'],
  trae: ['TRAE_*'],
  pi: ['PI_*'],
  warp: ['WARP_*'],
  'superai-cli': ['SUPER_AI_*'],
  'factory-droid': ['FACTORY_DROID_*'],
};

class EnvironmentManager {
  private sessions: Map<string, SessionEnvironment> = new Map();
  private globalOverrides: Record<string, string> = {};
  private globalSecrets: Set<string> = new Set();

  constructor() {}

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

    if (config.inherit) {
      // Safely handle missing CLI types if any future ones are added without updating map
      const cliSpecific = CLI_SPECIFIC_VARS[cliType] || [];
      const passthroughPatterns = [
        ...DEFAULT_PASSTHROUGH_VARS,
        ...config.passthrough,
        ...cliSpecific,
      ];

      for (const [key, value] of Object.entries(process.env)) {
        if (value && this.matchesPattern(key, passthroughPatterns)) {
          result[key] = value;
        }
      }
    }

    for (const [key, value] of Object.entries(this.globalOverrides)) {
      result[key] = value;
    }

    for (const [key, value] of Object.entries(config.variables)) {
      result[key] = this.expandVariables(value, result);
    }

    return result;
  }

  private matchesPattern(key: string, patterns: string[]): boolean {
    for (const pattern of patterns) {
      if (pattern.endsWith('*')) {
        const prefix = pattern.slice(0, -1);
        if (key.startsWith(prefix)) return true;
      } else if (pattern === key) {
        return true;
      }
    }
    return false;
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
      tormentnexus: [],
      opencode: [],
      antigravity: [],
      claude: ['ANTHROPIC_API_KEY'],
      aider: ['OPENAI_API_KEY'],
      cursor: [],
      continue: [],
      cody: ['SRC_ACCESS_TOKEN'],
      copilot: ['GITHUB_TOKEN'],
      custom: [],
      adrenaline: [],
      'amazon-q': [],
      'amazon-q-developer': [],
      'amp-code': [],
      auggie: [],
      'azure-openai': ['AZURE_OPENAI_API_KEY', 'AZURE_OPENAI_ENDPOINT'],
      bito: [],
      byterover: [],
      'claude-code': ['ANTHROPIC_API_KEY'],
      'code-codex': [],
      codebuff: [],
      codemachine: [],
      codex: [],
      crush: [],
      dolt: [],
      factory: [],
      gemini: ['GEMINI_API_KEY'],
      goose: [],
      grok: ['GROK_API_KEY'],
      jules: [],
      'kilo-code': [],
      kimi: ['KIMI_API_KEY'],
      llm: [],
      litellm: [],
      llamafile: [],
      manus: [],
      'mistral-vibe': ['MISTRAL_API_KEY'],
      ollama: [],
      'open-interpreter': [],
      'qwen-code': ['QWEN_API_KEY'],
      rowboatx: [],
      rovo: [],
      'shell-pilot': [],
      smithery: [],
      trae: [],
      pi: [],
      warp: [],
      'superai-cli': [],
      'factory-droid': [],
    };

    return required[cliType] || [];
  }

  validateEnvironmentForCLI(cliType: CLIType, env: Record<string, string>): { valid: boolean; missing: string[] } {
    const required = this.getRequiredVarsForCLI(cliType);
    const missing = required.filter(key => !env[key]);
    return { valid: missing.length === 0, missing };
  }
}

export const environmentManager = new EnvironmentManager();
