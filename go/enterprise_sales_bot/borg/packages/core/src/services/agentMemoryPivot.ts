import { getStructuredSessionSummary } from './sessionSummaryMemory.js';
import { getStructuredUserPrompt } from './sessionPromptMemory.js';

export type MemoryPivotKind = 'session' | 'tool' | 'concept' | 'file' | 'goal' | 'objective';

type PivotMetadata = Record<string, unknown> & {
    sessionId?: unknown;
    structuredObservation?: {
        toolName?: string;
        concepts?: string[];
        filesRead?: string[];
        filesModified?: string[];
    };
};

export type PivotMemoryRecord = {
    id: string;
    metadata: PivotMetadata;
    createdAt: Date;
    score?: number;
};

function getMemorySessionId(memory: PivotMemoryRecord): string | null {
    const summary = getStructuredSessionSummary(memory.metadata);
    if (summary?.sessionId?.trim()) {
        return summary.sessionId;
    }

    const prompt = getStructuredUserPrompt(memory.metadata);
    if (prompt?.sessionId?.trim()) {
        return prompt.sessionId;
    }

    const sessionId = memory.metadata.sessionId;
    return typeof sessionId === 'string' && sessionId.trim() ? sessionId : null;
}

function getMemoryToolName(memory: PivotMemoryRecord): string | null {
    const toolName = memory.metadata.structuredObservation?.toolName;
    return typeof toolName === 'string' && toolName.trim() ? toolName : null;
}

function getMemoryConcepts(memory: PivotMemoryRecord): string[] {
    const concepts = memory.metadata.structuredObservation?.concepts;
    return Array.isArray(concepts) ? concepts.filter((concept): concept is string => typeof concept === 'string') : [];
}

function getMemoryFiles(memory: PivotMemoryRecord): string[] {
    const observation = memory.metadata.structuredObservation;
    if (!observation) {
        return [];
    }

    return [...(observation.filesRead ?? []), ...(observation.filesModified ?? [])]
        .filter((value): value is string => typeof value === 'string' && value.trim().length > 0)
        .map((value) => value.replace(/\\/g, '/'));
}

function getMemoryGoals(memory: PivotMemoryRecord): string[] {
    const summary = getStructuredSessionSummary(memory.metadata);
    const prompt = getStructuredUserPrompt(memory.metadata);

    return [
        summary?.activeGoal,
        prompt?.activeGoal,
        prompt?.role === 'goal' ? prompt.content : undefined,
    ].filter((value): value is string => typeof value === 'string' && value.trim().length > 0);
}

function getMemoryObjectives(memory: PivotMemoryRecord): string[] {
    const summary = getStructuredSessionSummary(memory.metadata);
    const prompt = getStructuredUserPrompt(memory.metadata);

    return [
        summary?.lastObjective,
        prompt?.lastObjective,
        prompt?.role === 'objective' ? prompt.content : undefined,
    ].filter((value): value is string => typeof value === 'string' && value.trim().length > 0);
}

function normalizePivotValue(value: string): string {
    return value.trim().replace(/\\/g, '/').toLowerCase();
}

function scoreDirectPivotMatch(memory: PivotMemoryRecord, pivot: MemoryPivotKind, value: string): number {
    if (pivot === 'session') {
        const sessionId = getMemorySessionId(memory);
        return sessionId && sessionId.toLowerCase() === value ? 10 : 0;
    }

    if (pivot === 'tool') {
        const toolName = getMemoryToolName(memory);
        return toolName && toolName.toLowerCase() === value ? 10 : 0;
    }

    if (pivot === 'concept') {
        return getMemoryConcepts(memory).some((concept) => concept.toLowerCase() === value) ? 10 : 0;
    }

    if (pivot === 'goal') {
        return getMemoryGoals(memory).some((goal) => goal.toLowerCase() === value) ? 10 : 0;
    }

    if (pivot === 'objective') {
        return getMemoryObjectives(memory).some((objective) => objective.toLowerCase() === value) ? 10 : 0;
    }

    return getMemoryFiles(memory).some((file) => file.toLowerCase() === value) ? 10 : 0;
}

export function searchMemoryRecordsByPivot<T extends PivotMemoryRecord>(
    memories: T[],
    pivot: MemoryPivotKind,
    value: string,
    limit: number = 20,
): T[] {
    const normalizedValue = normalizePivotValue(value);
    if (!normalizedValue) {
        return [];
    }

    const directMatches = new Map<string, T>();
    const relatedSessionIds = new Set<string>();

    for (const memory of memories) {
        const score = scoreDirectPivotMatch(memory, pivot, normalizedValue);
        if (score <= 0) {
            continue;
        }

        directMatches.set(memory.id, {
            ...memory,
            score,
        });

        const sessionId = getMemorySessionId(memory);
        if (sessionId) {
            relatedSessionIds.add(sessionId.toLowerCase());
        }
    }

    if (pivot !== 'session' && relatedSessionIds.size > 0) {
        for (const memory of memories) {
            if (directMatches.has(memory.id)) {
                continue;
            }

            const sessionId = getMemorySessionId(memory)?.toLowerCase();
            if (!sessionId || !relatedSessionIds.has(sessionId)) {
                continue;
            }

            directMatches.set(memory.id, {
                ...memory,
                score: 4,
            });
        }
    }

    return Array.from(directMatches.values())
        .sort((left, right) => {
            if ((right.score ?? 0) !== (left.score ?? 0)) {
                return (right.score ?? 0) - (left.score ?? 0);
            }

            return right.createdAt.getTime() - left.createdAt.getTime();
        })
        .slice(0, limit);
}