import { getStructuredSessionSummary } from './sessionSummaryMemory.js';
import { getStructuredUserPrompt } from './sessionPromptMemory.js';

type TimelineMetadata = Record<string, unknown> & {
    sessionId?: unknown;
};

export type TimelineMemoryRecord = {
    id: string;
    metadata: TimelineMetadata;
    createdAt: Date;
    score?: number;
};

function getMemorySessionId(memory: TimelineMemoryRecord): string | null {
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

export function getMemoryTimelineWindow<T extends TimelineMemoryRecord>(
    memories: T[],
    sessionId: string,
    anchorTimestamp: number,
    before: number = 3,
    after: number = 3,
): T[] {
    const normalizedSessionId = sessionId.trim().toLowerCase();
    if (!normalizedSessionId || !Number.isFinite(anchorTimestamp)) {
        return [];
    }

    const sessionMemories = memories
        .filter((memory) => getMemorySessionId(memory)?.toLowerCase() === normalizedSessionId)
        .sort((left, right) => left.createdAt.getTime() - right.createdAt.getTime());

    if (sessionMemories.length === 0) {
        return [];
    }

    let anchorIndex = sessionMemories.findIndex((memory) => memory.createdAt.getTime() >= anchorTimestamp);
    if (anchorIndex === -1) {
        anchorIndex = sessionMemories.length - 1;
    } else if (anchorIndex > 0) {
        const candidate = sessionMemories[anchorIndex];
        const previous = sessionMemories[anchorIndex - 1];
        if (candidate && previous) {
            const candidateDistance = Math.abs(candidate.createdAt.getTime() - anchorTimestamp);
            const previousDistance = Math.abs(previous.createdAt.getTime() - anchorTimestamp);
            if (previousDistance <= candidateDistance) {
                anchorIndex -= 1;
            }
        }
    }

    const start = Math.max(0, anchorIndex - before);
    const end = Math.min(sessionMemories.length, anchorIndex + after + 1);
    return sessionMemories.slice(start, end);
}