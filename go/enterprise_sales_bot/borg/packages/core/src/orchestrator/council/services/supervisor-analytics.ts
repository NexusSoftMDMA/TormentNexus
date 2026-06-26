import { EventEmitter } from 'events';

interface SupervisorVote {
  debateId: string;
  timestamp: number;
  vote: 'approve' | 'reject' | 'abstain';
  confidence: number;
  responseTimeMs: number;
  tokensUsed: number;
  agreedWithConsensus: boolean;
}

interface SupervisorMetrics {
  totalDebates: number;
  totalVotes: number;
  approvalRate: number;
  rejectionRate: number;
  abstentionRate: number;
  avgConfidence: number;
  avgResponseTimeMs: number;
  totalTokensUsed: number;
  consensusAgreementRate: number;
  lastActiveTimestamp: number;
  streakData: {
    currentStreak: number;
    longestStreak: number;
    streakType: 'agree' | 'disagree' | null;
  };
}

interface DebateOutcome {
  debateId: string;
  timestamp: number;
  topic: string;
  consensusMode: string;
  outcome: 'approved' | 'rejected' | 'deadlock';
  participatingSupervisors: string[];
  roundCount: number;
  durationMs: number;
}

interface AnalyticsTimeRange {
  start: number;
  end: number;
}

interface SupervisorComparison {
  supervisor: string;
  metrics: SupervisorMetrics;
  rank: {
    byConsensusAgreement: number;
    byResponseTime: number;
    byConfidence: number;
    overall: number;
  };
}

interface InsightType {
  type: 'warning' | 'info' | 'recommendation';
  category: 'performance' | 'consensus' | 'reliability' | 'cost';
  message: string;
  supervisor?: string;
  data?: Record<string, unknown>;
}

interface TrendData {
  period: string;
  debateCount: number;
  avgConsensusRate: number;
  avgResponseTime: number;
  topPerformer: string;
}

class SupervisorAnalyticsService extends EventEmitter {
  private supervisorVotes: Map<string, SupervisorVote[]> = new Map();
  private debateOutcomes: DebateOutcome[] = [];
  private maxHistorySize = 10000;

  recordVote(
    supervisor: string,
    debateId: string,
    vote: 'approve' | 'reject' | 'abstain',
    confidence: number,
    responseTimeMs: number,
    tokensUsed: number,
    consensusVote?: 'approve' | 'reject'
  ): void {
    if (!this.supervisorVotes.has(supervisor)) {
      this.supervisorVotes.set(supervisor, []);
    }

    const votes = this.supervisorVotes.get(supervisor)!;
    const agreedWithConsensus = consensusVote ? vote === consensusVote : false;

    votes.push({
      debateId,
      timestamp: Date.now(),
      vote,
      confidence,
      responseTimeMs,
      tokensUsed,
      agreedWithConsensus,
    });

    if (votes.length > this.maxHistorySize) {
      votes.shift();
    }

    this.emit('voteRecorded', { supervisor, debateId, vote, agreedWithConsensus });
  }

  recordDebateOutcome(
    debateId: string,
    topic: string,
    consensusMode: string,
    outcome: 'approved' | 'rejected' | 'deadlock',
    participatingSupervisors: string[],
    roundCount: number,
    durationMs: number
  ): void {
    this.debateOutcomes.push({
      debateId,
      timestamp: Date.now(),
      topic,
      consensusMode,
      outcome,
      participatingSupervisors,
      roundCount,
      durationMs,
    });

    if (this.debateOutcomes.length > this.maxHistorySize) {
      this.debateOutcomes.shift();
    }

    this.emit('debateOutcomeRecorded', { debateId, outcome });
  }

