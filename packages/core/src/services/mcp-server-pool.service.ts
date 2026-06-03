import { ServerParameters } from "../types/mcp-admin/index.js";

import { configService } from "./config.service.js";
import { autoReconnectService } from "./auto-reconnect.service.js";
import { ConnectedClient, connecttormentnexusClient } from "./mcp-client.service.js";
import { serverErrorTracker } from "./server-error-tracker.service.js";
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from "../db/sqliteAvailability.js";

export interface McpServerPoolStatus {
    idle: number;
    active: number;
    activeSessionIds: string[];
    idleServerUuids: string[];
    currentActiveServerUuid: string | null;
    lastActiveServerSwitchAt: number | null;
}

export interface McpServerPoolLifecycleModes {
    lazySessionMode: boolean;
    singleActiveServerMode: boolean;
}

export interface McpServerPoolLifecycleEvent {
    id: number;
    timestamp: number;
    type:
        | 'session-created'
        | 'session-converted'
        | 'session-cleanup'
        | 'mode-updated'
        | 'single-active-prune'
        | 'server-crash'
        | 'active-server-switch';
    message: string;
    reasonCode?: string;
    sessionId?: string;
    serverUuid?: string;
    serverName?: string;
}

export class McpServerPool {
    // Singleton instance
    private static instance: McpServerPool | null = null;

    // Idle sessions: serverUuid -> ConnectedClient (no sessionId assigned yet)
    private idleSessions: Record<string, ConnectedClient> = {};

    // Active sessions: sessionId -> Record<serverUuid, ConnectedClient>
    private activeSessions: Record<string, Record<string, ConnectedClient>> = {};

    // Mapping: sessionId -> Set<serverUuid> for cleanup tracking
    private sessionToServers: Record<string, Set<string>> = {};

    // Session creation timestamps: sessionId -> timestamp
    private sessionTimestamps: Record<string, number> = {};

    // Server parameters cache: serverUuid -> ServerParameters
    private serverParamsCache: Record<string, ServerParameters> = {};

    // Track ongoing idle session creation to prevent duplicates
    private creatingIdleSessions: Set<string> = new Set();

    // Session cleanup timer
    private cleanupTimer: NodeJS.Timeout | null = null;
    private sqliteCleanupUnavailableLogged = false;

    // Default number of idle sessions per server UUID
    private readonly defaultIdleCount: number;
    // Lazy mode keeps downstream servers stopped until a tool is actually executed.
    private lazySessionMode: boolean;
    // Single-active mode ensures only one downstream server process is active per session.
    private singleActiveServerMode: boolean;
    // Bounded timeline of lifecycle decisions for dashboard/operator observability.
    private lifecycleEvents: McpServerPoolLifecycleEvent[] = [];
    private lifecycleEventSequence = 0;
    private readonly maxLifecycleEvents = 200;
    private currentActiveServerUuid: string | null = null;
    private lastActiveServerSwitchAt: number | null = null;

    private constructor(defaultIdleCount: number = 1) {
        this.defaultIdleCount = defaultIdleCount;
        this.lazySessionMode = process.env.TORMENTNEXUS_MCP_LAZY_SESSIONS !== 'false';
        this.singleActiveServerMode = process.env.TORMENTNEXUS_MCP_SINGLE_ACTIVE_SERVER !== 'false';
        this.recordLifecycleEvent({
            type: 'mode-updated',
            reasonCode: 'mode-initialized',
            message: `Initialized lifecycle modes (lazy=${this.lazySessionMode ? 'on' : 'off'}, single-active=${this.singleActiveServerMode ? 'on' : 'off'})`,
        });
        this.startCleanupTimer();
    }

    private recordLifecycleEvent(event: Omit<McpServerPoolLifecycleEvent, 'id' | 'timestamp'>): void {
        const next: McpServerPoolLifecycleEvent = {
            ...event,
            id: ++this.lifecycleEventSequence,
            timestamp: Date.now(),
        };

        this.lifecycleEvents.push(next);
        if (this.lifecycleEvents.length > this.maxLifecycleEvents) {
            this.lifecycleEvents.splice(0, this.lifecycleEvents.length - this.maxLifecycleEvents);
        }
    }

