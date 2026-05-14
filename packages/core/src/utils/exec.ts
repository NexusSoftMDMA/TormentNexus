import { spawn, SpawnOptions, ExecOptions } from "child_process";

export interface ExecResult {
    stdout: string;
    stderr: string;
    exitCode: number | null;
}

/**
 * Hardened execution utility that uses spawn with explicit argument arrays
 * to prevent shell injection.
 */
export function spawnAsync(
    command: string,
    args: string[] = [],
    options: SpawnOptions = {}
): Promise<ExecResult> {
    return new Promise((resolve, reject) => {
        let stdout = "";
        let stderr = "";

        const child = spawn(command, args, {
            ...options,
            shell: false, // Explicitly disable shell
        });

        child.stdout?.on("data", (data) => {
            stdout += data.toString();
        });

        child.stderr?.on("data", (data) => {
            stderr += data.toString();
        });

        child.on("error", (err) => {
            reject(err);
        });

        child.on("close", (code) => {
            resolve({
                stdout,
                stderr,
                exitCode: code,
            });
        });
    });
}

/**
 * Legacy compatibility layer.
 * @deprecated Use spawnAsync instead.
 */
export async function execAsync(command: string, options: ExecOptions = {}): Promise<{ stdout: string; stderr: string }> {
    const parts = command.split(" ");
    const cmd = parts[0];
    const args = parts.slice(1);

    const result = await spawnAsync(cmd, args, options as SpawnOptions);
    if (result.exitCode !== 0) {
        throw new Error(`Command failed: ${command}\n${result.stderr}`);
    }
    return { stdout: result.stdout, stderr: result.stderr };
}
