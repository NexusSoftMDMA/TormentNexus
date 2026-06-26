import { describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  connect: vi.fn(),
  listTools: vi.fn(),
  callTool: vi.fn(),
  close: vi.fn(),
  transports: [] as Array<{ command: string; args?: string[]; env?: Record<string, string> }>,
  httpTransports: [] as Array<{ url: string }>,
}));

vi.mock('@modelcontextprotocol/sdk/client/index.js', () => ({
  Client: vi.fn().mockImplementation(() => ({
    connect: mocks.connect,
    listTools: mocks.listTools,
    callTool: mocks.callTool,
    close: mocks.close,
  })),
}));

vi.mock('@modelcontextprotocol/sdk/client/stdio.js', () => ({
  StdioClientTransport: vi.fn().mockImplementation((params) => {
    mocks.transports.push(params);
    return { params };
  }),
}));

vi.mock('@modelcontextprotocol/sdk/client/streamableHttp.js', () => ({
  StreamableHTTPClientTransport: vi.fn().mockImplementation((url: URL) => {
    mocks.httpTransports.push({ url: url.toString() });
    return { url };
  }),
}));

const { StdioProxyConnector, sanitizeName, buildChildEnv } = await import('../src/connectors/proxy-stdio.js');
const { HttpProxyConnector } = await import('../src/connectors/proxy-http.js');
const { createProxyConnector } = await import('../src/connectors/factory.js');
const { isLikelyWriteTool } = await import('../src/connectors/write-policy.js');

describe('sanitizeName', () => {
  it('normalizes connector names for namespace prefixes', () => {
    expect(sanitizeName('GitHub Server')).toBe('github_server');
    expect(sanitizeName('@weird/name')).toBe('_weird_name');
    expect(sanitizeName('')).toBe('server');
  });
});

describe('buildChildEnv', () => {
  it('drops unsafe parent env vars and keeps the allowlist plus config overrides', () => {
    const original = { ...process.env };
    try {
      process.env.PATH = '/usr/bin';
      process.env.HOME = '/home/test';
      process.env.OPENAI_API_KEY = 'leak-me';
      process.env.AWS_SECRET_ACCESS_KEY = 'leak-me-too';

      const env = buildChildEnv({ CUSTOM: 'ok' });

      expect(env.PATH).toBe('/usr/bin');
      expect(env.HOME).toBe('/home/test');
      expect(env.CUSTOM).toBe('ok');
      expect(env.OPENAI_API_KEY).toBeUndefined();
      expect(env.AWS_SECRET_ACCESS_KEY).toBeUndefined();
    } finally {
      process.env = original;
    }
  });
});

describe('isLikelyWriteTool', () => {
  it('flags common write verbs and leaves reads alone', () => {
    expect(isLikelyWriteTool('write_file')).toBe(true);
    expect(isLikelyWriteTool('create_issue')).toBe(true);
    expect(isLikelyWriteTool('delete_row')).toBe(true);
    expect(isLikelyWriteTool('update_doc')).toBe(true);
    expect(isLikelyWriteTool('read_file')).toBe(false);
    expect(isLikelyWriteTool('search_notes')).toBe(false);
    expect(isLikelyWriteTool('list_tags')).toBe(false);
  });
});

