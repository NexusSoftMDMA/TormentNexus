
import { spawn } from 'child_process';
import { v4 as uuidv4 } from 'uuid';

export class CodeExecutorService {
    /**
     * Executes Python code in a separate process.
     * @param code The Python code to execute.
     * @param toolCallHandler Optional callback to handle tool calls from within the Python script (not fully implemented in this MVP).
     * @returns The stdout of the execution or throws an error with stderr.
     */
    async executeCode(code: string, _toolCallHandler?: (name: string, args: unknown) => Promise<unknown>): Promise<string> {
        return new Promise((resolve, reject) => {
            const pythonProcess = spawn('python', ['-c', code]);

            let stdout = '';
            let stderr = '';

            pythonProcess.stdout.on('data', (data) => {
                stdout += data.toString();
            });

            pythonProcess.stderr.on('data', (data) => {
                stderr += data.toString();
            });

            pythonProcess.on('close', (code) => {
                if (code !== 0) {
                    reject(new Error(`Python execution failed with code ${code}:\n${stderr}`));
                } else {
                    resolve(stdout);
                }
            });

            pythonProcess.on('error', (err) => {
                reject(new Error(`Failed to start Python process: ${err.message}`));
            });
        });
    }
}

export const codeExecutorService = new CodeExecutorService();
