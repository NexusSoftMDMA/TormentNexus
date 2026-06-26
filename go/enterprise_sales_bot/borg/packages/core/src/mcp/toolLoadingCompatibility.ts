import type { CallToolResult, Tool } from '@modelcontextprotocol/sdk/types.js';

interface WorkingSetLike {
    loadTool(name: string): string[];
    hydrateTool(name: string): string[];
    unloadTool(name: string): boolean;
    getLimits(): { maxLoadedTools: number; maxHydratedSchemas: number };
    listLoadedTools(): unknown;
}

function createTextResult(text: string, isError = false): CallToolResult {
    return {
        content: [{ type: 'text', text }],
        isError,
    };
}

export async function executeSearchToolsCompatibility<T>(
    args: Record<string, unknown>,
    searchTools: (query: string, limit: number) => Promise<T> | T,
): Promise<CallToolResult> {
    const query = typeof args.query === 'string' ? args.query : '';
    const limit = typeof args.limit === 'number' ? args.limit : 10;
    const results = await searchTools(query, limit);
    return createTextResult(JSON.stringify(results, null, 2));
}

export async function executeLoadToolCompatibility(
    args: Record<string, unknown>,
    hasTool: (name: string) => boolean,
    workingSet: Pick<WorkingSetLike, 'loadTool'>,
): Promise<CallToolResult> {
    const toolName = typeof args.name === 'string' ? args.name : '';
    if (!hasTool(toolName)) {
        return createTextResult(`Tool '${toolName}' not found.`, true);
    }

    const evicted = workingSet.loadTool(toolName);
    const message = evicted.length > 0
        ? `Tool '${toolName}' loaded. Evicted idle tools: ${evicted.join(', ')}.`
        : `Tool '${toolName}' loaded.`;
    return createTextResult(message);
}

export async function executeGetToolSchemaCompatibility(
    args: Record<string, unknown>,
    getTool: (name: string) => Tool | null,
    workingSet: Pick<WorkingSetLike, 'hydrateTool'>,
    serializeSchema: (tool: Tool, evictedHydratedTools: string[]) => Record<string, unknown>,
    missingToolMessage?: (name: string) => string,
    onHydrate?: (toolName: string, tool: Tool) => void,
): Promise<CallToolResult> {
    const toolName = typeof args.name === 'string' ? args.name : '';
    const tool = getTool(toolName);
    if (!tool) {
        return createTextResult(
            missingToolMessage ? missingToolMessage(toolName) : `Tool '${toolName}' not found.`,
            true,
        );
    }

    const evictedHydratedTools = workingSet.hydrateTool(toolName);
    onHydrate?.(toolName, tool);
    return createTextResult(JSON.stringify(serializeSchema(tool, evictedHydratedTools), null, 2));
}

export async function executeUnloadToolCompatibility(
    args: Record<string, unknown>,
    workingSet: Pick<WorkingSetLike, 'unloadTool'>,
): Promise<CallToolResult> {
    const toolName = typeof args.name === 'string' ? args.name : '';
    const removed = workingSet.unloadTool(toolName);
    return createTextResult(
        removed
            ? `Tool '${toolName}' unloaded from the current session.`
            : `Tool '${toolName}' was not loaded in the current session.`,
        !removed,
    );
}

export async function executeListLoadedToolsCompatibility(
    workingSet: Pick<WorkingSetLike, 'getLimits' | 'listLoadedTools'>,
): Promise<CallToolResult> {
    return createTextResult(JSON.stringify({
        limits: workingSet.getLimits(),
        tools: workingSet.listLoadedTools(),
    }, null, 2));
}

export async function executeListAllToolsCompatibility<T>(
    allTools: T,
): Promise<CallToolResult> {
    return createTextResult(JSON.stringify(allTools, null, 2));
}
