import {
    McpServerErrorStatusEnum,
    McpServerStatusEnum,
    ServerParameters,
} from "../types/mcp-admin/index.js";

import { mcpServersRepository } from "../db/repositories/index.js";
import { configService } from "./config.service.js";
import { mcpServerPool } from "./mcp-server-pool.service.js";
import { serverErrorTracker } from "./server-error-tracker.service.js";
import { serverHealthService } from "./server-health.service.js";
import { convertDbServerToParams } from "./utils.service.js";

export type ReconnectionStatus =
    | "IDLE"
    | "PENDING"
    | "RECONNECTING"
    | "SUCCEEDED"
    | "FAILED"
    | "CANCELLED"
    | "MAX_RETRIES_EXCEEDED";

export interface ReconnectionState {
    serverUuid: string;
    serverName: string;
    status: ReconnectionStatus;
    currentAttempt: number;
    maxAttempts: number;
    nextRetryAt: Date | null;
    lastAttemptAt: Date | null;
    lastError: string | null;
    totalAttempts: number;
    successfulReconnections: number;
}

export interface ReconnectionResult {
    serverUuid: string;
    success: boolean;
    attempt: number;
    error?: string;
    reconnectedAt?: string;
}

export interface ReconnectionConfig {
    maxAttempts: number;
    baseDelayMs: number;
    maxDelayMs: number;
    jitterFactor: number;
    autoReconnectOnCrash: boolean;
    autoReconnectOnHealthFailure: boolean;
}

const DEFAULT_CONFIG: ReconnectionConfig = {
    maxAttempts: 5,
    baseDelayMs: 1000,
    maxDelayMs: 60000,
    jitterFactor: 0.2,
    autoReconnectOnCrash: true,
    autoReconnectOnHealthFailure: true,
};

export class AutoReconnectService {
    private static instance: AutoReconnectService | null = null;

    private reconnectionStates: Map<string, ReconnectionState> = new Map();
    private reconnectionTimers: Map<string, NodeJS.Timeout> = new Map();
    private config: ReconnectionConfig = { ...DEFAULT_CONFIG };
    private isEnabled = true;

    private onReconnectSuccessListeners: Array<(serverUuid: string) => void> = [];
    private onReconnectFailureListeners: Array<
        (serverUuid: string, error: string) => void
    > = [];
    private onMaxRetriesExceededListeners: Array<(serverUuid: string) => void> =
        [];

    private constructor() { }

    static getInstance(): AutoReconnectService {
        if (!AutoReconnectService.instance) {
            AutoReconnectService.instance = new AutoReconnectService();
        }
        return AutoReconnectService.instance;
    }

    configure(config: Partial<ReconnectionConfig>): void {
        this.config = { ...this.config, ...config };
        console.log("AutoReconnectService configured:", this.config);
    }

    async loadConfigFromDb(): Promise<void> {
        try {
            const maxAttempts = await configService.getMcpMaxAttempts();
            const timeout = await configService.getMcpTimeout();

            this.config.maxAttempts = Math.max(maxAttempts, 1);
            this.config.maxDelayMs = Math.min(timeout * 2, 120000);

            console.log("AutoReconnectService loaded config from DB:", this.config);
        } catch (error) {
            console.warn(
                "Failed to load auto-reconnect config from DB, using defaults:",
                error,
            );
        }
    }

    setEnabled(enabled: boolean): void {
        this.isEnabled = enabled;
        if (!enabled) {
            this.cancelAllReconnections();
        }
        console.log(`AutoReconnectService ${enabled ? "enabled" : "disabled"}`);
    }

    isAutoReconnectEnabled(): boolean {
        return this.isEnabled;
    }

    scheduleReconnection(
        serverUuid: string,
        serverName: string,
        reason: "crash" | "health_failure" | "manual",
    ): void {
        if (!this.isEnabled) {
            console.log(
                `Auto-reconnect disabled, skipping reconnection for ${serverName}`,
            );
            return;
        }

        if (reason === "crash" && !this.config.autoReconnectOnCrash) {
            console.log(
                `Auto-reconnect on crash disabled, skipping for ${serverName}`,
            );
            return;
        }

        if (
            reason === "health_failure" &&
            !this.config.autoReconnectOnHealthFailure
        ) {
            console.log(
                `Auto-reconnect on health failure disabled, skipping for ${serverName}`,
            );
            return;
        }

        const existingState = this.reconnectionStates.get(serverUuid);
        if (
            existingState &&
            (existingState.status === "RECONNECTING" ||
                existingState.status === "PENDING")
        ) {
            console.log(
                `Reconnection already in progress for ${serverName}, skipping`,
            );
            return;
        }

        const state: ReconnectionState = {
            serverUuid,
            serverName,
            status: "PENDING",
            currentAttempt: 0,
            maxAttempts: this.config.maxAttempts,
            nextRetryAt: null,
            lastAttemptAt: null,
            lastError: null,
            totalAttempts: existingState?.totalAttempts ?? 0,
            successfulReconnections: existingState?.successfulReconnections ?? 0,
        };

        this.reconnectionStates.set(serverUuid, state);

        console.log(
            `Scheduled reconnection for ${serverName} (${serverUuid}) due to ${reason}`,
        );

        this.attemptReconnection(serverUuid);
    }

