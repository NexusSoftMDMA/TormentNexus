const MCP_SERVER_DEBUG_ENV = 'TORMENTNEXUS_MCP_SERVER_DEBUG';
const MCP_SERVER_DEBUG_NAMESPACE = 'tormentnexus:mcp-server';

export function isMcpServerDebugEnabled(env: NodeJS.ProcessEnv = process.env): boolean {
    const rawFlag = env[MCP_SERVER_DEBUG_ENV]?.trim().toLowerCase();
    if (rawFlag && ['1', 'true', 'yes', 'on'].includes(rawFlag)) {
        return true;
    }

    const debugNamespaces = env.DEBUG?.split(',').map((value) => value.trim().toLowerCase()).filter(Boolean) ?? [];
    return debugNamespaces.includes(MCP_SERVER_DEBUG_NAMESPACE);
}

export function mcpServerDebugLog(message: string, env: NodeJS.ProcessEnv = process.env): void {
    if (!isMcpServerDebugEnabled(env)) {
        return;
    }

    console.log(message);
}