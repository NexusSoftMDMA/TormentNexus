
import { PolicyEngine } from './PolicyEngine.js';

export type AutonomyLevel = 'low' | 'medium' | 'high';

export class PermissionManager {
    public autonomyLevel: AutonomyLevel;
    private policyEngine: PolicyEngine;

    constructor(autonomyLevel: AutonomyLevel = 'high') {
        this.autonomyLevel = autonomyLevel;
        this.policyEngine = new PolicyEngine(process.cwd());
    }

    setAutonomyLevel(level: AutonomyLevel) {
        this.autonomyLevel = level;
    }

    getAutonomyLevel(): AutonomyLevel {
        return this.autonomyLevel;
    }

    /**
     * Determines if a tool call requires user approval.
     */
    checkPermission(toolName: string, args: unknown): 'APPROVED' | 'DENIED' | 'NEEDS_CONSULTATION' {
        // 1. Policy Check (Hard Guardrails)
        const policyResult = this.policyEngine.check(toolName, args);
        if (!policyResult.allowed) {
            console.warn(`[PermissionManager] Access Denied by Policy: ${policyResult.reason}`);
            return 'DENIED';
        }

        // High Autonomy: Trust completely (if Policy allows)
        if (this.autonomyLevel === 'high') {
            return 'APPROVED';
        }

        const risk = this.assessRisk(toolName, args);

        if (this.autonomyLevel === 'medium') {
            // Medium: Allow low/medium, consult on high
            if (risk === 'high') return 'NEEDS_CONSULTATION';
            return 'APPROVED';
        }

        // Low Autonomy: Block high, consult on medium, allow low
        if (risk === 'high') return 'DENIED';
        if (risk === 'medium') return 'NEEDS_CONSULTATION';
        return 'APPROVED';
    }

    private assessRisk(toolName: string, args: unknown): 'low' | 'medium' | 'high' {
        // High Risk Tools (Modifying system, network, sensitive reads)
        if (toolName.includes('write_file') ||
            toolName.includes('execute_command') ||
            toolName.includes('install') ||
            toolName.includes('git_push')) {
            return 'high';
        }

        // Medium Risk (Read-only but potentially sensitive, or minor mods)
        if (toolName.includes('read_file') || toolName.includes('list_dir') || toolName.includes('read_page')) {
            return 'medium';
        }

        // Low Risk (Info, search, ping)
        if (toolName.includes('search') || toolName.includes('status') || toolName.includes('ping')) {
            return 'low';
        }

        // Default to high risk for unknown tools
        // But let's be nicer to harmless internal tools
        if (toolName.startsWith('vscode_')) return 'medium';

        return 'high';
    }
}
