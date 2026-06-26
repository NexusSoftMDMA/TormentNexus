import { describe, expect, it } from 'vitest';

import { getMemoryTimelineWindow } from './agentMemoryTimeline.js';

describe('getMemoryTimelineWindow', () => {
    const memories = [
        {
            id: 'prompt-1',
            createdAt: new Date('2026-03-12T10:00:00.000Z'),
            metadata: {
                structuredUserPrompt: {
                    role: 'goal',
                    content: 'Tighten the memory dashboard.',
                    promptNumber: 1,
                    sessionId: 'session-42',
                    contentHash: 'prompt-1',
                    recordedAt: Date.parse('2026-03-12T10:00:00.000Z'),
                },
            },
        },
        {
            id: 'summary-1',
            createdAt: new Date('2026-03-12T10:05:00.000Z'),
            metadata: {
                structuredSessionSummary: {
                    sessionId: 'session-42',
                    name: 'Memory dashboard',
                    status: 'running',
                    contentHash: 'summary-1',
                    recordedAt: Date.parse('2026-03-12T10:05:00.000Z'),
                },
            },
        },
        {
            id: 'observation-1',
            createdAt: new Date('2026-03-12T10:10:00.000Z'),
            metadata: {
                sessionId: 'session-42',
                structuredObservation: {
                    toolName: 'read_file',
                },
            },
        },
        {
            id: 'observation-2',
            createdAt: new Date('2026-03-12T10:20:00.000Z'),
            metadata: {
                sessionId: 'session-42',
                structuredObservation: {
                    toolName: 'apply_patch',
                },
            },
        },
        {
            id: 'observation-3',
            createdAt: new Date('2026-03-12T10:30:00.000Z'),
            metadata: {
                sessionId: 'session-42',
                structuredObservation: {
                    toolName: 'run_task',
                },
            },
        },
        {
            id: 'other-session',
            createdAt: new Date('2026-03-12T10:15:00.000Z'),
            metadata: {
                sessionId: 'session-99',
            },
        },
    ];

    it('returns a chronological same-session window around the nearest anchor record', () => {
        const results = getMemoryTimelineWindow(
            memories,
            'session-42',
            Date.parse('2026-03-12T10:18:00.000Z'),
            2,
            1,
        );

        expect(results.map((memory) => memory.id)).toEqual([
            'summary-1',
            'observation-1',
            'observation-2',
            'observation-3',
        ]);
    });

    it('returns an empty list when the session is unknown', () => {
        expect(getMemoryTimelineWindow(memories, 'missing-session', Date.now(), 2, 2)).toEqual([]);
    });

    it('uses the last record as the anchor when the requested timestamp is after the session timeline', () => {
        const results = getMemoryTimelineWindow(
            memories,
            'session-42',
            Date.parse('2026-03-12T11:00:00.000Z'),
            2,
            2,
        );

        expect(results.map((memory) => memory.id)).toEqual([
            'observation-1',
            'observation-2',
            'observation-3',
        ]);
    });
});