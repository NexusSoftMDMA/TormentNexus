import { spawnAsync } from "../utils/exec.js";
import path from 'path';

function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

export interface CommitLog {
    hash: string;
    author: string;
    date: string;
    message: string;
}

export interface GitStatus {
    branch: string;
    clean: boolean;
    modified: string[];
    staged: string[];
}

export class GitService {
    private cwd: string;

    constructor(cwd: string) {
        this.cwd = cwd;
    }

    private async run(args: string[]): Promise<string> {
        const result = await spawnAsync("git", args, { cwd: this.cwd });
        if (result.exitCode !== 0) {
            throw new Error(`Git command failed: git ${args.join(" ")}\n${result.stderr}`);
        }
        return result.stdout.trim();
    }

    async getLog(limit: number = 20): Promise<CommitLog[]> {
        try {
            // Format: %H|%an|%aI|%s
            const out = await this.run(["log", "-n", String(limit), "--pretty=format:%H|%an|%aI|%s"]);
            return out.split('\n').filter(Boolean).map(line => {
                const [hash, author, date, message] = line.split('|');
                return { hash, author, date, message };
            });
        } catch (e) {
            return [];
        }
    }

    async getStatus(): Promise<GitStatus> {
        try {
            const branch = await this.run(['rev-parse', '--abbrev-ref', 'HEAD']);
            const statusOut = await this.run(['status', '--porcelain']);

            const modified: string[] = [];
            const staged: string[] = [];

            statusOut.split('\n').filter(Boolean).forEach(line => {
                const code = line.substring(0, 2);
                const file = line.substring(3);
                if (code.includes('M') || code.includes('?')) modified.push(file);
                if (code.includes('A') || (code.includes('M') && code[0] !== ' ')) staged.push(file);
            });

            return {
                branch,
                clean: statusOut.length === 0,
                modified,
                staged
            };
        } catch (e) {
            return { branch: 'unknown', clean: false, modified: [], staged: [] };
        }
    }

    async revert(hash: string): Promise<string> {
        try {
            const r = await this.run(['revert', '--no-edit', hash]);
            return r;
        } catch (e: unknown) {
            throw new Error(`Failed to revert: ${getErrorMessage(e)}`);
        }
    }

    async resetTo(hash: string, mode: 'soft' | 'hard' = 'soft'): Promise<string> {
        try {
            const r = await this.run(['reset', `--${mode}`, hash]);
            return r;
        } catch (e: unknown) {
            throw new Error(`Failed to reset: ${getErrorMessage(e)}`);
        }
    }
}
