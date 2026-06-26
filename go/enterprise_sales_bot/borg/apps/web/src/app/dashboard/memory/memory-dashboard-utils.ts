export type MemorySearchMode = 'all' | 'facts' | 'observations' | 'prompts' | 'session_summaries';

export type MemoryTimelineGroup = {
    key: string;
    label: string;
    items: MemoryRecord[];
};

export type MemoryDetailSection = {
    title: string;
    body?: string;
    items?: string[];
};

export type RelatedMemoryRecord = {
    memory: MemoryRecord;
    score: number;
    reasons: string[];
};

export type MemoryWindowGroup = {
    key: 'earlier' | 'later';
    label: string;
    items: MemoryRecord[];
};

export type MemoryPivotAction = {
    key: string;
    label: string;
    query: string;
    mode: MemorySearchMode;
    group: 'session' | 'tool' | 'concept' | 'file' | 'goal' | 'objective';
    description: string;
};

export type MemoryPivotSection = {
    title: string;
    actions: MemoryPivotAction[];
};

export type MemoryRecord = {
    id?: string;
    content: string;
    createdAt?: Date | string | number;
    timestamp?: Date | string | number;
    score?: number;
    metadata?: Record<string, unknown> & {
        source?: string;
        type?: string;
        memoryKind?: string;
        structuredObservation?: {
            type?: string;
            title?: string;
            subtitle?: string;
            narrative?: string;
            facts?: string[];
            concepts?: string[];
            filesRead?: string[];
            filesModified?: string[];
            toolName?: string;
        };
        structuredUserPrompt?: {
            role?: string;
            content?: string;
            promptNumber?: number;
            sessionId?: string;
            activeGoal?: string | null;
            lastObjective?: string | null;
        };
        structuredSessionSummary?: {
            name?: string;
            sessionId?: string;
            status?: string;
            cliType?: string;
            activeGoal?: string | null;
            lastObjective?: string | null;
            restartCount?: number;
        };
    };
};

export type MemoryRecordKind = 'fact' | 'observation' | 'prompt' | 'session_summary';

export const MEMORY_SEARCH_MODES: Array<{ value: MemorySearchMode; label: string; description: string }> = [
    { value: 'all', label: 'All Records', description: 'Facts, observations, prompts, and summaries together.' },
    { value: 'facts', label: 'Facts', description: 'Manual facts and generic memory records in the selected tier.' },
    { value: 'observations', label: 'Observations', description: 'Structured runtime discoveries and tool activity.' },
    { value: 'prompts', label: 'Prompts', description: 'Captured user prompts, goals, and objectives.' },
    { value: 'session_summaries', label: 'Session Summaries', description: 'Structured session state, goals, and restart history.' },
];

export const MEMORY_MODEL_PILLARS = [
    {
        title: 'Facts live in tiers',
        description: 'Session, working, and long-term records remain TormentNexus’s storage backbone for manual and inferred facts.',
    },
    {
        title: 'Observations are structured',
        description: 'Runtime activity is normalized into typed observations with files, tool names, and fact extraction.',
    },
    {
        title: 'Prompts and summaries preserve intent',
        description: 'Captured user prompts and session summaries keep goals, objectives, and provenance visible.',
    },
    {
        title: 'TormentNexus is an adapter',
        description: 'The adapter remains useful for interchange, but TormentNexus-native records are now the source of truth.',
    },
] as const;

export function getMemoryRecordKind(memory: MemoryRecord): MemoryRecordKind {
    if (memory.metadata?.structuredSessionSummary) {
        return 'session_summary';
    }

    if (memory.metadata?.structuredUserPrompt) {
        return 'prompt';
    }

    if (memory.metadata?.structuredObservation) {
        return 'observation';
    }

    return 'fact';
}

export function filterMemoryRecords(memories: MemoryRecord[], mode: MemorySearchMode): MemoryRecord[] {
    if (mode === 'all') {
        return memories;
    }

    if (mode === 'facts') {
        return memories.filter((memory) => getMemoryRecordKind(memory) === 'fact');
    }

    if (mode === 'observations') {
        return memories.filter((memory) => getMemoryRecordKind(memory) === 'observation');
    }

    if (mode === 'prompts') {
        return memories.filter((memory) => getMemoryRecordKind(memory) === 'prompt');
    }

    return memories.filter((memory) => getMemoryRecordKind(memory) === 'session_summary');
}

