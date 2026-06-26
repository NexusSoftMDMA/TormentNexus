export interface MCPServerConfig {
    command: string;
    args: string[];
    enabled: boolean;
    env?: Record<string, string>;
    alwaysOn?: boolean;
}

export interface MCPToolDefinition {
    name: string;
    description?: string;
    inputSchema?: unknown;
    [key: string]: unknown;
}

export interface MCPAggregatedTool extends MCPToolDefinition {
    name: string;
    server: string;
    _originalName: string;
    serverDisplayName?: string;
    advertisedName?: string;
    serverTags?: string[];
    toolTags?: string[];
    semanticGroup?: string;
    semanticGroupLabel?: string;
    keywords?: string[];
    alwaysOn?: boolean;
}

export interface MCPClientLike {
    connect(): Promise<void>;
    listTools(): Promise<MCPToolDefinition[]>;
    callTool(toolName: string, args: unknown): Promise<unknown>;
    close(): Promise<void>;
}

export type MCPServerStatus = 'stopped' | 'connecting' | 'connected' | 'error' | 'restarting';

export type MCPServerWarmupStatus = 'idle' | 'scheduled' | 'warming' | 'ready' | 'failed';

export interface MCPServerState {
    name: string;
    config: MCPServerConfig;
    status: MCPServerStatus;
    toolCount: number;
    advertisedToolCount?: number;
    advertisedAlwaysOn?: boolean;
    advertisedSource?: 'database' | 'config' | 'empty';
    warmupStatus?: MCPServerWarmupStatus;
    lastConnectedAt?: number;
    lastError?: string;
  lastErrorAt?: number;
    restartCount: number;
}

export interface MCPTrafficEvent {
    server: string;
    method: 'tools/list' | 'tools/call';
    paramsSummary: string;
    latencyMs: number;
    success: boolean;
    timestamp: number;
    toolName?: string;
    error?: string;
}

export interface MCPAggregatorOptions {
    configPath?: string;
    createClient?: (name: string, config: MCPServerConfig) => MCPClientLike;
    restartDelayMs?: number;
    now?: () => number;
    maxTrafficEvents?: number;
    /**
     * When true, listAggregatedTools() will only return tools from already-connected
     * servers rather than eagerly spawning all configured server processes. Tool
     * execution (executeTool) still connects on-demand regardless of this flag.
     * This prevents all MCP binaries from launching at tool-listing time in lazy mode.
     */
    lazyMode?: boolean;
}
