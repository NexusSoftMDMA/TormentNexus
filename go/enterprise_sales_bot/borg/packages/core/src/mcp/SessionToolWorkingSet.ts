export interface SessionToolWorkingSetOptions {
    maxLoadedTools?: number;
    maxHydratedSchemas?: number;
    /** Tools idle longer than this (ms) are preferred eviction candidates. Default: 5 minutes. */
    idleEvictionThresholdMs?: number;
}

export interface LoadedToolState {
    name: string;
    hydrated: boolean;
    lastLoadedAt: number;
    lastHydratedAt: number | null;
    /** Unix ms of the most recent load, hydrate, or touchTool call for this tool. */
    lastAccessedAt: number;
}

export interface SessionToolEvictionEvent {
    toolName: string;
    timestamp: number;
    tier: 'loaded' | 'hydrated';
    idleEvicted: boolean;
    idleDurationMs: number;
}

export const DEFAULT_MAX_LOADED_TOOLS = 16;
export const DEFAULT_MAX_HYDRATED_SCHEMAS = 8;
const DEFAULT_IDLE_EVICTION_THRESHOLD_MS = 5 * 60 * 1000;
const MAX_EVICTION_HISTORY = 200;

/** Internal per-tool tracking entry. */
interface ToolEntry {
    loadedAt: number;
    accessedAt: number;
}

export class SessionToolWorkingSet {
    private maxLoadedTools: number;
    private maxHydratedSchemas: number;
    private idleEvictionThresholdMs: number;
    private readonly loadedTools = new Map<string, ToolEntry>();
    private readonly hydratedTools = new Map<string, ToolEntry>();
    private readonly alwaysLoadedTools = new Set<string>();
    private readonly alwaysLoadedState = new Map<string, ToolEntry>();
    private readonly evictionHistory: SessionToolEvictionEvent[] = [];

    constructor(options: SessionToolWorkingSetOptions = {}) {
        this.maxLoadedTools = options.maxLoadedTools ?? DEFAULT_MAX_LOADED_TOOLS;
        this.maxHydratedSchemas = options.maxHydratedSchemas ?? DEFAULT_MAX_HYDRATED_SCHEMAS;
        this.idleEvictionThresholdMs = options.idleEvictionThresholdMs ?? DEFAULT_IDLE_EVICTION_THRESHOLD_MS;
    }

    private touch(map: Map<string, ToolEntry>, name: string, timestamp: number): boolean {
        const entry = map.get(name);
        if (!entry) return false;
        entry.accessedAt = timestamp;
        return true;
    }

    /**
     * Picks the next eviction candidate with idle-first preference:
     * 1) if any non-pinned tools exceeded idle threshold, evict the stalest among those
     * 2) otherwise, evict standard LRU candidate
     */
    private pickEvictionCandidate(map: Map<string, ToolEntry>): { name: string; idleDurationMs: number; idleEvicted: boolean } | undefined {
        let idleCandidate: string | undefined;
        let idleCandidateAccessedAt = Infinity;
        let idleCandidateDuration = 0;

        let lruCandidate: string | undefined;
        let lruMinAccessedAt = Infinity;
        const now = Date.now();

        for (const [name, entry] of map) {
            if (this.alwaysLoadedTools.has(name)) {
                continue;
            }

            const idleDurationMs = Math.max(0, now - entry.accessedAt);
            const isIdle = idleDurationMs >= this.idleEvictionThresholdMs;

            if (isIdle && entry.accessedAt < idleCandidateAccessedAt) {
                idleCandidateAccessedAt = entry.accessedAt;
                idleCandidate = name;
                idleCandidateDuration = idleDurationMs;
            }

            if (entry.accessedAt < lruMinAccessedAt) {
                lruMinAccessedAt = entry.accessedAt;
                lruCandidate = name;
            }
        }

        if (idleCandidate) {
            return {
                name: idleCandidate,
                idleDurationMs: idleCandidateDuration,
                idleEvicted: true,
            };
        }

        if (!lruCandidate) {
            return undefined;
        }

        return {
            name: lruCandidate,
            idleDurationMs: Math.max(0, now - lruMinAccessedAt),
            idleEvicted: false,
        };
    }

    private recordEviction(event: SessionToolEvictionEvent): void {
        this.evictionHistory.push(event);
        if (this.evictionHistory.length > MAX_EVICTION_HISTORY) {
            this.evictionHistory.splice(0, this.evictionHistory.length - MAX_EVICTION_HISTORY);
        }
    }

    private ensureAlwaysLoadedState(name: string, timestamp: number): void {
        const entry = this.alwaysLoadedState.get(name);
        if (entry) {
            entry.accessedAt = timestamp;
            return;
        }

        this.alwaysLoadedState.set(name, {
            loadedAt: timestamp,
            accessedAt: timestamp,
        });
    }

    loadTool(name: string): string[] {
        if (this.alwaysLoadedTools.has(name)) {
            this.ensureAlwaysLoadedState(name, Date.now());
            return [];
        }

        const timestamp = Date.now();

        if (this.touch(this.loadedTools, name, timestamp)) {
            return [];
        }

        const evicted: string[] = [];

        while (this.loadedTools.size >= this.maxLoadedTools) {
            const candidate = this.pickEvictionCandidate(this.loadedTools);
            if (!candidate) break;
            this.loadedTools.delete(candidate.name);
            this.hydratedTools.delete(candidate.name);
            evicted.push(candidate.name);
            this.recordEviction({
                toolName: candidate.name,
                timestamp,
                tier: 'loaded',
                idleEvicted: candidate.idleEvicted,
                idleDurationMs: candidate.idleDurationMs,
            });
        }

        this.loadedTools.set(name, { loadedAt: timestamp, accessedAt: timestamp });
        return evicted;
    }

