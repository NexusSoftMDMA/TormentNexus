import { spawnSync } from 'node:child_process';
import process from 'node:process';

type DiscoveryServerLike = {
    name: string;
    type?: 'STDIO' | 'SSE' | 'STREAMABLE_HTTP' | null;
    command?: string | null;
    args?: string[] | null;
    env?: Record<string, string> | null;
    url?: string | null;
    headers?: Record<string, string> | null;
    bearerToken?: string | null;
};

type DiscoveryPreflightOptions = {
    platform?: NodeJS.Platform;
    commandExists?: (command: string, platform: NodeJS.Platform) => boolean;
};

const SAMPLE_VALUE_PATTERNS = [
    /YOUR_[A-Z0-9_]+_HERE/i,
    /postgres:\/\/user:password@localhost:5432\/dbname/i,
    /Bearer\s+YOUR_[A-Z0-9_]+_HERE/i,
];

function hasSampleValue(value: string | null | undefined): boolean {
    if (!value) {
        return false;
    }

    const trimmed = value.trim();
    if (!trimmed) {
        return false;
    }

    return SAMPLE_VALUE_PATTERNS.some((pattern) => pattern.test(trimmed));
}

function findPlaceholderFields(server: DiscoveryServerLike): string[] {
    const fields: string[] = [];

    if (hasSampleValue(server.command)) {
        fields.push('command');
    }

    for (const [index, arg] of (server.args ?? []).entries()) {
        if (hasSampleValue(arg)) {
            fields.push(`args[${index}]`);
        }
    }

    for (const [key, value] of Object.entries(server.env ?? {})) {
        if (hasSampleValue(value)) {
            fields.push(`env.${key}`);
        }
    }

    if (hasSampleValue(server.url)) {
        fields.push('url');
    }

    for (const [key, value] of Object.entries(server.headers ?? {})) {
        if (hasSampleValue(value)) {
            fields.push(`headers.${key}`);
        }
    }

    if (hasSampleValue(server.bearerToken)) {
        fields.push('bearerToken');
    }

    return fields;
}

function commandExistsOnPath(command: string, platform: NodeJS.Platform): boolean {
    const lookup = platform === 'win32' ? 'where.exe' : 'which';
    const result = spawnSync(lookup, [command], {
        stdio: 'ignore',
        windowsHide: true,
    });

    return result.status === 0;
}

export function getDiscoveryPreflightFailure(
    server: DiscoveryServerLike,
    options: DiscoveryPreflightOptions = {},
): string | null {
    const placeholderFields = findPlaceholderFields(server);
    if (placeholderFields.length > 0) {
        const preview = placeholderFields.slice(0, 4).join(', ');
        const suffix = placeholderFields.length > 4 ? ', …' : '';
        return `Discovery skipped because ${server.name} still contains placeholder or sample configuration values (${preview}${suffix}). Update the config and try again.`;
    }

    if ((server.type ?? 'STDIO') !== 'STDIO') {
        return null;
    }

    const command = server.command?.trim();
    if (!command) {
        return null;
    }

    const platform = options.platform ?? process.platform;
    const exists = (options.commandExists ?? commandExistsOnPath)(command, platform);
    if (!exists) {
        return `Discovery skipped because command "${command}" is not available on PATH for ${server.name}. Install it or update the server command before retrying.`;
    }

    return null;
}
