/**
 * Service to connect to and orchestrate the Bytedance DeerFlow Engine.
 * Note: The DeerFlow backend runs a Python FastAPI gateway locally on port 8001 (or 2026 via Nginx).
 */
export class DeerFlowBridgeService {
    private baseUrl: string;

    constructor() {
        // Ensure this maps to the Gateway API (8001) or the unified Nginx proxy if available
        this.baseUrl = process.env.DEER_FLOW_API_URL || 'http://localhost:8001/api';
    }

    /**
     * Checks if the DeerFlow backend is accessible.
     */
    async isAvailable(): Promise<boolean> {
        try {
            const res = await fetch(`${this.baseUrl}/models`, {
                method: 'GET',
                signal: AbortSignal.timeout(2000), // 2s timeout
            });
            return res.ok;
        } catch (e) {
            return false;
        }
    }

    /**
     * Retrieves the available LLM models configured in DeerFlow's config.yaml.
     */
    async getModels(): Promise<any[]> {
        const res = await fetch(`${this.baseUrl}/models`);
        if (!res.ok) throw new Error(`DeerFlow API Error: ${res.statusText}`);
        return await res.json();
    }

    /**
     * Retrieves the list of available DeerFlow skills.
     */
    async getSkills(): Promise<any[]> {
        const res = await fetch(`${this.baseUrl}/skills`);
        if (!res.ok) throw new Error(`DeerFlow API Error: ${res.statusText}`);
        return await res.json();
    }

    /**
     * Retrieves the persistent user memory context stored by DeerFlow across conversations.
     */
    async getMemoryStatus(): Promise<any> {
        const res = await fetch(`${this.baseUrl}/memory/status`);
        if (!res.ok) throw new Error(`DeerFlow API Error: ${res.statusText}`);
        return await res.json();
    }
}

// Export singleton instance
export const deerFlowBridgeService = new DeerFlowBridgeService();
