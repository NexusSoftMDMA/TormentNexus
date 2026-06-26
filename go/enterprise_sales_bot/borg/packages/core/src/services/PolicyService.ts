
export type PermissionLevel = 'root' | 'standard' | 'restricted';

export interface ToolPolicy {
    toolName: string;
    allowedRoles: PermissionLevel[];
    requiresConfirmation: boolean;
}

export class PolicyService {
    private policies: Map<string, ToolPolicy> = new Map();

    constructor() {
        // Default Policies
        this.addPolicy('write_file', ['root', 'standard'], true); // Risky
        this.addPolicy('run_command', ['root'], true); // Very Risky
        this.addPolicy('replace_file_content', ['root', 'standard'], true); // Risky

        // Safe Tools
        const safeTools = ['read_file', 'list_files', 'search_codebase', 'research_topic', 'chat_reply', 'vscode_get_status'];
        safeTools.forEach(t => this.addPolicy(t, ['root', 'standard', 'restricted'], false));
    }

    private addPolicy(tool: string, roles: PermissionLevel[], confirm: boolean) {
        this.policies.set(tool, { toolName: tool, allowedRoles: roles, requiresConfirmation: confirm });
    }

    public checkPermission(role: PermissionLevel, toolName: string): { allowed: boolean; reason?: string; requiresConfirmation?: boolean } {
        const policy = this.policies.get(toolName);

        // If no policy exists, assume restricted/unknown tools are unsafe -> require root?
        // Or default deny? Let's default ALLOW for beta velocity, but log warning.
        if (!policy) {
            return { allowed: true, requiresConfirmation: false };
        }

        if (!policy.allowedRoles.includes(role)) {
            return { allowed: false, reason: `Role '${role}' is not allowed to use '${toolName}'.` };
        }

        return { allowed: true, requiresConfirmation: policy.requiresConfirmation };
    }
}
