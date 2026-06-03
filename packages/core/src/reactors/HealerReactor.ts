import { EventBus, SystemEvent } from '../services/EventBus.js';
import { HealerService } from '../services/HealerService.js';

export function shouldIgnoreExpectedStartupError(errorLog: string): boolean {
    const normalized = errorLog.toLowerCase();

    const ignoredFragments = [
        'openai api key not configured',
        'bad control character in string literal in json',
        'failed to load mcp.jsonc',
        'error fetching saved scripts',
        'error fetching mcp servers from config',
        'sqliteerror: no such table: config',
        'insufficient balance',
        'credit balance is too low',
        'quota exceeded',
        'you exceeded your current quota',
        'too many requests',
        'rate limit',
        'retry in ',
        'fetch failed',
        'econnrefused',
        'no connection could be made because the target machine actively refused it',
        'could not connect to a chroma server',
    'cannot open database because the directory does not exist',
    'bobbybookmarks db not found',
    'tormentnexusingest',
    'sqlite runtime is unavailable',
    ];

    return ignoredFragments.some((fragment) => normalized.includes(fragment));
}

function getErrorLog(payload: unknown): string {
    if (!payload || typeof payload !== 'object') {
        return '';
    }

    const record = payload as Record<string, unknown>;
    const message = typeof record.message === 'string' ? record.message : '';
    const error = typeof record.error === 'string' ? record.error : '';
    const nestedErrorMessage = record.error && typeof record.error === 'object' && typeof (record.error as Record<string, unknown>).message === 'string'
        ? String((record.error as Record<string, unknown>).message)
        : '';
    const stack = typeof record.stack === 'string' ? record.stack : '';

    return [message, error, nestedErrorMessage, stack]
        .filter((value) => typeof value === 'string' && value.trim().length > 0)
        .join('\n');
}

export class HealerReactor {
    private eventBus: EventBus;
    private healerService: HealerService;
    private isHealing: boolean = false;
    private isStopped: boolean = false;
    private lastErrorTime: number = 0;
    private consecutiveFailures: number = 0;
    private readonly BASE_COOLDOWN_MS = 10000;
    private readonly MAX_COOLDOWN_MS = 300000;
    private idleTimer: NodeJS.Timeout | null = null;
    private readonly IDLE_THRESHOLD_MS = 60000; // 1 minute idle

    constructor(eventBus: EventBus, healerService: HealerService) {
        this.eventBus = eventBus;
        this.healerService = healerService;
    }

    public start() {
        console.log("[HealerReactor] 🛡️ Immune System Active. Listening for pathogens...");
        this.eventBus.subscribe('terminal:error', this.handleError.bind(this));
        this.eventBus.subscribe('agent:stop_healing', () => {
            console.log("[HealerReactor] 🛑 StopHook received. Healing suspended.");
            this.isStopped = true;
        });
        this.eventBus.subscribe('agent:resume_healing', () => {
            console.log("[HealerReactor] ▶️ ResumeHook received. Healing resumed.");
            this.isStopped = false;
        });

        // Idle Healer: periodically check for issues when system is quiet
        this.resetIdleTimer();
        this.eventBus.subscribe('*', () => this.resetIdleTimer());
    }

    private resetIdleTimer() {
        if (this.idleTimer) clearTimeout(this.idleTimer);
        this.idleTimer = setTimeout(() => this.onIdle(), this.IDLE_THRESHOLD_MS);
    }

    private onIdle() {
        if (this.isHealing || this.isStopped) return;
        console.log("[HealerReactor] 😴 System idle. Running routine diagnostics...");
        // In the future, this could trigger 'npm test' or 'tsc' globally to find hidden issues
    }

    private async handleError(event: SystemEvent) {
        if (this.isStopped) return;

        const errorLog = getErrorLog(event.payload);

        if (shouldIgnoreExpectedStartupError(errorLog)) {
            return;
        }

        const cooldown = Math.min(
            this.BASE_COOLDOWN_MS * Math.pow(2, this.consecutiveFailures),
            this.MAX_COOLDOWN_MS
        );
        const now = Date.now();
        if (this.isHealing || (now - this.lastErrorTime < cooldown)) {
            return;
        }

        console.log(`[HealerReactor] 🩺 Detected Pathogen! Initiating immune response...`);
        this.isHealing = true;
        this.lastErrorTime = now;

        try {
            this.eventBus.emitEvent('task:update', 'HealerReactor', { message: 'Diagnosing error...' });

            // Trigger the Healer Service with the new autonomous loop
            const success = await this.healerService.healAndVerify(errorLog);

            if (success) {
                console.log(`[HealerReactor] ✅ Pathogen neutralized.`);
                this.eventBus.emitEvent('system:healed', 'HealerReactor', { status: 'success' });
                this.consecutiveFailures = 0;
            } else {
                console.log(`[HealerReactor] ⚠️ Integration failed. Could not auto-heal.`);
                this.consecutiveFailures++;
            }

        } catch (error: any) {
            console.error("[HealerReactor] ❌ Immune System Failure:", error);
            this.consecutiveFailures++;
        } finally {
            this.isHealing = false;
        }
    }
}
