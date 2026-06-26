import { describe, test, expect, beforeEach, afterEach } from 'vitest';
import { HumanVetoService } from '../human-veto.js';
import type { CouncilDecision, DevelopmentTask } from '../types.js';

describe('HumanVetoService', () => {
  let veto: HumanVetoService;

  const mockTask: DevelopmentTask = {
    id: 'test-task-1',
    description: 'Test task',
    context: '',
    files: [],
  };

  const mockDecision: CouncilDecision = {
    approved: true,
    consensus: 0.75,
    weightedConsensus: 0.8,
    votes: [],
    reasoning: 'Test reasoning',
    dissent: [],
  };

  beforeEach(() => {
    veto = new HumanVetoService();
  });

  afterEach(() => {
    veto.cleanup();
  });

  describe('isEnabled/setEnabled', () => {
    test('defaults to disabled', () => {
      expect(veto.isEnabled()).toBe(false);
    });

    test('can be enabled', () => {
      veto.setEnabled(true);
      expect(veto.isEnabled()).toBe(true);
    });

    test('clears pending decisions when disabled', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      await veto.submitForVeto(mockTask, mockDecision);
      expect(veto.getAllPending().length).toBe(1);
      
      veto.setEnabled(false);
      expect(veto.getAllPending().length).toBe(0);
    });
  });

  describe('getConfig/updateConfig', () => {
    test('returns default config', () => {
      const config = veto.getConfig();
      expect(config.enabled).toBe(false);
      expect(config.timeoutMs).toBe(300000);
      expect(config.autoApproveOnTimeout).toBe(true);
    });

    test('updates config', () => {
      veto.updateConfig({ timeoutMs: 60000, autoApproveOnTimeout: false });
      const config = veto.getConfig();
      expect(config.timeoutMs).toBe(60000);
      expect(config.autoApproveOnTimeout).toBe(false);
    });
  });

  describe('submitForVeto', () => {
    test('returns requiresVeto: false when disabled', async () => {
      const result = await veto.submitForVeto(mockTask, mockDecision);
      expect(result.requiresVeto).toBe(false);
      expect(result.decisionId).toBeUndefined();
    });

    test('auto-approves high consensus decisions', async () => {
      veto.setEnabled(true);
      const highConsensusDecision = { ...mockDecision, weightedConsensus: 0.95 };
      
      const result = await veto.submitForVeto(mockTask, highConsensusDecision);
      expect(result.requiresVeto).toBe(false);
    });

    test('requires veto for low consensus decisions', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      const result = await veto.submitForVeto(mockTask, mockDecision);
      expect(result.requiresVeto).toBe(true);
      expect(result.decisionId).toBeDefined();
    });

    test('emits decision_pending event', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      let eventData: any = null;
      veto.on('decision_pending', (data) => { eventData = data; });
      
      const result = await veto.submitForVeto(mockTask, mockDecision);
      
      expect(eventData).not.toBeNull();
      expect(eventData.decisionId).toBe(result.decisionId);
    });
  });

  describe('processVeto', () => {
    test('approves decision', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      const { decisionId } = await veto.submitForVeto(mockTask, mockDecision);
      
      const result = await veto.processVeto(decisionId!, {
        action: 'approve',
        reason: 'Looks good',
      });
      
      expect(result).not.toBeNull();
      expect(result!.approved).toBe(true);
      expect(result!.reasoning).toContain('APPROVED');
      expect(result!.reasoning).toContain('Looks good');
    });

    test('rejects decision', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      const { decisionId } = await veto.submitForVeto(mockTask, mockDecision);
      
      const result = await veto.processVeto(decisionId!, {
        action: 'reject',
        reason: 'Security concern',
      });
      
      expect(result).not.toBeNull();
      expect(result!.approved).toBe(false);
      expect(result!.reasoning).toContain('REJECTED');
    });

    test('requests redebate', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      const { decisionId } = await veto.submitForVeto(mockTask, mockDecision);
      
      let redebateEvent: any = null;
      veto.on('redebate_requested', (data) => { redebateEvent = data; });
      
      const result = await veto.processVeto(decisionId!, {
        action: 'redebate',
        reason: 'Need more analysis',
      });
      
      expect(result).toBeNull();
      expect(redebateEvent).not.toBeNull();
      expect(redebateEvent.reason as string).toBe('Need more analysis');
    });

    test('returns null for unknown decision', async () => {
      const result = await veto.processVeto('unknown-id', { action: 'approve' });
      expect(result).toBeNull();
    });

    test('clears timeout after processing', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0, timeoutMs: 100 });
      
      const { decisionId } = await veto.submitForVeto(mockTask, mockDecision);
      await veto.processVeto(decisionId!, { action: 'approve' });
      
      await new Promise(resolve => setTimeout(resolve, 150));
      expect(veto.getHistory().length).toBe(1);
    });
  });

  describe('getPendingDecision/getAllPending', () => {
    test('returns pending decision by id', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      const { decisionId } = await veto.submitForVeto(mockTask, mockDecision);
      
      const pending = veto.getPendingDecision(decisionId!);
      expect(pending).toBeDefined();
      expect(pending!.id).toBe(decisionId!);
      expect(pending!.status).toBe('pending');
    });

    test('returns all pending decisions', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      await veto.submitForVeto(mockTask, mockDecision);
      await veto.submitForVeto({ ...mockTask, id: 'task-2' }, mockDecision);
      
      expect(veto.getAllPending().length).toBe(2);
    });
  });

  describe('getHistory', () => {
    test('returns processed decisions', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      const { decisionId } = await veto.submitForVeto(mockTask, mockDecision);
      await veto.processVeto(decisionId!, { action: 'approve' });
      
      const history = veto.getHistory();
      expect(history.length).toBe(1);
      expect(history[0].status).toBe('approved');
    });

    test('respects limit', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      for (let i = 0; i < 5; i++) {
        const { decisionId } = await veto.submitForVeto(
          { ...mockTask, id: `task-${i}` }, 
          mockDecision
        );
        await veto.processVeto(decisionId!, { action: 'approve' });
      }
      
      expect(veto.getHistory(3).length).toBe(3);
    });
  });

  describe('getStats', () => {
    test('returns correct statistics', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ minConsensusForAutoApprove: 1.0 });
      
      await veto.submitForVeto(mockTask, mockDecision);
      
      const stats = veto.getStats();
      expect(stats.enabled).toBe(true);
      expect(stats.pendingCount).toBe(1);
      expect(stats.historyCount).toBe(0);
    });
  });

  describe('timeout handling', () => {
    test('auto-approves on timeout when configured', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ 
        minConsensusForAutoApprove: 1.0,
        timeoutMs: 50,
        autoApproveOnTimeout: true,
      });
      
      let timeoutEvent: any = null;
      veto.on('decision_timeout', (data) => { timeoutEvent = data; });
      
      await veto.submitForVeto(mockTask, mockDecision);
      
      await new Promise(resolve => setTimeout(resolve, 100));
      
      expect(timeoutEvent).not.toBeNull();
      expect(timeoutEvent.autoApproved).toBe(true);
      expect(veto.getAllPending().length).toBe(0);
    });

    test('does not auto-approve when disabled', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ 
        minConsensusForAutoApprove: 1.0,
        timeoutMs: 50,
        autoApproveOnTimeout: false,
      });
      
      let timeoutEvent: any = null;
      veto.on('decision_timeout', (data) => { timeoutEvent = data; });
      
      await veto.submitForVeto(mockTask, mockDecision);
      
      await new Promise(resolve => setTimeout(resolve, 100));
      
      expect(timeoutEvent.autoApproved).toBe(false);
      expect(timeoutEvent.finalDecision).toBeNull();
    });
  });

  describe('requireVetoForRejections', () => {
    test('requires veto for rejected decisions when enabled', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ 
        minConsensusForAutoApprove: 0.5,
        requireVetoForRejections: true,
      });
      
      const rejectedDecision = { ...mockDecision, approved: false, weightedConsensus: 0.8 };
      const result = await veto.submitForVeto(mockTask, rejectedDecision);
      
      expect(result.requiresVeto).toBe(true);
    });

    test('auto-approves rejected decisions when disabled', async () => {
      veto.setEnabled(true);
      veto.updateConfig({ 
        minConsensusForAutoApprove: 0.5,
        requireVetoForRejections: false,
      });
      
      const rejectedDecision = { ...mockDecision, approved: false, weightedConsensus: 0.8 };
      const result = await veto.submitForVeto(mockTask, rejectedDecision);
      
      expect(result.requiresVeto).toBe(false);
    });
  });
});
