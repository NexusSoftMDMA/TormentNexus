import { spawnAsync } from '../utils/exec.js';
import fs from 'fs';
import path from 'path';

import { MCPAggregator } from '../mcp/MCPAggregator.js';

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
            const statusRes = await spawnAsync('git', ['submodule', 'status'], { cwd: this.rootDir });
            const configRes = await spawnAsync('git', ['config', '--file', '.gitmodules', '--get-regexp', 'path'], { cwd: this.rootDir });
            const urlRes = await spawnAsync('git', ['config', '--file', '.gitmodules', '--get-regexp', 'url'], { cwd: this.rootDir });

            const submodules: SubmoduleStatus[] = [];
            const lines = statusRes.stdout.trim().split('\n');

            for (const line of lines) {
                if (!line.trim()) continue;

                const match = line.match(/^([ +-])([0-9a-f]+)\s+(.+?)(?:\s+\((.+)\))?$/);
                if (match) {
                    const [, indicator, commit, subPath, branch] = match;
                    let status: SubmoduleStatus['status'] = 'clean';

                    if (indicator === '+') status = 'out-of-sync';
                    if (indicator === '-') status = 'missing';

                    const url = this.extractUrl(subPath, urlRes.stdout);
                    const fullPath = path.join(this.rootDir, subPath);
                    const { caps, startCommand } = this.detectCapabilities(subPath);
                    const isInstalled = fs.existsSync(path.join(fullPath, 'node_modules')) || fs.existsSync(path.join(fullPath, '.venv'));

                    submodules.push({
                        name: subPath.split('/').pop() || subPath,
                        path: subPath,
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
        const lines = configOutput.split('\n');
        for (const line of lines) {
            if (line.includes(submodulePath)) {
                const parts = line.split(' ');
                if (parts.length > 1) return parts[1];
            }
        }
        return undefined;
    }

    public async updateAll(): Promise<{ success: boolean; output: string }> {
        try {
            const result = await spawnAsync('git', ['submodule', 'update', '--init', '--recursive', '--remote'], { cwd: this.rootDir });
            return { success: true, output: result.stdout + result.stderr };
        } catch (error) {
            return { success: false, output: String(error) };
        }
    }

    public async installDependencies(submodulePath: string): Promise<{ success: boolean; output: string }> {
        const fullPath = path.join(this.rootDir, submodulePath);
        const isWin = process.platform === 'win32';
        try {
            if (fs.existsSync(path.join(fullPath, 'package.json'))) {
                const npm = isWin ? 'npm.cmd' : 'npm';
                const result = await spawnAsync(npm, ['install'], { cwd: fullPath });
                return { success: true, output: result.stdout + result.stderr };
            }
            if (fs.existsSync(path.join(fullPath, 'requirements.txt'))) {
                const pip = isWin ? 'pip.exe' : 'pip';
                const result = await spawnAsync(pip, ['install', '-r', 'requirements.txt'], { cwd: fullPath });
                return { success: true, output: result.stdout + result.stderr };
            }
            return { success: false, output: "No known package manager found (package.json or requirements.txt)" };
        } catch (error) {
            return { success: false, output: String(error) };
        }
    }

    public async buildSubmodule(submodulePath: string): Promise<{ success: boolean; output: string }> {
        const fullPath = path.join(this.rootDir, submodulePath);
        const isWin = process.platform === 'win32';
        try {
            if (fs.existsSync(path.join(fullPath, 'package.json'))) {
                const pkg = JSON.parse(fs.readFileSync(path.join(fullPath, 'package.json'), 'utf-8'));
                if (pkg.scripts && pkg.scripts.build) {
                    const npm = isWin ? 'npm.cmd' : 'npm';
                    const result = await spawnAsync(npm, ['run', 'build'], { cwd: fullPath });
                    return { success: true, output: result.stdout + result.stderr };
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
            const result = await spawnAsync('git', ['submodule', 'init', submodulePath], { cwd: this.rootDir });
            return { success: true, output: result.stdout + result.stderr };
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

                if (pkg.keywords && pkg.keywords.includes('mcp-server')) caps.push('mcp-server');
                if (pkg.dependencies && pkg.dependencies['@modelcontextprotocol/sdk']) caps.push('mcp-sdk');

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
