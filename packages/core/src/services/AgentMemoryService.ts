/**
 * Agent Memory Service - Tiered Memory System for Persistent Agent Context
 * 
 * Implements multi-tier memory architecture inspired by Mem0 and Letta:
 * - Session Memory: Ephemeral context within a conversation
 * - Working Memory: Task-relevant facts extracted during execution
 * - Long-term Memory: Persistent learnings across sessions
 * 
 * Features:
 * - Automatic memory extraction from conversations
 * - Relevance-based retrieval with temporal decay
 * - User/Agent/Project namespaces
 * - Memory consolidation (working -> long-term)
 */

import * as fs from 'fs';
import * as path from 'path';
import * as crypto from 'crypto';
import { MemoryManager } from './MemoryManager.js';
import {
    buildSessionSummaryContent,
    buildStructuredSessionSummary,
    getStructuredSessionSummary,
    type SessionSummaryInput,
} from './sessionSummaryMemory.js';
import { buildSessionBootstrapPrompt } from './sessionBootstrapMemory.js';
import { buildToolContextPayload } from './toolContextMemory.js';
import {
    buildStructuredUserPrompt,
    buildUserPromptContent,
    getStructuredUserPrompt,
    type UserPromptInput,
    type UserPromptRole,
} from './sessionPromptMemory.js';
import { searchMemoryRecordsByPivot } from './agentMemoryPivot.js';
import { getMemoryTimelineWindow } from './agentMemoryTimeline.js';
import { getCrossSessionMemoryLinks, type CrossSessionMemoryLink } from './agentMemoryConnections.js';

// Memory types
export type MemoryType = 'session' | 'working' | 'long_term';
export type MemoryNamespace = 'user' | 'agent' | 'project';

export interface Memory {
    id: string;
    type: MemoryType;
    namespace: MemoryNamespace;
    content: string;
    media?: MemoryMedia[]; // Multi-modal support
    metadata: MemoryMetadata;
    createdAt: Date;
    accessedAt: Date;
    accessCount: number;
    score?: number;  // Relevance score for retrieval
    ttl?: number;    // Time-to-live in ms (for session memory)
}

export interface MemoryMedia {
    type: 'image' | 'audio' | 'video' | 'file';
    url?: string;        // External or Local path
    dataUrl?: string;   // Base64 encoded
    blobHash?: string;  // Reference to binary storage
    mimeType?: string;
    description?: string; // OCR text or AI caption
}

export interface MemoryMetadata {
    source?: string;          // Where memory came from (tool, conversation, etc)
    tags?: string[];          // Categorization tags
    relatedMemories?: string[]; // Links to other memories
    confidence?: number;      // Extraction confidence (0-1)
    userId?: string;          // Associated user
    projectId?: string;       // Associated project
    sessionId?: string;       // Associated session ID
    [key: string]: unknown;
}

export interface MemorySearchOptions {
    namespace?: MemoryNamespace;
    type?: MemoryType;
    limit?: number;
    minScore?: number;
    tags?: string[];
    includeExpired?: boolean;
}

export interface MemoryServiceOptions {
    persistDir: string;
    sessionTTL?: number;        // Default session TTL (30 min)
    consolidationThreshold?: number;  // Access count to promote to long-term
    maxSessionMemories?: number;
    maxWorkingMemories?: number;
}

export type ObservationType = 'discovery' | 'decision' | 'progress' | 'warning' | 'fix';

export interface StructuredObservation {
    type: ObservationType;
    title: string;
    subtitle?: string;
    narrative: string;
    facts: string[];
    concepts: string[];
    filesRead: string[];
    filesModified: string[];
    toolName?: string;
    contentHash: string;
    recordedAt: number;
}

export interface ObservationInput {
    toolName?: string;
    title?: string;
    subtitle?: string;
    narrative?: string;
    rawInput?: unknown;
    rawOutput?: unknown;
    facts?: string[];
    concepts?: string[];
    filesRead?: string[];
    filesModified?: string[];
    type?: ObservationType;
    namespace?: MemoryNamespace;
    metadata?: MemoryMetadata;
}

export interface ObservationSearchOptions {
    namespace?: MemoryNamespace;
    type?: ObservationType;
    limit?: number;
    tags?: string[];
}

export interface ToolContextInput {
    toolName: string;
    args?: unknown;
    activeGoal?: string | null;
    lastObjective?: string | null;
}

export interface UserPromptSearchOptions {
    role?: UserPromptRole;
    limit?: number;
}

export type MemoryPivotKind = 'session' | 'tool' | 'concept' | 'file' | 'goal' | 'objective';

export interface MemoryPivotSearchInput {
    pivot: MemoryPivotKind;
    value: string;
    limit?: number;
}

export interface MemoryTimelineWindowInput {
    sessionId: string;
    anchorTimestamp: number;
    before?: number;
    after?: number;
}

export interface CrossSessionMemoryLinksInput {
    memoryId: string;
    limit?: number;
}

const OBSERVATION_DEDUP_WINDOW_MS = 30_000;
const OBSERVATION_MAX_FACTS = 5;
const OBSERVATION_MAX_CONCEPTS = 8;

