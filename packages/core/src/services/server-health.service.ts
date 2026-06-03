import {
    ConfigKey,
    HealthCheckResult,
    ServerHealthInfo,
    ServerHealthStatus,
} from "../types/mcp-admin/index.js";

import { mcpServersRepository } from "../db/repositories/index.js";
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from "../db/sqliteAvailability.js";
import { configService } from "./config.service.js";
import { autoReconnectService } from "./auto-reconnect.service.js";
import { connecttormentnexusClient } from "./mcp-client.service.js";
import { serverErrorTracker } from "./server-error-tracker.service.js";
import { convertDbServerToParams } from "./utils.service.js";

interface HealthState {
    serverUuid: string;
    serverName: string;
    status: ServerHealthStatus;
    lastChecked: Date | null;
    lastHealthy: Date | null;
    responseTimeMs: number | null;
    errorMessage: string | null;
    consecutiveFailures: number;
    toolCount: number | null;
}

export class ServerHealthService {
    private static instance: ServerHealthService | null = null;

    private healthStates: Map<string, HealthState> = new Map();
    private periodicCheckInterval: NodeJS.Timeout | null = null;
    private lastFullCheckAt: Date | null = null;
    private isRunning = false;

    private readonly DEFAULT_CHECK_INTERVAL_MS = 60000;
    private readonly DEFAULT_TIMEOUT_MS = 10000;
    private readonly DEFAULT_UNHEALTHY_THRESHOLD = 3;

    private constructor() { }

    static getInstance(): ServerHealthService {
        if (!ServerHealthService.instance) {
            ServerHealthService.instance = new ServerHealthService();
        }
        return ServerHealthService.instance;
    }

    private buildUnavailableResult(serverUuid: string, startTime: number, checkedAt: string, error: unknown): HealthCheckResult {
        return {
            serverUuid,
            success: false,
            responseTimeMs: Date.now() - startTime,
            errorMessage: formatOptionalSqliteFailure(
                "Server health persistence is unavailable",
                error,
            ),
            checkedAt,
        };
    }

    async checkServerHealth(serverUuid: string): Promise<HealthCheckResult> {
        const startTime = Date.now();
        const checkedAt = new Date().toISOString();

        let server;
        try {
            server = await mcpServersRepository.findByUuid(serverUuid);
        } catch (error) {
            if (isSqliteUnavailableError(error)) {
                const result = this.buildUnavailableResult(serverUuid, startTime, checkedAt, error);
                this.setServerStatus(serverUuid, serverUuid, "UNKNOWN");
                return result;
            }

            throw error;
        }
        if (!server) {
            return {
                serverUuid,
                success: false,
                responseTimeMs: Date.now() - startTime,
                errorMessage: "Server not found",
                checkedAt,
            };
        }

        const serverParams = await convertDbServerToParams(server);
        if (!serverParams) {
            return {
                serverUuid,
                success: false,
                responseTimeMs: Date.now() - startTime,
                errorMessage: "Failed to convert server parameters",
                checkedAt,
            };
        }

        const isInErrorState =
            await serverErrorTracker.isServerInErrorState(serverUuid);
        if (isInErrorState) {
            this.updateHealthState(serverUuid, server.name, {
                serverUuid,
                success: false,
                responseTimeMs: Date.now() - startTime,
                errorMessage: "Server is in ERROR state due to repeated crashes",
                checkedAt,
            });

            return {
                serverUuid,
                success: false,
                responseTimeMs: Date.now() - startTime,
                errorMessage: "Server is in ERROR state due to repeated crashes",
                checkedAt,
            };
        }

        this.setServerStatus(serverUuid, server.name, "CHECKING");

        try {
            const timeoutMs = await this.getTimeoutMs();

            const connectionResult = await Promise.race([
                this.performHealthCheck(serverParams),
                new Promise<{ success: false; errorMessage: string }>((resolve) =>
                    setTimeout(
                        () =>
                            resolve({
                                success: false,
                                errorMessage: "Health check timed out",
                            }),
                        timeoutMs,
                    ),
                ),
            ]);

            const responseTimeMs = Date.now() - startTime;

            const result: HealthCheckResult = {
                serverUuid,
                success: connectionResult.success,
                responseTimeMs,
                toolCount: connectionResult.success
                    ? connectionResult.toolCount
                    : undefined,
                errorMessage: connectionResult.success
                    ? undefined
                    // @ts-expect-error -- type mismatch from inferred schema types
                    : connectionResult.errorMessage,
                checkedAt,
            };

            this.updateHealthState(serverUuid, server.name, result);

            return result;
        } catch (error) {
            const responseTimeMs = Date.now() - startTime;
            const errorMessage =
                error instanceof Error ? error.message : "Unknown error";

            const result: HealthCheckResult = {
                serverUuid,
                success: false,
                responseTimeMs,
                errorMessage,
                checkedAt,
            };

            this.updateHealthState(serverUuid, server.name, result);

            return result;
        }
    }