  getSupervisorMetrics(supervisor: string, timeRange?: AnalyticsTimeRange): SupervisorMetrics {
    const votes = this.getFilteredVotes(supervisor, timeRange);

    if (votes.length === 0) {
      return this.getEmptyMetrics();
    }

    const approveCount = votes.filter((v) => v.vote === 'approve').length;
    const rejectCount = votes.filter((v) => v.vote === 'reject').length;
    const abstainCount = votes.filter((v) => v.vote === 'abstain').length;
    const agreedCount = votes.filter((v) => v.agreedWithConsensus).length;

    const streakData = this.calculateStreak(votes);

    return {
      totalDebates: new Set(votes.map((v) => v.debateId)).size,
      totalVotes: votes.length,
      approvalRate: approveCount / votes.length,
      rejectionRate: rejectCount / votes.length,
      abstentionRate: abstainCount / votes.length,
      avgConfidence: votes.reduce((sum, v) => sum + v.confidence, 0) / votes.length,
      avgResponseTimeMs: votes.reduce((sum, v) => sum + v.responseTimeMs, 0) / votes.length,
      totalTokensUsed: votes.reduce((sum, v) => sum + v.tokensUsed, 0),
      consensusAgreementRate: agreedCount / votes.length,
      lastActiveTimestamp: Math.max(...votes.map((v) => v.timestamp)),
      streakData,
    };
  }

  getAllSupervisorMetrics(timeRange?: AnalyticsTimeRange): Map<string, SupervisorMetrics> {
    const result = new Map<string, SupervisorMetrics>();
    for (const supervisor of this.supervisorVotes.keys()) {
      result.set(supervisor, this.getSupervisorMetrics(supervisor, timeRange));
    }
    return result;
  }

  compareSupervisors(timeRange?: AnalyticsTimeRange): SupervisorComparison[] {
    const allMetrics = this.getAllSupervisorMetrics(timeRange);
    const comparisons: SupervisorComparison[] = [];

    for (const [supervisor, metrics] of allMetrics) {
      comparisons.push({
        supervisor,
        metrics,
        rank: { byConsensusAgreement: 0, byResponseTime: 0, byConfidence: 0, overall: 0 },
      });
    }

    const byConsensus = [...comparisons].sort(
      (a, b) => b.metrics.consensusAgreementRate - a.metrics.consensusAgreementRate
    );
    const byResponseTime = [...comparisons].sort(
      (a, b) => a.metrics.avgResponseTimeMs - b.metrics.avgResponseTimeMs
    );
    const byConfidence = [...comparisons].sort(
      (a, b) => b.metrics.avgConfidence - a.metrics.avgConfidence
    );

    byConsensus.forEach((c, i) => (c.rank.byConsensusAgreement = i + 1));
    byResponseTime.forEach((c, i) => (c.rank.byResponseTime = i + 1));
    byConfidence.forEach((c, i) => (c.rank.byConfidence = i + 1));

    comparisons.forEach((c) => {
      c.rank.overall = Math.round(
        (c.rank.byConsensusAgreement + c.rank.byResponseTime + c.rank.byConfidence) / 3
      );
    });

    return comparisons.sort((a, b) => a.rank.overall - b.rank.overall);
  }

  generateInsights(timeRange?: AnalyticsTimeRange): InsightType[] {
    const insights: InsightType[] = [];
    const allMetrics = this.getAllSupervisorMetrics(timeRange);

    for (const [supervisor, metrics] of allMetrics) {
      if (metrics.consensusAgreementRate < 0.5 && metrics.totalVotes >= 10) {
        insights.push({
          type: 'warning',
          category: 'consensus',
          message: `${supervisor} frequently disagrees with consensus (${(metrics.consensusAgreementRate * 100).toFixed(1)}% agreement)`,
          supervisor,
          data: { agreementRate: metrics.consensusAgreementRate },
        });
      }

      if (metrics.avgResponseTimeMs > 30000 && metrics.totalVotes >= 5) {
        insights.push({
          type: 'warning',
          category: 'performance',
          message: `${supervisor} has slow response times (avg ${(metrics.avgResponseTimeMs / 1000).toFixed(1)}s)`,
          supervisor,
          data: { avgResponseTimeMs: metrics.avgResponseTimeMs },
        });
      }

      if (metrics.abstentionRate > 0.3 && metrics.totalVotes >= 10) {
        insights.push({
          type: 'info',
          category: 'reliability',
          message: `${supervisor} abstains frequently (${(metrics.abstentionRate * 100).toFixed(1)}%)`,
          supervisor,
          data: { abstentionRate: metrics.abstentionRate },
        });
      }

      if (metrics.avgConfidence > 0.9 && metrics.consensusAgreementRate > 0.8 && metrics.totalVotes >= 20) {
        insights.push({
          type: 'recommendation',
          category: 'performance',
          message: `${supervisor} is a top performer - consider prioritizing in team selection`,
          supervisor,
          data: { confidence: metrics.avgConfidence, agreementRate: metrics.consensusAgreementRate },
        });
      }
    }

    const recentDebates = this.getFilteredDebates(timeRange);
    const deadlockRate = recentDebates.filter((d) => d.outcome === 'deadlock').length / (recentDebates.length || 1);
    if (deadlockRate > 0.2 && recentDebates.length >= 10) {
      insights.push({
        type: 'warning',
        category: 'consensus',
        message: `High deadlock rate (${(deadlockRate * 100).toFixed(1)}%) - consider adjusting consensus mode or team composition`,
        data: { deadlockRate, debateCount: recentDebates.length },
      });
    }

    return insights;
  }