function uniqueStrings(values: Array<string | undefined | null>, maxItems?: number): string[] {
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

        if (typeof maxItems === 'number' && normalized.length >= maxItems) {
            break;
        }
    }

    return normalized;
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
    if (typeof value === 'string') {
        return value.trim().slice(0, maxLength);
    }

    if (value == null) {
        return '';
    }

    const serialized = safeJsonStringify(value);
    return serialized.trim().slice(0, maxLength);
}

function sanitizeSentence(value: string, maxLength: number = 140): string {
    return value.replace(/\s+/g, ' ').trim().slice(0, maxLength);
}

function firstUsefulSentence(value: string): string {
    const sentence = value.split(/(?<=[.!?])\s+/).find(part => part.trim().length > 0) ?? value;
    return sanitizeSentence(sentence);
}

function extractFacts(text: string, providedFacts: string[] = []): string[] {
    if (providedFacts.length > 0) {
        return uniqueStrings(providedFacts, OBSERVATION_MAX_FACTS);
    }

    const lineFacts = text
        .split(/\r?\n/)
        .map(line => line.replace(/^[-*\d.\s]+/, '').trim())
        .filter(line => line.length >= 12);

    if (lineFacts.length > 0) {
        return uniqueStrings(lineFacts, OBSERVATION_MAX_FACTS);
    }

    const sentenceFacts = text
        .split(/(?<=[.!?])\s+/)
        .map(sentence => sanitizeSentence(sentence, 180))
        .filter(sentence => sentence.length >= 12);

    return uniqueStrings(sentenceFacts, OBSERVATION_MAX_FACTS);
}

function inferObservationType(toolName: string | undefined, narrative: string, providedType?: ObservationType): ObservationType {
    if (providedType) {
        return providedType;
    }

    const loweredTool = (toolName ?? '').toLowerCase();
    const loweredNarrative = narrative.toLowerCase();

    if (/(error|failed|exception|warning|timeout)/.test(loweredNarrative)) {
        return 'warning';
    }

    if (/(fix|patch|repair|resolve)/.test(loweredTool) || /(fixed|resolved|patched)/.test(loweredNarrative)) {
        return 'fix';
    }

    if (/(read|grep|search|inspect|list|query)/.test(loweredTool)) {
        return 'discovery';
    }

    if (/(decide|choose|select|plan)/.test(loweredNarrative)) {
        return 'decision';
    }

    return 'progress';
}

function buildObservationTitle(toolName: string | undefined, narrative: string, providedTitle?: string): string {
    if (providedTitle?.trim()) {
        return sanitizeSentence(providedTitle);
    }

    if (toolName?.trim()) {
        return sanitizeSentence(`${toolName}: ${firstUsefulSentence(narrative)}`);
    }

    return firstUsefulSentence(narrative) || 'Recorded observation';
}

function buildObservationConcepts(
    toolName: string | undefined,
    type: ObservationType,
    metadata: MemoryMetadata,
    providedConcepts: string[] = [],
): string[] {
    const tagValues = Array.isArray(metadata.tags) ? metadata.tags : [];
    const toolTokens = (toolName ?? '')
        .split(/[^a-zA-Z0-9]+/)
        .map(token => token.trim().toLowerCase())
        .filter(token => token.length >= 3);

    return uniqueStrings([
        ...providedConcepts,
        ...tagValues,
        ...toolTokens,
        type,
    ], OBSERVATION_MAX_CONCEPTS);
}

function normalizeFileList(values: string[] = []): string[] {
    return uniqueStrings(values.map(value => value.replace(/\\/g, '/')), 20);
}

function buildObservationContent(observation: StructuredObservation): string {
    const factsBlock = observation.facts.length > 0
        ? `\nFacts:\n${observation.facts.map(fact => `- ${fact}`).join('\n')}`
        : '';
    const filesBlock = observation.filesRead.length > 0 || observation.filesModified.length > 0
        ? `\nFiles:\n${[
            ...observation.filesRead.map(file => `- read: ${file}`),
            ...observation.filesModified.map(file => `- modified: ${file}`),
        ].join('\n')}`
        : '';

    return `${observation.title}\n${observation.narrative}${factsBlock}${filesBlock}`.trim();
}

function getStructuredObservation(metadata: MemoryMetadata): StructuredObservation | null {
    const candidate = metadata.structuredObservation;

    if (!candidate || typeof candidate !== 'object') {
        return null;
    }

    const observation = candidate as Partial<StructuredObservation>;
    if (typeof observation.title !== 'string' || typeof observation.narrative !== 'string' || typeof observation.contentHash !== 'string') {
        return null;
    }

    const type = observation.type;
    if (type !== 'discovery' && type !== 'decision' && type !== 'progress' && type !== 'warning' && type !== 'fix') {
        return null;
    }

    return {
        type,
        title: observation.title,
        subtitle: typeof observation.subtitle === 'string' ? observation.subtitle : undefined,
        narrative: observation.narrative,
        facts: Array.isArray(observation.facts) ? observation.facts.filter((fact): fact is string => typeof fact === 'string') : [],
        concepts: Array.isArray(observation.concepts) ? observation.concepts.filter((concept): concept is string => typeof concept === 'string') : [],
        filesRead: Array.isArray(observation.filesRead) ? observation.filesRead.filter((file): file is string => typeof file === 'string') : [],
        filesModified: Array.isArray(observation.filesModified) ? observation.filesModified.filter((file): file is string => typeof file === 'string') : [],
        toolName: typeof observation.toolName === 'string' ? observation.toolName : undefined,
        contentHash: observation.contentHash,
        recordedAt: typeof observation.recordedAt === 'number' ? observation.recordedAt : Date.now(),
    };
}