describe('StdioProxyConnector', () => {
  it('initializes, caches tools, forwards calls, and closes the client', async () => {
    mocks.connect.mockResolvedValue(undefined);
    mocks.listTools.mockResolvedValue({
      tools: [
        {
          name: 'read_file',
          description: 'Read a file',
          inputSchema: { type: 'object', properties: { path: { type: 'string' } } },
        },
      ],
    });
    mocks.callTool.mockResolvedValue({ content: [{ type: 'text', text: 'ok' }] });
    mocks.close.mockResolvedValue(undefined);

    const connector = new StdioProxyConnector({
      name: 'Filesystem',
      command: 'npx',
      args: ['server'],
      env: { TEST_ENV: '1' },
      enabled: true,
    });

    expect(connector.id).toBe('proxy_filesystem');
    await connector.initialize();

    expect(mocks.transports.at(-1)).toMatchObject({
      command: 'npx',
      args: ['server'],
      env: expect.objectContaining({ TEST_ENV: '1' }),
    });
    await expect(connector.listTools()).resolves.toEqual([
      {
        name: 'read_file',
        description: 'Read a file',
        inputSchema: { type: 'object', properties: { path: { type: 'string' } } },
      },
    ]);

    await expect(connector.callTool('read_file', { path: '/tmp/a' })).resolves.toEqual({
      content: [{ type: 'text', text: 'ok' }],
    });
    expect(mocks.callTool).toHaveBeenCalledWith({ name: 'read_file', arguments: { path: '/tmp/a' } });

    await connector.shutdown();
    expect(mocks.close).toHaveBeenCalled();
  });

  it('rejects calls before initialization', async () => {
    const connector = new StdioProxyConnector({
      name: 'Filesystem',
      command: 'npx',
      args: [],
      env: {},
      enabled: true,
    });

    await expect(connector.callTool('read_file', {})).rejects.toThrow('Connector not initialized');
  });

  it('filters and blocks filesystem write tools unless write access is explicit', async () => {
    mocks.connect.mockResolvedValue(undefined);
    mocks.listTools.mockResolvedValue({
      tools: [
        {
          name: 'read_file',
          description: 'Read a file',
          inputSchema: { type: 'object', properties: {} },
        },
        {
          name: 'write_file',
          description: 'Write a file',
          inputSchema: { type: 'object', properties: {} },
        },
        {
          name: 'create_directory',
          description: 'Create a directory',
          inputSchema: { type: 'object', properties: {} },
        },
      ],
    });

    const connector = new StdioProxyConnector({
      name: 'filesystem',
      command: 'npx',
      args: ['-y', '@modelcontextprotocol/server-filesystem', '/tmp'],
      env: {},
    });

    await connector.initialize();

    await expect(connector.listTools()).resolves.toEqual([
      {
        name: 'read_file',
        description: 'Read a file',
        inputSchema: { type: 'object', properties: {} },
      },
    ]);
    await expect(connector.callTool('write_file', { path: '/tmp/file', content: 'nope' })).resolves.toMatchObject({
      isError: true,
      content: [{ type: 'text', text: expect.stringContaining('write access is disabled') }],
    });
  });

  it('filters write-like tools on non-filesystem proxies when writeAccess is false', async () => {
    mocks.connect.mockResolvedValue(undefined);
    mocks.listTools.mockResolvedValue({
      tools: [
        { name: 'search_issues', description: '', inputSchema: { type: 'object', properties: {} } },
        { name: 'create_issue', description: '', inputSchema: { type: 'object', properties: {} } },
        { name: 'delete_issue', description: '', inputSchema: { type: 'object', properties: {} } },
      ],
    });

    const connector = new StdioProxyConnector({
      name: 'github',
      command: 'npx',
      args: ['@modelcontextprotocol/server-github'],
      env: {},
      writeAccess: false,
    });

    await connector.initialize();

    await expect(connector.listTools()).resolves.toEqual([
      { name: 'search_issues', description: '', inputSchema: { type: 'object', properties: {} } },
    ]);
    await expect(connector.callTool('create_issue', {})).resolves.toMatchObject({
      isError: true,
      content: [{ type: 'text', text: expect.stringContaining('write access is disabled') }],
    });
  });

  it('allows filesystem write tools when write access is explicit', async () => {
    mocks.connect.mockResolvedValue(undefined);
    mocks.listTools.mockResolvedValue({
      tools: [
        {
          name: 'write_file',
          description: 'Write a file',
          inputSchema: { type: 'object', properties: {} },
        },
      ],
    });
    mocks.callTool.mockResolvedValue({ content: [{ type: 'text', text: 'wrote' }] });

    const connector = new StdioProxyConnector({
      name: 'filesystem',
      command: 'npx',
      args: ['-y', '@modelcontextprotocol/server-filesystem', '/tmp'],
      env: {},
      writeAccess: true,
    });

    await connector.initialize();

    await expect(connector.listTools()).resolves.toHaveLength(1);
    await expect(connector.callTool('write_file', { path: '/tmp/file', content: 'ok' })).resolves.toEqual({
      content: [{ type: 'text', text: 'wrote' }],
    });
  });
});

