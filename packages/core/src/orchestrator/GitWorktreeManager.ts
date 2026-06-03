
import { execAsync } from "../utils/exec.js";
import path from "path";
import fs from "fs";

interface WorktreeInfo {
    path: string;
    head?: string;
    branch?: string;
}

export class GitWorktreeManager {
    constructor(private rootDir: string) { }

    async listWorktrees(): Promise<WorktreeInfo[]> {
        const { stdout } = await execAsync("git worktree list --porcelain", { cwd: this.rootDir });
        const worktrees: WorktreeInfo[] = [];
        let current: Partial<WorktreeInfo> = {};

        stdout.split('\n').forEach(line => {
            if (line.startsWith('worktree ')) {
                if (current.path) worktrees.push(current as WorktreeInfo);
                current = { path: line.substring(9).trim() };
            } else if (line.startsWith('HEAD ')) {
                current.head = line.substring(5).trim();
            } else if (line.startsWith('branch ')) {
                current.branch = line.substring(7).trim();
            }
        });
        if (current.path) worktrees.push(current as WorktreeInfo);
        return worktrees;
    }

    async addWorktree(branch: string, relativePath: string): Promise<string> {
        const fullPath = path.resolve(this.rootDir, relativePath);

        // Ensure parent dir exists
        fs.mkdirSync(path.dirname(fullPath), { recursive: true });

        // Check if branch exists
        let command = `git worktree add ${fullPath} ${branch}`;
        try {
            await execAsync(`git show-ref --verify refs/heads/${branch}`, { cwd: this.rootDir });
        } catch (e: unknown) {
            // Branch doesn't exist, create it
            command = `git worktree add -b ${branch} ${fullPath}`;
        }

        console.log(`[GitWorktree] Adding worktree: ${command}`);
        await execAsync(command, { cwd: this.rootDir });
        return fullPath;
    }

    async removeWorktree(pathOrBranch: string, force: boolean = false): Promise<void> {
        let command = `git worktree remove ${pathOrBranch}`;
        if (force) command += " --force";

        console.log(`[GitWorktree] Removing worktree: ${command}`);
        await execAsync(command, { cwd: this.rootDir });
    }

    async createTaskEnvironment(taskId: string): Promise<string> {
        const branchName = `task/${taskId}`;
        const relativePath = `.tormentnexus/worktrees/${taskId}`;
        console.log(`[GitWorktree] Creating task environment: ${taskId} at ${relativePath}`);
        return this.addWorktree(branchName, relativePath);
    }

    async cleanupTaskEnvironment(taskId: string): Promise<void> {
        const relativePath = `.tormentnexus/worktrees/${taskId}`;
        const fullPath = path.resolve(this.rootDir, relativePath);
        console.log(`[GitWorktree] Cleaning up task environment: ${taskId}`);
        await this.removeWorktree(fullPath, true);
    }
}
