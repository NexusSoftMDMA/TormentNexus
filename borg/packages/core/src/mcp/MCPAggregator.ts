import fs from "fs";
import path from "path";

import { deriveSemanticCatalogForServer } from "./catalogMetadata.js";
import { MCPConfigStore } from "./configStore.js";
import { namespaceToolName, parseNamespacedToolName } from "./namespaces.js";
import { StdioClient } from "./StdioClient.js";
import { MCPTrafficInspector, summarizeParams } from "./trafficInspector.js";
import type {
	MCPAggregatedTool,
	MCPAggregatorOptions,
	MCPClientLike,
	MCPServerConfig,
	MCPServerState,
	MCPToolDefinition,
} from "./types.js";

type AdvertisedInventorySeed = {
	servers: Array<{
		name: string;
		alwaysOnAdvertised?: boolean;
		advertisedToolCount?: number;
	}>;
	source?: "database" | "config" | "empty";
};

export class MCPAggregator {
	public readonly clients: Map<string, MCPClientLike> = new Map();
	public readonly configPath: string;

	private readonly configStore: MCPConfigStore;
	private readonly createClient: (
		name: string,
		config: MCPServerConfig,
	) => MCPClientLike;
	private readonly restartDelayMs: number;
	private readonly now: () => number;
	private readonly serverStates: Map<string, MCPServerState> = new Map();
	private readonly trafficInspector: MCPTrafficInspector;
	/** Connection timeout per server (ms). Prevents a single broken downstream
	 *  server from blocking the entire aggregator for 60+ seconds. */
	private readonly connectTimeoutMs: number;
	// When true, listAggregatedTools() skips unconnected servers to avoid eager
	// binary spawning. Set via setLazyMode() once the lifecycle mode is known.
	private lazyMode: boolean = false;
	private initializationState: {
		inProgress: boolean;
		initialized: boolean;
		isLKG: boolean;
		lastStartedAt?: number;
		lastCompletedAt?: number;
		lastSuccessAt?: number;
		lastError?: string;
	} = {
		inProgress: false,
		initialized: false,
		isLKG: false,
	};

	constructor(options?: string | MCPAggregatorOptions) {
		const normalizedOptions =
			typeof options === "string" ? { configPath: options } : (options ?? {});

		this.configPath =
			normalizedOptions.configPath ||
			path.join(process.cwd(), "config", "mcp_servers.json");
		this.configStore = new MCPConfigStore(this.configPath);
		this.createClient =
			normalizedOptions.createClient ??
			((name: string, config: MCPServerConfig) =>
				new StdioClient(name, {
					...config,
					env: config.env ?? {},
				}));
		this.restartDelayMs = normalizedOptions.restartDelayMs ?? 5_000;
		this.connectTimeoutMs = 60_000; // 60s per-server connection timeout for npx-based servers
		this.now = normalizedOptions.now ?? (() => Date.now());
		this.trafficInspector = new MCPTrafficInspector(
			normalizedOptions.maxTrafficEvents ?? 200,
		);
		this.lazyMode = Boolean(normalizedOptions.lazyMode);
	}

	/**
	 * Update the lazy mode flag at runtime. When true, listAggregatedTools() will
	 * skip servers that are not already connected, preventing eager process spawning
	 * during tool listing. executeTool() always connects on demand regardless.
	 */
	public setLazyMode(lazyMode: boolean): void {
		this.lazyMode = lazyMode;
	}

	public getLazyMode(): boolean {
		return this.lazyMode;
	}

	public async initialize(): Promise<void> {
		this.initializationState.inProgress = true;
		this.initializationState.lastStartedAt = this.now();
		this.initializationState.lastError = undefined;
		this.initializationState.isLKG = false;
		try {
			// Check if primary exists, if not we'll be using LKG
			const primaryExists = fs.existsSync(this.configPath);
			const config = this.configStore.readAll();

			if (!primaryExists && Object.keys(config).length > 0) {
				this.initializationState.isLKG = true;
				console.warn(
					`[MCPAggregator] Primary config not found, initialized with LKG from ${this.configPath.replace(/\.json$/, ".lkg.json")}`,
				);
			}

			for (const [name, serverCfg] of Object.entries(config)) {
				this.serverStates.set(name, {
					name,
					config: serverCfg,
					status: serverCfg.enabled ? "stopped" : "stopped",
					toolCount: 0,
					restartCount: 0,
				});
			}
			this.initializationState.initialized = true;
			this.initializationState.lastCompletedAt = this.now();
			this.initializationState.lastSuccessAt =
				this.initializationState.lastCompletedAt;
		} catch (error) {
			this.initializationState.lastCompletedAt = this.now();
			this.initializationState.lastError =
				error instanceof Error ? error.message : String(error);
			console.error("[MCPAggregator] Failed to load config:", error);
		} finally {
			this.initializationState.inProgress = false;
		}
	}

