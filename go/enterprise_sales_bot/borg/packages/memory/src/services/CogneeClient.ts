import { spawn } from 'child_process';
import * as path from 'path';

export interface CogneeResponse {
    status?: string;
    results?: string[];
    error?: string;
}

export class CogneeClient {
    private scriptPath: string;

    constructor() {
        // Resolve path to the bridge script
        // Assuming this file is compiled to dist/services/CogneeClient.js
        // and scripts are in ../../scripts/ from src (or ../scripts dist layout)
        // If running via ts-node/tsx (source), __dirname is src/services
        // script is at ../../scripts/cognee_bridge.py
        this.scriptPath = path.resolve(__dirname, '../../scripts/cognee_bridge.py');
    }

    private async execute(command: string, payload: any): Promise<CogneeResponse> {
        return new Promise((resolve, reject) => {
            const script = path.basename(this.scriptPath);
            const dir = path.dirname(this.scriptPath);

            // Log for debugging
            // console.log(`[CogneeClient] Spawning python ${script} in ${dir}`);

            const py = spawn('python', [script], { cwd: dir });

            let stdout = '';
            let stderr = '';

            py.stdout.on('data', (data) => {
                stdout += data.toString();
            });

            py.stderr.on('data', (data) => {
                stderr += data.toString();
            });

            // Send payload as NDJSON
            try {
                py.stdin.write(JSON.stringify({ command, payload }) + '\n');
                py.stdin.end();
            } catch (e) {
                reject(new Error(`Failed to write to python process: ${e}`));
                return;
            }

            py.on('close', (code) => {
                if (code !== 0) {
                    reject(new Error(`Python process exited with code ${code}. Stderr: ${stderr}`));
                } else {
                    try {
                        const lines = stdout.trim().split('\n');
                        // Find the last valid JSON line
                        let result = null;
                        for (let i = lines.length - 1; i >= 0; i--) {
                            try {
                                const line = lines[i].trim();
                                if (line) {
                                    result = JSON.parse(line);
                                    break;
                                }
                            } catch (e) { }
                        }

                        if (result) {
                            resolve(result as CogneeResponse);
                        } else {
                            // If no JSON found, perhaps it was just empty generic output?
                            reject(new Error(`No valid JSON output from Cognee bridge. Stdout: ${stdout}`));
                        }
                    } catch (e) {
                        reject(new Error(`Failed to parse output: ${e}. Stdout: ${stdout}`));
                    }
                }
            });

            py.on('error', (err) => {
                reject(err);
            });
        });
    }

    public async add(text: string, dataset: string = "tormentnexus_memory"): Promise<void> {
        const response = await this.execute('add', { text, dataset });
        if (response.error) throw new Error(`Cognee Add Error: ${response.error}`);
    }

    public async cognify(dataset: string = "tormentnexus_memory"): Promise<void> {
        const response = await this.execute('cognify', { dataset });
        if (response.error) throw new Error(`Cognee Cognify Error: ${response.error}`);
    }

    public async search(query: string, type: string = "INSIGHTS"): Promise<string[]> {
        const response = await this.execute('search', { query, type });
        if (response.error) throw new Error(`Cognee Search Error: ${response.error}`);
        return response.results || [];
    }
}
