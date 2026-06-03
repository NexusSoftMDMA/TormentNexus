import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import type { MCPServerConfig } from "../config/TormentNexusConfig.js";
import { tormentnexusLogStore } from "../services/log-store.service.js";
import { ProcessManagedStdioTransport } from "../transports/process-managed.transport.js";

import { db } from "../db/index.js";
import { workspaceSecretsTable } from "../db/tormentnexus-schema.js";

export class StdioClient {
    private client: Client | null = null;
    private transport: ProcessManagedStdioTransport | null = null;
    public readonly name: string;
    private config: MCPServerConfig;

    constructor(name: string, config: MCPServerConfig) {
        this.name = name;
        this.config = config;
    }

    private isExpectedConnectionClose(error: unknown): boolean {
        if (!(error instanceof Error)) {
            return false;
        }

        const maybeCode = (error as NodeJS.ErrnoException).code;
        if (maybeCode === "ABORT_ERR" || maybeCode === "ECONNRESET") {
            return true;
        }

        const message = error.message.toLowerCase();
        return message.includes("connection closed") || message.includes("transport closed");
    }

    public async connect(): Promise<void> {
        if (!this.config.enabled) {
            console.log(`[StdioClient:${this.name}] Disabled in config.`);
            return;
        }

        console.log(`[StdioClient:${this.name}] Connecting to ${this.config.command} ${this.config.args.join(' ')}...`);

        const safeEnv: Record<string, string> = {};
        for (const [key, value] of Object.entries(process.env)) {
            if (value !== undefined) safeEnv[key] = value;
        }

        // Inject workspace secrets from the SQLite database
        try {
            const secrets = await db.select().from(workspaceSecretsTable);
            for (const secret of secrets) {
                if (secret.value !== undefined) {
                    safeEnv[secret.key] = secret.value;
                }
            }
        } catch (e: any) {
            console.warn(`[StdioClient:${this.name}] Failed to fetch workspace secrets:`, e.message);
        }

        // Config overrides secrets (to allow specific servers to use different keys if needed)
        Object.assign(safeEnv, this.config.env);

        this.transport = new ProcessManagedStdioTransport({
            command: this.config.command,
            args: this.config.args,
            env: safeEnv,
            stderr: 'pipe',
        });

        this.transport.stderr?.on('data', (chunk: Buffer) => {
            const message = chunk.toString().trim();
            if (!message) {
                return;
            }

            tormentnexusLogStore.addLog(this.name, 'error', message);
        });

        this.transport.stderr?.on('error', (error: Error) => {
            tormentnexusLogStore.addLog(this.name, 'error', 'stderr error', error);
        });

        this.transport.stdout?.on('data', (chunk: Buffer) => {
            const message = chunk.toString().trim();
            if (!message) {
                return;
            }

            tormentnexusLogStore.addLog(this.name, 'info', message);
        });

        this.transport.stdout?.on('error', (error: Error) => {
            tormentnexusLogStore.addLog(this.name, 'error', 'stdout error', error);
        });

        this.client = new Client({
            name: `tormentnexus-client-${this.name}`,
            version: "1.0.0",
        }, {
            capabilities: {
                // Client capabilities
            }
        });

        try {
            await this.client.connect(this.transport);
            console.log(`[StdioClient:${this.name}] Connected!`);
        } catch (error) {
            if (!this.isExpectedConnectionClose(error)) {
                console.error(`[StdioClient:${this.name}] Connection failed:`, error);
            }
            throw error;
        }
    }

    public async listTools(options?: { throwOnError?: boolean }): Promise<any[]> {
        if (!this.client) return [];
        try {
            const result = await this.client.listTools();
            return result.tools;
        } catch (error) {
            if (!this.isExpectedConnectionClose(error)) {
                console.error(`[StdioClient:${this.name}] Failed to list tools:`, error);
            }
            if (options?.throwOnError) {
                throw error;
            }
            return [];
        }
    }

    public async callTool(toolName: string, args: any): Promise<any> {
        if (!this.client) throw new Error(`[StdioClient:${this.name}] Not connected.`);
        return await this.client.callTool({
            name: toolName,
            arguments: args
        });
    }

    public async close(): Promise<void> {
        if (this.client) {
            await this.client.close().catch(() => undefined);
        }
        if (this.transport) {
            await this.transport.close().catch(() => undefined);
        }
    }
}
