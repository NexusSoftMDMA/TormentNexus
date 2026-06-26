import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';
import { CallToolResult, Connector, ToolDefinition } from './types.js';
import { createProxyToolPolicy, writeAccessDisabledResult } from './write-policy.js';

export interface StdioProxyConfig {
  name: string;
  command: string;
  args: string[];
  env: Record<string, string>;
  writeAccess?: boolean;
}

export class StdioProxyConnector implements Connector {
  readonly id: string;
  readonly displayName: string;
  private client?: Client;
  private transport?: StdioClientTransport;
  private tools: ToolDefinition[] = [];
  private readonly toolAllowed: (name: string) => boolean;

  constructor(private readonly config: StdioProxyConfig) {
    this.id = `proxy_${sanitizeName(config.name)}`;
    this.displayName = config.name;
    this.toolAllowed = createProxyToolPolicy(config);
  }

  async initialize(): Promise<void> {
    this.transport = new StdioClientTransport({
      command: this.config.command,
      args: this.config.args,
      env: buildChildEnv(this.config.env),
    });

    this.client = new Client(
      { name: `mvmt-proxy-${this.config.name}`, version: '0.1.0' },
      { capabilities: {} },
    );

    await this.client.connect(this.transport);

    const result = await this.client.listTools();
    this.tools = result.tools
      .map((tool) => ({
        name: tool.name,
        description: tool.description || '',
        inputSchema: tool.inputSchema,
      }))
      .filter((tool) => this.toolAllowed(tool.name));
  }

  async listTools(): Promise<ToolDefinition[]> {
    return this.tools;
  }

  async callTool(name: string, args: Record<string, unknown>): Promise<CallToolResult> {
    if (!this.client) throw new Error(`Connector not initialized: ${this.displayName}`);
    if (!this.toolAllowed(name)) {
      return writeAccessDisabledResult(name, this.displayName);
    }
    const result = await this.client.callTool({ name, arguments: args });
    return result as CallToolResult;
  }

  async shutdown(): Promise<void> {
    try {
      await this.client?.close();
    } catch {
      // The child process may already be gone.
    }
  }

}

export function sanitizeName(name: string): string {
  const sanitized = name.replace(/[^a-zA-Z0-9_]/g, '_').toLowerCase();
  return sanitized || 'server';
}

const SAFE_ENV_KEYS = new Set([
  'PATH',
  'HOME',
  'USER',
  'LOGNAME',
  'SHELL',
  'LANG',
  'LC_ALL',
  'LC_CTYPE',
  'LC_MESSAGES',
  'TZ',
  'TERM',
  'TMPDIR',
  'PWD',
  'SYSTEMROOT',
  'APPDATA',
  'LOCALAPPDATA',
  'PROGRAMFILES',
  'PROGRAMDATA',
  'USERPROFILE',
  'COMSPEC',
]);

export function buildChildEnv(configEnv: Record<string, string>): Record<string, string> {
  const base: Record<string, string> = {};
  for (const [key, value] of Object.entries(process.env)) {
    if (value === undefined) continue;
    if (SAFE_ENV_KEYS.has(key)) base[key] = value;
  }
  return { ...base, ...configEnv };
}