    private markActiveServer(serverUuid: string, sessionId: string): void {
        if (this.currentActiveServerUuid === serverUuid) {
            return;
        }

        const serverName = this.serverParamsCache[serverUuid]?.name;
        const serverLabel = this.serverParamsCache[serverUuid]?.name
            ? `${this.serverParamsCache[serverUuid].name} (${serverUuid})`
            : serverUuid;

        this.currentActiveServerUuid = serverUuid;
        this.lastActiveServerSwitchAt = Date.now();
        this.recordLifecycleEvent({
            type: 'active-server-switch',
            reasonCode: 'focus-shift',
            serverUuid,
            serverName,
            sessionId,
            message: `Active downstream focus switched to ${serverLabel} (session ${sessionId})`,
        });
    }

    private recomputeActiveServerFocus(): void {
        const activeServerUuids = new Set<string>();
        Object.values(this.activeSessions).forEach((sessionServers) => {
            Object.keys(sessionServers).forEach((serverUuid) => activeServerUuids.add(serverUuid));
        });

        if (activeServerUuids.size === 0) {
            this.currentActiveServerUuid = null;
            return;
        }

        if (this.currentActiveServerUuid && activeServerUuids.has(this.currentActiveServerUuid)) {
            return;
        }

        const next = activeServerUuids.values().next().value ?? null;
        this.currentActiveServerUuid = next;
    }

    /**
     * Get the singleton instance
     */
    static getInstance(defaultIdleCount: number = 1): McpServerPool {
        if (!McpServerPool.instance) {
            McpServerPool.instance = new McpServerPool(defaultIdleCount);
        }
        return McpServerPool.instance;
    }

    /**
     * Get or create a session for a specific MCP server
     */
    async getSession(
        sessionId: string,
        serverUuid: string,
        params: ServerParameters,
        namespaceUuid?: string,
    ): Promise<ConnectedClient | undefined> {
        // Update server params cache
        this.serverParamsCache[serverUuid] = params;

        // Check if we already have an active session for this sessionId and server
        if (this.activeSessions[sessionId]?.[serverUuid]) {
            this.markActiveServer(serverUuid, sessionId);
            this.recordLifecycleEvent({
                type: 'session-converted',
                reasonCode: 'reuse-active',
                sessionId,
                serverUuid,
                serverName: params.name,
                message: `Reused active session binding for server ${serverUuid} in session ${sessionId}`,
            });
            return this.activeSessions[sessionId][serverUuid];
        }

        if (this.singleActiveServerMode) {
            await this.cleanupOtherActiveServersGlobal(serverUuid);
        }

        // Initialize session if it doesn't exist
        if (!this.activeSessions[sessionId]) {
            this.activeSessions[sessionId] = {};
            this.sessionToServers[sessionId] = new Set();
            this.sessionTimestamps[sessionId] = Date.now();
        }

        // Check if we have an idle session for this server that we can convert
        const idleClient = this.idleSessions[serverUuid];
        if (idleClient) {
            // Convert idle session to active session
            delete this.idleSessions[serverUuid];
            this.activeSessions[sessionId][serverUuid] = idleClient;
            this.sessionToServers[sessionId].add(serverUuid);
            this.markActiveServer(serverUuid, sessionId);
            this.recordLifecycleEvent({
                type: 'session-converted',
                reasonCode: 'promote-idle',
                sessionId,
                serverUuid,
                serverName: params.name,
                message: `Promoted idle server ${serverUuid} into active session ${sessionId}`,
            });

            console.log(
                `Converted idle session to active for server ${serverUuid}, session ${sessionId}`,
            );

            // In eager mode we prewarm a replacement idle session.
            if (!this.lazySessionMode) {
                this.createIdleSessionAsync(serverUuid, params, namespaceUuid);
            }

            return idleClient;
        }

        // No idle session available, create a new connection
        const newClient = await this.createNewConnection(params, namespaceUuid);
        if (!newClient) {
            return undefined;
        }

        this.activeSessions[sessionId][serverUuid] = newClient;
        this.sessionToServers[sessionId].add(serverUuid);
        this.markActiveServer(serverUuid, sessionId);
        this.recordLifecycleEvent({
            type: 'session-created',
            reasonCode: 'create-new',
            sessionId,
            serverUuid,
            serverName: params.name,
            message: `Created new active downstream session for server ${serverUuid} in ${sessionId}`,
        });

        console.log(
            `Created new active session for server ${serverUuid}, session ${sessionId}`,
        );

        // In eager mode we prewarm one idle session for future reuse.
        if (!this.lazySessionMode) {
            this.createIdleSessionAsync(serverUuid, params, namespaceUuid);
        }

        return newClient;
    }

