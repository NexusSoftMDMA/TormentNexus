import { describe, expect, it } from 'vitest';

import { buildToolContextPayload } from './toolContextMemory.js';

describe('buildToolContextPayload', () => {
    it('prefers observations that match the exact tool and touched file', () => {
        const payload = buildToolContextPayload({
            toolName: 'read_file',
            args: { filePath: 'apps/web/src/app/dashboard/session/page.tsx' },
            observations: [
                {
                    title: 'Reviewed session dashboard route',
                    narrative: 'Confirmed the page passes metadata into the session details dialog.',
                    toolName: 'read_file',
                    filesRead: ['apps/web/src/app/dashboard/session/page.tsx'],
                    type: 'discovery',
                    recordedAt: 20,
                },
                {
                    title: 'Touched unrelated router',
                    narrative: 'Investigated billing router behavior.',
                    toolName: 'read_file',
                    filesRead: ['packages/core/src/routers/billingRouter.ts'],
                    type: 'discovery',
                    recordedAt: 30,
                },
            ],
            summaries: [],
        });

        expect(payload.observationCount).toBe(2);
        expect(payload.prompt).toContain('Reviewed session dashboard route');
        expect(payload.prompt).toContain('Relevant files: page.tsx');
    });

    it('falls back to matching summaries when observations are absent', () => {
        const payload = buildToolContextPayload({
            toolName: 'apply_patch',
            args: { filePath: 'packages/core/src/MCPServer.ts' },
            activeGoal: 'Wire native JIT memory before tool calls.',
            observations: [],
            summaries: [
                {
                    content: 'TormentNexus Dev Session ended after wiring session-start bootstrap metadata into supervised sessions.',
                    cliType: 'tabby',
                    status: 'stopped',
                    recordedAt: 100,
                },
                {
                    content: 'Billing fallback validation session.',
                    cliType: 'tabby',
                    status: 'stopped',
                    recordedAt: 10,
                },
            ],
        });

        expect(payload.summaryCount).toBe(1);
        expect(payload.prompt).toContain('Current goal: Wire native JIT memory before tool calls.');
        expect(payload.prompt).toContain('session-start bootstrap metadata');
    });

    it('returns a useful no-match prompt when memory is not relevant', () => {
        const payload = buildToolContextPayload({
            toolName: 'browser_open',
            args: { url: 'https://example.com' },
            observations: [
                {
                    title: 'Fixed billing route mismatch',
                    narrative: 'Patched billing route fallback behavior.',
                    toolName: 'apply_patch',
                    filesModified: ['packages/core/src/routers/billingRouter.ts'],
                    type: 'fix',
                },
            ],
            summaries: [],
        });

        expect(payload.observationCount).toBe(0);
        expect(payload.summaryCount).toBe(0);
        expect(payload.prompt).toContain('No strongly relevant prior memory was found');
    });
});