export interface ToolSearchCandidate {
    name: string;
    description?: string | null;
    serverName?: string | null;
    serverDisplayName?: string | null;
    originalName?: string | null;
    advertisedName?: string | null;
    serverTags?: string[];
    toolTags?: string[];
    semanticGroup?: string | null;
    semanticGroupLabel?: string | null;
    keywords?: string[];
    alwaysOn?: boolean;
    loaded?: boolean;
    hydrated?: boolean;
    deferred?: boolean;
}

export type ToolSearchProfile = 'web-research' | 'repo-coding' | 'browser-automation' | 'local-ops' | 'database';

export interface RankedToolSearchResult {
    name: string;
    description: string;
    serverName?: string;
    serverDisplayName?: string;
    originalName?: string;
    advertisedName?: string;
    serverTags?: string[];
    toolTags?: string[];
    semanticGroup?: string;
    semanticGroupLabel?: string;
    keywords?: string[];
    alwaysOn?: boolean;
    loaded: boolean;
    hydrated: boolean;
    deferred: boolean;
    requiresSchemaHydration: boolean;
    matchReason: string;
    score: number;
    autoLoaded?: boolean;
}

export interface ToolSearchAutoLoadDecision {
    toolName: string;
    reason: string;
    confidence: number;
    scoreGap: number;
    topScore: number;
    secondScore: number;
}

export type ToolSearchAutoLoadOutcome = 'loaded' | 'skipped' | 'not-applicable';

export interface ToolSearchAutoLoadEvaluation {
    evaluated: boolean;
    outcome: ToolSearchAutoLoadOutcome;
    decision: ToolSearchAutoLoadDecision | null;
    skipReason?: string;
    minConfidence?: number;
}

interface ToolSearchAutoLoadOptions {
    minConfidence?: number;
}

function normalizeText(value: string | null | undefined): string {
    return value?.trim().toLowerCase() ?? '';
}

function tokenizeQuery(query: string): string[] {
    return normalizeText(query)
        .split(/\s+/)
        .filter(Boolean);
}

function buildNoQueryScore(candidate: ToolSearchCandidate): number {
    let score = 0;

    if (candidate.alwaysOn) {
        score += 30;
    }

    if (candidate.loaded) {
        score += 20;
    }

    if (candidate.hydrated) {
        score += 10;
    }

    if (!candidate.deferred) {
        score += 2;
    }

    return score;
}

function buildProfileBoost(
    candidate: ToolSearchCandidate,
    profile: ToolSearchProfile | undefined,
): { boost: number; reason?: string } {
    if (!profile) {
        return { boost: 0 };
    }

    const searchable = normalizeText([
        candidate.name,
        candidate.originalName,
        candidate.advertisedName,
        candidate.description,
        candidate.serverName,
        candidate.serverDisplayName,
        candidate.semanticGroup,
        candidate.semanticGroupLabel,
        ...(candidate.serverTags ?? []),
        ...(candidate.toolTags ?? []),
        ...(candidate.keywords ?? []),
    ].join(' '));

    const profileKeywords: Record<ToolSearchProfile, string[]> = {
        'web-research': ['web', 'search', 'crawl', 'scrape', 'http', 'url', 'browser', 'docs'],
        'repo-coding': ['code', 'repo', 'git', 'lsp', 'symbol', 'test', 'build', 'diff'],
        'browser-automation': ['browser', 'playwright', 'dom', 'click', 'navigate', 'screenshot', 'page'],
        'local-ops': ['shell', 'terminal', 'process', 'filesystem', 'task', 'command', 'runtime'],
        'database': ['db', 'database', 'sql', 'sqlite', 'postgres', 'query', 'schema'],
    };

    const matches = profileKeywords[profile].filter((keyword) => searchable.includes(keyword));
    if (matches.length === 0) {
        return { boost: 0 };
    }

    const boost = Math.min(18, matches.length * 4);
    return {
        boost,
        reason: `${profile} profile boost (${matches.slice(0, 3).join(', ')})`,
    };
}

