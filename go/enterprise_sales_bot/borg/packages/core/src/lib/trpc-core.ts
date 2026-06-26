import { initTRPC } from "@trpc/server";
import { z } from "zod";

export type MCPServerRuntime = {
	directorConfig?: {
		demo_mode?: boolean;
	};
	workflowEngine?: unknown;
	marketplaceService?: unknown;
	[key: string]: any;
};

export type WorkflowEngineRuntime = {
	getGraph: (workflowId: string) =>
		| {
				nodes: Array<{
					id: string;
					label: string;
					data: { description?: string };
					type: "checkpoint" | "default";
				}>;
				edges: Array<{
					id: string;
					source: string;
					target: string;
					animated: boolean;
					label?: string;
				}>;
		  }
		| undefined;
	start: (
		workflowId: string,
		initialState: Record<string, unknown>,
	) => Promise<unknown>;
	listExecutions: () => unknown[];
	getExecution: (executionId: string) => unknown;
	getHistory: (executionId: string) => unknown[];
	resume: (executionId: string) => Promise<void>;
	pause: (executionId: string) => void;
	approve: (executionId: string) => Promise<void>;
	reject: (executionId: string, reason?: string) => void;
};

export type WorkflowDefinitionRuntime = {
	id: string;
	name?: string;
	description?: string;
	entryPoint?: string;
	nodes?: Map<string, unknown>;
	edges?: unknown[];
};

export type AutoTestResultRuntime = {
	status: string;
	timestamp: number;
	output?: string;
};

export type AutoTestServiceRuntime = {
	isRunning: boolean;
	testResults: Map<string, AutoTestResultRuntime>;
	start: () => Promise<void>;
	stop: () => void;
	findTestFile?: (filePath: string) => string | undefined;
	runTest?: (testFile: string) => void;
	repoGraph?: unknown;
};

export type HealerServiceRuntime = {
	analyzeError: (error: string, context: string) => unknown;
	heal: (error: string, context: string) => Promise<boolean>;
	getHistory: () => unknown;
	on: (event: "heal", listener: (data: unknown) => void) => void;
	off: (event: "heal", listener: (data: unknown) => void) => void;
};

export type GitServiceRuntime = {
	getLog: (limit?: number) => unknown;
	getStatus: () => unknown;
	revert: (hash: string) => Promise<unknown>;
};

export type PermissionManagerRuntime = {
	autonomyLevel: "low" | "medium" | "high";
	setAutonomyLevel: (level: "low" | "medium" | "high") => void;
};

export type DirectorRuntime = {
	getStatus?: () => unknown;
	getConfig?: () => unknown;
	updateConfig?: (config: Record<string, unknown>) => void;
	stopAutoDrive?: () => void;
	broadcast?: (message: string) => void;
	executeTask?: (task: string) => Promise<unknown>;
	startChatDaemon?: () => void;
	startWatchdog?: (intervalMs?: number) => void;
};

export type AuditServiceRuntime = {
	getLogs: (limit?: number) => unknown;
	query: (input: Record<string, unknown>) => unknown;
};

export type DarwinExperimentRuntime = {
	id: string;
};

export type DarwinServiceRuntime = {
	proposeMutation: (prompt: string, goal: string) => unknown;
	startExperiment: (
		mutationId: string,
		task: string,
	) => Promise<DarwinExperimentRuntime>;
	getStatus: () => unknown;
};

export type CouncilOrchestratorRuntime = {
	runConsensusSession: (proposal: string) => Promise<unknown>;
	lastResult: unknown;
};

export type CouncilSessionRuntime = {
	id: string;
	topic: string;
	status: "active" | "concluded";
	round: number;
	opinions: Array<{
		agentId: string;
		content: string;
		timestamp: number;
		round: number;
	}>;
	votes: Array<{
		agentId: string;
		choice: string;
		reason: string;
		timestamp: number;
	}>;
	createdAt: number;
};

export type CouncilServiceRuntime = {
	listSessions: () => CouncilSessionRuntime[];
	getSession: (id: string) => CouncilSessionRuntime | undefined;
};