    private async cleanupOtherActiveServersGlobal(keepServerUuid: string): Promise<void> {
        const cleanupOperations: Array<Promise<void>> = [];
        let cleanedActive = 0;
        let cleanedIdle = 0;

        Object.entries(this.activeSessions).forEach(([sessionId, sessionServers]) => {
            Object.entries(sessionServers)
                .filter(([uuid]) => uuid !== keepServerUuid)
                .forEach(([uuid, client]) => {
                    cleanupOperations.push((async () => {
                        try {
                            await client.cleanup();
                        } catch (error) {
                            console.error(`Error cleaning up stale active server ${uuid} for session ${sessionId}:`, error);
                        }
                        delete sessionServers[uuid];
                        this.sessionToServers[sessionId]?.delete(uuid);
                        cleanedActive += 1;
                    })());
                });
        });

        if (cleanupOperations.length > 0) {
            await Promise.allSettled(cleanupOperations);
        }

        Object.entries(this.idleSessions)
            .filter(([uuid]) => uuid !== keepServerUuid)
            .forEach(([uuid, client]) => {
                cleanupOperations.push((async () => {
                    try {
                        await client.cleanup();
                    } catch (error) {
                        console.error(`Error cleaning up stale idle server ${uuid}:`, error);
                    }
                    delete this.idleSessions[uuid];
                    cleanedIdle += 1;
                })());
            });

        if (cleanupOperations.length > 0) {
            await Promise.allSettled(cleanupOperations);
            this.recordLifecycleEvent({
                type: 'single-active-prune',
                reasonCode: 'single-active-policy',
                serverUuid: keepServerUuid,
                serverName: this.serverParamsCache[keepServerUuid]?.name,
                message: `Pruned stale downstream sessions for single-active policy (kept=${keepServerUuid}, activePruned=${cleanedActive}, idlePruned=${cleanedIdle})`,
            });
            this.recomputeActiveServerFocus();
        }
    }

    /**
     * Create a new connection for a server
     */
    private async createNewConnection(
        params: ServerParameters,
        namespaceUuid?: string,
    ): Promise<ConnectedClient | undefined> {
        console.log(
            `Creating new connection for server ${params.name} (${params.uuid}) with namespace: ${namespaceUuid || "none"}`,
        );

        const connectedClient = await connecttormentnexusClient(
            params,
            (exitCode, signal) => {
                console.log(
                    `Crash handler callback called for server ${params.name} (${params.uuid}) with namespace: ${namespaceUuid || "none"}`,
                );

                // Handle process crash - always set up crash handler
                if (namespaceUuid) {
                    // If we have a namespace context, use it
                    this.handleServerCrash(
                        params.uuid,
                        namespaceUuid,
                        exitCode,
                        signal,
                    ).catch((error) => {
                        console.error(
                            `Error handling server crash for ${params.uuid} in ${namespaceUuid}:`,
                            error,
                        );
                    });
                } else {
                    // If no namespace context, still track the crash globally
                    this.handleServerCrashWithoutNamespace(
                        params.uuid,
                        exitCode,
                        signal,
                    ).catch((error) => {
                        console.error(
                            `Error handling server crash for ${params.uuid} (no namespace):`,
                            error,
                        );
                    });
                }
            },
        );
        if (!connectedClient) {
            return undefined;
        }

        return connectedClient;
    }

    /**
     * Create an idle session for a server (blocking version for initial setup)
     */
    private async createIdleSession(
        serverUuid: string,
        params: ServerParameters,
        namespaceUuid?: string,
    ): Promise<void> {
        if (this.lazySessionMode) {
            return;
        }

        // Don't create if we already have an idle session for this server
        if (this.idleSessions[serverUuid]) {
            return;
        }

        const newClient = await this.createNewConnection(params, namespaceUuid);
        if (newClient) {
            this.idleSessions[serverUuid] = newClient;
            console.log(`Created idle session for server ${serverUuid}`);
        }
    }

