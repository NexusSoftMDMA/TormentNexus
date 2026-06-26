import { describe, it, expect, vi, beforeEach } from 'vitest';
import { SSEClient } from './SSEClient.js';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { SSEClientTransport } from '@modelcontextprotocol/sdk/client/sse.js';

// Mock the underlying SDK modules
vi.mock('@modelcontextprotocol/sdk/client/index.js', () => {
    return {
        Client: vi.fn().mockImplementation(function() { return {
            connect: vi.fn().mockResolvedValue(undefined),
            close: vi.fn().mockResolvedValue(undefined),
            listTools: vi.fn().mockResolvedValue({ tools: [{ name: 'test-tool' }] }),
            listPrompts: vi.fn().mockResolvedValue({ prompts: [{ name: 'test-prompt' }] }),
            listResources: vi.fn().mockResolvedValue({ resources: [{ uri: 'test-resource' }] }),
            callTool: vi.fn().mockResolvedValue({ content: 'result' }),
            ping: vi.fn().mockResolvedValue(true),
        }; }),
    };
});

vi.mock('@modelcontextprotocol/sdk/client/sse.js', () => {
    return {
        SSEClientTransport: vi.fn(),
    };
});

describe('SSEClient', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should throw an error if URL is missing', async () => {
        const client = new SSEClient('test-server', { enabled: true, type: 'SSE' });
        await expect(client.connect()).rejects.toThrow(/missing URL/);
    });

    it('should connect using SSEClientTransport without headers', async () => {
        const client = new SSEClient('test-server', {
            enabled: true,
            type: 'SSE',
            url: 'http://localhost:8080/sse',
        });

        await client.connect();

        expect(SSEClientTransport).toHaveBeenCalledWith(new URL('http://localhost:8080/sse'));
        expect(Client).toHaveBeenCalled();
    });

    it('should connect using SSEClientTransport with bearer token and headers', async () => {
        const client = new SSEClient('test-server', {
            enabled: true,
            type: 'SSE',
            url: 'http://localhost:8080/sse',
            bearerToken: 'my-token',
            headers: { 'X-Custom': 'value' },
        });

        await client.connect();

        expect(SSEClientTransport).toHaveBeenCalledWith(
            new URL('http://localhost:8080/sse'),
            expect.objectContaining({
                eventSourceInit: expect.objectContaining({
                    headers: expect.objectContaining({
                        Authorization: 'Bearer my-token',
                        'X-Custom': 'value',
                    }),
                }),
            })
        );
    });

    it('should route listTools and callTool to the underlying client', async () => {
        const client = new SSEClient('test-server', {
            enabled: true,
            type: 'SSE',
            url: 'http://localhost:8080/sse',
        });

        await client.connect();

        const tools = await client.listTools();
        expect(tools).toEqual([{ name: 'test-tool' }]);

        const result = await client.callTool('test-tool', { arg: 'val' });
        expect(result).toEqual({ content: 'result' });
    });
});