    private async performHealthCheck(
        serverParams: Parameters<typeof connecttormentnexusClient>[0],
    ): Promise<
        | { success: true; toolCount: number }
        | { success: false; errorMessage: string }
    > {
        try {
            const connectedClient = await connecttormentnexusClient(serverParams);

            if (!connectedClient) {
                return { success: false, errorMessage: "Failed to connect to server" };
            }

            try {
                const toolsResult = await connectedClient.client.listTools();
                const toolCount = toolsResult.tools?.length ?? 0;

                await connectedClient.cleanup();

                return { success: true, toolCount };
            } catch (error) {
                await connectedClient.cleanup().catch(() => { });
                return {
                    success: false,
                    errorMessage:
                        error instanceof Error ? error.message : "Failed to list tools",
                };
            }
        } catch (error) {
            return {
                success: false,
                errorMessage:
                    error instanceof Error ? error.message : "Connection failed",
            };
        }
    }

    async checkMultipleServers(
        serverUuids?: string[],
    ): Promise<HealthCheckResult[]> {
        let uuidsToCheck = serverUuids;

        if (!uuidsToCheck || uuidsToCheck.length === 0) {
            try {
                const allServers = await mcpServersRepository.findAll();
                uuidsToCheck = allServers.map((s) => s.uuid);
            } catch (error) {
                if (isSqliteUnavailableError(error)) {
                    this.lastFullCheckAt = new Date();
                    return [{
                        serverUuid: "catalog",
                        success: false,
                        responseTimeMs: 0,
                        errorMessage: formatOptionalSqliteFailure(
                            "Server health persistence is unavailable",
                            error,
                        ),
                        checkedAt: new Date().toISOString(),
                    }];
                }

                throw error;
            }
        }

        const results = await Promise.allSettled(
            (uuidsToCheck || []).map((uuid) => this.checkServerHealth(uuid)),
        );

        this.lastFullCheckAt = new Date();

        return results.map((result, index) => {
            if (result.status === "fulfilled") {
                return result.value;
            }
            return {
                serverUuid: uuidsToCheck![index],
                success: false,
                responseTimeMs: 0,
                errorMessage: result.reason?.message ?? "Unknown error",
                checkedAt: new Date().toISOString(),
            };
        });
    }

    private setServerStatus(
        serverUuid: string,
        serverName: string,
        status: ServerHealthStatus,
    ): void {
        const existing = this.healthStates.get(serverUuid);
        if (existing) {
            existing.status = status;
        } else {
            this.healthStates.set(serverUuid, {
                serverUuid,
                serverName,
                status,
                lastChecked: null,
                lastHealthy: null,
                responseTimeMs: null,
                errorMessage: null,
                consecutiveFailures: 0,
                toolCount: null,
            });
        }
    }

