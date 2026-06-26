
import { spawn } from 'child_process';
import path from 'path';

export class MergeService {
    constructor(private repoRoot: string) { }

    async mergeBranch(branchName: string): Promise<{ success: boolean; message: string; conflict?: boolean }> {
        console.log(`[MergeService] Attempting to merge ${branchName} into main...`);

        // 1. Git fetch/status
        // Assumes we are in Main repo or can run git commands targeting it
        const runGit = (args: string[]) => new Promise<{ stdout: string, stderr: string, code: number }>((resolve) => {
            const proc = spawn('git', args, { cwd: this.repoRoot });
            let stdout = '';
            let stderr = '';
            proc.stdout.on('data', d => stdout += d.toString());
            proc.stderr.on('data', d => stderr += d.toString());
            proc.on('close', code => resolve({ stdout, stderr, code: code || 0 }));
        });

        // Checkout main (ensure we are on main)
        await runGit(['checkout', 'main']);
        await runGit(['pull', 'origin', 'main']); // Optional

        // Merge
        const result = await runGit(['merge', branchName, '--no-ff', '-m', `Merge squad ${branchName}`]);

        if (result.code === 0) {
            return { success: true, message: `Successfully merged ${branchName}\n${result.stdout}` };
        } else {
            // Conflict?
            if (result.stdout.includes('CONFLICT') || result.stderr.includes('CONFLICT')) {
                // Abort merge to leave clean state
                await runGit(['merge', '--abort']);
                return { success: false, message: `Merge Conflict detected. Aborted.\n${result.stdout}`, conflict: true };
            }
            return { success: false, message: `Merge Failed: ${result.stderr}` };
        }
    }
}
