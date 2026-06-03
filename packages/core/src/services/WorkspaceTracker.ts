import fs from 'fs/promises';
import path from 'path';
import os from 'os';

export interface WorkspaceRecord {
    path: string;
    lastAccessedAt: number;
    name: string;
}

export class WorkspaceTracker {
    private readonly registryPath: string;

    constructor() {
        this.registryPath = path.join(os.homedir(), '.tormentnexus', 'workspaces.json');
    }

    public async registerWorkspace(workspacePath: string = process.cwd()): Promise<void> {
        try {
            await fs.mkdir(path.dirname(this.registryPath), { recursive: true });
            
            let workspaces: WorkspaceRecord[] = [];
            try {
                const data = await fs.readFile(this.registryPath, 'utf8');
                workspaces = JSON.parse(data) as WorkspaceRecord[];
            } catch {
                // File doesn't exist or is invalid
            }

            // Remove existing entry if it exists
            workspaces = workspaces.filter(w => w.path !== workspacePath);

            // Add to the front
            workspaces.unshift({
                path: workspacePath,
                name: path.basename(workspacePath),
                lastAccessedAt: Date.now(),
            });

            // Keep top 50 workspaces
            workspaces = workspaces.slice(0, 50);

            await fs.writeFile(this.registryPath, JSON.stringify(workspaces, null, 2), 'utf8');
            console.log(`[WorkspaceTracker] Registered workspace: ${workspacePath}`);
        } catch (error) {
            console.error('[WorkspaceTracker] Failed to register workspace:', error);
        }
    }

    public async listWorkspaces(): Promise<WorkspaceRecord[]> {
        try {
            const data = await fs.readFile(this.registryPath, 'utf8');
            const workspaces = JSON.parse(data) as WorkspaceRecord[];
            
            // Verify they still exist on disk
            const validWorkspaces: WorkspaceRecord[] = [];
            for (const ws of workspaces) {
                try {
                    const stat = await fs.stat(ws.path);
                    if (stat.isDirectory()) {
                        validWorkspaces.push(ws);
                    }
                } catch {
                    // Directory was deleted or moved
                }
            }
            return validWorkspaces;
        } catch {
            return [];
        }
    }
}

export const workspaceTracker = new WorkspaceTracker();
