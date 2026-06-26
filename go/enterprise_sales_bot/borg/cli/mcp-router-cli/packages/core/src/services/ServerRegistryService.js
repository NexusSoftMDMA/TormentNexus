/**
 * tormentnexus Server Registry Service
 *
 * Complements McpManager by providing:
 * - Installation of new MCP servers from registries (GitHub, npm)
 * - Update management (track upstream changes)
 * - Health monitoring and status checks
 * - Metadata enrichment (categories, tags, ratings)
 *
 * Works with:
 * - MCPRegistryService (for server discovery)
 * - McpManager (for server lifecycle)
 * - DatabaseManager (for persistence)
 */
import { EventEmitter } from 'events';
import { promises as fs } from 'fs';
import path from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';
import { DatabaseManager } from '../db/index.js';
import { MCPRegistryService } from './MCPRegistryService.js';
const execAsync = promisify(exec);
// ============================================
// ServerRegistryService Class
// ============================================
export class ServerRegistryService extends EventEmitter {
    dataDir;
    static instance;
    db;
    registryService;
    healthCheckInterval;
    HEALTH_CHECK_INTERVAL = 60000; // 1 minute
    mcpDir;
    constructor(dataDir) {
        super();
        this.dataDir = dataDir;
        this.db = DatabaseManager.getInstance(dataDir);
        this.registryService = MCPRegistryService.getInstance(dataDir);
        this.mcpDir = path.join(this.dataDir || process.cwd(), 'mcp-servers');
    }
    static getInstance(dataDir) {
        if (!ServerRegistryService.instance) {
            ServerRegistryService.instance = new ServerRegistryService(dataDir);
        }
        return ServerRegistryService.instance;
    }
    async initialize() {
        await fs.mkdir(this.mcpDir, { recursive: true });
        this.startHealthChecks();
    }
    // ============================================
    // Server Installation
    // ============================================
    /**
     * Install a server from registry definition
     */
    async installServer(name, options = {}) {
        console.log(`[ServerRegistryService] Installing server: ${name}`);
        const serverDef = this.registryService.getServer(name);
        if (!serverDef) {
            return {
                success: false,
                serverName: name,
                error: `Server not found in registry: ${name}`
            };
        }
        try {
            let serverConfig;
            if (options.type === 'npm' || serverDef.url.includes('npm')) {
                serverConfig = await this.installFromNpm(serverDef);
            }
            else {
                serverConfig = await this.installFromGithub(serverDef);
            }
            const dbServer = this.db.createMcpServer(serverConfig);
            this.emit('server:installed', { serverDef, dbServer });
            console.log(`[ServerRegistryService] Successfully installed: ${name}`);
            return {
                success: true,
                serverName: name,
                serverId: dbServer.id
            };
        }
        catch (e) {
            return {
                success: false,
                serverName: name,
                error: e?.message || 'Installation failed'
            };
        }
    }
    /**
     * Install from GitHub repository
     */
    async installFromGithub(serverDef) {
        const repoUrl = serverDef.repository || serverDef.url;
        const match = repoUrl.match(/github\.com\/([^/]+)\/([^/]+)/);
        if (!match) {
            throw new Error('Invalid GitHub repository URL');
        }
        const [, owner, repo] = match;
        const installPath = path.join(this.mcpDir, repo);
        if (await this.pathExists(installPath)) {
            console.log(`[ServerRegistryService] Repository already exists: ${installPath}`);
        }
        else {
            console.log(`[ServerRegistryService] Cloning repository: ${repoUrl}`);
            await execAsync(`git clone ${repoUrl} ${installPath}`);
        }
        const serverPath = await this.findMcpServerEntry(installPath);
        if (!serverPath) {
            throw new Error('No MCP server entry point found');
        }
        const now = Date.now();
        return {
            id: `${now}-${serverDef.name}`,
            name: serverDef.name,
            type: 'stdio',
            command: this.getInstallCommand(serverPath),
            args: [],
            description: serverDef.description,
            icon: serverDef.source,
            enabled: true,
            createdAt: now,
            updatedAt: now
        };
    }
    /**
     * Install from npm package
     */
    async installFromNpm(serverDef) {
        const npmUrl = serverDef.url;
        const match = npmUrl.match(/npmjs\.com\/package\/([^/]+)/);
        const packageName = match ? match[1] : serverDef.name.replace(/-/g, '_');
        console.log(`[ServerRegistryService] Installing npm package: ${packageName}`);
        await execAsync(`npm install ${packageName}`, { cwd: this.mcpDir });
        const now = Date.now();
        return {
            id: `${now}-${serverDef.name}`,
            name: serverDef.name,
            type: 'stdio',
            command: `node ${path.join(this.mcpDir, 'node_modules', packageName)}`,
            args: [],
            description: serverDef.description,
            icon: serverDef.source,
            enabled: true,
            createdAt: now,
            updatedAt: now
        };
    }
    /**
     * Find MCP server entry point in a directory
     */
    async findMcpServerEntry(dirPath) {
        const entries = await fs.readdir(dirPath);
        for (const entry of entries) {
            const fullPath = path.join(dirPath, entry);
            const stat = await fs.stat(fullPath);
            if (stat.isDirectory()) {
                const packageJsonPath = path.join(fullPath, 'package.json');
                if (await this.pathExists(packageJsonPath)) {
                    const packageJson = JSON.parse(await fs.readFile(packageJsonPath, 'utf-8'));
                    if (packageJson.bin) {
                        return path.join(fullPath, typeof packageJson.bin === 'string' ? packageJson.bin : Object.keys(packageJson.bin)[0]);
                    }
                }
            }
        }
        return dirPath;
    }
    /**
     * Get the install command for a server
     */
    getInstallCommand(serverPath) {
        return serverPath;
    }
    /**
     * Uninstall a server
     */
    async uninstallServer(serverId) {
        const server = this.db.getMcpServer(serverId);
        if (!server) {
            return false;
        }
        try {
            if (server.command) {
                const serverPath = path.parse(server.command).dir;
                if (await this.pathExists(serverPath)) {
                    await fs.rm(serverPath, { recursive: true, force: true });
                    console.log(`[ServerRegistryService] Removed server directory: ${serverPath}`);
                }
            }
            this.db.deleteMcpServer(serverId);
            this.emit('server:uninstalled', server);
            console.log(`[ServerRegistryService] Successfully uninstalled: ${server.name}`);
            return true;
        }
        catch (e) {
            console.error(`[ServerRegistryService] Failed to uninstall ${server.name}:`, e);
            return false;
        }
    }
    // ============================================
    // Update Management
    // ============================================
    /**
     * Check for updates on all installed servers
     */
    async checkUpdates() {
        const installedServers = this.db.getAllMcpServers();
        const updates = [];
        for (const server of installedServers) {
            try {
                const updateInfo = await this.checkServerUpdate(server);
                updates.push(updateInfo);
            }
            catch (e) {
                console.warn(`[ServerRegistryService] Failed to check updates for ${server.name}:`, e);
            }
        }
        this.emit('updates:checked', updates);
        return updates;
    }
    /**
     * Check for updates on a specific server
     */
    async checkServerUpdate(server) {
        if (server.type !== 'stdio') {
            return {
                serverName: server.name,
                hasUpdates: false
            };
        }
        const serverPath = path.parse(server.command || '').dir;
        if (!(await this.pathExists(path.join(serverPath, '.git')))) {
            return {
                serverName: server.name,
                hasUpdates: false
            };
        }
        try {
            await execAsync('git fetch', { cwd: serverPath });
            const { stdout: localHash } = await execAsync('git rev-parse HEAD', { cwd: serverPath });
            const { stdout: remoteHash } = await execAsync('git rev-parse @{u}', { cwd: serverPath });
            const hasUpdates = localHash.trim() !== remoteHash.trim();
            return {
                serverName: server.name,
                hasUpdates,
                commitHash: remoteHash.trim(),
                updatedAt: Date.now()
            };
        }
        catch (e) {
            return {
                serverName: server.name,
                hasUpdates: false,
                error: String(e)
            };
        }
    }
    /**
     * Update a server
     */
    async updateServer(serverId) {
        const server = this.db.getMcpServer(serverId);
        if (!server) {
            return {
                success: false,
                serverName: 'unknown',
                error: 'Server not found'
            };
        }
        if (server.type !== 'stdio') {
            return {
                success: false,
                serverName: server.name,
                error: 'Only STDIO servers can be updated'
            };
        }
        const serverPath = path.parse(server.command || '').dir;
        try {
            await execAsync('git pull', { cwd: serverPath });
            const updated = this.db.updateMcpServer(serverId, {
                updatedAt: Date.now()
            });
            this.emit('server:updated', updated);
            console.log(`[ServerRegistryService] Successfully updated: ${server.name}`);
            return {
                success: true,
                serverName: server.name,
                serverId: server.id
            };
        }
        catch (e) {
            return {
                success: false,
                serverName: server.name,
                error: e?.message || 'Update failed'
            };
        }
    }
    // ============================================
    // Health Monitoring
    // ============================================
    /**
     * Start periodic health checks
     */
    startHealthChecks() {
        if (this.healthCheckInterval) {
            clearInterval(this.healthCheckInterval);
        }
        this.healthCheckInterval = setInterval(() => {
            this.performHealthCheck();
        }, this.HEALTH_CHECK_INTERVAL);
    }
    /**
     * Stop health checks
     */
    stopHealthChecks() {
        if (this.healthCheckInterval) {
            clearInterval(this.healthCheckInterval);
            this.healthCheckInterval = undefined;
        }
    }
    /**
     * Perform health check on all servers
     */
    async performHealthCheck() {
        const servers = this.db.getAllMcpServers({ enabled: true });
        const results = [];
        for (const server of servers) {
            const result = await this.checkServerHealth(server);
            results.push(result);
        }
        this.emit('health:checked', results);
        return results;
    }
    /**
     * Check health of a specific server
     */
    async checkServerHealth(server) {
        const startTime = Date.now();
        try {
            if (server.type === 'stdio' && server.command) {
                const exists = await this.pathExists(server.command);
                if (!exists) {
                    return {
                        serverName: server.name,
                        status: 'error',
                        lastChecked: Date.now(),
                        error: 'Server binary not found'
                    };
                }
            }
            if (server.type === 'sse' || server.type === 'streamable_http') {
                if (!server.url) {
                    return {
                        serverName: server.name,
                        status: 'error',
                        lastChecked: Date.now(),
                        error: 'No URL configured'
                    };
                }
                try {
                    await fetch(server.url, { method: 'HEAD' });
                }
                catch (e) {
                    return {
                        serverName: server.name,
                        status: 'unhealthy',
                        lastChecked: Date.now(),
                        error: e?.message || 'Health check failed'
                    };
                }
            }
            return {
                serverName: server.name,
                status: 'healthy',
                latencyMs: Date.now() - startTime,
                lastChecked: Date.now()
            };
        }
        catch (e) {
            return {
                serverName: server.name,
                status: 'error',
                lastChecked: Date.now(),
                error: e?.message || 'Unknown error'
            };
        }
    }
    // ============================================
    // Metadata Management
    // ============================================
    /**
     * Update server metadata from registry
     */
    async updateServerMetadata(serverName) {
        const dbServer = this.db.getMcpServerByName(serverName);
        const registryServer = this.registryService.getServer(serverName);
        if (dbServer && registryServer) {
            this.db.updateMcpServer(dbServer.id, {
                description: registryServer.description,
                icon: registryServer.source,
                updatedAt: Date.now()
            });
            this.emit('server:metadata:updated', { dbServer, registryServer });
        }
    }
    /**
     * Update all server metadata from registry
     */
    async updateAllMetadata() {
        const installedServers = this.db.getAllMcpServers();
        let updated = 0;
        for (const server of installedServers) {
            try {
                await this.updateServerMetadata(server.name);
                updated++;
            }
            catch (e) {
                console.warn(`[ServerRegistryService] Failed to update metadata for ${server.name}:`, e);
            }
        }
        this.emit('metadata:updated', updated);
        return updated;
    }
    // ============================================
    // Utilities
    // ============================================
    async pathExists(p) {
        try {
            await fs.access(p);
            return true;
        }
        catch {
            return false;
        }
    }
    /**
     * Get all installed servers
     */
    getAllServers() {
        return this.db.getAllMcpServers();
    }
    /**
     * Get server by name
     */
    getServer(name) {
        return this.db.getMcpServerByName(name);
    }
    /**
     * Get server statistics
     */
    getStats() {
        const servers = this.getAllServers();
        const enabled = servers.filter(s => s.enabled).length;
        const byType = {
            stdio: 0,
            sse: 0,
            streamable_http: 0
        };
        for (const server of servers) {
            byType[server.type]++;
        }
        return {
            totalServers: servers.length,
            enabledServers: enabled,
            byType
        };
    }
}