	public getInitializationStatus() {
		return {
			...this.initializationState,
			connectedClientCount: this.clients.size,
			configuredServerCount: this.serverStates.size,
		};
	}

	public seedAdvertisedInventory(seed: AdvertisedInventorySeed): void {
		const advertisedByName = new Map(
			seed.servers.map((server) => [server.name, server]),
		);

		for (const [name, state] of this.serverStates.entries()) {
			const advertised = advertisedByName.get(name);
			if (!advertised) {
				continue;
			}

			this.serverStates.set(name, {
				...state,
				advertisedAlwaysOn: Boolean(advertised.alwaysOnAdvertised),
				advertisedToolCount:
					advertised.advertisedToolCount ?? state.advertisedToolCount ?? 0,
				advertisedSource: seed.source ?? state.advertisedSource,
				warmupStatus:
					state.warmupStatus ??
					(advertised.alwaysOnAdvertised ? "scheduled" : "idle"),
				config: {
					...state.config,
					alwaysOn: Boolean(
						advertised.alwaysOnAdvertised || state.config.alwaysOn,
					),
				},
			});
		}
	}

	public warmAdvertisedServers(): void {
		for (const [name, state] of this.getEnabledServerEntries()) {
			if (!state.advertisedAlwaysOn) {
				continue;
			}

			if (state.status === "connected") {
				this.serverStates.set(name, {
					...state,
					warmupStatus: "ready",
				});
				continue;
			}

			if (state.warmupStatus === "warming") {
				continue;
			}

			this.serverStates.set(name, {
				...state,
				warmupStatus: "scheduled",
			});

			void this.ensureConnectedClient(name)
				.then(() => {
					const nextState = this.serverStates.get(name);
					if (!nextState) {
						return;
					}

					this.serverStates.set(name, {
						...nextState,
						warmupStatus: "ready",
					});
				})
				.catch((error) => {
					const nextState = this.serverStates.get(name);
					if (!nextState) {
						return;
					}

					this.serverStates.set(name, {
						...nextState,
						warmupStatus: "failed",
						lastError: error instanceof Error ? error.message : String(error),
					});
				});
		}
	}

	public async connectToServer(
		name: string,
		config: MCPServerConfig,
	): Promise<void> {
		const state = this.upsertState(name, config);
		state.status = "connecting";
		state.warmupStatus = state.advertisedAlwaysOn
			? "warming"
			: state.warmupStatus;

		console.log(`[MCPAggregator] Connecting to downstream server: ${name}...`);
		try {
			const previousClient = this.clients.get(name);
			if (previousClient) {
				await previousClient.close();
			}

			const client = this.createClient(name, config);
			// Race the connection against a timeout so a single broken server
			// (e.g. npx failing due to SSL) doesn't block the entire aggregator.
			const connectPromise = client.connect();
			const timeoutPromise = new Promise<void>((_, reject) =>
				setTimeout(
					() =>
						reject(
							new Error(
								`Connection to '${name}' timed out after ${this.connectTimeoutMs}ms`,
							),
						),
					this.connectTimeoutMs,
				),
			);
			await Promise.race([connectPromise, timeoutPromise]);
			this.clients.set(name, client);
			state.status = "connected";
			state.lastConnectedAt = this.now();
			state.warmupStatus = state.advertisedAlwaysOn
				? "ready"
				: state.warmupStatus;
			state.lastError = undefined;
			console.log(`[MCPAggregator] ✓ Connected to ${name}`);
		} catch (error) {
			state.status = "error";
			state.warmupStatus = state.advertisedAlwaysOn
				? "failed"
				: state.warmupStatus;
			state.lastError = error instanceof Error ? error.message : String(error);
			state.lastErrorAt = this.now();
			console.error(`[MCPAggregator] ❌ Failed to connect to ${name}:`, error);
		}
	}

	public async executeTool(name: string, args: unknown): Promise<unknown> {
		const namespacedTool = parseNamespacedToolName(name);
		if (namespacedTool) {
			const client = await this.ensureConnectedClient(
				namespacedTool.serverName,
			);

			return this.callToolOnServer(
				namespacedTool.serverName,
				namespacedTool.toolName,
				args,
				client,
			);
		}

		for (const [serverName] of this.getEnabledServerEntries()) {
			// In lazy mode, skip servers not already connected.
			if (this.lazyMode && !this.clients.has(serverName)) continue;
			// Skip servers that recently failed to avoid blocking on every tool call.
			const serverState = this.serverStates.get(serverName);
			if (
				serverState &&
				serverState.status === "error" &&
				serverState.lastErrorAt
			) {
				const timeSinceError = this.now() - serverState.lastErrorAt;
				if (timeSinceError < 120_000) continue; // Cool-down: skip for 2 minutes
			}
			const client = await this.ensureConnectedClient(serverName);
			const tools = await this.listToolsForServer(serverName, client);
			if (tools.find((tool) => tool.name === name)) {
				return this.callToolOnServer(serverName, name, args, client);
			}
		}

		throw new Error(`Tool '${name}' not found in any connected MCP server.`);
	}

