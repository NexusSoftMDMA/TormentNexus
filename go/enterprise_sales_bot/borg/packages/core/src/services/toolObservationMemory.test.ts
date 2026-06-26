import { describe, expect, it } from 'vitest';

import { buildToolObservationInput, shouldCaptureToolObservation } from './toolObservationMemory.js';

describe('toolObservationMemory', () => {
    it('builds a structured observation for successful read tools', () => {
        const observation = buildToolObservationInput({
            toolName: 'read_file',
            args: { filePath: 'src/example.ts' },
            result: { content: [{ type: 'text', text: 'export const answer = 42;' }] },
            durationMs: 87,
        });

        expect(observation).toMatchObject({
            toolName: 'read_file',
            filesRead: ['src/example.ts'],
        });
        expect(observation?.facts).toContain('Duration: 87ms');
        expect(observation?.narrative).toContain('read_file completed in 87ms');
    });

    it('builds warning observations for failed write tools', () => {
        const observation = buildToolObservationInput({
            toolName: 'apply_patch',
            args: { filePath: 'src/example.ts' },
            error: new Error('Patch did not apply cleanly'),
            durationMs: 12,
        });

        expect(observation).toMatchObject({
            toolName: 'apply_patch',
            type: 'warning',
            filesModified: ['src/example.ts'],
        });
        expect(observation?.narrative).toContain('failed after 12ms');
        expect(observation?.facts).toContain('Failure: Patch did not apply cleanly');
    });

    it('skips low-signal control tools', () => {
        expect(shouldCaptureToolObservation('search_tools')).toBe(false);
        expect(buildToolObservationInput({
            toolName: 'search_tools',
            args: { query: 'foo' },
            result: { content: [{ type: 'text', text: '...' }] },
            durationMs: 5,
        })).toBeNull();
    });
});