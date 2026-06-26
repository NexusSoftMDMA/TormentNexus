import { IAIProvider, AIProviderConfig, ChatMessage, AIProviderResponse } from './ProviderInterface.js';

export class OpenAIProvider implements IAIProvider {
    name = 'openai';
    capabilities = ['vision', 'tools'] as any;

    private apiKey?: string;
    private baseUrl = 'https://api.openai.com/v1/chat/completions';

    initialize(config: AIProviderConfig): void {
        this.apiKey = config.apiKey;
        if (config.baseUrl) this.baseUrl = config.baseUrl;
    }

    async generateContent(
        modelId: string,
        messages: ChatMessage[],
        options?: { temperature?: number; maxTokens?: number; systemPrompt?: string; }
    ): Promise<AIProviderResponse> {
        if (!this.apiKey) throw new Error('OpenAI API key not configured');

        let finalMessages = [...messages];
        if (options?.systemPrompt) {
            finalMessages.unshift({ role: 'system', content: options.systemPrompt });
        }

        const requestBody = {
            model: modelId,
            temperature: options?.temperature ?? 0.7,
            max_tokens: options?.maxTokens,
            messages: finalMessages.map(m => ({
                role: m.role,
                content: m.content
            }))
        };

        const response = await fetch(this.baseUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.apiKey}`
            },
            body: JSON.stringify(requestBody)
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`OpenAI API error: ${response.status} ${errorText}`);
        }

        const data = await response.json();
        return {
            content: data.choices[0]?.message?.content || '',
            usage: {
                promptTokens: data.usage?.prompt_tokens || 0,
                completionTokens: data.usage?.completion_tokens || 0,
                totalTokens: data.usage?.total_tokens || 0
            },
            raw: data
        };
    }
}
