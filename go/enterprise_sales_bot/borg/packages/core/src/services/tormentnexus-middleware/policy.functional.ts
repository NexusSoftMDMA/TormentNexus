import { CallToolMiddleware } from "./functional-middleware.js";
import { policyService } from "../stubs/policy.service.stub.js";

export interface PolicyMiddlewareOptions {
    enabled: boolean;
}

export const createPolicyMiddleware = (
    options: PolicyMiddlewareOptions,
): CallToolMiddleware => {
    return (next) => async (request, context) => {
        if (!options.enabled) {
            return await next(request, context);
        }

        const params = request.params as typeof request.params & {
            _meta?: { policyId?: string };
        };
        const { name } = params;

        // Check if a policy ID is attached to the request (e.g. from run_agent)
        const policyId = params._meta?.policyId;

        if (policyId) {
            try {
                const policy = await policyService.getPolicy(policyId);
                if (policy) {
                    const allowed = policyService.evaluateAccess(policy, name);
                    if (!allowed) {
                        throw new Error(`Access denied to tool '${name}' by policy '${policyId}'.`);
                    }
                } else {
                    console.warn(`Policy ID ${policyId} provided but not found.`);
                }
            } catch (error) {
                // Re-throw denial or DB errors
                throw error;
            }
        }

        // Proceed if no policy ID (open access) OR if policy allowed it
        return await next(request, context);
    };
};
