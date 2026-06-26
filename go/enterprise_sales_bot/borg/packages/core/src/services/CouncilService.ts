
import { v4 as uuidv4 } from 'uuid';

export interface Opinion {
    agentId: string;
    content: string;
    timestamp: number;
    round: number;
}

export interface Vote {
    agentId: string;
    choice: string;
    reason: string;
    timestamp: number;
}

export interface CouncilSession {
    id: string;
    topic: string;
    status: 'active' | 'concluded';
    round: number;
    opinions: Opinion[];
    votes: Vote[];
    createdAt: number;
}

export type CouncilAgent = unknown;

export class CouncilService {
    private sessions: Map<string, CouncilSession> = new Map();
    private agents: Map<string, CouncilAgent> = new Map();

    constructor() { }

    public registerAgent(role: string, agent: CouncilAgent): void {
        this.agents.set(role, agent);
    }

    public startSession(topic: string): CouncilSession {
        const id = uuidv4();
        const session: CouncilSession = {
            id,
            topic,
            status: 'active',
            round: 1,
            opinions: [],
            votes: [],
            createdAt: Date.now()
        };
        this.sessions.set(id, session);
        return session;
    }

    public getSession(id: string): CouncilSession | undefined {
        return this.sessions.get(id);
    }

    public listSessions(): CouncilSession[] {
        return Array.from(this.sessions.values()).sort((a, b) => b.createdAt - a.createdAt);
    }

    public submitOpinion(sessionId: string, agentId: string, content: string): void {
        const session = this.sessions.get(sessionId);
        if (!session) throw new Error(`Council Session ${sessionId} not found`);
        if (session.status !== 'active') throw new Error(`Council Session ${sessionId} is concluded`);

        session.opinions.push({
            agentId,
            content,
            timestamp: Date.now(),
            round: session.round
        });
    }

    public advanceRound(sessionId: string): void {
        const session = this.sessions.get(sessionId);
        if (!session) throw new Error(`Council Session ${sessionId} not found`);
        session.round++;
    }

    public castVote(sessionId: string, agentId: string, choice: string, reason: string): void {
        const session = this.sessions.get(sessionId);
        if (!session) throw new Error(`Council Session ${sessionId} not found`);
        if (session.status !== 'active') throw new Error(`Council Session ${sessionId} is concluded`);

        session.votes.push({
            agentId,
            choice,
            reason,
            timestamp: Date.now()
        });
    }

    public concludeSession(sessionId: string): void {
        const session = this.sessions.get(sessionId);
        if (!session) throw new Error(`Council Session ${sessionId} not found`);
        session.status = 'concluded';
    }
}
