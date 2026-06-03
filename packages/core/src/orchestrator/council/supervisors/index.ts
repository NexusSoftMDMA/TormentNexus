import type { Supervisor, SupervisorConfig } from './types.js';
import { MockSupervisor } from './mock.js';
import { OpenAISupervisor } from './openai.js';
import { AnthropicSupervisor } from './anthropic.js';
import { GenericOpenAISupervisor } from './generic-openai.js';
import { TormentNexusSupervisor } from './tormentnexus.js';

export function createSupervisor(config: SupervisorConfig): Supervisor {
  const { provider } = config;

  // Prefer TormentNexus-native supervisor if possible (it uses the shared LlmService)
  if (['openai', 'anthropic', 'google', 'gemini', 'grok', 'xai', 'deepseek', 'qwen', 'moonshot', 'kimi'].includes(provider)) {
    return new TormentNexusSupervisor(config);
  }

  switch (provider) {
    case 'openai':
      return new OpenAISupervisor(config);

    case 'anthropic':
      return new AnthropicSupervisor(config);

    case 'deepseek':
      return new GenericOpenAISupervisor({
        ...config,
        baseURL: config.baseURL || 'https://api.deepseek.com/v1',
        apiKey: config.apiKey || process.env.DEEPSEEK_API_KEY,
      });

    case 'qwen':
      return new GenericOpenAISupervisor({
        ...config,
        baseURL: config.baseURL || 'https://dashscope.aliyuncs.com/compatible-mode/v1',
        apiKey: config.apiKey || process.env.QWEN_API_KEY,
      });

    case 'moonshot':
    case 'kimi':
      return new GenericOpenAISupervisor({
        ...config,
        baseURL: config.baseURL || 'https://api.moonshot.cn/v1',
        apiKey: config.apiKey || process.env.KIMI_API_KEY || process.env.MOONSHOT_API_KEY,
      });

    case 'grok':
    case 'xai':
      return new GenericOpenAISupervisor({
        ...config,
        baseURL: config.baseURL || 'https://api.x.ai/v1',
        apiKey: config.apiKey || process.env.GROK_API_KEY || process.env.XAI_API_KEY,
      });

    case 'gemini':
    case 'google':
      return new GenericOpenAISupervisor({
        ...config,
        baseURL: config.baseURL || 'https://generativelanguage.googleapis.com/v1beta/openai',
        apiKey: config.apiKey || process.env.GEMINI_API_KEY || process.env.GOOGLE_API_KEY,
      });

    case 'custom':
      if (!config.baseURL) {
        throw new Error(`Custom provider requires baseURL for supervisor: ${config.name}`);
      }
      return new GenericOpenAISupervisor(config);

    default:
      throw new Error(`Unknown provider: ${provider} for supervisor: ${config.name}`);
  }
}

export function createSupervisors(configs: SupervisorConfig[]): Supervisor[] {
  return configs.map(createSupervisor);
}

export function createMockSupervisor(name: string = 'MockSupervisor'): Supervisor {
  return new MockSupervisor({ name, provider: 'custom' });
}

export { MockSupervisor } from './mock.js';
export { OpenAISupervisor } from './openai.js';
export { AnthropicSupervisor } from './anthropic.js';
export { GenericOpenAISupervisor } from './generic-openai.js';
