
import { describe, it, expect } from 'vitest';
import { MemoryManager } from '../services/MemoryManager.js';

describe('Phase 26: Infinite Context V3', () => {
    it('should pruning long conversation history', () => {
        const memory = new MemoryManager(process.cwd());

        // Generate a synthetic conversation
        // System Prompt + 50 User/Assistant turns
        const messages: any[] = [
            { role: 'system', content: 'You are TormentNexus.' }
        ];

        for (let i = 0; i < 50; i++) {
            messages.push({ role: 'user', content: `Question ${i}: Tell me about the number ${i}.` });
            messages.push({ role: 'assistant', content: `Answer ${i}: The number ${i} is an integer.` });
        }

        const pruner = memory['pruner']; // Access private property or use public API
        // Wait, property is private. We should use public API.

        const initialSize = memory.getContextSize(messages);
        console.log(`Initial Context Size: ${initialSize} tokens`);

        // Prune to a very small limit ~ 200 tokens
        // Each message pair is roughly ~20 tokens. 50 pairs ~ 1000 tokens.
        const pruned = memory.pruneContext(messages, {
            maxTokens: 200,
            keepFirst: 1, // Keep system prompt
            keepLast: 4   // Keep last 2 exchanges (4 messages)
        });

        const finalSize = memory.getContextSize(pruned);
        console.log(`Pruned Context Size: ${finalSize} tokens`);

        const minimumPreservedMessages = [messages[0], ...messages.slice(-4)];
        const minimumAchievableSize = memory.getContextSize(minimumPreservedMessages);

        // Assertions
        expect(finalSize).toBeLessThanOrEqual(Math.max(200, minimumAchievableSize));
        expect(pruned[0].content).toBe('You are TormentNexus.'); // System prompt kept
        expect(pruned[pruned.length - 1].content).toContain('Answer 49'); // Last message kept
        expect(pruned.length).toBeLessThan(messages.length);
        expect(pruned.length).toBeGreaterThan(5); // At least 1 + 4
    });

    it('should not prune if within limits', () => {
        const memory = new MemoryManager(process.cwd());
        const messages = [
            { role: 'system', content: 'Hi' },
            { role: 'user', content: 'Hello' }
        ];

        const pruned = memory.pruneContext(messages, { maxTokens: 1000 });
        expect(pruned.length).toBe(2);
    });
});
