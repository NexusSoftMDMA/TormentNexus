
export interface LLMRequest {
    systemPrompt: string;
    messages: { role: 'user' | 'model', content: string }[];
    model?: string;
    temperature?: number;
}

export interface LLMResponse {
    content: string;
    usage?: {
        inputTokens: number;
        outputTokens: number;
    };
}

export interface LLMClient {
    generateContent(request: LLMRequest): Promise<LLMResponse>;
}
