import { describe, test, expect, beforeEach, vi } from 'vitest';
import { DebateHistoryService, type DebateRecord } from '../debate-history.js';
import type { DevelopmentTask, CouncilDecision, Vote } from '../types.js';

import { resetMockDb, createDrizzleMock, createSchemaMock } from './mock-db.js';

vi.mock('../db.js', () => ({
  dbService: {
    getDb: vi.fn(),
    getDrizzle: vi.fn(() => createDrizzleMock()),
    getSchema: vi.fn(() => createSchemaMock()),
    close: vi.fn(),
  }
}));

const { dbService } = await import('../db.js');
const mockGetDb = vi.mocked(dbService.getDb);
const mockGetDrizzle = vi.mocked(dbService.getDrizzle);
const mockGetSchema = vi.mocked(dbService.getSchema);

describe('DebateHistoryService', () => {
  let service: DebateHistoryService;

  const mockTask: DevelopmentTask = {
    id: 'task-1',
    description: 'Implement user authentication',
    context: 'Building a secure login system',
    files: ['auth.ts', 'user.ts'],
  };

  const mockVotes: Vote[] = [
    { supervisor: 'GPT-4', approved: true, confidence: 0.9, weight: 1, comment: 'Looks good' },
    { supervisor: 'Claude', approved: true, confidence: 0.85, weight: 1, comment: 'Approved with suggestions' },
    { supervisor: 'Gemini', approved: false, confidence: 0.7, weight: 1, comment: 'Needs more tests' },
  ];

  const mockDecision: CouncilDecision = {
    approved: true,
    consensus: 0.67,
    weightedConsensus: 0.68,
    votes: mockVotes,
    reasoning: 'Majority approved the implementation',
    dissent: ['Gemini suggested more tests'],
  };

  beforeEach(async () => {
    vi.clearAllMocks();
    resetMockDb();
    mockGetDrizzle.mockImplementation(() => createDrizzleMock());
    mockGetSchema.mockImplementation(() => createSchemaMock());
    service = new DebateHistoryService();
    service.updateConfig({ enabled: true, autoSave: true });
    await service.clearAll();
  });

  describe('saveDebate', () => {
    test('saves debate and returns record with id', async () => {
      const record = await service.saveDebate(mockTask, mockDecision, {
        debateRounds: 2,
        consensusMode: 'weighted',
      });

      expect(record.id).toMatch(/^debate_/);
      expect(record.task).toEqual(mockTask);
      expect(record.decision).toEqual(mockDecision);
      expect(record.metadata.debateRounds).toBe(2);
      expect(record.metadata.consensusMode).toBe('weighted');
      expect(record.metadata.supervisorCount).toBe(3);
      expect(record.metadata.participatingSupervisors).toEqual(['GPT-4', 'Claude', 'Gemini']);
    });

    test('emits debate_saved event', async () => {
      let eventData: any = null;
      service.on('debate_saved', (data) => { eventData = data; });

      await service.saveDebate(mockTask, mockDecision, {});

      expect(eventData).not.toBeNull();
      expect(eventData.task.id).toBe('task-1');
    });

    test('stores dynamic selection metadata', async () => {
      const record = await service.saveDebate(mockTask, mockDecision, {
        dynamicSelection: {
          enabled: true,
          taskType: 'security-audit',
          confidence: 0.95,
        },
      });

      expect(record.metadata.dynamicSelection?.enabled).toBe(true);
      expect(record.metadata.dynamicSelection?.taskType).toBe('security-audit');
      expect(record.metadata.dynamicSelection?.confidence).toBe(0.95);
    });
  });

  describe('getDebate', () => {
    test('retrieves saved debate by id', async () => {
      const saved = await service.saveDebate(mockTask, mockDecision, {});
      const retrieved = await service.getDebate(saved.id);

      expect(retrieved).toEqual(saved);
    });

    test('returns undefined for non-existent id', async () => {
      const result = await service.getDebate('non-existent-id');
      expect(result).toBeUndefined();
    });
  });

  describe('deleteRecord', () => {
    test('deletes existing record', async () => {
      const saved = await service.saveDebate(mockTask, mockDecision, {});
      const deleted = await service.deleteRecord(saved.id);

      expect(deleted).toBe(true);
      expect(await service.getDebate(saved.id)).toBeUndefined();
    });

    test('returns false for non-existent record', async () => {
      const deleted = await service.deleteRecord('non-existent');
      expect(deleted).toBe(false);
    });

    test('emits debate_deleted event', async () => {
      let eventData: any = null;
      service.on('debate_deleted', (data) => { eventData = data; });

      const saved = await service.saveDebate(mockTask, mockDecision, {});
      await service.deleteRecord(saved.id);

      expect(eventData).not.toBeNull();
      expect(eventData.id).toBe(saved.id);
    });
  });

  describe('queryDebates', () => {
    beforeEach(async () => {
      await service.saveDebate(mockTask, mockDecision, { sessionId: 'session-1', consensusMode: 'weighted' });
      await service.saveDebate(
        { ...mockTask, id: 'task-2' },
        { ...mockDecision, approved: false, consensus: 0.3 },
        { sessionId: 'session-2', consensusMode: 'unanimous' }
      );
      await service.saveDebate(
        { ...mockTask, id: 'task-3' },
        { ...mockDecision, consensus: 0.9 },
        { sessionId: 'session-1', consensusMode: 'weighted' }
      );
    });

    test('returns all records by default', async () => {
      const results = await service.queryDebates();
      expect(results.length).toBe(3);
    });

    test('filters by sessionId', async () => {
      const results = await service.queryDebates({ sessionId: 'session-1' });
      expect(results.length).toBe(2);
    });

    test('filters by approved status', async () => {
      const approved = await service.queryDebates({ approved: true });
      expect(approved.length).toBe(2);

      const rejected = await service.queryDebates({ approved: false });
      expect(rejected.length).toBe(1);
    });

    test('filters by supervisor name', async () => {
      const results = await service.queryDebates({ supervisorName: 'Claude' });
      expect(results.length).toBe(3);
    });

    test.skip('filters by consensus range', async () => {
      const highConsensus = await service.queryDebates({ minConsensus: 0.8 });
      expect(highConsensus.length).toBe(1);

      const lowConsensus = await service.queryDebates({ maxConsensus: 0.5 });
      expect(lowConsensus.length).toBe(1);
    });

    test.skip('sorts by timestamp descending by default', async () => {
      const results = await service.queryDebates();
      for (let i = 1; i < results.length; i++) {
        expect(results[i - 1].timestamp).toBeGreaterThanOrEqual(results[i].timestamp);
      }
    });

    test.skip('sorts by consensus', async () => {
      const results = await service.queryDebates({ sortBy: 'consensus', sortOrder: 'asc' });
      for (let i = 1; i < results.length; i++) {
        expect(results[i - 1].decision.consensus).toBeLessThanOrEqual(results[i].decision.consensus);
      }
    });

    test('applies pagination', async () => {
      const page1 = await service.queryDebates({ limit: 2, offset: 0 });
      const page2 = await service.queryDebates({ limit: 2, offset: 2 });

      expect(page1.length).toBe(2);
      expect(page2.length).toBe(1);
    });
  });

  describe('getStats', () => {
    test('returns empty stats for no records', async () => {
      const stats = await service.getStats();

      expect(stats.totalDebates).toBe(0);
      expect(stats.approvalRate).toBe(0);
      expect(stats.averageConsensus).toBe(0);
    });

    test('calculates correct statistics', async () => {
      await service.saveDebate(mockTask, mockDecision, { consensusMode: 'weighted' });
      await service.saveDebate(
        { ...mockTask, id: 'task-2' },
        { ...mockDecision, approved: false, consensus: 0.4 },
        { consensusMode: 'weighted' }
      );

      const stats = await service.getStats();

      expect(stats.totalDebates).toBe(2);
      expect(stats.approvedCount).toBe(1);
      expect(stats.rejectedCount).toBe(1);
      expect(stats.approvalRate).toBe(0.5);
      expect(stats.debatesByConsensusMode['weighted']).toBe(2);
    });

    test('groups by supervisor', async () => {
      await service.saveDebate(mockTask, mockDecision, {});

      const stats = await service.getStats();

      expect(stats.debatesBySupervisor['GPT-4']).toBe(1);
      expect(stats.debatesBySupervisor['Claude']).toBe(1);
      expect(stats.debatesBySupervisor['Gemini']).toBe(1);
    });
  });

  describe('getSupervisorVoteHistory', () => {
    test('returns empty history for unknown supervisor', async () => {
      const history = await service.getSupervisorVoteHistory('Unknown');

      expect(history.totalVotes).toBe(0);
      expect(history.approvals).toBe(0);
      expect(history.recentVotes).toEqual([]);
    });

    test('tracks supervisor voting patterns', async () => {
      await service.saveDebate(mockTask, mockDecision, {});
      await service.saveDebate({ ...mockTask, id: 'task-2' }, mockDecision, {});

      const gpt4History = await service.getSupervisorVoteHistory('GPT-4');
      expect(gpt4History.totalVotes).toBe(2);
      expect(gpt4History.approvals).toBe(2);
      expect(gpt4History.averageConfidence).toBe(0.9);

      const geminiHistory = await service.getSupervisorVoteHistory('Gemini');
      expect(geminiHistory.totalVotes).toBe(2);
      expect(geminiHistory.approvals).toBe(0);
      expect(geminiHistory.rejections).toBe(2);
    });

    test('returns recent votes limited to 10', async () => {
      for (let i = 0; i < 15; i++) {
        await service.saveDebate({ ...mockTask, id: `task-${i}` }, mockDecision, {});
      }

      const history = await service.getSupervisorVoteHistory('Claude');
      expect(history.recentVotes.length).toBe(10);
    });
  });

  describe('export functions', () => {
    beforeEach(async () => {
      await service.saveDebate(mockTask, mockDecision, { sessionId: 'test-session' });
    });

    test('exportToJson returns valid JSON', async () => {
      const json = await service.exportToJson();
      const parsed = JSON.parse(json);

      expect(Array.isArray(parsed)).toBe(true);
      expect(parsed.length).toBe(1);
      expect(parsed[0].task.id).toBe('task-1');
    });

    test('exportToCsv returns valid CSV', async () => {
      const csv = await service.exportToCsv();
      const lines = csv.split('\n');

      expect(lines.length).toBe(2);
      expect(lines[0]).toContain('id,timestamp');
      expect(lines[1]).toContain('task-1');
    });
  });

  describe('clearAll', () => {
    test('removes all records', async () => {
      await service.saveDebate(mockTask, mockDecision, {});
      await service.saveDebate({ ...mockTask, id: 'task-2' }, mockDecision, {});

      const count = await service.clearAll();

      expect(count).toBe(2);
      expect(await service.getRecordCount()).toBe(0);
    });

    test('emits cleared event', async () => {
      let eventData: any = null;
      service.on('cleared', (data) => { eventData = data; });

      await service.saveDebate(mockTask, mockDecision, {});
      await service.clearAll();

      expect(eventData).not.toBeNull();
      expect(eventData.count).toBe(1);
    });
  });

  describe('config management', () => {
    test('getConfig returns current config', async () => {
      const config = service.getConfig();

      expect(config).toHaveProperty('enabled');
      expect(config).toHaveProperty('maxRecords');
      expect(config).toHaveProperty('retentionDays');
    });

    test('updateConfig updates and returns new config', async () => {
      const updated = service.updateConfig({ maxRecords: 500 });

      expect(updated.maxRecords).toBe(500);
      expect(service.getConfig().maxRecords).toBe(500);
    });

    test('emits config_updated event', async () => {
      let eventData: any = null;
      service.on('config_updated', (data) => { eventData = data; });

      service.updateConfig({ retentionDays: 30 });

      expect(eventData).not.toBeNull();
      expect(eventData.retentionDays).toBe(30);
    });
  });

  describe('isEnabled/getRecordCount', () => {
    test('isEnabled reflects config', async () => {
      service.updateConfig({ enabled: true });
      expect(service.isEnabled()).toBe(true);

      service.updateConfig({ enabled: false });
      expect(service.isEnabled()).toBe(false);
    });

    test('getRecordCount returns correct count', async () => {
      expect(await service.getRecordCount()).toBe(0);

      await service.saveDebate(mockTask, mockDecision, {});
      expect(await service.getRecordCount()).toBe(1);

      await service.saveDebate({ ...mockTask, id: 'task-2' }, mockDecision, {});
      expect(await service.getRecordCount()).toBe(2);
    });

    test('initialize disables persistence cleanly when SQLite is unavailable', async () => {
      const sqliteUnavailable = new Error('SQLite runtime is unavailable');
      mockGetDrizzle.mockImplementation(() => {
        throw sqliteUnavailable;
      });
      const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      const storageErrors: Array<{ action: string; error: unknown }> = [];
      const initializedEvents: Array<{ recordCount: number; storageAvailable?: boolean }> = [];
      service.on('storage_error', (event) => {
        storageErrors.push(event as { action: string; error: unknown });
      });
      service.on('initialized', (event) => {
        initializedEvents.push(event as { recordCount: number; storageAvailable?: boolean });
      });

      await expect(service.initialize()).resolves.toBeUndefined();

      expect(service.isEnabled()).toBe(false);
      expect(storageErrors).toEqual([
        { action: 'initialize', error: sqliteUnavailable },
      ]);
      expect(initializedEvents).toContainEqual({ recordCount: 0, storageAvailable: false });
      expect(warnSpy).toHaveBeenCalledWith(
        '[DebateHistory] Disabling debate history persistence for this run because SQLite is unavailable:',
        'Debate history initialization failed: SQLite runtime is unavailable for this run.',
      );
    });

    test('initialize reports non-SQLite failures honestly', async () => {
      const initializationFailure = new Error('schema mismatch');
      mockGetDrizzle.mockImplementation(() => {
        throw initializationFailure;
      });
      const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      await expect(service.initialize()).resolves.toBeUndefined();

      expect(service.isEnabled()).toBe(false);
      expect(warnSpy).toHaveBeenCalledWith(
        '[DebateHistory] Disabling debate history persistence for this run due to initialization failure:',
        'Debate history initialization failed: schema mismatch',
      );
    });
  });

  describe('pruning', () => {
    test('prunes when exceeding maxRecords', async () => {
      service.updateConfig({ maxRecords: 3 });

      for (let i = 0; i < 5; i++) {
        await service.saveDebate({ ...mockTask, id: `task-${i}` }, mockDecision, {});
      }

      expect(await service.getRecordCount()).toBe(3);
    });
  });
});
