import { z } from "zod";

export const ServerHealthStatusEnum = z.enum([
    "HEALTHY",
    "UNHEALTHY",
    "UNKNOWN",
    "CHECKING",
]);

export type ServerHealthStatus = z.infer<typeof ServerHealthStatusEnum>;

export interface HealthCheckResult {
    serverUuid: string;
    success: boolean;
    responseTimeMs: number;
    toolCount?: number;
    errorMessage?: string;
    checkedAt: string;
}

export interface ServerHealthInfo {
    serverUuid: string;
    serverName: string;
    status: ServerHealthStatus;
    lastChecked: string | null;
    lastHealthy: string | null;
    responseTimeMs: number | null;
    errorMessage: string | null;
    consecutiveFailures: number;
    toolCount: number | null;
}
