import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export class Installer {
    private configPath: string;

    constructor(configPath?: string) {
        // Default to the path provided by the user if not specified
        this.configPath = configPath || 'C:\\Users\\hyper\\AppData\\Roaming\\Antigravity\\User\\mcp.json';
    }

    async install(): Promise<string> {
        console.error(`Attempting to install TormentNexus Supervisor to: ${this.configPath}`);

        try {
            await fs.access(this.configPath);
        } catch {
            throw new Error(`Antigravity config not found at: ${this.configPath}`);
        }

        const content = await fs.readFile(this.configPath, 'utf-8');
        const config = JSON.parse(content);

        // Calculate the absolute path to our own executable (dist/index.js)
        const scriptPath = path.resolve(__dirname, 'index.js');

        config.servers = config.servers || {};
        config.servers['tormentnexus-supervisor'] = {
            command: 'node',
            args: [scriptPath],
            env: {
                NODE_ENV: 'production'
            },
            disabled: false,
            autoApprove: []
        };

        await fs.writeFile(this.configPath, JSON.stringify(config, null, 4), 'utf-8');
        return `Successfully installed 'tormentnexus-supervisor' to ${this.configPath}`;
    }
}
