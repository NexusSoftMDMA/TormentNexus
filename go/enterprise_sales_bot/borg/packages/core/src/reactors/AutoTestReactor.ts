
import { EventBus, SystemEvent } from '../services/EventBus.js';
import { AutoTestService } from '../services/AutoTestService.js';

function getChangedPath(payload: unknown): string | null {
    if (!payload || typeof payload !== 'object') {
        return null;
    }

    const record = payload as Record<string, unknown>;
    return typeof record.path === 'string' ? record.path : null;
}

export class AutoTestReactor {
    private eventBus: EventBus;
    private autoTestService: AutoTestService;
    private isRunning: boolean = false;

    constructor(eventBus: EventBus, autoTestService: AutoTestService) {
        this.eventBus = eventBus;
        this.autoTestService = autoTestService;
    }

    public start() {
        console.log("[AutoTestReactor] Armed and listening for file changes...");
        // Listen for file changes
        this.eventBus.subscribe('file:change', this.handleFileChange.bind(this));
    }

    private async handleFileChange(event: SystemEvent) {
        const changedPath = getChangedPath(event.payload);
        if (!changedPath) {
            return;
        }

        // Only react to TS/TSX files
        if (!changedPath.match(/\.(ts|tsx)$/)) return;

        // Check if we are already running a test (debounce/throttle)
        if (this.isRunning) return;

        console.log(`[AutoTestReactor] Reacting to change in ${changedPath}`);
        this.isRunning = true;

        try {
            // Find related tests (Naive implementation: look for file with .test.ts)
            // In a real reactor, we'd use the dependency graph
            const testFile = changedPath.replace(/\.tsx?$/, '.test.ts');

            // For now, just log that we WOULD run tests
            this.eventBus.emitEvent('task:update', 'AutoTestReactor', { message: `Triggering tests for ${changedPath}` });

            // await this.autoTestService.runTests([testFile]); 
            // (AutoTestService needs to expose runTests for specific files)
        } catch (error) {
            console.error("[AutoTestReactor] Failed to react:", error);
        } finally {
            this.isRunning = false;
        }
    }
}
