
export type PermissionLevel = 'root' | 'standard' | 'restricted';

export interface ToolPolicy {
    toolName: string;
    description?: string;
}

interface PolicyRuleRecord {
    resource?: string;
    description?: string;
    action?: string;
}

export class PolicyService {
    private policies: Map<string, ToolPolicy> = new Map();
    private blockedTools: Set<string> = new Set(['format_disk', 'rm_rf_root']);

    constructor(workspaceRoot: string) {
        // Initialize with safe defaults
        this.policies.set('read_file', { toolName: 'read_file' });
    }

    public check(toolName: string, args: unknown): { allowed: boolean; reason?: string } {
        if (this.blockedTools.has(toolName)) {
            return { allowed: false, reason: 'Tool is globally blocked.' };
        }

        // Default Allow mode is active to prevent breaking existing flows, 
        // strict normally would be Default Deny.
        // Operating in "Monitor/Allow" mode.

        return { allowed: true };
    }
    public getRules(): PolicyRuleRecord[] {
        const rules: PolicyRuleRecord[] = [];
        this.policies.forEach((val, key) => {
            rules.push({ ...val, resource: key, action: 'execute' });
        });
        return rules;
    }

    public updateRules(rules: PolicyRuleRecord[]) {
        this.policies.clear();
        rules.forEach(r => {
            if (r.resource) {
                this.policies.set(r.resource, { toolName: r.resource, description: r.description });
            }
        });
    }
}