    /**
     * Create an idle session for a server asynchronously (non-blocking)
     */
    private createIdleSessionAsync(
        serverUuid: string,
        params: ServerParameters,
        namespaceUuid?: string,
    ): void {
        if (this.lazySessionMode) {
            return;
        }

        // Don't create if we already have an idle session or are already creating one
        if (
            this.idleSessions[serverUuid] ||
            this.creatingIdleSessions.has(serverUuid)
        ) {
            return;
        }

        // Mark that we're creating an idle session for this server
        this.creatingIdleSessions.add(serverUuid);

        // Create the session in the background (fire and forget)
        this.createNewConnection(params, namespaceUuid)
            .then((newClient) => {
                if (newClient && !this.idleSessions[serverUuid]) {
                    this.idleSessions[serverUuid] = newClient;
                    console.log(
                        `Created background idle session for server [${params.name}] ${serverUuid}`,
                    );
                } else if (newClient) {
                    // We already have an idle session, cleanup the extra one
                    newClient.cleanup().catch((error) => {
                        console.error(
                            `Error cleaning up extra idle session for ${serverUuid}:`,
                            error,
                        );
                    });
                }
            })
            .catch((error) => {
                console.error(
                    `Error creating background idle session for ${serverUuid}:`,
                    error,
                );
            })
            .finally(() => {
                // Remove from creating set
                this.creatingIdleSessions.delete(serverUuid);
            });
    }

    /**
     * Ensure idle sessions exist for all servers
     */
    async ensureIdleSessions(
        serverParams: Record<string, ServerParameters>,
        namespaceUuid?: string,
    ): Promise<void> {
        const promises = Object.entries(serverParams).map(
            async ([uuid, params]) => {
                if (!this.idleSessions[uuid]) {
                    await this.createIdleSession(uuid, params, namespaceUuid);
                }
            },
        );

        await Promise.allSettled(promises);
    }

    /**
     * Cleanup a session by sessionId
     */
    async cleanupSession(sessionId: string): Promise<void> {
        const activeSession = this.activeSessions[sessionId];
        if (!activeSession) {
            return;
        }

        // Cleanup all connections for this session
        await Promise.allSettled(
            Object.entries(activeSession).map(async ([_serverUuid, client]) => {
                await client.cleanup();
            }),
        );

        // Remove from active sessions
        delete this.activeSessions[sessionId];

        // Clean up session timestamp
        delete this.sessionTimestamps[sessionId];

        // Clean up session to servers mapping
        const serverUuids = this.sessionToServers[sessionId];
        if (serverUuids) {
            if (!this.lazySessionMode) {
                // For each server this session was using, create new idle sessions if needed (ASYNC - NON-BLOCKING)
                Array.from(serverUuids).forEach((serverUuid) => {
                    const params = this.serverParamsCache[serverUuid];
                    if (params) {
                        // Note: We don't have namespaceUuid here, so we can't track crashes properly
                        // This is a limitation of the current design - we'll need to pass namespaceUuid from the caller
                        this.createIdleSessionAsync(serverUuid, params);
                    }
                });
            }

            delete this.sessionToServers[sessionId];
        }

        this.recordLifecycleEvent({
            type: 'session-cleanup',
            reasonCode: 'cleanup-request',
            sessionId,
            message: `Cleaned up downstream session ${sessionId}`,
        });
        this.recomputeActiveServerFocus();

        console.log(`Cleaned up MCP server pool session ${sessionId}`);
    }

    /**
     * Cleanup all sessions
     */
    async cleanupAll(): Promise<void> {
        // Cleanup all active sessions
        const activeSessionIds = Object.keys(this.activeSessions);
        await Promise.allSettled(
            activeSessionIds.map((sessionId) => this.cleanupSession(sessionId)),
        );

        // Cleanup all idle sessions
        await Promise.allSettled(
            Object.entries(this.idleSessions).map(async ([_uuid, client]) => {
                await client.cleanup();
            }),
        );

        // Clear all state
        this.idleSessions = {};
        this.activeSessions = {};
        this.sessionToServers = {};
        this.sessionTimestamps = {};
        this.serverParamsCache = {};
        this.creatingIdleSessions.clear();
        this.currentActiveServerUuid = null;

        // Clear cleanup timer
        if (this.cleanupTimer) {
            clearInterval(this.cleanupTimer);
            this.cleanupTimer = null;
        }

        console.log("Cleaned up all MCP server pool sessions");
    }

