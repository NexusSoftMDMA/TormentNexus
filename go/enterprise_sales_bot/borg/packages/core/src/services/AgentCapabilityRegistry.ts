import { MeshService, SwarmMessageType } from '../mesh/MeshService.js';

export interface CachedCapabilities {
    capabilities: string[];
    role?: string;
    load?: number;
    cachedAt: number;
}

export class AgentCapabilityRegistry {
    private readonly mesh: MeshService;
    private readonly cacheTtlMs: number;
    private readonly now: () => number;
    private readonly cache = new Map<string, CachedCapabilities>();

    constructor(
        mesh: MeshService,
        options: {
            cacheTtlMs?: number;
            now?: () => number;
        } = {},
    ) {
        this.mesh = mesh;
        this.cacheTtlMs = options.cacheTtlMs ?? 5 * 60 * 1000;
        this.now = options.now ?? (() => Date.now());
    }

    public getCachedCapabilities(nodeId: string): CachedCapabilities | null {
        const cached = this.cache.get(nodeId);
        if (!cached) {
            return null;
        }

        if (this.now() - cached.cachedAt > this.cacheTtlMs) {
            this.cache.delete(nodeId);
            return null;
        }

        return cached;
    }

    public async discoverCapabilities(nodeId: string, timeoutMs: number = 5_000): Promise<CachedCapabilities> {
        const cached = this.getCachedCapabilities(nodeId);
        if (cached) {
            return cached;
        }

        const response = await this.mesh.request(
            nodeId,
            SwarmMessageType.CAPABILITY_QUERY,
            {},
            {
                timeoutMs,
                responseTypes: [SwarmMessageType.CAPABILITY_RESPONSE],
            },
        );

        const payload = response.payload as {
            capabilities?: unknown;
            role?: unknown;
            load?: unknown;
        };

        const next: CachedCapabilities = {
            capabilities: Array.isArray(payload?.capabilities)
                ? payload.capabilities.filter((value): value is string => typeof value === 'string')
                : [],
            role: typeof payload?.role === 'string' ? payload.role : undefined,
            load: typeof payload?.load === 'number' ? payload.load : undefined,
            cachedAt: this.now(),
        };

        this.cache.set(nodeId, next);
        return next;
    }

    public async findPeerForCapabilities(requiredCapabilities: string[], timeoutMs: number = 5_000): Promise<{
        nodeId: string;
        capabilities: string[];
        role?: string;
        load?: number;
    } | null> {
        const required = requiredCapabilities
            .map((value) => value.trim())
            .filter((value) => value.length > 0);

        if (required.length === 0) {
            return null;
        }

        for (const nodeId of this.mesh.getPeers()) {
            try {
                const discovered = await this.discoverCapabilities(nodeId, timeoutMs);
                const matches = required.every((capability) => discovered.capabilities.includes(capability));
                if (!matches) {
                    continue;
                }

                return {
                    nodeId,
                    capabilities: discovered.capabilities,
                    role: discovered.role,
                    load: discovered.load,
                };
            } catch {
                continue;
            }
        }

        return null;
    }
}
