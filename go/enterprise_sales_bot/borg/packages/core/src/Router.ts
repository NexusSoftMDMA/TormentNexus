import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { ListToolsResult, CallToolResult, Tool } from "@modelcontextprotocol/sdk/types.js";
import { tormentnexusLogStore } from "./services/log-store.service.js";
import { ProcessManagedStdioTransport } from "./transports/process-managed.transport.js";

/**
 * RouterConfig
 * Configuration for the MCP Router.
 * Currently supports an optional default provider, extensible for future routing policies.
 */
export interface RouterConfig {
    defaultProvider?: string;
}

/**
 * Router (MCP Aggregator)
 *
 * The central hub that manages connections to multiple MCP servers and provides
 * a unified tool interface. This is TormentNexus's "Meta-MCP" layer.
 *
 * Architecture:
 * - Maintains a Map of named MCP `Client` connections (e.g., "filesystem", "browser", "search").
 * - Each client connects via stdio transport to a child process running an MCP server.
 * - `listTools()` aggregates tools from ALL connected servers into a single flat list.
 * - `callTool()` uses a cached tool→client mapping for O(1) lookup (Phase 63 optimization).
 *
 * Performance:
 * - The `toolCache` avoids the previous O(N) linear scan that called `listTools()` on
 *   every client for every `callTool` invocation. Now it's O(1) with lazy cache refresh.
 */
export class Router {
    /** Map of server name → MCP Client connection */
    private clients: Map<string, Client> = new Map();

    /** Cache mapping tool name → owning client name for O(1) lookup in callTool */
    private toolCache: Map<string, string> = new Map();

    constructor(private config: RouterConfig = {}) {
        console.log("Router initialized");
    }

    /**
     * Connects to a local MCP server via stdio transport.
     * Idempotent — returns existing client if already connected.
     *
     * @param name - Unique identifier for this server (e.g., "filesystem", "browser")
     * @param command - The executable to spawn (e.g., "node", "python")
     * @param args - Arguments to pass to the command (e.g., ["./dist/server.js"])
     * @returns The connected MCP Client instance
     */
    async connectToServer(name: string, command: string, args: string[]) {
        if (this.clients.has(name)) return this.clients.get(name);

        const transport = new ProcessManagedStdioTransport({
            command,
            args,
            stderr: 'pipe',
        });

        transport.stderr?.on('data', (chunk: Buffer) => {
            const message = chunk.toString().trim();
            if (!message) {
                return;
            }

            tormentnexusLogStore.addLog(name, 'error', message);
        });

        transport.stderr?.on('error', (error: Error) => {
            tormentnexusLogStore.addLog(name, 'error', 'stderr error', error);
        });

        transport.stdout?.on('data', (chunk: Buffer) => {
            const message = chunk.toString().trim();
            if (!message) {
                return;
            }

            tormentnexusLogStore.addLog(name, 'info', message);
        });

        transport.stdout?.on('error', (error: Error) => {
            tormentnexusLogStore.addLog(name, 'error', 'stdout error', error);
        });

        const client = new Client(
            {
                name: "tormentnexus-router",
                version: "0.99.1",
            },
            {
                capabilities: {},
            }
        );

        await client.connect(transport);
        this.clients.set(name, client);
        console.log(`Connected to MCP Server: ${name}`);
        return client;
    }

    /**
     * Aggregates tools from all connected MCP servers.
     * Also rebuilds the toolCache for O(1) lookups in callTool.
     *
     * @returns Flat array of all available tools across all connected servers
     */
    async listTools(): Promise<Tool[]> {
        const allTools: Tool[] = [];
        this.toolCache.clear(); // Invalidate cache on refresh

        for (const [name, client] of this.clients.entries()) {
            try {
                const result = await client.listTools();
                // Populate cache: tool.name → server name
                // Future: prefix with `${name}__` for namespacing if tool names collide
                for (const tool of result.tools) {
                    this.toolCache.set(tool.name, name);
                    allTools.push(tool);
                }
            } catch (e) {
                console.error(`Failed to list tools from ${name}`, e);
            }
        }
        return allTools;
    }

    /**
     * Calls a tool by name, routing to the correct MCP server.
     * Uses cached tool→client mapping for O(1) lookup.
     * On cache miss, refreshes the cache by calling listTools() once.
     *
     * @param name - The tool name to invoke (e.g., "read_file", "search_web")
     * @param args - Arguments to pass to the tool
     * @returns The tool execution result
     * @throws Error if the tool is not found in any connected server
     */
    async callTool(name: string, args: any): Promise<CallToolResult> {
        // 1. Try Cache — O(1) lookup
        let clientName = this.toolCache.get(name);

        // 2. Cache miss — lazy refresh (happens once, then cached)
        if (!clientName) {
            await this.listTools();
            clientName = this.toolCache.get(name);
        }

        // 3. Execute on the owning client
        if (clientName) {
            const client = this.clients.get(clientName);
            if (client) {
                return (await client.callTool({ name, arguments: args })) as unknown as CallToolResult;
            }
        }

        throw new Error(`Tool ${name} not found in any connected MCP server.`);
    }

    /**
     * Returns a specific MCP client by name.
     * Useful for direct server interaction outside of the tool abstraction.
     */
    getClient(name: string): Client | undefined {
        return this.clients.get(name);
    }
}
