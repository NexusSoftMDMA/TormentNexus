import fs from 'fs/promises';
import path from 'path';
import { existsSync } from 'fs';
import { z } from 'zod';

/**
 * AutoConfig (Universal Client)
 * Automatically detects environment (Kubernetes, Docker, Local)
 * and configures the MCP client accordingly.
 */
const RuntimeEnvironmentSchema = z.enum(['k8s', 'docker', 'local']);
type RuntimeEnvironment = z.infer<typeof RuntimeEnvironmentSchema>;

const CommandServerSchema = z.object({
    command: z.string().min(1),
    args: z.array(z.string()),
    env: z.record(z.string()).optional(),
});

const SseServerSchema = z.object({
    url: z.string().url(),
    transport: z.literal('sse'),
});

const AutoConfigSchema = z.object({
    mcpServers: z.object({
        tormentnexus: z.union([CommandServerSchema, SseServerSchema]),
    }),
    ui: z.object({
        theme: z.string(),
        logs: z.enum(['verbose', 'json']),
    }),
    meta: z.object({
        environment: RuntimeEnvironmentSchema,
        generatedAt: z.string(),
    }),
});

export type GeneratedAutoConfig = z.infer<typeof AutoConfigSchema>;

export interface AutoConfigOptions {
    workspaceRoot?: string;
    sseUrl?: string;
    envAllowList?: string[];
}

export class AutoConfig {
    private static readonly DEFAULT_ALLOWLIST = ['NODE_ENV'];

    static async detectEnvironment(): Promise<RuntimeEnvironment> {
        if (process.env.KUBERNETES_SERVICE_HOST) {
            return 'k8s';
        }

        try {
            await fs.access('/.dockerenv');
            return 'docker';
        } catch {
            // continue
        }

        try {
            const cgroup = await fs.readFile('/proc/1/cgroup', 'utf-8');
            if (/docker|containerd|kubepods/i.test(cgroup)) {
                return 'docker';
            }
        } catch {
            // non-linux or unavailable cgroup path
        }

        return 'local';
    }

    private static filterEnvironment(allowList: string[] = AutoConfig.DEFAULT_ALLOWLIST): Record<string, string> {
        const allowSet = new Set(allowList.map((k) => k.trim()).filter(Boolean));
        const out: Record<string, string> = {};

        for (const key of allowSet) {
            const value = process.env[key];
            if (typeof value === 'string') {
                out[key] = value;
            }
        }

        return out;
    }

    private static resolveCoreEntry(workspaceRoot: string): { command: string; args: string[] } {
        const distEntry = path.join(workspaceRoot, 'packages', 'core', 'dist', 'index.js');
        const srcEntry = path.join(workspaceRoot, 'packages', 'core', 'src', 'index.ts');

        if (existsSync(distEntry)) {
            return { command: process.execPath, args: [distEntry] };
        }

        return { command: process.execPath, args: [srcEntry] };
    }

    static async generateConfig(options: AutoConfigOptions = {}): Promise<GeneratedAutoConfig> {
        const env = await this.detectEnvironment();
        console.log(`[AutoConfig] Detected Environment: ${env}`);

        const workspaceRoot = options.workspaceRoot ?? process.cwd();
        const envVars = this.filterEnvironment(options.envAllowList);

        const localServer = this.resolveCoreEntry(workspaceRoot);
        const localConfig = {
            command: localServer.command,
            args: localServer.args,
            ...(Object.keys(envVars).length > 0 ? { env: envVars } : {}),
        };

        const sseUrl = options.sseUrl ?? process.env.TORMENTNEXUS_MCP_SSE_URL ?? 'http://tormentnexus.default.svc.cluster.local:3000/sse';
        const serverConfig = env === 'k8s'
            ? { url: sseUrl, transport: 'sse' as const }
            : localConfig;

        const config: GeneratedAutoConfig = {
            mcpServers: {
                tormentnexus: serverConfig,
            },
            ui: {
                theme: env === 'k8s' ? 'dark-enterprise' : 'dark-modern',
                logs: env === 'local' ? 'verbose' : 'json'
            },
            meta: {
                environment: env,
                generatedAt: new Date().toISOString(),
            }
        };

        return AutoConfigSchema.parse(config);
    }

    static async writeConfig(config: GeneratedAutoConfig, outputPath: string): Promise<string> {
        const dir = path.dirname(outputPath);
        await fs.mkdir(dir, { recursive: true });
        await fs.writeFile(outputPath, JSON.stringify(config, null, 2), 'utf-8');
        return outputPath;
    }

    static async generateAndWrite(options: AutoConfigOptions & { outputPath: string }): Promise<string> {
        const config = await this.generateConfig(options);
        return this.writeConfig(config, options.outputPath);
    }
}
