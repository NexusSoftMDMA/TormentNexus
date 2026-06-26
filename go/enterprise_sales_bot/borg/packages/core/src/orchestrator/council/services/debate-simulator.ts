import { EventEmitter } from 'events';

interface StoredVote {
  supervisorId: string;
  vote: 'approve' | 'reject' | 'abstain';
  confidence: number;
  reasoning: string;
  tokensUsed: number;
  responseTimeMs: number;
}

interface StoredRound {
  roundNumber: number;
  votes: StoredVote[];
  timestamp: number;
}

interface StoredDebate {
  id: string;
  topic: string;
  context: string;
  consensusMode: string;
  team: string[];
  rounds: StoredRound[];
  outcome: 'approved' | 'rejected' | 'deadlock';
  finalVotes: Record<string, 'approve' | 'reject' | 'abstain'>;
  createdAt: number;
  durationMs: number;
  metadata?: Record<string, unknown>;
}

interface ReplayConfig {
  consensusMode?: string;
  team?: string[];
  maxRounds?: number;
  speedMultiplier?: number;
}

interface SimulationConfig {
  topic: string;
  context: string;
  consensusMode: string;
  team: string[];
  maxRounds?: number;
  mockResponses?: Record<string, { vote: 'approve' | 'reject' | 'abstain'; confidence: number; reasoning: string }>;
  randomizeVotes?: boolean;
  biasToward?: 'approve' | 'reject';
}

interface ReplayResult {
  originalDebate: StoredDebate;
  replayConfig: ReplayConfig;
  newOutcome: 'approved' | 'rejected' | 'deadlock';
  roundByRoundComparison: Array<{
    roundNumber: number;
    originalVotes: Record<string, string>;
    replayVotes: Record<string, string>;
    votesChanged: boolean;
  }>;
  outcomeChanged: boolean;
  analysis: {
    consensusModeImpact: string;
    teamCompositionImpact: string;
    recommendations: string[];
  };
}

interface SimulationResult {
  id: string;
  config: SimulationConfig;
  rounds: StoredRound[];
  outcome: 'approved' | 'rejected' | 'deadlock';
  finalVotes: Record<string, 'approve' | 'reject' | 'abstain'>;
  durationMs: number;
  isSimulation: true;
}

type ConsensusCalculator = (votes: StoredVote[]) => 'approved' | 'rejected' | 'deadlock' | 'continue';

class DebateSimulatorService extends EventEmitter {
  private storedDebates: Map<string, StoredDebate> = new Map();
  private simulations: Map<string, SimulationResult> = new Map();
  private consensusCalculators: Map<string, ConsensusCalculator> = new Map();

  constructor() {
    super();
    this.initializeConsensusCalculators();
  }

  private initializeConsensusCalculators(): void {
    this.consensusCalculators.set('majority', (votes) => {
      const approves = votes.filter((v) => v.vote === 'approve').length;
      const rejects = votes.filter((v) => v.vote === 'reject').length;
      const total = votes.length;
      if (approves > total / 2) return 'approved';
      if (rejects > total / 2) return 'rejected';
      return 'continue';
    });

    this.consensusCalculators.set('unanimous', (votes) => {
      const allApprove = votes.every((v) => v.vote === 'approve');
      const allReject = votes.every((v) => v.vote === 'reject');
      if (allApprove) return 'approved';
      if (allReject) return 'rejected';
      return 'continue';
    });

    this.consensusCalculators.set('supermajority', (votes) => {
      const approves = votes.filter((v) => v.vote === 'approve').length;
      const rejects = votes.filter((v) => v.vote === 'reject').length;
      const total = votes.length;
      if (approves >= total * 0.67) return 'approved';
      if (rejects >= total * 0.67) return 'rejected';
      return 'continue';
    });

    this.consensusCalculators.set('weighted', (votes) => {
      let approveWeight = 0;
      let rejectWeight = 0;
      for (const vote of votes) {
        if (vote.vote === 'approve') approveWeight += vote.confidence;
        else if (vote.vote === 'reject') rejectWeight += vote.confidence;
      }
      const totalWeight = approveWeight + rejectWeight;
      if (totalWeight === 0) return 'continue';
      if (approveWeight / totalWeight > 0.6) return 'approved';
      if (rejectWeight / totalWeight > 0.6) return 'rejected';
      return 'continue';
    });

    this.consensusCalculators.set('veto', (votes) => {
      const hasVeto = votes.some((v) => v.vote === 'reject' && v.confidence > 0.9);
      if (hasVeto) return 'rejected';
      const approves = votes.filter((v) => v.vote === 'approve').length;
      if (approves > votes.length / 2) return 'approved';
      return 'continue';
    });
  }

