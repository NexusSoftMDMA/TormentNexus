export interface ToolContextObservationCandidate {
    title?: string;
    narrative?: string;
    content?: string;
    type?: string;
    toolName?: string;
    concepts?: string[];
    filesRead?: string[];
    filesModified?: string[];
    recordedAt?: number;
}

export interface ToolContextSummaryCandidate {
    content?: string;
    cliType?: string;
    status?: string;
    sessionId?: string;
    recordedAt?: number;
}

export interface ToolContextPayload {
    toolName: string;
    query: string;
    matchedPaths: string[];
    observationCount: number;
    summaryCount: number;
    prompt: string;
}

function sanitizeLine(value: string, maxLength: number): string {
    return value.replace(/\s+/g, ' ').trim().slice(0, maxLength);
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

function basenameLike(filePath: string): string {
    const normalized = filePath.replace(/\\/g, '/').trim();
    const parts = normalized.split('/').filter(Boolean);
    return parts.at(-1) ?? normalized;
}

function normalizePathCandidate(value: unknown): string[] {
    if (typeof value === 'string') {
        const trimmed = value.trim();
        return trimmed ? [trimmed.replace(/\\/g, '/')] : [];
    }

    if (Array.isArray(value)) {
        return value.flatMap((entry) => normalizePathCandidate(entry));
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
        ...normalizePathCandidate(record.absolutePath),
        ...normalizePathCandidate(record.targetPath),
        ...normalizePathCandidate(record.oldPath),
        ...normalizePathCandidate(record.newPath),
        ...normalizePathCandidate(record.old_path),
        ...normalizePathCandidate(record.new_path),
        ...normalizePathCandidate(record.dirPath),
        ...normalizePathCandidate(record.paths),
        ...normalizePathCandidate(record.files),
    ], 10);
}

function extractTextCandidates(args: unknown): string[] {
    if (!args || typeof args !== 'object') {
        return [];
    }

    const record = args as Record<string, unknown>;
    return uniqueStrings([
        typeof record.query === 'string' ? record.query : null,
        typeof record.prompt === 'string' ? record.prompt : null,
        typeof record.description === 'string' ? record.description : null,
        typeof record.instructions === 'string' ? record.instructions : null,
        typeof record.goal === 'string' ? record.goal : null,
        typeof record.name === 'string' ? record.name : null,
    ], 6).map((value) => sanitizeLine(value, 120));
}

function tokenize(value: string): string[] {
    return value
        .toLowerCase()
        .split(/[^a-z0-9]+/)
        .map((part) => part.trim())
        .filter((part) => part.length >= 3 && !['tool', 'with', 'from', 'that', 'this', 'into', 'using', 'args'].includes(part));
}

function buildQuery(toolName: string, args: unknown): { query: string; tokens: string[]; paths: string[] } {
    const paths = extractPathCandidates(args);
    const pathTokens = paths.flatMap((filePath) => tokenize(basenameLike(filePath)));
    const argTokens = extractTextCandidates(args).flatMap((value) => tokenize(value));
    const toolTokens = tokenize(toolName);
    const tokens = uniqueStrings([...toolTokens, ...pathTokens, ...argTokens], 16);
    const query = uniqueStrings([
        toolName,
        ...paths.map((filePath) => basenameLike(filePath)),
        ...extractTextCandidates(args),
    ], 8).join(' ');

    return {
        query: sanitizeLine(query || toolName, 240),
        tokens,
        paths,
    };
}

function scoreText(text: string, tokens: string[]): number {
    const lowered = text.toLowerCase();
    let score = 0;

    for (const token of tokens) {
        if (lowered.includes(token)) {
            score += 2;
        }
    }

    return score;
}

function scoreObservation(
    observation: ToolContextObservationCandidate,
    toolName: string,
    tokens: string[],
    paths: string[],
): number {
    let score = 0;

    if (observation.toolName === toolName) {
        score += 8;
    } else if (typeof observation.toolName === 'string' && observation.toolName.includes(toolName)) {
        score += 4;
    }

    score += scoreText([observation.title, observation.narrative, observation.content, ...(observation.concepts ?? [])].filter(Boolean).join(' '), tokens);

    const fileTouches = [...(observation.filesRead ?? []), ...(observation.filesModified ?? [])];
    for (const path of paths) {
        const normalizedPath = path.toLowerCase();
        const basename = basenameLike(path).toLowerCase();
        if (fileTouches.some((candidate) => candidate.toLowerCase().includes(normalizedPath))) {
            score += 6;
            continue;
        }
        if (fileTouches.some((candidate) => basename && candidate.toLowerCase().includes(basename))) {
            score += 4;
        }
    }

    return score;
}

