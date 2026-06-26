import { describe, it, expect, vi } from 'vitest';
import { MCPAggregator } from './MCPAggregator.js';
import { StdioClient } from './StdioClient.js';
import { SSEClient } from './SSEClient.js';
import type { MCPServerConfig } from './types.js';

vi.mock('./StdioClient.js', () => ({
    StdioClient: vi.fn(),
}));

vi.mock('./SSEClient.js', () => ({
    SSEClient: vi.fn(),
}));

vi.mock('./configStore.js', () => ({
    MCPConfigStore: vi.fn().mockImplementation(function() { return {
        getServers: vi.fn().mockReturnValue({}),
        addServer: vi.fn(),
        updateServer: vi.fn(),
        removeServer: vi.fn(),
    }; }),
}));

vi.mock('./trafficInspector.js', () => ({
    MCPTrafficInspector: vi.fn(),
}));

describe('MCPAggregator Factory', () => {
    it('should create an SSEClient when type is SSE', () => {
        // Instantiate without explicitly overriding createClient
        const aggregator = new MCPAggregator({ configPath: 'mock.json' });

        // Access private createClient via any casting for testing
        const createClient = (aggregator as any).createClient;

        const config: MCPServerConfig = {
            enabled: true,
            type: 'SSE',
            url: 'http://localhost:8080/sse'
        };

        const client = createClient('test-server', config);

        expect(SSEClient).toHaveBeenCalledWith('test-server', config);
        // We cannot strict equal to instance of SSEClient due to mock, but checking the mock call is sufficient
    });

    it('should create an SSEClient when url is present', () => {
        const aggregator = new MCPAggregator({ configPath: 'mock.json' });
        const createClient = (aggregator as any).createClient;

        const config: MCPServerConfig = {
            enabled: true,
            url: 'http://localhost:8080/sse'
        };

        createClient('test-server', config);

        expect(SSEClient).toHaveBeenCalledWith('test-server', config);
    });

    it('should create a StdioClient by default for backwards compatibility', () => {
        const aggregator = new MCPAggregator({ configPath: 'mock.json' });
        const createClient = (aggregator as any).createClient;

        const config: MCPServerConfig = {
            enabled: true,
            command: 'npx',
            args: ['test']
        };

        createClient('test-server', config);

        expect(StdioClient).toHaveBeenCalledWith('test-server', expect.objectContaining({
            command: 'npx',
            args: ['test']
        }));
    });
});
