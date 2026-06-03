import { parseNamespacedToolName } from './namespaces.js';
import type { MCPAggregatedTool } from './types.js';

interface DirectMCPHandlerDecision {
    legacyProxyEnabled: boolean;
    legacyProxyInitFailed?: boolean;
}

interface ProxyExecutionDecision {
    legacyProxyEnabled: boolean;
    prefersAggregatorExecution: boolean;
}

export function parseBooleanFlag(value: string | undefined): boolean {
    if (!value) {
        return false;
    }

    return ['1', 'true', 'yes', 'on'].includes(value.trim().toLowerCase());
}

export function isLegacyProxyDisabled(env: NodeJS.ProcessEnv = process.env): boolean {
    return parseBooleanFlag(env.MCP_DISABLE_TORMENTNEXUS);
}

export function shouldUseLegacyProxy(env: NodeJS.ProcessEnv = process.env): boolean {
    if (isLegacyProxyDisabled(env)) {
        return false;
    }

    return parseBooleanFlag(env.MCP_ENABLE_LEGACY_TORMENTNEXUS_PROXY ?? env.MCP_ENABLE_TORMENTNEXUS_PROXY);
}

export function isToolNotFoundError(error: unknown): boolean {
    if (!(error instanceof Error)) {
        return false;
    }

    return /unknown tool|not found in any connected mcp server|no provider found/i.test(error.message);
}

export function shouldPreferAggregatorExecution(
    toolName: string,
    aggregatedTools: ReadonlyArray<Pick<MCPAggregatedTool, 'name' | '_originalName'>> = [],
): boolean {
    if (parseNamespacedToolName(toolName) !== null) {
        return true;
    }

    return aggregatedTools.some((tool) => tool.name === toolName || tool._originalName === toolName);
}

export function shouldUseDirectMCPHandlers(decision: DirectMCPHandlerDecision): boolean {
    if (!decision.legacyProxyEnabled) {
        return true;
    }

    return decision.legacyProxyInitFailed === true;
}

export function shouldAttemptLegacyProxyExecution(decision: ProxyExecutionDecision): boolean {
    if (!decision.legacyProxyEnabled) {
        return false;
    }

    return !decision.prefersAggregatorExecution;
}