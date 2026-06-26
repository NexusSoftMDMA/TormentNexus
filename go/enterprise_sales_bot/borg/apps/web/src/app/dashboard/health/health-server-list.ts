interface RecordLike {
    [key: string]: unknown;
}

export interface HealthServerListItem {
    uuid: string;
    configKey: string;
    name: string;
    transportType: string;
}

function asRecord(value: unknown): RecordLike | null {
    return value && typeof value === 'object' && !Array.isArray(value)
        ? (value as RecordLike)
        : null;
}

function toNonEmptyString(value: unknown): string | null {
    return typeof value === 'string' && value.trim().length > 0
        ? value.trim()
        : null;
}

export function normalizeHealthServers(servers: unknown): HealthServerListItem[] {
    if (!Array.isArray(servers)) {
        return [];
    }

    return servers
        .map((entry, index) => {
            const record = asRecord(entry);
            if (!record) {
                return null;
            }

            const configKey = toNonEmptyString(record.configKey) ?? `unknown-server-${index + 1}`;
            const uuid = toNonEmptyString(record.uuid) ?? `${configKey}-${index + 1}`;
            const name = toNonEmptyString(record.name) ?? configKey;
            const transportType = toNonEmptyString(record.transportType) ?? 'unknown';

            return {
                uuid,
                configKey,
                name,
                transportType,
            } satisfies HealthServerListItem;
        })
        .filter((entry): entry is HealthServerListItem => Boolean(entry));
}

export function getConnectedServerKeys(mcpStatus: unknown): string[] {
    const status = asRecord(mcpStatus);
    const servers = status ? asRecord(status.servers) : null;

    return servers ? Object.keys(servers) : [];
}