/**
 * Reason: Vector search metadata is flexible and can include unknown values.
 * What: Narrows metadata into safe, optional `type`/`namespace` values for memory reconstruction.
 * Why: Preserves behavior while removing broad untyped casts during search-result hydration.
 */
function parseMemoryMetadata(metadata: unknown): {
    normalized: MemoryMetadata;
    type?: MemoryType;
    namespace?: MemoryNamespace;
    createdAt?: number;
} {
    if (!metadata || typeof metadata !== 'object') {
        return { normalized: {} };
    }

    const normalized = metadata as MemoryMetadata;
    const typeValue = normalized.type;
    const namespaceValue = normalized.namespace;
    const createdAtValue = normalized.createdAt;

    const type = typeValue === 'session' || typeValue === 'working' || typeValue === 'long_term'
        ? typeValue
        : undefined;
    const namespace = namespaceValue === 'user' || namespaceValue === 'agent' || namespaceValue === 'project'
        ? namespaceValue
        : undefined;
    const createdAt = typeof createdAtValue === 'number' ? createdAtValue : undefined;

    return { normalized, type, namespace, createdAt };
}

/**
 * Simple in-memory vector similarity using TF-IDF-like approach
 * Production would use proper embeddings
 */
// SimpleVectorSearch replaced by MemoryManager

/**
 * Agent Memory Service - Main service class
 */
export class AgentMemoryService {
    private memories: Map<string, Memory> = new Map();
    private memoryManager: MemoryManager;
    private options: Required<MemoryServiceOptions>;
    private dirty = false;

    constructor(options: MemoryServiceOptions, memoryManager?: MemoryManager) {
        this.options = {
            persistDir: options.persistDir,
            sessionTTL: options.sessionTTL ?? 30 * 60 * 1000,  // 30 minutes
            consolidationThreshold: options.consolidationThreshold ?? 5,
            maxSessionMemories: options.maxSessionMemories ?? 100,
            maxWorkingMemories: options.maxWorkingMemories ?? 500,
        };

        if (memoryManager) {
            this.memoryManager = memoryManager;
        } else {
            // Initialize MemoryManager (parent of persistDir is workspaceRoot)
            const workspaceRoot = path.dirname(this.options.persistDir);
            this.memoryManager = new MemoryManager(workspaceRoot);
        }

        this.loadFromDisk();

        // Auto-Pickup: If this is a fresh session (no session memories), 
        // try to automatically load the most recent handoff.
        const sessionMemories = Array.from(this.memories.values()).filter(m => m.type === 'session');
        if (sessionMemories.length === 0) {
            this.autoPickupLatestHandoff().catch(e => 
                console.error("[AgentMemoryService] Auto-pickup failed:", e)
            );
        }

        this.startAutoHandoff();
    }

    private async autoPickupLatestHandoff() {
        try {
            const handoffDir = path.join(this.options.persistDir, '..', 'handoffs');
            if (!fs.existsSync(handoffDir)) return;

            const files = fs.readdirSync(handoffDir)
                .filter(f => f.startsWith('handoff_'))
                .sort((a, b) => b.localeCompare(a));

            if (files.length > 0) {
                const latestFile = files[0];
                const content = fs.readFileSync(path.join(handoffDir, latestFile), 'utf-8');
                const res = await this.pickupSession(content);
                if (res.success) {
                    console.log(`[AgentMemoryService] 🔄 Auto-picked up latest handoff: ${latestFile} (${res.count} items)`);
                }
            }
        } catch (e) {
            // Silent fail for auto-pickup
        }
    }

    /**
     * Generate unique memory ID
     */
    private generateId(): string {
        return crypto.randomBytes(8).toString('hex');
    }

    /**
     * Load memories from disk
     */
    private loadFromDisk(): void {
        const filePath = path.join(this.options.persistDir, 'memories.json');

        if (!fs.existsSync(this.options.persistDir)) {
            fs.mkdirSync(this.options.persistDir, { recursive: true });
        }

        if (fs.existsSync(filePath)) {
            try {
                const data = JSON.parse(fs.readFileSync(filePath, 'utf-8'));
                for (const mem of data.memories || []) {
                    mem.createdAt = new Date(mem.createdAt);
                    mem.accessedAt = new Date(mem.accessedAt);

                    if (mem.type === 'session' && mem.ttl) {
                        const age = Date.now() - mem.createdAt.getTime();
                        if (age > mem.ttl) continue;
                    }

                    this.memories.set(mem.id, mem);
                    // No need to manually add to vectorIndex here, 
                    // MemoryManager handles persistence of long-term/working.
                }
                console.log(`[AgentMemoryService] Loaded ${this.memories.size} session memories`);
            } catch (e) {
                console.error('[AgentMemoryService] Failed to load memories:', e);
            }
        }
    }

