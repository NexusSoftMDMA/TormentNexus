import { TRPCError } from '@trpc/server';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const fsMock = {
  readFileSync: vi.fn(),
  writeFileSync: vi.fn(),
  existsSync: vi.fn(),
};

vi.mock('fs', () => fsMock);

vi.mock('./sessionRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { sessionRouter: t.router({}) };
});

vi.mock('./councilRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { councilRouter: t.router({}) };
});

vi.mock('./quotaRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { quotaRouter: t.router({}) };
});

vi.mock('./historyRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { historyRouter: t.router({}) };
});

vi.mock('./visualRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { visualRouter: t.router({}) };
});

vi.mock('./smartPilotRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { smartPilotRouter: t.router({}) };
});

vi.mock('./evolutionRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { evolutionRouter: t.router({}) };
});

vi.mock('./hooksRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { hooksRouter: t.router({}) };
});

vi.mock('./fineTuneRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { fineTuneRouter: t.router({}) };
});

vi.mock('./ideRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { ideRouter: t.router({}) };
});

vi.mock('./rotationRouter.js', async () => {
  const { t } = await import('../../lib/trpc-core.js');
  return { rotationRouter: t.router({}) };
});

describe('councilRouter members truthfulness', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns an empty list when council.json is missing', async () => {
    const missing = new Error('missing');
    (missing as NodeJS.ErrnoException).code = 'ENOENT';
    fsMock.readFileSync.mockImplementation(() => {
      throw missing;
    });

    const { councilRouter } = await import('./index.js');
    const caller = councilRouter.createCaller({});

    await expect(caller.members()).resolves.toEqual([]);
  });

  it('returns a concise TRPC error when council.json is invalid', async () => {
    fsMock.readFileSync.mockReturnValue('{');

    const { councilRouter } = await import('./index.js');
    const caller = councilRouter.createCaller({});

    await expect(caller.members()).rejects.toMatchObject<Partial<TRPCError>>({
      message: 'Council configuration is unavailable: council.json contains invalid JSON.',
    });
  });
});
