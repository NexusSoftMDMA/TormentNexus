import { beforeEach, describe, expect, it, vi } from 'vitest';

const repoMocks = vi.hoisted(() => ({
    findServerByUuid: vi.fn(),
    getActiveRecipe: vi.fn(),
    startValidationRun: vi.fn(),
    finishValidationRun: vi.fn(),
    updateServerStatus: vi.fn(),
}));

vi.mock('../db/repositories/published-catalog.repo.js', () => ({
    publishedCatalogRepository: repoMocks,
}));

import { validatePublishedServer } from './published-catalog-validator.js';

describe('validatePublishedServer', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        repoMocks.findServerByUuid.mockResolvedValue({
            uuid: 'server-1',
            transport: 'stdio',
            confidence: 50,
            status: 'normalized',
        });
        repoMocks.getActiveRecipe.mockResolvedValue({
            template: {
                type: 'stdio',
                command: 'npx',
                args: ['example-mcp'],
            },
        });
        repoMocks.startValidationRun.mockResolvedValue({ uuid: 'run-1' });
        repoMocks.finishValidationRun.mockResolvedValue(undefined);
        repoMocks.updateServerStatus.mockResolvedValue(undefined);
    });

    it('classifies stdio recipes as stdio transport instead of unsafe', async () => {
        const result = await validatePublishedServer('server-1');

        expect(result).toMatchObject({
            server_uuid: 'server-1',
            run_uuid: 'run-1',
            outcome: 'skipped',
            failure_class: 'stdio_transport',
            findings_summary: {
                transport: 'stdio',
                validation_mode: 'not_probed',
            },
        });

        expect(repoMocks.finishValidationRun).toHaveBeenCalledWith(expect.objectContaining({
            uuid: 'run-1',
            outcome: 'skipped',
            failure_class: 'stdio_transport',
            findings_summary: {
                transport: 'stdio',
                validation_mode: 'not_probed',
            },
        }));
        expect(repoMocks.updateServerStatus).not.toHaveBeenCalled();
    });
});
