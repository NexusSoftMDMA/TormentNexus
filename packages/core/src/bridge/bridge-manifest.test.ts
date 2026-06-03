import { describe, expect, it } from 'vitest';

import {
    applyBridgeClientHello,
    buildBridgeManifest,
    createDefaultBridgeClient,
} from './bridge-manifest.js';

describe('bridge manifest helpers', () => {
    it('creates default bridge clients and applies hello metadata', () => {
        const initial = createDefaultBridgeClient('client-1', 1000);
        const registered = applyBridgeClientHello(initial, {
            clientType: 'vscode-extension',
            clientName: 'TormentNexus VS Code Bridge',
            version: '0.99.1',
            platform: 'VS Code 1.99',
            capabilities: ['chat.inject', 'editor.selection.read', 'chat.inject'],
            hookPhases: ['chat.submit', 'editor.selection', 'chat.submit'],
        }, 2000);

        expect(registered).toMatchObject({
            clientId: 'client-1',
            clientType: 'vscode-extension',
            clientName: 'TormentNexus VS Code Bridge',
            version: '0.99.1',
            platform: 'VS Code 1.99',
            capabilities: ['chat.inject', 'editor.selection.read'],
            hookPhases: ['chat.submit', 'editor.selection'],
            connectedAt: 1000,
            lastSeenAt: 2000,
        });
    });

    it('builds a stable manifest with sorted connected clients', () => {
        const alpha = applyBridgeClientHello(createDefaultBridgeClient('a', 1000), {
            clientType: 'browser-extension',
            clientName: 'TormentNexus Browser Bridge',
        }, 1500);
        const beta = applyBridgeClientHello(createDefaultBridgeClient('b', 1000), {
            clientType: 'vscode-extension',
            clientName: 'TormentNexus VS Code Bridge',
        }, 1500);

        const manifest = buildBridgeManifest([beta, alpha]);

        expect(manifest.protocolVersion).toBe('2026-03-10');
        expect(manifest.supportedCapabilities).toContain('bridge.websocket');
        expect(manifest.supportedHookPhases).toContain('chat.submit');
        expect(manifest.connectedClients.map((client) => client.clientName)).toEqual([
            'TormentNexus Browser Bridge',
            'TormentNexus VS Code Bridge',
        ]);
    });
});
