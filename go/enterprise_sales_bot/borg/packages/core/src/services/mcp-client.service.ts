import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { SSEClientTransport } from "@modelcontextprotocol/sdk/client/sse.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js"; // Assuming this exists in SDK or similar
// StreamableHTTPClientTransport may not be available in basic SDK. TormentNexus uses it.
// If not found, skip HTTP/SSE or use standard HTTP.
// Assuming it prevails based on active imports.

import { Transport } from "@modelcontextprotocol/sdk/shared/transport.js";
import { ServerParameters } from "../types/mcp-admin/index.js";

import {
    ProcessManagedStdioTransport,
    StdioServerParameters,
} from "../transports/process-managed.transport.js";
import { tormentnexusLogStore } from "./log-store.service.js";
import { serverErrorTracker } from "./server-error-tracker.service.js";
import { resolveEnvVariables } from "./utils.service.js";

const sleep = (time: number) =>
    new Promise<void>((resolve) => setTimeout(() => resolve(), time));

const toStringEnv = (
    env: Record<string, unknown> | undefined,
): Record<string, string> | undefined => {
    if (!env) {
        return undefined;
    }

    const mapped = Object.entries(env).reduce<Record<string, string>>(
        (acc, [key, value]) => {
            if (value === undefined || value === null) {
                return acc;
            }

            acc[key] = typeof value === "string" ? value : String(value);
            return acc;
        },
        {},
    );

    return Object.keys(mapped).length > 0 ? mapped : undefined;
};

export interface ConnectedClient {
    client: Client;
    cleanup: () => Promise<void>;
    onProcessCrash?: (exitCode: number | null, signal: string | null) => void;
}

/**
 * Transforms localhost URLs to use host.docker.internal when running inside Docker
 */
export const transformDockerUrl = (url: string): string => {
    if (process.env.TRANSFORM_LOCALHOST_TO_DOCKER_INTERNAL === "true") {
        const transformed = url.replace(
            /localhost|127\.0\.0\.1/g,
            "host.docker.internal",
        );
        return transformed;
    }
    return url;
};

export const createtormentnexusClient = (
    serverParams: ServerParameters,
): { client: Client | undefined; transport: Transport | undefined } => {
    let transport: Transport | undefined;

    // Create the appropriate transport based on server type
    // Default to "STDIO" if type is undefined
    if (!serverParams.type || serverParams.type === "STDIO") {
        // Resolve environment variable placeholders
        const resolvedEnv = serverParams.env
            ? resolveEnvVariables(serverParams.env)
            : undefined;

        const stdioParams: StdioServerParameters = {
            command: serverParams.command || "",
            args: serverParams.args || undefined,
            env: toStringEnv(resolvedEnv),
            stderr: "pipe",
        };
        transport = new ProcessManagedStdioTransport(stdioParams);

        // Handle stderr stream when set to "pipe"
        if ((transport as ProcessManagedStdioTransport).stderr) {
            const stderrStream = (transport as ProcessManagedStdioTransport).stderr;

            stderrStream?.on("data", (chunk: Buffer) => {
                tormentnexusLogStore.addLog(
                    serverParams.name,
                    "error",
                    chunk.toString().trim(),
                );
            });

            stderrStream?.on("error", (error: Error) => {
                tormentnexusLogStore.addLog(
                    serverParams.name,
                    "error",
                    "stderr error",
                    error,
                );
            });
        }

        // Also capture stdout for diagnostics. This includes MCP protocol frames and
        // server-side logs written to stdout by some tools.
        if ((transport as ProcessManagedStdioTransport).stdout) {
            const stdoutStream = (transport as ProcessManagedStdioTransport).stdout;

            stdoutStream?.on("data", (chunk: Buffer) => {
                const message = chunk.toString().trim();
                if (!message) {
                    return;
                }

                tormentnexusLogStore.addLog(
                    serverParams.name,
                    "info",
                    message,
                );
            });

            stdoutStream?.on("error", (error: Error) => {
                tormentnexusLogStore.addLog(
                    serverParams.name,
                    "error",
                    "stdout error",
                    error,
                );
            });
        }
    } else if (serverParams.type === "SSE" && serverParams.url) {
        // Transform the URL if TRANSFORM_LOCALHOST_TO_DOCKER_INTERNAL is set to "true"
        const transformedUrl = transformDockerUrl(serverParams.url);

        // Build headers: start with custom headers, then add auth header
        const headers: Record<string, string> = {
            ...(serverParams.headers || {}),
        };

        // Check for authentication - prioritize OAuth tokens, fallback to bearerToken
        const authToken =
            serverParams.oauth_tokens?.access_token || serverParams.bearerToken;
        if (authToken) {
            headers["Authorization"] = `Bearer ${authToken}`;
        }

        const hasHeaders = Object.keys(headers).length > 0;

        if (!hasHeaders) {
            transport = new SSEClientTransport(new URL(transformedUrl));
        } else {
            transport = new SSEClientTransport(new URL(transformedUrl), {
                eventSourceInit: {
                    fetch: (
                        url: Parameters<typeof fetch>[0],
                        init?: Parameters<typeof fetch>[1],
                    ) => {
                        const mergedHeaders: HeadersInit = {
                            ...(init?.headers
                                ? Object.fromEntries(new Headers(init.headers).entries())
                                : {}),
                            ...headers,
                        };

                        return fetch(url, {
                            ...init,
                            headers: mergedHeaders,
                        });
                    },
                },
                requestInit: {
                    headers
                }
            });
        }
    } else {
        tormentnexusLogStore.addLog(
            serverParams.name,
            "error",
            `Unsupported server type: ${serverParams.type}`,
        );
        return { client: undefined, transport: undefined };
    }

    const client = new Client(
        {
            name: "tormentnexus-client",
            version: "0.99.1",
        },
        {
            capabilities: {
                // Intentionally empty: this client does not require optional MCP client capabilities.
            },
        },
    );
    return { client, transport };
};

