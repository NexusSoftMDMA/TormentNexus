import { ConfigKey, ConfigKeyEnum } from "../types/mcp-admin/index.js";
import { configRepo } from "../db/repositories/index.js";

export const configService = {
    async isSignupDisabled(): Promise<boolean> {
        const config = await configRepo.get(ConfigKeyEnum.Enum.DISABLE_SIGNUP);
        return config === "true";
    },

    async setSignupDisabled(disabled: boolean): Promise<void> {
        await configRepo.set(
            ConfigKeyEnum.Enum.DISABLE_SIGNUP,
            disabled.toString()
        );
    },

    async isSsoSignupDisabled(): Promise<boolean> {
        const config = await configRepo.get(
            ConfigKeyEnum.Enum.DISABLE_SSO_SIGNUP,
        );
        return config === "true";
    },

    async setSsoSignupDisabled(disabled: boolean): Promise<void> {
        await configRepo.set(
            ConfigKeyEnum.Enum.DISABLE_SSO_SIGNUP,
            disabled.toString()
        );
    },

    async isBasicAuthDisabled(): Promise<boolean> {
        const config = await configRepo.get(
            ConfigKeyEnum.Enum.DISABLE_BASIC_AUTH,
        );
        return config === "true";
    },

    async setBasicAuthDisabled(disabled: boolean): Promise<void> {
        await configRepo.set(
            ConfigKeyEnum.Enum.DISABLE_BASIC_AUTH,
            disabled.toString()
        );
    },

    async getMcpResetTimeoutOnProgress(): Promise<boolean> {
        const config = await configRepo.get(
            ConfigKeyEnum.Enum.MCP_RESET_TIMEOUT_ON_PROGRESS,
        );
        return config === "true" || true;
    },

    async setMcpResetTimeoutOnProgress(enabled: boolean): Promise<void> {
        await configRepo.set(
            ConfigKeyEnum.Enum.MCP_RESET_TIMEOUT_ON_PROGRESS,
            enabled.toString()
        );
    },

    async getMcpTimeout(): Promise<number> {
        const config = await configRepo.get(ConfigKeyEnum.Enum.MCP_TIMEOUT);
        return config ? parseInt(config, 10) : 60000;
    },

    async setMcpTimeout(timeout: number): Promise<void> {
        await configRepo.set(
            ConfigKeyEnum.Enum.MCP_TIMEOUT,
            timeout.toString()
        );
    },

    async getMcpMaxTotalTimeout(): Promise<number> {
        const config = await configRepo.get(
            ConfigKeyEnum.Enum.MCP_MAX_TOTAL_TIMEOUT,
        );
        return config ? parseInt(config, 10) : 60000;
    },

    async setMcpMaxTotalTimeout(timeout: number): Promise<void> {
        await configRepo.set(
            ConfigKeyEnum.Enum.MCP_MAX_TOTAL_TIMEOUT,
            timeout.toString()
        );
    },

    async getMcpMaxAttempts(): Promise<number> {
        const config = await configRepo.get(
            ConfigKeyEnum.Enum.MCP_MAX_ATTEMPTS,
        );
        return config ? parseInt(config, 10) : 1;
    },

    async setMcpMaxAttempts(maxAttempts: number): Promise<void> {
        await configRepo.set(
            ConfigKeyEnum.Enum.MCP_MAX_ATTEMPTS,
            maxAttempts.toString()
        );
    },

    async getSessionLifetime(): Promise<number | null> {
        const config = await configRepo.get(
            ConfigKeyEnum.Enum.SESSION_LIFETIME,
        );
        if (!config) {
            return null; // No session lifetime set - infinite sessions
        }
        const lifetime = parseInt(config, 10);
        return isNaN(lifetime) ? null : lifetime;
    },

    async setSessionLifetime(lifetime?: number | null): Promise<void> {
        if (lifetime === null || lifetime === undefined) {
            // Remove the config to indicate infinite session lifetime
            await configRepo.delete(ConfigKeyEnum.Enum.SESSION_LIFETIME);
        } else {
            await configRepo.set(
                ConfigKeyEnum.Enum.SESSION_LIFETIME,
                lifetime.toString()
            );
        }
    },

    async getConfig(key: ConfigKey): Promise<string | null> {
        return await configRepo.get(key);
    },

    async setConfig(
        key: ConfigKey,
        value: string
    ): Promise<void> {
        await configRepo.set(key, value);
    },

    // async getAllConfigs(): Promise<
    //   Array<{ id: string; value: string; description?: string | null }>
    // > {
    //   return await configRepo.getAllConfigs(); // Need to implement getAll in repo or allow it. Repo didn't have getAll in my previous view.
    // },

    async getAuthProviders(): Promise<
        Array<{ id: string; name: string; enabled: boolean }>
    > {
        const providers = [];

        // Check if OIDC is configured
        const isOidcEnabled = !!(
            process.env.OIDC_CLIENT_ID &&
            process.env.OIDC_CLIENT_SECRET &&
            process.env.OIDC_DISCOVERY_URL
        );

        if (isOidcEnabled) {
            providers.push({
                id: "oidc",
                name: "OIDC",
                enabled: true,
            });
        }

        return providers;
    },

    /**
     * Get the memory limit for Code Mode execution in MB.
     * Defaults to 128MB.
     */
    getCodeExecutionMemoryLimit(): number {
        const envVal = process.env.CODE_EXECUTION_MEMORY_LIMIT;
        if (envVal) {
            const parsed = parseInt(envVal, 10);
            if (!isNaN(parsed) && parsed > 0) {
                return parsed;
            }
        }
        return 128;
    },

    /**
     * Validate that OPENAI_API_KEY is present if required features are enabled.
     * Logs a warning if not present.
     */
    validateOpenAiKey(): void {
        if (!process.env.OPENAI_API_KEY) {
            console.warn(
                "WARN: OPENAI_API_KEY is not set. Semantic Search and Autonomous Agent features will not work."
            );
        }
    }
};
