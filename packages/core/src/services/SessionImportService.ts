import crypto from 'crypto';
import fs from 'fs/promises';
import os from 'os';
import path from 'path';

import Database from 'better-sqlite3';
import fg from 'fast-glob';

import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@tormentnexus/ai';
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from '../db/sqliteAvailability.js';

import AgentMemoryService from './AgentMemoryService.js';
import {
    ImportedSessionStore,
    type ImportedSessionMemoryInput,
    type ImportedSessionMemoryKind,
    type ImportedSessionMemorySource,
    type ImportedSessionMaintenanceStats,
    type ImportedSessionRecord,
} from './ImportedSessionStore.js';

interface SessionImportServiceOptions {
    store?: ImportedSessionStore;
    includeHomeDirectories?: boolean;
    importIntervalMs?: number;
    maxFilesPerRoot?: number;
}

interface SessionDiscoveryRule {
    sourceTool: string;
    roots: string[];
    filePatterns: string[];
    fileNameHints?: string[];
    importAllFiles?: boolean;
    ignoredPathHints?: string[];
}

interface DiscoveryCandidate {
    sourceTool: string;
    sourcePath: string;
    sessionFormat: string;
    lastModifiedAt: number | null;
    externalSessionId?: string | null;
    transcript?: string;
    title?: string;
    workingDirectory?: string | null;
    metadata?: Record<string, unknown>;
}

interface NormalizedImportedSession {
    sessionId: string;
    title: string;
    sourceTool: string;
    sourcePath: string;
    sessionFormat: string;
    externalSessionId: string | null;
    workingDirectory: string;
    transcript: string;
    excerpt: string;
    transcriptHash: string;
    discoveredAt: number;
    importedAt: number;
    lastModifiedAt: number | null;
    normalizedSession: Record<string, unknown>;
    metadata: Record<string, unknown>;
}

interface ImportedSessionRetentionSummary {
    strategy: ImportedSessionMemorySource;
    transcriptLength: number;
    analyzedChars: number;
    durableMemoryCount: number;
    durableInstructionCount: number;
    archiveDisposition: 'archive_only';
    summary: string;
    salientTags: string[];
    keepArchivedCategories: string[];
    discardableCategories: string[];
}

interface ImportedSessionAnalysis {
    memories: ImportedSessionMemoryInput[];
    retentionSummary: ImportedSessionRetentionSummary;
}

interface ImportedInstructionDoc {
    path: string;
    updatedAt: number;
    size: number;
}

export interface SessionImportSummary {
    discoveredCount: number;
    importedCount: number;
    skippedCount: number;
    storedMemoryCount: number;
    instructionDocPath: string | null;
    tools: string[];
}

const IMPORTABLE_EXTENSIONS = new Set(['.md', '.mdx', '.txt', '.log', '.json', '.jsonl']);
const DISCOVERY_HINTS = ['session', 'sessions', 'chat', 'conversation', 'transcript', 'history', 'prompt', 'messages', 'handoff'];
const DEFAULT_SCAN_INTERVAL_MS = 15 * 60 * 1000;
const DEFAULT_MAX_FILES_PER_ROOT = 200;

function uniqueStrings(values: Array<string | undefined | null>, maxItems?: number): string[] {
    const seen = new Set<string>();
    const normalized: string[] = [];

    for (const value of values) {
        if (typeof value !== 'string') continue;
        const trimmed = value.replace(/\s+/g, ' ').trim();
        if (!trimmed) continue;
        const key = trimmed.toLowerCase();
        if (seen.has(key)) continue;
        seen.add(key);
        normalized.push(trimmed);

        if (typeof maxItems === 'number' && normalized.length >= maxItems) {
            break;
        }
    }

    return normalized;
}

function sanitizeSentence(value: string, maxLength: number = 220): string {
    return value.replace(/\s+/g, ' ').trim().slice(0, maxLength);
}

function extractJsonTextFragments(value: unknown): string[] {
    if (typeof value === 'string') {
        return [value];
    }

    if (Array.isArray(value)) {
        return value.flatMap(extractJsonTextFragments);
    }

    if (!value || typeof value !== 'object') {
        return [];
    }

    const record = value as Record<string, unknown>;
    const prioritizedKeys = [
        'content',
        'text',
        'message',
        'prompt',
        'response',
        'request',
        'input',
        'output',
        'result',
        'params',
        'arguments',
        'body',
        'summary',
        'title',
        'parts',
    ];
    const collected: string[] = [];

    for (const key of prioritizedKeys) {
        if (key in record) {
            collected.push(...extractJsonTextFragments(record[key]));
        }
    }

    for (const nestedKey of ['messages', 'conversation', 'transcript', 'entries', 'events', 'turns', 'items']) {
        if (nestedKey in record) {
            collected.push(...extractJsonTextFragments(record[nestedKey]));
        }
    }

    return collected;
}

type OpenAiLikeMessage = {
    role?: unknown;
    content?: unknown;
    tool_calls?: unknown;
    created_at?: unknown;
    timestamp?: unknown;
};

type ChatGptExportMessage = {
    id?: unknown;
    author?: unknown;
    create_time?: unknown;
    content?: unknown;
    metadata?: unknown;
};

type ChatGptExportMappingNode = {
    id?: unknown;
    parent?: unknown;
    children?: unknown;
    message?: unknown;
};

function extractOpenAiLikeMessages(value: unknown): OpenAiLikeMessage[] {
    if (Array.isArray(value)) {
        return value.filter((entry): entry is OpenAiLikeMessage => Boolean(entry) && typeof entry === 'object');
    }

    if (!value || typeof value !== 'object') {
        return [];
    }

    const record = value as Record<string, unknown>;
    for (const key of ['messages', 'conversation', 'transcript', 'entries', 'items']) {
        const nested = record[key];
        if (Array.isArray(nested)) {
            return nested.filter((entry): entry is OpenAiLikeMessage => Boolean(entry) && typeof entry === 'object');
        }
    }

    return [];
}

function normalizeChatGptExportMessage(value: unknown): ChatGptExportMessage | null {
    if (!value || typeof value !== 'object') {
        return null;
    }

    return value as ChatGptExportMessage;
}

function extractChatGptExportMessages(value: unknown): ChatGptExportMessage[] {
    if (!value || typeof value !== 'object') {
        return [];
    }

    const record = value as Record<string, unknown>;
    const mapping = record.mapping;
    if (!mapping || typeof mapping !== 'object' || Array.isArray(mapping)) {
        return [];
    }

    const nodes = mapping as Record<string, ChatGptExportMappingNode>;
    const currentNode = typeof record.current_node === 'string' && record.current_node.trim()
        ? record.current_node
        : null;

    if (currentNode) {
        const ordered: ChatGptExportMessage[] = [];
        const visited = new Set<string>();
        let cursor: string | null = currentNode;

        while (cursor && !visited.has(cursor)) {
            visited.add(cursor);
            const node: ChatGptExportMappingNode | undefined = nodes[cursor];
            if (!node || typeof node !== 'object') {
                break;
            }

            const message = normalizeChatGptExportMessage(node.message);
            if (message) {
                ordered.push(message);
            }

            cursor = typeof node.parent === 'string' && node.parent.trim()
                ? node.parent
                : null;
        }

        return ordered.reverse();
    }

    return Object.values(nodes)
        .map((node) => normalizeChatGptExportMessage(node?.message))
        .filter((message): message is ChatGptExportMessage => Boolean(message))
        .sort((left, right) => {
            const leftTime = toTimestampMs(left.create_time) ?? 0;
            const rightTime = toTimestampMs(right.create_time) ?? 0;
            return leftTime - rightTime;
        });
}

function formatChatGptExportTranscript(value: unknown): string | null {
    const messages = extractChatGptExportMessages(value);
    if (messages.length === 0) {
        return null;
    }

    const lines: string[] = [];

    for (const message of messages) {
        const author = message.author && typeof message.author === 'object'
            ? message.author as Record<string, unknown>
            : null;
        const role = typeof author?.role === 'string' && author.role.trim()
            ? author.role.trim().toLowerCase()
            : 'user';

        if (role !== 'user' && role !== 'assistant' && role !== 'system') {
            continue;
        }

        const content = uniqueStrings(extractJsonTextFragments(message.content), 20).join('\n').trim();
        if (!content) {
            continue;
        }

        const label = role === 'assistant'
            ? 'Assistant'
            : role === 'system'
                ? 'System'
                : 'User';
        lines.push(`${label}: ${content}`);
    }

    return lines.length > 0 ? lines.join('\n\n') : null;
}