  getTrends(periodDays: number = 7, bucketCount: number = 7): TrendData[] {
    const now = Date.now();
    const periodMs = periodDays * 24 * 60 * 60 * 1000;
    const bucketSize = periodMs / bucketCount;
    const trends: TrendData[] = [];

    for (let i = 0; i < bucketCount; i++) {
      const bucketStart = now - periodMs + i * bucketSize;
      const bucketEnd = bucketStart + bucketSize;
      const timeRange = { start: bucketStart, end: bucketEnd };

      const debates = this.getFilteredDebates(timeRange);
      const allMetrics = this.getAllSupervisorMetrics(timeRange);

      let totalConsensusRate = 0;
      let totalResponseTime = 0;
      let supervisorCount = 0;
      let topPerformer = '';
      let topScore = 0;

      for (const [supervisor, metrics] of allMetrics) {
        if (metrics.totalVotes > 0) {
          totalConsensusRate += metrics.consensusAgreementRate;
          totalResponseTime += metrics.avgResponseTimeMs;
          supervisorCount++;

          const score = metrics.consensusAgreementRate * 0.6 + (1 - metrics.avgResponseTimeMs / 60000) * 0.4;
          if (score > topScore) {
            topScore = score;
            topPerformer = supervisor;
          }
        }
      }

      trends.push({
        period: new Date(bucketStart).toISOString().split('T')[0],
        debateCount: debates.length,
        avgConsensusRate: supervisorCount > 0 ? totalConsensusRate / supervisorCount : 0,
        avgResponseTime: supervisorCount > 0 ? totalResponseTime / supervisorCount : 0,
        topPerformer,
      });
    }

    return trends;
  }

  getDebateStatistics(timeRange?: AnalyticsTimeRange): {
    total: number;
    byOutcome: Record<string, number>;
    byConsensusMode: Record<string, number>;
    avgRounds: number;
    avgDurationMs: number;
    participationBySupervisor: Record<string, number>;
  } {
    const debates = this.getFilteredDebates(timeRange);

    const byOutcome: Record<string, number> = { approved: 0, rejected: 0, deadlock: 0 };
    const byConsensusMode: Record<string, number> = {};
    const participationBySupervisor: Record<string, number> = {};

    let totalRounds = 0;
    let totalDuration = 0;

    for (const debate of debates) {
      byOutcome[debate.outcome] = (byOutcome[debate.outcome] || 0) + 1;
      byConsensusMode[debate.consensusMode] = (byConsensusMode[debate.consensusMode] || 0) + 1;
      totalRounds += debate.roundCount;
      totalDuration += debate.durationMs;

      for (const supervisor of debate.participatingSupervisors) {
        participationBySupervisor[supervisor] = (participationBySupervisor[supervisor] || 0) + 1;
      }
    }

    return {
      total: debates.length,
      byOutcome,
      byConsensusMode,
      avgRounds: debates.length > 0 ? totalRounds / debates.length : 0,
      avgDurationMs: debates.length > 0 ? totalDuration / debates.length : 0,
      participationBySupervisor,
    };
  }

