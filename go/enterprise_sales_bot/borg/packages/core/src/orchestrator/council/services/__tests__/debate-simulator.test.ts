import { describe, it, expect, beforeEach } from 'vitest';
import { debateSimulator } from '../debate-simulator.js';

const createTestDebate = (id: string, outcome: 'approved' | 'rejected' | 'deadlock' = 'approved') => ({
  id,
  topic: 'Test topic',
  context: 'Test context',
  consensusMode: 'majority',
  team: ['gpt-4', 'claude', 'gemini'],
  rounds: [
    {
      roundNumber: 1,
      votes: [
        { supervisorId: 'gpt-4', vote: 'approve' as const, confidence: 0.9, reasoning: 'Good', tokensUsed: 300, responseTimeMs: 1500 },
        { supervisorId: 'claude', vote: 'approve' as const, confidence: 0.8, reasoning: 'OK', tokensUsed: 250, responseTimeMs: 1200 },
        { supervisorId: 'gemini', vote: 'reject' as const, confidence: 0.6, reasoning: 'Issues', tokensUsed: 200, responseTimeMs: 1000 },
      ],
      timestamp: Date.now(),
    },
  ],
  outcome,
  finalVotes: { 'gpt-4': 'approve' as const, 'claude': 'approve' as const, 'gemini': 'reject' as const },
  createdAt: Date.now(),
  durationMs: 5000,
});

