import { describe, expect, it } from 'vitest';

import { summarizeCachedInventory } from './startupInventorySummary.js';

describe('summarizeCachedInventory', () => {
    it('preserves config-backed advertised counts and always-on posture', () => {
        const summary = summarizeCachedInventory({
            source: 'config',
            snapshotUpdatedAt: '2026-03-12T17:00:00.000Z',
            toolCounts: new Map(),
            servers: [
                {
                    uuid: 'config:always-on:0',
                    name: 'always-on',
                    alwaysOnAdvertised: true,
                },
                {
                    uuid: 'config:on-demand:1',
                    name: 'on-demand',
                    alwaysOnAdvertised: false,
                },
            ] as any,
            tools: [
                {
                    name: 'always-on__search',
                    server: 'always-on',
                    alwaysOn: true,
                },
                {
                    name: 'always-on__read',
                    server: 'always-on',
                    alwaysOn: true,
                },
                {
                    name: 'on-demand__browse',
                    server: 'on-demand',
                    alwaysOn: false,
                },
            ] as any,
        });

        expect(summary).toEqual({
            source: 'config',
            snapshotUpdatedAt: '2026-03-12T17:00:00.000Z',
            serverCount: 2,
            toolCount: 3,
            alwaysOnServerCount: 1,
            alwaysOnToolCount: 2,
        });
    });

    it('keeps empty snapshots empty without inventing cached readiness', () => {
        const summary = summarizeCachedInventory({
            source: 'empty',
            snapshotUpdatedAt: null,
            toolCounts: new Map(),
            servers: [],
            tools: [],
        });

        expect(summary).toEqual({
            source: 'empty',
            snapshotUpdatedAt: null,
            serverCount: 0,
            toolCount: 0,
            alwaysOnServerCount: 0,
            alwaysOnToolCount: 0,
        });
    });
});