    /**
     * Save memories to disk
     */
    private saveToDisk(): void {
        if (!this.dirty) return;

        const filePath = path.join(this.options.persistDir, 'memories.json');
        const data = {
            version: 1,
            savedAt: new Date().toISOString(),
            memories: Array.from(this.memories.values()),
        };

        fs.writeFileSync(filePath, JSON.stringify(data, null, 2));
        this.dirty = false;
    }

    /**
     * Add a memory
     */
    async add(
        content: string,
        type: MemoryType,
        namespace: MemoryNamespace,
        metadata: MemoryMetadata = {}
    ): Promise<Memory> {
        const memory: Memory = {
            id: this.generateId(),
            type,
            namespace,
            content,
            metadata,
            createdAt: new Date(),
            accessedAt: new Date(),
            accessCount: 0,
            ttl: type === 'session' ? this.options.sessionTTL : undefined,
        };

        this.memories.set(memory.id, memory);
        this.dirty = true;

        // If working or long-term, also persist to vector DB
        if (type !== 'session') {
            await this.memoryManager.saveContext(content, {
                ...metadata,
                id: memory.id,
                type,
                namespace,
                createdAt: memory.createdAt.getTime()
            });
        }

        this.enforceMemoryLimits();
        this.scheduleSave();
        return memory;
    }

    /**
     * Get a memory by ID
     */
    get(id: string): Memory | null {
        const memory = this.memories.get(id);
        if (!memory) return null;

        // Check expiration
        if (memory.type === 'session' && memory.ttl) {
            const age = Date.now() - memory.createdAt.getTime();
            if (age > memory.ttl) {
                this.delete(id);
                return null;
            }
        }

        // Update access metadata
        memory.accessedAt = new Date();
        memory.accessCount++;
        this.dirty = true;

        // Check for consolidation
        this.checkConsolidation(memory);

        return memory;
    }

    /**
     * Delete a memory
     */
    async delete(id: string): Promise<boolean> {
        const existed = this.memories.delete(id);
        if (existed) {
            this.dirty = true;
        }

        // Also delete from vector DB
        try {
            await this.memoryManager.deleteContext(id);
        } catch (e) {
            // Might not exist in vector DB if it was only a session memory
        }

        return existed;
    }

    /**
     * Search memories using hybrid strategy (Local sessions + Vector DB)
     */
    async search(query: string, options: MemorySearchOptions = {}): Promise<Memory[]> {
        const limit = options.limit ?? 10;

        // 1. Get results from Vector DB
        const vectorResults = await this.memoryManager.search(query, limit * 2);
        const mappedResults: Memory[] = vectorResults.map(r => {
            const parsed = parseMemoryMetadata(r.metadata);
            return {
                id: r.id,
                content: r.content,
                type: parsed.type ?? 'long_term',
                namespace: parsed.namespace ?? 'project',
                metadata: parsed.normalized,
                createdAt: new Date(parsed.createdAt ?? Date.now()),
                accessedAt: new Date(),
                accessCount: 0, // We could track this in metadata if needed
                score: r.score
            };
        });

        // 2. Add current session memories (which might not be indexed yet)
        const sessionMemories = Array.from(this.memories.values())
            .filter(m => m.type === 'session');

        // For session memories, we do a simple title/content match if needed, 
        // or just include the most recent ones. 
        // Better: Index session memories too? For now, we rely on the fact that 
        // they are small.

        let results = [...mappedResults, ...sessionMemories];

        // Filter by type/namespace
        if (options.type) {
            results = results.filter(m => m.type === options.type);
        }
        if (options.namespace) {
            results = results.filter(m => m.namespace === options.namespace);
        }
        if (options.tags?.length) {
            results = results.filter(m =>
                options.tags!.some(tag => m.metadata.tags?.includes(tag))
            );
        }

        // Simple relevance boost for session memories if they mention query terms
        const queryTerms = query.toLowerCase().split(/\s+/).filter(t => t.length > 2);
        for (const res of results) {
            if (res.type === 'session' && !res.score) {
                let sessionScore = 0;
                for (const term of queryTerms) {
                    if (res.content.toLowerCase().includes(term)) sessionScore += 1;
                }
                res.score = sessionScore;
            }
        }

        // Apply temporal decay
        const now = Date.now();
        for (const memory of results) {
            const ageHours = (now - memory.accessedAt.getTime()) / (1000 * 60 * 60);
            const decayFactor = Math.exp(-ageHours / 24);
            memory.score = (memory.score ?? 0) * (0.5 + 0.5 * decayFactor);
        }

        return results
            .sort((a, b) => (b.score ?? 0) - (a.score ?? 0))
            .slice(0, limit);
    }

    /**
     * Get recent memories
     */
    getRecent(limit: number = 10, options: MemorySearchOptions = {}): Memory[] {
        let memories = Array.from(this.memories.values());

        if (options.type) {
            memories = memories.filter(m => m.type === options.type);
        }
        if (options.namespace) {
            memories = memories.filter(m => m.namespace === options.namespace);
        }

        return memories
            .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
            .slice(0, limit);
    }

