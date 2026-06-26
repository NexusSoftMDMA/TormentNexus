export type ToolPreferences = {
    importantTools: string[];
    alwaysLoadedTools: string[];
    autoLoadMinConfidence: number;
    /** Maximum number of tools the session working set will hold before LRU eviction. Range: 4..64. Default 16. */
    maxLoadedTools: number;
    /** Maximum number of hydrated schemas the session working set will hold before LRU eviction. Range: 2..32. Default 8. */
    maxHydratedSchemas: number;
};

type ToolSelectionSettings = {
    importantTools?: unknown;
    alwaysLoadedTools?: unknown;
    autoLoadMinConfidence?: unknown;
    maxLoadedTools?: unknown;
    maxHydratedSchemas?: unknown;
};

type ToolPreferenceDisplayFields = {
    name: string;
    description: string;
    server: string;
    alwaysOn?: boolean;
    matchReason?: string;
    score?: number;
    rank?: number;
};

function normalizeToolNames(value: unknown): string[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return Array.from(new Set(value
        .filter((item): item is string => typeof item === 'string')
        .map((name) => name.trim())
        .filter(Boolean)));
}

function normalizeAutoLoadMinConfidence(value: unknown): number {
    if (typeof value !== 'number' || Number.isNaN(value)) {
        return 0.85;
    }

    return Math.max(0.5, Math.min(0.99, value));
}

function normalizeMaxLoadedTools(value: unknown): number {
    if (typeof value !== 'number' || Number.isNaN(value)) {
        return 16;
    }

    return Math.max(4, Math.min(64, Math.round(value)));
}

function normalizeMaxHydratedSchemas(value: unknown): number {
    if (typeof value !== 'number' || Number.isNaN(value)) {
        return 8;
    }

    return Math.max(2, Math.min(32, Math.round(value)));
}

export function normalizeToolPreferences(value: { importantTools?: unknown; alwaysLoadedTools?: unknown; autoLoadMinConfidence?: unknown; maxLoadedTools?: unknown; maxHydratedSchemas?: unknown } | null | undefined): ToolPreferences {
    return {
        importantTools: normalizeToolNames(value?.importantTools),
        alwaysLoadedTools: normalizeToolNames(value?.alwaysLoadedTools),
        autoLoadMinConfidence: normalizeAutoLoadMinConfidence(value?.autoLoadMinConfidence),
        maxLoadedTools: normalizeMaxLoadedTools(value?.maxLoadedTools),
        maxHydratedSchemas: normalizeMaxHydratedSchemas(value?.maxHydratedSchemas),
    };
}

export function readToolPreferencesFromSettings(settings: ToolSelectionSettings | null | undefined): ToolPreferences {
    return normalizeToolPreferences({
        importantTools: settings?.importantTools,
        alwaysLoadedTools: settings?.alwaysLoadedTools,
        autoLoadMinConfidence: settings?.autoLoadMinConfidence,
        maxLoadedTools: settings?.maxLoadedTools,
        maxHydratedSchemas: settings?.maxHydratedSchemas,
    });
}

export function buildToolPreferenceSettings(
    existingSettings: Record<string, unknown>,
    nextPreferences: ToolPreferences,
): Record<string, unknown> {
    const normalized = normalizeToolPreferences(nextPreferences);

    return {
        ...existingSettings,
        toolSelection: {
            ...(existingSettings.toolSelection && typeof existingSettings.toolSelection === 'object'
                ? existingSettings.toolSelection as Record<string, unknown>
                : {}),
            importantTools: normalized.importantTools,
            alwaysLoadedTools: normalized.alwaysLoadedTools,
            autoLoadMinConfidence: normalized.autoLoadMinConfidence,
            maxLoadedTools: normalized.maxLoadedTools,
            maxHydratedSchemas: normalized.maxHydratedSchemas,
        },
    };
}

export function mergeToolPreferences<T extends ToolPreferenceDisplayFields>(
    results: T[],
    preferences: ToolPreferences,
    catalog: Array<{ name: string; description: string; server: string; alwaysOn?: boolean }>,
): Array<T & { important: boolean; alwaysShow: boolean; alwaysLoaded: boolean; alwaysOn: boolean }> {
    const importantSet = new Set(preferences.importantTools);
    const alwaysLoadedSet = new Set(preferences.alwaysLoadedTools);
    const merged = new Map<string, T & { important: boolean; alwaysShow: boolean; alwaysLoaded: boolean; alwaysOn: boolean }>();

    for (const item of results) {
        const alwaysOn = Boolean(item.alwaysOn);
        merged.set(item.name, {
            ...item,
            important: importantSet.has(item.name),
            alwaysShow: importantSet.has(item.name) || alwaysOn,
            alwaysLoaded: alwaysLoadedSet.has(item.name),
            alwaysOn,
        });
    }

    for (const catalogTool of catalog.filter((tool) => Boolean(tool.alwaysOn))) {
        if (merged.has(catalogTool.name)) {
            const current = merged.get(catalogTool.name)!;
            merged.set(catalogTool.name, {
                ...current,
                alwaysOn: true,
                alwaysShow: true,
                matchReason: current.matchReason ?? 'advertised because the tool is marked always-on',
            });
            continue;
        }

        merged.set(catalogTool.name, {
            ...(catalogTool as T),
            matchReason: 'advertised because the tool is marked always-on',
            score: 9_500,
            rank: 0,
            important: importantSet.has(catalogTool.name),
            alwaysShow: true,
            alwaysLoaded: alwaysLoadedSet.has(catalogTool.name),
            alwaysOn: true,
        });
    }

    for (const toolName of preferences.importantTools) {
        if (merged.has(toolName)) {
            continue;
        }

        const catalogTool = catalog.find((tool) => tool.name === toolName);
        if (!catalogTool) {
            continue;
        }

        merged.set(toolName, {
            ...(catalogTool as T),
            matchReason: 'pinned as important (always shown)',
            score: 10_000,
            rank: 0,
            important: true,
            alwaysShow: true,
            alwaysLoaded: alwaysLoadedSet.has(toolName),
            alwaysOn: Boolean(catalogTool.alwaysOn),
        });
    }

    const values = Array.from(merged.values());
    values.sort((left, right) => {
        if (left.alwaysLoaded !== right.alwaysLoaded) {
            return left.alwaysLoaded ? -1 : 1;
        }

        if (left.alwaysOn !== right.alwaysOn) {
            return left.alwaysOn ? -1 : 1;
        }

        if (left.important !== right.important) {
            return left.important ? -1 : 1;
        }

        return (left.rank ?? Number.MAX_SAFE_INTEGER) - (right.rank ?? Number.MAX_SAFE_INTEGER);
    });

    return values.map((item, index) => ({
        ...item,
        rank: item.rank && item.rank > 0 ? item.rank : index + 1,
    }));
}