	public async listAggregatedTools(): Promise<MCPAggregatedTool[]> {
		const allTools: MCPAggregatedTool[] = [];

		for (const [serverName] of this.getEnabledServerEntries()) {
			// In lazy mode, skip servers that are not already connected. Spawning every
			// configured binary just to enumerate available tools defeats deferred startup.
			// Callers such as the MCP router prefer the database-cached tool inventory;
			// this path is only hit as a last resort where the cache is empty.
			if (this.lazyMode && !this.clients.has(serverName)) {
				continue;
			}
			// Skip servers in error state (recently failed) to avoid blocking.
			const aggServerState = this.serverStates.get(serverName);
			if (
				aggServerState &&
				aggServerState.status === "error" &&
				aggServerState.lastErrorAt
			) {
				const timeSinceError = this.now() - aggServerState.lastErrorAt;
				if (timeSinceError < 120_000) continue;
			}

			const client = await this.ensureConnectedClient(serverName);
			const tools = await this.listToolsForServer(serverName, client);
			const namespaced = tools.map((tool) =>
				this.namespaceTool(serverName, tool),
			);
			allTools.push(...namespaced);
		}

		return allTools;
	}

	public async addServerConfig(
		name: string,
		config: MCPServerConfig,
	): Promise<void> {
		this.configStore.upsert(name, config);
		this.upsertState(name, config);
		await this.connectToServer(name, config);
		console.log(`[MCPAggregator] Saved config for ${name}`);
	}

	public async removeServerConfig(name: string): Promise<void> {
		const client = this.clients.get(name);
		if (client) {
			await client.close();
			this.clients.delete(name);
		}

		this.configStore.remove(name);
		this.serverStates.delete(name);
	}

	public async notifyServerExit(name: string, error?: Error): Promise<void> {
		const state = this.serverStates.get(name);
		if (!state) {
			return;
		}

		this.clients.delete(name);
		state.status = "restarting";
		state.restartCount += 1;
		state.lastError = error?.message;

		if (this.restartDelayMs > 0) {
			await new Promise((resolve) => setTimeout(resolve, this.restartDelayMs));
		}

		await this.connectToServer(name, state.config);
	}

	public async listServers(): Promise<MCPServerState[]> {
		const config = this.configStore.readAll();
		for (const [name, serverConfig] of Object.entries(config)) {
			this.upsertState(name, serverConfig);
		}

		return Array.from(this.serverStates.values()).map((state) => ({
			...state,
		}));
	}

	public async searchTools(query: string): Promise<MCPAggregatedTool[]> {
		const normalizedQuery = query.trim().toLowerCase();
		const tools = await this.listAggregatedTools();

		if (!normalizedQuery) {
			return tools;
		}

		return tools.filter((tool) => {
			const haystack =
				`${tool.name} ${tool._originalName} ${tool.advertisedName ?? ""} ${tool.description ?? ""} ${tool.server} ${tool.serverDisplayName ?? ""} ${(tool.serverTags ?? []).join(" ")} ${(tool.toolTags ?? []).join(" ")} ${tool.semanticGroup ?? ""} ${tool.semanticGroupLabel ?? ""} ${(tool.keywords ?? []).join(" ")}`.toLowerCase();
			return haystack.includes(normalizedQuery);
		});
	}

	public getTrafficEvents() {
		return this.trafficInspector.getEvents();
	}

	public async shutdown(): Promise<void> {
		for (const client of this.clients.values()) {
			await client.close();
		}
		this.clients.clear();
	}

	private upsertState(name: string, config: MCPServerConfig): MCPServerState {
		const existing = this.serverStates.get(name);
		const nextState: MCPServerState = existing
			? { ...existing, config }
			: {
					name,
					config,
					status: config.enabled ? "stopped" : "stopped",
					toolCount: 0,
					advertisedToolCount: 0,
					advertisedAlwaysOn: Boolean(config.alwaysOn),
					warmupStatus: config.alwaysOn ? "scheduled" : "idle",
					restartCount: 0,
				};

		nextState.config = {
			...nextState.config,
			alwaysOn: nextState.config.alwaysOn ?? nextState.advertisedAlwaysOn,
		};
		nextState.advertisedAlwaysOn = Boolean(
			nextState.advertisedAlwaysOn || nextState.config.alwaysOn,
		);
		nextState.warmupStatus =
			nextState.warmupStatus ??
			(nextState.advertisedAlwaysOn ? "scheduled" : "idle");

		this.serverStates.set(name, nextState);
		return nextState;
	}

