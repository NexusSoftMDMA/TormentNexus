import type { CallToolResult, Tool } from '@modelcontextprotocol/sdk/types.js';

import { SessionToolWorkingSet } from './SessionToolWorkingSet.js';
import { getToolLoadingDefinitions } from './toolLoadingDefinitions.js';
import {
    executeGetToolSchemaCompatibility,
    executeListLoadedToolsCompatibility,
    executeLoadToolCompatibility,
    executeSearchToolsCompatibility,
    executeUnloadToolCompatibility,
} from './toolLoadingCompatibility.js';
import {
    pickAutoLoadCandidate,
    rankToolSearchCandidates,
    type RankedToolSearchResult,
} from './toolSearchRanking.js';
import type { ToolContextPayload } from '../services/toolContextMemory.js';

type SearchableTool = Tool & {
    server?: string;
    serverDisplayName?: string;
    originalName?: string;
    advertisedName?: string;
    serverTags?: string[];
    toolTags?: string[];
    semanticGroup?: string;
    semanticGroupLabel?: string;
    keywords?: string[];
    alwaysOn?: boolean;
};

function createTextResult(text: string, isError = false): CallToolResult {
    return {
        content: [{ type: 'text', text }],
        isError,
    };
}

export class NativeSessionMetaTools {
    private readonly workingSet: SessionToolWorkingSet;
    private readonly catalog = new Map<string, Tool>();
    private toolContextResolver?: (input: { toolName: string; args?: Record<string, unknown> }) => ToolContextPayload | null;

    constructor(
        workingSet: SessionToolWorkingSet = new SessionToolWorkingSet(),
        options: {
            toolContextResolver?: (input: { toolName: string; args?: Record<string, unknown> }) => ToolContextPayload | null;
        } = {},
    ) {
        this.workingSet = workingSet;
        this.toolContextResolver = options.toolContextResolver;
    }

    public setToolContextResolver(
        resolver?: (input: { toolName: string; args?: Record<string, unknown> }) => ToolContextPayload | null,
    ): void {
        this.toolContextResolver = resolver;
    }

    public refreshCatalog(tools: Tool[]): void {
        this.catalog.clear();
        // Always include base loading/meta tools in catalog
        this.listToolDefinitions().forEach((tool) => {
            this.catalog.set(tool.name, tool);
        });
        tools.forEach((tool) => {
            this.catalog.set(tool.name, tool);
        });
    }

    public listToolDefinitions(): Tool[] {
        return getToolLoadingDefinitions();
    }

    public getVisibleLoadedTools(): Tool[] {
        return this.workingSet.listLoadedTools()
            .map((entry) => {
                const tool = this.catalog.get(entry.name);
                if (!tool) {
                    return null;
                }

                if (entry.hydrated) {
                    return tool;
                }

                return this.toMinimalTool(tool);
            })
            .filter((tool): tool is Tool => tool !== null);
    }

    public getLoadedToolNames(): string[] {
        return this.workingSet.getLoadedToolNames();
    }

    public setAlwaysLoadedTools(names: string[]): void {
        this.workingSet.setAlwaysLoadedTools(names.filter((name) => this.catalog.has(name)));
    }

    public hasTool(name: string): boolean {
        return this.catalog.has(name);
    }

    public loadToolIntoSession(name: string): { loaded: boolean; evicted: string[] } {
        if (!this.catalog.has(name)) {
            return { loaded: false, evicted: [] };
        }

        return {
            loaded: true,
            evicted: this.workingSet.loadTool(name),
        };
    }

    public touchLoadedTool(name: string): boolean {
        return this.workingSet.touchTool(name);
    }

