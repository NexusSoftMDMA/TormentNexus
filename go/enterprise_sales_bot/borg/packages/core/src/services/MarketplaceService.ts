import { z } from 'zod';
import { McpmRegistry, RegistryItem } from '../skills/McpmRegistry.js';
import { McpmInstaller } from '../skills/McpmInstaller.js';
import { MeshService, SwarmMessage, SwarmMessageType } from '../mesh/MeshService.js';
import { Registry } from '@tormentnexus/mcp-registry';
import path from 'path';

export const MarketplaceEntrySchema = z.object({
    id: z.string(),
    name: z.string(),
    description: z.string(),
    author: z.string().optional(),
    type: z.enum(['agent', 'tool', 'skill']),
    source: z.enum(['official', 'community', 'local']),
    url: z.string().optional(), // Git URL or Mesh Key
    verified: z.boolean().default(false),
    peerCount: z.number().default(0),
    installed: z.boolean().default(false),
    tags: z.array(z.string()).default([])
});

export type MarketplaceEntry = z.infer<typeof MarketplaceEntrySchema>;

export class MarketplaceService {
    private legacyRegistry: McpmRegistry;
    private mcpRegistry: ReturnType<typeof Registry.prototype.list>;
    private installer: McpmInstaller;
    private meshService?: MeshService;
    private installDir: string;
    private peerTools: Map<string, { lastSeen: number, entries: MarketplaceEntry[] }> = new Map();

    constructor(installDir: string, meshService?: MeshService) {
        this.installDir = installDir;
        this.legacyRegistry = new McpmRegistry();

        // Internal MCP Server registry
        const registry = new Registry();
        this.mcpRegistry = registry.list();

        this.installer = new McpmInstaller(installDir);
        this.meshService = meshService;

        if (this.meshService) {
            this.meshService.on('message', this.handleMeshMessage.bind(this));
        }
    }

    private handleMeshMessage(msg: SwarmMessage) {
        if (msg.type === SwarmMessageType.MARKETPLACE_PUBLISH || msg.type === SwarmMessageType.MARKETPLACE_RESPONSE) {
            const entries = msg.payload as MarketplaceEntry[];
            this.peerTools.set(msg.sender, { lastSeen: Date.now(), entries });
        } else if (msg.type === SwarmMessageType.MARKETPLACE_QUERY) {
            // Send back our local tools
            this.list().then(localTools => {
                if (this.meshService) {
                    this.meshService.sendResponse(msg, SwarmMessageType.MARKETPLACE_RESPONSE, localTools.filter(x => x.source === 'official'));
                }
            }).catch(e => console.error("MarketplaceService failed to resolve local tools for query", e));
        }
    }

    /**
     * Aggregates tools from Registry (Official) and Mesh (Community)
     */
    async list(filter?: string): Promise<MarketplaceEntry[]> {
        // Load legacy skills
        const officialItems = await this.legacyRegistry.search(filter || '');

        const entries: MarketplaceEntry[] = officialItems.map(item => ({
            id: item.name,
            name: item.name,
            description: "Official Skill",
            author: "TormentNexus Ecosystem",
            type: 'skill', // Legacy are mostly skills
            source: 'official',
            url: item.url,
            verified: true,
            peerCount: 1,
            installed: false,
            tags: item.tags || []
        }));

        // Load new MCP Servers from registry.json
        for (const server of this.mcpRegistry) {
            // Apply simple filter if provided
            if (filter && !server.name.toLowerCase().includes(filter.toLowerCase()) &&
                !server.description.toLowerCase().includes(filter.toLowerCase())) {
                continue;
            }

            entries.push({
                id: server.package, // Package name is unique ID
                name: server.name,
                description: server.description,
                author: "MCP Registry",
                type: 'tool', // MCP servers are essentially tool providers
                source: 'official',
                url: `https://www.npmjs.com/package/${server.package}`,
                verified: true,
                peerCount: 1,
                installed: false,
                tags: ['mcp', server.type]
            });
        }

        // Mesh Discovery (Future: Listen for announcements)
        if (this.meshService) {
            // Housekeeping: remove peers not seen in 120 seconds
            const now = Date.now();
            for (const [peerId, data] of this.peerTools.entries()) {
                if (now - data.lastSeen > 120000) {
                    this.peerTools.delete(peerId);
                } else {
                    // Inject community tags
                    const formatted = data.entries.map(e => ({...e, source: 'community' as const, author: `Node ${peerId.substring(0,8)}`}));
                    entries.push(...formatted);
                }
            }
        }

        // Check installation status
        for (const entry of entries) {
            entry.installed = await this.checkInstalled(entry.id);
        }

        return entries;
    }

    async install(id: string): Promise<string> {
        // Delegate to existing McpmInstaller
        // It currently takes a 'skillName' which matches the registry 'name'
        return this.installer.install(id);
    }

    async publish(manifest: Partial<MarketplaceEntry>): Promise<string> {
        if (!this.meshService) {
            throw new Error("MeshService not available for publishing.");
        }
        
        const entry: MarketplaceEntry = {
            id: manifest.id || 'unknown',
            name: manifest.name || 'Unknown Tool',
            description: manifest.description || '',
            author: manifest.author || 'TormentNexus Ecosystem',
            type: manifest.type || 'tool',
            source: 'community',
            url: manifest.url,
            verified: false,
            peerCount: 1,
            installed: false,
            tags: manifest.tags || []
        };
        
        this.meshService.broadcast(SwarmMessageType.MARKETPLACE_PUBLISH, [entry]);
        return "Published to Mesh.";
    }

    private async checkInstalled(id: string): Promise<boolean> {
        const fs = await import('fs/promises');

        // 1. Check if it's an MCP server installed in mcp.json
        try {
            // Assume mcp.json is near the root or in a config dir
            // The default TormentNexus Controller writes to mcp.json in the current working directory or a specific config path
            const mcpJsonPath = path.join(process.cwd(), 'mcp.json');
            const mcpJsonRaw = await fs.readFile(mcpJsonPath, 'utf-8');
            const mcpConfig = JSON.parse(mcpJsonRaw);

            // The ID is the package name. In mcp.json, servers are usually named after the package without scopes, or similar.
            // But we can check if the command includes the package name id
            if (mcpConfig && mcpConfig.mcpServers) {
                for (const [serverName, serverConfig] of Object.entries(mcpConfig.mcpServers)) {
                    // Check if the command or args reference the package ID
                    const configStr = JSON.stringify(serverConfig);
                    if (configStr.includes(id)) {
                        return true;
                    }
                }
            }
        } catch {
            // mcp.json might not exist or be accessible, fallback to legacy check
        }

        // 2. Legacy check if dir exists in installDir
        try {
            await fs.access(path.join(this.installDir, id));
            return true;
        } catch {
            return false;
        }
    }
}