export function getMemoryBadgeLabel(memory: MemoryRecord): string {
    const summary = memory.metadata?.structuredSessionSummary;
    if (summary) {
        return summary.status?.trim() || 'session summary';
    }

    const prompt = memory.metadata?.structuredUserPrompt;
    if (prompt) {
        return prompt.role?.trim() || 'prompt';
    }

    const observation = memory.metadata?.structuredObservation;
    if (observation) {
        return observation.type?.trim() || 'observation';
    }

    return String(memory.metadata?.memoryKind || memory.metadata?.type || 'fact');
}

export function getMemoryTitle(memory: MemoryRecord): string {
    const summary = memory.metadata?.structuredSessionSummary;
    if (summary) {
        return summary.name?.trim() || summary.sessionId?.trim() || 'Unnamed session';
    }

    const prompt = memory.metadata?.structuredUserPrompt;
    if (prompt) {
        if (typeof prompt.promptNumber === 'number') {
            return `Prompt #${prompt.promptNumber}`;
        }
        return prompt.role?.trim() ? `${prompt.role} capture` : 'Captured prompt';
    }

    const observation = memory.metadata?.structuredObservation;
    if (observation) {
        return observation.title?.trim() || 'Untitled observation';
    }

    const firstLine = memory.content.split(/\r?\n/).find((line) => line.trim().length > 0)?.trim();
    return firstLine?.slice(0, 96) || 'Stored fact';
}

export function getMemoryPreview(memory: MemoryRecord): string {
    const summary = memory.metadata?.structuredSessionSummary;
    if (summary) {
        return summary.activeGoal?.trim() || summary.lastObjective?.trim() || memory.content;
    }

    const prompt = memory.metadata?.structuredUserPrompt;
    if (prompt) {
        return prompt.content?.trim() || prompt.activeGoal?.trim() || prompt.lastObjective?.trim() || memory.content;
    }

    const observation = memory.metadata?.structuredObservation;
    if (observation) {
        return observation.narrative?.trim() || memory.content;
    }

    return memory.content;
}

export function getMemoryTimestamp(memory: MemoryRecord): number {
    const candidate = memory.createdAt ?? memory.timestamp;

    if (candidate instanceof Date) {
        return candidate.getTime();
    }

    if (typeof candidate === 'number') {
        return candidate;
    }

    if (typeof candidate === 'string') {
        const parsed = Date.parse(candidate);
        return Number.isNaN(parsed) ? Date.now() : parsed;
    }

    return Date.now();
}

export function getMemoryProvenance(memory: MemoryRecord): string[] {
    const tokens: string[] = [];
    const summary = memory.metadata?.structuredSessionSummary;
    const prompt = memory.metadata?.structuredUserPrompt;
    const observation = memory.metadata?.structuredObservation;

    if (memory.metadata?.source) {
        tokens.push(`source=${String(memory.metadata.source)}`);
    }

    if (observation?.toolName) {
        tokens.push(`tool=${observation.toolName}`);
    }

    if (summary?.cliType) {
        tokens.push(`cli=${summary.cliType}`);
    }

    const sessionId = summary?.sessionId || prompt?.sessionId;
    if (sessionId) {
        tokens.push(`session=${sessionId}`);
    }

    if (observation?.filesRead?.length) {
        tokens.push(`read=${observation.filesRead.length}`);
    }

    if (observation?.filesModified?.length) {
        tokens.push(`modified=${observation.filesModified.length}`);
    }

    if (observation?.facts?.length) {
        tokens.push(`facts=${observation.facts.length}`);
    }

    if (typeof summary?.restartCount === 'number' && summary.restartCount > 0) {
        tokens.push(`restarts=${summary.restartCount}`);
    }

    return tokens;
}

export function getMemoryRecordKey(memory: MemoryRecord): string {
    if (memory.id?.trim()) {
        return memory.id;
    }

    return [
        getMemoryTimestamp(memory),
        getMemoryTitle(memory),
        getMemoryBadgeLabel(memory),
    ].join('::');
}

