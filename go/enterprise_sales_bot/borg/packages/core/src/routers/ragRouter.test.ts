import { TRPCError } from '@trpc/server';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const memoryManagerMock = {};
const intakeMock = {
    ingestFile: vi.fn(),
    ingestText: vi.fn(),
};

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getMemoryManager: () => memoryManagerMock,
    };
});

vi.mock('../services/rag/EmbeddingService.js', () => ({
    EmbeddingService: vi.fn(),
}));

vi.mock('../services/rag/DocumentIntakeService.js', () => ({
    DocumentIntakeService: vi.fn(() => intakeMock),
}));

describe('ragRouter truthfulness', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('throws a TRPC error when file ingestion fails', async () => {
        intakeMock.ingestFile.mockRejectedValueOnce(new Error('file intake offline'));

        const { ragRouter } = await import('./ragRouter.js');
        const caller = ragRouter.createCaller({});

        await expect(caller.ingestFile({
            filePath: 'README.md',
            userId: 'default',
            chunkSize: 1000,
            chunkOverlap: 200,
            strategy: 'recursive',
        })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Document ingestion failed: file intake offline',
        });
    });

    it('throws a TRPC error when text ingestion fails', async () => {
        intakeMock.ingestText.mockRejectedValueOnce(new Error('text intake offline'));

        const { ragRouter } = await import('./ragRouter.js');
        const caller = ragRouter.createCaller({});

        await expect(caller.ingestText({
            text: 'hello',
            sourceName: 'note.txt',
            userId: 'default',
            chunkSize: 1000,
            chunkOverlap: 200,
            strategy: 'recursive',
        })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Document ingestion failed: text intake offline',
        });
    });
});