    private async attemptReconnection(serverUuid: string): Promise<void> {
        const state = this.reconnectionStates.get(serverUuid);
        if (!state) {
            console.error(`No reconnection state found for ${serverUuid}`);
            return;
        }

        if (state.currentAttempt >= state.maxAttempts) {
            state.status = "MAX_RETRIES_EXCEEDED";
            console.error(
                `Max reconnection attempts (${state.maxAttempts}) exceeded for ${state.serverName}`,
            );
            this.notifyMaxRetriesExceeded(serverUuid);
            return;
        }

        state.currentAttempt++;
        state.totalAttempts++;
        state.status = "RECONNECTING";
        state.lastAttemptAt = new Date();

        console.log(
            `Attempting reconnection for ${state.serverName} (attempt ${state.currentAttempt}/${state.maxAttempts})`,
        );

        try {
            const server = await mcpServersRepository.findByUuid(serverUuid);
            if (!server) {
                throw new Error("Server not found in database");
            }

            const serverParams = await convertDbServerToParams(server);
            if (!serverParams) {
                throw new Error("Failed to convert server parameters");
            }

            await serverErrorTracker.resetServerErrorState(serverUuid);
            await mcpServerPool.invalidateIdleSession(serverUuid, serverParams);

            const healthResult =
                await serverHealthService.checkServerHealth(serverUuid);

            if (healthResult.success) {
                state.status = "SUCCEEDED";
                state.lastError = null;
                state.currentAttempt = 0;
                state.nextRetryAt = null;
                state.successfulReconnections++;

                console.log(
                    `Successfully reconnected to ${state.serverName} after ${state.totalAttempts} attempts`,
                );
                this.notifyReconnectSuccess(serverUuid);
            } else {
                throw new Error(healthResult.errorMessage ?? "Health check failed");
            }
        } catch (error) {
            const errorMessage =
                error instanceof Error ? error.message : "Unknown error";
            state.lastError = errorMessage;

            console.error(
                `Reconnection attempt ${state.currentAttempt} failed for ${state.serverName}: ${errorMessage}`,
            );

            if (state.currentAttempt < state.maxAttempts) {
                const delayMs = this.calculateBackoffDelay(state.currentAttempt);
                state.status = "PENDING";
                state.nextRetryAt = new Date(Date.now() + delayMs);

                console.log(`Scheduling retry for ${state.serverName} in ${delayMs}ms`);

                const existingTimer = this.reconnectionTimers.get(serverUuid);
                if (existingTimer) {
                    clearTimeout(existingTimer);
                }

                const timer = setTimeout(() => {
                    this.attemptReconnection(serverUuid);
                }, delayMs);

                this.reconnectionTimers.set(serverUuid, timer);
            } else {
                state.status = "MAX_RETRIES_EXCEEDED";
                console.error(
                    `All reconnection attempts exhausted for ${state.serverName}`,
                );
                this.notifyMaxRetriesExceeded(serverUuid);
            }

            this.notifyReconnectFailure(serverUuid, errorMessage);
        }
    }

    // Exponential backoff: baseDelay * 2^(attempt-1) with jitter to prevent thundering herd
    private calculateBackoffDelay(attempt: number): number {
        const exponentialDelay = this.config.baseDelayMs * Math.pow(2, attempt - 1);
        const cappedDelay = Math.min(exponentialDelay, this.config.maxDelayMs);
        const jitter = cappedDelay * this.config.jitterFactor * Math.random();
        return Math.round(cappedDelay + jitter);
    }