    /**
     * Add session memory (ephemeral)
     */
    async addSession(content: string, metadata: MemoryMetadata = {}): Promise<Memory> {
        return await this.add(content, 'session', 'agent', metadata);
    }

    /**
     * Add working memory (task-relevant)
     */
    async addWorking(content: string, namespace: MemoryNamespace = 'project', metadata: MemoryMetadata = {}): Promise<Memory> {
        return await this.add(content, 'working', namespace, metadata);
    }

    /**
     * Add long-term memory (persistent)
     */
    async addLongTerm(content: string, namespace: MemoryNamespace = 'project', metadata: MemoryMetadata = {}): Promise<Memory> {
        return await this.add(content, 'long_term', namespace, metadata);
    }

    /**
     * Record a structured observation inspired by the tormentnexus observation pipeline.
     * This is intentionally heuristic for now: it normalizes raw tool output into a typed,
     * deduplicated working-memory record that TormentNexus can search immediately.
     */
    async recordObservation(input: ObservationInput): Promise<Memory> {
        const namespace = input.namespace ?? 'project';
        const metadata = input.metadata ?? {};
        const rawInputText = normalizeText(input.rawInput, 400);
        const rawOutputText = normalizeText(input.rawOutput, 1200);
        const narrative = sanitizeSentence(
            input.narrative
            ?? rawOutputText
            ?? rawInputText
            ?? input.title
            ?? 'Recorded observation',
            600,
        );
        const type = inferObservationType(input.toolName, narrative, input.type);
        const title = buildObservationTitle(input.toolName, narrative, input.title);
        const facts = extractFacts(rawOutputText || narrative, input.facts);
        const concepts = buildObservationConcepts(input.toolName, type, metadata, input.concepts);
        const filesRead = normalizeFileList(input.filesRead);
        const filesModified = normalizeFileList(input.filesModified);

        const contentHash = crypto.createHash('sha256').update(safeJsonStringify({
            toolName: input.toolName ?? null,
            title,
            narrative,
            rawInput: rawInputText,
            rawOutput: rawOutputText,
            facts,
            concepts,
            filesRead,
            filesModified,
            type,
        })).digest('hex');

        const existing = Array.from(this.memories.values()).find(memory => {
            const observation = getStructuredObservation(memory.metadata);
            if (!observation) return false;
            if (observation.contentHash !== contentHash) return false;
            return Date.now() - memory.createdAt.getTime() <= OBSERVATION_DEDUP_WINDOW_MS;
        });

        if (existing) {
            existing.accessedAt = new Date();
            existing.accessCount += 1;
            this.dirty = true;
            this.scheduleSave();
            return existing;
        }

        const observation: StructuredObservation = {
            type,
            title,
            subtitle: input.subtitle ? sanitizeSentence(input.subtitle) : undefined,
            narrative,
            facts,
            concepts,
            filesRead,
            filesModified,
            toolName: input.toolName,
            contentHash,
            recordedAt: Date.now(),
        };

        const memoryMetadata: MemoryMetadata = {
            ...metadata,
            source: metadata.source ?? input.toolName ?? 'observation',
            tags: uniqueStrings([...(metadata.tags ?? []), ...concepts]),
            structuredObservation: observation,
            observationType: type,
            observationHash: contentHash,
            toolName: input.toolName,
            filesRead,
            filesModified,
            rawInput: rawInputText || undefined,
            rawOutput: rawOutputText || undefined,
        };

        return await this.add(buildObservationContent(observation), 'working', namespace, memoryMetadata);
    }

    async captureSessionSummary(input: SessionSummaryInput): Promise<Memory> {
        const metadata = (input.metadata ?? {}) as MemoryMetadata;
        const logTail = uniqueStrings((input.logTail ?? []).map((line) => sanitizeSentence(line, 240)), 12);
        const structuredSummary = buildStructuredSessionSummary(input, Date.now(), logTail);

        const existing = Array.from(this.memories.values()).find((memory) => {
            const summary = getStructuredSessionSummary(memory.metadata);
            if (!summary) return false;
            if (summary.contentHash !== structuredSummary.contentHash) return false;
            return Date.now() - memory.createdAt.getTime() <= OBSERVATION_DEDUP_WINDOW_MS;
        });

        if (existing) {
            existing.accessedAt = new Date();
            existing.accessCount += 1;
            this.dirty = true;
            this.scheduleSave();
            return existing;
        }

        return await this.add(buildSessionSummaryContent(structuredSummary), 'long_term', 'project', {
            ...metadata,
            source: metadata.source ?? 'session_summary',
            section: metadata.section ?? 'general',
            tags: uniqueStrings([...(Array.isArray(metadata.tags) ? metadata.tags : []), 'session-summary', structuredSummary.cliType, structuredSummary.status]),
            sessionId: structuredSummary.sessionId,
            structuredSessionSummary: structuredSummary,
            memoryKind: 'session_summary',
        });
    }