export type MemoryManagerRuntime = {
	saveContext: (
		content: string,
		metadata?: Record<string, unknown>,
	) => Promise<unknown>;
	search?: (query: string, limit?: number) => Promise<unknown[]>;
	listContexts: () => Promise<unknown[]>;
	searchSymbols?: (query: string, limit?: number) => Promise<unknown[]>;
	getContext?: (id: string) => Promise<unknown>;
	deleteContext?: (id: string) => Promise<void>;
	getAllSymbols?: () => Promise<unknown[]>;
};

export type AgentMemoryServiceRuntime = {
	getStats: () => unknown;
	search: (
		query: string,
		options?: Record<string, unknown>,
	) => Promise<unknown[]>;
	add: (
		content: string,
		type?: unknown,
		source?: string,
		metadata?: Record<string, unknown>,
	) => Promise<void>;
	recordObservation?: (input: Record<string, unknown>) => Promise<unknown>;
	captureSessionSummary?: (input: Record<string, unknown>) => Promise<unknown>;
	getRecentObservations?: (
		limit?: number,
		options?: Record<string, unknown>,
	) => unknown[];
	searchObservations?: (
		query: string,
		options?: Record<string, unknown>,
	) => Promise<unknown[]>;
	getRecentSessionSummaries?: (limit?: number) => unknown[];
	searchSessionSummaries?: (
		query: string,
		limit?: number,
	) => Promise<unknown[]>;
	getSessionBootstrap?: (options?: {
		activeGoal?: string | null;
		lastObjective?: string | null;
		toolAdvertisementLines?: string[];
	}) => unknown;
	getToolContext?: (input: {
		toolName: string;
		args?: unknown;
		activeGoal?: string | null;
		lastObjective?: string | null;
	}) => unknown;
	captureUserPrompt?: (input: Record<string, unknown>) => Promise<unknown>;
	getRecentUserPrompts?: (
		limit?: number,
		options?: Record<string, unknown>,
	) => unknown[];
	searchUserPrompts?: (
		query: string,
		options?: Record<string, unknown>,
	) => Promise<unknown[]>;
	searchByPivot?: (input: Record<string, unknown>) => unknown[];
	getTimelineWindow?: (input: Record<string, unknown>) => unknown[];
	getCrossSessionLinks?: (input: Record<string, unknown>) => unknown[];
};

export type SessionImportServiceRuntime = {
	scanAndImport: (options?: { force?: boolean }) => Promise<unknown>;
	listImportedSessions: (limit?: number) => unknown[];
	getImportedSession: (id: string) => unknown | null;
	getImportedMaintenanceStats?: () => unknown;
	listInstructionDocs?: () => Promise<unknown[]>;
};

export type SymbolPinServiceRuntime = {
	list: () => unknown[];
	pin: (input: Record<string, unknown>) => unknown;
	unpin: (id: string) => unknown;
	updatePriority: (id: string, priority: number) => unknown;
	addNotes: (id: string, notes: string) => unknown;
	clear: () => unknown;
	forFile: (filePath: string) => unknown[];
};

export type ConfigManagerRuntime = {
	loadConfig: () => Record<string, unknown>;
	saveConfig: (config: Record<string, unknown>) => void;
};

export type CommandRegistryRuntime = {
	execute: (
		input: string,
	) => Promise<{ handled?: boolean; output?: string } | null | undefined>;
	getCommands?: () => Array<{ name: string; description?: string }>;
};

export type ContextManagerRuntime = {
	list: () => unknown[];
	add: (filePath: string) => unknown;
	remove: (filePath: string) => unknown;
	clear: () => unknown;
	getContextPrompt: () => string;
};

export type AutoDevServiceRuntime = {
	startLoop: (input: Record<string, unknown>) => Promise<string>;
	cancelLoop: (loopId: string) => boolean;
	getLoops: () => unknown[];
	getLoop: (loopId: string) => unknown;
	clearCompleted: () => number;
};

export type ShellServiceRuntime = {
	logCommand: (input: Record<string, unknown>) => Promise<string>;
	queryHistory: (
		query: string,
		limit?: number,
	) => Promise<unknown[]> | unknown[];
	getSystemHistory: (limit?: number) => Promise<unknown[]> | unknown[];
	executeWithContext?: (
		command: string,
		options?: Record<string, unknown>,
	) => Promise<unknown>;
};

export type EventBusRuntime = {
	getHistory: (
		limit?: number,
	) => Array<{ timestamp: number; [key: string]: unknown }>;
};

