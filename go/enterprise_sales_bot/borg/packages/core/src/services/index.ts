/**
 * Services Index - Exports all core services for easy import
 */

// Core Infrastructure Services (Phase 51)
export { LSPService, LanguageServer } from './LSPService.js';
export type {
    LanguageServerConfig,
    Position,
    Range,
    Location,
    SymbolInformation,
    DocumentSymbol,
    SymbolKind,
    LSPMessage,
} from './LSPService.js';

export { PlanService, DiffSandbox } from './PlanService.js';
export type {
    PlanMode,
    FileDiff,
    DiffHunk,
    DiffSandboxState,
    Checkpoint,
    PlanServiceOptions,
} from './PlanService.js';

export { CodeModeService, ToolRegistry, CodeExecutor, createDefaultTools } from './CodeModeService.js';
export type {
    ToolFunction,
    ToolDefinition,
    ExecutionResult,
    CodeModeOptions,
} from './CodeModeService.js';

// Existing Services
export { AutoDevService } from './AutoDevService.js';
export { AutoTestService } from './AutoTestService.js';
export { GitService } from './GitService.js';
export { HealerService } from './HealerService.js';
export { KnowledgeService } from './KnowledgeService.js';
export { MemoryManager } from './MemoryManager.js';
export { MetricsService } from './MetricsService.js';
export { RepoGraphService } from './RepoGraphService.js';
export { ResearchService } from './ResearchService.js';
export { ShellService } from './ShellService.js';
export { SymbolPinService } from './SymbolPinService.js';

// Phase 53: Agent Memory
export { AgentMemoryService } from './AgentMemoryService.js';
export type {
    Memory,
    MemoryType,
    MemoryNamespace,
    MemoryMetadata,
    MemorySearchOptions,
    MemoryServiceOptions,
} from './AgentMemoryService.js';

// Phase 80: Swarm Missions
export { MissionService } from './MissionService.js';
export type { SwarmMission } from './MissionService.js';
