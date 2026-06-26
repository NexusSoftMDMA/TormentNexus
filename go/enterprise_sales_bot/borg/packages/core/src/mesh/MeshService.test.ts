import { afterEach, describe, expect, it } from 'vitest';

import { MeshService, SwarmMessageType } from './MeshService.js';

describe('MeshService', () => {
    const instances: MeshService[] = [];

    afterEach(() => {
        while (instances.length > 0) {
            instances.pop()?.destroy();
        }
    });

    it('correlates direct request/response messages', async () => {
        const requester = new MeshService();
        const responder = new MeshService();
        instances.push(requester, responder);

        responder.on('message', (message) => {
            if (message.type === SwarmMessageType.CAPABILITY_QUERY) {
                responder.sendResponse(message, SwarmMessageType.CAPABILITY_RESPONSE, {
                    capabilities: ['coder', 'research'],
                });
            }
        });

        const response = await requester.request(
            responder.nodeId,
            SwarmMessageType.CAPABILITY_QUERY,
            { topic: 'capabilities' },
            {
                timeoutMs: 500,
                responseTypes: [SwarmMessageType.CAPABILITY_RESPONSE],
            },
        );

        expect(response.type).toBe(SwarmMessageType.CAPABILITY_RESPONSE);
        expect(response.target).toBe(requester.nodeId);
        expect(response.sender).toBe(responder.nodeId);
        expect(response.correlationId).toBeDefined();
        expect(response.payload).toEqual({
            capabilities: ['coder', 'research'],
        });
    });

    it('rejects request when no matching response arrives before timeout', async () => {
        const requester = new MeshService();
        const responder = new MeshService();
        instances.push(requester, responder);

        await expect(requester.request(
            responder.nodeId,
            SwarmMessageType.DIRECT_MESSAGE,
            { ping: true },
            {
                timeoutMs: 20,
                responseTypes: [SwarmMessageType.TASK_RESULT],
            },
        )).rejects.toThrow(/timed out/i);
    });
});
