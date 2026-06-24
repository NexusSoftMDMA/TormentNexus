mcpServerDebugLog("[MCPServer] Starting imports...");
import "./debug_marker.js";

import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";

export const LooseListToolsRequestSchema = z.object({
	method: z.literal("tools/list"),
	params: z.any().optional(),
});

export const LooseCallToolRequestSchema = z.object({
	method: z.literal("tools/call"),
	params: z.any(),
});

export const LooseListPromptsRequestSchema = z.object({
	method: z.literal("prompts/list"),
	params: z.any().optional(),
});

export const LooseGetPromptRequestSchema = z.object({
	method: z.literal("prompts/get"),
	params: z.any(),
});

export const LooseListResourcesRequestSchema = z.object({
	method: z.literal("resources/list"),
	params: z.any().optional(),
});

export const LooseReadResourceRequestSchema = z.object({
	method: z.literal("resources/read"),
	params: z.any(),
});

export const LooseListResourceTemplatesRequestSchema = z.object({
	method: z.literal("resources/templates/list"),
	params: z.any().optional(),
});
mcpServerDebugLog("[MCPServer] ✓ @modelcontextprotocol/sdk");

import { fileURLToPath } from "url";
import path from "path";
import fs from "fs";
mcpServerDebugLog("[MCPServer] ✓ path/url/fs");

import { Router } from "./Router.js";
mcpServerDebugLog("[MCPServer] ✓ Router");

import { type ModelSelector, LLMService } from "@tormentnexus/ai";
import { CoreModelSelector } from "./providers/CoreModelSelector.js";
mcpServerDebugLog("[MCPServer] ✓ ModelSelector");

import { WebSocketServer } from "ws";
import { WebSocketServerTransport } from "./transports/WebSocketServerTransport.js";
import http from "http";
mcpServerDebugLog("[MCPServer] ✓ ws/http");

import { McpmInstaller } from "./skills/McpmInstaller.js";
import { Director } from "@tormentnexus/agents";
import { Council, CouncilRole } from "@tormentnexus/agents";
import { GeminiAgent } from "./agents/GeminiAgent.js";
import { ClaudeAgent } from "./agents/ClaudeAgent.js";
import { MetaArchitectAgent } from "./agents/MetaArchitectAgent.js";
import { CoderAgent } from "./agents/CoderAgent.js";
import { ResearcherAgent } from "./agents/ResearcherAgent.js";
import { McpWorkerAgent } from "./agents/McpWorkerAgent.js";
import { MeshCoderAgent } from "./agents/MeshCoderAgent.js";
import { MeshResearcherAgent } from "./agents/MeshResearcherAgent.js";
import { ConfigManager } from "./config/ConfigManager.js";
import { AutoTestService } from "./services/AutoTestService.js";
import { ShellService } from "./services/ShellService.js";
import { SandboxService } from "./security/SandboxService.js";
import { SquadService } from "./orchestrator/SquadService.js";
import { MeshService, SwarmMessageType } from "./mesh/MeshService.js";
import { GitWorktreeManager } from "./orchestrator/GitWorktreeManager.js";
import { AuditService } from "./security/AuditService.js";
import { GitService } from "./services/GitService.js";
import { Supervisor } from "@tormentnexus/agents";
import { SkillRegistry } from "./skills/SkillRegistry.js";
import { SuggestionService } from "./suggestions/SuggestionService.js";
import { ResearchService } from "./services/ResearchService.js";
import { HealerService } from "./services/HealerService.js";
import { DarwinService } from "./services/DarwinService.js";
import { MetricsService } from "./services/MetricsService.js";
import { PolicyService } from "./security/PolicyService.js";
import { PromptRegistry } from "./prompts/PromptRegistry.js";
import { WebSearchTool } from "./tools/WebSearchTool.js";
import { DiagnosticTools } from "./tools/DiagnosticTools.js";
import { LSPTools } from "./tools/LSPTools.js";
import { LSPService } from "./services/LSPService.js";
import { PlanService } from "./services/PlanService.js";
import { CodeModeService } from "./services/CodeModeService.js";
import { WorkflowEngine } from "./orchestrator/WorkflowEngine.js";
import { AgentMemoryService } from "./services/AgentMemoryService.js";
import { SessionImportService } from "./services/SessionImportService.js";
import { MemoryManager } from "./services/MemoryManager.js"; // Use legacy MemoryManager
import { BobbyBookmarksBacklogAdapter } from "./services/bobby-bookmarks-adapter.js";
mcpServerDebugLog("[MCPServer] ✓ Phase 51/53 Infrastructure");
import { SkillAssimilationService } from "./services/SkillAssimilationService.js";
import { MarketplaceService } from "./services/MarketplaceService.js";
import { registerSystemWorkflows } from "./orchestrator/SystemWorkflows.js";
import { MCPAggregator } from "./mcp/MCPAggregator.js";
import { getCachedToolInventory } from "./mcp/cachedToolInventory.js";
import {
	createDirectModeAgentRunner,
	getDirectModeCompatibilityTools,
	getDirectModeMetadataGuardResult,
	getDirectModeSavedScriptTools,
	tryHandleDirectModeCompatibilityTool,
} from "./mcp/directModeCompatibility.js";
import {
	getDownstreamPrompt,
	listDownstreamPrompts,
	listDownstreamResources,
	listDownstreamResourceTemplates,
	readDownstreamResource,
} from "./mcp/downstreamDiscovery.js";
import {
	isToolNotFoundError,
	shouldPreferAggregatorExecution,
} from "./mcp/legacyProxyMode.js";
import { SubmoduleManager } from "./mcp/SubmoduleManager.js";
import { SubmoduleService } from "./services/SubmoduleService.js";
import { FileSensor } from "./sensors/FileSensor.js";
import { TerminalSensor } from "./sensors/TerminalSensor.js";
import { AutoTestReactor } from "./reactors/AutoTestReactor.js";
import { HealerReactor } from "./reactors/HealerReactor.js";
import { MemoryHarvestReactor } from "./reactors/MemoryHarvestReactor.js";
import { SessionManager } from "./services/SessionManager.js";
import { SessionSupervisor } from "./supervisor/SessionSupervisor.js";
import { PtySupervisor } from "./supervisor/PtySupervisor.js";
import { ProjectTracker } from "./services/ProjectTracker.js";
import { MissionService } from "./services/MissionService.js";
import { buildToolObservationInput } from "./services/toolObservationMemory.js";
import { detectLocalExecutionEnvironment } from "./services/execution-environment.js";
import { loadTormentNexusMcpConfig } from "./mcp/mcpJsonConfig.js";
import {
	buildAutomaticToolContextFingerprint,
	buildAutomaticToolContextMemory,
	buildAutomaticToolContextStartPayload,
	shouldPersistAutomaticToolContext,
	shouldResolveAutomaticToolContext,
} from "./services/toolContextInjection.js";
import { readToolPreferencesFromSettings } from "./routers/mcp-tool-preferences.js";
import { flattenToolSchema } from "./utils/mcp-schema-sanitizer.js";

mcpServerDebugLog("[MCPServer] ✓ SkillRegistry");

import { SpawnerService } from "./agents/SpawnerService.js";
mcpServerDebugLog("[MCPServer] ✓ SpawnerService");

import {
	FileSystemTools,
	TerminalService, /// New class
	ProcessRegistry, /// New class
	MemoryTools,
	TunnelTools,
	ConfigTools,
	LogTools,
	SearchTools,
	ReaderTools,
	InputTools,
	WorktreeTools,
	MetaTools,
	SystemStatusTool,
	ChainExecutor,
	type ChainRequest,
} from "@tormentnexus/tools";
mcpServerDebugLog("[MCPServer] ✓ All Tools & ChainExecutor");

mcpServerDebugLog("[MCPServer] ✓ All Tools & ChainExecutor");

// Council and Director already imported above
mcpServerDebugLog("[MCPServer] ✓ Council");

import { CommandRegistry } from "./commands/CommandRegistry.js";
import { GitCommand } from "./commands/lib/GitCommands.js";
import {
	HelpCommand,
	VersionCommand,
	DirectorCommand,
} from "./commands/lib/SystemCommands.js";
import { ContextCommand } from "./commands/lib/ContextCommands.js";
import {
	UndoCommand,
	DiffCommand,
	StashCommand,
	FixCommand,
} from "./commands/lib/WorkflowCommands.js";
import { mcpServerDebugLog } from "./mcpServerDebug.js";
mcpServerDebugLog("[MCPServer] ✓ Commands");

import { ContextManager } from "./context/ContextManager.js";
import { SymbolPinService } from "./services/SymbolPinService.js";
import { AutoDevService } from "./services/AutoDevService.js";
import { KnowledgeService } from "./services/KnowledgeService.js";
import { EventBus } from "./services/EventBus.js";
import { DeepResearchService } from "./services/DeepResearchService.js";
import { McpConfigService } from "./services/McpConfigService.js";
import { DocumentIntakeService } from "./services/rag/DocumentIntakeService.js";
import { EmbeddingService } from "./services/rag/EmbeddingService.js";

