import { TRPCError } from '@trpc/server';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('node:fs/promises', () => ({
    default: {
        readFile: vi.fn(),
        writeFile: vi.fn(),
    },
}));

describe('researchRouter ingestion queue truthfulness', () => {
    beforeEach(() => {
        vi.resetModules();
        vi.clearAllMocks();
    });

    it('returns a concise TRPC error when ingestion status data is unreadable', async () => {
        const fs = (await import('node:fs/promises')).default;
        vi.mocked(fs.readFile).mockImplementation(async (path: any) => {
            const target = String(path);
            if (target.endsWith('scripts\\ingestion-status.json')) {
                return '{invalid json';
            }

            if (target.endsWith('TORMENTNEXUS_MASTER_INDEX.jsonc')) {
                return '{"categories":{}}';
            }

            throw new Error(`Unexpected read: ${target}`);
        });

        const { researchRouter } = await import('./researchRouter.js');
        const caller = researchRouter.createCaller({});

        await expect(caller.ingestionQueue()).rejects.toMatchObject<Partial<TRPCError>>({
            message: "Research ingestion queue is unavailable: Expected property name or '}' in JSON at position 1 (line 1 column 2)",
        });
    });
});
