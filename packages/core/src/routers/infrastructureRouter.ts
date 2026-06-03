import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { exec } from 'node:child_process';
import { promisify } from 'node:util';
import fs from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';

const execAsync = promisify(exec);
const LEGACY_INFRA_BINARY = ['mcp', 'enetes'].join('');
const INFRA_BINARY = process.env.TORMENTNEXUS_INFRA_BINARY?.trim() || LEGACY_INFRA_BINARY;
const INFRA_SUBMODULE_DIR = process.env.TORMENTNEXUS_INFRA_SUBMODULE?.trim() || LEGACY_INFRA_BINARY;

export const infrastructureRouter = t.router({
    /**
     * Get the current status of the TormentNexus infrastructure daemon / binary.
     */
    getInfrastructureStatus: publicProcedure.query(async () => {
        try {
            const binPath = path.join(process.cwd(), '..', '..', 'submodules', INFRA_SUBMODULE_DIR, 'bin', INFRA_BINARY);

            let isInstalled = false;
            try {
                await fs.access(binPath);
                isInstalled = true;
            } catch {
                try {
                    await execAsync(`${INFRA_BINARY} --version`);
                    isInstalled = true;
                } catch {
                    isInstalled = false;
                }
            }

            // Check config files
            const configPath = path.join(os.homedir(), '.config', 'mcpetes', 'config.yaml');
            let hasConfig = false;
            try {
                await fs.access(configPath);
                hasConfig = true;
            } catch {
                hasConfig = false;
            }

            return {
                installed: isInstalled,
                hasConfig,
                daemonActive: false, // Would check port 3000/3010 if running UI
                version: isInstalled ? "latest" : null
            };
        } catch (error) {
            return {
                installed: false,
                hasConfig: false,
                daemonActive: false,
                version: null,
                error: (error as Error).message
            };
        }
    }),

    /**
     * Run the infrastructure health check command.
     */
    runDoctor: adminProcedure.mutation(async () => {
        try {
            const { stdout, stderr } = await execAsync(`${INFRA_BINARY} doctor`);
            return { success: true, output: stdout || stderr };
        } catch (error: any) {
            return { success: false, output: error.stdout || error.stderr || error.message };
        }
    }),

    /**
     * Apply configurations across all clients
     */
    applyConfigurations: adminProcedure.mutation(async () => {
        try {
            const { stdout, stderr } = await execAsync(`${INFRA_BINARY} apply`);
            return { success: true, output: stdout || stderr };
        } catch (error: any) {
            return { success: false, output: error.stdout || error.stderr || error.message };
        }
    }),
});
