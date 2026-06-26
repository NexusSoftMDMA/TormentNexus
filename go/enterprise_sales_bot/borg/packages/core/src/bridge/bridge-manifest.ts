export type BridgeClientType = 'browser-extension' | 'vscode-extension' | 'cli-adapter' | 'unknown';

export interface BridgeClientHello {
    type: 'TORMENTNEXUS_CLIENT_HELLO';
    clientType?: BridgeClientType;
    clientName?: string;
    version?: string;
    platform?: string;
    capabilities?: string[];
    hookPhases?: string[];
}

export interface RegisteredBridgeClient {
    clientId: string;
    clientType: BridgeClientType;
    clientName: string;
    version?: string;
    platform?: string;
    capabilities: string[];
    hookPhases: string[];
    connectedAt: number;
    lastSeenAt: number;
}

export interface BridgeManifest {
    protocolVersion: string;
    supportedClientTypes: BridgeClientType[];
    supportedCapabilities: string[];
    supportedHookPhases: string[];
    connectedClients: RegisteredBridgeClient[];
}

export const SUPPORTED_BRIDGE_CLIENT_TYPES: BridgeClientType[] = [
    'browser-extension',
    'vscode-extension',
    'cli-adapter',
    'unknown',
];

export const SUPPORTED_BRIDGE_CAPABILITIES = [
    'bridge.websocket',
    'memory.capture',
    'rag.ingest',
    'session.attach',
    'context.push',
    'chat.inject',
    'command.execute',
    'editor.selection.read',
    'terminal.buffer.read',
    'browser.page.capture',
    'browser.history.read',
    'browser.debug.cdp',
] as const;

export const SUPPORTED_BRIDGE_HOOK_PHASES = [
    'session.start',
    'session.end',
    'user.activity',
    'context.capture',
    'memory.capture',
    'chat.submit',
    'editor.selection',
    'terminal.output',
    'browser.page.absorb',
    'browser.chat.surface',
] as const;

function uniqueStrings(values: string[] | undefined): string[] {
    if (!Array.isArray(values)) {
        return [];
    }

    return [...new Set(values.map((value) => String(value).trim()).filter(Boolean))].sort((left, right) => left.localeCompare(right));
}

export function createDefaultBridgeClient(clientId: string, now = Date.now()): RegisteredBridgeClient {
    return {
        clientId,
        clientType: 'unknown',
        clientName: 'Unknown bridge client',
        capabilities: [],
        hookPhases: [],
        connectedAt: now,
        lastSeenAt: now,
    };
}

export function applyBridgeClientHello(
    existing: RegisteredBridgeClient,
    hello: Partial<BridgeClientHello>,
    now = Date.now(),
): RegisteredBridgeClient {
    const capabilities = Array.isArray(hello.capabilities)
        ? uniqueStrings(hello.capabilities)
        : existing.capabilities;
    const hookPhases = Array.isArray(hello.hookPhases)
        ? uniqueStrings(hello.hookPhases)
        : existing.hookPhases;

    return {
        ...existing,
        clientType: hello.clientType ?? existing.clientType,
        clientName: hello.clientName?.trim() || existing.clientName,
        version: hello.version?.trim() || existing.version,
        platform: hello.platform?.trim() || existing.platform,
        capabilities,
        hookPhases,
        lastSeenAt: now,
    };
}

export function buildBridgeManifest(clients: RegisteredBridgeClient[]): BridgeManifest {
    return {
        protocolVersion: '2026-03-10',
        supportedClientTypes: [...SUPPORTED_BRIDGE_CLIENT_TYPES],
        supportedCapabilities: [...SUPPORTED_BRIDGE_CAPABILITIES],
        supportedHookPhases: [...SUPPORTED_BRIDGE_HOOK_PHASES],
        connectedClients: [...clients].sort((left, right) => left.clientName.localeCompare(right.clientName)),
    };
}
