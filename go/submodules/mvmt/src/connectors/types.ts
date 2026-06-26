export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: unknown;
}

export interface TextToolContent {
  type: 'text';
  text: string;
}

export interface ImageToolContent {
  type: 'image';
  data: string;
  mimeType: string;
}

export interface CallToolResult {
  content: Array<TextToolContent | ImageToolContent>;
  isError?: boolean;
}

export interface Connector {
  readonly id: string;
  readonly displayName: string;
  initialize(): Promise<void>;
  listTools(): Promise<ToolDefinition[]>;
  callTool(name: string, args: Record<string, unknown>): Promise<CallToolResult>;
  shutdown(): Promise<void>;
}
