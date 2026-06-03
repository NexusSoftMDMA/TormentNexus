import {
    GetPromptResultSchema,
    ListPromptsResultSchema,
    ListResourcesResultSchema,
    ListResourceTemplatesResultSchema,
    ReadResourceResultSchema,
    type Prompt,
    type Resource,
    type ResourceTemplate,
} from '@modelcontextprotocol/sdk/types.js';

import type { ConnectedClient } from '../services/mcp-client.service.js';
import { getMcpServers } from '../services/mcp-config-discovery.service.js';
import { mcpServerPool } from '../services/mcp-server-pool.service.js';
import { parseToolName } from '../services/tool-name-parser.service.js';
import { sanitizeName } from '../services/common-utils.js';
import type { ServerParameters } from '../types/mcp-admin/index.js';

const DISCOVERY_TIMEOUT_MS = 2_000;

async function withTimeout<T>(promise: Promise<T>, label: string): Promise<T> {
    let timer: NodeJS.Timeout | null = null;

    try {
        return await Promise.race([
            promise,
            new Promise<T>((_, reject) => {
                timer = setTimeout(() => {
                    reject(new Error(`${label} timed out after ${DISCOVERY_TIMEOUT_MS}ms`));
                }, DISCOVERY_TIMEOUT_MS);
            }),
        ]);
    } finally {
        if (timer) {
            clearTimeout(timer);
        }
    }
}

export interface DownstreamDiscoveryContext {
    namespaceUuid: string;
    sessionId: string;
    includeInactiveServers?: boolean;
}

interface DownstreamServerVisit {
    uuid: string;
    params: ServerParameters;
    session: ConnectedClient;
    serverName: string;
}

export function isSameServerInstance(
    params: { name?: string; url?: string | null },
    namespaceUuid: string,
): boolean {
    return params.name === `tormentnexus-unified-${namespaceUuid}`;
}

async function visitEligibleServers(
    context: DownstreamDiscoveryContext,
    logScope: string,
    visit: (server: DownstreamServerVisit) => Promise<void>,
): Promise<void> {
    const serverParams = await getMcpServers(
        context.namespaceUuid,
        context.includeInactiveServers ?? false,
    );
    const visitedServers = new Set<string>();

    const eligibleServers = Object.entries(serverParams).filter(([uuid, params]) => {
        if (visitedServers.has(uuid)) {
            console.log(
                `Skipping already visited server in ${logScope}: ${params.name || uuid}`,
            );
            return false;
        }

        if (isSameServerInstance(params, context.namespaceUuid)) {
            console.log(
                `Skipping self-referencing server in ${logScope}: ${params.name || uuid}`,
            );
            return false;
        }

        visitedServers.add(uuid);
        return true;
    });

    await Promise.allSettled(
        eligibleServers.map(async ([uuid, params]) => {
            const session = await withTimeout(
                mcpServerPool.getSession(
                    context.sessionId,
                    uuid,
                    params,
                    context.namespaceUuid,
                ),
                `Session bootstrap for ${params.name || uuid}`,
            ).catch((error) => {
                console.error(`Error connecting to ${params.name || uuid} during ${logScope}:`, error);
                return null;
            });
            if (!session) {
                return;
            }

            const actualServerName =
                session.client.getServerVersion()?.name || params.name || '';
            const ourServerName = `tormentnexus-unified-${context.namespaceUuid}`;

            if (actualServerName === ourServerName) {
                console.log(
                    `Skipping self-referencing TormentNexus server in ${logScope}: "${actualServerName}"`,
                );
                return;
            }

            const serverName = params.name || actualServerName;
            await visit({
                uuid,
                params,
                session,
                serverName,
            });
        }),
    );
}

export async function listDownstreamPrompts(options: {
    context: DownstreamDiscoveryContext;
    cursor?: string;
    meta?: Record<string, unknown>;
    promptToClient: Record<string, ConnectedClient>;
}): Promise<{ prompts: Prompt[]; nextCursor?: string }> {
    const prompts: Prompt[] = [];

    await visitEligibleServers(options.context, 'prompts', async ({ session, serverName }) => {
        const capabilities = session.client.getServerCapabilities();
        if (!capabilities?.prompts) {
            return;
        }

        try {
            const result = await withTimeout(
                session.client.request(
                    {
                        method: 'prompts/list',
                        params: {
                            cursor: options.cursor,
                            _meta: options.meta,
                        },
                    },
                    ListPromptsResultSchema as any,
                ),
                `Prompt discovery for ${serverName}`,
            );

            if (!result.prompts) {
                return;
            }

            const promptsWithSource = result.prompts.map((prompt: Prompt) => {
                const promptName = `${sanitizeName(serverName)}__${prompt.name}`;
                options.promptToClient[promptName] = session;
                return {
                    ...prompt,
                    name: promptName,
                    description: prompt.description || '',
                };
            });

            prompts.push(...promptsWithSource);
        } catch (error) {
            console.error(`Error fetching prompts from: ${serverName}`, error);
        }
    });

    return {
        prompts,
        nextCursor: options.cursor,
    };
}

