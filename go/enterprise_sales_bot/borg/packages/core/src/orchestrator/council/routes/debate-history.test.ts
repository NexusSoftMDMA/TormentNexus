import { beforeEach, describe, expect, it, vi } from 'vitest';

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
  exportToJson: vi.fn(),
  exportToCsv: vi.fn(),
  clearAll: vi.fn(),
  initialize: vi.fn(),
};

vi.mock('../services/debate-history.js', () => ({
  debateHistory: mockDebateHistory,
}));

describe('debate history routes', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
  });

  it('returns 503 when debate history storage is unavailable', async () => {
    mockDebateHistory.getRecordCount.mockRejectedValue(
      new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
    );

    const { debateHistoryRoutes } = await import('./debate-history.js');
    const response = await debateHistoryRoutes.request('http://localhost/status');
    const body = await response.json();

    expect(response.status).toBe(503);
    expect(body).toEqual({
      success: false,
      error: 'Debate history is unavailable: SQLite runtime is unavailable for this run.',
    });
  });
});
