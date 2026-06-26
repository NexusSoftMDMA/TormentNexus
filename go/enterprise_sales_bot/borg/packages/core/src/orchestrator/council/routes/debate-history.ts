import { Hono, type Context } from 'hono';
import { debateHistory, type DebateQueryOptions } from '../services/debate-history.js';
import type { TaskType } from './types.js';
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from '../../../db/sqliteAvailability.js';

export const debateHistoryRoutes = new Hono();

function handleDebateHistoryRouteError(c: Context, error: unknown) {
  if (isSqliteUnavailableError(error)) {
    return c.json({
      success: false,
      error: formatOptionalSqliteFailure('Debate history is unavailable', error),
    }, 503);
  }

  throw error;
}

debateHistoryRoutes.get('/status', async (c) => {
  try {
    return c.json({
      success: true,
      data: {
        enabled: debateHistory.isEnabled(),
        recordCount: await debateHistory.getRecordCount(),
        storageSize: debateHistory.getStorageSize(),
        config: debateHistory.getConfig(),
      },
    });
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.get('/config', (c) => {
  return c.json({
    success: true,
    data: debateHistory.getConfig(),
  });
});

debateHistoryRoutes.post('/config', async (c) => {
  const body = await c.req.json();
  const config = debateHistory.updateConfig(body);
  return c.json({
    success: true,
    data: config,
  });
});

debateHistoryRoutes.post('/toggle', async (c) => {
  const body = await c.req.json<{ enabled?: boolean }>();
  const enabled = body.enabled ?? !debateHistory.isEnabled();
  debateHistory.updateConfig({ enabled });
  return c.json({
    success: true,
    data: { enabled: debateHistory.isEnabled() },
  });
});

debateHistoryRoutes.get('/stats', async (c) => {
  try {
    return c.json({
      success: true,
      data: await debateHistory.getStats(),
    });
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.get('/list', async (c) => {
  const query = c.req.query();
  
  const options: DebateQueryOptions = {
    sessionId: query.sessionId,
    taskType: query.taskType as TaskType | undefined,
    approved: query.approved === 'true' ? true : query.approved === 'false' ? false : undefined,
    supervisorName: query.supervisorName,
    fromTimestamp: query.fromTimestamp ? parseInt(query.fromTimestamp, 10) : undefined,
    toTimestamp: query.toTimestamp ? parseInt(query.toTimestamp, 10) : undefined,
    minConsensus: query.minConsensus ? parseFloat(query.minConsensus) : undefined,
    maxConsensus: query.maxConsensus ? parseFloat(query.maxConsensus) : undefined,
    limit: query.limit ? parseInt(query.limit, 10) : undefined,
    offset: query.offset ? parseInt(query.offset, 10) : undefined,
    sortBy: query.sortBy as 'timestamp' | 'consensus' | 'duration' | undefined,
    sortOrder: query.sortOrder as 'asc' | 'desc' | undefined,
  };

  try {
    const records = await debateHistory.queryDebates(options);
    return c.json({
      success: true,
      data: records,
      meta: {
        count: records.length,
        totalRecords: await debateHistory.getRecordCount(),
      },
    });
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.get('/debates/:id', async (c) => {
  try {
    const id = c.req.param('id');
    const record = await debateHistory.getDebate(id);
    
    if (!record) {
      return c.json({ success: false, error: 'Debate not found' }, 404);
    }
    
    return c.json({
      success: true,
      data: record,
    });
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.delete('/debates/:id', async (c) => {
  try {
    const id = c.req.param('id');
    const deleted = await debateHistory.deleteRecord(id);
    
    if (!deleted) {
      return c.json({ success: false, error: 'Debate not found' }, 404);
    }
    
    return c.json({
      success: true,
      data: { deleted: true, id },
    });
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.get('/supervisor/:name', async (c) => {
  try {
    const name = c.req.param('name');
    const history = await debateHistory.getSupervisorVoteHistory(name);
    return c.json({
      success: true,
      data: history,
    });
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.get('/export/json', async (c) => {
  const query = c.req.query();
  
  const options: DebateQueryOptions = {
    sessionId: query.sessionId,
    approved: query.approved === 'true' ? true : query.approved === 'false' ? false : undefined,
    limit: query.limit ? parseInt(query.limit, 10) : undefined,
  };

  try {
    const json = await debateHistory.exportToJson(options);
    
    c.header('Content-Type', 'application/json');
    c.header('Content-Disposition', 'attachment; filename="debate-history.json"');
    return c.body(json);
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.get('/export/csv', async (c) => {
  const query = c.req.query();
  
  const options: DebateQueryOptions = {
    sessionId: query.sessionId,
    approved: query.approved === 'true' ? true : query.approved === 'false' ? false : undefined,
    limit: query.limit ? parseInt(query.limit, 10) : undefined,
  };

  try {
    const csv = await debateHistory.exportToCsv(options);
    
    c.header('Content-Type', 'text/csv');
    c.header('Content-Disposition', 'attachment; filename="debate-history.csv"');
    return c.body(csv);
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.delete('/clear', async (c) => {
  try {
    const count = await debateHistory.clearAll();
    return c.json({
      success: true,
      data: { cleared: count },
    });
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});

debateHistoryRoutes.post('/initialize', async (c) => {
  try {
    await debateHistory.initialize();
    return c.json({
      success: true,
      data: {
        initialized: true,
        recordCount: await debateHistory.getRecordCount(),
      },
    });
  } catch (error) {
    return handleDebateHistoryRouteError(c, error);
  }
});
