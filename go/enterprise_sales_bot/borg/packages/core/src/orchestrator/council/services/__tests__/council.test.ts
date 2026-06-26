import { describe, test, expect, beforeEach, vi } from 'vitest';
import { SupervisorCouncil } from '../council.js';

vi.mock('../db.js', () => ({
  dbService: {
    getDb: vi.fn(),
    getDrizzle: vi.fn(() => {
      const dbMock: any = {
        insert: vi.fn().mockReturnThis(),
        values: vi.fn().mockReturnThis(),
        select: vi.fn().mockReturnThis(),
        from: vi.fn().mockReturnThis(),
        where: vi.fn().mockReturnThis(),
        returning: vi.fn().mockReturnThis(),
        delete: vi.fn().mockReturnThis(),
        update: vi.fn().mockReturnThis(),
        set: vi.fn().mockReturnThis(),
        limit: vi.fn().mockReturnThis(),
        orderBy: vi.fn().mockReturnThis(),
        then: vi.fn((resolve) => resolve([])),
      };
      return dbMock;
    }),
    getSchema: vi.fn(() => new Proxy({}, {
      get: (target, prop) => ({ name: prop })
    })),
    close: vi.fn(),
  }
}));
import type { Supervisor, Message, DevelopmentTask } from '../types.js';

function createMockSupervisor(name: string, responses: string[]): Supervisor {
  let callIndex = 0;
  return {
    name,
    provider: 'mock',
    async chat(_messages: Message[]): Promise<string> {
      const response = responses[callIndex] || responses[responses.length - 1];
      callIndex++;
      return response;
    },
    async isAvailable(): Promise<boolean> {
      return true;
    },
  };
}

function createTask(overrides: Partial<DevelopmentTask> = {}): DevelopmentTask {
  return {
    id: 'task-1',
    description: 'Test task',
    context: 'Testing context',
    files: ['test.ts'],
    ...overrides,
  };
}

