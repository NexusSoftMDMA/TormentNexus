import { EventEmitter } from 'events';
import { dbService } from './db.js';
import type { CouncilConfig, ConsensusMode, DevelopmentTask, CouncilDecision } from './types.js';
import { eq, and, sql, desc } from 'drizzle-orm';

/**
 * Workspace configuration and state
 */
export interface Workspace {
  id: string;
  name: string;
  description?: string;
  path: string;
  createdAt: Date;
  updatedAt: Date;
  config: WorkspaceConfig;
  metadata: WorkspaceMetadata;
  status: WorkspaceStatus;
}

export interface WorkspaceConfig {
  defaultConsensusMode: ConsensusMode;
  defaultDebateRounds: number;
  consensusThreshold: number;
  supervisorTeam: string[];
  leadSupervisor?: string;
  autoSaveDebates: boolean;
  maxConcurrentDebates: number;
  budgetLimit?: number;
  tags: string[];
}

export interface WorkspaceMetadata {
  totalDebates: number;
  approvedDebates: number;
  rejectedDebates: number;
  totalTokensUsed: number;
  estimatedCost: number;
  lastDebateAt?: Date;
  averageDebateDuration: number;
}

export type WorkspaceStatus = 'active' | 'paused' | 'archived';

export interface WorkspaceDebate {
  workspaceId: string;
  debateId: string;
  task: DevelopmentTask;
  decision?: CouncilDecision;
  startedAt: Date;
  completedAt?: Date;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
}

export interface WorkspaceStats {
  workspaceId: string;
  period: { start: Date; end: Date };
  debates: {
    total: number;
    approved: number;
    rejected: number;
    pending: number;
  };
  performance: {
    avgDurationMs: number;
    avgConsensus: number;
    avgConfidence: number;
  };
  supervisors: {
    mostActive: string[];
    highestAgreement: string[];
  };
  tokens: {
    total: number;
    perDebate: number;
  };
  cost: {
    total: number;
    perDebate: number;
  };
}

export interface WorkspaceComparison {
  workspaces: string[];
  metrics: {
    workspaceId: string;
    name: string;
    debates: number;
    approvalRate: number;
    avgConsensus: number;
    avgDuration: number;
    totalCost: number;
  }[];
  ranking: {
    byApprovalRate: string[];
    byConsensus: string[];
    byEfficiency: string[];
  };
}

/**
 * WorkspaceManager - Manage multiple project workspaces with isolated configurations
 */
export class WorkspaceManagerService extends EventEmitter {
  private activeWorkspaceId: string | null = null;

  constructor() {
    super();
  }

  // ============ Workspace CRUD ============

  async createWorkspace(
    name: string,
    path: string,
    config?: Partial<WorkspaceConfig>,
    description?: string
  ): Promise<Workspace> {
    const id = this.generateId();
    const now = new Date();

    const workspace: Workspace = {
      id,
      name,
      description,
      path,
      createdAt: now,
      updatedAt: now,
      config: {
        defaultConsensusMode: config?.defaultConsensusMode ?? 'weighted',
        defaultDebateRounds: config?.defaultDebateRounds ?? 2,
        consensusThreshold: config?.consensusThreshold ?? 0.7,
        supervisorTeam: config?.supervisorTeam ?? [],
        leadSupervisor: config?.leadSupervisor,
        autoSaveDebates: config?.autoSaveDebates ?? true,
        maxConcurrentDebates: config?.maxConcurrentDebates ?? 3,
        budgetLimit: config?.budgetLimit,
        tags: config?.tags ?? [],
      },
      metadata: {
        totalDebates: 0,
        approvedDebates: 0,
        rejectedDebates: 0,
        totalTokensUsed: 0,
        estimatedCost: 0,
        averageDebateDuration: 0,
      },
      status: 'active',
    };

    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();

    await db.insert(schema.councilWorkspacesTable).values({
      id: workspace.id,
      name: workspace.name,
      path: workspace.path,
      status: workspace.status,
      config: { config: workspace.config, metadata: workspace.metadata },
      description: workspace.description || null,
      createdAt: workspace.createdAt,
      updatedAt: workspace.updatedAt
    });

    this.emit('workspace:created', workspace);
    return workspace;
  }