export type SuggestionRuntime = {
	id: string;
	title: string;
	payload?: {
		tool?: string;
		args?: Record<string, unknown>;
	};
};

export type SuggestionServiceRuntime = {
	getPendingSuggestions: () => SuggestionRuntime[];
	resolveSuggestion: (
		id: string,
		decision: "APPROVED" | "REJECTED",
	) => SuggestionRuntime | null | undefined;
	clearAll?: () => void;
};

export type SquadServiceRuntime = {
	listMembers: () => Promise<unknown[]>;
	spawnMember: (branch: string, goal: string) => Promise<unknown>;
	killMember: (branch: string) => Promise<unknown>;
	messageMember: (branch: string, message: string) => Promise<unknown>;
	toggleIndexer: (enabled: boolean) => Promise<boolean>;
	getIndexerStatus: () => { running: boolean; indexing: boolean };
};

export type SessionStateRuntime = {
	isAutoDriveActive: boolean;
	activeGoal: string | null;
	lastObjective: string | null;
	lastHeartbeat: number;
};

export type SessionManagerRuntime = {
	getState: () => SessionStateRuntime;
	updateState: (state: Partial<SessionStateRuntime>) => void;
	save: () => void;
	clearSession: () => void;
	touch: () => void;
};

export type SessionSupervisorLogRuntime = {
	timestamp: number;
	stream: "stdout" | "stderr" | "system";
	message: string;
};

export type SessionExecutionPolicyRuntime = {
	requestedProfile: "auto" | "powershell" | "posix" | "compatibility";
	effectiveProfile: "powershell" | "posix" | "compatibility" | "fallback";
	shellId: string | null;
	shellLabel: string | null;
	shellFamily: "powershell" | "cmd" | "posix" | "wsl" | null;
	shellPath: string | null;
	supportsPowerShell: boolean;
	supportsPosixShell: boolean;
	reason: string;
};

export type SessionSupervisorSessionRuntime = {
	id: string;
	name: string;
	cliType: string;
	command: string;
	args: string[];
	env: Record<string, string>;
	executionProfile: "auto" | "powershell" | "posix" | "compatibility";
	executionPolicy: SessionExecutionPolicyRuntime | null;
	requestedWorkingDirectory: string;
	workingDirectory: string;
	worktreePath?: string;
	autoRestart: boolean;
	isolateWorktree: boolean;
	status:
		| "created"
		| "starting"
		| "running"
		| "stopping"
		| "stopped"
		| "restarting"
		| "error";
	createdAt: number;
	startedAt?: number;
	stoppedAt?: number;
	scheduledRestartAt?: number;
	lastActivityAt: number;
	restartCount: number;
	maxRestartAttempts: number;
	lastExitCode?: number;
	lastExitSignal?: string;
	lastError?: string;
	metadata: Record<string, unknown>;
	logs: SessionSupervisorLogRuntime[];
};

export type SessionSupervisorRuntime = {
	listSessions: () => SessionSupervisorSessionRuntime[];
	getSession: (id: string) => SessionSupervisorSessionRuntime | undefined;
	createSession: (
		input: Record<string, unknown>,
	) => Promise<SessionSupervisorSessionRuntime>;
	startSession: (id: string) => Promise<SessionSupervisorSessionRuntime>;
	updateSessionMetadata: (
		id: string,
		metadataPatch: Record<string, unknown>,
	) => SessionSupervisorSessionRuntime;
	stopSession: (
		id: string,
		options?: { force?: boolean },
	) => Promise<SessionSupervisorSessionRuntime>;
	restartSession: (id: string) => Promise<SessionSupervisorSessionRuntime>;
	getSessionLogs: (id: string, limit?: number) => SessionSupervisorLogRuntime[];
	getAttachInfo: (id: string) => {
		id: string;
		pid?: number;
		command: string;
		args: string[];
		cwd: string;
		status: string;
		attachable: boolean;
	};
	getSessionHealth: (id: string) => {
		status: string;
		lastCheck: number;
		consecutiveFailures: number;
		restartCount: number;
		lastRestartAt?: number;
		nextRestartAt?: number;
		lastExitCode?: number;
		lastExitSignal?: string;
		errorMessage?: string;
	};
	restoreSessions: () => SessionSupervisorSessionRuntime[];
	getRestoreStatus: () => {
		lastRestoreAt?: number;
		restoredSessionCount: number;
		autoResumeCount: number;
	};
};

