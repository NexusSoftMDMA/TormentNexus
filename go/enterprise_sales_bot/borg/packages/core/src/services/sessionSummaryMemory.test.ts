import { describe, expect, it } from 'vitest';

import {
    buildSessionSummaryContent,
    buildStructuredSessionSummary,
    createSessionSummaryContentHash,
    getStructuredSessionSummary,
} from './sessionSummaryMemory.js';

describe('sessionSummaryMemory helpers', () => {
    it('builds stable structured summaries and readable content', () => {
        const summary = buildStructuredSessionSummary({
            sessionId: 'session-123',
            name: 'TormentNexus Dev Session',
            cliType: 'tabby',
            workingDirectory: 'C:/repo',
            status: 'stopped',
            restartCount: 1,
            lastExitCode: 0,
            activeGoal: 'Ship native memory summaries',
            logTail: ['server started', 'tests passed'],
        }, 1234567890, ['server started', 'tests passed']);

        const { contentHash: _contentHash, ...summaryWithoutHash } = summary;

        expect(summary.contentHash).toBe(createSessionSummaryContentHash(summaryWithoutHash));
        expect(buildSessionSummaryContent(summary)).toContain('TormentNexus Dev Session (tabby) ended with status stopped.');
        expect(buildSessionSummaryContent(summary)).toContain('Goal: Ship native memory summaries');
    });

    it('reads structured session summaries from memory metadata safely', () => {
        expect(getStructuredSessionSummary({
            structuredSessionSummary: {
                sessionId: 'session-123',
                status: 'stopped',
                restartCount: 0,
                logTail: ['done'],
                contentHash: 'session-1',
                recordedAt: 123,
            },
        })).toMatchObject({
            sessionId: 'session-123',
            status: 'stopped',
            logTail: ['done'],
        });

        expect(getStructuredSessionSummary({ structuredSessionSummary: { nope: true } })).toBeNull();
    });
});
