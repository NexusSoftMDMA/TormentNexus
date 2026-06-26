import { IAIProvider, AIProviderConfig, ChatMessage, AIProviderResponse } from './ProviderInterface.js';

export class GeminiProvider implements IAIProvider {
    name = 'gemini';
    capabilities = ['vision', 'tools', 'long-context'] as any;

    private apiKey?: string;

    initialize(config: AIProviderConfig): void {
        this.apiKey = config.apiKey;
    }

    async generateContent(
        modelId: string,
        messages: ChatMessage[],
        options?: { temperature?: number; maxTokens?: number; systemPrompt?: string; }
    ): Promise<AIProviderResponse> {
        if (!this.apiKey) throw new Error('Gemini API key not configured');

        const baseUrl = `https://generativelanguage.googleapis.com/v1beta/models/${modelId}:generateContent?key=${this.apiKey}`;

        let systemInstruction;
        if (options?.systemPrompt) {
            systemInstruction = {
                parts: [{ text: options.systemPrompt }]
            };
        }

        const contents = messages.filter(m => m.role !== 'system').map(m => ({
            role: m.role === 'assistant' ? 'model' : 'user',
            parts: [{ text: m.content }]
        }));

        const requestBody = {
            systemInstruction,
            contents,
            generationConfig: {
                temperature: options?.temperature ?? 0.7,
                maxOutputTokens: options?.maxTokens,
            }
        };

        const response = await fetch(baseUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Gemini API error: ${response.status} ${errorText}`);
        }

        const data = await response.json();
        const text = data.candidates?.[0]?.content?.parts?.[0]?.text || '';

        return {
            content: text,
            usage: {
                promptTokens: data.usageMetadata?.promptTokenCount || 0,
                completionTokens: data.usageMetadata?.candidatesTokenCount || 0,
                totalTokens: data.usageMetadata?.totalTokenCount || 0
            },
            raw: data
        };
    }
}