    /**
     * Get pool status for monitoring
     */
    getPoolStatus(): McpServerPoolStatus {
        const idle = Object.keys(this.idleSessions).length;
        const active = Object.keys(this.activeSessions).reduce(
            (total, sessionId) =>
                total + Object.keys(this.activeSessions[sessionId]).length,
            0,
        );

        return {
            idle,
            active,
            activeSessionIds: Object.keys(this.activeSessions),
            idleServerUuids: Object.keys(this.idleSessions),
            currentActiveServerUuid: this.currentActiveServerUuid,
            lastActiveServerSwitchAt: this.lastActiveServerSwitchAt,
        };
    }

    getLifecycleModes(): McpServerPoolLifecycleModes {
        return {
            lazySessionMode: this.lazySessionMode,
            singleActiveServerMode: this.singleActiveServerMode,
        };
    }

    async setLifecycleModes(next: Partial<McpServerPoolLifecycleModes>): Promise<McpServerPoolLifecycleModes> {
        const previousLazyMode = this.lazySessionMode;
        const previousSingleActiveMode = this.singleActiveServerMode;

        if (typeof next.lazySessionMode === 'boolean') {
            this.lazySessionMode = next.lazySessionMode;
        }

        if (typeof next.singleActiveServerMode === 'boolean') {
            this.singleActiveServerMode = next.singleActiveServerMode;
        }

        // If lazy mode has just been enabled, proactively drain idle sessions.
        if (!previousLazyMode && this.lazySessionMode) {
            await Promise.allSettled(
                Object.entries(this.idleSessions).map(async ([serverUuid, client]) => {
                    try {
                        await client.cleanup();
                    } catch (error) {
                        console.error(`Error cleaning up idle session while enabling lazy mode for ${serverUuid}:`, error);
                    }
                    delete this.idleSessions[serverUuid];
                }),
            );
        }

        if (previousLazyMode !== this.lazySessionMode || previousSingleActiveMode !== this.singleActiveServerMode) {
            this.recordLifecycleEvent({
                type: 'mode-updated',
                reasonCode: 'mode-change',
                message: `Updated lifecycle modes (lazy=${this.lazySessionMode ? 'on' : 'off'}, single-active=${this.singleActiveServerMode ? 'on' : 'off'})`,
            });
        }

        return this.getLifecycleModes();
    }

    getLifecycleEvents(limit: number = 25): McpServerPoolLifecycleEvent[] {
        const normalizedLimit = Math.max(1, Math.min(limit, this.maxLifecycleEvents));
        return this.lifecycleEvents.slice(-normalizedLimit).reverse();
    }

    /**
     * Get active session connections for a specific session (for debugging/monitoring)
     */
    getSessionConnections(
        sessionId: string,
    ): Record<string, ConnectedClient> | undefined {
        return this.activeSessions[sessionId];
    }

    /**
     * Get all active session IDs (for debugging/monitoring)
     */
    getActiveSessionIds(): string[] {
        return Object.keys(this.activeSessions);
    }

    /**
     * Invalidate and refresh idle session for a specific server
     * This should be called when a server's parameters (command, args, etc.) change
     */
    async invalidateIdleSession(
        serverUuid: string,
        params: ServerParameters,
        namespaceUuid?: string,
    ): Promise<void> {
        console.log(`Invalidating idle session for server ${serverUuid}`);

        // Update server params cache
        this.serverParamsCache[serverUuid] = params;

        // Cleanup existing idle session if it exists
        const existingIdleSession = this.idleSessions[serverUuid];
        if (existingIdleSession) {
            try {
                await existingIdleSession.cleanup();
                console.log(
                    `Cleaned up existing idle session for server ${serverUuid}`,
                );
            } catch (error) {
                console.error(
                    `Error cleaning up existing idle session for server ${serverUuid}:`,
                    error,
                );
            }
            delete this.idleSessions[serverUuid];
        }

        // Remove from creating set if it's in progress
        this.creatingIdleSessions.delete(serverUuid);

        // Create a new idle session with updated parameters
        await this.createIdleSession(serverUuid, params, namespaceUuid);
    }

