import { describe, expect, it } from 'vitest';

import {
    buildStructuredUserPrompt,
    buildUserPromptContent,
    getStructuredUserPrompt,
} from './sessionPromptMemory.js';

describe('sessionPromptMemory', () => {
    it('normalizes structured user prompts', () => {
        const prompt = buildStructuredUserPrompt({
            content: '  Ship provider fallback next.  ',
            role: 'goal',
            promptNumber: 4,
            activeGoal: 'Ship provider fallback next.',
            lastObjective: 'Finish memory hooks',
            recordedAt: 100,
            contentHash: 'abc123',
        });

        expect(prompt).toMatchObject({
            role: 'goal',
            content: 'Ship provider fallback next.',
            promptNumber: 4,
            activeGoal: 'Ship provider fallback next.',
            lastObjective: 'Finish memory hooks',
            contentHash: 'abc123',
            recordedAt: 100,
        });
    });

    it('renders prompt content with contextual lines', () => {
        const content = buildUserPromptContent({
            role: 'objective',
            content: 'Wire goal saving into structured prompt memory.',
            promptNumber: 2,
            activeGoal: 'Improve tormentnexus parity',
            lastObjective: 'Add pre-tool memory injection',
            contentHash: 'hash',
            recordedAt: 50,
        });

        expect(content).toContain('Captured objective prompt #2');
        expect(content).toContain('Improve tormentnexus parity');
        expect(content).toContain('Add pre-tool memory injection');
    });

    it('extracts structured prompt metadata safely', () => {
        const parsed = getStructuredUserPrompt({
            structuredUserPrompt: {
                role: 'prompt',
                content: 'Review the MCP router roadmap.',
                promptNumber: 3,
                contentHash: 'hash-2',
                recordedAt: 77,
            },
        });

        expect(parsed).toMatchObject({
            role: 'prompt',
            content: 'Review the MCP router roadmap.',
            promptNumber: 3,
            contentHash: 'hash-2',
            recordedAt: 77,
        });
    });
});