function formatOpenAiLikeTranscript(value: unknown): string | null {
    const chatGptExportTranscript = formatChatGptExportTranscript(value);
    if (chatGptExportTranscript) {
        return chatGptExportTranscript;
    }

    const messages = extractOpenAiLikeMessages(value);
    if (messages.length === 0) {
        return null;
    }

    const lines: string[] = [];

    for (const message of messages) {
        const role = typeof message.role === 'string' && message.role.trim()
            ? message.role.trim().toLowerCase()
            : 'user';

        if (role !== 'user' && role !== 'assistant') {
            continue;
        }

        const content = uniqueStrings(extractJsonTextFragments(message.content), 20).join('\n').trim();
        const toolCalls = Array.isArray(message.tool_calls)
            ? message.tool_calls
                .map((entry) => {
                    if (!entry || typeof entry !== 'object') return null;
                    const toolCall = entry as Record<string, unknown>;
                    const toolFunction = toolCall.function;
                    if (!toolFunction || typeof toolFunction !== 'object') return null;
                    const name = typeof (toolFunction as Record<string, unknown>).name === 'string'
                        ? (toolFunction as Record<string, unknown>).name
                        : null;
                    return name ? `[Tool Use: ${name}]` : null;
                })
                .filter((entry): entry is string => Boolean(entry))
            : [];

        const fragments = [content, ...toolCalls].filter((entry) => entry.trim().length > 0);
        if (fragments.length === 0) {
            continue;
        }

        const label = role === 'assistant' ? 'Assistant' : 'User';
        lines.push(`${label}: ${fragments.join('\n')}`);
    }

    return lines.length > 0 ? lines.join('\n\n') : null;
}

function parseTranscriptContent(filePath: string, content: string): string {
    const extension = path.extname(filePath).toLowerCase();

    if (extension === '.jsonl') {
        const lines = content
            .split(/\r?\n/)
            .map((line) => line.trim())
            .filter(Boolean);

        const extracted = lines.flatMap((line) => {
            try {
                return extractJsonTextFragments(JSON.parse(line));
            } catch {
                return [line];
            }
        });

        return uniqueStrings(extracted, 400).join('\n');
    }

    if (extension === '.json') {
        try {
            const parsed = JSON.parse(content);
            const openAiLikeTranscript = formatOpenAiLikeTranscript(parsed);
            if (openAiLikeTranscript) {
                return openAiLikeTranscript;
            }
            const extracted = uniqueStrings(extractJsonTextFragments(parsed), 400);
            if (extracted.length > 0) {
                return extracted.join('\n');
            }
        } catch {
            return content;
        }
    }

    return content;
}