    /**
     * Invalidate and refresh idle sessions for multiple servers
     */
    async invalidateIdleSessions(
        serverParams: Record<string, ServerParameters>,
        namespaceUuid?: string,
    ): Promise<void> {
        const promises = Object.entries(serverParams).map(([serverUuid, params]) =>
            this.invalidateIdleSession(serverUuid, params, namespaceUuid),
        );

        await Promise.allSettled(promises);
    }

    /**
     * Clean up idle session for a specific server without creating a new one
     * This should be called when a server is being deleted
     */
    async cleanupIdleSession(serverUuid: string): Promise<void> {
        console.log(`Cleaning up idle session for server ${serverUuid}`);

        // Cleanup existing idle session if it exists
        const existingIdleSession = this.idleSessions[serverUuid];
        if (existingIdleSession) {
            try {
                await existingIdleSession.cleanup();
                console.log(`Cleaned up idle session for server ${serverUuid}`);
            } catch (error) {
                console.error(
                    `Error cleaning up idle session for server ${serverUuid}:`,
                    error,
                );
            }
            delete this.idleSessions[serverUuid];
        }

        // Remove from creating set if it's in progress
        this.creatingIdleSessions.delete(serverUuid);

        // Remove from server params cache
        delete this.serverParamsCache[serverUuid];
    }

    /**
     * Ensure idle session exists for a newly created server
     * This should be called when a new server is created
     */
    async ensureIdleSessionForNewServer(
        serverUuid: string,
        params: ServerParameters,
        namespaceUuid?: string,
    ): Promise<void> {
        console.log(`Ensuring idle session exists for new server ${serverUuid}`);

        // Update server params cache
        this.serverParamsCache[serverUuid] = params;

        // Only create if we don't already have one
        if (
            !this.idleSessions[serverUuid] &&
            !this.creatingIdleSessions.has(serverUuid)
        ) {
            await this.createIdleSession(serverUuid, params, namespaceUuid);
        }
    }

    /**
     * Handle server process crash
     */
    async handleServerCrash(
        serverUuid: string,
        namespaceUuid: string,
        exitCode: number | null,
        signal: string | null,
    ): Promise<void> {
        console.warn(
            `Handling server crash for ${serverUuid} in namespace ${namespaceUuid}`,
        );

        await serverErrorTracker.recordServerCrash(serverUuid, exitCode, signal);
        await this.cleanupServerSessions(serverUuid);
        this.recordLifecycleEvent({
            type: 'server-crash',
            reasonCode: 'process-exit',
            serverUuid,
            serverName: this.serverParamsCache[serverUuid]?.name,
            message: `Detected downstream server crash for ${serverUuid} in namespace ${namespaceUuid} (exit=${exitCode ?? 'null'}, signal=${signal ?? 'null'})`,
        });

        const serverName =
            this.serverParamsCache[serverUuid]?.name ?? `server-${serverUuid}`;
        autoReconnectService.scheduleReconnection(serverUuid, serverName, "crash");
    }

    /**
     * Handle server process crash without namespace context
     * This is used when servers are created without a specific namespace
     */
    async handleServerCrashWithoutNamespace(
        serverUuid: string,
        exitCode: number | null,
        signal: string | null,
    ): Promise<void> {
        console.warn(
            `Handling server crash for ${serverUuid} (no namespace context)`,
        );

        console.log(`Recording crash for server ${serverUuid}`);
        await serverErrorTracker.recordServerCrash(serverUuid, exitCode, signal);
        await this.cleanupServerSessions(serverUuid);
        this.recordLifecycleEvent({
            type: 'server-crash',
            reasonCode: 'process-exit',
            serverUuid,
            serverName: this.serverParamsCache[serverUuid]?.name,
            message: `Detected downstream server crash for ${serverUuid} (exit=${exitCode ?? 'null'}, signal=${signal ?? 'null'})`,
        });

        const serverName =
            this.serverParamsCache[serverUuid]?.name ?? `server-${serverUuid}`;
        autoReconnectService.scheduleReconnection(serverUuid, serverName, "crash");
    }