  storeDebate(debate: StoredDebate): void {
    this.storedDebates.set(debate.id, debate);
    this.emit('debateStored', { id: debate.id });
  }

  getStoredDebate(id: string): StoredDebate | undefined {
    return this.storedDebates.get(id);
  }

  listStoredDebates(limit?: number): StoredDebate[] {
    const debates = Array.from(this.storedDebates.values())
      .sort((a, b) => b.createdAt - a.createdAt);
    return limit ? debates.slice(0, limit) : debates;
  }

  replayDebate(debateId: string, config: ReplayConfig): ReplayResult {
    const original = this.storedDebates.get(debateId);
    if (!original) {
      throw new Error(`Debate ${debateId} not found`);
    }

    const consensusMode = config.consensusMode || original.consensusMode;
    const team = config.team || original.team;
    const maxRounds = config.maxRounds || original.rounds.length;

    const calculator = this.consensusCalculators.get(consensusMode) || this.consensusCalculators.get('majority')!;

    const roundByRoundComparison: ReplayResult['roundByRoundComparison'] = [];
    let newOutcome: 'approved' | 'rejected' | 'deadlock' = 'deadlock';

    for (let i = 0; i < Math.min(maxRounds, original.rounds.length); i++) {
      const originalRound = original.rounds[i];
      const filteredVotes = originalRound.votes.filter((v) => team.includes(v.supervisorId));

      const originalVotes: Record<string, string> = {};
      const replayVotes: Record<string, string> = {};

      for (const vote of originalRound.votes) {
        originalVotes[vote.supervisorId] = vote.vote;
      }
      for (const vote of filteredVotes) {
        replayVotes[vote.supervisorId] = vote.vote;
      }

      const votesChanged = JSON.stringify(originalVotes) !== JSON.stringify(replayVotes);

      roundByRoundComparison.push({
        roundNumber: i + 1,
        originalVotes,
        replayVotes,
        votesChanged,
      });

      const result = calculator(filteredVotes);
      if (result !== 'continue') {
        newOutcome = result;
        break;
      }
    }

    const outcomeChanged = newOutcome !== original.outcome;

    const analysis = this.analyzeReplay(original, config, newOutcome, outcomeChanged);

    const result: ReplayResult = {
      originalDebate: original,
      replayConfig: config,
      newOutcome,
      roundByRoundComparison,
      outcomeChanged,
      analysis,
    };

    this.emit('replayCompleted', { debateId, outcomeChanged });
    return result;
  }

