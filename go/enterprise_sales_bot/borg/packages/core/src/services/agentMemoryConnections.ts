import { getStructuredSessionSummary } from './sessionSummaryMemory.js';
import { getStructuredUserPrompt } from './sessionPromptMemory.js';

const INTENT_STOP_WORDS = new Set([
    'a',
    'an',
    'and',
    'are',
    'as',
    'at',
    'be',
    'been',
    'being',
    'but',
    'by',
    'for',
    'from',
    'had',
    'has',
    'have',
    'in',
    'into',
    'is',
    'it',
    'its',
    'of',
    'on',
    'or',
    'that',
    'the',
    'their',
    'them',
    'these',
    'this',
    'those',
    'to',
    'was',
    'were',
    'with',
]);

type ConnectionMetadata = Record<string, unknown> & {
    source?: unknown;
    sessionId?: unknown;
    structuredObservation?: {
        toolName?: string;
        concepts?: string[];
        filesRead?: string[];
        filesModified?: string[];
    };
};

export type ConnectionMemoryRecord = {
    id: string;
    metadata: ConnectionMetadata;
    createdAt: Date;
};

export type CrossSessionMemoryLink<T extends ConnectionMemoryRecord = ConnectionMemoryRecord> = {
    memory: T;
    score: number;
    reasons: string[];
};

