import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { mcpServersRepository, oauthRepository, oauthSessionsRepository } from '../db/repositories/index.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';
import {
    OAuthClientCreateInputSchema,
    OAuthSessionCreateInputSchema,
    OAuthTokensSchema,
} from '../types/tormentnexus/oauth.zod.js';

const ParsedOAuthStateSchema = z.object({
    mcpServerUuid: z.string().min(1),
    tokenEndpoint: z.string().url().optional(),
    redirectUri: z.string().url().optional(),
    clientId: z.string().optional(),
});

function parseOAuthState(rawState: string): z.infer<typeof ParsedOAuthStateSchema> {
    const trimmed = rawState.trim();

    if (/^[0-9a-fA-F-]{36}$/.test(trimmed)) {
        return { mcpServerUuid: trimmed };
    }

    try {
        const parsed = JSON.parse(trimmed);
        return ParsedOAuthStateSchema.parse(parsed);
    } catch {
        throw new Error('Invalid OAuth state payload');
    }
}

function resolveTokenEndpoint(serverUrl?: string | null): string {
    if (!serverUrl) {
        throw new Error('Token endpoint not provided and MCP server URL is unavailable');
    }

    const normalized = serverUrl.endsWith('/') ? serverUrl.slice(0, -1) : serverUrl;
    return `${normalized}/oauth/token`;
}

const clientsRouter = t.router({
    create: adminProcedure
        .input(OAuthClientCreateInputSchema)
        .mutation(async ({ input }) => {
            try {
                return await oauthRepository.createClient(input);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('OAuth store is unavailable', error);
            }
        }),

    get: adminProcedure
        .input(z.object({ clientId: z.string() }))
        .query(async ({ input }) => {
            try {
                return await oauthRepository.findClientById(input.clientId);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('OAuth store is unavailable', error);
            }
        }),
});

const sessionsRouter = t.router({
    upsert: adminProcedure
        .input(OAuthSessionCreateInputSchema)
        .mutation(async ({ input }) => {
            try {
                return await oauthSessionsRepository.upsert(input);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('OAuth store is unavailable', error);
            }
        }),

    getByServer: adminProcedure
        .input(z.object({ mcpServerUuid: z.string() }))
        .query(async ({ input }) => {
            try {
                return await oauthSessionsRepository.findByMcpServerUuid(input.mcpServerUuid);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('OAuth store is unavailable', error);
            }
        }),
});

export const oauthRouter = t.router({
    clients: clientsRouter,
    sessions: sessionsRouter,
    exchange: publicProcedure
        .input(z.object({ code: z.string(), state: z.string() }))
        .mutation(async ({ input }) => {
            try {
                const parsedState = parseOAuthState(input.state);

                const existingSession = await oauthSessionsRepository.findByMcpServerUuid(parsedState.mcpServerUuid);
                if (!existingSession) {
                    throw new Error('No pending OAuth session found for this state');
                }

                const clientInfo = existingSession.client_information;
                const clientId = parsedState.clientId ?? clientInfo?.client_id;
                if (!clientId) {
                    throw new Error('OAuth client_id missing from session/state');
                }

                const client = await oauthRepository.findClientById(clientId);
                if (!client) {
                    throw new Error(`OAuth client not found: ${clientId}`);
                }

                const mcpServer = await mcpServersRepository.findByUuid(parsedState.mcpServerUuid);
                const tokenEndpoint = parsedState.tokenEndpoint ?? resolveTokenEndpoint(mcpServer?.url ?? null);
                const redirectUri = parsedState.redirectUri ?? client.redirect_uris?.[0];

                if (!redirectUri) {
                    throw new Error('OAuth redirect_uri unavailable for token exchange');
                }

                const body = new URLSearchParams({
                    grant_type: 'authorization_code',
                    code: input.code,
                    redirect_uri: redirectUri,
                    client_id: clientId,
                });

                if (client.client_secret) {
                    body.set('client_secret', client.client_secret);
                }

                if (existingSession.code_verifier) {
                    body.set('code_verifier', existingSession.code_verifier);
                }

                const response = await fetch(tokenEndpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                        Accept: 'application/json',
                    },
                    body: body.toString(),
                });

                const raw = await response.text();
                let parsed: unknown;

                try {
                    parsed = JSON.parse(raw);
                } catch {
                    parsed = { error: 'invalid_json', raw };
                }

                if (!response.ok) {
                    const providerMessage = typeof parsed === 'object' && parsed !== null && 'error_description' in parsed
                        ? String((parsed as Record<string, unknown>).error_description)
                        : typeof parsed === 'object' && parsed !== null && 'error' in parsed
                            ? String((parsed as Record<string, unknown>).error)
                            : `HTTP ${response.status}`;

                    throw new Error(`OAuth token exchange failed: ${providerMessage}`);
                }

                const tokens = OAuthTokensSchema.parse(parsed);

                await oauthSessionsRepository.upsert({
                    mcp_server_uuid: parsedState.mcpServerUuid,
                    client_information: existingSession.client_information,
                    code_verifier: existingSession.code_verifier,
                    tokens,
                });

                return { success: true, tokens };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('OAuth store is unavailable', error);
            }
        }),
});
