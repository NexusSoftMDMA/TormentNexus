import { IMemoryProvider, Memory } from '../../interfaces/IMemoryProvider.js';
import { JsonMemoryProvider } from './JsonMemoryProvider.js';
import { RedundantMemoryManager } from './RedundantMemoryManager.js';

export type MemoryProviderMode = 'json' | 'postgres' | 'redundant';

export type MemoryPipelineSummary = {
    configuredMode: MemoryProviderMode;
    providerNames: string[];
    providerCount: number;
    sectionedStoreEnabled: boolean;
};

export class MemoryManager implements IMemoryProvider {
    private provider: IMemoryProvider;
    private readonly providerType: MemoryProviderMode;

    constructor(workspaceRoot: string, providerType: MemoryProviderMode = 'redundant') {
        this.providerType = providerType;
        // Provider selection:
        // - 'redundant' (default): Fan-out to ALL providers (JSON + sectioned store + future vector).
        //   This ensures maximum data durability and search recall.
        // - 'json': Legacy single-provider mode using flat-file JSON.
        // - 'postgres': Placeholder for future DB-backed provider.
        if (providerType === 'redundant') {
            this.provider = new RedundantMemoryManager(workspaceRoot);
        } else {
            this.provider = new JsonMemoryProvider(workspaceRoot);
        }
    }

    getPipelineSummary(): MemoryPipelineSummary {
        if (this.provider instanceof RedundantMemoryManager) {
            const providerNames = this.provider.getProviderNames();
            return {
                configuredMode: this.providerType,
                providerNames,
                providerCount: providerNames.length,
                sectionedStoreEnabled: providerNames.includes('sectioned-store'),
            };
        }

        return {
            configuredMode: this.providerType,
            providerNames: ['json'],
            providerCount: 1,
            sectionedStoreEnabled: false,
        };
    }

    async init(): Promise<void> {
        await this.provider.init();
    }

    async saveMemory(content: string, metadata: Record<string, any>, userId: string, agentId?: string): Promise<Memory> {
        return this.provider.saveMemory(content, metadata, userId, agentId);
    }

    async searchMemories(query: string, userId: string, limit?: number, threshold?: number): Promise<Memory[]> {
        return this.provider.searchMemories(query, userId, limit, threshold);
    }

    async listMemories(userId: string, limit?: number, offset?: number): Promise<Memory[]> {
        return this.provider.listMemories(userId, limit, offset);
    }

    async deleteMemory(uuid: string, userId: string): Promise<void> {
        return this.provider.deleteMemory(uuid, userId);
    }
}
