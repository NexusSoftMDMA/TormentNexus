import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/visual-workflows.repo.js', () => ({
    visualWorkflowsRepo: {
        updateWorkflow: vi.fn(),
        createWorkflow: vi.fn(),
        getWorkflow: vi.fn(),
        listWorkflows: vi.fn(),
    },
}));

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual('../lib/trpc-core.js');
    return {
        ...actual,
        getWorkflowEngine: vi.fn(),
        getWorkflowDefinitions: vi.fn(),
    };
});

const { workflowRouter } = await import('./workflowRouter.js');
const { visualWorkflowsRepo } = await import('../db/repositories/visual-workflows.repo.js');
const { getWorkflowEngine, getWorkflowDefinitions } = await import('../lib/trpc-core.js');

function createCaller() {
    return workflowRouter.createCaller({} as never);
}

describe('workflowRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(visualWorkflowsRepo);
    const getWorkflowEngineMock = vi.mocked(getWorkflowEngine);
    const getWorkflowDefinitionsMock = vi.mocked(getWorkflowDefinitions);

    beforeEach(() => {
        vi.clearAllMocks();
        getWorkflowEngineMock.mockReturnValue(null);
        getWorkflowDefinitionsMock.mockReturnValue([]);
    });

    it('surfaces a clear error for loadCanvas when SQLite is unavailable', async () => {
        repositoryMocks.getWorkflow.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.loadCanvas({ id: 'workflow-1' })).rejects.toMatchObject({
            message: 'Workflow canvas storage is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for saveCanvas when SQLite is unavailable', async () => {
        repositoryMocks.createWorkflow.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.saveCanvas({
            name: 'Test workflow',
            nodes: [],
            edges: [],
        })).rejects.toMatchObject({
            message: 'Workflow canvas storage is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for workflow definitions when the engine is unavailable', async () => {
        const caller = createCaller();

        await expect(caller.list()).rejects.toMatchObject({
            message: 'Workflow engine is unavailable for this run.',
        });
    });

    it('surfaces a clear error for workflow executions when the engine is unavailable', async () => {
        const caller = createCaller();

        await expect(caller.listExecutions()).rejects.toMatchObject({
            message: 'Workflow engine is unavailable for this run.',
        });
        await expect(caller.getExecution({ executionId: 'exec-1' })).rejects.toMatchObject({
            message: 'Workflow engine is unavailable for this run.',
        });
        await expect(caller.getHistory({ executionId: 'exec-1' })).rejects.toMatchObject({
            message: 'Workflow engine is unavailable for this run.',
        });
    });

    it('surfaces a clear not-found error for missing workflow graphs', async () => {
        getWorkflowEngineMock.mockReturnValue({
            getGraph: vi.fn().mockReturnValue(undefined),
        } as never);

        const caller = createCaller();

        await expect(caller.getGraph({ workflowId: 'missing-workflow' })).rejects.toMatchObject({
            message: 'Workflow "missing-workflow" was not found.',
        });
    });
});
