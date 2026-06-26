import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../../lib/trpc-core.js';
import { debateHistory } from '../../orchestrator/council/services/debate-history.js';
import { rethrowSqliteUnavailableAsTrpc } from '../sqliteTrpc.js';

const debateQueryOptionsSchema = z.object({
  sessionId: z.string().optional(),
  taskType: z.string().optional(),
  approved: z.boolean().optional(),
  supervisorName: z.string().optional(),
  fromTimestamp: z.number().optional(),
  toTimestamp: z.number().optional(),
  minConsensus: z.number().optional(),
  maxConsensus: z.number().optional(),
  limit: z.number().optional(),
  offset: z.number().optional(),
  sortBy: z.enum(['timestamp', 'consensus', 'duration']).optional(),
  sortOrder: z.enum(['asc', 'desc']).optional(),
});

export const historyRouter = t.router({
  status: publicProcedure.query(async () => {
    try {
      return {
        enabled: debateHistory.isEnabled(),
        recordCount: await debateHistory.getRecordCount(),
        storageSize: debateHistory.getStorageSize(),
        config: debateHistory.getConfig(),
      };
    } catch (error) {
      rethrowSqliteUnavailableAsTrpc('Council debate history is unavailable', error);
    }
  }),

  getConfig: publicProcedure.query(async () => {
    return debateHistory.getConfig();
  }),

  updateConfig: adminProcedure.input(z.any()).mutation(async ({ input }) => {
    return debateHistory.updateConfig(input);
  }),

  toggle: adminProcedure.input(z.object({ enabled: z.boolean().optional() })).mutation(async ({ input }) => {
    const enabled = input.enabled ?? !debateHistory.isEnabled();
    debateHistory.updateConfig({ enabled });
    return { enabled: debateHistory.isEnabled() };
  }),

  stats: publicProcedure.query(async () => {
    try {
      return await debateHistory.getStats();
    } catch (error) {
      rethrowSqliteUnavailableAsTrpc('Council debate history is unavailable', error);
    }
  }),

  list: publicProcedure.input(debateQueryOptionsSchema).query(async ({ input }) => {
    try {
      const records = await debateHistory.queryDebates(input as any);
      return {
        records,
        meta: {
          count: records.length,
          totalRecords: await debateHistory.getRecordCount(),
        },
      };
    } catch (error) {
      rethrowSqliteUnavailableAsTrpc('Council debate history is unavailable', error);
    }
  }),

  get: publicProcedure.input(z.object({ id: z.string() })).query(async ({ input }) => {
    try {
      const record = await debateHistory.getDebate(input.id);
      if (!record) throw new Error('Debate not found');
      return record;
    } catch (error) {
      rethrowSqliteUnavailableAsTrpc('Council debate history is unavailable', error);
    }
  }),

  delete: adminProcedure.input(z.object({ id: z.string() })).mutation(async ({ input }) => {
    try {
      const deleted = await debateHistory.deleteRecord(input.id);
      if (!deleted) throw new Error('Debate not found');
      return { deleted: true, id: input.id };
    } catch (error) {
      rethrowSqliteUnavailableAsTrpc('Council debate history is unavailable', error);
    }
  }),

  supervisorHistory: publicProcedure.input(z.object({ name: z.string() })).query(async ({ input }) => {
    try {
      return await debateHistory.getSupervisorVoteHistory(input.name);
    } catch (error) {
      rethrowSqliteUnavailableAsTrpc('Council debate history is unavailable', error);
    }
  }),

  clear: adminProcedure.mutation(async () => {
    try {
      const count = await debateHistory.clearAll();
      return { cleared: count };
    } catch (error) {
      rethrowSqliteUnavailableAsTrpc('Council debate history is unavailable', error);
    }
  }),

  initialize: adminProcedure.mutation(async () => {
    try {
      await debateHistory.initialize();
      return {
        initialized: true,
        recordCount: await debateHistory.getRecordCount(),
      };
    } catch (error) {
      rethrowSqliteUnavailableAsTrpc('Council debate history is unavailable', error);
    }
  }),
});
