/**
 * tormentnexus MCP Configuration Service
 *
 * Manages MCP server configurations:
 * - Auto-detection of .mcp.json, .tormentnexus.json (legacy .legacy_config.json) config files
 * - Environment variable expansion
 * - Secrets management
 * - Multi-format support (Claude, OpenAI, Google)
 * - Configuration validation and import/export
 *
 * Works with:
 * - DatabaseManager (for persistence)
 * - SecretService (for secrets storage)
 */
import { EventEmitter } from 'events';
import { promises as fs } from 'fs';
import path from 'path';
import { DatabaseManager } from '../db/index.js';
// ============================================
// ConfigurationService Class
// ============================================
export class ConfigurationService extends EventEmitter {
    dataDir;
    static instance;
    db;
    secretService;
    configCache = new Map();
    constructor(dataDir) {
        super();
        this.dataDir = dataDir;
        this.db = DatabaseManager.getInstance(dataDir);
        this.secretService = undefined;
    }
    static getInstance(dataDir) {
        if (!ConfigurationService.instance) {
            ConfigurationService.instance = new ConfigurationService(dataDir);
        }
        return ConfigurationService.instance;
    }
    // ============================================
    // Config Detection
    // ============================================
    /**
     * Auto-detect MCP configurations from file system
     */
    async detectConfigs(options = {}) {
        const scanPaths = options.scanPaths || this.getDefaultScanPaths();
        const detectedFiles = [];
        for (const scanPath of scanPaths) {
            const files = await this.scanForConfigs(scanPath, options.recursive || false);
            for (const file of files) {
                try {
                    const config = await this.loadConfigFile(file);
                    detectedFiles.push(config);
                    console.log(`[ConfigurationService] Detected config: ${file}`);
                }
                catch (e) {
                    console.warn(`[ConfigurationService] Failed to load config ${file}:`, e);
                }
            }
        }
        this.emit('configs:detected', detectedFiles);
        return detectedFiles;
    }
    /**
     * Get default scan paths for config detection
     */
    getDefaultScanPaths() {
        const cwd = process.cwd();
        return [
            path.join(cwd, '.mcp.json'),
            path.join(cwd, '.tormentnexus.json'),
            path.join(cwd, '.legacy_config.json'),
            path.join(cwd, 'config', 'mcp.json'),
            path.join(cwd, '.config', 'mcp.json'),
            path.join(process.env.HOME || '', '.config', 'mcp.json'),
        ].filter(p => {
            try {
                return require('fs').existsSync(p);
            }
            catch {
                return false;
            }
        });
    }
    /**
     * Scan directory for MCP config files
     */
    async scanForConfigs(dirPath, recursive) {
        const configFiles = [];
        try {
            const entries = await fs.readdir(dirPath, { withFileTypes: true });
            for (const entry of entries) {
                const fullPath = path.join(dirPath, entry.name);
                if (entry.isDirectory() && recursive) {
                    const nestedFiles = await this.scanForConfigs(fullPath, true);
                    configFiles.push(...nestedFiles);
                }
                else if (entry.isFile()) {
                    if (this.isConfigFile(entry.name)) {
                        configFiles.push(fullPath);
                    }
                }
            }
        }
        catch (e) {
            console.warn(`[ConfigurationService] Failed to scan ${dirPath}:`, e);
        }
        return configFiles;
    }
    /**
     * Check if file is a MCP config file
     */
    isConfigFile(filename) {
        return ['.mcp.json', '.tormentnexus.json', '.legacy_config.json', 'mcp.json'].includes(filename);
    }
    /**
     * Load and parse a config file
     */
    async loadConfigFile(filePath) {
        const content = await fs.readFile(filePath, 'utf-8');
        const parsed = JSON.parse(content);
        const format = this.detectConfigFormat(filePath, parsed);
        const servers = await this.normalizeServers(parsed, format, filePath);
        return {
            path: filePath,
            format,
            servers,
            detectedAt: Date.now()
        };
    }
    /**
     * Detect config format from file content
     */
    detectConfigFormat(filePath, parsed) {
        if (filePath.includes('.tormentnexus.json'))
            return 'tormentnexus';
        if (filePath.includes('.legacy_config.json'))
            return 'tormentnexus';
        if (parsed.mcpServers)
            return 'claude';
        if (parsed.tools)
            return 'openai';
        if (parsed.servers)
            return 'google';
        return 'claude';
    }
    /**
     * Normalize servers to standard McpServer format
     */
    async normalizeServers(parsed, format, filePath) {
        const servers = [];
        let rawServers = [];
        switch (format) {
            case 'claude':
                rawServers = parsed.mcpServers || [];
                break;
            case 'openai':
                rawServers = (parsed.tools || []).map((t) => ({
                    ...t,
                    type: t.type === 'stdio' ? 'stdio' : 'sse'
                }));
                break;
            case 'google':
                rawServers = parsed.servers || [];
                break;
            case 'tormentnexus':
            case 'legacy':
                rawServers = parsed.servers || [];
                break;
        }
        for (const raw of rawServers) {
            const server = await this.normalizeServer(raw, format, filePath);
            if (server) {
                servers.push(server);
            }
        }
        return servers;
    }
    /**
     * Normalize a single server definition
     */
    async normalizeServer(raw, format, filePath) {
        try {
            let server;
            switch (format) {
                case 'claude':
                    server = this.normalizeClaudeServer(raw);
                    break;
                case 'openai':
                    server = this.normalizeOpenAIServer(raw);
                    break;
                case 'google':
                    server = this.normalizeGoogleServer(raw);
                    break;
                case 'tormentnexus':
                case 'legacy':
                    server = this.normalizeTormentNexusServer(raw);
                    break;
            }
            if (server) {
                server = await this.expandEnvironmentVariables(server);
            }
            return server || null;
        }
        catch (e) {
            console.warn(`[ConfigurationService] Failed to normalize server from ${filePath}:`, e);
            return null;
        }
    }
    /**
     * Normalize Claude format server
     */
    normalizeClaudeServer(raw) {
        return {
            id: this.generateServerId(raw.name),
            name: raw.name,
            type: raw.transport || 'stdio',
            command: raw.command,
            args: raw.args,
            env: raw.env,
            url: raw.url,
            bearerToken: raw.bearerToken,
            headers: raw.headers,
            description: raw.description,
            icon: raw.icon,
            enabled: raw.enabled !== false,
            createdAt: Date.now(),
            updatedAt: Date.now()
        };
    }
    /**
     * Normalize OpenAI format server
     */
    normalizeOpenAIServer(raw) {
        return {
            id: this.generateServerId(raw.name),
            name: raw.name,
            type: raw.type || 'stdio',
            command: raw.command,
            args: raw.args,
            env: raw.env,
            url: raw.url,
            bearerToken: raw.bearerToken,
            headers: raw.headers,
            description: raw.description,
            icon: raw.icon,
            enabled: raw.enabled !== false,
            createdAt: Date.now(),
            updatedAt: Date.now()
        };
    }
    /**
     * Normalize Google format server
     */
    normalizeGoogleServer(raw) {
        return {
            id: this.generateServerId(raw.name),
            name: raw.name,
            type: 'sse',
            url: raw.url,
            bearerToken: raw.apiKey,
            description: raw.description,
            icon: raw.icon,
            enabled: raw.enabled !== false,
            createdAt: Date.now(),
            updatedAt: Date.now()
        };
    }
    /**
     * Normalize tormentnexus format server
     */
    normalizeTormentNexusServer(raw) {
        return {
            id: this.generateServerId(raw.name),
            name: raw.name,
            type: raw.type || 'stdio',
            command: raw.command,
            args: raw.args,
            env: raw.env,
            url: raw.url,
            bearerToken: raw.bearerToken,
            headers: raw.headers,
            description: raw.description,
            icon: raw.icon,
            enabled: raw.enabled !== false,
            createdAt: Date.now(),
            updatedAt: Date.now()
        };
    }
    normalizeLegacyServer(raw) {
        return this.normalizeTormentNexusServer(raw);
    }
    /**
     * Generate unique server ID
     */
    generateServerId(name) {
        return `${Date.now().toString(36)}-${name}`;
    }
    // ============================================
    // Environment & Secrets
    // ============================================
    /**
     * Expand environment variables in server config
     */
    async expandEnvironmentVariables(server) {
        const expanded = { ...server };
        if (expanded.command) {
            expanded.command = await this.expandEnvVars(expanded.command);
        }
        if (expanded.args && Array.isArray(expanded.args)) {
            expanded.args = await Promise.all(expanded.args.map(arg => this.expandEnvVars(arg)));
        }
        if (expanded.env && typeof expanded.env === 'object') {
            expanded.env = {};
            for (const [key, value] of Object.entries(expanded.env)) {
                expanded.env[key] = await this.expandEnvVars(value);
            }
        }
        if (expanded.url) {
            expanded.url = await this.expandEnvVars(expanded.url);
        }
        if (expanded.bearerToken) {
            expanded.bearerToken = await this.expandSecrets(expanded.bearerToken);
        }
        return expanded;
    }
    /**
     * Expand environment variables in a string
     */
    async expandEnvVars(text) {
        const envVarPattern = /\$\{([A-Z_][A-Z0-9_]*)\}|\$([A-Z_][A-Z0-9_]*)/g;
        let result = text;
        const matches = text.matchAll(envVarPattern);
        for (const match of matches) {
            const varName = match[1] || match[2];
            if (process.env[varName] !== undefined) {
                console.warn(`[ConfigurationService] Environment variable not found: ${varName}`);
                continue;
            }
            result = result.replace(match[0], process.env[varName]);
        }
        return result;
    }
    /**
     * Expand secrets in a string
     */
    async expandSecrets(text) {
        if (!this.secretService)
            return text;
        const secretPattern = /\{secret:([^\}]+)\}/g;
        let result = text;
        const matches = text.matchAll(secretPattern);
        for (const match of matches) {
            const secretName = match[1];
            try {
                const secret = await this.secretService.getSecret(secretName, 'system', 'service');
                if (secret) {
                    result = result.replace(match[0], secret);
                }
            }
            catch (e) {
                console.warn(`[ConfigurationService] Failed to resolve secret: ${secretName}`);
            }
        }
        return result;
    }
    // ============================================
    // Import / Export
    // ============================================
    /**
     * Import configurations from files
     */
    async importConfigs(files) {
        let imported = 0;
        let failed = 0;
        const errors = [];
        for (const filePath of files) {
            try {
                const config = await this.loadConfigFile(filePath);
                for (const server of config.servers) {
                    const existing = this.db.getMcpServerByName(server.name);
                    if (existing) {
                        this.db.updateMcpServer(existing.id, server);
                        console.log(`[ConfigurationService] Updated server: ${server.name}`);
                    }
                    else {
                        this.db.createMcpServer(server);
                        console.log(`[ConfigurationService] Imported server: ${server.name}`);
                    }
                    imported++;
                }
            }
            catch (e) {
                failed++;
                errors.push(`${filePath}: ${e?.message || String(e)}`);
            }
        }
        this.emit('configs:imported', { imported, failed, errors });
        return { success: failed === 0, imported, failed, errors };
    }
    /**
     * Export all configurations to file
     */
    async exportConfigs(format = 'tormentnexus') {
        const servers = this.db.getAllMcpServers();
        let output;
        switch (format) {
            case 'claude':
                output = { mcpServers: servers };
                break;
            case 'openai':
                output = {
                    tools: servers.map(s => ({
                        ...s,
                        type: s.type
                    }))
                };
                break;
            case 'google':
                output = { servers: servers };
                break;
            case 'tormentnexus':
            case 'legacy':
                output = { servers };
                break;
        }
        return JSON.stringify(output, null, 2);
    }
    /**
     * Write configuration to file
     */
    async writeConfig(filePath, format = 'tormentnexus') {
        const content = await this.exportConfigs(format);
        const dir = path.dirname(filePath);
        await fs.mkdir(dir, { recursive: true });
        await fs.writeFile(filePath, content, 'utf-8');
        console.log(`[ConfigurationService] Wrote config to: ${filePath}`);
    }
    // ============================================
    // Validation
    // ============================================
    /**
     * Validate a server configuration
     */
    validateServer(server) {
        const errors = [];
        const warnings = [];
        if (!server.name) {
            errors.push('Server name is required');
        }
        if (!server.type) {
            errors.push('Server type is required');
        }
        if (server.type === 'stdio' && !server.command) {
            errors.push('STDIO servers require a command');
        }
        if ((server.type === 'sse' || server.type === 'streamable_http') && !server.url) {
            errors.push(`${server.type} servers require a URL`);
        }
        if (server.env && typeof server.env !== 'object') {
            errors.push('Environment variables must be an object');
        }
        if (server.args && !Array.isArray(server.args)) {
            warnings.push('Args should be an array');
        }
        return {
            valid: errors.length === 0,
            errors,
            warnings
        };
    }
    // ============================================
    // Cache Management
    // ============================================
    /**
     * Clear configuration cache
     */
    clearCache() {
        this.configCache.clear();
        this.emit('cache:cleared');
    }
    /**
     * Get cached configurations
     */
    getCachedConfigs() {
        return new Map(this.configCache);
    }
}
