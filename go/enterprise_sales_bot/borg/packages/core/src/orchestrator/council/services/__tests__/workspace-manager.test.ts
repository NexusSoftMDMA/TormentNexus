import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { WorkspaceManagerService, type WorkspaceConfig } from '../workspace-manager.js';
import { createDrizzleMock, createSchemaMock, resetMockDb } from './mock-db.js';

vi.mock('../db.js', () => {
    return {
        dbService: {
            getDb: vi.fn(),
            getDrizzle: vi.fn(() => createDrizzleMock()),
            getSchema: vi.fn(() => createSchemaMock()),
            close: vi.fn(),
        }
    };
});
import type { CouncilDecision } from '../types.js';

describe('WorkspaceManagerService', () => {
  let manager: WorkspaceManagerService;

  beforeEach(async () => {
    resetMockDb();
    manager = new WorkspaceManagerService();
  });

  describe('workspace CRUD', () => {
    it('should create a workspace with default config', async () => {
      const workspace = await manager.createWorkspace('Test Project', '/path/to/project');
      
      expect(workspace.id).toMatch(/^ws_/);
      expect(workspace.name).toBe('Test Project');
      expect(workspace.path).toBe('/path/to/project');
      expect(workspace.status).toBe('active');
      expect(workspace.config.defaultConsensusMode).toBe('weighted');
      expect(workspace.config.defaultDebateRounds).toBe(2);
      expect(workspace.metadata.totalDebates).toBe(0);
    });

    it('should create a workspace with custom config', async () => {
      const workspace = await manager.createWorkspace('Custom Project', '/custom/path', {
        defaultConsensusMode: 'unanimous',
        defaultDebateRounds: 3,
        supervisorTeam: ['gpt-4', 'claude'],
        budgetLimit: 100,
      }, 'A custom project');
      
      expect(workspace.config.defaultConsensusMode).toBe('unanimous');
      expect(workspace.config.defaultDebateRounds).toBe(3);
      expect(workspace.config.supervisorTeam).toEqual(['gpt-4', 'claude']);
      expect(workspace.config.budgetLimit).toBe(100);
      expect(workspace.description).toBe('A custom project');
    });

    it('should get workspace by id', async () => {
      const created = await manager.createWorkspace('Test', '/test');
      const retrieved = await manager.getWorkspace(created.id);
      
      expect(retrieved).toBeDefined();
      expect(retrieved?.name).toBe('Test');
    });

    it('should get workspace by path', async () => {
      await manager.createWorkspace('Test', '/unique/path');
      const retrieved = await manager.getWorkspaceByPath('/unique/path');
      
      expect(retrieved).toBeDefined();
      expect(retrieved?.name).toBe('Test');
    });

    it('should return undefined for non-existent workspace', async () => {
      expect(await manager.getWorkspace('non-existent')).toBeUndefined();
    });

    it('should get all workspaces', async () => {
      await manager.createWorkspace('Project 1', '/path1');
      await manager.createWorkspace('Project 2', '/path2');
      
      const all = await manager.getAllWorkspaces();
      expect(all.length).toBe(2);
    });

    it('should update workspace', async () => {
      const workspace = await manager.createWorkspace('Original', '/path');
      const updated = await manager.updateWorkspace(workspace.id, { name: 'Updated' });
      
      expect(updated?.name).toBe('Updated');
      expect(updated?.updatedAt.getTime()).toBeGreaterThanOrEqual(workspace.createdAt.getTime());
    });

    it('should update workspace config', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      const updated = await manager.updateWorkspaceConfig(workspace.id, { defaultDebateRounds: 5 });
      
      expect(updated?.config.defaultDebateRounds).toBe(5);
    });

    it('should delete workspace', async () => {
      const workspace = await manager.createWorkspace('ToDelete', '/path');
      const deleted = await manager.deleteWorkspace(workspace.id);
      
      expect(deleted).toBe(true);
      expect(await manager.getWorkspace(workspace.id)).toBeUndefined();
    });

    it('should archive workspace', async () => {
      const workspace = await manager.createWorkspace('ToArchive', '/path');
      const archived = await manager.archiveWorkspace(workspace.id);
      
      expect(archived?.status).toBe('archived');
    });
  });

  describe('workspace filtering', () => {
    beforeEach(async () => {
      const ws1 = await manager.createWorkspace('Active 1', '/path1', { tags: ['frontend'] });
      const ws2 = await manager.createWorkspace('Active 2', '/path2', { tags: ['backend'] });
      await manager.createWorkspace('Paused', '/path3', { tags: ['frontend'] });
      await manager.updateWorkspace((await manager.getAllWorkspaces())[2].id, { status: 'paused' });
    });

    it('should filter by status', async () => {
      const active = await manager.getWorkspacesByStatus('active');
      expect(active.length).toBe(2);
    });

    it('should filter by tag', async () => {
      const frontend = await manager.getWorkspacesByTag('frontend');
      expect(frontend.length).toBe(2);
    });
  });

  describe('active workspace', () => {
    it('should set and get active workspace', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      const success = await manager.setActiveWorkspace(workspace.id);
      
      expect(success).toBe(true);
      expect((await manager.getActiveWorkspace())?.id).toBe(workspace.id);
    });

    it('should not set archived workspace as active', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      await manager.archiveWorkspace(workspace.id);
      
      const success = await manager.setActiveWorkspace(workspace.id);
      expect(success).toBe(false);
    });

    it('should clear active workspace', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      await manager.setActiveWorkspace(workspace.id);
      manager.clearActiveWorkspace();
      
      expect(await manager.getActiveWorkspace()).toBeUndefined();
    });
  });

  describe('debate tracking', () => {
    it('should start a debate', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      const debate = await manager.startDebate(workspace.id, {
        id: 'task-1',
        description: 'Test task',
        context: 'Test context',
        files: ['file.ts'],
      });
      
      expect(debate).toBeDefined();
      expect(debate?.status).toBe('in_progress');
      expect(debate?.workspaceId).toBe(workspace.id);
    });

    it('should respect concurrent debate limit', async () => {
      const workspace = await manager.createWorkspace('Test', '/path', { maxConcurrentDebates: 1 });
      
      await manager.startDebate(workspace.id, { id: '1', description: '', context: '', files: [] });
      const second = await manager.startDebate(workspace.id, { id: '2', description: '', context: '', files: [] });
      
      expect(second).toBeUndefined();
    });

    it('should complete a debate and update metadata', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      const debate = await manager.startDebate(workspace.id, {
        id: 'task-1',
        description: 'Test',
        context: '',
        files: [],
      });
      
      const decision: CouncilDecision = {
        approved: true,
        consensus: 0.8,
        weightedConsensus: 0.85,
        votes: [],
        reasoning: 'Approved',
        dissent: [],
      };
      
      const completed = await manager.completeDebate(workspace.id, debate!.debateId, decision, 1000, 0.05);
      
      expect(completed?.status).toBe('completed');
      expect(completed?.decision).toBeDefined();
      
      const updatedWorkspace = await manager.getWorkspace(workspace.id);
      expect(updatedWorkspace?.metadata.totalDebates).toBe(1);
      expect(updatedWorkspace?.metadata.approvedDebates).toBe(1);
      expect(updatedWorkspace?.metadata.totalTokensUsed).toBe(1000);
    });

    it('should fail a debate', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      const debate = await manager.startDebate(workspace.id, {
        id: 'task-1',
        description: 'Test',
        context: '',
        files: [],
      });
      
      const failed = await manager.failDebate(workspace.id, debate!.debateId, 'API error');
      
      expect(failed?.status).toBe('failed');
    });

    it('should get workspace debates', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      await manager.startDebate(workspace.id, { id: '1', description: '', context: '', files: [] });
      await manager.startDebate(workspace.id, { id: '2', description: '', context: '', files: [] });
      
      const debates = await manager.getWorkspaceDebates(workspace.id);
      expect(debates.length).toBe(2);
    });

    it('should get all active debates across workspaces', async () => {
      const ws1 = await manager.createWorkspace('Test 1', '/path1');
      const ws2 = await manager.createWorkspace('Test 2', '/path2');
      
      await manager.startDebate(ws1.id, { id: '1', description: '', context: '', files: [] });
      await manager.startDebate(ws2.id, { id: '2', description: '', context: '', files: [] });
      
      const active = await manager.getAllActiveDebates();
      expect(active.length).toBe(2);
    });
  });

  describe('statistics', () => {
    it('should calculate workspace stats', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      const debate = await manager.startDebate(workspace.id, {
        id: 'task-1',
        description: 'Test',
        context: '',
        files: [],
      });
      
      await manager.completeDebate(workspace.id, debate!.debateId, {
        approved: true,
        consensus: 0.9,
        weightedConsensus: 0.9,
        votes: [
          { supervisor: 'gpt-4', approved: true, confidence: 0.9, weight: 1, comment: '' },
        ],
        reasoning: '',
        dissent: [],
      }, 500, 0.02);
      
      const stats = await manager.getWorkspaceStats(workspace.id, 30);
      
      expect(stats).toBeDefined();
      expect(stats?.debates.total).toBe(1);
      expect(stats?.debates.approved).toBe(1);
    });

    it('should compare workspaces', async () => {
      const ws1 = await manager.createWorkspace('Project 1', '/path1');
      const ws2 = await manager.createWorkspace('Project 2', '/path2');
      
      const comparison = await manager.compareWorkspaces([ws1.id, ws2.id]);
      
      expect(comparison.workspaces.length).toBe(2);
      expect(comparison.metrics.length).toBe(2);
      expect(comparison.ranking.byApprovalRate.length).toBe(2);
    });
  });

  describe('bulk operations', () => {
    it('should pause all workspaces', async () => {
      await manager.createWorkspace('Test 1', '/path1');
      await manager.createWorkspace('Test 2', '/path2');
      
      const count = await manager.pauseAllWorkspaces();
      
      expect(count).toBe(2);
      expect((await manager.getWorkspacesByStatus('paused')).length).toBe(2);
    });

    it('should resume all workspaces', async () => {
      await manager.createWorkspace('Test 1', '/path1');
      await manager.createWorkspace('Test 2', '/path2');
      await manager.pauseAllWorkspaces();
      
      const count = await manager.resumeAllWorkspaces();
      
      expect(count).toBe(2);
      expect((await manager.getWorkspacesByStatus('active')).length).toBe(2);
    });
  });

  describe('config cloning', () => {
    it('should clone config from one workspace to another', async () => {
      const source = await manager.createWorkspace('Source', '/source', {
        defaultDebateRounds: 5,
        consensusThreshold: 0.9,
      });
      const target = await manager.createWorkspace('Target', '/target');
      
      await manager.cloneWorkspaceConfig(source.id, target.id);
      
      const updated = await manager.getWorkspace(target.id);
      expect(updated?.config.defaultDebateRounds).toBe(5);
      expect(updated?.config.consensusThreshold).toBe(0.9);
    });
  });

  describe('export/import', () => {
    it('should export workspace with debates', async () => {
      const workspace = await manager.createWorkspace('Test', '/path');
      await manager.startDebate(workspace.id, { id: '1', description: '', context: '', files: [] });
      
      const exported = await manager.exportWorkspace(workspace.id);
      
      expect(exported).toBeDefined();
      expect(exported?.workspace.name).toBe('Test');
      expect(exported?.debates.length).toBe(1);
    });

    it('should import workspace', async () => {
      const workspace = await manager.createWorkspace('Original', '/original');
      const exported = await manager.exportWorkspace(workspace.id);
      if (!exported) throw new Error('Exported is undefined');
      
      const imported = await manager.importWorkspace(exported);
      
      expect(imported.name).toBe('Original (imported)');
      expect(imported.id).not.toBe(workspace.id);
    });
  });

  describe('cleanup', () => {
    it('should clear all workspaces', async () => {
      await manager.createWorkspace('Test 1', '/path1');
      await manager.createWorkspace('Test 2', '/path2');
      
      await manager.clearAllWorkspaces();
      
      expect((await manager.getAllWorkspaces()).length).toBe(0);
    });
  });
});