export function getMemorySessionId(memory: MemoryRecord): string | null {
    const summary = memory.metadata?.structuredSessionSummary;
    if (summary?.sessionId?.trim()) {
        return summary.sessionId;
    }

    const prompt = memory.metadata?.structuredUserPrompt;
    if (prompt?.sessionId?.trim()) {
        return prompt.sessionId;
    }

    const sessionId = memory.metadata?.sessionId;
    return typeof sessionId === 'string' && sessionId.trim() ? sessionId : null;
}

function getMemoryConcepts(memory: MemoryRecord): string[] {
    const observation = memory.metadata?.structuredObservation;
    return Array.isArray(observation?.concepts)
        ? observation.concepts.filter((value): value is string => typeof value === 'string' && value.trim().length > 0)
        : [];
}

function getMemoryFiles(memory: MemoryRecord): string[] {
    const observation = memory.metadata?.structuredObservation;
    const values = [
        ...(Array.isArray(observation?.filesRead) ? observation.filesRead : []),
        ...(Array.isArray(observation?.filesModified) ? observation.filesModified : []),
    ];

    return values.filter((value): value is string => typeof value === 'string' && value.trim().length > 0);
}

function getMemoryToolName(memory: MemoryRecord): string | null {
    const toolName = memory.metadata?.structuredObservation?.toolName;
    return typeof toolName === 'string' && toolName.trim() ? toolName : null;
}

function getMemoryGoals(memory: MemoryRecord): string[] {
    const summary = memory.metadata?.structuredSessionSummary;
    const prompt = memory.metadata?.structuredUserPrompt;

    return [
        summary?.activeGoal,
        prompt?.activeGoal,
        prompt?.role === 'goal' ? prompt.content : undefined,
    ].filter((value): value is string => typeof value === 'string' && value.trim().length > 0);
}

function getMemoryObjectives(memory: MemoryRecord): string[] {
    const summary = memory.metadata?.structuredSessionSummary;
    const prompt = memory.metadata?.structuredUserPrompt;

    return [
        summary?.lastObjective,
        prompt?.lastObjective,
        prompt?.role === 'objective' ? prompt.content : undefined,
    ].filter((value): value is string => typeof value === 'string' && value.trim().length > 0);
}

function getUniqueStrings(values: string[]): string[] {
    const seen = new Set<string>();
    const results: string[] = [];

    for (const value of values) {
        const normalized = value.trim();
        if (!normalized) {
            continue;
        }

        const key = normalized.toLowerCase();
        if (seen.has(key)) {
            continue;
        }

        seen.add(key);
        results.push(normalized);
    }

    return results;
}

function getOverlap(left: string[], right: string[]): string[] {
    const rightSet = new Set(right.map((value) => value.toLowerCase()));
    return left.filter((value) => rightSet.has(value.toLowerCase()));
}

export function getRelatedMemoryRecords(current: MemoryRecord, memories: MemoryRecord[], limit: number = 4): RelatedMemoryRecord[] {
    const currentKey = getMemoryRecordKey(current);
    const currentSessionId = getMemorySessionId(current);
    const currentToolName = getMemoryToolName(current);
    const currentSource = typeof current.metadata?.source === 'string' ? current.metadata.source : null;
    const currentConcepts = getMemoryConcepts(current);
    const currentFiles = getMemoryFiles(current);

    const related: RelatedMemoryRecord[] = [];

    for (const candidate of memories) {
        if (getMemoryRecordKey(candidate) === currentKey) {
            continue;
        }

        let score = 0;
        const reasons: string[] = [];

        const candidateSessionId = getMemorySessionId(candidate);
        if (currentSessionId && candidateSessionId && currentSessionId === candidateSessionId) {
            score += 5;
            reasons.push(`same session (${currentSessionId})`);
        }

        const candidateToolName = getMemoryToolName(candidate);
        if (currentToolName && candidateToolName && currentToolName === candidateToolName) {
            score += 3;
            reasons.push(`same tool (${currentToolName})`);
        }

        const candidateSource = typeof candidate.metadata?.source === 'string' ? candidate.metadata.source : null;
        if (currentSource && candidateSource && currentSource === candidateSource) {
            score += 2;
            reasons.push(`same source (${currentSource})`);
        }

        const sharedConcepts = getOverlap(currentConcepts, getMemoryConcepts(candidate));
        if (sharedConcepts.length) {
            score += Math.min(sharedConcepts.length, 2) * 2;
            reasons.push(`shared concepts: ${sharedConcepts.slice(0, 2).join(', ')}`);
        }

        const sharedFiles = getOverlap(currentFiles, getMemoryFiles(candidate));
        if (sharedFiles.length) {
            score += Math.min(sharedFiles.length, 2) * 2;
            reasons.push(`shared file: ${sharedFiles[0]}`);
        }

        if (score > 0) {
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

            return getMemoryTimestamp(right.memory) - getMemoryTimestamp(left.memory);
        })
        .slice(0, limit);
}