function scoreCandidate(
    candidate: ToolSearchCandidate,
    normalizedQuery: string,
    queryTokens: string[],
    profile: ToolSearchProfile | undefined,
): { score: number; matchReason: string } | null {
    if (!normalizedQuery) {
        const profileBoost = buildProfileBoost(candidate, profile);
        return {
            score: buildNoQueryScore(candidate) + profileBoost.boost,
            matchReason: profileBoost.reason
                ?? (candidate.loaded
                    ? 'already loaded in the current session'
                    : 'available tool in the current catalog'),
        };
    }

    const normalizedName = normalizeText(candidate.name);
    const normalizedOriginalName = normalizeText(candidate.originalName);
    const normalizedAdvertisedName = normalizeText(candidate.advertisedName);
    const normalizedDescription = normalizeText(candidate.description);
    const normalizedServerName = normalizeText(candidate.serverName);
    const normalizedServerDisplayName = normalizeText(candidate.serverDisplayName);
    const normalizedSemanticGroup = normalizeText(candidate.semanticGroup);
    const normalizedSemanticGroupLabel = normalizeText(candidate.semanticGroupLabel);
    const normalizedTagText = normalizeText([
        ...(candidate.serverTags ?? []),
        ...(candidate.toolTags ?? []),
        ...(candidate.keywords ?? []),
    ].join(' '));

    let score = 0;
    let matchReason = '';

    if (normalizedName === normalizedQuery) {
        score += 120;
        matchReason = 'exact tool name match';
    } else if (normalizedOriginalName === normalizedQuery) {
        score += 115;
        matchReason = 'exact original tool name match';
    } else if (normalizedAdvertisedName === normalizedQuery) {
        score += 110;
        matchReason = 'exact advertised tool name match';
    } else if (normalizedName.startsWith(normalizedQuery)) {
        score += 90;
        matchReason = 'tool name prefix match';
    } else if (normalizedOriginalName.startsWith(normalizedQuery)) {
        score += 85;
        matchReason = 'original tool name prefix match';
    } else if (normalizedAdvertisedName.startsWith(normalizedQuery)) {
        score += 80;
        matchReason = 'advertised tool name prefix match';
    } else if (normalizedName.includes(normalizedQuery)) {
        score += 70;
        matchReason = 'tool name contains query';
    } else if (normalizedOriginalName.includes(normalizedQuery)) {
        score += 65;
        matchReason = 'original tool name contains query';
    } else if (normalizedAdvertisedName.includes(normalizedQuery)) {
        score += 60;
        matchReason = 'advertised tool name contains query';
    } else if (normalizedTagText.includes(normalizedQuery)) {
        score += 58;
        matchReason = 'semantic tag match';
    } else if (normalizedSemanticGroupLabel.includes(normalizedQuery) || normalizedSemanticGroup.includes(normalizedQuery)) {
        score += 55;
        matchReason = 'semantic group match';
    } else if (normalizedDescription.includes(normalizedQuery)) {
        score += 45;
        matchReason = 'description contains query';
    } else if (normalizedServerDisplayName.includes(normalizedQuery)) {
        score += 35;
        matchReason = 'advertised server name contains query';
    } else if (normalizedServerName.includes(normalizedQuery)) {
        score += 30;
        matchReason = 'server name contains query';
    }

    const tokenMatches = queryTokens.filter((token) => (
        normalizedName.includes(token)
        || normalizedOriginalName.includes(token)
        || normalizedAdvertisedName.includes(token)
        || normalizedDescription.includes(token)
        || normalizedServerDisplayName.includes(token)
        || normalizedServerName.includes(token)
        || normalizedSemanticGroup.includes(token)
        || normalizedSemanticGroupLabel.includes(token)
        || normalizedTagText.includes(token)
    ));

    if (tokenMatches.length === 0 && score === 0) {
        return null;
    }

    score += tokenMatches.length * 6;

    if (!matchReason) {
        matchReason = tokenMatches.length > 1
            ? `matched ${tokenMatches.length} query keywords`
            : 'matched a query keyword';
    }

    if (candidate.loaded) {
        score += 5;
    }

    if (candidate.alwaysOn) {
        score += 4;
    }

    if (candidate.hydrated) {
        score += 3;
    }

    const profileBoost = buildProfileBoost(candidate, profile);
    score += profileBoost.boost;

    if (profileBoost.reason && profile && !matchReason.includes(profile)) {
        matchReason = `${matchReason}; ${profileBoost.reason}`;
    }

    return { score, matchReason };
}

function compareResults(left: RankedToolSearchResult, right: RankedToolSearchResult): number {
    if (right.score !== left.score) {
        return right.score - left.score;
    }

    if (left.loaded !== right.loaded) {
        return left.loaded ? -1 : 1;
    }

    if (left.hydrated !== right.hydrated) {
        return left.hydrated ? -1 : 1;
    }

    return left.name.localeCompare(right.name);
}

