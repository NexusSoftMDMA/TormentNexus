import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockGetDb = vi.fn();
const mockBroadcast = vi.fn();

vi.mock('../db.js', () => ({
  dbService: {
    getDb: mockGetDb,
  },
}));

vi.mock('../ws-manager.js', () => ({
  wsManager: {
    broadcast: mockBroadcast,
  },
}));

describe('collectiveMemory', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
  });

  it('surfaces a concise degraded-mode error when SQLite is unavailable', async () => {
    mockGetDb.mockImplementation(() => {
      throw new Error(
        'SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)',
      );
    });

    const { collectiveMemory } = await import('../collective-memory.js');

    await expect(collectiveMemory.getAllFacts()).rejects.toThrow(
      'Collective memory is unavailable: SQLite runtime is unavailable for this run.',
    );
    expect(mockBroadcast).not.toHaveBeenCalled();
  });
});