function classifyMemoryKind(content: string): ImportedSessionMemoryKind {
    const lowered = content.toLowerCase();
    return /(always|never|prefer|should|must|avoid|remember to|do not|don't|use\b)/.test(lowered)
        ? 'instruction'
        : 'memory';
}

function deriveTags(content: string, sourceTool: string): string[] {
    const lowered = content.toLowerCase();
    const tags = [sourceTool, classifyMemoryKind(content)];

    if (/(port|localhost|127\.0\.0\.1|http|ws:|wss:)/.test(lowered)) tags.push('networking');
    if (/(memory|context|session|history)/.test(lowered)) tags.push('memory');
    if (/(sqlite|database|db)/.test(lowered)) tags.push('database');
    if (/(build|typecheck|test|vitest|tsc)/.test(lowered)) tags.push('validation');
    if (/(mcp|tool|server|catalog)/.test(lowered)) tags.push('mcp');
    if (/(dashboard|ui|widget|page)/.test(lowered)) tags.push('ui');

    return uniqueStrings(tags);
}

function heuristicMemoryExtraction(text: string, sourceTool: string): ImportedSessionMemoryInput[] {
    const candidateLines = text
        .split(/\r?\n/)
        .map((line) => sanitizeSentence(line, 240))
        .filter((line) => line.length >= 24)
        .filter((line) => /(use|prefer|should|must|avoid|remember|fixed|fix|discovered|default|path|port|error|warning|supports|requires)/i.test(line));

    const candidateSentences = text
        .split(/(?<=[.!?])\s+/)
        .map((sentence) => sanitizeSentence(sentence, 220))
        .filter((sentence) => sentence.length >= 30)
        .filter((sentence) => /(use|prefer|should|must|avoid|remember|fixed|default|path|port|error|warning|supports|requires)/i.test(sentence));

    const facts = uniqueStrings([...candidateLines, ...candidateSentences], 6);

    return facts.map((fact) => ({
        kind: classifyMemoryKind(fact),
        content: fact,
        tags: deriveTags(fact, sourceTool),
        source: 'heuristic',
        metadata: {
            extraction: 'heuristic',
        },
    }));
}

function buildRetentionSummary(
    session: NormalizedImportedSession,
    parsedMemories: ImportedSessionMemoryInput[],
    analyzedChars: number,
    strategy: ImportedSessionMemorySource,
    overrides?: Partial<ImportedSessionRetentionSummary>,
): ImportedSessionRetentionSummary {
    const durableInstructionCount = parsedMemories.filter((memory) => memory.kind === 'instruction').length;
    const durableMemoryCount = parsedMemories.length - durableInstructionCount;
    const defaultSummary = parsedMemories.length > 0
        ? `Archived full transcript; promoted ${parsedMemories.length} durable item(s) to fast memory while keeping the remaining context compressed.`
        : 'Archived full transcript for historical reference only; no durable memories were promoted.';

    return {
        strategy,
        transcriptLength: session.transcript.length,
        analyzedChars,
        durableMemoryCount,
        durableInstructionCount,
        archiveDisposition: 'archive_only',
        summary: sanitizeSentence(overrides?.summary ?? defaultSummary, 240),
        salientTags: uniqueStrings(
            overrides?.salientTags ?? parsedMemories.flatMap((memory) => memory.tags ?? []),
            10,
        ),
        keepArchivedCategories: uniqueStrings(
            overrides?.keepArchivedCategories ?? ['conversational-context', 'implementation-detail', 'historical-trace'],
            8,
        ),
        discardableCategories: uniqueStrings(
            overrides?.discardableCategories ?? ['greetings', 'small-talk', 'duplicate-restatements'],
            8,
        ),
    };
}

function parseJsonStringArray(value: unknown): string[] {
    if (Array.isArray(value)) {
        return uniqueStrings(value.filter((entry): entry is string => typeof entry === 'string'));
    }

    if (typeof value !== 'string' || !value.trim()) {
        return [];
    }

    try {
        const parsed = JSON.parse(value);
        return Array.isArray(parsed)
            ? uniqueStrings(parsed.filter((entry): entry is string => typeof entry === 'string'))
            : [];
    } catch {
        return [];
    }
}

function parseJsonRecord(value: unknown): Record<string, unknown> {
    if (value && typeof value === 'object' && !Array.isArray(value)) {
        return value as Record<string, unknown>;
    }

    if (typeof value !== 'string' || !value.trim()) {
        return {};
    }

    try {
        const parsed = JSON.parse(value);
        return parsed && typeof parsed === 'object' && !Array.isArray(parsed)
            ? parsed as Record<string, unknown>
            : {};
    } catch {
        return {};
    }
}

function toTimestampMs(value: unknown): number | null {
    if (typeof value === 'number' && Number.isFinite(value)) {
        return value > 1_000_000_000_000 ? value : value * 1000;
    }

    if (typeof value !== 'string' || !value.trim()) {
        return null;
    }

    const parsed = Date.parse(value);
    return Number.isNaN(parsed) ? null : parsed;
}

function isMissingFileError(error: unknown): boolean {
    if (!error || typeof error !== 'object') {
        return false;
    }

    const record = error as Record<string, unknown>;
    const code = typeof record.code === 'string' ? record.code : '';
    return code === 'ENOENT';
}

function extractWorkingDirectory(metadata: Record<string, unknown>): string | null {
    const candidates = [
        metadata.cwd,
        metadata.workingDirectory,
        metadata.workspacePath,
        metadata.repoPath,
        metadata.projectPath,
    ];

    for (const candidate of candidates) {
        if (typeof candidate === 'string' && candidate.trim()) {
            return candidate.trim();
        }
    }

    return null;
}

function buildTormentNexusLedgerTranscript(row: Record<string, unknown>): string {
    const lines = [
        `TormentNexus session ledger for project ${sanitizeSentence(String(row.project ?? 'default'), 120)}`,
        typeof row.summary === 'string' ? row.summary : '',
    ];

    if (typeof row.event_type === 'string' && row.event_type.trim() && row.event_type !== 'session') {
        lines.push(`Event type: ${row.event_type}`);
    }

    if (typeof row.confidence_score === 'number' && Number.isFinite(row.confidence_score)) {
        lines.push(`Confidence score: ${row.confidence_score}`);
    }

    if (typeof row.importance === 'number' && Number.isFinite(row.importance) && row.importance > 0) {
        lines.push(`Importance: ${row.importance}`);
    }

    const todos = parseJsonStringArray(row.todos);
    if (todos.length > 0) {
        lines.push('Open TODOs:', ...todos.map((todo) => `- ${todo}`));
    }

    const decisions = parseJsonStringArray(row.decisions);
    if (decisions.length > 0) {
        lines.push('Decisions:', ...decisions.map((decision) => `- ${decision}`));
    }

    const filesChanged = parseJsonStringArray(row.files_changed);
    if (filesChanged.length > 0) {
        lines.push('Files changed:', ...filesChanged.map((fileName) => `- ${fileName}`));
    }

    const keywords = parseJsonStringArray(row.keywords);
    if (keywords.length > 0) {
        lines.push(`Keywords: ${keywords.join(', ')}`);
    }

    if (typeof row.created_at === 'string' && row.created_at.trim()) {
        lines.push(`Created at: ${row.created_at}`);
    }

    if (typeof row.role === 'string' && row.role.trim()) {
        lines.push(`Role: ${row.role}`);
    }

    return lines.filter(Boolean).join('\n');
}

function buildTormentNexusHandoffTranscript(row: Record<string, unknown>): string {
    const lines = [
        `TormentNexus handoff for project ${sanitizeSentence(String(row.project ?? 'default'), 120)}`,
        typeof row.last_summary === 'string' ? row.last_summary : '',
        typeof row.key_context === 'string' ? row.key_context : '',
    ];

    const pendingTodo = parseJsonStringArray(row.pending_todo);
    if (pendingTodo.length > 0) {
        lines.push('Open TODOs:', ...pendingTodo.map((todo) => `- ${todo}`));
    }

    const activeDecisions = parseJsonStringArray(row.active_decisions);
    if (activeDecisions.length > 0) {
        lines.push('Active decisions:', ...activeDecisions.map((decision) => `- ${decision}`));
    }

    const keywords = parseJsonStringArray(row.keywords);
    if (keywords.length > 0) {
        lines.push(`Keywords: ${keywords.join(', ')}`);
    }

    if (typeof row.active_branch === 'string' && row.active_branch.trim()) {
        lines.push(`Active branch: ${row.active_branch}`);
    }

    if (typeof row.version === 'number') {
        lines.push(`Version: ${row.version}`);
    }

    if (typeof row.updated_at === 'string' && row.updated_at.trim()) {
        lines.push(`Updated at: ${row.updated_at}`);
    }

    return lines.filter(Boolean).join('\n');
}

function parseJsonValue(value: unknown): unknown {
    if (typeof value !== 'string' || !value.trim()) {
        return value;
    }

    try {
        return JSON.parse(value);
    } catch {
        return value;
    }
}

function buildLlmLogText(primary: unknown, jsonFallback?: unknown): string {
    if (typeof primary === 'string' && primary.trim()) {
        return primary.trim();
    }

    const extracted = uniqueStrings(extractJsonTextFragments(parseJsonValue(jsonFallback)), 40);
    return extracted.join('\n').trim();
}

function buildLlmConversationTranscript(
    rows: Array<Record<string, unknown>>,
    toolCallsByResponse: Map<string, string[]>,
    toolResultsByResponse: Map<string, string[]>,
): string {
    const entries: string[] = [];

    for (const row of rows) {
        const responseId = typeof row.id === 'string' ? row.id : '';
        const systemText = buildLlmLogText(row.system);
        const promptText = buildLlmLogText(row.prompt, row.prompt_json);
        const responseText = buildLlmLogText(row.response, row.response_json);
        const toolCalls = responseId ? toolCallsByResponse.get(responseId) ?? [] : [];
        const toolResults = responseId ? toolResultsByResponse.get(responseId) ?? [] : [];
        const blocks: string[] = [];

        if (systemText) {
            blocks.push(`System: ${systemText}`);
        }

        if (promptText) {
            blocks.push(`User: ${promptText}`);
        }

        const assistantFragments = [
            responseText,
            ...toolCalls.map((name) => `[Tool Call: ${name}]`),
            ...toolResults.map((name) => `[Tool Result: ${name}]`),
        ].filter((fragment) => fragment.trim().length > 0);

        if (assistantFragments.length > 0) {
            blocks.push(`Assistant: ${assistantFragments.join('\n')}`);
        }

        if (blocks.length > 0) {
            entries.push(blocks.join('\n\n'));
        }
    }

    return entries.join('\n\n');
}

async function pathExists(candidatePath: string): Promise<boolean> {
    try {
        await fs.stat(candidatePath);
        return true;
    } catch {
        return false;
    }
}

export class SessionImportService {
    private llmService: LLMService;
    private memoryService: AgentMemoryService;
    private workspaceRoot: string;
    private store: ImportedSessionStore;
    private includeHomeDirectories: boolean;
    private importIntervalMs: number;
    private maxFilesPerRoot: number;
    private scanTimer: NodeJS.Timeout | null = null;
    private docsDir: string;

    constructor(
        llmService: LLMService,
        memoryService: AgentMemoryService,
        workspaceRoot: string = process.cwd(),
        options: SessionImportServiceOptions = {},
    ) {
        this.llmService = llmService;
        this.memoryService = memoryService;
        this.workspaceRoot = workspaceRoot;
        this.store = options.store ?? new ImportedSessionStore(
            path.join(this.workspaceRoot, '.tormentnexus', 'imported_sessions', 'archive'),
        );
        this.includeHomeDirectories = options.includeHomeDirectories ?? true;
        this.importIntervalMs = options.importIntervalMs ?? DEFAULT_SCAN_INTERVAL_MS;
        this.maxFilesPerRoot = options.maxFilesPerRoot ?? DEFAULT_MAX_FILES_PER_ROOT;
        this.docsDir = path.join(this.workspaceRoot, '.tormentnexus', 'imported_sessions', 'docs');

        if (typeof (this.store as ImportedSessionStore & { compactInlineTranscripts?: unknown }).compactInlineTranscripts === 'function') {
            try {
                this.store.compactInlineTranscripts(250);
            } catch (error) {
                console.warn(formatOptionalSqliteFailure(
                    '[SessionImport] Failed to compact inline imported transcripts',
                    error,
                ));
            }
        }

        if (typeof (this.store as ImportedSessionStore & { backfillRetentionSummaries?: unknown }).backfillRetentionSummaries === 'function') {
            try {
                this.store.backfillRetentionSummaries((session) => {
                    const summary = buildRetentionSummary(
                        {
                            sessionId: session.id,
                            title: session.title ?? path.basename(session.sourcePath),
                            sourceTool: session.sourceTool,
                            sourcePath: session.sourcePath,
                            sessionFormat: session.sessionFormat,
                            externalSessionId: session.externalSessionId,
                            workingDirectory: session.workingDirectory ?? this.workspaceRoot,
                            transcript: session.transcript,
                            excerpt: session.excerpt ?? '',
                            transcriptHash: session.transcriptHash,
                            discoveredAt: session.discoveredAt,
                            importedAt: session.importedAt,
                            lastModifiedAt: session.lastModifiedAt,
                            normalizedSession: session.normalizedSession,
                            metadata: session.metadata,
                        },
                        session.parsedMemories.map((memory) => ({
                            kind: memory.kind,
                            content: memory.content,
                            tags: memory.tags,
                            source: memory.source,
                            metadata: memory.metadata,
                        })),
                        session.transcript.length > 16_000 ? 16_000 : session.transcript.length,
                        'heuristic',
                    );

                    return {
                        strategy: summary.strategy,
                        transcriptLength: summary.transcriptLength,
                        analyzedChars: summary.analyzedChars,
                        durableMemoryCount: summary.durableMemoryCount,
                        durableInstructionCount: summary.durableInstructionCount,
                        archiveDisposition: summary.archiveDisposition,
                        summary: summary.summary,
                        salientTags: summary.salientTags,
                        keepArchivedCategories: summary.keepArchivedCategories,
                        discardableCategories: summary.discardableCategories,
                    };
                }, 250);
            } catch (error) {
                console.warn(formatOptionalSqliteFailure(
                    '[SessionImport] Failed to backfill imported transcript retention summaries',
                    error,
                ));
            }
        }
    }

    public startAutoImport(): void {
        if (this.scanTimer) {
            return;
        }

        void this.runAutoImportIteration();
        this.scanTimer = setInterval(() => {
            void this.runAutoImportIteration();
        }, this.importIntervalMs);
    }

    private async runAutoImportIteration(): Promise<void> {
        try {
            await this.scanAndImport();
        } catch (error) {
            console.warn(formatOptionalSqliteFailure(
                '[SessionImport] Auto-import iteration failed',
                error,
            ));
        }
    }

    public async scanAndImport(options: { force?: boolean } = {}): Promise<SessionImportSummary> {
        const candidates = await this.discoverCandidates();
        let importedCount = 0;
        let skippedCount = 0;
        let storedMemoryCount = 0;
        const tools = new Set<string>();

        for (const candidate of candidates) {
            tools.add(candidate.sourceTool);
            let imported: ImportedSessionRecord | null = null;
            try {
                imported = await this.importCandidate(candidate, options.force ?? false);
            } catch (error) {
                if (isMissingFileError(error)) {
                    console.warn(`[SessionImport] Skipping vanished session source ${candidate.sourcePath}`);
                    skippedCount += 1;
                    continue;
                }

                if (isSqliteUnavailableError(error)) {
                    console.warn(formatOptionalSqliteFailure(
                        `[SessionImport] Skipping imported-session persistence for ${candidate.sourcePath}`,
                        error,
                    ));
                    skippedCount += 1;
                    continue;
                }

                throw error;
            }
            if (!imported) {
                skippedCount += 1;
                continue;
            }

            importedCount += 1;
            storedMemoryCount += imported.parsedMemories.length;
        }

        const instructionDocPath = await this.writeInstructionDocs();

        return {
            discoveredCount: candidates.length,
            importedCount,
            skippedCount,
            storedMemoryCount,
            instructionDocPath,
            tools: Array.from(tools).sort(),
        };
    }

    public listImportedSessions(limit: number = 50): ImportedSessionRecord[] {
        return this.store.listImportedSessions(limit);
    }

    public getImportedMaintenanceStats(): ImportedSessionMaintenanceStats {
        return this.store.getMaintenanceStats();
    }

    public getImportedSession(id: string): ImportedSessionRecord | null {
        return this.store.getImportedSession(id);
    }

    public async listInstructionDocs(): Promise<ImportedInstructionDoc[]> {
        if (!(await pathExists(this.docsDir))) {
            return [];
        }

        const entries = await fs.readdir(this.docsDir, { withFileTypes: true });
        const docs: ImportedInstructionDoc[] = [];

        for (const entry of entries) {
            if (!entry.isFile() || !entry.name.endsWith('.md')) continue;
            const docPath = path.join(this.docsDir, entry.name);
            const stat = await fs.stat(docPath);
            docs.push({
                path: docPath,
                updatedAt: stat.mtimeMs,
                size: stat.size,
            });
        }

        return docs.sort((left, right) => right.updatedAt - left.updatedAt);
    }

    private getDiscoveryRules(): SessionDiscoveryRule[] {
        const homeDir = os.homedir();
        const appData = process.env.APPDATA ?? '';
        const localAppData = process.env.LOCALAPPDATA ?? '';

        const workspaceRules: SessionDiscoveryRule[] = [
            {
                sourceTool: 'aider',
                roots: [path.join(this.workspaceRoot, '.aider.chat.history.md'), path.join(this.workspaceRoot, '.aider')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['aider', 'history', 'chat'],
            },
            {
                sourceTool: 'claude-code',
                roots: [path.join(this.workspaceRoot, '.claude')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['claude', 'session', 'chat', 'conversation', 'transcript'],
            },
            {
                sourceTool: 'cursor',
                roots: [path.join(this.workspaceRoot, '.cursor')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['cursor', 'session', 'chat', 'conversation', 'history'],
                ignoredPathHints: ['cursor-retrieval', 'workspace.json', 'embeddable_files', 'high_level_folder_description'],
            },
            {
                sourceTool: 'windsurf',
                roots: [path.join(this.workspaceRoot, '.windsurf')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['windsurf', 'session', 'chat', 'conversation', 'history'],
                ignoredPathHints: ['workspace.json'],
            },
            {
                sourceTool: 'opencode',
                roots: [path.join(this.workspaceRoot, '.opencode'), path.join(this.workspaceRoot, '.docs', 'ai-logs')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['opencode', 'session', 'chat', 'conversation', 'history', 'log'],
            },
            {
                sourceTool: 'codex',
                roots: [path.join(this.workspaceRoot, '.codex')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['codex', 'session', 'chat', 'conversation', 'history'],
            },
            {
                sourceTool: 'gemini',
                roots: [path.join(this.workspaceRoot, '.gemini')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['gemini', 'session', 'chat', 'conversation', 'history'],
            },
            {
                sourceTool: 'copilot-cli',
                roots: [path.join(this.workspaceRoot, '.copilot', 'session-state')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'handoff', 'checkpoint', 'history'],
            },
            {
                sourceTool: 'openai',
                roots: [
                    path.join(this.workspaceRoot, '.openai'),
                    path.join(this.workspaceRoot, '.chatgpt'),
                    path.join(this.workspaceRoot, 'ChatGPT'),
                    path.join(this.workspaceRoot, 'OpenAI'),
                ],
                filePatterns: ['**/*.{json,jsonl,md,txt,log}'],
                fileNameHints: ['openai', 'chatgpt', 'conversation', 'history', 'export', 'session', 'messages'],
            },
            {
                sourceTool: 'tormentnexus',
                roots: [path.join(this.workspaceRoot, '.tormentnexus')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'memory', 'handoff', 'history'],
            },
            {
                sourceTool: 'vscode-extensions',
                roots: [path.join(this.workspaceRoot, '.vscode')],
                filePatterns: ['**/*.{json,jsonl,md,txt,log}'],
                fileNameHints: ['session', 'chat', 'conversation', 'history', 'transcript', 'messages'],
                ignoredPathHints: ['settings.json', 'extensions.json', 'workspace.json'],
            },
        ];

        if (!this.includeHomeDirectories) {
            return workspaceRules;
        }

        return [
            ...workspaceRules,
            {
                sourceTool: 'claude-code',
                roots: [path.join(homeDir, '.claude'), path.join(appData, 'Claude')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'chat', 'conversation', 'transcript', 'history'],
                ignoredPathHints: ['analytics.json', 'marketplacemanifest.json', 'plugins', 'projects'],
            },
            {
                sourceTool: 'aider',
                roots: [path.join(homeDir, '.aider.chat.history.md'), path.join(homeDir, '.aider')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['aider', 'history', 'chat'],
                ignoredPathHints: ['analytics.json', 'model_prices.json', 'installs.json'],
            },
            {
                sourceTool: 'cursor',
                roots: [path.join(appData, 'Cursor', 'User', 'workspaceStorage'), path.join(localAppData, 'Cursor', 'User', 'workspaceStorage')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'chat', 'conversation', 'history'],
                ignoredPathHints: ['cursor-retrieval', 'workspace.json', 'embeddable_files', 'high_level_folder_description'],
            },
            {
                sourceTool: 'windsurf',
                roots: [path.join(appData, 'Windsurf', 'User', 'workspaceStorage')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'chat', 'conversation', 'history'],
                ignoredPathHints: ['workspace.json'],
            },
            {
                sourceTool: 'copilot-cli',
                roots: [path.join(homeDir, '.copilot', 'session-state')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'handoff', 'checkpoint', 'history'],
            },
            {
                sourceTool: 'opencode',
                roots: [path.join(homeDir, '.opencode')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['opencode', 'session', 'chat', 'conversation', 'history', 'log', 'message'],
            },
            {
                sourceTool: 'openai',
                roots: [
                    path.join(homeDir, '.openai'),
                    path.join(homeDir, '.chatgpt'),
                    path.join(homeDir, 'ChatGPT'),
                    path.join(homeDir, 'OpenAI'),
                    path.join(homeDir, 'Documents', 'ChatGPT'),
                    path.join(homeDir, 'Documents', 'OpenAI'),
                ],
                filePatterns: ['**/*.{json,jsonl,md,txt,log}'],
                fileNameHints: ['openai', 'chatgpt', 'conversation', 'history', 'export', 'session', 'messages'],
            },
            {
                sourceTool: 'copilot-chat',
                roots: [
                    path.join(appData, 'Code', 'User', 'globalStorage', 'emptyWindowChatSessions'),
                    path.join(appData, 'Code - Insiders', 'User', 'globalStorage', 'emptyWindowChatSessions'),
                ],
                filePatterns: ['**/*.{json,jsonl,md,txt,log}'],
                importAllFiles: true,
            },
            {
                sourceTool: 'copilot-chat',
                roots: [
                    path.join(appData, 'Code', 'User', 'globalStorage', 'github.copilot-chat'),
                    path.join(appData, 'Code - Insiders', 'User', 'globalStorage', 'github.copilot-chat'),
                ],
                filePatterns: ['**/*.{json,jsonl,md,txt,log}'],
                fileNameHints: ['session', 'chat', 'conversation', 'copilot'],
                ignoredPathHints: [
                    'commandembeddings',
                    'settingembeddings',
                    'memory-tool',
                    'ask-agent',
                    'explore-agent',
                    'plan-agent',
                    'debugcommand',
                    'copilotcli\\copilot.',
                    'copilotcli.session.metadata',
                ],
            },
            {
                sourceTool: 'codex',
                roots: [path.join(homeDir, '.codex')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'chat', 'conversation', 'history'],
            },
            {
                sourceTool: 'gemini',
                roots: [path.join(homeDir, '.gemini')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'chat', 'conversation', 'history'],
            },
            {
                sourceTool: 'antigravity',
                roots: [path.join(homeDir, '.gemini', 'antigravity', 'brain')],
                filePatterns: ['**/*.{json,jsonl,log,md,txt}'],
                importAllFiles: true,
            },
            {
                sourceTool: 'tormentnexus',
                roots: [path.join(homeDir, '.tormentnexus')],
                filePatterns: ['**/*.{md,txt,log,json,jsonl}'],
                fileNameHints: ['session', 'memory', 'handoff', 'history'],
            },
            {
                sourceTool: 'vscode-extensions',
                roots: [
                    path.join(appData, 'Code', 'User', 'globalStorage'),
                    path.join(appData, 'Code - Insiders', 'User', 'globalStorage'),
                ],
                filePatterns: ['**/*.{json,jsonl,md,txt,log}'],
                fileNameHints: ['session', 'chat', 'conversation', 'history', 'transcript', 'messages'],
                ignoredPathHints: [
                    'emptywindowchatsessions',
                    'github.copilot-chat',
                    'commandembeddings',
                    'settingembeddings',
                    'memory-tool',
                    'ask-agent',
                    'explore-agent',
                    'plan-agent',
                    'debugcommand',
                    'copilotcli\\copilot.',
                    'copilotcli.session.metadata',
                ],
            },
        ];
    }

    private async discoverCandidates(): Promise<DiscoveryCandidate[]> {
        const discovered = new Map<string, DiscoveryCandidate>();

        for (const rule of this.getDiscoveryRules()) {
            for (const rootPath of rule.roots) {
                if (!rootPath) continue;
                if (!(await pathExists(rootPath))) continue;

                const stat = await fs.stat(rootPath);
                if (stat.isFile()) {
                    if (this.isImportableSessionPath(rootPath, rule.fileNameHints)) {
                        const structuredCandidates = await this.expandStructuredCandidates(rule.sourceTool, rootPath, stat.mtimeMs);
                        if (structuredCandidates.length > 0) {
                            for (const candidate of structuredCandidates) {
                                discovered.set(candidate.sourcePath, candidate);
                            }
                        } else {
                            discovered.set(rootPath, {
                                sourceTool: rule.sourceTool,
                                sourcePath: rootPath,
                                sessionFormat: path.extname(rootPath).replace(/^\./, '') || 'text',
                                lastModifiedAt: stat.mtimeMs,
                                metadata: rule.sourceTool === 'antigravity'
                                    ? {
                                        antigravityImportSurface: 'experimental',
                                        antigravityDiscoveryRoot: 'brain',
                                        antigravitySource: 'reverse-engineered',
                                    }
                                    : undefined,
                            });
                        }
                    }
                    continue;
                }

                const matches = await fg(rule.filePatterns, {
                    cwd: rootPath,
                    absolute: true,
                    onlyFiles: true,
                    unique: true,
                    deep: 6,
                    ignore: ['**/node_modules/**', '**/.git/**', '**/dist/**', '**/build/**', '**/coverage/**'],
                });

                for (const match of matches.slice(0, this.maxFilesPerRoot)) {
                    if (!this.isImportableSessionPath(
                        match,
                        rule.fileNameHints,
                        rule.importAllFiles ?? false,
                        rule.ignoredPathHints,
                    )) continue;
                    let matchStat;
                    try {
                        matchStat = await fs.stat(match);
                    } catch (error) {
                        if (isMissingFileError(error)) {
                            continue;
                        }
                        throw error;
                    }
                    const structuredCandidates = await this.expandStructuredCandidates(rule.sourceTool, match, matchStat.mtimeMs);
                    if (structuredCandidates.length > 0) {
                        for (const candidate of structuredCandidates) {
                            discovered.set(candidate.sourcePath, candidate);
                        }
                    } else {
                        discovered.set(match, {
                            sourceTool: rule.sourceTool,
                            sourcePath: match,
                            sessionFormat: path.extname(match).replace(/^\./, '') || 'text',
                            lastModifiedAt: matchStat.mtimeMs,
                            metadata: rule.sourceTool === 'antigravity'
                                ? {
                                    antigravityImportSurface: 'experimental',
                                    antigravityDiscoveryRoot: 'brain',
                                    antigravitySource: 'reverse-engineered',
                                }
                                : undefined,
                        });
                    }
                }
            }
        }

        for (const candidate of await this.discoverTormentNexusDatabaseCandidates()) {
            discovered.set(candidate.sourcePath, candidate);
        }

        for (const candidate of await this.discoverLlmDatabaseCandidates()) {
            discovered.set(candidate.sourcePath, candidate);
        }

        return Array.from(discovered.values()).sort((left, right) => (right.lastModifiedAt ?? 0) - (left.lastModifiedAt ?? 0));
    }

    private async expandStructuredCandidates(
        sourceTool: string,
        sourcePath: string,
        lastModifiedAt: number | null,
    ): Promise<DiscoveryCandidate[]> {
        if (sourceTool !== 'openai' || path.extname(sourcePath).toLowerCase() !== '.json') {
            return [];
        }

        let parsed: unknown;
        try {
            parsed = JSON.parse(await fs.readFile(sourcePath, 'utf-8'));
        } catch {
            return [];
        }

        const conversations = Array.isArray(parsed)
            ? parsed
            : parsed && typeof parsed === 'object' && Array.isArray((parsed as Record<string, unknown>).conversations)
                ? (parsed as Record<string, unknown>).conversations as unknown[]
                : [];

        const candidates: DiscoveryCandidate[] = [];

        for (const entry of conversations) {
            if (!entry || typeof entry !== 'object') {
                continue;
            }

            const conversation = entry as Record<string, unknown>;
            const transcript = formatChatGptExportTranscript(conversation)?.trim();
            if (!transcript) {
                continue;
            }

            const conversationId = typeof conversation.id === 'string' && conversation.id.trim()
                ? conversation.id.trim()
                : typeof conversation.conversation_id === 'string' && conversation.conversation_id.trim()
                    ? conversation.conversation_id.trim()
                    : null;
            const title = typeof conversation.title === 'string' && conversation.title.trim()
                ? sanitizeSentence(conversation.title, 120)
                : conversationId
                    ? `ChatGPT export ${conversationId}`
                    : path.basename(sourcePath, path.extname(sourcePath));
            const candidateSourcePath = conversationId
                ? `${sourcePath}#conversation:${conversationId}`
                : `${sourcePath}#conversation:${candidates.length + 1}`;

            candidates.push({
                sourceTool,
                sourcePath: candidateSourcePath,
                sessionFormat: 'chatgpt-export',
                lastModifiedAt: toTimestampMs(conversation.update_time) ?? toTimestampMs(conversation.create_time) ?? lastModifiedAt,
                externalSessionId: conversationId,
                transcript,
                title,
                workingDirectory: path.dirname(sourcePath),
                metadata: {
                    openaiExportType: 'chatgpt-conversations',
                    openaiExportSource: sourcePath,
                    openaiConversationId: conversationId,
                    openaiCreateTime: toTimestampMs(conversation.create_time),
                    openaiUpdateTime: toTimestampMs(conversation.update_time),
                },
            });
        }

        return candidates;
    }

    private async discoverTormentNexusDatabaseCandidates(): Promise<DiscoveryCandidate[]> {
        const roots = [
            path.join(this.workspaceRoot, '.tormentnexus-mcp', 'data.db'),
            this.includeHomeDirectories ? path.join(os.homedir(), '.tormentnexus-mcp', 'data.db') : null,
        ].filter((root): root is string => Boolean(root));

        const discovered: DiscoveryCandidate[] = [];
        const seen = new Set<string>();

        for (const dbPath of roots) {
            if (!(await pathExists(dbPath))) {
                continue;
            }

            try {
                const tormentnexusDb = new Database(dbPath, { readonly: true, fileMustExist: true });

                try {
                    const ledgerRows = tormentnexusDb
                        .prepare('SELECT * FROM session_ledger ORDER BY datetime(created_at) DESC LIMIT ?')
                        .all(this.maxFilesPerRoot) as Array<Record<string, unknown>>;

                    for (const row of ledgerRows) {
                        const id = typeof row.id === 'string' ? row.id : null;
                        if (!id) continue;

                        const sourcePath = `${dbPath}#session_ledger:${id}`;
                        if (seen.has(sourcePath)) continue;
                        seen.add(sourcePath);

                        const transcript = buildTormentNexusLedgerTranscript(row).trim();
                        if (!transcript) continue;

                        const metadata: Record<string, unknown> = {
                            tormentnexusProject: row.project,
                            tormentnexusRole: row.role,
                            tormentnexusTable: 'session_ledger',
                            tormentnexusConversationId: row.conversation_id,
                            tormentnexusEventType: typeof row.event_type === 'string' && row.event_type.trim()
                                ? row.event_type
                                : 'session',
                            tormentnexusConfidenceScore: typeof row.confidence_score === 'number' && Number.isFinite(row.confidence_score)
                                ? row.confidence_score
                                : null,
                            tormentnexusImportance: typeof row.importance === 'number' && Number.isFinite(row.importance)
                                ? row.importance
                                : 0,
                        };

                        if (
                            metadata.tormentnexusEventType === 'correction'
                            && typeof metadata.tormentnexusImportance === 'number'
                            && metadata.tormentnexusImportance >= 3
                            && typeof row.summary === 'string'
                            && row.summary.trim()
                        ) {
                            metadata.behavioralWarnings = [sanitizeSentence(row.summary, 220)];
                        }

                        discovered.push({
                            sourceTool: 'tormentnexus-mcp',
                            sourcePath,
                            sessionFormat: 'tormentnexus-ledger',
                            lastModifiedAt: toTimestampMs(row.created_at),
                            externalSessionId: typeof row.conversation_id === 'string' && row.conversation_id.trim()
                                ? row.conversation_id
                                : id,
                            transcript,
                            title: typeof row.summary === 'string' && row.summary.trim()
                                ? sanitizeSentence(row.summary, 120)
                                : `TormentNexus ledger ${id}`,
                            workingDirectory: path.dirname(dbPath),
                            metadata,
                        });
                    }
                } catch {
                    // Table may not exist yet in a freshly initialized TormentNexus install.
                }

                try {
                    const handoffRows = tormentnexusDb
                        .prepare('SELECT * FROM session_handoffs ORDER BY datetime(COALESCE(updated_at, created_at)) DESC LIMIT ?')
                        .all(this.maxFilesPerRoot) as Array<Record<string, unknown>>;

                    for (const row of handoffRows) {
                        const project = typeof row.project === 'string' ? row.project : null;
                        if (!project) continue;

                        const sourcePath = `${dbPath}#session_handoffs:${project}`;
                        if (seen.has(sourcePath)) continue;
                        seen.add(sourcePath);

                        const transcript = buildTormentNexusHandoffTranscript(row).trim();
                        if (!transcript) continue;

                        const parsedMetadata = parseJsonRecord(row.metadata);
                        const metadata = {
                            tormentnexusProject: project,
                            tormentnexusTable: 'session_handoffs',
                            tormentnexusVersion: row.version,
                            ...parsedMetadata,
                        };

                        discovered.push({
                            sourceTool: 'tormentnexus-mcp',
                            sourcePath,
                            sessionFormat: 'tormentnexus-handoff',
                            lastModifiedAt: toTimestampMs(row.updated_at) ?? toTimestampMs(row.created_at),
                            externalSessionId: `handoff:${project}`,
                            transcript,
                            title: `TormentNexus handoff ${project}`,
                            workingDirectory: extractWorkingDirectory(parsedMetadata) ?? path.dirname(dbPath),
                            metadata,
                        });
                    }
                } catch {
                    // Table may not exist in minimal TormentNexus setups.
                }

                tormentnexusDb.close();
            } catch (error) {
                console.warn(formatOptionalSqliteFailure(
                    `[SessionImport] Failed to inspect TormentNexus database at ${dbPath}`,
                    error,
                ));
            }
        }

        return discovered;
    }

    private async discoverLlmDatabaseCandidates(): Promise<DiscoveryCandidate[]> {
        const homeDir = os.homedir();
        const appData = process.env.APPDATA ?? '';
        const localAppData = process.env.LOCALAPPDATA ?? '';
        const llmUserPath = process.env.LLM_USER_PATH?.trim();
        const llmUserDbPath = llmUserPath
            ? (path.extname(llmUserPath).toLowerCase() === '.db' ? llmUserPath : path.join(llmUserPath, 'logs.db'))
            : null;
        const roots = uniqueStrings([
            path.join(this.workspaceRoot, '.llm', 'logs.db'),
            path.join(homeDir, '.llm', 'logs.db'),
            path.join(homeDir, '.config', 'io.datasette.llm', 'logs.db'),
            path.join(homeDir, '.local', 'share', 'io.datasette.llm', 'logs.db'),
            path.join(homeDir, 'Library', 'Application Support', 'io.datasette.llm', 'logs.db'),
            appData ? path.join(appData, 'io.datasette.llm', 'logs.db') : null,
            localAppData ? path.join(localAppData, 'io.datasette.llm', 'logs.db') : null,
            llmUserDbPath,
        ]);

        const discovered: DiscoveryCandidate[] = [];
        const seen = new Set<string>();

        for (const dbPath of roots) {
            if (!(await pathExists(dbPath))) {
                continue;
            }

            try {
                // Check file size before opening
                const MAX_DB_FILE_SIZE = 10 * 1024 * 1024; // 10MB
                const stat = await fs.stat(dbPath);
                if (stat.size > MAX_DB_FILE_SIZE) {
                    console.warn(`[SessionImport] Skipping database too large for scanning (>10MB): ${dbPath} (${stat.size} bytes)`);
                    continue;
                }

                const llmDb = new Database(dbPath, { readonly: true, fileMustExist: true });

                try {
                    const hasResponsesTable = Boolean(
                        llmDb.prepare("SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ? LIMIT 1").pluck().get('responses'),
                    );
                    if (!hasResponsesTable) {
                        continue;
                    }

                    const hasConversationsTable = Boolean(
                        llmDb.prepare("SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ? LIMIT 1").pluck().get('conversations'),
                    );
                    const hasToolCallsTable = Boolean(
                        llmDb.prepare("SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ? LIMIT 1").pluck().get('tool_calls'),
                    );
                    const hasToolResultsTable = Boolean(
                        llmDb.prepare("SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ? LIMIT 1").pluck().get('tool_results'),
                    );

                    const toolCallsByResponse = new Map<string, string[]>();
                    if (hasToolCallsTable) {
                        const toolCallRows = llmDb.prepare('SELECT response_id, name FROM tool_calls ORDER BY id ASC').all() as Array<Record<string, unknown>>;
                        for (const row of toolCallRows) {
                            const responseId = typeof row.response_id === 'string' ? row.response_id : null;
                            const name = typeof row.name === 'string' ? row.name.trim() : '';
                            if (!responseId || !name) continue;
                            const existing = toolCallsByResponse.get(responseId) ?? [];
                            existing.push(name);
                            toolCallsByResponse.set(responseId, existing);
                        }
                    }

                    const toolResultsByResponse = new Map<string, string[]>();
                    if (hasToolResultsTable) {
                        const toolResultRows = llmDb.prepare('SELECT response_id, name FROM tool_results ORDER BY id ASC').all() as Array<Record<string, unknown>>;
                        for (const row of toolResultRows) {
                            const responseId = typeof row.response_id === 'string' ? row.response_id : null;
                            const name = typeof row.name === 'string' ? row.name.trim() : '';
                            if (!responseId || !name) continue;
                            const existing = toolResultsByResponse.get(responseId) ?? [];
                            existing.push(name);
                            toolResultsByResponse.set(responseId, existing);
                        }
                    }

                    if (hasConversationsTable) {
                        const conversationRows = llmDb
                            .prepare(`
                                SELECT c.id, c.name, c.model, MAX(r.datetime_utc) AS last_activity
                                FROM conversations c
                                JOIN responses r ON r.conversation_id = c.id
                                GROUP BY c.id, c.name, c.model
                                ORDER BY datetime(MAX(r.datetime_utc)) DESC
                                LIMIT ?
                            `)
                            .all(this.maxFilesPerRoot) as Array<Record<string, unknown>>;

                        for (const row of conversationRows) {
                            const conversationId = typeof row.id === 'string' ? row.id : null;
                            if (!conversationId) continue;

                            const sourcePath = `${dbPath}#conversation:${conversationId}`;
                            if (seen.has(sourcePath)) continue;
                            seen.add(sourcePath);

                            const responseRows = llmDb
                                .prepare(`
                                    SELECT id, model, prompt, system, prompt_json, response, response_json, datetime_utc, input_tokens, output_tokens, resolved_model
                                    FROM responses
                                    WHERE conversation_id = ?
                                    ORDER BY datetime(datetime_utc) ASC, id ASC
                                `)
                                .all(conversationId) as Array<Record<string, unknown>>;
                            const transcript = buildLlmConversationTranscript(responseRows, toolCallsByResponse, toolResultsByResponse).trim();
                            if (!transcript) continue;

                            const inputTokens = responseRows.reduce((sum, entry) => sum + (typeof entry.input_tokens === 'number' ? entry.input_tokens : 0), 0);
                            const outputTokens = responseRows.reduce((sum, entry) => sum + (typeof entry.output_tokens === 'number' ? entry.output_tokens : 0), 0);
                            discovered.push({
                                sourceTool: 'llm-cli',
                                sourcePath,
                                sessionFormat: 'llm-conversation',
                                lastModifiedAt: toTimestampMs(row.last_activity),
                                externalSessionId: conversationId,
                                transcript,
                                title: typeof row.name === 'string' && row.name.trim()
                                    ? sanitizeSentence(row.name, 120)
                                    : `llm conversation ${conversationId}`,
                                workingDirectory: path.dirname(dbPath),
                                metadata: {
                                    llmConversationId: conversationId,
                                    llmConversationModel: row.model,
                                    llmResponseCount: responseRows.length,
                                    llmInputTokens: inputTokens,
                                    llmOutputTokens: outputTokens,
                                    llmDatabasePath: dbPath,
                                },
                            });
                        }
                    }

                    const remainingSlots = Math.max(this.maxFilesPerRoot - discovered.length, 0);
                    if (remainingSlots > 0) {
                        const orphanRows = llmDb
                            .prepare(`
                                SELECT id, model, prompt, system, prompt_json, response, response_json, datetime_utc, input_tokens, output_tokens, resolved_model
                                FROM responses
                                WHERE conversation_id IS NULL
                                ORDER BY datetime(datetime_utc) DESC, id DESC
                                LIMIT ?
                            `)
                            .all(remainingSlots) as Array<Record<string, unknown>>;

                        for (const row of orphanRows) {
                            const responseId = typeof row.id === 'string' ? row.id : null;
                            if (!responseId) continue;

                            const sourcePath = `${dbPath}#response:${responseId}`;
                            if (seen.has(sourcePath)) continue;
                            seen.add(sourcePath);

                            const transcript = buildLlmConversationTranscript([row], toolCallsByResponse, toolResultsByResponse).trim();
                            if (!transcript) continue;

                            const promptText = buildLlmLogText(row.prompt, row.prompt_json);
                            discovered.push({
                                sourceTool: 'llm-cli',
                                sourcePath,
                                sessionFormat: 'llm-response',
                                lastModifiedAt: toTimestampMs(row.datetime_utc),
                                externalSessionId: responseId,
                                transcript,
                                title: promptText
                                    ? sanitizeSentence(promptText, 120)
                                    : `llm response ${responseId}`,
                                workingDirectory: path.dirname(dbPath),
                                metadata: {
                                    llmResponseId: responseId,
                                    llmModel: row.model,
                                    llmResolvedModel: row.resolved_model,
                                    llmInputTokens: typeof row.input_tokens === 'number' ? row.input_tokens : 0,
                                    llmOutputTokens: typeof row.output_tokens === 'number' ? row.output_tokens : 0,
                                    llmDatabasePath: dbPath,
                                },
                            });
                        }
                    }
                } finally {
                    llmDb.close();
                }
            } catch (error) {
                console.warn(formatOptionalSqliteFailure(
                    `[SessionImport] Failed to inspect llm logs database at ${dbPath}`,
                    error,
                ));
            }
        }

        return discovered;
    }

    private isGeneratedImportPath(filePath: string): boolean {
        const normalizedPath = path.resolve(filePath).toLowerCase();
        const ignoredRoots = [
            path.join(this.workspaceRoot, '.tormentnexus', 'imported_sessions'),
            path.join(os.homedir(), '.tormentnexus', 'imported_sessions'),
        ]
            .map((root) => path.resolve(root).toLowerCase());

        return ignoredRoots.some((root) => normalizedPath === root || normalizedPath.startsWith(`${root}${path.sep}`));
    }

    private isImportableSessionPath(
        filePath: string,
        fileNameHints: string[] = [],
        importAllFiles: boolean = false,
        ignoredPathHints: string[] = [],
    ): boolean {
        if (this.isGeneratedImportPath(filePath)) {
            return false;
        }

        const normalized = filePath.toLowerCase();
        if (ignoredPathHints.some((hint) => normalized.includes(hint.toLowerCase()))) {
            return false;
        }

        const extension = path.extname(normalized);
        if (!IMPORTABLE_EXTENSIONS.has(extension)) {
            return false;
        }

        if (importAllFiles) {
            return true;
        }

        const hints = uniqueStrings([...DISCOVERY_HINTS, ...fileNameHints]);
        return hints.some((hint) => normalized.includes(hint.toLowerCase()));
    }

    private async importCandidate(candidate: DiscoveryCandidate, force: boolean): Promise<ImportedSessionRecord | null> {
        // Fast metadata-based check to skip unchanged files
        const MAX_SESSION_FILE_SIZE = 10 * 1024 * 1024; // 10MB
        let sourceSize: number | null = null;
        let sourceMtime: number | null = null;
        
        if (!candidate.sourcePath.includes('#')) {
            try {
                const stat = await fs.stat(candidate.sourcePath);
                sourceSize = stat.size;
                sourceMtime = stat.mtimeMs;
                
                if (sourceSize > MAX_SESSION_FILE_SIZE) {
                    console.warn(`[SessionImport] Skipping file too large for import (>10MB): ${candidate.sourcePath} (${sourceSize} bytes)`);
                    return null;
                }
                
                if (!force && this.store.hasMatchingSource(candidate.sourcePath, sourceSize, sourceMtime)) {
                    return null;
                }
            } catch (e) {
                // Ignore stat errors here, let the read handle it
            }
        }

        let transcript: string;
        if (typeof candidate.transcript === 'string') {
            transcript = candidate.transcript.trim();
        } else {
            try {
                transcript = parseTranscriptContent(candidate.sourcePath, await fs.readFile(candidate.sourcePath, 'utf-8')).trim();
            } catch (error) {
                if (isMissingFileError(error)) {
                    console.warn(`[SessionImport] Session source disappeared before import: ${candidate.sourcePath}`);
                    return null;
                }
                throw error;
            }
        }
        if (!transcript) {
            return null;
        }

        const transcriptHash = crypto.createHash('sha256').update(transcript).digest('hex');
        if (!force && this.store.hasTranscriptHash(transcriptHash)) {
            return null;
        }

        const normalizedSession = this.buildNormalizedSession(candidate, transcript, transcriptHash);
        
        // Use already fetched stat for the record if available
        const analysis = await this.analyzeImportedSession(normalizedSession);
        const parsedMemories = analysis.memories.map((memory) => ({
            ...memory,
            metadata: {
                ...(memory.metadata ?? {}),
                sourceTool: normalizedSession.sourceTool,
                path: normalizedSession.sourcePath,
                sessionId: normalizedSession.sessionId,
            },
        }));
        const imported = this.store.upsertSession({
            sourceTool: normalizedSession.sourceTool,
            sourcePath: normalizedSession.sourcePath,
            sourceSize,
            sourceMtime,
            externalSessionId: normalizedSession.externalSessionId,
            title: normalizedSession.title,
            sessionFormat: normalizedSession.sessionFormat,
            transcript: normalizedSession.transcript,
            excerpt: normalizedSession.excerpt,
            workingDirectory: normalizedSession.workingDirectory,
            transcriptHash: normalizedSession.transcriptHash,
            normalizedSession: normalizedSession.normalizedSession,
            metadata: {
                ...normalizedSession.metadata,
                retentionSummary: analysis.retentionSummary,
            },
            discoveredAt: normalizedSession.discoveredAt,
            importedAt: normalizedSession.importedAt,
            lastModifiedAt: normalizedSession.lastModifiedAt,
            parsedMemories,
        });

        await this.memoryService.captureSessionSummary({
            sessionId: normalizedSession.sessionId,
            name: normalizedSession.title,
            cliType: normalizedSession.sourceTool,
            workingDirectory: normalizedSession.workingDirectory,
            status: 'imported',
            startedAt: normalizedSession.lastModifiedAt ?? normalizedSession.importedAt,
            stoppedAt: normalizedSession.importedAt,
            lastActivityAt: normalizedSession.lastModifiedAt ?? normalizedSession.importedAt,
            logTail: normalizedSession.transcript.split(/\r?\n/).slice(-12),
            metadata: {
                source: 'session_import',
                sourceTool: normalizedSession.sourceTool,
                importedSessionId: imported.id,
                sourcePath: normalizedSession.sourcePath,
                sessionFormat: normalizedSession.sessionFormat,
            },
        });

        for (const memory of parsedMemories) {
            await this.memoryService.addLongTerm(memory.content, 'project', {
                source: 'session_import',
                sourceTool: normalizedSession.sourceTool,
                importedSessionId: imported.id,
                sessionId: normalizedSession.sessionId,
                path: normalizedSession.sourcePath,
                tags: memory.tags,
                section: memory.kind === 'instruction' ? 'instructions' : 'general',
                memoryKind: memory.kind,
                extractionSource: memory.source,
            });
        }

        return imported;
    }

    private buildNormalizedSession(candidate: DiscoveryCandidate, transcript: string, transcriptHash: string): NormalizedImportedSession {
        const importedAt = Date.now();
        const lines = transcript.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
        const sourceFilePath = candidate.sourcePath.split('#')[0];
        const title = sanitizeSentence(candidate.title || lines[0] || path.basename(sourceFilePath), 120) || path.basename(sourceFilePath);
        const excerpt = sanitizeSentence(lines.slice(0, 3).join(' ').slice(0, 320), 320);
        const externalSessionId = candidate.externalSessionId ?? path.basename(sourceFilePath, path.extname(sourceFilePath));
        const sessionId = `imported:${candidate.sourceTool}:${transcriptHash.slice(0, 12)}`;
        const workingDirectory = candidate.workingDirectory ?? path.dirname(sourceFilePath);

        return {
            sessionId,
            title,
            sourceTool: candidate.sourceTool,
            sourcePath: candidate.sourcePath,
            sessionFormat: candidate.sessionFormat,
            externalSessionId,
            workingDirectory,
            transcript,
            excerpt,
            transcriptHash,
            discoveredAt: importedAt,
            importedAt,
            lastModifiedAt: candidate.lastModifiedAt,
            normalizedSession: {
                sessionId,
                title,
                sourceTool: candidate.sourceTool,
                sourcePath: candidate.sourcePath,
                sessionFormat: candidate.sessionFormat,
                externalSessionId,
                transcriptHash,
                lineCount: lines.length,
                contentLength: transcript.length,
                excerpt,
                importedAt,
                lastModifiedAt: candidate.lastModifiedAt,
            },
            metadata: {
                ...(candidate.metadata ?? {}),
                sourceTool: candidate.sourceTool,
                sourcePath: candidate.sourcePath,
                sessionFormat: candidate.sessionFormat,
                contentLength: transcript.length,
                lineCount: lines.length,
            },
        };
    }

    private async analyzeImportedSession(session: NormalizedImportedSession): Promise<ImportedSessionAnalysis> {
        const transcriptTail = session.transcript.length > 16_000
            ? session.transcript.slice(-16_000)
            : session.transcript;
        const heuristicMemories = heuristicMemoryExtraction(session.transcript, session.sourceTool);

        if (!process.env.OPENAI_API_KEY?.trim()) {
            return {
                memories: heuristicMemories,
                retentionSummary: buildRetentionSummary(session, heuristicMemories, transcriptTail.length, 'heuristic'),
            };
        }

        const prompt = `
You are TormentNexus's session-import memory extractor and archive-retention analyst.
Given an imported transcript from ${session.sourceTool}, extract up to 6 durable technical memories or operator instructions and summarize what should remain archive-only.
Project context:
- working directory: ${session.workingDirectory}
- source path: ${session.sourcePath}
- session title: ${session.title}
- session format: ${session.sessionFormat}
Return JSON only as an object:
{
  "memories": [
    {
      "fact": "Use port 4000 for the TormentNexus control plane.",
      "tags": ["networking", "runtime"],
      "kind": "instruction"
    }
  ],
  "retention": {
    "summary": "Promote operator defaults; keep the rest in compressed archive storage.",
    "salientTags": ["networking", "runtime"],
    "keepArchivedCategories": ["conversational-context", "historical-trace"],
    "discardableCategories": ["greetings", "duplicate-restatements"]
  }
}

Rules:
- Only emit durable facts, decisions, defaults, paths, ports, operational guidance, or architectural instructions.
- kind must be "memory" or "instruction".
- Return an empty memories array if nothing durable exists.
- Assume the full transcript remains archived in compressed form.

Transcript:
${transcriptTail}
        `.trim();

        try {
            const response = await this.llmService.generateText('openrouter', DEFAULT_OPENROUTER_FREE_MODEL, 'Extract durable memories and instructions.', prompt);
            const jsonText = typeof response.content === 'string' ? response.content : '';
            const start = jsonText.indexOf('{');
            const end = jsonText.lastIndexOf('}');

            if (start !== -1 && end !== -1) {
                const parsed = JSON.parse(jsonText.slice(start, end + 1)) as {
                    memories?: Array<{
                        fact?: string;
                        tags?: string[];
                        kind?: string;
                    }>;
                    retention?: {
                        summary?: string;
                        salientTags?: string[];
                        keepArchivedCategories?: string[];
                        discardableCategories?: string[];
                    };
                };

                const normalized = (parsed.memories ?? [])
                    .filter((entry) => typeof entry.fact === 'string' && entry.fact.trim().length > 0)
                    .map((entry) => {
                        const fact = sanitizeSentence(entry.fact as string, 240);
                        return {
                            kind: entry.kind === 'instruction' ? 'instruction' : classifyMemoryKind(fact),
                            content: fact,
                            tags: uniqueStrings([...(entry.tags ?? []), ...deriveTags(fact, session.sourceTool)], 8),
                            source: 'llm' as ImportedSessionMemorySource,
                            metadata: {
                                extraction: 'llm',
                            },
                        };
                    });

                return {
                    memories: normalized,
                    retentionSummary: buildRetentionSummary(
                        session,
                        normalized,
                        transcriptTail.length,
                        'llm',
                        {
                            summary: parsed.retention?.summary,
                            salientTags: uniqueStrings(
                                [
                                    ...(parsed.retention?.salientTags ?? []),
                                    ...normalized.flatMap((memory) => memory.tags ?? []),
                                ],
                                10,
                            ),
                            keepArchivedCategories: parsed.retention?.keepArchivedCategories,
                            discardableCategories: parsed.retention?.discardableCategories,
                        },
                    ),
                };
            }
        } catch (error) {
            console.warn(`[SessionImport] LLM extraction failed for ${session.sourcePath}: ${error instanceof Error ? error.message : String(error)}`);
        }

        return {
            memories: heuristicMemories,
            retentionSummary: buildRetentionSummary(session, heuristicMemories, transcriptTail.length, 'heuristic'),
        };
    }

    private async writeInstructionDocs(): Promise<string | null> {
        let instructions;
        try {
            instructions = this.store.listInstructionMemories(250);
        } catch (error) {
            if (isSqliteUnavailableError(error)) {
                console.warn(formatOptionalSqliteFailure(
                    '[SessionImport] Skipping imported instruction docs for this run',
                    error,
                ));
                return null;
            }

            throw error;
        }

        if (instructions.length === 0) {
            return null;
        }

        await fs.mkdir(this.docsDir, { recursive: true });
        const docPath = path.join(this.docsDir, 'auto-imported-agent-instructions.md');
        const generatedAt = new Date().toISOString();
        const lines = [
            '# Auto-imported Agent Instructions',
            '',
            `Generated at: ${generatedAt}`,
            '',
            'These instructions were derived automatically from imported sessions across supported tools.',
            '',
            '## Durable instructions',
            '',
            ...instructions.map((instruction) => {
                const sourceTool = typeof instruction.metadata?.sourceTool === 'string' ? instruction.metadata.sourceTool : 'unknown';
                const sourcePath = typeof instruction.metadata?.path === 'string' ? instruction.metadata.path : 'unknown source';
                return `- **${sourceTool}** — ${instruction.content} _(source: \`${sourcePath}\`)_`;
            }),
            '',
        ];

        await fs.writeFile(docPath, `${lines.join('\n')}\n`, 'utf-8');
        return docPath;
    }
}