  getVotingPatterns(supervisor: string, timeRange?: AnalyticsTimeRange): {
    voteDistribution: Record<string, number>;
    confidenceDistribution: { low: number; medium: number; high: number };
    hourlyActivity: Record<number, number>;
    recentVotes: Array<{ debateId: string; vote: string; timestamp: number }>;
  } {
    const votes = this.getFilteredVotes(supervisor, timeRange);

    const voteDistribution: Record<string, number> = { approve: 0, reject: 0, abstain: 0 };
    const confidenceDistribution = { low: 0, medium: 0, high: 0 };
    const hourlyActivity: Record<number, number> = {};

    for (const vote of votes) {
      voteDistribution[vote.vote]++;

      if (vote.confidence < 0.4) confidenceDistribution.low++;
      else if (vote.confidence < 0.7) confidenceDistribution.medium++;
      else confidenceDistribution.high++;

      const hour = new Date(vote.timestamp).getHours();
      hourlyActivity[hour] = (hourlyActivity[hour] || 0) + 1;
    }

    const recentVotes = votes
      .slice(-20)
      .reverse()
      .map((v) => ({ debateId: v.debateId, vote: v.vote, timestamp: v.timestamp }));

    return { voteDistribution, confidenceDistribution, hourlyActivity, recentVotes };
  }

  getSummary(): {
    totalSupervisors: number;
    totalDebates: number;
    totalVotes: number;
    activeSupervisors: string[];
    topPerformers: string[];
    insightCount: number;
  } {
    const last24h = { start: Date.now() - 24 * 60 * 60 * 1000, end: Date.now() };
    const comparisons = this.compareSupervisors(last24h);
    const insights = this.generateInsights(last24h);

    const activeSupervisors = comparisons
      .filter((c) => c.metrics.totalVotes > 0)
      .map((c) => c.supervisor);

    const topPerformers = comparisons
      .filter((c) => c.rank.overall <= 3)
      .map((c) => c.supervisor);

    let totalVotes = 0;
    for (const votes of this.supervisorVotes.values()) {
      totalVotes += votes.length;
    }

    return {
      totalSupervisors: this.supervisorVotes.size,
      totalDebates: this.debateOutcomes.length,
      totalVotes,
      activeSupervisors,
      topPerformers,
      insightCount: insights.length,
    };
  }

  clearHistory(): void {
    this.supervisorVotes.clear();
    this.debateOutcomes = [];
    this.emit('historyCleared', {});
  }

  private getFilteredVotes(supervisor: string, timeRange?: AnalyticsTimeRange): SupervisorVote[] {
    const votes = this.supervisorVotes.get(supervisor) || [];
    if (!timeRange) return votes;
    return votes.filter((v) => v.timestamp >= timeRange.start && v.timestamp <= timeRange.end);
  }

  private getFilteredDebates(timeRange?: AnalyticsTimeRange): DebateOutcome[] {
    if (!timeRange) return this.debateOutcomes;
    return this.debateOutcomes.filter(
      (d) => d.timestamp >= timeRange.start && d.timestamp <= timeRange.end
    );
  }

  private calculateStreak(votes: SupervisorVote[]): SupervisorMetrics['streakData'] {
    if (votes.length === 0) {
      return { currentStreak: 0, longestStreak: 0, streakType: null };
    }

    let currentStreak = 0;
    let longestStreak = 0;
    let streakType: 'agree' | 'disagree' | null = null;
    let currentType: 'agree' | 'disagree' | null = null;

    const sortedVotes = [...votes].sort((a, b) => b.timestamp - a.timestamp);

    for (const vote of sortedVotes) {
      const type = vote.agreedWithConsensus ? 'agree' : 'disagree';

      if (currentType === null) {
        currentType = type;
        streakType = type;
        currentStreak = 1;
      } else if (type === currentType) {
        currentStreak++;
      } else {
        if (currentStreak > longestStreak) {
          longestStreak = currentStreak;
        }
        currentStreak = 1;
        currentType = type;
      }
    }

    if (currentStreak > longestStreak) {
      longestStreak = currentStreak;
    }

    return {
      currentStreak,
      longestStreak,
      streakType,
    };
  }

  private getEmptyMetrics(): SupervisorMetrics {
    return {
      totalDebates: 0,
      totalVotes: 0,
      approvalRate: 0,
      rejectionRate: 0,
      abstentionRate: 0,
      avgConfidence: 0,
      avgResponseTimeMs: 0,
      totalTokensUsed: 0,
      consensusAgreementRate: 0,
      lastActiveTimestamp: 0,
      streakData: { currentStreak: 0, longestStreak: 0, streakType: null },
    };
  }
}

export const supervisorAnalytics = new SupervisorAnalyticsService();