    public async handleToolCall(name: string, args: Record<string, unknown>): Promise<CallToolResult | null> {
        if (name === 'search_tools') {
            return await executeSearchToolsCompatibility(args, (query, limit) => this.searchTools(query, limit));
        }

        if (name === 'load_tool') {
            return await executeLoadToolCompatibility(args, (toolName) => this.catalog.has(toolName), this.workingSet);
        }

        if (name === 'get_tool_schema') {
            return await executeGetToolSchemaCompatibility(
                args,
                (toolName) => this.catalog.get(toolName) ?? null,
                this.workingSet,
                (tool, evictedHydratedTools) => ({
                    name: tool.name,
                    description: tool.description ?? '',
                    inputSchema: tool.inputSchema ?? { type: 'object', properties: {} },
                    evictedHydratedTools,
                }),
            );
        }

        if (name === 'get_tool_context') {
            const toolName = typeof args.name === 'string' ? args.name : '';
            if (!toolName) {
                return createTextResult('Tool context lookup requires a downstream tool name.', true);
            }

            if (!this.toolContextResolver) {
                return createTextResult('Tool context resolver is not available in this TormentNexus session.', true);
            }

            const payload = this.toolContextResolver({
                toolName,
                args: typeof args.arguments === 'object' && args.arguments !== null
                    ? args.arguments as Record<string, unknown>
                    : undefined,
            });

            return createTextResult(JSON.stringify(payload ?? {
                toolName,
                query: toolName,
                matchedPaths: [],
                observationCount: 0,
                summaryCount: 0,
                prompt: `JIT tool context for ${toolName}:\nNo relevant prior memory was found.`,
            }));
        }

        if (name === 'unload_tool') {
            return await executeUnloadToolCompatibility(args, this.workingSet);
        }

        if (name === 'list_loaded_tools') {
            return await executeListLoadedToolsCompatibility(this.workingSet);
        }
        if (name === 'list_all_tools') {
            const baseToolNames = new Set(this.listToolDefinitions().map((t) => t.name));
            const toolsList = Array.from(this.catalog.values())
                .filter((tool) => !baseToolNames.has(tool.name))
                .map((tool) => {
                    const searchableTool = tool as SearchableTool;
                    const loaded = this.workingSet.isLoaded(tool.name);
                    return {
                        name: tool.name,
                        description: tool.description ?? '',
                        alwaysOn: Boolean(searchableTool.alwaysOn),
                        loaded,
                        inputSchema: tool.inputSchema ?? { type: 'object', properties: {} },
                    };
                });

            const summary = {
                total: toolsList.length,
                loaded: toolsList.filter((t) => t.loaded).length,
                alwaysOn: toolsList.filter((t) => t.alwaysOn).length,
            };

            return createTextResult(JSON.stringify({
                summary,
                tools: toolsList,
            }));
        }
        if (name === 'set_capacity') {
            const maxLoadedTools = typeof args.maxLoadedTools === 'number' ? args.maxLoadedTools : undefined;
            const maxHydratedSchemas = typeof args.maxHydratedSchemas === 'number' ? args.maxHydratedSchemas : undefined;
            const idleEvictionThresholdMs = typeof args.idleEvictionThresholdMs === 'number' ? args.idleEvictionThresholdMs : undefined;
            this.workingSet.reconfigure({
                maxLoadedTools: maxLoadedTools ?? this.workingSet.getLimits().maxLoadedTools,
                maxHydratedSchemas: maxHydratedSchemas ?? this.workingSet.getLimits().maxHydratedSchemas,
                idleEvictionThresholdMs: idleEvictionThresholdMs ?? this.workingSet.getLimits().idleEvictionThresholdMs,
            });
            const limits = this.workingSet.getLimits();
            return createTextResult(`Capacity updated: maxLoadedTools=${limits.maxLoadedTools}, maxHydratedSchemas=${limits.maxHydratedSchemas}, idleEvictionThresholdMs=${limits.idleEvictionThresholdMs}`);
        }
        if (name === 'get_eviction_history') {
            const history = this.workingSet.getEvictionHistory();
            return createTextResult(JSON.stringify(history));
        }
        if (name === 'clear_eviction_history') {
            this.workingSet.clearEvictionHistory();
            return createTextResult('Cleared eviction history.');
        }

        return null;
    }

    private searchTools(query: string, limit: number): RankedToolSearchResult[] {
        const rankedResults = rankToolSearchCandidates(
            Array.from(this.catalog.values()).map((tool) => {
                const searchableTool = tool as SearchableTool;
                return {
                name: tool.name,
                description: tool.description ?? '',
                serverName: searchableTool.server,
                serverDisplayName: searchableTool.serverDisplayName,
                originalName: searchableTool.originalName,
                advertisedName: searchableTool.advertisedName,
                serverTags: searchableTool.serverTags,
                toolTags: searchableTool.toolTags,
                semanticGroup: searchableTool.semanticGroup,
                semanticGroupLabel: searchableTool.semanticGroupLabel,
                keywords: searchableTool.keywords,
                alwaysOn: searchableTool.alwaysOn,
                loaded: this.workingSet.isLoaded(tool.name),
                hydrated: this.workingSet.isHydrated(tool.name),
                deferred: !this.workingSet.isHydrated(tool.name),
                };
            }),
            query,
            limit,
        );

        const autoLoadDecision = pickAutoLoadCandidate(rankedResults, query);
        if (!autoLoadDecision) {
            return rankedResults;
        }

        const { loaded, evicted } = this.loadToolIntoSession(autoLoadDecision.toolName);
        if (!loaded) {
            return rankedResults;
        }

        return rankedResults.map((result) => {
            if (result.name === autoLoadDecision.toolName) {
                return {
                    ...result,
                    loaded: true,
                    autoLoaded: true,
                    matchReason: `${result.matchReason}; ${autoLoadDecision.reason}`,
                };
            }

            if (evicted.includes(result.name)) {
                return {
                    ...result,
                    loaded: false,
                };
            }

            return result;
        });
    }

    /**
     * Inject tools predicted by the ConversationalToolInjector into the working
     * set. Only tools present in the catalog are loaded; unknown names are
     * silently skipped. Always-on tools are not affected.
     *
     * Returns the names of tools that were actually loaded (newly added to the
     * working set, not previously loaded).
     */
    public injectConversationalTools(names: string[]): string[] {
        const loaded: string[] = [];
        for (const name of names) {
            if (!this.catalog.has(name)) {
                continue;
            }
            if (this.workingSet.isLoaded(name)) {
                continue; // already visible
            }
            const { loaded: wasLoaded } = this.loadToolIntoSession(name);
            if (wasLoaded) {
                loaded.push(name);
            }
        }
        return loaded;
    }

    /**
     * Returns a compact snapshot of every tool in the catalog, suitable for
     * passing to the ConversationalToolInjector for LLM-based prediction.
     */
    public getCatalogSnapshot(): Array<{
        name: string;
        description: string;
        alwaysOn: boolean;
        loaded: boolean;
        serverTags: string[];
        toolTags: string[];
        semanticGroup: string;
        keywords: string[];
    }> {
        return Array.from(this.catalog.values()).map((tool) => {
            const st = tool as SearchableTool;
            return {
                name: tool.name,
                description: tool.description ?? '',
                alwaysOn: Boolean(st.alwaysOn),
                loaded: this.workingSet.isLoaded(tool.name),
                serverTags: st.serverTags ?? [],
                toolTags: st.toolTags ?? [],
                semanticGroup: st.semanticGroup ?? '',
                keywords: st.keywords ?? [],
            };
        });
    }

    private toMinimalTool(tool: Tool): Tool {
        return {
            name: tool.name,
            description: `[Deferred] ${tool.description ?? 'No description'}`,
            inputSchema: {
                type: 'object',
                properties: {},
            },
        };
    }
}