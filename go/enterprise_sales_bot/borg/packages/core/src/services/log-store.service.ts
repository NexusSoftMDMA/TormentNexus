export interface TormentNexusLogEntry {
    id: string;
    timestamp: Date;
    serverName: string;
    level: "error" | "info" | "warn";
    message: string;
    error?: string;
}

class TormentNexusLogStore {
    private logs: TormentNexusLogEntry[] = [];
    private readonly maxLogs = 1000; // Keep only the last 1000 logs
    private readonly listeners: Set<(log: TormentNexusLogEntry) => void> = new Set();

    addLog(
        serverName: string,
        level: TormentNexusLogEntry["level"],
        message: string,
        error?: unknown,
    ) {
        const logEntry: TormentNexusLogEntry = {
            id: crypto.randomUUID(),
            timestamp: new Date(),
            serverName,
            level,
            message,
            error: error
                ? error instanceof Error
                    ? error.message
                    : String(error)
                : undefined,
        };

        // Add to logs array
        this.logs.push(logEntry);

        // Keep only the last maxLogs entries
        if (this.logs.length > this.maxLogs) {
            this.logs = this.logs.slice(-this.maxLogs);
        }

        // Also log to console for debugging
        const fullMessage = `[TormentNexus][${serverName}] ${message}`;
        switch (level) {
            case "error":
                console.error(fullMessage, error || "");
                break;
            case "warn":
                console.warn(fullMessage, error || "");
                break;
            case "info":
                console.info(fullMessage, error || "");
                break;
        }

        // Notify listeners
        this.listeners.forEach((listener) => {
            try {
                listener(logEntry);
            } catch (err) {
                console.error("Error notifying log listener:", err);
            }
        });
    }

    getLogs(limit?: number): TormentNexusLogEntry[] {
        const logsToReturn = limit ? this.logs.slice(-limit) : this.logs;
        return [...logsToReturn].reverse(); // Return newest first
    }

    clearLogs(): void {
        this.logs = [];
    }

    addListener(listener: (log: TormentNexusLogEntry) => void): () => void {
        this.listeners.add(listener);
        return () => this.listeners.delete(listener);
    }

    getLogCount(): number {
        return this.logs.length;
    }
}

// Singleton instance
export const tormentnexusLogStore = new TormentNexusLogStore();
