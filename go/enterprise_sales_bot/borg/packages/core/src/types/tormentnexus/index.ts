/**
 * @file index.ts
 * @module packages/core/src/types/tormentnexus/index
 *
 * WHAT:
 * Barrel file exporting all TormentNexus Zod types.
 *
 * WHY:
 * Simplifies imports across the application.
 */

export * from "./api-keys.zod.js";
export * from "./config-schemas.zod.js";
export * from "./endpoints.zod.js";
export * from "./logs.zod.js";
export * from "./mcp-servers.zod.js";
export * from "./namespaces.zod.js";
export * from "./oauth.zod.js";
export * from "./policies.zod.js";
export * from "./saved-scripts.zod.js";
export * from "./tool-sets.zod.js";
export * from "./tools.zod.js";
export * from "./server-health.js";

export type ServerParameters = {
	uuid: string;
	name: string;
	description?: string;
	type?: "STDIO" | "SSE" | "STREAMABLE_HTTP";
	command?: string | null;
	args?: string[];
	env?: Record<string, unknown>;
	url?: string | null;
	created_at?: string;
	status?: string;
	error_status?: string;
	stderr?: "overlapped" | "pipe" | "ignore" | "inherit";
	oauth_tokens?: {
		access_token: string;
		token_type: string;
		expires_in?: number;
		scope?: string;
		refresh_token?: string;
	} | null;
	bearerToken?: string | null;
	headers?: Record<string, string>;
	user_id?: string | null;
};

