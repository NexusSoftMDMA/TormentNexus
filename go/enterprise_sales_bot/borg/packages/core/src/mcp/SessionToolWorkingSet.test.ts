import { afterEach, describe, expect, it, vi } from 'vitest';

import { SessionToolWorkingSet } from './SessionToolWorkingSet.js';

describe('SessionToolWorkingSet', () => {
    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('records loaded-tier evictions with idle metadata', () => {
        const nowSpy = vi.spyOn(Date, 'now');
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 2,
            maxHydratedSchemas: 1,
            idleEvictionThresholdMs: 1_000,
        });

        nowSpy.mockReturnValueOnce(0);
        workingSet.loadTool('a');

        nowSpy.mockReturnValueOnce(100);
        workingSet.loadTool('b');

        nowSpy.mockReturnValueOnce(2_000);
        const evicted = workingSet.loadTool('c');

        expect(evicted).toEqual(['a']);
        const history = workingSet.getEvictionHistory();
        expect(history).toHaveLength(1);
        expect(history[0]).toMatchObject({
            toolName: 'a',
            tier: 'loaded',
            idleEvicted: true,
        });
    });

    it('applies reconfigure limits immediately and supports clearing history', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 4,
            maxHydratedSchemas: 3,
        });

        workingSet.loadTool('a');
        workingSet.loadTool('b');
        workingSet.loadTool('c');

        workingSet.reconfigure({ maxLoadedTools: 2, maxHydratedSchemas: 1 });

        expect(workingSet.getLoadedToolNames().length).toBeLessThanOrEqual(2);
        expect(workingSet.getEvictionHistory().length).toBeGreaterThan(0);

        workingSet.clearEvictionHistory();
        expect(workingSet.getEvictionHistory()).toEqual([]);
    });

    it('tracks always-loaded tool recency metadata', () => {
        const nowSpy = vi.spyOn(Date, 'now');
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 2,
            maxHydratedSchemas: 1,
        });

        nowSpy.mockReturnValueOnce(1_000);
        workingSet.setAlwaysLoadedTools(['browser__pinned']);

        const initialState = workingSet.listLoadedTools().find((tool) => tool.name === 'browser__pinned');
        expect(initialState).toBeTruthy();
        expect(initialState?.lastLoadedAt).toBe(1_000);
        expect(initialState?.lastAccessedAt).toBe(1_000);

        nowSpy.mockReturnValueOnce(2_500);
        expect(workingSet.touchTool('browser__pinned')).toBe(true);

        const updatedState = workingSet.listLoadedTools().find((tool) => tool.name === 'browser__pinned');
        expect(updatedState?.lastLoadedAt).toBe(1_000);
        expect(updatedState?.lastAccessedAt).toBe(2_500);
    });

    it('returns eviction history most recent first', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 1,
            maxHydratedSchemas: 1,
        });

        workingSet.loadTool('alpha');
        workingSet.loadTool('beta'); // evicts alpha
        workingSet.loadTool('gamma'); // evicts beta

        const history = workingSet.getEvictionHistory();
        expect(history).toHaveLength(2);
        expect(history[0]?.toolName).toBe('beta');
        expect(history[1]?.toolName).toBe('alpha');
    });

    it('bounds eviction history to the configured retention size', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 1,
            maxHydratedSchemas: 1,
        });

        for (let index = 0; index < 220; index += 1) {
            workingSet.loadTool(`tool-${index}`);
        }

        const history = workingSet.getEvictionHistory();
        expect(history).toHaveLength(200);
        expect(history[0]?.toolName).toBe('tool-218');
        expect(history[history.length - 1]?.toolName).toBe('tool-19');
    });
});
