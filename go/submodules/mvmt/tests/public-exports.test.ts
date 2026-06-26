import { describe, expect, it } from 'vitest';

describe('public package exports', () => {
  it('keeps mount runtime exports and hides legacy connector helpers', async () => {
    const api = await import('../src/index.js');

    expect(api).toHaveProperty('start');
    expect(api).toHaveProperty('setupConfig');
    expect(api).toHaveProperty('ConfigSchema');
    expect(api).toHaveProperty('LocalFolderMountSchema');
    expect(api).toHaveProperty('ToolRouter');
    expect(api).toHaveProperty('createMcpServer');

    expect(api).not.toHaveProperty('addConnector');
    expect(api).not.toHaveProperty('getConnectorSetupStatuses');
    expect(api).not.toHaveProperty('listConnectors');
    expect(api).not.toHaveProperty('createProxyConnector');
    expect(api).not.toHaveProperty('upsertProxyConfig');
    expect(api).not.toHaveProperty('ProxySchema');
  });
});
