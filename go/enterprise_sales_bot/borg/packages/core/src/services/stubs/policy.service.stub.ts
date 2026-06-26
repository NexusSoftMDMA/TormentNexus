export interface Policy { }

export const policyService = {
    getPolicy: async (_policyId: string): Promise<Policy | null> => {
        // Compatibility placeholder: no persisted policy backend is active in this runtime path.
        // Returning null keeps callers operational while policy storage/evaluation is being implemented.
        return null;
    },
    evaluateAccess: (_policy: unknown, _toolName: string): boolean => {
        // Compatibility placeholder: pass-through evaluator.
        // NOTE: This does not enforce allow/deny semantics yet.
        // UI surfaces should communicate that policy enforcement is currently in audit/preview mode.
        return true;
    }
};
