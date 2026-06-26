import { TRPCError } from '@trpc/server';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const mockDebateHistory = {
  isEnabled: vi.fn(() => true),
  getRecordCount: vi.fn(),
  getStorageSize: vi.fn(() => 0),
  getConfig: vi.fn(() => ({ enabled: true })),
  updateConfig: vi.fn(),
  getStats: vi.fn(),
  queryDebates: vi.fn(),
  getDebate: vi.fn(),
  deleteRecord: vi.fn(),
  getSupervisorVoteHistory: vi.fn(),
  clearAll: vi.fn(),
  initialize: vi.fn(),
};

vi.mock('../../orchestrator/council/services/debate-history.js', () => ({
  debateHistory: mockDebateHistory,
}));

describe('historyRouter degraded debate history handling', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('returns a concise TRPC error when status cannot read debate history', async () => {
    mockDebateHistory.getRecordCount.mockRejectedValueOnce(
      new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
    );

    const { historyRouter } = await import('./historyRouter.js');
    const caller = historyRouter.createCaller({});

    await expect(caller.status()).rejects.toMatchObject<Partial<TRPCError>>({
      message: 'Council debate history is unavailable: SQLite runtime is unavailable for this run.',
    });
  });

  it('returns a concise TRPC error when list cannot query debate records', async () => {
    mockDebateHistory.queryDebates.mockRejectedValueOnce(
      new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
    );

    const { historyRouter } = await import('./historyRouter.js');
    const caller = historyRouter.createCaller({});

    await expect(caller.list({})).rejects.toMatchObject<Partial<TRPCError>>({
      message: 'Council debate history is unavailable: SQLite runtime is unavailable for this run.',
    });
  });

  it('returns a concise TRPC error when stats cannot aggregate debate history', async () => {
    mockDebateHistory.getStats.mockRejectedValueOnce(
      new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
    );

    const { historyRouter } = await import('./historyRouter.js');
    const caller = historyRouter.createCaller({});

    await expect(caller.stats()).rejects.toMatchObject<Partial<TRPCError>>({
      message: 'Council debate history is unavailable: SQLite runtime is unavailable for this run.',
    });
  });
});
