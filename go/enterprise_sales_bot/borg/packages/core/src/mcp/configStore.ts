import fs from 'fs';
import path from 'path';

import type { MCPServerConfig } from './types.js';

export class MCPConfigStore {
    private readonly lkgPath: string;

    constructor(private readonly configPath: string) {
        this.lkgPath = configPath.replace(/\.json$/, '.lkg.json');
    }

    public readAll(): Record<string, MCPServerConfig> {
        if (!fs.existsSync(this.configPath)) {
            return this.readLKG();
        }

        try {
            const raw = fs.readFileSync(this.configPath, 'utf8');
            const parsed = JSON.parse(raw) as Record<string, MCPServerConfig>;
            
            // If we successfully read and parsed, ensure LKG is updated
            this.writeLKG(parsed);
            
            return parsed;
        } catch (error) {
            console.error(`[MCPConfigStore] Failed to read primary config at ${this.configPath}, falling back to LKG:`, error);
            return this.readLKG();
        }
    }

    public readLKG(): Record<string, MCPServerConfig> {
        if (!fs.existsSync(this.lkgPath)) {
            return {};
        }

        try {
            const raw = fs.readFileSync(this.lkgPath, 'utf8');
            return JSON.parse(raw) as Record<string, MCPServerConfig>;
        } catch (error) {
            console.error(`[MCPConfigStore] Failed to read LKG config at ${this.lkgPath}:`, error);
            return {};
        }
    }

    public writeAll(config: Record<string, MCPServerConfig>): void {
        const dir = path.dirname(this.configPath);
        if (!fs.existsSync(dir)) {
            fs.mkdirSync(dir, { recursive: true });
        }

        fs.writeFileSync(this.configPath, JSON.stringify(config, null, 2));
        this.writeLKG(config);
    }

    private writeLKG(config: Record<string, MCPServerConfig>): void {
        const dir = path.dirname(this.lkgPath);
        if (!fs.existsSync(dir)) {
            fs.mkdirSync(dir, { recursive: true });
        }
        fs.writeFileSync(this.lkgPath, JSON.stringify(config, null, 2));
    }

    public upsert(name: string, config: MCPServerConfig): void {
        const current = this.readAll();
        current[name] = config;
        this.writeAll(current);
    }

    public remove(name: string): void {
        const current = this.readAll();
        delete current[name];
        this.writeAll(current);
    }
}
