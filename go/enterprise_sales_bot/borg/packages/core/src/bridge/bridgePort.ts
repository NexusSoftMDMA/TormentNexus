const DEFAULT_TORMENTNEXUS_BRIDGE_PORT = 3001;

function normalizePort(value: string | undefined): number | null {
    if (typeof value !== 'string') {
        return null;
    }

    const trimmed = value.trim();
    if (!/^\d+$/u.test(trimmed)) {
        return null;
    }

    const parsed = Number(trimmed);
    if (!Number.isInteger(parsed) || parsed <= 0 || parsed > 65535) {
        return null;
    }

    return parsed;
}

export function getDefaultBridgePort(): number {
    return DEFAULT_TORMENTNEXUS_BRIDGE_PORT;
}

export function resolveBridgePort(env: NodeJS.ProcessEnv = process.env): number {
    return normalizePort(env.TORMENTNEXUS_BRIDGE_PORT)
        ?? normalizePort(env.TORMENTNEXUS_CORE_BRIDGE_PORT)
        ?? DEFAULT_TORMENTNEXUS_BRIDGE_PORT;
}

export function getBridgeHttpBase(port: number = resolveBridgePort()): string {
    return `http://127.0.0.1:${port}`;
}

export function getBridgeHealthUrl(port: number = resolveBridgePort()): string {
    return `${getBridgeHttpBase(port)}/health`;
}

export function getBridgeToolExecuteUrl(port: number = resolveBridgePort()): string {
    return `${getBridgeHttpBase(port)}/tool/execute`;
}

export function getBridgeStreamUrl(port: number = resolveBridgePort()): string {
    return `${getBridgeHttpBase(port)}/api/mesh/stream`;
}

export function getBridgeWebSocketUrl(port: number = resolveBridgePort()): string {
    return `ws://127.0.0.1:${port}`;
}
