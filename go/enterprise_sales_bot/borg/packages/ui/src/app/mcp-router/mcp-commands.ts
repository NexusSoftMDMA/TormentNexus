/**
 * TORMENTNEXUS MCP Router - CLI Integration
 *
 * Integrates MCP Router CLI commands into the main tormentnexus CLI
 */

import { spawn } from 'child_process';

interface SpawnResult {
    stdout: string;
    stderr: string;
}

function execAsync(command: string, args: string[], options: any): Promise<SpawnResult> {
    return new Promise((resolve, reject) => {
        const child = spawn(command, args, options);
        let stdout = '';
        let stderr = '';

        if (child.stdout) {
            child.stdout.on('data', (data) => {
                stdout += data.toString();
            });
        }

        if (child.stderr) {
            child.stderr.on('data', (data) => {
                stderr += data.toString();
            });
        }

        child.on('close', (code) => {
            if (code === 0) {
                resolve({ stdout, stderr });
            } else {
                reject(new Error(`Command failed with code ${code}: ${stderr}`));
            }
        });

        child.on('error', (err) => {
            reject(err);
        });
    });
}

const CLI_PATH = process.env.MCP_ROUTER_CLI || './cli/mcp-router-cli/dist/mcp-router-cli.js';