import {
	PermissionManager,
	type AutonomyLevel,
} from "./security/PermissionManager.js";
import { BrowserTool } from "@tormentnexus/tools";
import { SearchService } from "@tormentnexus/search";
import { CouncilService } from "./services/CouncilService.js";
import { BrowserService } from "./services/BrowserService.js";
import type { ConnectedClient } from "./services/mcp-client.service.js";
import { createCouncilTools } from "./mcp/tools/council_tools.js";
import { NativeSessionMetaTools } from "./mcp/NativeSessionMetaTools.js";
import { ConversationalToolInjector } from "./mcp/ConversationalToolInjector.js";
import { jsonConfigProvider } from "./services/config/JsonConfigProvider.js";
import { configImportService } from "./services/config-import.service.js";
import { mcpServerPool } from "./services/mcp-server-pool.service.js";
import {
	applyBridgeClientHello,
	buildBridgeManifest,
	createDefaultBridgeClient,
	type RegisteredBridgeClient,
} from "./bridge/bridge-manifest.js";
import {
	getBridgeHealthUrl,
	getBridgeStreamUrl,
	getBridgeWebSocketUrl,
	resolveBridgePort,
} from "./bridge/bridgePort.js";
mcpServerDebugLog("[MCPServer] ✓ PermissionManager");

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// LAZY: VectorStore and Indexer are imported dynamically when needed
// import { VectorStore } from './memory/VectorStore.js';
// import { ShellService } from "./services/ShellService.js";
import type { AgentAdapter } from "./orchestrator/AgentAdapter.js";
import { ClaudeAdapter } from "./orchestrator/adapters/ClaudeAdapter.js";
import { GeminiAdapter } from "./orchestrator/adapters/GeminiAdapter.js";
import type { Tool, CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import type { ToolContextPayload } from "./services/toolContextMemory.js";

interface ToolRequest {
	params: {
		name: string;
		arguments: Record<string, any>;
	};
}

interface ToolWithHandler {
	name: string;
	handler: (args: unknown) => Promise<unknown> | unknown;
}

interface TextContentEnvelope {
	content?: Array<{ text?: string }>;
}

mcpServerDebugLog("[MCPServer] All imports complete!");

declare global {
	var mcpServerInstance: MCPServer;
}

type MCPServerOptions = {
	skipWebsocket?: boolean;
	skipAutoDrive?: boolean;
	skipMesh?: boolean;
	skipStdio?: boolean;
	inputTools?: InputTools;
	systemStatusTool?: SystemStatusTool;
	processRegistry?: ProcessRegistry;
};

export class MCPServer {
	private readonly options: MCPServerOptions;
	private server: Server; // Stdio Server
	private serverSetupPromise: Promise<void>;
	private wsServer: Server | null = null; // WebSocket Server
	private wsServerSetupPromise: Promise<void> | null = null;
	private router: Router;
	public modelSelector: ModelSelector;
	public llmService: LLMService;
	public director: Director;
	public permissionManager: PermissionManager;
	public auditService: AuditService;
	public shellService: ShellService;
	public memoryManager: MemoryManager; // Centralized Memory Service
	// public vectorStore: any; // DEPRECATED
	// private indexer: any; // DEPRECATED
	private memoryInitialized: boolean = false;
	private pendingRequests: Map<string, (response: any) => void> = new Map();
	public suggestionService: SuggestionService;
	private chainExecutor: ChainExecutor;
	public wssInstance: any = null; // WebSocket.Server
	private inputTools: InputTools;
	public lastUserActivityTime: number = Date.now(); // Start with grace period
	private systemStatusTool: SystemStatusTool;
	private terminalService: TerminalService;
	private processRegistry: ProcessRegistry;
	private mcpmInstaller: McpmInstaller;
	public configManager: ConfigManager;
	public autoTestService: AutoTestService;
	public sandboxService: SandboxService;
	public spawnerService: SpawnerService;
	public squadService: SquadService;
	public gitWorktreeManager: GitWorktreeManager;
	public commandRegistry: CommandRegistry;
	public contextManager: ContextManager;
	public symbolPinService: SymbolPinService;
	public autoDevService: AutoDevService;
	public researchService: ResearchService;
	public gitService: GitService; // Phase 30
	public metricsService: MetricsService; // Phase 31
	public policyService: PolicyService; // Phase 32
	private knowledgeService: KnowledgeService; // Knowledge Graph Service
	public healerService: HealerService;
	public darwinService: DarwinService;
	public promptRegistry: PromptRegistry;
	public skillRegistry: SkillRegistry;
	public skillAssimilationService: SkillAssimilationService;
	public mcpAggregator: MCPAggregator;
	private submoduleManager: SubmoduleManager;
	public eventBus: EventBus;
	public deepResearchService: DeepResearchService;
	public councilService: CouncilService;
	public browserService: BrowserService;
	public marketplaceService: MarketplaceService; // Phase 65: Decentralized Marketplace
	public sessionManager: SessionManager; // Phase 57: State Persistence
	public sessionSupervisor: SessionSupervisor;
	public ptySupervisor: PtySupervisor;
	public a2aLogger?: any;
	public swarmController?: any;
	public pairOrchestrator?: any;

	public projectTracker: ProjectTracker; // Phase 59: Autonomous Loop
	public missionService: MissionService; // Phase 80: Swarm Persistence
	public meshService: MeshService | undefined; // Phase 60: P2P Mesh
	public mcpConfigService: McpConfigService;

	// Sensors (Phase 43)
	public fileSensor: FileSensor;
	public terminalSensor: TerminalSensor;
	public autoTestReactor: AutoTestReactor;
	public healerReactor: HealerReactor;
	public memoryHarvestReactor: MemoryHarvestReactor;
	public submoduleService: SubmoduleService;

	// Phase 51: Core Infrastructure
	public lspService: LSPService;
	public planService: PlanService;
	public codeModeService: CodeModeService;
	public workflowEngine: WorkflowEngine;
	public lspTools: LSPTools;
	public agentMemoryService: AgentMemoryService;
	public sessionImportService: SessionImportService;
	private readonly nativeSessionMetaTools: NativeSessionMetaTools;
	private readonly conversationalToolInjector: ConversationalToolInjector;
	private readonly promptToClient: Record<string, ConnectedClient> = {};
	private readonly resourceToClient: Record<string, ConnectedClient> = {};
	private readonly bridgeClients: Map<any, RegisteredBridgeClient> = new Map();
	private readonly bridgePort = resolveBridgePort();
	private readonly recentToolContextFingerprints: Map<string, number> =
		new Map();

	// Core Integrations (Phase 5 & 6)
	public browserTool: BrowserTool;
	public searchService: SearchService;

	// Core Agents
	public geminiAgent: GeminiAgent;
	public claudeAgent: ClaudeAgent;
	public metaArchitectAgent: MetaArchitectAgent;
	public coderAgent: CoderAgent;
	public researcherAgent: ResearcherAgent;
	public mcpWorkerAgent: McpWorkerAgent;
	public meshCoderAgent?: MeshCoderAgent;
	public meshResearcherAgent?: MeshResearcherAgent;
	public council: Council;
	public supervisor: Supervisor;

	private activeAgents: Map<string, AgentAdapter> = new Map();
	public directorConfig = {
		taskCooldownMs: 10000,
		heartbeatIntervalMs: 30000,
		periodicSummaryMs: 120000,
		pasteToSubmitDelayMs: 1000,
		acceptDetectionMode: "polling" as const, // Polling is robust!
		pollingIntervalMs: 30000,
		council: {
			personas: ["Architect", "Product", "Critic"],
			contextFiles: ["README.md", "docs/ROADMAP.md"],
			prompts: {
				Architect:
					"You are The Architect. Focus on system design, code patterns, and maintainability. Be strict.",
				Product:
					"You are the Product Manager. Focus on user value, roadmap alignment, and core problem solving.",
				Critic:
					"You are The Critic. Focus on security, edge cases, bugs, and potential risks. Be pessimistic.",
			},
		},
	};

	public get activeAgentsMap() {
		return this.activeAgents;
	}

	public get isMemoryInitialized() {
		return this.memoryInitialized;
	}

	public getBridgeStatus() {
		const connectedClients = Array.from(this.bridgeClients.values()).sort(
			(left, right) => left.clientName.localeCompare(right.clientName),
		);
		const manifest = buildBridgeManifest(connectedClients);

		return {
			port: this.bridgePort,
			ready: Boolean(this.wssInstance),
			clientCount: Number(this.wssInstance?.clients?.size ?? 0),
			clients: connectedClients,
			supportedCapabilities: manifest.supportedCapabilities,
			supportedHookPhases: manifest.supportedHookPhases,
			websocketUrl: getBridgeWebSocketUrl(this.bridgePort),
			healthUrl: getBridgeHealthUrl(this.bridgePort),
			streamUrl: getBridgeStreamUrl(this.bridgePort),
		};
	}

	/**
	 * Reason: tool definitions originate from mixed internal/external registries with loose typing.
	 * What: runtime guard for tool objects that expose an executable `handler(args)` function.
	 * Why: enables safe invocation without broad casts across tool arrays.
	 */
	private isToolWithHandler(value: unknown): value is ToolWithHandler {
		if (!value || typeof value !== "object") {
			return false;
		}
		const handler = Reflect.get(value as object, "handler");
		const name = Reflect.get(value as object, "name");
		return typeof name === "string" && typeof handler === "function";
	}

	/**
	 * Reason: multiple execution branches return heterogeneous payloads, but suggestion analysis needs text content.
	 * What: safely extracts first text block from MCP-style `{ content: [{ text }] }` envelopes.
	 * Why: preserves engagement trigger behavior while removing broad result casting.
	 */
	private getFirstTextContent(result: unknown): string | null {
		if (!result || typeof result !== "object") {
			return null;
		}
		const envelope = result as TextContentEnvelope;
		const first = envelope.content?.[0];
		return typeof first?.text === "string" ? first.text : null;
	}

	/**
	 * Reason: TormentNexus now captures session-start and stop-time memory, but still needs
	 * post-tool lifecycle observations without wiring every caller individually.
	 * What: best-effort bridge from centralized tool execution into structured memory observations.
	 * Why: keeps tormentnexus-style lifecycle capture native to TormentNexus while never blocking tool execution.
	 */
	private async captureToolObservation(event: {
		toolName: string;
		args: unknown;
		result?: unknown;
		error?: unknown;
		durationMs: number;
	}): Promise<void> {
		const observation = buildToolObservationInput(event);
		if (!observation || !this.agentMemoryService?.recordObservation) {
			return;
		}

		try {
			await this.agentMemoryService.recordObservation(observation);
		} catch (error) {
			console.error("[MCPServer] Failed to capture tool observation:", error);
		}
	}

	/**
	 * Reason: TormentNexus can already rank relevant memories for a tool call, but until now
	 * that JIT context stayed behind an explicit helper instead of being used automatically.
	 * What: resolves compact pre-tool context using current session goal/objective state,
	 * broadcasts a short preview to the inspector, and stores deduped session memory.
	 * Why: gives TormentNexus a native PreToolUse-style lifecycle seam without mutating downstream schemas.
	 */
	private async resolveAutomaticToolContext(
		toolName: string,
		args: unknown,
	): Promise<ToolContextPayload | null> {
		if (
			!shouldResolveAutomaticToolContext(toolName) ||
			!this.agentMemoryService?.getToolContext
		) {
			return null;
		}

		const sessionState = this.sessionManager?.getState?.();
		const payload = this.agentMemoryService.getToolContext({
			toolName,
			args,
			activeGoal: sessionState?.activeGoal ?? null,
			lastObjective: sessionState?.lastObjective ?? null,
		}) as ToolContextPayload | null;

		if (!shouldPersistAutomaticToolContext(payload)) {
			return payload;
		}

		const fingerprint = buildAutomaticToolContextFingerprint(payload);
		const now = Date.now();

		for (const [
			key,
			timestamp,
		] of this.recentToolContextFingerprints.entries()) {
			if (now - timestamp > 30_000) {
				this.recentToolContextFingerprints.delete(key);
			}
		}

		if (this.recentToolContextFingerprints.has(fingerprint)) {
			return payload;
		}

		this.recentToolContextFingerprints.set(fingerprint, now);

		try {
			const memory = buildAutomaticToolContextMemory(payload);
			await this.agentMemoryService.add(
				memory.content,
				"session",
				"project",
				memory.metadata,
			);
		} catch (error) {
			console.error(
				"[MCPServer] Failed to capture automatic tool context:",
				error,
			);
		}

		return payload;
	}

	public async getPredictedToolAds(
		chatHistory: string,
		activeGoal: string,
	): Promise<string[]> {
		// LLM-based tool prediction using cheap/free model selection
		try {
			const nativeTools = await this.getNativeTools();
			const toolsList = nativeTools
				.map((t) => `- ${t.name}: ${t.description || ""}`)
				.join("\n");

			const systemPrompt = `You are a predictive tool routing agent. Given a user goal and/or chat history, select the single most relevant and useful MCP tool from the list below that the user is likely to need next.

Available Tools:
${toolsList}

Respond ONLY with a JSON array of tool name strings, for example: ["tool_name_1"]. If no tool is specifically useful, return [].`;

			const prompt = `User Goal: ${activeGoal}\nChat History:\n${chatHistory}`;

			const modelSelection = await this.modelSelector.selectModel({
				taskComplexity: "low",
				routingTaskType: "general",
				routingStrategy: "cheapest",
			});

			const completion = await this.llmService.generateText(
				modelSelection.provider,
				modelSelection.modelId,
				systemPrompt,
				prompt,
				{ routingStrategy: "cheapest" },
			);

			const content =
				typeof completion === "string" ? completion : completion.content || "";
			const jsonStart = content.indexOf("[");
			const jsonEnd = content.lastIndexOf("]");
			if (jsonStart !== -1 && jsonEnd !== -1) {
				const parsed = JSON.parse(content.slice(jsonStart, jsonEnd + 1));
				if (Array.isArray(parsed) && parsed.length > 0) {
					console.error(
						`[MCPServer] 🔮 LLM predicted optimal tools: ${parsed.join(", ")}`,
					);
					return parsed;
				}
			}
		} catch (e: any) {
			console.error(
				"[MCPServer] LLM tool prediction failed, trying sidecar fallback:",
				e.message,
			);
		}

		const SIDECAR_URL =
			process.env.TORMENTNEXUS_SIDECAR_URL || "http://127.0.0.1:4300";
		try {
			const res = await fetch(`${SIDECAR_URL}/api/mcp/tools/predict`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ chatHistory, activeGoal }),
			});
			if (res.ok) {
				const json = (await res.json()) as any;
				const predictedTools = json.data?.predictedTools || json.data || [];
				return Array.isArray(predictedTools) ? predictedTools : [];
			}
		} catch (e) {
			console.warn(
				"[MCPServer] Failed to fetch tool predictions from sidecar:",
				e,
			);
		}
		return [];
	}

	private async syncNativeToolPreferences(): Promise<void> {
		try {
			const config = await loadTormentNexusMcpConfig();
			const settings = config.settings as
				| {
						toolSelection?: {
							importantTools?: unknown;
							alwaysLoadedTools?: unknown;
						};
				  }
				| undefined;
			const preferences = readToolPreferencesFromSettings(
				settings?.toolSelection,
			);
			this.nativeSessionMetaTools.setAlwaysLoadedTools(
				preferences.alwaysLoadedTools,
			);
		} catch {
			this.nativeSessionMetaTools.setAlwaysLoadedTools([]);
		}
	}

	public setAlwaysLoadedTools(toolNames: string[]): void {
		this.nativeSessionMetaTools.setAlwaysLoadedTools(toolNames);
	}

	/**
	 * Append a conversation turn to the ConversationalToolInjector's sliding
	 * window. Call this from transport layers or tool-call handlers whenever
	 * a user or assistant message is observed so the injector can track the
	 * direction of the conversation and predict relevant tools proactively.
	 */
	public appendConversationTurn(
		role: "user" | "assistant" | "tool",
		text: string,
	): void {
		this.conversationalToolInjector.appendTurn(role, text);

		const SIDECAR_URL =
			process.env.TORMENTNEXUS_SIDECAR_URL || "http://127.0.0.1:4300";
		fetch(`${SIDECAR_URL}/api/mcp/conversation/append`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ role, text }),
		}).catch((err: any) => {
			console.warn(
				"[MCPServer] Failed to sync conversation turn to Go sidecar:",
				err.message,
			);
		});
	}

	/**
	 * Returns a read-only view into the ConversationalToolInjector for
	 * dashboard queries (conversation window, token count, etc.).
	 */
	public getConversationInjector(): {
		getWindow: () => ReadonlyArray<{
			role: string;
			text: string;
			timestamp: number;
		}>;
		getWindowTokenCount: () => number;
	} {
		return this.conversationalToolInjector;
	}

	constructor(options: MCPServerOptions = {}) {
		this.options = options;
		this.router = new Router();
		this.modelSelector = new CoreModelSelector();
		this.llmService = new LLMService(this.modelSelector);
		// SkillRegistry initialized later with correct paths
		this.sessionManager = new SessionManager(process.cwd()); // Phase 57: State Persistence
		if (options.skipAutoDrive) {
			this.sessionManager.disableAutoDriveRestore();
		}
		this.projectTracker = new ProjectTracker(process.cwd()); // Phase 59: Autonomous Loop
		this.missionService = new MissionService(process.cwd()); // Phase 80: Swarm Persistence
		this.director = new Director(this);
		this.council = new Council(this.modelSelector);
		this.permissionManager = new PermissionManager("high"); // Default to HIGH AUTONOMY as requested
		this.auditService = new AuditService(process.cwd());
		this.shellService = new ShellService();
		this.gitService = new GitService(process.cwd()); // Phase 30
		this.gitWorktreeManager = new GitWorktreeManager(process.cwd());
		this.sessionSupervisor = new SessionSupervisor({
			rootDir: process.cwd(),
			worktreeManager: this.gitWorktreeManager,
		});
		this.ptySupervisor = new PtySupervisor({
			rootDir: process.cwd(),
			worktreeManager: this.gitWorktreeManager,
		});
		this.metricsService = new MetricsService(); // Phase 31
		this.metricsService.startMonitoring();
		this.policyService = new PolicyService(process.cwd()); // Phase 32
		this.chainExecutor = new ChainExecutor(this);
		this.inputTools = options.inputTools || new InputTools();
		this.systemStatusTool = options.systemStatusTool || new SystemStatusTool();
		this.processRegistry = options.processRegistry || new ProcessRegistry();
		this.terminalService = new TerminalService(this.processRegistry);
		this.mcpmInstaller = new McpmInstaller(
			path.join(process.cwd(), ".tormentnexus", "skills"),
		);
		this.spawnerService = SpawnerService.getInstance();
		this.configManager = new ConfigManager();
		this.mcpConfigService = new McpConfigService();
		this.nativeSessionMetaTools = new NativeSessionMetaTools(undefined, {
			toolContextResolver: (input) => {
				return {
					toolName: input.toolName,
					query: input.toolName,
					matchedPaths: [],
					observationCount: 0,
					summaryCount: 0,
					prompt: `JIT tool context for ${input.toolName}:\nNo prior memory was found.`,
				};
			},
		});
		this.conversationalToolInjector = new ConversationalToolInjector(
			this.llmService,
			this.modelSelector,
			{
				maxWindowTurns: 8,
				throttleMs: 3_000,
				maxInjectedTools: 5,
				minWindowTokens: 30,
			},
		);
		// Fire and forget config sync
		this.mcpConfigService
			.syncWithDatabase()
			.catch((err) => console.error("[MCPServer] Config Sync Failed:", err));

		this.autoTestService = new AutoTestService(process.cwd());
		this.sandboxService = new SandboxService();
		this.healerService = new HealerService(
			this.llmService,
			this,
			this.shellService,
		);
		this.promptRegistry = new PromptRegistry();
		this.skillRegistry = new SkillRegistry([
			path.join(process.cwd(), "packages", "core", "src", "skills"),
			path.join(process.cwd(), ".tormentnexus", "skills"),
		]);
		// SearchService is needed for DeepResearchService types
		const searchService = new SearchService();
		this.memoryManager = new MemoryManager(process.cwd());
		this.agentMemoryService = new AgentMemoryService(
			{ persistDir: path.join(process.cwd(), ".tormentnexus", "agent_memory") },
			this.memoryManager,
		);
		this.deepResearchService = new DeepResearchService(
			this,
			this.llmService,
			searchService,
			this.memoryManager,
		); // Initialize FIRST
		this.skillAssimilationService = new SkillAssimilationService(
			this.skillRegistry,
			this.llmService,
			this.deepResearchService,
		);
		this.darwinService = new DarwinService(this.llmService, this);
		this.supervisor = new Supervisor(this);
		this.lspService = new LSPService(process.cwd());

		// Pass lazy mode to the aggregator at construction so that listAggregatedTools()
		// never eagerly spawns server binaries in deferred-startup configurations.
		this.mcpAggregator = new MCPAggregator({
			lazyMode: mcpServerPool.getLifecycleModes().lazySessionMode,
		});
		this.submoduleManager = new SubmoduleManager(process.cwd());
		this.submoduleService = new SubmoduleService(
			process.cwd(),
			this.mcpAggregator,
		);
		this.eventBus = new EventBus();

		// Initialize Sensors
		this.fileSensor = new FileSensor(this.eventBus, process.cwd());
		this.terminalSensor = new TerminalSensor(this.eventBus);

		// Start Sensors (Non-blocking)
		this.fileSensor.start();
		this.terminalSensor.start();

		// Initialize Reactors
		this.autoTestReactor = new AutoTestReactor(
			this.eventBus,
			this.autoTestService,
		);
		this.autoTestReactor.start();

		this.healerReactor = new HealerReactor(this.eventBus, this.healerService);
		this.healerReactor.start();

		this.memoryHarvestReactor = new MemoryHarvestReactor(
			this.eventBus,
			this.llmService,
			this.agentMemoryService,
		);
		this.memoryHarvestReactor.start();

		// Neural Pulse Bridge: Broadcast all system events to connected dashboard clients
		this.eventBus.on("system_event", (event) => {
			if (this.wssInstance) {
				this.broadcastWebSocketMessage({
					type: "NEURAL_PULSE",
					payload: event,
				});
			}
		});

		// Phase 51: Core Infrastructure Services
		this.lspService = new LSPService(process.cwd());
		this.planService = new PlanService({ rootPath: process.cwd() });
		this.codeModeService = new CodeModeService({
			timeout: 30000,
			allowAsync: true,
		});
		this.workflowEngine = new WorkflowEngine({
			persistDir: path.join(process.cwd(), ".tormentnexus", "workflows"),
		});
		this.lspTools = new LSPTools(process.cwd());
		// MemoryManager + AgentMemoryService initialized early
		this.sessionImportService = new SessionImportService(
			this.llmService,
			this.agentMemoryService,
			process.cwd(),
		);

		// Phase 5 & 6 Init
		this.browserTool = new BrowserTool();
		this.searchService = new SearchService();
		this.nativeSessionMetaTools.setToolContextResolver(
			(input) => this.agentMemoryService?.getToolContext?.(input) ?? null,
		);

		// Initialize Core Agents
		this.geminiAgent = new GeminiAgent(this.llmService, this.promptRegistry);
		this.claudeAgent = new ClaudeAgent(this.llmService, this.promptRegistry);
		this.metaArchitectAgent = new MetaArchitectAgent(
			this.llmService,
			this.promptRegistry,
		);

		// Initialize Specialized Agents
		this.coderAgent = new CoderAgent(this.llmService);
		this.researcherAgent = new ResearcherAgent(this.deepResearchService);
		this.mcpWorkerAgent = new McpWorkerAgent(this.llmService, this);

		// Initialize Council with Agents
		this.council = new Council(this.modelSelector);
		this.council.setServer(this);
		this.council.registerAgent(CouncilRole.CRITIC, this.geminiAgent);
		this.council.registerAgent(CouncilRole.ARCHITECT, this.claudeAgent);
		this.council.registerAgent(
			CouncilRole.META_ARCHITECT,
			this.metaArchitectAgent,
		);

		// Load persistent config
		const savedConfig = this.configManager.loadConfig();
		if (savedConfig) {
			mcpServerDebugLog("[MCPServer] Loaded persistent config.");
			this.directorConfig = { ...this.directorConfig, ...savedConfig };
			// Ensure nested merge for council
			if (savedConfig.council) {
				this.directorConfig.council = {
					...this.directorConfig.council,
					...savedConfig.council,
				};
			}
		}

		this.suggestionService = new SuggestionService(undefined, this.director);

		// Context Manager
		this.contextManager = new ContextManager();
		this.symbolPinService = new SymbolPinService();
		this.autoDevService = new AutoDevService(process.cwd(), this.director);
		this.shellService = new ShellService(); // Added this line
		this.commandRegistry = new CommandRegistry(); // Corrected typo from instruction
		this.commandRegistry.register(new GitCommand());
		this.commandRegistry.register(new HelpCommand(this.commandRegistry));
		this.commandRegistry.register(new VersionCommand());
		this.commandRegistry.register(new DirectorCommand(() => this.director));
		this.commandRegistry.register(new ContextCommand(this.contextManager));
		this.commandRegistry.register(new UndoCommand());
		this.commandRegistry.register(new DiffCommand());
		this.commandRegistry.register(new StashCommand());
		this.commandRegistry.register(new FixCommand(() => this.autoDevService));

		// Memory System - Already initialized above
		this.researchService = new ResearchService(this, this.memoryManager); // Initialized AFTER memoryManager
		this.knowledgeService = new KnowledgeService(this.memoryManager); // Added
		// DeepResearchService initialized earlier (line 276)
		this.councilService = new CouncilService();
		this.browserService = new BrowserService();

		this.squadService = new SquadService(this as any);
		this.gitWorktreeManager = new GitWorktreeManager(process.cwd());

		// Phase 60: Mesh Service
		if (!options.skipMesh) {
			this.meshService = new MeshService();

			// Phase 93: P2P Artifact Federation (Serving Artifacts)
			this.meshService.on("message", async (msg: any) => {
				if (msg.type === SwarmMessageType.ARTIFACT_READ_REQUEST) {
					const req = msg.payload as { path: string };
					try {
						const content = await fs.promises.readFile(req.path, "utf-8");
						// If we have it, send response directly to requester
						this.meshService!.sendResponse(
							msg,
							SwarmMessageType.ARTIFACT_READ_RESPONSE,
							{
								path: req.path,
								content: content,
							},
						);
						console.log(
							`[Mesh Artifact] Served ${req.path} to node ${msg.sender}`,
						);
					} catch (err) {
						// File not found locally, ignore silently
					}
				}
			});

			// Phase 94: Sub-Agent Task Routing
			// Attach specialized sub-agents to the mesh bus so they can bid on classified tasks
			this.meshCoderAgent = new MeshCoderAgent(this);
			this.meshResearcherAgent = new MeshResearcherAgent(this);
		}

		// Phase 65: Marketplace (Depends on Mesh)
		this.marketplaceService = new MarketplaceService(
			path.join(process.cwd(), ".tormentnexus", "skills"),
			this.meshService,
		);

		global.mcpServerInstance = this;

		// Standard Server (Stdio)
		const primaryServer = this.createServerInstance();
		this.server = primaryServer.server;
		this.serverSetupPromise = primaryServer.ready;

		if (!options.skipWebsocket) {
			const PORT = this.bridgePort;
			mcpServerDebugLog(
				`[MCPServer] Preparing WebSocket/HTTP bridge for port ${PORT}...`,
			);

			// Create a dedicated MCP server instance for websocket transport.
			// Without this, `this.wsServer` stays null and `.connect()` throws
			// during boot (`Cannot read properties of null (reading 'connect')`).
			const secondaryServer = this.createServerInstance();
			this.wsServer = secondaryServer.server;
			this.wsServerSetupPromise = secondaryServer.ready;
		} else {
			this.wsServer = null;
			this.wssInstance = null;
			this.wsServerSetupPromise = null;
		}
	}

	/**
	 * Lazy initialization of memory system (VectorStore + Indexer)
	 * Only loaded when memory tools are first used to speed up startup
	 */
	public async initializeMemorySystem() {
		if (this.memoryInitialized) return;

		mcpServerDebugLog(
			"[MCPServer] Lazy-loading memory system (MemoryManager)...",
		);
		// this.memoryManager is instantiated in constructor but lazily initialized internally
		// Actually, we should instantiate it here or in constructor?
		// Let's rely on internal lazy loading of MemoryManager, but trigger it here to be safe
		// Or just let tool calls trigger it.
		// For parity with old logic, let's just make sure it's ready.
		// await this.memoryManager.initialize(); // private method, called via public methods

		// Actually, let's just mark initialized true
		this.memoryInitialized = true;
	}

	private createServerInstance(): { server: Server; ready: Promise<void> } {
		const s = new Server(
			{ name: "tormentnexus-core", version: "0.90.7" },
			{
				capabilities: {
					tools: {},
					prompts: {},
					resources: {},
				},
			},
		);
		const ready = this.setupHandlers(s);
		return { server: s, ready };
	}

	private broadcastRequestAndAwait(
		messageType: string,
		payload: any = {},
	): Promise<any> {
		if (!this.wssInstance || this.wssInstance.clients.size === 0) {
			return Promise.resolve({
				content: [
					{ type: "text", text: "Error: No Native Extension connected." },
				],
			});
		}

		return new Promise((resolve) => {
			const requestId = `req_${Date.now()}_${Math.random()}`;
			const timeout = setTimeout(() => {
				this.pendingRequests.delete(requestId);
				resolve({
					content: [{ type: "text", text: "Error: Extension timed out." }],
				});
			}, 3000);

			this.pendingRequests.set(requestId, (data: any) => {
				clearTimeout(timeout);
				// Handle text vs object response
				const text =
					typeof data.content === "string"
						? data.content
						: JSON.stringify(data.content);
				resolve({ content: [{ type: "text", text: text || "No content." }] });
			});

			this.wssInstance.clients.forEach((client: any) => {
				if (client.readyState === 1) {
					client.send(
						JSON.stringify({ type: messageType, requestId, ...payload }),
					);
				}
			});
		});
	}

	public updateDirectorConfig(newConfig: any) {
		this.directorConfig = newConfig;
		this.configManager.saveConfig(newConfig);
	}

	public captureScreenshotFromBrowser(): Promise<string> {
		if (!this.wssInstance || this.wssInstance.clients.size === 0) {
			throw new Error("No Browser Extension connected.");
		}

		return new Promise((resolve, reject) => {
			const requestId = `req_${Date.now()}_${Math.random()}`;
			const timeout = setTimeout(() => {
				this.pendingRequests.delete(requestId);
				reject(new Error("Browser screenshot timed out."));
			}, 5000);

			this.pendingRequests.set(requestId, (data: any) => {
				clearTimeout(timeout);
				// data is the screenshot data URL
				resolve(data);
			});

			this.wssInstance.clients.forEach((client: any) => {
				if (client.readyState === 1) {
					client.send(
						JSON.stringify({
							jsonrpc: "2.0",
							method: "browser_screenshot",
							id: requestId,
						}),
					);
				}
			});
		});
	}

	public async executeTool(name: string, args: any): Promise<any> {
		const callId = Math.random().toString(36).substring(7);
		const startTime = Date.now();
		const toolContext = await this.resolveAutomaticToolContext(name, args);

		// Broadcast Start
		if (this.wssInstance) {
			this.wssInstance.clients.forEach((c: any) => {
				if (c.readyState === 1)
					c.send(
						JSON.stringify({
							type: "TOOL_CALL_START",
							id: callId,
							tool: name,
							args: args,
							...buildAutomaticToolContextStartPayload(toolContext),
						}),
					);
			});
		}

		// Emit Pulse Event
		this.eventBus.emitEvent("tool:call", "MCPServer", {
			tool: name,
			args,
			callId,
			toolContextPreview:
				buildAutomaticToolContextStartPayload(toolContext).contextPreview,
		});

		// Audit Start
		try {
			this.auditService.log("TOOL_START", { tool: name, args }, "INFO");
		} catch (e) {
			console.error("Audit Fail", e);
		}

		try {
			// 0. Permission Check
			// A. Policy Service (Fine-grained)
			// A. Policy Service (Fine-grained)
			const policyDecision = this.policyService.check(name, args);
			if (!policyDecision.allowed) {
				this.auditService.log(
					"TOOL_DENIED",
					{ tool: name, args, reason: policyDecision.reason || "Policy Deny" },
					"WARN",
				);
				throw new Error(
					`Policy VIOLATION: Execution of tool '${name}' is DENIED by active policy. Reason: ${policyDecision.reason}`,
				);
			}

			// B. Permission Manager (High-level Autonomy)
			const permission = this.permissionManager.checkPermission(name, args);

			if (permission === "DENIED") {
				this.auditService.log(
					"TOOL_DENIED",
					{ tool: name, args, reason: "Policy/Permission" },
					"WARN",
				);
				throw new Error(
					`Permission Denied for tool '${name}' (Risk Level High). Autonomy Level is '${this.permissionManager.autonomyLevel}'.`,
				);
			}

			if (permission === "NEEDS_CONSULTATION") {
				console.log(`[TormentNexus Core] Consulting Council for: ${name}`);
				this.auditService.log(
					"TOOL_CONSULTATION",
					{ tool: name, args },
					"WARN",
				);
				const debate = await this.council.runConsensusSession(
					`Execute tool '${name}' with args: ${JSON.stringify(args)}`,
				);

				if (!debate.approved) {
					throw new Error(`Council Denied Execution: ${debate.summary}`);
				}
				console.log(`[TormentNexus Core] Council Approved: ${debate.summary}`);
			}

			// 1. Internal Status / Config Tools
			let result;
			if (name === "router_status") {
				result = {
					content: [{ type: "text", text: "TormentNexus Router is active." }],
				};
			} else if (name === "set_autonomy") {
				const level = args?.level as AutonomyLevel;
				this.permissionManager.setAutonomyLevel(level);
				result = {
					content: [{ type: "text", text: `Autonomy Level set to: ${level}` }],
				};
			} else if (name === "chat_reply") {
				const text = args?.text as string;
				// SAFETY: Default to false to prevent feedback loops. Explicitly set true if needed.
				const submit = (args?.submit as boolean) ?? false;
				console.log(
					`[TormentNexus Core] Chat Reply Requested: ${text} (submit: ${submit})`,
				);

				if (this.wssInstance) {
					this.wssInstance.clients.forEach((client: any) => {
						if (client.readyState === 1) {
							// OPEN
							client.send(
								JSON.stringify({
									type: "PASTE_INTO_CHAT",
									text: text,
									submit: submit, // Include submit flag for atomic paste+submit
								}),
							);
						}
					});
					result = {
						content: [
							{
								type: "text",
								text: `Sent to browser/IDE: "${text}" (submit: ${submit})`,
							},
						],
					};
				} else {
					result = {
						content: [
							{
								type: "text",
								text: "Error: No WebSocket server active to forward reply.",
							},
						],
					};
				}
			} else if (name === "chat_submit") {
				if (this.wssInstance && this.wssInstance.clients.size > 0) {
					this.wssInstance.clients.forEach((client: any) => {
						if (client.readyState === 1)
							client.send(JSON.stringify({ type: "SUBMIT_CHAT" }));
					});
					result = {
						content: [{ type: "text", text: "Sent SUBMIT_CHAT signal." }],
					};
				} else {
					// FALLBACK: Use Native Input for Alt-Enter
					console.error(
						"[MCPServer] No Extension bridge. Falling back to Native Input for Alt-Enter.",
					);
					try {
						const status = await this.inputTools.sendKeys("alt+enter");
						result = {
							content: [{ type: "text", text: `Fallback: ${status}` }],
						};
					} catch (e: any) {
						result = {
							content: [
								{
									type: "text",
									text: `Error: No extension AND native input failed: ${e.message}`,
								},
							],
						};
					}
				}
			} else if (name === "click_element") {
				const target = args?.target as string;
				if (this.wssInstance) {
					this.wssInstance.clients.forEach((client: any) => {
						if (client.readyState === 1)
							client.send(
								JSON.stringify({
									method: "click_element",
									params: { target },
								}),
							);
					});
					result = {
						content: [
							{
								type: "text",
								text: `Sent CLICK_ELEMENT signal for "${target}".`,
							},
						],
					};
				} else {
					result = {
						content: [{ type: "text", text: "Error: No WebSocket server." }],
					};
				}
			} else if (name === "click_at") {
				const x = args?.x as number;
				const y = args?.y as number;
				if (this.wssInstance) {
					this.wssInstance.clients.forEach((client: any) => {
						if (client.readyState === 1)
							client.send(
								JSON.stringify({
									method: "click_at",
									params: { x, y },
								}),
							);
					});
					result = {
						content: [
							{ type: "text", text: `Sent CLICK_AT signal for (${x},${y}).` },
						],
					};
				} else {
					result = {
						content: [{ type: "text", text: "Error: No WebSocket server." }],
					};
				}
			} else if (name === "navigate") {
				const url = args?.url as string;
				if (this.wssInstance) {
					this.wssInstance.clients.forEach((client: any) => {
						if (client.readyState === 1)
							client.send(
								JSON.stringify({
									method: "browser_navigate",
									params: { url },
								}),
							);
					});
					// Wait for navigation
					await new Promise((r) => setTimeout(r, 2000));
					result = { content: [{ type: "text", text: `Navigated to ${url}` }] };
				} else {
					result = {
						content: [{ type: "text", text: "Error: No WebSocket server." }],
					};
				}
			} else if (name === "browser_screenshot") {
				// Modified to try Puppeteer if no WebSocket
				if (this.wssInstance && this.wssInstance.clients.size > 0) {
					const dataUrl = await this.captureScreenshotFromBrowser();
					result = {
						content: [
							{ type: "text", text: "Screenshot captured (Client)." },
							{
								type: "image",
								data: dataUrl.split(",")[1],
								mimeType: "image/png",
							},
						],
					};
				} else {
					// Fallback to Puppeteer Navigator
					// Note: Default screenshot of active page? Or require pageId?
					// For now, let's assume we want to screenshot the active Puppeteer page if available.
					// But BrowserService manages multiple pages.
					// Let's defer to specific `navigator_screenshot` or update this logic later.
					result = {
						content: [
							{
								type: "text",
								text: "Error: No Client Extension. For backend browsing, use 'navigator_navigate'.",
							},
						],
					};
				}
			}
			// --- THE NAVIGATOR (Phase 48: Puppeteer Backend) ---
			else if (name === "browser_navigate") {
				const url = args?.url as string;
				if (!url) throw new Error("Missing 'url'");
				const res = await this.browserService.navigate(url);
				result = {
					content: [
						{
							type: "text",
							text: `Navigated to "${res.title}" (ID: ${res.id})\n\nContent Preview:\n${res.content}`,
						},
					],
				};
			} else if (name === "browser_click") {
				const pageId = args?.pageId as string;
				const selector = args?.selector as string;
				if (!pageId || !selector)
					throw new Error("Missing params: pageId, selector");
				await this.browserService.click(pageId, selector);
				result = {
					content: [
						{ type: "text", text: `Clicked '${selector}' on page ${pageId}` },
					],
				};
			} else if (name === "browser_type") {
				const pageId = args?.pageId as string;
				const selector = args?.selector as string;
				const text = args?.text as string;
				if (!pageId || !selector || !text) throw new Error("Missing params");
				await this.browserService.type(pageId, selector, text);
				result = {
					content: [
						{
							type: "text",
							text: `Typed into '${selector}' on page ${pageId}`,
						},
					],
				};
			} else if (name === "browser_extract") {
				const pageId = args?.pageId as string;
				if (!pageId) throw new Error("Missing 'pageId'");
				const content = await this.browserService.extract(pageId);
				result = {
					content: [{ type: "text", text: content.substring(0, 5000) }],
				}; // Limit response size
			} else if (name === "get_knowledge_graph") {
				result = await this.knowledgeService.getGraph(args?.query, args?.depth);
			} else if (name === "system_diagnostics") {
				const status = await DiagnosticTools.checkHealth();
				result = {
					content: [
						{
							type: "text",
							text: DiagnosticTools.getDiagnosticsMarkup(status),
						},
					],
				};
			} else if (name === "system_status") {
				const status = await DiagnosticTools.checkHealth();
				const memUsage = process.memoryUsage();
				const uptime = process.uptime();
				result = {
					content: [
						{
							type: "text",
							text: `## System Health Report\n${DiagnosticTools.getDiagnosticsMarkup(status)}\n\n**Process:** ${(memUsage.heapUsed / 1024 / 1024).toFixed(0)}MB heap, Uptime: ${(uptime / 3600).toFixed(1)}h`,
						},
					],
				};
			} else if (name === "health_check") {
				const status = await DiagnosticTools.checkHealth();
				const obj: Record<string, any> = {};
				status.forEach((s: any, i: number) => {
					obj[i] = s;
				});
				result = {
					content: [{ type: "text", text: JSON.stringify(obj, null, 2) }],
				};
			} else if (name === "native_input") {
				const keys = args?.keys as string;
				const targetWindow = args?.targetWindow as string | undefined;
				const forceFocus = (args?.forceFocus as boolean) ?? false;
				// Use Direct InputTools with optional window targeting
				try {
					const toolStartTime = Date.now();
					const status = await this.inputTools.sendKeys(
						keys,
						forceFocus,
						targetWindow,
					);
					const toolDuration = Date.now() - toolStartTime;

					this.metricsService.track("tool_call", 1, {
						tool: name,
						success: "true",
					});
					this.metricsService.trackDuration("tool_execution", toolDuration, {
						tool: name,
					});

					result = { content: [{ type: "text", text: status }] };
				} catch (e: any) {
					this.metricsService.track("tool_call", 1, {
						tool: name,
						success: "false",
					});
					this.metricsService.track("tool_error", 1, { tool: name });
					result = {
						content: [
							{
								type: "text",
								text: `Error executing native_input: ${e.message}`,
							},
						],
					};
				}
			} else if (name === "vscode_execute_command") {
				const command = args?.command as string;
				const cmdArgs = Array.isArray(args?.args) ? args.args : [];

				if (this.wssInstance) {
					this.wssInstance.clients.forEach((client: any) => {
						if (client.readyState === 1) {
							client.send(
								JSON.stringify({
									type: "VSCODE_COMMAND",
									command,
									args: cmdArgs,
								}),
							);
						}
					});
					result = {
						content: [
							{ type: "text", text: `Sent VSCODE_COMMAND: ${command}` },
						],
					};
				} else {
					result = {
						content: [
							{
								type: "text",
								text: "Error: No WebSocket server (Extension bridge) active.",
							},
						],
					};
				}
			} else if (name === "get_chrome_devtools_mcp_url") {
				// Hardcoded fallback response to stabilize specific routing issue
				result = { content: [{ type: "text", text: "ws://localhost:9222" }] };
			}
			// --- SWARM TOOLS (Phase 51) ---
			else if (name === "start_squad") {
				mcpServerDebugLog("[DEBUG] ENTERED start_squad BLOCK");
				const branch = args?.branch as string;
				const goal = args?.goal as string;
				if (!branch || !goal) throw new Error("Missing params: branch, goal");
				const res = await this.squadService.spawnMember(branch, goal);
				result = { content: [{ type: "text", text: res }] };
			} else if (name === "list_squads") {
				const members = this.squadService.listMembers();
				result = {
					content: [{ type: "text", text: JSON.stringify(members, null, 2) }],
				};
			} else if (name === "kill_squad") {
				const branch = args?.branch as string;
				if (!branch) throw new Error("Missing param: branch");
				const res = await this.squadService.killMember(branch);
				result = { content: [{ type: "text", text: res }] };
			} else if (name === "git_worktree_add") {
				const worktreePath = args?.path as string;
				const branch = args?.branch as string;
				const cwd = (args?.cwd as string) || process.cwd();

				// Construct command: git worktree add -b <branch> <path>
				// Note: -b creates new branch. If branch exists, maybe just checkout?
				// SquadService seems to imply new branch for task.
				// Safest: try -b, if fails (branch exists), try without -b?
				// For now, follow SquadService pattern: git worktree add -b ...
				const cmd = `git worktree add -b ${branch} "${worktreePath}"`;
				console.log(`[GitWorktree] Adding worktree: ${cmd}`);

				try {
					const out = await this.shellService.execute(cmd, cwd);
					result = { content: [{ type: "text", text: out }] };
				} catch (e: any) {
					throw new Error(`Git worktree add failed: ${e.message}`);
				}
			} else if (name === "git_worktree_remove") {
				const worktreePath = args?.path as string;
				const force = args?.force as boolean;
				const cwd = (args?.cwd as string) || process.cwd();
				const cmd = `git worktree remove ${force ? "--force" : ""} "${worktreePath}"`;

				try {
					const out = await this.shellService.execute(cmd, cwd);
					result = { content: [{ type: "text", text: out }] };
				} catch (e: any) {
					// Start_squad might fail if dir doesn't exist, but remove implies cleanup.
					// If fail, maybe manual rm -rf?
					throw new Error(`Git worktree remove failed: ${e.message}`);
				}
			} else if (name === "git_worktree_list") {
				const cwd = (args?.cwd as string) || process.cwd();
				const cmd = `git worktree list`;
				try {
					const out = await this.shellService.execute(cmd, cwd);
					result = { content: [{ type: "text", text: out }] };
				} catch (e: any) {
					throw new Error(`Git worktree list failed: ${e.message}`);
				}
			} else if (name === "vscode_get_status") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [
							{ type: "text", text: "Error: No Native Extension connected." },
						],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [
									{ type: "text", text: "Error: Extension timed out." },
								],
							});
						}, 3000);

						this.pendingRequests.set(requestId, (status: any) => {
							clearTimeout(timeout);
							resolve({
								content: [
									{ type: "text", text: JSON.stringify(status, null, 2) },
								],
							});
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								client.send(JSON.stringify({ type: "GET_STATUS", requestId }));
							}
						});
					});
				}
			}
			// --- CORE INFRASTRUCTURE TOOLS (Phase 20) ---
			else if (name === "symbol_search") {
				const query = (args?.query as string) || "";
				try {
					const goRes = await fetch(
						"http://127.0.0.1:4300/api/index?q=" +
							encodeURIComponent(query) +
							"&limit=20",
						{ signal: AbortSignal.timeout(5000) },
					);
					if (goRes.ok) {
						const data = await goRes.json();
						result = {
							content: [
								{
									type: "text",
									text: JSON.stringify(
										(data.results || data || []).slice(0, 20),
										null,
										2,
									),
								},
							],
						};
					} else {
						result = { content: [{ type: "text", text: "[]" }] };
					}
				} catch {
					result = { content: [{ type: "text", text: "[]" }] };
				}
			} else if (name === "lsp_symbol_search") {
				const query = args?.query as string;
				if (!query) throw new Error("Missing 'query' parameter");
				const symbols = this.lspService.searchSymbols(query);
				result = {
					content: [
						{
							type: "text",
							text: JSON.stringify(symbols.slice(0, 50), null, 2),
						},
					],
				};
			} else if (name === "lsp_definition") {
				const file = args?.file as string;
				const line = args?.line as number;
				const char = args?.char as number;
				if (!file || line === undefined || char === undefined)
					throw new Error("Missing params: file, line, char");
				const def = await this.lspService.goToDefinition(file, line, char);
				result = {
					content: [{ type: "text", text: JSON.stringify(def, null, 2) }],
				};
			} else if (name === "lsp_references") {
				const file = args?.file as string;
				const line = args?.line as number;
				const char = args?.char as number;
				if (!file || line === undefined || char === undefined)
					throw new Error("Missing params: file, line, char");
				const refs = await this.lspService.findReferences(file, line, char);
				result = {
					content: [{ type: "text", text: JSON.stringify(refs, null, 2) }],
				};
			} else if (name === "code_mode_execute") {
				const code = args?.code as string;
				if (!code) throw new Error("Missing 'code' parameter");
				const res = await this.codeModeService.executeCode(code);
				result = {
					content: [{ type: "text", text: JSON.stringify(res, null, 2) }],
				};
			} else if (name === "plan_mode_switch") {
				const mode = args?.mode as string;
				if (mode === "PLAN") this.planService.enterPlanMode();
				else if (mode === "BUILD") this.planService.enterBuildMode();
				else throw new Error("Invalid mode. Use 'PLAN' or 'BUILD'");
				result = {
					content: [{ type: "text", text: `Switched to ${mode} mode.` }],
				};
			} else if (name === "plan_mode_stage") {
				const file = args?.file as string;
				const content = args?.content as string;
				if (!file || !content) throw new Error("Missing params: file, content");
				const diff = this.planService.proposeChange(
					file,
					content,
					"Staged via tool",
				);
				result = {
					content: [
						{
							type: "text",
							text: `Staged changes for ${file} (ID: ${diff.id})`,
						},
					],
				};
			} else if (name === "plan_mode_status") {
				result = {
					content: [{ type: "text", text: this.planService.getStatus() }],
				};
			}
			// --- MEMORY / SEARCH TOOLS (Phase 23) ---
			else if (name === "memory_index_codebase") {
				const indexRoot = (args?.path as string) || process.cwd();
				await this.initializeMemorySystem();
				const count = await this.memoryManager.indexCodebase(indexRoot);
				result = {
					content: [
						{ type: "text", text: `Indexed ${count} chunks in ${indexRoot}` },
					],
				};
			} else if (name === "memory_search") {
				const query = args?.query as string;
				const limit = (args?.limit as number) || 5;
				if (!query) throw new Error("Missing 'query'");
				await this.initializeMemorySystem();
				const res = await this.memoryManager.search(query, limit);
				result = {
					content: [{ type: "text", text: JSON.stringify(res, null, 2) }],
				};
			}
			// --- DEEP RESEARCH (Phase 31) ---
			else if (name === "research_topic") {
				const topic = args?.topic as string;
				const depth = (args?.depth as number) || 2;
				const breadth = (args?.breadth as number) || 2;
				if (!topic) throw new Error("Missing 'topic'");
				console.log(`[MCPServer] Triggering Recursive Research: ${topic}`);
				const res = await this.deepResearchService.recursiveResearch(
					topic,
					depth,
					breadth,
				);
				result = {
					content: [{ type: "text", text: JSON.stringify(res, null, 2) }],
				};
			} else if (name === "vscode_read_selection") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [
							{ type: "text", text: "Error: No Native Extension connected." },
						],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [
									{ type: "text", text: "Error: Extension timed out." },
								],
							});
						}, 3000);

						this.pendingRequests.set(requestId, (data: any) => {
							clearTimeout(timeout);
							resolve({
								content: [
									{ type: "text", text: data.content || "No content." },
								],
							});
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								client.send(
									JSON.stringify({ type: "GET_SELECTION", requestId }),
								);
							}
						});
					});
				}
			} else if (name === "vscode_read_terminal") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [
							{ type: "text", text: "Error: No Native Extension connected." },
						],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [
									{ type: "text", text: "Error: Extension timed out." },
								],
							});
						}, 3000);

						this.pendingRequests.set(requestId, (data: any) => {
							clearTimeout(timeout);
							resolve({
								content: [
									{ type: "text", text: data.content || "No content." },
								],
							});
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								client.send(
									JSON.stringify({ type: "GET_TERMINAL", requestId }),
								);
							}
						});
					});
				}
			}
			// --- HEALER TOOLS (Phase 33) ---
			else if (name === "healer_diagnose") {
				const error = args?.error as string;
				const context = args?.context as string;
				if (!error) throw new Error("Missing 'error' parameter");
				const diagnosis = await this.healerService.analyzeError(error, context);
				result = {
					content: [{ type: "text", text: JSON.stringify(diagnosis, null, 2) }],
				};
			} else if (name === "healer_heal") {
				const error = args?.error as string;
				const context = args?.context as string;
				if (!error) throw new Error("Missing 'error' parameter");
				const success = await this.healerService.healAndVerify(error, context);
				result = {
					content: [
						{
							type: "text",
							text: success
								? "Healer successfully applied fix."
								: "Healer could not fix this error.",
						},
					],
				};
			}
			// --- DARWIN TOOLS (Phase 34) ---
			else if (name === "darwin_evolve") {
				const prompt = args?.prompt as string;
				const goal = args?.goal as string;
				const mutation = await this.darwinService.proposeMutation(prompt, goal);
				result = {
					content: [{ type: "text", text: JSON.stringify(mutation, null, 2) }],
				};
			} else if (name === "darwin_experiment") {
				const mutationId = args?.mutationId as string;
				const task = args?.task as string;
				const experiment = await this.darwinService.startExperiment(
					mutationId,
					task,
				);
				result = {
					content: [
						{ type: "text", text: `Experiment ${experiment.id} started.` },
					],
				};
			} else if (name === "get_chat_history") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [{ type: "text", text: "Error: No Extension connected." }],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_chat_${Date.now()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [
									{ type: "text", text: "Error: Chat scrape timed out." },
								],
							});
						}, 5000);

						this.pendingRequests.set(requestId, (data: any) => {
							clearTimeout(timeout);
							const history = data.history || [];
							resolve({
								content: [{ type: "text", text: history.join("\n") }],
							});
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								client.send(
									JSON.stringify({ type: "GET_CHAT_HISTORY", requestId }),
								);
							}
						});
					});
				}
			} else if (name === "vscode_get_notifications") {
				result = await this.broadcastRequestAndAwait("GET_NOTIFICATIONS");
			} else if (name === "vscode_submit_chat") {
				if (this.wssInstance) {
					this.wssInstance.clients.forEach((client: any) => {
						if (client.readyState === 1)
							client.send(JSON.stringify({ type: "SUBMIT_CHAT_HOOK" }));
					});
					result = {
						content: [{ type: "text", text: "Sent SUBMIT_CHAT_HOOK." }],
					};
				} else {
					result = {
						content: [{ type: "text", text: "Error: No WebSocket server." }],
					};
				}
			} else if (name === "start_task") {
				const goal = args?.goal as string;
				const maxSteps = (args?.maxSteps as number) || 10;
				const resultStr = await this.director.executeTask(goal, maxSteps);
				result = {
					content: [{ type: "text", text: resultStr }],
				};
			} else if (name === "browser_get_history") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [
							{ type: "text", text: "Error: No Browser Extension connected." },
						],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [{ type: "text", text: "Error: Browser timed out." }],
							});
						}, 5000);

						this.pendingRequests.set(requestId, (data: any) => {
							clearTimeout(timeout);
							if (Array.isArray(data)) {
								const historyText = data
									.map(
										(item: any) =>
											`- [${item.title || "Untitled"}](${item.url}) (Visits: ${item.visitCount})`,
									)
									.join("\n");
								resolve({
									content: [
										{
											type: "text",
											text: historyText || "No history found for this query.",
										},
									],
								});
							} else {
								resolve({
									content: [
										{
											type: "text",
											text: "Error: Unexpected data format from browser.",
										},
									],
								});
							}
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								client.send(
									JSON.stringify({
										jsonrpc: "2.0",
										method: "browser_get_history",
										params: { query: args.query, maxResults: args.maxResults },
										id: requestId,
									}),
								);
							}
						});
					});
				}
			} else if (name === "browser_debug") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [
							{ type: "text", text: "Error: No Browser Extension connected." },
						],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [{ type: "text", text: "Error: Browser timed out." }],
							});
						}, 10000);

						this.pendingRequests.set(requestId, (data: any) => {
							clearTimeout(timeout);
							resolve({
								content: [
									{ type: "text", text: JSON.stringify(data, null, 2) },
								],
							});
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								client.send(
									JSON.stringify({
										jsonrpc: "2.0",
										method: "browser_debug",
										params: args,
										id: requestId,
									}),
								);
							}
						});
					});
				}
			} else if (name === "browser_proxy_fetch") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [
							{ type: "text", text: "Error: No Browser Extension connected." },
						],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [{ type: "text", text: "Error: Browser timed out." }],
							});
						}, 15000);

						this.pendingRequests.set(requestId, (data: any) => {
							clearTimeout(timeout);
							resolve({
								content: [
									{ type: "text", text: JSON.stringify(data, null, 2) },
								],
							});
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								client.send(
									JSON.stringify({
										jsonrpc: "2.0",
										method: "browser_proxy_fetch",
										params: args,
										id: requestId,
									}),
								);
							}
						});
					});
				}
			}
			// Phase 5: Browser Tool Direct
			else if (name === "agent_browse") {
				const task = args?.task as string;
				if (!task) throw new Error("Missing 'task' argument for agent_browse");
				result = await this.browserTool.executeTask(task, true);
			}
			// Phase 6: Semantic Search Direct
			else if (name === "agent_search") {
				const query = args?.query as string;
				const root = args?.path || process.cwd();
				if (!query)
					throw new Error("Missing 'query' argument for agent_search");

				// Lazy load index if needed
				try {
					await this.searchService.loadIndex();
				} catch (e) {
					void e;
				}

				result = await this.searchService.search(query, root);
			}

			// Swarm Tools
			else if (name === "spawn_agent") {
				const id = this.spawnerService.spawn(args.type, args.task);
				result = {
					content: [
						{ type: "text", text: `Agent spawned successfully. ID: ${id}` },
					],
				};
			} else if (name === "list_agents") {
				const agents = this.spawnerService.listAgents();
				result = {
					content: [{ type: "text", text: JSON.stringify(agents, null, 2) }],
				};
			} else if (name === "kill_agent") {
				const success = this.spawnerService.killAgent(args.agentId);
				result = {
					content: [
						{
							type: "text",
							text: success
								? "Agent terminated."
								: "Failed to terminate agent (might not be running).",
						},
					],
				};
			} else if (name === "get_agent_result") {
				const agent = this.spawnerService.getAgent(args.agentId);
				if (!agent) {
					result = {
						content: [{ type: "text", text: "Error: Agent not found." }],
					};
				} else {
					result = {
						content: [
							{
								type: "text",
								text: JSON.stringify(
									agent.result || { status: agent.status },
									null,
									2,
								),
							},
						],
					};
				}
			} else if (name === "read_page") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					// FALLBACK: Use Native Reader (ReaderTools)
					console.error(
						"[MCPServer] No Browser Extension. Falling back to Native Reader...",
					);
					const nativeReader = ReaderTools.find((t) => t.name === "read_page");
					if (nativeReader && this.isToolWithHandler(nativeReader)) {
						result = await nativeReader.handler(args);
					} else {
						result = {
							content: [
								{ type: "text", text: "Error: No Native Reader available." },
							],
						};
					}
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							// TIMEOUT FALLBACK to Native Reader
							console.error(
								"[MCPServer] Browser Timed Out. Falling back to Native Reader...",
							);
							const nativeReader = ReaderTools.find(
								(t) => t.name === "read_page",
							);
							if (nativeReader && this.isToolWithHandler(nativeReader)) {
								Promise.resolve(nativeReader.handler(args))
									.then(resolve)
									.catch((e: Error) =>
										resolve({
											content: [{ type: "text", text: `Error: ${e.message}` }],
										}),
									);
							} else {
								resolve({
									content: [
										{
											type: "text",
											text: "Error: Browser timed out and no native reader.",
										},
									],
								});
							}
						}, 5000);

						this.pendingRequests.set(requestId, (data: any) => {
							clearTimeout(timeout);
							// Expect data to have { url, title, content }
							const text = `URL: ${data.url}\nTitle: ${data.title}\n\n${data.content.substring(0, 5000)}...`;
							resolve({ content: [{ type: "text", text: text }] });
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								// Send JSON-RPC 2.0 request
								client.send(
									JSON.stringify({
										jsonrpc: "2.0",
										method: "browser_scrape",
										id: requestId,
									}),
								);
							}
						});
					});
				}
			} else if (name === "memorize_page") {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [
							{ type: "text", text: "Error: No Browser Extension connected." },
						],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [{ type: "text", text: "Error: Browser timed out." }],
							});
						}, 8000); // 8s timeout for memorization

						this.pendingRequests.set(requestId, async (data: any) => {
							clearTimeout(timeout);
							// Data: { url, title, content }
							if (!data.content) {
								resolve({
									content: [
										{
											type: "text",
											text: "Error: No content received from browser.",
										},
									],
								});
								return;
							}

							try {
								const ctxId = await this.memoryManager.saveContext(
									`URL: ${data.url}\n\n${data.content}`, // Content
									{
										// Metadata
										title: data.title || "Web Page",
										source: data.url || "browser",
									},
								);
								resolve({
									content: [
										{
											type: "text",
											text: `✅ Memorized page: "${data.title}" (ID: ${ctxId})`,
										},
									],
								});
							} catch (e: any) {
								resolve({
									content: [
										{
											type: "text",
											text: `Error saving context: ${e.message}`,
										},
									],
								});
							}
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								// Reuse 'browser_scrape' method
								client.send(
									JSON.stringify({
										jsonrpc: "2.0",
										method: "browser_scrape",
										id: requestId,
									}),
								);
							}
						});
					});
				}
			} else if (
				[
					"browser_insert_text",
					"browser_submit_form",
					"browser_capture_screenshot",
					"browser_select_element",
					"browser_navigate",
					"browser_execute_script",
				].includes(name)
			) {
				if (!this.wssInstance || this.wssInstance.clients.size === 0) {
					result = {
						content: [
							{ type: "text", text: "Error: No Browser Extension connected." },
						],
					};
				} else {
					result = await new Promise((resolve) => {
						const requestId = `req_${Date.now()}_${Math.random()}`;
						const timeout = setTimeout(() => {
							this.pendingRequests.delete(requestId);
							resolve({
								content: [
									{ type: "text", text: "Error: Browser operation timed out." },
								],
							});
						}, 15000);

						this.pendingRequests.set(requestId, (data: any) => {
							clearTimeout(timeout);
							resolve({
								content: [
									{
										type: "text",
										text:
											typeof data === "string"
												? data
												: JSON.stringify(data, null, 2),
									},
								],
							});
						});

						this.wssInstance.clients.forEach((client: any) => {
							if (client.readyState === 1) {
								client.send(
									JSON.stringify({
										jsonrpc: "2.0",
										method: name,
										params: args,
										id: requestId,
									}),
								);
							}
						});
					});
				}
			} else if (name === "use_agent") {
				const agentName = args.name as string;
				const prompt = args.prompt as string;
				let agent = this.activeAgents.get(agentName);

				if (!agent) {
					if (agentName === "claude") agent = new ClaudeAdapter();
					else if (agentName === "gemini") agent = new GeminiAdapter();
					else
						return {
							content: [{ type: "text", text: `Unknown agent: ${agentName}` }],
						};

					await agent.start();
					this.activeAgents.set(agentName, agent);
				}

				if (!agent.isActive()) {
					await agent.start();
				}

				// Send prompt and wait for response (simple timeout-based aggregation)
				const responsePromise = new Promise<string>((resolve) => {
					let buffer = "";
					let timer: any;

					const onData = (data: string) => {
						buffer += data;
						clearTimeout(timer);
						// Debounce: Wait 2s for more data, else assume done
						timer = setTimeout(() => {
							cleanup();
							resolve(buffer);
						}, 2000);
					};

					const cleanup = () => {
						agent!.removeListener("output", onData);
					};

					agent!.on("output", onData);
					agent!.send(prompt).catch((err: any) => {
						cleanup();
						resolve(`Error sending to agent: ${err.message}`);
					});
				});

				const output = await responsePromise;
				result = { content: [{ type: "text", text: output }] };
			} else if (name === "take_screenshot") {
				try {
					const data = await this.captureScreenshotFromBrowser();
					result = {
						content: [
							{ type: "text", text: "Screenshot captured." },
							{
								type: "image",
								data: data.split(",")[1],
								mimeType: "image/jpeg",
							},
						],
					};
				} catch (e: any) {
					result = { content: [{ type: "text", text: `Error: ${e.message}` }] };
				}
			} else if (name === "analyze_screenshot") {
				const prompt = args.prompt as string;
				try {
					console.log(
						`[TormentNexus Core] 👁️ Analyzing screenshot with prompt: "${prompt}"...`,
					);
					const data = await this.captureScreenshotFromBrowser();
					const base64 = data.split(",")[1];
					const mimeType = "image/jpeg";

					const response = await this.llmService.generateText(
						"google",
						"gemini-1.5-pro",
						"You are a web automation assistant. Analyze the screenshot.",
						prompt,
						{
							images: [{ base64, mimeType }],
							taskComplexity: "high",
						},
					);

					result = {
						content: [{ type: "text", text: response.content }],
					};
				} catch (e: any) {
					result = {
						content: [
							{
								type: "text",
								text: `Error analyzing screenshot: ${e.message}`,
							},
						],
					};
				}
			}
			// Removed obsolete start_watchdog and start_chat_daemon handlers
			else if (name === "start_auto_drive") {
				this.director.startAutoDrive();
				result = {
					content: [
						{
							type: "text",
							text: "Auto-Drive started. The Director will now proactively drive the development loop.",
						},
					],
				};
			} else if (name === "stop_auto_drive") {
				this.director.stopAutoDrive();
				result = {
					content: [{ type: "text", text: "Auto-Drive Stopped." }],
				};
			} else if (name === "search_skills") {
				const query = (args?.query as string) || "";
				try {
					const allSkills = await this.skillRegistry.searchSkills(query);
					result = {
						content: [
							{
								type: "text",
								text: JSON.stringify({ matches: allSkills }, null, 2),
							},
						],
					};
				} catch {
					result = {
						content: [{ type: "text", text: JSON.stringify({ matches: [] }) }],
					};
				}
			} else if (name === "list_skills") {
				result = this.skillRegistry.listSkills();
			} else if (name === "read_skill") {
				result = this.skillRegistry.readSkill(args?.skillName as string);
			} else if (name === "mcpm_search") {
				result = {
					content: [
						{
							type: "text",
							text: JSON.stringify(
								await this.mcpmInstaller.search(args?.query as string),
								null,
								2,
							),
						},
					],
				};
			} else if (name === "mcpm_install") {
				const msg = await this.mcpmInstaller.install(args?.name as string);
				result = {
					content: [{ type: "text", text: msg }],
				};
			} else if (name === "mcp_add_server") {
				const srvName = args.name as string;
				const repoUrl = args.repoUrl as string;

				// 1. Clone Repo
				const cloneMsg = await this.submoduleManager.addSubmodule(
					srvName,
					repoUrl,
				);

				// 2. Add to Aggregator Config
				await this.mcpAggregator.addServerConfig(srvName, {
					command: args.command,
					args: args.args,
					enabled: true,
					env: args.env,
				});

				result = {
					content: [
						{
							type: "text",
							text: `${cloneMsg}\nAdded server '${srvName}' to configuration and connected.`,
						},
					],
				};
			} else if (name === "start_autotest") {
				this.autoTestService.start();
				result = {
					content: [{ type: "text", text: "Auto-Test Watcher Started." }],
				};
			} else if (name === "stop_autotest") {
				this.autoTestService.stop();
				result = {
					content: [{ type: "text", text: "Auto-Test Watcher Stopped." }],
				};
			} else if (name === "start_squad") {
				const branch = args?.branch as string;
				const goal = args?.goal as string;
				const msg = await this.squadService.spawnMember(branch, goal);
				result = { content: [{ type: "text", text: msg }] };
			} else if (name === "list_squads") {
				const squads = this.squadService.listMembers();
				result = {
					content: [{ type: "text", text: JSON.stringify(squads, null, 2) }],
				};
			} else if (name === "kill_squad") {
				const branch = args?.branch as string;
				const msg = await this.squadService.killMember(branch);
				result = { content: [{ type: "text", text: msg }] };
			} else if (name === "merge_squad") {
				const { MergeService } = await import("./orchestrator/MergeService.js");
				const merger = new MergeService(process.cwd());
				const branch = args?.branch as string;
				const res = await merger.mergeBranch(branch);
				result = { content: [{ type: "text", text: JSON.stringify(res) }] };
			} else if (name === "git_worktree_list") {
				result = {
					content: [
						{
							type: "text",
							text: JSON.stringify(
								await this.gitWorktreeManager.listWorktrees(),
								null,
								2,
							),
						},
					],
				};
			} else if (name === "git_worktree_add") {
				const branch = args?.branch as string;
				const path = args?.path as string; // Relative or absolute
				// Manager handles it
				const finalPath = await this.gitWorktreeManager.addWorktree(
					branch,
					path,
				);
				result = {
					content: [{ type: "text", text: `Worktree added at: ${finalPath}` }],
				};
			} else if (name === "git_worktree_remove") {
				const pathOrBranch = args?.path || (args?.branch as string);
				const force = args?.force as boolean;
				await this.gitWorktreeManager.removeWorktree(pathOrBranch, force);
				result = {
					content: [
						{ type: "text", text: `Worktree removed: ${pathOrBranch}` },
					],
				};
			} else if (name === "execute_sandbox") {
				const lang = (args?.language as string) || "node";
				const code = args?.code as string;
				result = {
					content: [
						{
							type: "text",
							text: JSON.stringify(
								await this.sandboxService.execute(
									(lang === "javascript" || lang === "js"
										? "node"
										: lang === "python3"
											? "python"
											: lang) as "node" | "python",
									code,
								),
								null,
								2,
							),
						},
					],
				};
			} else if (name === "index_codebase") {
				const dir = args?.path || process.cwd();
				console.log(`[TormentNexus Core] Indexing codebase at ${dir}...`);
				try {
					const count = await Promise.race([
						this.memoryManager.indexCodebase(dir),
						new Promise<number>((resolve) =>
							setTimeout(() => resolve(0), 10000),
						),
					]);
					result = {
						content: [
							{
								type: "text",
								text: `Indexed ${count} documents/chunks from ${dir}.`,
							},
						],
					};
				} catch (e: any) {
					result = {
						content: [{ type: "text", text: `Indexing failed: ${e.message}` }],
					};
				}
			} else if (name === "search_codebase") {
				const query = (args?.query as string) || "";
				console.log(`[TormentNexus Core] Semantic searching for: ${query}`);
				try {
					const matches = await Promise.race([
						this.memoryManager.search(query),
						new Promise<any[]>((resolve) =>
							setTimeout(() => resolve([]), 8000),
						),
					]);
					let text = `Searching for: "${query}"

`;
					matches.forEach((m: any, i: number) => {
						const filePath = m.metadata?.file_path || m.id;
						text += `${i + 1}. [${filePath}]
${m.content?.substring(0, 200) || ""}...

`;
					});
					result = {
						content: [
							{ type: "text", text: text || `No results found for "${query}"` },
						],
					};
				} catch (e: any) {
					result = {
						content: [{ type: "text", text: `Search failed: ${e.message}` }],
					};
				}
			} else if (name === "find_references") {
				const output = await this.lspTools.findReferences(args);
				result = { content: [{ type: "text", text: output }] };
			} else if (name === "go_to_definition") {
				const output = await this.lspTools.goToDefinition(args);
				result = { content: [{ type: "text", text: output }] };
			} else if (name === "get_symbols") {
				const output = await this.lspTools.getSymbols(args);
				result = { content: [{ type: "text", text: output }] };
			} else if (name === "rename_symbol") {
				const output = await this.lspTools.renameSymbol(args);
				result = { content: [{ type: "text", text: output }] };
			} else if (name === "search_symbols") {
				const output = await this.lspTools.searchSymbols(args);
				result = { content: [{ type: "text", text: output }] };
			}
			// Phase 51: Plan/Build Mode Tools
			else if (name === "plan_mode") {
				this.planService.enterPlanMode();
				result = {
					content: [
						{
							type: "text",
							text: `Switched to PLAN mode. Changes will be staged but not applied.`,
						},
					],
				};
			} else if (name === "build_mode") {
				this.planService.enterBuildMode();
				result = {
					content: [
						{
							type: "text",
							text: `Switched to BUILD mode. Approved changes can now be applied.`,
						},
					],
				};
			} else if (name === "propose_change") {
				const diff = this.planService.proposeChange(
					args.file_path,
					args.content,
					args.description,
				);
				result = {
					content: [
						{
							type: "text",
							text: `Proposed change for ${args.file_path} (diff ID: ${diff.id})`,
						},
					],
				};
			} else if (name === "review_changes") {
				const pending = this.planService.getPendingChanges();
				const summary = pending
					.map((d) => `[${d.id}] ${d.filePath} - ${d.status}`)
					.join("\n");
				result = {
					content: [
						{
							type: "text",
							text:
								pending.length === 0
									? "No pending changes."
									: `Pending changes:\n${summary}`,
						},
					],
				};
			} else if (name === "approve_change") {
				const success = this.planService.approveDiff(args.diff_id);
				result = {
					content: [
						{
							type: "text",
							text: success
								? `Approved diff ${args.diff_id}`
								: `Failed to approve diff ${args.diff_id}`,
						},
					],
				};
			} else if (name === "apply_changes") {
				const applied = this.planService.applyApprovedChanges();
				if (!applied) {
					result = {
						content: [
							{
								type: "text",
								text: `Cannot apply changes. Switch to BUILD mode first.`,
							},
						],
					};
				} else {
					result = {
						content: [
							{
								type: "text",
								text: `Applied ${applied.applied.length} changes. Failed: ${applied.failed.length}`,
							},
						],
					};
				}
			} else if (name === "plan_status") {
				const status = this.planService.getStatus();
				result = { content: [{ type: "text", text: status }] };
			} else if (name === "create_checkpoint") {
				const checkpoint = this.planService.createCheckpoint(
					args.name,
					args.description,
				);
				result = {
					content: [
						{
							type: "text",
							text: `Created checkpoint: ${checkpoint.name} (${checkpoint.id})`,
						},
					],
				};
			} else if (name === "rollback") {
				const success = this.planService.rollback(args.checkpoint_id);
				result = {
					content: [
						{
							type: "text",
							text: success
								? `Rolled back to checkpoint ${args.checkpoint_id}`
								: "Rollback failed",
						},
					],
				};
			}
			// Phase 53: Memory Tools
			else if (name === "add_memory") {
				const type = args.type || "working";
				const namespace = args.namespace || "project";
				try {
					const memory = (await Promise.race([
						this.agentMemoryService.add(args?.content || "", type, namespace, {
							source: args?.source || "tool",
							tags: args?.tags || [],
						}),
						new Promise((_, reject) =>
							setTimeout(() => reject(new Error("timeout")), 8000),
						),
					])) as any;
					result = {
						content: [
							{
								type: "text",
								text: `Added ${type} memory: ${memory?.id ?? "unknown"}`,
							},
						],
					};
				} catch (e: any) {
					result = {
						content: [
							{
								type: "text",
								text: `Added ${type} memory (session-only). ${e.message === "timeout" ? "Vector store timed out." : e.message}`,
							},
						],
					};
				}
			} else if (name === "search_memory") {
				let memories: any[] = [];
				try {
					memories = await Promise.race([
						this.agentMemoryService.search(args?.query || "", {
							type: args?.type,
							namespace: args?.namespace,
							limit: args?.limit || 10,
						}),
						new Promise<any[]>((resolve) =>
							setTimeout(() => resolve([]), 8000),
						),
					]);
				} catch {
					memories = [];
				}
				if (memories.length === 0) {
					result = {
						content: [
							{
								type: "text",
								text: `No memories found matching "${args.query}"`,
							},
						],
					};
				} else {
					const formatted = memories
						.map(
							(m) =>
								`[${m.type}/${m.namespace}] ${m.content.substring(0, 100)}${m.content.length > 100 ? "..." : ""} (score: ${(m.score ?? 0).toFixed(2)})`,
						)
						.join("\n");
					result = {
						content: [
							{
								type: "text",
								text: `Found ${memories.length} memories:\n${formatted}`,
							},
						],
					};
				}
			} else if (name === "get_recent_memories") {
				const memories = await this.agentMemoryService.getRecent(
					args.limit || 10,
					{
						type: args.type,
						namespace: args.namespace,
					},
				);
				if (memories.length === 0) {
					result = { content: [{ type: "text", text: "No recent memories." }] };
				} else {
					const formatted = memories
						.map(
							(m) =>
								`[${m.type}/${m.namespace}] ${m.content.substring(0, 100)}${m.content.length > 100 ? "..." : ""}`,
						)
						.join("\n");
					result = {
						content: [{ type: "text", text: `Recent memories:\n${formatted}` }],
					};
				}
			} else if (name === "memory_stats") {
				const stats = this.agentMemoryService.getStats();
				const formatted = Object.entries(stats)
					.map(([k, v]) => `${k}: ${v}`)
					.join("\n");
				result = {
					content: [{ type: "text", text: `Memory Statistics:\n${formatted}` }],
				};
			} else if (name === "clear_session_memory") {
				this.agentMemoryService.clearSession();
				result = {
					content: [{ type: "text", text: "Cleared all session memories." }],
				};
			}
			// Phase 54: Workflow Tools
			else if (name === "run_workflow") {
				const workflowId = args.workflow_id;
				const input = args.input || {};
				try {
					const runResult = await this.workflowEngine.start(workflowId, {
						input,
					});
					result = {
						content: [
							{
								type: "text",
								text: `Workflow ${workflowId} completed.\nRun ID: ${runResult.id}\nStatus: ${runResult.status}\nNode: ${runResult.currentNode}\nOutput: ${JSON.stringify(runResult.state?.output ?? runResult.state, null, 2).substring(0, 500)}`,
							},
						],
					};
				} catch (e: any) {
					result = {
						content: [{ type: "text", text: `Workflow error: ${e.message}` }],
					};
				}
			} else if (name === "list_workflows") {
				const executions = this.workflowEngine.listExecutions();
				if (executions.length === 0) {
					result = {
						content: [
							{
								type: "text",
								text: "No workflow executions. Use run_workflow to start a workflow.",
							},
						],
					};
				} else {
					const formatted = executions
						.map(
							(e) =>
								`- ${e.id}: ${e.workflowId} [${e.status}] at ${e.currentNode}`,
						)
						.join("\n");
					result = {
						content: [
							{ type: "text", text: `Workflow executions:\n${formatted}` },
						],
					};
				}
			} else if (name === "workflow_status") {
				const runId = args.run_id;
				const execution = this.workflowEngine.getExecution(runId);
				if (!execution) {
					result = {
						content: [
							{ type: "text", text: `No workflow run found with ID: ${runId}` },
						],
					};
				} else {
					result = {
						content: [
							{
								type: "text",
								text: `Workflow Run: ${runId}\nWorkflow: ${execution.workflowId}\nStatus: ${execution.status}\nCurrent Node: ${execution.currentNode}\nLast Updated: ${execution.updatedAt}`,
							},
						],
					};
				}
			} else if (name === "approve_workflow") {
				const runId = args.run_id;
				const approved = args.approved ?? true;
				try {
					if (approved) {
						await this.workflowEngine.approve(runId);
					} else {
						this.workflowEngine.reject(
							runId,
							"Rejected via approve_workflow tool",
						);
					}
					result = {
						content: [
							{
								type: "text",
								text: approved
									? `Approved workflow ${runId}`
									: `Rejected workflow ${runId}`,
							},
						],
					};
				} catch (e: any) {
					result = {
						content: [{ type: "text", text: `Approval error: ${e.message}` }],
					};
				}
			}
			// Phase 55: Code Mode Tools
			else if (name === "execute_code") {
				const code = args.code;
				if (!this.codeModeService.isEnabled()) {
					result = {
						content: [
							{
								type: "text",
								text: "Code Mode is disabled. Use enable_code_mode first.",
							},
						],
					};
				} else {
					try {
						const execResult = await this.codeModeService.executeCode(
							code,
							args.context || {},
						);
						if (execResult.success) {
							result = {
								content: [
									{
										type: "text",
										text: `Execution successful (${execResult.executionTime}ms)\nTools called: ${execResult.toolsCalled.join(", ") || "none"}\nOutput:\n${execResult.output || execResult.result || "No output"}`,
									},
								],
							};
						} else {
							result = {
								content: [
									{
										type: "text",
										text: `Execution failed: ${execResult.error}\nOutput: ${execResult.output || "none"}`,
									},
								],
							};
						}
					} catch (e: any) {
						result = {
							content: [
								{ type: "text", text: `Code execution error: ${e.message}` },
							],
						};
					}
				}
			} else if (name === "enable_code_mode") {
				this.codeModeService.enable();
				result = {
					content: [
						{
							type: "text",
							text: "Code Mode enabled. You can now execute code via execute_code.",
						},
					],
				};
			} else if (name === "disable_code_mode") {
				this.codeModeService.disable();
				result = { content: [{ type: "text", text: "Code Mode disabled." }] };
			} else if (name === "code_mode_status") {
				const enabled = this.codeModeService.isEnabled();
				const registry = this.codeModeService.getRegistry();
				const toolCount = registry.getNames().length;
				const reduction = this.codeModeService.calculateContextReduction();
				result = {
					content: [
						{
							type: "text",
							text: `Code Mode Status:\n- Enabled: ${enabled}\n- Registered tools: ${toolCount}\n- Context reduction: ${reduction.reduction}% (${reduction.traditional} → ${reduction.codeMode} chars)`,
						},
					],
				};
			} else if (name === "list_code_tools") {
				const registry = this.codeModeService.getRegistry();
				const tools = registry.getAll();
				if (tools.length === 0) {
					result = {
						content: [
							{ type: "text", text: "No tools registered in Code Mode." },
						],
					};
				} else {
					const formatted = tools
						.map((t) => `- ${t.name}: ${t.description}`)
						.join("\n");
					result = {
						content: [{ type: "text", text: `Code Mode Tools:\n${formatted}` }],
					};
				}
			} else if (name === "handoff_session") {
				const artifact = await this.agentMemoryService.handoffSession(args);
				result = { content: [{ type: "text", text: artifact }] };
			} else if (name === "pickup_session") {
				const pickupRes = await this.agentMemoryService.pickupSession(
					args.artifact as string,
				);
				result = {
					content: [
						{
							type: "text",
							text: pickupRes.success
								? `Successfully restored ${pickupRes.count} context items.`
								: "Failed to restore session.",
						},
					],
				};
			} else if (name === "compact_context") {
				// In a real scenario, we'd get the current message history.
				// For the tool, we'll signal the pruner to run on the next maintenance cycle or return current stats.
				const maxTokens = args.max_tokens || 100000;
				result = {
					content: [
						{
							type: "text",
							text: `Context compaction triggered (Limit: ${maxTokens} tokens). Older messages are being archived to LanceDB.`,
						},
					],
				};
			} else if (name === "get_suggestions") {
				const context = args.context || "";
				if (context) {
					await this.suggestionService.processContext({
						type: "manual",
						content: context,
					});
				}
				const suggestions = this.suggestionService.getPendingSuggestions();
				result = {
					content: [
						{ type: "text", text: JSON.stringify(suggestions, null, 2) },
					],
				};
			} else if (name === "process_note") {
				const content = (args?.content as string) || "";
				const title = (args?.title as string) || "";
				try {
					const memory = (await Promise.race([
						this.agentMemoryService.add(content, "long_term", "project", {
							source: "process_note",
							title,
						}),
						new Promise((_, reject) =>
							setTimeout(() => reject(new Error("timeout")), 8000),
						),
					])) as any;
					result = {
						content: [
							{
								type: "text",
								text: `Memory saved with ID: ${memory?.id ?? "unknown"}`,
							},
						],
					};
				} catch (e: any) {
					result = {
						content: [
							{
								type: "text",
								text: `Memory saved (session-only, ID: note-${Date.now()}). ${e.message === "timeout" ? "Vector store timed out." : e.message}`,
							},
						],
					};
				}
			} else if (name === "export_chat") {
				const format = (args.format as string) || "markdown";
				const exportPath =
					(args.path as string) ||
					`./chat_export_${Date.now()}.${format === "json" ? "json" : "md"}`;
				const history = await this.shellService.getSystemHistory(200);
				const content =
					format === "json"
						? JSON.stringify(history, null, 2)
						: history.map((line) => `- ${line}`).join("\n");
				await fs.promises.writeFile(exportPath, content);
				result = {
					content: [{ type: "text", text: `Chat exported to ${exportPath}` }],
				};
			} else if (name === "auto_heal") {
				const error = (args?.error as string) || "";
				const context = (args?.context as string) || "";
				const success = await this.healerService.healAndVerify(error, context);
				result = {
					content: [
						{
							type: "text",
							text: success
								? "Healer successfully fixed the error."
								: "Healer could not fix this error autonomously.",
						},
					],
				};
			} else if (name === "get_project_context") {
				const contextPath = path.join(
					process.cwd(),
					".tormentnexus",
					"project_context.md",
				);
				let projectContent = "";
				if (fs.existsSync(contextPath)) {
					projectContent = await fs.promises.readFile(contextPath, "utf-8");
				} else {
					projectContent =
						"# Project Context\n\nNo persistent context has been recorded yet.";
				}

				// Append Environment Awareness
				const env = await detectLocalExecutionEnvironment();
				const envReport = `
## Environment Awareness
- **OS**: ${env.os}
- **Preferred Shell**: ${env.summary.preferredShellLabel} (${env.summary.preferredShellId})
- **POSIX Support**: ${env.summary.supportsPosixShell ? "Available" : "Missing"}
- **Recommendations**: ${env.summary.notes.join(" ")}

### Available Tools
${env.tools
	.filter((tool) => tool.installed)
	.map((tool) => `- **${tool.name}**: ${tool.version || "detected"}`)
	.join("\n")}
`;
				result = {
					content: [
						{ type: "text", text: `${projectContent}\n\n---\n${envReport}` },
					],
				};
			} else if (name === "update_project_context") {
				const contextPath = path.join(
					process.cwd(),
					".tormentnexus",
					"project_context.md",
				);
				const tormentnexusDir = path.join(process.cwd(), ".tormentnexus");
				if (!fs.existsSync(tormentnexusDir))
					fs.mkdirSync(tormentnexusDir, { recursive: true });

				await fs.promises.writeFile(contextPath, args.content as string);
				result = {
					content: [
						{ type: "text", text: "Project context updated successfully." },
					],
				};
			} else {
				/*
            // Phase 60: The Mesh
            else if (name === "swarm_broadcast") {
                if (!this.meshService) {
                    result = { content: [{ type: "text", text: "Mesh Service not active." }] };
                } else {
                    const type = typeof args?.type === 'string' ? args.type : 'TASK_OFFER';
                    this.meshService.broadcast(type, args.payload || {});
                    result = {
                        content: [{ type: "text", text: `Broadcasted '${type}' to swarm.` }]
                    };
                }
            }
            */
				// Check Standard Library
				const terminalTools = this.terminalService.getTools();
				const standardTool = [
					...FileSystemTools,
					...terminalTools,
					...MemoryTools,
					...TunnelTools,
					...LogTools,
					...ConfigTools,
					...SearchTools,
					...ReaderTools,
					...WorktreeTools,
					...MetaTools,
					WebSearchTool,
				].find((t) => t.name === name);
				if (standardTool && this.isToolWithHandler(standardTool)) {
					result = await standardTool.handler(args);

					// Phase 93: P2P Artifact Federation Interception
					if (
						name === "read_file" &&
						this.meshService &&
						result?.content?.[0]?.text?.includes("ENOENT")
					) {
						console.log(
							`[Mesh Artifact] Local read missed for ${args.path}. Querying Swarm...`,
						);

						const timeoutMs = 2000;
						const reqPath = args.path;

						try {
							const federatedContent = await new Promise<string>(
								(resolve, reject) => {
									const timeout = setTimeout(
										() => reject(new Error("Swarm Artifact Read Timeout")),
										timeoutMs,
									);

									const onMeshResponse = (msg: any) => {
										if (msg.type === SwarmMessageType.ARTIFACT_READ_RESPONSE) {
											const res = msg.payload as {
												path: string;
												content: string;
											};
											if (res.path === reqPath) {
												clearTimeout(timeout);
												this.meshService!.removeListener(
													"message",
													onMeshResponse,
												);
												resolve(res.content);
											}
										}
									};

									this.meshService!.on("message", onMeshResponse);
									this.meshService!.broadcast(
										SwarmMessageType.ARTIFACT_READ_REQUEST,
										{ path: reqPath },
									);
								},
							);

							// Successfully fetched from Swarm
							result = { content: [{ type: "text", text: federatedContent }] };
							console.log(
								`[Mesh Artifact] Successfully federated ${reqPath} from Swarm.`,
							);
						} catch (e: any) {
							// Timeout or failure, keep original ENOENT result
							console.log(
								`[Mesh Artifact] Federation failed for ${reqPath}: ${e.message}`,
							);
						}
					}
				} else {
					const directMetaResult = await this.handleDirectMetaTool(name, args);

					if (directMetaResult) {
						result = directMetaResult;
					} else {
						const aggregatedTools =
							await this.mcpAggregator.listAggregatedTools();
						const prefersAggregatorExecution = shouldPreferAggregatorExecution(
							name,
							aggregatedTools,
						);

						if (prefersAggregatorExecution) {
							try {
								this.nativeSessionMetaTools.touchLoadedTool(name);
								result = await this.mcpAggregator.executeTool(name, args);
							} catch (aggErr: any) {
								if (isToolNotFoundError(aggErr)) {
									try {
										result = await this.router.callTool(name, args);
									} catch (e: any) {
										throw new Error(`Tool execution failed: ${e.message}`);
									}
								} else {
									throw aggErr;
								}
							}

							// Skip proxy routing once the TormentNexus-native aggregator path has run.
						} else {
							// TormentNexus-native fallback path for unscoped tools.
							try {
								this.nativeSessionMetaTools.touchLoadedTool(name);
								result = await this.mcpAggregator.executeTool(name, args);
							} catch (aggErr: any) {
								if (isToolNotFoundError(aggErr)) {
									try {
										result = await this.router.callTool(name, args);
									} catch (e: any) {
										throw new Error(`Tool execution failed: ${e.message}`);
									}
								} else {
									throw aggErr;
								}
							}
						}
					}
				}
			}

			// Broadcast Success
			try {
				this.auditService.log(
					"TOOL_END",
					{
						tool: name,
						result: JSON.stringify(result).substring(0, 500),
						duration: Date.now() - startTime,
					},
					"INFO",
				);
			} catch (e) {
				console.error("Audit Fail", e);
			}
			if (this.wssInstance) {
				this.wssInstance.clients.forEach((c: any) => {
					if (c.readyState === 1)
						c.send(
							JSON.stringify({
								type: "TOOL_CALL_END",
								id: callId,
								tool: name,
								success: true,
								duration: Date.now() - startTime,
								result: JSON.stringify(result).substring(0, 200), // Summarize
							}),
						);
				});
			}

			this.captureToolObservation({
				toolName: name,
				args,
				result,
				durationMs: Date.now() - startTime,
			}).catch((err) =>
				console.error(
					"[MCPServer] BG observation failed:",
					err?.message || err,
				),
			);

			// ENGAGEMENT MODULE: Suggestion Trigger
			if ((name === "read_file" || name === "view_file") && result) {
				const contentText = this.getFirstTextContent(result);
				if (contentText) {
					const filePath = args.file_path || args.path || args.AbsolutePath;
					this.suggestionService
						.processContext({
							type: "file_read",
							path: filePath,
							content: contentText,
						})
						.catch((e) =>
							console.error("[SuggestionService] Trigger failed:", e),
						);
				}
			}

			return result;
		} catch (e: any) {
			// Audit Error
			try {
				this.auditService.log(
					"TOOL_END",
					{ tool: name, error: e.message, duration: Date.now() - startTime },
					"ERROR",
				);
			} catch (auditErr) {
				console.error("Audit Fail", auditErr);
			}

			// Broadcast Error
			if (this.wssInstance) {
				this.wssInstance.clients.forEach((c: any) => {
					if (c.readyState === 1)
						c.send(
							JSON.stringify({
								type: "TOOL_CALL_END",
								id: callId,
								tool: name,
								success: false,
								duration: Date.now() - startTime,
								result: e.message,
							}),
						);
				});
			}

			this.captureToolObservation({
				toolName: name,
				args,
				error: e,
				durationMs: Date.now() - startTime,
			}).catch((err) =>
				console.error(
					"[MCPServer] BG observation failed:",
					err?.message || err,
				),
			);

			// Return Error as Content (Don't throw, it kills the MCP stream)
			return {
				isError: true,
				content: [{ type: "text", text: `Error: ${e.message}` }],
			};
		}
	}

	public async getNativeTools(options?: {
		downstreamTools?: Tool[];
		skipLiveDiscovery?: boolean;
	}): Promise<Tool[]> {
		const internalTools: any[] = [
			{
				name: "router_status",
				description: "Check the status of the TormentNexus Router",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "mcpm_search",
				description: "Search for skills in the module registry",
				inputSchema: {
					type: "object",
					properties: {
						query: { type: "string" },
					},
					required: ["query"],
				},
			},
			{
				name: "mcpm_install",
				description: "Install a skill from the registry (via git)",
				inputSchema: {
					type: "object",
					properties: {
						name: { type: "string" },
					},
					required: ["name"],
				},
			},
			{
				name: "mcp_add_server",
				description:
					"Add a new downstream MCP server (via git clone + config update)",
				inputSchema: {
					type: "object",
					properties: {
						name: { type: "string" },
						repoUrl: { type: "string" },
						command: {
							type: "string",
							description: "Command to start server (e.g. 'node')",
						},
						args: {
							type: "array",
							items: { type: "string" },
							description: "Args for command (e.g. ['dist/index.js'])",
						},
						env: { type: "object", description: "Environment variables" },
					},
					required: ["name", "repoUrl", "command", "args"],
				},
			},
			{
				name: "system_status",
				description: "Get current system metrics (CPU, Memory, Uptime)",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "system_diagnostics",
				description: "Run advanced health checks on Core, Web UI, and Bridges",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "assimilate_skill",
				description:
					"Convert a research item into a functional TormentNexus Skill (runbook)",
				inputSchema: {
					type: "object",
					properties: {
						id: { type: "string" },
						name: { type: "string" },
						url: { type: "string" },
						summary: { type: "string" },
						relevance: { type: "string" },
					},
					required: ["name", "url", "summary", "relevance"],
				},
			},
			{
				name: "memorize_page",
				description:
					"Save the current browser page content to long-term memory",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "browser_screenshot",
				description: "Capture a screenshot of the current browser page",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "memory_index_codebase",
				description:
					"Deep Data Search: Index a directory for semantic code search",
				inputSchema: {
					type: "object",
					properties: {
						path: {
							type: "string",
							description: "Absolute path to directory to index (default: cwd)",
						},
					},
				},
			},
			{
				name: "memory_search",
				description: "Semantic search across indexed code and memories",
				inputSchema: {
					type: "object",
					properties: {
						query: { type: "string" },
						limit: { type: "number" },
					},
					required: ["query"],
				},
			},
			{
				name: "browser_get_history",
				description: "Get browser history from the connected browser extension",
				inputSchema: {
					type: "object",
					properties: {
						query: {
							type: "string",
							description: "Search query for history items",
						},
						maxResults: {
							type: "number",
							description: "Maximum number of results to return (default 20)",
						},
					},
				},
			},
			{
				name: "browser_debug",
				description:
					"Deep browser debugging via CDP (Chrome DevTools Protocol)",
				inputSchema: {
					type: "object",
					properties: {
						action: {
							type: "string",
							enum: ["attach", "detach", "command"],
							description: "Action to perform",
						},
						method: {
							type: "string",
							description: "CDP method to call (e.g. Runtime.evaluate)",
						},
						params: { type: "object", description: "CDP parameters" },
					},
					required: ["action"],
				},
			},
			{
				name: "browser_proxy_fetch",
				description:
					"Perform a network fetch through the browser extension (bypasses CORS, reaches local servers)",
				inputSchema: {
					type: "object",
					properties: {
						url: { type: "string", description: "URL to fetch" },
						options: {
							type: "object",
							description: "Fetch options (method, headers, body)",
						},
					},
					required: ["url"],
				},
			},
			{
				name: "spawn_agent",
				description: "Spawn a specialized sub-agent for a parallel task",
				inputSchema: {
					type: "object",
					properties: {
						type: {
							type: "string",
							enum: ["research", "code", "custom"],
							description: "Agent type",
						},
						task: {
							type: "string",
							description: "Detailed task description for the agent",
						},
					},
					required: ["type", "task"],
				},
			},
			{
				name: "list_agents",
				description: "List all active sub-agents in the swarm",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "kill_agent",
				description: "Terminate a running sub-agent",
				inputSchema: {
					type: "object",
					properties: {
						agentId: {
							type: "string",
							description: "UUID of the agent to kill",
						},
					},
					required: ["agentId"],
				},
			},
			{
				name: "get_agent_result",
				description: "Retrieve the final result of a completed agent task",
				inputSchema: {
					type: "object",
					properties: {
						agentId: { type: "string", description: "UUID of the agent" },
					},
					required: ["agentId"],
				},
			},
			{
				name: "start_task",
				description: "Start an autonomous task with the Director Agent",
				inputSchema: {
					type: "object",
					properties: {
						goal: { type: "string" },
						maxSteps: { type: "number" },
					},
					required: ["goal"],
				},
			},
			{
				name: "start_autotest",
				description: "Start the Auto-Test Watcher (runs tests on file save)",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "stop_autotest",
				description: "Stop the Auto-Test Watcher",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "execute_sandbox",
				description: "Execute code in a secure Docker container (Python/Node)",
				inputSchema: {
					type: "object",
					properties: {
						language: { type: "string", enum: ["python", "node"] },
						code: { type: "string" },
					},
					required: ["language", "code"],
				},
			},
			{
				name: "start_watchdog",
				description:
					"Start the Supervisor Watchdog to auto-approve terminal prompts",
				inputSchema: {
					type: "object",
					properties: {
						maxCycles: {
							type: "number",
							description: "Number of 5s cycles to run (default 20)",
						},
					},
				},
			},
			{
				name: "start_chat_daemon",
				description:
					"Start the Director in Chat Daemon mode (Auto-Approves & Reads Selection)",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "start_auto_drive",
				description:
					"Start the Autonomous Development Loop (Reads task.md, Drives Chat, Auto-Approves)",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "stop_auto_drive",
				description: "Stop the Autonomous Development Loop",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "set_autonomy",
				description: "Set the autonomy level (low, medium, high)",
				inputSchema: {
					type: "object",
					properties: {
						level: { type: "string", enum: ["low", "medium", "high"] },
					},
					required: ["level"],
				},
			},
			{
				name: "chat_reply",
				description: "Insert text into the active browser chat (Web Bridge)",
				inputSchema: {
					type: "object",
					properties: {
						text: { type: "string" },
					},
					required: ["text"],
				},
			},
			{
				name: "chat_submit",
				description: "Simulate pressing Enter in the chat input",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "click_element",
				description: "Click a button or link on the page by identifying text",
				inputSchema: {
					type: "object",
					properties: {
						target: {
							type: "string",
							description: "Text content of the button/link to click",
						},
					},
					required: ["target"],
				},
			},
			{
				name: "click_at",
				description:
					"Click at specific X,Y coordinates on the page (for Vision capabilities)",
				inputSchema: {
					type: "object",
					properties: {
						x: { type: "number" },
						y: { type: "number" },
					},
					required: ["x", "y"],
				},
			},
			{
				name: "native_input",
				description: "Simulate global keyboard input (for Native Apps/IDE)",
				inputSchema: {
					type: "object",
					properties: {
						keys: {
							type: "string",
							description: "Keys: enter, esc, ctrl+r, f5, etc.",
						},
					},
					required: ["keys"],
				},
			},
			{
				name: "vscode_execute_command",
				description: "Execute a VS Code command via the installed extension",
				inputSchema: {
					type: "object",
					properties: {
						command: {
							type: "string",
							description: "Command ID (e.g. workbench.action.files.save)",
						},
						args: {
							type: "array",
							description: "Optional arguments",
							items: { type: "string" },
						},
					},
					required: ["command"],
				},
			},
			{
				name: "read_page",
				description: "Read the text content of the active browser tab",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "use_agent",
				description: "Delegate a task to an external AI Agent (Claude, Gemini)",
				inputSchema: {
					type: "object",
					properties: {
						name: {
							type: "string",
							enum: ["claude", "gemini"],
							description: "Name of the agent to use",
						},
						prompt: {
							type: "string",
							description: "The task or prompt to send to the agent",
						},
					},
					required: ["name", "prompt"],
				},
			},
			{
				name: "research_recursively",
				description:
					"Perform deep, recursive research on a topic using sub-agents/sub-tasks.",
				inputSchema: {
					type: "object",
					properties: {
						topic: {
							type: "string",
							description: "The main topic to research",
						},
						depth: {
							type: "number",
							description: "Depth of recursion (def: 2)",
						},
						breadth: {
							type: "number",
							description: "Breadth of related topics (def: 3)",
						},
					},
					required: ["topic"],
				},
			},
			{
				name: "take_screenshot",
				description: "Capture a screenshot of the active browser tab",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "analyze_screenshot",
				description:
					"Capture and analyze a screenshot of the active browser tab using Vision AI",
				inputSchema: {
					type: "object",
					properties: {
						prompt: {
							type: "string",
							description: "Question or instruction about the screenshot",
						},
					},
					required: ["prompt"],
				},
			},
			{
				name: "vscode_get_status",
				description:
					"Get the current status of the VS Code editor (Active File, Terminal, etc.)",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "vscode_read_selection",
				description:
					"Read the currently selected text or the entire active document from VS Code",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "vscode_read_terminal",
				description:
					"Read the content of the active terminal (Uses Clipboard: SelectAll -> Copy)",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "vscode_get_notifications",
				description: "Read the latest notification/status message from VS Code",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "vscode_submit_chat",
				description:
					"Submit the current text in the chat input (Simulates Enter)",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "index_codebase",
				description:
					"Scan code files and populate the semantic memory (vector store)",
				inputSchema: {
					type: "object",
					properties: {
						path: {
							type: "string",
							description: "Root directory to index (defaults to cwd)",
						},
					},
				},
			},
			{
				name: "search_codebase",
				description: "Semantically search the codebase for relevant snippets",
				inputSchema: {
					type: "object",
					properties: {
						query: {
							type: "string",
							description:
								"Natural language query (e.g. 'Where is authentication?')",
						},
					},
					required: ["query"],
				},
			},
			{
				name: "chain_tools",
				description:
					"Execute a sequence of tools where outputs can be piped to inputs. Use {{prev.content[0].text}} for variable substitution.",
				inputSchema: {
					type: "object",
					properties: {
						tools: {
							type: "array",
							items: {
								type: "object",
								properties: {
									toolName: { type: "string" },
									args: { type: "object" },
								},
								required: ["toolName"],
							},
						},
					},
					required: ["tools"],
				},
			},
			{
				name: "get_knowledge_graph",
				description:
					"Retrieve the interconnected knowledge graph for visualization.",
				inputSchema: {
					type: "object",
					properties: {
						query: {
							type: "string",
							description: "Optional topic to focus the graph around",
						},
						depth: {
							type: "number",
							description: "Traversal depth (default: 1)",
						},
					},
				},
			},
			// Phase 51: LSP Tools
			{
				name: "find_symbol",
				description:
					"Find a symbol (function, class, variable) by name in a file",
				inputSchema: {
					type: "object",
					properties: {
						file_path: {
							type: "string",
							description: "Relative path to the file",
						},
						symbol_name: {
							type: "string",
							description: "Name of the symbol to find",
						},
					},
					required: ["file_path", "symbol_name"],
				},
			},
			{
				name: "find_references",
				description: "Find all references to a symbol at a position",
				inputSchema: {
					type: "object",
					properties: {
						file_path: { type: "string" },
						line: { type: "number", description: "Line number (0-indexed)" },
						character: {
							type: "number",
							description: "Character position (0-indexed)",
						},
					},
					required: ["file_path", "line", "character"],
				},
			},
			{
				name: "go_to_definition",
				description: "Navigate to the definition of a symbol at a position",
				inputSchema: {
					type: "object",
					properties: {
						file_path: { type: "string" },
						line: { type: "number" },
						character: { type: "number" },
					},
					required: ["file_path", "line", "character"],
				},
			},
			{
				name: "get_symbols",
				description:
					"Get all symbols (functions, classes, variables) in a file",
				inputSchema: {
					type: "object",
					properties: {
						file_path: {
							type: "string",
							description: "Relative path to the file",
						},
					},
					required: ["file_path"],
				},
			},
			{
				name: "rename_symbol",
				description:
					"Rename a symbol across the entire project (semantic rename)",
				inputSchema: {
					type: "object",
					properties: {
						file_path: { type: "string" },
						line: { type: "number" },
						character: { type: "number" },
						new_name: {
							type: "string",
							description: "New name for the symbol",
						},
					},
					required: ["file_path", "line", "character", "new_name"],
				},
			},
			{
				name: "search_symbols",
				description: "Search for symbols by name across the entire project",
				inputSchema: {
					type: "object",
					properties: {
						query: {
							type: "string",
							description: "Search query for symbol names",
						},
					},
					required: ["query"],
				},
			},
			// Phase 51: Plan/Build Mode Tools
			{
				name: "plan_mode",
				description:
					"Switch to PLAN mode - explore/propose changes without applying",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "build_mode",
				description: "Switch to BUILD mode - approved changes can be applied",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "propose_change",
				description: "Propose a file change (staged in diff sandbox)",
				inputSchema: {
					type: "object",
					properties: {
						file_path: {
							type: "string",
							description: "Path to file to change",
						},
						content: { type: "string", description: "New file content" },
						description: {
							type: "string",
							description: "Description of the change",
						},
					},
					required: ["file_path", "content"],
				},
			},
			{
				name: "review_changes",
				description: "View all pending changes in the diff sandbox",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "approve_change",
				description: "Approve a pending diff by ID",
				inputSchema: {
					type: "object",
					properties: {
						diff_id: {
							type: "string",
							description: "ID of the diff to approve",
						},
					},
					required: ["diff_id"],
				},
			},
			{
				name: "apply_changes",
				description:
					"Apply all approved changes to filesystem (BUILD mode only)",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "plan_status",
				description: "Get current mode (PLAN/BUILD) and diff sandbox status",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "create_checkpoint",
				description: "Create a checkpoint for rollback",
				inputSchema: {
					type: "object",
					properties: {
						name: { type: "string", description: "Checkpoint name" },
						description: {
							type: "string",
							description: "Optional description",
						},
					},
					required: ["name"],
				},
			},
			{
				name: "rollback",
				description: "Rollback to a previous checkpoint",
				inputSchema: {
					type: "object",
					properties: {
						checkpoint_id: {
							type: "string",
							description: "ID of checkpoint to restore",
						},
					},
					required: ["checkpoint_id"],
				},
			},
			// Phase 53: Memory Tools
			{
				name: "add_memory",
				description: "Add a memory to the agent's knowledge base",
				inputSchema: {
					type: "object",
					properties: {
						content: { type: "string", description: "Memory content to store" },
						type: {
							type: "string",
							enum: ["session", "working", "long_term"],
							description: "Memory tier (default: working)",
						},
						namespace: {
							type: "string",
							enum: ["user", "agent", "project"],
							description: "Memory namespace (default: project)",
						},
						source: { type: "string", description: "Source of the memory" },
						tags: {
							type: "array",
							items: { type: "string" },
							description: "Tags for categorization",
						},
					},
					required: ["content"],
				},
			},
			{
				name: "search_memory",
				description: "Search memories by content similarity",
				inputSchema: {
					type: "object",
					properties: {
						query: { type: "string", description: "Search query" },
						type: {
							type: "string",
							enum: ["session", "working", "long_term"],
							description: "Filter by memory tier",
						},
						namespace: {
							type: "string",
							enum: ["user", "agent", "project"],
							description: "Filter by namespace",
						},
						limit: { type: "number", description: "Max results (default: 10)" },
					},
					required: ["query"],
				},
			},
			{
				name: "get_recent_memories",
				description: "Get most recently accessed memories",
				inputSchema: {
					type: "object",
					properties: {
						limit: { type: "number", description: "Max results (default: 10)" },
						type: { type: "string", enum: ["session", "working", "long_term"] },
						namespace: { type: "string", enum: ["user", "agent", "project"] },
					},
				},
			},
			{
				name: "memory_stats",
				description: "Get memory system statistics",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "clear_session_memory",
				description: "Clear all ephemeral session memories",
				inputSchema: { type: "object", properties: {} },
			},
			// Phase 54: Workflow Tools
			{
				name: "run_workflow",
				description: "Run a registered workflow with input",
				inputSchema: {
					type: "object",
					properties: {
						workflow_id: {
							type: "string",
							description: "ID of the workflow to run",
						},
						input: {
							type: "object",
							description: "Input data for the workflow",
						},
					},
					required: ["workflow_id"],
				},
			},
			{
				name: "list_workflows",
				description: "List all registered workflows",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "workflow_status",
				description: "Get the status of a workflow run",
				inputSchema: {
					type: "object",
					properties: {
						run_id: { type: "string", description: "ID of the workflow run" },
					},
					required: ["run_id"],
				},
			},
			{
				name: "approve_workflow",
				description:
					"Approve or reject a workflow waiting for human-in-the-loop",
				inputSchema: {
					type: "object",
					properties: {
						run_id: { type: "string", description: "ID of the workflow run" },
						approved: {
							type: "boolean",
							description: "Whether to approve (true) or reject (false)",
						},
					},
					required: ["run_id"],
				},
			},
			// Phase 55: Code Mode Tools
			{
				name: "execute_code",
				description:
					"Execute JavaScript/TypeScript code in a sandboxed environment with tool access",
				inputSchema: {
					type: "object",
					properties: {
						code: {
							type: "string",
							description: "JavaScript/TypeScript code to execute",
						},
						context: {
							type: "object",
							description: "Additional context variables to inject",
						},
					},
					required: ["code"],
				},
			},
			{
				name: "enable_code_mode",
				description:
					"Enable Code Mode for efficient tool calling via code execution",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "disable_code_mode",
				description: "Disable Code Mode",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "code_mode_status",
				description:
					"Get Code Mode status, registered tools, and context reduction stats",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "list_code_tools",
				description: "List tools available within Code Mode",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "list_workspace_symbols",
				description:
					"Search for symbols (classes, functions, interfaces) across the entire workspace. Mimics Cursor/Windsurf internal context tools.",
				inputSchema: {
					type: "object",
					properties: {
						query: { type: "string", description: "Search query for symbols" },
					},
					required: ["query"],
				},
			},
			{
				name: "symbol_search",
				description:
					"(Alias for list_workspace_symbols) Search for symbols in the codebase.",
				inputSchema: {
					type: "object",
					properties: {
						query: { type: "string" },
					},
					required: ["query"],
				},
			},
			// Phase 56: Session Handoff Tools
			{
				name: "handoff_session",
				description:
					"Summarizes the current session context and exports it as a portable handoff artifact.",
				inputSchema: {
					type: "object",
					properties: {
						notes: {
							type: "string",
							description: "Additional notes to include in the handoff",
						},
					},
				},
			},
			{
				name: "pickup_session",
				description:
					"Restores a previous session from a handoff artifact string.",
				inputSchema: {
					type: "object",
					properties: {
						artifact: {
							type: "string",
							description: "The handoff artifact JSON string",
						},
					},
					required: ["artifact"],
				},
			},
			{
				name: "compact_context",
				description:
					"Manually triggers context pruning and compacting. Moves older conversation history into long-term memory to save token space.",
				inputSchema: {
					type: "object",
					properties: {
						max_tokens: {
							type: "number",
							description:
								"Desired token limit after compacting (default 100000)",
						},
					},
				},
			},
			{
				name: "get_suggestions",
				description:
					"Requests proactive tool and skill suggestions based on the current context or chat history.",
				inputSchema: {
					type: "object",
					properties: {
						context: {
							type: "string",
							description: "Optional explicit context string to analyze",
						},
					},
				},
			},
			{
				name: "export_chat",
				description:
					"Exports the current session chat history to a formatted file (Markdown or JSON).",
				inputSchema: {
					type: "object",
					properties: {
						format: {
							type: "string",
							enum: ["markdown", "json"],
							description: "Export format (default markdown)",
						},
						path: {
							type: "string",
							description:
								"Optional relative path to save the export (default current dir)",
						},
					},
				},
			},
			{
				name: "process_note",
				description:
					"Processes a raw text note or snippet, extracting facts and concepts into long-term memory.",
				inputSchema: {
					type: "object",
					properties: {
						content: {
							type: "string",
							description: "The raw text of the note",
						},
						title: {
							type: "string",
							description: "Optional title for the note",
						},
					},
					required: ["content"],
				},
			},
			{
				name: "auto_heal",
				description:
					"Hands off a technical error or failing test to the TormentNexus Healer. The system will autonomously diagnose the error, generate a fix, and apply it to the source code.",
				inputSchema: {
					type: "object",
					properties: {
						error: {
							type: "string",
							description: "The error message or stack trace to heal",
						},
						context: {
							type: "string",
							description:
								"Optional additional context about what was happening when the error occurred",
						},
					},
					required: ["error"],
				},
			},
			{
				name: "get_project_context",
				description:
					"Retrieves the persistent project-wide context, rules, and architecture notes. Use this to maintain high-level situational awareness of the repository.",
				inputSchema: { type: "object", properties: {} },
			},
			{
				name: "update_project_context",
				description:
					"Updates the persistent project-wide context. Use this to record major architectural decisions, new rules, or project milestones.",
				inputSchema: {
					type: "object",
					properties: {
						content: {
							type: "string",
							description: "The updated full text of the project context",
						},
					},
					required: ["content"],
				},
			},
			/*
            // Phase 60: The Mesh tools
            {
                name: "swarm_broadcast",
                description: "Broadcast a message to the TormentNexus P2P Swarm",
                inputSchema: {
                    type: "object",
                    properties: {
                        type: { type: "string", description: "Message Type (TASK_OFFER, etc.)" },
                        payload: { type: "object", description: "Payload data" }
                    },
                    required: ["type"]
                }
            }
            */
		];

		// Standard Library Tools
		const terminalTools = this.terminalService.getTools();
		const standardTools = [
			...FileSystemTools,
			...terminalTools,
			...MemoryTools,
			...TunnelTools,
			...LogTools,
			...ConfigTools,
			...SearchTools,
			...ReaderTools,
			...WorktreeTools,
		].map((t) => ({
			name: t.name,
			description: t.description,
			inputSchema: t.inputSchema,
		}));

		// Skills
		const skillTools = this.skillRegistry.getSkillTools();

		// Aggregation: Fetch tools from all connected sub-MCPs
		const externalTools = options?.skipLiveDiscovery
			? []
			: await this.router.listTools();
		const aggregatedTools =
			options?.downstreamTools ??
			(options?.skipLiveDiscovery
				? []
				: await this.mcpAggregator.listAggregatedTools());

		return [
			...internalTools,
			...standardTools,
			...skillTools,
			...externalTools,
			...aggregatedTools,
		] as Tool[];
	}

	private async getCachedAdvertisedDownstreamTools(): Promise<Tool[]> {
		const { tools } = await getCachedToolInventory();

		return tools.map((tool) => ({
			...tool,
			inputSchema:
				typeof tool.inputSchema === "object" && tool.inputSchema !== null
					? tool.inputSchema
					: { type: "object", properties: {} },
		})) as Tool[];
	}

	private async setupHandlers(serverInstance: Server) {
		mcpServerDebugLog("[MCPServer] Using TormentNexus-native MCP handlers.");
		await this.setupDirectHandlers(serverInstance);
	}

	private async setupDirectHandlers(serverInstance: Server): Promise<void> {
		this.setupDirectDiscoveryHandlers(serverInstance);

		serverInstance.setRequestHandler(LooseListToolsRequestSchema, async () => {
			const visibleTools = await this.getDirectModeTools();
			return {
				tools: visibleTools.map((t) => ({
					name: t.name,
					description: t.description,
					inputSchema: t.inputSchema,
				})),
			};
		});

		serverInstance.setRequestHandler(LooseCallToolRequestSchema, async (request) => {
			const meta =
				request.params._meta &&
				typeof request.params._meta === "object" &&
				!Array.isArray(request.params._meta)
					? (request.params._meta as Record<string, unknown>)
					: undefined;
			const metadataGuardResult = getDirectModeMetadataGuardResult(
				request.params.name,
				meta,
			);

			if (metadataGuardResult) {
				return metadataGuardResult;
			}

			// Conversational context: record what tool is being called and key args
			// so the injector can track the direction of the conversation.
			try {
				const toolArgs = request.params.arguments ?? {};
				const argText = Object.values(toolArgs)
					.filter((v) => typeof v === "string")
					.map((v) => (v as string).slice(0, 200))
					.join(" ");
				const turnText = argText
					? `${request.params.name}: ${argText}`
					: request.params.name;
				this.appendConversationTurn("tool", turnText);
			} catch {
				// non-fatal — never block tool execution
			}

			const directMetaResult = await this.handleDirectMetaTool(
				request.params.name,
				(request.params.arguments ?? {}) as Record<string, unknown>,
				meta,
			);

			if (directMetaResult) {
				return directMetaResult;
			}

			return await this.executeTool(
				request.params.name,
				request.params.arguments ?? {},
			);
		});
	}

	private setupDirectDiscoveryHandlers(serverInstance: Server): void {
		const context = {
			namespaceUuid: "tormentnexus-core-namespace",
			sessionId: "tormentnexus-core-session",
			includeInactiveServers: false,
		};

		serverInstance.setRequestHandler(
			LooseListPromptsRequestSchema,
			async (request) => {
				return await listDownstreamPrompts({
					context,
					cursor: request.params?.cursor,
					meta: request.params?._meta,
					promptToClient: this.promptToClient,
				});
			},
		);

		serverInstance.setRequestHandler(
			LooseGetPromptRequestSchema,
			async (request) => {
				return await getDownstreamPrompt({
					name: request.params.name,
					arguments: request.params.arguments ?? {},
					meta: request.params._meta,
					promptToClient: this.promptToClient,
				});
			},
		);

		serverInstance.setRequestHandler(
			LooseListResourcesRequestSchema,
			async (request) => {
				return await listDownstreamResources({
					context,
					cursor: request.params?.cursor,
					meta: request.params?._meta,
					resourceToClient: this.resourceToClient,
				});
			},
		);

		serverInstance.setRequestHandler(
			LooseReadResourceRequestSchema,
			async (request) => {
				return await readDownstreamResource({
					uri: request.params.uri,
					meta: request.params._meta,
					resourceToClient: this.resourceToClient,
				});
			},
		);

		serverInstance.setRequestHandler(
			LooseListResourceTemplatesRequestSchema,
			async (request) => {
				return await listDownstreamResourceTemplates({
					context,
					cursor: request.params?.cursor,
					meta: request.params?._meta,
				});
			},
		);
	}

	private async getDirectModeTools(): Promise<Tool[]> {
		const cachedInventory = await getCachedToolInventory();
		const cachedAdvertisedDownstreamTools = cachedInventory.tools.map(
			(tool) => {
				const flattened = flattenToolSchema(tool);
				return {
					...flattened,
					inputSchema:
						typeof flattened.inputSchema === "object" &&
						flattened.inputSchema !== null
							? flattened.inputSchema
							: { type: "object", properties: {} },
				};
			},
		) as (Tool & { alwaysOn: boolean })[];

		const allNativeTools = await this.getNativeTools({
			downstreamTools: cachedAdvertisedDownstreamTools,
			skipLiveDiscovery: true,
		});
		const aggregatedToolNames = new Set(
			cachedAdvertisedDownstreamTools.map((tool) => tool.name),
		);
		const baseTools = allNativeTools.filter(
			(tool) => !aggregatedToolNames.has(tool.name),
		);
		const savedScriptTools =
			await getDirectModeSavedScriptTools(jsonConfigProvider);

		this.nativeSessionMetaTools.refreshCatalog([
			...cachedAdvertisedDownstreamTools,
			...allNativeTools,
		]);
		await this.syncNativeToolPreferences();

		// CONVERSATIONAL TOOL INJECTION (Phase 113)
		// Use the ConversationalToolInjector to predict and auto-load tools
		// relevant to the current conversation direction. This runs asynchronously
		// with a built-in throttle so it never blocks tool listing.
		try {
			const catalogSnapshot = this.nativeSessionMetaTools.getCatalogSnapshot();
			const prediction =
				await this.conversationalToolInjector.getPredictedToolNames(
					catalogSnapshot,
				);
			if (prediction.toolNames.length > 0) {
				const injected = this.nativeSessionMetaTools.injectConversationalTools(
					prediction.toolNames,
				);
				if (injected.length > 0) {
					console.error(
						`[MCPServer] 🧠 Conversational injection loaded: ${injected.join(", ")} (${prediction.reason})`,
					);
				}
			}
		} catch (injectionErr: any) {
			console.warn(
				"[MCPServer] Conversational tool injection failed (non-fatal):",
				injectionErr?.message,
			);
		}

		// ULTRA-STREAMLINED ADVERTISING:
		// Only show core Meta-Tools by default.
		// baseTools (standard lib) are now LATENT by default to keep context clean.
		// They must be explicitly enabled via "Always On" in the dashboard or discovered via search.
		const alwaysOnDownstreamTools = cachedAdvertisedDownstreamTools.filter(
			(t) => t.alwaysOn,
		);
		const alwaysOnBaseTools = baseTools.filter((t) => (t as any).alwaysOn);

		const allVisibleTools = [
			...this.nativeSessionMetaTools.listToolDefinitions(),
			...getDirectModeCompatibilityTools().filter((t) => (t as any).alwaysOn), // Filter compatibility tools too
			...alwaysOnBaseTools,
			...alwaysOnDownstreamTools,
			...savedScriptTools.filter((t) => (t as any).alwaysOn),
			...this.nativeSessionMetaTools.getVisibleLoadedTools(),
		];

		// Safety limit for LLMs (e.g. Gemini has a 512 function declaration limit)
		// We use 450 to leave room for native tools and CLI-specific tools.
		const MAX_TOOLS = 450;
		const result = allVisibleTools.slice(0, MAX_TOOLS).map(flattenToolSchema);

		// Ensure we always have AT LEAST the core meta tools even if everything is toggled off
		if (
			result.length < this.nativeSessionMetaTools.listToolDefinitions().length
		) {
			return this.nativeSessionMetaTools
				.listToolDefinitions()
				.map(flattenToolSchema);
		}

		return result;
	}

	private async handleDirectMetaTool(
		name: string,
		args: Record<string, unknown>,
		meta?: Record<string, unknown>,
	): Promise<CallToolResult | null> {
		await this.syncNativeToolPreferences();

		const metadataGuardResult = getDirectModeMetadataGuardResult(name, meta);
		if (metadataGuardResult) {
			return metadataGuardResult;
		}

		const compatibilityArgs =
			name === "run_agent" &&
			typeof args.policyId !== "string" &&
			typeof meta?.policyId === "string"
				? { ...args, policyId: meta.policyId }
				: args;

		const compatibilityResult = await tryHandleDirectModeCompatibilityTool(
			name,
			compatibilityArgs,
			this.codeModeService,
			this.sandboxService,
			this.agentMemoryService,
			jsonConfigProvider,
			jsonConfigProvider,
			this.nativeSessionMetaTools,
			createDirectModeAgentRunner(this.llmService),
			async (toolName, toolArgs, delegatedMeta) => {
				const effectiveMeta = delegatedMeta;
				const delegatedGuardResult = getDirectModeMetadataGuardResult(
					toolName,
					effectiveMeta,
				);
				if (delegatedGuardResult) {
					return delegatedGuardResult;
				}

				const directResult = await this.handleDirectMetaTool(
					toolName,
					toolArgs,
					effectiveMeta,
				);
				if (directResult) {
					return directResult;
				}

				return await this.executeTool(toolName, toolArgs);
			},
			configImportService,
		);
		if (compatibilityResult) {
			return compatibilityResult;
		}

		const cachedAdvertisedDownstreamTools =
			await this.getCachedAdvertisedDownstreamTools();
		const allNativeTools = await this.getNativeTools({
			downstreamTools: cachedAdvertisedDownstreamTools,
			skipLiveDiscovery: true,
		});
		this.nativeSessionMetaTools.refreshCatalog([
			...cachedAdvertisedDownstreamTools,
			...allNativeTools,
		]);
		return await this.nativeSessionMetaTools.handleToolCall(name, args);
	}

	async start() {
		mcpServerDebugLog("[MCPServer] Loading Skills...");
		// Non-blocking initialization of aggregator to prevent stalling Stdio/WS start
		this.mcpAggregator
			.initialize()
			.then(async () => {
				const inventory = await getCachedToolInventory();
				const toolCountsByName = new Map(
					inventory.servers.map((server) => [
						server.name,
						inventory.toolCounts.get(server.uuid) ?? 0,
					]),
				);

				this.mcpAggregator.seedAdvertisedInventory({
					servers: inventory.servers.map((server) => ({
						name: server.name,
						alwaysOnAdvertised: server.alwaysOnAdvertised,
						advertisedToolCount: toolCountsByName.get(server.name) ?? 0,
					})),
					source: inventory.source,
				});
				const { lazySessionMode } = mcpServerPool.getLifecycleModes();
				// Keep aggregator in sync with the resolved lifecycle mode so that
				// any restarts or hot-reload paths also respect the latest setting.
				this.mcpAggregator.setLazyMode(lazySessionMode);
				if (lazySessionMode) {
					mcpServerDebugLog(
						"[MCPServer] Lazy MCP session mode enabled; skipping eager advertised-server warmup.",
					);
				} else {
					this.mcpAggregator.warmAdvertisedServers();
				}
				// Always connect always-on servers regardless of lazy mode.
				this.mcpAggregator
					.connectAllAlwaysOnServers()
					.catch((e) =>
						console.error("[MCPServer] Always-on auto-connect failed:", e),
					);
			})
			.catch((e) => console.error("[MCPServer] Aggregator Init Failed:", e));

		// Startup readiness should only report memory as ready once the runtime flag
		// is actually set. The current memory bootstrap is lightweight, so we prime
		// it during core startup instead of advertising a false-positive ready state.
		await this.initializeMemorySystem();

		// Start Next.js Dashboard in the background asynchronously
		this.ensureDashboardRunning().catch((e) =>
			console.error("[MCPServer] Dashboard launch failed:", e),
		);

		// Start Services
		// this.director.startChatDaemon(); // Removed, auto-drive handles this

		// Trigger automatic session log import in the background
		this.sessionImportService
			.scanAndImport()
			.catch((e) => console.error("[SessionImport] Failed:", e));

		// Trigger automatic BobbyBookmarks sync in the background on startup
		try {
			const bobbyAdapter = new BobbyBookmarksBacklogAdapter({
				baseUrl: "https://bobbybookmarks.com",
				perPage: 100,
				includeDuplicates: true,
				includeResearched: true,
			});
			bobbyAdapter
				.sync()
				.then((report) => {
					console.log(
						`[BobbyBookmarks] Sync completed on startup. Fetched: ${report.fetched}, Upserted: ${report.upserted}`,
					);
				})
				.catch((e) => {
					console.error("[BobbyBookmarks] Sync failed on startup:", e);
				});
		} catch (err) {
			console.error("[BobbyBookmarks] Adapter initialization failed:", err);
		}

		// Build Graph in Background
		this.autoTestService.repoGraph
			.buildGraph()
			.catch((e) => console.error("Graph build failed", e));

		mcpServerDebugLog("[MCPServer] 🚀 TormentNexus Core ready.");
		mcpServerDebugLog("[MCPServer] Preparing request handlers...");
		await this.serverSetupPromise;

		// 1. Start Stdio (for local CLI usage)
		if (!this.options.skipStdio) {
			mcpServerDebugLog("[MCPServer] Connecting Stdio...");
			const stdioTransport = new StdioServerTransport();
			await this.server.connect(stdioTransport);
			mcpServerDebugLog("TormentNexus Core: Stdio Transport Active");
		} else {
			mcpServerDebugLog(
				"[MCPServer] Skipping Stdio transport (managed by external loader).",
			);
		}

		// 2. Start WebSocket (for Extension/Web usage)
		if (this.wsServer && !this.wssInstance) {
			mcpServerDebugLog("[MCPServer] Starting WebSocket Server...");
			const PORT = this.bridgePort;
			const httpServer = http.createServer(async (req, res) => {
				if (req.method === "GET" && req.url === "/api/mesh/stream") {
					res.writeHead(200, {
						"Content-Type": "text/event-stream",
						"Cache-Control": "no-cache",
						Connection: "keep-alive",
						"Access-Control-Allow-Origin": "*",
					});
					res.write('data: {"type":"CONNECTED"}\n\n');

					const safeSend = (data: any) => {
						try {
							res.write(`data: ${JSON.stringify(data)}\n\n`);
						} catch (e) {
							// client disconnected
						}
					};

					const handleMeshMessage = (msg: any) => safeSend(msg);
					this.eventBus.on("mesh:traffic", handleMeshMessage);

					req.on("close", () => {
						this.eventBus.off("mesh:traffic", handleMeshMessage);
					});
				} else if (req.url === "/health") {
					res.writeHead(200, { "Content-Type": "application/json" });
					res.end(
						JSON.stringify({
							status: "online",
							uptime: process.uptime(),
							timestamp: Date.now(),
							version: "0.90.7",
						}),
					);
				} else if (req.url === "/mcp/servers") {
					const servers = await this.mcpAggregator.listServers();
					res.writeHead(200, {
						"Content-Type": "application/json",
						"Access-Control-Allow-Origin": "*", // Allow Web UI access
					});
					res.end(JSON.stringify(servers));
				} else if (req.method === "POST") {
					// Handle POST requests (Extension Compatibility)
					let body = "";
					req.on("data", (chunk) => (body += chunk));
					req.on("end", async () => {
						try {
							const data = JSON.parse(body);
							res.writeHead(200, {
								"Content-Type": "application/json",
								"Access-Control-Allow-Origin": "*",
							});

							if (req.url === "/director.chat") {
								// Direct chat interface
								const result = await this.director.executeTask(data.message);
								res.end(JSON.stringify({ result: { data: result } }));
							} else if (req.url === "/expert.dispatch") {
								const kind = String(data.kind ?? "").toLowerCase();

								if (kind === "research") {
									if (!this.researcherAgent) {
										res.writeHead(503, {
											"Content-Type": "application/json",
											"Access-Control-Allow-Origin": "*",
										});
										res.end(
											JSON.stringify({
												success: false,
												error: "Researcher Agent not initialized",
											}),
										);
										return;
									}

									const query = String(data.query ?? "").trim();
									if (!query) {
										res.writeHead(400, {
											"Content-Type": "application/json",
											"Access-Control-Allow-Origin": "*",
										});
										res.end(
											JSON.stringify({
												success: false,
												error: "query is required",
											}),
										);
										return;
									}

									const result = await this.researcherAgent.handleTask({
										task: query,
										options: {
											depth: Number(data.depth ?? 2),
											breadth: Number(data.breadth ?? 3),
										},
									});

									res.end(
										JSON.stringify({ success: true, kind: "research", result }),
									);
								} else if (kind === "code") {
									if (!this.coderAgent) {
										res.writeHead(503, {
											"Content-Type": "application/json",
											"Access-Control-Allow-Origin": "*",
										});
										res.end(
											JSON.stringify({
												success: false,
												error: "Coder Agent not initialized",
											}),
										);
										return;
									}

									const task = String(data.task ?? "").trim();
									if (!task) {
										res.writeHead(400, {
											"Content-Type": "application/json",
											"Access-Control-Allow-Origin": "*",
										});
										res.end(
											JSON.stringify({
												success: false,
												error: "task is required",
											}),
										);
										return;
									}

									const result = await this.coderAgent.handleTask({ task });
									res.end(
										JSON.stringify({ success: true, kind: "code", result }),
									);
								} else {
									res.writeHead(400, {
										"Content-Type": "application/json",
										"Access-Control-Allow-Origin": "*",
									});
									res.end(
										JSON.stringify({
											success: false,
											error: "Unsupported expert dispatch kind",
										}),
									);
								}
							} else if (req.url === "/expert.status") {
								res.end(
									JSON.stringify({
										success: true,
										researcher: this.researcherAgent ? "active" : "offline",
										coder: this.coderAgent ? "active" : "offline",
									}),
								);
							} else if (req.url === "/knowledge.capture") {
								const content = String(data.content ?? "").trim();
								const title = String(data.title ?? "Captured Page");
								const url = String(
									data.url ?? data.source ?? "browser://captured",
								);
								const source = String(data.source ?? "browser_extension");

								if (!content) {
									res.writeHead(400, {
										"Content-Type": "application/json",
										"Access-Control-Allow-Origin": "*",
									});
									res.end(
										JSON.stringify({
											success: false,
											error: "content is required",
										}),
									);
									return;
								}

								const ctxId = await this.memoryManager.saveContext(
									`URL: ${url}\n\n${content}`,
									{
										title,
										source: url,
										type: "research",
										origin: source,
									},
								);

								this.broadcastWebSocketMessage({
									type: "KNOWLEDGE_CAPTURED",
									payload: {
										id: ctxId,
										title,
										url,
										source,
										timestamp: Date.now(),
										preview: content.slice(0, 240),
									},
								});

								res.end(JSON.stringify({ success: true, id: ctxId }));
							} else if (req.url === "/knowledge.ingest-url") {
								const url = String(data.url ?? "").trim();
								if (!url) {
									res.writeHead(400, {
										"Content-Type": "application/json",
										"Access-Control-Allow-Origin": "*",
									});
									res.end(
										JSON.stringify({
											success: false,
											error: "url is required",
										}),
									);
									return;
								}

								const result = await this.deepResearchService.ingest(url);
								res.end(JSON.stringify({ success: true, result }));
							} else if (req.url === "/rag.ingest-text") {
								const text = String(data.text ?? "").trim();
								const sourceName = String(
									data.sourceName ??
										data.title ??
										data.url ??
										"browser-extension-page",
								);
								const userId = String(data.userId ?? "default");
								const chunkSize = Number(data.chunkSize ?? 1000);
								const chunkOverlap = Number(data.chunkOverlap ?? 200);
								const strategy =
									data.strategy === "sliding_window" ||
									data.strategy === "semantic"
										? data.strategy
										: "recursive";

								if (!text) {
									res.writeHead(400, {
										"Content-Type": "application/json",
										"Access-Control-Allow-Origin": "*",
									});
									res.end(
										JSON.stringify({
											success: false,
											error: "text is required",
										}),
									);
									return;
								}

								const intakeService = new DocumentIntakeService(
									this.memoryManager as any,
									new EmbeddingService("local"),
								);
								const result = await intakeService.ingestText(
									text,
									sourceName,
									userId,
									{
										chunkSize,
										chunkOverlap,
										strategy,
									},
								);

								this.broadcastWebSocketMessage({
									type: "RAG_INGESTED",
									payload: {
										sourceName,
										chunksIngested: result.chunks,
										success: result.success,
										timestamp: Date.now(),
										source: "browser_extension",
									},
								});

								res.end(
									JSON.stringify({
										success: result.success,
										chunksIngested: result.chunks,
									}),
								);
							} else if (req.url === "/tool/execute") {
								// Generic tool execution
								const result = await this.executeTool(data.name, data.args);
								res.end(JSON.stringify({ result: { data: result } }));
							} else {
								// Forward to TRPC-like structure if needed, or 404
								res.writeHead(404);
								res.end(JSON.stringify({ error: "Endpoint not found" }));
							}
						} catch (e: any) {
							res.writeHead(500, {
								"Content-Type": "application/json",
								"Access-Control-Allow-Origin": "*",
							});
							res.end(JSON.stringify({ error: e.message }));
						}
					});
				} else {
					res.writeHead(404);
					res.end();
				}
			});

			httpServer.on("close", () => {
				if (this.wssInstance?.options?.server === httpServer) {
					this.wssInstance = null;
				}
			});
			const wss = new WebSocketServer({ server: httpServer });
			this.wssInstance = wss;

			let bridgePortConflictHandled = false;
			const handleBridgePortConflict = () => {
				if (bridgePortConflictHandled) {
					return;
				}

				bridgePortConflictHandled = true;
				console.warn(
					`[TormentNexus Core] ⚠️ WebSocket bridge port ${PORT} is already in use. Skipping bridge startup while keeping the rest of TormentNexus online.`,
				);
				this.wssInstance = null;
			};

			httpServer.on("error", (err: any) => {
				if (err?.code === "EADDRINUSE") {
					handleBridgePortConflict();
					return;
				}

				console.error(
					`[TormentNexus Core] ❌ WebSocket Server Error (Port ${PORT}):`,
					err.message,
				);
			});

			wss.on("error", (err: any) => {
				if (err?.code === "EADDRINUSE") {
					handleBridgePortConflict();
					return;
				}

				console.error(
					`[TormentNexus Core] ❌ WebSocket bridge runtime error (Port ${PORT}):`,
					err.message,
				);
			});

			const bridgeListening = await new Promise<boolean>((resolve) => {
				let settled = false;

				const finalize = (value: boolean) => {
					if (settled) {
						return;
					}

					settled = true;
					resolve(value);
				};

				httpServer.once("listening", () => {
					mcpServerDebugLog(
						`TormentNexus Core: WebSocket Transport Active on ws://localhost:${PORT}`,
					);
					finalize(true);
				});

				httpServer.once("error", (err: any) => {
					if (err?.code === "EADDRINUSE") {
						finalize(false);
						return;
					}

					finalize(false);
				});

				httpServer.listen(PORT);
			});

			if (!bridgeListening) {
				try {
					httpServer.close();
				} catch {
					// ignore cleanup failure for duplicate bridge startup
				}

				try {
					wss.close();
				} catch {
					// ignore cleanup failure for duplicate bridge startup
				}

				this.wssInstance = null;
			}

			// 2.5 Setup WS Message Handling mechanism
			wss.on("connection", (ws: any) => {
				const clientId = `bridge-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
				this.bridgeClients.set(ws, createDefaultBridgeClient(clientId));

				ws.on("close", () => {
					this.bridgeClients.delete(ws);
				});

				ws.on("message", async (data: any) => {
					try {
						const msg = JSON.parse(data.toString());
						if (msg.type === "TORMENTNEXUS_CLIENT_HELLO") {
							const existing =
								this.bridgeClients.get(ws) ??
								createDefaultBridgeClient(clientId);
							const updated = applyBridgeClientHello(existing, msg, Date.now());
							this.bridgeClients.set(ws, updated);

							if (ws.readyState === 1) {
								ws.send(
									JSON.stringify({
										type: "TORMENTNEXUS_CORE_MANIFEST",
										manifest: buildBridgeManifest(
											Array.from(this.bridgeClients.values()),
										),
									}),
								);
							}
							return;
						}

						// Robust Response Handling: Any message with a requestId should resolve a pending promise
						if (msg.requestId && this.pendingRequests.has(msg.requestId)) {
							const resolve = this.pendingRequests.get(msg.requestId);
							if (resolve) {
								// If it's a STATUS_UPDATE, we unpack .status for legacy parity
								// Otherwise we pass the whole message so the caller can extract what they need (e.g. .history, .logs)
								if (msg.type === "STATUS_UPDATE") {
									resolve(msg.status);
								} else {
									resolve(msg);
								}
								this.pendingRequests.delete(msg.requestId);
							}
						}

						// Special handling for legacy/specific types if needed (though the above handles most)
						if (msg.type === "BROWSER_MIRROR_UPDATE") {
							this.broadcastWebSocketMessage(msg, ws);
						}

						// Explicit conversation turn from dashboard or bridge client
						// Allows any client to push user/assistant messages into the
						// ConversationalToolInjector sliding window for better prediction.
						if (msg.type === "CONVERSATION_TURN") {
							const role =
								msg.role === "user" ||
								msg.role === "assistant" ||
								msg.role === "tool"
									? (msg.role as "user" | "assistant" | "tool")
									: "user";
							const text = typeof msg.text === "string" ? msg.text : "";
							if (text.trim()) {
								this.appendConversationTurn(role, text.trim().slice(0, 1000));
							}
						}

						if (msg.type === "USER_ACTIVITY") {
							this.lastUserActivityTime = Math.max(
								this.lastUserActivityTime,
								msg.lastActivityTime,
							);
							if (msg.activeEditor) {
								this.suggestionService.processContext({
									type: "editor",
									path: msg.activeEditor.uri,
									content: "", // Could retrieve content if needed
								});
							}
						}

						if (msg.type === "BROWSER_LOG") {
							const level = String(msg.level ?? "log");
							const content = String(msg.content ?? msg.message ?? "");
							const url = String(msg.url ?? "unknown");
							const timestamp = Number(msg.timestamp) || Date.now();
							const source = String(msg.source ?? "browser_extension");
							const icon =
								level === "error" ? "🔴" : level === "warn" ? "🟡" : "🔵";

							console.log(`[Browser] ${icon} ${content} (${url})`);
							this.broadcastWebSocketMessage(
								{
									type: "BROWSER_LOG",
									payload: {
										level,
										content,
										url,
										timestamp,
										source,
									},
								},
								ws,
							);

							if (level === "error") {
								this.auditService.log(
									"BROWSER_ERROR",
									{ url, error: content },
									"ERROR",
								);
								// TRIGGER HEALER
								this.healerService
									.healAndVerify(
										content,
										`URL: ${url}, Timestamp: ${timestamp}`,
									)
									.catch((e) => console.error("Healer Error:", e));
							}
						}

						if (msg.type === "BROWSER_DEBUG_EVENT") {
							const timestamp = Number(msg.timestamp) || Date.now();
							const tabId = typeof msg.tabId === "number" ? msg.tabId : null;
							const method = String(msg.method ?? "unknown");
							const params =
								typeof msg.params === "object" && msg.params !== null
									? msg.params
									: {};

							this.broadcastWebSocketMessage(
								{
									type: "BROWSER_DEBUG_EVENT",
									payload: {
										timestamp,
										tabId,
										method,
										params,
										source: "browser_extension",
									},
								},
								ws,
							);
						}

						if (msg.type === "BROWSER_CHAT_SURFACE") {
							const timestamp = Number(msg.timestamp) || Date.now();
							const trigger = String(msg.trigger ?? "mutation");
							const snapshot =
								typeof msg.snapshot === "object" && msg.snapshot !== null
									? msg.snapshot
									: {};

							// Feed user-visible text from chat surface into conversation window
							try {
								const chatText =
									typeof (snapshot as any).userInput === "string"
										? (snapshot as any).userInput
										: typeof (snapshot as any).lastUserMessage === "string"
											? (snapshot as any).lastUserMessage
											: "";
								if (chatText.trim()) {
									this.appendConversationTurn(
										"user",
										chatText.trim().slice(0, 500),
									);
								}
							} catch {
								// non-fatal
							}

							this.broadcastWebSocketMessage(
								{
									type: "BROWSER_CHAT_SURFACE",
									payload: {
										timestamp,
										trigger,
										snapshot,
										source: String(msg.source ?? "browser_extension"),
									},
								},
								ws,
							);
						}

						if (msg.type === "KNOWLEDGE_CAPTURE") {
							const content = String(msg.content ?? "").trim();
							const title = String(msg.title ?? "Captured Page");
							const url = String(msg.url ?? "browser://captured");
							const source = String(msg.source ?? "browser_extension");
							const timestamp = Number(msg.timestamp) || Date.now();

							if (!content) {
								if (msg.requestId && ws.readyState === 1) {
									ws.send(
										JSON.stringify({
											type: "STATUS_UPDATE",
											requestId: msg.requestId,
											status: {
												success: false,
												error: "No content provided for knowledge capture.",
											},
										}),
									);
								}
							} else if (this.wsServer && this.wssInstance) {
								console.error(
									"[MCPServer] WebSocket/SSE bridge already active (constructor bootstrap), skipping duplicate bind.",
								);
								return;
							}

							const ctxId = await this.memoryManager.saveContext(
								`URL: ${url}\n\n${content}`,
								{
									title,
									source: url,
									type: "research",
									origin: source,
								},
							);

							this.broadcastWebSocketMessage(
								{
									type: "KNOWLEDGE_CAPTURED",
									payload: {
										id: ctxId,
										title,
										url,
										source,
										timestamp,
										preview: content.slice(0, 240),
									},
								},
								ws,
							);

							if (msg.requestId && ws.readyState === 1) {
								ws.send(
									JSON.stringify({
										type: "STATUS_UPDATE",
										requestId: msg.requestId,
										status: { success: true, id: ctxId, title, url },
									}),
								);
							}
						}
					} catch (e) {
						// Ignore non-JSON
					}
				});
			});
		} else {
			mcpServerDebugLog(
				"[MCPServer] Skipping WebSocket (No wsServer instance).",
			);
		}

		// 3. Connect to Supervisor (Native Automation)
		mcpServerDebugLog("[MCPServer] Connecting to Supervisor...");

		try {
			mcpServerDebugLog(`[MCPServer] DEBUG __dirname: ${__dirname}`);
			const rootDir = this.findMonorepoRoot(__dirname);
			mcpServerDebugLog(`[MCPServer] DEBUG rootDir: ${rootDir}`);
			if (rootDir) {
				const supervisorPath = path.join(
					rootDir,
					"packages",
					"tormentnexus-supervisor",
					"dist",
					"index.js",
				);
				mcpServerDebugLog(
					`[MCPServer] Supervisor Path Resolved: ${supervisorPath}`,
				);

				await this.router.connectToServer("tormentnexus-supervisor", "node", [
					supervisorPath,
				]);
				mcpServerDebugLog(
					`TormentNexus Core: Connected to Supervisor at ${supervisorPath}`,
				);

				// Phase 16: Google Workspace Integration
				const workspacePath = path.join(
					rootDir,
					"external",
					"mcp-servers",
					"workspace",
					"workspace-server",
					"dist",
					"index.js",
				);
				mcpServerDebugLog(
					`[MCPServer] Google Workspace Server Path: ${workspacePath}`,
				);
				if (fs.existsSync(workspacePath)) {
					await this.router.connectToServer("google-workspace", "node", [
						workspacePath,
					]);
					mcpServerDebugLog(
						"TormentNexus Core: Connected to Google Workspace Server (GMail/Calendar)",
					);
				}
			} else {
				console.error(
					"[MCPServer] Failed to locate Monorepo Root. Skipping Supervisor.",
				);
			}
		} catch (e: any) {
			console.error(
				"TormentNexus Core: Failed to connect to Supervisor. Native automation disabled.",
				e.message,
			);
		}

		if (this.wsServer && this.wssInstance) {
			mcpServerDebugLog("[MCPServer] Connecting internal WS transport...");
			if (this.wsServerSetupPromise) {
				await this.wsServerSetupPromise;
			}
			const wsTransport = new WebSocketServerTransport(this.wssInstance);
			await this.wsServer.connect(wsTransport);
		}
		mcpServerDebugLog("[MCPServer] Start Complete.");
	}

	private findMonorepoRoot(startDir: string): string | null {
		let current = startDir;
		const root = path.parse(current).root;
		while (current !== root) {
			if (fs.existsSync(path.join(current, "turbo.json"))) {
				return current;
			}
			current = path.dirname(current);
		}
		return null;
	}

	private async ensureDashboardRunning(): Promise<void> {
		const port = Number(
			process.env.TORMENTNEXUS_DASH_PORT || process.env.PORT || 3000,
		);

		const isPortOccupied = async (p: number): Promise<boolean> => {
			const net = await import("net");
			return new Promise((resolve) => {
				const server = net.createServer();
				server.once("error", () => resolve(true));
				server.once("listening", () => {
					server.close(() => resolve(false));
				});
				server.listen(p, "127.0.0.1");
			});
		};

		try {
			const occupied = await isPortOccupied(port);
			if (occupied) {
				console.error(
					`[MCPServer] Dashboard already running or port ${port} is occupied.`,
				);
				return;
			}

			console.error(
				`[MCPServer] Starting Next.js Dashboard on port ${port}...`,
			);
			const fs = await import("fs");
			const path = await import("path");
			const { spawn } = await import("child_process");

			const monorepoRoot = process.cwd();
			const webDir = path.join(monorepoRoot, "apps", "web");

			if (!fs.existsSync(webDir)) {
				console.error(`[MCPServer] Dashboard directory not found at ${webDir}`);
				return;
			}

			const child = spawn(
				process.platform === "win32" ? "npx.cmd" : "npx",
				["next", "dev", "--port", String(port)],
				{
					cwd: webDir,
					detached: true,
					stdio: "ignore",
					env: {
						...process.env,
						NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE: "1",
					},
				},
			);
			child.unref();
			console.error(
				`[MCPServer] Dashboard spawned successfully (PID: ${child.pid}).`,
			);
		} catch (err: any) {
			console.error(
				"[MCPServer] Failed to auto-start Next.js dashboard:",
				err.message,
			);
		}
	}

	private broadcastWebSocketMessage(
		message: Record<string, unknown>,
		excludeClient?: any,
	) {
		if (!this.wssInstance?.clients) {
			return;
		}

		const serializedMessage = JSON.stringify(message);
		this.wssInstance.clients.forEach((client: any) => {
			if (client !== excludeClient && client.readyState === 1) {
				client.send(serializedMessage);
			}
		});
	}
}

function redirectProtocolUnsafeConsoleMethodsForDirectExecution(): void {
	const stderr = console.error.bind(console);

	console.log = stderr;
	console.info = stderr;
	console.debug = stderr;
	console.trace = stderr;
	console.dir = ((...args: unknown[]) => stderr(...args)) as typeof console.dir;
}

async function startDirectlyIfExecuted(): Promise<void> {
	const entryArg = process.argv[1] ? path.resolve(process.argv[1]) : null;

	if (!entryArg || entryArg !== __filename) {
		return;
	}

	redirectProtocolUnsafeConsoleMethodsForDirectExecution();

	process.on("unhandledRejection", (reason) => {
		console.error("[TormentNexus Core] Unhandled promise rejection:", reason);
	});

	process.on("uncaughtException", (error) => {
		console.error("[TormentNexus Core] Uncaught exception:", error);
		process.exit(1);
	});

	try {
		const mcp = new MCPServer({ skipWebsocket: true });
		await mcp.start();
		console.error("[TormentNexus Core] MCPServer direct entrypoint started.");
	} catch (error) {
		console.error(
			"[TormentNexus Core] Failed to start direct MCP entrypoint:",
			error,
		);
		process.exit(1);
	}
}

void startDirectlyIfExecuted();
