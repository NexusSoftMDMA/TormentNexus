import { describe, expect, it } from 'vitest';

import { MemoryManager } from './MemoryManager.js';
import { RedundantMemoryManager } from './RedundantMemoryManager.js';

describe('MemoryManager pipeline summaries', () => {
    it('reports the sectioned store as active in the default redundant pipeline', () => {
        const manager = new MemoryManager('C:/tormentnexus-workspace');

        expect(manager.getPipelineSummary()).toEqual({
            configuredMode: 'redundant',
            providerNames: ['json', 'sectioned-store'],
            providerCount: 2,
            sectionedStoreEnabled: true,
        });
    });

    it('reports the sectioned store as inactive in json-only mode', () => {
        const manager = new MemoryManager('C:/tormentnexus-workspace', 'json');

        expect(manager.getPipelineSummary()).toEqual({
            configuredMode: 'json',
            providerNames: ['json'],
            providerCount: 1,
            sectionedStoreEnabled: false,
        });
    });
});

describe('RedundantMemoryManager provider registration', () => {
    it('lists the built-in provider names in fan-out order', () => {
        const manager = new RedundantMemoryManager('C:/tormentnexus-workspace');

        expect(manager.getProviderNames()).toEqual(['json', 'sectioned-store']);
    });
});