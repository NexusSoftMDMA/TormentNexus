export interface McpServerConfig {
    name: string;
    type: 'stdio' | 'sse';
    command?: string;
    args?: string[];
    env?: Record<string, string>;
    url?: string;
    disabled?: boolean;
}

export interface SavedScriptConfig {
    uuid?: string;
    name: string;
    code: string;
    description?: string | null;
}

export interface SavedToolSetConfig {
    uuid?: string;
    name: string;
    tools: string[];
    description?: string | null;
}

export interface IConfigProvider {
    loadMcpServers(): Promise<McpServerConfig[]>;
    saveMcpServers(servers: McpServerConfig[]): Promise<void>;
    getSettings(): Promise<Record<string, any>>;
    loadAlwaysVisibleTools(): Promise<string[]>;
    saveAlwaysVisibleTools(toolNames: string[]): Promise<string[]>;
    loadScripts(): Promise<SavedScriptConfig[]>;
    saveScript(script: SavedScriptConfig): Promise<SavedScriptConfig>;
    deleteScript(nameOrUuid: string): Promise<void>;
    loadToolSets(): Promise<SavedToolSetConfig[]>;
    saveToolSet(toolSet: SavedToolSetConfig): Promise<SavedToolSetConfig>;
    deleteToolSet(nameOrUuid: string): Promise<void>;
}
