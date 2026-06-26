import { IAIProvider, AIProviderConfig, ChatMessage, AIProviderResponse } from './ProviderInterface.js';

export class AnthropicProvider implements IAIProvider {
    name = 'anthropic';
    capabilities = ['vision', 'tools', 'long-context'] as any;

    private apiKey?: string;
    private baseUrl = 'https://api.anthropic.com/v1/messages';

    initialize(config: AIProviderConfig): void {
        this.apiKey = config.apiKey;
        if (config.baseUrl) this.baseUrl = config.baseUrl;
    }

    async generateContent(
        modelId: string,
        messages: ChatMessage[],
        options?: { temperature?: number; maxTokens?: number; systemPrompt?: string; }
    ): Promise<AIProviderResponse> {
        if (!this.apiKey) throw new Error('Anthropic API key not configured');

        const systemMessages = messages.filter(m => m.role === 'system').map(m => m.content).join('\n');
        const finalSystem = options?.systemPrompt ? `${options.systemPrompt}\n${systemMessages}` : systemMessages;

        const filteredMessages = messages.filter(m => m.role !== 'system');

        const requestBody = {
            model: modelId,
            max_tokens: options?.maxTokens || 4096,
            temperature: options?.temperature ?? 0.7,
            system: finalSystem || undefined,
            messages: filteredMessages.map(m => ({
                role: m.role,
                content: m.content
            }))
        };

        const response = await fetch(this.baseUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'x-api-key': this.apiKey,
                'anthropic-version': '2023-06-01'
            },
            body: JSON.stringify(requestBody)
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Anthropic API error: ${response.status} ${errorText}`);
        }

        const data = await response.json();
        return {
            content: data.content[0]?.text || '',
            usage: {
                promptTokens: data.usage?.input_tokens || 0,
                completionTokens: data.usage?.output_tokens || 0,
                totalTokens: (data.usage?.input_tokens || 0) + (data.usage?.output_tokens || 0)
            },
            raw: data
        };
    }
}
