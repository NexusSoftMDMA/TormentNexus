import { describe, it, expect, beforeEach, vi } from 'vitest';
import { CollaborativeDebateManagerService } from '../collaborative-debate-manager.js';
import type { CouncilDecision, DevelopmentTask } from '../types.js';

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

describe('CollaborativeDebateManagerService', () => {
  let manager: CollaborativeDebateManagerService;

  beforeEach(() => {
    manager = new CollaborativeDebateManagerService();
  });

  describe('Debate Lifecycle', () => {
    it('should create a debate', () => {
      const debate = manager.createDebate({
        title: 'Test Debate',
        description: 'Test description',
        taskType: 'code-review',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      expect(debate.id.startsWith('collab_')).toBe(true);
      expect(debate.title).toBe('Test Debate');
      expect(debate.status).toBe('draft');
      expect(debate.participants.size).toBe(1);
      expect(debate.participants.get('user1')?.role).toBe('owner');
    });

    it('should list debates with filters', () => {
      manager.createDebate({
        title: 'Debate 1',
        description: 'Desc',
        taskType: 'code-review',
        ownerId: 'user1',
        ownerName: 'User One',
        tags: ['important'],
      });
      manager.createDebate({
        title: 'Debate 2',
        description: 'Desc',
        taskType: 'security-audit',
        ownerId: 'user2',
        ownerName: 'User Two',
      });

      expect(manager.listDebates({ taskType: 'code-review' }).length).toBe(1);
      expect(manager.listDebates({ participantId: 'user1' }).length).toBe(1);
      expect(manager.listDebates({ tag: 'important' }).length).toBe(1);
    });

    it('should start inviting', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      const updated = manager.startInviting(debate.id);
      expect(updated.status).toBe('inviting');
    });

    it('should start debate with enough participants', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 2, maxParticipants: 10 },
      });

      manager.addParticipantDirectly({
        debateId: debate.id,
        addedBy: 'user1',
        participantId: 'user2',
        participantName: 'User Two',
        role: 'voter',
      });

      const started = manager.startDebate(debate.id);
      expect(started.status).toBe('in_progress');
      expect(started.startedAt).toBeDefined();
    });

    it('should reject starting with insufficient participants', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 3, maxParticipants: 10 },
      });

      expect(() => manager.startDebate(debate.id)).toThrow('at least 3 participants');
    });

    it('should start voting', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate.id);
      const voting = manager.startVoting(debate.id);

      expect(voting.status).toBe('voting');
      expect(voting.votingStartedAt).toBeDefined();
      expect(voting.expiresAt).toBeDefined();
    });

    it('should finalize debate', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate.id);
      manager.startVoting(debate.id);
      manager.submitVote({
        debateId: debate.id,
        participantId: 'user1',
        approved: true,
        confidence: 0.9,
        reasoning: 'LGTM',
      });

      const finalized = manager.finalizeDebate(debate.id);
      expect(finalized.status).toBe('completed');
      expect(finalized.finalDecision).toBeDefined();
      expect(finalized.finalDecision?.approved).toBe(true);
    });

    it('should cancel debate', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      const cancelled = manager.cancelDebate(debate.id, 'No longer needed');
      expect(cancelled.status).toBe('cancelled');
    });
  });

  describe('Participant Management', () => {
    it('should create invite token', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      const token = manager.createInvite({
        debateId: debate.id,
        inviterId: 'user1',
        role: 'voter',
      });

      expect(token.token.startsWith('inv_')).toBe(true);
      expect(token.debateId).toBe(debate.id);
      expect(token.role).toBe('voter');
    });

    it('should join with token', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      const token = manager.createInvite({
        debateId: debate.id,
        inviterId: 'user1',
        role: 'voter',
      });

      const result = manager.joinWithToken({
        token: token.token,
        participantId: 'user2',
        participantName: 'User Two',
      });

      expect(result.participant.id).toBe('user2');
      expect(result.debate.participants.size).toBe(2);
    });

    it('should add participant directly', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      const participant = manager.addParticipantDirectly({
        debateId: debate.id,
        addedBy: 'user1',
        participantId: 'user2',
        participantName: 'User Two',
        role: 'admin',
      });

      expect(participant.role).toBe('admin');
      expect(participant.status).toBe('joined');
    });

    it('should remove participant', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      manager.addParticipantDirectly({
        debateId: debate.id,
        addedBy: 'user1',
        participantId: 'user2',
        participantName: 'User Two',
        role: 'voter',
      });

      manager.removeParticipant(debate.id, 'user2', 'user1');
      const updated = manager.getDebate(debate.id);
      expect(updated?.participants.get('user2')?.status).toBe('kicked');
    });

    it('should not allow removing owner', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      expect(() => manager.removeParticipant(debate.id, 'user1', 'user1')).toThrow('Cannot remove owner');
    });

    it('should update participant role', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      manager.addParticipantDirectly({
        debateId: debate.id,
        addedBy: 'user1',
        participantId: 'user2',
        participantName: 'User Two',
        role: 'voter',
      });

      const updated = manager.updateParticipantRole(debate.id, 'user2', 'admin', 'user1');
      expect(updated.role).toBe('admin');
    });

    it('should get participant debates', () => {
      manager.createDebate({
        title: 'Debate 1',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });
      manager.createDebate({
        title: 'Debate 2',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      const debates = manager.getParticipantDebates('user1');
      expect(debates.length).toBe(2);
    });
  });

  describe('Messaging', () => {
    it('should add message', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate.id);
      const message = manager.addMessage({
        debateId: debate.id,
        participantId: 'user1',
        content: 'Hello world!',
      });

      expect(message.id.startsWith('msg_')).toBe(true);
      expect(message.content).toBe('Hello world!');
    });

    it('should edit message', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate.id);
      const message = manager.addMessage({
        debateId: debate.id,
        participantId: 'user1',
        content: 'Original',
      });

      const edited = manager.editMessage(debate.id, message.id, 'user1', 'Edited content');
      expect(edited.content).toBe('Edited content');
      expect(edited.edited).toBe(true);
    });

    it('should add reaction', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate.id);
      const message = manager.addMessage({
        debateId: debate.id,
        participantId: 'user1',
        content: 'Great idea!',
      });

      manager.addReaction(debate.id, message.id, 'user1', '👍');
      const messages = manager.getMessages(debate.id);
      expect(messages[0].reactions?.get('👍')).toContain('user1');
    });

    it('should get messages since timestamp', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate.id);
      manager.addMessage({ debateId: debate.id, participantId: 'user1', content: 'First' });
      
      const midTime = Date.now();
      manager.addMessage({ debateId: debate.id, participantId: 'user1', content: 'Second' });

      const recent = manager.getMessages(debate.id, midTime - 1);
      expect(recent.length).toBe(2);
    });

    it('should prevent observers from posting', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10, allowObservers: true },
      });

      manager.addParticipantDirectly({
        debateId: debate.id,
        addedBy: 'user1',
        participantId: 'observer1',
        participantName: 'Observer',
        role: 'observer',
      });

      manager.startDebate(debate.id);

      expect(() => manager.addMessage({
        debateId: debate.id,
        participantId: 'observer1',
        content: 'Should fail',
      })).toThrow('Observers cannot post');
    });
  });

  describe('Voting', () => {
    it('should submit vote', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate.id);
      manager.startVoting(debate.id);

      const vote = manager.submitVote({
        debateId: debate.id,
        participantId: 'user1',
        approved: true,
        confidence: 0.85,
        reasoning: 'Looks good to me',
        concerns: ['Minor issue'],
        suggestions: ['Consider refactoring'],
      });

      expect(vote.approved).toBe(true);
      expect(vote.confidence).toBe(0.85);
    });

    it('should allow revoting when enabled', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10, allowRevoting: true },
      });

      manager.startDebate(debate.id);
      manager.startVoting(debate.id);

      manager.submitVote({
        debateId: debate.id,
        participantId: 'user1',
        approved: false,
        confidence: 0.5,
        reasoning: 'Initial vote',
      });

      const revote = manager.submitVote({
        debateId: debate.id,
        participantId: 'user1',
        approved: true,
        confidence: 0.9,
        reasoning: 'Changed my mind',
      });

      expect(revote.approved).toBe(true);
      expect(manager.getDebate(debate.id)?.humanVotes.length).toBe(1);
    });

    it('should reject revoting when disabled', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10, allowRevoting: false },
      });

      manager.startDebate(debate.id);
      manager.startVoting(debate.id);

      manager.submitVote({
        debateId: debate.id,
        participantId: 'user1',
        approved: false,
        confidence: 0.5,
        reasoning: 'Initial vote',
      });

      expect(() => manager.submitVote({
        debateId: debate.id,
        participantId: 'user1',
        approved: true,
        confidence: 0.9,
        reasoning: 'Try to change',
      })).toThrow('revoting not allowed');
    });

    it('should prevent observers from voting', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.addParticipantDirectly({
        debateId: debate.id,
        addedBy: 'user1',
        participantId: 'observer1',
        participantName: 'Observer',
        role: 'observer',
      });

      manager.startDebate(debate.id);
      manager.startVoting(debate.id);

      expect(() => manager.submitVote({
        debateId: debate.id,
        participantId: 'observer1',
        approved: true,
        confidence: 0.5,
        reasoning: 'Should fail',
      })).toThrow('Observers cannot vote');
    });

    it('should get votes with anonymity during voting', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 2, maxParticipants: 10, anonymousVoting: true },
      });

      manager.addParticipantDirectly({
        debateId: debate.id,
        addedBy: 'user1',
        participantId: 'user2',
        participantName: 'User Two',
        role: 'voter',
      });

      manager.startDebate(debate.id);
      manager.startVoting(debate.id);

      manager.submitVote({
        debateId: debate.id,
        participantId: 'user2',
        approved: true,
        confidence: 0.9,
        reasoning: 'My secret vote',
      });

      const votes = manager.getVotes(debate.id, 'user2');
      expect(votes[0].participantId).toBe('anonymous');
    });
  });

  describe('AI Integration', () => {
    it('should set AI supervisor votes', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      manager.setAISupervisorVotes(debate.id, [
        { supervisor: 'GPT-4', approved: true, confidence: 0.9, weight: 1, comment: 'LGTM' },
        { supervisor: 'Claude', approved: true, confidence: 0.85, weight: 1, comment: 'Approved' },
      ]);

      const updated = manager.getDebate(debate.id);
      expect(updated?.aiSupervisorVotes?.length).toBe(2);
    });

    it('should include AI votes in decision calculation', () => {
      const debate = manager.createDebate({
        title: 'Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.setAISupervisorVotes(debate.id, [
        { supervisor: 'GPT-4', approved: true, confidence: 0.9, weight: 1, comment: 'LGTM' },
      ]);

      manager.startDebate(debate.id);
      manager.startVoting(debate.id);

      manager.submitVote({
        debateId: debate.id,
        participantId: 'user1',
        approved: false,
        confidence: 0.8,
        reasoning: 'Concerns',
      });

      const finalized = manager.finalizeDebate(debate.id);
      expect(finalized.finalDecision?.aiApprovalRate).toBe(1);
      expect(finalized.finalDecision?.humanApprovalRate).toBe(0);
    });
  });

  describe('Statistics', () => {
    it('should return correct statistics', () => {
      const debate1 = manager.createDebate({
        title: 'Debate 1',
        description: 'Desc',
        taskType: 'code-review',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate1.id);
      manager.startVoting(debate1.id);
      manager.submitVote({
        debateId: debate1.id,
        participantId: 'user1',
        approved: true,
        confidence: 0.9,
        reasoning: 'Good',
      });
      manager.finalizeDebate(debate1.id);

      const debate2 = manager.createDebate({
        title: 'Debate 2',
        description: 'Desc',
        taskType: 'security-audit',
        ownerId: 'user2',
        ownerName: 'User Two',
      });

      const stats = manager.getStatistics();
      expect(stats.totalDebates).toBe(2);
      expect(stats.completedDebates).toBe(1);
      expect(stats.activeDebates).toBe(0);
    });
  });

  describe('Export/Import', () => {
    it('should export debate', () => {
      const debate = manager.createDebate({
        title: 'Export Test',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
        config: { minParticipants: 1, maxParticipants: 10 },
      });

      manager.startDebate(debate.id);
      manager.addMessage({ debateId: debate.id, participantId: 'user1', content: 'Test message' });

      const exported = manager.exportDebate(debate.id);
      expect(exported.debate.title).toBe('Export Test');
      expect(exported.debate.participants.length).toBe(1);
    });

    it('should import debate', () => {
      const debate = manager.createDebate({
        title: 'Original',
        description: 'Desc',
        taskType: 'general',
        ownerId: 'user1',
        ownerName: 'User One',
      });

      const exported = manager.exportDebate(debate.id);
      const imported = manager.importDebate(exported);

      expect(imported.id).not.toBe(debate.id);
      expect(imported.title).toBe('Original');
      expect(imported.status).toBe('completed');
    });
  });
});