	private getEnabledServerEntries(): Array<[string, MCPServerState]> {
		return Array.from(this.serverStates.entries()).filter(([, state]) =>
			Boolean(state.config.enabled),
		);
	}

	private async ensureConnectedClient(name: string): Promise<MCPClientLike> {
		const existingClient = this.clients.get(name);
		if (existingClient) {
			const state = this.serverStates.get(name);
			if (state && state.advertisedAlwaysOn && state.warmupStatus !== "ready") {
				this.serverStates.set(name, {
					...state,
					warmupStatus: "ready",
				});
			}
			return existingClient;
		}

		const state = this.serverStates.get(name);
		if (!state || !state.config.enabled) {
			throw new Error(`Tool '${name}' not found in any connected MCP server.`);
		}

		await this.connectToServer(name, state.config);
		const connectedClient = this.clients.get(name);
		if (!connectedClient) {
			throw new Error(`Failed to connect to downstream MCP server '${name}'.`);
		}

		return connectedClient;
	}

	private namespaceTool(
		serverName: string,
		tool: MCPToolDefinition,
	): MCPAggregatedTool {
		const derived = deriveSemanticCatalogForServer({
			serverName,
			tools: [
				{
					name: tool.name,
					title: typeof tool.title === "string" ? tool.title : null,
					description: tool.description ?? null,
					inputSchema:
						typeof tool.inputSchema === "object" && tool.inputSchema !== null
							? (tool.inputSchema as Record<string, unknown>)
							: null,
					alwaysOn: false,
				},
			],
		});
		const derivedTool = derived.tools[0];

		return {
			...tool,
			server: serverName,
			name: namespaceToolName(serverName, tool.name),
			_originalName: tool.name,
			description: `[${serverName}] ${tool.description ?? ""}`.trim(),
			serverDisplayName: derivedTool?.serverDisplayName ?? serverName,
			advertisedName:
				derivedTool?.advertisedName ?? namespaceToolName(serverName, tool.name),
			serverTags: derivedTool?.serverTags ?? [],
			toolTags: derivedTool?.toolTags ?? [],
			semanticGroup: derivedTool?.semanticGroup ?? "general-utility",
			semanticGroupLabel: derivedTool?.semanticGroupLabel ?? "general utility",
			keywords: derivedTool?.keywords ?? [],
			alwaysOn: Boolean(derivedTool?.alwaysOn),
		};
	}

	private async listToolsForServer(
		serverName: string,
		client: MCPClientLike,
	): Promise<MCPToolDefinition[]> {
		const start = this.now();

		try {
			const tools = await client.listTools();
			const state = this.serverStates.get(serverName);
			if (state) {
				state.toolCount = tools.length;
				state.status = "connected";
			}

			this.trafficInspector.record({
				server: serverName,
				method: "tools/list",
				paramsSummary: "",
				latencyMs: this.now() - start,
				success: true,
				timestamp: this.now(),
			});

			return tools;
		} catch (error) {
			const state = this.serverStates.get(serverName);
			if (state) {
				state.status = "error";
				state.lastError =
					error instanceof Error ? error.message : String(error);
			}

			this.trafficInspector.record({
				server: serverName,
				method: "tools/list",
				paramsSummary: "",
				latencyMs: this.now() - start,
				success: false,
				timestamp: this.now(),
				error: error instanceof Error ? error.message : String(error),
			});

			console.error(
				`[MCPAggregator] Error listing tools from ${serverName}:`,
				error,
			);
			return [];
		}
	}

	private async callToolOnServer(
		serverName: string,
		toolName: string,
		args: unknown,
		client: MCPClientLike,
	): Promise<unknown> {
		const start = this.now();

		try {
			const result = await client.callTool(toolName, args);
			const state = this.serverStates.get(serverName);
			if (state) {
				state.status = "connected";
				state.lastError = undefined;
			}

			this.trafficInspector.record({
				server: serverName,
				method: "tools/call",
				paramsSummary: summarizeParams(args),
				latencyMs: this.now() - start,
				success: true,
				timestamp: this.now(),
				toolName,
			});
			return result;
		} catch (error) {
			const state = this.serverStates.get(serverName);
			if (state) {
				state.lastError =
					error instanceof Error ? error.message : String(error);
			}

			this.trafficInspector.record({
				server: serverName,
				method: "tools/call",
				paramsSummary: summarizeParams(args),
				latencyMs: this.now() - start,
				success: false,
				timestamp: this.now(),
				toolName,
				error: error instanceof Error ? error.message : String(error),
			});

			throw error;
		}
	}
}