export async function getDownstreamPrompt(options: {
    name: string;
    arguments?: Record<string, unknown>;
    meta?: Record<string, unknown>;
    promptToClient: Record<string, ConnectedClient>;
}) {
    const clientForPrompt = options.promptToClient[options.name];
    if (!clientForPrompt) {
        throw new Error(`Unknown prompt: ${options.name}`);
    }

    const parsed = parseToolName(options.name);
    if (!parsed) {
        throw new Error(`Invalid prompt name format: ${options.name}`);
    }

    return await clientForPrompt.client.request(
        {
            method: 'prompts/get',
            params: {
                name: parsed.originalToolName,
                arguments: options.arguments || {},
                _meta: options.meta,
            },
        },
        GetPromptResultSchema,
    );
}

export async function listDownstreamResources(options: {
    context: DownstreamDiscoveryContext;
    cursor?: string;
    meta?: Record<string, unknown>;
    resourceToClient: Record<string, ConnectedClient>;
}): Promise<{ resources: Resource[]; nextCursor?: string }> {
    const resources: Resource[] = [];

    await visitEligibleServers(options.context, 'resources', async ({ session, serverName }) => {
        const capabilities = session.client.getServerCapabilities();
        if (!capabilities?.resources) {
            return;
        }

        try {
            const result = await withTimeout(
                session.client.request(
                    {
                        method: 'resources/list',
                        params: {
                            cursor: options.cursor,
                            _meta: options.meta,
                        },
                    },
                    ListResourcesResultSchema as any,
                ),
                `Resource discovery for ${serverName}`,
            );

            if (!result.resources) {
                return;
            }

            const resourcesWithSource = result.resources.map((resource: Resource) => {
                options.resourceToClient[resource.uri] = session;
                return {
                    ...resource,
                    name: resource.name || '',
                };
            });

            resources.push(...resourcesWithSource);
        } catch (error) {
            console.error(`Error fetching resources from: ${serverName}`, error);
        }
    });

    return {
        resources,
        nextCursor: options.cursor,
    };
}

export async function readDownstreamResource(options: {
    uri: string;
    meta?: Record<string, unknown>;
    resourceToClient: Record<string, ConnectedClient>;
}) {
    const clientForResource = options.resourceToClient[options.uri];
    if (!clientForResource) {
        throw new Error(`Unknown resource: ${options.uri}`);
    }

    return await clientForResource.client.request(
        {
            method: 'resources/read',
            params: {
                uri: options.uri,
                _meta: options.meta,
            },
        },
        ReadResourceResultSchema,
    );
}

export async function listDownstreamResourceTemplates(options: {
    context: DownstreamDiscoveryContext;
    cursor?: string;
    meta?: Record<string, unknown>;
}): Promise<{ resourceTemplates: ResourceTemplate[]; nextCursor?: string }> {
    const resourceTemplates: ResourceTemplate[] = [];

    await visitEligibleServers(
        options.context,
        'resource templates',
        async ({ session, serverName }) => {
            const capabilities = session.client.getServerCapabilities();
            if (!capabilities?.resources) {
                return;
            }

            try {
                const result = await withTimeout(
                    session.client.request(
                        {
                            method: 'resources/templates/list',
                            params: {
                                cursor: options.cursor,
                                _meta: options.meta,
                            },
                        },
                        ListResourceTemplatesResultSchema as any,
                    ),
                    `Resource template discovery for ${serverName}`,
                );

                if (!result.resourceTemplates) {
                    return;
                }

                const templatesWithSource = result.resourceTemplates.map(
                    (template: ResourceTemplate) => ({
                        ...template,
                        name: template.name || '',
                    }),
                );
                resourceTemplates.push(...templatesWithSource);
            } catch (error) {
                console.error(
                    `Error fetching resource templates from: ${serverName}`,
                    error,
                );
            }
        },
    );

    return {
        resourceTemplates,
        nextCursor: options.cursor,
    };
}
