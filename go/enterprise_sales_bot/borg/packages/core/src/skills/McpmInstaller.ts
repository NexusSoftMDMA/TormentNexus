import { McpmRegistry } from './McpmRegistry.js';
import { Registry as McpServerRegistry } from '@tormentnexus/mcp-registry';
import { spawn } from 'child_process';
import path from 'path';
import fs from 'fs/promises';

function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

export class McpmInstaller {
    private legacyRegistry: McpmRegistry;
    private mcpRegistry: McpServerRegistry;
    private installDir: string;

    constructor(installDir: string) {
        this.legacyRegistry = new McpmRegistry();
        this.mcpRegistry = new McpServerRegistry();
        this.installDir = installDir;
    }

    async install(skillName: string): Promise<string> {
        // 1. Check if it's an MCP server from the new registry
        // The frontend passes the package name as the ID
        const mcpServer = this.mcpRegistry.list().find(s => s.package === skillName || s.name === skillName);

        if (mcpServer) {
            console.log(`[McpmInstaller] Installing MCP Server: ${mcpServer.package}`);
            try {
                // Execute the internal tormentnexus CLI command
                await this.runCommand(`npx tormentnexus mcp install ${mcpServer.package}`);
                return `Successfully installed MCP Server '${mcpServer.name}'`;
            } catch (e: unknown) {
                throw new Error(`Failed to install MCP server: ${getErrorMessage(e)}`);
            }
        }

        // 2. Fallback to legacy Git-based skills
        const results = await this.legacyRegistry.search(skillName);
        const match = results.find(r => r.name.toLowerCase() === skillName.toLowerCase()) || results[0];

        if (!match) {
            throw new Error(`Skill or Server '${skillName}' not found in any registry.`);
        }

        const targetPath = path.join(this.installDir, match.name);

        try {
            await fs.access(targetPath);
            return `Skill '${match.name}' is already installed at ${targetPath}. (Skipping download)`;
        } catch { } // Proceeds if doesn't exist

        await fs.mkdir(this.installDir, { recursive: true });

        console.log(`[McpmInstaller] Cloning ${match.url} to ${targetPath}...`);
        try {
            await this.runCommand(`git clone ${match.url} "${targetPath}"`);
            return `Successfully installed '${match.name}' from ${match.url}`;
        } catch (e: unknown) {
            throw new Error(`Failed to clone skill: ${getErrorMessage(e)}`);
        }
    }

    async search(query: string) {
        return this.legacyRegistry.search(query);
    }

    private runCommand(cmd: string): Promise<void> {
        return new Promise((resolve, reject) => {
            const p = spawn(cmd, { shell: true, stdio: 'inherit' });
            p.on('close', (code) => code === 0 ? resolve() : reject(new Error(`Command failed: ${cmd}`)));
            p.on('error', reject);
        });
    }
}
