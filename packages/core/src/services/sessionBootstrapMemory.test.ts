import { describe, expect, it } from 'vitest';

import { buildSessionBootstrapPrompt } from './sessionBootstrapMemory.js';

describe('sessionBootstrapMemory', () => {
    it('builds a compact memory bootstrap prompt from summaries and observations', () => {
        const payload = buildSessionBootstrapPrompt({
            activeGoal: 'Finish native tormentnexus start injection',
            lastObjective: 'Reuse recent findings instead of re-discovering them',
            toolAdvertisementLines: [
                'search_tools — semantically search downstream tools for the current topic',
                'auto_call_tool — search and execute a matching tool in one shot',
            ],
            summaries: [
                { content: 'Previous session finished the MCP auto-load slice and validated core typecheck.' },
            ],
            observations: [
                {
                    title: 'Session router already stores stop summaries',
                    type: 'discovery',
                    toolName: 'read_file',
                },
            ],
        });

        expect(payload.summaryCount).toBe(1);
        expect(payload.observationCount).toBe(1);
        expect(payload.toolAdvertisementCount).toBe(2);
        expect(payload.prompt).toContain('Current goal: Finish native tormentnexus start injection');
        expect(payload.prompt).toContain('Suggested tools for the current topic:');
        expect(payload.prompt).toContain('search_tools');
        expect(payload.prompt).toContain('Recent session summaries:');
        expect(payload.prompt).toContain('Relevant observations:');
    });
});