describe('SupervisorCouncil', () => {
  let council: SupervisorCouncil;

  beforeEach(() => {
    council = new SupervisorCouncil({
      supervisors: [],
      debateRounds: 2,
      consensusThreshold: 0.5,
      weightedVoting: true,
    });
  });

  describe('addSupervisor', () => {
    test('adds supervisor with default weight', () => {
      const supervisor = createMockSupervisor('Test', ['response']);
      council.addSupervisor(supervisor);
      
      expect(council.getSupervisors()).toHaveLength(1);
      expect(council.getSupervisorWeight('Test')).toBe(1.0);
    });

    test('adds supervisor with custom weight', () => {
      const supervisor = createMockSupervisor('Test', ['response']);
      council.addSupervisor(supervisor, 1.5);
      
      expect(council.getSupervisorWeight('Test')).toBe(1.5);
    });
  });

  describe('setSupervisorWeight', () => {
    test('clamps weight between 0 and 2', () => {
      const supervisor = createMockSupervisor('Test', ['response']);
      council.addSupervisor(supervisor);
      
      council.setSupervisorWeight('Test', -1);
      expect(council.getSupervisorWeight('Test')).toBe(0);
      
      council.setSupervisorWeight('Test', 5);
      expect(council.getSupervisorWeight('Test')).toBe(2);
    });
  });

  describe('debate', () => {
    test('returns auto-approve when no supervisors available', async () => {
      const result = await council.debate(createTask());
      
      expect(result.approved).toBe(true);
      expect(result.consensus).toBe(1.0);
      expect(result.votes).toHaveLength(0);
    });

    test('parses APPROVE vote correctly', async () => {
      const supervisor = createMockSupervisor('Approver', [
        'Initial opinion',
        'Refined opinion',
        'VOTE: APPROVE\nCONFIDENCE: 0.9\nREASONING: Looks good',
      ]);
      council.addSupervisor(supervisor);
      
      const result = await council.debate(createTask());
      
      expect(result.votes[0].approved).toBe(true);
      expect(result.votes[0].confidence).toBe(0.9);
    });

    test('parses REJECT vote correctly', async () => {
      const supervisor = createMockSupervisor('Rejecter', [
        'Initial opinion',
        'Refined opinion',
        'VOTE: REJECT\nCONFIDENCE: 0.8\nREASONING: Has issues',
      ]);
      council.addSupervisor(supervisor);
      
      const result = await council.debate(createTask());
      
      expect(result.votes[0].approved).toBe(false);
      expect(result.votes[0].confidence).toBe(0.8);
    });

    test('calculates simple consensus correctly', async () => {
      const approver = createMockSupervisor('Approver', [
        'Opinion 1', 'Refined 1', 'VOTE: APPROVE\nCONFIDENCE: 1.0',
      ]);
      const rejecter = createMockSupervisor('Rejecter', [
        'Opinion 2', 'Refined 2', 'VOTE: REJECT\nCONFIDENCE: 1.0',
      ]);
      council.addSupervisor(approver);
      council.addSupervisor(rejecter);
      
      const result = await council.debate(createTask());
      
      expect(result.consensus).toBe(0.5);
    });

    test('calculates weighted consensus with confidence', async () => {
      const highConfApprover = createMockSupervisor('HighConf', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 1.0',
      ]);
      const lowConfRejecter = createMockSupervisor('LowConf', [
        'Opinion', 'Refined', 'VOTE: REJECT\nCONFIDENCE: 0.2',
      ]);
      council.addSupervisor(highConfApprover, 1.0);
      council.addSupervisor(lowConfRejecter, 1.0);
      
      const result = await council.debate(createTask());
      
      expect(result.weightedConsensus).toBeGreaterThan(0.4);
      expect(result.weightedConsensus).toBe(0.5);
    });

    test('respects supervisor weights', async () => {
      const lightApprover = createMockSupervisor('Light', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 1.0',
      ]);
      const heavyRejecter = createMockSupervisor('Heavy', [
        'Opinion', 'Refined', 'VOTE: REJECT\nCONFIDENCE: 1.0',
      ]);
      council.addSupervisor(lightApprover, 0.5);
      council.addSupervisor(heavyRejecter, 2.0);
      
      const result = await council.debate(createTask());
      
      expect(result.weightedConsensus).toBeLessThan(0.5);
    });

    test('tracks strong dissent (high confidence rejections)', async () => {
      const approver = createMockSupervisor('Approver', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 0.6',
      ]);
      const strongRejecter = createMockSupervisor('StrongRejecter', [
        'Opinion', 'Refined', 'VOTE: REJECT\nCONFIDENCE: 0.9\nREASONING: Critical security flaw',
      ]);
      council.addSupervisor(approver);
      council.addSupervisor(strongRejecter);
      
      const result = await council.debate(createTask());
      
      expect(result.dissent).toHaveLength(1);
      expect(result.dissent![0]).toContain('StrongRejecter');
    });

    test('does not track weak dissent (low confidence rejections)', async () => {
      const approver = createMockSupervisor('Approver', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 0.8',
      ]);
      const weakRejecter = createMockSupervisor('WeakRejecter', [
        'Opinion', 'Refined', 'VOTE: REJECT\nCONFIDENCE: 0.5',
      ]);
      council.addSupervisor(approver);
      council.addSupervisor(weakRejecter);
      
      const result = await council.debate(createTask());
      
      expect(result.dissent).toHaveLength(0);
    });
  });

  describe('parseConfidence', () => {
    test('handles percentage format', async () => {
      const supervisor = createMockSupervisor('Test', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 85',
      ]);
      council.addSupervisor(supervisor);
      
      const result = await council.debate(createTask());
      
      expect(result.votes[0].confidence).toBe(0.85);
    });

    test('handles alternative confidence format', async () => {
      const supervisor = createMockSupervisor('Test', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nconfidence 0.75\nREASONING: Good',
      ]);
      council.addSupervisor(supervisor);
      
      const result = await council.debate(createTask());
      
      expect(result.votes[0].confidence).toBe(0.75);
    });

    test('defaults to 0.7 when confidence not specified', async () => {
      const supervisor = createMockSupervisor('Test', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nREASONING: Good',
      ]);
      council.addSupervisor(supervisor);
      
      const result = await council.debate(createTask());
      
      expect(result.votes[0].confidence).toBe(0.7);
    });
  });

  describe('consensus threshold', () => {
    test('approves when weighted consensus meets threshold', async () => {
      council.setConsensusThreshold(0.5);
      
      const approver1 = createMockSupervisor('A1', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 0.9',
      ]);
      const approver2 = createMockSupervisor('A2', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 0.8',
      ]);
      const rejecter = createMockSupervisor('R1', [
        'Opinion', 'Refined', 'VOTE: REJECT\nCONFIDENCE: 0.5',
      ]);
      
      council.addSupervisor(approver1);
      council.addSupervisor(approver2);
      council.addSupervisor(rejecter);
      
      const result = await council.debate(createTask());
      
      expect(result.approved).toBe(true);
    });

    test('rejects when weighted consensus below threshold', async () => {
      council.setConsensusThreshold(0.8);
      
      const approver = createMockSupervisor('A1', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 0.6',
      ]);
      const rejecter = createMockSupervisor('R1', [
        'Opinion', 'Refined', 'VOTE: REJECT\nCONFIDENCE: 0.9',
      ]);
      
      council.addSupervisor(approver);
      council.addSupervisor(rejecter);
      
      const result = await council.debate(createTask());
      
      expect(result.approved).toBe(false);
    });
  });

  describe('weighted voting toggle', () => {
    test('uses simple consensus when weightedVoting disabled', async () => {
      council.setWeightedVoting(false);
      council.setConsensusThreshold(0.5);
      
      const highConfApprover = createMockSupervisor('HighConf', [
        'Opinion', 'Refined', 'VOTE: APPROVE\nCONFIDENCE: 0.1',
      ]);
      const lowConfRejecter = createMockSupervisor('LowConf', [
        'Opinion', 'Refined', 'VOTE: REJECT\nCONFIDENCE: 0.9',
      ]);
      council.addSupervisor(highConfApprover);
      council.addSupervisor(lowConfRejecter);
      
      const result = await council.debate(createTask());
      
      expect(result.consensus).toBe(0.5);
      expect(result.approved).toBe(true);
    });
  });

  describe('clearSupervisors', () => {
    test('removes all supervisors and weights', () => {
      council.addSupervisor(createMockSupervisor('A', ['r']), 1.5);
      council.addSupervisor(createMockSupervisor('B', ['r']), 0.8);
      
      council.clearSupervisors();
      
      expect(council.getSupervisors()).toHaveLength(0);
      expect(council.getSupervisorWeight('A')).toBe(1.0);
    });
  });
});
