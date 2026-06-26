import type { Supervisor, SupervisorConfig, Message } from './types.js';
import { fetchWithRetry } from './retry.js';

interface OpenAIChatMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

interface OpenAIChatResponse {
  choices: Array<{
    message: {
      content: string;
    };
  }>;
}

export class OpenAISupervisor implements Supervisor {
  name: string;
  provider: string;

  private apiKey: string;
  private model: string;
  private baseURL: string;
  private systemPrompt?: string;
  private temperature: number;

  constructor(config: SupervisorConfig) {
    this.name = config.name;
    this.provider = config.provider;
    this.apiKey = config.apiKey || process.env.OPENAI_API_KEY || '';
    this.model = config.model || 'gpt-4o';
    this.baseURL = config.baseURL || 'https://api.openai.com/v1';
    this.systemPrompt = config.systemPrompt;
    this.temperature = config.temperature ?? 0.7;
  }

  async chat(messages: Message[]): Promise<string> {
    if (!this.apiKey) {
      throw new Error(`OpenAI API key not configured for supervisor: ${this.name}`);
    }

    const openaiMessages: OpenAIChatMessage[] = [];

    if (this.systemPrompt) {
      openaiMessages.push({
        role: 'system',
        content: this.systemPrompt,
      });
    }

    for (const msg of messages) {
      openaiMessages.push({
        role: msg.role,
        content: msg.content,
      });
    }

    const response = await fetchWithRetry(this.name, `${this.baseURL}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify({
        model: this.model,
        messages: openaiMessages,
        temperature: this.temperature,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`OpenAI API error (${response.status}): ${error}`);
    }

    const data = (await response.json()) as OpenAIChatResponse;
    return data.choices[0]?.message?.content || '';
  }

  async isAvailable(): Promise<boolean> {
    if (!this.apiKey) {
      return false;
    }

    try {
      const response = await fetch(`${this.baseURL}/models`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
        },
      });
      return response.ok;
    } catch {
      return false;
    }
  }
}