function getMemorySessionId(memory: ConnectionMemoryRecord): string | null {
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

function getMemoryToolName(memory: ConnectionMemoryRecord): string | null {
    const toolName = memory.metadata.structuredObservation?.toolName;
    return typeof toolName === 'string' && toolName.trim() ? toolName : null;
}

function getMemoryConcepts(memory: ConnectionMemoryRecord): string[] {
    const concepts = memory.metadata.structuredObservation?.concepts;
    return Array.isArray(concepts)
        ? concepts.filter((concept): concept is string => typeof concept === 'string' && concept.trim().length > 0)
        : [];
}

function getNormalizedGoalSignals(memory: ConnectionMemoryRecord): string[] {
    const prompt = getStructuredUserPrompt(memory.metadata);
    const summary = getStructuredSessionSummary(memory.metadata);
    const values = [
        prompt?.activeGoal,
        prompt?.lastObjective,
        prompt?.role === 'goal' || prompt?.role === 'objective' ? prompt.content : undefined,
        summary?.activeGoal,
        summary?.lastObjective,
    ];

    const seen = new Set<string>();
    const normalized: string[] = [];

    for (const value of values) {
        if (typeof value !== 'string') {
            continue;
        }

        const candidate = value.trim().toLowerCase();
        if (!candidate || seen.has(candidate)) {
            continue;
        }

        seen.add(candidate);
        normalized.push(candidate);
    }

    return normalized;
}

function getSessionGoalSignals<T extends ConnectionMemoryRecord>(memories: T[], sessionId: string | null): string[] {
    if (!sessionId) {
        return [];
    }

    const seen = new Set<string>();
    const signals: string[] = [];

    for (const memory of memories) {
        if (getMemorySessionId(memory) !== sessionId) {
            continue;
        }

        for (const signal of getNormalizedGoalSignals(memory)) {
            if (seen.has(signal)) {
                continue;
            }

            seen.add(signal);
            signals.push(signal);
        }
    }

    return signals;
}

function getMemoryFiles(memory: ConnectionMemoryRecord): string[] {
    const observation = memory.metadata.structuredObservation;
    if (!observation) {
        return [];
    }

    return [...(observation.filesRead ?? []), ...(observation.filesModified ?? [])]
        .filter((value): value is string => typeof value === 'string' && value.trim().length > 0)
        .map((value) => value.replace(/\\/g, '/'));
}

function getUniqueOverlap(left: string[], right: string[]): string[] {
    const rightSet = new Set(right.map((value) => value.toLowerCase()));
    const seen = new Set<string>();
    const overlap: string[] = [];

    for (const value of left) {
        const normalized = value.toLowerCase();
        if (!rightSet.has(normalized) || seen.has(normalized)) {
            continue;
        }

        seen.add(normalized);
        overlap.push(value);
    }

    return overlap;
}

function normalizeIntentToken(token: string): string {
    const trimmed = token.trim().toLowerCase();
    if (!trimmed) {
        return '';
    }

    if (trimmed.length > 6 && trimmed.endsWith('ing')) {
        return trimmed.slice(0, -3);
    }

    if (trimmed.length > 5 && trimmed.endsWith('ed')) {
        return trimmed.slice(0, -2);
    }

    if (trimmed.length > 5 && trimmed.endsWith('es')) {
        return trimmed.slice(0, -2);
    }

    if (trimmed.length > 4 && trimmed.endsWith('s')) {
        return trimmed.slice(0, -1);
    }

    return trimmed;
}

function getIntentTokens(value: string): string[] {
    const seen = new Set<string>();
    const tokens: string[] = [];

    for (const rawToken of value.split(/[^a-zA-Z0-9]+/)) {
        const normalized = normalizeIntentToken(rawToken);
        if (normalized.length < 4 || INTENT_STOP_WORDS.has(normalized) || seen.has(normalized)) {
            continue;
        }

        seen.add(normalized);
        tokens.push(normalized);
    }

    return tokens;
}

type IntentThemeMatch = {
    left: string;
    right: string;
    sharedTokens: string[];
};

function getIntentThemeMatches(leftSignals: string[], rightSignals: string[]): IntentThemeMatch[] {
    const matches: IntentThemeMatch[] = [];
    const seen = new Set<string>();

    for (const left of leftSignals) {
        const leftTokens = getIntentTokens(left);
        if (leftTokens.length < 2) {
            continue;
        }

        for (const right of rightSignals) {
            if (left === right) {
                continue;
            }

            const rightTokens = getIntentTokens(right);
            if (rightTokens.length < 2) {
                continue;
            }

            const sharedTokens = getUniqueOverlap(leftTokens, rightTokens);
            const overlapRatio = sharedTokens.length / Math.min(leftTokens.length, rightTokens.length);
            const containsOther = left.includes(right) || right.includes(left);

            if (sharedTokens.length < 2 && !containsOther) {
                continue;
            }

            if (!containsOther && overlapRatio < 0.5) {
                continue;
            }

            const key = `${left}::${right}`;
            if (seen.has(key)) {
                continue;
            }

            seen.add(key);
            matches.push({
                left,
                right,
                sharedTokens,
            });
        }
    }

    return matches;
}

export function getCrossSessionMemoryLinks<T extends ConnectionMemoryRecord>(
    memories: T[],
    memoryId: string,
    limit: number = 5,
): Array<CrossSessionMemoryLink<T>> {
    const anchor = memories.find((memory) => memory.id === memoryId);
    if (!anchor) {
        return [];
    }

    const anchorSessionId = getMemorySessionId(anchor);
    const anchorToolName = getMemoryToolName(anchor);
    const anchorSource = typeof anchor.metadata.source === 'string' ? anchor.metadata.source : null;
    const anchorConcepts = getMemoryConcepts(anchor);
    const anchorFiles = getMemoryFiles(anchor);
    const anchorGoalSignals = getSessionGoalSignals(memories, anchorSessionId);

    const related: Array<CrossSessionMemoryLink<T>> = [];

    for (const candidate of memories) {
        if (candidate.id === anchor.id) {
            continue;
        }

        const candidateSessionId = getMemorySessionId(candidate);
        if (!candidateSessionId || (anchorSessionId && candidateSessionId === anchorSessionId)) {
            continue;
        }

        let score = 0;
        const reasons: string[] = [];

        const sharedConcepts = getUniqueOverlap(anchorConcepts, getMemoryConcepts(candidate));
        if (sharedConcepts.length) {
            score += Math.min(sharedConcepts.length, 2) * 3;
            reasons.push(`shared concepts: ${sharedConcepts.slice(0, 2).join(', ')}`);
        }

        const sharedFiles = getUniqueOverlap(anchorFiles, getMemoryFiles(candidate));
        if (sharedFiles.length) {
            score += Math.min(sharedFiles.length, 2) * 3;
            reasons.push(`shared file: ${sharedFiles[0]}`);
        }

        const candidateGoalSignals = getSessionGoalSignals(memories, candidateSessionId);
        const sharedGoalSignals = getUniqueOverlap(anchorGoalSignals, candidateGoalSignals);
        if (sharedGoalSignals.length) {
            score += Math.min(sharedGoalSignals.length, 2) * 4;
            reasons.push(`shared goal/objective: ${sharedGoalSignals[0]}`);
        } else {
            const sharedIntentThemes = getIntentThemeMatches(anchorGoalSignals, candidateGoalSignals);
            if (sharedIntentThemes.length) {
                const topTheme = sharedIntentThemes[0];
                score += Math.min(sharedIntentThemes.length, 2) * 3;
                reasons.push(`similar goal/objective theme: ${topTheme.sharedTokens.slice(0, 2).join(', ')}`);
            }
        }

        const candidateToolName = getMemoryToolName(candidate);
        if (anchorToolName && candidateToolName && anchorToolName === candidateToolName) {
            score += 2;
            reasons.push(`same tool (${anchorToolName})`);
        }

        const candidateSource = typeof candidate.metadata.source === 'string' ? candidate.metadata.source : null;
        if (anchorSource && candidateSource && anchorSource === candidateSource) {
            score += 1;
            reasons.push(`same source (${anchorSource})`);
        }

        if (score > 0) {
            reasons.push(`other session (${candidateSessionId})`);
            related.push({
                memory: candidate,
                score,
                reasons,
            });
        }
    }

    return related
        .sort((left, right) => {
            if (right.score !== left.score) {
                return right.score - left.score;
            }

            return right.memory.createdAt.getTime() - left.memory.createdAt.getTime();
        })
        .slice(0, limit);
}