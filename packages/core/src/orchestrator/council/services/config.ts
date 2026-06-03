import type { CouncilConfig, SupervisorConfig, SessionTemplate, LogRotationConfig, HealthCheckConfig, CrashRecoveryConfig, CLIType } from './types.js';
import { existsSync, readFileSync, writeFileSync, mkdirSync } from 'fs';
import { join, dirname } from 'path';

const DEFAULT_ORCHESTRATOR_CONFIG_DIR = join(process.cwd(), '.tormentnexus', 'orchestrator');
const LEGACY_CONFIG_DIR = process.env.AUTOPILOT_CONFIG_DIR || join(process.cwd(), '.autopilot');
const CONFIG_DIR = process.env.TORMENTNEXUS_ORCHESTRATOR_CONFIG_DIR
  || process.env.AUTOPILOT_CONFIG_DIR
  || (existsSync(join(DEFAULT_ORCHESTRATOR_CONFIG_DIR, 'config.json')) ? DEFAULT_ORCHESTRATOR_CONFIG_DIR : undefined)
  || (existsSync(join(LEGACY_CONFIG_DIR, 'config.json')) ? LEGACY_CONFIG_DIR : DEFAULT_ORCHESTRATOR_CONFIG_DIR);
const CONFIG_FILE = join(CONFIG_DIR, 'config.json');

export interface PersistenceConfig {
  enabled: boolean;
  filePath: string;
  autoSaveIntervalMs: number;
  autoResumeOnStart: boolean;
  maxPersistedSessions: number;
}

export interface AutopilotConfig {
  council: CouncilConfig;
  server: {
    port: number;
    host: string;
    corsOrigins?: string;
  };
  sessions: {
    basePort: number;
    maxSessions: number;
    pollInterval: number;
    defaultCLI: CLIType;
  };
  persistence: PersistenceConfig;
  templates: SessionTemplate[];
  logRotation: LogRotationConfig;
  healthCheck: HealthCheckConfig;
  crashRecovery: CrashRecoveryConfig;
  environment: {
    passthrough: string[];
    globals: Record<string, string>;
  };
}

const DEFAULT_CONFIG: AutopilotConfig = {
  council: {
    supervisors: [],
    debateRounds: 2,
    consensusThreshold: 0.7,
    enabled: true,
    smartPilot: false,
  },
  server: {
    port: 3847,
    host: '0.0.0.0',
  },
  sessions: {
    basePort: 4096,
    maxSessions: 10,
    pollInterval: 10000,
    defaultCLI: 'opencode',
  },
  persistence: {
    enabled: true,
    filePath: '.tormentnexus/orchestrator/sessions.json',
    autoSaveIntervalMs: 5000,
    autoResumeOnStart: true,
    maxPersistedSessions: 100,
  },
  templates: [
    {
      name: 'default',
      description: 'Standard session with no special configuration',
      supervisors: [],
      tags: ['default'],
    },
    {
      name: 'review',
      description: 'Code review focused session',
      supervisors: [],
      tags: ['review', 'quality'],
    },
    {
      name: 'debug',
      description: 'Debugging and troubleshooting session',
      supervisors: [],
      tags: ['debug', 'troubleshoot'],
    },
    {
      name: 'feature',
      description: 'Feature development session',
      supervisors: [],
      tags: ['feature', 'development'],
    },
  ],
  logRotation: {
    maxLogsPerSession: 1000,
    maxLogAgeMs: 24 * 60 * 60 * 1000,
    pruneIntervalMs: 60000,
  },
  healthCheck: {
    enabled: true,
    intervalMs: 10000,
    timeoutMs: 5000,
    maxFailures: 3,
  },
  crashRecovery: {
    enabled: true,
    maxRestartAttempts: 3,
    restartDelayMs: 2000,
    backoffMultiplier: 2,
    maxBackoffMs: 30000,
  },
  environment: {
    passthrough: [],
    globals: {},
  },
};

