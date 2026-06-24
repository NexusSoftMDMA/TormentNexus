function redirectProtocolUnsafeConsoleMethods(): void {
    const stderr = console.error.bind(console);
    console.log = stderr;
    console.info = stderr;
    console.debug = stderr;
    console.trace = stderr;
    console.dir = ((...args: unknown[]) => stderr(...args)) as typeof console.dir;
}

async function main() {
    // MCP stdio requires stdout to remain pristine JSON-RPC output only.
    redirectProtocolUnsafeConsoleMethods();

    process.on('unhandledRejection', (reason) => {
        console.error('[TormentNexus Core] Unhandled promise rejection:', reason);
    });
    process.on('uncaughtException', (error) => {
        console.error('[TormentNexus Core] Uncaught exception:', error);
        process.exit(1);
    });

    try {
        // PHASE 1: Immediately connect stdio transport with a lightweight MCP server.
        // This ensures the client (e.g., Gemini CLI) receives its initialize response
        // within ~500ms instead of waiting 30-60s for the full MCPServer to load.
        console.error('[TormentNexus Core] Phase 1: Connecting stdio transport immediately...');

        const { Server } = await import('@modelcontextprotocol/sdk/server/index.js');
        const { StdioServerTransport } = await import('@modelcontextprotocol/sdk/server/stdio.js');
        const { z } = await import('zod');

        const LooseListToolsRequestSchema = z.object({
            method: z.literal("tools/list"),
            params: z.any().optional(),
        });
        const LooseCallToolRequestSchema = z.object({
            method: z.literal("tools/call"),
            params: z.any(),
        });
        const LooseListPromptsRequestSchema = z.object({
            method: z.literal("prompts/list"),
            params: z.any().optional(),
        });
        const LooseGetPromptRequestSchema = z.object({
            method: z.literal("prompts/get"),
            params: z.any(),
        });
        const LooseListResourcesRequestSchema = z.object({
            method: z.literal("resources/list"),
            params: z.any().optional(),
        });
        const LooseReadResourceRequestSchema = z.object({
            method: z.literal("resources/read"),
            params: z.any(),
        });
        const LooseListResourceTemplatesRequestSchema = z.object({
            method: z.literal("resources/templates/list"),
            params: z.any().optional(),
        });

        let mcpServer: any = null;
        let serverReady = false;

        const lightweightServer = new Server(
            { name: "tormentnexus-core", version: "0.99.1" },
            { capabilities: { tools: {}, prompts: {}, resources: {} } }
        );

        // Tools/list handler: waits for full server to be ready (up to 30s)
        // so the client receives the complete tool set on first discovery.
        // This prevents "No prompts or tools found" errors from MCP clients.
        lightweightServer.setRequestHandler(LooseListToolsRequestSchema, async () => {
            // Wait for the full MCPServer to finish loading
            const maxWaitMs = 30_000;
            const pollIntervalMs = 200;
            const deadline = Date.now() + maxWaitMs;
            while (!serverReady) {
                if (Date.now() >= deadline) {
                    console.error('[TormentNexus Core] Timed out waiting for MCPServer to initialize');
                    break;
                }
                await new Promise(r => setTimeout(r, pollIntervalMs));
            }
            if (serverReady && mcpServer) {
                try {
                    const rawTools = await mcpServer.getDirectModeTools();
                    const tools = rawTools.map((t: any) => ({
                        name: t.name,
                        description: t.description,
                        inputSchema: t.inputSchema,
                    }));
                    return { tools };
                } catch (e: any) {
                    console.error('[TormentNexus Core] Error in delegated tools/list:', e.message);
                }
            }
            // Fallback: return a loading indicator if server init timed out
            return {
                tools: [{
                    name: "tormentnexus_loading",
                    description: "TormentNexus is still initializing. Please retry in a moment.",
                    inputSchema: { type: "object", properties: {} }
                }]
            };
        });

        // Tools/call handler: delegates to the full MCPServer once ready,
        // or waits briefly for initialization to complete.
        lightweightServer.setRequestHandler(LooseCallToolRequestSchema, async (request: any) => {
            // Wait up to 30s for the server to be ready
            const maxWaitMs = 30_000;
            const pollIntervalMs = 200;
            const deadline = Date.now() + maxWaitMs;
            while (!serverReady) {
                if (Date.now() >= deadline) break;
                await new Promise(r => setTimeout(r, pollIntervalMs));
            }
            if (serverReady && mcpServer) {
                try {
                    return await mcpServer.executeTool(request.params.name, request.params.arguments ?? {});
                } catch (e: any) {
                    console.error('[TormentNexus Core] Error in delegated tools/call:', e.message);
                    return {
                        content: [{ type: "text", text: `Error executing tool: ${e.message}` }],
                        isError: true,
                    };
                }
            }
            return {
                content: [{ type: "text", text: `TormentNexus is still initializing. Please retry in a few seconds.` }],
                isError: true,
            };
        });

        // Placeholder prompt/resource handlers
        lightweightServer.setRequestHandler(LooseListPromptsRequestSchema, async () => ({ prompts: [] }));
        lightweightServer.setRequestHandler(LooseGetPromptRequestSchema, async () => ({ messages: [] }));
        lightweightServer.setRequestHandler(LooseListResourcesRequestSchema, async () => ({ resources: [] }));
        lightweightServer.setRequestHandler(LooseReadResourceRequestSchema, async () => ({ contents: [] }));
        lightweightServer.setRequestHandler(LooseListResourceTemplatesRequestSchema, async () => ({ resourceTemplates: [] }));

        const stdioTransport = new StdioServerTransport();
        await lightweightServer.connect(stdioTransport);
        console.error('[TormentNexus Core] Phase 1 complete: Stdio transport connected.');

        // PHASE 2: Load the heavy MCPServer module in the background.
        console.error('[TormentNexus Core] Phase 2: Loading full MCPServer (this takes 10-30s)...');
        const startTime = Date.now();

        const { ensureBackgroundCoreRunning } = await import('./backgroundCoreBootstrap.js');
        void ensureBackgroundCoreRunning({
            waitForReady: false,
            log: (message, ...optionalParams) => console.error(message, ...optionalParams),
        }).then((result) => {
            if (result.status === 'spawned') {
                console.error(`[TormentNexus Core] Background control-plane bootstrap requested (PID: ${result.pid ?? 'unknown'}).`);
            }
        }).catch((error) => {
            console.error('[TormentNexus Core] Background control-plane bootstrap failed:', error);
        });

        const { MCPServer } = await import('./MCPServer.js');
        const elapsed = Date.now() - startTime;
        console.error(`[TormentNexus Core] MCPServer module loaded in ${elapsed}ms.`);

        // Create MCPServer with skipStdio since we already connected the transport
        mcpServer = new MCPServer({ skipWebsocket: true, skipStdio: true });
        await mcpServer.start();

        serverReady = true;
        console.error(`[TormentNexus Core] Phase 2 complete: Full server ready in ${Date.now() - startTime}ms.`);
        console.error("[TormentNexus Core] MCP Server Stdio Entry Point Started.");

        // Notify client to reload tools list now that the full server is initialized
        try {
            await lightweightServer.notification({
                method: "notifications/tools/list_changed"
            });
            console.error("[TormentNexus Core] Sent notifications/tools/list_changed to client successfully.");
        } catch (notificationErr: any) {
            console.error("[TormentNexus Core] Failed to send notifications/tools/list_changed:", notificationErr.message);
        }

    } catch (err) {
        console.error("Failed to start MCP server:", err);
        process.exit(1);
    }
}

main();
