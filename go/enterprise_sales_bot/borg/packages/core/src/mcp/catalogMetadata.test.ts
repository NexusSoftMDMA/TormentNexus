import { describe, expect, it } from 'vitest';

import { deriveSemanticCatalogForServer } from './catalogMetadata.js';

describe('catalogMetadata semantic derivation', () => {
    it('derives server and tool tags from server descriptions and tool inventory', () => {
        const derived = deriveSemanticCatalogForServer({
            serverName: 'browser',
            description: 'Playwright browser automation for pages, tabs, screenshots, and navigation',
            tools: [
                {
                    name: 'open_tab',
                    description: 'Open a page URL in a new browser tab',
                    inputSchema: { type: 'object', properties: { url: { type: 'string' } } },
                },
                {
                    name: 'screenshot_page',
                    description: 'Capture a screenshot of the current page',
                    inputSchema: { type: 'object', properties: { fullPage: { type: 'boolean' } } },
                },
            ],
        });

        expect(derived.serverTags).toContain('browser');
        expect(derived.serverDisplayName).toContain('browser');
        expect(derived.tools[0]?.toolTags).toContain('browser');
        expect(derived.tools[0]?.semanticGroup).toBe('browser-automation');
        expect(derived.tools[0]?.advertisedName).toContain('open_tab');
    });

    it('propagates always-on status from the server to all derived tools', () => {
        const derived = deriveSemanticCatalogForServer({
            serverName: 'memory',
            description: 'Store and recall notes',
            alwaysOn: true,
            tools: [
                {
                    name: 'recall',
                    description: 'Recall memory entries',
                },
            ],
        });

        expect(derived.alwaysOn).toBe(true);
        expect(derived.tools[0]?.alwaysOn).toBe(true);
    });
});
