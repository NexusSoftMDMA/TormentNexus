import { describe, expect, it } from 'vitest';
import { getConnectorSetupStatuses } from '../src/cli/connectors.js';
import { MvmtConfig } from '../src/config/schema.js';
import { upsertProxyConfig } from '../src/connectors/setup-utils.js';

const baseConfig: MvmtConfig = {
  version: 1,
  server: {
    port: 4141,
    allowedOrigins: [],
    access: 'local',
  },
  proxy: [],
  mounts: [],
};

describe('connector setup helpers', () => {
  it('reports unconfigured local connector setups', () => {
    const statuses = getConnectorSetupStatuses(baseConfig);

    expect(statuses.map((status) => [status.name, status.configured, status.addable])).toEqual([
      ['filesystem', false, false],
    ]);
  });

  it('reports configured local connector setups', () => {
    const statuses = getConnectorSetupStatuses({
      ...baseConfig,
      mounts: [
        {
          name: 'workspace',
          type: 'local_folder',
          path: '/workspace',
          root: '/workspace',
          description: '',
          guidance: '',
          exclude: [],
          protect: [],
          writeAccess: false,
          enabled: true,
        },
      ],
    });

    expect(statuses.map((status) => [status.name, status.configured])).toEqual([
      ['filesystem', true],
    ]);
  });

  it('upserts proxy config by connector name', () => {
    const config = upsertProxyConfig(
      {
        ...baseConfig,
        proxy: [
          {
            name: 'Search',
            transport: 'stdio',
            command: '/old/python',
            args: [],
            env: {},
            writeAccess: false,
            enabled: true,
          },
        ],
      },
      {
        name: 'search',
        transport: 'stdio',
        command: '/new/python',
        args: ['server'],
        env: {},
        writeAccess: true,
        enabled: true,
      },
    );

    expect(config.proxy).toEqual([
      {
        name: 'search',
        transport: 'stdio',
        command: '/new/python',
        args: ['server'],
        env: {},
        writeAccess: true,
        enabled: true,
      },
    ]);
  });
});
