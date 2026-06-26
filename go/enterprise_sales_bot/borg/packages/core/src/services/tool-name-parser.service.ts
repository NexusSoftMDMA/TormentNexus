export interface ParsedToolName {
    serverName: string;
    originalToolName: string;
}

/**
 * Parse a TormentNexus tool name into server prefix and tool name components
 *
 * @param toolName - Full tool name (e.g., "Server__tool" or "Parent__Child__tool")
 * @returns Parsed components or null if invalid format
 */
export function parseToolName(toolName: string): ParsedToolName | null {
    const firstDoubleUnderscoreIndex = toolName.indexOf("__");
    if (firstDoubleUnderscoreIndex === -1) {
        return null;
    }

    // The first __ is always the separator between the top-level server prefix and the forwarded tool name
    // Everything before the first __ is the server prefix (which may contain nested servers)
    // Everything after the first __ is the forwarded tool name (which may itself include additional prefixes)
    const serverName = toolName.substring(0, firstDoubleUnderscoreIndex);
    const originalToolName = toolName.substring(firstDoubleUnderscoreIndex + 2);

    return {
        serverName,
        originalToolName,
    };
}

/**
 * Create a tool name from server prefix and tool name
 *
 * @param serverName - Server prefix (can be nested like "Parent__Child")
 * @param toolName - Tool name
 * @returns Full tool name
 */
export function createToolName(serverName: string, toolName: string): string {
    return `${serverName}__${toolName}`;
}
