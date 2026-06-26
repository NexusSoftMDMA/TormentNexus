
export interface AIProviderConfig {
    apiKey: string;
    organization?: string;
    baseUrl?: string;
}

export interface ChatMessage {
    role: 'system' | 'user' | 'assistant';
    content: string;
}

export interface AIProviderResponse {
    content: string;
    usage?: {
        promptTokens: number;
        completionTokens: number;
        totalTokens: number;
    };
    raw?: any;
}

export interface IAIProvider {
    name: string;
    capabilities: string[];

    initialize(config: AIProviderConfig): void;

    generateContent(
        modelId: string,
        messages: ChatMessage[],
        options?: {
            temperature?: number;
            maxTokens?: number;
            systemPrompt?: string;
        }
    ): Promise<AIProviderResponse>;
}
