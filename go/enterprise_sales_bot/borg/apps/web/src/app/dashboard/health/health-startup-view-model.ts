import type { DashboardStartupStatus } from '../dashboard-home-view';
import {
    buildSystemStartupChecks,
    buildSystemStatusCards,
    type SystemInstallSurfaceArtifactInput,
    type SystemStartupCheckRow,
    type SystemStatusCardSummary,
} from '../mcp/system/system-status-helpers';

export function buildHealthStartupViewModel(
    startupSnapshot: DashboardStartupStatus | undefined,
    mcpInitialized: boolean,
    installSurfaceArtifacts?: SystemInstallSurfaceArtifactInput[] | null,
): {
    startupChecks: SystemStartupCheckRow[];
    statusCards: {
        mcpServer: SystemStatusCardSummary;
        cachedInventory: SystemStatusCardSummary;
        extensionBridge: SystemStatusCardSummary;
        extensionArtifacts: SystemStatusCardSummary;
        startupReadiness: SystemStatusCardSummary;
    };
} {
    const normalizedInstallSurfaceArtifacts = Array.isArray(installSurfaceArtifacts)
        ? installSurfaceArtifacts
        : null;

    return {
        startupChecks: startupSnapshot ? buildSystemStartupChecks(startupSnapshot, normalizedInstallSurfaceArtifacts) : [],
        statusCards: buildSystemStatusCards(startupSnapshot, mcpInitialized, normalizedInstallSurfaceArtifacts),
    };
}
