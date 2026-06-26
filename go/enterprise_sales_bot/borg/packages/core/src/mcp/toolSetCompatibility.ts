import type { CallToolResult } from '@modelcontextprotocol/sdk/types.js';
import type { SavedToolSetConfig } from '../interfaces/IConfigProvider.js';

interface ToolSetSessionLike {
    getLoadedToolNames(): string[];
    hasTool(name: string): boolean;
    loadToolIntoSession(name: string): { loaded: boolean; evicted: string[] };
}

interface ToolSetStoreLike {
    loadToolSets(): Promise<SavedToolSetConfig[]>;
    saveToolSet(toolSet: SavedToolSetConfig): Promise<unknown>;
}

export async function saveToolSetCompatibility(
    name: string,
    description: string | null,
    session: Pick<ToolSetSessionLike, 'getLoadedToolNames'>,
    saveToolSet: (toolSet: SavedToolSetConfig) => Promise<unknown>,
): Promise<CallToolResult> {
    const tools = session.getLoadedToolNames();
    if (tools.length === 0) {
        return {
            isError: true,
            content: [{ type: 'text', text: 'No tools currently loaded to save.' }],
        };
    }

    const toolSet: SavedToolSetConfig = {
        name,
        description,
        tools,
    };

    await saveToolSet(toolSet);
    return {
        isError: false,
        content: [{ type: 'text', text: `Tool Set '${toolSet.name}' saved with ${tools.length} tools.` }],
    };
}

export async function loadToolSetCompatibility(
    toolSetName: string,
    toolSets: SavedToolSetConfig[],
    session: Pick<ToolSetSessionLike, 'hasTool' | 'loadToolIntoSession'>,
): Promise<CallToolResult> {
    const toolSet = toolSets.find((candidate) => candidate.name === toolSetName);

    if (!toolSet) {
        return {
            isError: true,
            content: [{ type: 'text', text: `Tool Set '${toolSetName}' not found.` }],
        };
    }

    let loadedCount = 0;
    const missing: string[] = [];

    for (const toolName of toolSet.tools) {
        if (!session.hasTool(toolName)) {
            missing.push(toolName);
            continue;
        }

        const loadResult = session.loadToolIntoSession(toolName);
        if (loadResult.loaded) {
            loadedCount += 1;
        }
    }

    let message = `Loaded ${loadedCount} tools from set '${toolSet.name}'.`;
    if (missing.length > 0) {
        message += ` Warning: ${missing.length} tools could not be found (might be offline): ${missing.join(', ')}`;
    }

    return {
        isError: false,
        content: [{ type: 'text', text: message }],
    };
}

export function listToolSetsCompatibility(toolSets: SavedToolSetConfig[]): CallToolResult {
    return {
        isError: false,
        content: [{
            type: 'text',
            text: JSON.stringify(toolSets.map((toolSet) => ({
                name: toolSet.name,
                description: toolSet.description ?? '',
                tools: toolSet.tools,
                toolCount: toolSet.tools.length,
            })), null, 2),
        }],
    };
}

export async function executeCompatibleSaveToolSet(
    args: Record<string, unknown>,
    session: Pick<ToolSetSessionLike, 'getLoadedToolNames'>,
    toolSetStore: Pick<ToolSetStoreLike, 'saveToolSet'>,
): Promise<CallToolResult> {
    try {
        return await saveToolSetCompatibility(
            typeof args.name === 'string' ? args.name : '',
            typeof args.description === 'string' ? args.description : null,
            session,
            toolSetStore.saveToolSet,
        );
    } catch (error) {
        return {
            isError: true,
            content: [{ type: 'text', text: `Failed to save tool set: ${error instanceof Error ? error.message : String(error)}` }],
        };
    }
}

export async function executeCompatibleLoadToolSet(
    args: Record<string, unknown>,
    session: Pick<ToolSetSessionLike, 'hasTool' | 'loadToolIntoSession'>,
    toolSetStore: Pick<ToolSetStoreLike, 'loadToolSets'>,
): Promise<CallToolResult> {
    try {
        return await loadToolSetCompatibility(
            typeof args.name === 'string' ? args.name : '',
            await toolSetStore.loadToolSets(),
            session,
        );
    } catch (error) {
        return {
            isError: true,
            content: [{ type: 'text', text: `Failed to load tool set: ${error instanceof Error ? error.message : String(error)}` }],
        };
    }
}

export async function executeCompatibleListToolSets(
    toolSetStore: Pick<ToolSetStoreLike, 'loadToolSets'>,
): Promise<CallToolResult> {
    try {
        return listToolSetsCompatibility(await toolSetStore.loadToolSets());
    } catch (error) {
        return {
            isError: true,
            content: [{ type: 'text', text: `Failed to list tool sets: ${error instanceof Error ? error.message : String(error)}` }],
        };
    }
}
