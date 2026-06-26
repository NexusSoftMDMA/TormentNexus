import type { ObservationInput } from './AgentMemoryService.js';

export interface ToolObservationEvent {
    toolName: string;
    args?: unknown;
    result?: unknown;
    error?: unknown;
    durationMs: number;
}

const LOW_SIGNAL_TOOL_NAMES = new Set([
    'router_status',
    'set_autonomy',
    'chat_reply',
    'chat_submit',
    'click_element',
    'click_at',
    'native_input',
    'vscode_execute_command',
    'vscode_get_status',
    'vscode_get_notifications',
    'vscode_submit_chat',
    'memory_stats',
    'add_memory',
    'search_memory',
    'get_recent_memories',
    'clear_session_memory',
    'list_loaded_tools',
    'list_all_tools',
    'load_tool',
    'unload_tool',
    'get_tool_schema',
    'get_tool_context',
    'search_tools',
    'code_mode_status',
    'list_code_tools',
    'plan_status',
    'list_workflows',
    'workflow_status',
    'list_agents',
    'list_squads',
    'git_worktree_list',
    'system_status',
    'get_status',
]);

const READ_TOOL_PATTERNS = [
    /(^|_)(read|view|search|grep|find|list|get|query|inspect|diagnose|extract|fetch|research|navigate)(_|$)/i,
    /^browser_/i,
    /^lsp_/i,
];

const WRITE_TOOL_PATTERNS = [
    /(^|_)(write|create|update|rename|edit|apply|patch|delete|remove|save|merge|rollback|approve|fix)(_|$)/i,
    /^git_worktree_(add|remove)$/i,
    /^propose_change$/i,
];

function sanitizeText(value: string, maxLength: number): string {
    return value.replace(/\s+/g, ' ').trim().slice(0, maxLength);
}

function safeJsonStringify(value: unknown): string {
    const seen = new WeakSet<object>();

    return JSON.stringify(value, (_key, currentValue) => {
        if (typeof currentValue === 'object' && currentValue !== null) {
            if (seen.has(currentValue)) {
                return '[Circular]';
            }
            seen.add(currentValue);
        }

        return currentValue;
    }, 2);
}

function normalizeText(value: unknown, maxLength: number = 800): string {
    if (value instanceof Error) {
        return sanitizeText(value.message, maxLength);
    }

    if (typeof value === 'string') {
        return sanitizeText(value, maxLength);
    }

    if (value == null) {
        return '';
    }

    return sanitizeText(safeJsonStringify(value), maxLength);
}

function uniqueStrings(values: Array<string | undefined | null>, maxItems: number = 12): string[] {
    const seen = new Set<string>();
    const normalized: string[] = [];

    for (const value of values) {
        if (typeof value !== 'string') continue;
        const trimmed = value.trim();
        if (!trimmed) continue;
        const key = trimmed.toLowerCase();
        if (seen.has(key)) continue;
        seen.add(key);
        normalized.push(trimmed);
        if (normalized.length >= maxItems) break;
    }

    return normalized;
}

function extractTextFromContentArray(value: unknown): string[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value
        .map((entry) => {
            if (!entry || typeof entry !== 'object') {
                return null;
            }
            const text = Reflect.get(entry as object, 'text');
            return typeof text === 'string' ? sanitizeText(text, 400) : null;
        })
        .filter((text): text is string => Boolean(text));
}

function summarizeToolResult(result: unknown): string {
    if (!result || typeof result !== 'object') {
        return normalizeText(result, 600);
    }

    const content = Reflect.get(result as object, 'content');
    const textBlocks = extractTextFromContentArray(content);
    if (textBlocks.length > 0) {
        return sanitizeText(textBlocks.join(' '), 600);
    }

    const message = Reflect.get(result as object, 'message');
    if (typeof message === 'string') {
        return sanitizeText(message, 600);
    }

    return normalizeText(result, 600);
}

function normalizePathCandidate(value: unknown): string[] {
    if (typeof value === 'string') {
        const trimmed = value.trim();
        return trimmed ? [trimmed.replace(/\\/g, '/')] : [];
    }

    if (Array.isArray(value)) {
        return value.flatMap((item) => normalizePathCandidate(item));
    }

    return [];
}

function extractPathCandidates(args: unknown): string[] {
    if (!args || typeof args !== 'object') {
        return [];
    }

    const record = args as Record<string, unknown>;
    return uniqueStrings([
        ...normalizePathCandidate(record.path),
        ...normalizePathCandidate(record.file),
        ...normalizePathCandidate(record.filePath),
        ...normalizePathCandidate(record.AbsolutePath),
        ...normalizePathCandidate(record.absolutePath),
        ...normalizePathCandidate(record.targetPath),
        ...normalizePathCandidate(record.old_path),
        ...normalizePathCandidate(record.oldPath),
        ...normalizePathCandidate(record.new_path),
        ...normalizePathCandidate(record.newPath),
        ...normalizePathCandidate(record.dirPath),
        ...normalizePathCandidate(record.paths),
        ...normalizePathCandidate(record.files),
    ], 20);
}

function inferFileTouches(toolName: string, args: unknown): { filesRead: string[]; filesModified: string[] } {
    const candidates = extractPathCandidates(args);
    const isReadTool = READ_TOOL_PATTERNS.some((pattern) => pattern.test(toolName));
    const isWriteTool = WRITE_TOOL_PATTERNS.some((pattern) => pattern.test(toolName));

    return {
        filesRead: isReadTool ? candidates : [],
        filesModified: isWriteTool ? candidates : [],
    };
}

export function shouldCaptureToolObservation(toolName: string): boolean {
    return !LOW_SIGNAL_TOOL_NAMES.has(toolName);
}

export function buildToolObservationInput(event: ToolObservationEvent): ObservationInput | null {
    if (!shouldCaptureToolObservation(event.toolName)) {
        return null;
    }

    const failed = event.error != null;
    const resultSummary = failed
        ? normalizeText(event.error, 600)
        : summarizeToolResult(event.result);
    const { filesRead, filesModified } = inferFileTouches(event.toolName, event.args);

    const facts = uniqueStrings([
        `Tool: ${event.toolName}`,
        `Duration: ${event.durationMs}ms`,
        filesRead.length > 0 ? `Files read: ${filesRead.join(', ')}` : null,
        filesModified.length > 0 ? `Files modified: ${filesModified.join(', ')}` : null,
        failed && resultSummary ? `Failure: ${resultSummary}` : null,
    ], 5);

    const narrative = failed
        ? `${event.toolName} failed after ${event.durationMs}ms. ${resultSummary}`.trim()
        : `${event.toolName} completed in ${event.durationMs}ms. ${resultSummary}`.trim();

    return {
        toolName: event.toolName,
        title: failed ? `${event.toolName} failed` : `${event.toolName} completed`,
        narrative: sanitizeText(narrative, 600),
        rawInput: event.args,
        rawOutput: failed ? event.error : event.result,
        facts,
        concepts: uniqueStrings([
            'tool-execution',
            event.toolName,
            failed ? 'failure' : 'success',
            ...event.toolName.split(/[^a-zA-Z0-9]+/).map((token) => token.toLowerCase()),
        ], 8),
        filesRead,
        filesModified,
        type: failed ? 'warning' : undefined,
        metadata: {
            source: 'tool_execution',
            tags: uniqueStrings(['tool-execution', event.toolName, failed ? 'failure' : 'success'], 8),
            durationMs: event.durationMs,
            success: !failed,
        },
    };
}
