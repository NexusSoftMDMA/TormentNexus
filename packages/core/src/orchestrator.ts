
console.log("[Core:Orchestrator] Starting imports...");
import { spawn } from 'child_process';

import express from 'express';
console.log("[Core:Orchestrator] ✓ express");
import cors from 'cors';
console.log("[Core:Orchestrator] ✓ cors");
import { createExpressMiddleware } from '@trpc/server/adapters/express';
import { appRouter } from './trpc.js';
console.log("[Core:Orchestrator] ✓ trpc.js");
import { ingestPublishedCatalog } from './services/published-catalog-ingestor.js';
import { InputTools, SystemStatusTool } from '@tormentnexus/tools';
import { MCPServer } from './MCPServer.js';
import { listenExpress } from './orchestrator-listen.js';
import { resolveSupervisorEntryPath } from './orchestratorPaths.js';
import { resolveBridgePort } from './bridge/bridgePort.js';
import { councilApp } from './orchestrator/council/node-index.js';
import { jsonConfigProvider } from './services/config/JsonConfigProvider.js';
import { codeExecutorService } from './services/CodeExecutorService.js';

export const name = "@tormentnexus/core";

export interface StartOrchestratorOptions {
    host?: string;
    trpcPort?: number;
    startSupervisor?: boolean;
    startMcp?: boolean;
    autoDrive?: boolean;
}

