/**
 * tormentnexus MCP Session Service
 *
 * Enhanced session lifecycle management for MCP servers:
 * - Auto-start servers on system boot
 * - Auto-restart crashed servers
 * - Keep-alive heartbeat monitoring
 * - Latency tracking and performance metrics
 * - Client registry for multi-client sessions
 *
 * Works with:
 * - McpManager (for server lifecycle)
 * - ServerRegistryService (for health checks)
 * - DatabaseManager (for persistence)
 */
import { EventEmitter } from 'events';
import { DatabaseManager } from '../db/index.js';
// ============================================
// McpSessionService Class
// ============================================
export class McpSessionService extends EventEmitter {
    dataDir;
    static instance;
    db;
    sessions = new Map();
    metrics = new Map();
    heartbeatInterval;
    latencyHistory = new Map();
    config = {
        autoStart: true,
        autoRestart: true,
        restartDelay: 5000,
        restartMaxAttempts: 3,
        heartbeatInterval: 30000,
        heartbeatTimeout: 120000
    };
    constructor(dataDir) {
        super();
        this.dataDir = dataDir;
        this.db = DatabaseManager.getInstance(dataDir);
    }
    static getInstance(dataDir) {
        if (!McpSessionService.instance) {
            McpSessionService.instance = new McpSessionService(dataDir);
        }
        return McpSessionService.instance;
    }
    // ============================================
    // Lifecycle Management
    // ============================================
    /**
     * Initialize session service and start auto-start servers
     */
    async initialize() {
        console.log('[McpSessionService] Initializing session service...');
        await this.loadPersistedSessions();
        await this.startAutoStartServers();
        this.startHeartbeat();
        console.log(`[McpSessionService] Initialized with ${this.sessions.size} sessions`);
    }
    /**
     * Load persisted sessions from database
     */
    async loadPersistedSessions() {
        const servers = this.db.getAllMcpServers({ enabled: true });
        for (const server of servers) {
            const session = {
                sessionId: this.generateSessionId(),
                serverId: server.id,
                serverName: server.name,
                status: 'stopped',
                startTime: undefined,
                lastHeartbeat: undefined,
                clients: [],
                autoRestart: this.config.autoRestart,
                restartDelay: this.config.restartDelay,
                errorCount: 0
            };
            this.sessions.set(server.id, session);
            this.initializeMetrics(server.name);
        }
    }
    /**
     * Start servers marked for auto-start
     */
    async startAutoStartServers() {
        const sessionsToStart = Array.from(this.sessions.values())
            .filter(s => s.autoRestart);
        for (const session of sessionsToStart) {
            try {
                await this.startSession(session.serverId);
                console.log(`[McpSessionService] Auto-started: ${session.serverName}`);
            }
            catch (e) {
                console.error(`[McpSessionService] Failed to auto-start ${session.serverName}:`, e);
            }
        }
    }
    /**
     * Start a session for a server
     */
    async startSession(serverId) {
        const session = this.sessions.get(serverId);
        if (!session) {
            throw new Error(`Session not found for server: ${serverId}`);
        }
        if (session.status === 'running') {
            return session;
        }
        session.status = 'starting';
        session.startTime = Date.now();
        session.errorCount = 0;
        this.emit('session:starting', session);
        return session;
    }
    /**
     * Mark session as running
     */
    markSessionRunning(serverId, clientId) {
        const session = this.sessions.get(serverId);
        if (!session)
            return null;
        session.status = 'running';
        session.lastHeartbeat = Date.now();
        if (clientId && !session.clients.includes(clientId)) {
            session.clients.push(clientId);
        }
        this.emit('session:started', session);
        return session;
    }
    /**
     * Pause a session
     */
    pauseSession(serverId) {
        const session = this.sessions.get(serverId);
        if (!session)
            return null;
        session.status = 'paused';
        this.emit('session:paused', session);
        return session;
    }
    /**
     * Resume a paused session
     */
    resumeSession(serverId) {
        const session = this.sessions.get(serverId);
        if (!session)
            return null;
        session.status = 'running';
        session.lastHeartbeat = Date.now();
        this.emit('session:resumed', session);
        return session;
    }
    /**
     * Stop a session
     */
    stopSession(serverId) {
        const session = this.sessions.get(serverId);
        if (!session)
            return null;
        session.status = 'stopped';
        session.startTime = undefined;
        session.clients = [];
        this.emit('session:stopped', session);
        return session;
    }
    /**
     * Terminate a session (force cleanup)
     */
    terminateSession(serverId, reason) {
        const session = this.sessions.get(serverId);
        if (!session)
            return null;
        session.status = 'stopped';
        session.lastError = reason;
        this.emit('session:terminated', { session, reason });
        return session;
    }
    // ============================================
    // Auto-Restart
    // ============================================
    /**
     * Handle server error and potentially restart
     */
    async handleServerError(serverId, error) {
        const session = this.sessions.get(serverId);
        if (!session || !session.autoRestart) {
            return;
        }
        session.errorCount++;
        session.lastError = error;
        session.status = 'error';
        this.emit('session:error', { session, error });
        if (session.errorCount <= this.config.restartMaxAttempts) {
            console.log(`[McpSessionService] Restarting ${session.serverName} (attempt ${session.errorCount})`);
            await new Promise(resolve => setTimeout(resolve, session.restartDelay));
            await this.startSession(serverId);
        }
        else {
            console.error(`[McpSessionService] Max restart attempts reached for ${session.serverName}`);
            session.autoRestart = false;
        }
    }
    /**
     * Reset error count for a server
     */
    resetErrorCount(serverId) {
        const session = this.sessions.get(serverId);
        if (session) {
            session.errorCount = 0;
            session.lastError = undefined;
        }
    }
    // ============================================
    // Health & Heartbeat
    // ============================================
    /**
     * Start heartbeat monitoring
     */
    startHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
        }
        this.heartbeatInterval = setInterval(() => {
            this.performHeartbeat();
        }, this.config.heartbeatInterval);
    }
    /**
     * Stop heartbeat monitoring
     */
    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = undefined;
        }
    }
    /**
     * Perform heartbeat check on all running sessions
     */
    performHeartbeat() {
        const now = Date.now();
        const runningSessions = Array.from(this.sessions.values())
            .filter(s => s.status === 'running');
        for (const session of runningSessions) {
            if (!session.lastHeartbeat) {
                continue;
            }
            const elapsed = now - session.lastHeartbeat;
            if (elapsed > this.config.heartbeatTimeout) {
                console.warn(`[McpSessionService] Timeout for ${session.serverName} (last heartbeat: ${elapsed}ms ago)`);
                if (session.autoRestart) {
                    this.handleServerError(session.serverId, 'Heartbeat timeout');
                }
                else {
                    session.status = 'stopped';
                    this.emit('session:timeout', session);
                }
            }
        }
    }
    /**
     * Update heartbeat for a server
     */
    updateHeartbeat(serverId, latencyMs) {
        const session = this.sessions.get(serverId);
        if (!session)
            return;
        session.lastHeartbeat = Date.now();
        if (latencyMs !== undefined) {
            this.recordLatency(session.serverName, latencyMs);
            session.latencyMs = latencyMs;
        }
        if (session.status === 'starting') {
            session.status = 'running';
            this.emit('session:started', session);
        }
    }
    // ============================================
    // Latency & Metrics
    // ============================================
    /**
     * Record latency measurement
     */
    recordLatency(serverName, latencyMs) {
        const history = this.latencyHistory.get(serverName) || [];
        history.push(latencyMs);
        if (history.length > 100) {
            history.shift();
        }
        this.latencyHistory.set(serverName, history);
    }
    /**
     * Initialize metrics for a server
     */
    initializeMetrics(serverName) {
        if (!this.metrics.has(serverName)) {
            this.metrics.set(serverName, []);
        }
    }
    /**
     * Get performance metrics for a server
     */
    getMetrics(serverName) {
        const history = this.latencyHistory.get(serverName);
        if (!history || history.length === 0) {
            return null;
        }
        const session = Array.from(this.sessions.values())
            .find(s => s.serverName === serverName);
        if (!session || !session.startTime) {
            return null;
        }
        const avgLatency = history.reduce((a, b) => a + b, 0) / history.length;
        const maxLatency = Math.max(...history);
        const minLatency = Math.min(...history);
        const uptime = Date.now() - session.startTime;
        return {
            serverName,
            avgLatencyMs: avgLatency,
            maxLatencyMs: maxLatency,
            minLatencyMs: minLatency,
            requestCount: history.length,
            errorRate: session.errorCount > 0 ? session.errorCount / history.length : 0,
            uptimeMs: uptime
        };
    }
    /**
     * Get all metrics
     */
    getAllMetrics() {
        return new Map(this.metrics);
    }
    /**
     * Clear metrics history
     */
    clearMetrics(serverName) {
        if (serverName) {
            this.latencyHistory.delete(serverName);
            this.metrics.delete(serverName);
        }
        else {
            this.latencyHistory.clear();
            this.metrics.clear();
        }
        this.emit('metrics:cleared', { serverName });
    }
    // ============================================
    // Client Management
    // ============================================
    /**
     * Register a client with a session
     */
    registerClient(serverId, clientId) {
        const session = this.sessions.get(serverId);
        if (!session)
            return false;
        if (!session.clients.includes(clientId)) {
            session.clients.push(clientId);
            this.emit('client:registered', { session, clientId });
            return true;
        }
        return false;
    }
    /**
     * Unregister a client from a session
     */
    unregisterClient(serverId, clientId) {
        const session = this.sessions.get(serverId);
        if (!session)
            return false;
        const index = session.clients.indexOf(clientId);
        if (index !== -1) {
            session.clients.splice(index, 1);
            this.emit('client:unregistered', { session, clientId });
            return true;
        }
        return false;
    }
    /**
     * Get all clients for a session
     */
    getClients(serverId) {
        const session = this.sessions.get(serverId);
        return session?.clients || [];
    }
    /**
     * Get all clients across all sessions
     */
    getAllClients() {
        const allClients = new Set();
        for (const session of this.sessions.values()) {
            session.clients.forEach(c => allClients.add(c));
        }
        return allClients;
    }
    // ============================================
    // Configuration
    // ============================================
    /**
     * Update session configuration
     */
    updateConfig(updates) {
        this.config = { ...this.config, ...updates };
        this.emit('config:updated', this.config);
    }
    /**
     * Get current configuration
     */
    getConfig() {
        return { ...this.config };
    }
    /**
     * Set auto-start for a server
     */
    setAutoStart(serverId, enabled) {
        const session = this.sessions.get(serverId);
        if (session) {
            session.autoRestart = enabled;
        }
    }
    // ============================================
    // Utilities
    // ============================================
    /**
     * Generate unique session ID
     */
    generateSessionId() {
        return `mcp-session-${Date.now().toString(36)}-${Math.random().toString(36).substring(2, 9)}`;
    }
    /**
     * Get session by server ID
     */
    getSession(serverId) {
        return this.sessions.get(serverId);
    }
    /**
     * Get all sessions
     */
    getAllSessions() {
        return Array.from(this.sessions.values());
    }
    /**
     * Get sessions by status
     */
    getSessionsByStatus(status) {
        return Array.from(this.sessions.values()).filter(s => s.status === status);
    }
    /**
     * Get session statistics
     */
    getStats() {
        const sessions = this.getAllSessions();
        const running = sessions.filter(s => s.status === 'running').length;
        const stopped = sessions.filter(s => s.status === 'stopped').length;
        const error = sessions.filter(s => s.status === 'error').length;
        const totalClients = sessions.reduce((sum, s) => sum + s.clients.length, 0);
        return {
            totalSessions: sessions.length,
            running,
            stopped,
            error,
            totalClients
        };
    }
    /**
     * Shutdown all sessions
     */
    async shutdown() {
        console.log('[McpSessionService] Shutting down all sessions...');
        this.stopHeartbeat();
        for (const session of this.sessions.values()) {
            if (session.status === 'running' || session.status === 'starting') {
                try {
                    await this.stopSession(session.serverId);
                }
                catch (e) {
                    console.error(`[McpSessionService] Failed to stop ${session.serverName}:`, e);
                }
            }
        }
        this.emit('shutdown', this.getStats());
    }
}
