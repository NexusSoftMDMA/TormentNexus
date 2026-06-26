
import { EventBus } from '../services/EventBus.js';

function shouldIgnoreInternalDiagnostic(message: string): boolean {
    const trimmed = message.trim();

    if (!trimmed) {
        return true;
    }

    const internalPrefixes = [
        '[HealerReactor]',
        '[TerminalSensor]',
        '[FileSensor]',
        '[EventBus]',
        '[TormentNexus Core] Unhandled promise rejection:',
        '[TormentNexus Core] Uncaught exception:',
    ];

    return internalPrefixes.some((prefix) => trimmed.startsWith(prefix));
}

export function splitTerminalSensorBuffer(buffer: string, force: boolean = false): { lines: string[]; remaining: string } {
    const normalized = buffer.replace(/\r\n/g, '\n');
    const parts = normalized.split('\n');
    const trailing = parts.pop() ?? '';

    if (force) {
        return {
            lines: trailing ? [...parts, trailing] : parts,
            remaining: '',
        };
    }

    return {
        lines: parts,
        remaining: trailing,
    };
}

export class TerminalSensor {
    private eventBus: EventBus;
    private originalStderrWrite: typeof process.stderr.write | null = null;
    private isInterceptingStderr = false;
    private stderrBuffer = '';

    constructor(eventBus: EventBus) {
        this.eventBus = eventBus;
    }

    private emitBufferedErrorLine(line: string): void {
        const normalizedLine = line.trim();
        if (!normalizedLine) {
            return;
        }

        const lowered = normalizedLine.toLowerCase();
        if (!shouldIgnoreInternalDiagnostic(normalizedLine) && (lowered.includes('error') || lowered.includes('exception') || normalizedLine.includes('❌'))) {
            this.eventBus.emitEvent('terminal:error', 'TerminalSensor', { message: normalizedLine });
        }
    }

    private flushStderrBuffer(force: boolean = false): void {
        if (!this.stderrBuffer) {
            return;
        }

        const { lines, remaining } = splitTerminalSensorBuffer(this.stderrBuffer, force);

        for (const line of lines) {
            this.emitBufferedErrorLine(line);
        }

        this.stderrBuffer = remaining;
    }

    private captureStderrChunk(chunk: string): void {
        this.stderrBuffer += chunk;
        this.flushStderrBuffer(false);
    }

    public start() {
        console.log("[TerminalSensor] Hooking stderr...");

        // Hook stderr to capture errors
        this.originalStderrWrite = process.stderr.write.bind(process.stderr);

        /**
         * Reason: we need to observe terminal error output in real-time for event emission.
         * What: wraps stderr writes, emits `terminal:error` on heuristic matches, then forwards bytes unchanged.
         * Why: preserves normal stderr behavior while enabling lightweight telemetry without unsafe casts.
         */
        const patchedWrite: typeof process.stderr.write = (chunk, encoding?, callback?) => {
            if (this.isInterceptingStderr) {
                if (!this.originalStderrWrite) {
                    return false;
                }

                return this.originalStderrWrite(chunk, encoding as never, callback as never);
            }

            const str = chunk.toString();
            this.isInterceptingStderr = true;

            try {
                this.captureStderrChunk(str);

                // Pass through to original stderr
                if (!this.originalStderrWrite) {
                    return false;
                }

                return this.originalStderrWrite(chunk, encoding as never, callback as never);
            } finally {
                this.isInterceptingStderr = false;
            }
        };

        process.stderr.write = patchedWrite;
    }

    public stop() {
        this.flushStderrBuffer(true);
        if (this.originalStderrWrite) {
            process.stderr.write = this.originalStderrWrite;
            this.originalStderrWrite = null;
        }

    }
}
