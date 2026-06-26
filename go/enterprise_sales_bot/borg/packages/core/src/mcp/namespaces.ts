const TOOL_NAMESPACE_SEPARATOR = '__';

export function namespaceToolName(serverName: string, toolName: string): string {
    return `${serverName}${TOOL_NAMESPACE_SEPARATOR}${toolName}`;
}

export function parseNamespacedToolName(toolName: string): { serverName: string; toolName: string } | null {
    const separatorIndex = toolName.indexOf(TOOL_NAMESPACE_SEPARATOR);
    if (separatorIndex <= 0 || separatorIndex === toolName.length - TOOL_NAMESPACE_SEPARATOR.length) {
        return null;
    }

    return {
        serverName: toolName.slice(0, separatorIndex),
        toolName: toolName.slice(separatorIndex + TOOL_NAMESPACE_SEPARATOR.length),
    };
}