export type QuotaUsageRuntime = {
	provider: string;
	cost: number;
	requests: number;
};

export type QuotaInfoRuntime = {
	tier?: string;
	limit?: number | null;
	used?: number;
	remaining?: number | null;
	resetDate?: string | null;
	rateLimitRpm?: number | null;
	/** @see ProviderAuthTruth */
	authTruth?: string;
	/** @see QuotaDataConfidence */
	quotaConfidence?: string;
	/** ISO-8601 timestamp of last real-time fetch from provider, or null. */
	quotaRefreshedAt?: string | null;
};

export type QuotaServiceRuntime = {
	getUsageByModel: () => QuotaUsageRuntime[];
	getQuota?: (provider: string) => QuotaInfoRuntime | undefined;
};

export type ModelInfoRuntime = {
	id: string;
	provider: string;
	model?: string;
	name?: string;
	inputPrice?: number | null;
	outputPrice?: number | null;
	contextWindow?: number | null;
	tier?: string;
	recommended?: boolean;
	reason?: string;
};

export type ModelSelectorRuntime = {
	getQuotaService: () => QuotaServiceRuntime;
	getAvailableModels?: () => ModelInfoRuntime[];
	getFallbackChain?: (options?: {
		routingTaskType?: string;
	}) => ModelInfoRuntime[];
	getDepletedModels?: () => Array<{
		key: string;
		provider: string;
		modelId: string;
		depletedAt: number;
		retryAfter: number;
		isPermanent: boolean;
		coolsDownAt: string;
	}>;
};

export type CostStatsRuntime = {
	estimatedCostUSD: number;
	totalRequests?: number;
	dailyHistory?: Array<{ date: string; cost: number; requests: number }>;
};

export type LLMServiceRuntime = {
	getCostStats: () => CostStatsRuntime;
	modelSelector: ModelSelectorRuntime;
	generate?: (
		prompt: string,
		options?: Record<string, unknown>,
	) => Promise<{ text?: string; latencyMs?: number }>;
	generateText?: (
		provider: string,
		model: string,
		systemPrompt: string,
		prompt: string,
		options?: Record<string, unknown>,
	) => Promise<{ content?: string; text?: string; latencyMs?: number }>;
};

export type ResearchServiceRuntime = {
	research: (topic: string, depth?: number) => Promise<unknown>;
	ingest: (url: string) => Promise<unknown>;
};

export type KnowledgeGraphResponseRuntime = {
	content: Array<{ type: string; text: string }>;
};

export type KnowledgeServiceRuntime = {
	getGraph: (
		query?: string,
		depth?: number,
	) => Promise<KnowledgeGraphResponseRuntime>;
};

export type DeepResearchServiceRuntime = {
	ingest: (url: string) => Promise<string>;
	recursiveResearch?: (
		topic: string,
		depth?: number,
		maxBreadth?: number,
	) => Promise<unknown>;
	generateQueries?: (topic: string) => Promise<string[]>;
};

export type ProjectTrackerRuntime = {
	getStatus: () => {
		taskId: string;
		status: string;
		progress: number;
		currentTask: string;
	};
};

export type LspServiceRuntime = {
	getStatus: () => unknown;
	findSymbol: (
		filePath: string,
		symbolName: string,
	) => Promise<unknown> | unknown;
	findReferences: (
		filePath: string,
		line: number,
		character: number,
	) => Promise<unknown> | unknown;
	getSymbols: (filePath: string) => Promise<unknown> | unknown;
	searchSymbols: (query: string) => Promise<unknown> | unknown;
	indexProject: () => Promise<void> | void;
};

export type MCPSimpleServerRuntime = {
	name?: string;
	status?: string;
	toolCount?: number;
	advertisedToolCount?: number;
	advertisedAlwaysOn?: boolean;
	advertisedSource?: "database" | "config" | "empty";
	warmupStatus?: "idle" | "scheduled" | "warming" | "ready" | "failed";
	lastConnectedAt?: number;
	lastError?: string;
	tools?: unknown[];
	command?: string;
	args?: string[];
	config?: {
		command?: string;
		args?: string[];
		env?: Record<string, string>;
	};
};

