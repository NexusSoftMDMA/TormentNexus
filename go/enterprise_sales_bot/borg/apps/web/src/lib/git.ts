import fs from 'fs/promises';
import path from 'path';
import { exec } from 'child_process';
import util from 'util';

const execAsync = util.promisify(exec);

export interface SubmoduleInfo {
    path: string;
    url: string;
    status: 'clean' | 'dirty' | 'missing' | 'error';
    branch?: string;
    lastCommit?: string;
    lastCommitDate?: string;
    lastCommitMessage?: string;
    version?: string;
    pkgName?: string;
}

export async function getSubmodules(workspaceRoot: string): Promise<SubmoduleInfo[]> {
    const gitmodulesPath = path.join(workspaceRoot, '.gitmodules');

    try {
        const content = await fs.readFile(gitmodulesPath, 'utf-8');
        const submodules: SubmoduleInfo[] = [];

        // Regex to parse .gitmodules
        // [submodule "path/to/sub"]
        // 	path = path/to/sub
        // 	url = ...

        const lines = content.split('\n');
        let currentPath = '';
        let currentUrl = '';

        for (const line of lines) {
            if (line.trim().startsWith('path = ')) {
                currentPath = line.trim().substring(7).trim();
            } else if (line.trim().startsWith('url = ')) {
                currentUrl = line.trim().substring(6).trim();

                if (currentPath && currentUrl) {
                    submodules.push({
                        path: currentPath,
                        url: currentUrl,
                        status: 'missing' // Default
                    });
                    currentPath = '';
                    currentUrl = '';
                }
            }
        }

        // Now check status for each (in parallel chunks to likely improve perf but not kill CPU)
        // Limiting concurrency is good practice.
        const results = await Promise.all(submodules.map(async sub => {
            const fullPath = path.join(workspaceRoot, sub.path);
            return checkSubmoduleStatus(fullPath, sub);
        }));

        return results;

    } catch {
        // console.error("Error reading .gitmodules:", e); // Silent fail safe
        return [];
    }
}

async function checkSubmoduleStatus(fullPath: string, info: SubmoduleInfo): Promise<SubmoduleInfo> {
    try {
        // Check if directory exists
        try {
            await fs.access(fullPath);
        } catch {
            return { ...info, status: 'missing' };
        }

        // Check git status
        const { stdout } = await execAsync('git status --porcelain', { cwd: fullPath });
        const isDirty = stdout.trim().length > 0;

        // Get Head
        const { stdout: head } = await execAsync('git rev-parse --short HEAD', { cwd: fullPath });

        // Get Date
        const { stdout: date } = await execAsync('git log -1 --format=%cd --date=relative', { cwd: fullPath });

        // Get Message
        const { stdout: message } = await execAsync('git log -1 --format=%s', { cwd: fullPath });

        // Try to read package.json for version/name
        let version = 'unknown';
        let pkgName = '';
        try {
            const pkgPath = path.join(fullPath, 'package.json');
            const pkgContent = await fs.readFile(pkgPath, 'utf-8');
            const pkg = JSON.parse(pkgContent);
            version = pkg.version || 'unknown';
            pkgName = pkg.name || '';
        } catch {
            // Not a Node.js repo or no package.json
        }

        return {
            ...info,
            status: isDirty ? 'dirty' : 'clean',
            lastCommit: head.trim(),
            lastCommitDate: date.trim(),
            lastCommitMessage: message.trim(),
            version,
            pkgName
        };

    } catch (e) {
        return { ...info, status: 'error' };
    }
}