export function rankToolSearchCandidates(
    candidates: ToolSearchCandidate[],
    query: string,
    limit: number,
    profile?: ToolSearchProfile,
): RankedToolSearchResult[] {
    const normalizedQuery = normalizeText(query);
    const queryTokens = tokenizeQuery(query);
    const safeLimit = Math.max(1, limit);
    const rankedCandidates: Array<RankedToolSearchResult | null> = candidates
        .map((candidate) => {
            const ranking = scoreCandidate(candidate, normalizedQuery, queryTokens, profile);
            if (!ranking) {
                return null;
            }

            const loaded = candidate.loaded ?? false;
            const hydrated = candidate.hydrated ?? false;
            const deferred = candidate.deferred ?? false;

            return {
                name: candidate.name,
                description: candidate.description ?? '',
                serverName: candidate.serverName ?? undefined,
                serverDisplayName: candidate.serverDisplayName ?? undefined,
                originalName: candidate.originalName ?? undefined,
                advertisedName: candidate.advertisedName ?? undefined,
                serverTags: candidate.serverTags ?? [],
                toolTags: candidate.toolTags ?? [],
                semanticGroup: candidate.semanticGroup ?? undefined,
                semanticGroupLabel: candidate.semanticGroupLabel ?? undefined,
                keywords: candidate.keywords ?? [],
                alwaysOn: candidate.alwaysOn ?? false,
                loaded,
                hydrated,
                deferred,
                requiresSchemaHydration: deferred && !hydrated,
                matchReason: ranking.matchReason,
                score: ranking.score,
            };
        });

    return rankedCandidates
        .filter((candidate): candidate is RankedToolSearchResult => candidate !== null)
        .sort(compareResults)
        .slice(0, safeLimit);
}

export function pickAutoLoadCandidate(
    results: RankedToolSearchResult[],
    query: string,
    options?: ToolSearchAutoLoadOptions,
): ToolSearchAutoLoadDecision | null {
    return evaluateAutoLoadCandidate(results, query, options).decision;
}

export function evaluateAutoLoadCandidate(
    results: RankedToolSearchResult[],
    query: string,
    options?: ToolSearchAutoLoadOptions,
): ToolSearchAutoLoadEvaluation {
    const normalizedQuery = normalizeText(query);
    if (!normalizedQuery || results.length === 0) {
        return {
            evaluated: false,
            outcome: 'not-applicable',
            decision: null,
            skipReason: 'no query or ranked results available',
        };
    }

    const [topResult, secondResult] = results;
    if (!topResult || topResult.loaded) {
        return {
            evaluated: false,
            outcome: 'not-applicable',
            decision: null,
            skipReason: topResult?.loaded
                ? 'top result already loaded'
                : 'no top-ranked result available',
        };
    }

    const scoreGap = topResult.score - (secondResult?.score ?? 0);
    const hasExactMatch = topResult.matchReason.includes('exact');
    const hasPrefixMatch = topResult.matchReason.includes('prefix');
    const hasStrongKeywordMatch = topResult.score >= 125 && scoreGap >= 18;

    if (!(hasExactMatch || hasPrefixMatch || hasStrongKeywordMatch)) {
        return {
            evaluated: true,
            outcome: 'skipped',
            decision: null,
            skipReason: 'top result did not meet exact/prefix/strong-keyword auto-load criteria',
        };
    }

    if (hasPrefixMatch && topResult.score < 90) {
        return {
            evaluated: true,
            outcome: 'skipped',
            decision: null,
            skipReason: 'prefix match score below minimum threshold',
        };
    }

    if (!hasExactMatch && scoreGap < 10) {
        return {
            evaluated: true,
            outcome: 'skipped',
            decision: null,
            skipReason: 'top result too ambiguous relative to second result',
        };
    }

    const baseConfidence = hasExactMatch
        ? 0.95
        : hasPrefixMatch
            ? 0.87
            : 0.82;

    const confidence = Math.max(0, Math.min(0.99, baseConfidence + (scoreGap >= 20 ? 0.04 : scoreGap >= 12 ? 0.02 : 0)));
    const minConfidence = Math.max(0, Math.min(0.99, options?.minConfidence ?? 0.85));

    if (confidence < minConfidence) {
        return {
            evaluated: true,
            outcome: 'skipped',
            decision: null,
            skipReason: 'confidence below configured auto-load threshold',
            minConfidence,
        };
    }

    return {
        evaluated: true,
        outcome: 'loaded',
        decision: {
            toolName: topResult.name,
            reason: `auto-loaded after ${topResult.matchReason}`,
            confidence,
            scoreGap,
            topScore: topResult.score,
            secondScore: secondResult?.score ?? 0,
        },
        minConfidence,
    };
}