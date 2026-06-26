import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockCollectiveMemory = {
  getAllFacts: vi.fn(),
  searchFacts: vi.fn(),
  storeFact: vi.fn(),
  recallFact: vi.fn(),
};

vi.mock('../services/collective-memory.js', () => ({
  collectiveMemory: mockCollectiveMemory,
}));

describe('council memory routes', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
  });

  it('returns 503 when collective memory is unavailable', async () => {
    mockCollectiveMemory.getAllFacts.mockRejectedValue(
      new Error('Collective memory is unavailable: SQLite runtime is unavailable for this run.'),
    );

    const { default: app } = await import('./memory.js');
    const response = await app.request('http://localhost/facts');
    const body = await response.json();

    expect(response.status).toBe(503);
    expect(body).toEqual({
      success: false,
      error: 'Collective memory is unavailable: SQLite runtime is unavailable for this run.',
    });
  });
});