    async captureUserPrompt(input: UserPromptInput & { metadata?: MemoryMetadata }): Promise<Memory> {
        const metadata = input.metadata ?? {};
        const content = normalizeText(input.content, 400);
        const role = input.role ?? 'prompt';
        const contentHash = crypto.createHash('sha256').update(safeJsonStringify({
            content,
            role,
            sessionId: input.sessionId ?? null,
            activeGoal: input.activeGoal ?? null,
            lastObjective: input.lastObjective ?? null,
        })).digest('hex');

        const existing = Array.from(this.memories.values()).find((memory) => {
            const prompt = getStructuredUserPrompt(memory.metadata);
            if (!prompt) return false;
            if (prompt.contentHash !== contentHash) return false;
            return Date.now() - memory.createdAt.getTime() <= OBSERVATION_DEDUP_WINDOW_MS;
        });

        if (existing) {
            existing.accessedAt = new Date();
            existing.accessCount += 1;
            this.dirty = true;
            this.scheduleSave();
            return existing;
        }

        const promptNumber = this.getRecentUserPrompts(100).length + 1;
        const structuredPrompt = buildStructuredUserPrompt({
            ...input,
            content,
            role,
            promptNumber,
            recordedAt: Date.now(),
            contentHash,
        });

        return await this.add(buildUserPromptContent(structuredPrompt), 'long_term', 'project', {
            ...metadata,
            source: metadata.source ?? 'user_prompt',
            sessionId: structuredPrompt.sessionId,
            tags: uniqueStrings([...(Array.isArray(metadata.tags) ? metadata.tags : []), 'user-prompt', role]),
            structuredUserPrompt: structuredPrompt,
            memoryKind: 'user_prompt',
            promptRole: role,
        });
    }

    async searchSessionSummaries(query: string, limit: number = 10): Promise<Memory[]> {
        const results = await this.search(query, {
            limit,
            namespace: 'project',
            type: 'long_term',
            tags: ['session-summary'],
        });

        return results.filter((memory) => getStructuredSessionSummary(memory.metadata) !== null);
    }

    getRecentSessionSummaries(limit: number = 10): Memory[] {
        return Array.from(this.memories.values())
            .filter((memory) => memory.type === 'long_term' && getStructuredSessionSummary(memory.metadata) !== null)
            .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
            .slice(0, limit);
    }

    getSessionBootstrap(options: {
        activeGoal?: string | null;
        lastObjective?: string | null;
        toolAdvertisementLines?: string[];
    } = {}) {
        const summaries = this.getRecentSessionSummaries(3);
        const observations = this.getRecentObservations(5);

        return buildSessionBootstrapPrompt({
            activeGoal: options.activeGoal,
            lastObjective: options.lastObjective,
            toolAdvertisementLines: options.toolAdvertisementLines,
            summaries: summaries.map((memory) => ({ content: memory.content })),
            observations: observations.map((memory) => {
                const observation = getStructuredObservation(memory.metadata);
                return {
                    title: observation?.title,
                    narrative: observation?.narrative,
                    type: observation?.type,
                    toolName: observation?.toolName,
                    content: memory.content,
                };
            }),
        });
    }

    async searchUserPrompts(query: string, options: UserPromptSearchOptions = {}): Promise<Memory[]> {
        const results = await this.search(query, {
            limit: options.limit,
            namespace: 'project',
            type: 'long_term',
            tags: ['user-prompt'],
        });

        return results.filter((memory) => {
            const prompt = getStructuredUserPrompt(memory.metadata);
            if (!prompt) return false;
            if (options.role && prompt.role !== options.role) return false;
            return true;
        });
    }

    getRecentUserPrompts(limit: number = 10, options: UserPromptSearchOptions = {}): Memory[] {
        return Array.from(this.memories.values())
            .filter((memory) => {
                if (memory.type !== 'long_term' || memory.namespace !== 'project') {
                    return false;
                }
                const prompt = getStructuredUserPrompt(memory.metadata);
                if (!prompt) return false;
                if (options.role && prompt.role !== options.role) return false;
                return true;
            })
            .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
            .slice(0, limit);
    }

    getToolContext(input: ToolContextInput) {
        const observations = Array.from(this.memories.values())
            .map((memory) => {
                const observation = getStructuredObservation(memory.metadata);
                if (!observation) {
                    return null;
                }

                return {
                    title: observation.title,
                    narrative: observation.narrative,
                    content: memory.content,
                    type: observation.type,
                    toolName: observation.toolName,
                    concepts: observation.concepts,
                    filesRead: observation.filesRead,
                    filesModified: observation.filesModified,
                    recordedAt: observation.recordedAt,
                };
            })
            .filter((observation): observation is NonNullable<typeof observation> => observation !== null);

        const summaries = Array.from(this.memories.values())
            .map((memory) => {
                const summary = getStructuredSessionSummary(memory.metadata);
                if (!summary) {
                    return null;
                }

                return {
                    content: memory.content,
                    cliType: summary.cliType,
                    status: summary.status,
                    sessionId: summary.sessionId,
                    recordedAt: summary.stoppedAt ?? memory.createdAt.getTime(),
                };
            })
            .filter((summary): summary is NonNullable<typeof summary> => summary !== null);

        return buildToolContextPayload({
            toolName: input.toolName,
            args: input.args,
            activeGoal: input.activeGoal,
            lastObjective: input.lastObjective,
            observations,
            summaries,
        });
    }

