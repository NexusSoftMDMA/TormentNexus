import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { spawnAsync } from '../utils/exec.js';
import fs from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';

const LEGACY_INFRA_BINARY = ['mcp', 'enetes'].join('');
const INFRA_BINARY = process.env.BORG_INFRA_BINARY?.trim() || LEGACY_INFRA_BINARY;
const INFRA_SUBMODULE_DIR = process.env.BORG_INFRA_SUBMODULE?.trim() || LEGACY_INFRA_BINARY;

export const infrastructureRouter = t.router({
    /**
     * Get the current status of the Borg infrastructure daemon / binary.
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
                    const result = await spawnAsync(INFRA_BINARY, ['--version']);
                    isInstalled = result.exitCode === 0;
                } catch {
                    isInstalled = false;
                }
            }

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
                daemonActive: false,
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
            const result = await spawnAsync(INFRA_BINARY, ['doctor']);
            return { success: result.exitCode === 0, output: result.stdout || result.stderr };
        } catch (error: any) {
            return { success: false, output: error.message };
        }
    }),

    /**
     * Apply configurations across all clients
     */
    applyConfigurations: adminProcedure.mutation(async () => {
        try {
            const result = await spawnAsync(INFRA_BINARY, ['apply']);
            return { success: result.exitCode === 0, output: result.stdout || result.stderr };
        } catch (error: any) {
            return { success: false, output: error.message };
        }
    }),
});
