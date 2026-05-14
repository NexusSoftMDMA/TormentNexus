import { spawnAsync } from "../utils/exec.js";
import path from "path";
import fs from "fs";

interface WorktreeInfo {
    path: string;
    head?: string;
    branch?: string;
}

export class GitWorktreeManager {
    constructor(private rootDir: string) { }

    private async runGit(args: string[]): Promise<string> {
        const result = await spawnAsync("git", args, { cwd: this.rootDir });
        if (result.exitCode !== 0) {
            throw new Error(`Git command failed: git ${args.join(" ")}\n${result.stderr}`);
        }
        return result.stdout.trim();
    }

    async listWorktrees(): Promise<WorktreeInfo[]> {
        const stdout = await this.runGit(["worktree", "list", "--porcelain"]);
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
        let exists = false;
        try {
            await this.runGit(['show-ref', '--verify', `refs/heads/${branch}`]);
            exists = true;
        } catch (e: unknown) {
            // Branch doesn't exist
        }

        const args = exists
            ? ['worktree', 'add', fullPath, branch]
            : ['worktree', 'add', '-b', branch, fullPath];

        console.log(`[GitWorktree] Adding worktree: git ${args.join(" ")}`);
        await this.runGit(args);
        return fullPath;
    }

    async removeWorktree(pathOrBranch: string, force: boolean = false): Promise<void> {
        const args = ['worktree', 'remove', pathOrBranch];
        if (force) args.push("--force");

        console.log(`[GitWorktree] Removing worktree: git ${args.join(" ")}`);
        await this.runGit(args);
    }

    async createTaskEnvironment(taskId: string): Promise<string> {
        const branchName = `task/${taskId}`;
        const relativePath = `.borg/worktrees/${taskId}`;
        console.log(`[GitWorktree] Creating task environment: ${taskId} at ${relativePath}`);
        return this.addWorktree(branchName, relativePath);
    }

    async cleanupTaskEnvironment(taskId: string): Promise<void> {
        const relativePath = `.borg/worktrees/${taskId}`;
        const fullPath = path.resolve(this.rootDir, relativePath);
        console.log(`[GitWorktree] Cleaning up task environment: ${taskId}`);
        await this.removeWorktree(fullPath, true);
    }
}
