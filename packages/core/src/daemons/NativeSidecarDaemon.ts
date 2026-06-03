import { EventBus } from '../services/EventBus.js';
import { a2aBroker } from '@tormentnexus/agents';

export class NativeSidecarDaemon {
    private lastSeenTimestamp: number = 0;
    private reconnectTimeout: NodeJS.Timeout | null = null;
    private retryCount: number = 0;
    private readonly maxRetryDelay = 30000;

    constructor(
        private readonly eventBus: EventBus,
        private readonly sidecarUrl: string = process.env.TORMENTNEXUS_SIDECAR_URL || 'http://localhost:4300'
    ) {}

    async start() {
        console.log(`[NativeSidecarDaemon] Starting bridge to Go sidecar at ${this.sidecarUrl}`);
        this.connect();
    }

    private connect() {
        // In a real implementation, we'd use EventSource or a persistent fetch loop
        // for the /api/sse endpoint. For this task, we focus on the replay logic.
        this.pollHistory();
    }

    private async pollHistory() {
        try {
            const url = `${this.sidecarUrl}/api/sse/history?since=${this.lastSeenTimestamp}`;
            const response = await fetch(url);
            if (!response.ok) throw new Error(`HTTP ${response.status}`);

            const result = await response.json();
            if (result.success && Array.isArray(result.data)) {
                for (const event of result.data) {
                    this.handleEvent(event);
                    this.lastSeenTimestamp = Math.max(this.lastSeenTimestamp, event.timestamp);
                }
            }

            this.retryCount = 0;
        } catch (error) {
            this.retryCount++;
            const delay = Math.min(1000 * Math.pow(2, this.retryCount), this.maxRetryDelay);
            console.warn(`[NativeSidecarDaemon] Bridge error, retrying in ${delay}ms:`, (error as Error).message);

            this.reconnectTimeout = setTimeout(() => this.pollHistory(), delay);
            return;
        }

        this.reconnectTimeout = setTimeout(() => this.pollHistory(), 5000);
    }

    private handleEvent(event: any) {
        if (event.type === 'a2a:signal') {
            const payload = event.payload;
            if (payload && payload.message) {
                // Skip heartbeat logging to reduce noise (frequent and expected)
                if (payload.message.type !== 'HEARTBEAT') {
                console.log(`[NativeSidecarDaemon] Replaying A2A signal: ${payload.message.type}`);
                }
                a2aBroker.routeMessage(payload.message).catch(e =>
                    console.error('[NativeSidecarDaemon] Failed to route replayed message:', e)
                );
            }
        }
    }

    stop() {
        if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);
    }
}
