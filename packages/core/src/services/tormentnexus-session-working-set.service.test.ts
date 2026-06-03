import { describe, expect, it } from 'vitest';

import { SessionToolWorkingSet } from './tormentnexus-session-working-set.service.js';

describe('SessionToolWorkingSet', () => {
    it('uses tighter default limits for the session working set', () => {
        const workingSet = new SessionToolWorkingSet();

        expect(workingSet.getLimits()).toEqual({
            maxLoadedTools: 16,
            maxHydratedSchemas: 8,
        });
    });

    it('evicts the least recently used loaded tool when the hard loaded cap is exceeded', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 3,
            maxHydratedSchemas: 2,
        });

        workingSet.loadTool('alpha');
        workingSet.loadTool('beta');
        workingSet.loadTool('gamma');
        workingSet.loadTool('beta');

        const evicted = workingSet.loadTool('delta');

        expect(evicted).toEqual(['alpha']);
        expect(workingSet.getLoadedToolNames()).toEqual(['gamma', 'beta', 'delta']);
    });

    it('evicts the least recently used hydrated schema independently of loaded metadata', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 4,
            maxHydratedSchemas: 2,
        });

        workingSet.hydrateTool('alpha');
        workingSet.hydrateTool('beta');
        const evicted = workingSet.hydrateTool('gamma');

        expect(evicted).toEqual(['alpha']);
        expect(workingSet.isLoaded('alpha')).toBe(true);
        expect(workingSet.isHydrated('alpha')).toBe(false);
        expect(workingSet.isHydrated('beta')).toBe(true);
        expect(workingSet.isHydrated('gamma')).toBe(true);
    });

    it('supports explicit unload and loaded-tool inspection', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 4,
            maxHydratedSchemas: 2,
        });

        workingSet.loadTool('alpha');
        workingSet.hydrateTool('beta');

        const snapshot = workingSet.listLoadedTools();

        expect(snapshot.map((item) => ({ name: item.name, hydrated: item.hydrated }))).toEqual([
            { name: 'alpha', hydrated: false },
            { name: 'beta', hydrated: true },
        ]);

        expect(workingSet.unloadTool('beta')).toBe(true);
        expect(workingSet.listLoadedTools().map((item) => item.name)).toEqual(['alpha']);
        expect(workingSet.unloadTool('missing')).toBe(false);
    });

    it('refreshes LRU order when an already loaded tool is actually used', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 3,
            maxHydratedSchemas: 2,
        });

        workingSet.loadTool('alpha');
        workingSet.loadTool('beta');
        workingSet.loadTool('gamma');

        expect(workingSet.touchTool('alpha')).toBe(true);

        const evicted = workingSet.loadTool('delta');

        expect(evicted).toEqual(['beta']);
        expect(workingSet.getLoadedToolNames()).toEqual(['gamma', 'alpha', 'delta']);
    });

    it('keeps always-loaded tools visible and protected from eviction or full unload', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 2,
            maxHydratedSchemas: 1,
        });

        workingSet.setAlwaysLoadedTools(['pinned']);
        workingSet.loadTool('alpha');
        const evicted = workingSet.loadTool('beta');

        expect(evicted).toEqual([]);
        expect(workingSet.getLoadedToolNames()).toEqual(['pinned', 'alpha', 'beta']);
        expect(workingSet.isLoaded('pinned')).toBe(true);
        expect(workingSet.unloadTool('pinned')).toBe(false);

        workingSet.hydrateTool('pinned');
        expect(workingSet.isHydrated('pinned')).toBe(true);
        expect(workingSet.unloadTool('pinned')).toBe(true);
        expect(workingSet.isLoaded('pinned')).toBe(true);
        expect(workingSet.isHydrated('pinned')).toBe(false);
    });

    it('records evicted tools in the bounded eviction history with tier=loaded', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 2,
            maxHydratedSchemas: 2,
        });

        workingSet.loadTool('alpha');
        workingSet.loadTool('beta');
        workingSet.loadTool('gamma'); // evicts alpha

        const history = workingSet.getEvictionHistory();
        expect(history).toHaveLength(1);
        expect(history[0].toolName).toBe('alpha');
        expect(history[0].tier).toBe('loaded');
        expect(typeof history[0].timestamp).toBe('number');
    });

    it('records hydrated schema evictions separately with tier=hydrated', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 4,
            maxHydratedSchemas: 2,
        });

        workingSet.hydrateTool('alpha');
        workingSet.hydrateTool('beta');
        workingSet.hydrateTool('gamma'); // evicts alpha schema only

        const history = workingSet.getEvictionHistory();
        expect(history).toHaveLength(1);
        expect(history[0].toolName).toBe('alpha');
        expect(history[0].tier).toBe('hydrated');
        // alpha is still loaded (metadata) even though schema was evicted
        expect(workingSet.isLoaded('alpha')).toBe(true);
    });

    it('returns history most recent first', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 2,
            maxHydratedSchemas: 2,
        });

        workingSet.loadTool('a');
        workingSet.loadTool('b');
        workingSet.loadTool('c'); // evicts a
        workingSet.loadTool('d'); // evicts b

        const history = workingSet.getEvictionHistory();
        expect(history[0].toolName).toBe('b');
        expect(history[1].toolName).toBe('a');
    });

    it('clearEvictionHistory empties the ring buffer', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 1,
            maxHydratedSchemas: 1,
        });

        workingSet.loadTool('alpha');
        workingSet.loadTool('beta'); // evicts alpha

        expect(workingSet.getEvictionHistory()).toHaveLength(1);
        workingSet.clearEvictionHistory();
        expect(workingSet.getEvictionHistory()).toHaveLength(0);
    });

    it('reconfigure updates capacity limits without clearing loaded tools', () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 8,
            maxHydratedSchemas: 4,
        });

        workingSet.loadTool('alpha');
        workingSet.loadTool('beta');

        workingSet.reconfigure({ maxLoadedTools: 32, maxHydratedSchemas: 16 });

        expect(workingSet.getLimits()).toEqual({ maxLoadedTools: 32, maxHydratedSchemas: 16 });
        // Previously loaded tools are still present.
        expect(workingSet.isLoaded('alpha')).toBe(true);
        expect(workingSet.isLoaded('beta')).toBe(true);
    });

    it('reconfigure clamps inputs to valid bounds', () => {
        const workingSet = new SessionToolWorkingSet();

        workingSet.reconfigure({ maxLoadedTools: 200, maxHydratedSchemas: 0 });

        const limits = workingSet.getLimits();
        expect(limits.maxLoadedTools).toBe(64);       // clamped to max
        expect(limits.maxHydratedSchemas).toBe(2);    // clamped to min
    });
});
