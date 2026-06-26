import { EventEmitter } from 'events';
import { dbService } from './db.js';
import type { CouncilDecision, DevelopmentTask, Vote, ConsensusMode, TaskType } from './types.js';
import { eq, lt, and, gte, lte, desc, asc, sql } from 'drizzle-orm';
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from '../../../db/sqliteAvailability.js';

/**
 * A complete record of a council debate
 */
export interface DebateRecord {
  id: string;
  timestamp: number;
  task: DevelopmentTask;
  decision: CouncilDecision;
  metadata: DebateMetadata;
}

/**
 * Metadata about the debate context
 */
export interface DebateMetadata {
  sessionId?: string;
  debateRounds: number;
  consensusMode: ConsensusMode;
  leadSupervisor?: string;
  dynamicSelection?: {
    enabled: boolean;
    taskType?: TaskType;
    confidence?: number;
  };
  durationMs: number;
  supervisorCount: number;
  participatingSupervisors: string[];
}

/**
 * Query options for searching debate history
 */
export interface DebateQueryOptions {
  sessionId?: string;
  taskType?: TaskType;
  approved?: boolean;
  supervisorName?: string;
  fromTimestamp?: number;
  toTimestamp?: number;
  minConsensus?: number;
  maxConsensus?: number;
  limit?: number;
  offset?: number;
  sortBy?: 'timestamp' | 'consensus' | 'duration';
  sortOrder?: 'asc' | 'desc';
}

/**
 * Statistics about debate history
 */
export interface DebateStats {
  totalDebates: number;
  approvedCount: number;
  rejectedCount: number;
  approvalRate: number;
  averageConsensus: number;
  averageDurationMs: number;
  debatesByTaskType: Record<string, number>;
  debatesBySupervisor: Record<string, number>;
  debatesByConsensusMode: Record<string, number>;
  oldestDebate?: number;
  newestDebate?: number;
}

/**
 * Configuration for debate history persistence
 */
export interface DebateHistoryConfig {
  enabled: boolean;
  storageDir: string;
  maxRecords: number;
  autoSave: boolean;
  retentionDays: number;
}

/**
 * Service for persisting and querying council debate history
 */
export class DebateHistoryService extends EventEmitter {
  private config: DebateHistoryConfig = {
    enabled: true,
    storageDir: './data/debate-history',
    maxRecords: 1000,
    autoSave: true,
    retentionDays: 90,
  };

  private initialized = false;
  private storageAvailable = true;

  constructor() {
    super();
  }

  /**
   * Initialize the service and load existing records
   */
  async initialize(): Promise<void> {
    if (this.initialized) return;
    this.initialized = true;

    try {
      const recordCount = await this.getRecordCount();
      this.storageAvailable = true;
      this.emit('initialized', { recordCount });
    } catch (error) {
      this.storageAvailable = false;
      this.config.enabled = false;
      this.emit('storage_error', { action: 'initialize', error });
      const message = isSqliteUnavailableError(error)
        ? '[DebateHistory] Disabling debate history persistence for this run because SQLite is unavailable:'
        : '[DebateHistory] Disabling debate history persistence for this run due to initialization failure:';
      console.warn(message, formatOptionalSqliteFailure('Debate history initialization failed', error));
      this.emit('initialized', { recordCount: 0, storageAvailable: false });
    }
  }

  /**
   * Generate a unique ID for a debate record
   */
  private generateId(): string {
    const timestamp = Date.now().toString(36);
    const random = Math.random().toString(36).substring(2, 8);
    return `debate_${timestamp}_${random}`;
  }