function loadSupervisorsFromEnv(): SupervisorConfig[] {
  const supervisors: SupervisorConfig[] = [];

  if (process.env.OPENAI_API_KEY) {
    supervisors.push({
      name: 'GPT-4o',
      provider: 'openai',
      model: process.env.OPENAI_MODEL || 'gpt-4o',
      apiKey: process.env.OPENAI_API_KEY,
    });
  }

  if (process.env.ANTHROPIC_API_KEY) {
    supervisors.push({
      name: 'Claude',
      provider: 'anthropic',
      model: process.env.ANTHROPIC_MODEL || 'claude-3-5-sonnet-20241022',
      apiKey: process.env.ANTHROPIC_API_KEY,
    });
  }

  if (process.env.DEEPSEEK_API_KEY) {
    supervisors.push({
      name: 'DeepSeek',
      provider: 'deepseek',
      model: process.env.DEEPSEEK_MODEL || 'deepseek-chat',
      apiKey: process.env.DEEPSEEK_API_KEY,
    });
  }

  if (process.env.GEMINI_API_KEY) {
    supervisors.push({
      name: 'Gemini',
      provider: 'gemini',
      model: process.env.GEMINI_MODEL || 'gemini-pro',
      apiKey: process.env.GEMINI_API_KEY,
    });
  }

  if (process.env.GROK_API_KEY || process.env.XAI_API_KEY) {
    supervisors.push({
      name: 'Grok',
      provider: 'grok',
      model: process.env.GROK_MODEL || 'grok-beta',
      apiKey: process.env.GROK_API_KEY || process.env.XAI_API_KEY,
    });
  }

  if (process.env.QWEN_API_KEY) {
    supervisors.push({
      name: 'Qwen',
      provider: 'qwen',
      model: process.env.QWEN_MODEL || 'qwen-plus',
      apiKey: process.env.QWEN_API_KEY,
    });
  }

  if (process.env.KIMI_API_KEY || process.env.MOONSHOT_API_KEY) {
    supervisors.push({
      name: 'Kimi',
      provider: 'kimi',
      model: process.env.KIMI_MODEL || 'moonshot-v1-8k',
      apiKey: process.env.KIMI_API_KEY || process.env.MOONSHOT_API_KEY,
    });
  }

  return supervisors;
}

function getEnvValue(...names: string[]): string | undefined {
  for (const name of names) {
    const value = process.env[name];
    if (typeof value === 'string' && value.length > 0) {
      return value;
    }
  }
  return undefined;
}

export function loadConfig(): AutopilotConfig {
  let config = { ...DEFAULT_CONFIG };

  if (existsSync(CONFIG_FILE)) {
    try {
      const fileContent = readFileSync(CONFIG_FILE, 'utf-8');
      const fileConfig = JSON.parse(fileContent) as Partial<AutopilotConfig>;
      config = {
        ...config,
        ...fileConfig,
        council: { ...config.council, ...fileConfig.council },
        server: { ...config.server, ...fileConfig.server },
        sessions: { ...config.sessions, ...fileConfig.sessions },
        persistence: { ...config.persistence, ...fileConfig.persistence },
        logRotation: { ...config.logRotation, ...fileConfig.logRotation },
        healthCheck: { ...config.healthCheck, ...fileConfig.healthCheck },
        crashRecovery: { ...config.crashRecovery, ...fileConfig.crashRecovery },
        environment: { ...config.environment, ...fileConfig.environment },
        templates: fileConfig.templates || config.templates,
      };
    } catch {
      console.warn(`Failed to load config from ${CONFIG_FILE}, using defaults`);
    }
  }

  const envSupervisors = loadSupervisorsFromEnv();
  if (envSupervisors.length > 0) {
    const existingNames = new Set(config.council.supervisors.map(s => s.name));
    for (const sup of envSupervisors) {
      if (!existingNames.has(sup.name)) {
        config.council.supervisors.push(sup);
      }
    }
  }

  const orchestratorPort = getEnvValue('TORMENTNEXUS_ORCHESTRATOR_PORT', 'AUTOPILOT_PORT');
  if (orchestratorPort) {
    config.server.port = parseInt(orchestratorPort);
  }
  const orchestratorHost = getEnvValue('TORMENTNEXUS_ORCHESTRATOR_HOST', 'AUTOPILOT_HOST');
  if (orchestratorHost) {
    config.server.host = orchestratorHost;
  }
  const basePort = getEnvValue('TORMENTNEXUS_ORCHESTRATOR_BASE_PORT', 'AUTOPILOT_BASE_PORT');
  if (basePort) {
    config.sessions.basePort = parseInt(basePort);
  }
  const debateRounds = getEnvValue('TORMENTNEXUS_ORCHESTRATOR_DEBATE_ROUNDS', 'AUTOPILOT_DEBATE_ROUNDS');
  if (debateRounds) {
    config.council.debateRounds = parseInt(debateRounds);
  }
  const consensusThreshold = getEnvValue('TORMENTNEXUS_ORCHESTRATOR_CONSENSUS', 'AUTOPILOT_CONSENSUS');
  if (consensusThreshold) {
    config.council.consensusThreshold = parseFloat(consensusThreshold);
  }
  const smartPilot = getEnvValue('TORMENTNEXUS_ORCHESTRATOR_SMART_PILOT', 'AUTOPILOT_SMART_PILOT');
  if (smartPilot) {
    config.council.smartPilot = smartPilot === 'true';
  }

  return config;
}

export function saveConfig(config: AutopilotConfig): void {
  if (!existsSync(CONFIG_DIR)) {
    mkdirSync(CONFIG_DIR, { recursive: true });
  }

  const sanitizedConfig = {
    ...config,
    council: {
      ...config.council,
      supervisors: config.council.supervisors.map(s => ({
        ...s,
        apiKey: undefined,
      })),
    },
  };

  writeFileSync(CONFIG_FILE, JSON.stringify(sanitizedConfig, null, 2));
}

export function updateCouncilConfig(updates: Partial<CouncilConfig>): AutopilotConfig {
  const config = loadConfig();
  config.council = { ...config.council, ...updates };
  saveConfig(config);
  return config;
}
