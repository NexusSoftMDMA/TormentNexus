/**
 * tormentnexus MCP Registry Service
 *
 * Discovers and manages MCP servers from external registries:
 * - Punkpeye/AppCypher (appcypher.github.io/mcp)
 * - wong2 registry (github.com/wong2/mcp-directory)
 * - appcypher registry
 * - ToolSDK (toolsdk.org)
 * - Docker Hub
 * - Playbooks (github.com/openplaybooks/mcp-playbook)
 *
 * Features:
 * - Registry scanning and server discovery
 * - Server metadata extraction
 * - Category classification
 * - Rating and relevance assessment
 * - Auto-installation support
 */
import { EventEmitter } from 'events';
import { DatabaseManager } from '../db/index.js';
// ============================================
// MCPRegistryService Class
// ============================================
export class MCPRegistryService extends EventEmitter {
    dataDir;
    static instance;
    // Built-in registry sources
    registrySources = [
        {
            name: 'Punkpeye/AppCypher',
            url: 'https://raw.githubusercontent.com/appcypher/mcp-directory/main/registry.json',
            type: 'json',
            description: 'Comprehensive MCP server directory'
        },
        {
            name: 'wong2',
            url: 'https://raw.githubusercontent.com/wong2/mcp-directory/main/registry.json',
            type: 'json',
            description: 'wong2\'s MCP server directory'
        },
        {
            name: 'ToolSDK',
            url: 'https://toolsdk.org/api/mcp/servers',
            type: 'api',
            description: 'ToolSDK official registry'
        }
    ];
    // Cache for discovered servers
    discoveredServers = new Map();
    lastSyncTime = 0;
    db;
    constructor(dataDir) {
        super();
        this.dataDir = dataDir;
        this.initDatabase();
    }
    static getInstance(dataDir) {
        if (!MCPRegistryService.instance) {
            MCPRegistryService.instance = new MCPRegistryService(dataDir);
        }
        return MCPRegistryService.instance;
    }
    initDatabase() {
        try {
            this.db = DatabaseManager.getInstance(this.dataDir);
        }
        catch (e) {
            console.error('[MCPRegistryService] Database initialization failed:', e);
        }
    }
    // ============================================
    // Registry Discovery
    // ============================================
    /**
     * Discover servers from all configured registries
     */
    async discoverAll() {
        console.log('[MCPRegistryService] Starting registry discovery...');
        const allServers = [];
        for (const source of this.registrySources) {
            try {
                const servers = await this.discoverFromRegistry(source);
                allServers.push(...servers);
                console.log(`[MCPRegistryService] Found ${servers.length} servers from ${source.name}`);
            }
            catch (e) {
                console.warn(`[MCPRegistryService] Failed to discover from ${source.name}:`, e);
            }
        }
        // Deduplicate servers by name
        const uniqueServers = this.deduplicateServers(allServers);
        // Update cache
        this.discoveredServers.clear();
        for (const server of uniqueServers) {
            this.discoveredServers.set(server.name, server);
        }
        this.lastSyncTime = Date.now();
        // Mark installed servers
        await this.markInstalledServers();
        this.emit('discovery:complete', uniqueServers);
        console.log(`[MCPRegistryService] Discovery complete: ${uniqueServers.length} unique servers`);
        return uniqueServers;
    }
    /**
     * Discover servers from a specific registry source
     */
    async discoverFromRegistry(source) {
        switch (source.type) {
            case 'json':
                return await this.discoverFromJson(source.url, source.name);
            case 'github':
                return await this.discoverFromGithub(source.url, source.name);
            case 'api':
                return await this.discoverFromApi(source.url, source.name);
            default:
                return [];
        }
    }
    /**
     * Fetch and parse JSON registry
     */
    async discoverFromJson(url, sourceName) {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Failed to fetch ${url}: ${response.statusText}`);
        }
        const data = await response.json();
        const servers = Array.isArray(data) ? data : (data.servers || []);
        return servers.map((server) => this.normalizeServerDefinition(server, sourceName));
    }
    /**
     * Discover from GitHub API
     */
    async discoverFromGithub(url, sourceName) {
        // TODO: Implement GitHub API discovery
        // For now, return empty array
        return [];
    }
    /**
     * Discover from API endpoint
     */
    async discoverFromApi(url, sourceName) {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`API request failed: ${response.statusText}`);
        }
        const data = await response.json();
        const servers = Array.isArray(data) ? data : (data.servers || []);
        return servers.map((server) => this.normalizeServerDefinition(server, sourceName));
    }
    /**
     * Normalize server definition to standard format
     */
    normalizeServerDefinition(raw, source) {
        // Handle various registry formats
        const name = raw.name || raw.title || raw.id || 'unknown';
        const description = raw.description || raw.desc || raw.readme || '';
        return {
            name,
            url: raw.url || raw.github_url || raw.repository || '',
            repository: raw.repository || raw.github_url || '',
            description,
            category: this.extractCategory(raw),
            tags: this.extractTags(raw),
            version: raw.version,
            stars: raw.stars || raw.star_count,
            forks: raw.forks,
            license: raw.license,
            rating: this.calculateRating(raw),
            installed: false,
            source
        };
    }
    /**
     * Extract category from server metadata
     */
    extractCategory(raw) {
        // Priority: explicit category > tags > description analysis > default
        if (raw.category)
            return raw.category;
        if (raw.type)
            return raw.type;
        // Analyze description for keywords
        const desc = (raw.description || '').toLowerCase();
        if (desc.includes('file') || desc.includes('filesystem'))
            return 'file-system';
        if (desc.includes('database') || desc.includes('db'))
            return 'database';
        if (desc.includes('memory') || desc.includes('vector'))
            return 'memory';
        if (desc.includes('browser') || desc.includes('web'))
            return 'browser';
        if (desc.includes('code') || desc.includes('execute'))
            return 'code-execution';
        if (desc.includes('search') || desc.includes('index'))
            return 'search';
        if (desc.includes('finance') || desc.includes('financial'))
            return 'finance';
        if (desc.includes('chat') || desc.includes('messaging'))
            return 'communication';
        return 'general';
    }
    /**
     * Extract tags from server metadata
     */
    extractTags(raw) {
        const tags = [];
        if (raw.tags && Array.isArray(raw.tags)) {
            tags.push(...raw.tags);
        }
        if (raw.features && Array.isArray(raw.features)) {
            tags.push(...raw.features);
        }
        return [...new Set(tags)]; // Deduplicate
    }
    /**
     * Calculate rating based on various factors
     */
    calculateRating(raw) {
        let rating = 0;
        // Popularity signals
        const stars = raw.stars || raw.star_count || 0;
        if (stars > 1000)
            rating += 5;
        else if (stars > 500)
            rating += 4;
        else if (stars > 100)
            rating += 3;
        else if (stars > 50)
            rating += 2;
        else if (stars > 10)
            rating += 1;
        // Forks
        const forks = raw.forks || 0;
        if (forks > 100)
            rating += 2;
        else if (forks > 50)
            rating += 1;
        // License (prefer open source)
        if (raw.license && ['MIT', 'Apache-2.0', 'BSD'].includes(raw.license)) {
            rating += 1;
        }
        // Has README/Documentation
        if (raw.description || raw.readme) {
            rating += 1;
        }
        return Math.min(rating, 5); // Cap at 5
    }
    /**
     * Deduplicate servers by name
     */
    deduplicateServers(servers) {
        const uniqueMap = new Map();
        for (const server of servers) {
            const existing = uniqueMap.get(server.name);
            const serverRating = server.rating ?? 0;
            const existingRating = existing?.rating ?? 0;
            if (!existing || serverRating > existingRating) {
                uniqueMap.set(server.name, server);
            }
        }
        return Array.from(uniqueMap.values());
    }
    /**
     * Mark servers that are already installed
     */
    async markInstalledServers() {
        if (!this.db)
            return;
        const installedServers = this.db.getAllMcpServers();
        const installedNames = new Set(installedServers.map(s => s.name));
        for (const server of this.discoveredServers.values()) {
            server.installed = installedNames.has(server.name);
        }
    }
    // ============================================
    // Server Search & Filtering
    // ============================================
    /**
     * Get all discovered servers
     */
    getAllServers() {
        return Array.from(this.discoveredServers.values());
    }
    /**
     * Search discovered servers by query
     */
    searchServers(query, options = {}) {
        const lowerQuery = query.toLowerCase();
        const allServers = this.getAllServers();
        return allServers.filter(server => {
            // Name match
            if (server.name.toLowerCase().includes(lowerQuery))
                return true;
            // Description match
            if (server.description.toLowerCase().includes(lowerQuery))
                return true;
            // Tag match
            if (server.tags.some(tag => tag.toLowerCase().includes(lowerQuery)))
                return true;
            // Category match
            if (server.category.toLowerCase().includes(lowerQuery))
                return true;
            return false;
        }).filter(server => {
            // Apply filters
            if (options.includeCategories && !options.includeCategories.includes(server.category)) {
                return false;
            }
            if (options.excludeCategories && options.excludeCategories.includes(server.category)) {
                return false;
            }
            const rating = server.rating ?? 0;
            if (options.minRating && rating < options.minRating) {
                return false;
            }
            if (!options.includeInstalled && !server.installed) {
                return false;
            }
            return true;
        });
    }
    /**
     * Get servers by category
     */
    getServersByCategory(category) {
        return this.getAllServers().filter(server => server.category === category);
    }
    /**
     * Get all available categories
     */
    getCategories() {
        const categories = new Set();
        for (const server of this.discoveredServers.values()) {
            categories.add(server.category);
        }
        return Array.from(categories).sort();
    }
    /**
     * Get registry statistics
     */
    getStats() {
        const servers = this.getAllServers();
        const categories = this.getCategories();
        const installedCount = servers.filter(s => s.installed).length;
        return {
            totalServers: servers.length,
            installedServers: installedCount,
            categories,
            lastSync: this.lastSyncTime
        };
    }
    // ============================================
    // Installation Support
    // ============================================
    /**
     * Prepare server for installation
     * Returns the URL to clone or npm package to install
     */
    prepareInstallation(serverName) {
        const server = this.discoveredServers.get(serverName);
        if (!server)
            return null;
        // Determine installation type
        if (server.repository && server.repository.includes('github.com')) {
            return { type: 'github', source: server.repository };
        }
        else if (server.url.includes('npmjs.com') || server.url.includes('npm')) {
            return { type: 'npm', source: server.url };
        }
        else {
            return { type: 'url', source: server.url };
        }
    }
    /**
     * Get server definition by name
     */
    getServer(name) {
        return this.discoveredServers.get(name);
    }
    /**
     * Add custom registry source
     */
    addRegistrySource(source) {
        this.registrySources.push(source);
        this.emit('registry:added', source);
    }
    /**
     * Remove registry source
     */
    removeRegistrySource(name) {
        const index = this.registrySources.findIndex(s => s.name === name);
        if (index !== -1) {
            this.registrySources.splice(index, 1);
            this.emit('registry:removed', name);
            return true;
        }
        return false;
    }
    /**
     * Get all registry sources
     */
    getRegistrySources() {
        return [...this.registrySources];
    }
    /**
     * Clear server cache
     */
    clearCache() {
        this.discoveredServers.clear();
        this.lastSyncTime = 0;
        this.emit('cache:cleared');
    }
    /**
     * Force refresh registry
     */
    async refresh() {
        this.clearCache();
        return await this.discoverAll();
    }
}