    /**
     * Clean up all sessions for a specific server
     */
    private async cleanupServerSessions(serverUuid: string): Promise<void> {
        // Clean up idle session
        const idleSession = this.idleSessions[serverUuid];
        if (idleSession) {
            try {
                await idleSession.cleanup();
                console.log(`Cleaned up idle session for crashed server ${serverUuid}`);
            } catch (error) {
                console.error(
                    `Error cleaning up idle session for crashed server ${serverUuid}:`,
                    error,
                );
            }
            delete this.idleSessions[serverUuid];
        }

        // Clean up active sessions that use this server
        for (const [sessionId, sessionServers] of Object.entries(
            this.activeSessions,
        )) {
            if (sessionServers[serverUuid]) {
                try {
                    await sessionServers[serverUuid].cleanup();
                    console.log(
                        `Cleaned up active session ${sessionId} for crashed server ${serverUuid}`,
                    );
                } catch (error) {
                    console.error(
                        `Error cleaning up active session ${sessionId} for crashed server ${serverUuid}:`,
                        error,
                    );
                }
                delete sessionServers[serverUuid];
                this.sessionToServers[sessionId]?.delete(serverUuid);
            }
        }

        // Remove from creating set
        this.creatingIdleSessions.delete(serverUuid);
        this.recomputeActiveServerFocus();
    }

    /**
     * Check if a server is in error state
     */
    async isServerInErrorState(serverUuid: string): Promise<boolean> {
        return await serverErrorTracker.isServerInErrorState(serverUuid);
    }

    /**
     * Reset error state for a server (e.g., after manual recovery)
     */
    async resetServerErrorState(serverUuid: string): Promise<void> {
        // Reset crash attempts and error status
        await serverErrorTracker.resetServerErrorState(serverUuid);

        console.log(`Reset error state for server ${serverUuid}`);
    }

    /**
     * Start the automatic cleanup timer for expired sessions
     */
    private startCleanupTimer(): void {
        // Check for expired sessions every 5 minutes
        this.cleanupTimer = setInterval(
            async () => {
                await this.cleanupExpiredSessions();
            },
            5 * 60 * 1000,
        ); // 5 minutes
    }

    /**
     * Clean up expired sessions based on session lifetime setting
     * Preserves idle sessions for warmup
     */
    private async cleanupExpiredSessions(): Promise<void> {
        try {
            const sessionLifetime = await configService.getSessionLifetime();
            this.sqliteCleanupUnavailableLogged = false;

            // If session lifetime is null, sessions are infinite - skip cleanup
            if (sessionLifetime === null) {
                return;
            }

            const now = Date.now();
            const expiredSessionIds: string[] = [];

            // Find expired active sessions (never clean up idle sessions as they are for warmup)
            for (const [sessionId, timestamp] of Object.entries(
                this.sessionTimestamps,
            )) {
                // Only clean up active sessions, never idle sessions
                if (this.activeSessions[sessionId] && now - timestamp > sessionLifetime) {
                    expiredSessionIds.push(sessionId);
                }
            }

            // Clean up expired sessions
            if (expiredSessionIds.length > 0) {
                console.log(
                    `Cleaning up ${expiredSessionIds.length} expired MCP server pool sessions: ${expiredSessionIds.join(", ")}`,
                );

                await Promise.allSettled(
                    expiredSessionIds.map((sessionId) => this.cleanupSession(sessionId)),
                );
            }
        } catch (error) {
            if (isSqliteUnavailableError(error)) {
                if (!this.sqliteCleanupUnavailableLogged) {
                    console.warn(formatOptionalSqliteFailure(
                        "[McpServerPool] Skipping automatic session cleanup",
                        error,
                    ));
                    this.sqliteCleanupUnavailableLogged = true;
                }
                return;
            }
            console.error("Error during automatic session cleanup:", error);
        }
    }

    /**
     * Get session age in milliseconds
     */
    getSessionAge(sessionId: string): number | undefined {
        const timestamp = this.sessionTimestamps[sessionId];
        return timestamp ? Date.now() - timestamp : undefined;
    }

    /**
     * Check if a session is expired
     */
    async isSessionExpired(sessionId: string): Promise<boolean> {
        const age = this.getSessionAge(sessionId);
        if (age === undefined) return false;

        const sessionLifetime = await configService.getSessionLifetime();
        if (sessionLifetime === null) {
            return false;
        }

        return age > sessionLifetime;
    }
}

// Create a singleton instance
export const mcpServerPool = McpServerPool.getInstance();