    hydrateTool(name: string): string[] {
        this.loadTool(name);

        const evicted: string[] = [];
        const timestamp = Date.now();

        if (this.hydratedTools.has(name)) {
            this.touch(this.hydratedTools, name, timestamp);
            return evicted;
        }

        while (this.hydratedTools.size >= this.maxHydratedSchemas) {
            const candidate = this.pickEvictionCandidate(this.hydratedTools);
            if (!candidate) break;
            this.hydratedTools.delete(candidate.name);
            evicted.push(candidate.name);
            this.recordEviction({
                toolName: candidate.name,
                timestamp,
                tier: 'hydrated',
                idleEvicted: candidate.idleEvicted,
                idleDurationMs: candidate.idleDurationMs,
            });
        }

        this.hydratedTools.set(name, { loadedAt: timestamp, accessedAt: timestamp });
        return evicted;
    }

    touchTool(name: string): boolean {
        const timestamp = Date.now();
        let touchedLoaded = false;
        if (this.alwaysLoadedTools.has(name)) {
            this.ensureAlwaysLoadedState(name, timestamp);
            touchedLoaded = true;
        } else {
            touchedLoaded = this.touch(this.loadedTools, name, timestamp);
        }

        const touchedHydrated = this.touch(this.hydratedTools, name, timestamp);
        return touchedLoaded || touchedHydrated;
    }

    unloadTool(name: string): boolean {
        if (this.alwaysLoadedTools.has(name)) {
            return this.hydratedTools.delete(name);
        }

        const removedLoaded = this.loadedTools.delete(name);
        const removedHydrated = this.hydratedTools.delete(name);
        return removedLoaded || removedHydrated;
    }

    isLoaded(name: string): boolean {
        return this.alwaysLoadedTools.has(name) || this.loadedTools.has(name);
    }

    isHydrated(name: string): boolean {
        return this.hydratedTools.has(name);
    }

    getLoadedToolNames(): string[] {
        return [
            ...this.alwaysLoadedTools,
            ...Array.from(this.loadedTools.keys()).filter((name) => !this.alwaysLoadedTools.has(name)),
        ];
    }

    listLoadedTools(): LoadedToolState[] {
        return this.getLoadedToolNames().map((name) => {
            const loadedEntry = this.loadedTools.get(name) ?? this.alwaysLoadedState.get(name);
            const hydratedEntry = this.hydratedTools.get(name);
            return {
                name,
                hydrated: this.isHydrated(name),
                lastLoadedAt: loadedEntry?.loadedAt ?? 0,
                lastHydratedAt: hydratedEntry?.loadedAt ?? null,
                lastAccessedAt: loadedEntry?.accessedAt ?? 0,
            };
        });
    }

    getLimits(): { maxLoadedTools: number; maxHydratedSchemas: number; idleEvictionThresholdMs: number } {
        return {
            maxLoadedTools: this.maxLoadedTools,
            maxHydratedSchemas: this.maxHydratedSchemas,
            idleEvictionThresholdMs: this.idleEvictionThresholdMs,
        };
    }

    reconfigure(options: SessionToolWorkingSetOptions): void {
        if (typeof options.maxLoadedTools === 'number') {
            this.maxLoadedTools = Math.max(1, Math.round(options.maxLoadedTools));
        }

        if (typeof options.maxHydratedSchemas === 'number') {
            this.maxHydratedSchemas = Math.max(1, Math.round(options.maxHydratedSchemas));
        }

        if (typeof options.idleEvictionThresholdMs === 'number') {
            this.idleEvictionThresholdMs = Math.max(0, Math.round(options.idleEvictionThresholdMs));
        }

        const timestamp = Date.now();

        while (this.loadedTools.size > this.maxLoadedTools) {
            const candidate = this.pickEvictionCandidate(this.loadedTools);
            if (!candidate) {
                break;
            }

            this.loadedTools.delete(candidate.name);
            this.hydratedTools.delete(candidate.name);
            this.recordEviction({
                toolName: candidate.name,
                timestamp,
                tier: 'loaded',
                idleEvicted: candidate.idleEvicted,
                idleDurationMs: candidate.idleDurationMs,
            });
        }

        while (this.hydratedTools.size > this.maxHydratedSchemas) {
            const candidate = this.pickEvictionCandidate(this.hydratedTools);
            if (!candidate) {
                break;
            }

            this.hydratedTools.delete(candidate.name);
            this.recordEviction({
                toolName: candidate.name,
                timestamp,
                tier: 'hydrated',
                idleEvicted: candidate.idleEvicted,
                idleDurationMs: candidate.idleDurationMs,
            });
        }
    }

    getEvictionHistory(): SessionToolEvictionEvent[] {
        return [...this.evictionHistory].reverse();
    }

    clearEvictionHistory(): void {
        this.evictionHistory.length = 0;
    }

    setAlwaysLoadedTools(names: string[]): void {
        const timestamp = Date.now();
        const previousAlwaysLoadedState = new Map(this.alwaysLoadedState);
        this.alwaysLoadedTools.clear();
        this.alwaysLoadedState.clear();

        for (const name of names) {
            this.alwaysLoadedTools.add(name);
            const existingEntry = previousAlwaysLoadedState.get(name)
                ?? this.loadedTools.get(name)
                ?? this.hydratedTools.get(name);

            this.alwaysLoadedState.set(name, {
                loadedAt: existingEntry?.loadedAt ?? timestamp,
                accessedAt: existingEntry?.accessedAt ?? timestamp,
            });
        }

        for (const loadedName of Array.from(this.loadedTools.keys())) {
            if (this.alwaysLoadedTools.has(loadedName)) {
                this.loadedTools.delete(loadedName);
            }
        }

        this.reconfigure({});
    }
}