describe('HttpProxyConnector', () => {
  it('initializes against an HTTP MCP URL', async () => {
    mocks.connect.mockResolvedValue(undefined);
    mocks.listTools.mockResolvedValue({
      tools: [
        {
          name: 'remote_tool',
          description: 'Remote tool',
          inputSchema: { type: 'object', properties: {} },
        },
      ],
    });
    mocks.callTool.mockResolvedValue({ content: [{ type: 'text', text: 'remote ok' }] });

    const connector = new HttpProxyConnector({
      name: 'Supabase',
      url: 'https://mcp.example.test/mcp',
      env: {},
    });

    expect(connector.id).toBe('proxy_supabase');
    await connector.initialize();

    expect(mocks.httpTransports.at(-1)).toEqual({ url: 'https://mcp.example.test/mcp' });
    await expect(connector.listTools()).resolves.toEqual([
      {
        name: 'remote_tool',
        description: 'Remote tool',
        inputSchema: { type: 'object', properties: {} },
      },
    ]);

    await expect(connector.callTool('remote_tool', {})).resolves.toEqual({
      content: [{ type: 'text', text: 'remote ok' }],
    });
    await expect(connector.shutdown()).resolves.toBeUndefined();
  });

  it('filters and blocks HTTP proxy write tools when writeAccess is false', async () => {
    mocks.connect.mockResolvedValue(undefined);
    mocks.listTools.mockResolvedValue({
      tools: [
        { name: 'search_records', description: '', inputSchema: { type: 'object', properties: {} } },
        { name: 'create_record', description: '', inputSchema: { type: 'object', properties: {} } },
        { name: 'delete_record', description: '', inputSchema: { type: 'object', properties: {} } },
      ],
    });

    const connector = createProxyConnector({
      name: 'remote',
      transport: 'http',
      url: 'https://mcp.example.test/mcp',
      args: [],
      env: {},
      enabled: true,
      writeAccess: false,
    });

    expect(connector).toBeInstanceOf(HttpProxyConnector);
    await connector!.initialize();

    await expect(connector!.listTools()).resolves.toEqual([
      { name: 'search_records', description: '', inputSchema: { type: 'object', properties: {} } },
    ]);

    mocks.callTool.mockClear();
    await expect(connector!.callTool('create_record', {})).resolves.toMatchObject({
      isError: true,
      content: [{ type: 'text', text: expect.stringContaining('write access is disabled') }],
    });
    expect(mocks.callTool).not.toHaveBeenCalled();
  });
});

describe('createProxyConnector', () => {
  it('creates stdio and HTTP proxy connectors from config', () => {
    expect(
      createProxyConnector({
        name: 'filesystem',
        transport: 'stdio',
        command: 'npx',
        args: ['server'],
        env: {},
        enabled: true,
      })?.id,
    ).toBe('proxy_filesystem');

    expect(
      createProxyConnector({
        name: 'remote',
        transport: 'http',
        url: 'https://mcp.example.test/mcp',
        args: [],
        env: {},
        enabled: true,
      })?.id,
    ).toBe('proxy_remote');
  });

  it('returns undefined for incomplete proxy config', () => {
    expect(
      createProxyConnector({
        name: 'bad',
        transport: 'stdio',
        args: [],
        env: {},
        enabled: true,
      } as any),
    ).toBeUndefined();
  });
});