export const connecttormentnexusClient = async (
    serverParams: ServerParameters,
    onProcessCrash?: (exitCode: number | null, signal: string | null) => void,
): Promise<ConnectedClient | undefined> => {
    const waitFor = 5000;

    // Get max attempts from server error tracker instead of hardcoding
    const maxAttempts = await serverErrorTracker.getServerMaxAttempts(
        serverParams.uuid,
    );
    let count = 0;
    let retry = true;

    console.log(
        `Connecting to server ${serverParams.name} (${serverParams.uuid}) with max attempts: ${maxAttempts}`,
    );

    while (retry) {
        try {
            // Check if server is already in error state before attempting connection
            const isInErrorState = await serverErrorTracker.isServerInErrorState(
                serverParams.uuid,
            );
            if (isInErrorState) {
                console.warn(
                    `Server ${serverParams.name} (${serverParams.uuid}) is already in ERROR state, skipping connection attempt`,
                );
                return undefined;
            }

            // Create fresh client and transport for each attempt
            const { client, transport } = createtormentnexusClient(serverParams);
            if (!client || !transport) {
                return undefined;
            }

            // Set up process crash detection for STDIO transports BEFORE connecting
            if (transport instanceof ProcessManagedStdioTransport) {
                console.log(
                    `Setting up crash handler for server ${serverParams.name} (${serverParams.uuid})`,
                );
                transport.onprocesscrash = (exitCode, signal) => {
                    console.warn(
                        `Process crashed for server ${serverParams.name} (${serverParams.uuid}): code=${exitCode}, signal=${signal}`,
                    );

                    // Notify the pool about the crash
                    if (onProcessCrash) {
                        console.log(
                            `Calling onProcessCrash callback for server ${serverParams.name} (${serverParams.uuid})`,
                        );
                        onProcessCrash(exitCode, signal);
                    } else {
                        console.warn(
                            `No onProcessCrash callback provided for server ${serverParams.name} (${serverParams.uuid})`,
                        );
                    }
                };
            }

            await client.connect(transport);

            return {
                client,
                cleanup: async () => {
                    await transport.close();
                    await client.close();
                },
                onProcessCrash: (exitCode, signal) => {
                    console.warn(
                        `Process crash detected for server ${serverParams.name} (${serverParams.uuid}): code=${exitCode}, signal=${signal}`,
                    );

                    // Notify the pool about the crash
                    if (onProcessCrash) {
                        onProcessCrash(exitCode, signal);
                    }
                },
            };
        } catch (error) {
            tormentnexusLogStore.addLog(
                "client",
                "error",
                `Error connecting to TormentNexus client (attempt ${count + 1}/${maxAttempts})`,
                error,
            );
            count++;
            retry = count < maxAttempts;
            if (retry) {
                await sleep(waitFor);
            }
        }
    }

    return undefined;
};
