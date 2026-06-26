
import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs';
import path from 'path';

import { MCPAggregator } from '../mcp/MCPAggregator.js';

const execAsync = promisify(exec);

export interface SubmoduleStatus {
    name: string;
    path: string;
    commit: string;
    branch?: string;
    status: 'clean' | 'modified' | 'out-of-sync' | 'missing';
    url?: string;
    capabilities?: string[];
    isInstalled?: boolean;
    isBuilt?: boolean;
    startCommand?: string;
}

export class SubmoduleService {
    private rootDir: string;
    private mcpAggregator?: MCPAggregator;

    constructor(rootDir: string = process.cwd(), mcpAggregator?: MCPAggregator) {
        this.rootDir = rootDir;
        this.mcpAggregator = mcpAggregator;
    }

    public async listSubmodules(): Promise<SubmoduleStatus[]> {
        const gitModulesPath = path.join(this.rootDir, '.gitmodules');
        if (!fs.existsSync(gitModulesPath)) {
            return [];
        }

        try {
            const { stdout: statusOutput } = await execAsync('git submodule status', { cwd: this.rootDir });
            const { stdout: configOutput } = await execAsync('git config --file .gitmodules --get-regexp path', { cwd: this.rootDir });
            const { stdout: urlOutput } = await execAsync('git config --file .gitmodules --get-regexp url', { cwd: this.rootDir });

            const submodules: SubmoduleStatus[] = [];
            const lines = statusOutput.trim().split('\n');

            for (const line of lines) {
                if (!line.trim()) continue;

                const match = line.match(/^([ +-])([0-9a-f]+)\s+(.+?)(?:\s+\((.+)\))?$/);
                if (match) {
                    const [, indicator, commit, path, branch] = match;
                    let status: SubmoduleStatus['status'] = 'clean';

                    if (indicator === '+') status = 'out-of-sync';
                    if (indicator === '-') status = 'missing';
                    // We'd need `git status --porcelain` inside the submodule to check for 'modified' accurately without being expensive

                    // Find URL
                    const url = this.extractUrl(path, urlOutput);
                    const fullPath = require('path').join(this.rootDir, path);
                    const { caps, startCommand } = this.detectCapabilities(path);
                    const isInstalled = fs.existsSync(require('path').join(fullPath, 'node_modules')) || fs.existsSync(require('path').join(fullPath, '.venv')); // rough check

                    submodules.push({
                        name: path.split('/').pop() || path,
                        path,
                        commit,
                        branch: branch || 'HEAD',
                        status,
                        url,
                        capabilities: caps,
                        isInstalled,
                        startCommand
                    });
                }
            }
            return submodules;
        } catch (error) {
            console.error('Failed to list submodules:', error);
            const message = error instanceof Error ? error.message : String(error);
            throw new Error(message);
        }
    }

    private extractUrl(submodulePath: string, configOutput: string): string | undefined {
        // configOutput lines look like: submodule.packages/foo.url https://...
        const lines = configOutput.split('\n');
        // This is a naive match, robust implementation would match the submodule section name
        for (const line of lines) {
            if (line.includes(submodulePath)) { // simplified
                const parts = line.split(' ');
                if (parts.length > 1) return parts[1];
            }
        }
        return undefined;
    }

    public async updateAll(): Promise<{ success: boolean; output: string }> {
        try {
            const { stdout, stderr } = await execAsync('git submodule update --init --recursive --remote', { cwd: this.rootDir });
            return { success: true, output: stdout + stderr };
        } catch (error) {
            return { success: false, output: String(error) };
        }
    }
    public async installDependencies(submodulePath: string): Promise<{ success: boolean; output: string }> {
        const fullPath = path.join(this.rootDir, submodulePath);
        try {
            // Check for package.json
            if (fs.existsSync(path.join(fullPath, 'package.json'))) {
                const { stdout, stderr } = await execAsync('npm install', { cwd: fullPath });
                return { success: true, output: stdout + stderr };
            }
            // Check for requirements.txt (Python)
            if (fs.existsSync(path.join(fullPath, 'requirements.txt'))) {
                const { stdout, stderr } = await execAsync('pip install -r requirements.txt', { cwd: fullPath });
                return { success: true, output: stdout + stderr };
            }
            return { success: false, output: "No known package manager found (package.json or requirements.txt)" };
        } catch (error) {
            return { success: false, output: String(error) };
        }
    }

    public async buildSubmodule(submodulePath: string): Promise<{ success: boolean; output: string }> {
        const fullPath = path.join(this.rootDir, submodulePath);
        try {
            if (fs.existsSync(path.join(fullPath, 'package.json'))) {
                const pkg = JSON.parse(fs.readFileSync(path.join(fullPath, 'package.json'), 'utf-8'));
                if (pkg.scripts && pkg.scripts.build) {
                    const { stdout, stderr } = await execAsync('npm run build', { cwd: fullPath });
                    return { success: true, output: stdout + stderr };
                }
                return { success: true, output: "No build script found, assuming raw source is fine." };
            }
            return { success: false, output: "No package.json found." };
        } catch (error) {
            return { success: false, output: String(error) };
        }
    }

    public async enableSubmodule(submodulePath: string): Promise<{ success: boolean; output: string }> {
        try {
            const { stdout, stderr } = await execAsync(`git submodule init ${submodulePath}`, { cwd: this.rootDir });
            return { success: true, output: stdout + stderr };
        } catch (error) {
            return { success: false, output: String(error) };
        }
    }

    public detectCapabilities(submodulePath: string): { caps: string[], startCommand?: string } {
        const fullPath = path.join(this.rootDir, submodulePath);
        const caps: string[] = [];
        let startCommand: string | undefined;

        try {
            if (fs.existsSync(path.join(fullPath, 'package.json'))) {
                const pkg = JSON.parse(fs.readFileSync(path.join(fullPath, 'package.json'), 'utf-8'));

                // 1. Detect MCP
                if (pkg.keywords && pkg.keywords.includes('mcp-server')) {
                    caps.push('mcp-server');
                }
                if (pkg.dependencies && pkg.dependencies['@modelcontextprotocol/sdk']) {
                    caps.push('mcp-sdk');
                }

                // 2. Detect Start Command
                if (pkg.scripts && pkg.scripts.start) {
                    startCommand = 'npm start';
                } else if (pkg.bin) {
                    if (typeof pkg.bin === 'string') {
                        startCommand = `node ${pkg.bin}`;
                    } else if (typeof pkg.bin === 'object') {
                        const firstBin = Object.values(pkg.bin)[0];
                        if (firstBin) startCommand = `node ${firstBin}`;
                    }
                } else if (pkg.main) {
                    startCommand = `node ${pkg.main}`;
                }
            } else if (fs.existsSync(path.join(fullPath, 'requirements.txt'))) {
                // Python detection
                caps.push('python');
                if (fs.existsSync(path.join(fullPath, 'main.py'))) {
                    startCommand = 'python main.py';
                } else if (fs.existsSync(path.join(fullPath, 'app.py'))) {
                    startCommand = 'python app.py';
                }
            }
        } catch (e) {
            // ignore
        }
        return { caps, startCommand };
    }
}
