import { describe, expect, it } from 'vitest';

import { searchMemoryRecordsByPivot } from './agentMemoryPivot.js';

describe('searchMemoryRecordsByPivot', () => {
    const observation = {
        id: 'observation-1',
        createdAt: new Date('2026-03-12T12:00:00.000Z'),
        metadata: {
            sessionId: 'session-42',
            structuredObservation: {
                toolName: 'grep_search',
                concepts: ['billing', 'routing'],
                filesRead: ['packages/core/src/routers/billingRouter.ts'],
                filesModified: ['apps/web/src/app/dashboard/memory/page.tsx'],
            },
        },
    };

    const prompt = {
        id: 'prompt-1',
        createdAt: new Date('2026-03-12T11:30:00.000Z'),
        metadata: {
            structuredUserPrompt: {
                role: 'goal',
                content: 'Verify the billing routing fallback flow.',
                promptNumber: 3,
                sessionId: 'session-42',
                contentHash: 'prompt-hash',
                recordedAt: Date.parse('2026-03-12T11:30:00.000Z'),
            },
        },
    };

    const summary = {
        id: 'summary-1',
        createdAt: new Date('2026-03-12T11:45:00.000Z'),
        metadata: {
            structuredSessionSummary: {
                sessionId: 'session-42',
                name: 'Billing routing cleanup',
                status: 'running',
                cliType: 'claude',
                activeGoal: 'Keep the billing routing flow truthful.',
                contentHash: 'summary-hash',
                recordedAt: Date.parse('2026-03-12T11:45:00.000Z'),
            },
        },
    };

    const unrelated = {
        id: 'observation-2',
        createdAt: new Date('2026-03-12T12:05:00.000Z'),
        metadata: {
            sessionId: 'session-99',
            structuredObservation: {
                toolName: 'read_file',
                concepts: ['roadmap'],
                filesRead: ['ROADMAP.md'],
                filesModified: [],
            },
        },
    };

    const objectivePrompt = {
        id: 'prompt-2',
        createdAt: new Date('2026-03-12T11:40:00.000Z'),
        metadata: {
            structuredUserPrompt: {
                role: 'objective',
                content: 'Patch the billing router fallback path.',
                promptNumber: 4,
                sessionId: 'session-42',
                contentHash: 'objective-hash',
                recordedAt: Date.parse('2026-03-12T11:40:00.000Z'),
            },
        },
    };

    it('returns direct and session-related records for a tool pivot', () => {
        const results = searchMemoryRecordsByPivot([
            observation,
            prompt,
            summary,
            unrelated,
        ], 'tool', 'grep_search', 10);

        expect(results.map((memory) => memory.id)).toEqual([
            'observation-1',
            'summary-1',
            'prompt-1',
        ]);
        expect(results[0]?.score).toBe(10);
        expect(results[1]?.score).toBe(4);
        expect(results[2]?.score).toBe(4);
    });

    it('matches file pivots using normalized paths', () => {
        const results = searchMemoryRecordsByPivot([
            observation,
            unrelated,
        ], 'file', 'apps\\web\\src\\app\\dashboard\\memory\\page.tsx', 10);

        expect(results).toHaveLength(1);
        expect(results[0]?.id).toBe('observation-1');
        expect(results[0]?.score).toBe(10);
    });

    it('returns only exact session matches for session pivots', () => {
        const results = searchMemoryRecordsByPivot([
            observation,
            prompt,
            summary,
            unrelated,
        ], 'session', 'session-42', 10);

        expect(results.map((memory) => memory.id)).toEqual([
            'observation-1',
            'summary-1',
            'prompt-1',
        ]);
        expect(results.every((memory) => memory.score === 10)).toBe(true);
    });

    it('matches goal pivots from summary intent anchors and includes session context', () => {
        const results = searchMemoryRecordsByPivot([
            observation,
            prompt,
            summary,
            unrelated,
        ], 'goal', 'Keep the billing routing flow truthful.', 10);

        expect(results.map((memory) => memory.id)).toEqual([
            'summary-1',
            'observation-1',
            'prompt-1',
        ]);
        expect(results[0]?.score).toBe(10);
        expect(results[1]?.score).toBe(4);
        expect(results[2]?.score).toBe(4);
    });

    it('matches objective pivots from explicit objective prompts', () => {
        const results = searchMemoryRecordsByPivot([
            observation,
            prompt,
            summary,
            objectivePrompt,
            unrelated,
        ], 'objective', 'Patch the billing router fallback path.', 10);

        expect(results.map((memory) => memory.id)).toEqual([
            'prompt-2',
            'observation-1',
            'summary-1',
            'prompt-1',
        ]);
        expect(results[0]?.score).toBe(10);
        expect(results.slice(1).every((memory) => memory.score === 4)).toBe(true);
    });
});