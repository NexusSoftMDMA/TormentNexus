import { describe, it, expect, beforeEach } from 'vitest';
import { supervisorAnalytics } from '../supervisor-analytics.js';

describe('SupervisorAnalyticsService', () => {
  beforeEach(() => {
    supervisorAnalytics.clearHistory();
  });

  describe('recordVote', () => {
    it('should record a vote', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500, 'approve');
      
      const metrics = supervisorAnalytics.getSupervisorMetrics('gpt-4');
      expect(metrics.totalVotes).toBe(1);
      expect(metrics.approvalRate).toBe(1);
    });

    it('should track consensus agreement', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500, 'approve');
      supervisorAnalytics.recordVote('gpt-4', 'debate-2', 'reject', 0.8, 1200, 400, 'approve');
      
      const metrics = supervisorAnalytics.getSupervisorMetrics('gpt-4');
      expect(metrics.consensusAgreementRate).toBe(0.5);
    });

    it('should track multiple supervisors', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500);
      supervisorAnalytics.recordVote('claude', 'debate-1', 'reject', 0.8, 1200, 400);
      
      const allMetrics = supervisorAnalytics.getAllSupervisorMetrics();
      expect(allMetrics.size).toBe(2);
    });
  });

  describe('recordDebateOutcome', () => {
    it('should record a debate outcome', () => {
      supervisorAnalytics.recordDebateOutcome(
        'debate-1',
        'Code review',
        'majority',
        'approved',
        ['gpt-4', 'claude'],
        3,
        5000
      );
      
      const stats = supervisorAnalytics.getDebateStatistics();
      expect(stats.total).toBe(1);
      expect(stats.byOutcome.approved).toBe(1);
    });

    it('should track participation', () => {
      supervisorAnalytics.recordDebateOutcome(
        'debate-1',
        'Topic 1',
        'majority',
        'approved',
        ['gpt-4', 'claude', 'gemini'],
        2,
        3000
      );
      supervisorAnalytics.recordDebateOutcome(
        'debate-2',
        'Topic 2',
        'unanimous',
        'rejected',
        ['gpt-4', 'claude'],
        4,
        8000
      );
      
      const stats = supervisorAnalytics.getDebateStatistics();
      expect(stats.participationBySupervisor['gpt-4']).toBe(2);
      expect(stats.participationBySupervisor['claude']).toBe(2);
      expect(stats.participationBySupervisor['gemini']).toBe(1);
    });
  });

  describe('getSupervisorMetrics', () => {
    it('should return empty metrics for unknown supervisor', () => {
      const metrics = supervisorAnalytics.getSupervisorMetrics('unknown');
      expect(metrics.totalVotes).toBe(0);
      expect(metrics.approvalRate).toBe(0);
    });

    it('should calculate correct rates', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500);
      supervisorAnalytics.recordVote('gpt-4', 'debate-2', 'approve', 0.8, 1200, 400);
      supervisorAnalytics.recordVote('gpt-4', 'debate-3', 'reject', 0.7, 1000, 300);
      supervisorAnalytics.recordVote('gpt-4', 'debate-4', 'abstain', 0.5, 800, 200);
      
      const metrics = supervisorAnalytics.getSupervisorMetrics('gpt-4');
      expect(metrics.approvalRate).toBe(0.5);
      expect(metrics.rejectionRate).toBe(0.25);
      expect(metrics.abstentionRate).toBe(0.25);
    });

    it('should calculate average confidence and response time', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.8, 1000, 500);
      supervisorAnalytics.recordVote('gpt-4', 'debate-2', 'approve', 0.6, 2000, 400);
      
      const metrics = supervisorAnalytics.getSupervisorMetrics('gpt-4');
      expect(metrics.avgConfidence).toBe(0.7);
      expect(metrics.avgResponseTimeMs).toBe(1500);
    });

    it('should filter by time range', () => {
      const now = Date.now();
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500);
      
      const futureRange = { start: now + 1000, end: now + 2000 };
      const metrics = supervisorAnalytics.getSupervisorMetrics('gpt-4', futureRange);
      expect(metrics.totalVotes).toBe(0);
    });

    it('should track total tokens used', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500);
      supervisorAnalytics.recordVote('gpt-4', 'debate-2', 'reject', 0.8, 1200, 300);
      
      const metrics = supervisorAnalytics.getSupervisorMetrics('gpt-4');
      expect(metrics.totalTokensUsed).toBe(800);
    });
  });

  describe('compareSupervisors', () => {
    it('should rank supervisors', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1000, 500, 'approve');
      supervisorAnalytics.recordVote('gpt-4', 'debate-2', 'approve', 0.9, 1000, 500, 'approve');
      supervisorAnalytics.recordVote('claude', 'debate-1', 'reject', 0.5, 3000, 400, 'approve');
      supervisorAnalytics.recordVote('claude', 'debate-2', 'reject', 0.5, 3000, 400, 'approve');
      
      const comparisons = supervisorAnalytics.compareSupervisors();
      expect(comparisons[0].supervisor).toBe('gpt-4');
      expect(comparisons[0].rank.byConsensusAgreement).toBe(1);
    });

    it('should return empty array when no supervisors', () => {
      const comparisons = supervisorAnalytics.compareSupervisors();
      expect(comparisons.length).toBe(0);
    });
  });

  describe('generateInsights', () => {
    it('should generate warning for low consensus agreement', () => {
      for (let i = 0; i < 12; i++) {
        supervisorAnalytics.recordVote('gpt-4', `debate-${i}`, 'reject', 0.8, 1500, 500, 'approve');
      }
      
      const insights = supervisorAnalytics.generateInsights();
      const consensusWarning = insights.find(
        (i) => i.category === 'consensus' && i.supervisor === 'gpt-4'
      );
      expect(consensusWarning).toBeDefined();
      expect(consensusWarning?.type).toBe('warning');
    });

    it('should generate warning for slow response times', () => {
      for (let i = 0; i < 6; i++) {
        supervisorAnalytics.recordVote('gpt-4', `debate-${i}`, 'approve', 0.8, 35000, 500);
      }
      
      const insights = supervisorAnalytics.generateInsights();
      const perfWarning = insights.find(
        (i) => i.category === 'performance' && i.type === 'warning'
      );
      expect(perfWarning).toBeDefined();
    });

    it('should generate recommendation for top performers', () => {
      for (let i = 0; i < 25; i++) {
        supervisorAnalytics.recordVote('gpt-4', `debate-${i}`, 'approve', 0.95, 1500, 500, 'approve');
      }
      
      const insights = supervisorAnalytics.generateInsights();
      const recommendation = insights.find(
        (i) => i.type === 'recommendation' && i.supervisor === 'gpt-4'
      );
      expect(recommendation).toBeDefined();
    });

    it('should warn about high deadlock rate', () => {
      for (let i = 0; i < 15; i++) {
        const outcome = i < 5 ? 'deadlock' : 'approved';
        supervisorAnalytics.recordDebateOutcome(
          `debate-${i}`,
          'Topic',
          'majority',
          outcome as 'deadlock' | 'approved',
          ['gpt-4'],
          2,
          3000
        );
      }
      
      const insights = supervisorAnalytics.generateInsights();
      const deadlockWarning = insights.find(
        (i) => i.message.includes('deadlock')
      );
      expect(deadlockWarning).toBeDefined();
    });
  });

  describe('getTrends', () => {
    it('should return trend data', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500, 'approve');
      supervisorAnalytics.recordDebateOutcome(
        'debate-1',
        'Topic',
        'majority',
        'approved',
        ['gpt-4'],
        2,
        3000
      );
      
      const trends = supervisorAnalytics.getTrends(7, 7);
      expect(trends.length).toBe(7);
      expect(trends[6].debateCount).toBeGreaterThanOrEqual(0);
    });
  });

  describe('getVotingPatterns', () => {
    it('should return vote distribution', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500);
      supervisorAnalytics.recordVote('gpt-4', 'debate-2', 'approve', 0.8, 1200, 400);
      supervisorAnalytics.recordVote('gpt-4', 'debate-3', 'reject', 0.7, 1000, 300);
      
      const patterns = supervisorAnalytics.getVotingPatterns('gpt-4');
      expect(patterns.voteDistribution.approve).toBe(2);
      expect(patterns.voteDistribution.reject).toBe(1);
    });

    it('should return confidence distribution', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.2, 1500, 500);
      supervisorAnalytics.recordVote('gpt-4', 'debate-2', 'approve', 0.5, 1200, 400);
      supervisorAnalytics.recordVote('gpt-4', 'debate-3', 'approve', 0.9, 1000, 300);
      
      const patterns = supervisorAnalytics.getVotingPatterns('gpt-4');
      expect(patterns.confidenceDistribution.low).toBe(1);
      expect(patterns.confidenceDistribution.medium).toBe(1);
      expect(patterns.confidenceDistribution.high).toBe(1);
    });

    it('should return recent votes', () => {
      for (let i = 0; i < 25; i++) {
        supervisorAnalytics.recordVote('gpt-4', `debate-${i}`, 'approve', 0.9, 1500, 500);
      }
      
      const patterns = supervisorAnalytics.getVotingPatterns('gpt-4');
      expect(patterns.recentVotes.length).toBe(20);
    });
  });

  describe('getDebateStatistics', () => {
    it('should calculate average rounds and duration', () => {
      supervisorAnalytics.recordDebateOutcome('debate-1', 'T1', 'majority', 'approved', ['a'], 2, 2000);
      supervisorAnalytics.recordDebateOutcome('debate-2', 'T2', 'unanimous', 'rejected', ['a'], 4, 4000);
      
      const stats = supervisorAnalytics.getDebateStatistics();
      expect(stats.avgRounds).toBe(3);
      expect(stats.avgDurationMs).toBe(3000);
    });

    it('should track by consensus mode', () => {
      supervisorAnalytics.recordDebateOutcome('debate-1', 'T1', 'majority', 'approved', ['a'], 2, 2000);
      supervisorAnalytics.recordDebateOutcome('debate-2', 'T2', 'majority', 'approved', ['a'], 2, 2000);
      supervisorAnalytics.recordDebateOutcome('debate-3', 'T3', 'unanimous', 'rejected', ['a'], 3, 3000);
      
      const stats = supervisorAnalytics.getDebateStatistics();
      expect(stats.byConsensusMode.majority).toBe(2);
      expect(stats.byConsensusMode.unanimous).toBe(1);
    });
  });

  describe('getSummary', () => {
    it('should return summary statistics', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500, 'approve');
      supervisorAnalytics.recordVote('claude', 'debate-1', 'approve', 0.8, 1200, 400, 'approve');
      supervisorAnalytics.recordDebateOutcome('debate-1', 'T1', 'majority', 'approved', ['gpt-4', 'claude'], 2, 3000);
      
      const summary = supervisorAnalytics.getSummary();
      expect(summary.totalSupervisors).toBe(2);
      expect(summary.totalDebates).toBe(1);
      expect(summary.totalVotes).toBe(2);
    });
  });

  describe('clearHistory', () => {
    it('should clear all data', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500);
      supervisorAnalytics.recordDebateOutcome('debate-1', 'T1', 'majority', 'approved', ['gpt-4'], 2, 3000);
      
      supervisorAnalytics.clearHistory();
      
      const summary = supervisorAnalytics.getSummary();
      expect(summary.totalSupervisors).toBe(0);
      expect(summary.totalDebates).toBe(0);
    });
  });

  describe('streak calculation', () => {
    it('should track agreement streaks', () => {
      supervisorAnalytics.recordVote('gpt-4', 'debate-1', 'approve', 0.9, 1500, 500, 'approve');
      supervisorAnalytics.recordVote('gpt-4', 'debate-2', 'approve', 0.9, 1500, 500, 'approve');
      supervisorAnalytics.recordVote('gpt-4', 'debate-3', 'approve', 0.9, 1500, 500, 'approve');
      
      const metrics = supervisorAnalytics.getSupervisorMetrics('gpt-4');
      expect(metrics.streakData.currentStreak).toBe(3);
      expect(metrics.streakData.streakType).toBe('agree');
    });
  });
});