  /**
   * Save a debate record
   */
  async saveDebate(
    task: DevelopmentTask,
    decision: CouncilDecision,
    metadata: Partial<DebateMetadata>
  ): Promise<DebateRecord> {
    const id = this.generateId();
    const timestamp = Date.now();
    
    const record: DebateRecord = {
      id,
      timestamp,
      task,
      decision,
      metadata: {
        sessionId: metadata.sessionId,
        debateRounds: metadata.debateRounds ?? 2,
        consensusMode: metadata.consensusMode ?? 'weighted',
        leadSupervisor: metadata.leadSupervisor,
        dynamicSelection: metadata.dynamicSelection,
        durationMs: metadata.durationMs ?? 0,
        supervisorCount: decision.votes.length,
        participatingSupervisors: decision.votes.map(v => v.supervisor),
      },
    };

    if (this.config.enabled && this.storageAvailable) {
      await this.persistRecord(record);
      await this.pruneOldRecords();
    }

    this.emit('debate_saved', record);
    return record;
  }

  private async pruneOldRecords(): Promise<void> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const cutoffTime = new Date(Date.now() - (this.config.retentionDays * 24 * 60 * 60 * 1000));

    // Delete old records
    await db.delete(schema.councilDebatesTable).where(lt(schema.councilDebatesTable.timestamp, cutoffTime));

