import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StreamableHTTPClientTransport } from '@modelcontextprotocol/sdk/client/streamableHttp.js';
import { CallToolResult, Connector, ToolDefinition } from './types.js';
import { sanitizeName } from './proxy-stdio.js';
import { createProxyToolPolicy, writeAccessDisabledResult } from './write-policy.js';

export interface HttpProxyConfig {
  name: string;
  url: string;
  env: Record<string, string>;
  writeAccess?: boolean;
}

export class HttpProxyConnector implements Connector {
  readonly id: string;
  readonly displayName: string;
  private client?: Client;
  private transport?: StreamableHTTPClientTransport;
  private tools: ToolDefinition[] = [];
  private readonly toolAllowed: (name: string) => boolean;

  constructor(private readonly config: HttpProxyConfig) {
    this.id = `proxy_${sanitizeName(config.name)}`;
    this.displayName = config.name;
    this.toolAllowed = createProxyToolPolicy(config);
  }

  async initialize(): Promise<void> {
    this.transport = new StreamableHTTPClientTransport(new URL(this.config.url));

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
      // The connection may already be closed.
    }
  }
}
