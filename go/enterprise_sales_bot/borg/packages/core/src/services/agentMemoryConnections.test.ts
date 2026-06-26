import { describe, expect, it } from 'vitest';

import { getCrossSessionMemoryLinks } from './agentMemoryConnections.js';

describe('getCrossSessionMemoryLinks', () => {
    const memories = [
        {
            id: 'anchor',
            createdAt: new Date('2026-03-12T10:00:00.000Z'),
            metadata: {
                source: 'search',
                sessionId: 'session-1',
                structuredObservation: {
                    toolName: 'grep_search',
                    concepts: ['memory', 'router'],
                    filesRead: ['packages/core/src/routers/memoryRouter.ts'],
                    filesModified: ['apps/web/src/app/dashboard/memory/page.tsx'],
                },
            },
        },
        {
            id: 'same-session',
            createdAt: new Date('2026-03-12T10:05:00.000Z'),
            metadata: {
                sessionId: 'session-1',
                structuredObservation: {
                    toolName: 'grep_search',
                    concepts: ['memory'],
                    filesRead: ['packages/core/src/routers/memoryRouter.ts'],
                    filesModified: [],
                },
            },
        },
        {
            id: 'anchor-summary',
            createdAt: new Date('2026-03-12T10:06:00.000Z'),
            metadata: {
                structuredSessionSummary: {
                    sessionId: 'session-1',
                    status: 'running',
                    activeGoal: 'Keep the memory dashboard truthful.',
                    lastObjective: 'Surface related records clearly.',
                },
            },
        },
        {
            id: 'cross-session-best',
            createdAt: new Date('2026-03-12T11:00:00.000Z'),
            metadata: {
                source: 'search',
                sessionId: 'session-2',
                structuredObservation: {
                    toolName: 'grep_search',
                    concepts: ['memory', 'router'],
                    filesRead: ['packages/core/src/routers/memoryRouter.ts'],
                    filesModified: ['apps/web/src/app/dashboard/memory/page.tsx'],
                },
            },
        },
        {
            id: 'cross-session-secondary',
            createdAt: new Date('2026-03-12T12:00:00.000Z'),
            metadata: {
                sessionId: 'session-3',
                structuredObservation: {
                    toolName: 'read_file',
                    concepts: ['memory'],
                    filesRead: ['README.md'],
                    filesModified: [],
                },
            },
        },
        {
            id: 'prompt-session',
            createdAt: new Date('2026-03-12T09:30:00.000Z'),
            metadata: {
                structuredUserPrompt: {
                    role: 'goal',
                    content: 'Keep the memory dashboard truthful.',
                    promptNumber: 2,
                    sessionId: 'session-4',
                    contentHash: 'prompt-session-hash',
                    recordedAt: Date.parse('2026-03-12T09:30:00.000Z'),
                },
            },
        },
        {
            id: 'summary-session',
            createdAt: new Date('2026-03-12T12:30:00.000Z'),
            metadata: {
                structuredSessionSummary: {
                    sessionId: 'session-5',
                    status: 'running',
                    activeGoal: 'Keep the memory dashboard truthful.',
                    lastObjective: 'Surface related records clearly.',
                },
            },
        },
        {
            id: 'similar-goal-session',
            createdAt: new Date('2026-03-12T12:45:00.000Z'),
            metadata: {
                structuredSessionSummary: {
                    sessionId: 'session-6',
                    status: 'running',
                    activeGoal: 'Keep memory dashboard views truthful and coherent.',
                    lastObjective: 'Clarify related records in the inspector.',
                },
            },
        },
    ];

    it('returns only other-session records ranked by shared evidence', () => {
        const results = getCrossSessionMemoryLinks(memories, 'anchor', 5);

        expect(results.map((entry) => entry.memory.id)).toEqual([
            'cross-session-best',
            'summary-session',
            'similar-goal-session',
            'prompt-session',
            'cross-session-secondary',
        ]);
        expect(results[0]?.reasons).toEqual(expect.arrayContaining([
            'shared concepts: memory, router',
            'shared file: packages/core/src/routers/memoryRouter.ts',
            'same tool (grep_search)',
            'same source (search)',
            'other session (session-2)',
        ]));
        expect(results.some((entry) => entry.memory.id === 'same-session')).toBe(false);
        expect(results.find((entry) => entry.memory.id === 'summary-session')?.reasons).toEqual(expect.arrayContaining([
            'shared goal/objective: keep the memory dashboard truthful.',
            'other session (session-5)',
        ]));
        expect(results.find((entry) => entry.memory.id === 'prompt-session')?.reasons).toEqual(expect.arrayContaining([
            'shared goal/objective: keep the memory dashboard truthful.',
            'other session (session-4)',
        ]));
        expect(results.find((entry) => entry.memory.id === 'similar-goal-session')?.reasons).toEqual(expect.arrayContaining([
            'similar goal/objective theme: keep, memory',
            'other session (session-6)',
        ]));
    });

    it('returns an empty list when the anchor record is unknown', () => {
        expect(getCrossSessionMemoryLinks(memories, 'missing-anchor', 5)).toEqual([]);
    });
});