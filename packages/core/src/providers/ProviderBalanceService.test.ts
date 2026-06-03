import { afterEach, describe, expect, it, vi } from 'vitest';

import { ProviderBalanceService } from './ProviderBalanceService.js';
import type { ProviderBalanceConnectionSource } from './types.js';

describe('ProviderBalanceService', () => {
    afterEach(() => {
        vi.unstubAllGlobals();
    });

    it('normalizes GitHub Copilot quota snapshots into TormentNexus provider quotas', async () => {
        const source: ProviderBalanceConnectionSource = {
            async getConnection(provider) {
                if (provider !== 'github') {
                    return null;
                }

                return {
                    id: 'github-1',
                    provider: 'github',
                    authMethod: 'oauth',
                    accessToken: 'gho_test',
                };
            },
        };

        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            async json() {
                return {
                    copilot_plan: 'copilot_pro',
                    quota_reset_date: '2026-03-31T00:00:00.000Z',
                    quota_snapshots: {
                        chat: {
                            entitlement: 300,
                            remaining: 120,
                        },
                    },
                };
            },
        }));

        const service = new ProviderBalanceService({ connectionSource: source });
        const snapshots = await service.fetchSnapshots();
        const github = snapshots.find((snapshot) => snapshot.provider === 'github-copilot');

        expect(github).toMatchObject({
            provider: 'github-copilot',
            name: 'GitHub Copilot',
            authMethod: 'oauth',
            used: 180,
            limit: 300,
            remaining: 120,
            tier: 'copilot_pro',
            availability: 'available',
            source: 'balance',
            connectionId: 'github-1',
        });
        expect(github?.windows).toEqual([
            expect.objectContaining({
                key: 'chat',
                limit: 300,
                remaining: 120,
            }),
        ]);
    });

    it('normalizes Codex percent windows and preserves unsupported connections as missing auth cards', async () => {
        const source: ProviderBalanceConnectionSource = {
            async getConnection(provider) {
                if (provider === 'codex') {
                    return {
                        id: 'codex-1',
                        provider: 'codex',
                        authMethod: 'oauth',
                        accessToken: 'openai_test',
                        metadata: {
                            workspaceId: 'ws_123',
                        },
                    };
                }

                return null;
            },
        };

        vi.stubGlobal('fetch', vi.fn().mockImplementation(async (url: string) => {
            if (url.includes('chatgpt.com/backend-api/wham/usage')) {
                return {
                    ok: true,
                    async json() {
                        return {
                            plan_type: 'chatgpt_plus',
                            rate_limit: {
                                primary_window: { used_percent: 25, reset_at: 1_774_329_600 },
                                secondary_window: { used_percent: 70, reset_at: 1_774_588_800 },
                            },
                        };
                    },
                };
            }

            throw new Error(`Unexpected fetch ${url}`);
        }));

        const service = new ProviderBalanceService({ connectionSource: source });
        const snapshots = await service.fetchSnapshots();
        const codex = snapshots.find((snapshot) => snapshot.provider === 'openai');
        const antigravity = snapshots.find((snapshot) => snapshot.provider === 'antigravity');

        expect(codex).toMatchObject({
            provider: 'openai',
            limit: 100,
            used: 70,
            remaining: 30,
            tier: 'chatgpt_plus',
        });
        expect(codex?.windows).toEqual([
            expect.objectContaining({ key: 'session', used: 25, remaining: 75 }),
            expect.objectContaining({ key: 'weekly', used: 70, remaining: 30 }),
        ]);
        expect(antigravity).toMatchObject({
            configured: false,
            authenticated: false,
            availability: 'missing_auth',
            provider: 'antigravity',
        });
    });
});