export type MCPAggregatedToolRuntime = {
	name: string;
	description?: string;
	server?: string;
	serverDisplayName?: string;
	advertisedName?: string;
	serverTags?: string[];
	toolTags?: string[];
	semanticGroup?: string;
	semanticGroupLabel?: string;
	keywords?: string[];
	alwaysOn?: boolean;
	inputSchema?: unknown;
};

export type MCPAggregatorRuntime = {
	listServers?: () => Promise<MCPSimpleServerRuntime[]>;
	listAggregatedTools?: () => Promise<MCPAggregatedToolRuntime[]>;
	searchTools?: (query: string) => Promise<MCPAggregatedToolRuntime[]>;
	getTrafficEvents?: () => unknown[];
	addServerConfig?: (
		name: string,
		config: {
			command: string;
			args: string[];
			env?: Record<string, string>;
			enabled?: boolean;
		},
	) => Promise<void>;
	removeServerConfig?: (name: string) => Promise<void>;
	getInitializationStatus?: () => {
		inProgress: boolean;
		initialized: boolean;
		lastStartedAt?: number;
		lastCompletedAt?: number;
		lastSuccessAt?: number;
		lastError?: string;
		connectedClientCount: number;
		configuredServerCount: number;
	};
	clients?: Map<string, { close?: () => Promise<void> | void }>;
	configPath?: string;
};

export type McpConfigServiceRuntime = {
	getStatus: () => {
		inProgress: boolean;
		lastStartedAt?: number;
		lastCompletedAt?: number;
		lastSuccessAt?: number;
		lastError?: string;
		lastServerCount: number;
		lastToolCount: number;
	};
};

export type SubmoduleServiceRuntime = {
	listSubmodules: () => Promise<unknown[]>;
	updateAll: () => Promise<{ success: boolean; output: string }>;
	installDependencies: (
		submodulePath: string,
	) => Promise<{ success: boolean; output: string }>;
	buildSubmodule: (
		submodulePath: string,
	) => Promise<{ success: boolean; output: string }>;
	enableSubmodule: (
		submodulePath: string,
	) => Promise<{ success: boolean; output: string }>;
	detectCapabilities: (submodulePath: string) => {
		caps: string[];
		startCommand?: string;
	};
};

export type SkillRuntime = {
	id: string;
	name: string;
	description: string;
	content: string;
	path: string;
};

export type SkillRegistryRuntime = {
	getSkills: () => SkillRuntime[];
	readSkill: (skillName: string) => Promise<unknown>;
	createSkill: (
		id: string,
		name: string,
		description: string,
	) => Promise<unknown>;
	saveSkill: (id: string, content: string) => Promise<unknown>;
};

export type SkillAssimilationServiceRuntime = {
	assimilate: (request: {
		topic: string;
		docsUrl?: string;
		autoInstall?: boolean;
	}) => Promise<{ success: boolean; toolName?: string; logs: string[] }>;
};

export type BrowserServiceRuntime = {
	getStatus: () => { active: boolean; pageCount: number; pageIds: string[] };
	close: (pageId: string) => Promise<void>;
	closeAll: () => Promise<void>;
};

/*
export type MeshServiceRuntime = {
    nodeId: string;
    getStatus: () => { nodeId: string; peerCount: number; peerIds: string[] };
    broadcast: (type: string, payload: unknown) => void;
};
*/

export type MarketplaceEntryRuntime = {
	id: string;
	name: string;
	description: string;
	author?: string;
	type: "agent" | "tool" | "skill";
	source: "official" | "community" | "local";
	url?: string;
	verified: boolean;
	peerCount: number;
	installed: boolean;
	tags: string[];
};

export type MarketplaceServiceRuntime = {
	list: (filter?: string) => Promise<MarketplaceEntryRuntime[]>;
	install: (id: string) => Promise<string>;
	publish: (manifest: Partial<MarketplaceEntryRuntime>) => Promise<string>;
};

export const t = initTRPC.create();

/**
 * Reason: strict catch variables are `unknown`, so direct `.message` access is unsafe.
 * What: narrows unknown error values into a stable string message.
 * Why: preserves existing authorization branching while removing broad catch typing.
 */
