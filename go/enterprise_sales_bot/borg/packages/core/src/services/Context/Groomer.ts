import { ChatMessage } from '@tormentnexus/ai';

export class ContextGroomer {
    private maxTokens: number;

    constructor(maxTokens: number = 8000) {
        this.maxTokens = maxTokens;
    }

    public compressContext(messages: ChatMessage[]): ChatMessage[] {
        // Very rudimentary token estimation (approx 4 chars per token)
        const estimateTokens = (text: string) => text.length / 4;

        let currentTokens = 0;
        const result: ChatMessage[] = [];

        // Always keep the system message
        const systemMessage = messages.find(m => m.role === 'system');
        if (systemMessage) {
            currentTokens += estimateTokens(systemMessage.content);
            result.push(systemMessage);
        }

        // Keep the most recent messages that fit
        const recentMessages = messages.filter(m => m.role !== 'system').reverse();
        const keptRecent: ChatMessage[] = [];

        for (const msg of recentMessages) {
            const tokens = estimateTokens(msg.content);
            if (currentTokens + tokens > this.maxTokens) {
                // Insert a summary placeholder if we run out of room
                keptRecent.unshift({
                    role: 'system',
                    content: '[System Note: Earlier conversation context was compressed/pruned due to length limits.]'
                });
                break;
            }
            currentTokens += tokens;
            keptRecent.unshift(msg);
        }

        return [...result, ...keptRecent];
    }
}