    // Delete excess records
    const count = await this.getRecordCount();
    if (count > this.config.maxRecords) {
        const excess = count - this.config.maxRecords;
        const oldRecords = await db.select({ id: schema.councilDebatesTable.id })
            .from(schema.councilDebatesTable)
            .orderBy(asc(schema.councilDebatesTable.timestamp))
            .limit(excess);

        for (const row of oldRecords) {
            await this.deleteRecord(row.id);
        }

        if (oldRecords.length > 0) {
             this.emit('pruned', { count: oldRecords.length });
        }
    }
  }

  /**
   * Persist a single record to SQLite
   */
  private async persistRecord(record: DebateRecord): Promise<void> {
    try {
      const db = dbService.getDrizzle();
      const schema = dbService.getSchema();
      
      await db.insert(schema.councilDebatesTable).values({
        id: record.id,
        title: record.task.description.substring(0, 255),
        sessionId: record.metadata.sessionId || null,
        taskType: record.metadata.dynamicSelection?.taskType || 'general',
        status: 'completed',
        consensus: record.decision.consensus,
        weightedConsensus: record.decision.weightedConsensus || record.decision.consensus,
        outcome: record.decision.approved ? 'approved' : 'rejected',
        rounds: record.metadata.debateRounds,
        timestamp: new Date(record.timestamp),
        data: record
      });
    } catch (error) {
      this.emit('storage_error', { action: 'persist', recordId: record.id, error });
    }
  }

  /**
   * Delete a record from SQLite
   */
  async deleteRecord(id: string): Promise<boolean> {
    if (!this.config.enabled || !this.storageAvailable) return false;
    
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const result = await db.delete(schema.councilDebatesTable).where(eq(schema.councilDebatesTable.id, id));

    const deleted = (result as any).changes > 0;
    if (deleted) {
      this.emit('debate_deleted', { id });
    }

    return deleted;
  }

  /**
   * Get a single debate record by ID
   */
  async getDebate(id: string): Promise<DebateRecord | undefined> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilDebatesTable).where(eq(schema.councilDebatesTable.id, id)).limit(1);

    if (rows.length > 0) {
      return rows[0].data as DebateRecord;
    }
    return undefined;
  }

  /**
   * Query debate records with filters
   */
  async queryDebates(options: DebateQueryOptions = {}): Promise<DebateRecord[]> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    
    let where: any = undefined;
    const conditions = [];

    if (options.sessionId) {
      conditions.push(eq(schema.councilDebatesTable.sessionId, options.sessionId));
    }

    if (options.taskType) {
      conditions.push(eq(schema.councilDebatesTable.taskType, options.taskType));
    }

    if (options.approved !== undefined) {
      conditions.push(eq(schema.councilDebatesTable.outcome, options.approved ? 'approved' : 'rejected'));
    }

    if (options.fromTimestamp) {
      conditions.push(gte(schema.councilDebatesTable.timestamp, new Date(options.fromTimestamp)));
    }

    if (options.toTimestamp) {
      conditions.push(lte(schema.councilDebatesTable.timestamp, new Date(options.toTimestamp)));
    }

    if (options.minConsensus !== undefined) {
      conditions.push(gte(schema.councilDebatesTable.consensus, options.minConsensus));
    }

    if (options.maxConsensus !== undefined) {
      conditions.push(lte(schema.councilDebatesTable.consensus, options.maxConsensus));
    }

    if (conditions.length > 0) {
      where = and(...conditions);
    }

    const limit = options.limit ?? 50;
    const offset = options.offset ?? 0;
    
    let query = db.select().from(schema.councilDebatesTable);
    if (where) query = query.where(where) as any;

    const sortBy = options.sortBy ?? 'timestamp';
    const sortOrder = options.sortOrder ?? 'desc';
    
    const sortCol = sortBy === 'consensus' ? schema.councilDebatesTable.consensus : 
                   sortBy === 'duration' ? schema.councilDebatesTable.id : // duration not in schema, fallback to ID or similar
                   schema.councilDebatesTable.timestamp;

    query = query.orderBy(sortOrder === 'desc' ? desc(sortCol) : asc(sortCol))
                 .limit(limit)
                 .offset(offset) as any;

    const rows = await query;
    let results = rows.map(r => r.data as DebateRecord);

    // Post-filter for supervisor name as it's inside the JSON/metadata
    if (options.supervisorName) {
      results = results.filter(r =>
        r.metadata.participatingSupervisors.includes(options.supervisorName!)
      );
    }

    return results;
  }

  /**
   * Get statistics about debate history
   */
  async getStats(): Promise<DebateStats> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    
    const totalCountResult = await db.select({ count: sql<number>`count(*)` }).from(schema.councilDebatesTable);
    const totalDebates = totalCountResult[0]?.count || 0;
    
    const approvedCountResult = await db.select({ count: sql<number>`count(*)` })
        .from(schema.councilDebatesTable)
        .where(eq(schema.councilDebatesTable.outcome, 'approved'));
    const approvedCount = approvedCountResult[0]?.count || 0;
    
    const rejectedCount = totalDebates - approvedCount;

    const avgConsensusResult = await db.select({ avg: sql<number>`avg(consensus)` }).from(schema.councilDebatesTable);
    const avgConsensus = avgConsensusResult[0]?.avg || 0;

    const taskTypeStats = await db.select({ 
        taskType: schema.councilDebatesTable.taskType, 
        count: sql<number>`count(*)` 
    }).from(schema.councilDebatesTable).groupBy(schema.councilDebatesTable.taskType);
    
    const debatesByTaskType: Record<string, number> = {};
    taskTypeStats.forEach(r => debatesByTaskType[r.taskType] = r.count);

    // Load full data for complex aggregations (safe for MVP scale)
    const allDebates = await this.queryDebates({ limit: 10000 });

    const debatesBySupervisor: Record<string, number> = {};
    const debatesByConsensusMode: Record<string, number> = {};
    let totalDuration = 0;

    for (const record of allDebates) {
        totalDuration += record.metadata.durationMs;

        for (const supervisor of record.metadata.participatingSupervisors) {
            debatesBySupervisor[supervisor] = (debatesBySupervisor[supervisor] ?? 0) + 1;
        }

        const mode = record.metadata.consensusMode;
        debatesByConsensusMode[mode] = (debatesByConsensusMode[mode] ?? 0) + 1;
    }

    const timestampResult = await db.select({ 
        min: sql<number>`min(timestamp)`, 
        max: sql<number>`max(timestamp)` 
    }).from(schema.councilDebatesTable);

    return {
      totalDebates,
      approvedCount,
      rejectedCount,
      approvalRate: totalDebates > 0 ? approvedCount / totalDebates : 0,
      averageConsensus: avgConsensus,
      averageDurationMs: totalDebates > 0 ? totalDuration / totalDebates : 0,
      debatesByTaskType,
      debatesBySupervisor,
      debatesByConsensusMode,
      oldestDebate: timestampResult[0]?.min,
      newestDebate: timestampResult[0]?.max,
    };
  }

  /**
   * Get vote patterns for a specific supervisor
   */
  async getSupervisorVoteHistory(supervisorName: string): Promise<{
    totalVotes: number;
    approvals: number;
    rejections: number;
    averageConfidence: number;
    recentVotes: Array<{ debateId: string; approved: boolean; confidence: number; timestamp: number }>;
  }> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilDebatesTable).orderBy(desc(schema.councilDebatesTable.timestamp)).limit(1000);

    const votes: Array<{ debateId: string; vote: Vote; timestamp: number }> = [];

    for (const row of rows) {
      const record = row.data as DebateRecord;
      const vote = record.decision.votes.find(v => v.supervisor === supervisorName);
      if (vote) {
        votes.push({ debateId: record.id, vote, timestamp: record.timestamp });
      }
    }

    if (votes.length === 0) {
      return {
        totalVotes: 0,
        approvals: 0,
        rejections: 0,
        averageConfidence: 0,
        recentVotes: [],
      };
    }

    const approvals = votes.filter(v => v.vote.approved).length;
    const totalConfidence = votes.reduce((sum, v) => sum + v.vote.confidence, 0);

    return {
      totalVotes: votes.length,
      approvals,
      rejections: votes.length - approvals,
      averageConfidence: totalConfidence / votes.length,
      recentVotes: votes.slice(0, 10).map(v => ({
        debateId: v.debateId,
        approved: v.vote.approved,
        confidence: v.vote.confidence,
        timestamp: v.timestamp,
      })),
    };
  }

  /**
   * Export debate history to JSON
   */
  async exportToJson(options: DebateQueryOptions = {}): Promise<string> {
    const records = await this.queryDebates({ ...options, limit: options.limit ?? 10000 });
    return JSON.stringify(records, null, 2);
  }

  /**
   * Export debate history to CSV
   */
  async exportToCsv(options: DebateQueryOptions = {}): Promise<string> {
    const records = await this.queryDebates({ ...options, limit: options.limit ?? 10000 });
    
    const headers = [
      'id',
      'timestamp',
      'task_id',
      'task_description',
      'approved',
      'consensus',
      'weighted_consensus',
      'consensus_mode',
      'supervisor_count',
      'participating_supervisors',
      'duration_ms',
      'session_id',
      'task_type',
    ];

    const rows = records.map(r => [
      r.id,
      new Date(r.timestamp).toISOString(),
      r.task.id,
      `"${r.task.description.replace(/"/g, '""')}"`,
      r.decision.approved,
      r.decision.consensus.toFixed(3),
      (r.decision.weightedConsensus ?? r.decision.consensus).toFixed(3),
      r.metadata.consensusMode,
      r.metadata.supervisorCount,
      `"${r.metadata.participatingSupervisors.join(', ')}"`,
      r.metadata.durationMs,
      r.metadata.sessionId ?? '',
      r.metadata.dynamicSelection?.taskType ?? 'general',
    ]);

    return [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
  }

  /**
   * Clear all debate history
   */
  async clearAll(): Promise<number> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const result = await db.delete(schema.councilDebatesTable);
    const count = (result as any).changes || 0;
    this.emit('cleared', { count });
    return count;
  }

  /**
   * Get current configuration
   */
  getConfig(): DebateHistoryConfig {
    return { ...this.config };
  }

  /**
   * Update configuration
   */
  updateConfig(updates: Partial<DebateHistoryConfig>): DebateHistoryConfig {
    this.config = { ...this.config, ...updates };
    this.emit('config_updated', this.config);
    return this.getConfig();
  }

  /**
   * Check if service is enabled
   */
  isEnabled(): boolean {
    return this.config.enabled && this.storageAvailable;
  }

  /**
   * Get total record count
   */
  async getRecordCount(): Promise<number> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const result = await db.select({ count: sql<number>`count(*)` }).from(schema.councilDebatesTable);
    return result[0]?.count || 0;
  }

  /**
   * Get storage size in bytes
   */
  getStorageSize(): number {
    return 0; // Not easily available in SQLite without fs stats which varies by mode
  }
}

export const debateHistory = new DebateHistoryService();