export async function startOrchestrator(options: StartOrchestratorOptions = {}) {
    console.log(`[Core] Initializing ${name}...`);

    const host = options.host ?? '0.0.0.0';
    const trpcPort = options.trpcPort ?? 4100;
    const startSupervisor = options.startSupervisor ?? false;
    const startMcp = options.startMcp ?? true;
    const autoDrive = options.autoDrive ?? false;

    console.log("[Core] 1. Starting Express/tRPC/REST...");
    // 1. Start tRPC Server (Dashboard API)
    const app = express();
    
    // Standard CORS for Express (handles tRPC and static routes)
    app.use(cors({
        origin: (origin, callback) => {
            if (!origin) return callback(null, true);
            if (origin.includes('localhost') || origin.includes('127.0.0.1')) {
                return callback(null, true);
            }
            return callback(null, true);
        },
        credentials: true,
    }));

    // REST API: /api/scripts — must be registered BEFORE the /api/* wildcard
    // so these specific routes take priority over the Hono bridge.
    app.get('/api/scripts', async (_req, res) => {
        try {
            const scripts = await jsonConfigProvider.loadScripts();
            res.json({ success: true, data: scripts });
        } catch (err: any) {
            res.json({ success: true, data: [], meta: { fallback: true, reason: err?.message } });
        }
    });
    app.get('/api/scripts/get', async (req, res) => {
        try {
            const uuid = String(req.query.uuid ?? '');
            const scripts = await jsonConfigProvider.loadScripts();
            const found = scripts.find((s: any) => s.uuid === uuid);
            res.json({ success: true, data: found ?? null });
        } catch (err: any) {
            res.json({ success: false, error: err?.message });
        }
    });
    app.post('/api/scripts/create', express.json(), async (req, res) => {
        try {
            const script = await jsonConfigProvider.saveScript(req.body);
            res.json({ success: true, data: script });
        } catch (err: any) {
            res.json({ success: false, error: err?.message });
        }
    });
    app.post('/api/scripts/update', express.json(), async (req, res) => {
        try {
            const scripts = await jsonConfigProvider.loadScripts();
            const existing = scripts.find((s: any) => s.uuid === req.body.uuid);
            if (!existing) { res.json({ success: false, error: 'Script not found' }); return; }
            const updated = { ...existing, ...req.body };
            await jsonConfigProvider.saveScript(updated);
            res.json({ success: true, data: updated });
        } catch (err: any) {
            res.json({ success: false, error: err?.message });
        }
    });
    app.post('/api/scripts/delete', express.json(), async (req, res) => {
        try {
            await jsonConfigProvider.deleteScript(req.body.uuid);
            res.json({ success: true });
        } catch (err: any) {
            res.json({ success: false, error: err?.message });
        }
    });
    app.post('/api/scripts/execute', express.json(), async (req, res) => {
        try {
            const scripts = await jsonConfigProvider.loadScripts();
            const found = scripts.find((s: any) => s.uuid === req.body.uuid);
            if (!found) { res.json({ success: false, error: 'Script not found' }); return; }
            const result = await codeExecutorService.executeCode(found.code);
            res.json({ success: true, data: result });
        } catch (err: any) {
            res.json({ success: false, error: err?.message });
        }
    });

    // Bridge REST API from Council (Hono) into Express
    // This unifies both systems on the same port (3847)
    app.all('/api/*', async (req, res) => {
        try {
            // Convert Express Request to Web Standard Request for Hono
            const protocol = req.protocol;
            const host = req.get('host');
            const url = `${protocol}://${host}${req.originalUrl}`;
            
            const webReq = new Request(url, {
                method: req.method,
                headers: new Headers(req.headers as Record<string, string>),
                body: ['POST', 'PUT', 'PATCH'].includes(req.method) ? JSON.stringify(req.body) : undefined,
            });

            // Call Hono's fetch handler
            const honoRes = await (councilApp as any).fetch(webReq);
            
            // Convert Web Standard Response back to Express Response
            res.status(honoRes.status);
            honoRes.headers.forEach((value: string, key: string) => {
                res.setHeader(key, value);
            });
            
            const body = await honoRes.text();
            res.send(body);
        } catch (error) {
            console.error('[Core:Bridge] Hono bridging error:', error);
            res.status(500).json({ success: false, error: 'Internal server error in API bridge' });
        }
    });

    // Health endpoint — must precede TRPC so probes don't fall through to
    // middleware that calls getMcpServer() (which throws before init).
    app.get('/health', (_req, res) => {
        res.json({
            status: 'ok',
            name: '@tormentnexus/core',
            uptime: process.uptime(),
            timestamp: Date.now(),
            mcpReady: !!global.mcpServerInstance,
        });
    });

    // (scripts routes moved above /api/* wildcard)

    // tRPC middleware
    app.use('/trpc', createExpressMiddleware({
        router: appRouter,
        createContext: () => ({}),
    }));

    await listenExpress(app, trpcPort, host);
    console.log(`[Core] Unified API Server running at http://${host}:${trpcPort}/trpc and /api`);

    // 1.1. Schedule automatic catalog ingestion (startup + 24h interval)
    scheduleCatalogIngestion();

    // 1.5. Start Supervisor (Native Input / Watchdog)
    if (startSupervisor) {
        try {
            console.log("[Core] 1.5 Starting TormentNexus Supervisor...");
            const supervisorPath = resolveSupervisorEntryPath();
            if (!supervisorPath) {
                console.warn("[Core] TormentNexus Supervisor build not found. Skipping supervisor startup.");
            } else {
                const supervisor = spawn('node', [supervisorPath], {
                    stdio: 'inherit',
                    detached: false
                });

                supervisor.on('error', (err) => console.error("[Supervisor] Failed to start:", err));
                console.log(`[Core] Supervisor running (PID: ${supervisor.pid})`);
            }
        } catch (e) {
            console.error("[Core] Failed to spawn Supervisor:", e);
        }
    }

    // 2. Start MCP Server (Bridged: Stdio + WebSocket)
    if (startMcp) {
        try {
            console.log("[Core] 2. Instantiating MCPServer...");
            const inputTools = new InputTools();
            const systemStatusTool = new SystemStatusTool();
            const mcp = new MCPServer({
                inputTools,
                systemStatusTool,
                skipAutoDrive: !autoDrive,
                skipStdio: true,
            });

            console.log("[Core] 3. Starting MCPServer...");
            await mcp.start();
            console.log("[Core] MCPServer Started.");

            // Auto-Start the Director in Auto-Drive Mode (High Autonomy)
            if (autoDrive) {
                console.log("[Core] 4. Scheduling Auto-Drive...");
                // We use a small delay to ensure connections are ready
                setTimeout(() => {
                    console.log("[Core] Triggering Auto-Drive Tool...");
                    mcp.executeTool('start_auto_drive', {}).catch(e => console.error("Failed to auto-start Director:", e));
                }, 3000);
            }

        } catch (err) {
            console.error("Failed to start MCP server:", err);
            throw err;
        }
    } else {
        // Lightweight MCP init — just create the instance for tRPC routers to use
        try {
            console.log("[Core] 2. Instantiating lightweight MCPServer (no stdio/bridge)...");
            const inputTools = new InputTools();
            const systemStatusTool = new SystemStatusTool();
            const mcp = new MCPServer({
                inputTools,
                systemStatusTool,
                skipAutoDrive: true,
                skipStdio: true,
            });
            global.mcpServerInstance = mcp;

            // Load server configs from mcp.jsonc into the aggregator
            try {
                const { loadTormentNexusMcpConfig } = await import('./mcp/mcpJsonConfig.js');
                const config = await loadTormentNexusMcpConfig();
                const serverConfigs = config.mcpServers ?? {};
                let loaded = 0;
                for (const [name, serverConfig] of Object.entries(serverConfigs)) {
                    const sc = serverConfig as any;
                    if (sc.command) {
                        mcp.mcpAggregator.addServerConfig(name, {
                            command: sc.command,
                            args: sc.args ?? [],
                            env: sc.env ?? {},
                            enabled: sc.enabled !== false,
                        }).catch(() => {}); // Don't block on connection failures
                        loaded++;
                    }
                }
                console.log(`[Core] Loaded ${loaded} server configs from mcp.jsonc into aggregator`);
            } catch (e) {
                console.log('[Core] Could not load mcp.jsonc configs:', (e as Error).message);
            }

            console.log("[Core] MCPServer initialized (lightweight — no tool discovery).");
        } catch (err) {
            console.error("[Core] Failed to initialize lightweight MCP:", err);
        }
    }

    return {
        host,
        trpcPort,
        bridgePort: startMcp ? resolveBridgePort() : null,
    };
}

// ---------------------------------------------------------------------------
// Catalog Ingestion Scheduler
// ---------------------------------------------------------------------------
const INGEST_INTERVAL_MS = 24 * 60 * 60 * 1000; // 24 hours
const INGEST_STARTUP_DELAY_MS = 10_000;           // 10 seconds after startup

function scheduleCatalogIngestion(): void {
    const run = () => {
        console.log("[Core:Catalog] Running scheduled catalog ingestion...");
        ingestPublishedCatalog()
            .then(report => {
                const adapterNames = report.results.map(r => r.source).join(", ");
                console.log(
                    `[Core:Catalog] Ingestion complete — upserted: ${report.total_upserted}, ` +
                    `errors: ${report.total_errors}, adapters: ${adapterNames}`
                );
            })
            .catch(err => {
                console.warn("[Core:Catalog] Scheduled ingestion failed (non-fatal):", err?.message ?? err);
            });
    };

    setTimeout(() => {
        run();
        setInterval(run, INGEST_INTERVAL_MS);
    }, INGEST_STARTUP_DELAY_MS);

    console.log(`[Core:Catalog] Catalog ingestion scheduled (startup delay: ${INGEST_STARTUP_DELAY_MS / 1000}s, interval: 24h)`);
}
