import { describe, expect, it } from 'vitest';

import {
    buildToolPreferenceSettings,
    mergeToolPreferences,
    normalizeToolPreferences,
    readToolPreferencesFromSettings,
} from './mcp-tool-preferences.js';

describe('mcp tool preferences helpers', () => {
    it('normalizes duplicate and empty preference entries', () => {
        expect(normalizeToolPreferences({
            importantTools: [' github__issues ', '', 'github__issues', 'browser__open'],
            alwaysLoadedTools: [' browser__open ', 'browser__open', ''],
            autoLoadMinConfidence: 2,
        })).toEqual({
            importantTools: ['github__issues', 'browser__open'],
            alwaysLoadedTools: ['browser__open'],
            autoLoadMinConfidence: 0.99,
            maxLoadedTools: 16,
            maxHydratedSchemas: 8,
        });
    });

    it('reads both pinned and always-loaded tools from settings', () => {
        expect(readToolPreferencesFromSettings({
            importantTools: ['github__issues'],
            alwaysLoadedTools: ['browser__open'],
            autoLoadMinConfidence: 0.9,
        })).toEqual({
            importantTools: ['github__issues'],
            alwaysLoadedTools: ['browser__open'],
            autoLoadMinConfidence: 0.9,
            maxLoadedTools: 16,
            maxHydratedSchemas: 8,
        });
    });

    it('merges always-loaded tools ahead of normal search results', () => {
        const merged = mergeToolPreferences([
            {
                name: 'github__issues',
                description: 'Search GitHub issues',
                server: 'github',
                rank: 2,
            },
            {
                name: 'browser__open',
                description: 'Open a browser page',
                server: 'browser',
                rank: 1,
            },
        ], {
            importantTools: ['github__issues'],
            alwaysLoadedTools: ['browser__open'],
            autoLoadMinConfidence: 0.85,
        }, [
            {
                name: 'github__issues',
                description: 'Search GitHub issues',
                server: 'github',
            },
            {
                name: 'browser__open',
                description: 'Open a browser page',
                server: 'browser',
                alwaysOn: true,
            },
        ]);

        expect(merged.map((tool) => ({
            name: tool.name,
            important: tool.important,
            alwaysLoaded: tool.alwaysLoaded,
            alwaysOn: tool.alwaysOn,
        }))).toEqual([
            {
                name: 'browser__open',
                important: false,
                alwaysLoaded: true,
                alwaysOn: true,
            },
            {
                name: 'github__issues',
                important: true,
                alwaysLoaded: false,
                alwaysOn: false,
            },
        ]);
    });

    it('always advertises always-on catalog tools even when they are not in the initial result set', () => {
        const merged = mergeToolPreferences([], {
            importantTools: [],
            alwaysLoadedTools: [],
            autoLoadMinConfidence: 0.85,
        }, [
            {
                name: 'memory__recall',
                description: 'Recall a saved memory',
                server: 'memory',
                alwaysOn: true,
            },
        ]);

        expect(merged[0]).toMatchObject({
            name: 'memory__recall',
            alwaysOn: true,
            alwaysShow: true,
        });
        expect(merged[0]?.matchReason).toContain('always-on');
    });

    it('persists both important and always-loaded settings', () => {
        expect(buildToolPreferenceSettings({
            unrelated: true,
            toolSelection: {
                previous: 'keep-me',
            },
        }, {
            importantTools: ['github__issues'],
            alwaysLoadedTools: ['browser__open'],
            autoLoadMinConfidence: 0.9,
            maxLoadedTools: 16,
            maxHydratedSchemas: 8,
        })).toEqual({
            unrelated: true,
            toolSelection: {
                previous: 'keep-me',
                importantTools: ['github__issues'],
                alwaysLoadedTools: ['browser__open'],
                autoLoadMinConfidence: 0.9,
                maxLoadedTools: 16,
                maxHydratedSchemas: 8,
            },
        });
    });
});