    async searchObservations(query: string, options: ObservationSearchOptions = {}): Promise<Memory[]> {
        const results = await this.search(query, {
            namespace: options.namespace,
            limit: options.limit,
            tags: options.tags,
            type: 'working',
        });

        return results.filter(memory => {
            const observation = getStructuredObservation(memory.metadata);
            if (!observation) return false;
            if (options.type && observation.type !== options.type) return false;
            if (options.namespace && memory.namespace !== options.namespace) return false;
            return true;
        });
    }

    getRecentObservations(limit: number = 10, options: ObservationSearchOptions = {}): Memory[] {
        return Array.from(this.memories.values())
            .filter(memory => {
                const observation = getStructuredObservation(memory.metadata);
                if (!observation) return false;
                if (options.namespace && memory.namespace !== options.namespace) return false;
                if (options.type && observation.type !== options.type) return false;
                return true;
            })
            .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
            .slice(0, limit);
    }

    searchByPivot(input: MemoryPivotSearchInput): Memory[] {
        return searchMemoryRecordsByPivot(Array.from(this.memories.values()), input.pivot, input.value, input.limit ?? 20);
    }

    getTimelineWindow(input: MemoryTimelineWindowInput): Memory[] {
        return getMemoryTimelineWindow(
            Array.from(this.memories.values()),
            input.sessionId,
            input.anchorTimestamp,
            input.before ?? 3,
            input.after ?? 3,
        );
    }

    getCrossSessionLinks(input: CrossSessionMemoryLinksInput): Array<CrossSessionMemoryLink<Memory>> {
        return getCrossSessionMemoryLinks(Array.from(this.memories.values()), input.memoryId, input.limit ?? 5);
    }

    /**
     * Check if memory should be consolidated to long-term
     */
    private async checkConsolidation(memory: Memory): Promise<void> {
        if (memory.type === 'working' &&
            memory.accessCount >= this.options.consolidationThreshold) {
            // Promote to long-term memory
            memory.type = 'long_term';
            memory.ttl = undefined;
            this.dirty = true;

            // Sync update to vector DB
            await this.memoryManager.saveContext(memory.content, {
                ...memory.metadata,
                id: memory.id,
                type: 'long_term',
                consolidatedAt: Date.now()
            });

            console.log(`[AgentMemoryService] Consolidated memory ${memory.id} to long-term`);
        }
    }

    /**
     * Enforce memory limits by removing old memories
     */
    private enforceMemoryLimits(): void {
        const session = this.getByType('session');
        const working = this.getByType('working');

        // Remove excess session memories (oldest first)
        if (session.length > this.options.maxSessionMemories) {
            const toRemove = session
                .sort((a, b) => a.createdAt.getTime() - b.createdAt.getTime())
                .slice(0, session.length - this.options.maxSessionMemories);
            for (const m of toRemove) {
                this.delete(m.id);
            }
        }

        // Remove excess working memories (least accessed first)
        if (working.length > this.options.maxWorkingMemories) {
            const toRemove = working
                .sort((a, b) => a.accessCount - b.accessCount)
                .slice(0, working.length - this.options.maxWorkingMemories);
            for (const m of toRemove) {
                this.delete(m.id);
            }
        }
    }

    /**
     * Get memories by type
     */
    getByType(type: MemoryType): Memory[] {
        return Array.from(this.memories.values()).filter(m => m.type === type);
    }

    /**
     * Get memories by namespace
     */
    getByNamespace(namespace: MemoryNamespace): Memory[] {
        return Array.from(this.memories.values()).filter(m => m.namespace === namespace);
    }

    /**
     * Clear all session memories
     */
    clearSession(): void {
        for (const [id, memory] of this.memories) {
            if (memory.type === 'session') {
                this.memories.delete(id);
            }
        }
        this.dirty = true;
    }

    /**
     * Get memory statistics
     */
    getStats(): Record<string, number> {
        const byType: Record<string, number> = { session: 0, working: 0, long_term: 0 };
        const byNamespace: Record<string, number> = { user: 0, agent: 0, project: 0 };
        const byObservationType: Record<ObservationType, number> = {
            discovery: 0,
            decision: 0,
            progress: 0,
            warning: 0,
            fix: 0,
        };
        let observationCount = 0;
        let promptCount = 0;
        let sessionSummaryCount = 0;
        const observationHashes = new Set<string>();

        for (const memory of this.memories.values()) {
            byType[memory.type]++;
            byNamespace[memory.namespace]++;

            const observation = getStructuredObservation(memory.metadata);
            if (observation) {
                observationCount++;
                byObservationType[observation.type]++;
                observationHashes.add(observation.contentHash);
            }

            if (getStructuredUserPrompt(memory.metadata)) {
                promptCount++;
            }

            if (getStructuredSessionSummary(memory.metadata)) {
                sessionSummaryCount++;
            }
        }

        // We could also mix in stats from memoryManager if needed

        return {
            totalCount: this.memories.size,
            sessionCount: byType.session,
            workingCount: byType.working,
            longTermCount: byType.long_term,
            observationCount,
            uniqueObservationCount: observationHashes.size,
            promptCount,
            sessionSummaryCount,
            ...byType,
            ...byNamespace,
            ...byObservationType,
        };
    }

