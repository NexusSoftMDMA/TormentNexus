// DB removed
import { Tool } from "@modelcontextprotocol/sdk/types.js";
import { parseToolName } from "../tool-name-parser.service.js";
import {
    CallToolMiddleware,
    ListToolsMiddleware,
} from "./functional-middleware.js";

/**
 * Configuration for the tool overrides middleware
 */
export interface ToolOverridesConfig {
    cacheEnabled?: boolean;
    cacheTTL?: number; // milliseconds
    persistentCacheOnListTools?: boolean; // if true, cache never expires when set via list tools
}

/**
 * Tool override information
 */
interface ToolOverride {
    overrideName?: string | null;
    overrideTitle?: string | null;
    overrideDescription?: string | null;
    overrideAnnotations?: Record<string, unknown> | null;
}

function mergeAnnotations(
    original: Tool["annotations"],
    namespaceOverrides?: Record<string, unknown> | null,
): Tool["annotations"] | undefined {
    if (!namespaceOverrides || Object.keys(namespaceOverrides).length === 0) {
        return original;
    }

    const baseAnnotations = (original ? { ...original } : {}) as Record<
        string,
        unknown
    >;

    for (const [key, value] of Object.entries(namespaceOverrides)) {
        baseAnnotations[key] = value;
    }

    return baseAnnotations as Tool["annotations"];
}

// Override resolver
async function getToolOverrides(
    namespaceUuid: string,
    serverName: string,
    toolName: string,
    useCache: boolean = true,
    isPersistent: boolean = false,
): Promise<ToolOverride | null> {
    return null;
}

/**
 * Apply overrides to tools
 */
async function applyToolOverrides(
    tools: Tool[],
    namespaceUuid: string,
    useCache: boolean = true,
    isPersistent: boolean = false,
): Promise<Tool[]> {
    if (!tools || tools.length === 0) {
        return tools;
    }

    return tools;
}

/**
 * Map override name back to original name for tool calls
 */
export async function mapOverrideNameToOriginal(
    toolName: string,
    namespaceUuid: string,
    useCache: boolean = true,
): Promise<string> {
    return toolName;
}

/**
 * Creates a List Tools middleware that applies tool name/description overrides
 */
export function createToolOverridesListToolsMiddleware(
    config: ToolOverridesConfig = {},
): ListToolsMiddleware {
    const useCache = config.cacheEnabled ?? true;
    const isPersistent = config.persistentCacheOnListTools ?? false;

    return (handler) => {
        return async (request, context) => {
            // Call the original handler to get the tools
            const response = await handler(request, context);

            // Apply overrides to the tools
            if (response.tools) {
                const overriddenTools = await applyToolOverrides(
                    response.tools,
                    context.namespaceUuid,
                    useCache,
                    isPersistent,
                );

                return {
                    ...response,
                    tools: overriddenTools,
                };
            }

            return response;
        };
    };
}

/**
 * Creates a Call Tool middleware that maps override names back to original names
 */
export function createToolOverridesCallToolMiddleware(
    config: ToolOverridesConfig = {},
): CallToolMiddleware {
    const useCache = config.cacheEnabled ?? true;

    return (handler) => {
        return async (request, context) => {
            // Map override name back to original name if needed
            const originalToolName = await mapOverrideNameToOriginal(
                request.params.name,
                context.namespaceUuid,
                useCache,
            );

            // Create a new request with the original tool name
            const modifiedRequest = {
                ...request,
                params: {
                    ...request.params,
                    name: originalToolName,
                },
            };

            // Call the original handler with the modified request
            return handler(modifiedRequest, context);
        };
    };
}

/**
 * Utility function to clear override cache
 */
export function clearOverrideCache(namespaceUuid?: string): void {
    // No-op as cache is disabled
}
