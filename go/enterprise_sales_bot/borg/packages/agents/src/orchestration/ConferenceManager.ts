import { DebateEngine, type Participant as DebateParticipant } from './DebateEngine.js';

export class ConferenceManager {
  private engine: DebateEngine | null = null;
  private maxRounds: number;
  private participants: DebateParticipant[] = [];
  private turns: { participantId: string; content: string }[] = [];

  constructor(maxRounds: number = 3) {
    this.maxRounds = maxRounds;
  }

  /** Set the LLM-backed debate engine. Call before runConferenceIteration. */
  setEngine(engine: DebateEngine): void {
    this.engine = engine;
  }

  registerParticipant(p: DebateParticipant): void {
    this.participants.push(p);
  }

  setupCouncil(): void {
    this.registerParticipant({
      id: 'p1',
      role: 'Planner',
      name: 'Planner',
      model: 'claude-3-7-sonnet',
      systemPrompt: 'You are the Planner. Analyze the objective and propose a step-by-step strategy.',
    });
    this.registerParticipant({
      id: 'p2',
      role: 'Implementer',
      name: 'Implementer',
      model: 'gemini-2.5-flash',
      systemPrompt: "You are the Implementer. Review the Planner's strategy and identify potential technical blockers or suggest concrete implementations.",
    });
    this.registerParticipant({
      id: 'p3',
      role: 'Reviewer',
      name: 'Reviewer',
      model: 'gpt-4o',
      systemPrompt: 'You are the Reviewer. Critically analyze the proposed plans and implementations for security, performance, and correctness flaws.',
    });
  }

  addTurn(participantId: string, content: string): void {
    this.turns.push({ participantId, content });
  }

  getTranscript(): string {
    return this.turns.map((t) => `[${t.participantId}]: ${t.content}`).join('\n');
  }

  async runConferenceIteration(objective: string): Promise<string> {
    this.addTurn('p1', `Initial analysis for objective: ${objective}`);
    if (this.engine && this.participants.length > 0) {
      const result = await this.engine.runDebate(objective, this.participants, this.maxRounds);
      return result.summary;
    }
    return this.getTranscript();
  }
}