function getErrorMessage(error: unknown): string {
	if (error instanceof Error) {
		return error.message;
	}

	if (typeof error === "string") {
		return error;
	}

	return "";
}

// Mock RBAC Middleware
const isAdmin = t.middleware(async ({ next, ctx }) => {
	// In a real app, check ctx.user.role via JWT/Session
	// For local desktop app, we default to ADMIN unless config says 'demo_mode'
	try {
		const config = getMcpServer().directorConfig;
		if (config?.demo_mode) {
			throw new Error("UNAUTHORIZED: Demo Mode enabled. Action restricted.");
		}
	} catch (error: unknown) {
		// If server not initialized, strict mode might block, or we allow.
		// Logic: If error is "UNAUTHORIZED", rethrow.
		const message = getErrorMessage(error);
		if (message.startsWith("UNAUTHORIZED")) throw error;
		// Else ignore (server missing)
	}
	return next();
});

export const publicProcedure = t.procedure;
export const adminProcedure = t.procedure.use(isAdmin);

/**
 * Typed accessor for the global MCPServer instance.
 * Uses the global declaration from MCPServer.ts.
 */
export function getMcpServer(): MCPServerRuntime {
	const server = global.mcpServerInstance;
	if (server) return server as unknown as MCPServerRuntime;
	// Return empty object — callers must check for undefined properties
	return {} as unknown as MCPServerRuntime;
}

export function isMcpReady(): boolean {
	return !!global.mcpServerInstance;
}

/**
 * Safely invoke a service method, returning a fallback when MCP isn't initialized.
 * Usage: safeServiceCall(() => getService().method(), fallbackValue)
 */
export function safeServiceCall<T>(fn: () => T, fallback: T): T {
	try {
		const result = fn();
		return (result === undefined || result === null) ? fallback : result;
	} catch {
		return fallback;
	}
}

export function getWorkflowEngine(): WorkflowEngineRuntime | null {
	const engine = getMcpServer().workflowEngine;
	return (engine ?? null) as WorkflowEngineRuntime | null;
}

export function getWorkflowDefinitions(
	engine: WorkflowEngineRuntime,
): WorkflowDefinitionRuntime[] {
	const maybeMap = (
		engine as unknown as { workflows?: Map<string, WorkflowDefinitionRuntime> }
	).workflows;
	if (!maybeMap) return [];
	return Array.from(maybeMap.values());
}

export function getAutoTestService(): AutoTestServiceRuntime {
	return getMcpServer().autoTestService as AutoTestServiceRuntime;
}

export function getHealerService(): HealerServiceRuntime {
	return getMcpServer().healerService as HealerServiceRuntime;
}

export function getGitService(): GitServiceRuntime {
	return getMcpServer().gitService as GitServiceRuntime;
}

export function getPermissionManager(): PermissionManagerRuntime {
	return getMcpServer().permissionManager as PermissionManagerRuntime;
}

export function getDirectorRuntime(): DirectorRuntime {
	return getMcpServer().director as DirectorRuntime;
}

export function getAuditService(): AuditServiceRuntime {
	return getMcpServer().auditService as AuditServiceRuntime;
}

export function getDarwinService(): DarwinServiceRuntime {
	return getMcpServer().darwinService as DarwinServiceRuntime;
}

export function getCouncilOrchestrator(): CouncilOrchestratorRuntime {
	return getMcpServer().council as unknown as CouncilOrchestratorRuntime;
}

export function getCouncilService(): CouncilServiceRuntime {
	return getMcpServer().councilService as CouncilServiceRuntime;
}

export function getSupervisorCouncil(): Promise<unknown> {
	// This returns the migrated SupervisorCouncil singleton
	return import("../orchestrator/council/services/council.js").then(
		(m) => m.council,
	);
}

export function getCouncilSessionManager(): Promise<unknown> {
	return import("../orchestrator/council/services/session-manager.js").then(
		(m) => m.sessionManager,
	);
}

export function getCouncilWsManager(): Promise<unknown> {
	return import("../orchestrator/council/services/ws-manager.js").then(
		(m) => m.wsManager,
	);
}

export function getCouncilHierarchy(): Promise<unknown> {
	return import("../orchestrator/council/services/council-hierarchy.js").then(
		(m) => m.councilHierarchy,
	);
}

export function getMemoryManager(): MemoryManagerRuntime {
	return getMcpServer().memoryManager as MemoryManagerRuntime;
}