describe('DebateSimulatorService', () => {
  beforeEach(() => {
    debateSimulator.clearStoredDebates();
    debateSimulator.clearSimulations();
  });

  describe('storeDebate', () => {
    it('should store a debate', () => {
      const debate = createTestDebate('debate-1');
      debateSimulator.storeDebate(debate);
      
      const stored = debateSimulator.getStoredDebate('debate-1');
      expect(stored).toBeDefined();
      expect(stored?.id).toBe('debate-1');
    });

    it('should list stored debates', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      debateSimulator.storeDebate(createTestDebate('debate-2'));
      
      const debates = debateSimulator.listStoredDebates();
      expect(debates.length).toBe(2);
    });

    it('should respect limit when listing', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      debateSimulator.storeDebate(createTestDebate('debate-2'));
      debateSimulator.storeDebate(createTestDebate('debate-3'));
      
      const debates = debateSimulator.listStoredDebates(2);
      expect(debates.length).toBe(2);
    });
  });

  describe('replayDebate', () => {
    it('should replay with same config and get same outcome', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1', 'approved'));
      
      const result = debateSimulator.replayDebate('debate-1', {});
      expect(result.newOutcome).toBe('approved');
      expect(result.outcomeChanged).toBe(false);
    });

    it('should change outcome with different consensus mode', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1', 'approved'));
      
      const result = debateSimulator.replayDebate('debate-1', { consensusMode: 'unanimous' });
      expect(result.newOutcome).not.toBe('approved');
    });

    it('should filter team when specified', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      
      const result = debateSimulator.replayDebate('debate-1', { team: ['gpt-4', 'claude'] });
      expect(result.roundByRoundComparison[0].replayVotes).not.toHaveProperty('gemini');
    });

    it('should throw for non-existent debate', () => {
      expect(() => debateSimulator.replayDebate('nonexistent', {})).toThrow('not found');
    });

    it('should include analysis in result', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      
      const result = debateSimulator.replayDebate('debate-1', { consensusMode: 'weighted' });
      expect(result.analysis).toBeDefined();
      expect(result.analysis.consensusModeImpact).toContain('weighted');
    });
  });

  describe('simulateDebate', () => {
    it('should create a simulation', () => {
      const result = debateSimulator.simulateDebate({
        topic: 'Test',
        context: 'Context',
        consensusMode: 'majority',
        team: ['gpt-4', 'claude'],
        randomizeVotes: true,
      });
      
      expect(result.id).toMatch(/^sim_/);
      expect(result.isSimulation).toBe(true);
      expect(result.rounds.length).toBeGreaterThan(0);
    });

    it('should use mock responses when provided', () => {
      const result = debateSimulator.simulateDebate({
        topic: 'Test',
        context: 'Context',
        consensusMode: 'majority',
        team: ['gpt-4', 'claude'],
        mockResponses: {
          'gpt-4': { vote: 'approve', confidence: 0.95, reasoning: 'Mock approve' },
          'claude': { vote: 'reject', confidence: 0.85, reasoning: 'Mock reject' },
        },
      });
      
      expect(result.finalVotes['gpt-4']).toBe('approve');
      expect(result.finalVotes['claude']).toBe('reject');
    });

    it('should store simulation for retrieval', () => {
      const result = debateSimulator.simulateDebate({
        topic: 'Test',
        context: 'Context',
        consensusMode: 'majority',
        team: ['gpt-4'],
      });
      
      const retrieved = debateSimulator.getSimulation(result.id);
      expect(retrieved).toBeDefined();
      expect(retrieved?.id).toBe(result.id);
    });

    it('should list simulations', () => {
      debateSimulator.simulateDebate({
        topic: 'Test 1',
        context: 'Context',
        consensusMode: 'majority',
        team: ['gpt-4'],
      });
      debateSimulator.simulateDebate({
        topic: 'Test 2',
        context: 'Context',
        consensusMode: 'unanimous',
        team: ['claude'],
      });
      
      const simulations = debateSimulator.listSimulations();
      expect(simulations.length).toBe(2);
    });

    it('should bias toward specified outcome', () => {
      const approveResults = [];
      for (let i = 0; i < 10; i++) {
        const result = debateSimulator.simulateDebate({
          topic: 'Test',
          context: 'Context',
          consensusMode: 'majority',
          team: ['a', 'b', 'c'],
          randomizeVotes: true,
          biasToward: 'approve',
        });
        approveResults.push(result.outcome === 'approved');
      }
      
      const approvedCount = approveResults.filter(Boolean).length;
      expect(approvedCount).toBeGreaterThan(3);
    });
  });

  describe('runWhatIfAnalysis', () => {
    it('should run multiple scenarios', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      
      const results = debateSimulator.runWhatIfAnalysis('debate-1', [
        { consensusMode: 'majority' },
        { consensusMode: 'unanimous' },
        { consensusMode: 'weighted' },
      ]);
      
      expect(results.length).toBe(3);
      results.forEach((r) => {
        expect(r.scenario).toBeDefined();
        expect(r.result).toBeDefined();
      });
    });
  });

  describe('compareConsensusModes', () => {
    it('should compare all consensus modes', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      
      const results = debateSimulator.compareConsensusModes('debate-1');
      
      expect(results.majority).toBeDefined();
      expect(results.unanimous).toBeDefined();
      expect(results.supermajority).toBeDefined();
      expect(results.weighted).toBeDefined();
      expect(results.veto).toBeDefined();
    });

    it('should return outcome and rounds needed', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      
      const results = debateSimulator.compareConsensusModes('debate-1');
      
      expect(results.majority.outcome).toBeDefined();
      expect(results.majority.roundsNeeded).toBeDefined();
    });
  });

  describe('findOptimalTeam', () => {
    it('should find team for target outcome', () => {
      const debate = createTestDebate('debate-1');
      debate.rounds[0].votes = [
        { supervisorId: 'gpt-4', vote: 'approve', confidence: 0.9, reasoning: 'Good', tokensUsed: 300, responseTimeMs: 1500 },
        { supervisorId: 'claude', vote: 'approve', confidence: 0.8, reasoning: 'OK', tokensUsed: 250, responseTimeMs: 1200 },
        { supervisorId: 'gemini', vote: 'reject', confidence: 0.6, reasoning: 'Issues', tokensUsed: 200, responseTimeMs: 1000 },
      ];
      debateSimulator.storeDebate(debate);
      
      const result = debateSimulator.findOptimalTeam('debate-1', 'approved', 2);
      
      if (result) {
        expect(result.outcome).toBe('approved');
        expect(result.team.length).toBeGreaterThanOrEqual(2);
      }
    });

    it('should return null for non-existent debate', () => {
      const result = debateSimulator.findOptimalTeam('nonexistent', 'approved');
      expect(result).toBeNull();
    });
  });

  describe('getSummary', () => {
    it('should return summary statistics', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      debateSimulator.simulateDebate({
        topic: 'Test',
        context: 'Context',
        consensusMode: 'majority',
        team: ['gpt-4'],
      });
      
      const summary = debateSimulator.getSummary();
      expect(summary.storedDebatesCount).toBe(1);
      expect(summary.simulationsCount).toBe(1);
      expect(summary.availableConsensusModes.length).toBeGreaterThan(0);
    });
  });

  describe('clearStoredDebates', () => {
    it('should clear all stored debates', () => {
      debateSimulator.storeDebate(createTestDebate('debate-1'));
      debateSimulator.storeDebate(createTestDebate('debate-2'));
      
      debateSimulator.clearStoredDebates();
      
      expect(debateSimulator.listStoredDebates().length).toBe(0);
    });
  });

  describe('clearSimulations', () => {
    it('should clear all simulations', () => {
      debateSimulator.simulateDebate({
        topic: 'Test',
        context: 'Context',
        consensusMode: 'majority',
        team: ['gpt-4'],
      });
      
      debateSimulator.clearSimulations();
      
      expect(debateSimulator.listSimulations().length).toBe(0);
    });
  });

  describe('consensus calculators', () => {
    it('should approve with majority', () => {
      const debate = createTestDebate('debate-1');
      debate.rounds[0].votes = [
        { supervisorId: 'gpt-4', vote: 'approve', confidence: 0.9, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
        { supervisorId: 'claude', vote: 'approve', confidence: 0.8, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
        { supervisorId: 'gemini', vote: 'reject', confidence: 0.7, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
      ];
      debateSimulator.storeDebate(debate);
      
      const result = debateSimulator.replayDebate('debate-1', { consensusMode: 'majority' });
      expect(result.newOutcome).toBe('approved');
    });

    it('should deadlock with unanimous when mixed', () => {
      const debate = createTestDebate('debate-1');
      debate.rounds[0].votes = [
        { supervisorId: 'gpt-4', vote: 'approve', confidence: 0.9, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
        { supervisorId: 'claude', vote: 'approve', confidence: 0.8, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
        { supervisorId: 'gemini', vote: 'reject', confidence: 0.7, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
      ];
      debateSimulator.storeDebate(debate);
      
      const result = debateSimulator.replayDebate('debate-1', { consensusMode: 'unanimous' });
      expect(result.newOutcome).toBe('deadlock');
    });

    it('should reject with high-confidence veto', () => {
      const debate = createTestDebate('debate-1');
      debate.rounds[0].votes = [
        { supervisorId: 'gpt-4', vote: 'approve', confidence: 0.9, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
        { supervisorId: 'claude', vote: 'approve', confidence: 0.8, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
        { supervisorId: 'gemini', vote: 'reject', confidence: 0.95, reasoning: '', tokensUsed: 100, responseTimeMs: 500 },
      ];
      debateSimulator.storeDebate(debate);
      
      const result = debateSimulator.replayDebate('debate-1', { consensusMode: 'veto' });
      expect(result.newOutcome).toBe('rejected');
    });
  });
});