    /**
     * Export memories to JSON
     */
    export(): string {
        return JSON.stringify({
            exportedAt: new Date().toISOString(),
            memories: Array.from(this.memories.values()),
        }, null, 2);
    }

    /**
     * Import memories from JSON
     */
    async import(jsonData: string): Promise<number> {
        const data = JSON.parse(jsonData);
        let count = 0;

        for (const mem of data.memories || []) {
            mem.createdAt = new Date(mem.createdAt);
            mem.accessedAt = new Date(mem.accessedAt);
            mem.id = this.generateId();  // Assign new ID

            this.memories.set(mem.id, mem);

            if (mem.type !== 'session') {
                await this.memoryManager.saveContext(mem.content, {
                    ...mem.metadata,
                    id: mem.id,
                    type: mem.type
                });
            }
            count++;
        }

        this.dirty = true;
        return count;
    }

    /**
     * Handoff Session — Summarizes the current session and exports it as a portable artifact.
     */
    async handoffSession(metadata: Record<string, any> = {}): Promise<string> {
        const stats = this.getStats();
        const summaries = this.memories.values();
        const recentEvents = Array.from(summaries)
            .filter(m => m.type === 'session')
            .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
            .slice(0, 20);

        const handoffArtifact = {
            version: "0.99.1",
            timestamp: Date.now(),
            sessionId: metadata.sessionId || 'current',
            stats,
            recentContext: recentEvents.map(m => ({
                content: m.content,
                metadata: m.metadata
            })),
            notes: metadata.notes || ''
        };

        return JSON.stringify(handoffArtifact, null, 2);
    }

    /**
     * Pickup Session — Restores context from a handoff artifact.
     */
    async pickupSession(artifactJson: string): Promise<{ success: boolean; count: number }> {
        try {
            const artifact = JSON.parse(artifactJson);
            let count = 0;

            if (artifact.recentContext && Array.isArray(artifact.recentContext)) {
                for (const item of artifact.recentContext) {
                    await this.add(item.content, 'session', 'project', item.metadata);
                    count++;
                }
            }

            return { success: true, count };
        } catch (e) {
            console.error("[AgentMemoryService] Failed to pickup session:", e);
            return { success: false, count: 0 };
        }
    }

    /**
     * Schedule auto-save
     */
    private saveTimeout: NodeJS.Timeout | null = null;
    private handoffInterval: NodeJS.Timeout | null = null;

    private startAutoHandoff() {
        // Every 5 minutes, save a background handoff artifact
        this.handoffInterval = setInterval(async () => {
            try {
                const artifact = await this.handoffSession({ notes: 'Automatic background handoff' });
                const handoffDir = path.join(this.options.persistDir, '..', 'handoffs');
                if (!fs.existsSync(handoffDir)) fs.mkdirSync(handoffDir, { recursive: true });
                
                const filename = `handoff_${Date.now()}.json`;
                fs.writeFileSync(path.join(handoffDir, filename), artifact);
                
                // Keep only last 10 automatic handoffs
                const files = fs.readdirSync(handoffDir)
                    .filter(f => f.startsWith('handoff_'))
                    .sort((a, b) => b.localeCompare(a));
                
                if (files.length > 10) {
                    files.slice(10).forEach(f => fs.unlinkSync(path.join(handoffDir, f)));
                }

                // Autonomous Context Compacting:
                // If active session memory is getting dense, trigger an automatic prune
                const sessionMemories = Array.from(this.memories.values()).filter(m => m.type === 'session');
                const HEURISTIC_TOKEN_THRESHOLD = 80000; // ~100 messages * 800 tokens
                
                if (sessionMemories.length > 50) {
                    console.log(`[AgentMemoryService] 🤖 Autonomous Compaction triggered (${sessionMemories.length} session items). Archiving old context...`);
                    // We call handoffSession internally to ensure a snapshot is safe, 
                    // then we could prune. For now, we'll emit a system event so the UI knows.
                    this.memoryManager.pruneContext(
                        Array.from(this.memories.values()).map(m => ({ role: 'user', content: m.content })),
                        { maxTokens: HEURISTIC_TOKEN_THRESHOLD }
                    );
                }
            } catch (e) {
                console.error("[AgentMemoryService] Auto-handoff failed:", e);
            }
        }, 5 * 60 * 1000);
    }

    private scheduleSave(): void {
        if (this.saveTimeout) return;
        this.saveTimeout = setTimeout(() => {
            this.saveToDisk();
            this.saveTimeout = null;
        }, 5000);  // Batch saves every 5 seconds
    }

    /**
     * Force save to disk
     */
    flush(): void {
        if (this.saveTimeout) {
            clearTimeout(this.saveTimeout);
            this.saveTimeout = null;
        }
        this.saveToDisk();
    }

    /**
     * Shutdown and save
     */
    shutdown(): void {
        this.flush();
    }
}

export default AgentMemoryService;
