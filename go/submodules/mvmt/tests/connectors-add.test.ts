import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  ConfigSchema,
  DEFAULT_MOUNT_EXCLUDE_PATTERNS,
  DEFAULT_MOUNT_PROTECT_PATTERNS,
  MvmtConfig,
} from '../src/config/schema.js';

const mocks = vi.hoisted(() => ({
  expandHome: vi.fn((value: string) => value),
  getConfigPath: vi.fn(),
  loadConfig: vi.fn(),
  saveConfig: vi.fn(),
  getConnectorSetupDefinition: vi.fn(),
  detect: vi.fn(),
  prompt: vi.fn(),
  apply: vi.fn(),
  isConfigured: vi.fn(),
}));

vi.mock('../src/config/loader.js', () => ({
  expandHome: mocks.expandHome,
  getConfigPath: mocks.getConfigPath,
  loadConfig: mocks.loadConfig,
  saveConfig: mocks.saveConfig,
}));

vi.mock('../src/connectors/setup-registry.js', () => ({
  getConnectorSetupDefinition: mocks.getConnectorSetupDefinition,
  getSetupRegistry: vi.fn(() => []),
}));

const { addConnector } = await import('../src/cli/connectors.js');

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

describe('addConnector', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'log').mockImplementation(() => {});
    mocks.getConfigPath.mockReturnValue('/tmp/mvmt-config.yaml');
    mocks.loadConfig.mockReturnValue(baseConfig);
    mocks.isConfigured.mockReturnValue(false);
  });

  it('uses the setup registry to detect, prompt, apply, and save connector config', async () => {
    const detected = { paths: ['/Users/me/project'] };
    const input = { paths: ['/Users/me/project'], writeAccess: false };
    const nextConfig: MvmtConfig = {
      ...baseConfig,
      mounts: [
        {
          name: 'project',
          type: 'local_folder',
          path: '/project',
          root: '/Users/me/project',
          description: '',
          guidance: '',
          exclude: [...DEFAULT_MOUNT_EXCLUDE_PATTERNS],
          protect: [...DEFAULT_MOUNT_PROTECT_PATTERNS],
          writeAccess: false,
          enabled: true,
        },
      ],
    };

    mocks.detect.mockResolvedValue(detected);
    mocks.prompt.mockResolvedValue(input);
    mocks.apply.mockReturnValue(nextConfig);
    mocks.getConnectorSetupDefinition.mockReturnValue({
      id: 'filesystem',
      displayName: 'Filesystem',
      isAddable: true,
      detect: mocks.detect,
      prompt: mocks.prompt,
      isConfigured: mocks.isConfigured,
      apply: mocks.apply,
    });

    await addConnector('filesystem');

    expect(mocks.detect).toHaveBeenCalledTimes(1);
    expect(mocks.prompt).toHaveBeenCalledWith(detected);
    expect(mocks.apply).toHaveBeenCalledWith(baseConfig, input);
    expect(mocks.saveConfig).toHaveBeenCalledWith('/tmp/mvmt-config.yaml', nextConfig);
    expect(() => ConfigSchema.parse(mocks.saveConfig.mock.calls[0][1])).not.toThrow();
  });
});