  simulateDebate(config: SimulationConfig): SimulationResult {
    const id = `sim_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    const startTime = Date.now();
    const rounds: StoredRound[] = [];
    const maxRounds = config.maxRounds || 3;
    const calculator = this.consensusCalculators.get(config.consensusMode) || this.consensusCalculators.get('majority')!;

    let outcome: 'approved' | 'rejected' | 'deadlock' = 'deadlock';
    const finalVotes: Record<string, 'approve' | 'reject' | 'abstain'> = {};

    for (let roundNum = 1; roundNum <= maxRounds; roundNum++) {
      const votes: StoredVote[] = [];

      for (const supervisor of config.team) {
        const vote = this.generateMockVote(supervisor, config, roundNum);
        votes.push(vote);
        finalVotes[supervisor] = vote.vote;
      }

      rounds.push({
        roundNumber: roundNum,
        votes,
        timestamp: Date.now(),
      });

      const result = calculator(votes);
      if (result !== 'continue') {
        outcome = result;
        break;
      }
    }

    const simulation: SimulationResult = {
      id,
      config,
      rounds,
      outcome,
      finalVotes,
      durationMs: Date.now() - startTime,
      isSimulation: true,
    };

    this.simulations.set(id, simulation);
    this.emit('simulationCompleted', { id, outcome });
    return simulation;
  }

  runWhatIfAnalysis(
    debateId: string,
    scenarios: ReplayConfig[]
  ): Array<{ scenario: ReplayConfig; result: ReplayResult }> {
    const results: Array<{ scenario: ReplayConfig; result: ReplayResult }> = [];

    for (const scenario of scenarios) {
      const result = this.replayDebate(debateId, scenario);
      results.push({ scenario, result });
    }

    this.emit('whatIfAnalysisCompleted', { debateId, scenarioCount: scenarios.length });
    return results;
  }

  compareConsensusModes(debateId: string): Record<string, { outcome: string; roundsNeeded: number }> {
    const modes = ['majority', 'unanimous', 'supermajority', 'weighted', 'veto'];
    const results: Record<string, { outcome: string; roundsNeeded: number }> = {};

    for (const mode of modes) {
      const replay = this.replayDebate(debateId, { consensusMode: mode });
      results[mode] = {
        outcome: replay.newOutcome,
        roundsNeeded: replay.roundByRoundComparison.length,
      };
    }

    return results;
  }

  findOptimalTeam(
    debateId: string,
    targetOutcome: 'approved' | 'rejected',
    minTeamSize: number = 3
  ): { team: string[]; outcome: string } | null {
    const original = this.storedDebates.get(debateId);
    if (!original) return null;

    const allSupervisors = original.team;
    const combinations = this.getCombinations(allSupervisors, minTeamSize);

    for (const team of combinations) {
      const result = this.replayDebate(debateId, { team });
      if (result.newOutcome === targetOutcome) {
        return { team, outcome: result.newOutcome };
      }
    }

    return null;
  }

  getSimulation(id: string): SimulationResult | undefined {
    return this.simulations.get(id);
  }

  listSimulations(limit?: number): SimulationResult[] {
    const sims = Array.from(this.simulations.values())
      .sort((a, b) => {
        const aTime = a.rounds[0]?.timestamp || 0;
        const bTime = b.rounds[0]?.timestamp || 0;
        return bTime - aTime;
      });
    return limit ? sims.slice(0, limit) : sims;
  }

  clearStoredDebates(): void {
    this.storedDebates.clear();
    this.emit('debatesCleared', {});
  }

  clearSimulations(): void {
    this.simulations.clear();
    this.emit('simulationsCleared', {});
  }

  getSummary(): {
    storedDebatesCount: number;
    simulationsCount: number;
    availableConsensusModes: string[];
  } {
    return {
      storedDebatesCount: this.storedDebates.size,
      simulationsCount: this.simulations.size,
      availableConsensusModes: Array.from(this.consensusCalculators.keys()),
    };
  }

  private generateMockVote(
    supervisor: string,
    config: SimulationConfig,
    _roundNum: number
  ): StoredVote {
    if (config.mockResponses && config.mockResponses[supervisor]) {
      const mock = config.mockResponses[supervisor];
      return {
        supervisorId: supervisor,
        vote: mock.vote,
        confidence: mock.confidence,
        reasoning: mock.reasoning,
        tokensUsed: Math.floor(Math.random() * 500) + 200,
        responseTimeMs: Math.floor(Math.random() * 2000) + 500,
      };
    }

    let vote: 'approve' | 'reject' | 'abstain';
    if (config.randomizeVotes) {
      const rand = Math.random();
      const bias = config.biasToward === 'approve' ? 0.6 : config.biasToward === 'reject' ? 0.4 : 0.5;
      if (rand < bias) vote = 'approve';
      else if (rand < 0.95) vote = 'reject';
      else vote = 'abstain';
    } else {
      vote = 'approve';
    }

    return {
      supervisorId: supervisor,
      vote,
      confidence: Math.random() * 0.4 + 0.6,
      reasoning: `Simulated vote from ${supervisor}`,
      tokensUsed: Math.floor(Math.random() * 500) + 200,
      responseTimeMs: Math.floor(Math.random() * 2000) + 500,
    };
  }

  private analyzeReplay(
    original: StoredDebate,
    config: ReplayConfig,
    newOutcome: string,
    outcomeChanged: boolean
  ): ReplayResult['analysis'] {
    const recommendations: string[] = [];

    let consensusModeImpact = 'No change in consensus mode';
    if (config.consensusMode && config.consensusMode !== original.consensusMode) {
      consensusModeImpact = `Changed from ${original.consensusMode} to ${config.consensusMode}`;
      if (outcomeChanged) {
        recommendations.push(
          `Consider using ${config.consensusMode} consensus mode for similar topics`
        );
      }
    }

    let teamCompositionImpact = 'No change in team composition';
    if (config.team && config.team.length > 0) {
      const added = config.team.filter((t) => !original.team.includes(t));
      const removed = original.team.filter((t) => !config.team!.includes(t));
      if (added.length > 0 || removed.length > 0) {
        teamCompositionImpact = `Added: ${added.join(', ') || 'none'}, Removed: ${removed.join(', ') || 'none'}`;
        if (outcomeChanged) {
          recommendations.push(
            'Team composition significantly impacts outcomes. Review supervisor selection criteria.'
          );
        }
      }
    }

    if (!outcomeChanged) {
      recommendations.push('The outcome is robust to the tested configuration changes.');
    }

    return {
      consensusModeImpact,
      teamCompositionImpact,
      recommendations,
    };
  }

  private getCombinations<T>(arr: T[], minSize: number): T[][] {
    const results: T[][] = [];

    const combine = (start: number, combo: T[]) => {
      if (combo.length >= minSize) {
        results.push([...combo]);
      }
      for (let i = start; i < arr.length; i++) {
        combo.push(arr[i]);
        combine(i + 1, combo);
        combo.pop();
      }
    };

    combine(0, []);
    return results;
  }
}

export const debateSimulator = new DebateSimulatorService();