    async triggerReconnection(serverUuid: string): Promise<ReconnectionResult> {
        const server = await mcpServersRepository.findByUuid(serverUuid);
        if (!server) {
            return {
                serverUuid,
                success: false,
                attempt: 0,
                error: "Server not found",
            };
        }

        this.cancelReconnection(serverUuid);
        this.scheduleReconnection(serverUuid, server.name, "manual");

        await new Promise((resolve) => setTimeout(resolve, 100));

        const state = this.reconnectionStates.get(serverUuid);
        if (!state) {
            return {
                serverUuid,
                success: false,
                attempt: 0,
                error: "Failed to initialize reconnection",
            };
        }

        return {
            serverUuid,
            success: state.status === "SUCCEEDED",
            attempt: state.currentAttempt,
            error: state.lastError ?? undefined,
            reconnectedAt:
                state.status === "SUCCEEDED" ? new Date().toISOString() : undefined,
        };
    }

    cancelReconnection(serverUuid: string): void {
        const timer = this.reconnectionTimers.get(serverUuid);
        if (timer) {
            clearTimeout(timer);
            this.reconnectionTimers.delete(serverUuid);
        }

        const state = this.reconnectionStates.get(serverUuid);
        if (state) {
            state.status = "CANCELLED";
            state.nextRetryAt = null;
        }

        console.log(`Cancelled reconnection for ${serverUuid}`);
    }

    cancelAllReconnections(): void {
        for (const timer of this.reconnectionTimers.values()) {
            clearTimeout(timer);
        }
        this.reconnectionTimers.clear();

        for (const state of this.reconnectionStates.values()) {
            if (state.status === "PENDING" || state.status === "RECONNECTING") {
                state.status = "CANCELLED";
                state.nextRetryAt = null;
            }
        }

        console.log("Cancelled all reconnection attempts");
    }

    getReconnectionState(serverUuid: string): ReconnectionState | undefined {
        return this.reconnectionStates.get(serverUuid);
    }

    getAllReconnectionStates(): ReconnectionState[] {
        return Array.from(this.reconnectionStates.values());
    }

    getPendingCount(): number {
        return Array.from(this.reconnectionStates.values()).filter(
            (s) => s.status === "PENDING" || s.status === "RECONNECTING",
        ).length;
    }

    getReconnectionSummary(): {
        totalTracked: number;
        pending: number;
        reconnecting: number;
        succeeded: number;
        failed: number;
        cancelled: number;
        maxRetriesExceeded: number;
    } {
        const states = Array.from(this.reconnectionStates.values());
        return {
            totalTracked: states.length,
            pending: states.filter((s) => s.status === "PENDING").length,
            reconnecting: states.filter((s) => s.status === "RECONNECTING").length,
            succeeded: states.filter((s) => s.status === "SUCCEEDED").length,
            failed: states.filter((s) => s.status === "FAILED").length,
            cancelled: states.filter((s) => s.status === "CANCELLED").length,
            maxRetriesExceeded: states.filter(
                (s) => s.status === "MAX_RETRIES_EXCEEDED",
            ).length,
        };
    }

    resetReconnectionState(serverUuid: string): void {
        this.cancelReconnection(serverUuid);
        this.reconnectionStates.delete(serverUuid);
        console.log(`Reset reconnection state for ${serverUuid}`);
    }

    clearAllStates(): void {
        this.cancelAllReconnections();
        this.reconnectionStates.clear();
        console.log("Cleared all reconnection states");
    }

    onReconnectSuccess(listener: (serverUuid: string) => void): void {
        this.onReconnectSuccessListeners.push(listener);
    }

    onReconnectFailure(
        listener: (serverUuid: string, error: string) => void,
    ): void {
        this.onReconnectFailureListeners.push(listener);
    }

    onMaxRetriesExceeded(listener: (serverUuid: string) => void): void {
        this.onMaxRetriesExceededListeners.push(listener);
    }

    private notifyReconnectSuccess(serverUuid: string): void {
        for (const listener of this.onReconnectSuccessListeners) {
            try {
                listener(serverUuid);
            } catch (error) {
                console.error("Error in reconnect success listener:", error);
            }
        }
    }

    private notifyReconnectFailure(serverUuid: string, error: string): void {
        for (const listener of this.onReconnectFailureListeners) {
            try {
                listener(serverUuid, error);
            } catch (e) {
                console.error("Error in reconnect failure listener:", e);
            }
        }
    }

    private notifyMaxRetriesExceeded(serverUuid: string): void {
        for (const listener of this.onMaxRetriesExceededListeners) {
            try {
                listener(serverUuid);
            } catch (error) {
                console.error("Error in max retries exceeded listener:", error);
            }
        }
    }

    getConfig(): ReconnectionConfig {
        return { ...this.config };
    }
}

export const autoReconnectService = AutoReconnectService.getInstance();