export function groupMemoryWindowAroundAnchor(anchor: MemoryRecord, memories: MemoryRecord[]): MemoryWindowGroup[] {
    const anchorTimestamp = getMemoryTimestamp(anchor);
    const earlier = memories
        .filter((memory) => getMemoryTimestamp(memory) < anchorTimestamp)
        .sort((left, right) => getMemoryTimestamp(right) - getMemoryTimestamp(left));
    const later = memories
        .filter((memory) => getMemoryTimestamp(memory) >= anchorTimestamp)
        .sort((left, right) => getMemoryTimestamp(left) - getMemoryTimestamp(right));

    const groups: MemoryWindowGroup[] = [];

    if (earlier.length) {
        groups.push({
            key: 'earlier',
            label: 'Earlier in session',
            items: earlier,
        });
    }

    if (later.length) {
        groups.push({
            key: 'later',
            label: 'Later in session',
            items: later,
        });
    }

    return groups;
}

export function getMemoryPivotSections(memory: MemoryRecord): MemoryPivotSection[] {
    const sections: MemoryPivotSection[] = [];
    const sessionId = getMemorySessionId(memory);
    const toolName = getMemoryToolName(memory);
    const goals = getUniqueStrings(getMemoryGoals(memory));
    const objectives = getUniqueStrings(getMemoryObjectives(memory));
    const concepts = getUniqueStrings(getMemoryConcepts(memory));
    const files = getUniqueStrings(getMemoryFiles(memory));

    if (sessionId) {
        sections.push({
            title: 'Session pivots',
            actions: [
                {
                    key: `session:${sessionId}`,
                    label: sessionId,
                    query: sessionId,
                    mode: 'all',
                    group: 'session',
                    description: 'Re-query all records tied to this session identifier.',
                },
            ],
        });
    }

    if (toolName) {
        sections.push({
            title: 'Tool pivots',
            actions: [
                {
                    key: `tool:${toolName}`,
                    label: toolName,
                    query: toolName,
                    mode: 'all',
                    group: 'tool',
                    description: 'Search all related records anchored to observations from this tool.',
                },
            ],
        });
    }

    if (goals.length) {
        sections.push({
            title: 'Goal pivots',
            actions: goals.map((goal) => ({
                key: `goal:${goal}`,
                label: goal,
                query: goal,
                mode: 'all',
                group: 'goal',
                description: 'Search all related records anchored to this active goal.',
            })),
        });
    }

    if (objectives.length) {
        sections.push({
            title: 'Objective pivots',
            actions: objectives.map((objective) => ({
                key: `objective:${objective}`,
                label: objective,
                query: objective,
                mode: 'all',
                group: 'objective',
                description: 'Search all related records anchored to this recent objective.',
            })),
        });
    }

    if (concepts.length) {
        sections.push({
            title: 'Concept pivots',
            actions: concepts.map((concept) => ({
                key: `concept:${concept}`,
                label: concept,
                query: concept,
                mode: 'all',
                group: 'concept',
                description: 'Search all related records anchored to this concept.',
            })),
        });
    }

    if (files.length) {
        sections.push({
            title: 'File pivots',
            actions: files.map((file) => ({
                key: `file:${file}`,
                label: file,
                query: file,
                mode: 'all',
                group: 'file',
                description: 'Search all related records anchored to this file.',
            })),
        });
    }

    return sections;
}

export function sortMemoryRecordsByTimestamp(memories: MemoryRecord[]): MemoryRecord[] {
    return [...memories].sort((left, right) => getMemoryTimestamp(right) - getMemoryTimestamp(left));
}