function scoreSummary(summary: ToolContextSummaryCandidate, tokens: string[], paths: string[]): number {
    const combined = [summary.content, summary.cliType, summary.status].filter(Boolean).join(' ');
    let score = scoreText(combined, tokens);

    for (const path of paths) {
        const basename = basenameLike(path).toLowerCase();
        if (basename && combined.toLowerCase().includes(basename)) {
            score += 3;
        }
    }

    return score;
}

function formatObservationLine(observation: ToolContextObservationCandidate): string {
    const title = sanitizeLine(observation.title ?? observation.narrative ?? observation.content ?? 'Relevant observation', 160);
    const detail = sanitizeLine(observation.narrative ?? observation.content ?? '', 220);
    const tags = uniqueStrings([observation.type, observation.toolName], 3).join(' · ');
    const suffix = detail && detail !== title ? ` — ${detail}` : '';
    return tags ? `${title}${suffix} (${tags})` : `${title}${suffix}`;
}

function formatSummaryLine(summary: ToolContextSummaryCandidate): string {
    return sanitizeLine(summary.content ?? 'Relevant prior session summary', 240);
}

export function buildToolContextPayload(input: {
    toolName: string;
    args?: unknown;
    activeGoal?: string | null;
    lastObjective?: string | null;
    observations: ToolContextObservationCandidate[];
    summaries: ToolContextSummaryCandidate[];
}): ToolContextPayload {
    const toolName = sanitizeLine(input.toolName, 120);
    const { query, tokens, paths } = buildQuery(toolName, input.args);

    const matchedObservations = input.observations
        .map((observation) => ({
            observation,
            score: scoreObservation(observation, toolName, tokens, paths),
            recordedAt: observation.recordedAt ?? 0,
        }))
        .filter((entry) => entry.score > 0)
        .sort((left, right) => right.score - left.score || right.recordedAt - left.recordedAt)
        .slice(0, 4)
        .map((entry) => entry.observation);

    const matchedSummaries = input.summaries
        .map((summary) => ({
            summary,
            score: scoreSummary(summary, tokens, paths),
            recordedAt: summary.recordedAt ?? 0,
        }))
        .filter((entry) => entry.score > 0)
        .sort((left, right) => right.score - left.score || right.recordedAt - left.recordedAt)
        .slice(0, 2)
        .map((entry) => entry.summary);

    const fallbackSummaries = matchedSummaries.length > 0
        ? matchedSummaries
        : input.summaries
            .slice()
            .sort((left, right) => (right.recordedAt ?? 0) - (left.recordedAt ?? 0))
            .slice(0, 1);

    const matchedPaths = uniqueStrings([
        ...paths,
        ...matchedObservations.flatMap((observation) => [...(observation.filesRead ?? []), ...(observation.filesModified ?? [])]),
    ], 8);

    const sections = [
        `JIT tool context for ${toolName}:`,
        input.activeGoal ? `Current goal: ${sanitizeLine(input.activeGoal, 180)}` : null,
        input.lastObjective ? `Last objective: ${sanitizeLine(input.lastObjective, 180)}` : null,
        query ? `Focus query: ${query}` : null,
        matchedPaths.length > 0 ? `Relevant files: ${matchedPaths.map((filePath) => basenameLike(filePath)).join(', ')}` : null,
        matchedObservations.length > 0 ? 'Potentially relevant observations:' : null,
        ...matchedObservations.map((observation) => `- ${formatObservationLine(observation)}`),
        fallbackSummaries.length > 0 ? 'Potentially relevant session summaries:' : null,
        ...fallbackSummaries.map((summary) => `- ${formatSummaryLine(summary)}`),
        matchedObservations.length === 0 && fallbackSummaries.length === 0
            ? 'No strongly relevant prior memory was found for this tool call.'
            : 'Use only the parts that still match the current intent; ignore stale context.',
    ].filter((line): line is string => Boolean(line));

    return {
        toolName,
        query,
        matchedPaths,
        observationCount: matchedObservations.length,
        summaryCount: fallbackSummaries.length,
        prompt: sections.join('\n'),
    };
}