import type { CallToolResult } from '@modelcontextprotocol/sdk/types.js';

export function getAllowedToolsMetadataGuardResult(
    toolName: string,
    meta?: Record<string, unknown>,
): CallToolResult | null {
    const allowedTools = meta?.allowedTools;

    if (Array.isArray(allowedTools) && !allowedTools.includes(toolName)) {
        return {
            content: [{
                type: 'text',
                text: `Access denied: Tool '${toolName}' is not in the allowed tools list for this agent.`,
            }],
            isError: true,
        };
    }

    return null;
}