function formatTimelineDateLabel(timestamp: number, now: number): string {
    const date = new Date(timestamp);
    const startOfDate = new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
    const startOfNow = new Date(now);
    const startOfToday = new Date(startOfNow.getFullYear(), startOfNow.getMonth(), startOfNow.getDate()).getTime();
    const diffDays = Math.round((startOfToday - startOfDate) / 86_400_000);

    if (diffDays === 0) {
        return 'Today';
    }

    if (diffDays === 1) {
        return 'Yesterday';
    }

    return date.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        year: date.getFullYear() === new Date(now).getFullYear() ? undefined : 'numeric',
    });
}

export function groupMemoryRecordsByDay(memories: MemoryRecord[], now: number = Date.now()): MemoryTimelineGroup[] {
    const groups = new Map<string, MemoryTimelineGroup>();

    for (const memory of sortMemoryRecordsByTimestamp(memories)) {
        const timestamp = getMemoryTimestamp(memory);
        const date = new Date(timestamp);
        const key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
        const existing = groups.get(key);

        if (existing) {
            existing.items.push(memory);
            continue;
        }

        groups.set(key, {
            key,
            label: formatTimelineDateLabel(timestamp, now),
            items: [memory],
        });
    }

    return Array.from(groups.values());
}

function pushSection(sections: MemoryDetailSection[], section: MemoryDetailSection): void {
    if (section.body?.trim()) {
        sections.push(section);
        return;
    }

    if (section.items?.length) {
        sections.push(section);
    }
}

export function getMemoryDetailSections(memory: MemoryRecord): MemoryDetailSection[] {
    const sections: MemoryDetailSection[] = [];
    const summary = memory.metadata?.structuredSessionSummary;
    const prompt = memory.metadata?.structuredUserPrompt;
    const observation = memory.metadata?.structuredObservation;

    if (observation) {
        pushSection(sections, {
            title: 'Narrative',
            body: observation.narrative?.trim(),
        });
        pushSection(sections, {
            title: 'Subtitle',
            body: observation.subtitle?.trim(),
        });
        pushSection(sections, {
            title: 'Extracted facts',
            items: observation.facts,
        });
        pushSection(sections, {
            title: 'Concepts',
            items: observation.concepts,
        });
        pushSection(sections, {
            title: 'Files read',
            items: observation.filesRead,
        });
        pushSection(sections, {
            title: 'Files modified',
            items: observation.filesModified,
        });
    } else if (prompt) {
        pushSection(sections, {
            title: 'Prompt content',
            body: prompt.content?.trim() || memory.content,
        });
        pushSection(sections, {
            title: 'Intent anchors',
            items: [prompt.activeGoal ?? '', prompt.lastObjective ?? ''].filter(Boolean),
        });
    } else if (summary) {
        pushSection(sections, {
            title: 'Active goal',
            body: summary.activeGoal?.trim(),
        });
        pushSection(sections, {
            title: 'Last objective',
            body: summary.lastObjective?.trim(),
        });
        pushSection(sections, {
            title: 'Runtime details',
            items: [
                summary.sessionId ? `Session: ${summary.sessionId}` : '',
                summary.cliType ? `CLI: ${summary.cliType}` : '',
                summary.status ? `Status: ${summary.status}` : '',
                typeof summary.restartCount === 'number' ? `Restarts: ${summary.restartCount}` : '',
            ].filter(Boolean),
        });
    } else {
        pushSection(sections, {
            title: 'Stored content',
            body: memory.content,
        });
    }

    if (memory.content.trim()) {
        const preview = getMemoryPreview(memory).trim();
        const content = memory.content.trim();
        if (content !== preview) {
            pushSection(sections, {
                title: 'Canonical record',
                body: content,
            });
        }
    }

    return sections;
}

export function getMemoryModeHint(mode: MemorySearchMode, tier: 'session' | 'working' | 'long_term'): string {
    if (mode === 'all') {
        return `Showing every record TormentNexus can surface for the ${tier} tier.`;
    }

    if (mode === 'facts') {
        return `Showing manual and generic fact records stored in the ${tier} tier.`;
    }

    if (mode === 'observations') {
        return 'Showing structured runtime observations from TormentNexus-native tool and workflow capture.';
    }

    if (mode === 'prompts') {
        return 'Showing structured prompt and goal captures recorded for operator intent.';
    }

    return 'Showing structured session summaries with status, goals, and restart history.';
}
