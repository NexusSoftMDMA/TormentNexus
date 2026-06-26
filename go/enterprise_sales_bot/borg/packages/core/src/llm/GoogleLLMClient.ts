import { LLMClient, LLMRequest, LLMResponse } from "./types.js";

export class GoogleLLMClient implements LLMClient {
    private apiKey: string;

    constructor(apiKey: string) {
        this.apiKey = apiKey;
    }

    async generateContent(request: LLMRequest): Promise<LLMResponse> {
        const model = request.model || "gemini-1.5-flash";
        const url = `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent?key=${this.apiKey}`;

        // Convert common format to Gemini REST format
        const contents = request.messages.map(m => ({
            role: m.role === 'user' ? 'user' : 'model',
            parts: [{ text: m.content }]
        }));

        if (request.systemPrompt) {
            // Note: Gemini 1.5 supports system instruction. 
            // For simplicity in this REST client, we are prepending it or using the specific field if available.
            // Let's use the 'systemInstruction' field for newer models.
        }

        const payload: any = {
            contents,
            generationConfig: {
                temperature: request.temperature || 0.7
            }
        };

        if (request.systemPrompt) {
            payload.systemInstruction = {
                parts: [{ text: request.systemPrompt }]
            };
        }

        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Gemini API Error: ${response.status} - ${errorText}`);
        }

        const data = await response.json();
        const content = data.candidates?.[0]?.content?.parts?.[0]?.text || "";

        return {
            content,
            usage: {
                inputTokens: data.usageMetadata?.promptTokenCount || 0,
                outputTokens: data.usageMetadata?.candidatesTokenCount || 0
            }
        };
    }
}