    private updateHealthState(
        serverUuid: string,
        serverName: string,
        result: HealthCheckResult,
    ): void {
        const existing = this.healthStates.get(serverUuid);
        const now = new Date();

        if (result.success) {
            this.healthStates.set(serverUuid, {
                serverUuid,
                serverName,
                status: "HEALTHY",
                lastChecked: now,
                lastHealthy: now,
                responseTimeMs: result.responseTimeMs,
                errorMessage: null,
                consecutiveFailures: 0,
                toolCount: result.toolCount ?? null,
            });
        } else {
            const consecutiveFailures = (existing?.consecutiveFailures ?? 0) + 1;
            const unhealthyThreshold = this.DEFAULT_UNHEALTHY_THRESHOLD;
            const isNowUnhealthy = consecutiveFailures >= unhealthyThreshold;

            this.healthStates.set(serverUuid, {
                serverUuid,
                serverName,
                status: isNowUnhealthy ? "UNHEALTHY" : "UNKNOWN",
                lastChecked: now,
                lastHealthy: existing?.lastHealthy ?? null,
                responseTimeMs: result.responseTimeMs,
                errorMessage: result.errorMessage ?? null,
                consecutiveFailures,
                toolCount: existing?.toolCount ?? null,
            });

            if (isNowUnhealthy) {
                autoReconnectService.scheduleReconnection(
                    serverUuid,
                    serverName,
                    "health_failure",
                );
            }
        }
    }

    getServerHealth(serverUuid: string): ServerHealthInfo | undefined {
        const state = this.healthStates.get(serverUuid);
        if (!state) return undefined;

        return this.stateToHealthInfo(state);
    }

    getAllServerHealth(): ServerHealthInfo[] {
        return Array.from(this.healthStates.values()).map((state) =>
            this.stateToHealthInfo(state),
        );
    }

    private stateToHealthInfo(state: HealthState): ServerHealthInfo {
        return {
            serverUuid: state.serverUuid,
            serverName: state.serverName,
            status: state.status,
            lastChecked: state.lastChecked?.toISOString() ?? null,
            lastHealthy: state.lastHealthy?.toISOString() ?? null,
            responseTimeMs: state.responseTimeMs,
            errorMessage: state.errorMessage,
            consecutiveFailures: state.consecutiveFailures,
            toolCount: state.toolCount,
        };
    }

    getHealthSummary(): {
        totalServers: number;
        healthyCount: number;
        unhealthyCount: number;
        unknownCount: number;
        checkingCount: number;
        lastFullCheckAt: string | null;
    } {
        const states = Array.from(this.healthStates.values());

        return {
            totalServers: states.length,
            healthyCount: states.filter((s) => s.status === "HEALTHY").length,
            unhealthyCount: states.filter((s) => s.status === "UNHEALTHY").length,
            unknownCount: states.filter((s) => s.status === "UNKNOWN").length,
            checkingCount: states.filter((s) => s.status === "CHECKING").length,
            lastFullCheckAt: this.lastFullCheckAt?.toISOString() ?? null,
        };
    }

    async startPeriodicChecks(): Promise<void> {
        if (this.isRunning) {
            console.log("Periodic health checks already running");
            return;
        }

        this.isRunning = true;
        const intervalMs = await this.getCheckIntervalMs();

        console.log(
            `Starting periodic health checks with interval: ${intervalMs}ms`,
        );

        await this.checkMultipleServers();

        this.periodicCheckInterval = setInterval(async () => {
            try {
                await this.checkMultipleServers();
            } catch (error) {
                console.error("Error during periodic health check:", error);
            }
        }, intervalMs);
    }

    stopPeriodicChecks(): void {
        if (this.periodicCheckInterval) {
            clearInterval(this.periodicCheckInterval);
            this.periodicCheckInterval = null;
        }
        this.isRunning = false;
        console.log("Stopped periodic health checks");
    }

    private async getCheckIntervalMs(): Promise<number> {
        try {
            const config = await configService.getConfig(
                "HEALTH_CHECK_INTERVAL" as ConfigKey,
            );
            if (config) {
                const parsed = parseInt(config, 10);
                if (!isNaN(parsed) && parsed >= 5000) {
                    return parsed;
                }
            }
        } catch { }
        return this.DEFAULT_CHECK_INTERVAL_MS;
    }

    private async getTimeoutMs(): Promise<number> {
        try {
            const timeout = await configService.getMcpTimeout();
            return Math.min(timeout, this.DEFAULT_TIMEOUT_MS);
        } catch {
            return this.DEFAULT_TIMEOUT_MS;
        }
    }

    clearHealthState(serverUuid: string): void {
        this.healthStates.delete(serverUuid);
    }

    clearAllHealthStates(): void {
        this.healthStates.clear();
        this.lastFullCheckAt = null;
    }
}

export const serverHealthService = ServerHealthService.getInstance();
