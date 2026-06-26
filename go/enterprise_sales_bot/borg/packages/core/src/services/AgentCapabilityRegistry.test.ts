import { afterEach, describe, expect, it } from 'vitest';

import { MeshService, SwarmMessageType } from '../mesh/MeshService.js';
import { AgentCapabilityRegistry } from './AgentCapabilityRegistry.js';

describe('AgentCapabilityRegistry', () => {
    const instances: MeshService[] = [];

    afterEach(() => {
        while (instances.length > 0) {
            instances.pop()?.destroy();
        }
    });

    it('discovers and caches peer capabilities over the mesh', async () => {
        const requester = new MeshService();
        const responder = new MeshService();
        instances.push(requester, responder);

        responder.on('message', (message) => {
            if (message.type === SwarmMessageType.CAPABILITY_QUERY) {
                responder.sendResponse(message, SwarmMessageType.CAPABILITY_RESPONSE, {
                    role: 'worker',
                    capabilities: ['research', 'git'],
                    load: 0.25,
                });
            }
        });

        (requester as unknown as { knownNodes: Set<string> }).knownNodes.add(responder.nodeId);

        const registry = new AgentCapabilityRegistry(requester, {
            cacheTtlMs: 60_000,
            now: () => 1_000,
        });

        const discovered = await registry.discoverCapabilities(responder.nodeId, 500);
        expect(discovered).toMatchObject({
            role: 'worker',
            capabilities: ['research', 'git'],
            load: 0.25,
            cachedAt: 1_000,
        });

        expect(registry.getCachedCapabilities(responder.nodeId)).toMatchObject({
            capabilities: ['research', 'git'],
        });
    });

    it('finds a peer matching required capabilities', async () => {
        const requester = new MeshService();
        const responder = new MeshService();
        instances.push(requester, responder);

        responder.on('message', (message) => {
            if (message.type === SwarmMessageType.CAPABILITY_QUERY) {
                responder.sendResponse(message, SwarmMessageType.CAPABILITY_RESPONSE, {
                    role: 'coder',
                    capabilities: ['typescript', 'git', 'mcp'],
                    load: 0.1,
                });
            }
        });

        (requester as unknown as { knownNodes: Set<string> }).knownNodes.add(responder.nodeId);

        const registry = new AgentCapabilityRegistry(requester);
        const match = await registry.findPeerForCapabilities(['git', 'mcp'], 500);

        expect(match).toMatchObject({
            nodeId: responder.nodeId,
            role: 'coder',
        });
        expect(match?.capabilities).toEqual(expect.arrayContaining(['git', 'mcp']));
    });
});
