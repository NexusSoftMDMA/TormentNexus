import { beforeEach, describe, expect, it, vi } from 'vitest';

const existsSyncMock = vi.fn();
const readFileSyncMock = vi.fn();
const writeFileSyncMock = vi.fn();

vi.mock('node:fs', () => ({
    existsSync: (...args: unknown[]) => existsSyncMock(...args),
    readFileSync: (...args: unknown[]) => readFileSyncMock(...args),
    writeFileSync: (...args: unknown[]) => writeFileSyncMock(...args),
}));

const { settingsRouter } = await import('./settingsRouter.js');

function createCaller() {
    return settingsRouter.createCaller({} as never);
}

describe('settingsRouter provider key mutations', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        delete process.env.OPENAI_API_KEY;
    });

    it('updates provider keys in process env and workspace env files', async () => {
        existsSyncMock.mockImplementation((filePath: string) => (
            filePath.endsWith('.env')
            || filePath.endsWith('packages\\core\\.env')
        ));
        readFileSyncMock.mockImplementation((filePath: string) => (
            filePath.endsWith('packages\\core\\.env')
                ? 'CORE_ONLY="1"\n'
                : 'EXISTING="1"\n'
        ));

        const caller = createCaller();
        const result = await caller.updateProviderKey({ provider: 'openai', key: 'sk-live-1234' });

        expect(process.env.OPENAI_API_KEY).toBe('sk-live-1234');
        expect(result).toEqual({ success: true, updatedKey: 'OPENAI_API_KEY' });
        expect(writeFileSyncMock.mock.calls).toEqual([
            [expect.stringMatching(/packages\\core\\\.env$/), 'CORE_ONLY="1"\nOPENAI_API_KEY="sk-live-1234"\n', 'utf8'],
            [expect.stringMatching(/packages\\core\\packages\\core\\\.env$/), 'CORE_ONLY="1"\nOPENAI_API_KEY="sk-live-1234"\n', 'utf8'],
        ]);
    });

    it('removes provider keys from process env and env files', async () => {
        process.env.OPENAI_API_KEY = 'sk-live-1234';
        existsSyncMock.mockImplementation((filePath: string) => (
            filePath.endsWith('.env')
            || filePath.endsWith('packages\\core\\.env')
        ));
        readFileSyncMock.mockImplementation((filePath: string) => (
            filePath.endsWith('packages\\core\\.env')
                ? 'OPENAI_API_KEY="sk-live-1234"\nCORE_ONLY="1"\n'
                : 'EXISTING="1"\nOPENAI_API_KEY="sk-live-1234"\n'
        ));

        const caller = createCaller();
        const result = await caller.removeProviderKey({ provider: 'openai' });

        expect(process.env.OPENAI_API_KEY).toBeUndefined();
        expect(result).toEqual({ success: true, removedKey: 'OPENAI_API_KEY', removedAny: true });
        expect(writeFileSyncMock.mock.calls).toEqual([
            [expect.stringMatching(/packages\\core\\\.env$/), 'CORE_ONLY="1"\n', 'utf8'],
            [expect.stringMatching(/packages\\core\\packages\\core\\\.env$/), 'CORE_ONLY="1"\n', 'utf8'],
        ]);
    });
});
