import type { MemoryMetadata } from './AgentMemoryService.js';
import type { ToolContextPayload } from './toolContextMemory.js';

const TOOL_CONTEXT_EXCLUDED_NAMES = new Set([
    'search_tools',
    'load_tool',
    'unload_tool',
    'get_tool_schema',
    'list_loaded_tools',
    'list_all_tools',
    'get_tool_context',
    'memory_stats',
    'add_memory',
    'search_memory',
    'get_recent_memories',
    'clear_session_memory',
    'router_status',
    'set_autonomy',
    'plan_status',
    'code_mode_status',
]);

function sanitizeLine(value: string, maxLength: number): string {
    return value.replace(/\s+/g, ' ').trim().slice(0, maxLength);
}

export function shouldResolveAutomaticToolContext(toolName: string): boolean {
    return !TOOL_CONTEXT_EXCLUDED_NAMES.has(toolName);
}

export function shouldPersistAutomaticToolContext(payload: ToolContextPayload | null | undefined): payload is ToolContextPayload {
    if (!payload) {
        return false;
    }

    return payload.observationCount > 0 || payload.summaryCount > 0 || payload.matchedPaths.length > 0;
}

export function buildAutomaticToolContextFingerprint(payload: ToolContextPayload): string {
    return JSON.stringify({
        toolName: payload.toolName,
        query: payload.query,
        matchedPaths: payload.matchedPaths,
        observationCount: payload.observationCount,
        summaryCount: payload.summaryCount,
    });
}

export function buildAutomaticToolContextPreview(payload: ToolContextPayload): string {
    const parts = [
        payload.observationCount > 0 ? `${payload.observationCount} observation${payload.observationCount === 1 ? '' : 's'}` : null,
        payload.summaryCount > 0 ? `${payload.summaryCount} summary${payload.summaryCount === 1 ? '' : 'summaries'}` : null,
        payload.matchedPaths.length > 0 ? `${payload.matchedPaths.length} file hit${payload.matchedPaths.length === 1 ? '' : 's'}` : null,
    ].filter((value): value is string => Boolean(value));

    if (parts.length === 0) {
        return 'No relevant JIT tool context';
    }

    return `JIT context: ${parts.join(' · ')}`;
}

export function buildAutomaticToolContextMemory(payload: ToolContextPayload): {
    content: string;
    metadata: MemoryMetadata;
} {
    return {
        content: payload.prompt,
        metadata: {
            source: 'pre_tool_context',
            tags: ['tool-context', 'pre-tool', payload.toolName],
            toolName: payload.toolName,
            query: payload.query,
            matchedPaths: payload.matchedPaths,
            observationCount: payload.observationCount,
            summaryCount: payload.summaryCount,
            preview: buildAutomaticToolContextPreview(payload),
        },
    };
}

export function buildAutomaticToolContextStartPayload(payload: ToolContextPayload | null): {
    contextPreview?: string;
    contextMatchedPaths?: string[];
    contextObservationCount?: number;
    contextSummaryCount?: number;
} {
    if (!payload || !shouldPersistAutomaticToolContext(payload)) {
        return {};
    }

    return {
        contextPreview: sanitizeLine(buildAutomaticToolContextPreview(payload), 120),
        contextMatchedPaths: payload.matchedPaths.slice(0, 4),
        contextObservationCount: payload.observationCount,
        contextSummaryCount: payload.summaryCount,
    };
}
