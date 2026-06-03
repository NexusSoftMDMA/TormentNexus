import fs from 'fs';
import path from 'path';

type ConfigData = Record<string, unknown> & {
    council?: {
        personas: string[];
        contextFiles: string[];
        prompts: {
            Architect: string;
            Product: string;
            Critic: string;
        };
    };
    mcpServers?: unknown[];
};

export class ConfigManager {
    private configPath: string;

    constructor() {
        // Use user home or workspace root? 
        // MCPServer uses process.cwd()/.tormentnexus/skills
        // Let's use process.cwd()/.tormentnexus/config.json
        this.configPath = path.join(process.cwd(), '.tormentnexus', 'config.json');
    }

    loadConfig(): ConfigData | null {
        try {
            if (fs.existsSync(this.configPath)) {
                const content = fs.readFileSync(this.configPath, 'utf-8');
                const parsed = JSON.parse(content) as unknown;
                if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
                    return parsed as ConfigData;
                }
            }
        } catch (e) {
            console.error("[ConfigManager] Failed to load config:", e);
        }
        return null;
    }

    saveConfig(config: Record<string, unknown>) {
        try {
            const dir = path.dirname(this.configPath);
            if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
            fs.writeFileSync(this.configPath, JSON.stringify(config, null, 2));
            console.log("[ConfigManager] Saved config to", this.configPath);
        } catch (e) {
            console.error("[ConfigManager] Failed to save config:", e);
        }
    }
}