  async getWorkspace(id: string): Promise<Workspace | undefined> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilWorkspacesTable).where(eq(schema.councilWorkspacesTable.id, id)).limit(1);
    if (rows.length === 0) return undefined;
    return this.mapRowToWorkspace(rows[0]);
  }

  async getWorkspaceByPath(path: string): Promise<Workspace | undefined> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilWorkspacesTable).where(eq(schema.councilWorkspacesTable.path, path)).limit(1);
    if (rows.length === 0) return undefined;
    return this.mapRowToWorkspace(rows[0]);
  }

  async getAllWorkspaces(): Promise<Workspace[]> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilWorkspacesTable);
    return rows.map(r => this.mapRowToWorkspace(r));
  }

  async getWorkspacesByStatus(status: WorkspaceStatus): Promise<Workspace[]> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilWorkspacesTable).where(eq(schema.councilWorkspacesTable.status, status));
    return rows.map(r => this.mapRowToWorkspace(r));
  }

  async getWorkspacesByTag(tag: string): Promise<Workspace[]> {
    const all = await this.getAllWorkspaces();
    return all.filter(w => w.config.tags && w.config.tags.includes(tag));
  }

  private mapRowToWorkspace(row: any): Workspace {
    const data = row.config as any;
    return {
      id: row.id,
      name: row.name,
      path: row.path,
      status: row.status as WorkspaceStatus,
      description: row.description,
      createdAt: row.createdAt,
      updatedAt: row.updatedAt,
      config: data.config,
      metadata: data.metadata,
    };
  }

  async updateWorkspace(id: string, updates: Partial<Omit<Workspace, 'id' | 'createdAt'>>): Promise<Workspace | undefined> {
    const workspace = await this.getWorkspace(id);
    if (!workspace) return undefined;

    const updated: Workspace = {
      ...workspace,
      ...updates,
      id: workspace.id,
      createdAt: workspace.createdAt,
      updatedAt: new Date(),
      config: updates.config ? { ...workspace.config, ...updates.config } : workspace.config,
      metadata: updates.metadata ? { ...workspace.metadata, ...updates.metadata } : workspace.metadata,
    };

    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();

    await db.update(schema.councilWorkspacesTable)
      .set({
        name: updated.name,
        path: updated.path,
        status: updated.status,
        config: { config: updated.config, metadata: updated.metadata },
        description: updated.description || null,
        updatedAt: updated.updatedAt,
      })
      .where(eq(schema.councilWorkspacesTable.id, updated.id));

    this.emit('workspace:updated', updated);
    return updated;
  }

  async updateWorkspaceConfig(id: string, config: Partial<WorkspaceConfig>): Promise<Workspace | undefined> {
    const workspace = await this.getWorkspace(id);
    if (!workspace) return undefined;

    return this.updateWorkspace(id, {
      config: { ...workspace.config, ...config },
    });
  }

  async deleteWorkspace(id: string): Promise<boolean> {
    const workspace = await this.getWorkspace(id);
    if (!workspace) return false;

    if (this.activeWorkspaceId === id) {
      this.activeWorkspaceId = null;
    }

    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    await db.delete(schema.councilWorkspacesTable).where(eq(schema.councilWorkspacesTable.id, id));

    this.emit('workspace:deleted', { id, name: workspace.name });
    return true;
  }

  async archiveWorkspace(id: string): Promise<Workspace | undefined> {
    return this.updateWorkspace(id, { status: 'archived' });
  }

  // ============ Active Workspace ============

  async setActiveWorkspace(id: string): Promise<boolean> {
    const workspace = await this.getWorkspace(id);
    if (!workspace || workspace.status !== 'active') return false;

    this.activeWorkspaceId = id;
    this.emit('workspace:activated', workspace);
    return true;
  }

  async getActiveWorkspace(): Promise<Workspace | undefined> {
    if (!this.activeWorkspaceId) return undefined;
    return this.getWorkspace(this.activeWorkspaceId);
  }

  clearActiveWorkspace(): void {
    this.activeWorkspaceId = null;
    this.emit('workspace:deactivated');
  }

  // ============ Debate Tracking ============

  async startDebate(workspaceId: string, task: DevelopmentTask): Promise<WorkspaceDebate | undefined> {
    const workspace = await this.getWorkspace(workspaceId);
    if (!workspace) return undefined;

    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();

    // Check concurrent limit
    const activeCountResult = await db.select({ count: sql<number>`count(*)` })
        .from(schema.councilDebatesTable)
        .where(and(
            eq(schema.councilDebatesTable.workspaceId, workspaceId),
            eq(schema.councilDebatesTable.status, 'in_progress')
        ));
    const activeCount = activeCountResult[0]?.count || 0;

    if (activeCount >= workspace.config.maxConcurrentDebates) {
      this.emit('workspace:debate:limit_reached', { workspaceId, limit: workspace.config.maxConcurrentDebates });
      return undefined;
    }

    const debateId = this.generateId();
    const startedAt = new Date();

    const debate: WorkspaceDebate = {
      workspaceId,
      debateId,
      task,
      startedAt,
      status: 'in_progress',
    };

    // Insert into debates table
    await db.insert(schema.councilDebatesTable).values({
      id: debateId,
      title: task.description.substring(0, 255),
      workspaceId,
      taskType: 'general',
      status: 'in_progress',
      timestamp: startedAt,
      data: debate,
      outcome: 'pending' // need to provide a value for non-null column
    });

    this.emit('workspace:debate:started', debate);
    return debate;
  }

  async completeDebate(
    workspaceId: string,
    debateId: string,
    decision: CouncilDecision,
    tokensUsed: number = 0,
    cost: number = 0
  ): Promise<WorkspaceDebate | undefined> {
    const workspace = await this.getWorkspace(workspaceId);
    if (!workspace) return undefined;

    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilDebatesTable).where(eq(schema.councilDebatesTable.id, debateId)).limit(1);

    if (rows.length === 0) return undefined;

    const existingDebate = rows[0].data as WorkspaceDebate;
    const completedAt = new Date();
    const duration = completedAt.getTime() - new Date(existingDebate.startedAt).getTime();

    const updatedDebate: WorkspaceDebate = {
      ...existingDebate,
      decision,
      completedAt,
      status: 'completed',
    };

    // Update debate record
    await db.update(schema.councilDebatesTable)
      .set({
        status: 'completed',
        outcome: decision.approved ? 'approved' : 'rejected',
        consensus: decision.consensus,
        weightedConsensus: decision.weightedConsensus ?? decision.consensus,
        data: updatedDebate
      })
      .where(eq(schema.councilDebatesTable.id, debateId));

    // Update workspace metadata
    const totalDebates = workspace.metadata.totalDebates + 1;
    const previousTotalDuration = workspace.metadata.averageDebateDuration * workspace.metadata.totalDebates;
      
    await this.updateWorkspace(workspaceId, {
        metadata: {
          ...workspace.metadata,
          totalDebates,
          approvedDebates: workspace.metadata.approvedDebates + (decision.approved ? 1 : 0),
          rejectedDebates: workspace.metadata.rejectedDebates + (decision.approved ? 0 : 1),
          totalTokensUsed: workspace.metadata.totalTokensUsed + tokensUsed,
          estimatedCost: workspace.metadata.estimatedCost + cost,
          lastDebateAt: completedAt,
          averageDebateDuration: (previousTotalDuration + duration) / totalDebates,
        },
    });

    this.emit('workspace:debate:completed', updatedDebate);
    return updatedDebate;
  }

  async failDebate(workspaceId: string, debateId: string, error: string): Promise<WorkspaceDebate | undefined> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilDebatesTable).where(eq(schema.councilDebatesTable.id, debateId)).limit(1);
    if (rows.length === 0) return undefined;

    const existingDebate = rows[0].data as WorkspaceDebate;
    const updatedDebate: WorkspaceDebate = {
      ...existingDebate,
      completedAt: new Date(),
      status: 'failed',
    };

    await db.update(schema.councilDebatesTable)
      .set({ status: 'failed', data: updatedDebate })
      .where(eq(schema.councilDebatesTable.id, debateId));

    this.emit('workspace:debate:failed', { debate: updatedDebate, error });
    return updatedDebate;
  }

  async getWorkspaceDebates(workspaceId: string, limit: number = 50): Promise<WorkspaceDebate[]> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilDebatesTable)
      .where(eq(schema.councilDebatesTable.workspaceId, workspaceId))
      .orderBy(desc(schema.councilDebatesTable.timestamp))
      .limit(limit);

    return rows.map(r => r.data as WorkspaceDebate);
  }

  async getActiveDebates(workspaceId: string): Promise<WorkspaceDebate[]> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilDebatesTable)
      .where(and(
          eq(schema.councilDebatesTable.workspaceId, workspaceId),
          eq(schema.councilDebatesTable.status, 'in_progress')
      ));

    return rows.map(r => r.data as WorkspaceDebate);
  }

  async getAllActiveDebates(): Promise<WorkspaceDebate[]> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const rows = await db.select().from(schema.councilDebatesTable)
      .where(eq(schema.councilDebatesTable.status, 'in_progress'));

    return rows.map(r => r.data as WorkspaceDebate);
  }

  // ============ Statistics & Analytics ============

  async getWorkspaceStats(workspaceId: string, periodDays: number = 30): Promise<WorkspaceStats | undefined> {
    const workspace = await this.getWorkspace(workspaceId);
    if (!workspace) return undefined;

    // Basic stats from metadata
    return {
      workspaceId,
      period: { start: new Date(), end: new Date() },
      debates: {
        total: workspace.metadata.totalDebates,
        approved: workspace.metadata.approvedDebates,
        rejected: workspace.metadata.rejectedDebates,
        pending: 0,
      },
      performance: {
        avgDurationMs: workspace.metadata.averageDebateDuration,
        avgConsensus: 0,
        avgConfidence: 0,
      },
      supervisors: {
        mostActive: [],
        highestAgreement: [],
      },
      tokens: {
        total: workspace.metadata.totalTokensUsed,
        perDebate: 0,
      },
      cost: {
        total: workspace.metadata.estimatedCost,
        perDebate: 0,
      },
    };
  }

  async compareWorkspaces(workspaceIds: string[]): Promise<WorkspaceComparison> {
    const metrics = await Promise.all(workspaceIds.map(async id => {
      const stats = await this.getWorkspaceStats(id);
      const workspace = await this.getWorkspace(id);
      if (!stats || !workspace) return null;

      return {
        workspaceId: id,
        name: workspace.name,
        debates: stats.debates.total,
        approvalRate: stats.debates.total > 0 ? stats.debates.approved / stats.debates.total : 0,
        avgConsensus: stats.performance.avgConsensus,
        avgDuration: stats.performance.avgDurationMs,
        totalCost: stats.cost.total,
      };
    }));

    const validMetrics = metrics.filter((m): m is NonNullable<typeof m> => m !== null);

    const byApprovalRate = [...validMetrics].sort((a, b) => b.approvalRate - a.approvalRate).map(m => m.workspaceId);
    const byConsensus = [...validMetrics].sort((a, b) => b.avgConsensus - a.avgConsensus).map(m => m.workspaceId);
    const byEfficiency = [...validMetrics].sort((a, b) => a.totalCost - b.totalCost).map(m => m.workspaceId); // Lower cost is better

    return {
        workspaces: workspaceIds,
        metrics: validMetrics,
        ranking: { byApprovalRate, byConsensus, byEfficiency }
    };
  }

  // ============ Bulk Operations ============

  async pauseAllWorkspaces(): Promise<number> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const result = await db.update(schema.councilWorkspacesTable)
      .set({ status: 'paused' })
      .where(eq(schema.councilWorkspacesTable.status, 'active'));
    
    const count = (result as any).changes || 0;
    this.emit('workspaces:paused_all', { count });
    return count;
  }

  async resumeAllWorkspaces(): Promise<number> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    const result = await db.update(schema.councilWorkspacesTable)
      .set({ status: 'active' })
      .where(eq(schema.councilWorkspacesTable.status, 'paused'));
    
    const count = (result as any).changes || 0;
    this.emit('workspaces:resumed_all', { count });
    return count;
  }

  // ============ Config Templates ============

  async cloneWorkspaceConfig(sourceId: string, targetId: string): Promise<boolean> {
    const source = await this.getWorkspace(sourceId);
    const target = await this.getWorkspace(targetId);
    if (!source || !target) return false;

    await this.updateWorkspaceConfig(targetId, { ...source.config });
    return true;
  }

  // ============ Helpers ============

  private generateId(): string {
    return `ws_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  // ============ Export/Import ============

  async exportWorkspace(id: string): Promise<{ workspace: Workspace; debates: WorkspaceDebate[] } | undefined> {
    const workspace = await this.getWorkspace(id);
    if (!workspace) return undefined;

    const debates = await this.getWorkspaceDebates(id, 1000);

    return {
      workspace,
      debates,
    };
  }

  async importWorkspace(data: { workspace: Workspace; debates: WorkspaceDebate[] }): Promise<Workspace> {
    const newId = this.generateId();
    const now = new Date();

    const imported: Workspace = {
      ...data.workspace,
      id: newId,
      name: `${data.workspace.name} (imported)`,
      createdAt: now,
      updatedAt: now,
    };

    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();

    await db.insert(schema.councilWorkspacesTable).values({
      id: imported.id,
      name: imported.name,
      path: imported.path,
      status: imported.status,
      config: { config: imported.config, metadata: imported.metadata },
      description: imported.description || null,
      createdAt: imported.createdAt,
      updatedAt: imported.updatedAt
    });

    this.emit('workspace:imported', imported);
    return imported;
  }

  // ============ Cleanup ============

  async clearAllWorkspaces(): Promise<void> {
    const db = dbService.getDrizzle();
    const schema = dbService.getSchema();
    // Also clear debates associated with workspaces
    await db.delete(schema.councilDebatesTable).where(sql`workspace_id IS NOT NULL`);
    await db.delete(schema.councilWorkspacesTable);
    this.activeWorkspaceId = null;
    this.emit('workspaces:cleared');
  }
}

export const workspaceManager = new WorkspaceManagerService();