export function getAgentMemoryService(): AgentMemoryServiceRuntime | undefined {
	return getMcpServer().agentMemoryService as
		| AgentMemoryServiceRuntime
		| undefined;
}

export function getSessionImportService():
	| SessionImportServiceRuntime
	| undefined {
	return getMcpServer().sessionImportService as
		| SessionImportServiceRuntime
		| undefined;
}

export function getSymbolPinService(): SymbolPinServiceRuntime | undefined {
	return getMcpServer().symbolPinService as SymbolPinServiceRuntime | undefined;
}

export function getConfigManager(): ConfigManagerRuntime | undefined {
	return getMcpServer().configManager as ConfigManagerRuntime | undefined;
}

export function getCommandRegistry(): CommandRegistryRuntime {
	return getMcpServer().commandRegistry as CommandRegistryRuntime;
}

export function getContextManager(): ContextManagerRuntime | undefined {
	return getMcpServer().contextManager as ContextManagerRuntime | undefined;
}

export function getAutoDevService(): AutoDevServiceRuntime | undefined {
	return getMcpServer().autoDevService as AutoDevServiceRuntime | undefined;
}

export function getShellService(): ShellServiceRuntime | undefined {
	return getMcpServer().shellService as ShellServiceRuntime | undefined;
}

export function getEventBus(): EventBusRuntime | undefined {
	return getMcpServer().eventBus as EventBusRuntime | undefined;
}

export function getSuggestionService(): SuggestionServiceRuntime {
	return getMcpServer().suggestionService as SuggestionServiceRuntime;
}

export function getSquadService(): SquadServiceRuntime | undefined {
	return getMcpServer().squadService as SquadServiceRuntime | undefined;
}

export function getSessionManager(): SessionManagerRuntime {
	return getMcpServer().sessionManager as SessionManagerRuntime;
}

export function getSessionSupervisor(): SessionSupervisorRuntime {
	return getMcpServer().sessionSupervisor as SessionSupervisorRuntime;
}

export function getLLMService(): LLMServiceRuntime {
	return getMcpServer().llmService as LLMServiceRuntime;
}

export function getKnowledgeService(): KnowledgeServiceRuntime {
	return getMcpServer().knowledgeService as KnowledgeServiceRuntime;
}

export function getDeepResearchService(): DeepResearchServiceRuntime {
	return getMcpServer().deepResearchService as DeepResearchServiceRuntime;
}

export function getResearchService(): ResearchServiceRuntime | undefined {
	return getMcpServer().researchService as ResearchServiceRuntime | undefined;
}

export function getProjectTracker(): ProjectTrackerRuntime | undefined {
	return getMcpServer().projectTracker as ProjectTrackerRuntime | undefined;
}

export function getLspService(): LspServiceRuntime | undefined {
	return getMcpServer().lspService as LspServiceRuntime | undefined;
}

export function getMcpAggregator(): MCPAggregatorRuntime | undefined {
	return getMcpServer().mcpAggregator as MCPAggregatorRuntime | undefined;
}

export function getSubmoduleService(): SubmoduleServiceRuntime | undefined {
	return getMcpServer().submoduleService as SubmoduleServiceRuntime | undefined;
}

export function getSkillRegistry(): SkillRegistryRuntime | undefined {
	return getMcpServer().skillRegistry as SkillRegistryRuntime | undefined;
}

export function getSkillAssimilationService():
	| SkillAssimilationServiceRuntime
	| undefined {
	return getMcpServer().skillAssimilationService as
		| SkillAssimilationServiceRuntime
		| undefined;
}

export function getBrowserService(): BrowserServiceRuntime | undefined {
	return getMcpServer().browserService as BrowserServiceRuntime | undefined;
}

export function getMcpConfigService(): McpConfigServiceRuntime | undefined {
	return getMcpServer().mcpConfigService as McpConfigServiceRuntime | undefined;
}

/*
export function getMeshService(): MeshServiceRuntime | undefined {
    return getMcpServer().meshService as MeshServiceRuntime | undefined;
}
*/

export function getMarketplaceService(): MarketplaceServiceRuntime | undefined {
	return getMcpServer().marketplaceService as
		| MarketplaceServiceRuntime
		| undefined;
}
