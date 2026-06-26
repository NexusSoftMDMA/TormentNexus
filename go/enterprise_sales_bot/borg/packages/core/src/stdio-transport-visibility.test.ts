import { beforeEach, describe, expect, it, vi } from 'vitest';

const connectMock = vi.fn();
const closeMock = vi.fn();
const transportCloseMock = vi.fn();
const transportInstances: Array<{
    params: Record<string, unknown>;
    stderrOn: ReturnType<typeof vi.fn>;
    stdoutOn: ReturnType<typeof vi.fn>;
}> = [];

vi.mock('@modelcontextprotocol/sdk/client/index.js', () => {
    class MockClient {
        connect = connectMock;
        close = closeMock;
        listTools = vi.fn().mockResolvedValue({ tools: [] });
        callTool = vi.fn();
        constructor(_info: unknown, _options: unknown) {}
    }

    return { Client: MockClient };
});

vi.mock('./transports/process-managed.transport.js', () => {
    class MockProcessManagedStdioTransport {
        public readonly stderr = { on: vi.fn() };
        public readonly stdout = { on: vi.fn() };

        constructor(public readonly params: Record<string, unknown>) {
            transportInstances.push({
                params,
                stderrOn: this.stderr.on,
                stdoutOn: this.stdout.on,
            });
        }

        async close(): Promise<void> {
            transportCloseMock();
        }
    }

    return { ProcessManagedStdioTransport: MockProcessManagedStdioTransport };
});

vi.mock('./services/log-store.service.js', () => ({
    tormentnexusLogStore: {
        addLog: vi.fn(),
    },
}));

describe('managed stdio transport visibility', () => {
    beforeEach(() => {
        connectMock.mockReset().mockResolvedValue(undefined);
        closeMock.mockReset().mockResolvedValue(undefined);
        transportCloseMock.mockReset();
        transportInstances.length = 0;
    });

    it.skip('uses the managed transport for legacy Router connections', async () => {
        const { Router } = await import('./Router.js');
        const router = new Router();

        await router.connectToServer('demo-router', 'node', ['demo.js']);

        expect(connectMock).toHaveBeenCalledTimes(1);
        expect(transportInstances).toHaveLength(1);
        expect(transportInstances[0]?.params).toMatchObject({
            command: 'node',
            args: ['demo.js'],
            stderr: 'pipe',
        });
        expect(transportInstances[0]?.stderrOn).toHaveBeenCalledWith('data', expect.any(Function));
        expect(transportInstances[0]?.stderrOn).toHaveBeenCalledWith('error', expect.any(Function));
        expect(transportInstances[0]?.stdoutOn).toHaveBeenCalledWith('data', expect.any(Function));
        expect(transportInstances[0]?.stdoutOn).toHaveBeenCalledWith('error', expect.any(Function));
    });

    it('uses the managed transport for aggregated stdio clients and closes both transport and client', async () => {
        const { StdioClient } = await import('./mcp/StdioClient.js');
        const client = new StdioClient('demo-aggregator', {
            command: 'python',
            args: ['server.py'],
            env: { TEST_ENV: '1' },
            enabled: true,
        });

        await client.connect();
        await client.close();

        expect(connectMock).toHaveBeenCalledTimes(1);
        expect(transportInstances).toHaveLength(1);
        expect(transportInstances[0]?.params).toMatchObject({
            command: 'python',
            args: ['server.py'],
            stderr: 'pipe',
        });
        expect(transportInstances[0]?.params.env).toMatchObject({ TEST_ENV: '1' });
        expect(transportInstances[0]?.stderrOn).toHaveBeenCalledWith('data', expect.any(Function));
        expect(transportInstances[0]?.stdoutOn).toHaveBeenCalledWith('data', expect.any(Function));
        expect(transportCloseMock).toHaveBeenCalledTimes(1);
        expect(closeMock).toHaveBeenCalledTimes(1);
    });
});