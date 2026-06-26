export interface CouncilOpinionRow {
    agentId: string;
    content: string;
    timestamp: number;
    round: number;
}

export interface CouncilVoteRow {
    agentId: string;
    choice: string;
    reason: string;
    timestamp: number;
}

export interface CouncilSessionRow {
    id: string;
    topic: string;
    status: 'active' | 'concluded';
    round: number;
    opinions: CouncilOpinionRow[];
    votes: CouncilVoteRow[];
    createdAt: number;
}

function isObject(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null;
}

function asFiniteNumber(value: unknown, fallback = 0): number {
    return typeof value === 'number' && Number.isFinite(value) ? value : fallback;
}

function normalizeStatus(value: unknown): 'active' | 'concluded' {
    return value === 'concluded' ? 'concluded' : 'active';
}

function normalizeOpinions(payload: unknown): CouncilOpinionRow[] {
    if (!Array.isArray(payload)) {
        return [];
    }

    return payload.reduce<CouncilOpinionRow[]>((acc, item, index) => {
        if (!isObject(item)) {
            return acc;
        }

        const rawAgentId = typeof item.agentId === 'string' ? item.agentId.trim() : '';

        acc.push({
            agentId: rawAgentId.length > 0 ? rawAgentId : `agent-${index}`,
            content: typeof item.content === 'string' ? item.content : '',
            timestamp: asFiniteNumber(item.timestamp),
            round: asFiniteNumber(item.round),
        });

        return acc;
    }, []);
}

function normalizeVotes(payload: unknown): CouncilVoteRow[] {
    if (!Array.isArray(payload)) {
        return [];
    }

    return payload.reduce<CouncilVoteRow[]>((acc, item, index) => {
        if (!isObject(item)) {
            return acc;
        }

        const rawAgentId = typeof item.agentId === 'string' ? item.agentId.trim() : '';

        acc.push({
            agentId: rawAgentId.length > 0 ? rawAgentId : `voter-${index}`,
            choice: typeof item.choice === 'string' ? item.choice : '',
            reason: typeof item.reason === 'string' ? item.reason : '',
            timestamp: asFiniteNumber(item.timestamp),
        });

        return acc;
    }, []);
}

export function normalizeCouncilSessions(payload: unknown): CouncilSessionRow[] {
    if (!Array.isArray(payload)) {
        return [];
    }

    return payload.reduce<CouncilSessionRow[]>((acc, item, index) => {
        if (!isObject(item)) {
            return acc;
        }

        const rawId = typeof item.id === 'string' ? item.id.trim() : '';
        const rawTopic = typeof item.topic === 'string' ? item.topic.trim() : '';

        acc.push({
            id: rawId.length > 0 ? rawId : `session-${index}`,
            topic: rawTopic.length > 0 ? rawTopic : 'Untitled debate',
            status: normalizeStatus(item.status),
            round: asFiniteNumber(item.round),
            opinions: normalizeOpinions(item.opinions),
            votes: normalizeVotes(item.votes),
            createdAt: asFiniteNumber(item.createdAt),
        });

        return acc;
    }, []);
}
